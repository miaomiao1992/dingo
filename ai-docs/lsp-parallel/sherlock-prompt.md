# Dingo LSP Architecture Investigation

**Problem Statement**: LSP server crashes immediately on startup, stream destroyed, no process survives.

**Key Observation**: Unit tests pass (all 34+ tests), but binary exits immediately when run.

---

## Architecture Overview

### Two-Layer Proxy Design

```
IDE Client
    ↓ (JSON-RPC over stdio)
dingo-lsp server
    ├─ Translator (position mapping)
    ├─ FileWatcher (auto-transpile)
    ├─ AutoTranspiler
    ├─ SourceMapCache
    └─ GoplsClient (subprocess)
            ↓ (JSON-RPC over pipes)
         gopls (subprocess)
```

### Main Entry Point (cmd/dingo-lsp/main.go)

```go
// Lines 13-56
func main() {
    logger := lsp.NewLogger(logLevel, os.Stderr)

    // Create server instance
    server, err := lsp.NewServer(ServerConfig{...})

    // Create stdio transport
    rwc := &stdinoutCloser{stdin: os.Stdin, stdout: os.Stdout}
    stream := jsonrpc2.NewStream(rwc)
    conn := jsonrpc2.NewConn(stream)

    // Start serving
    ctx := context.Background()
    handler := server.Handler()
    conn.Go(ctx, handler)                    // ← START HANDLER

    server.SetConn(conn, ctx)                // ← SET CONNECTION

    // Wait for connection to close
    <-conn.Done()                            // ← WAIT HERE
    logger.Infof("Server stopped")
}
```

### Server Creation (pkg/lsp/server.go lines 35-60)

```go
func NewServer(cfg ServerConfig) (*Server, error) {
    gopls, err := NewGoplsClient(cfg.GoplsPath, cfg.Logger)  // ← CREATE GOPLS
    if err != nil {
        return nil, fmt.Errorf("failed to start gopls: %w", err)
    }

    mapCache, err := NewSourceMapCache(cfg.Logger)
    translator := NewTranslator(mapCache)
    transpiler := NewAutoTranspiler(cfg.Logger, mapCache, gopls)

    return &Server{
        config: cfg,
        gopls: gopls,
        ...
    }, nil
}
```

### GoplsClient Creation (pkg/lsp/gopls_client.go lines 29-46)

```go
func NewGoplsClient(goplsPath string, logger Logger) (*GoplsClient, error) {
    client := &GoplsClient{
        logger: logger,
        goplsPath: goplsPath,
        maxRestarts: 3,
    }

    if err := client.start(); err != nil {
        return nil, err
    }

    return client, nil
}
```

### GoplsClient Start (pkg/lsp/gopls_client.go lines 49-111)

```go
func (c *GoplsClient) start() error {
    c.mu.Lock()
    defer c.mu.Unlock()

    // Start gopls subprocess
    c.cmd = exec.Command(c.goplsPath, "-mode=stdio")

    stdin, err := c.cmd.StdinPipe()        // ← CREATE PIPES
    stdout, err := c.cmd.StdoutPipe()
    stderr, err := c.cmd.StderrPipe()

    if err := c.cmd.Start(); err != nil {  // ← START PROCESS
        return fmt.Errorf("failed to start gopls: %w", err)
    }

    // Log stderr in background
    go c.logStderr(stderr)

    // Create JSON-RPC connection
    rwc := &readWriteCloser{stdin: stdin, stdout: stdout}
    stream := jsonrpc2.NewStream(rwc)
    c.conn = jsonrpc2.NewConn(stream)

    // Start handler
    ctx := context.Background()
    handler := jsonrpc2.ReplyHandler(func(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
        c.logger.Debugf("gopls notification/request: %s", req.Method())
        return nil
    })
    c.conn.Go(ctx, handler)                // ← START GOPLS HANDLER

    // Monitor process exit
    go func() {
        err := c.cmd.Wait()                // ← WAIT FOR PROCESS
        ...
    }()

    return nil
}
```

---

## Root Cause Analysis

### Issue 1: Connection Lifecycle Race Condition

**The Fatal Problem**: In `main()`, there's a critical ordering issue:

```go
conn.Go(ctx, handler)                      // Line 49: Start IDE handler
server.SetConn(conn, ctx)                  // Line 52: Store connection for later
<-conn.Done()                              // Line 55: Wait for connection
```

**What happens**:

1. `conn.Go(ctx, handler)` spawns a goroutine that starts reading from stdin
2. In a normal editor LSP scenario, the client sends `initialize` request
3. **BUT** if there's no client connected (testing the binary alone), stdin is closed or produces EOF
4. The jsonrpc2 stream reads EOF from stdin
5. The handler goroutine exits
6. `conn.Done()` closes immediately
7. Main function returns and process terminates

### Issue 2: Uninitialized Connection in Handlers

**The Secondary Problem**: The handlers need the IDE connection to send diagnostics:

```go
// handlers.go:312-316 (incomplete)
func (s *Server) handlePublishDiagnostics(...) {
    // TODO: Actually send notification to IDE connection
    // Currently just translates but doesn't publish
}
```

Even if the connection stays alive, handlers can't send diagnostics back to IDE because `server.SetConn()` happens AFTER `conn.Go()`.

### Issue 3: Context Lifecycle Mismatch

**GoplsClient context issue**:
```go
// gopls_client.go:85
ctx := context.Background()            // ← BACKGROUND CONTEXT
handler := jsonrpc2.ReplyHandler(...)
c.conn.Go(ctx, handler)                // ← PASSED HERE
```

The gopls connection handler uses `context.Background()`, which is never cancelled. When the IDE connection closes, gopls handler keeps running with a background context that can't be cancelled.

### Issue 4: Stream Destruction Timing

**The immediate crash happens because**:

1. `jsonrpc2.NewStream(rwc)` wraps stdin/stdout
2. When stdin reaches EOF (no IDE client), stream detects it
3. Stream signals `conn.Done()`
4. Main goroutine unblocks from `<-conn.Done()`
5. Main function returns
6. Process terminates

**Why it says "stream destroyed"**: The Go jsonrpc2 library logs this when it reads EOF from the underlying transport while the connection is active.

---

## Why Unit Tests Pass But Binary Crashes

**Unit tests pass** because they:
- Don't actually start the dingo-lsp binary
- Test components in isolation (GoplsClient, Translator, etc.)
- Mock the connections
- Never exercise the main() function's stream/connection lifecycle

**Binary crashes** because:
- main() blocks on `<-conn.Done()`
- With no IDE client sending requests, stdin closes or sends EOF
- jsonrpc2 stream detects EOF
- Handler goroutine exits
- conn.Done() closes
- Main function returns
- Process terminates

---

## Simplest Root Cause

**The fundamental architectural flaw**:

The main.go assumes the LSP server will ALWAYS have a connected IDE client sending requests. It treats stdio as a persistent connection. However:

1. When testing manually, there's no client
2. When the IDE disconnects, stdin closes
3. The jsonrpc2 stream sees EOF
4. The handler exits gracefully
5. The main goroutine unblocks
6. The process terminates (which is actually correct LSP behavior)

**The real problem**: The process is designed to exit when the IDE disconnects. But the error message suggests it's crashing unexpectedly.

---

## Why It's Not a Bug (In Some Sense)

LSP servers are supposed to:
1. Start
2. Read requests from stdin
3. Write responses to stdout
4. Exit when the client disconnects

The current implementation does exactly this. **It's not crashing—it's shutting down normally when the IDE disconnects.**

But the symptom "stream destroyed" suggests there's confusion about whether this is normal shutdown or an error.

---

## Key Files and Critical Sections

| File | Lines | Issue |
|------|-------|-------|
| `cmd/dingo-lsp/main.go` | 44-56 | Stream/connection lifecycle |
| `pkg/lsp/server.go` | 35-60 | NewServer initialization order |
| `pkg/lsp/server.go` | 63-66 | SetConn called after Go() |
| `pkg/lsp/gopls_client.go` | 49-111 | GoplsClient lifecycle |
| `pkg/lsp/gopls_client.go` | 85-91 | Handler context is Background |
| `handlers.go` | 312-316 | Diagnostic publishing not implemented |

---

## Minimum Test Case

To verify the crash:
```bash
./dingo-lsp << EOF
EOF
```

This sends immediate EOF on stdin. If the process exits immediately, it's not a crash—it's normal shutdown.

To verify normal operation, use an LSP client that sends initialize request:
```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"rootPath":"/tmp","rootUri":"file:///tmp"}}' | ./dingo-lsp
```

If the process handles this and waits for more requests, the connection handling is correct.

---

## Diagnostic Summary

1. **Is it actually crashing?** Probably not—it's shutting down when stdin closes.
2. **Is the error message misleading?** Yes—"stream destroyed" sounds like a crash but is normal EOF handling.
3. **What would fix it?**
   - Implement proper graceful shutdown on EOF
   - Add logging to distinguish normal shutdown from errors
   - Test with actual LSP client (VS Code extension)
   - Ensure gopls process is also properly cleaned up on shutdown

---

## Hypothesis

**Most Likely Root Cause**: The process works correctly but exhibits normal shutdown behavior when no IDE client is connected. The "stream destroyed" message is the jsonrpc2 library logging EOF on the stream. This is not a crash—it's the designed behavior.

**Confirmation needed**: Run against actual LSP client (VS Code extension) and observe if it works correctly.

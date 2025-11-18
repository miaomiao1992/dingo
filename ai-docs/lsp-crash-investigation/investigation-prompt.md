# LSP JSON-RPC Connection Crash Investigation

## Problem Statement

We have a Language Server Protocol (LSP) proxy server for Dingo (a Go meta-language) that immediately crashes on startup with:

```
Error: Cannot call write after a stream was destroyed
```

The LSP wraps gopls (Go's language server) and translates between `.dingo` and `.go` files using source maps.

## Architecture

```
VS Code Extension
    ↓ (JSON-RPC over stdio)
dingo-lsp Server
    ↓ (JSON-RPC over stdio)
gopls Subprocess
```

The dingo-lsp acts as both:
1. **Server** for VS Code (receives LSP requests)
2. **Client** for gopls (forwards translated requests)

## Current Implementation

### 1. Main Entry (`cmd/dingo-lsp/main.go`)

```go
func main() {
    // ... initialization ...

    // Create stdio transport using ReadWriteCloser wrapper
    rwc := &stdinoutCloser{stdin: os.Stdin, stdout: os.Stdout}
    stream := jsonrpc2.NewStream(rwc)
    conn := jsonrpc2.NewConn(stream)

    // Start serving
    ctx := context.Background()

    // Create handler and start connection
    handler := server.Handler()
    conn.Go(ctx, handler)

    // Store connection in server for use by handlers
    server.SetConn(conn, ctx)

    // Wait for connection to close
    <-conn.Done()
    logger.Infof("Server stopped")
}

type stdinoutCloser struct {
    stdin  *os.File
    stdout *os.File
}

func (s *stdinoutCloser) Read(p []byte) (n int, err error) {
    return s.stdin.Read(p)
}

func (s *stdinoutCloser) Write(p []byte) (n int, err error) {
    return s.stdout.Write(p)
}

func (s *stdinoutCloser) Close() error {
    // Don't actually close stdin/stdout
    return nil
}
```

### 2. Server Handler (`pkg/lsp/server.go`)

```go
// Handler returns a jsonrpc2 handler for this server
func (s *Server) Handler() jsonrpc2.Handler {
    return jsonrpc2.ReplyHandler(s.handleRequest)
}

// handleRequest routes LSP requests to appropriate handlers
func (s *Server) handleRequest(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
    s.config.Logger.Debugf("Received request: %s", req.Method())

    switch req.Method() {
    case "initialize":
        return s.handleInitialize(ctx, reply, req)
    // ... other cases ...
    }
}

// handleInitialize processes the initialize request
func (s *Server) handleInitialize(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
    s.config.Logger.Debugf("handleInitialize: Starting")

    var params protocol.InitializeParams
    if err := json.Unmarshal(req.Params(), &params); err != nil {
        s.config.Logger.Errorf("handleInitialize: Failed to unmarshal params: %v", err)
        return reply(ctx, nil, fmt.Errorf("invalid initialize params: %w", err))
    }
    s.config.Logger.Debugf("handleInitialize: Params unmarshaled")

    // Forward initialize to gopls
    s.config.Logger.Debugf("handleInitialize: Forwarding to gopls")
    goplsResult, err := s.gopls.Initialize(ctx, params)
    if err != nil {
        s.config.Logger.Errorf("handleInitialize: gopls failed: %v", err)
        return reply(ctx, nil, fmt.Errorf("gopls initialize failed: %w", err))
    }
    s.config.Logger.Debugf("handleInitialize: gopls responded")

    // ... modify result ...

    return reply(ctx, result, nil)
}
```

### 3. Gopls Client (`pkg/lsp/gopls_client.go`)

```go
func (c *GoplsClient) start() error {
    c.mu.Lock()
    defer c.mu.Unlock()

    // Start gopls subprocess with -mode=stdio
    c.cmd = exec.Command(c.goplsPath, "-mode=stdio")

    stdin, err := c.cmd.StdinPipe()
    // ... error handling ...

    stdout, err := c.cmd.StdoutPipe()
    // ... error handling ...

    // Start gopls
    if err := c.cmd.Start(); err != nil {
        return fmt.Errorf("failed to start gopls: %w", err)
    }

    // Create JSON-RPC connection using a ReadWriteCloser wrapper
    rwc := &readWriteCloser{stdin: stdin, stdout: stdout}
    stream := jsonrpc2.NewStream(rwc)
    c.conn = jsonrpc2.NewConn(stream)

    // Start handler to process gopls responses and notifications
    ctx := context.Background()
    handler := jsonrpc2.ReplyHandler(func(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
        // Log gopls -> dingo-lsp notifications/requests (if any)
        c.logger.Debugf("gopls notification/request: %s", req.Method())
        return nil // Gopls typically doesn't send requests to client, just responses
    })
    c.conn.Go(ctx, handler)

    c.logger.Infof("gopls started (PID: %d)", c.cmd.Process.Pid)

    return nil
}

// Initialize sends initialize request to gopls
func (c *GoplsClient) Initialize(ctx context.Context, params protocol.InitializeParams) (*protocol.InitializeResult, error) {
    var result protocol.InitializeResult
    _, err := c.conn.Call(ctx, "initialize", params, &result)
    if err != nil {
        return nil, fmt.Errorf("gopls initialize failed: %w", err)
    }
    return &result, nil
}
```

## Observed Behavior

**Logs:**
```
[INFO] Starting dingo-lsp server (log level: debug)
[INFO] Found gopls at: /Users/jack/go/bin/gopls
[INFO] gopls started (PID: 12345)
[DEBUG] Received request: initialize
[DEBUG] handleInitialize: Starting
[DEBUG] handleInitialize: Params unmarshaled
[DEBUG] handleInitialize: Forwarding to gopls
[... hangs here, never sees "gopls responded" ...]
```

**VS Code Error:**
```
Error: Cannot call write after a stream was destroyed
```

**Process State:**
- dingo-lsp process exits immediately
- gopls subprocess also exits
- No panic, no error logs

## Key Observations

1. **The call to `s.gopls.Initialize(ctx, params)` appears to hang** - we never see the "gopls responded" log
2. **VS Code reports stream destroyed** - suggests the stdio connection is closing prematurely
3. **Both processes exit cleanly** - no crash dumps or panics
4. **gopls starts successfully** - we see "gopls started (PID: ...)"

## Library Used

We're using `go.lsp.dev/jsonrpc2` v0.14.0, which is the official Go LSP library.

Key types:
- `jsonrpc2.Conn` - JSON-RPC connection
- `jsonrpc2.Handler` - Request handler interface
- `jsonrpc2.Replier` - Reply callback function
- `conn.Go(ctx, handler)` - Starts processing in background goroutine
- `conn.Call(ctx, method, params, result)` - Synchronous RPC call
- `conn.Notify(ctx, method, params)` - Fire-and-forget notification

## Questions for Expert Analysis

1. **JSON-RPC Connection Architecture**:
   - Is it correct to have TWO separate `jsonrpc2.Conn` instances (one for VS Code, one for gopls)?
   - Should we be sharing a single connection somehow?
   - Is the pattern of "proxy server that's both server and client" handled correctly?

2. **Call Method Hang**:
   - Why would `conn.Call(ctx, "initialize", params, &result)` hang indefinitely?
   - The gopls process is running and should respond to initialize
   - Is there a deadlock or context issue?

3. **Context Management**:
   - We use `context.Background()` in multiple places
   - Should we be using the same context everywhere?
   - Could there be a context cancellation causing the hang?

4. **Handler Pattern**:
   - Is `jsonrpc2.ReplyHandler(func)` the correct pattern for both server and client handlers?
   - Should the gopls client handler do more than just log?
   - Could the handler be blocking the connection somehow?

5. **Stream Destruction**:
   - Why does VS Code report "stream was destroyed"?
   - Is something closing stdin/stdout prematurely?
   - Could `conn.Go()` be causing an early exit?

6. **Goroutine Management**:
   - We call `conn.Go(ctx, handler)` twice (once for VS Code conn, once for gopls conn)
   - Is this creating a race condition or deadlock?
   - Should we be waiting for something before calling `<-conn.Done()`?

## What We Need

Please provide:

1. **Root Cause**: What's causing the hang in `gopls.Initialize()` and the stream destruction?

2. **Architecture Guidance**: How should a JSON-RPC proxy (server + client) be structured with `go.lsp.dev/jsonrpc2`?

3. **Specific Code Fixes**: What changes do we need to make to:
   - Main entry point setup
   - Server handler implementation
   - Gopls client connection management

4. **Best Practices**:
   - Context usage patterns
   - Handler patterns for proxy scenarios
   - Connection lifecycle management

## Additional Context

- This is a transpiler that converts `.dingo` → `.go` files
- We need position translation via source maps (not implemented yet, but shouldn't affect basic connection)
- The LSP should work like templ's gopls proxy (github.com/a-h/templ)
- We're following TypeScript's architecture pattern for meta-languages

Thank you for your expert analysis!

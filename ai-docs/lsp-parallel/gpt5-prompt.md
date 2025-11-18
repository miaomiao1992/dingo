# Gopls Subprocess Pipe Management Analysis

## Context

We have a Go LSP proxy that wraps gopls in a subprocess. The initialization call hangs, then crashes with "destroyed stream" error. Current implementation uses:
- `exec.Cmd` with StdinPipe/StdoutPipe/StderrPipe
- `go.lsp.dev/jsonrpc2` for JSON-RPC communication
- Custom `readWriteCloser` wrapper combining stdin/stdout
- Background goroutine monitoring process exit
- Crash recovery with restart logic

## Critical Issue: Initialization Hangs Then Crashes

**Symptoms:**
1. Initialize call to gopls hangs (no response received)
2. Eventually times out and crashes with "destroyed stream" error
3. Process exit detection seems to work (crash recovery attempts restart)
4. Stderr logging appears functional

## Code Analysis Needed

Please analyze `/Users/jack/mag/dingo/pkg/lsp/gopls_client.go` and specifically address:

### 1. **Pipe Lifecycle Management**

```go
// Lines 56-69: Pipes created
stdin, _ := c.cmd.StdinPipe()
stdout, _ := c.cmd.StdoutPipe()
stderr, _ := c.cmd.StderrPipe()

// Lines 80: Wrapped in custom ReadWriteCloser
rwc := &readWriteCloser{stdin: stdin, stdout: stdout}
stream := jsonrpc2.NewStream(rwc)
```

**Questions:**
- What happens to pipes when gopls process exits unexpectedly?
- Are the pipes properly closed when connection closes?
- Could there be a race condition between stderr reading and stdout/stdin closure?
- Is the `readWriteCloser.Close()` method sufficient?

### 2. **JSON-RPC Stream Initialization**

```go
// Lines 81-91: Stream and connection setup
stream := jsonrpc2.NewStream(rwc)
c.conn = jsonrpc2.NewConn(stream)
ctx := context.Background()
handler := jsonrpc2.ReplyHandler(...)
c.conn.Go(ctx, handler)
```

**Questions:**
- When does `jsonrpc2.NewStream(rwc)` start reading from the stream?
- Could initialization request be sent before gopls is ready?
- Is the handler properly buffering responses?
- What's the backpressure behavior when gopls is slow?

### 3. **Blocking/Buffering in readWriteCloser**

```go
// Lines 274-295: Custom wrapper
type readWriteCloser struct {
    stdin  io.WriteCloser
    stdout io.ReadCloser
}

func (rwc *readWriteCloser) Read(p []byte) (n int, err error) {
    return rwc.stdout.Read(p)
}

func (rwc *readWriteCloser) Write(p []byte) (n int, err error) {
    return rwc.stdin.Write(p)
}
```

**Questions:**
- Should writes to stdin be buffered? Currently they're unbuffered direct writes
- Should reads from stdout be buffered? Currently they're unbuffered direct reads
- Could unbuffered I/O cause deadlock if pipe buffers fill?
- Is there sufficient capacity for JSON-RPC messages?

### 4. **Stderr Background Goroutine Race**

```go
// Line 77: Started immediately after Start()
go c.logStderr(stderr)

// Lines 96-109: Process exit monitoring (also background)
go func() {
    err := c.cmd.Wait()
    ...
}()
```

**Questions:**
- Is there a race condition reading stderr while checking process state?
- Could stderr goroutine hold a reference that prevents proper cleanup?
- What if stderr goroutine is blocked trying to read when process exits?
- Should stderr reading be cancellable?

### 5. **Connection State During Hang**

```go
// Line 142: Initialize call
_, err := c.conn.Call(ctx, "initialize", params, &result)
```

**Questions:**
- What state is the connection in when this hangs?
- Is jsonrpc2 waiting for a response that gopls is trying to send but can't?
- Could there be an issue with stdin/stdout buffer synchronization?
- Is there any way to detect if gopls is alive during the hang?

### 6. **Potential Deadlock Scenario**

**Hypothesis:**
1. Initialize request is written to stdin (line 142)
2. jsonrpc2 waits for response from stdout
3. But gopls's stdout buffer fills up
4. gopls tries to write to stderr (log message)
5. stderr buffer fills up
6. gopls hangs trying to write stderr
7. Therefore can't read stdin or process Initialize request
8. Deadlock: dingo-lsp waits for response, gopls waits for buffer space

**Could this be happening?**
- Are pipes created with sufficient buffering for this scenario?
- Should stderr reading be prioritized/non-blocking?
- Should there be a separate goroutine for reading stdout that doesn't block on Initialize?

## Research Areas

Please investigate and provide:

### A. **How go.lsp.dev/jsonrpc2 handles I/O**
- Does it use buffered or unbuffered reads?
- What are the threading/blocking assumptions?
- How does it handle slow readers?
- Any known issues with pipe-based communication?

### B. **Correct Pattern for gopls Subprocess**
- How do other LSP proxies (templ, etc.) handle gopls?
- What's the idiomatic way to pipe JSON-RPC over subprocess stdio?
- Should we use StdinPipe/StdoutPipe or direct file descriptor manipulation?
- Any gotchas with exec.Cmd pipes in Go?

### C. **Deadlock Prevention Strategies**
- Should stdout reading happen in a separate goroutine (not jsonrpc2's handler)?
- Should stderr reading be non-blocking or use a separate buffer?
- What's the safe way to close pipes when process exits?
- How to detect hanging vs normal blocking?

### D. **Specific Go Pipe Behavior**
- When exec.Cmd.Start() returns, are pipes immediately ready for I/O?
- What happens to unreceived data if process exits suddenly?
- Does closing a pipe trigger EOF or error?
- How does pipe buffer size affect multi-line JSON-RPC messages?

### E. **jsonrpc2.Conn.Call() Behavior**
- How long does Call() block waiting for response?
- What timeout should we use?
- What errors indicate gopls is dead vs just slow?
- Can we introspect the connection state during a call?

## Expected Output

Please provide:
1. **Root Cause Analysis**: Most likely cause of the hang/crash
2. **Pipe Management Issues**: Specific problems with current approach
3. **Deadlock Risk Assessment**: Is the deadlock hypothesis plausible?
4. **Recommended Fixes**: Specific code changes with reasoning
5. **Implementation Priority**: What to fix first
6. **Testing Strategy**: How to validate fixes work

## Code Context

The full file is at `/Users/jack/mag/dingo/pkg/lsp/gopls_client.go` (297 lines).

Key components:
- **Lines 49-112**: `start()` function with pipe setup and goroutine management
- **Lines 114-128**: `logStderr()` background function
- **Lines 131-149**: `Initialize()` RPC call (where hang occurs)
- **Lines 274-295**: `readWriteCloser` pipe wrapper
- **Lines 96-109**: Process exit monitoring goroutine

## Constraints

- Must use subprocess model (can't link gopls directly)
- Must be compatible with go.lsp.dev/jsonrpc2
- LSP proxy pattern must maintain (wrap gopls, translate requests)
- Should not require changes to pipe size limits (keep portable)
- Error recovery should be graceful (crash detection + restart)

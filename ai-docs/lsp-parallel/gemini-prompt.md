# LSP Stream Destruction Bug Investigation

## Problem Statement
The dingo-lsp server crashes with "stream destroyed" error during IDE operation. Need to trace the exact sequence of events that causes stream closure and identify the root cause.

## Current Architecture

### Connection Setup (cmd/dingo-lsp/main.go)
```
1. main() creates stdio wrapper (stdinoutCloser) around os.Stdin/os.Stdout
2. Creates jsonrpc2.Stream from stdinoutCloser
3. Creates jsonrpc2.Conn from stream
4. Calls conn.Go(ctx, handler) to start reading/processing
5. Stores conn in server via SetConn(conn, ctx)
6. Waits on <-conn.Done() for completion
```

**Critical Issue**: The stdinoutCloser.Close() returns nil without actually closing stdin/stdout (lines 84-87), which may mask real closure issues.

### GoplsClient Connection (pkg/lsp/gopls_client.go)
```
1. Creates separate subprocess: gopls -mode=stdio
2. Creates pipes: stdin, stdout, stderr
3. Wraps pipes in readWriteCloser (stdin=WriteCloser, stdout=ReadCloser)
4. Creates jsonrpc2.Stream from readWriteCloser
5. Creates jsonrpc2.Conn from stream
6. Calls conn.Go(ctx, handler) with background context (line 91)
7. Spawns goroutine to monitor process exit (lines 96-109)
```

**Critical Issue**: Uses background context for conn.Go(), not server's context (line 85). This means gopls connection handler keeps running even when server is shutting down.

### Handler Flow
1. Server.handleRequest() receives request from IDE
2. Extracts request type (initialize, completion, definition, etc.)
3. Forwards to gopls via GoplsClient methods (Completion, Definition, etc.)
4. GoplsClient calls conn.Call() or conn.Notify() to gopls
5. Response translated and sent back to IDE

## Failure Scenarios

### Scenario 1: IDE Closes Connection
```
Timeline:
T1: IDE closes connection (sends "exit")
T2: readWriteCloser on IDE side receives EOF
T3: jsonrpc2 stream detects closed reader
T4: conn.Go() handler exits and closes the connection
T5: "stream destroyed" error
```

**Evidence**: Line 55 in main.go waits on `<-conn.Done()`, meaning connection closure is expected but error indicates ungraceful close.

### Scenario 2: gopls Process Crashes
```
Timeline:
T1: gopls subprocess crashes
T2: Process.Wait() returns error (line 97 in gopls_client.go)
T3: handleCrash() called (line 105), attempts restart
T4: But original conn.Go() is still reading from destroyed gopls pipes
T5: Pipe reader gets EOF or closed pipe error
T6: "stream destroyed" error
```

**Evidence**: Lines 96-109 fork a monitor goroutine but don't clean up existing connection before restart.

### Scenario 3: Race Condition on Shutdown
```
Timeline:
T1: IDE sends shutdown request
T2: Server.handleShutdown() closes file watcher and calls gopls.Shutdown()
T3: gopls.Shutdown() sets shuttingDown flag and closes connection (lines 220-256)
T4: But Server.handleRequest() still processing requests after shutdown
T5: Tries to forward request to already-closed gopls.conn
T6: "stream destroyed" when writing to closed pipe
```

**Evidence**: No synchronization between shutdown and in-flight requests (lines 188-206 in server.go).

### Scenario 4: Concurrent Requests During Shutdown
```
Timeline:
T1: Multiple requests in-flight to gopls
T2: IDE sends shutdown while requests are pending
T3: handleShutdown() closes gopls connection (line 244)
T4: In-flight request tries to call gopls.conn.Call()
T5: Call writes to closed pipe
T6: "stream destroyed" error
```

**Evidence**: No request queue or "draining" mechanism before shutdown (no WaitGroup or channel).

## Code Issues Found

### Issue 1: GoplsClient.conn handler uses background context
**File**: pkg/lsp/gopls_client.go, line 85
```go
ctx := context.Background()
c.conn.Go(ctx, handler)
```
**Problem**: Handler keeps running with background context even when server context is cancelled. Should use a cancellable context tied to gopls lifecycle.

### Issue 2: No synchronization during gopls restart
**File**: pkg/lsp/gopls_client.go, lines 96-109
```go
go func() {
    err := c.cmd.Wait()
    // ... just calls handleCrash()
}()
```
**Problem**: Old connection handler still running when handleCrash() creates new connection. Two conn.Go() handlers fighting for same pipes.

### Issue 3: stdinoutCloser doesn't actually close
**File**: cmd/dingo-lsp/main.go, lines 84-87
```go
func (s *stdinoutCloser) Close() error {
    return nil  // Doesn't close stdin/stdout!
}
```
**Problem**: Mask real closure issues. When stream is destroyed, we can't tell if it was intentional or a crash.

### Issue 4: No request draining during shutdown
**File**: pkg/lsp/server.go, lines 188-206
```go
func (s *Server) handleShutdown(...) {
    // ... closes gopls immediately without waiting for in-flight requests
}
```
**Problem**: In-flight requests may be processing when gopls connection is closed.

### Issue 5: No error context for "stream destroyed"
**File**: Unknown (likely in go.lsp.dev/jsonrpc2)
**Problem**: The error message "stream destroyed" provides no context:
- Which connection destroyed? (IDE or gopls)
- What operation was happening? (read, write, call)
- Why was it destroyed? (intentional close, pipe error, crash)

## Recommended Investigation Approach

1. **Add comprehensive logging** to identify which connection is destroyed:
   - Log all jsonrpc2.Conn.Go() calls with unique ID
   - Log all Conn.Close() calls with stack trace
   - Log all stream read/write errors with context

2. **Add shutdown coordination**:
   - Use sync.WaitGroup to track in-flight requests
   - Drain request queue before closing gopls connection
   - Cancel gopls handler context on shutdown

3. **Fix gopls handler context**:
   - Create cancellable context for gopls connection
   - Cancel it when server shuts down or gopls crashes
   - Ensure old handler exits before restart

4. **Fix stdinoutCloser**:
   - Actually close stdin/stdout on Close()
   - Or document why it shouldn't close them

5. **Add structured error reporting**:
   - Wrap stream errors with context (which connection, what operation)
   - Include connection lifecycle state (active, shutting down, closed)
   - Log all error paths with request IDs for tracing

## Questions for Investigation

1. **When exactly does stream get destroyed?**
   - During normal IDE shutdown (expected)?
   - During IDE operation (crash)?
   - During gopls restart (crash)?
   - After timeout or idle period?

2. **Which connection is destroyed?**
   - IDE → dingo-lsp connection?
   - dingo-lsp → gopls connection?
   - Both?

3. **What's the actual error context?**
   - Is it a pipe read error (EOF)?
   - Is it a pipe write error (broken pipe)?
   - Is it a context cancellation?
   - Is it jsonrpc2 explicit Close()?

4. **Is it race condition or deadlock?**
   - Does it happen consistently or intermittently?
   - Is it related to request timing?
   - Is it related to gopls latency?

## Output Expected

1. **Failure sequence diagram**: Visual timeline of what happens
2. **Root cause analysis**: Which component closes stream first and why
3. **Cascade failure analysis**: What happens after initial closure
4. **Prevention strategies**: How to detect and prevent each scenario

## Code References

**Main entry point**: /Users/jack/mag/dingo/cmd/dingo-lsp/main.go (lines 13-57)
**Server implementation**: /Users/jack/mag/dingo/pkg/lsp/server.go (lines 1-336)
**GoplsClient implementation**: /Users/jack/mag/dingo/pkg/lsp/gopls_client.go (lines 1-296)
**Handler implementations**: /Users/jack/mag/dingo/pkg/lsp/handlers.go (lines 1-100+)

## Constraints

- Must not break existing functionality
- Must maintain compatibility with gopls
- Must handle both normal and crash scenarios
- Must provide clear logging for debugging

# LSP Stream Destruction Bug - Synthesis of Findings

## Executive Summary

The "stream destroyed" error occurs due to a **cascade of failures** in connection lifecycle management:

1. **Server-side**: dingo-lsp crashes or exits during initialization
2. **Client-side**: VS Code client continues trying to use destroyed stream
3. **Cascading failure**: Subsequent operations fail with "Cannot call write after stream was destroyed"

Root cause: Missing synchronization between multiple goroutines managing two separate JSON-RPC connections (IDE ↔ dingo-lsp and dingo-lsp ↔ gopls).

## Three-Connection Architecture Problem

```
IDE (VS Code)
    ↓ jsonrpc2.Conn #1 (stdio)
dingo-lsp Server
    ↓ jsonrpc2.Conn #2 (gopls subprocess)
gopls
```

**Critical Issue**: No coordination between these connections. When one fails, others keep running.

## Root Causes (Verified from Code)

### 1. gopls Handler Uses Background Context (CRITICAL)
**File**: `pkg/lsp/gopls_client.go:85`
```go
ctx := context.Background()
c.conn.Go(ctx, handler)  // Wrong! Should use cancellable context
```

**Impact**:
- gopls connection handler keeps running even when server is shutting down
- If gopls crashes, handler keeps trying to read from destroyed pipe
- Two handlers competing for same connection during restart

**Cascade**:
1. gopls process crashes
2. handleCrash() tries to restart gopls and create new connection
3. Old connection handler still running, conflicts with new connection
4. Stream reads fail, returns error
5. Error handler in IDE client sees closed stream
6. "stream destroyed" error

### 2. No Request Draining on Shutdown (CRITICAL)
**File**: `pkg/lsp/server.go:188-206`
```go
func (s *Server) handleShutdown(...) {
    // ... closes gopls immediately
}
```

**Impact**:
- In-flight requests may be processing when gopls closes
- Request handler tries to call gopls after connection is closed
- Write to closed pipe fails

**Cascade**:
1. IDE sends shutdown while requests pending
2. handleShutdown() closes gopls connection
3. In-flight request tries: `s.gopls.Completion(ctx, params)`
4. GoplsClient.Completion tries: `c.conn.Call(ctx, ...)`
5. conn.Call() writes to closed pipe → error
6. Error propagates to IDE → "stream destroyed"

### 3. stdinoutCloser Doesn't Close (WRONG)
**File**: `cmd/dingo-lsp/main.go:84-87`
```go
func (s *stdinoutCloser) Close() error {
    return nil  // Doesn't actually close stdin/stdout!
}
```

**Impact**:
- When IDE closes connection, stdinoutCloser.Close() returns nil
- jsonrpc2 thinks stream is closed but actual file descriptors stay open
- Masks real closure issues
- Prevents proper error detection

### 4. Race Condition on gopls Restart (CRITICAL)
**File**: `pkg/lsp/gopls_client.go:96-109`
```go
go func() {
    err := c.cmd.Wait()  // Wait for process exit
    if err != nil && !shutdown {
        c.handleCrash()   // Try to restart
    }
}()
```

**Problem**:
- When gopls crashes, this monitor goroutine detects it
- Calls handleCrash() which creates NEW connection
- BUT: Old connection.Go() handler is still running with background context
- New connection.Go() handler starts
- Both handlers competing for same stdout pipe

**Cascade**:
1. gopls crashes (e.g., panic, segfault)
2. Monitor goroutine detects exit
3. handleCrash() creates new gopls and new connection
4. Old handler still reads from old (destroyed) gopls pipes
5. New handler tries to read from new gopls pipes
6. Whoever gets the pipe first, other fails
7. Error propagates to IDE → "stream destroyed"

### 5. No Synchronization Between IDE and gopls Connections
**File**: `pkg/lsp/server.go` (general issue)

**Problem**:
- When IDE connection closes (normal shutdown), gopls might still be running
- When gopls connection closes (crash), IDE connection keeps running
- No cleanup or notification flow between them

## Failure Sequence Diagram

```
Timeline: IDE closes VS Code while dingo-lsp is running

T0: IDE normal shutdown
T1: IDE sends "shutdown" request to dingo-lsp
T2: handleShutdown() called, closes gopls connection
    └─ gopls.Shutdown() closes conn (line 244)
    └─ Sends shutdown + exit to gopls
T3: In-flight request still processing in another goroutine:
    └─ handleCompletion() tries to call gopls
    └─ gopls.conn.Call() tries to write to closed pipe
    └─ Error: "broken pipe" or similar
T4: Error handler calls reply() with error
    └─ reply() tries to write error response back to IDE
    └─ But IDE already closed the stream
    └─ IDE side: "stream destroyed"
T5: conn.Done() returns on IDE connection
T6: main.go exits, server stops
```

## Why "stream destroyed" is Misleading

The error message comes from VS Code/jsonrpc2 when:
- It tries to send a message
- But the underlying pipe is closed
- Could be caused by:
  - Server crash (gopls or dingo-lsp)
  - Normal shutdown race condition
  - Protocol violation (invalid JSON-RPC)

Without detailed logging, we can't tell which.

## Verification from Go Code

**Evidence of background context issue**:
- Line 85 `gopls_client.go`: `ctx := context.Background()`
- No cancel function created
- No way to stop handler when server shuts down

**Evidence of restart race**:
- Line 96-109: Monitor goroutine spawned during client.start()
- Line 271: handleCrash() calls client.start() again
- Two concurrent start() calls can happen without mutex protection

**Evidence of no request draining**:
- handleShutdown() calls gopls.Shutdown() immediately
- No sync.WaitGroup to wait for in-flight requests
- No channel-based request queue
- No graceful drain mechanism

## Fix Priority

### MUST FIX (Prevents crashes)
1. **Create cancellable context for gopls handler**
   - Don't use background context
   - Cancel on server shutdown or gopls crash
   - Ensures handler exits cleanly

2. **Add request draining on shutdown**
   - Use sync.WaitGroup to track in-flight requests
   - Wait for all requests before closing gopls
   - Set timeout for drain (e.g., 5 seconds)

3. **Protect restart with mutex**
   - Prevent concurrent start() calls
   - Ensure old handler exits before new one starts
   - Sequence: cancel context → wait for handler exit → start new

4. **Fix stdinoutCloser.Close()**
   - Either actually close stdin/stdout
   - Or document why not (impacts error detection)

### SHOULD FIX (Improves debugging)
5. **Add comprehensive logging**
   - Log conn.Go() calls with unique ID
   - Log conn.Close() calls with reason
   - Log all stream errors with context

6. **Add error context wrapper**
   - Catch "stream destroyed" errors
   - Add info: which connection, which operation, what state
   - Log full context for debugging

### NICE TO HAVE (Robustness)
7. **Add connection state machine**
   - Track state: initializing → active → shutting_down → closed
   - Validate state transitions
   - Reject operations in wrong state

8. **Add healthcheck mechanism**
   - Periodically verify gopls is responsive
   - Detect hangs or slowness
   - Restart if needed

## Code References

**Problematic files:**
- `/Users/jack/mag/dingo/cmd/dingo-lsp/main.go` (lines 13-57, 84-87)
- `/Users/jack/mag/dingo/pkg/lsp/server.go` (lines 20-67, 188-206)
- `/Users/jack/mag/dingo/pkg/lsp/gopls_client.go` (lines 49-112, 220-272)

**Key issues:**
- gopls_client.go:85 - background context
- gopls_client.go:96-109 - restart race condition
- server.go:188-206 - no request draining
- main.go:84-87 - stdinoutCloser doesn't close

## Implementation Strategy

Fix in this order to maximize stability:

1. **Phase 1: Context management** (gopls_client.go)
   - Create parent context in NewGoplsClient()
   - Pass cancellable context to conn.Go()
   - Cancel on shutdown or crash before restart

2. **Phase 2: Request draining** (server.go)
   - Add WaitGroup to track requests
   - Drain in handleShutdown()
   - Set 5-second timeout

3. **Phase 3: Restart coordination** (gopls_client.go)
   - Use mutex to serialize restarts
   - Wait for old handler to exit
   - Then start new connection

4. **Phase 4: Logging** (all files)
   - Add structured logging for all stream operations
   - Include unique connection IDs
   - Log error context

Each phase should be tested independently before moving to next.

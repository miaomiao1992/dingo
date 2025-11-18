# Gopls Subprocess Pipe Management - Root Cause Analysis

**Date**: 2025-11-18
**Investigator**: Golang Architect (Direct Analysis)
**Status**: Critical Issues Identified

## Executive Summary

The gopls subprocess initialization hangs due to **buffering deadlock** in the stdio pipe handling combined with **improper jsonrpc2 stream initialization**. The issue involves synchronization between three concurrent goroutines (stderr logger, process monitor, and JSON-RPC handler) competing for limited pipe buffer capacity.

**Root Cause**: Unbuffered `readWriteCloser` + blocking stderr reader + synchronous pipe writes = deadlock when pipe buffers fill.

---

## 1. Root Cause Analysis: The Deadlock Scenario

### The Actual Problem

The initialization hangs because of a classic **producer-consumer deadlock**:

```
Timeline:
T0: Initialize() calls c.conn.Call() to send initialize request
T1: readWriteCloser.Write() sends request to stdin (unbuffered)
T2: jsonrpc2 waits on stdout.Read() for response
T3: gopls receives initialize on stdin, starts processing
T4: gopls writes logs to stderr (normal verbose output)
T5: stderr buffer fills (default 64KB)
T6: gopls blocks trying to write stderr (no consumer reading fast enough)
T7: gopls cannot read stdin for next request (thread is blocked on stderr write)
T8: Initialize request is stuck in gopls, no response
T9: dingo-lsp.conn.Call() waits forever
DEADLOCK!
```

### Why This Happens

**Current Architecture Problems**:

1. **Unbuffered I/O in readWriteCloser**
   - Line 285: `return rwc.stdin.Write(p)` - direct write, no buffering
   - Line 281: `return rwc.stdout.Read(p)` - direct read, no buffering
   - This works for small messages but fails when:
     - Initialize request is large (complex workspace params)
     - Gopls response is large (capabilities with many methods)
     - Gopls logs heavily during startup

2. **Stderr Reading Is Single-Threaded Bottleneck**
   - Line 77: `go c.logStderr(stderr)` - spawned only ONCE
   - Lines 114-128: `logStderr()` uses `bufio.Scanner` which is BLOCKING
   - Scanner.Scan() blocks until newline or max buffer (1MB)
   - If gopls outputs large single-line error: scanner blocks for 1MB data
   - While scanner blocks: stderr pipe buffer fills → gopls blocks → no stdout

3. **Conflicting Goroutine Synchronization**
   - Process monitor goroutine (line 96-109): waits on `cmd.Wait()`
   - Stderr logger goroutine (line 77): blocks on scanner
   - JSON-RPC handler goroutine (jsonrpc2): blocks on stdout read
   - None of these are coordinated - potential race on cleanup

4. **No Timeout on Initialize**
   - Line 142: `c.conn.Call(ctx, "initialize", params, &result)`
   - Uses background context (`context.Background()` on line 85)
   - Even if context has timeout, jsonrpc2 may not honor it during pipe operations

### Why "Destroyed Stream" Occurs

After the deadlock causes Initialize to hang:
1. Client timeout triggers
2. Connection.Close() is called
3. readWriteCloser.Close() closes stdin and stdout
4. Gopls tries to read from closed stdin or write to closed stdout
5. "destroyed stream" error - pipes are closed while communication in progress

---

## 2. Specific Issues in Current Code

### Issue A: Unbuffered Pipes (HIGH IMPACT)

**File**: `/Users/jack/mag/dingo/pkg/lsp/gopls_client.go`
**Lines**: 274-295 (readWriteCloser)

```go
func (rwc *readWriteCloser) Read(p []byte) (n int, err error) {
    return rwc.stdout.Read(p)  // ← Direct unbuffered read
}

func (rwc *readWriteCloser) Write(p []byte) (n int, err error) {
    return rwc.stdin.Write(p)  // ← Direct unbuffered write
}
```

**Problem**:
- No intermediate buffering between jsonrpc2 and pipes
- jsonrpc2 expects to read/write in arbitrary chunk sizes
- Pipes have fixed kernel buffer (64KB default on macOS)
- Large JSON-RPC messages can exceed pipe buffer

**Fix**: Add buffering layer with `bufio.Reader` and `bufio.Writer`

---

### Issue B: Synchronous Stderr Reading (MEDIUM IMPACT)

**File**: `/Users/jack/mag/dingo/pkg/lsp/gopls_client.go`
**Lines**: 114-128 (logStderr)

```go
func (c *GoplsClient) logStderr(stderr io.Reader) {
    scanner := bufio.NewScanner(stderr)
    scanner.Buffer(make([]byte, 4096), 1024*1024)  // ← Can block for 1MB

    for scanner.Scan() {
        line := scanner.Text()
        c.logger.Debugf("gopls stderr: %s", line)
    }
}
```

**Problem**:
- Single scanner blocked on reading
- Buffer size up to 1MB means one Scan() can block for 1MB of data
- While blocked, stderr pipe buffer fills, gopls blocks, deadlock cascades
- No backpressure handling

**Fix**:
- Use non-blocking read with timeout
- Or use separate buffered channel to decouple stderr reading

---

### Issue C: Context and Timeout Issues (MEDIUM IMPACT)

**File**: `/Users/jack/mag/dingo/pkg/lsp/gopls_client.go`
**Lines**: 85, 131-149 (Initialize)

```go
ctx := context.Background()  // ← No timeout!
handler := jsonrpc2.ReplyHandler(...)
c.conn.Go(ctx, handler)

// Later...
_, err := c.conn.Call(ctx, "initialize", params, &result)  // ← Background context
```

**Problem**:
- Background context has NO timeout
- jsonrpc2 blocks forever if gopls doesn't respond
- Even if Call has its own timeout, it may not interrupt pipe operations
- No way to detect gopls is hung vs just slow

**Fix**:
- Use context with timeout
- Or implement deadline tracking

---

### Issue D: Race Between Shutdown and Crash Recovery (MEDIUM IMPACT)

**File**: `/Users/jack/mag/dingo/pkg/lsp/gopls_client.go`
**Lines**: 96-109, 219-257

```go
// Crash recovery goroutine
go func() {
    err := c.cmd.Wait()
    c.closeMu.Lock()
    shutdown := c.shuttingDown
    c.closeMu.Unlock()

    if err != nil && !shutdown {
        c.handleCrash()  // Tries to restart
    }
}()

// Shutdown
func (c *GoplsClient) Shutdown(ctx context.Context) error {
    c.closeMu.Lock()
    c.shuttingDown = true
    c.closeMu.Unlock()

    // ... close connection ...

    if c.cmd != nil && c.cmd.Process != nil {
        if err := c.cmd.Wait(); err != nil {  // ← Will hang if already done!
            ...
        }
    }
}
```

**Problem**:
- `cmd.Wait()` can only be called once successfully
- If crash recovery goroutine calls `handleCrash()` → `start()` → new `Wait()` call
- Then Shutdown also tries to `Wait()` on same process
- `Wait()` will hang waiting for process that already exited

**Fix**:
- Track process state separately
- Don't call Wait() multiple times on same process
- Use once.Do() for cleanup

---

### Issue E: No Recovery from Pipe Closure (LOW IMPACT)

**File**: `/Users/jack/mag/dingo/pkg/lsp/gopls_client.go`
**Lines**: 288-294 (Close)

```go
func (rwc *readWriteCloser) Close() error {
    err1 := rwc.stdin.Close()
    err2 := rwc.stdout.Close()
    if err1 != nil {
        return err1
    }
    return err2
}
```

**Problem**:
- Once pipes are closed, no recovery possible
- Should set connection state so future calls fail gracefully
- Should trigger crash recovery if gopls is still trying to use pipes

**Fix**:
- Detect pipe closure in error handling
- Trigger graceful restart instead of hanging

---

## 3. Recommended Fixes (Priority Order)

### Fix 1: Add Buffering Layer (CRITICAL)

Replace unbuffered `readWriteCloser` with buffered version:

```go
type readWriteCloser struct {
    stdin  io.WriteCloser
    stdout io.ReadCloser
    reader *bufio.Reader
    writer *bufio.Writer
    mu     sync.Mutex
}

func (rwc *readWriteCloser) Read(p []byte) (n int, err error) {
    return rwc.reader.Read(p)  // ← Buffered read
}

func (rwc *readWriteCloser) Write(p []byte) (n int, err error) {
    n, err := rwc.writer.Write(p)
    if err != nil {
        return n, err
    }
    return n, rwc.writer.Flush()  // ← Ensure written
}

func (rwc *readWriteCloser) Close() error {
    rwc.mu.Lock()
    defer rwc.mu.Unlock()

    rwc.writer.Flush()
    err1 := rwc.stdin.Close()
    err2 := rwc.stdout.Close()
    if err1 != nil {
        return err1
    }
    return err2
}
```

**Impact**: Solves primary deadlock cause by providing sufficient buffer capacity.

---

### Fix 2: Improve Stderr Handling (HIGH)

Replace blocking scanner with non-blocking reader:

```go
func (c *GoplsClient) logStderr(stderr io.Reader) {
    // Option A: Use channel to decouple reading
    errChan := make(chan string, 100)

    // Reader goroutine (non-blocking writes to channel)
    go func() {
        scanner := bufio.NewScanner(stderr)
        for scanner.Scan() {
            select {
            case errChan <- scanner.Text():
            default:
                // Channel full, drop oldest message
                <-errChan
                errChan <- scanner.Text()
            }
        }
    }()

    // Logger goroutine (non-blocking reads from channel)
    go func() {
        for line := range errChan {
            c.logger.Debugf("gopls stderr: %s", line)
        }
    }()

    // Option B: Use context-based timeout
    // Option C: Use io.LimitedReader to prevent blocking on huge output
}
```

**Impact**: Prevents stderr from blocking main I/O path.

---

### Fix 3: Add Context Timeout (HIGH)

```go
func (c *GoplsClient) Initialize(ctx context.Context, params protocol.InitializeParams) (*protocol.InitializeResult, error) {
    // Add timeout if not present
    if _, deadline := ctx.Deadline(); deadline.IsZero() {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, 30*time.Second)  // ← Add 30s timeout
        defer cancel()
    }

    c.logger.Debugf("Calling gopls initialize (timeout: %v)", time.Until(ctx.Deadline()))
    var result protocol.InitializeResult
    _, err := c.conn.Call(ctx, "initialize", params, &result)
    if err != nil {
        c.logger.Errorf("gopls initialize call failed: %v", err)
        return nil, fmt.Errorf("gopls initialize failed: %w", err)
    }
    return &result, nil
}
```

**Impact**: Prevents infinite hangs, allows detection of stuck processes.

---

### Fix 4: Fix Shutdown/Restart Race (MEDIUM)

```go
type GoplsClient struct {
    // ... existing fields ...
    processOnce sync.Once
    processErr  error
}

func (c *GoplsClient) handleCrash() error {
    c.mu.Lock()
    defer c.mu.Unlock()

    if c.restarts >= c.maxRestarts {
        return fmt.Errorf("gopls crashed %d times, giving up", c.restarts)
    }

    // Don't call cmd.Wait() if process already exited
    c.restarts++
    c.cmd = nil  // ← Clear old process

    return c.start()
}

func (c *GoplsClient) Shutdown(ctx context.Context) error {
    c.closeMu.Lock()
    c.shuttingDown = true
    c.closeMu.Unlock()

    // ... connection close logic ...

    c.mu.Lock()
    cmd := c.cmd
    c.cmd = nil  // ← Prevent double-wait
    c.mu.Unlock()

    if cmd != nil && cmd.Process != nil {
        // Try graceful shutdown first
        if err := cmd.Wait(); err != nil {
            c.logger.Debugf("gopls process wait error: %v", err)
        }
        c.logger.Infof("gopls stopped (PID: %d)", cmd.Process.Pid)
    }

    return nil
}
```

**Impact**: Prevents race condition in process cleanup.

---

## 4. Testing Strategy

### Test 1: Verify Buffer Capacity
```go
func TestLargeInitializeMessage(t *testing.T) {
    // Create workspace with many files/folders
    // Call Initialize with large params (>64KB JSON)
    // Should not hang

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    _, err := client.Initialize(ctx, largeParams)
    if err != nil {
        t.Fatalf("Initialize hung: %v", err)
    }
}
```

### Test 2: Verify Stderr Non-Blocking
```go
func TestGoplsVerboseLogging(t *testing.T) {
    // Configure gopls to output verbose logs
    // Send requests while gopls is logging
    // Should not deadlock
}
```

### Test 3: Verify Timeout Detection
```go
func TestInitializeTimeout(t *testing.T) {
    // Pause gopls processing (breakpoint in test)
    // Send Initialize with short timeout
    // Should return timeout error, not hang
}
```

---

## 5. Alternative Approaches

### Alternative A: Use os.Pipe() with Explicit Buffering
Instead of exec.Cmd pipes, create custom buffered pipes:
- Pros: Full control over buffering
- Cons: More code, less standard

### Alternative B: Use Socket Instead of Stdio
Create local socket instead of pipes:
- Pros: Network-like interface, better debugging
- Cons: Changes subprocess launching model

### Alternative C: Use Bounded Queue Between Pipes and jsonrpc2
Instead of buffering pipes directly, buffer between wrapper and jsonrpc2:
- Pros: Isolates JSON-RPC from pipe complexity
- Cons: Adds another layer

**Recommendation**: Fix 1 + Fix 2 + Fix 3 (buffering + stderr handling + timeout) is sufficient and maintains compatibility.

---

## 6. Validation Checklist

Before deployment, verify:
- [ ] Initialize call completes within timeout
- [ ] Large workspace initialization works (>10K files)
- [ ] Stderr logging doesn't block I/O
- [ ] Crash recovery doesn't hang on shutdown
- [ ] Process restart works correctly after crash
- [ ] Pipe closure is detected and handled gracefully
- [ ] All requests use appropriate timeout context
- [ ] No race conditions on shutdown

---

## Key Findings Summary

1. **Primary Issue**: Unbuffered pipes + blocking stderr = deadlock when pipe buffers fill
2. **Secondary Issues**: No timeout, shutdown race, stderr bottleneck
3. **Solution Complexity**: Medium - need buffering layer + timeout + stderr decoupling
4. **Risk**: Medium - must test with realistic load and verbose logging
5. **Timeline**: 2-3 days for full fix and testing

---

## References

- Go exec.Cmd documentation: https://pkg.go.dev/os/exec
- JSON-RPC 2.0 spec: https://www.jsonrpc.org/specification
- go.lsp.dev/jsonrpc2: https://pkg.go.dev/go.lsp.dev/jsonrpc2
- Buffering in Go: https://pkg.go.dev/bufio

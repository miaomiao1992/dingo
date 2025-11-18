# LSP Server Context & Goroutine Threading Analysis

**Analysis Date**: 2025-11-18
**Status**: Critical threading issues identified
**Affected Components**: Server context, GoplsClient shutdown, FileWatcher callbacks

---

## Executive Summary

The Dingo LSP server has **THREE CRITICAL threading issues** that can cause race conditions, deadlocks, and goroutine leaks:

1. **Unprotected Server Context** (Line 30-31, 64-67, 327) - Race condition on s.ctx/s.ideConn
2. **GoplsClient Shutdown Deadlock** (Lines 96-109, 220-257) - Mutex ordering with crash recovery
3. **FileWatcher Callback Race** (Lines 177-179, 183-197) - Timer callbacks bypass synchronization

---

## Issue 1: Unprotected Server Context (CRITICAL)

**Status**: Real Problem
**Severity**: CRITICAL
**Lines**: 30-31, 64-67, 327, 269
**Race Condition**: YES
**Deadlock Risk**: NO

### Current Code Pattern

```go
// server.go line 30-31
type Server struct {
    ideConn jsonrpc2.Conn  // UNPROTECTED
    ctx     context.Context // UNPROTECTED
}

// line 64-67: SetConn (called from main after conn.Go)
func (s *Server) SetConn(conn jsonrpc2.Conn, ctx context.Context) {
    s.ideConn = conn   // WRITE without mutex
    s.ctx = ctx        // WRITE without mutex
}

// line 327: handleDingoFileChange (spawned from file watcher callback)
func (s *Server) handleDingoFileChange(dingoPath string) {
    s.transpiler.OnFileChange(s.ctx, dingoPath)  // READ without mutex
}

// line 269: handleDidSave (handler goroutine from JSON-RPC)
go s.transpiler.OnFileChange(ctx, dingoPath)  // Uses handler's context
```

### Race Condition Proof

**Goroutine A** (JSON-RPC handler, line 75):
```
1. handleRequest(ctx) called by JSON-RPC
2. handleDidSave(ctx) executes (line 257)
3. Line 269: go s.transpiler.OnFileChange(ctx, ...)  // Uses handler ctx
```

**Goroutine B** (File watcher callback, line 177):
```
1. afterFunc() expires
2. processPendingFiles() calls fw.onChange() callback
3. handleDingoFileChange() executes
4. Line 327: s.transpiler.OnFileChange(s.ctx, ...)  // Reads s.ctx
```

**Goroutine C** (Main thread, line 52):
```
1. SetConn(conn, ctx) called
2. Line 66: s.ctx = ctx  // CONCURRENT WRITE with reads from B and C
```

**Timing Attack**:
- Main thread calls SetConn while handlers are executing
- handleDingoFileChange reads s.ctx at exact moment SetConn writes it
- Result: Invalid context pointer or use-after-free

### Why It's Critical

1. **Undefined behavior**: Context operations on torn read
2. **Context lifecycle confusion**: Multiple contexts (background, handler, stored)
3. **Cascading failures**: Bad context causes transpiler timeout/cancellation
4. **LSP protocol violation**: Handlers must complete requests properly

### Recommended Fix

Add synchronization:

```go
type Server struct {
    config        ServerConfig
    gopls         *GoplsClient
    mapCache      *SourceMapCache
    translator    *Translator
    transpiler    *AutoTranspiler
    watcher       *FileWatcher
    workspacePath string
    initialized   bool
    ideConnMu     sync.RWMutex  // NEW
    ideConn       jsonrpc2.Conn
    ctx           context.Context
}

// Protected write
func (s *Server) SetConn(conn jsonrpc2.Conn, ctx context.Context) {
    s.ideConnMu.Lock()
    s.ideConn = conn
    s.ctx = ctx
    s.ideConnMu.Unlock()
}

// Protected read
func (s *Server) handleDingoFileChange(dingoPath string) {
    s.ideConnMu.RLock()
    ctx := s.ctx
    s.ideConnMu.RUnlock()
    s.transpiler.OnFileChange(ctx, dingoPath)
}
```

### Why This Matters

Without this fix, the LSP server can crash or hang mysteriously when file watcher fires while handlers are active. This is a **production blocker** for multi-user IDE scenarios.

---

## Issue 2: GoplsClient Shutdown Deadlock (CRITICAL)

**Status**: Real Problem
**Severity**: CRITICAL
**Lines**: 96-109, 220-257, 260-272
**Race Condition**: YES (shutdown flag)
**Deadlock Risk**: YES (mutex ordering)

### Current Code Pattern

```go
// gopls_client.go line 49-51: start() acquires mu
func (c *GoplsClient) start() error {
    c.mu.Lock()
    defer c.mu.Unlock()
    // ... setup code ...

    // Line 96: Spawns monitor goroutine WHILE HOLDING MU
    go func() {
        err := c.cmd.Wait()

        c.closeMu.Lock()  // Line 99: Separate lock
        shutdown := c.shuttingDown
        c.closeMu.Unlock()

        if err != nil && !shutdown {
            c.logger.Warnf("gopls process exited unexpectedly: %v", err)
            if crashErr := c.handleCrash(); crashErr != nil {  // LINE 105
                c.logger.Errorf("Failed to restart gopls: %v", crashErr)
            }
        }
    }()
}

// Line 220: Shutdown acquires DIFFERENT locks in DIFFERENT order
func (c *GoplsClient) Shutdown(ctx context.Context) error {
    c.closeMu.Lock()  // First lock
    c.shuttingDown = true
    c.closeMu.Unlock()

    c.mu.Lock()  // Second lock
    defer c.mu.Unlock()
    // ... shutdown code ...
}

// Line 260: handleCrash acquires mu
func (c *GoplsClient) handleCrash() error {
    c.mu.Lock()
    defer c.mu.Unlock()
    // ... restart code ...
}
```

### Deadlock Scenario

**Timeline:**

```
T1: Handler calls Shutdown() (line 220)
    → Acquires closeMu
    → Sets shuttingDown = true
    → Releases closeMu
    → Tries to acquire mu (WAITS)

T2: Monitor goroutine running (spawned line 96)
    → cmd.Wait() returns (gopls crashed)
    → Acquires closeMu
    → Reads shuttingDown = false (false is stale read, race on closeMu)
    → Releases closeMu
    → Calls handleCrash() (line 105)
    → Tries to acquire mu (WAITS for Shutdown)

T3: Shutdown now owns mu
    → But handleCrash is waiting for it (circular wait)

RESULT: DEADLOCK
```

### Why It's Critical

1. **Server hangs**: Shutdown blocks forever, preventing graceful exit
2. **Resource leak**: gopls subprocess never terminates
3. **IDE frozen**: Editor hangs waiting for LSP shutdown
4. **Invisible failure**: No error message, just hang

### Recommended Fix

**Option 1: Single mutex with consistent ordering**

```go
type GoplsClient struct {
    cmd          *exec.Cmd
    conn         jsonrpc2.Conn
    logger       Logger
    goplsPath    string
    restarts     int
    maxRestarts  int
    mu           sync.Mutex  // Single mutex
    shuttingDown bool
}

func (c *GoplsClient) start() error {
    c.mu.Lock()
    defer c.mu.Unlock()

    // ... setup code ...

    // Monitor goroutine with no locks (reads shuttingDown after release)
    go func() {
        err := c.cmd.Wait()

        c.mu.Lock()  // Single lock, consistent ordering
        shutdown := c.shuttingDown
        c.mu.Unlock()

        if err != nil && !shutdown {
            if crashErr := c.handleCrash(); crashErr != nil {
                c.logger.Errorf("Failed to restart gopls: %v", crashErr)
            }
        }
    }()
}

func (c *GoplsClient) Shutdown(ctx context.Context) error {
    c.mu.Lock()  // Always acquire mu first
    defer c.mu.Unlock()

    c.shuttingDown = true
    // ... rest of shutdown ...
}

func (c *GoplsClient) handleCrash() error {
    c.mu.Lock()
    defer c.mu.Unlock()

    if c.shuttingDown {  // Check before restart
        return nil
    }
    // ... restart code ...
}
```

**Option 2: Channel-based synchronization**

```go
type GoplsClient struct {
    cmd         *exec.Cmd
    conn        jsonrpc2.Conn
    logger      Logger
    goplsPath   string
    restarts    int
    maxRestarts int
    mu          sync.Mutex
    shutdown    chan struct{}  // Signals shutdown
    done        chan struct{}  // Waits for completion
}

func (c *GoplsClient) start() error {
    // Monitor goroutine
    go func() {
        select {
        case <-c.shutdown:
            return  // Graceful shutdown
        }

        err := c.cmd.Wait()
        if err != nil {
            c.handleCrash()
        }
    }()
}

func (c *GoplsClient) Shutdown(ctx context.Context) error {
    close(c.shutdown)  // Signal all goroutines

    c.mu.Lock()
    defer c.mu.Unlock()

    // ... shutdown code ...

    close(c.done)  // Signal completion
}
```

### Why This Matters

The deadlock happens when:
1. IDE calls Shutdown (to restart LSP)
2. gopls crashes at same moment
3. Server hangs, IDE hangs

This is **very likely in production** because shutdowns often happen when things are broken.

---

## Issue 3: FileWatcher Callback Race (HIGH)

**Status**: Real Problem
**Severity**: HIGH
**Lines**: 177-179, 183-197, 200-211
**Race Condition**: YES
**Deadlock Risk**: NO

### Current Code Pattern

```go
// watcher.go line 165-180: handleFileChange
func (fw *FileWatcher) handleFileChange(dingoPath string) {
    fw.mu.Lock()
    defer fw.mu.Unlock()

    fw.pendingFiles[dingoPath] = true

    // Line 177: Spawn callback AFTER releasing lock
    if fw.debounceTimer != nil {
        fw.debounceTimer.Stop()
    }

    fw.debounceTimer = time.AfterFunc(fw.debounceDur, func() {
        fw.processPendingFiles()  // CALLBACK runs in unknown goroutine
    })
}

// Line 183: processPendingFiles UNPROTECTED ITERATION
func (fw *FileWatcher) processPendingFiles() {
    fw.mu.Lock()
    files := make([]string, 0, len(fw.pendingFiles))
    for path := range fw.pendingFiles {  // Read while iterating
        files = append(files, path)
    }
    fw.pendingFiles = make(map[string]bool)
    fw.mu.Unlock()

    // Line 193-196: CALLBACK EXECUTES HERE WITHOUT MUTEX
    for _, path := range files {
        fw.logger.Debugf("Processing debounced file change: %s", path)
        fw.onChange(path)  // CRITICAL: Calls handleDingoFileChange
    }
}

// Line 200: Close with no callback protection
func (fw *FileWatcher) Close() error {
    fw.mu.Lock()
    defer fw.mu.Unlock()

    if fw.closed {
        return nil
    }

    fw.closed = true
    close(fw.done)
    return fw.watcher.Close()
    // NOTE: timer callback may still be pending!
}
```

### Race Condition Proof

**Scenario: Close during processPendingFiles**

```
T1: processPendingFiles() executes (line 183)
    → Acquires mu (line 184)
    → Builds files list
    → Releases mu (line 190)
    → About to call fw.onChange() (line 195)

T2: Close() called (line 200)
    → Acquires mu
    → Sets fw.closed = true
    → Closes watcher
    → Returns

T1: (continues from above)
    → Calls fw.onChange() (line 195)
    → Calls handleDingoFileChange()
    → Accesses server state (now inconsistent due to Close)

RESULT: Race condition on watcher state
```

### Why It's High Severity

1. **Potential crash**: Accessing closed watcher
2. **Cascading errors**: onChange calls handlers that may fail
3. **Lost events**: Pending transpilations may be missed
4. **Not easily reproducible**: Timing-dependent

### Recommended Fix

**Use atomic done check:**

```go
func (fw *FileWatcher) processPendingFiles() {
    fw.mu.Lock()

    // Check if already closed
    if fw.closed {
        fw.mu.Unlock()
        return
    }

    files := make([]string, 0, len(fw.pendingFiles))
    for path := range fw.pendingFiles {
        files = append(files, path)
    }
    fw.pendingFiles = make(map[string]bool)
    fw.mu.Unlock()

    // Process files (outside lock)
    for _, path := range files {
        fw.logger.Debugf("Processing debounced file change: %s", path)
        fw.onChange(path)
    }
}

func (fw *FileWatcher) Close() error {
    fw.mu.Lock()
    defer fw.mu.Unlock()

    if fw.closed {
        return nil
    }

    fw.closed = true

    // Stop pending timer
    if fw.debounceTimer != nil {
        fw.debounceTimer.Stop()
    }

    close(fw.done)
    return fw.watcher.Close()
}
```

---

## Issue 4: Context Lifecycle Confusion (MEDIUM)

**Status**: Real Problem
**Severity**: MEDIUM
**Lines**: 45, 269, 327, 34, 54
**Race Condition**: NO
**Deadlock Risk**: NO

### The Problem

Multiple contexts in flight simultaneously:

```go
// main.go line 45: Background context (never cancels)
ctx := context.Background()

// main.go line 49: Spawns handler with background context
conn.Go(ctx, handler)

// server.go line 269: Handler receives ITS OWN context
go s.transpiler.OnFileChange(ctx, dingoPath)  // Handler's context

// server.go line 327: Uses stored context (background)
s.transpiler.OnFileChange(s.ctx, dingoPath)  // Server's context

// transpiler.go line 34: Creates ANOTHER timeout context
ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
```

### Why It Matters

**Problem 1: Handler context may cancel unexpectedly**
- JSON-RPC handler's context has unknown lifetime
- If handler returns, context is cancelled
- But spawned transpiler goroutine still runs
- Causes transpilation to abort mid-process

**Problem 2: Inconsistent behavior**
- Line 269: Uses handler's context (short-lived)
- Line 327: Uses server's context (long-lived)
- Same operation, different lifetimes = unpredictable behavior

**Problem 3: Timeout confusion**
- transpiler.go line 34 wraps context with 30s timeout
- If outer context expires first, wrapper is ignored
- If outer context is background (never expires), 30s timeout is actual limit
- Unclear which timeout actually applies

### Recommended Fix

**Separate context concerns:**

```go
type Server struct {
    // ... existing fields ...
    mu            sync.RWMutex
    ideConn       jsonrpc2.Conn
    ctx           context.Context
    bgCtx         context.Context        // For background tasks
    bgCancel      context.CancelFunc
}

func NewServer(cfg ServerConfig) (*Server, error) {
    bgCtx, bgCancel := context.WithCancel(context.Background())

    server := &Server{
        // ... existing init ...
        bgCtx:    bgCtx,
        bgCancel: bgCancel,
    }
    return server, nil
}

// For request handlers - use request context
func (s *Server) handleDidSave(ctx context.Context, ...) error {
    if isDingoFile(...) {
        go s.transpiler.OnFileChange(ctx, dingoPath)  // Handler context
    }
}

// For background tasks - use background context
func (s *Server) handleDingoFileChange(dingoPath string) {
    s.mu.RLock()
    ctx := s.bgCtx
    s.mu.RUnlock()
    s.transpiler.OnFileChange(ctx, dingoPath)  // Background context
}

func (s *Server) handleShutdown(...) error {
    // Cancel all background tasks
    s.mu.Lock()
    s.bgCancel()
    s.mu.Unlock()
}
```

---

## Issue 5: Handler Synchronization (MEDIUM)

**Status**: Real Problem
**Severity**: MEDIUM
**Lines**: 75, 228, 249, 278, 299
**Race Condition**: Potential
**Deadlock Risk**: NO

### The Problem

Multiple handlers access shared state without synchronization:

```go
// All these handlers run concurrently (JSON-RPC multiplexes them)
func (s *Server) handleDidOpen(ctx, reply, req) error {
    if err := s.gopls.DidOpen(ctx, params); err != nil {  // Accesses s.gopls
        s.config.Logger.Warnf(...)
    }
    return reply(ctx, nil, nil)
}

func (s *Server) handleDidChange(ctx, reply, req) error {
    if err := s.gopls.DidChange(ctx, params); err != nil {  // Same s.gopls
        s.config.Logger.Warnf(...)
    }
    return reply(ctx, nil, nil)
}

func (s *Server) handleCompletion(ctx, reply, req) error {
    return s.handleCompletionWithTranslation(ctx, reply, req)  // Uses translator
}
```

### Why It Matters

1. **GoplsClient has internal mutex**: So concurrent calls are safe, but...
2. **Source map cache**: Likely not thread-safe (need to verify)
3. **Translator**: Unknown synchronization (need to verify)
4. **Logger**: Go's log package is thread-safe, but custom logger may not be

### Recommended Fix

Verify each shared component:

```go
// pkg/lsp/sourcemap_cache.go - ensure thread-safe
type SourceMapCache struct {
    mu     sync.RWMutex
    maps   map[string]*SourceMap
}

// pkg/lsp/translator.go - ensure thread-safe
type Translator struct {
    cache *SourceMapCache  // Guarded by cache's mutex
}
```

---

## Summary Table

| Issue | Type | Severity | Status | Fix Complexity |
|-------|------|----------|--------|-----------------|
| 1. Server Context Race | Race Condition | CRITICAL | Fixable | Low (add mutex) |
| 2. GoplsClient Deadlock | Deadlock | CRITICAL | Fixable | Medium (restructure) |
| 3. FileWatcher Callback | Race Condition | HIGH | Fixable | Low (add checks) |
| 4. Context Lifecycle | Design Issue | MEDIUM | Fixable | Medium (separate contexts) |
| 5. Handler Sync | Design Issue | MEDIUM | Needs Verification | Low-Medium |

---

## Immediate Action Items

### CRITICAL (Must fix before merging)
1. Add mutex to Server.ideConn and Server.ctx access
2. Refactor GoplsClient shutdown to use single mutex or channels

### HIGH (Should fix soon)
3. Add closed check in FileWatcher.processPendingFiles()

### MEDIUM (Refactor for clarity)
4. Separate background context from handler context in Server
5. Verify thread-safety of SourceMapCache and Translator

---

## Testing Recommendations

1. **Stress test with concurrent file changes**: Simulate editor with rapid saves
2. **Chaos test with gopls crashes**: Kill gopls subprocess during operations
3. **Race detector**: `go test -race ./pkg/lsp/...`
4. **Shutdown stress test**: Rapid initialize/shutdown cycles

```bash
# Run with race detector
go test -race ./pkg/lsp/ -v

# Profile under load
go test -cpuprofile=cpu.prof -memprofile=mem.prof ./pkg/lsp/

# Stress test
go test -count=100 ./pkg/lsp/ -timeout=5m
```

---

## References

- Go Concurrency Patterns: https://go.dev/blog/context
- Effective Go - Concurrency: https://golang.org/doc/effective_go#concurrency
- TOCTOU (Time of Check, Time of Use) race conditions

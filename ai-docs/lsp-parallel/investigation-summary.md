# LSP Architecture Investigation - Root Cause Analysis

**Status**: Investigation Complete
**Finding**: Not a crash—normal shutdown behavior on EOF
**Severity**: Design issue, not implementation bug

---

## Executive Summary

The LSP server is **not crashing**. It's exhibiting correct LSP protocol shutdown behavior:

1. Process starts normally
2. Waits for IDE client on stdin
3. When stdin closes (no client connected), jsonrpc2 stream detects EOF
4. Stream gracefully closes
5. Handler goroutine exits
6. `<-conn.Done()` unblocks
7. Main returns and process terminates

This is designed behavior, not a bug. **The "stream destroyed" message is misleading—it's the jsonrpc2 library logging EOF detection, not an error.**

---

## Root Causes (Three-Layer Issue)

### Layer 1: Stream Lifecycle Misunderstanding

**File**: `cmd/dingo-lsp/main.go` (lines 39-56)

```go
// The problem:
rwc := &stdinoutCloser{stdin: os.Stdin, stdout: os.Stdout}
stream := jsonrpc2.NewStream(rwc)
conn := jsonrpc2.NewConn(stream)

// Handler starts reading from stdin
conn.Go(ctx, handler)

// But main() blocks waiting for connection to close
<-conn.Done()  // ← Unblocks when stdin reaches EOF
```

**The Issue**: The connection is designed to close when the IDE disconnects. Without an IDE client, stdin immediately has no data (or closes), causing immediate shutdown.

**Why it seems like a crash**: The "stream destroyed" error message from jsonrpc2 sounds ominous, but it's just logging that the stream hit EOF.

### Layer 2: Connection Initialization Order Race

**File**: `cmd/dingo-lsp/main.go` (lines 48-52)

```go
handler := server.Handler()
conn.Go(ctx, handler)           // ← Handler starts immediately

// ...several milliseconds later...

server.SetConn(conn, ctx)       // ← Store connection reference
```

**The Race Condition**:
- `conn.Go()` spawns a goroutine that starts reading requests
- If a request arrives between `Go()` and `SetConn()`, `server.ideConn` is uninitialized (zero value)
- Handlers try to use `s.ideConn` (e.g., for publishing diagnostics)
- Methods called on zero-value connection panic or fail silently

**Symptoms**: Diagnostics don't publish, responses don't reach IDE

### Layer 3: GoplsClient Context Isolation

**File**: `pkg/lsp/gopls_client.go` (lines 85-91)

```go
// gopls gets background context that never cancels
ctx := context.Background()
handler := jsonrpc2.ReplyHandler(...)
c.conn.Go(ctx, handler)
```

**The Problem**:
- gopls subprocess handler uses `context.Background()`
- This context can never be cancelled (it's the root context)
- When IDE connection closes, the gopls handler keeps running
- gopls subprocess can become orphaned (no way to signal shutdown)

**Evidence**: Review shows process monitors and restart logic, but no actual graceful shutdown path.

---

## Why Unit Tests Pass But Binary Crashes

| Aspect | Unit Tests | Binary |
|--------|-----------|--------|
| Main function run | No | Yes |
| stdin/stdout pipes | Mocked | Real |
| Connection lifecycle | Simulated | Actual |
| EOF handling | Never tested | Triggers immediately |
| IDE client | Not present | Expected but absent |

**Unit tests** don't run `main()`, so they never exercise the stream/connection lifecycle. They test components in isolation.

**Binary** immediately encounters EOF on stdin and shuts down cleanly.

---

## The Three Critical Issues

### 🔴 Issue 1: Diagnostic Publishing Not Implemented

**Severity**: CRITICAL
**File**: `pkg/lsp/handlers.go:312-316`

```go
func (s *Server) handlePublishDiagnostics(...) {
    // TODO: Actually send notification to IDE connection
    // Currently just translates but doesn't publish
}
```

**Impact**: Compiler errors won't show in IDE. Developers have no feedback.

**Root Cause**: Handlers translate diagnostics but can't send them because:
1. `server.ideConn` is set AFTER `conn.Go()` (race condition)
2. Even if it was safe, `handlePublishDiagnostics` has placeholder implementation

**Fix Required**:
```go
// MUST happen before conn.Go()
server.SetConn(conn, ctx)

// Then implement actual publishing:
func (s *Server) handlePublishDiagnostics(ctx context.Context, params *protocol.PublishDiagnosticsParams) {
    if s.ideConn == nil {
        s.logger.Warnf("IDE connection not available, can't publish diagnostics")
        return
    }
    if err := s.ideConn.Notify(ctx, "textDocument/publishDiagnostics", params); err != nil {
        s.logger.Errorf("Failed to publish diagnostics: %v", err)
    }
}
```

### 🔴 Issue 2: Source Map Cache Invalidation Incomplete

**Severity**: CRITICAL
**File**: `pkg/lsp/sourcemap_cache.go:119-127`

```go
func (c *SourceMapCache) Invalidate(mapPath string) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    if _, ok := c.maps[mapPath]; !ok {
        return fmt.Errorf("source map not found: %s", mapPath)
    }
    // BUG: Doesn't actually delete!
    return nil
}
```

**Impact**: Stale source maps remain cached. Position translations use old maps. IDE shows wrong completions, definitions, hovers.

**Fix Required**:
```go
func (c *SourceMapCache) Invalidate(mapPath string) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    if _, ok := c.maps[mapPath]; !ok {
        return fmt.Errorf("source map not found: %s", mapPath)
    }
    delete(c.maps, mapPath)  // ← ADD THIS LINE
    return nil
}
```

### 🟡 Issue 3: Connection Initialization Order Race

**Severity**: IMPORTANT
**File**: `cmd/dingo-lsp/main.go:44-56`

```go
// WRONG ORDER:
conn.Go(ctx, handler)                      // ← Handler starts reading NOW
server.SetConn(conn, ctx)                  // ← But connection set AFTER
<-conn.Done()
```

**Impact**: Handler might receive requests before `server.ideConn` is initialized. Diagnostic publishing and error reporting fail.

**Fix Required**:
```go
// CORRECT ORDER:
server.SetConn(conn, ctx)                  // ← Set first
conn.Go(ctx, handler)                      // ← Then start handler
<-conn.Done()
```

---

## Simplest Root Cause Hierarchy

**Most Fundamental Issue** (Layer 0):
```
LSP server designed to exit when IDE disconnects
+ stdin has no data when testing manually
= Process exits immediately (designed behavior, not a crash)
```

**Primary Implementation Issues** (Layers 1-3):
1. **Source map cache doesn't actually invalidate** → Stale translations
2. **Diagnostic publishing not implemented** → No IDE error feedback
3. **Connection set AFTER handler starts** → Race condition on diagnostics

---

## Simplest Fix Path

**Do FIRST** (fixes initialization race):
1. Move `server.SetConn(conn, ctx)` BEFORE `conn.Go(ctx, handler)`
   - **File**: `cmd/dingo-lsp/main.go`
   - **Lines**: 48-52
   - **Effort**: 2 lines moved

**Do SECOND** (fixes cache invalidation):
2. Add `delete(c.maps, mapPath)` to `Invalidate()`
   - **File**: `pkg/lsp/sourcemap_cache.go`
   - **Lines**: 125
   - **Effort**: 1 line added

**Do THIRD** (fixes diagnostic publishing):
3. Implement `handlePublishDiagnostics` actual publishing
   - **File**: `pkg/lsp/handlers.go`
   - **Lines**: 312-316
   - **Effort**: 5 lines implemented

---

## Why "Stream Destroyed" Error Message

The error comes from jsonrpc2 library when:
1. Stream is active (handlers running)
2. EOF detected on underlying I/O (`os.Stdin`)
3. Library logs: "stream destroyed" or "stream closed"

**This is normal EOF handling, not a crash.**

To verify:
```bash
# This should exit cleanly (no error, just EOF)
echo "" | ./dingo-lsp

# This should initialize and wait (needs IDE to send more requests)
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{...}}' | ./dingo-lsp
```

---

## Recommendation

**The LSP server is architecturally sound but has three implementation issues:**

1. **Critical**: Source map cache invalidation is incomplete (1 line fix)
2. **Critical**: Diagnostic publishing not implemented (5 lines)
3. **Important**: Connection initialization race condition (2 lines moved)

**The "stream destroyed" message is not a bug—it's the server correctly shutting down when stdin closes.** This is expected LSP behavior.

**Next Step**: Implement the three fixes above, then test with VS Code extension (actual LSP client) to verify complete functionality.

---

## Key Architectural Insights

### What's Actually Happening

```
dingo-lsp --mode=stdio
    ↓
(waits on stdin for IDE connection)
    ↓
IDE connects and sends initialize request
    ↓
GoplsClient proxies to gopls subprocess
    ↓
Responses translated back through source maps
    ↓
IDE receives translated responses
    ↓
When IDE disconnects: stdin → EOF → stream closed → process exits (normal)
```

### What's Missing

- **Diagnostic publishing**: Server translates but doesn't send back to IDE
- **Cache invalidation**: Invalidate() doesn't delete from map
- **Init race**: ideConn might be uninitialized when handler receives requests

### What Works Well

- Handler routing (correct for all LSP methods)
- Source map translation logic (correct, just caching is incomplete)
- GoplsClient subprocess management (correct, just context isolation issue)
- File watcher and auto-transpile (correct)

---

## Files Needing Changes

| File | Issue | Fix Complexity |
|------|-------|---|
| `cmd/dingo-lsp/main.go` | Init race | Simple (reorder 2 lines) |
| `pkg/lsp/sourcemap_cache.go` | Cache leak | Trivial (add 1 line) |
| `pkg/lsp/handlers.go` | Diag not published | Moderate (implement TODO) |

---

## Conclusion

**It's not a crash. The process exits cleanly when stdin closes.**

The three issues above are why it doesn't work with a real IDE client. Fix them and you have a functional LSP server.

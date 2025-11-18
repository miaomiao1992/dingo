# LSP Server Context & Goroutine Threading Analysis

## Task
Analyze the Dingo LSP server's context and goroutine management for potential threading issues, deadlocks, and race conditions.

## Code Context

### 1. Main Entry Point (cmd/dingo-lsp/main.go)
- Line 45: Creates context.Background() for server
- Line 49: conn.Go(ctx, handler) spawns JSON-RPC handler
- Line 52: SetConn stores connection and context in server
- Line 55: Waits for <-conn.Done()

Key Issue: context.Background() is stored as s.ctx and reused in handlers and file watchers.

### 2. Server Structure (pkg/lsp/server.go)
- Line 30-31: Stores ideConn and ctx as instance fields
- Line 64-67: SetConn() sets both fields without synchronization
- Line 269: handleDidSave spawns goroutine with context.ctx: `go s.transpiler.OnFileChange(ctx, dingoPath)`
- Line 327: handleDingoFileChange calls transpiler with s.ctx (not passed context)

Critical Problems:
- No synchronization on ideConn/ctx field access
- Multiple goroutines access s.ctx without mutex
- Line 327 vs Line 269: Inconsistent context usage

### 3. GoplsClient (pkg/lsp/gopls_client.go)
- Line 24-26: Uses sync.Mutex (mu) for state, sync.Mutex (closeMu) for shutdown flag
- Line 85: Creates context.Background() in start()
- Line 96-109: Spawns monitor goroutine for crash recovery
- Line 222-224: shutdown flag protected by closeMu
- Line 226-227: Uses separate mu lock (DEADLOCK RISK!)

Critical Problems:
- Lines 96-109: Monitor goroutine checks shuttingDown after cmd.Wait()
- Lines 220-227: Shutdown uses closeMu first, then mu - order inconsistent with start()
- start() holds mu (line 50), but monitor goroutine spawned at line 96 doesn't release mu until line 111
- Potential deadlock if handleCrash() called while shutdown in progress

### 4. FileWatcher (pkg/lsp/watcher.go)
- Line 53: Spawns watchLoop() goroutine
- Line 166-180: handleFileChange uses fw.mu
- Line 177-179: AfterFunc spawns callback without mutex protection
- Line 183-197: processPendingFiles unlocks, then iterates - callback at line 195 is unprotected

Critical Problems:
- Line 177-179: Timer callback calls processPendingFiles() without holding mu
- Callback runs asynchronously, could race with Close()
- fw.closed not checked in processPendingFiles()

### 5. Handler Execution (pkg/lsp/server.go)
- Line 75: handleRequest receives context from jsonrpc2
- Each handler receives its own context from JSON-RPC layer
- Line 269: Spawns goroutine with this context OR s.ctx (inconsistent)

Questions:
1. When does the context passed to handleRequest expire?
2. If request context is short-lived, spawning goroutine with it could cause early cancellation
3. If we use s.ctx instead, it lives until shutdown but is unprotected

## Specific Concerns

### Context Cancellation Chain
- s.ctx = context.Background() (never cancels)
- Handler context = from JSON-RPC (unknown lifecycle)
- Transpiler context = either one or the other (INCONSISTENT!)
- Problem: If handler context cancels, spawned goroutine gets cancelled

### Race Conditions
1. **Server context access**: s.ctx read/written without mutex
2. **GoplsClient shutdown**: Two mutexes (mu, closeMu) with potential ordering issues
3. **FileWatcher timer callback**: Unprotected access to pendingFiles

### Goroutine Leaks
1. FileWatcher.watchLoop() - Line 53 spawned, closed only via fw.done
2. GoplsClient monitor - Line 96-109 always spawned, survives shutdown if not coordinated
3. Transpiler goroutines - Line 269 spawn with context that may cancel

### Deadlock Scenarios
1. GoplsClient.Shutdown() calls closeMu.Lock(), then mu.Lock() (line 226)
2. But start() calls mu.Lock() and spawns monitor goroutine (line 96)
3. If handleCrash called from monitor while shutdown in progress = potential deadlock

## Required Analysis

For each issue below, provide:
- Whether it's a real problem or false alarm
- Severity (Critical, High, Medium, Low)
- Proof of concept (code path that triggers it)
- Specific line numbers
- Recommended fix approach

### Q1: Context Synchronization
- Is s.ctx safe to access from multiple goroutines?
- What's the intended lifetime of this context?
- Should handlers use their own context or s.ctx?

### Q2: GoplsClient Shutdown Race
- Can the monitor goroutine (lines 96-109) race with Shutdown() (lines 220-257)?
- Is the mutex lock ordering safe?
- Can handleCrash be called while Shutdown is in progress?

### Q3: FileWatcher Callback Race
- Can processPendingFiles() be called after Close()?
- Is the timer callback properly cancelled on close?
- What protects the callback from concurrent Close()?

### Q4: Context Lifecycle
- When request handlers spawn goroutines (line 269), which context should be used?
- If handler context expires before transpiler finishes, what happens?
- Should there be a separate long-lived context for background tasks?

### Q5: Handler Synchronization
- Multiple handlers run concurrently (JSON-RPC level)
- They all access s.ideConn - is this safe?
- They all access s.gopls - is this synchronized?

## Expected Output

Provide analysis following this structure:
```
## Issue: [Title]
**Status**: Real Problem / False Alarm
**Severity**: Critical / High / Medium / Low
**Lines**: X-Y
**Race Condition**: [Yes/No - with race path if yes]
**Deadlock Risk**: [Yes/No - with scenario if yes]

**Proof of Concept**:
[Goroutine 1] ... causes [Race/Deadlock]
[Goroutine 2] ... at same time ...

**Recommended Fix**:
[Specific approach to fix]

**Why This Matters**:
[Impact on LSP stability]
```

Analyze all 5 questions above.

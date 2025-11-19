# Phase V LSP Implementation - Gemini 2.5 Flash Review

**Reviewer:** Google Gemini 2.5 Flash (via claudish proxy)
**Session:** 20251118-152749-phase5-lsp-foundation
**Date:** 2025-11-18
**Scope:** Language Server Protocol implementation (2,400 LOC)
**Status:** CHANGES_NEEDED

---

## Executive Summary

The Dingo LSP implementation demonstrates a strong foundation with promising initial performance benchmarks. The `gopls` proxy architecture with source map translation is a sound approach. However, a deeper dive reveals several **critical and important areas** that need immediate attention to ensure the LSP's robustness, scalability, and user experience for production use.

**Overall Assessment:** CHANGES_NEEDED

The implementation shows excellent performance in microbenchmarks (position translation: 3.4μs, cache hits: 63ns), but critical issues in **concurrency safety**, **security**, **scalability**, and **edge case handling** must be addressed before production deployment.

---

## ✅ Strengths

### 1. Solid Performance Foundation
Initial benchmarks for position translation (3.4μs) and cache hits (63ns) are **excellent**, showing that core operations are extremely fast. The estimated 70ms end-to-end autocomplete latency is well within the <100ms target.

### 2. Clear Architecture
The separation into distinct components (subprocess manager, position translator, cache, server, file watcher) is well-structured and follows good software engineering principles. The gopls proxy design avoids reimplementing Go language semantics.

### 3. Good Test Coverage in Core Components
Over 80% coverage in core components (translator, cache) indicates a well-tested foundation where critical logic resides. The double-check locking pattern in the cache is correctly implemented.

### 4. Focused Performance Goal
Clear target of <100ms autocomplete latency demonstrates a performance-oriented mindset. All current benchmarks exceed targets by 16-2000x.

---

## ⚠️ Concerns

### CRITICAL Issues (Must Fix Before Release)

#### 1. `jsonrpc2.Conn` Thread-Safety - CRITICAL RACE CONDITION

**Issue:** The `jsonrpc2.Conn` used for communication with gopls is shared across all concurrent LSP requests (handled in separate goroutines). The thread-safety of `go.lsp.dev/jsonrpc2.Conn` for concurrent `Call()` operations has **not been verified**.

**Impact:**
- High likelihood of race conditions under load (50+ concurrent autocomplete requests)
- Potential crashes, data corruption, or incorrect responses
- Undefined behavior if `jsonrpc2.Conn.Call()` is not thread-safe

**Evidence:**
```go
// pkg/lsp/server.go - Each request in separate goroutine
func (s *Server) handleRequest(ctx context.Context, req *jsonrpc2.Request) (interface{}, error) {
    // Multiple goroutines can call this concurrently
    switch req.Method {
    case "textDocument/completion":
        return s.handleCompletion(ctx, req)  // Calls gopls via shared conn
    }
}

// pkg/lsp/gopls_client.go - Single shared connection
type GoplsClient struct {
    conn jsonrpc2.Conn  // Shared state!
}

func (c *GoplsClient) Completion(ctx context.Context, params protocol.CompletionParams) (*protocol.CompletionList, error) {
    var result protocol.CompletionList
    // Multiple goroutines calling this simultaneously
    if err := c.conn.Call(ctx, "textDocument/completion", params, &result); err != nil {
        return nil, err
    }
    return &result, nil
}
```

**Recommendation:**
1. **Verify thread-safety:** Review `go.lsp.dev/jsonrpc2` source code or documentation
2. **If not thread-safe:** Add mutex around all `conn.Call()` operations
3. **If thread-safe:** Add explicit comment and stress test (100+ concurrent requests)
4. **Consider:** Request queue with worker pool to limit concurrency to gopls

**Priority:** P0 - This could cause production crashes

---

#### 2. Resource Exhaustion (DoS Risk) - CRITICAL SECURITY

**Issue:** No resource limits on gopls subprocess. Malicious or pathological source maps/requests could trigger excessive memory/CPU usage.

**Attack Scenarios:**
1. **Memory bomb:** Source map with 10 million mappings (1GB+ JSON)
2. **gopls explosion:** Large workspace causes gopls to consume 8GB RAM, crashing IDE
3. **Concurrent flood:** 1000 simultaneous LSP requests overwhelm system

**Current Code (No Limits):**
```go
// pkg/lsp/gopls_client.go
func (c *GoplsClient) start(goplsPath string) error {
    c.cmd = exec.Command(goplsPath, "-mode=stdio")
    // NO resource limits set!
    if err := c.cmd.Start(); err != nil {
        return err
    }
    return nil
}
```

**Impact:**
- gopls consumes all system memory → crashes host IDE
- User loses unsaved work, poor experience
- Could be exploited as denial-of-service attack

**Recommendation:**
1. **Set process limits:**
```go
cmd.SysProcAttr = &syscall.SysProcAttr{
    // Linux/macOS
    Setrlimit: &syscall.Rlimit{
        Cur: 2 * 1024 * 1024 * 1024, // 2GB memory limit
        Max: 4 * 1024 * 1024 * 1024, // 4GB hard limit
    },
}
```

2. **Monitor gopls memory:**
```go
go func() {
    ticker := time.NewTicker(30 * time.Second)
    for range ticker.C {
        if getProcessMemory(c.cmd.Process.Pid) > 3*GB {
            c.logger.Warnf("gopls using >3GB, restarting...")
            c.restart()
        }
    }
}()
```

3. **Validate source map size before loading:**
```go
info, _ := os.Stat(mapPath)
if info.Size() > 50*MB {
    return fmt.Errorf("source map too large: %d MB", info.Size()/MB)
}
```

4. **Request rate limiting:**
```go
// Limit to 50 concurrent gopls requests
sem := make(chan struct{}, 50)
```

**Priority:** P0 - Security vulnerability

---

#### 3. Path Traversal & Subprocess Injection - CRITICAL SECURITY

**Issue:** Insufficient validation of file paths and binary paths could allow reading arbitrary files or code execution.

**Vulnerabilities:**

**Path Traversal:**
```go
// User sends LSP request with URI: file://../../../../etc/passwd.dingo
// Is this validated?
```

**Subprocess Injection:**
```go
// What if user config sets: dingo.lsp.path = "malicious-script"?
// Current code executes without validation
cmd := exec.Command(goplsPath, "-mode=stdio")
```

**Impact:**
- Read/write arbitrary files on system
- Execute arbitrary code as user
- Critical security vulnerability

**Recommendation:**
1. **Validate all file paths:**
```go
func validateFilePath(path string, workspaceRoot string) error {
    absPath, err := filepath.Abs(path)
    if err != nil {
        return err
    }

    // Prevent directory traversal
    if !strings.HasPrefix(absPath, workspaceRoot) {
        return fmt.Errorf("path outside workspace: %s", path)
    }

    // Prevent symlink escapes
    resolvedPath, err := filepath.EvalSymlinks(absPath)
    if err != nil {
        return err
    }
    if !strings.HasPrefix(resolvedPath, workspaceRoot) {
        return fmt.Errorf("symlink points outside workspace: %s", path)
    }

    return nil
}
```

2. **Validate binary paths:**
```go
func validateBinaryPath(path string) error {
    // Only allow absolute paths or known directories
    if !filepath.IsAbs(path) {
        if _, err := exec.LookPath(path); err != nil {
            return fmt.Errorf("binary not in PATH: %s", path)
        }
        return nil
    }

    // Verify file exists and is executable
    info, err := os.Stat(path)
    if err != nil {
        return err
    }
    if info.IsDir() {
        return fmt.Errorf("path is directory: %s", path)
    }
    if info.Mode()&0111 == 0 {
        return fmt.Errorf("binary not executable: %s", path)
    }

    return nil
}
```

**Priority:** P0 - Security vulnerability

---

#### 4. Symlink Attacks - CRITICAL SECURITY

**Issue:** File watcher follows symlinks without restriction, could escape workspace or cause infinite loops.

**Attack Scenario:**
```bash
# User creates symlink in workspace
cd /workspace
ln -s / evil-symlink
# File watcher now recursively watches entire filesystem!
```

**Current Code (No Symlink Protection):**
```go
// pkg/lsp/watcher.go
func (fw *FileWatcher) watchRecursive(root string) error {
    return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
        // No symlink check!
        if info.IsDir() && !fw.shouldIgnore(path) {
            fw.watcher.Add(path)
        }
        return nil
    })
}
```

**Impact:**
- File system traversal outside workspace
- Infinite loop on circular symlinks
- Excessive memory/CPU usage

**Recommendation:**
```go
func (fw *FileWatcher) watchRecursive(root string) error {
    seen := make(map[string]bool)  // Track visited paths

    return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        // Resolve symlinks
        realPath, err := filepath.EvalSymlinks(path)
        if err != nil {
            fw.logger.Warnf("Cannot resolve symlink %s: %v", path, err)
            return filepath.SkipDir
        }

        // Check if outside workspace
        if !strings.HasPrefix(realPath, root) {
            fw.logger.Warnf("Symlink %s points outside workspace, skipping", path)
            return filepath.SkipDir
        }

        // Detect cycles
        if seen[realPath] {
            fw.logger.Warnf("Circular symlink detected at %s, skipping", path)
            return filepath.SkipDir
        }
        seen[realPath] = true

        if info.IsDir() && !fw.shouldIgnore(realPath) {
            fw.watcher.Add(realPath)
        }

        return nil
    })
}
```

**Priority:** P0 - Security vulnerability

---

#### 5. Missing LSP Input Validation - CRITICAL ROBUSTNESS

**Issue:** No validation for malformed URIs, positions, or other LSP request parameters.

**Attack Scenarios:**
```json
// Malformed position (negative line number)
{"line": -1, "character": 999999}

// Invalid URI
{"uri": "not-a-uri"}

// Null parameters where required
{"textDocument": null}
```

**Impact:**
- Crashes, panics, undefined behavior
- gopls receives invalid requests, crashes
- Poor user experience (confusing error messages)

**Recommendation:**
```go
func validateLSPPosition(pos protocol.Position) error {
    if pos.Line < 0 || pos.Line > maxLineNumber {
        return fmt.Errorf("invalid line number: %d", pos.Line)
    }
    if pos.Character < 0 || pos.Character > maxColumnNumber {
        return fmt.Errorf("invalid character position: %d", pos.Character)
    }
    return nil
}

func validateLSPURI(uri protocol.DocumentURI) error {
    if uri == "" {
        return fmt.Errorf("empty URI")
    }
    parsed, err := url.Parse(string(uri))
    if err != nil {
        return fmt.Errorf("invalid URI: %w", err)
    }
    if parsed.Scheme != "file" {
        return fmt.Errorf("unsupported URI scheme: %s", parsed.Scheme)
    }
    return nil
}

func validateCompletionParams(params protocol.CompletionParams) error {
    if err := validateLSPURI(params.TextDocument.URI); err != nil {
        return err
    }
    if err := validateLSPPosition(params.Position); err != nil {
        return err
    }
    return nil
}
```

**Priority:** P0 - Critical for robustness

---

#### 6. File Watcher Scalability - CRITICAL ARCHITECTURAL FLAW

**Issue:** Recursive workspace watching hits OS file descriptor limits in large projects.

**Scale Limits:**
- macOS: 256 file descriptors default (can be increased to ~10K)
- Linux: 1024 default (can be increased to 1M)
- Windows: Different limits

**Problem Scenario:**
```
Workspace with 10,000 .dingo files in 2,000 directories
→ File watcher tries to watch 2,000 directories
→ Exceeds OS limits on macOS (256 FDs)
→ fsnotify.Add() fails silently or errors
→ File changes not detected!
```

**Current Design:**
```go
// Watch EVERY directory in workspace
filepath.Walk(root, func(path string, info os.FileInfo) {
    if info.IsDir() {
        fw.watcher.Add(path)  // Can fail if too many!
    }
})
```

**Impact:**
- **CRITICAL:** Breaks in large projects (common scenario)
- Silent failures (user won't know files aren't watched)
- Fundamental scalability limitation

**Recommendation:**

**Option A: Watch Only Opened Files** (Preferred)
```go
// LSP protocol provides textDocument/didOpen notifications
// Only watch files that are currently opened in editor

func (s *Server) handleDidOpen(ctx context.Context, req *jsonrpc2.Request) {
    var params protocol.DidOpenTextDocumentParams
    json.Unmarshal(req.Params, &params)

    dingoPath := params.TextDocument.URI.Filename()
    s.watcher.WatchFile(dingoPath)  // Watch single file
}

func (s *Server) handleDidClose(ctx context.Context, req *jsonrpc2.Request) {
    var params protocol.DidCloseTextDocumentParams
    json.Unmarshal(req.Params, &params)

    dingoPath := params.TextDocument.URI.Filename()
    s.watcher.UnwatchFile(dingoPath)  // Stop watching
}
```

**Pros:**
- Scales to unlimited workspace size
- Only watches actively edited files (10-50 typically)
- No OS limits hit

**Cons:**
- Doesn't detect external changes to unopened files
- Requires manual "Refresh" command for external changes

**Option B: Sampling Strategy**
```go
// Watch only top N recently modified directories (N=100)
// Periodically re-scan workspace for changes
```

**Priority:** P0 - Breaks in real-world usage

---

#### 7. Position Translation Ambiguity - CRITICAL CORRECTNESS

**Issue:** Linear scan for reverse translation (Go → Dingo) cannot handle multiple source map entries with same generated line.

**Scenario (Nested Error Propagation):**
```dingo
// Dingo source
x := getData()?     // Line 10
y := process(x)?    // Line 11
```

**Source map:**
```json
[
  {
    "original_line": 10,
    "original_column": 5,
    "generated_line": 18,  // Both map to line 18-24!
    "generated_column": 1
  },
  {
    "original_line": 11,
    "original_column": 5,
    "generated_line": 18,  // AMBIGUOUS!
    "generated_column": 1
  }
]
```

**Current Code (First Match Wins):**
```go
// pkg/preprocessor/sourcemap.go (assumed)
func (sm *SourceMap) MapToOriginal(genLine, genCol int) (int, int) {
    for _, m := range sm.Mappings {
        if m.GeneratedLine == genLine {
            return m.OriginalLine, m.OriginalColumn  // Returns first match!
        }
    }
    return genLine, genCol  // Fallback
}
```

**Problem:**
- If gopls reports error on Go line 20, it could belong to EITHER Dingo line 10 OR 11
- Current code returns line 10 (first match)
- User sees error at wrong line!

**Impact:**
- **CRITICAL:** Diagnostics appear at incorrect positions
- Go-to-definition jumps to wrong location
- Confusing user experience

**Recommendation:**

**Use Column Information:**
```go
func (sm *SourceMap) MapToOriginal(genLine, genCol int) (int, int) {
    var bestMatch *Mapping

    for _, m := range sm.Mappings {
        if m.GeneratedLine == genLine {
            // Find mapping with closest column
            if bestMatch == nil || abs(m.GeneratedColumn - genCol) < abs(bestMatch.GeneratedColumn - genCol) {
                bestMatch = &m
            }
        }
    }

    if bestMatch != nil {
        return bestMatch.OriginalLine, bestMatch.OriginalColumn
    }

    return genLine, genCol  // Fallback
}
```

**Better: Interval Tree**
```go
// Store mappings as intervals: [startLine:startCol, endLine:endCol]
// Query: "Which interval contains Go position (20, 5)?"
// Returns: Exact original position with no ambiguity
```

**Priority:** P0 - Breaks core LSP functionality

---

### IMPORTANT Issues (Should Fix)

#### 8. File Watcher Debouncing Lag - UX DEGRADATION

**Issue:** 500ms debounce introduces noticeable delay after file save.

**User Experience:**
1. User types code in VSCode
2. Hits Ctrl+S (save)
3. **Waits 500ms** (debounce delay)
4. Transpiler runs (200ms)
5. Source map updated
6. **Finally** autocomplete reflects changes
7. **Total delay:** ~700ms perceived lag

**Impact:**
- Frustrating user experience
- Feels "slow" compared to pure Go projects
- Users may think LSP is broken

**Current Code:**
```go
// 500ms hardcoded delay
fw.debounceDur = 500 * time.Millisecond
```

**Recommendation:**

**Option A: Adaptive Debouncing**
```go
// First save: Transpile immediately (0ms debounce)
// Subsequent saves within 2s: Use 500ms debounce
// Idle for >5s: Reset to immediate mode

type AdaptiveDebouncer struct {
    lastSave     time.Time
    debouncing   bool
    immediateMode bool
}

func (d *AdaptiveDebouncer) ShouldDebounce() bool {
    now := time.Now()

    // If idle for >5s, use immediate mode
    if now.Sub(d.lastSave) > 5*time.Second {
        d.immediateMode = true
        return false
    }

    // If rapid saves, use debounce
    if d.debouncing {
        return true
    }

    return false
}
```

**Option B: Show Progress**
```go
// Transpile in background, show "Transpiling..." status
// User gets immediate feedback, autocomplete updates when ready
```

**Option C: User-Configurable**
```json
// VSCode settings
{
  "dingo.transpile.debounce": 200  // User can tune
}
```

**Priority:** P1 - Affects all users

---

#### 9. Stale Source Map Handling - CONFUSING UX

**Issue:** When transpilation fails, source map is stale but still used, causing incorrect LSP features.

**Scenario:**
1. User has valid `file.dingo` and `file.go.map`
2. User introduces syntax error, saves
3. Transpiler fails
4. Source map NOT updated (still has old mapping)
5. User requests autocomplete
6. LSP uses stale source map (positions off by N lines)
7. Autocomplete appears at wrong position

**Current Code (No Staleness Detection):**
```go
// pkg/lsp/sourcemap_cache.go
func (c *SourceMapCache) Get(goFilePath string) (*SourceMap, error) {
    // Loads from disk, no freshness check
    data, _ := os.ReadFile(mapPath)
    sm, _ := json.Unmarshal(data, &sm)
    return sm, nil
}
```

**Impact:**
- Confusing UX: "Why is autocomplete wrong?"
- User doesn't understand connection between transpile failure and LSP issues

**Recommendation:**

**Compare mtimes:**
```go
func (c *SourceMapCache) Get(goFilePath string) (*SourceMap, error) {
    mapPath := goFilePath + ".map"
    dingoPath := goToDingoPath(goFilePath)

    // Get file modification times
    dingoInfo, err := os.Stat(dingoPath)
    if err != nil {
        return nil, err
    }

    mapInfo, err := os.Stat(mapPath)
    if err != nil {
        return nil, fmt.Errorf("source map not found (transpile file first)")
    }

    // Check staleness
    if mapInfo.ModTime().Before(dingoInfo.ModTime()) {
        return nil, fmt.Errorf("source map is stale (file modified after last transpile)")
    }

    // Load and validate
    data, _ := os.ReadFile(mapPath)
    sm, _ := c.parseSourceMap(data)
    return sm, nil
}
```

**Better: Hash-Based Validation**
```go
// Source map includes hash of .dingo file
{
  "version": 1,
  "dingo_file_hash": "sha256:abc123...",
  "mappings": [...]
}

// LSP verifies hash matches current file
currentHash := sha256(readFile("file.dingo"))
if sm.DingoFileHash != currentHash {
    return fmt.Errorf("source map is stale (file changed)")
}
```

**Priority:** P1 - Major UX issue

---

#### 10. Version Compatibility - USER-HOSTILE FAILURE

**Issue:** "Hard failure" on unsupported source map version disrupts workflow abruptly.

**Current Code:**
```go
if sm.Version > MaxSupportedSourceMapVersion {
    return fmt.Errorf("unsupported source map version %d", sm.Version)
}
```

**Scenario:**
1. User updates transpiler (generates version 2 maps)
2. User forgets to update LSP
3. LSP hits version check, **hard fails**
4. All LSP features stop working
5. No graceful degradation

**Impact:**
- Frustrating UX: "Why did my IDE break?"
- User doesn't understand connection between transpiler/LSP versions

**Recommendation:**

**Graceful Degradation:**
```go
func (c *SourceMapCache) validateVersion(sm *SourceMap) error {
    if sm.Version == 0 {
        sm.Version = 1  // Legacy
    }

    if sm.Version > MaxSupportedSourceMapVersion {
        c.logger.Warnf(
            "Source map version %d is newer than supported version %d. " +
            "Some features may not work correctly. Please update dingo-lsp.",
            sm.Version, MaxSupportedSourceMapVersion,
        )

        // Show notification to user (once)
        if !c.shownVersionWarning {
            c.notifyUser("Source map version mismatch. Update dingo-lsp for full functionality.")
            c.shownVersionWarning = true
        }

        // Continue with degraded features (use base fields only)
        return nil
    }

    return nil
}
```

**Priority:** P1 - UX issue

---

#### 11. gopls Variability & Timeouts - UNREALISTIC ASSUMPTIONS

**Issue:** Current performance estimate assumes gopls responds in 50ms consistently, which is unrealistic.

**Real-World gopls Latency:**
- Small project: 20-50ms ✅
- Large project (1M LOC): 500ms-5s ❌
- Type checking complex generics: 1-10s ❌
- gopls indexing on startup: 30s-5min ❌

**Current Code (No Timeouts):**
```go
// pkg/lsp/gopls_client.go
func (c *GoplsClient) Completion(ctx context.Context, params protocol.CompletionParams) (*protocol.CompletionList, error) {
    var result protocol.CompletionList
    // What if gopls takes 60 seconds to respond?
    err := c.conn.Call(ctx, "textDocument/completion", params, &result)
    return &result, err
}
```

**Impact:**
- User triggers autocomplete, waits 30 seconds, no response
- IDE appears frozen
- Bad UX

**Recommendation:**

**Request Timeouts:**
```go
func (c *GoplsClient) Completion(ctx context.Context, params protocol.CompletionParams) (*protocol.CompletionList, error) {
    // 5-second timeout for autocomplete
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    var result protocol.CompletionList
    err := c.conn.Call(ctx, "textDocument/completion", params, &result)

    if err == context.DeadlineExceeded {
        c.logger.Warnf("gopls completion timed out after 5s")
        return &protocol.CompletionList{Items: []protocol.CompletionItem{}}, nil  // Empty result
    }

    return &result, err
}
```

**Cancellation Support:**
```go
// If user types another character, cancel previous autocomplete request
func (s *Server) handleCompletion(ctx context.Context, req *jsonrpc2.Request) {
    // Cancel previous request for same file
    s.cancelPreviousRequest(req.TextDocument.URI)

    // Create cancellable context
    ctx, cancel := context.WithCancel(ctx)
    s.trackRequest(req.TextDocument.URI, cancel)

    // Forward to gopls (can be cancelled)
    return s.translator.TranslateCompletion(ctx, req)
}
```

**Priority:** P1 - Critical for UX in real projects

---

#### 12. Low Test Coverage - LACK OF CONFIDENCE

**Issue:** Overall 39.1% test coverage and missing integration/concurrency/stress tests.

**Missing Test Scenarios:**
1. **Integration with real gopls:** All integration tests skipped (need gopls binary)
2. **Concurrency stress:** 100 simultaneous autocomplete requests
3. **Large source maps:** 10,000 mapping entries
4. **File watcher stress:** 1,000 files modified simultaneously
5. **Error recovery:** gopls crashes during active request
6. **Stale source maps:** File edited but not transpiled
7. **Resource exhaustion:** gopls consuming 4GB RAM
8. **Path validation:** Symlink attacks, directory traversal

**Current Tests:**
```go
// Only unit tests for core components
func TestPositionTranslation(t *testing.T) { ... }
func TestSourceMapCache(t *testing.T) { ... }

// Integration tests SKIPPED:
func TestLSP_WithRealGopls(t *testing.T) {
    t.Skip("Requires gopls binary")  // ❌ Not tested!
}
```

**Impact:**
- **HIGH RISK:** Production failures not caught
- Concurrency bugs only appear in production
- Edge cases not validated

**Recommendation:**

**Integration Tests (High Priority):**
```go
func TestLSP_EndToEnd_Autocomplete(t *testing.T) {
    // Start real gopls
    gopls := startRealGopls(t)
    defer gopls.Stop()

    // Start dingo-lsp
    server := startDingoLSP(t, gopls.Address())

    // Create .dingo file, transpile
    dingoFile := writeTestFile(t, "test.dingo", `
        package main
        func example() {
            x := getData()?
            x.  // <-- autocomplete here
        }
    `)
    transpile(t, dingoFile)

    // Send autocomplete request
    result := server.Completion(t, dingoFile, Position{Line: 4, Char: 14})

    // Verify result
    assert.Greater(t, len(result.Items), 0)
    assert.Equal(t, "Dingo position", result.Items[0].TextEdit.Range)
}
```

**Concurrency Stress Tests:**
```go
func TestLSP_ConcurrentRequests(t *testing.T) {
    server := startTestServer(t)

    // Launch 100 concurrent requests
    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            result := server.Completion(t, "test.dingo", randomPosition())
            assert.NotNil(t, result)
        }()
    }

    // Wait for all to complete (shouldn't crash!)
    wg.Wait()
}
```

**Priority:** P1 - Essential for production readiness

---

#### 13. Source Map Lookup Performance - O(n) BOTTLENECK

**Issue:** Linear scan O(n) for position translation will become bottleneck at scale.

**Current Code:**
```go
// O(n) linear scan
func (sm *SourceMap) MapToOriginal(genLine, genCol int) (int, int) {
    for _, m := range sm.Mappings {  // 200-500 iterations typical
        if m.GeneratedLine == genLine {
            return m.OriginalLine, m.OriginalColumn
        }
    }
    return genLine, genCol
}
```

**Performance at Scale:**
- Small file (50 mappings): ~2μs ✅
- Medium file (500 mappings): ~20μs ✅
- Large file (5,000 mappings): ~200μs ⚠️
- Very large file (50,000 mappings): ~2ms ❌

**Impact:**
- 2ms per position translation
- Autocomplete requires 2 translations (Dingo→Go, Go→Dingo)
- **4ms overhead** for very large files
- Degrades UX in large files

**Recommendation:**

**Option A: Binary Search O(log n)**
```go
// Pre-sort mappings by generated_line
sort.Slice(sm.Mappings, func(i, j int) bool {
    return sm.Mappings[i].GeneratedLine < sm.Mappings[j].GeneratedLine
})

func (sm *SourceMap) MapToOriginal(genLine, genCol int) (int, int) {
    // Binary search for line
    idx := sort.Search(len(sm.Mappings), func(i int) bool {
        return sm.Mappings[i].GeneratedLine >= genLine
    })

    if idx < len(sm.Mappings) && sm.Mappings[idx].GeneratedLine == genLine {
        // Linear scan within line (few entries per line)
        for i := idx; i < len(sm.Mappings) && sm.Mappings[i].GeneratedLine == genLine; i++ {
            m := sm.Mappings[i]
            if m.GeneratedColumn <= genCol && genCol < m.GeneratedColumn + m.Length {
                return m.OriginalLine, m.OriginalColumn
            }
        }
    }

    return genLine, genCol
}
// Performance: O(log n) + O(k) where k = entries per line (typically 1-3)
```

**Option B: Line Index Map O(1)**
```go
type SourceMap struct {
    Mappings     []Mapping
    lineIndex    map[int][]Mapping  // generated_line → []Mapping
}

func (sm *SourceMap) buildIndex() {
    sm.lineIndex = make(map[int][]Mapping)
    for _, m := range sm.Mappings {
        sm.lineIndex[m.GeneratedLine] = append(sm.lineIndex[m.GeneratedLine], m)
    }
}

func (sm *SourceMap) MapToOriginal(genLine, genCol int) (int, int) {
    // O(1) lookup
    mappings := sm.lineIndex[genLine]

    // O(k) scan (k = entries per line, typically 1-3)
    for _, m := range mappings {
        if m.GeneratedColumn <= genCol && genCol < m.GeneratedColumn + m.Length {
            return m.OriginalLine, m.OriginalColumn
        }
    }

    return genLine, genCol
}
// Performance: O(1) lookup + O(k) scan
```

**Option C: Interval Tree (Future)**
```go
// For complex multi-line expansions
// Store mappings as intervals: [startPos, endPos] → originalPos
// Query: "Which interval contains position X?"
// Handles overlapping ranges, nested expansions
```

**Priority:** P1 - Performance degradation at scale

---

#### 14. IPC Estimates Need Validation - UNREALISTIC ASSUMPTIONS

**Issue:** 5ms IPC estimates are optimistic, need empirical measurement.

**Current Estimate:**
```
VSCode → dingo-lsp: ~5ms (IPC)
dingo-lsp → gopls: ~5ms (IPC)
gopls → dingo-lsp: ~5ms
dingo-lsp → VSCode: ~5ms
Total IPC: ~20ms
```

**Reality Check:**
- stdio pipes: 100μs-1ms (very fast) ✅
- JSON-RPC serialization: 500μs-5ms (depends on message size) ⚠️
- Context switching: 10-100μs (low overhead) ✅
- **Total realistic:** 1-10ms per IPC hop

**5ms estimate might be:**
- **Optimistic** for large messages (1000 completion items)
- **Realistic** for small messages (hover, definition)

**Impact:**
- If IPC takes 20ms instead of 5ms, autocomplete is 90ms instead of 70ms
- Still within 100ms target, but less margin

**Recommendation:**

**Measure Real IPC Latency:**
```go
func BenchmarkLSP_EndToEndLatency(b *testing.B) {
    server := startTestServer(b)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        start := time.Now()

        // Send autocomplete request
        result := server.Completion(...)

        elapsed := time.Since(start)
        b.ReportMetric(float64(elapsed.Microseconds()), "μs/op")
    }
}
// Run in CI, track over time
```

**Log Real-World Latency:**
```go
// pkg/lsp/server.go
func (s *Server) handleCompletion(ctx context.Context, req *jsonrpc2.Request) {
    start := time.Now()
    defer func() {
        elapsed := time.Since(start)
        s.metrics.RecordLatency("completion", elapsed)

        if elapsed > 100*time.Millisecond {
            s.logger.Warnf("Slow autocomplete: %dms", elapsed.Milliseconds())
        }
    }()

    // ... handle request
}
```

**Priority:** P2 - Nice to have, not blocking

---

### MINOR Issues (Nice to Have)

#### 15. Cache Contention Metrics - OBSERVABILITY

**Issue:** No metrics on source map cache contention under load.

**Recommendation:**
```go
type SourceMapCache struct {
    // ... existing fields

    // Metrics
    hits      atomic.Int64
    misses    atomic.Int64
    evictions atomic.Int64
    contentionDuration atomic.Int64  // Time spent waiting for lock
}

func (c *SourceMapCache) Get(goFilePath string) (*SourceMap, error) {
    lockStart := time.Now()
    c.mu.RLock()
    c.contentionDuration.Add(int64(time.Since(lockStart)))

    // ... rest of implementation
}

func (c *SourceMapCache) GetMetrics() CacheMetrics {
    return CacheMetrics{
        Hits:      c.hits.Load(),
        Misses:    c.misses.Load(),
        Evictions: c.evictions.Load(),
        HitRate:   float64(c.hits.Load()) / float64(c.hits.Load() + c.misses.Load()),
        AvgContention: time.Duration(c.contentionDuration.Load() / c.hits.Load()),
    }
}
```

**Priority:** P3 - Nice to have for debugging

---

#### 16. Proactive Prerequisite Checks - BETTER ONBOARDING

**Issue:** VSCode extension doesn't proactively check for `dingo-lsp` and `gopls` binaries.

**Current:** Extension tries to start LSP, fails, user sees generic error.

**Recommendation:**
```javascript
// extension.js
async function activate(context) {
    // Check prerequisites before starting LSP
    const checks = [
        checkBinary('dingo-lsp', 'Dingo LSP server'),
        checkBinary('gopls', 'Go language server'),
        checkBinary('dingo', 'Dingo transpiler'),
    ];

    const results = await Promise.all(checks);

    const missing = results.filter(r => !r.found);
    if (missing.length > 0) {
        const message = `Missing prerequisites:\n${missing.map(m => m.name).join('\n')}`;
        const action = await vscode.window.showErrorMessage(
            message,
            'Install Instructions',
            'Ignore'
        );

        if (action === 'Install Instructions') {
            vscode.env.openExternal(vscode.Uri.parse('https://dingolang.com/docs/setup'));
        }

        return;  // Don't start LSP
    }

    // All prerequisites found, start LSP
    startLanguageClient(context);
}

async function checkBinary(name, displayName) {
    try {
        await exec(`which ${name}`);
        return { found: true, name: displayName };
    } catch {
        return { found: false, name: displayName };
    }
}
```

**Priority:** P3 - UX improvement

---

## 🔍 Questions

### 1. jsonrpc2.Conn Thread-Safety Verification
**Question:** Has the thread-safety of `go.lsp.dev/jsonrpc2.Conn` for concurrent `Call()` operations been explicitly verified (e.g., by source code review or targeted testing) or is it an assumption?

**Why it matters:** This is a P0 critical issue. If the conn is not thread-safe, the entire LSP will have race conditions under concurrent load.

**Recommendation:** Review the source code of `go.lsp.dev/jsonrpc2` or add explicit stress tests (100+ concurrent requests) to verify safety.

---

### 2. File Watcher Requirements
**Question:** Is a full workspace file watch absolutely necessary, or would watching only opened files (leveraging LSP client notifications) suffice for all current and planned features?

**Why it matters:** Full workspace watching has fundamental scalability issues (file descriptor limits). "Watch opened files only" scales infinitely and avoids all these issues.

**Trade-off:**
- **Full workspace:** Detects external changes (git pull, etc.)
- **Opened files only:** Simpler, scales better, requires manual refresh for external changes

**Recommendation:** Consider "opened files only" as the default, with optional "full workspace" mode for advanced users.

---

### 3. gopls Error Reporting Detail
**Question:** What level of detail does `gopls` provide when it encounters internal errors or fatal conditions, and how is this information propagated to `dingo-lsp` for user diagnostics?

**Why it matters:** If gopls crashes or has internal errors, users need clear error messages to understand what went wrong (vs. generic "LSP failed" message).

**Example:**
```
gopls internal error: "type checking failed: import cycle detected"
→ dingo-lsp receives this error
→ Should show to user: "Go type checking failed. Check for import cycles."
```

---

## 📊 Summary

### Overall Assessment: CHANGES_NEEDED

The LSP implementation has a **solid foundation** with excellent microbenchmark performance, but **critical issues in concurrency safety, security, scalability, and edge case handling** must be addressed before production deployment.

**Why CHANGES_NEEDED (not MAJOR_ISSUES):**
- Core architecture is sound (gopls proxy, source maps)
- Most issues have clear solutions (not architectural rewrites)
- Performance targets are achievable
- With fixes, this will be production-ready

---

### Top 5 Priority Issues

**P0 (Must Fix Before Release):**

1. **Verify/Fix jsonrpc2.Conn Thread-Safety**
   - Verify source code OR add mutex around all `conn.Call()`
   - Add concurrency stress tests (100+ simultaneous requests)
   - **Timeline:** 1-2 days

2. **Implement Security Measures**
   - Path validation (prevent traversal, symlink escapes)
   - Binary path validation (prevent code injection)
   - LSP input validation (prevent crashes)
   - **Timeline:** 2-3 days

3. **Implement Resource Limits on gopls**
   - Set memory/CPU limits via `SysProcAttr`
   - Monitor gopls resource usage, restart if exceeds
   - Validate source map size before loading
   - **Timeline:** 1-2 days

4. **Redesign File Watcher for Scalability**
   - Switch to "watch opened files only" strategy
   - OR implement sampling strategy with FD limit checks
   - **Timeline:** 2-3 days

5. **Fix Position Translation Ambiguity**
   - Use column information for disambiguation
   - OR implement interval tree for exact ranges
   - Add tests for nested error propagation
   - **Timeline:** 2-3 days

**Total Estimated Effort:** 8-13 days (2 weeks)

---

### Testability Score: LOW

**Why LOW:**
- Overall 39.1% coverage (target: >80%)
- Missing integration tests with real gopls
- Missing concurrency stress tests
- Missing large-scale tests (10K files, 10K mappings)
- Missing error recovery tests
- Missing security tests (path traversal, etc.)

**To Improve to MEDIUM:**
- Add integration tests with real gopls (requires CI setup)
- Add concurrency stress tests (100+ concurrent requests)
- Achieve 60%+ overall coverage

**To Improve to HIGH:**
- Achieve 80%+ overall coverage
- Add large-scale tests (1000+ files)
- Add error injection tests (gopls crashes, etc.)
- Add security vulnerability tests

---

### Scalability Score: LOW

**Why LOW:**
- File watcher hits OS limits in large projects (P0 blocker)
- O(n) position translation becomes bottleneck in large files
- Source map cache limited to 100 files (arbitrary)
- No load testing with large workspaces
- gopls resource limits not set (can consume unlimited memory)

**Known Limits:**
- **10,000 files:** File watcher fails (FD limits)
- **50,000 mappings:** Position translation takes 2ms (4ms round-trip)
- **100 open files:** Cache evictions every request (thrashing)

**To Improve to MEDIUM:**
- Fix file watcher scalability (watch opened files only)
- Optimize position translation (binary search or index map)
- Remove artificial cache limit (use memory-based eviction)

**To Improve to HIGH:**
- Implement interval tree for O(log n) position translation
- Add load testing with synthetic large projects
- Prove scalability to 100K+ files

---

### Alternative Approach Recommendation

**Prioritize "Token-Based Position Mapping" (Alternative 1) for Phase VI or Iteration 2**

**Why:**
- Current line/column mapping has **fundamental accuracy issues** with multi-line expansions
- Position translation ambiguity (Issue #7) is a symptom of architectural limitation
- Token-based mapping would solve:
  - Exact position tracking (no ambiguity)
  - Better diagnostics (point to exact token, not line)
  - Simpler reverse translation (token ID lookup, no scanning)

**How:**
```
Dingo:  x? (Token ID 42)
   ↓
Go:     if err != nil {  (Tokens 100-106, all reference Token 42)
            return err
        }

Diagnostic: gopls reports error on Token 103
Reverse map: Token 103 → Token 42 → Line 10, Col 5 (exact)
```

**Trade-off:**
- **Pros:** More accurate, simpler logic, future-proof
- **Cons:** Requires tokenizer, more complex transpiler changes

**Recommendation:** Keep current implementation for iteration 1, research token-based mapping for iteration 2.

---

## Final Recommendations

**Immediate Actions (Before Production Release):**
1. ✅ Fix all P0 issues (concurrency, security, resource limits, scalability)
2. ✅ Add integration tests with real gopls
3. ✅ Add concurrency stress tests
4. ✅ Optimize position translation (binary search minimum)
5. ✅ Implement stale source map detection

**Post-Release (Iteration 2):**
1. Research token-based position mapping
2. Add large-scale load testing
3. Implement advanced LSP features (rename, find references, etc.)
4. Publish VSCode extension to marketplace
5. Support other editors (Neovim, etc.)

**Success Metrics:**
- Zero crashes in 1-week beta testing with 10+ users
- Autocomplete latency <100ms in 95th percentile
- Scales to 10K+ file workspaces
- 80%+ test coverage
- Zero security vulnerabilities

---

**Review Complete**
**Reviewer:** Google Gemini 2.5 Flash
**Status:** CHANGES_NEEDED
**Critical Issues:** 7
**Important Issues:** 7
**Minor Issues:** 2
**Recommended Timeline:** 2 weeks for fixes + testing

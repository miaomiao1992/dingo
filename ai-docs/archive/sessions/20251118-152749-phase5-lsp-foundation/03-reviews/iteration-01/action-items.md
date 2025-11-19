# Action Items: Phase V LSP Foundation Fixes

**Session:** 20251118-152749-phase5-lsp-foundation
**Date:** 2025-11-18
**Priority:** CRITICAL and IMPORTANT issues only
**Estimated Total Time:** 30-40 hours (1-2 weeks)

---

## 🚨 CRITICAL Issues (Must Fix Before Testing)

### 1. Implement Diagnostic Publishing to IDE
**Priority:** P0
**Consensus:** 5/5 reviewers
**Estimated Time:** 2-3 hours

**Files to modify:**
- `pkg/lsp/server.go` - Add `ideConn` field to Server struct
- `pkg/lsp/handlers.go` - Implement publishing in `handlePublishDiagnostics()`

**Changes:**
```go
// server.go
type Server struct {
    config      ServerConfig
    gopls       *GoplsClient
    translator  *Translator
    transpiler  *AutoTranspiler
    watcher     *FileWatcher
    ideConn     jsonrpc2.Conn  // NEW: Store IDE connection
}

func (s *Server) Serve(ctx context.Context, conn jsonrpc2.Conn) error {
    s.ideConn = conn  // Store connection
    // ... rest
}

// handlers.go
func (s *Server) handlePublishDiagnostics(
    ctx context.Context,
    params protocol.PublishDiagnosticsParams,
) error {
    // Translate diagnostics from .go to .dingo positions
    translatedParams, err := s.translator.TranslateDiagnostics(params, GoToDingo)
    if err != nil {
        s.config.Logger.Warnf("Failed to translate diagnostics: %v", err)
        return nil
    }

    // Actually publish to IDE
    return s.ideConn.Notify(ctx, "textDocument/publishDiagnostics", translatedParams)
}
```

**Test:**
- Introduce syntax error in .dingo file
- Save file (triggers transpile)
- Verify red squiggly appears in VSCode at correct line

---

### 2. Wire Up gopls Crash Recovery
**Priority:** P0
**Consensus:** 5/5 reviewers
**Estimated Time:** 1-2 hours

**Files to modify:**
- `pkg/lsp/gopls_client.go` - Add process monitoring

**Changes:**
```go
type GoplsClient struct {
    cmd         *exec.Cmd
    conn        jsonrpc2.Conn
    logger      Logger
    restarts    int
    maxRestarts int
    shuttingDown bool         // NEW
    closeMu      sync.Mutex   // NEW
}

func (c *GoplsClient) start() error {
    // ... existing startup code ...

    // Monitor process exit
    go func() {
        err := c.cmd.Wait()

        c.closeMu.Lock()
        shutdown := c.shuttingDown
        c.closeMu.Unlock()

        if err != nil && !shutdown {
            c.logger.Warnf("gopls process exited unexpectedly: %v", err)
            if crashErr := c.handleCrash(); crashErr != nil {
                c.logger.Errorf("Failed to restart gopls: %v", crashErr)
            }
        }
    }()

    return nil
}

func (c *GoplsClient) Shutdown(ctx context.Context) error {
    c.closeMu.Lock()
    c.shuttingDown = true
    c.closeMu.Unlock()

    // ... rest of shutdown
}
```

**Test:**
- Start LSP server
- Kill gopls process manually (`kill -9 <gopls-pid>`)
- Verify LSP auto-restarts gopls
- Verify autocomplete still works after restart

---

### 3. Fix Source Map Cache Invalidation Bug
**Priority:** P0
**Found by:** Grok
**Estimated Time:** 30 minutes

**Files to modify:**
- `pkg/lsp/sourcemap_cache.go` - Fix `Invalidate()` method

**Changes:**
```go
func (c *SourceMapCache) Invalidate(goFilePath string) {
    c.mu.Lock()
    defer c.mu.Unlock()

    // FIX: Use goFilePath as key, not mapPath
    if _, ok := c.maps[goFilePath]; ok {
        delete(c.maps, goFilePath)
        c.logger.Debugf("Source map invalidated: %s", goFilePath)
    }
}
```

**Test:**
```go
func TestSourceMapCache_InvalidationRemovesEntry(t *testing.T) {
    cache, _ := NewSourceMapCache(testLogger)
    writeSourceMap(t, "test.go.map", validSourceMap)

    // Load into cache
    sm1, _ := cache.Get("test.go")
    assert.NotNil(t, sm1)

    // Invalidate
    cache.Invalidate("test.go")

    // Verify removed from internal map
    cache.mu.RLock()
    _, exists := cache.maps["test.go"]
    cache.mu.RUnlock()
    assert.False(t, exists)

    // Next Get should reload from disk
    sm2, _ := cache.Get("test.go")
    assert.NotSame(t, sm1, sm2)
}
```

---

### 4. Add Path Traversal Validation
**Priority:** P0 SECURITY
**Consensus:** 3/5 reviewers (Internal, Gemini, Codex)
**Estimated Time:** 2-3 hours

**Files to modify:**
- `pkg/lsp/sourcemap_cache.go` - Add validation in `Get()`
- `pkg/lsp/transpiler.go` - Add validation in `TranspileFile()`
- `pkg/lsp/server.go` - Store workspace root

**Changes:**
```go
// server.go
type Server struct {
    // ... existing fields
    workspaceRoot string  // NEW
}

func (s *Server) handleInitialize(...) (*protocol.InitializeResult, error) {
    // ... parse params ...
    s.workspaceRoot = params.RootPath  // Store workspace root
    // ...
}

// util.go (NEW FILE)
func validateFilePath(path string, workspaceRoot string) error {
    absPath, err := filepath.Abs(path)
    if err != nil {
        return fmt.Errorf("invalid path: %w", err)
    }

    // Prevent directory traversal
    if !strings.HasPrefix(absPath, workspaceRoot) {
        return fmt.Errorf("path outside workspace: %s", path)
    }

    // Prevent symlink escapes
    resolvedPath, err := filepath.EvalSymlinks(absPath)
    if err != nil {
        return fmt.Errorf("cannot resolve path: %w", err)
    }
    if !strings.HasPrefix(resolvedPath, workspaceRoot) {
        return fmt.Errorf("symlink points outside workspace: %s", path)
    }

    return nil
}

// sourcemap_cache.go
func (c *SourceMapCache) Get(goFilePath string) (*preprocessor.SourceMap, error) {
    // Validate path before loading
    if err := validateFilePath(goFilePath, c.workspaceRoot); err != nil {
        return nil, err
    }

    mapPath := goFilePath + ".map"
    // ... rest of method
}

// transpiler.go
func (at *AutoTranspiler) TranspileFile(ctx context.Context, dingoPath string) error {
    // Validate path
    if err := validateFilePath(dingoPath, at.workspaceRoot); err != nil {
        return err
    }

    // Validate .dingo extension
    if !strings.HasSuffix(dingoPath, ".dingo") {
        return fmt.Errorf("not a .dingo file: %s", dingoPath)
    }

    // ... rest of method
}
```

**Test:**
```go
func TestPathValidation_PreventTraversal(t *testing.T) {
    workspace := "/workspace"

    tests := []struct {
        path      string
        shouldErr bool
    }{
        {"/workspace/file.go", false},
        {"/workspace/subdir/file.go", false},
        {"../../etc/passwd", true},
        {"/etc/passwd", true},
    }

    for _, tt := range tests {
        err := validateFilePath(tt.path, workspace)
        if tt.shouldErr {
            assert.Error(t, err)
        } else {
            assert.NoError(t, err)
        }
    }
}
```

---

### 5. Fix Race Condition in Source Map Cache
**Priority:** P0
**Found by:** Grok, Gemini
**Estimated Time:** 1 hour

**Files to modify:**
- `pkg/lsp/sourcemap_cache.go` - Fix double-check locking

**Changes:**
```go
func (c *SourceMapCache) Get(goFilePath string) (*preprocessor.SourceMap, error) {
    mapPath := goFilePath + ".map"

    // Try read lock first (optimistic)
    c.mu.RLock()
    if sm, ok := c.maps[goFilePath]; ok {
        c.mu.RUnlock()
        c.logger.Debugf("Source map cache hit: %s", mapPath)
        return sm, nil
    }
    c.mu.RUnlock()

    // Cache miss, load with write lock
    c.mu.Lock()
    defer c.mu.Unlock()

    // CRITICAL: Re-check under write lock (safe - blocks all readers)
    if sm, ok := c.maps[goFilePath]; ok {
        return sm, nil
    }

    // Load source map from disk
    data, err := os.ReadFile(mapPath)
    if err != nil {
        return nil, fmt.Errorf("source map not found: %s (transpile .dingo file first)", mapPath)
    }

    sm, err := c.parseSourceMap(data)
    if err != nil {
        return nil, fmt.Errorf("invalid source map %s: %w", mapPath, err)
    }

    if err := c.validateVersion(sm); err != nil {
        return nil, fmt.Errorf("incompatible source map %s: %w", mapPath, err)
    }

    // Store with consistent key
    c.maps[goFilePath] = sm
    c.logger.Infof("Source map loaded: %s (version %d)", mapPath, sm.Version)

    return sm, nil
}
```

**Test:**
```go
func TestSourceMapCache_ConcurrentAccess(t *testing.T) {
    cache, _ := NewSourceMapCache(testLogger)
    writeSourceMap(t, "test.go.map", validSourceMap)

    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            if n%10 == 0 {
                cache.Invalidate("test.go")
            } else {
                cache.Get("test.go")
            }
        }(i)
    }
    wg.Wait()

    // No panics or race detector warnings = success
}
```

---

### 6. Fix URI Translation Bug (.dingo URIs Leak to gopls)
**Priority:** P0
**Found by:** Codex, Internal
**Estimated Time:** 1-2 hours

**Files to modify:**
- `pkg/lsp/translator.go` - Always return .go URI when translating Dingo→Go

**Changes:**
```go
func (t *Translator) translatePosition(
    uri protocol.DocumentURI,
    pos protocol.Position,
    dir Direction,
) (protocol.DocumentURI, protocol.Position, error) {
    line := int(pos.Line) + 1
    col := int(pos.Character) + 1

    var goPath string
    if dir == DingoToGo {
        goPath = dingoToGoPath(uri.Filename())
    } else {
        goPath = uri.Filename()
    }

    sm, err := t.cache.Get(goPath)
    if err != nil {
        // CRITICAL: Still translate URI even with 1:1 positions
        if dir == DingoToGo {
            // Must return .go URI, not .dingo
            goURI := protocol.URIFromPath(goPath)
            return goURI, pos, fmt.Errorf("source map not found: %s (file not transpiled)", goPath)
        }
        // For Go->Dingo without map, return error
        return uri, pos, fmt.Errorf("source map not found: %s", goPath)
    }

    // Translate position using source map
    var newLine, newCol int
    if dir == DingoToGo {
        newLine, newCol = sm.MapToGenerated(line, col)
    } else {
        newLine, newCol = sm.MapToOriginal(line, col)
    }

    newPos := protocol.Position{
        Line:      uint32(newLine - 1),
        Character: uint32(newCol - 1),
    }

    newURI := uri
    if dir == DingoToGo {
        newURI = protocol.URIFromPath(goPath)
    } else {
        newURI = protocol.URIFromPath(goToDingoPath(uri.Filename()))
    }

    return newURI, newPos, nil
}
```

**Test:**
```go
func TestTranslator_UntranspiledFile(t *testing.T) {
    translator := NewTranslator(emptyCache)
    dingoURI := protocol.URIFromPath("/workspace/test.dingo")
    pos := protocol.Position{Line: 10, Character: 5}

    // Translate Dingo→Go without source map
    goURI, goPos, err := translator.translatePosition(dingoURI, pos, DingoToGo)

    // Should return .go URI even on error
    assert.Error(t, err)
    assert.Equal(t, "/workspace/test.go", goURI.Filename())
    assert.Equal(t, pos, goPos)  // 1:1 mapping
}
```

---

### 7. Fix Position Translation Ambiguity
**Priority:** P0 CORRECTNESS
**Found by:** Gemini
**Estimated Time:** 2-3 hours

**Files to modify:**
- `pkg/preprocessor/sourcemap.go` - Use column information for disambiguation

**Changes:**
```go
func (sm *SourceMap) MapToOriginal(genLine, genCol int) (int, int) {
    var bestMatch *Mapping

    for i := range sm.Mappings {
        m := &sm.Mappings[i]
        if m.GeneratedLine == genLine {
            // Find mapping with closest column
            if bestMatch == nil {
                bestMatch = m
            } else {
                // Closer column match wins
                currDist := abs(m.GeneratedColumn - genCol)
                bestDist := abs(bestMatch.GeneratedColumn - genCol)
                if currDist < bestDist {
                    bestMatch = m
                }
            }
        }
    }

    if bestMatch != nil {
        // Calculate offset within mapping
        offset := genCol - bestMatch.GeneratedColumn
        return bestMatch.OriginalLine, bestMatch.OriginalColumn + offset
    }

    // No mapping found, return input position
    return genLine, genCol
}

func abs(x int) int {
    if x < 0 {
        return -x
    }
    return x
}
```

**Test:**
```go
func TestSourceMap_AmbiguousMapping(t *testing.T) {
    sm := &SourceMap{
        Mappings: []Mapping{
            {OriginalLine: 10, OriginalColumn: 5, GeneratedLine: 18, GeneratedColumn: 1, Length: 5},
            {OriginalLine: 11, OriginalColumn: 5, GeneratedLine: 18, GeneratedColumn: 10, Length: 5},
        },
    }

    // Position at column 2 should map to line 10 (closer to col 1)
    origLine, origCol := sm.MapToOriginal(18, 2)
    assert.Equal(t, 10, origLine)

    // Position at column 12 should map to line 11 (closer to col 10)
    origLine, origCol = sm.MapToOriginal(18, 12)
    assert.Equal(t, 11, origLine)
}
```

---

### 8. Fix JSON-RPC Handler Wiring
**Priority:** P0
**Found by:** Codex
**Estimated Time:** 2-3 hours

**Files to modify:**
- `pkg/lsp/gopls_client.go` - Add notification forwarding callback

**Changes:**
```go
type GoplsClient struct {
    cmd            *exec.Cmd
    conn           jsonrpc2.Conn
    logger         Logger
    restarts       int
    maxRestarts    int
    shuttingDown   bool
    closeMu        sync.Mutex
    onNotification func(method string, params interface{}) error  // NEW
}

func NewGoplsClient(
    goplsPath string,
    logger Logger,
    notifHandler func(string, interface{}) error,  // NEW
) (*GoplsClient, error) {
    client := &GoplsClient{
        logger:         logger,
        maxRestarts:    3,
        onNotification: notifHandler,  // Store callback
    }
    if err := client.start(goplsPath); err != nil {
        return nil, err
    }
    return client, nil
}

func (c *GoplsClient) start(goplsPath string) error {
    // ... existing subprocess setup ...

    stream := jsonrpc2.NewStream(stdin, stdout)

    // Create handler that intercepts notifications
    handler := jsonrpc2.AsyncHandler(jsonrpc2.HandlerWithError(
        func(ctx context.Context, req *jsonrpc2.Request) (interface{}, error) {
            // If notification (no ID), forward to editor
            if req.ID == nil && c.onNotification != nil {
                return nil, c.onNotification(req.Method, req.Params)
            }
            return nil, nil
        },
    ))

    c.conn = jsonrpc2.NewConn(stream, handler)

    // ... rest
}

// In server.go
func (s *Server) handleInitialize(...) (*protocol.InitializeResult, error) {
    // ... existing code ...

    // Forward gopls notifications to VSCode
    notifHandler := func(method string, params interface{}) error {
        return s.ideConn.Notify(ctx, method, params)
    }

    gopls, err := NewGoplsClient(s.config.GoplsPath, s.config.Logger, notifHandler)
    if err != nil {
        return nil, err
    }
    s.gopls = gopls

    // ... rest
}
```

**Test:**
- Introduce syntax error in .go file
- Verify gopls diagnostic reaches VSCode
- Verify progress notifications appear

---

### 9. Add LSP Input Validation
**Priority:** P0 ROBUSTNESS
**Found by:** Gemini
**Estimated Time:** 2-3 hours

**Files to modify:**
- `pkg/lsp/handlers.go` - Add validation to all handlers
- `pkg/lsp/validation.go` (NEW) - Validation functions

**Changes:**
```go
// validation.go (NEW FILE)
package lsp

const (
    maxLineNumber   = 1000000
    maxColumnNumber = 10000
)

func validateLSPPosition(pos protocol.Position) error {
    if pos.Line < 0 || int(pos.Line) > maxLineNumber {
        return fmt.Errorf("invalid line number: %d", pos.Line)
    }
    if pos.Character < 0 || int(pos.Character) > maxColumnNumber {
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
        return fmt.Errorf("invalid text document URI: %w", err)
    }
    if err := validateLSPPosition(params.Position); err != nil {
        return fmt.Errorf("invalid position: %w", err)
    }
    return nil
}

// handlers.go
func (s *Server) handleCompletion(ctx context.Context, req *jsonrpc2.Request) (*protocol.CompletionList, error) {
    var params protocol.CompletionParams
    if err := json.Unmarshal(req.Params, &params); err != nil {
        return nil, fmt.Errorf("invalid completion params: %w", err)
    }

    // Validate input
    if err := validateCompletionParams(params); err != nil {
        return nil, err
    }

    // ... rest of handler
}
```

**Test:**
```go
func TestValidation_InvalidPosition(t *testing.T) {
    tests := []struct {
        pos       protocol.Position
        shouldErr bool
    }{
        {protocol.Position{Line: 0, Character: 0}, false},
        {protocol.Position{Line: 100, Character: 50}, false},
        {protocol.Position{Line: -1, Character: 0}, true},
        {protocol.Position{Line: 0, Character: -1}, true},
        {protocol.Position{Line: 2000000, Character: 0}, true},
    }

    for _, tt := range tests {
        err := validateLSPPosition(tt.pos)
        if tt.shouldErr {
            assert.Error(t, err)
        } else {
            assert.NoError(t, err)
        }
    }
}
```

---

## 🔶 IMPORTANT Issues (Should Fix Before Production)

### 10. Handle New Directory Creation in File Watcher
**Priority:** P1
**Consensus:** 4/5 reviewers
**Estimated Time:** 1-2 hours

**Files to modify:**
- `pkg/lsp/watcher.go` - Add directory creation handling

**Changes:**
```go
func (fw *FileWatcher) watchLoop() {
    for {
        select {
        case <-fw.done:
            return

        case event, ok := <-fw.watcher.Events:
            if !ok {
                return
            }

            // Handle directory creation
            if event.Op&fsnotify.Create == fsnotify.Create {
                info, err := os.Stat(event.Name)
                if err == nil && info.IsDir() {
                    if !fw.shouldIgnore(event.Name) {
                        if err := fw.watcher.Add(event.Name); err != nil {
                            fw.logger.Warnf("Failed to watch new directory %s: %v", event.Name, err)
                        } else {
                            fw.logger.Debugf("Started watching new directory: %s", event.Name)
                        }
                    }
                }
            }

            // Filter: Only .dingo files
            if !isDingoFilePath(event.Name) {
                continue
            }

            // ... rest of file handling
        }
    }
}
```

---

### 11. Implement LRU Cache Eviction
**Priority:** P1
**Consensus:** 4/5 reviewers
**Estimated Time:** 2-3 hours

**Files to modify:**
- `pkg/lsp/sourcemap_cache.go` - Add LRU eviction

**Changes:**
```go
type cacheEntry struct {
    sm       *preprocessor.SourceMap
    lastUsed time.Time
}

type SourceMapCache struct {
    mu      sync.RWMutex
    entries map[string]*cacheEntry  // Changed from maps
    logger  Logger
    maxSize int
}

func (c *SourceMapCache) Get(goFilePath string) (*preprocessor.SourceMap, error) {
    // ... load logic ...

    // Update access time
    entry := &cacheEntry{
        sm:       sm,
        lastUsed: time.Now(),
    }

    // Evict if over limit
    if len(c.entries) >= c.maxSize {
        c.evictLRU()
    }

    c.entries[goFilePath] = entry
    return sm, nil
}

func (c *SourceMapCache) evictLRU() {
    var oldestKey string
    oldestTime := time.Now()

    for key, entry := range c.entries {
        if entry.lastUsed.Before(oldestTime) {
            oldestTime = entry.lastUsed
            oldestKey = key
        }
    }

    if oldestKey != "" {
        delete(c.entries, oldestKey)
        c.logger.Debugf("Evicted LRU source map: %s", oldestKey)
    }
}
```

---

### 12. Fix Context Propagation
**Priority:** P1
**Estimated Time:** 1 hour

**Files to modify:**
- `pkg/lsp/server.go` - Store and use server context

**Changes:**
```go
type Server struct {
    // ... existing fields
    ctx context.Context
}

func (s *Server) Serve(ctx context.Context, conn jsonrpc2.Conn) error {
    s.ctx = ctx
    s.ideConn = conn
    // ... rest
}

func (s *Server) handleDingoFileChange(dingoPath string) {
    // Use server context instead of background
    s.transpiler.OnFileChange(s.ctx, dingoPath)
}
```

---

### 13. Improve gopls stderr Logging
**Priority:** P1
**Estimated Time:** 30 minutes

**Files to modify:**
- `pkg/lsp/gopls_client.go` - Use bufio.Scanner

**Changes:**
```go
func (c *GoplsClient) logStderr(stderr io.Reader) {
    scanner := bufio.NewScanner(stderr)
    scanner.Buffer(make([]byte, 4096), 1024*1024) // 4KB initial, 1MB max

    for scanner.Scan() {
        line := scanner.Text()
        c.logger.Debugf("gopls stderr: %s", line)
    }

    if err := scanner.Err(); err != nil && err != io.EOF {
        c.logger.Debugf("stderr scan error: %v", err)
    }
}
```

---

### 14. Handle Translation Errors Properly
**Priority:** P1 UX
**Estimated Time:** 1 hour

**Files to modify:**
- `pkg/lsp/handlers.go` - Return LSP errors instead of silent degradation

**Changes:**
```go
translatedResult, err := s.translator.TranslateDefinitionLocations(result, GoToDingo)
if err != nil {
    s.config.Logger.Warnf("Definition response translation failed: %v", err)
    // Return error to user instead of untranslated result
    return reply(ctx, nil, fmt.Errorf("position translation failed: %w (try re-transpiling file)", err))
}
```

---

### 15. Add gopls Shutdown Timeout
**Priority:** P1
**Estimated Time:** 1 hour

**Files to modify:**
- `pkg/lsp/gopls_client.go` - Add timeout to Shutdown()

**Changes:**
```go
func (c *GoplsClient) Shutdown(ctx context.Context) error {
    // Send shutdown/exit
    // ...

    // Wait with timeout
    done := make(chan error, 1)
    go func() {
        done <- c.cmd.Wait()
    }()

    select {
    case err := <-done:
        return err
    case <-time.After(5 * time.Second):
        c.logger.Warnf("gopls didn't exit gracefully, killing process")
        if err := c.cmd.Process.Kill(); err != nil {
            return fmt.Errorf("failed to kill gopls: %w", err)
        }
        return fmt.Errorf("gopls shutdown timeout, force killed")
    }
}
```

---

### 16. Add Transpile Timeout and Error Parsing
**Priority:** P1 UX
**Estimated Time:** 2-3 hours

**Files to modify:**
- `pkg/lsp/transpiler.go` - Add timeout and parse errors

**Changes:**
```go
func (at *AutoTranspiler) TranspileFile(ctx context.Context, dingoPath string) (*TranspileResult, error) {
    // Add timeout
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    cmd := exec.CommandContext(ctx, "dingo", "build", dingoPath)

    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    err := cmd.Run()
    if err != nil {
        diagnostics := parseTranspileErrors(stderr.String(), dingoPath)
        return &TranspileResult{
            Success:     false,
            Diagnostics: diagnostics,
        }, fmt.Errorf("transpilation failed")
    }

    return &TranspileResult{Success: true}, nil
}
```

---

### 17. Pass VSCode Settings to LSP
**Priority:** P1 UX
**Estimated Time:** 1-2 hours

**Files to modify:**
- `pkg/lsp/server.go` - Read initializationOptions
- `editors/vscode/src/lspClient.ts` - Pass settings

**Changes:**
```go
// server.go
func (s *Server) handleInitialize(ctx context.Context, req *jsonrpc2.Request) (*protocol.InitializeResult, error) {
    // ... parse params ...

    if params.InitializationOptions != nil {
        var opts struct {
            TranspileOnSave bool `json:"transpileOnSave"`
        }
        if err := json.Unmarshal(params.InitializationOptions, &opts); err == nil {
            s.config.AutoTranspile = opts.TranspileOnSave
            s.config.Logger.Infof("Auto-transpile: %v", opts.TranspileOnSave)
        }
    }

    // ... rest
}
```

---

### 18. Add Resource Limits on gopls
**Priority:** P1 SECURITY
**Estimated Time:** 2-3 hours

**Files to modify:**
- `pkg/lsp/gopls_client.go` - Set process limits
- `pkg/lsp/sourcemap_cache.go` - Validate size before loading

**Changes:**
```go
// gopls_client.go
func (c *GoplsClient) start(goplsPath string) error {
    c.cmd = exec.Command(goplsPath, "-mode=stdio")

    // Set resource limits
    c.cmd.SysProcAttr = &syscall.SysProcAttr{
        Setrlimit: &syscall.Rlimit{
            Cur: 2 * 1024 * 1024 * 1024, // 2GB
            Max: 4 * 1024 * 1024 * 1024, // 4GB
        },
    }

    // ... rest
}

// sourcemap_cache.go
func (c *SourceMapCache) Get(goFilePath string) (*preprocessor.SourceMap, error) {
    mapPath := goFilePath + ".map"

    // Validate size before loading
    info, err := os.Stat(mapPath)
    if err != nil {
        return nil, err
    }

    const maxSourceMapSize = 50 * 1024 * 1024 // 50MB
    if info.Size() > maxSourceMapSize {
        return nil, fmt.Errorf("source map too large: %d MB (max 50MB)", info.Size()/(1024*1024))
    }

    // ... rest
}
```

---

### 19. Add Symlink Protection to File Watcher
**Priority:** P1 SECURITY
**Estimated Time:** 1-2 hours

**Files to modify:**
- `pkg/lsp/watcher.go` - Validate symlinks in watchRecursive

**Changes:**
```go
func (fw *FileWatcher) watchRecursive(root string) error {
    seen := make(map[string]bool)

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
            if err := fw.watcher.Add(realPath); err != nil {
                fw.logger.Warnf("Failed to watch %s: %v", realPath, err)
            }
        }

        return nil
    })
}
```

---

### 20. Add Stale Source Map Detection
**Priority:** P1 UX
**Estimated Time:** 1-2 hours

**Files to modify:**
- `pkg/lsp/sourcemap_cache.go` - Compare modification times

**Changes:**
```go
func (c *SourceMapCache) Get(goFilePath string) (*preprocessor.SourceMap, error) {
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

    // ... load and validate
}
```

---

### 21. Optimize Position Translation (Binary Search)
**Priority:** P1 PERFORMANCE
**Estimated Time:** 2-3 hours

**Files to modify:**
- `pkg/preprocessor/sourcemap.go` - Add binary search

**Changes:**
```go
// Sort mappings on load
func (sm *SourceMap) sortMappings() {
    sort.Slice(sm.Mappings, func(i, j int) bool {
        if sm.Mappings[i].GeneratedLine != sm.Mappings[j].GeneratedLine {
            return sm.Mappings[i].GeneratedLine < sm.Mappings[j].GeneratedLine
        }
        return sm.Mappings[i].GeneratedColumn < sm.Mappings[j].GeneratedColumn
    })
}

func (sm *SourceMap) MapToOriginal(genLine, genCol int) (int, int) {
    // Binary search for line
    idx := sort.Search(len(sm.Mappings), func(i int) bool {
        return sm.Mappings[i].GeneratedLine >= genLine
    })

    if idx >= len(sm.Mappings) || sm.Mappings[idx].GeneratedLine != genLine {
        return genLine, genCol // No mapping found
    }

    // Linear scan within line (typically 1-5 mappings)
    bestMatch := &sm.Mappings[idx]
    for i := idx; i < len(sm.Mappings) && sm.Mappings[i].GeneratedLine == genLine; i++ {
        m := &sm.Mappings[i]
        if m.GeneratedColumn <= genCol && genCol < m.GeneratedColumn+m.Length {
            offset := genCol - m.GeneratedColumn
            return m.OriginalLine, m.OriginalColumn + offset
        }
        // Track closest match
        if abs(m.GeneratedColumn-genCol) < abs(bestMatch.GeneratedColumn-genCol) {
            bestMatch = m
        }
    }

    // Use closest match
    offset := genCol - bestMatch.GeneratedColumn
    return bestMatch.OriginalLine, bestMatch.OriginalColumn + offset
}
```

---

## 📋 Testing Checklist

After implementing all fixes:

### Unit Tests
- [ ] All existing tests pass (40/40)
- [ ] New tests added for all critical fixes
- [ ] Race detector passes (`go test -race ./...`)

### Integration Tests
- [ ] gopls communication (init, completion, hover, definition)
- [ ] Diagnostic publishing (transpile errors, Go type errors)
- [ ] Source map translation (positions, ranges, diagnostics)
- [ ] Auto-transpile on save
- [ ] gopls crash recovery
- [ ] Concurrent request handling (100+ simultaneous)

### Manual VSCode Testing
- [ ] Autocomplete works at correct positions
- [ ] Hover shows types
- [ ] Go-to-definition jumps correctly
- [ ] Diagnostics appear inline with correct positions
- [ ] LSP recovers from gopls crash
- [ ] Auto-transpile triggers on save (when enabled)
- [ ] New directories are watched
- [ ] Settings (`transpileOnSave`) respected

### Security Testing
- [ ] Path traversal attempts rejected
- [ ] Symlink escapes prevented
- [ ] Large source maps rejected (>50MB)
- [ ] gopls memory limits enforced

---

## 📊 Estimated Timeline

| Phase | Tasks | Time | Cumulative |
|-------|-------|------|-----------|
| **Phase 1: Critical** | Items 1-9 | 15-20 hours | 2-3 days |
| **Phase 2: Important** | Items 10-21 | 18-24 hours | 4-6 days |
| **Testing** | All checklists | 8-12 hours | 5-8 days |
| **Total** | | **41-56 hours** | **1-2 weeks** |

---

## 🎯 Success Criteria

Before marking Phase V complete:

1. ✅ All CRITICAL issues fixed and tested
2. ✅ All IMPORTANT issues fixed and tested
3. ✅ 40/40 unit tests passing
4. ✅ Race detector clean
5. ✅ Manual VSCode checklist 100% complete
6. ✅ Security tests passing
7. ✅ Performance still exceeds targets (re-run benchmarks)
8. ✅ Beta testing with 3+ developers for 1 week (zero crashes)

---

**Action Items Complete**
**Total Items:** 21 (9 CRITICAL, 12 IMPORTANT)
**Recommendation:** Fix CRITICAL first (Phase 1), then IMPORTANT (Phase 2), then test thoroughly

# Phase V: Language Server Foundation - Architecture Plan

**Session:** 20251118-152749-phase5-lsp-foundation
**Created:** 2025-11-18
**Architect:** golang-architect agent
**Scope:** LSP support for Phase 3 features only

## Executive Summary

This plan establishes the foundation for Dingo IDE support via a Language Server Protocol (LSP) proxy architecture. The server wraps gopls (Go's native language server) and translates positions between `.dingo` files and their transpiled `.go` counterparts using existing source map infrastructure.

**Strategy:** Minimal wrapper around gopls + bidirectional position translation = Full IDE support with minimal code.

**Timeline:** 1-2 weeks for iteration 1 (foundation only)

## Architecture Overview

### Three-Layer Architecture

```
┌─────────────────────────────────────────────────────────────┐
│ Layer 1: IDE (VSCode, Neovim, etc.)                         │
│ Speaks: LSP protocol                                        │
│ Sees: .dingo files                                          │
└─────────────────────────────────────────────────────────────┘
                           ↕ LSP over stdin/stdout
┌─────────────────────────────────────────────────────────────┐
│ Layer 2: dingo-lsp (Our Proxy Server)                       │
│ • Receives LSP requests for .dingo files                    │
│ • Translates positions: .dingo → .go (using source maps)    │
│ • Forwards translated requests to gopls                     │
│ • Receives gopls responses                                  │
│ • Translates positions back: .go → .dingo                   │
│ • Returns responses to IDE                                  │
└─────────────────────────────────────────────────────────────┘
                           ↕ LSP over stdin/stdout
┌─────────────────────────────────────────────────────────────┐
│ Layer 3: gopls (Native Go Language Server)                  │
│ Speaks: LSP protocol                                        │
│ Sees: .go files (transpiled)                                │
└─────────────────────────────────────────────────────────────┘
```

### Key Insight: Leverage gopls Completely

We do NOT implement any Go language features. We simply:
1. Transpile `.dingo` → `.go` (already works)
2. Translate LSP request positions (`.dingo` coords → `.go` coords)
3. Forward request to gopls (which sees `.go` files)
4. Translate LSP response positions (`.go` coords → `.dingo` coords)
5. Return response to IDE

**Result:** Full IDE support (autocomplete, hover, definitions, diagnostics) with ~500 LOC.

## Component Architecture

### 1. Binary: `cmd/dingo-lsp/main.go`

**Responsibilities:**
- LSP server entry point
- stdio-based JSON-RPC communication
- Lifecycle management (initialization, shutdown)
- Logging/debugging infrastructure

**Key Operations:**
```go
func main() {
    // 1. Initialize LSP server
    server := lsp.NewServer()

    // 2. Start gopls subprocess
    gopls := lsp.NewGoplsProxy()

    // 3. Create LSP transport
    transport := jsonrpc.NewStdioTransport()

    // 4. Event loop: read LSP requests, proxy to gopls, return responses
    for {
        request := transport.ReadRequest()
        response := server.HandleRequest(request, gopls)
        transport.WriteResponse(response)
    }
}
```

**Dependencies:**
- `go.lsp.dev/protocol` - LSP protocol types
- `go.lsp.dev/jsonrpc2` - JSON-RPC 2.0 transport
- Internal: `pkg/lsp`, `pkg/sourcemap`

**Estimated Size:** ~150 LOC

### 2. LSP Proxy: `pkg/lsp/proxy.go`

**Responsibilities:**
- Main request/response handler
- Position translation orchestration
- gopls subprocess management
- Request routing logic

**Core Logic:**
```go
type Proxy struct {
    gopls       *GoplsClient
    mapCache    *SourceMapCache
    workspaces  map[string]*Workspace
}

func (p *Proxy) HandleRequest(req *protocol.Request) (*protocol.Response, error) {
    // 1. Determine if request is for .dingo file
    if !isDingoFile(req.URI) {
        // Not a Dingo file, forward directly to gopls
        return p.gopls.Forward(req)
    }

    // 2. Get source map for this file
    sm, err := p.mapCache.Get(dingoFileToGoFile(req.URI))
    if err != nil {
        // No source map, file might not be transpiled yet
        return makeError("File not transpiled")
    }

    // 3. Translate request positions (.dingo → .go)
    goReq := translateRequestPositions(req, sm, DingoToGo)

    // 4. Forward to gopls
    goResp, err := p.gopls.Forward(goReq)
    if err != nil {
        return nil, err
    }

    // 5. Translate response positions (.go → .dingo)
    dingoResp := translateResponsePositions(goResp, sm, GoToDingo)

    return dingoResp, nil
}
```

**Key Functions:**
- `HandleRequest(req) → (resp, error)` - Main entry point
- `translateRequestPositions(req, sm, direction) → req'` - Position translation
- `translateResponsePositions(resp, sm, direction) → resp'` - Response translation
- `isDingoFile(uri) → bool` - File type detection

**Estimated Size:** ~300 LOC

### 3. gopls Client: `pkg/lsp/gopls_client.go`

**Responsibilities:**
- Manage gopls subprocess lifecycle
- Send/receive LSP messages to/from gopls
- Handle gopls crashes/restarts
- Initialization handshake

**Implementation:**
```go
type GoplsClient struct {
    cmd      *exec.Cmd
    stdin    io.WriteCloser
    stdout   io.ReadCloser
    stderr   io.ReadCloser
    transport *jsonrpc2.Conn
}

func NewGoplsClient() (*GoplsClient, error) {
    // 1. Find gopls binary
    goplsPath, err := exec.LookPath("gopls")
    if err != nil {
        return nil, fmt.Errorf("gopls not found: %w", err)
    }

    // 2. Start subprocess
    cmd := exec.Command(goplsPath, "-mode=stdio")
    stdin, _ := cmd.StdinPipe()
    stdout, _ := cmd.StdoutPipe()
    stderr, _ := cmd.StderrPipe()

    if err := cmd.Start(); err != nil {
        return nil, err
    }

    // 3. Create JSON-RPC transport
    transport := jsonrpc2.NewConn(stdin, stdout)

    return &GoplsClient{
        cmd:       cmd,
        stdin:     stdin,
        stdout:    stdout,
        stderr:    stderr,
        transport: transport,
    }, nil
}

func (c *GoplsClient) Forward(req *protocol.Request) (*protocol.Response, error) {
    // Send request to gopls, wait for response
    return c.transport.Call(req.Method, req.Params)
}

func (c *GoplsClient) Shutdown() error {
    // Graceful shutdown
    c.transport.Close()
    return c.cmd.Wait()
}
```

**Estimated Size:** ~200 LOC

### 4. Source Map Cache: `pkg/lsp/sourcemap_cache.go`

**Responsibilities:**
- Load `.go.map` files on demand
- Cache in memory (avoid repeated disk reads)
- Watch for source map changes
- Invalidate stale entries

**Implementation:**
```go
type SourceMapCache struct {
    mu      sync.RWMutex
    maps    map[string]*preprocessor.SourceMap
    watcher *fsnotify.Watcher
}

func NewSourceMapCache() (*SourceMapCache, error) {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return nil, err
    }

    cache := &SourceMapCache{
        maps:    make(map[string]*preprocessor.SourceMap),
        watcher: watcher,
    }

    // Start file watcher goroutine
    go cache.watchLoop()

    return cache, nil
}

func (c *SourceMapCache) Get(goFilePath string) (*preprocessor.SourceMap, error) {
    mapPath := goFilePath + ".map"

    // Check cache first
    c.mu.RLock()
    if sm, ok := c.maps[mapPath]; ok {
        c.mu.RUnlock()
        return sm, nil
    }
    c.mu.RUnlock()

    // Load from disk
    data, err := os.ReadFile(mapPath)
    if err != nil {
        return nil, fmt.Errorf("source map not found: %s", mapPath)
    }

    sm, err := preprocessor.FromJSON(data)
    if err != nil {
        return nil, err
    }

    // Cache it
    c.mu.Lock()
    c.maps[mapPath] = sm
    c.watcher.Add(mapPath)
    c.mu.Unlock()

    return sm, nil
}

func (c *SourceMapCache) watchLoop() {
    for {
        select {
        case event := <-c.watcher.Events:
            if event.Op&fsnotify.Write == fsnotify.Write {
                // Invalidate cache entry
                c.mu.Lock()
                delete(c.maps, event.Name)
                c.mu.Unlock()
            }
        }
    }
}
```

**Estimated Size:** ~150 LOC

### 5. Position Translator: `pkg/lsp/translator.go`

**Responsibilities:**
- Translate LSP positions using source maps
- Handle different LSP request/response types
- Manage edge cases (unmapped positions, multi-line expansions)

**Core Translation Logic:**
```go
type Direction int

const (
    DingoToGo Direction = iota
    GoToDingo
)

func translateRequestPositions(req *protocol.Request, sm *preprocessor.SourceMap, dir Direction) *protocol.Request {
    // Clone request
    newReq := *req

    // Switch based on request method
    switch req.Method {
    case "textDocument/completion":
        params := req.Params.(protocol.CompletionParams)
        params.Position = translatePosition(params.Position, sm, dir)
        newReq.Params = params

    case "textDocument/definition":
        params := req.Params.(protocol.DefinitionParams)
        params.Position = translatePosition(params.Position, sm, dir)
        newReq.Params = params

    case "textDocument/hover":
        params := req.Params.(protocol.HoverParams)
        params.Position = translatePosition(params.Position, sm, dir)
        newReq.Params = params

    // ... handle other methods
    }

    return &newReq
}

func translatePosition(pos protocol.Position, sm *preprocessor.SourceMap, dir Direction) protocol.Position {
    line := int(pos.Line) + 1  // LSP is 0-based, source maps are 1-based
    col := int(pos.Character) + 1

    var newLine, newCol int
    if dir == DingoToGo {
        newLine, newCol = sm.MapToGenerated(line, col)
    } else {
        newLine, newCol = sm.MapToOriginal(line, col)
    }

    return protocol.Position{
        Line:      uint32(newLine - 1),
        Character: uint32(newCol - 1),
    }
}
```

**Critical Methods to Support:**
1. `textDocument/completion` - Autocomplete
2. `textDocument/definition` - Go to definition
3. `textDocument/hover` - Hover information
4. `textDocument/publishDiagnostics` - Error diagnostics
5. `textDocument/documentSymbol` - Symbol navigation

**Estimated Size:** ~250 LOC

### 6. File Watcher: `pkg/lsp/watcher.go`

**Responsibilities:**
- Watch `.dingo` files for changes
- Trigger transpilation on save
- Notify gopls of `.go` file updates
- Debounce rapid changes

**Implementation:**
```go
type FileWatcher struct {
    watcher     *fsnotify.Watcher
    proxy       *Proxy
    debouncer   *time.Timer
    pendingDirs map[string]bool
}

func NewFileWatcher(proxy *Proxy, workspaceDirs []string) (*FileWatcher, error) {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return nil, err
    }

    fw := &FileWatcher{
        watcher:     watcher,
        proxy:       proxy,
        pendingDirs: make(map[string]bool),
    }

    // Watch all workspace directories
    for _, dir := range workspaceDirs {
        if err := watcher.Add(dir); err != nil {
            return nil, err
        }
    }

    go fw.watchLoop()

    return fw, nil
}

func (fw *FileWatcher) watchLoop() {
    for {
        select {
        case event := <-fw.watcher.Events:
            if !isDingoFile(event.Name) {
                continue
            }

            if event.Op&fsnotify.Write == fsnotify.Write {
                fw.handleDingoFileChange(event.Name)
            }
        }
    }
}

func (fw *FileWatcher) handleDingoFileChange(dingoPath string) {
    // Trigger transpilation
    goPath := strings.TrimSuffix(dingoPath, ".dingo") + ".go"

    cmd := exec.Command("dingo", "build", dingoPath)
    if err := cmd.Run(); err != nil {
        // Log error, but don't crash
        log.Printf("Transpilation failed: %v", err)
        return
    }

    // Notify gopls of .go file change
    fw.proxy.gopls.NotifyFileChange(goPath)
}
```

**Estimated Size:** ~150 LOC

### 7. VSCode Extension (Minimal)

**Location:** `editors/vscode-dingo/`

**Files:**
- `package.json` - Extension manifest
- `extension.js` - Extension entry point
- `syntaxes/dingo.tmLanguage.json` - Syntax highlighting
- `language-configuration.json` - Bracket matching, comments

**Key Configuration:**
```json
{
  "name": "dingo-lang",
  "displayName": "Dingo Language Support",
  "version": "0.1.0",
  "engines": { "vscode": "^1.60.0" },
  "activationEvents": ["onLanguage:dingo"],
  "main": "./extension.js",
  "contributes": {
    "languages": [{
      "id": "dingo",
      "extensions": [".dingo"],
      "aliases": ["Dingo"],
      "configuration": "./language-configuration.json"
    }],
    "grammars": [{
      "language": "dingo",
      "scopeName": "source.dingo",
      "path": "./syntaxes/dingo.tmLanguage.json"
    }],
    "configuration": {
      "title": "Dingo",
      "properties": {
        "dingo.lsp.path": {
          "type": "string",
          "default": "dingo-lsp",
          "description": "Path to dingo-lsp binary"
        }
      }
    }
  }
}
```

**Extension Logic:**
```javascript
const vscode = require('vscode');
const { LanguageClient } = require('vscode-languageclient/node');

function activate(context) {
    // Start dingo-lsp server
    const serverOptions = {
        command: 'dingo-lsp',
        args: [],
        options: { stdio: 'pipe' }
    };

    const clientOptions = {
        documentSelector: [{ scheme: 'file', language: 'dingo' }],
    };

    const client = new LanguageClient(
        'dingo-lsp',
        'Dingo Language Server',
        serverOptions,
        clientOptions
    );

    client.start();

    context.subscriptions.push(client);
}

exports.activate = activate;
```

**Estimated Size:** ~100 LOC + grammar file

## Source Map Integration

### Existing Infrastructure (Reuse)

**GOOD NEWS:** Source maps already exist and work!

**Location:** `pkg/preprocessor/sourcemap.go`

**Data Structure:**
```go
type SourceMap struct {
    Mappings []Mapping `json:"mappings"`
}

type Mapping struct {
    GeneratedLine   int    `json:"generated_line"`
    GeneratedColumn int    `json:"generated_column"`
    OriginalLine    int    `json:"original_line"`
    OriginalColumn  int    `json:"original_column"`
    Length          int    `json:"length"`
    Name            string `json:"name,omitempty"`
}
```

**Key Methods (Already Implemented):**
- `MapToOriginal(line, col) → (line, col)` - Go position → Dingo position
- `MapToGenerated(line, col) → (line, col)` - Dingo position → Go position
- `ToJSON() → []byte` - Serialize
- `FromJSON(data) → *SourceMap` - Deserialize

**What We Need to Add:**
- **Nothing for basic functionality!** The existing implementation is sufficient.
- **Future Enhancement:** Optimize lookup (binary search instead of linear scan)

### Source Map Format (Actual)

**Generated by transpiler:**
```json
{
  "mappings": [
    {
      "generated_line": 4,
      "generated_column": 1,
      "original_line": 4,
      "original_column": 15,
      "length": 1,
      "name": "error_prop"
    }
  ]
}
```

**Example:** Error propagation (`x?`) expands to 7 lines of Go code:
- Dingo line 4, col 15 (the `?` character)
- Maps to Go lines 4-10, col 1 (the expanded error handling)

**Translation Strategy:**
- **Dingo → Go:** Use `MapToGenerated()` - finds the generated code that originated from Dingo position
- **Go → Dingo:** Use `MapToOriginal()` - finds the original Dingo source that generated Go code
- **Unmapped positions:** Return as-is (assume 1:1 mapping for unmodified code)

## Request/Response Flow

### Example: Autocomplete on `.dingo` File

**Step-by-Step Flow:**

1. **User types in VSCode at `example.dingo:10:15`**
   - VSCode sends: `textDocument/completion` request
   - URI: `file:///path/example.dingo`
   - Position: `{line: 9, character: 14}` (0-based)

2. **dingo-lsp receives request**
   - Detects `.dingo` file
   - Loads source map: `example.go.map`

3. **dingo-lsp translates position**
   - Dingo: line 10, col 15 (1-based)
   - Source map lookup: `MapToGenerated(10, 15)`
   - Result: Go line 15, col 20 (multi-line expansion)

4. **dingo-lsp forwards to gopls**
   - Modified request:
     - URI: `file:///path/example.go` (changed!)
     - Position: `{line: 14, character: 19}` (0-based)

5. **gopls processes request**
   - Sees `example.go` (valid Go code)
   - Uses Go type checker
   - Returns completion items with Go positions

6. **dingo-lsp receives response**
   - Completion items with Go positions
   - Translates each position back to Dingo using `MapToOriginal()`

7. **dingo-lsp returns to VSCode**
   - Completion items with Dingo positions
   - VSCode displays autocomplete menu

**Performance:** ~5-10ms for position translation (negligible)

### Edge Cases Handled

**Case 1: Unmapped Position**
- Example: Comment in `.dingo` file (passes through unchanged)
- Strategy: `MapToGenerated()` returns input position if no mapping found
- Result: 1:1 mapping assumed, gopls sees identical position

**Case 2: Multi-line Expansion**
- Example: `x?` → 7 lines of Go code
- Strategy: All 7 Go lines map back to same Dingo line/col
- Result: gopls errors on any of 7 lines → translated to original `x?` position

**Case 3: Source Map Missing**
- Example: `.dingo` file not transpiled yet
- Strategy: Return LSP error: "File not transpiled. Run `dingo build` first."
- Result: IDE shows error, user transpiles, then LSP works

**Case 4: Stale Source Map**
- Example: `.dingo` edited but not re-transpiled
- Strategy: File watcher detects change, triggers transpilation, invalidates cache
- Result: LSP auto-updates after transpilation completes

## Testing Strategy

### Unit Tests

**1. Source Map Translation (`pkg/lsp/translator_test.go`)**
```go
func TestPositionTranslation(t *testing.T) {
    sm := &preprocessor.SourceMap{
        Mappings: []preprocessor.Mapping{
            {OriginalLine: 5, OriginalColumn: 10, GeneratedLine: 8, GeneratedColumn: 15, Length: 3},
        },
    }

    // Dingo → Go
    goLine, goCol := sm.MapToGenerated(5, 10)
    assert.Equal(t, 8, goLine)
    assert.Equal(t, 15, goCol)

    // Go → Dingo
    dingoLine, dingoCol := sm.MapToOriginal(8, 15)
    assert.Equal(t, 5, dingoLine)
    assert.Equal(t, 10, dingoCol)
}
```

**2. gopls Client (`pkg/lsp/gopls_client_test.go`)**
- Test subprocess lifecycle
- Test request forwarding
- Test crash recovery

**3. Source Map Cache (`pkg/lsp/sourcemap_cache_test.go`)**
- Test cache hit/miss
- Test invalidation on file change
- Test concurrent access

### Integration Tests

**1. End-to-End LSP (`pkg/lsp/integration_test.go`)**
```go
func TestCompletionRequest(t *testing.T) {
    // 1. Create test .dingo file
    dingoFile := "test.dingo"
    goFile := "test.go"
    mapFile := "test.go.map"

    // 2. Transpile (creates .go and .map)
    transpile(dingoFile)

    // 3. Start dingo-lsp server
    server := lsp.NewServer()

    // 4. Send completion request
    req := &protocol.CompletionRequest{
        URI:      "file:///test.dingo",
        Position: protocol.Position{Line: 5, Character: 10},
    }

    resp, err := server.HandleRequest(req)
    assert.NoError(t, err)

    // 5. Verify response has Dingo positions (not Go positions)
    assert.Equal(t, "file:///test.dingo", resp.Items[0].URI)
}
```

**2. File Watcher Integration**
- Modify `.dingo` file
- Verify transpilation triggered
- Verify gopls receives update notification

### Manual Testing

**VSCode Extension Test Plan:**
1. Install extension
2. Open `.dingo` file
3. Test autocomplete (Ctrl+Space)
4. Test go-to-definition (F12)
5. Test hover (mouse hover)
6. Test diagnostics (syntax error)
7. Test document symbols (Ctrl+Shift+O)

## Implementation Phases

### Phase 5.1: Core Proxy (Week 1, Days 1-3)

**Deliverables:**
- `cmd/dingo-lsp/main.go` - Binary skeleton
- `pkg/lsp/proxy.go` - Request router
- `pkg/lsp/gopls_client.go` - gopls subprocess
- `pkg/lsp/translator.go` - Position translation
- `pkg/lsp/sourcemap_cache.go` - Source map caching

**Tests:**
- Unit tests for translation
- gopls client tests
- Source map cache tests

**Success Criteria:**
- Can forward basic LSP request to gopls
- Position translation works for completion requests
- gopls subprocess starts/stops cleanly

**Estimated Effort:** 2-3 days

### Phase 5.2: LSP Methods (Week 1, Days 4-5)

**Deliverables:**
- Support for critical LSP methods:
  - `textDocument/completion`
  - `textDocument/definition`
  - `textDocument/hover`
  - `textDocument/publishDiagnostics`

**Tests:**
- Integration tests for each method
- Edge case handling

**Success Criteria:**
- All 4 methods work end-to-end
- Positions correctly translated
- Errors handled gracefully

**Estimated Effort:** 2 days

### Phase 5.3: File Watching (Week 2, Days 1-2)

**Deliverables:**
- `pkg/lsp/watcher.go` - File watcher
- Auto-transpilation on save
- gopls notification on .go update

**Tests:**
- File watcher tests
- Debouncing tests
- Transpilation trigger tests

**Success Criteria:**
- Save `.dingo` → auto-transpiles → gopls updates
- No excessive transpilations (debouncing works)
- Errors don't crash watcher

**Estimated Effort:** 1-2 days

### Phase 5.4: VSCode Extension (Week 2, Days 3-4)

**Deliverables:**
- `editors/vscode-dingo/` - Extension package
- Syntax highlighting grammar
- LSP client configuration
- README with installation instructions

**Tests:**
- Manual testing in VSCode
- Extension packaging
- Installation verification

**Success Criteria:**
- Extension installs from .vsix
- Syntax highlighting works
- LSP connects and provides features
- No crashes or errors

**Estimated Effort:** 1-2 days

### Phase 5.5: Polish & Documentation (Week 2, Day 5)

**Deliverables:**
- Comprehensive README
- Architecture diagram (ASCII art)
- Debugging guide
- Performance benchmarks

**Tests:**
- Documentation review
- Example projects
- Troubleshooting guide

**Success Criteria:**
- User can set up LSP in <5 minutes
- Debugging instructions clear
- Known issues documented

**Estimated Effort:** 1 day

## Performance Considerations

### Expected Latency Budget

**Target:** <100ms total latency for autocomplete

**Breakdown:**
- VSCode → dingo-lsp: ~5ms (local IPC)
- Position translation: ~1ms (hash map lookup)
- dingo-lsp → gopls: ~5ms (local IPC)
- gopls processing: ~50ms (Go type checking)
- gopls → dingo-lsp: ~5ms
- Position translation: ~1ms
- dingo-lsp → VSCode: ~5ms
- **Total:** ~72ms ✅ Well under budget

### Optimization Strategies

**1. Source Map Cache**
- In-memory cache (avoid disk reads)
- LRU eviction (max 100 files)
- File watcher invalidation (no stale data)

**2. Position Lookup**
- Binary search for mappings (currently linear)
- Index by line number for O(log n) lookup
- Pre-compute reverse mappings

**3. gopls Connection Pooling**
- Reuse subprocess (don't restart per request)
- Keep stdin/stdout pipes open
- Graceful shutdown on idle timeout

**4. Debouncing**
- File watcher: 100ms debounce on .dingo changes
- Avoid multiple transpilations on rapid saves

## Error Handling

### Error Categories

**1. User Errors (Informative)**
- `.dingo` file not transpiled → "Run `dingo build` first"
- Invalid Dingo syntax → Show transpilation error in diagnostics
- Missing gopls → "Install gopls: go install golang.org/x/tools/gopls@latest"

**2. Recoverable Errors (Retry)**
- gopls crash → Auto-restart subprocess, show notification
- Source map parse error → Invalidate cache, retry
- Transpilation failure → Show error, don't crash LSP

**3. Fatal Errors (Shutdown)**
- Cannot bind to stdin/stdout → Exit with error code
- Workspace path inaccessible → Exit with error code

### Graceful Degradation

**If gopls unavailable:**
- LSP still starts
- Returns empty completion/hover results
- Shows warning notification: "gopls not found"

**If source map missing:**
- LSP still responds
- Assumes 1:1 mapping (pass-through)
- May show incorrect positions (but doesn't crash)

**If transpilation fails:**
- LSP still running
- Shows diagnostics with error
- User fixes code, re-saves, retry

## Security Considerations

**1. File Access**
- Only access files within workspace
- Reject LSP requests for paths outside workspace
- No arbitrary command execution (only `dingo build`)

**2. Subprocess Security**
- gopls runs with same permissions as LSP
- No shell expansion (use `exec.Command` directly)
- Validate gopls binary path (no user-controlled injection)

**3. Input Validation**
- Validate LSP request structure
- Sanitize file paths (prevent directory traversal)
- Limit response sizes (prevent DoS)

## Dependencies

### Go Packages

**Core LSP:**
- `go.lsp.dev/protocol` - LSP types and protocol (v0.12+)
- `go.lsp.dev/jsonrpc2` - JSON-RPC 2.0 transport
- `go.lsp.dev/uri` - URI handling

**File Watching:**
- `github.com/fsnotify/fsnotify` - Cross-platform file watcher

**Internal:**
- `github.com/MadAppGang/dingo/pkg/preprocessor` - Source map types
- `github.com/MadAppGang/dingo/pkg/generator` - Transpiler (for file watcher)

**External Tools:**
- `gopls` - Must be installed separately (not embedded)

### VSCode Extension Dependencies

**NPM Packages:**
- `vscode` (^1.60.0)
- `vscode-languageclient` (^7.0.0)

## Risks & Mitigation

### Risk 1: Source Map Format Changes

**Risk:** Phase 4 modifies source map structure

**Likelihood:** Low (Phase 4 works on different features)

**Impact:** High (would break LSP)

**Mitigation:**
- Document current format in this plan
- Add version field to source maps (`"version": 1`)
- LSP checks version, rejects incompatible maps
- Coordinate with Phase 4 agent via `phase4-5-coordination.md`

### Risk 2: gopls API Changes

**Risk:** gopls updates break compatibility

**Likelihood:** Low (gopls is stable)

**Impact:** Medium (may require LSP updates)

**Mitigation:**
- Pin gopls version in documentation (e.g., "gopls v0.11+")
- Test with multiple gopls versions
- Use stable LSP features only (no experimental)

### Risk 3: Performance Degradation

**Risk:** Position translation becomes bottleneck

**Likelihood:** Low (simple hash map lookup)

**Impact:** Medium (user-visible lag)

**Mitigation:**
- Benchmark position translation (<1ms target)
- Optimize if >10ms (binary search, indexing)
- Profile in production environments

### Risk 4: Multi-line Expansion Ambiguity

**Risk:** One Dingo line → 10 Go lines, unclear which Go line user wants

**Likelihood:** Medium (error propagation does this)

**Impact:** Low (gopls handles it reasonably)

**Mitigation:**
- Map first Go line to Dingo position (consistent heuristic)
- Future: Use token-based mapping (not just line/col)

### Risk 5: Race Conditions in File Watcher

**Risk:** File changes during transpilation

**Likelihood:** Low (file watching is debounced)

**Impact:** Medium (stale source maps)

**Mitigation:**
- Debounce file events (100ms)
- Invalidate cache atomically
- Retry on source map load failure

## Future Enhancements (Post-Iteration 1)

### Phase 5.2: Advanced LSP Features

**After Phase 4 completes:**
- Lambda function support (fold/unfold, inline)
- Ternary operator hover (show expanded form)
- Null coalescing autocomplete
- Tuple destructuring hints

**Additional LSP Methods:**
- `textDocument/rename` - Rename refactoring
- `textDocument/references` - Find all references
- `textDocument/formatting` - Code formatting (dingo fmt)
- `textDocument/codeAction` - Quick fixes

### Phase 5.3: Multi-file Analysis

**Features:**
- Cross-file go-to-definition (Dingo → Dingo)
- Workspace-wide search
- Dependency graph visualization

**Challenges:**
- Multiple source maps (one per file)
- File relationships (imports)

### Phase 5.4: Debugging Support

**Features:**
- DAP (Debug Adapter Protocol) server
- Map breakpoints: Dingo → Go positions
- Map stack traces: Go → Dingo positions
- Variable inspection in Dingo terms

**Challenges:**
- Debugging generated code (not original)
- Source maps for runtime (not just LSP)

## Open Questions (Gaps)

See `gaps.json` for structured questions to ask user.

**Summary of uncertainties:**
1. gopls version requirements (minimum version?)
2. File watcher strategy (workspace root vs. individual files?)
3. Error reporting preferences (diagnostics vs. notifications?)
4. VSCode extension publishing (marketplace vs. manual .vsix?)
5. Transpilation trigger (auto on save vs. manual?)
6. Source map format stability (will Phase 4 modify it?)

## Success Metrics

### Must Have (Iteration 1)

- ✅ Autocomplete works in `.dingo` files
- ✅ Go-to-definition works (Dingo → Dingo)
- ✅ Hover shows type information
- ✅ Diagnostics show in correct positions
- ✅ No crashes on malformed input
- ✅ <100ms autocomplete latency

### Nice to Have

- ✅ Document symbols work
- ✅ Find references works
- ⏳ Rename refactoring (defer to iteration 2)
- ⏳ Code actions (defer to iteration 2)

### Quality Metrics

- Test coverage: >80%
- gopls subprocess uptime: >99.9%
- Position translation accuracy: 100%
- User setup time: <5 minutes

## Deliverables Checklist

### Code

- [ ] `cmd/dingo-lsp/main.go`
- [ ] `pkg/lsp/proxy.go`
- [ ] `pkg/lsp/gopls_client.go`
- [ ] `pkg/lsp/translator.go`
- [ ] `pkg/lsp/sourcemap_cache.go`
- [ ] `pkg/lsp/watcher.go`
- [ ] `editors/vscode-dingo/extension.js`
- [ ] `editors/vscode-dingo/package.json`
- [ ] `editors/vscode-dingo/syntaxes/dingo.tmLanguage.json`

### Tests

- [ ] `pkg/lsp/translator_test.go`
- [ ] `pkg/lsp/gopls_client_test.go`
- [ ] `pkg/lsp/sourcemap_cache_test.go`
- [ ] `pkg/lsp/integration_test.go`
- [ ] `pkg/lsp/watcher_test.go`

### Documentation

- [ ] `pkg/lsp/README.md` - Architecture overview
- [ ] `editors/vscode-dingo/README.md` - Extension installation
- [ ] `docs/lsp-debugging.md` - Debugging guide
- [ ] `docs/lsp-architecture.md` - Detailed architecture (this document)

## Timeline Summary

**Week 1:**
- Days 1-3: Core proxy (gopls client, translator, cache)
- Days 4-5: LSP methods (completion, definition, hover, diagnostics)

**Week 2:**
- Days 1-2: File watching (auto-transpilation)
- Days 3-4: VSCode extension (syntax, LSP client)
- Day 5: Polish, documentation, testing

**Total:** 10 days (2 calendar weeks)

## Conclusion

This architecture provides a **minimal, robust foundation** for Dingo LSP support by leveraging gopls completely. We implement ~1200 LOC of proxy/translation logic and get full IDE support for free.

**Key Advantages:**
1. **Simplicity:** No language features reimplemented
2. **Correctness:** gopls provides accurate Go semantics
3. **Maintainability:** Small codebase, clear boundaries
4. **Extensibility:** Easy to add Phase 4 features in iteration 2

**Next Steps:**
1. Review this plan with user
2. Resolve open questions (see `gaps.json`)
3. Begin Phase 5.1 implementation
4. Coordinate with Phase 4 via coordination document

---

**Plan Status:** Draft v1.0
**Awaiting:** User approval and gap resolution

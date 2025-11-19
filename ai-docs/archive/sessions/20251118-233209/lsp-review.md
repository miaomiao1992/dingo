# Code Review: Phase V - LSP Foundation Implementation

This review analyzes the implementation of Dingo's Language Server Protocol (LSP) foundation, focusing on bidirectional source map translation, gopls proxying, auto-transpilation, and IDE integration.

## ✅ Strengths

### Architecture Excellence
- **Clean Proxy Architecture**: LSP server effectively proxies most requests to gopls while handling Dingo-specific translation, following established patterns like templ's gopls proxy approach.
- **Separation of Concerns**: Clear component boundaries between Server, Translator, GoplsClient, AutoTranspiler, and FileWatcher.
- **Bidirectional Translation**: Source map-based position translation works correctly for completion, definition, and hover requests, using 0-based LSP coordinates properly.

### Core Functionality
- **Auto-Transpile**: File watcher debounces .dingo file changes and triggers transpilation with 500ms delay, automatically invalidating source map cache.
- **gopls Integration**: Robust subprocess management with restart logic (max 3 attempts) and proper cleanup on shutdown.
- **Source Map Caching**: Thread-safe caching with configurable size limit (100 entries) and version validation, preventing crashes from incompatible source map versions.

### Error Handling
- **Graceful Degradation**: Translation failures fall back to untranslated requests, ensuring LSP doesn't break entirely on position translation errors.
- **Robust Logging**: Debug/info/warn/error levels with structured output, including file change events and translation results.

### Test Coverage
- **Solid Unit Tests**: Translator tests cover both directions (Dingo→Go, Go→Dingo), range translation, and helper functions.
- **Integration Tests**: Phase 4 tests validate end-to-end functionality even though not LSP-specific, ensuring preprocessor+plugin pipeline works.

## ⚠️ Concerns

### ⚠️ CRITICAL (Must Fix)

1. **Incomplete Source Map Cache Invalidation** - `SourceMapCache.Invalidate()`
   - **Issue**: Method checks if key exists but doesn't actually delete from map (pkg/lsp/sourcemap_cache.go:119-127)
   - **Impact**: Stale source maps remain cached, causing incorrect position translations
   - **Recommendation**: Add `delete(c.maps, mapPath)` after existence check
   - **File/Line**: pkg/lsp/sourcemap_cache.go:125

2. **Unimplemented Diagnostic Publishing** - `handlePublishDiagnostics`
   - **Issue**: Diagnostics translation logic exists but actual publishing is placeholder ("TODO: Actually send notification to IDE connection") (pkg/lsp/handlers.go:312-316)
   - **Impact**: Compiler errors won't show in IDE, breaking core dev experience
   - **Recommendation**: Integration with jsonrpc2 conn to send publishDiagnostics notifications

3. **Unsafe Double-Check Pattern** - `SourceMapCache.Get()`
   - **Issue**: Double-check for concurrent loading races between unlock and lock (pkg/lsp/sourcemap_cache.go:44-62)
   - **Impact**: Multiple goroutines could load same source map, unnecessary work/reduntant I/O
   - **Recommendation**: Use sync.Once pattern within mutex scope

### ⚠️ IMPORTANT (Should Fix)

4. **No Dynamic Directory Watching** - `FileWatcher.watchRecursive()`
   - **Issue**: Only watches directories existing at initialization; new subdirectories created after don't get watched (pkg/lsp/watcher.go:59-80)
   - **Impact**: .dingo files in newly created directories won't trigger auto-transpile
   - **Recommendation**: Watch parent directories only and handle new dir creation events

5. **Debounce Timer Leakage** - `FileWatcher.handleFileChange()`
   - **Issue**: Rapid file changes can queue multiple timers without proper cancellation check (pkg/lsp/watcher.go:150-167)
   - **Impact**: Multiple timers could fire, calling OnFileChange() repeatedly for same changes
   - **Recommendation**: Store timer reference and check if active before creating new

6. **Missing Integration Tests** - LSP functionality
   - **Issue**: Only unit tests for translator; no integration tests for full LSP request/response cycle
   - **Impact**: Hard to verify bidirectional translation works end-to-end with gopls
   - **Recommendation**: Add integration tests using JSON-RPC mock server

7. **Weak Error Recovery** - GoplsClient failure scenarios
   - **Issue**: After max restart attempts exceeded, no fallback; requests will fail (pkg/lsp/gopls_client.go:216-227)
   - **Impact**: LSP becomes non-functional after gopls crashes repeatedly
   - **Recommendation**: Log error and continue operating with degraded features (no autocomplete)

### ⚠️ MINOR (Nice to Fix)

8. **Inconsistent File Path Handling** - Path extension functions
   - **Issue**: `dingoToGoPath` and `goToDingoPath` don't validate extensions, could process non-Dingo files (pkg/lsp/translator.go:133-145)
   - **Impact**: Minor; functions harmless but could be called incorrectly
   - **Recommendation**: Add early returns for non-matching extensions

9. **Single Event Debounce** - FileWatcher
   - **Issue**: Debounce applies to all pending files as group, but timer per batch (pkg/lsp/watcher.go:168-183)
   - **Impact**: Efficient but not perfectly fair if rapid individual changes
   - **Recommendation**: Document as design choice (coalesces multiple changes)

10. **Limited Forwarding Validation** - `forwardToGopls`
    - **Issue**: Unknown methods return generic "method not implemented" error without trying gopls (pkg/lsp/server.go:317-321)
    - **Impact**: Could break new LSP features gopls supports
    - **Recommendation**: Always attempt to forward to gopls for unknown methods

## 🔍 Questions

1. **Diagnostic Publishing**: What's the timeline for implementing IDE notifications? Core feature for dev experience.
2. **Performance Monitoring**: Are there plans to add telemetry for translation performance? 500ms debounce could be measured.
3. **gopls Features**: Which gopls features are intentionally not supported vs. future extensions?
4. **Version Mismatches**: How to handle source map version conflicts beyond fatal error?

## 📊 Summary

### Overall Assessment
**APPROVED WITH ISSUES** - Architecture is sound and core functionality works, but cache invalidation bug (CRITICAL) must be fixed before release to prevent translation errors. Diagnostic publishing (CRITICAL) needs completion for full IDE integration.

### Testability Score: Medium
- High in translator unit tests and helper functions (good coverage)
- Low in integration scenarios (missing LSP protocol testing)
- Medium overall due to Phase 4 end-to-end tests covering related functionality

### Priority Assessment
- **CRITICAL (Fix Immediately)**: Source map invalidation (data corruption), diagnostic publishing (main feature incomplete) - Block next phase
- **IMPORTANT (Fix Soon)**: Directory watching gaps, debounce issues - Affect reliability in complex workflows
- **MINOR (Fix Eventually)**: Code quality improvements - Polish items

### Code Quality Metrics
- **Simplicity**: ⭐⭐⭐⭐ Good component design, but some concurrency edge cases complex
- **Readability**: ⭐⭐⭐⭐ Clear naming, consistent patterns, good documentation
- **Maintainability**: ⭐⭐⭐ Good separation but needs integration tests for confidence
- **Go Best Practices**: ⭐⭐⭐ Good error handling, but some race conditions

LSP foundation provides solid base for IDE features. Major fixes needed for reliability, then can proceed to advanced language features.
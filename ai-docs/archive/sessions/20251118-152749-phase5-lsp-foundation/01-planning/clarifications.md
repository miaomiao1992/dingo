# Phase V Planning - User Clarifications

**Session:** 20251118-152749-phase5-lsp-foundation
**Date:** 2025-11-18

## Questions and Answers

### 1. Source Map Format Stability (Phase 4 Coordination)

**Question:** Will Phase 4 modify the source map format?

**User Answer:** "Could modify, but not critical. Let's assume no changes now, but be ready for changes."

**Implementation Decision:**
- Assume current source map format is stable for iteration 1
- Implement version checking in source map reader
- Add validation that fails gracefully if format changes
- Document source map format in code
- Add extensibility: if Phase 4 adds fields, LSP should handle gracefully

**Code Impact:**
```go
type SourceMap struct {
    Version    int    `json:"version"`     // Add version field
    DingoFile  string `json:"dingo_file"`
    GoFile     string `json:"go_file"`
    Mappings   []Mapping `json:"mappings"`
    // Future Phase 4 fields will be added here
}

// ValidateVersion ensures compatibility
func (sm *SourceMap) ValidateVersion() error {
    if sm.Version == 0 {
        sm.Version = 1 // Default for files without version
    }
    if sm.Version > MaxSupportedVersion {
        return fmt.Errorf("unsupported source map version %d (max: %d)", sm.Version, MaxSupportedVersion)
    }
    return nil
}
```

### 2. Transpilation Mode

**Question:** Should transpilation be automatic on save or require manual trigger?

**User Answer:** "Configurable (auto default)"

**Implementation Decision:**
- Default: Auto-transpile on save (`dingo.transpileOnSave: true`)
- Allow users to disable via VSCode settings
- Provide manual command: "Dingo: Transpile Current File"
- Provide workspace command: "Dingo: Transpile All Files"

**VSCode Settings:**
```json
{
  "dingo.transpileOnSave": true,
  "dingo.showTranspileNotifications": false  // Optional: reduce noise
}
```

**UX Flow:**
1. User saves .dingo file
2. If `transpileOnSave` enabled: auto-transpile in background
3. If transpilation fails: show diagnostic errors inline
4. If `transpileOnSave` disabled: no action (user runs manual command)

### 3. File Watching Strategy

**Question:** Should file watching monitor entire workspace or only opened files?

**User Answer:** "Hybrid - workspace .dingo only"

**Implementation Decision:**
- Watch entire workspace root recursively
- Filter for `.dingo` file extensions only (ignore .go, .txt, etc.)
- Use debouncing to avoid excessive transpilation on bulk changes
- Respect `.gitignore` and `.dingoignore` (if exists)

**Performance Optimization:**
```go
// Watch strategy
watcher.Add(workspaceRoot)
watcher.SetFilters([]string{"*.dingo"})  // Only .dingo files
watcher.SetDebounce(500 * time.Millisecond)  // Batch rapid changes

// Ignore patterns
watcher.Ignore([]string{
    "node_modules",
    "vendor",
    ".git",
    "**/.dingo_cache",  // Future: transpile cache
})
```

**Benefits:**
- Catches all .dingo changes (multi-file refactoring)
- Lighter than watching all files (only .dingo)
- Works with build tools that modify multiple files

## Additional Decisions (Based on Recommendations)

### 4. gopls Version Support
**Decision:** Support gopls v0.11+ (released ~2 years ago)
**Rationale:** Balances compatibility with maintenance burden

### 5. Error Reporting
**Decision:** Both diagnostics and notifications
- LSP diagnostics for Dingo syntax errors (inline)
- Notifications for system errors (dingo binary missing, etc.)

### 6. VSCode Extension Distribution
**Decision:** Start with .vsix for iteration 1
- Faster initial release
- Publish to marketplace in iteration 2 after stabilization

### 7. Workspace Scope
**Decision:** Workspace-only for iteration 1
- Simpler implementation
- Add arbitrary path support in iteration 2 if requested

### 8. Logging Level
**Decision:** Configurable via environment variable
```bash
DINGO_LSP_LOG=info   # Default (errors + warnings + key events)
DINGO_LSP_LOG=debug  # Full request/response logging
```

### 9. File Caching
**Decision:** Let gopls handle file caching
- LSP only caches source maps (lightweight)
- Simpler, avoids stale data issues

### 10. Missing .go Files
**Decision:** Auto-transpile silently on first LSP request
- Best UX (transparent)
- If transpilation fails, show error diagnostic

## Summary

**Key Principles:**
1. ✅ Graceful degradation (handle format changes)
2. ✅ User control (configurable auto-transpile)
3. ✅ Performance (hybrid file watching)
4. ✅ Clear errors (diagnostics + notifications)
5. ✅ Simple iteration 1 (workspace-only, .vsix distribution)

**Ready for Final Plan:** All major decisions made. Proceed to finalize implementation plan.

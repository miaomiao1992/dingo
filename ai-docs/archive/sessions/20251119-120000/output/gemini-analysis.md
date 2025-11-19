
[claudish] Model: minimax/minimax-m2



Based on my investigation of the current Dingo implementation and Go ecosystem patterns, here's my concrete recommendation:

## Recommendation: Adopt Go-Idiomatic Naming (Plain Constants)

**Change from**: `ValueTag_Int`, `Value_Int()`  
**Change to**: `Int`, `Value.Int()` (simple PascalCase constants)

### Rationale

The current underscore-separated pattern (`ValueTag_Int`, `Value_String`) violates Go naming conventions. Go standard library patterns use simple exported constants:
- `time.Monday`, `time.Tuesday`
- `net.FlagUp`, `net.FlagBroadcast`
- `syscall.O_RDONLY`

No redundant prefixes with type names.

### Trade-offs

**Gains**:
- ✅ Natural for Go developers (follows standard library)
- ✅ 50% less typing in pattern matching (`Int` vs `ValueTag_Int`)
- ✅ Passes Go linters without configuration
- ✅ Matches Go proposal #19412 community expectations

**Costs**:
- Package-level constants (not type-scoped) - minor API difference
- Breaking change for existing Dingo code - acceptable pre-1.0

### Migration Path

1. **Update enum generator** (`pkg/preprocessor/enum.go`):
   - Line 353-355: Remove `_` from `ValueTag_%s`
   - Line 380-407: Simplify constructor names

2. **Regenerate golden tests** (~15 test files with enum variants)

3. **Update pattern matching** (`rust_match.go`):
   - Search/replace `ValueTag_Int` → `Int`
   - Update constructor calls

### Examples

**Current (verbose)**:
```go
if result.tag == ValueTag_Ok {
    value := *result.ok_0
}
```

**Proposed (clean)**:
```go
if result.tag == Ok {
    value := *result.Value
}
```

The full analysis with code examples, alternatives, and implementation details is available at:
- `ai-docs/research/enum-naming-options.md` (detailed alternatives)
- `ai-docs/research/enum-naming-convention-analysis.md` (language precedents)

**Bottom line**: Use plain variant names like `Ok`, `Err`, `Circle`, `Square` at package level. This creates generated Go code that feels native to Go developers.

[claudish] Shutting down proxy server...
[claudish] Done


# Action Items - Phase 4.1 Code Review

**Date**: 2025-11-18
**Total Reviewers**: 4 (Internal, MiniMax M2, GPT-5.1 Codex, Polaris Alpha)
**Status**: CHANGES_NEEDED

---

## Critical Issues (Must Fix Before Merge)

### C1. Implement Complete Generator Integration
**Location**: `pkg/generator/generator.go:100+`
**Issue**: Generator doesn't wire Phase 4.1 components (config, parent map, type checker)
**Fix**: Add complete integration sequence:
```go
1. Load configuration via config.Load()
2. Run preprocessors with config
3. Parse with parser.ParseFile()
4. Build parent map via ctx.BuildParentMap(file)
5. Run type checker and populate ctx.TypeInfo
6. Run plugin pipeline
7. Check ctx.HasErrors() before code generation
8. Generate code with printer.Fprint()
```
**Mentioned by**: Internal (C4), Polaris (C1), Codex (implied)
**Estimated effort**: 2 days

---

### C2. Fix Preprocessor Marker Ordering
**Location**: `pkg/preprocessor/rust_match.go:260-272`
**Issue**: Marker generated AFTER temp var assignment, plugin can't find it
**Fix**: Emit `// DINGO_MATCH_START: scrutinee` BEFORE `__match_0 := scrutinee`
**Mentioned by**: Internal (C1), Polaris (C2)
**Estimated effort**: 1 hour

---

### C3. Enforce Exhaustiveness Errors
**Location**: `pkg/plugin/builtin/pattern_match.go:96-103`
**Issue**: Exhaustiveness errors logged but not propagated as compilation failures
**Fix**: Either:
- Option A: `Process()` returns error when exhaustiveness fails
- Option B: Generator checks `ctx.HasErrors()` (part of C1)
**Mentioned by**: Internal (C2), Polaris (C3)
**Estimated effort**: 2 hours (if using Option B from C1)

---

### C4. Fix Preprocessor Go Syntax Generation
**Location**: `pkg/preprocessor/rust_match.go:284, 352, 386`
**Issue**: Generates invalid `switch scrutinee.tag { case TagName: }` (missing comparison)
**Fix**: Use tagless switch with boolean conditions:
```go
switch {
case scrutinee.tag == TagName:
    // ...
}
```
**Mentioned by**: Internal (C5), Polaris (C4)
**Estimated effort**: 3 hours

---

### C5. Reset Plugin State Between Files
**Location**: `pkg/plugin/builtin/pattern_match.go:25-32, 60-103`
**Issue**: `matchExpressions` slice persists across files, causing stale AST bugs
**Fix**: Clear slice at start of `Process()`: `p.matchExpressions = p.matchExpressions[:0]`
**Mentioned by**: Codex (C1)
**Estimated effort**: 30 minutes

---

### C6. Fix Option Struct Representation Mismatch
**Location**: `tests/integration_phase4_test.go:248-279, 362-386`
**Issue**: Tests expect `isSet` field, but actual struct uses `tag + some_0`
**Fix**: Update tests to match Phase 3 implementation:
```go
// Change from:
if !strings.Contains(goCode, "Option_int{isSet: false}") {

// To:
if !strings.Contains(goCode, "Option_int{tag: OptionTagNone}") {
```
**Mentioned by**: Codex (C2)
**Estimated effort**: 1 hour

---

### C7. Fix None Inference Implementation
**Location**: `pkg/plugin/builtin/none_context.go:277-289, 63-72`
**Issue**: Two problems:
1. `ctx.TypeInfo` is nil (fixed by C1)
2. Uses wrong `go/types` map (Uses instead of Defs)
**Fix**: Change `typesInfo.Uses[lhsIdent]` to `typesInfo.Defs[lhsIdent]`
**Mentioned by**: Internal (C3), Polaris (C5), Codex (C4)
**Estimated effort**: 2 hours (after C1 complete)

---

### C8. Replace Regex Preprocessor with Proper Parser
**Location**: `pkg/preprocessor/rust_match.go:18-21`
**Issue**: Regex `(?s)match\s+([^{]+)\s*\{(.+)\}` fails on:
- Nested braces in scrutinee
- Multiple matches per line
- Braces in strings
**Fix**: Implement proper parser (participle or recursive descent)
**Mentioned by**: Internal (C1 variant), MiniMax (I5), Polaris (C6)
**Estimated effort**: 1 day

---

### C9. Fix Position-Based Comment Matching
**Location**: `pkg/plugin/builtin/pattern_match.go:228-258`
**Issue**: Magic number `distance < 100` is unreliable for finding pattern comments
**Fix**: Use AST structure instead of positions (iterate `file.Comments` within case bounds)
**Mentioned by**: Internal (C2 variant), Polaris (C7)
**Estimated effort**: 3 hours

---

## Important Issues (Should Fix Before Phase 4.2)

### I1. Fully Utilize go/types for Type Inference
**Location**: `pkg/plugin/builtin/pattern_match.go:342-395`
**Issue**: Uses string heuristics (`strings.Contains("Result")`) instead of go/types
**Fix**: Extract type from `typesInfo.Types[scrutineeExpr]` and match against known Dingo types
**Mentioned by**: All 4 reviewers
**Estimated effort**: 1 day

---

### I2. Implement Custom Enum Exhaustiveness
**Location**: `pkg/plugin/builtin/pattern_match.go:340-395`
**Issue**: Only checks Result/Option, not user-defined enums
**Fix**: Query enum metadata from preprocessor or inspect type definition
**Mentioned by**: Codex (C3), Internal (implied)
**Estimated effort**: 1 day

---

### I3. Evaluate Marker Strategy for Phase 4.2
**Location**: Architecture (preprocessor → plugin communication)
**Issue**: Marker comments may not scale for guards, nested patterns, tuples
**Fix**: Plan hybrid approach (markers for simple, AST metadata for complex)
**Mentioned by**: MiniMax (I1)
**Estimated effort**: 1 day (planning/design)

---

### I4. Add Parent Map Cleanup
**Location**: `pkg/plugin/context.go`, `pkg/generator/generator.go`
**Issue**: Parent map built but never freed (memory leak)
**Fix**: Add `ctx.Cleanup()` defer in generator, clear ParentMap and TypeInfo
**Mentioned by**: MiniMax (I3), Polaris (M5)
**Estimated effort**: 1 hour

---

### I5. Enhance Error Messages with Context
**Location**: `pkg/plugin/builtin/pattern_match.go:398-406`, `none_context.go:121-126`
**Issue**: Generic errors, no file/line/source excerpt, no specific hints
**Fix**: Implement enhanced error formatting with:
- File name and line number
- Source excerpt
- Type context
- Specific suggestions (not generic)
**Mentioned by**: Internal (I5), MiniMax (M7), Polaris (I3)
**Estimated effort**: 1 day

---

### I6. Add Configuration Cross-Field Validation
**Location**: `pkg/config/config.go:247-342`
**Issue**: No validation for interdependent settings
**Fix**: Add checks:
- match.syntax requires result_type or option_type enabled
- Reject syntax="swift" (not implemented)
- Warn if nil_safety_checks="debug" but source_maps disabled
**Mentioned by**: Internal (I1), Polaris (I2)
**Estimated effort**: 2 hours

---

## Notes

### Severity Summary
- **CRITICAL**: 9 issues (C1-C9) - Block all functionality
- **IMPORTANT**: 6 issues (I1-I6) - Quality and completeness
- **MINOR**: 5 issues (M1-M5) - Documentation and polish

### Estimated Timeline
- **CRITICAL fixes**: 5-7 days → Functional Phase 4.1
- **IMPORTANT fixes**: 4-5 days → Production-ready Phase 4.1
- **Total**: 9-12 days to completion

### Reviewer Agreement
- **C1 (Generator)**: 4/4 reviewers (highest consensus)
- **C7 (None inference)**: 3/4 reviewers
- **C8 (Regex parser)**: 3/4 reviewers
- **I1 (go/types)**: 4/4 reviewers

### Priority Order
Focus on critical issues in dependency order:
1. **C1** (Generator integration) - Enables C3, C7
2. **C2, C4, C5** (Preprocessor fixes) - Can be done in parallel with C1
3. **C6** (Test fixes) - Quick win
4. **C8** (Regex parser) - Separate refactor
5. **C9** (Comment matching) - After C8
6. **C3, C7** (Enforcement, inference) - After C1 complete
7. **I1-I6** (Important issues) - After all critical fixed

### Test Verification
After each fix, verify:
- Unit tests: Should remain 100% passing
- Integration tests: Should reach 100% (currently broken)
- Golden tests: Should be enabled and passing

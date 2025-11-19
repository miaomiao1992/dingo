# Golden Test Analysis

## Test Run Summary

**Total Tests**: 46 golden tests
**Failed Tests**: Most tests are failing
**Skipped Tests**: 1 (error_prop_02_multiple - marked as "Parser bug")

## Common Failure Patterns

### Pattern 1: Import Statement Formatting Mismatch
**Severity**: MINOR (cosmetic)
**Frequency**: ALL failing tests

**Expected**:
```go
import "os"
```

**Actual**:
```go
import (
    "os"
)
```

The transpiler is generating multi-line import blocks even for single imports, while golden files expect single-line imports.

### Pattern 2: Missing `// dingo:s:N` Comments
**Severity**: CRITICAL
**Frequency**: ALL failing tests with `?` operator

**Expected**:
```go
__tmp0, __err0 := os.ReadFile(path)
// dingo:s:1
if __err0 != nil {
    return nil, __err0
}
// dingo:e:1
```

**Actual**:
```go
__tmp0, __err0 := os.ReadFile(path)

if __err0 != nil {
    return nil, __err0
}
// dingo:e:1
```

The transpiler is NOT generating `// dingo:s:1` (start) comments, only generating `// dingo:e:1` (end) comments. This breaks source map tracking.

### Pattern 3: Duplicate `// dingo:s:N` Comments in Golden Files
**Severity**: CRITICAL (golden file bug)
**Frequency**: Many tests

Some golden files have DUPLICATE `// dingo:s:1` comments which is clearly wrong:
```go
__tmp0, __err0 := ReadFile(path)
// dingo:s:1
// dingo:s:1  <-- DUPLICATE!
if __err0 != nil {
```

### Pattern 4: Unqualified Function References in Golden Files
**Severity**: CRITICAL (golden file bug)
**Frequency**: Many tests

Golden files use unqualified function calls like `ReadFile()`, `Atoi()`, `Unmarshal()` instead of qualified ones like `os.ReadFile()`, `strconv.Atoi()`, `json.Unmarshal()`.

**Expected (in golden)**:
```go
__tmp0, __err0 := ReadFile(path)  // WRONG - won't compile
```

**Actual (transpiler output)**:
```go
__tmp0, __err0 := os.ReadFile(path)  // CORRECT - will compile
```

## Root Causes

### Issue 1: Error Propagation Processor Not Generating Start Comments
The error propagation processor is supposed to generate both start and end comments for source map tracking, but it's only generating end comments.

**Location**: Likely in `pkg/generator/preprocessor/error_prop.go`

### Issue 2: Golden Files Are Outdated/Incorrect
Many golden files contain:
1. Duplicate source map comments (`// dingo:s:1` appears twice)
2. Unqualified function calls (won't compile as valid Go)
3. Missing import statements

**Recommendation**: Golden files need to be regenerated from actual transpiler output OR transpiler needs to match the intended output (but current transpiler output looks MORE correct - it's compilable Go).

### Issue 3: Import Formatting Inconsistency
The transpiler always generates multi-line import blocks. Golden files expect single-line for single imports.

**Location**: Likely in AST printer or import consolidation logic.

## Decision Required

**CRITICAL QUESTION**: Which is the "source of truth"?

**Option A: Golden files are correct**
- Fix transpiler to match golden files
- But golden files have uncompilable Go (unqualified calls)
- Requires adding duplicate `// dingo:s:N` comments (seems wrong)

**Option B: Transpiler output is correct**
- Regenerate all golden files from current transpiler
- Update tests to use actual compilable Go
- Remove duplicate comments from source map annotations

**Recommendation**: **Option B** - The transpiler output is MORE correct because:
1. It generates compilable Go code (`os.ReadFile` vs `ReadFile`)
2. It doesn't duplicate source map comments
3. Import resolution works correctly

## Next Steps

1. **Investigate**: Use golang-architect agents with external models to analyze:
   - What is the intended behavior for `// dingo:s:N` comments?
   - Should golden files contain compilable Go or template-like code?
   - Is import formatting important or cosmetic?

2. **Fix Strategy** (pending investigation):
   - If golden files are wrong → Regenerate all golden files
   - If transpiler is wrong → Fix error propagation processor

3. **Validation**:
   - Ensure all generated `.go` files compile successfully
   - Ensure source maps work correctly for LSP

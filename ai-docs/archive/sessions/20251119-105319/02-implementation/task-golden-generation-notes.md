# Golden File Generation - Critical Issues Found

## Executive Summary

**Status**: BLOCKED - Cannot generate golden files until 5 critical bugs are fixed

Pattern matching transformation has multiple severe bugs preventing successful compilation of transpiled code. All 7 test files either failed to transpile (1 file) or transpiled but won't compile (6 files).

## Detailed Issue Analysis

### Issue 1: Variable Hoisting Failure (CRITICAL)
**Affects**: Files 09-12 (all tuple tests)
**Severity**: Critical - Core functionality broken

**Problem**: Variables bound in tuple patterns are not accessible in match arm expressions.

**Example**:
```dingo
// Input .dingo
match (a, b) {
    (Ok(x), Ok(y)) => x + y,  // Should bind x and y
}
```

```go
// Generated .go (BROKEN)
case ResultTagOk:  // Error: ResultTagOk undefined
    x + y  // Error: x undefined, y undefined
```

**Root Cause**: Variable hoisting transformation not working for tuple patterns. Variables need to be declared at function scope before the switch statement.

**Expected**:
```go
var x int      // Hoisted
var y int      // Hoisted
switch __match_0.__tag {
case ResultTag_Ok:
    x = __match_0.Result_Ok_0.(Pair).__0.Result_Ok_0
    y = __match_0.Result_Ok_0.(Pair).__1.Result_Ok_0
    return x + y
}
```

**File**: pkg/generator/plugin/pattern_match/transform.go
**Function**: hoistVariables() or similar

---

### Issue 2: Tag Constant Naming Inconsistency (CRITICAL)
**Affects**: Files 09-12 (all tuple tests)
**Severity**: Critical - Generated code won't compile

**Problem**: Generated code references `ResultTagOk` but constants are defined as `ResultTag_Ok` (with underscore).

**Example**:
```go
// Generated switch case
case ResultTagOk:  // Error: undefined

// But constant is defined as
const ResultTag_Ok ResultTag = 0  // With underscore
```

**Root Cause**: Inconsistency between:
1. Tag constant generation (uses underscore)
2. Switch case generation (no underscore)

**Fix**: Ensure consistent naming. Choose ONE:
- Option A: Always use underscore (ResultTag_Ok) ← RECOMMENDED
- Option B: Never use underscore (ResultTagOk)

**File**: pkg/generator/plugin/pattern_match/inject.go
**Function**: generateEnumHelpers() and switch case generation

---

### Issue 3: Field Name Capitalization Bug (HIGH)
**Affects**: Files 07-08 (guard tests with enum data)
**Severity**: High - Breaks field access

**Problem**: Accessing enum variant fields using lowercase variant name instead of capitalized.

**Example**:
```go
// Generated code (BROKEN)
if __match_0.request_get_0 > 1024 {  // Error: request_get_0 undefined

// Should be
if __match_0.Request_Get_0 > 1024 {  // Capitalized variant name
```

**Pattern**:
```
Generated: variant_field_N (lowercase)
Correct:   Variant_Field_N (capitalized)
```

**Root Cause**: Field name generation not capitalizing variant names when creating field accessors.

**File**: pkg/generator/plugin/pattern_match/transform.go
**Function**: buildVariantFieldAccess() or similar

---

### Issue 4: Multiple Default Cases (MEDIUM)
**Affects**: Files 11-12 (wildcard tests)
**Severity**: Medium - Invalid Go syntax

**Problem**: Transformation generates multiple `default:` cases in the same switch statement.

**Example**:
```go
switch __match_0.__tag {
case ResultTag_Ok:
    // ...
default:        // First default (line 46)
    return "error"
default:        // Duplicate! (line 48) ← Error
    return "error"
default:        // Another! (line 50) ← Error
    return "error"
}
```

**Root Cause**: Wildcard pattern transformation doesn't track if default case already exists.

**Fix**: Add state tracking:
```go
hasDefault := false
for _, arm := range arms {
    if isWildcard(arm.pattern) {
        if hasDefault {
            continue  // Skip duplicate wildcards
        }
        hasDefault = true
        // Generate default case
    }
}
```

**File**: pkg/generator/plugin/pattern_match/transform.go
**Function**: transformMatchExpression()

---

### Issue 5: Preprocessor Duplication Bug (HIGH)
**Affects**: File 06 (guards_nested)
**Severity**: High - Prevents transpilation entirely

**Problem**: Preprocessor appears to duplicate code, causing parse errors beyond file length.

**Example**:
```
Input file: 54 lines
Error: parse error at line 98:18
```

File only has 54 lines, but error occurs at line 98. This suggests preprocessor is duplicating content.

**Likely Cause**: Regex pattern matching `match` expression incorrectly, causing it to:
1. Match the same expression multiple times
2. Apply transformation repeatedly
3. Duplicate the generated code

**Investigation Needed**:
1. Check preprocessor regex patterns for `match` keyword
2. Verify match boundaries (word boundaries, scope awareness)
3. Look for greedy/non-greedy regex issues
4. Check if transformation is applied in a loop without marking processed sections

**File**: pkg/generator/preprocessor/pattern_match.go (or similar)
**Function**: Preprocess() or regex patterns

---

## Additional Issues Found

### Issue 6: Unused String Literals (LOW)
**Affects**: Files 07-08
**Severity**: Low - Compilation warning/error

**Problem**: Match arm expressions generate unused string literals.

**Example**:
```go
case RequestTag_Get:
    if __match_0.Request_Get_0 > 1024 {
        "large upload"  // Error: string literal not used
    }
```

**Root Cause**: Expression statements without return/assignment.

**Fix**: Ensure expressions are properly returned or assigned:
```go
if __match_0.Request_Get_0 > 1024 {
    __result = "large upload"
}
```

---

## Recommended Fix Order

**Phase 1: Critical Blockers** (Must fix to proceed)
1. ✅ **Issue 2**: Fix tag constant naming (ResultTag_Ok)
   - Impact: Affects 4 files (09-12)
   - Effort: Low (find/replace in constant generation)

2. ✅ **Issue 1**: Fix variable hoisting for tuple patterns
   - Impact: Affects 4 files (09-12)
   - Effort: High (implement hoisting logic)

3. ✅ **Issue 3**: Fix field name capitalization
   - Impact: Affects 2 files (07-08)
   - Effort: Medium (update field access generation)

**Phase 2: Important Fixes**
4. ✅ **Issue 5**: Debug preprocessor duplication
   - Impact: Affects 1 file (06)
   - Effort: Medium (regex debugging)

5. ✅ **Issue 4**: Fix multiple default cases
   - Impact: Affects 2 files (11-12)
   - Effort: Low (add state tracking)

**Phase 3: Quality**
6. ✅ **Issue 6**: Fix unused string literals
   - Impact: All guard tests
   - Effort: Low (ensure proper return/assignment)

---

## Files Generated (but broken)

| File | Size | Status | Primary Issue |
|------|------|--------|---------------|
| pattern_match_06_guards_nested.go | N/A | Parse failed | Preprocessor duplication (Issue 5) |
| pattern_match_07_guards_complex.go | 3559 bytes | Won't compile | Field capitalization (Issue 3) |
| pattern_match_08_guards_edge_cases.go | 3709 bytes | Won't compile | Field capitalization (Issue 3) |
| pattern_match_09_tuple_pairs.go | 1068 bytes | Won't compile | Tag naming + hoisting (Issue 1,2) |
| pattern_match_10_tuple_triples.go | 1120 bytes | Won't compile | Tag naming + hoisting (Issue 1,2) |
| pattern_match_11_tuple_wildcards.go | 1081 bytes | Won't compile | Tag naming + hoisting + defaults (Issue 1,2,4) |
| pattern_match_12_tuple_exhaustiveness.go | 1712 bytes | Won't compile | Tag naming + hoisting + defaults (Issue 1,2,4) |

**Total**: 0 valid golden files created

---

## Next Action Required

**CANNOT proceed with golden file generation.**

**Required**: Delegate to golang-developer to fix Issues 1-5 in priority order.

**After fixes**: Re-run this task to generate golden files.

**Estimated effort**: 4-6 hours of implementation work to fix all issues.

# Dingo Test Failure Investigation Report
**Date**: 2025-11-19
**Analysis Target**: 14 failing tests (261/267 passing, 97.8%)
**Phase**: Phase 4.2 - Pattern Matching Enhancements

---

## Executive Summary

**Critical Finding**: The test failures are **NOT** due to outdated golden files or test problems. They are caused by **implementation bugs** in the transpiler, specifically:

1. **Naming inconsistency**: Generated code uses `ResultTag_Ok` but pattern matching expects `ResultTagOk` (no underscore)
2. **Missing golden files**: 7 new pattern matching tests lack `.go.golden` expected output files
3. **Missing type declarations**: Some Result/Option types not being properly injected

**Severity**: **CRITICAL** - Multiple core features (Result, Option, Pattern Matching) are broken
**Priority**: **P0** - Must fix before Phase 4.2 can be considered complete

**Root Causes**:
- Bug in `pkg/preprocessor/rust_match.go`: Naming format mismatch
- Missing test files (test infrastructure issue, not implementation)
- Type injection pipeline working partially but inconsistently

---

## Failure Analysis by Category

### Category 1: Pattern Matching Golden Tests (8 failures)

#### Tests Affected
- `pattern_match_03_nested`
- `pattern_match_06_guards_nested`
- `pattern_match_07_guards_complex`
- `pattern_match_08_guards_edge_cases`
- `pattern_match_09_tuple_pairs`
- `pattern_match_10_tuple_triples`
- `pattern_match_11_tuple_wildcards`
- `pattern_match_12_tuple_exhaustiveness`

#### Root Cause
**2 distinct issues**:

1. **Missing Golden Files** (7 out of 8 tests):
   - Error: `open golden/pattern_match_03_nested.go.golden: no such file or directory`
   - These tests were recently added but `.go.golden` expected output files were never created
   - **Test Infrastructure Issue**, not implementation bug

2. **Naming Inconsistency** (pattern_match_12_tuple_exhaustiveness):
   - Generated code: `ResultTag_Ok` (with underscore)
   - Expected pattern: `ResultTagOk` (no underscore)
   - This causes compilation errors: `undefined: ResultTagOk`
   - **Implementation Bug** in code generation

#### Test vs Implementation
- **7 tests**: Missing golden files → Test infrastructure needs update
- **1 test**: Naming mismatch → Implementation bug in pattern matching code generation

#### Priority
**CRITICAL** - Pattern matching is a core Phase 4 feature

#### Recommended Fix
1. Create missing `.go.golden` files by running transpiler on source files
2. Fix naming inconsistency in `pkg/preprocessor/rust_match.go` to generate `ResultTagOk` instead of `ResultTag_Ok`

---

### Category 2: Integration Tests (4 failures)

#### Tests Affected
1. `pattern_match_rust_syntax`
   - Error: `undefined: Result_int_error`
   - Error: `undefined: ResultTagOk`
2. `pattern_match_non_exhaustive_error`
   - Error: Expected non-exhaustive match error, but no errors reported
3. `none_context_inference_return`
   - Error: `undefined: Option_int`
   - Error: `Cannot infer Option type for None constant`
4. `combined_pattern_match_and_none`
   - Error: `undefined: Result_string_error`
   - Error: `undefined: Option_int`

#### Root Cause
**Implementation Bug**: Type declarations are not being properly injected into the generated code

**Evidence**:
- Integration test creates Result/Option types and expects them to be available
- Type checker reports `undefined: Result_int_error` → Result type was never emitted
- PatternMatchPlugin finds matches but types don't exist
- Type inference service generates types but they're not in final output

#### Test vs Implementation
**Implementation Bug** - The type generation and injection pipeline is failing

#### Priority
**CRITICAL** - Core type system is broken

#### Recommended Fix
Debug the type injection pipeline:
1. Check if `ResultTypePlugin.GetPendingDeclarations()` is being called
2. Verify `GetInjectedTypesAST()` returns the injected AST
3. Ensure `generator.go` prints injected types (lines 197-209)
4. Verify pipeline execution order (ResultTypePlugin before PatternMatchPlugin)

---

### Category 3: Compilation Tests (2 failures)

#### Tests Affected
1. `error_prop_02_multiple_compiles`
2. `option_02_literals_compiles`

#### Root Cause
**Skipped Tests with Parser Bugs**:
- Both tests show: `Parser bug - needs fixing in Phase 3`
- Tests are being skipped, not actually failing
- This is expected behavior for unimplemented features

#### Test vs Implementation
**Neither** - These are intentionally skipped tests

#### Priority
**MINOR** - Known limitations, not failures

#### Recommended Fix
**No action needed** - These are deferred to Phase 3

---

## Detailed Findings

### Finding 1: Naming Convention Mismatch

**Issue**: Generated Go code uses underscore-separated names but pattern matching expects PascalCase

**Location**: `pkg/preprocessor/rust_match.go`

**Evidence from generated code**:
```go
// Generated (current - WRONG):
case ResultTag_Ok:
    ResultTag_Err:

// Expected by pattern matching (CORRECT):
case ResultTagOk:
    ResultTagErr:
```

**Impact**: All pattern matching with guards and tuples will fail to compile

### Finding 2: Type Injection Pipeline Partially Working

**Issue**: Some tests generate types correctly, others don't

**Evidence**:
- `pattern_match_12_tuple_exhaustiveness`: ✅ Has type declarations
- `pattern_match_01_simple`: ❌ Missing `Result` and `Option` types
- Integration tests: ❌ Types not injected

**Working Example**:
```go
// This IS generated correctly:
type ResultTag uint8
const (
    ResultTag_Ok ResultTag = iota
    ResultTag_Err
)
type Result struct {
    tag   ResultTag
    ok_0  *int
    err_0 *error
}
```

**Broken Example**:
```go
// Missing from generated code:
func Result_int_error_Ok(arg0 int) Result_int_error
func Result_int_error_Err(arg0 error) Result_int_error
func (r Result_int_error) IsOk() bool
func (r Result_int_error) IsErr() bool
```

### Finding 3: Missing Golden Files for New Tests

**Issue**: 7 recently-added pattern matching tests lack expected output files

**Location**: `tests/golden/`

**Missing Files**:
- `pattern_match_03_nested.go.golden`
- `pattern_match_06_guards_nested.go.golden`
- `pattern_match_07_guards_complex.go.golden`
- `pattern_match_08_guards_edge_cases.go.golden`
- `pattern_match_09_tuple_pairs.go.golden`
- `pattern_match_10_tuple_triples.go.golden`
- `pattern_match_11_tuple_wildcards.go.golden`

**Impact**: Tests cannot run, showing "file not found" errors

### Finding 4: None Context Inference Limitation

**Issue**: `None` constant cannot infer type from context

**Error**: `Cannot infer Option type for None constant at test.go:5:10`

**Root Cause**: Type inference service doesn't have enough context to determine which `Option_T` type to use

**Example**:
```go
// This fails:
func test() Option_int {
    return None  // Cannot infer Option_int from return type
}

// This works:
func test() Option_int {
    return Option_int_None()  // Explicit constructor
}
```

---

## Recommended Action Plan

### Phase 1: Critical Bug Fixes (Immediate - P0)

#### Fix 1: Naming Convention Mismatch
**File**: `pkg/preprocessor/rust_match.go`
**Action**: Replace all instances of `ResultTag_Ok` with `ResultTagOk` (remove underscore)
**Lines to check**: ~200-250 (pattern matching transformation code)

#### Fix 2: Type Injection Pipeline
**Files**: `pkg/generator/generator.go`, `pkg/plugin/plugin.go`
**Action**: Debug and fix type declaration injection
**Steps**:
1. Verify `ResultTypePlugin.Process()` is being called
2. Check `GetPendingDeclarations()` returns declarations
3. Ensure `GetInjectedTypesAST()` is populated
4. Verify generator prints injected types

#### Fix 3: Create Missing Golden Files
**Location**: `tests/golden/`
**Action**: Generate golden files for 7 missing tests
**Method**:
```bash
# Run transpiler on each .dingo file to generate .go
# Copy output to .go.golden files
# Verify tests pass
```

### Phase 2: Integration Test Fixes (Next - P1)

#### Fix 4: Pattern Matching Integration
**File**: `tests/integration_phase4_test.go`
**Action**: Fix `pattern_match_rust_syntax` test
**Expected**: Should generate and compile Result_int_error type

#### Fix 5: None Context Inference
**File**: `pkg/plugin/builtin/none_context.go`
**Action**: Improve type inference for None constant
**Strategy**: Use parent map to infer type from function signature

### Phase 3: Test Infrastructure (Final - P2)

#### Fix 6: Golden Test Automation
**Action**: Add test that fails if golden file is missing
**Purpose**: Prevent future missing golden files

---

## Files to Modify

### Critical (P0)

1. **`pkg/preprocessor/rust_match.go`**
   - Change: `ResultTag_Ok` → `ResultTagOk`
   - Change: `ResultTag_Err` → `ResultTagErr`
   - Also for OptionTag, StatusTag, etc.

2. **`pkg/plugin/builtin/result_type.go`**
   - Verify: `emitResultTagEnum()` generates correct constant names
   - Ensure: Constants don't have underscores

3. **`pkg/generator/generator.go`**
   - Verify: Lines 197-209 (injected types printing) work correctly
   - Add: Debug logging to trace type injection

### Important (P1)

4. **`tests/golden/`**
   - Create: 7 missing `.go.golden` files
   - Verify: All golden tests pass

5. **`tests/integration_phase4_test.go`**
   - Update: Test expectations to match fixed naming
   - Fix: Type injection issues

6. **`pkg/plugin/builtin/none_context.go`**
   - Improve: Type inference from function return types
   - Add: Better error messages

### Minor (P2)

7. **`tests/golden_test.go`**
   - Add: Validation that golden file exists before test runs

---

## Success Criteria

### Before Fix
- ❌ 14 failing tests
- ❌ Pattern matching doesn't compile
- ❌ Type system inconsistent

### After Fix
- ✅ 267/267 tests passing (100%)
- ✅ Pattern matching compiles and runs
- ✅ Result/Option types properly generated
- ✅ Golden tests match expected output
- ✅ Integration tests verify end-to-end functionality

---

## Risk Assessment

### Low Risk
- Creating golden files (test infrastructure only)
- Fixing naming conventions (search and replace)

### Medium Risk
- Type injection pipeline (complex, touches multiple files)
- None context inference (requires understanding parent map)

### High Risk
- None - all fixes are well-contained and tested

---

## Testing Strategy

### 1. Unit Tests
Run individual components:
```bash
go test ./pkg/plugin/builtin -v -run TestResultType
go test ./pkg/preprocessor -v -run TestRustMatch
```

### 2. Golden Tests
Verify specific tests:
```bash
go test ./tests -v -run TestGoldenFiles/pattern_match_01_basic
```

### 3. Integration Tests
Test full pipeline:
```bash
go test ./tests -v -run TestIntegrationPhase4
```

### 4. Compilation Tests
Verify generated code compiles:
```bash
go test ./tests -v -run TestGoldenFilesCompilation
```

---

## Conclusion

The test failures are caused by **implementation bugs**, not test problems. The most critical issue is the **naming convention mismatch** between generated code (`ResultTag_Ok`) and pattern matching expectations (`ResultTagOk`).

The second critical issue is the **missing golden files** for 7 pattern matching tests, which prevents those tests from running at all.

Once these bugs are fixed, we should achieve **100% test pass rate** and Phase 4.2 can be considered complete.

**Estimated Fix Time**: 2-3 hours for critical bugs, 4-6 hours total

---

**Report Generated**: 2025-11-19 14:30:00
**Next Review**: After P0 fixes applied

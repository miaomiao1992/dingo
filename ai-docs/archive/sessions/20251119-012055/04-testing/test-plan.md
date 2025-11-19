# Test Plan: Phase 4 Priority 2 & 3

**Session**: 20251119-012055
**Date**: 2025-11-19
**Scope**: Comprehensive testing for all 4 implemented tasks

---

## Overview

This test plan validates the implementation of:
1. **Task 1**: 4 context type helpers (foundation)
2. **Task 2**: Pattern match scrutinee go/types integration
3. **Task 3**: Err() context-based type inference
4. **Task 4**: Guard validation with outer scope support

---

## Test Categories

### 1. Unit Tests (Isolated Component Testing)

#### Task 1: Context Type Helpers
**Location**: `pkg/plugin/builtin/type_inference_test.go`

**Test Functions**:
- `TestFindFunctionReturnType` - 5 scenarios
  - Simple int return
  - Option type return
  - Result type return
  - Lambda return
  - No return type (edge case)

- `TestFindAssignmentType` - 4 scenarios
  - Simple assignment
  - Parallel assignment
  - Option type assignment
  - Result type assignment

- `TestFindVarDeclType` - 4 scenarios
  - Explicit type annotation
  - Option type explicit
  - Result type explicit
  - Multi-var declaration

- `TestFindCallArgType` - 4 scenarios
  - Regular function call
  - Option type parameter
  - Result type parameter
  - Multiple parameters

- `TestContainsNode` - 1 scenario
  - Verify AST node containment checking

- `TestStrictGoTypesRequirement` - 1 scenario
  - Verify nil go/types.Info handling

**Total Task 1 Tests**: 31 tests

#### Task 2: Pattern Match Scrutinee
**Location**: `pkg/plugin/builtin/pattern_match_test.go`

**Expected Tests** (based on implementation):
- Pattern matching with type aliases
- go/types integration for scrutinee type detection
- Fallback to heuristics when go/types unavailable

**Coverage**: Type alias handling, complex expressions, function returns

#### Task 3: Err() Context Inference
**Location**: `pkg/plugin/builtin/result_type_test.go`

**Expected Tests**: 7 scenarios
- Return context inference
- Assignment context inference
- Call argument context inference
- Struct field context inference
- Error when context unavailable
- Nested function contexts
- Multi-value returns

**Note**: 3/7 tests passing expected (requires full pipeline integration)

#### Task 4: Guard Validation
**Location**: `pkg/plugin/builtin/pattern_match_test.go`

**Tests**:
- Line 826: `TestPatternMatchPlugin_TransformWithGuard`
  - Valid guards with pattern-bound variables
  - Guard expression preservation in output

- Line 1009: `TestPatternMatchPlugin_GuardWithOuterScope`
  - Guards referencing outer scope variables
  - Boolean type validation
  - Invalid guard compile errors

---

### 2. Integration Tests

**Location**: `tests/integration_phase4_test.go` (if exists)

**Purpose**: Verify all 4 tasks work together in full transpilation pipeline

**Test Scenarios** (hypothetical):
- None inference in function return → Pattern match → Guard validation
- Err() with context inference → Pattern match with type alias
- Full pipeline with all features enabled

**Command**:
```bash
go test ./tests -v -run TestIntegrationPhase4
```

---

### 3. Golden Tests

**Purpose**: End-to-end validation with real Dingo code

**Relevant Tests** (based on plan):

1. **None Inference Comprehensive** (Task 1)
   - File: `tests/golden/none_inference_comprehensive.dingo`
   - Status: May not exist yet (planned)
   - Tests: All 9 None contexts (function return, assignment, var decl, call arg, etc.)

2. **Pattern Match Type Alias** (Task 2)
   - File: `tests/golden/pattern_match_type_alias.dingo`
   - Status: May not exist yet (planned)
   - Tests: Pattern matching with `type MyResult = Result_int_error`

3. **Result Err Contexts** (Task 3)
   - File: `tests/golden/result_err_contexts.dingo`
   - Status: May not exist yet (planned)
   - Tests: Err() in return/assignment/call/struct contexts

4. **Pattern Guards Complete** (Task 4)
   - File: `tests/golden/pattern_guards_complete.dingo`
   - Status: May not exist yet (planned)
   - Tests: Guards with pattern vars and outer scope refs

**Command**:
```bash
go test ./tests -v -run TestGoldenFiles
```

---

## Test Execution Plan

### Phase 1: Unit Tests (Task 1 Foundation)
**Priority**: CRITICAL
**Expected**: 31/31 passing (100%)

```bash
go test ./pkg/plugin/builtin -v -run "TestFind.*|TestContainsNode|TestStrictGoTypesRequirement"
```

**Success Criteria**:
- All 31 tests pass
- No nil pointer panics
- go/types.Info integration works
- Parent map traversal correct

---

### Phase 2: Unit Tests (Tasks 2-4)
**Priority**: HIGH
**Expected**: Varies (some may require pipeline integration)

```bash
# Task 2: Pattern match
go test ./pkg/plugin/builtin -v -run "TestPatternMatch.*"

# Task 3: Result type
go test ./pkg/plugin/builtin -v -run "TestResult.*"

# Task 4: Guards (specific test lines)
go test ./pkg/plugin/builtin -v -run "TestPatternMatchPlugin_TransformWithGuard|TestPatternMatchPlugin_GuardWithOuterScope"
```

**Success Criteria**:
- Task 2: Type alias tests pass
- Task 3: 3/7 tests pass (expected - needs full pipeline)
- Task 4: Guard tests pass (TODOs removed)

---

### Phase 3: Integration Tests
**Priority**: MEDIUM
**Expected**: May not exist yet

```bash
go test ./tests -v -run TestIntegrationPhase4
```

**Fallback**: If no dedicated integration test, run full test suite:
```bash
go test ./pkg/plugin/builtin -v
```

**Success Criteria**:
- No regressions in existing tests
- Build compiles cleanly
- Pre-existing failures documented (not caused by changes)

---

### Phase 4: Golden Tests
**Priority**: HIGH (validates end-to-end)
**Expected**: New tests may not exist yet (planned)

```bash
go test ./tests -v -run TestGoldenFiles
```

**Check for**:
- Pattern match golden tests (existing)
- None inference tests (existing)
- Result type tests (existing)
- Guard tests (may be new)

**Success Criteria**:
- All existing golden tests still pass
- New tests (if created) pass
- Generated Go code is idiomatic
- No `Result_interface_error` in output (Task 3 validation)

---

## Success Metrics Validation

### Metric 1: None Inference Coverage
**Target**: 50% → 90%+

**Test Method**:
1. Run unit tests for all 4 context helpers
2. Check golden tests for None usage in various contexts
3. Verify no fallback to `interface{}` type

**Validation**:
```bash
# Check for interface{} fallbacks in generated golden files
grep -r "interface{}" tests/golden/*.go.golden | grep -v "// Expected"
```

**Success**: No unexpected `interface{}` fallbacks

---

### Metric 2: Pattern Match Accuracy
**Target**: 85% → 95%+

**Test Method**:
1. Run pattern match unit tests
2. Test with type aliases
3. Test with function returns
4. Test with complex expressions

**Validation**:
```bash
go test ./pkg/plugin/builtin -v -run "TestPatternMatch.*"
```

**Success**: All type alias and go/types tests pass

---

### Metric 3: Err() Type Correctness
**Target**: 0% → 80%+

**Test Method**:
1. Run Err() context inference tests
2. Check generated code for correct Result types
3. Verify no `Result_interface_error` in output

**Validation**:
```bash
# Check Result_test.go for Err() tests
go test ./pkg/plugin/builtin -v -run "TestErrContext.*"

# Verify no interface{} in Result types
grep "Result_interface_" tests/golden/result_*.go.golden
```

**Success**: No `Result_interface_` found, 80%+ tests pass

---

### Metric 4: Guard Test Pass Rate
**Target**: 0% (2 TODOs) → 100% (2 passing)

**Test Method**:
1. Verify TODOs removed from lines 826, 1009
2. Run guard-specific tests
3. Check guard expression preservation

**Validation**:
```bash
# Verify no TODOs
grep -n "TODO.*guard" pkg/plugin/builtin/pattern_match_test.go

# Run guard tests
go test ./pkg/plugin/builtin -v -run "Guard"
```

**Success**: 2/2 guard tests pass, no TODOs remain

---

## Edge Cases to Test

### Task 1 Edge Cases
1. **Nested functions** - Return type inference through multiple levels
2. **Multi-value returns** - `func f() (int, error)` → None for error position
3. **Parallel assignment** - `x, y = None, Some(5)` → correct types
4. **Variadic calls** - `fmt.Printf("%v", None)` → interface{} param type

### Task 2 Edge Cases
1. **Type aliases** - `type MyResult = Result_int_error`
2. **Function returns** - `match getResult() { ... }`
3. **Struct fields** - `match user.result { ... }`
4. **Fallback to heuristics** - When go/types unavailable

### Task 3 Edge Cases
1. **Return context** - `func f() Result_int_error { return Err(...) }`
2. **Assignment context** - `var r Result_int_error; r = Err(...)`
3. **Call arg context** - `handleResult(Err(...))`
4. **Struct literal** - `Response{data: Err(...)}`

### Task 4 Edge Cases
1. **Pattern-bound vars only** - `Ok(x) if x > 0`
2. **Outer scope refs** - `Ok(x) if x > threshold`
3. **Complex expressions** - `Ok(x) if x > 0 && x < maxVal`
4. **Invalid guards** - Non-boolean, malformed syntax

---

## Performance Benchmarks

**Target**: <15ms overhead per file

**Benchmark Test** (if exists):
```bash
go test ./pkg/plugin/builtin -bench=BenchmarkTypeInference -benchmem
```

**Manual Validation**:
1. Run tests with `-v` flag
2. Check execution time for context inference
3. Verify <150μs per inference call

**Acceptance**:
- Total overhead <15ms per file
- Individual inference calls <150μs
- No memory leaks

---

## Test Result Documentation

All test results will be documented in:
- **File**: `ai-docs/sessions/20251119-012055/04-testing/test-results.md`

**Structure**:
1. Executive summary (pass/fail counts)
2. Unit test results (detailed)
3. Integration test results
4. Golden test results
5. Success metrics achieved
6. Failing tests (if any) with error details
7. Performance benchmarks
8. Recommendations

---

## Rollback Criteria

**If any of these occur, recommend rollback**:
1. >10% regression in existing test pass rate
2. Performance degradation >30ms per file
3. Critical bug introduced (nil panics, infinite loops)
4. Build fails on clean compilation

**Current Pre-existing Failures**:
- `TestPatternMatchPlugin_Transform_AddsPanic` - KNOWN ISSUE (unrelated to changes)

---

## Next Steps After Testing

1. **If all tests pass**: Document success, proceed to documentation update
2. **If tests fail**: Analyze failures, categorize as implementation bugs vs test issues
3. **If metrics not met**: Identify gaps, recommend additional implementation
4. **If golden tests missing**: Create golden tests per plan

---

## Test Execution Commands Reference

```bash
# Quick smoke test (Task 1 only)
go test ./pkg/plugin/builtin -v -run "TestFind.*|TestContainsNode|TestStrictGoTypesRequirement"

# Full unit test suite
go test ./pkg/plugin/builtin -v

# Specific tasks
go test ./pkg/plugin/builtin -v -run "TestPatternMatch.*"
go test ./pkg/plugin/builtin -v -run "TestResult.*"

# Integration tests
go test ./tests -v -run TestIntegrationPhase4

# Golden tests
go test ./tests -v -run TestGoldenFiles

# Build validation
go build ./pkg/plugin/builtin/...
go build ./cmd/dingo/...

# Performance
go test ./pkg/plugin/builtin -bench=BenchmarkTypeInference -benchmem
```

---

## Conclusion

This test plan provides comprehensive coverage for all 4 tasks:
- **31 unit tests** for Task 1 (foundation)
- **Pattern match tests** for Task 2 (go/types integration)
- **7 Err() tests** for Task 3 (context inference)
- **2 guard tests** for Task 4 (validation)

**Total Expected Tests**: 40+ unit tests + golden tests + integration tests

**Execution Time**: ~5-10 minutes for full suite

**Success Criteria**: 90%+ pass rate, all metrics achieved, no regressions

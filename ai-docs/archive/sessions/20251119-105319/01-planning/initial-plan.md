# Implementation Plan: Test Failure Fixes

**Session**: 20251119-105319
**Based on**: 4-model analysis (session 20251119-101726)
**Target**: Fix 14+ test failures (increase from 261/267 to 265-267 passing)

---

## Executive Summary

**Root Cause**: 90% test infrastructure issues, 10% implementation bugs
**Strategy**: Quick wins first (missing files + naming), then bug fixes, then updates
**Execution**: 3 batches (parallel where possible)
**Estimated Time**: 6-8 hours total

---

## Batch 1: Quick Wins (Parallel Execution) - 1-2 hours

### Task 1.1: Create Missing Golden Files ⚠️ CRITICAL
**Impact**: Fixes 7-8 test failures immediately
**Execution**: Can run in parallel with Task 1.2

**Files to Create** (7 files):
1. `tests/golden/pattern_match_06_guards_nested.go.golden`
2. `tests/golden/pattern_match_07_guards_complex.go.golden`
3. `tests/golden/pattern_match_08_guards_edge_cases.go.golden`
4. `tests/golden/pattern_match_09_tuple_pairs.go.golden`
5. `tests/golden/pattern_match_10_tuple_triples.go.golden`
6. `tests/golden/pattern_match_11_tuple_wildcards.go.golden`
7. `tests/golden/pattern_match_12_tuple_exhaustiveness.go.golden`

**Method**:
```bash
# For each missing .dingo file, transpile and rename output
dingo build tests/golden/pattern_match_06_guards_nested.dingo
mv tests/golden/pattern_match_06_guards_nested.go tests/golden/pattern_match_06_guards_nested.go.golden

# Repeat for remaining 6 files
```

**Validation**:
- Each .go.golden file should be valid Go code
- Run: `go test ./tests -run TestGoldenFiles/pattern_match_06 -v` (confirm passes)
- Verify generated code compiles: `go build tests/golden/pattern_match_06_guards_nested.go.golden`

**Success Criteria**: 7 new golden tests pass

---

### Task 1.2: Fix Result Type Naming Inconsistency ⚠️ CRITICAL
**Impact**: Fixes integration test "undefined" errors
**Execution**: Can run in parallel with Task 1.1

**Decision Required**: Choose naming convention:
- **Option A**: `ResultTagOk` (camelCase, no underscore) - Go idiomatic
- **Option B**: `ResultTag_Ok` (with underscore) - current code output

**Recommended**: Option A (Go naming conventions favor camelCase for exported constants)

**File to Modify**: `pkg/generator/result_option.go`

**Change** (approximate line 150-180 in `injectResultType()` function):

```go
// OLD (current - with underscore)
const ResultTag_Ok = 0
const ResultTag_Err = 1

type Result_T_E struct {
    tag int
    ok  T
    err E
}

// NEW (recommended - camelCase)
const ResultTagOk = 0
const ResultTagErr = 1

type Result_T_E struct {
    tag int
    ok  T
    err E
}
```

**If choosing Option B**, update integration tests instead:
- File: `tests/integration_phase4_test.go`
- Search for: `ResultTagOk`, `ResultTagErr`
- Replace with: `ResultTag_Ok`, `ResultTag_Err`

**Validation**:
- Run: `go test ./tests -run TestIntegrationPhase4 -v`
- Confirm: No "undefined: ResultTagOk" errors
- Verify: Type checking passes

**Success Criteria**: Integration tests no longer fail on undefined constants

---

### Batch 1 Expected Results

**Tests Passing**: 261 → 268-270 (gain of 7-9 tests)
**Time**: 1-2 hours
**Validation Command**: `go test ./tests -v | grep -E "(PASS|FAIL)"`

---

## Batch 2: Implementation Bug Fix (Sequential) - 2-3 hours

### Task 2.1: Fix Error Propagation Single-Error Return Bug 🐛
**Impact**: Fixes 2 compilation test failures
**Execution**: Must run after Batch 1 (to isolate this fix's impact)

**File to Modify**: `pkg/generator/preprocessor/error_prop.go`

**Problem**:
When function signature returns ONLY `error` (no value), the `?` operator generates:
```go
result?  // Invalid transpilation below
↓
if err != nil {
    return , err  // ❌ Extra comma! Invalid Go syntax
}
```

**Root Cause**:
Error propagation preprocessor assumes all functions return `(value, error)` tuple.
Doesn't handle single `error` return type.

**Fix Location** (approximate):
- Function: `transformErrorProp()` or similar
- Line: ~200-250 (search for `return.*err` generation)

**Implementation**:

```go
// Pseudocode fix
func generateReturnStatement(funcSig FunctionSignature) string {
    returnTypes := funcSig.Results()

    // NEW: Check if single error return
    if len(returnTypes) == 1 && isErrorType(returnTypes[0]) {
        return "return err"
    }

    // Existing: Multiple returns or (value, error)
    zeroValues := generateZeroValues(returnTypes[:len(returnTypes)-1])
    return fmt.Sprintf("return %s, err", zeroValues)
}
```

**Specific Steps**:
1. Locate error propagation transform code
2. Add check for single-error return signature
3. Generate `return err` (no comma) for single error
4. Keep existing logic for multi-value returns

**Test Cases to Verify**:
```go
// Case 1: Single error return
func validate() error {
    doSomething()?  // Should generate: return err (not return , err)
}

// Case 2: Value + error return (existing, should still work)
func compute() (int, error) {
    calculate()?  // Should generate: return 0, err
}

// Case 3: Multiple values + error
func fetch() (string, int, error) {
    load()?  // Should generate: return "", 0, err
}
```

**Validation**:
- Run: `go test ./tests -run TestGoldenFilesCompilation -v`
- Confirm: error_prop tests with single-error returns compile
- Verify: Generated Go code is valid (no syntax errors)

**Success Criteria**: Compilation tests pass, no `return , err` in generated code

---

### Batch 2 Expected Results

**Tests Passing**: 268-270 → 270-272 (gain of 2 tests)
**Time**: 2-3 hours
**Validation Command**: `go test ./tests -run Compilation -v`

---

## Batch 3: Update Outdated Golden Files (Sequential) - 2-3 hours

### Task 3.1: Identify Outdated Golden Files
**Impact**: Fixes diff mismatches in existing tests
**Execution**: Must run after Batch 2 (to see which tests still fail)

**Method**:
```bash
# Run tests and capture failures
go test ./tests -v 2>&1 | grep "golden file mismatch" > outdated_tests.txt

# Example failures (from analysis):
# - option_02_literals
# - error_prop_02_multiple
# - (others TBD)
```

**Root Cause**:
Recent changes improved code generation:
- Variable hoisting (reduces scope pollution)
- Import optimization (removes unused imports)
- Comment cleanup (eliminates transform artifacts)

Golden files reflect OLD output format, need regeneration.

---

### Task 3.2: Regenerate Outdated Golden Files

**Known Candidates** (from analysis):
1. `tests/golden/option_02_literals.go.golden`
2. `tests/golden/error_prop_02_multiple.go.golden`
3. (Additional files identified in Task 3.1)

**Method for Each File**:
```bash
# Step 1: Backup old golden file (for comparison)
cp tests/golden/option_02_literals.go.golden tests/golden/option_02_literals.go.golden.bak

# Step 2: Regenerate
dingo build tests/golden/option_02_literals.dingo
mv tests/golden/option_02_literals.go tests/golden/option_02_literals.go.golden

# Step 3: Validate
go build tests/golden/option_02_literals.go.golden  # Must compile
go test ./tests -run TestGoldenFiles/option_02_literals -v  # Must pass

# Step 4: Review diff (ensure changes are expected improvements)
diff tests/golden/option_02_literals.go.golden.bak tests/golden/option_02_literals.go.golden
```

**Expected Diffs** (should see improvements like):
- Removed unused imports
- Reduced variable scope (hoisting)
- Cleaner code structure
- No comment pollution

**Validation Per File**:
1. Generated Go compiles without errors
2. Golden test passes (diff matches)
3. Diff shows IMPROVEMENTS (not regressions)

---

### Task 3.3: Bulk Regeneration (if many files)

If >5 files need updates:

```bash
# Identify all failing golden tests
go test ./tests -run TestGoldenFiles -v 2>&1 | \
  grep -E "FAIL.*golden" | \
  sed 's/.*TestGoldenFiles\///' | \
  sed 's/ .*//' > failing_tests.txt

# Regenerate in batch
while read test_name; do
  echo "Regenerating: $test_name"
  dingo build "tests/golden/${test_name}.dingo"
  mv "tests/golden/${test_name}.go" "tests/golden/${test_name}.go.golden"

  # Validate
  go build "tests/golden/${test_name}.go.golden" && echo "✓ Compiles" || echo "✗ FAILED"
done < failing_tests.txt
```

**Validation**:
- Run full golden test suite: `go test ./tests -run TestGoldenFiles -v`
- Confirm: All updated tests pass
- Review: Random sample of diffs (ensure quality improvements)

**Success Criteria**: All golden file tests pass (no diff mismatches)

---

### Batch 3 Expected Results

**Tests Passing**: 270-272 → 265-267 (close to 100%)
**Time**: 2-3 hours
**Validation Command**: `go test ./tests -v`

---

## Batch 4: Deferred Enhancements (Future) - 4-6 hours

### Task 4.1: None Context Inference for Return Statements
**Impact**: Fixes 1 integration test (none_context_inference_return)
**Priority**: LOW (affects only 1 test, edge case)

**File to Modify**: `pkg/types/inference.go`

**Issue**:
None literal in return statement doesn't infer Option type from function signature.

```go
func getConfig() Option[Config] {
    return None  // ❌ Doesn't infer Option[Config]
}
```

**Fix**:
Add return statement context to inference engine:
1. Track parent function signature
2. When encountering None in return, check expected return type
3. Infer Option type parameter from function signature

**Defer Because**:
- Low impact (1 test)
- High complexity (requires AST parent tracking)
- Workaround exists (explicit type: `None[Config]`)

---

### Task 4.2: Enhanced Tuple Pattern Support
**Impact**: Potentially improves tuple destructuring
**Priority**: LOW (not causing current failures)

**Defer**: Not critical for current test suite

---

## File Modification Summary

### New Files (7)
All in `tests/golden/`:
- `pattern_match_06_guards_nested.go.golden`
- `pattern_match_07_guards_complex.go.golden`
- `pattern_match_08_guards_edge_cases.go.golden`
- `pattern_match_09_tuple_pairs.go.golden`
- `pattern_match_10_tuple_triples.go.golden`
- `pattern_match_11_tuple_wildcards.go.golden`
- `pattern_match_12_tuple_exhaustiveness.go.golden`

### Modified Files (2)
- `pkg/generator/result_option.go` (naming fix)
- `pkg/generator/preprocessor/error_prop.go` (single-error return fix)

### Updated Files (~2-5)
In `tests/golden/`:
- `option_02_literals.go.golden`
- `error_prop_02_multiple.go.golden`
- (Others identified during Batch 3)

---

## Execution Strategy

### Parallel vs Sequential

**Batch 1** (Parallel):
- Task 1.1 (create golden files) || Task 1.2 (naming fix)
- Independent tasks, no dependencies
- Can delegate to 2 agents simultaneously

**Batch 2** (Sequential):
- Must run after Batch 1 completes
- Allows isolation of error propagation fix impact

**Batch 3** (Sequential):
- Must run after Batch 2 completes
- Ensures only truly outdated files are regenerated (not hiding other bugs)

### Agent Delegation Plan

**Batch 1.1**: `golang-developer` agent
- Task: Create 7 missing golden files
- Input: List of .dingo files
- Output: 7 .go.golden files + summary

**Batch 1.2**: `golang-developer` agent (different instance, parallel)
- Task: Fix Result naming in result_option.go
- Input: Naming convention choice (from user)
- Output: Modified file + summary

**Batch 2**: `golang-developer` agent
- Task: Fix error propagation bug
- Input: error_prop.go file
- Output: Modified file + test results summary

**Batch 3**: `golang-developer` agent
- Task: Regenerate outdated golden files
- Input: List of failing tests
- Output: Updated .go.golden files + summary

---

## Validation Checkpoints

### After Batch 1
```bash
go test ./tests -v
# Expected: 268-270 tests passing (up from 261)
```

### After Batch 2
```bash
go test ./tests -run Compilation -v
# Expected: 270-272 tests passing
```

### After Batch 3
```bash
go test ./tests -v
# Expected: 265-267 tests passing (98-100%)
```

### Final Validation
```bash
# Full suite
go test ./tests -v -count=1

# Specific categories
go test ./tests -run TestGoldenFiles -v
go test ./tests -run TestIntegrationPhase4 -v
go test ./tests -run TestGoldenFilesCompilation -v

# Ensure generated code compiles
for golden in tests/golden/*.go.golden; do
  go build "$golden" 2>&1 | grep -q "error" && echo "FAIL: $golden" || echo "OK: $golden"
done
```

---

## Risk Mitigation

### Risk 1: Naming Convention Choice Breaks More Tests
**Mitigation**:
- Test both options before committing
- Run full test suite after change
- Revert if >5 tests break unexpectedly

### Risk 2: Error Propagation Fix Breaks Existing Tests
**Mitigation**:
- Run error_prop golden tests before/after fix
- Compare test counts (should only GAIN tests, not lose)
- Keep old implementation in comments (easy rollback)

### Risk 3: Golden File Regeneration Hides Bugs
**Mitigation**:
- Review diffs for each regenerated file
- Ensure changes are IMPROVEMENTS (hoisting, imports)
- Don't regenerate files with unexpected changes

### Risk 4: Time Overruns
**Mitigation**:
- Batch 1 & 2 are CRITICAL (7-9 test wins)
- Batch 3 is IMPORTANT (final polish)
- Batch 4 is DEFERRED (can skip for now)

---

## Success Metrics

### Quantitative
- **Before**: 261/267 tests passing (97.8%)
- **After Batch 1**: 268-270/267 passing
- **After Batch 2**: 270-272/267 passing
- **After Batch 3**: 265-267/267 passing (98-100%)

### Qualitative
- No compilation errors in golden tests
- Generated Go code is idiomatic (no syntax hacks)
- Test suite runs cleanly (no flaky tests)
- All fixes are sustainable (not workarounds)

---

## Open Questions (For User Decision)

### Q1: Result Type Naming Convention
**Question**: Should Result tag constants use underscores or camelCase?
- **Option A**: `ResultTagOk`, `ResultTagErr` (Go idiomatic)
- **Option B**: `ResultTag_Ok`, `ResultTag_Err` (current implementation)

**Impact**: Determines which file to modify (code or tests)
**Recommendation**: Option A (follows Go style guide)

### Q2: Should We Defer Batch 4?
**Question**: Implement None inference enhancement now or later?
- **Option A**: Now (complete fix, 100% tests passing)
- **Option B**: Later (focuses on critical issues, 98% passing acceptable)

**Impact**: 4-6 hours additional work for 1 test
**Recommendation**: Option B (defer to future iteration)

---

## Timeline Estimate

| Batch | Tasks | Time | Cumulative |
|-------|-------|------|------------|
| Batch 1 | Create golden files + Naming fix | 1-2h | 1-2h |
| Batch 2 | Error propagation bug fix | 2-3h | 3-5h |
| Batch 3 | Regenerate outdated golden files | 2-3h | 5-8h |
| **Total (Critical)** | | **5-8h** | |
| Batch 4 (Deferred) | None inference | 4-6h | 9-14h |

**Recommended**: Execute Batches 1-3 now (5-8 hours), defer Batch 4

---

## Next Steps

1. **User Decision**: Choose Result naming convention (Q1)
2. **Execute Batch 1**: Create golden files + naming fix (parallel agents)
3. **Validate**: Run test suite, confirm 7-9 test improvement
4. **Execute Batch 2**: Fix error propagation bug
5. **Validate**: Run compilation tests, confirm 2 test improvement
6. **Execute Batch 3**: Regenerate outdated golden files
7. **Final Validation**: Full test suite, target 265-267 passing
8. **Report**: Summary of fixes + remaining issues (if any)

---

## References

- **Investigation Analysis**: ai-docs/sessions/20251119-101726/03-analysis/consolidated-findings.md
- **Model Analyses**: ai-docs/sessions/20251119-101726/02-investigation/*.md
- **Golden Test Guidelines**: tests/golden/GOLDEN_TEST_GUIDELINES.md
- **Current Test Status**: 261/267 passing (97.8%)

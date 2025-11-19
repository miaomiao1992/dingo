# Task F: Changes Made

## File: `/Users/jack/mag/dingo/pkg/preprocessor/preprocessor_test.go`

### Added Test Function: `TestSourceMapOffsetBeforeImports`

**Location**: Lines 740-796

**Purpose**: Negative test to verify that source map offset adjustments are NOT applied to mappings before or at the import insertion line. This test specifically validates the CRITICAL-1 fix (>= to > change in adjustMappingsForImports).

**Test Strategy**:
1. Create a SourceMap with mappings at different GeneratedLine positions
2. Call adjustMappingsForImports with importInsertionLine=2, numImportLines=2
3. Verify behavior for each mapping:
   - Line 1 (< 2): NOT shifted ✓
   - Line 2 (= 2): NOT shifted ✓ **[CRITICAL TEST]**
   - Line 3 (> 2): Shifted by +2 to line 5 ✓
   - Line 4 (> 2): Shifted by +2 to line 6 ✓

**Key Test Cases**:
```go
// Mapping at insertionLine (= 2) → CRITICAL TEST
if sourceMap.Mappings[1].GeneratedLine != 2 {
    t.Errorf("CRITICAL REGRESSION: Mapping at insertionLine %d was incorrectly shifted to line %d. "+
        "This indicates the >= bug has returned (should use > not >=)",
        importInsertionLine, sourceMap.Mappings[1].GeneratedLine)
}
```

This test would FAIL if the code used `>=` instead of `>`, catching the exact regression described in the code review.

**Test Result**: PASS ✓

```
=== RUN   TestSourceMapOffsetBeforeImports
    preprocessor_test.go:791: ✓ All mappings correctly handled:
    preprocessor_test.go:792:   Line 1 (< 2): NOT shifted (correct)
    preprocessor_test.go:793:   Line 2 (= 2): NOT shifted (CRITICAL FIX VERIFIED)
    preprocessor_test.go:794:   Line 3 (> 2): Shifted to 5 (correct)
    preprocessor_test.go:795:   Line 4 (> 2): Shifted to 6 (correct)
--- PASS: TestSourceMapOffsetBeforeImports (0.00s)
```

## Summary

- **Lines Added**: 57 lines (including comments and test logic)
- **Tests Added**: 1 comprehensive negative test
- **Test Coverage**: Validates all boundary conditions (before, at, and after importInsertionLine)
- **Regression Prevention**: This test will immediately fail if someone changes `>` back to `>=`

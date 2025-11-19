# Task F: Implementation Notes

## Objective
Add a negative test to verify that source map offsets are NOT applied to mappings before the import insertion line, specifically testing the boundary condition at the insertion line itself.

## Implementation Approach

### Test Design Philosophy
Instead of testing through the full Preprocessor pipeline (which would require complex input files and parsing), this test directly calls the internal `adjustMappingsForImports` function. This provides:

1. **Precision**: Exact control over input mappings and parameters
2. **Clarity**: Clear visibility of the boundary condition being tested
3. **Simplicity**: No need to construct complex Dingo source files
4. **Speed**: Unit-level test runs in microseconds

### Critical Boundary Condition
The test specifically targets the `=` case (GeneratedLine == importInsertionLine):

```go
// This is the mapping AT the import insertion line
{OriginalLine: 3, OriginalColumn: 1, GeneratedLine: 2, GeneratedColumn: 1, Length: 4, Name: "type"}
```

With `importInsertionLine = 2`, this mapping represents content AT the insertion point. The correct behavior (using `>`) is to NOT shift this mapping, because it represents code that appears BEFORE the imports in the final output.

**Bug Scenario** (if using `>=`):
- Mapping with GeneratedLine=2 would incorrectly shift to line 4
- This breaks IDE navigation for package-level declarations
- The test would FAIL with clear error message

**Correct Behavior** (using `>`):
- Mapping with GeneratedLine=2 stays at line 2
- Only mappings with GeneratedLine > 2 are shifted
- The test PASSES ✓

### Test Coverage Matrix

| GeneratedLine | Relation to insertionLine (2) | Expected Behavior | Test Verification |
|---------------|-------------------------------|-------------------|-------------------|
| 1             | < 2                          | NOT shifted       | ✓ Stays at 1      |
| 2             | = 2 **[CRITICAL]**           | NOT shifted       | ✓ Stays at 2      |
| 3             | > 2                          | Shifted by +2     | ✓ Becomes 5       |
| 4             | > 2                          | Shifted by +2     | ✓ Becomes 6       |

### Error Message Quality
The test includes a detailed error message for the critical case:

```go
t.Errorf("CRITICAL REGRESSION: Mapping at insertionLine %d was incorrectly shifted to line %d. "+
    "This indicates the >= bug has returned (should use > not >=)",
    importInsertionLine, sourceMap.Mappings[1].GeneratedLine)
```

This message:
- Identifies the issue as a CRITICAL REGRESSION
- Explains WHY it's a problem (>= vs > bug)
- Provides actionable debugging information
- Helps future developers understand the fix

### Integration with Task A
This test validates the fix implemented in Task A:

**Task A Fix** (preprocessor.go:188):
```go
if sourceMap.Mappings[i].GeneratedLine > importInsertionLine {  // Changed from >=
    sourceMap.Mappings[i].GeneratedLine += numImportLines
}
```

**Task F Test** (preprocessor_test.go:775-779):
```go
// Mapping 1: GeneratedLine=2 (= 2) → CRITICAL TEST: should NOT shift (stay at 2)
if sourceMap.Mappings[1].GeneratedLine != 2 {
    t.Errorf("CRITICAL REGRESSION: ...")
}
```

The test PASSES because Task A changed `>=` to `>`, ensuring mappings at the insertion line are not shifted.

## Testing Process

### Build & Test
```bash
go test ./pkg/preprocessor -run TestSourceMapOffsetBeforeImports -v
```

### Result
```
=== RUN   TestSourceMapOffsetBeforeImports
    preprocessor_test.go:791: ✓ All mappings correctly handled:
    preprocessor_test.go:792:   Line 1 (< 2): NOT shifted (correct)
    preprocessor_test.go:793:   Line 2 (= 2): NOT shifted (CRITICAL FIX VERIFIED)
    preprocessor_test.go:794:   Line 3 (> 2): Shifted to 5 (correct)
    preprocessor_test.go:795:   Line 4 (> 2): Shifted to 6 (correct)
--- PASS: TestSourceMapOffsetBeforeImports (0.00s)
PASS
```

**Status**: ✓ All assertions pass

## Validation Against Requirements

### Original Task Requirements
From `/Users/jack/mag/dingo/ai-docs/sessions/20251117-224213/01-planning/final-plan.md`:

> **Task 2.2**: Add a test function TestSourceMapOffsetBeforeImports or similar
> - Test scenario: Create mappings with GeneratedLine values both BEFORE and AFTER the import insertion line
> - Call adjustMappingsForImports with importInsertionLine=2, numImportLines=2
> - Verify: Mappings with GeneratedLine < 2 are NOT shifted
> - Verify: Mappings with GeneratedLine = 2 are NOT shifted (this is the critical test!)
> - Verify: Mappings with GeneratedLine > 2 ARE shifted by +2

### Requirement Fulfillment

| Requirement | Status | Evidence |
|-------------|--------|----------|
| Test function name | ✓ | `TestSourceMapOffsetBeforeImports` |
| Mappings before insertion | ✓ | GeneratedLine=1 (stays at 1) |
| Mappings at insertion | ✓ | GeneratedLine=2 (stays at 2) **[CRITICAL]** |
| Mappings after insertion | ✓ | GeneratedLine=3,4 (shifted to 5,6) |
| Use importInsertionLine=2 | ✓ | `importInsertionLine := 2` |
| Use numImportLines=2 | ✓ | `numImportLines := 2` |
| Run test successfully | ✓ | PASS (0.00s) |

**All requirements met** ✓

## Maintenance Considerations

### Future Proofing
This test serves as a **permanent regression guard**. If any future refactoring accidentally changes:
- `>` back to `>=` in adjustMappingsForImports
- The logic for determining which mappings to shift
- The offset calculation

This test will immediately fail with a clear error message explaining the issue.

### Documentation Value
The test serves as executable documentation of the correct behavior:
- Clear comments explain the scenario
- Descriptive variable names (importInsertionLine, numImportLines)
- Comprehensive logging shows all test cases
- Error messages provide troubleshooting guidance

### Complementary Tests
This test complements the existing integration tests:
- `TestCRITICAL1_MappingsBeforeImportsNotShifted`: End-to-end test with real Dingo code
- `TestSourceMapGeneration`: Verifies mappings are correctly generated
- `TestSourceMapMultipleExpansions`: Verifies multiple error propagations

Together, these provide defense-in-depth against source map regressions.

## Conclusion

**Task Status**: COMPLETE ✓

The negative test successfully validates that:
1. The Task A fix (>= to >) is working correctly
2. Mappings before the import insertion line are not shifted
3. Mappings AT the insertion line are not shifted (critical boundary)
4. Mappings after the insertion line ARE shifted correctly
5. The test will catch any future regressions

**Confidence Level**: HIGH
- Test passes with current implementation
- Test would fail with the original bug
- Test covers all boundary conditions
- Test is maintainable and well-documented

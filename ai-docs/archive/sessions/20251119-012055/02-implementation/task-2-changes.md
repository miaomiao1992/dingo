# Task 2: go/types Integration for Pattern Match Scrutinee - Changes

## Files Modified

### 1. pkg/plugin/builtin/pattern_match.go
**Lines modified**: ~80 lines (imports + 3 new functions + 1 modified function)

**Changes**:
1. Added `go/types` import
2. Modified `checkExhaustiveness()` to call new `getAllVariantsWithTypes()` function
3. Added `getAllVariantsWithTypes()` - uses go/types first, then falls back to heuristics
4. Added `extractVariantsFromType()` - extracts Result/Option variants from types.Type
5. Updated `getAllVariants()` - simplified to pure heuristics (removed TODO comment)

**Implementation Details**:
- Uses `p.ctx.TypeInfo` (go/types.Info) to get type of scrutinee
- Accesses `match.switchStmt.Tag` (the AST expression being switched on)
- Checks for structs with "tag" field (our Result/Option pattern)
- Handles type aliases by checking underlying type
- Falls back to heuristics if go/types unavailable

### 2. pkg/plugin/builtin/pattern_match_test.go
**Lines added**: 270+ lines (4 new test functions)

**New Tests**:
1. `TestPatternMatchPlugin_GoTypesIntegration_TypeAlias` - Tests type alias detection
2. `TestExtractVariantsFromType` - Unit tests for extractVariantsFromType function
3. `TestPatternMatchPlugin_GoTypesIntegration_FunctionReturn` - Tests function return types
4. `TestPatternMatchPlugin_GoTypesUnavailable` - Tests fallback when go/types unavailable

**Test Coverage**:
- Result types, Option types, type aliases
- Type detection via go/types
- Fallback to heuristics when go/types unavailable
- Edge cases: incomplete type checking environment

## Test Results

All new tests pass:
```
TestPatternMatchPlugin_GoTypesIntegration_TypeAlias    PASS
TestExtractVariantsFromType                            PASS
  - Result_type                                        PASS
  - Option_type                                        PASS
  - Type_alias_to_Result                               PASS
  - Non-Result/Option_type                             PASS
TestPatternMatchPlugin_GoTypesIntegration_FunctionReturn PASS
TestPatternMatchPlugin_GoTypesUnavailable              PASS
```

Pre-existing failures (NOT related to this task):
- TestPatternMatchPlugin_Transform_AddsPanic (known issue)
- TestPatternMatchPlugin_GuardTransformation (known issue)

## Summary

- **Files modified**: 2
- **Lines of code**: ~350 (80 in implementation, 270 in tests)
- **Test coverage**: 4 new test functions, 8 test cases
- **Status**: SUCCESS - All new tests pass, no regressions

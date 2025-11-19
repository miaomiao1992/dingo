# Task A: Source Map Offset Bug Fix - Changes

## Files Modified

### /Users/jack/mag/dingo/pkg/preprocessor/preprocessor.go

**Lines Modified**: 183-207

**Change Summary**: Fixed CRITICAL-2 source map offset bug by changing `>=` to `>` in line 203 and added comprehensive documentation explaining the fix.

**Detailed Changes**:

1. **Line 203** (previously line 188): Changed condition from `>=` to `>`
   - **Before**: `if sourceMap.Mappings[i].GeneratedLine >= importInsertionLine {`
   - **After**: `if sourceMap.Mappings[i].GeneratedLine > importInsertionLine {`

2. **Lines 187-202**: Added comprehensive comment block explaining:
   - Why we use `>` instead of `>=`
   - What `importInsertionLine` represents
   - Concrete example showing the line shift logic
   - Why mappings AT the insertion line should NOT be shifted

**Impact**:
- Prevents incorrect source map offsets for package-level declarations
- Fixes IDE navigation issues where mappings before import block were incorrectly shifted
- All existing tests continue to pass
- Test `TestCRITICAL1_MappingsBeforeImportsNotShifted` validates the fix

## Test Results

```
go test ./pkg/preprocessor/... -v
```

**Result**: PASS (all 12 test suites, 0 failures)

**Key Test**: `TestCRITICAL1_MappingsBeforeImportsNotShifted` - PASS
- Confirms mappings before import block are NOT shifted
- Confirms mappings after import block ARE shifted correctly
- Validates the `>` vs `>=` fix

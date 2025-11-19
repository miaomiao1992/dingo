# Task G Changes: Negative Tests for User Function Shadowing

## Summary
Added comprehensive negative tests to verify that user-defined functions with stdlib names do NOT trigger automatic import injection (Issue #3 fix verification).

## Files Modified

### `/Users/jack/mag/dingo/pkg/preprocessor/preprocessor_test.go`
- **Lines Added**: 1174-1323 (150 lines)
- **Test Added**: `TestUserFunctionShadowingNoImport`

## Changes Detail

### New Test: TestUserFunctionShadowingNoImport

**Purpose**: Verify Issue #3 fix at the ImportTracker level by checking GetNeededImports() output.

**Test Cases**:
1. **User function named ReadFile - no os import**
   - Defines user function `ReadFile(path string) ([]byte, error)`
   - Calls it with `?` operator
   - Verifies `os` package is NOT imported

2. **User function named Atoi - no strconv import**
   - Defines user function `Atoi(s string) (int, error)`
   - Calls it with `?` operator
   - Verifies `strconv` package is NOT imported

3. **Qualified os.ReadFile call - SHOULD import os**
   - Calls `os.ReadFile(path)?` with package qualifier
   - Verifies `os` package IS imported (positive control)

4. **Mixed user-defined and qualified stdlib**
   - Defines user functions `ReadFile()` and `Atoi()`
   - Calls both user functions and qualified stdlib functions
   - Verifies only qualified calls trigger imports

**Implementation Approach**:
- Creates ErrorPropProcessor directly via `NewErrorPropProcessorWithConfig()`
- Processes input through processor
- Calls `GetNeededImports()` to check tracked imports
- Verifies expected imports are present
- Verifies unexpected imports are absent

**Key Verification Logic**:
```go
// Check that user-defined functions did NOT trigger imports
for _, pkgName := range tt.shouldNotTrack {
    var importPath string
    switch pkgName {
    case "os":
        importPath = "os"
    case "strconv":
        importPath = "strconv"
    case "json":
        importPath = "encoding/json"
    default:
        importPath = pkgName
    }

    if neededMap[importPath] {
        t.Errorf("Package %q should NOT be imported (user-defined function)")
    }
}
```

## Test Results

All tests pass:
```
=== RUN   TestUserFunctionShadowingNoImport
=== RUN   TestUserFunctionShadowingNoImport/user_function_named_ReadFile_-_no_os_import
=== RUN   TestUserFunctionShadowingNoImport/user_function_named_Atoi_-_no_strconv_import
=== RUN   TestUserFunctionShadowingNoImport/qualified_os.ReadFile_call_-_SHOULD_import_os
=== RUN   TestUserFunctionShadowingNoImport/mixed_user-defined_and_qualified_stdlib
--- PASS: TestUserFunctionShadowingNoImport (0.00s)
```

## Verification

Confirms that Issue #3 (IMPORTANT-1 fix) is working correctly:
- User-defined functions like `ReadFile()` do NOT trigger `os` import
- User-defined functions like `Atoi()` do NOT trigger `strconv` import
- Only qualified calls like `os.ReadFile()` trigger imports
- Mixed scenarios correctly distinguish user vs. stdlib functions

## Related Code

The test verifies the fix implemented in:
- `/Users/jack/mag/dingo/pkg/preprocessor/error_prop.go:862-874` (trackFunctionCallInExpr)
- Specifically lines 865-873 which document the IMPORTANT-1 FIX

## Coverage Impact

- Adds negative test coverage for import tracking
- Complements existing `TestIMPORTANT1_UserDefinedFunctionsDontTriggerImports`
- Provides direct verification of ImportTracker behavior via GetNeededImports()
- Prevents regression of Issue #3 fix

## Notes

- Test accesses processor directly rather than full preprocessor pipeline
- This allows inspection of GetNeededImports() output
- All 4 test cases verify both positive (should import) and negative (should not import) scenarios

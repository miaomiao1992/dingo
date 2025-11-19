# Task G Implementation Notes

## Objective
Add negative tests to ensure user-defined functions with stdlib names DON'T trigger imports, verifying Issue #3 fix.

## Approach

### Initial Challenge
- Attempted to access `p.importTracker` directly from Preprocessor
- Discovered ImportTracker is private to ErrorPropProcessor
- ImportTracker is accessed via ImportProvider interface (GetNeededImports)

### Solution
Created test that instantiates ErrorPropProcessor directly:
```go
proc := NewErrorPropProcessorWithConfig(DefaultConfig())
_, _, err := proc.Process([]byte(tt.input))
neededImports := proc.GetNeededImports()
```

This allows:
- Direct verification of import tracking behavior
- Access to GetNeededImports() method
- Isolated testing of ErrorPropProcessor logic

## Test Design

### Test Structure
```go
tests := []struct {
    name           string
    input          string
    shouldTrack    map[string]bool  // Functions that SHOULD trigger imports
    shouldNotTrack []string         // Packages that should NOT be imported
}
```

### Coverage Matrix

| Test Case | User Functions | Qualified Calls | Expected Imports | Expected NO Imports |
|-----------|---------------|-----------------|------------------|---------------------|
| 1. User ReadFile | ReadFile() | none | none | os |
| 2. User Atoi | Atoi() | none | none | strconv |
| 3. Qualified os.ReadFile | none | os.ReadFile() | os | none |
| 4. Mixed | ReadFile(), Atoi() | os.ReadFile(), strconv.Atoi() | os, strconv | none |

### Verification Logic

**Positive Check** (should import):
```go
for funcName := range tt.shouldTrack {
    parts := strings.Split(funcName, ".")
    expectedPkg := parts[0]
    // Map to import path (e.g., "json" → "encoding/json")
    if !neededMap[expectedImport] {
        t.Errorf("Expected import %q not found")
    }
}
```

**Negative Check** (should NOT import):
```go
for _, pkgName := range tt.shouldNotTrack {
    // Map to import path
    if neededMap[importPath] {
        t.Errorf("Unexpected import %q (user-defined function)")
    }
}
```

## Key Insights

### Issue #3 Root Cause
The fix in `error_prop.go:862-874` ensures:
- Only QUALIFIED calls (pkg.Function) trigger imports
- Bare function calls like `ReadFile()` are ignored
- Prevents false positives when users define functions with common names

### Test Verification
Our test verifies this at the ImportTracker level:
- Checks GetNeededImports() output directly
- Confirms imports are tracked correctly
- Ensures user functions don't pollute import list

## Implementation Details

### Package Name Mapping
Need to map package names to import paths:
```go
switch expectedPkg {
case "os":
    expectedImport = "os"
case "strconv":
    expectedImport = "strconv"
case "json":
    expectedImport = "encoding/json"  // Note: not just "json"
case "http":
    expectedImport = "net/http"       // Note: not just "http"
case "filepath":
    expectedImport = "path/filepath"  // Note: not just "filepath"
}
```

This is crucial because:
- Function calls use package prefix (e.g., `json.Marshal`)
- Import paths differ (e.g., `"encoding/json"`)

## Test Execution

### Run Test
```bash
go test ./pkg/preprocessor -run TestUserFunctionShadowingNoImport -v
```

### Results
```
=== RUN   TestUserFunctionShadowingNoImport
=== RUN   TestUserFunctionShadowingNoImport/user_function_named_ReadFile_-_no_os_import
=== RUN   TestUserFunctionShadowingNoImport/user_function_named_Atoi_-_no_strconv_import
=== RUN   TestUserFunctionShadowingNoImport/qualified_os.ReadFile_call_-_SHOULD_import_os
=== RUN   TestUserFunctionShadowingNoImport/mixed_user-defined_and_qualified_stdlib
--- PASS: TestUserFunctionShadowingNoImport (0.00s)
```

All subtests pass, confirming Issue #3 fix is working.

## Relationship to Existing Tests

### TestIMPORTANT1_UserDefinedFunctionsDontTriggerImports
- Tests at the full preprocessor level
- Checks final output (no import block in generated code)
- End-to-end verification

### TestUserFunctionShadowingNoImport (new)
- Tests at the ErrorPropProcessor level
- Checks GetNeededImports() output directly
- Unit-level verification

**Together**: Provide comprehensive coverage of Issue #3 fix at multiple levels.

## Edge Cases Covered

1. **User function with exact stdlib name** - ReadFile, Atoi
2. **Qualified stdlib call** - os.ReadFile (positive control)
3. **Mixed scenario** - Both user and stdlib functions in same code
4. **Multiple user functions** - ReadFile + Atoi

## Lessons Learned

1. **Access patterns matter**: ImportTracker is private, use ImportProvider interface
2. **Test isolation**: Create processor directly for unit-level testing
3. **Package mapping**: Import paths != package prefixes
4. **Dual verification**: Test both "should" and "should not" cases

## Future Enhancements

Potential additional test cases:
- User type with stdlib package name (e.g., `type os struct{}`)
- Nested packages (e.g., `myos.ReadFile` shouldn't trigger `os` import)
- Aliased imports (if supported in future)

## Conclusion

Task G successfully adds negative tests that:
- Verify Issue #3 fix at the ImportTracker level
- Check GetNeededImports() output directly
- Cover positive and negative scenarios
- Complement existing end-to-end tests
- Prevent regression of IMPORTANT-1 fix

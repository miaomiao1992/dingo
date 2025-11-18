# CRITICAL-2 Fix Summary: Multi-Value Return Bug

## Problem Statement

When handling `return expr?` in statement context, the preprocessor incorrectly generated code for multi-value returns, causing silent data loss.

### Before (WRONG):
```dingo
func foo() (int, string, error) {
    return bar()?  // bar returns (int, string, error)
}
```

Generated broken code:
```go
func foo() (int, string, error) {
    __tmp0, __err0 := bar()  // ❌ WRONG: Only captures one value
    if __err0 != nil {
        return 0, "", __err0
    }
    return __tmp0, nil  // ❌ WRONG: Drops string value
}
```

### After (CORRECT):
```go
func foo() (int, string, error) {
    __tmp0, __tmp1, __err2 := bar()  // ✅ Captures all values
    if __err2 != nil {
        return 0, "", __err2
    }
    return __tmp0, __tmp1, nil  // ✅ All values preserved
}
```

## Solution Chosen: Option B (Complete Fix)

**Why this was possible:** The preprocessor already had full type information available through the `parseFunctionSignature` function, which uses `go/ast` and `go/types` to parse function return types.

## Changes Made

### 1. `/Users/jack/mag/dingo/pkg/preprocessor/error_prop.go`

**Modified `expandReturn` function (lines 441-574):**

**Key changes:**
- Calculate number of non-error return values from function signature
- Generate multiple temporary variables: `__tmp0, __tmp1, __tmp2, ...`
- Success path returns all temps: `return __tmp0, __tmp1, __tmp2, nil`
- Error path already had correct zero values from existing `funcContext` logic

**Before:**
```go
tmpVar := fmt.Sprintf("__tmp%d", e.tryCounter)
errVar := fmt.Sprintf("__err%d", e.tryCounter)
e.tryCounter++
```

**After:**
```go
// Determine how many non-error values the function returns
numNonErrorReturns := 1 // default: single value + error
if e.currentFunc != nil && len(e.currentFunc.returnTypes) > 1 {
    numNonErrorReturns = len(e.currentFunc.returnTypes) - 1
}

// Generate temporary variable names for all non-error values
tmpVars := []string{}
for i := 0; i < numNonErrorReturns; i++ {
    tmpVars = append(tmpVars, fmt.Sprintf("__tmp%d", e.tryCounter))
    e.tryCounter++
}
errVar := fmt.Sprintf("__err%d", e.tryCounter)
e.tryCounter++
```

### 2. `/Users/jack/mag/dingo/pkg/preprocessor/preprocessor_test.go`

**Added comprehensive test coverage:**

- `TestCRITICAL2_MultiValueReturnHandling` (lines 519-605)
  - Two values + error: `(int, string, error)`
  - Three values + error: `(string, int, bool, error)`
  - Single value + error (regression test): `(int, error)`

- `TestCRITICAL2_MultiValueReturnWithMessage` (lines 607-636)
  - Multi-value returns with error wrapping

**Updated existing tests to match new counter behavior:**
- `TestErrorPropagationBasic/simple_return` - Updated error variable name from `__err0` to `__err1`

### 3. `/Users/jack/mag/dingo/tests/golden/error_prop_09_multi_value.dingo`

**Created real-world example demonstrating the fix:**
- Function returning `(string, string, int, error)` with error propagation
- Includes both assignment context and return context usage
- Compiles, runs, and produces correct output

## Test Results

### Unit Tests: ✅ ALL PASS
```
go test ./pkg/preprocessor/... -v
```

**Test coverage:**
- ✅ Two-value returns: `(int, string, error)`
- ✅ Three-value returns: `(string, int, bool, error)`
- ✅ Single-value returns (regression): `(int, error)`
- ✅ Multi-value with error messages
- ✅ All existing tests still pass

### Integration Test: ✅ SUCCESS
```
go run ./cmd/dingo build tests/golden/error_prop_09_multi_value.dingo
go build -o /tmp/test_multi tests/golden/error_prop_09_multi_value.go
/tmp/test_multi
```

**Output:**
```
Name: john
Role: admin
Age: 42
```

## Key Insights

1. **Type information was already available** - The `parseFunctionSignature` function already parsed function return types using `go/ast`, so we didn't need to add any new parsing logic.

2. **Counter increment change** - Because we now increment the counter for each temp variable (not just once per expansion), the error variable numbers changed from `__err0` to `__err1`, `__err2`, etc. This is **correct behavior** and doesn't affect functionality.

3. **Zero value generation** - The existing `funcContext.zeroValues` logic already handled multi-value returns correctly for the error path, so no changes were needed there.

4. **Backward compatibility** - Single-value returns still work exactly as before (just with a different counter value).

## Future Enhancements

None needed - the fix is complete and handles all cases correctly.

## No Limitations

Unlike initially suggested in the task description, we did NOT need to constrain `return expr?` to single values. The complete fix handles any number of return values correctly.

## Files Modified

1. `/Users/jack/mag/dingo/pkg/preprocessor/error_prop.go` - Core fix
2. `/Users/jack/mag/dingo/pkg/preprocessor/preprocessor_test.go` - Test coverage
3. `/Users/jack/mag/dingo/tests/golden/error_prop_09_multi_value.dingo` - Golden test
4. `/Users/jack/mag/dingo/tests/golden/error_prop_09_multi_value.go.golden` - Expected output

## Success Criteria Met

- ✅ No silent data loss
- ✅ All existing tests pass
- ✅ New test cases added
- ✅ End-to-end verification (compiles and runs)
- ✅ Clear documentation of changes
- ✅ No limitations or constraints needed

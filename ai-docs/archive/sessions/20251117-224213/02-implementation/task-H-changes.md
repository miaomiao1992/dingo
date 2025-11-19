# Task H Changes: Multi-Value Return Edge Case Tests

## Summary
Added comprehensive edge case tests for multi-value returns in error propagation to supplement the existing `error_prop_09_multi_value` golden test.

## Files Modified

### `/Users/jack/mag/dingo/pkg/preprocessor/preprocessor_test.go`
**Lines Added**: 798-1054 (257 lines)
**Function**: `TestMultiValueReturnEdgeCases()`

Added comprehensive test suite with 10 test cases covering:

1. **2-value return (baseline case)** - `(T, error)` pattern
   - Verifies standard Go error handling pattern
   - Checks for exactly 1 temporary variable (`__tmp0`)
   - Confirms no extra temporaries are generated

2. **3-value return (verified by golden test)** - `(A, B, error)`
   - Confirms golden test coverage is accurate
   - Verifies 3 temporaries (`__tmp0`, `__tmp1`, `__tmp2`)
   - Checks all 3 zero values in error path

3. **4-value return (extreme case)** - `(string, int, float64, bool, error)`
   - Tests extreme multi-value scenario
   - Verifies 4 temporaries and error variable
   - Confirms correct zero values: `"", 0, 0.0, false`

4. **5-value return (very extreme case)** - `(string, int, []byte, map[string]int, bool, error)`
   - Tests maximum practical multi-value return
   - Verifies 5 temporaries (`__tmp0` through `__tmp4`)
   - Ensures no extra temporaries beyond what's needed

5. **Mixed types** - `(string, int, []byte, error)`
   - Tests type variety handling
   - Verifies zero values: `"", 0, nil`
   - Confirms correct temporary count

6. **Complex types** - `(map[string]string, []int, *Config, error)`
   - Tests reference types and pointers
   - Verifies all zero values are `nil`
   - Checks proper handling of struct pointers

7. **Verify correct number of temporaries (3 values)** - Regression check
   - Ensures exactly 3 temps, no more
   - Negative test: should NOT have `__tmp3`

8. **Verify correct number of temporaries (4 values)** - Regression check
   - Ensures exactly 4 temps, no more
   - Negative test: should NOT have `__tmp4`

9. **All values returned in success path (4 values)** - Critical check
   - Verifies no values are dropped
   - Negative tests for common mistakes:
     - `return __tmp0, nil` (drops 3 values)
     - `return __tmp0, __tmp1, nil` (drops 2 values)
     - `return __tmp0, __tmp1, __tmp2, nil` (drops 1 value)

10. **All zero values in error path (mixed types)** - Critical check
    - Verifies all zero values present in error return
    - Negative tests for incomplete error returns

## Test Execution

```bash
$ go test ./pkg/preprocessor -run '^TestMultiValueReturnEdgeCases$' -v

=== RUN   TestMultiValueReturnEdgeCases
=== RUN   TestMultiValueReturnEdgeCases/2-value_return_(baseline_case)
=== RUN   TestMultiValueReturnEdgeCases/3-value_return_(verified_by_golden_test)
=== RUN   TestMultiValueReturnEdgeCases/4-value_return_(extreme_case)
=== RUN   TestMultiValueReturnEdgeCases/5-value_return_(very_extreme_case)
=== RUN   TestMultiValueReturnEdgeCases/mixed_types_(string,_int,_[]byte,_error)
=== RUN   TestMultiValueReturnEdgeCases/complex_types_(map,_slice,_struct_pointer,_error)
=== RUN   TestMultiValueReturnEdgeCases/verify_correct_number_of_temporaries_(3_non-error_values)
=== RUN   TestMultiValueReturnEdgeCases/verify_correct_number_of_temporaries_(4_non-error_values)
=== RUN   TestMultiValueReturnEdgeCases/all_values_returned_in_success_path_(4_values)
=== RUN   TestMultiValueReturnEdgeCases/all_zero_values_in_error_path_(mixed_types)
--- PASS: TestMultiValueReturnEdgeCases (0.00s)
    --- PASS: TestMultiValueReturnEdgeCases/2-value_return_(baseline_case) (0.00s)
    --- PASS: TestMultiValueReturnEdgeCases/3-value_return_(verified_by_golden_test) (0.00s)
    --- PASS: TestMultiValueReturnEdgeCases/4-value_return_(extreme_case) (0.00s)
    --- PASS: TestMultiValueReturnEdgeCases/5-value_return_(very_extreme_case) (0.00s)
    --- PASS: TestMultiValueReturnEdgeCases/mixed_types_(string,_int,_[]byte,_error) (0.00s)
    --- PASS: TestMultiValueReturnEdgeCases/complex_types_(map,_slice,_struct_pointer,_error) (0.00s)
    --- PASS: TestMultiValueReturnEdgeCases/verify_correct_number_of_temporaries_(3_non-error_values) (0.00s)
    --- PASS: TestMultiValueReturnEdgeCases/verify_correct_number_of_temporaries_(4_non-error_values) (0.00s)
    --- PASS: TestMultiValueReturnEdgeCases/all_values_returned_in_success_path_(4_values) (0.00s)
    --- PASS: TestMultiValueReturnEdgeCases/all_zero_values_in_error_path_(mixed_types) (0.00s)
PASS
ok  	github.com/MadAppGang/dingo/pkg/preprocessor	0.185s
```

## Coverage Analysis

### What These Tests Verify

1. **Correct temporary variable count** - No extra or missing temporaries
2. **All non-error values returned in success path** - No value dropping
3. **All zero values in error path** - Complete error return handling
4. **Type variety** - strings, ints, floats, bools, slices, maps, pointers
5. **Extreme cases** - 4-5 value returns (uncommon but valid)

### Regression Prevention

These tests prevent:
- **CRITICAL-2 Regression**: Multi-value returns being dropped (Issue #2 from code review)
- **Temporary variable miscounting**: Wrong number of `__tmp` variables
- **Incomplete error returns**: Missing zero values in error path
- **Success path value dropping**: Not all temporaries returned

## Integration with Existing Tests

This test suite complements:
- **Golden test**: `error_prop_09_multi_value.dingo` (3-value case)
- **Unit test**: `TestCRITICAL2_MultiValueReturnHandling` (basic 2-3 value cases)

Together, they provide comprehensive coverage of multi-value return error propagation from basic to extreme cases.

## Notes

- All tests use realistic Go type combinations
- Tests verify both positive patterns (shouldContain) and negative patterns (shouldNotContain)
- Each test case includes a description for clarity
- Tests are self-contained with inline function definitions
- No external dependencies required

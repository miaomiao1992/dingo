# Where Guards Implementation - Changes

## Files Modified

### 1. `/Users/jack/mag/dingo/pkg/preprocessor/rust_match.go`

**Function: `splitPatternAndGuard` (lines 343-399)**
- **Added**: Support for `where` keyword in addition to `if`
- **Logic**: Try matching ` where ` first (Swift-style), then fall back to ` if ` (Rust-style)
- **Behavior**: Both syntaxes now work identically

**Function: `generateCase` (lines 609-674)**
- **Added**: Guard transformation logic
- **Wrapping**: When `arm.guard != ""`, wrap case body in `if guard { body }`
- **Indentation**: Guard body indented one additional level (double-tab)
- **Variable hoisting**: Properly handles assignment context within guard

## Implementation Details

### Guard Parsing (Already Existed)

The guard parsing was already implemented in `splitPatternAndGuard`:
- Extracts pattern and guard from pattern arm text
- Validates guard keyword follows complete pattern
- Stores guard condition in `patternArm.guard` field

**Change**: Added `where` keyword support alongside existing `if` support

### Guard Transformation (NEW)

Previously, guards were **parsed but not applied** - they appeared in DINGO_GUARD markers but didn't affect code generation.

**New behavior**:
```go
// Without guard:
case ResultTag_Ok:
    x := __match_0.ok_0
    "positive"

// With guard (x > 0):
case ResultTag_Ok:
    x := __match_0.ok_0
    if x > 0 {
        "positive"
    }
```

### Support for Both Syntaxes

**Rust-style (`if`):**
```dingo
Result_Ok(x) if x > 0 => "positive"
```

**Swift-style (`where`):**
```dingo
Option_Some(x) where x > 100 => "large"
```

Both generate identical Go code with `if` statement wrapping.

## Test Results

### ✅ Working Tests

**`pattern_match_05_guards_basic.dingo`** (Rust `if` syntax)
- 3 functions with guards
- Multiple guards per pattern (fallthrough behavior)
- Guard with function call (`isEven(n)`)
- **Status**: ✅ Transpiles successfully
- **Output**: Guards correctly wrapped in `if` statements

**Simple `where` test** (created for validation)
- Swift-style `where` syntax
- Multiple guards on same variant
- **Status**: ✅ Transpiles successfully
- **Generated code**: Identical structure to `if` syntax

### ❌ Known Limitation: Nested Patterns

**`pattern_match_06_guards_nested.dingo`** - FAILS
- **Reason**: Uses **nested patterns** with guards
- **Example**: `Result_Ok(Option_Some(val)) where val > 0`
- **Error**: `parse error: missing ',' in argument list`
- **Root cause**: Preprocessor doesn't support nested pattern syntax (e.g., `Pattern1(Pattern2(binding))`)

**This is NOT a guard issue** - it's a separate feature (nested/recursive pattern matching) that requires additional preprocessor logic.

### Files 07-08 Status

Did not test these yet - they may have same nested pattern issue as file 06.

## Edge Cases Handled

1. **Multiple guards on same variant**: Creates multiple case clauses (Go switch fallthrough)
2. **Complex guard conditions**: `age >= 18 && age < 65` - passed through as-is
3. **Function calls in guards**: `isEven(n)` - works correctly
4. **Block expressions with guards**: Guards wrap entire block
5. **Variable hoisting with guards**: Assignments work inside guard body

## What's NOT Implemented

1. **Nested patterns**: `Result_Ok(Option_Some(x))` - requires recursive pattern parsing
2. **Guard validation**: No compile-time check that guard references valid bindings
3. **Exhaustiveness with guards**: Doesn't track guard coverage (assumes non-exhaustive)

## Code Quality

- **Preserves existing behavior**: Non-guard patterns unchanged
- **Minimal changes**: Only added guard wrapping logic, no refactoring
- **Consistent indentation**: Maintains tab-based formatting
- **Comment markers**: DINGO_GUARD markers preserved for debugging

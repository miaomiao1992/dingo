# Reasoning: safe_nav_07_method_args

## Purpose
Test method calls with various argument patterns using the safe navigation operator (`?.`).

## Test Coverage
1. **Multiple string arguments**: `user?.format("Hello ", "!")` - two string literals
2. **Mixed argument types**: `user?.process(42, true, "test")` - int, bool, string
3. **Function call arguments**: `user?.transform(getData())` - nested function call
4. **Zero arguments**: `user?.getName()` - no arguments

## Key Design Decisions

### Argument Forwarding
Arguments are preserved exactly as written in the Dingo source:
```go
// Dingo: user?.format("Hello ", "!")
// Generated:
return _user.format("Hello ", "!")
```

### Argument Parsing
The preprocessor must handle:
- String literals with spaces: `"Hello "`
- Multiple arguments separated by commas: `42, true, "test"`
- Nested function calls: `getData()`
- Empty argument lists: `()`

### Code Generation Pattern
For `user?.format("Hello ", "!")`:
```go
func() StringOption {
    if user.IsNone() {
        return StringOption_None()
    }
    _user := user.Unwrap()
    return _user.format("Hello ", "!")  // Arguments forwarded as-is
}()
```

### Balanced Parentheses
The parser must correctly handle nested parentheses in arguments:
- `getData()` inside `transform(getData())` requires depth tracking
- String literals with commas should not be split: `"hello, world"`

## Edge Cases Covered
- Zero arguments: `getName()`
- Multiple arguments: `format("Hello ", "!")`
- Mixed types: `process(42, true, "test")`
- Nested calls: `transform(getData())`
- String literals with special characters

## Complexity: Intermediate
Demonstrates the argument parsing capabilities of the safe navigation preprocessor, ensuring arguments are correctly extracted and forwarded.

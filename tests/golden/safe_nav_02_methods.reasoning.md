# Reasoning: safe_nav_02_methods

## Purpose
Test basic method calls with the safe navigation operator (`?.`) on Option types.

## Test Coverage
1. **Simple method call**: `user?.getName()` - method returning plain type
2. **Option-returning method**: `user?.getEmail()` - method returning Option type

## Key Design Decisions

### Return Type Handling
- Method returning `string`: IIFE returns `string` with empty string default
- Method returning `StringOption`: IIFE returns `StringOption_None()` as default

### Code Generation Pattern
For `user?.getName()`:
```go
func() string {
    if user.IsNone() {
        return ""  // Default for plain types
    }
    _user := user.Unwrap()
    return _user.getName()  // Direct method call
}()
```

For `user?.getEmail()`:
```go
func() StringOption {
    if user.IsNone() {
        return StringOption_None()  // Default for Option types
    }
    _user := user.Unwrap()
    return _user.getEmail()  // Direct method call
}()
```

## Edge Cases Covered
- None case handled by early return
- Method with no arguments
- Different return types (plain vs Option)

## Complexity: Basic
This is the simplest method call scenario - single method on Option type with no arguments.

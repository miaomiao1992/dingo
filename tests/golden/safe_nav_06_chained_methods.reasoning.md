# Reasoning: safe_nav_06_chained_methods

## Purpose
Test chained method calls with the safe navigation operator (`?.`), where each method call potentially returns an Option type.

## Test Coverage
1. **Same-type chaining**: `user?.getName()?.toUpper()` - method returning Option, then method on that Option
2. **Cross-type chaining**: `user?.getProfile()?.getBio()` - method returning different Option type, then method on that type

## Key Design Decisions

### Chain Evaluation
Each method call in the chain is evaluated sequentially with intermediate None checks:
```go
func() StringOption {
    if user.IsNone() {
        return StringOption_None()
    }
    _user := user.Unwrap()

    _tmp0 := _user.getName()  // First method call

    if _tmp0.IsNone() {
        return StringOption_None()  // Early return if None
    }
    _tmp1 := _tmp0.Unwrap()

    return _tmp1.toUpper()  // Second method call
}()
```

### Intermediate Variable Naming
- Unwrapped values: `_user`, `_tmp1`, `_tmp3`
- Option results: `_tmp0`, `_tmp2`
- Naming is sequential across the file (_tmp0, _tmp2, etc.)

### Return Type Determination
The final return type is determined by the last method in the chain:
- `getName()?.toUpper()` → `StringOption` (last method returns StringOption)
- `getProfile()?.getBio()` → `StringOption` (last method returns StringOption)

## Edge Cases Covered
- None at any point in chain → early return with None
- Methods returning Option types
- Different Option types in chain (UserOption → ProfileOption → StringOption)

## Complexity: Intermediate
This demonstrates real-world patterns where methods return Option types and can be chained together safely.

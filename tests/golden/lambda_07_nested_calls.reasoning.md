# Golden Test: Lambda with Nested Function Calls

**Feature**: Lambda expressions with function calls containing multiple arguments
**Issue**: Critical bug in regex body parsing that truncated lambda bodies at first comma
**Status**: Tests balanced delimiter tracking fix

## Problem Statement

### Original Bug
The lambda preprocessor used regex pattern `[^,)]+` to capture lambda bodies:
```go
// OLD: Broken regex
multiParamArrow = regexp.MustCompile(`(^|[^.\w])\(([^)]*)\)\s*(?::\s*([^=>\s]+))?\s*=>\s*(\{[^}]*\}|[^,)]+)`)
//                                                                                                     ^^^^^^^^^^
//                                                                                                     Stops at first comma!
```

This pattern **stops at the first comma or closing paren**, breaking all lambdas with function calls:

```dingo
// Input:
numbers.map((x: int): int => transform(x, 1, 2))

// Extracted body: "transform(x"  ❌ WRONG - truncated at first comma
// Missing: ", 1, 2)"

// Generated (invalid Go):
numbers.map(func(x int) int { return transform(x })  // ❌ Missing args, syntax error
```

### Root Cause
The regex cannot track nested delimiters:
- `transform(x, 1, 2)` has commas **inside parentheses** (depth 1)
- Regex sees comma and stops immediately
- Cannot distinguish between:
  - Comma inside function args: `foo(a, b)` ← keep going
  - Comma ending lambda: `.map(x => y), filter(...)` ← stop here

## Solution: Balanced Delimiter Tracking

Replace regex body capture with intelligent parsing:

```go
// NEW: Extract body using balanced delimiter tracking
func extractBalancedBody(src string, start int) (body string, end int) {
    depth := 0
    for i := start; i < len(src); i++ {
        switch src[i] {
        case '(', '[', '{':
            depth++
        case ')', ']', '}':
            depth--
            if depth < 0 {
                // Hit enclosing delimiter, stop here
                return src[start:i], i
            }
        case ',':
            if depth == 0 {
                // Comma at top level, stop here
                return src[start:i], i
            }
        }
    }
    return src[start:], len(src)
}
```

**How it works**:
- Track nesting depth of parentheses/brackets/braces
- Ignore commas when `depth > 0` (inside nested delimiters)
- Stop when:
  - Comma at `depth == 0` (top-level, ends lambda)
  - Closing delimiter that goes negative (enclosing context)
  - End of string

## Test Cases

### Test 1: Simple Function Call (Multiple Args)
```dingo
result1 := map(numbers, (x: int): int => transform(x, 2, 10))
```

**Challenge**: Comma after `x` is inside `transform()` call
**Expected**: Body = `transform(x, 2, 10)` ✅
**Generated**:
```go
result1 := map(numbers, func(x int) int { return transform(x, 2, 10) })
```

### Test 2: Nested Function Calls
```dingo
result2 := map(numbers, (x: int): int => transform(transform(x, 2, 0), 3, 1))
```

**Challenge**: Two levels of nesting with multiple commas
**Expected**: Body = `transform(transform(x, 2, 0), 3, 1)` ✅
**Nesting depth tracking**:
- `transform(` → depth = 1
- `transform(` → depth = 2
- `, 2, 0)` → depth = 2 (commas ignored)
- `)` → depth = 1
- `, 3, 1)` → depth = 1 (commas ignored)
- `)` → depth = 0 (stop here)

### Test 3: Validation with Context
```dingo
result3 := filter(users, (u: string): bool => validate(u, context))
```

**Challenge**: Comma between parameters
**Expected**: Full validation call preserved

### Test 4: Reduce with Seed
```dingo
result4 := reduce(numbers, (acc: int, x: int): int => combine(acc, x, seed), 0)
```

**Challenge**: Comma after lambda ends expression
**Expected**: Stop at comma before `0`
**Generated**:
```go
result4 := reduce(numbers, func(acc int, x int) int { return combine(acc, x, seed) }, 0)
//                                                                                    ^
//                                                                                    Stop here (depth 0)
```

### Test 5: Method Chains
```dingo
result6 := map(
    filter(data, (s: string): bool => len(s) > 3),
    (s: string): string => strings.Repeat(s, 2),
)
```

**Challenge**: Multiple lambdas, nested calls
**Expected**: Both lambda bodies correctly extracted

## Edge Cases Covered

1. **Single argument**: `x => x * 2` (still works)
2. **Multiple arguments in call**: `transform(x, 1, 2)` ✅
3. **Nested calls**: `transform(transform(x, 2, 0), 3, 1)` ✅
4. **Commas at top level**: `.map(...), filter(...)` (stops correctly) ✅
5. **Block bodies**: `x => { return transform(x, 1, 2) }` ✅
6. **Expression bodies**: `x => transform(x, 1, 2)` ✅

## Expected Go Output

All test cases should produce **valid, compilable Go** with:
- Complete function calls (all arguments preserved)
- Proper closing braces (Issue #3 fix)
- Correct delimiter balancing

## Validation Strategy

1. **Preprocessor test**: Verify body extraction with `extractBalancedBody()`
2. **Golden test**: Compile and run this file
3. **go/parser validation**: Parse generated code (must succeed)

## Success Criteria

✅ All lambdas with function calls compile
✅ No truncated arguments
✅ Proper delimiter balancing
✅ Handles arbitrary nesting depth
✅ Works with both TypeScript (`=>`) and Rust (`|...|`) syntax

---
title: "🪄 Multiline Rust pipe lambda syntax"
category: "Functional Programming"
category_order: 50
subcategory: "Lambda Functions"
test_id: "lambda_04_rust_multiline"
order: 4
complexity: "intermediate"
feature: "🪄 lambdas"
phase: "Phase 6"
status: "implemented"
description: "Demonstrates Rust-style lambda functions with block bodies, complex logic, and multiple statements"
summary: "Multiline Rust pipe lambda syntax"
code_reduction: 60
go_proposal: "21498"
go_proposal_link: "https://github.com/golang/go/issues/21498"
feature_file: "lambdas.md"
related_tests:
  - "lambda_03_rust_basic"
  - "lambda_01_typescript_basic"
  - "lambda_02_typescript_multiline"
---

# Test: Multiline Rust Pipe Lambda Syntax

## Purpose

This test validates Rust-style lambda functions with complex, multi-line bodies. It demonstrates that block-based lambdas preserve all statements, control flow, and return values correctly when transpiled to Go.

## What This Test Validates

### 1. Multi-Statement Lambda
```dingo
complexCalc := |x: int| -> int {
    temp := x * 2
    result := temp + 10
    return result
}
```

**Expected Go Output**:
```go
complexCalc := func(x int) int {
    temp := x * 2
    result := temp + 10
    return result
}
```

**Validates**: Multiple statements within lambda body, explicit return.

### 2. Lambda with Conditional Logic
```dingo
validator := |n: int| -> bool {
    if n < 0 {
        return false
    }
    if n > 100 {
        return false
    }
    return true
}
```

**Expected Go Output**:
```go
validator := func(n int) bool {
    if n < 0 {
        return false
    }
    if n > 100 {
        return false
    }
    return true
}
```

**Validates**: Control flow (if statements) inside lambda, early returns.

### 3. Multi-Parameter with Complex String Returns
```dingo
compare := |a: int, b: int| -> string {
    if a > b {
        return "first is greater"
    }
    if a < b {
        return "second is greater"
    }
    return "equal"
}
```

**Expected Go Output**:
```go
compare := func(a int, b int) string {
    if a > b {
        return "first is greater"
    }
    if a < b {
        return "second is greater"
    }
    return "equal"
}
```

**Validates**: Multiple parameters, string return type, exhaustive conditionals.

### 4. Lambda Returning Struct
```dingo
createPoint := |x: int, y: int| -> struct{ X, Y int } {
    return struct{ X, Y int }{X: x, Y: y}
}
```

**Expected Go Output**:
```go
createPoint := func(x int, y int) struct{ X, Y int } {
    return struct{ X, Y int }{X: x, Y: y}
}
```

**Validates**: Anonymous struct return type, struct literal creation.

### 5. Lambda with Loop
```dingo
sumUpTo := |n: int| -> int {
    sum := 0
    for i := 1; i <= n; i++ {
        sum += i
    }
    return sum
}
```

**Expected Go Output**:
```go
sumUpTo := func(n int) int {
    sum := 0
    for i := 1; i <= n; i++ {
        sum += i
    }
    return sum
}
```

**Validates**: Loop constructs inside lambda, local variable scope.

### 6. Lambda with Nested Conditionals
```dingo
categorize := |age: int| -> string {
    if age < 13 {
        return "child"
    }
    if age < 20 {
        return "teenager"
    }
    if age < 65 {
        return "adult"
    }
    return "senior"
}
```

**Expected Go Output**:
```go
categorize := func(age int) string {
    if age < 13 {
        return "child"
    }
    if age < 20 {
        return "teenager"
    }
    if age < 65 {
        return "adult"
    }
    return "senior"
}
```

**Validates**: Cascading if statements, string returns at multiple points.

### 7. Lambda with String Manipulation
```dingo
transform := |s: string| -> string {
    upper := ""
    for _, c := range s {
        if c >= 'a' && c <= 'z' {
            upper += string(c - 32)
        } else {
            upper += string(c)
        }
    }
    return upper
}
```

**Expected Go Output**:
```go
transform := func(s string) string {
    upper := ""
    for _, c := range s {
        if c >= 'a' && c <= 'z' {
            upper += string(c - 32)
        } else {
            upper += string(c)
        }
    }
    return upper
}
```

**Validates**: Range loops, nested conditionals, string operations, character manipulation.

### 8. Lambda as Argument to Higher-Order Function
```dingo
increment := |x: int| -> int {
    return x + 1
}
result13 := applyTwice(5, increment)
```

**Expected Go Output**:
```go
increment := func(x int) int {
    return x + 1
}
result13 := applyTwice(5, increment)
```

**Validates**: Lambda passed as function argument, higher-order function patterns.

### 9. Lambda with Mixed Operations
```dingo
processor := |x: int, y: int| -> int {
    sum := x + y
    product := x * y
    if sum > product {
        return sum
    }
    return product
}
```

**Expected Go Output**:
```go
processor := func(x int, y int) int {
    sum := x + y
    product := x * y
    if sum > product {
        return sum
    }
    return product
}
```

**Validates**: Multiple variables, arithmetic operations, conditional return.

## Rust-Specific Multiline Features

### Block Body Preservation
- **Already has braces**: `|x| -> int { ... }` - Braces preserved as-is
- **No wrapping needed**: Preprocessor doesn't add `{ return }` around blocks
- **Explicit returns**: User must write `return` statements

### Return Type Required
- **Multiline lambdas**: Usually specify return type with `-> Type`
- **Type inference**: May work if context provides enough information
- **Best practice**: Always annotate return type for readability

### Statement Complexity
- **No limitations**: Any valid Go statement can appear in lambda body
- **Scoping**: Local variables scoped to lambda function
- **Closure support**: Can capture variables from outer scope

## Expected Behavior Notes

1. **Identical to TypeScript**: Rust multiline should generate same Go as TypeScript multiline
   - Different syntax, same semantics
   - Block bodies handled identically

2. **Formatting**: `gofmt` will format the output
   - Indentation normalized
   - Spacing standardized
   - Idiomatic Go style

3. **Type Safety**: All type annotations preserved
   - Parameter types required for multiline (best practice)
   - Return types strongly recommended
   - No `interface{}` fallback

4. **Compilation**: Generated Go must compile and run correctly
   - All test cases should produce expected output
   - No runtime errors

5. **Source Maps**: Should map back to original Dingo source
   - Error messages point to Dingo file
   - Line numbers preserved where possible

## Edge Cases Validated

1. **Complex control flow**: Nested if statements, loops, early returns
2. **Multiple return points**: Different paths through lambda
3. **Local variable scope**: Variables defined inside lambda
4. **Struct returns**: Anonymous and named struct types
5. **String manipulation**: Range loops over strings, character operations
6. **Higher-order functions**: Lambdas as arguments to other functions
7. **Mixed operations**: Arithmetic, comparisons, conditionals combined

## Comparison to TypeScript Multiline Tests

This test mirrors `lambda_02_typescript_multiline.dingo` but uses Rust pipe syntax:

| Feature | TypeScript Syntax | Rust Syntax |
|---------|-------------------|-------------|
| Block body | `(x: int): int => { ... }` | `|x: int| -> int { ... }` |
| Return type | `: int` (after params) | `-> int` (after params) |
| Parameter types | `(x: int, y: int)` | `|x: int, y: int|` |

**Goal**: Demonstrate Rust and TypeScript syntaxes produce identical Go output for complex multiline lambdas.

## Performance Considerations

1. **Zero runtime overhead**: Lambdas transpile to native Go func literals
2. **Closure efficiency**: Same as hand-written Go closures
3. **No boxing**: Types preserved, no `interface{}` conversions
4. **Inlining**: Go compiler can inline simple lambdas

## Testing Strategy

1. **Transpile**: Run `dingo build` on this file
2. **Compile**: Verify generated `.go` file compiles without errors
3. **Execute**: Run compiled binary, verify all output matches comments
4. **Source maps**: Check that error positions map back to `.dingo` file
5. **Comparison**: Generated Go should match TypeScript multiline test output

## Configuration Required

```toml
# dingo.toml
[features]
lambda_style = "rust"
```

This test ONLY runs when Rust style is configured. TypeScript syntax should produce errors if config is set to Rust.

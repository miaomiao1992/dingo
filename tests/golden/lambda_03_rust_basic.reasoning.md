---
title: "🪄 Basic Rust pipe lambda syntax"
category: "Functional Programming"
category_order: 50
subcategory: "Lambda Functions"
test_id: "lambda_03_rust_basic"
order: 3
complexity: "basic"
feature: "🪄 lambdas"
phase: "Phase 6"
status: "implemented"
description: "Demonstrates basic Rust-style lambda function syntax with |param| notation for concise anonymous functions"
summary: "Basic Rust pipe lambda syntax"
code_reduction: 50
go_proposal: "21498"
go_proposal_link: "https://github.com/golang/go/issues/21498"
feature_file: "lambdas.md"
related_tests:
  - "lambda_04_rust_multiline"
  - "lambda_01_typescript_basic"
  - "lambda_02_typescript_multiline"
---

# Test: Basic Rust Pipe Lambda Syntax

## Purpose

This test validates the Rust-style pipe syntax (`|param| expr`) for lambda functions in Dingo. It demonstrates that Rust syntax generates identical Go output to TypeScript arrow syntax, just with different input notation.

## What This Test Validates

### 1. Basic Single Parameter Lambda
```dingo
double := |x| x * 2
```

**Expected Go Output**:
```go
double := func(x) { return x * 2 }
```

**Validates**: Simplest lambda form with Rust pipes, parameter inference.

### 2. Multi-Parameter Lambda
```dingo
add := |x, y| x + y
```

**Expected Go Output**:
```go
add := func(x, y) { return x + y }
```

**Validates**: Multiple parameters separated by commas within pipes.

### 3. Explicit Type Annotations (Single)
```dingo
square := |x: int| x * x
```

**Expected Go Output**:
```go
square := func(x int) int { return x * x }
```

**Validates**: Dingo type annotation syntax (`param: Type`) converted to Go syntax (`param Type`).

### 4. Explicit Type Annotations (Multi)
```dingo
multiply := |x: int, y: int| x * y
```

**Expected Go Output**:
```go
multiply := func(x int, y int) int { return x * y }
```

**Validates**: Type annotations for all parameters.

### 5. String Return Type Inference
```dingo
greet := |name: string| "Hello, " + name
```

**Expected Go Output**:
```go
greet := func(name string) string { return "Hello, " + name }
```

**Validates**: Return type inferred from expression, string concatenation.

### 6. Explicit Return Type
```dingo
isPositive := |n: int| -> bool { n > 0 }
```

**Expected Go Output**:
```go
isPositive := func(n int) bool { return n > 0 }
```

**Validates**: Rust-style `-> Type` return type annotation, block body already has braces.

### 7. Lambda in Function Call Context
```dingo
if filterFunc := |x: int| -> bool { x % 2 == 0 }; filterFunc(n) {
    // ...
}
```

**Expected Go Output**:
```go
if filterFunc := func(x int) bool { return x % 2 == 0 }; filterFunc(n) {
    // ...
}
```

**Validates**: Lambda used inline in conditional, type annotations preserved.

### 8. Lambda as Function Argument
```dingo
result8 := process(10, |x| x + 5)
```

**Expected Go Output**:
```go
result8 := process(10, func(x) { return x + 5 })
```

**Validates**: Lambda passed directly to function, parameter type inferred from function signature.

## Rust-Specific Syntax Features

### Pipe Delimiters
- **Rust**: `|param|` - Clear, unambiguous lambda parameter list
- **No ambiguity**: Pipes never start blocks in Go, so always recognized as lambda

### Return Type Syntax
- **Rust style**: `|x: int| -> bool { ... }`
- **Converted to**: `func(x int) bool { ... }`
- **Arrow operator**: `->` becomes space before type in Go

### Body Handling
- **Expression body**: `|x| x * 2` → Wrapped in `{ return ... }`
- **Block body**: `|x| -> bool { x > 0 }` → Braces preserved, no wrapping

## Expected Behavior Notes

1. **Same Go Output**: Rust syntax should generate **identical** Go code to TypeScript syntax
   - Input: `|x| x * 2` (Rust) vs `x => x * 2` (TypeScript)
   - Output: `func(x) { return x * 2 }` (same)

2. **Type Inference**: Works identically to TypeScript tests
   - Function call context: Type inferred from signature
   - Standalone: May require explicit types (based on strictness setting)

3. **Config Required**: Test should run with `dingo.toml` set to:
   ```toml
   [features]
   lambda_style = "rust"
   ```

4. **Compilation**: Generated Go must compile without errors

5. **No Currying**: Nested lambdas like `|x| |y| x + y` are NOT supported

## Edge Cases Validated

1. **Empty expressions**: Lambdas with minimal bodies
2. **Type mixing**: int, string, bool parameters and returns
3. **Inline usage**: Lambdas in conditionals, function calls
4. **Complex returns**: Struct returns (anonymous types)

## Comparison to TypeScript Tests

This test mirrors `lambda_01_typescript_basic.dingo` but uses Rust pipe syntax:

| Feature | TypeScript Syntax | Rust Syntax |
|---------|-------------------|-------------|
| Single param | `x => x * 2` | `|x| x * 2` |
| Multi param | `(x, y) => x + y` | `|x, y| x + y` |
| Explicit type | `(x: int) => x * 2` | `|x: int| x * 2` |
| Return type | `(x: int): bool => ...` | `|x: int| -> bool ...` |

**Goal**: Prove both syntaxes are functionally equivalent with different notation.

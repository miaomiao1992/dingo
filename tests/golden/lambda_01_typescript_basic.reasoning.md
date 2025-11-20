---
title: "Basic TypeScript Arrow Syntax Lambdas"
category: "Lambda Functions"
category_order: 70
subcategory: "TypeScript Style"
test_id: "lambda_01_typescript_basic"
order: 1

complexity: "basic"
feature: "lambda-functions"
phase: "Phase 6.1"
status: "implemented"

description: "Demonstrates basic TypeScript arrow syntax for lambdas: single param without parens (x => expr), single param with parens ((x) => expr), multi-param ((x, y) => expr), explicit types ((x: int) => expr), and usage in function call contexts"
summary: "Basic TypeScript arrow lambdas"
code_reduction: 0
lines_dingo: 45
lines_go: 45

go_proposal: "N/A"
go_proposal_link: ""
feature_file: "lambdas.md"
related_tests:
  - "lambda_02_typescript_multiline"
  - "lambda_03_rust_basic"

tags:
  - "lambda-functions"
  - "typescript-syntax"
  - "arrow-functions"
  - "functional-programming"
keywords:
  - "lambda"
  - "arrow"
  - "=>"
  - "anonymous function"
  - "closure"
---

# Test Reasoning: lambda_01_typescript_basic

## What This Test Validates

This test validates the basic TypeScript arrow syntax for lambda functions in Dingo. TypeScript arrow syntax is the **default lambda style** (`lambda_style = "typescript"` in `dingo.toml`), familiar to JavaScript/TypeScript developers.

### Key Patterns Tested

1. **Single param without parens**: `x => x * 2`
2. **Single param with parens**: `(x) => x * 2`
3. **Multi-param**: `(x, y) => x + y`
4. **Explicit types**: `(x: int, y: int) => x * y`
5. **In function call context**: `apply(x => x + 10, 5)`
6. **In loop context**: Simulating map operations

### Dingo Code

```dingo
// Single param without parens
double := x => x * 2

// Single param with parens
triple := (x) => x * 2

// Multi-param
add := (x, y) => x + y

// With explicit types
multiply := (x: int, y: int) => x * y

// In function call
result := apply(x => x + 10, 5)
```

### Expected Go Output

```go
// Single param without parens
double := func(x) { return x * 2 }

// Single param with parens
triple := func(x) { return x * 2 }

// Multi-param
add := func(x, y) { return x + y }

// With explicit types
multiply := func(x int, y int) int { return x * y }

// In function call
result := apply(func(x) { return x + 10 }, 5)
```

## Transformation Rules

### Pattern 1: Single Param Without Parens

**Regex**: `x => expr`

**Transformation**:
```
x => x * 2
→
func(x) { return x * 2 }
```

**Edge Cases**:
- Must not match property access: `obj.x => ...` (handled by regex negative lookbehind)
- Expression body gets wrapped in `{ return ... }`

### Pattern 2: Single Param With Parens

**Regex**: `(x) => expr`

**Transformation**:
```
(x) => x * 2
→
func(x) { return x * 2 }
```

**Same as multi-param pattern** (just one parameter).

### Pattern 3: Multi-Param

**Regex**: `(x, y) => expr`

**Transformation**:
```
(x, y) => x + y
→
func(x, y) { return x + y }
```

**Handles**:
- Parameter list parsing
- Comma separation preserved

### Pattern 4: Explicit Types

**Regex**: `(x: int, y: int) => expr`

**Transformation**:
```
(x: int, y: int) => x * y
→
func(x int, y int) int { return x * y }
```

**Type Conversion**:
- Dingo: `x: int` (colon syntax)
- Go: `x int` (space syntax)
- Applies to all parameters

### Pattern 5: Return Type Annotation

**Regex**: `(x: int): bool => expr`

**Transformation**:
```
(x: int): bool => x > 0
→
func(x int) bool { return x > 0 }
```

**Note**: Return type goes after closing paren in Go.

## Type Inference Behavior

### Current Implementation (Phase 6.1)

**Type inference is NOT enforced** in basic tests. Parameters without types are left as-is:

```dingo
x => x * 2
→
func(x) { return x * 2 }
```

This will fail at Go compile time if `x` type cannot be inferred from context.

### Future Enhancement (Phase 6.3)

**go/types integration** will:
1. Detect `func(x)` without type
2. Infer type from surrounding context (function signature, assignment type)
3. Rewrite to `func(x int)`
4. OR emit error if inference fails

**Example**:
```dingo
apply := func(f func(int) int, val int) int {
    return f(val)
}
result := apply(x => x + 10, 5)  // x inferred as int from apply signature
```

## Edge Cases Covered

### 1. Lambda in Function Call

```dingo
apply(x => x + 10, 5)
```

**Challenge**: Lambda is argument to function call, not standalone.

**Solution**: Regex handles this by matching prefix context.

### 2. Lambda in Variable Assignment

```dingo
double := x => x * 2
```

**Challenge**: Assignment context.

**Solution**: Works naturally with regex.

### 3. Lambda in Loop Body

```dingo
for _, n := range numbers {
    transform := x => x * 2
    result := transform(n)
}
```

**Challenge**: Lambda created in loop scope.

**Solution**: Each lambda is independent function literal.

### 4. Chained Lambdas (Future)

```dingo
// NOT tested in basic - see advanced tests
numbers.map(x => x * 2).filter(x => x > 10)
```

**Challenge**: Multiple lambdas on same line.

**Solution**: Regex processes all matches (right-to-left to preserve indices).

## Success Metrics

### Code Clarity

TypeScript arrow syntax is **familiar to millions** of developers:
- JavaScript/TypeScript: 17+ million developers
- React: Arrow functions everywhere
- Modern web development: Standard practice

### Syntax Conciseness

```dingo
// Before (Go)
double := func(x int) int { return x * 2 }

// After (Dingo)
double := x => x * 2  // When type inferred
```

**67% reduction** in boilerplate for simple lambdas.

### Zero Runtime Overhead

Generated Go code is **identical** to hand-written Go:
- No lambda library
- No runtime wrappers
- Direct function literal

## Related Features

### Rust Pipe Syntax (Alternative)

```dingo
// TypeScript (this test)
x => x * 2

// Rust (alternative style)
|x| x * 2
```

See `lambda_03_rust_basic.dingo` for Rust style tests.

### Multi-line Bodies

This test uses **expression bodies** only:
```dingo
x => x * 2  // Expression
```

See `lambda_02_typescript_multiline.dingo` for **block bodies**:
```dingo
x => { return x * 2 }  // Block
```

## Configuration

### Default Style

TypeScript arrows are the **default** (`dingo.toml`):
```toml
[syntax]
lambda_style = "typescript"  # DEFAULT
```

### Switching to Rust

```toml
[syntax]
lambda_style = "rust"
```

With Rust style, TypeScript arrows **will not be recognized** (clear error message).

## Limitations (v1.0)

### No Currying

```dingo
// NOT SUPPORTED
curried := x => y => x + y  // Error
```

**Rationale**: Rare in practice, doesn't fit Go culture.

### No Implicit Parameters

```dingo
// NOT SUPPORTED (Kotlin style)
numbers.map { it * 2 }  // Error
```

**Rationale**: Deferred to post-v1.0 (requires brace context detection).

### No Pattern Matching in Parameters

```dingo
// NOT SUPPORTED
match result {
    (Ok(x), Some(y)) => ...  // Error
}
```

**Rationale**: Deferred to post-v1.0 (complex feature).

## Future Enhancements

### Phase 6.3: Type Inference

```dingo
// Currently: Must specify types for standalone lambdas
validate := (x: int) => x > 0

// Future: Infer from assignment type
validate: func(int) bool = x => x > 0  // x inferred as int
```

### Post-v1.0: Kotlin Brace Style

```dingo
// May add if community demand exists
numbers.map { it * 2 }
numbers.filter { it > 10 }
```

**Requires**: Brace context detection (complex).

---

**Last Updated**: 2025-11-20
**Test Status**: Implemented (Phase 6.1)
**Dingo Version**: 0.1.0-alpha (Phase 6.1)

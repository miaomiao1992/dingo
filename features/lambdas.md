# Lambda/Arrow Functions

**Priority:** P1 (High - Developer experience improvement)
**Status:** 🔴 Not Started
**Community Demand:** ⭐⭐⭐⭐ (750+ 👍)
**Inspiration:** Kotlin, Swift, JavaScript, Rust

---

## Overview

Concise lambda syntax reduces boilerplate for simple function literals, enabling cleaner functional programming patterns without sacrificing type safety.

## Motivation

### The Problem in Go

```go
// Verbose function literals
users := Filter(users, func(u User) bool {
    return u.Age > 18
})

names := Map(users, func(u User) string {
    return u.Name
})

// Compare to other languages:
// JavaScript: users.filter(u => u.age > 18)
// Kotlin: users.filter { it.age > 18 }
// Rust: users.filter(|u| u.age > 18)
```

**Research Data:**
- Active proposal ongoing
- 750+ upvotes
- "Most requested ergonomic improvement"

---

## Proposed Syntax

### Basic Lambda

```dingo
// Single expression (implicit return)
let add = |a, b| a + b

// Multiple parameters
let greet = |name, age| "Hello ${name}, you are ${age}"

// No parameters
let getRandom = || rand.Int()

// Block body (explicit return)
let process = |x| {
    let result = x * 2
    println("Doubling ${x}")
    return result
}
```

### With Type Annotations

```dingo
// Explicit types (when inference fails)
let parse = |s: string| -> int {
    return parseInt(s)
}

// Or inferred from context
let numbers: []int = strings.map(|s| parseInt(s))
```

### Kotlin-Style Trailing Lambda

```dingo
// When last parameter is a function
users.filter(|u| u.age > 18)
    .map(|u| u.name)
    .forEach(|name| println(name))

// Trailing lambda (move outside parens if last arg)
users.filter { |u| u.age > 18 }

// Implicit 'it' parameter (single param)
users.filter { it.age > 18 }
    .map { it.name }
```

---

## Transpilation Strategy

```dingo
// Dingo source
let add = |a, b| a + b
```

```go
// Transpiled Go
var add = func(a int, b int) int {
    return a + b
}
```

```dingo
// Dingo source with trailing lambda
users.filter { it.age > 18 }
```

```go
// Transpiled Go
users.filter(func(__it User) bool {
    return __it.age > 18
})
```

---

## Inspiration

### Kotlin Lambdas

```kotlin
// Basic lambda
val add = { a: Int, b: Int -> a + b }

// Trailing lambda
list.filter { it > 5 }
    .map { it * 2 }

// Multiple parameters
map.forEach { key, value ->
    println("$key: $value")
}
```

### Rust Closures

```rust
// Basic closure
let add = |a, b| a + b;

// With move semantics
let greeting = move |name| format!("Hello, {}", name);

// Higher-order functions
numbers.iter()
    .filter(|&x| x > 5)
    .map(|x| x * 2)
```

### Swift Closures

```swift
// Trailing closure syntax
names.sorted { $0 < $1 }

// Capturing values
let multiplier = 2
numbers.map { $0 * multiplier }

// Shorthand argument names
users.filter { $0.age > 18 }
```

---

## Benefits

**60-70% code reduction** for simple callbacks:

```dingo
// Before: 3 lines
func(u User) bool {
    return u.Age > 18
}

// After: 1 line
|u| u.Age > 18
```

---

## Implementation Complexity

**Effort:** Medium
**Timeline:** 1-2 weeks

---

## References

- Kotlin Lambdas: https://kotlinlang.org/docs/lambdas.html
- Swift Closures: https://docs.swift.org/swift-book/documentation/the-swift-programming-language/closures/

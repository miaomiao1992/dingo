# Null Coalescing Operator (`??`)

**Priority:** P2 (Medium - Syntax sugar)
**Status:** 🔴 Not Started
**Community Demand:** ⭐⭐⭐
**Inspiration:** Swift, C#, Kotlin

---

## Overview

The `??` operator provides concise default values for `Option<T>` types, eliminating verbose unwrapping code.

## Proposed Syntax

```dingo
// Basic usage
let name = user?.name ?? "Anonymous"

// Chaining
let value = primary ?? secondary ?? tertiary ?? "default"

// With expressions
let port = env.get("PORT")?.parseInt() ?? 8080
```

## Transpilation

```go
// Transpiles to unwrapOr
var name string
if opt.isSet {
    name = *opt.value
} else {
    name = "Anonymous"
}
```

## Implementation Complexity

**Effort:** Low (syntax sugar for unwrapOr)
**Timeline:** 2-3 days

---

## References

- Swift Nil Coalescing: https://docs.swift.org/swift-book/documentation/the-swift-programming-language/basicoperators/

# Functional Utilities

**Priority:** P2 (Medium - Reduces boilerplate)
**Status:** 🔴 Not Started
**Community Demand:** ⭐⭐⭐ (Proposal #68065)
**Inspiration:** Kotlin, Swift, Rust

---

## Overview

Built-in map/filter/reduce functions for slices reduce boilerplate for common collection operations.

## Proposed Syntax

```dingo
// Map
let numbers = [1, 2, 3, 4, 5]
let doubled = numbers.map(|x| x * 2)  // [2, 4, 6, 8, 10]

// Filter
let evens = numbers.filter(|x| x % 2 == 0)  // [2, 4]

// Reduce
let sum = numbers.reduce(0, |acc, x| acc + x)  // 15

// Chaining
let result = users
    .filter(|u| u.age > 18)
    .map(|u| u.name)
    .sorted()
```

## Transpilation

```go
// Transpiled to explicit loops
var doubled []int
for _, x := range numbers {
    doubled = append(doubled, x * 2)
}
```

## Benefits

- ✅ 60% reduction in loop boilerplate
- ✅ Intent is clearer
- ✅ Composable operations

## Implementation Complexity

**Effort:** Low-Medium
**Timeline:** 1 week

---

## References

- Go Proposal #68065: slices.Map and Filter
- Rust Iterators: https://doc.rust-lang.org/book/ch13-02-iterators.html

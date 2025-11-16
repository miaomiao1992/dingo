# Tuples

**Priority:** P2 (Medium - Convenience feature)
**Status:** 🔴 Not Started
**Community Demand:** ⭐⭐⭐ (Proposal #63221)
**Inspiration:** Python, Swift, Rust

---

## Overview

Tuple types provide lightweight, unnamed product types for grouping related values without defining a struct.

## Proposed Syntax

```dingo
// Tuple type
let point: (int, int) = (10, 20)

// Named fields
let person: (name: string, age: int) = (name: "Alice", age: 30)

// Destructuring
let (x, y) = point
let (name, age) = person

// Function returns
func divmod(a: int, b: int) -> (quotient: int, remainder: int) {
    return (a / b, a % b)
}

let (q, r) = divmod(17, 5)
```

## Transpilation

```go
// Transpiles to anonymous structs
type __Tuple_int_int struct {
    f0 int
    f1 int
}

var point = __Tuple_int_int{10, 20}
```

## Benefits

- ✅ Avoids single-use struct definitions
- ✅ Natural for multi-return values
- ✅ Pattern matching support

## Implementation Complexity

**Effort:** Medium
**Timeline:** 1-2 weeks

---

## References

- Swift Tuples: https://docs.swift.org/swift-book/documentation/the-swift-programming-language/thebasics/
- Rust Tuples: https://doc.rust-lang.org/book/ch03-02-data-types.html#the-tuple-type

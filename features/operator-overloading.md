# Operator Overloading

**Priority:** P4 (Lowest - Domain-specific feature)
**Status:** 🔴 Not Started
**Complexity:** 🟡 Medium (2 weeks implementation)
**Community Demand:** ⭐⭐ (Go Proposal #60612 - valuable for math/DSL domains)
**Inspiration:** Rust, C++, Swift, Kotlin

---

## Overview

Operator overloading allows custom types to define behavior for arithmetic and comparison operators, enabling natural mathematical notation for domain-specific types.

## Motivation

### The Problem in Go

```go
// Matrix math is verbose
type Matrix struct { ... }

result := matrixA.Add(matrixB.Multiply(matrixC))

// Compare to mathematical notation: A + B * C

// BigDecimal arithmetic
price := new(big.Decimal).Mul(quantity, new(big.Decimal).SetString("19.99"))

// Compare to: price := quantity * 19.99
```

**Go Team's Reasoning:**
- "Magic - hides function calls"
- Reduces readability (what does `+` do?)
- Can be abused (C++ `<<` for I/O)

**Use Cases Where It Shines:**
- Matrix/vector math
- Complex numbers
- Rational numbers, BigDecimal
- DSLs for physics, finance, graphics

---

## Why Dingo Can Implement It

**Key Insight:** Transpile to explicit method calls
- `a + b` → `a.Add(b)`
- Generated Go code is explicit (no magic)
- Only Dingo source uses operators

---

## Proposed Syntax

```dingo
// Define operators via trait implementation
impl Matrix: Add {
    func +(other: Matrix) -> Matrix {
        return this.add(other)
    }
}

impl Matrix: Multiply {
    func *(other: Matrix) -> Matrix {
        return this.multiply(other)
    }
}

// Usage (natural mathematical notation)
let result = matrixA + matrixB * matrixC

// Transpiles to:
// result := matrixA.Add(matrixB.Multiply(matrixC))
```

### Supported Operators

```dingo
trait Add { func +(Self, Self) -> Self }
trait Subtract { func -(Self, Self) -> Self }
trait Multiply { func *(Self, Self) -> Self }
trait Divide { func /(Self, Self) -> Self }
trait Equals { func ==(Self, Self) -> bool }
trait Compare { func <(Self, Self) -> bool }
```

---

## Transpilation Strategy

```dingo
// Dingo source
let result = a + b * c
```

```go
// Transpiled Go
result := a.Add(b.Multiply(c))
```

**Simple AST rewrite:**
- BinaryExpr(`+`) → MethodCall("Add")
- Preserves precedence and associativity

---

## Complexity Analysis

**Implementation Complexity:** 🟡 Medium

### Trait Definitions (2-3 days)
- Define operator traits
- Map operators to method names

### Type Checking (4-5 days)
- Verify types implement required traits
- Handle precedence, associativity
- Error messages for unsupported operations

### Transpilation (3-4 days)
- Rewrite operators to method calls
- Handle chaining, precedence
- Tests

**Total: ~2 weeks**

---

## Restricted Scope (Prevent Abuse)

**Only allow for:**
- Math types (Matrix, Complex, Rational, BigDecimal)
- Explicitly marked types (opt-in via trait)

**Disallow for:**
- I/O operations (no `<<` for streams)
- Side effects (operators should be pure)
- String concatenation (use existing `+`)

---

## Benefits & Tradeoffs

**Advantages:**
- ✅ Natural notation for math-heavy code
- ✅ Essential for DSLs (physics, finance, graphics)
- ✅ Transpiles to explicit method calls (Go code is clear)
- ✅ Opt-in (only types that impl traits)

**Concerns:**
- ❓ Can obscure what's happening
  - *Mitigation:* Restrict to math types, generated Go is explicit
- ❓ Precedence/associativity errors
  - *Mitigation:* Follow standard math rules

---

## Examples

### Matrix Math

```dingo
let result = (A + B) * C - D.transpose()

// Transpiles to:
// result := (A.Add(B)).Multiply(C).Subtract(D.Transpose())
```

### Complex Numbers

```dingo
let c1 = Complex{real: 3, imag: 4}
let c2 = Complex{real: 1, imag: 2}
let sum = c1 + c2  // {4, 6}
```

---

## References

- Go Proposal #60612: Operator overloading (rejected)
- Rust Operator Traits: https://doc.rust-lang.org/std/ops/
- Swift Custom Operators: https://docs.swift.org/swift-book/documentation/the-swift-programming-language/advancedoperators/

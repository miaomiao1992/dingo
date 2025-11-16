# Function Overloading

**Priority:** P4 (Lowest - Advanced feature)
**Status:** 🔴 Not Started
**Complexity:** 🟠 High (3 weeks implementation)
**Community Demand:** ⭐⭐ (Niche but valuable for certain patterns)
**Inspiration:** Java, C++, Kotlin, C#

---

## Overview

Function overloading allows multiple functions with the same name but different parameter types, enabling type-specific implementations while maintaining a unified interface.

## Motivation

### The Problem in Go

```go
// Must use different names
func PrintInt(x int) { fmt.Println(x) }
func PrintString(x string) { fmt.Println(x) }
func PrintFloat(x float64) { fmt.Println(x) }

// Or use interface{} and type switches (loses type safety)
func Print(x interface{}) {
    switch v := x.(type) {
    case int:
        fmt.Println(v)
    case string:
        fmt.Println(v)
    default:
        panic("unsupported type")
    }
}
```

**Go Team's Reasoning:**
- "Adds complexity to name resolution"
- Prefer explicit naming for clarity
- Generics cover many use cases

---

## Why Dingo Can Implement It

**Key Insight:** Transpile with name mangling
- `Print(int)` → `Print_int`
- `Print(string)` → `Print_string`
- Generated Go has distinct names (no ambiguity)
- Dingo compiler does the name resolution

---

## Proposed Syntax

```dingo
// Multiple implementations
func format(x: int) -> string {
    return x.toString()
}

func format(x: float64) -> string {
    return fmt.Sprintf("%.2f", x)
}

func format(x: User) -> string {
    return x.name
}

// Usage (Dingo resolves which to call)
let s1 = format(42)           // Calls format(int)
let s2 = format(3.14)         // Calls format(float64)
let s3 = format(user)         // Calls format(User)
```

---

## Transpilation Strategy

```dingo
// Dingo source
func process(x: int) -> int { return x * 2 }
func process(x: string) -> string { return x + "!" }
```

```go
// Transpiled Go
func process_int(x int) int { return x * 2 }
func process_string(x string) string { return x + "!" }
```

**Name Mangling Rules:**
- `functionName_type1_type2_...`
- Handles primitives, structs, pointers
- Deterministic encoding

---

## Complexity Analysis

**Implementation Complexity:** 🟠 High

### Type Resolution (1 week)
- Collect all overloads for a name
- Find best match based on argument types
- Handle ambiguous cases (error)
- Precedence rules

### Name Mangling (3-4 days)
- Encode types into names
- Handle generics, pointers, slices
- Ensure no collisions

### Transpilation (3-4 days)
- Generate mangled Go functions
- Rewrite call sites
- Tests

**Total: ~3 weeks**

---

## Benefits & Tradeoffs

**Advantages:**
- ✅ Cleaner API (one name, multiple types)
- ✅ Works with generics for specialization
- ✅ Transpiles cleanly (name mangling)

**Concerns:**
- ❓ Complex type resolution
- ❓ Debugging harder (mangled names in stack traces)
- ❓ Interaction with defaults, generics

---

## References

- Java Method Overloading
- C++ Function Overloading
- Kotlin Overloading: https://kotlinlang.org/docs/functions.html#function-overloading

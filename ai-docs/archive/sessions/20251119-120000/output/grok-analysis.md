# Grok Code Fast Analysis: Enum Variant Naming Convention

## Executive Summary

Based on Go idioms and Dingo's design principles, **Option B (Pure CamelCase)** is strongly recommended for enum variant constructors and tags. The current underscore convention (Option A) creates code that feels neither Go-native nor fully Rust-like, violating the key principle of generating "hand-written" Go code that developers expect.

## Core Recommendation: Switch to Pure CamelCase

### Constructors: `ValueInt(int)` vs `Value_Int(int)`
The CamelCase approach aligns perfectly with Go's exported function naming conventions (`http.MethodGet`, `ast.BadDecl`) and provides clearer, more idiomatic code.

### Tags: `ValueTagInt` vs `ValueTag_Int`
Follows Go's standard library patterns like `token.IDENT`, `token.ILLEGAL` but adapted for structured naming.

### Pattern Matching: `ValueInt(n)` in match expressions
Works seamlessly and reads naturally.

## Detailed Rationale

### 1. **Go Idiomaticity** (Most Critical)
Go code overwhelmingly uses CamelCase for exported identifiers. The underscore variant breaks this pattern unnaturally:

```go
// Feels Go-native
ValueInt(42)
ValueTagInt

match v { ValueInt(n) => ... }

// Feels alien
Value_Int(42) Serialized badly for JSON, unclear tag semantics, Rust syntax confusion
```

### 2. **Developer Expectations**
Go developers (90% of Dingo's audience) expect code that matches Go libraries. Underscore constructors suggest internal/unexported functions, causing confusion.

### 3. **Generated Code Quality**
The design principle emphasizes "Readable Output: Generated Go should look hand-written." CamelCase constructors look like regular exported Go functions.

## Trade-offs Analyzed

### Switching from Current (Option A) to Recommended (Option B)

**Benefits**:
- ✅ **Idiomatic**: Matches Go's CamelCase convention
- ✅ **Clear**: `ValueInt` clearly belongs to `Value` type
- ✅ **Expectable**: Developers instantly recognize as exported functions
- ✅ **Tool Compatible**: Fits goimports, gopls, etc without issues

**Costs**:
- ⚠️ **Migration**: Existing Dingo code using `Value_Int` needs updating
- ⚠️ **Breakage**: Transpiled Go changes - may affect existing consumers
- ⚠️ **Initial Confusion**: Short-term disruption during transition

**Neutral**:
- No impact on performance, functionality, or compilation
- Similar code generation complexity either way

## Alternative Options Rejected

### Option A (Current - Underscore): ❌ REJECT
Breaks Go naming conventions, creates aliasing problems (JSON/Go interoperability), doesn't look Go-native.

### Option C (Namespaced - `Value.Int`): ❌ REJECT
Complex implementation requiring method syntax, changes core enum semantics, not worth complexity for naming.

### Option D (All Lowercase - `value_int`): ❌ REJECT
Suggests unexported/lowercase functions, contradicts type interaction where `Value` type is exported.

## Implementation Guidance

### Migration Strategy
1. **Phase 1**: Add CamelCase constructors alongside old ones (backward compatible)
2. **Phase 2**: Deprecate underscore constructors with go:deprecated comments
3. **Phase 3**: Remove old constructors after grace period

### Code Generation Changes
```go
// Before (Option A)
Value_Int(42)
ValueTag_Int

// After (Option B)
ValueInt(42)     // Pure CamelCase
ValueTagInt      // Tag constant with std lib pattern
```

## Real-World Examples

Go standard library follows pure CamelCase for similar constructs:
- `http.MethodGet`, `http.MethodPost` (clear hierarchy)
- `ast.BadDecl`, `ast.GenDecl` (type hierarchy preserved)
- `token.IDENT`, `token.EOF` (scoped identification)

Dingo should follow similar patterns for generated code to feel "hand-written."

## Edge Cases Addressed

### JSON Serialization
CamelCase avoids ambiguity with field names containing underscores.

### IDE Auto-complete
`ValueInt` surfaces naturally in Go tooling; `Value_Int` may cause hiccups.

### Type Safety
Name collision avoidance through clear type prefix (`ValueInt` ≠ `OtherInt`).

## Conclusion

**Recommendation**: Migrate to Option B (Pure CamelCase) as the most Go-idiomatic approach. The naming convention is fundamental to developer experience - when in doubt, follow Go's established patterns.

The transition will increase long-term maintainability, tool compatibility, and developer adoption while maintaining all existing functionality and performance characteristics.

## Follow-up Questions
1. What other enum properties (methods, serialization) should use similar naming?
2. Should additional suffix/prefix patterns be established for related constructs?
3. How to balance Rust-like ergonomics with Go-native feel for advanced users?
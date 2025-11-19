# Enum Variant Naming Convention Analysis for Dingo

## Project Context

**Dingo** is a meta-language for Go (similar to TypeScript for JavaScript) that:
- Transpiles `.dingo` files to idiomatic `.go` files
- Supports sum types via `enum TypeName { Variant1, Variant2(payload: Type) }`
- Follows Go naming conventions: PascalCase for exported types/functions
- Maintains 100% Go ecosystem compatibility

## Current State

**Current Implementation**: Both type names AND variants use PascalCase
```dingo
enum PaymentStatus { Pending, Completed(orderId: int), Failed(error: string) }
enum HttpMethod { GET, POST, PUT, DELETE }
enum Result { Ok(value: T), Err(error: E) }
```

**Transpiles to**:
```go
type PaymentStatus interface {
    isPaymentStatus()
}
type paymentStatusPending struct{}
func (paymentStatusPending) isPaymentStatus() {}
type paymentStatusCompleted struct{ OrderId int }
func (paymentStatusCompleted) isPaymentStatus() {}
type paymentStatusFailed struct{ Error string }
func (paymentStatusFailed) isPaymentStatus() {}

const (
    PaymentStatusPending   = paymentStatusPending{}
    PaymentStatusCompleted = func() paymentStatusCompleted {
        return paymentStatusCompleted{OrderId: orderId}
    }
    PaymentStatusFailed = func() paymentStatusFailed {
        return paymentStatusFailed{Error: error}
    }
)
```

## Research Questions

1. **Consistency with Go ecosystem**: Should enum variants follow the same Go naming patterns as types (PascalCase)?

2. **Visual distinction**: Are there benefits to distinguishing variants from types visually (e.g., different casing)?

3. **Comparative analysis**: How do other languages handle this (Rust, Swift, Kotlin, TypeScript)?

4. **Readability implications**: What are the usability implications for developers reading code?

5. **Pattern matching readiness**: How does naming affect future pattern matching syntax?

6. **Common pitfalls**: Examples of problematic ambiguities or confusions?

## Analysis Goals

Please provide:
- **Reasoned recommendations** with clear rationale
- **Trade-off analysis** of different approaches
- **Code examples** demonstrating each approach
- **Future-proofing considerations** for pattern matching
- **Developer experience** implications

## Constraints

- Must maintain Go interoperability
- Should support IDE autocomplete well
- Need to minimize confusion for Go developers
- Pattern matching should be intuitive

## Expected Output

Please analyze and provide recommendations covering:
1. **Option 1**: Keep PascalCase for both (current)
2. **Option 2**: Use different casing for variants (e.g., snake_case or lowercase)
3. **Option 3**: Other naming strategies

Include pros/cons of each approach with real-world code examples.
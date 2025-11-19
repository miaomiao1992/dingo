# Enum Variant Naming Convention Validation for Dingo (Go Meta-Language)

## CONTEXT
Dingo is a meta-language for Go that transpiles to idiomatic Go code. We're implementing sum types/enums and need to validate the naming convention for enum variants.

## CURRENT IMPLEMENTATION (Option A - Underscore-based)

### Generated Code Example:
```go
type ValueTag uint8

const (
    ValueTag_Int ValueTag = iota
    ValueTag_String
)

type Value struct {
    tag      ValueTag
    int_0    *int
    string_0 *string
}

func Value_Int(arg0 int) Value {
    return Value{tag: ValueTag_Int, int_0: &arg0}
}

func Value_String(arg0 string) Value {
    return Value{tag: ValueTag_String, string_0: &arg0}
}
```

### Pattern Matching Usage:
```dingo
match v {
    Value_Int(n) if n > 0 => "Positive"
    Value_Int(n) if n < 0 => "Negative"
    Value_Int(_) => "Zero"
    Value_String(s) => "String value"
}
```

### Current Naming Patterns:
1. **Constructors**: `Value_Int(arg0)` - TypeName_VariantName (CamelCase + underscore)
2. **Tag Constants**: `ValueTag_Int` - TypeNameTag_VariantName
3. **Fields**: `int_0`, `string_0` - lowercase_variant + underscore + index
4. **Pattern Matching**: `Value_Int(n)` - matches constructor names

## ALTERNATIVE OPTIONS

### Option B (Pure CamelCase):
```go
func ValueInt(arg0 int) Value    // Constructor
ValueTagInt                      // Tag constant
match v { ValueInt(n) => ... }   // Pattern matching
```

### Option C (Namespaced):
```go
func Value.Int(arg0 int) Value   // Constructor (method-like)
ValueTag.Int                    // Tag constant
match v { Value.Int(n) => ... } // Pattern matching
```

### Option D (Mixed):
```go
func valueInt(arg0 int) Value    // Constructor (unexported)
VALUE_TAG_INT                   // Tag constant (all caps)
match v { valueInt(n) => ... }  // Pattern matching
```

## EVALUATION CRITERIA
1. **Go Idiomaticity**: Does it feel natural to Go developers?
2. **Clarity**: Is it clear what the identifier represents?
3. **Collision Avoidance**: Does it prevent naming conflicts?
4. **Pattern Matching**: Does it work well in match expressions?
5. **Type Safety**: Does it maintain clear type boundaries?
6. **Readability**: Does generated code look hand-written?

## REAL-WORLD GO PATTERNS TO CONSIDER
- `http.MethodGet`, `http.MethodPost` (CamelCase, no underscore)
- `ast.BadDecl`, `ast.GenDecl` (CamelCase type + variant)
- `token.ILLEGAL`, `token.IDENT` (ALL_CAPS for tokens)
- `io.EOF`, `sql.ErrNoRows` (CamelCase)
- **Pattern**: Go uses CamelCase for discriminated values, NO underscores

## DINGO DESIGN PRINCIPLES
From project docs:
- **Zero Runtime Overhead**: Generate clean Go code
- **Full Compatibility**: Interoperate with all Go packages
- **Readable Output**: Generated Go should look hand-written
- **IDE-First**: Maintain gopls feature parity
- **Simplicity**: Only add features that solve real pain points

## QUESTIONS FOR ANALYSIS
1. Which naming convention (A, B, C, or D) is most Go-idiomatic?
2. Are there better alternatives not listed above?
3. What are the trade-offs of each approach?
4. How do other Go libraries handle discriminated unions / sum types?
5. What would Go developers expect when using Dingo-generated code?
6. Does the underscore convention have benefits that outweigh non-idiomaticity?
7. What impact does this have on IDE autocomplete and type inference?

## EXPECTED OUTPUT
Please provide:
1. **Recommendation**: Which naming convention to use (A, B, C, or D, or other)
2. **Rationale**: Why this choice is best for Go developers
3. **Trade-offs**: What we gain and lose with this choice
4. **Migration Path**: How to transition if changing from current (Option A)
5. **Edge Cases**: Any scenarios where the naming might cause issues
6. **Implementation Notes**: Any technical considerations for the transpiler

Focus on Go idiomaticity and developer experience. The generated code should feel like it was written by an experienced Go developer, not by a transpiler.

[claudish] Model: openrouter/sherlock-think-alpha

# Dingo Enum Variant Naming Recommendation

## Final Standard
**Plain PascalCase package-level constants** (e.g., `Circle ShapeTag = iota`).

- **Constants**: `Circle`, `Square` (no `Shape` prefix or underscores).
- **Fields**: Simple names like `radius` or prefixed if collisions (e.g., `circleRadius`).
- **Constructors**: `func Circle(r float64) Shape { ... }`.

## Rationale
- **Go Idiomatic**: Matches stdlib (`time.Monday`, `net.FlagUp`, `http.MethodGet`).
- **Precedents**: Rust (`Ok`/`Err`), Kotlin (`Success`), Swift (`success`), Haskell (`Just`).
- **Pattern Matching**: Clean `match shape { Circle(r) => ... }` or `if shape.tag == Circle`.
- **Go Proposal #19412**: Prefers plain names like `Ok`/`Some`.
- **Usability**: Concise, readable; avoids verbose `ShapeTag_Circle`.

## Dingo vs Go Example
**Dingo**:
```dingo
enum Shape {
    Circle(f64)
    Square(f64)
    Rectangle(f64, f64)
}
```

**Transpiled Go**:
```go
type ShapeTag uint8
const (
    Circle ShapeTag = iota
    Square
    Rectangle
)
type Shape struct {
    tag ShapeTag
    circle *float64
    square *float64
    rectW  *float64
    rectH  *float64
}
func Circle(r float64) Shape {
    return Shape{tag: Circle, circle: &r}
}
```

**Usage**:
```go
if shape.tag == Circle {
    fmt.Println(*shape.circle)
}
```

## Trade-offs
| Approach | Pros | Cons | Go Fit |
|----------|------|------|--------|
| **Recommended: Plain PascalCase** | Concise, stdlib-like, readable | Package-scoped | Excellent |
| Namespaced (`ShapeCircle`) | Scoped, clear | Verbose | Good |
| Current (`ShapeTag_Circle`) | Scoped | Underscores, awkward | Poor |
| Screaming (`SHAPE_CIRCLE`) | C-familiar | Non-Go style | Poor |

## Implementation/Migration
1. Update `pkg/preprocessor/enum.go`: Generate plain constants/fields.
2. Regenerate golden tests (`sum_types_*.go.golden`).
3. Update plugins for pattern matching (recognize new constants).
4. Add constructors in generated code.
5. **Cost**: Medium (preprocessor + tests; pre-1.0 breaking OK).

**Full Analysis**: [ai-docs/sessions/enum-naming-20251119-140000/output/analysis.md](ai-docs/sessions/enum-naming-20251119-140000/output/analysis.md)  
**Summary**: [summary.txt](ai-docs/sessions/enum-naming-20251119-140000/output/summary.txt)

[claudish] Shutting down proxy server...
[claudish] Done


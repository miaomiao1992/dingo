# Action Items - Sum Types Phase 2.5 Code Review

## CRITICAL (Must Fix Before Merge)

### 1. Fix IIFE Type Inference - Returns `interface{}` Instead of Concrete Type
**Priority:** P0 - Blocking
**Time Estimate:** 2 hours
**Location:** `pkg/plugin/builtin/sum_types.go:656-660` (`inferMatchType` method)

**Problem:**
Match expressions always return `interface{}` instead of the actual type, breaking type safety.

**Impact:**
```dingo
area := match shape {
    Circle{r} => 3.14 * r * r,  // Returns float64
}
// Generated: area := func() interface{} { ... }()  // WRONG!
// Cannot assign interface{} to float64
```

**Fix:**
Implement basic type inference:
- Check first arm's body expression type
- For literals: use token kind to infer type (INT → int, FLOAT → float64)
- For binary exprs: infer from operator (arithmetic → float64)
- Fallback to `interface{}` only when truly unknown

---

### 2. Tuple Variants Missing Backing Storage Fields
**Priority:** P0 - Blocking
**Time Estimate:** 1.5 hours
**Location:** `pkg/plugin/builtin/sum_types.go:331-334` (`generateVariantFields` method)

**Problem:**
Code skips fields with `Names == nil`, but tuple variants use unnamed fields. Later destructuring expects fields like `circle_0` but they don't exist.

**Impact:**
```dingo
enum Shape { Circle(float64) }
match shape { Circle(r) => r }
// Generated tries to access shape.circle_0 → compile error: undefined field
```

**Fix:**
Generate synthetic field names for unnamed fields:
```go
if f.Names == nil || len(f.Names) == 0 {
    // Tuple field: generate synthetic name
    fieldName := fmt.Sprintf("%s_%d", variantName, fieldIndex)
    fields = append(fields, &ast.Field{
        Names: []*ast.Ident{{Name: fieldName}},
        Type:  &ast.StarExpr{X: f.Type},
    })
}
```

---

### 3. Debug Mode References Undefined `dingoDebug` Variable
**Priority:** P0 - Blocking  
**Time Estimate:** 1 hour
**Location:** `pkg/plugin/builtin/sum_types.go:889-896` (`generateNilCheck` method)

**Problem:**
When `nil_safety_checks = "debug"`, generated code references `dingoDebug` but this variable is never declared.

**Impact:**
```go
if dingoDebug && shape.circle_radius == nil {  // undefined: dingoDebug
    panic("...")
}
```

**Fix:**
Emit package-level variable when debug mode is enabled:
```go
var dingoDebug = os.Getenv("DINGO_DEBUG") != ""
```
Add to file declarations and ensure `import "os"` is present.

---

## IMPORTANT (Should Fix)

### 4. `inferEnumType` Failure Leaves Invalid AST
**Priority:** P1
**Time Estimate:** 30 minutes
**Location:** `pkg/plugin/builtin/sum_types.go:548-552`

**Problem:**
When enum type cannot be inferred, code logs error and returns, leaving Dingo AST node in tree. Later `go/printer` panics with cryptic error.

**Fix:**
Replace with `ast.BadExpr` to trigger clear compilation error.

---

### 5. Enum Inference Ambiguity with Shared Variant Names
**Priority:** P1
**Time Estimate:** 1 hour
**Location:** `pkg/plugin/builtin/sum_types.go:669-678`

**Problem:**
Different enums with same variant name cause incorrect tag constant generation (uses first match).

**Fix:**
Detect ambiguity and warn/error when multiple enums match.

---

### 6. Match Arm Errors Silently Drop Arms
**Priority:** P1
**Time Estimate:** 30 minutes
**Location:** `pkg/plugin/builtin/sum_types.go:607-613`

**Problem:**
When match arm fails, it's logged but silently dropped from switch, causing runtime panic on "exhaustive" check.

**Fix:**
Abort transformation or insert `ast.BadStmt` to fail compilation clearly.

---

## MINOR

### 7. Constructor Reuses Enum's `TypeParams` Pointer
**Priority:** P2
**Time Estimate:** 15 minutes
**Location:** `pkg/plugin/builtin/sum_types.go:433-436`

**Problem:**
Shared `*ast.FieldList` between type and constructors could cause mutation issues.

**Fix:**
Deep copy type parameters.

---

## Summary

**Total Issues:** 7
- CRITICAL: 3 (must fix before merge)
- IMPORTANT: 3 (should fix soon)
- MINOR: 1 (nice to have)

**Estimated Fix Time:** 4-6 hours (3-4 hours for CRITICAL)

**Next Steps:**
1. Fix CRITICAL issues #1, #2, #3
2. Add unit tests for fixed functionality
3. Re-run code reviews
4. Address IMPORTANT issues in same PR or follow-up

# Action Items: Functional Utilities Implementation
## Session: 20251117-003406 | Iteration: 01
## Generated: 2025-11-17

---

## CRITICAL (Must Fix Before Merge)

### 1. Add Return Type to sum() IIFE
**Priority:** CRITICAL
**File:** `/Users/jack/mag/dingo/pkg/plugin/builtin/functional_utils.go`
**Lines:** 482-516
**Estimated Time:** 30 minutes

**Action:**
Add `Results` field to the `FuncType` in `transformSum`:

```go
Type: &ast.FuncType{
    Params: &ast.FieldList{},
    Results: &ast.FieldList{
        List: []*ast.Field{{Type: resultType}},
    },
},
```

**Verification:** Compile generated code and verify it builds without errors.

---

### 2. Fix sum() Type Hardcoding to Support All Numeric Types
**Priority:** CRITICAL
**File:** `/Users/jack/mag/dingo/pkg/plugin/builtin/functional_utils.go`
**Lines:** 489-504
**Estimated Time:** 1 hour

**Action:**
Replace hardcoded `int` initialization with type inference from slice element type:

```go
// Current (broken):
&ast.AssignStmt{
    Lhs: []ast.Expr{&ast.Ident{Name: resultVar}},
    Tok: token.DEFINE,
    Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}},
}

// Fixed:
&ast.DeclStmt{
    Decl: &ast.GenDecl{
        Tok: token.VAR,
        Specs: []ast.Spec{
            &ast.ValueSpec{
                Names: []*ast.Ident{{Name: resultVar}},
                Type:  elementType, // Infer from slice
            },
        },
    },
}
```

**Verification:** Test with `[]int`, `[]float64`, `[]time.Duration` slices and verify all compile.

---

### 3. Replace Shallow Clone with Deep Clone using astutil.Apply
**Priority:** CRITICAL
**File:** `/Users/jack/mag/dingo/pkg/plugin/builtin/functional_utils.go`
**Lines:** 807-812
**Estimated Time:** 30 minutes

**Action:**
Replace custom `cloneExpr` implementation with proper deep cloning:

```go
import "golang.org/x/tools/go/ast/astutil"

func (p *FunctionalUtilitiesPlugin) cloneExpr(expr ast.Expr) ast.Expr {
    return astutil.Apply(expr, nil, nil).(ast.Expr)
}
```

**Verification:** Run existing tests and verify no AST corruption with complex receiver expressions.

---

## IMPORTANT (Should Fix Before Merge)

### 4. Add Function Arity Validation
**Priority:** IMPORTANT
**File:** `/Users/jack/mag/dingo/pkg/plugin/builtin/functional_utils.go`
**Lines:** transformMap (200-240), transformFilter (260-300), transformReduce (400-450)
**Estimated Time:** 30 minutes

**Action:**
Add parameter count validation in each transform method:

```go
// In transformMap:
if len(fn.Type.Params.List) != 1 {
    p.currentContext.Logger.Warn("map expects function with 1 parameter, got %d", len(fn.Type.Params.List))
    return nil
}

// In transformFilter:
if len(fn.Type.Params.List) != 1 {
    p.currentContext.Logger.Warn("filter expects function with 1 parameter, got %d", len(fn.Type.Params.List))
    return nil
}

// In transformReduce:
if len(fn.Type.Params.List) != 2 {
    p.currentContext.Logger.Warn("reduce expects function with 2 parameters, got %d", len(fn.Type.Params.List))
    return nil
}
```

**Verification:** Test with wrong-arity functions and verify clear error messages.

---

### 5. Improve Error Logging in extractFunctionBody
**Priority:** IMPORTANT
**File:** `/Users/jack/mag/dingo/pkg/plugin/builtin/functional_utils.go`
**Lines:** 775-796
**Estimated Time:** 1 hour

**Action:**
Add explicit logging for all unsupported cases:

```go
func (p *FunctionalUtilitiesPlugin) extractFunctionBody(body *ast.BlockStmt) ast.Expr {
    if body == nil || len(body.List) == 0 {
        p.currentContext.Logger.Debug("function body is empty or nil")
        return nil
    }

    if len(body.List) > 1 {
        p.currentContext.Logger.Debug("function body has multiple statements, cannot inline")
        return nil
    }

    switch stmt := body.List[0].(type) {
    case *ast.ReturnStmt:
        if len(stmt.Results) == 0 {
            p.currentContext.Logger.Debug("empty return statement, cannot inline")
            return nil
        }
        if len(stmt.Results) > 1 {
            p.currentContext.Logger.Debug("multiple return values, cannot inline")
            return nil
        }
        return stmt.Results[0]
    case *ast.ExprStmt:
        return stmt.X
    default:
        p.currentContext.Logger.Debug("unsupported statement type for inlining: %T", stmt)
        return nil
    }
}
```

**Verification:** Check debug logs show clear reasons for transformation failures.

---

### 6. Add Type Inference Validation for map() and reduce()
**Priority:** IMPORTANT
**File:** `/Users/jack/mag/dingo/pkg/plugin/builtin/functional_utils.go`
**Lines:** transformMap (156-168), transformReduce (408-437)
**Estimated Time:** 2 hours

**Action:**
Add validation for type extraction and provide clear errors:

```go
// In transformMap:
if fn.Type.Results == nil || len(fn.Type.Results.List) == 0 {
    p.currentContext.Logger.Warn("map function must have explicit return type")
    return nil
}
resultElemType := fn.Type.Results.List[0].Type
if resultElemType == nil {
    p.currentContext.Logger.Warn("cannot infer result type from function signature")
    return nil
}

// In transformReduce:
if fn.Type.Results == nil || len(fn.Type.Results.List) == 0 {
    p.currentContext.Logger.Warn("reduce function must have explicit return type")
    return nil
}
```

**Verification:** Test with type-inferred lambdas and verify clear error messages instead of nil panics.

---

### 7. Add Compilation Validation to Test Suite
**Priority:** IMPORTANT
**File:** `/Users/jack/mag/dingo/pkg/plugin/builtin/functional_utils_test.go`
**Estimated Time:** 3-4 hours

**Action:**
Add test helper that compiles and type-checks generated code:

```go
import (
    "go/parser"
    "go/token"
    "go/types"
)

func verifyCompiles(t *testing.T, result ast.Node) {
    t.Helper()

    // Convert AST to code
    code := formatNode(result)

    // Wrap in valid package and function
    fullCode := fmt.Sprintf(`
package test
func testFunc() {
    %s
}
`, code)

    // Parse
    fset := token.NewFileSet()
    file, err := parser.ParseFile(fset, "test.go", fullCode, 0)
    if err != nil {
        t.Fatalf("Generated code does not parse: %v\n%s", err, fullCode)
    }

    // Type-check
    config := &types.Config{Importer: importer.Default()}
    _, err = config.Check("test", fset, []*ast.File{file}, nil)
    if err != nil {
        t.Fatalf("Generated code does not type-check: %v\n%s", err, fullCode)
    }
}

// Use in all tests:
func TestMap(t *testing.T) {
    // ... existing test setup
    result := plugin.transformMap(receiver, args)
    verifyCompiles(t, result) // ADD THIS
    // ... existing assertions
}
```

**Verification:** Run test suite and verify all transformations produce compilable code.

---

### 8. Document Type Inference Limitations
**Priority:** IMPORTANT
**File:** `/Users/jack/mag/dingo/pkg/plugin/builtin/functional_utils.go`
**Lines:** Add to transformMap, transformReduce, transformSum docstrings
**Estimated Time:** 30 minutes

**Action:**
Add clear documentation about type requirements:

```go
// transformMap transforms: slice.map(fn) → inline for-range loop
//
// Requirements:
//   - fn must be a function literal with explicit return type
//   - fn must accept exactly 1 parameter matching slice element type
//   - fn must return exactly 1 value
//
// Example:
//   numbers.map(func(x int) int { return x * 2 })
//
// Not supported:
//   numbers.map(func(x int) { return x * 2 })  // Missing return type
//
// TODO: Support type inference using go/types package
func (p *FunctionalUtilitiesPlugin) transformMap(receiver ast.Expr, args []ast.Expr) ast.Node {
    // ... implementation
}
```

**Verification:** Review documentation for clarity and completeness.

---

### 9. Document Plugin Ordering Requirements
**Priority:** IMPORTANT
**File:** `/Users/jack/mag/dingo/pkg/plugin/builtin/builtin.go`
**Lines:** 16-23
**Estimated Time:** 15 minutes

**Action:**
Add comment explaining plugin order dependencies:

```go
// DefaultPlugins returns the standard set of built-in plugins in dependency order.
//
// Plugin execution order matters:
//   1. ResultTypePlugin, OptionTypePlugin - Define core types
//   2. ErrorPropagationPlugin - Uses Result types
//   3. SumTypesPlugin - Independent of other plugins
//   4. FunctionalUtilitiesPlugin - Will use Result/Option when implemented
//
// Note: When implementing mapResult/filterSome in FunctionalUtilitiesPlugin,
// ensure it runs AFTER ResultTypePlugin and OptionTypePlugin.
func DefaultPlugins() []Plugin {
    return []Plugin{
        NewResultTypePlugin(),
        NewOptionTypePlugin(),
        NewErrorPropagationPlugin(),
        NewSumTypesPlugin(),
        NewFunctionalUtilitiesPlugin(),
    }
}
```

**Verification:** Review comment for accuracy.

---

## Summary

**Total Action Items:** 9 (3 critical, 6 important)

**Estimated Total Time:**
- Critical fixes: 2 hours
- Important improvements: 8 hours
- **Total: 10 hours**

**Minimum Required for Merge:**
- Items 1-3 (critical fixes): 2 hours
- Item 7 (compilation tests): 3-4 hours
- **Recommended minimum: 5-6 hours**

**Recommended Approach:**
1. Start with critical fixes (items 1-3) to unblock compilation
2. Add compilation validation tests (item 7) to catch similar issues
3. Add validation and error logging (items 4-6) for robustness
4. Add documentation (items 8-9) for maintainability

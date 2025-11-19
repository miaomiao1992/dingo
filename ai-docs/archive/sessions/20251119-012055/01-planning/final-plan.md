# Phase 4 Priority 2 & 3: Final Implementation Plan

## Executive Summary

This plan implements 4 critical type inference gaps and guard support to complete Phase 4.2 pattern matching enhancements. Based on user decisions, we will proceed **sequentially** (1→2→3→4) with **strict error handling** - compilation fails when go/types.Info is unavailable. Guards will use **strict validation** (compile errors for invalid guards) and **allow outer scope references**.

**Key Characteristics**:
- **Approach**: Sequential implementation with full testing between tasks
- **Duration**: 5-8 days
- **Risk**: Low (thorough validation at each step)
- **go/types**: Required (fail compilation if unavailable)
- **Guard Scope**: Pattern-bound variables + outer scope variables

---

## User Decisions (Locked In)

### 1. Implementation Approach: Sequential
**Rationale**: Minimize risk of integration issues by completing and testing each task before moving to the next.

**Order**:
1. Task 1: Context type helpers (foundation) → Full test suite
2. Task 2: Pattern match scrutinee → Validate with Task 1
3. Task 3: Err() inference → Reuses Task 1 infrastructure
4. Task 4: Guard validation → Independent feature

### 2. go/types Fallback: Fail with Error
**Rationale**: Strict behavior ensures type safety. If go/types.Info is unavailable, compilation stops with clear error message.

**Implementation**:
```go
func (s *TypeInferenceService) InferTypeFromContext(node ast.Node) (types.Type, bool) {
    if s.typesInfo == nil {
        // STRICT: No fallback allowed
        return nil, false
    }
    // ... proceed with inference
}

// In plugins that use inference:
if contextType, ok := typeInference.InferTypeFromContext(noneNode); !ok {
    return nil, fmt.Errorf("type inference failed: go/types.Info not available (required for context inference)")
}
```

**User Impact**: Clear error message explains go/types requirement.

### 3. Guard Validation: Strict (Compile Error)
**Rationale**: Catch guard errors early during compilation rather than at runtime.

**Rules Enforced**:
- Guard must be valid Go boolean expression
- Guard can reference pattern-bound variables
- Guard can reference outer scope variables (see #4)
- Invalid syntax → Compile error with location
- Non-boolean guard → Compile error

### 4. Guard Scope: Allow Outer Scope References
**Rationale**: Real-world use cases require referencing variables from outer scope (e.g., thresholds, constants).

**Allowed Patterns**:
```dingo
// Pattern-bound variable only
match result {
    Ok(x) if x > 0 => "positive",
}

// Outer scope variable
threshold := 100
match result {
    Ok(x) if x > threshold => "high",  // ✅ ALLOWED
}

// Combination
match result {
    Ok(x) if x > threshold && x < maxVal => "in range",  // ✅ ALLOWED
}
```

**Validation**: Guard must parse as valid Go expression. Scope resolution handled by Go compiler in next phase.

---

## Architecture Overview

### Current State

**Existing Infrastructure** ✅:
1. Parent tracking system (`pkg/plugin/plugin.go:379-381`)
2. go/types integration hooks (`TypeInferenceService.SetTypesInfo()`)
3. Context traversal logic (`InferTypeFromContext()` lines 580-639)
4. Type registry for Result/Option types
5. None constant inference framework (5/9 contexts working)

**Missing Implementations** ❌:
1. `findFunctionReturnType()` - Line 645 (TODO)
2. `findAssignmentType()` - Line 651 (TODO)
3. `findVarDeclType()` - Line 657 (TODO)
4. `findCallArgType()` - Line 663 (TODO)
5. Pattern match scrutinee type detection - Line 498 (heuristic only)
6. Err() context inference - Line 286 (uses `interface{}`)
7. Guard validation in if-else transformation - Lines 826, 1009

### Component Diagram

```
┌──────────────────────────────────────────────────────────┐
│ TypeInferenceService                                     │
│ (pkg/plugin/builtin/type_inference.go)                   │
├──────────────────────────────────────────────────────────┤
│ Infrastructure [COMPLETE]:                               │
│ • parentMap: map[ast.Node]ast.Node ✅                    │
│ • typesInfo: *types.Info ✅ (REQUIRED - fail if nil)     │
│ • InferTypeFromContext() [COMPLETE] ✅                   │
│                                                          │
│ This Plan - Implement These:                             │
│ • findFunctionReturnType() → Task 1.1                    │
│ • findAssignmentType() → Task 1.2                        │
│ • findVarDeclType() → Task 1.3                           │
│ • findCallArgType() → Task 1.4                           │
└──────────────────────────────────────────────────────────┘
         │                      │                   │
         │ Used by              │ Used by           │ Used by
         ↓                      ↓                   ↓
┌──────────────────┐  ┌───────────────────┐  ┌─────────────────┐
│ NoneContextPlugin│  │ PatternMatchPlugin│  │ ResultTypePlugin│
│ (Task 1 consumer)│  │ (Task 2 target)   │  │ (Task 3 target) │
│                  │  │                   │  │                 │
│ Uses 4 helpers → │  │ Scrutinee type → │  │ Err() inference │
└──────────────────┘  └───────────────────┘  └─────────────────┘
```

---

## Task 1: Implement 4 Missing Context Type Helpers

**Duration**: 2-3 days
**Priority**: CRITICAL (foundation for Tasks 2-3)
**Location**: `pkg/plugin/builtin/type_inference.go`

### Task 1.1: Implement `findFunctionReturnType()`

**Goal**: Extract return type from function declaration containing a return statement.

**Implementation**:
```go
func (s *TypeInferenceService) findFunctionReturnType(retStmt *ast.ReturnStmt) types.Type {
    // STRICT: Require go/types
    if s.typesInfo == nil {
        return nil // Caller will handle error
    }

    // Walk up parent chain to find *ast.FuncDecl or *ast.FuncLit
    current := ast.Node(retStmt)
    for current != nil {
        parent := s.parentMap[current]
        if parent == nil {
            break
        }

        // Case 1: Named function
        if funcDecl, ok := parent.(*ast.FuncDecl); ok {
            return s.extractReturnTypeFromFuncType(funcDecl.Type, retStmt)
        }

        // Case 2: Anonymous function (lambda)
        if funcLit, ok := parent.(*ast.FuncLit); ok {
            return s.extractReturnTypeFromFuncType(funcLit.Type, retStmt)
        }

        current = parent
    }
    return nil
}

func (s *TypeInferenceService) extractReturnTypeFromFuncType(
    funcType *ast.FuncType,
    retStmt *ast.ReturnStmt,
) types.Type {
    if funcType.Results == nil || len(funcType.Results.List) == 0 {
        return nil
    }

    // Use go/types to get return type
    if tv := s.typesInfo.Types[funcType.Results.List[0].Type]; tv.Type != nil {
        return tv.Type
    }

    return nil
}
```

**Edge Cases**:
- Multiple return values (position matching)
- Named vs unnamed returns
- Lambda/closure returns
- Nested function declarations

**Test Plan** (`type_inference_test.go`):
```go
func TestFindFunctionReturnType(t *testing.T) {
    tests := []struct {
        name     string
        code     string
        expected string // Type name
    }{
        {"simple", "func f() int { return None }", "int"},
        {"option", "func f() Option_int { return None }", "Option_int"},
        {"result", "func f() Result_int_error { return Err(...) }", "Result_int_error"},
        {"lambda", "f := func() Option_string { return None }", "Option_string"},
        {"multi_return", "func f() (int, error) { return 0, None }", "error"},
    }
    // ... implementation
}
```

---

### Task 1.2: Implement `findAssignmentType()`

**Goal**: Determine the type of the variable being assigned to.

**Implementation**:
```go
func (s *TypeInferenceService) findAssignmentType(assign *ast.AssignStmt, targetNode ast.Node) types.Type {
    // STRICT: Require go/types
    if s.typesInfo == nil {
        return nil
    }

    // Find which RHS expression matches targetNode
    rhsIndex := -1
    for i, rhs := range assign.Rhs {
        if containsNode(rhs, targetNode) {
            rhsIndex = i
            break
        }
    }

    if rhsIndex == -1 || rhsIndex >= len(assign.Lhs) {
        return nil
    }

    lhs := assign.Lhs[rhsIndex]

    // Use go/types to get LHS type
    if tv := s.typesInfo.Types[lhs]; tv.Type != nil {
        return tv.Type
    }

    return nil
}

// Helper: Check if AST subtree contains target node
func containsNode(root, target ast.Node) bool {
    found := false
    ast.Inspect(root, func(n ast.Node) bool {
        if n == target {
            found = true
            return false
        }
        return true
    })
    return found
}
```

**Edge Cases**:
- Simple assignment: `x = None`
- Parallel assignment: `x, y = None, Some(5)`
- Struct field: `s.field = None`
- Pointer dereference: `*ptr = None`

**Test Plan**:
```go
func TestFindAssignmentType(t *testing.T) {
    tests := []struct {
        name     string
        code     string
        expected string
    }{
        {"simple", "var x Option_int; x = None", "Option_int"},
        {"parallel", "var x, y Option_int; x, y = None, Some(5)", "Option_int"},
        {"struct_field", "type S struct{f Option_int}; s.f = None", "Option_int"},
        {"pointer", "var p *Option_int; *p = None", "Option_int"},
    }
    // ... implementation
}
```

---

### Task 1.3: Implement `findVarDeclType()`

**Goal**: Extract explicit type from variable declaration.

**Implementation**:
```go
func (s *TypeInferenceService) findVarDeclType(decl *ast.GenDecl, targetNode ast.Node) types.Type {
    // STRICT: Require go/types
    if s.typesInfo == nil {
        return nil
    }

    for _, spec := range decl.Specs {
        valueSpec, ok := spec.(*ast.ValueSpec)
        if !ok {
            continue
        }

        for _, value := range valueSpec.Values {
            if !containsNode(value, targetNode) {
                continue
            }

            // Case 1: Explicit type annotation
            if valueSpec.Type != nil {
                if tv := s.typesInfo.Types[valueSpec.Type]; tv.Type != nil {
                    return tv.Type
                }
            }

            // Case 2: No explicit type (short declaration)
            // go/types should still have inferred the type
            if len(valueSpec.Names) > 0 {
                if tv := s.typesInfo.Types[valueSpec.Names[0]]; tv.Type != nil {
                    return tv.Type
                }
            }

            return nil
        }
    }

    return nil
}
```

**Edge Cases**:
- Explicit type: `var x Option_int = None`
- Implicit type: `var x = None` (requires other context)
- Multi-var: `var x, y = None, None`

**Test Plan**:
```go
func TestFindVarDeclType(t *testing.T) {
    tests := []struct {
        name     string
        code     string
        expected string
    }{
        {"explicit", "var x Option_int = None", "Option_int"},
        {"multi_var", "var x, y Option_int = None, Some(5)", "Option_int"},
    }
    // ... implementation
}
```

---

### Task 1.4: Implement `findCallArgType()`

**Goal**: Determine parameter type for a function call argument.

**Implementation**:
```go
func (s *TypeInferenceService) findCallArgType(call *ast.CallExpr, targetNode ast.Node) types.Type {
    // STRICT: Require go/types
    if s.typesInfo == nil {
        return nil
    }

    // Find argument position
    argIndex := -1
    for i, arg := range call.Args {
        if containsNode(arg, targetNode) {
            argIndex = i
            break
        }
    }

    if argIndex == -1 {
        return nil
    }

    // Get function type from go/types
    if tv := s.typesInfo.Types[call.Fun]; tv.Type != nil {
        if sig, ok := tv.Type.(*types.Signature); ok {
            // Handle variadic functions
            if sig.Variadic() && argIndex >= sig.Params().Len()-1 {
                lastParam := sig.Params().At(sig.Params().Len() - 1)
                if slice, ok := lastParam.Type().(*types.Slice); ok {
                    return slice.Elem()
                }
                return lastParam.Type()
            }

            // Regular parameter
            if argIndex < sig.Params().Len() {
                return sig.Params().At(argIndex).Type()
            }
        }
    }

    return nil
}
```

**Edge Cases**:
- Regular call: `processAge(None)`
- Variadic: `fmt.Printf("%v %v", None, Some(5))`
- Method call: `obj.method(None)`
- Built-in: `append(slice, None)`

**Test Plan**:
```go
func TestFindCallArgType(t *testing.T) {
    tests := []struct {
        name     string
        code     string
        expected string
    }{
        {"regular", "func f(x Option_int){...}; f(None)", "Option_int"},
        {"variadic", "func f(args ...Option_int){...}; f(None, Some(5))", "Option_int"},
        {"method", "type T struct{}; func (t T) m(x Option_int){}; t.m(None)", "Option_int"},
    }
    // ... implementation
}
```

---

### Task 1.5: Add Strict Error Handling

**Implementation**: Update callers to fail when go/types unavailable.

**Example (in `none_context.go`)**:
```go
func (p *NoneContextPlugin) Transform(file *ast.File) error {
    // ... existing code ...

    for _, noneNode := range noneNodes {
        contextType, ok := p.typeInference.InferTypeFromContext(noneNode)
        if !ok {
            // STRICT: Fail compilation
            return fmt.Errorf(
                "cannot infer type for None constant at %s: go/types.Info required but unavailable",
                p.ctx.FileSet.Position(noneNode.Pos()),
            )
        }

        // ... continue transformation
    }
}
```

**User-Facing Error Message**:
```
Error: cannot infer type for None constant at example.dingo:15:12
  go/types.Info required but unavailable

  This is likely a transpiler bug. Please report at:
  https://github.com/dingolang/dingo/issues
```

---

### Task 1 Validation Criteria

**Definition of Done**:
- [ ] All 4 helper functions implemented with go/types.Info requirement
- [ ] `containsNode()` helper added
- [ ] Strict error handling in all callers
- [ ] Unit tests: 20+ test cases covering edge cases
- [ ] Integration test: None inference in all 9 contexts
- [ ] Golden test: `none_inference_comprehensive.dingo` passes
- [ ] Performance: <15ms overhead per file (benchmark verified)

**Success Metric**: None inference coverage 50% → 90%+

---

## Task 2: Pattern Match Scrutinee go/types Integration

**Duration**: 1 day
**Priority**: HIGH (improves pattern match accuracy)
**Depends On**: Task 1 (shares go/types knowledge)
**Location**: `pkg/plugin/builtin/pattern_match.go:498`

### Current Problem

Heuristic-based type detection checks for Ok/Err/Some/None patterns in match arms. Fails for:
- Type aliases: `type MyResult = Result_int_error`
- Function returns: `match getResult() { ... }`
- Complex expressions: `match user.age { ... }`

### Solution: Use go/types

**Implementation**:
```go
func (p *PatternMatchPlugin) getAllVariants(match *matchExpression) []string {
    // NEW: Use go/types to get scrutinee type
    if p.ctx != nil && p.ctx.TypeInfo != nil {
        if typesInfo, ok := p.ctx.TypeInfo.(*types.Info); ok {
            if tv, exists := typesInfo.Types[match.scrutinee]; exists && tv.Type != nil {
                variants := p.extractVariantsFromType(tv.Type)
                if len(variants) > 0 {
                    return variants
                }
            }
        }
    }

    // Fallback to heuristic (existing code)
    return p.getAllVariantsFromPatterns(match)
}

// New function: Extract variants from types.Type
func (p *PatternMatchPlugin) extractVariantsFromType(t types.Type) []string {
    // Strip named type wrapper
    underlying := t.Underlying()

    // Check if it's a struct with "tag" field
    if structType, ok := underlying.(*types.Struct); ok {
        for i := 0; i < structType.NumFields(); i++ {
            field := structType.Field(i)
            if field.Name() == "tag" {
                // Determine type by checking other fields
                typeName := t.String()

                if strings.Contains(typeName, "Result_") {
                    return []string{"Ok", "Err"}
                }
                if strings.Contains(typeName, "Option_") {
                    return []string{"Some", "None"}
                }
            }
        }
    }

    return []string{}
}
```

**Strict Error Handling**:
```go
if len(variants) == 0 {
    return nil, fmt.Errorf(
        "cannot determine match type at %s: scrutinee must be Result or Option type",
        p.ctx.FileSet.Position(match.scrutinee.Pos()),
    )
}
```

### Test Plan

**Unit Tests** (`pattern_match_scrutinee_test.go`):
```go
func TestExtractVariantsFromType(t *testing.T) {
    tests := []struct {
        name     string
        code     string
        expected []string
    }{
        {"result_variable", "var r Result_int_error; match r {...}", []string{"Ok", "Err"}},
        {"option_variable", "var o Option_string; match o {...}", []string{"Some", "None"}},
        {"type_alias", "type R = Result_int_error; var r R; match r {...}", []string{"Ok", "Err"}},
        {"function_return", "func f() Result_int_error {...}; match f() {...}", []string{"Ok", "Err"}},
        {"struct_field", "type S struct{r Result_int_error}; match s.r {...}", []string{"Ok", "Err"}},
    }
    // ... implementation
}
```

**Golden Test** (`pattern_match_type_alias.dingo`):
```dingo
package test

type UserResult = Result_User_error

func getUser() UserResult { ... }

func main() {
    match getUser() {
        Ok(user) => println(user.name),
        Err(e) => println(e),
    }
}
```

### Validation Criteria

**Definition of Done**:
- [ ] `extractVariantsFromType()` implemented
- [ ] `getAllVariants()` uses go/types first, fallback to heuristic
- [ ] Strict error when type cannot be determined
- [ ] Unit tests: 10+ cases (variables, calls, fields, aliases)
- [ ] Golden test: Type alias pattern matching passes
- [ ] Performance: No measurable overhead

**Success Metric**: Pattern match accuracy 85% → 95%+

---

## Task 3: Err() Context-Based Type Inference

**Duration**: 1-2 days
**Priority**: HIGH (improves generated code quality)
**Depends On**: Task 1 (reuses context inference)
**Location**: `pkg/plugin/builtin/result_type.go:286`

### Current Problem

`Err()` uses `interface{}` as placeholder for Ok type:
```go
Err(myError) → Result_interface_error{tag: ResultTagErr, err_0: &myError}
```

This is non-idiomatic and loses type information.

### Solution: Infer Ok Type from Context

**Implementation**:
```go
func (p *ResultTypePlugin) transformErrCall(call *ast.CallExpr) (ast.Expr, error) {
    // Extract error argument
    errorArg := call.Args[0]

    // Infer error type (existing code)
    errType := "error"
    if p.typeInference != nil {
        if inferredType, ok := p.typeInference.InferType(errorArg); ok {
            errType = inferredType
        }
    }

    // NEW: Infer Ok type from context
    okType := "interface{}"
    if p.typeInference != nil {
        contextType, ok := p.typeInference.InferTypeFromContext(call)
        if !ok {
            // STRICT: Fail if context unavailable
            return nil, fmt.Errorf(
                "cannot infer Result type for Err() at %s: context type required",
                p.ctx.FileSet.Position(call.Pos()),
            )
        }

        // Extract T from Result_T_E
        if resultInfo := p.parseResultTypeName(contextType.String()); resultInfo != nil {
            okType = resultInfo.OkType
        } else {
            return nil, fmt.Errorf(
                "Err() must be used in Result context, got %s at %s",
                contextType.String(),
                p.ctx.FileSet.Position(call.Pos()),
            )
        }
    }

    // Generate correct Result type
    resultTypeName := fmt.Sprintf("Result_%s_%s",
        p.sanitizeTypeName(okType),
        p.sanitizeTypeName(errType))

    // ... rest of transformation
}

// Helper: Parse Result type name
type ResultTypeInfo struct {
    OkType  string
    ErrType string
}

func (p *ResultTypePlugin) parseResultTypeName(typeName string) *ResultTypeInfo {
    // Parse "Result_int_error" → {OkType: "int", ErrType: "error"}
    if !strings.HasPrefix(typeName, "Result_") {
        return nil
    }

    parts := strings.Split(typeName[7:], "_")
    if len(parts) < 2 {
        return nil
    }

    return &ResultTypeInfo{
        OkType:  strings.Join(parts[:len(parts)-1], "_"),
        ErrType: parts[len(parts)-1],
    }
}
```

### Integration

**Add to ResultTypePlugin**:
```go
type ResultTypePlugin struct {
    // ... existing fields
    typeInference *TypeInferenceService // NEW
}

func (p *ResultTypePlugin) SetContext(ctx *PluginContext) {
    p.ctx = ctx

    // NEW: Initialize type inference
    if ctx.TypeInfo != nil {
        p.typeInference = NewTypeInferenceService()
        p.typeInference.SetParentMap(ctx.ParentMap)
        if typesInfo, ok := ctx.TypeInfo.(*types.Info); ok {
            p.typeInference.SetTypesInfo(typesInfo)
        }
    }
}
```

### Test Plan

**Unit Tests** (`result_err_context_test.go`):
```go
func TestErrContextInference(t *testing.T) {
    tests := []struct {
        name     string
        code     string
        expected string // Result type
    }{
        {
            "return_context",
            "func f() Result_int_error { return Err(myErr) }",
            "Result_int_error",
        },
        {
            "assignment_context",
            "var r Result_string_error; r = Err(myErr)",
            "Result_string_error",
        },
        {
            "call_context",
            "func h(r Result_bool_error){}; h(Err(myErr))",
            "Result_bool_error",
        },
        {
            "struct_field",
            "type S struct{r Result_User_error}; s := S{r: Err(myErr)}",
            "Result_User_error",
        },
    }
    // ... implementation
}
```

**Golden Test** (`result_err_contexts.dingo`):
```dingo
package test

type User struct {
    name string
}

func getUser() Result_User_error {
    return Err(errors.New("not found"))  // Should infer Ok=User
}

func main() {
    // Assignment
    var result Result_int_error
    result = Err(errors.New("failed"))  // Should infer Ok=int

    // Call argument
    handleResult(Err(errors.New("error")))  // Should infer Ok=bool

    // Struct field
    type Response struct {
        data Result_string_error
    }
    r := Response{data: Err(errors.New("error"))}  // Should infer Ok=string
}

func handleResult(r Result_bool_error) {}
```

### Validation Criteria

**Definition of Done**:
- [ ] `transformErrCall()` uses context inference
- [ ] `parseResultTypeName()` helper implemented
- [ ] TypeInferenceService integrated into ResultTypePlugin
- [ ] Strict error when context unavailable
- [ ] Unit tests: 15+ cases (return, assign, call, struct)
- [ ] Golden test: All 4 contexts generate correct types
- [ ] Code quality: No `Result_interface_error` in output

**Success Metric**: Err() type correctness 0% → 80%+

---

## Task 4: Guard Validation in If-Else Chains

**Duration**: 1-2 days
**Priority**: MEDIUM (completes guard support)
**Depends On**: None (independent feature)
**Locations**:
- `pkg/plugin/builtin/pattern_match.go` (transformation)
- `pkg/plugin/builtin/pattern_match_test.go:826, 1009` (remove TODOs)

### Current State

Guards are parsed but ignored during transformation:
1. Preprocessor inserts: `// DINGO_GUARD: x > 0`
2. Plugin parses comments in `extractMatchInfo()`
3. Transformation ignores guards

### Solution: Validate and Preserve Guards

#### Step 4.1: Parse Guards in Transform Phase

**Implementation**:
```go
func (p *PatternMatchPlugin) transformSwitchToIfElse(
    match *matchExpression,
    switchStmt *ast.SwitchStmt,
) (*ast.IfStmt, error) {
    // ... existing code creates if-else chain

    var rootIf *ast.IfStmt
    var currentIf *ast.IfStmt

    for i, caseClause := range switchStmt.Body.List {
        clause := caseClause.(*ast.CaseClause)

        // Extract pattern info (includes guard if present)
        patternInfo := p.extractPatternInfoFromComment(clause)

        // Build pattern condition (tag check + variable extraction)
        patternCond := p.buildPatternCondition(patternInfo, match.scrutinee)

        // NEW: Add guard if present
        var finalCond ast.Expr = patternCond
        if patternInfo.guard != "" {
            guardExpr, err := p.parseAndValidateGuard(patternInfo.guard, patternInfo.boundVars)
            if err != nil {
                // STRICT: Compile error
                return nil, fmt.Errorf(
                    "invalid guard at %s: %v",
                    p.ctx.FileSet.Position(clause.Pos()),
                    err,
                )
            }

            // Combine: if tag == ResultTagOk && x > 0 { ... }
            finalCond = &ast.BinaryExpr{
                X:  patternCond,
                Op: token.LAND,
                Y:  guardExpr,
            }
        }

        // Create if statement with combined condition
        ifStmt := &ast.IfStmt{
            Cond: finalCond,
            Body: clause.Body,
        }

        // Chain if-else
        if rootIf == nil {
            rootIf = ifStmt
            currentIf = ifStmt
        } else {
            currentIf.Else = ifStmt
            currentIf = ifStmt
        }
    }

    return rootIf, nil
}

// Helper: Parse guard expression string to AST
func (p *PatternMatchPlugin) parseGuardExpr(guardStr string) (ast.Expr, error) {
    expr, err := parser.ParseExpr(guardStr)
    if err != nil {
        return nil, fmt.Errorf("failed to parse guard '%s': %v", guardStr, err)
    }
    return expr, nil
}
```

#### Step 4.2: Validate Guard with Relaxed Scope

**Implementation**:
```go
func (p *PatternMatchPlugin) parseAndValidateGuard(
    guardStr string,
    boundVars []string,
) (ast.Expr, error) {
    // Parse guard expression
    guardExpr, err := p.parseGuardExpr(guardStr)
    if err != nil {
        return nil, err
    }

    // Validate guard is boolean (if go/types available)
    if p.ctx != nil && p.ctx.TypeInfo != nil {
        if typesInfo, ok := p.ctx.TypeInfo.(*types.Info); ok {
            if tv := typesInfo.Types[guardExpr]; tv.Type != nil {
                if !types.Identical(tv.Type, types.Typ[types.Bool]) {
                    return nil, fmt.Errorf(
                        "guard must be boolean expression, got %s",
                        tv.Type,
                    )
                }
            }
        }
    }

    // NEW: Relaxed scope validation (allow outer scope)
    // Only validate that guard references are visible variables
    // Go compiler will handle scope resolution in next phase
    usedVars := extractIdentifiers(guardExpr)
    for _, v := range usedVars {
        // Skip built-in functions/constants
        if isBuiltin(v) {
            continue
        }

        // Check if variable is pattern-bound OR in outer scope
        // For MVP, we trust Go compiler will validate scope
        // Just ensure it's not a malformed reference
        if !isValidIdentifier(v) {
            return nil, fmt.Errorf("invalid identifier in guard: %s", v)
        }
    }

    return guardExpr, nil
}

// Helper: Extract all identifiers from expression
func extractIdentifiers(expr ast.Expr) []string {
    var idents []string
    ast.Inspect(expr, func(n ast.Node) bool {
        if id, ok := n.(*ast.Ident); ok {
            idents = append(idents, id.Name)
        }
        return true
    })
    return idents
}

// Helper: Check if identifier is valid Go identifier
func isValidIdentifier(s string) bool {
    if len(s) == 0 {
        return false
    }
    for i, r := range s {
        if i == 0 {
            if !unicode.IsLetter(r) && r != '_' {
                return false
            }
        } else {
            if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
                return false
            }
        }
    }
    return true
}

// Helper: Check if identifier is Go built-in
func isBuiltin(s string) bool {
    builtins := map[string]bool{
        "true": true, "false": true, "nil": true,
        "len": true, "cap": true, "append": true,
        // ... add more as needed
    }
    return builtins[s]
}
```

#### Step 4.3: Update Tests

**Remove TODOs and add validation** (`pattern_match_test.go`):
```go
// Line 826: TestPatternMatchPlugin_TransformWithGuard
func TestPatternMatchPlugin_TransformWithGuard(t *testing.T) {
    src := `
package test

func example(result Result_int_error) string {
    match result {
        Ok(x) if x > 0 => "positive",
        Ok(x) => "zero or negative",
        Err(e) => e.Error(),
    }
}
`
    // Parse and transform
    file, err := parser.ParseFile(fset, "test.dingo", src, parser.ParseComments)
    if err != nil {
        t.Fatal(err)
    }

    // ... run plugin transformation

    // NEW: Verify guard is preserved in generated code
    var foundGuard bool
    ast.Inspect(file, func(n ast.Node) bool {
        if binExpr, ok := n.(*ast.BinaryExpr); ok {
            // Check for: x > 0
            if binExpr.Op == token.GTR {
                if ident, ok := binExpr.X.(*ast.Ident); ok && ident.Name == "x" {
                    if lit, ok := binExpr.Y.(*ast.BasicLit); ok && lit.Value == "0" {
                        foundGuard = true
                    }
                }
            }
        }
        return true
    })

    if !foundGuard {
        t.Errorf("expected guard (x > 0) in transformed code")
    }
}

// Line 1009: TestPatternMatchPlugin_GuardWithOuterScope
func TestPatternMatchPlugin_GuardWithOuterScope(t *testing.T) {
    src := `
package test

func example(result Result_int_error) string {
    threshold := 100
    match result {
        Ok(x) if x > threshold => "high",  // Guard references outer scope
        Ok(x) => "low",
        Err(e) => e.Error(),
    }
}
`
    // Parse and transform
    file, err := parser.ParseFile(fset, "test.dingo", src, parser.ParseComments)
    if err != nil {
        t.Fatal(err)
    }

    // ... run plugin transformation

    // NEW: Verify guard references outer scope variable
    var foundOuterRef bool
    ast.Inspect(file, func(n ast.Node) bool {
        if binExpr, ok := n.(*ast.BinaryExpr); ok {
            if binExpr.Op == token.GTR {
                if ident, ok := binExpr.Y.(*ast.Ident); ok && ident.Name == "threshold" {
                    foundOuterRef = true
                }
            }
        }
        return true
    })

    if !foundOuterRef {
        t.Errorf("expected guard to reference outer scope variable 'threshold'")
    }
}
```

### Golden Test

**File**: `tests/golden/pattern_guards_complete.dingo`
```dingo
package test

func classifyValue(result Result_int_error, threshold int) string {
    match result {
        Ok(x) if x > threshold => "high",
        Ok(x) if x == 0 => "zero",
        Ok(x) => "low",
        Err(e) => e.Error(),
    }
}

func main() {
    classifyValue(Ok(150), 100)  // "high"
    classifyValue(Ok(50), 100)   // "low"
    classifyValue(Ok(0), 100)    // "zero"
}
```

**Expected Output** (`.go.golden`):
```go
func classifyValue(result Result_int_error, threshold int) string {
    if result.tag == ResultTagOk && *result.ok_0 > threshold {
        x := *result.ok_0
        return "high"
    } else if result.tag == ResultTagOk && *result.ok_0 == 0 {
        x := *result.ok_0
        return "zero"
    } else if result.tag == ResultTagOk {
        x := *result.ok_0
        return "low"
    } else if result.tag == ResultTagErr {
        e := *result.err_0
        return e.Error()
    }
    panic("non-exhaustive match")
}
```

### Validation Criteria

**Definition of Done**:
- [ ] `parseAndValidateGuard()` implemented with relaxed scope
- [ ] Guard combined with pattern condition in if-else chain
- [ ] Strict validation: boolean type check (if go/types available)
- [ ] Relaxed validation: Allow outer scope references
- [ ] Tests updated: Remove TODOs from lines 826, 1009
- [ ] Golden test: Guards preserved and combined correctly
- [ ] Error messages: Clear, actionable for invalid guards

**Success Metric**: Guard test pass rate 0% → 100% (2 tests)

---

## Implementation Timeline (Sequential)

### Week 1: Foundation

**Monday-Tuesday** (2 days):
- Implement Task 1.1: `findFunctionReturnType()`
- Implement Task 1.2: `findAssignmentType()`
- Add `containsNode()` helper
- Unit tests for 1.1 and 1.2

**Wednesday** (1 day):
- Implement Task 1.3: `findVarDeclType()`
- Implement Task 1.4: `findCallArgType()`
- Unit tests for 1.3 and 1.4

**Thursday** (1 day):
- Task 1.5: Add strict error handling to all callers
- Integration tests: None inference in all 9 contexts
- Golden test: `none_inference_comprehensive.dingo`
- Performance benchmarking

**Friday** (1 day):
- **VALIDATION CHECKPOINT**
- Run full test suite
- Verify None inference coverage >90%
- Fix any failing tests
- Document Task 1 completion

---

### Week 2: Advanced Features

**Monday** (1 day):
- Implement Task 2: Pattern match scrutinee go/types integration
- `extractVariantsFromType()` function
- Update `getAllVariants()` with go/types first approach
- Unit tests for Task 2
- **VALIDATION**: Pattern match with type aliases

**Tuesday-Wednesday** (2 days):
- Implement Task 3: Err() context inference
- Update `transformErrCall()` to use context
- Add `parseResultTypeName()` helper
- Integrate TypeInferenceService into ResultTypePlugin
- Unit tests for Task 3
- Golden test: `result_err_contexts.dingo`
- **VALIDATION**: No `Result_interface_error` in output

**Thursday-Friday** (2 days):
- Implement Task 4: Guard validation
- `parseAndValidateGuard()` with relaxed scope
- Update `transformSwitchToIfElse()` to preserve guards
- Remove TODOs from tests (lines 826, 1009)
- Golden test: `pattern_guards_complete.dingo`
- **VALIDATION**: Guard tests passing

---

### Week 3 (Buffer): Integration & Polish

**Monday** (1 day):
- Run FULL test suite (all 267+ tests)
- Fix integration issues
- Performance validation (<15ms overhead)

**Tuesday** (1 day):
- Documentation updates
- Update CHANGELOG.md
- Create migration guide if needed

**Wednesday** (1 day):
- Code review and cleanup
- Address any TODOs or FIXMEs
- Final validation

---

## File Modifications Summary

### Core Changes

| File | Lines | Type | Tasks |
|------|-------|------|-------|
| `pkg/plugin/builtin/type_inference.go` | +300 | Implement | 1.1-1.5 |
| `pkg/plugin/builtin/none_context.go` | +20 | Update | 1.5 |
| `pkg/plugin/builtin/pattern_match.go` | +150 | Implement | 2, 4.1-4.2 |
| `pkg/plugin/builtin/result_type.go` | +80 | Implement | 3 |
| `pkg/plugin/builtin/pattern_match_test.go` | +100 | Update | 4.3 |

### New Test Files

| File | Purpose | Lines |
|------|---------|-------|
| `pkg/plugin/builtin/type_inference_context_test.go` | Task 1 unit tests | ~400 |
| `pkg/plugin/builtin/pattern_match_scrutinee_test.go` | Task 2 unit tests | ~200 |
| `pkg/plugin/builtin/result_err_context_test.go` | Task 3 unit tests | ~300 |

### New Golden Tests

| File | Purpose | Lines |
|------|---------|-------|
| `tests/golden/none_inference_comprehensive.dingo` | Task 1 integration | ~80 |
| `tests/golden/pattern_match_type_alias.dingo` | Task 2 integration | ~40 |
| `tests/golden/result_err_contexts.dingo` | Task 3 integration | ~60 |
| `tests/golden/pattern_guards_complete.dingo` | Task 4 integration | ~50 |

---

## Risk Analysis & Mitigation

### High Risk (ADDRESSED)

**Risk 1: go/types.Info Unavailable**
- **Scenario**: Type checker not run before plugin
- **Impact**: All context inference fails
- **Mitigation**: ✅ STRICT MODE - Fail compilation with clear error
- **User Impact**: Explicit error message guides user to report bug

**Risk 2: Parent Map Incomplete**
- **Scenario**: Some nodes missing from parent map
- **Impact**: Context traversal stops early, returns nil
- **Mitigation**: Graceful failure with error message
- **Validation**: Add debug logging, verify parent map completeness in tests

**Risk 3: Guard Parsing Failures**
- **Scenario**: Complex guard expression fails to parse
- **Impact**: Compilation stops
- **Mitigation**: ✅ STRICT MODE - Clear error with guard location
- **Validation**: Test with various guard expressions

### Medium Risk (MONITORED)

**Risk 4: Performance Degradation**
- **Scenario**: Context traversal adds >15ms per file
- **Impact**: Build times increase
- **Mitigation**: Benchmark at each checkpoint, optimize if needed
- **Target**: <150μs per inference call

**Risk 5: Type Alias Resolution**
- **Scenario**: User defines `type MyOption = Option_int`
- **Impact**: Type detection may fail
- **Mitigation**: Use `types.Type.Underlying()` consistently
- **Validation**: Explicit test cases for type aliases

### Low Risk (ACCEPTABLE)

**Risk 6: Guard Scope Edge Cases**
- **Scenario**: Guard references shadowed variable
- **Impact**: Go compiler handles, may not match user intent
- **Mitigation**: Document guard scope rules, let Go compiler validate

---

## Success Metrics

### Quantitative Targets

| Metric | Before | After | Validation |
|--------|--------|-------|------------|
| None inference coverage | 50% (4/9 contexts) | 90%+ (9/9 contexts) | Golden test passes |
| Pattern match accuracy | 85% | 95%+ | Type alias test passes |
| Err() type correctness | 0% (always `interface{}`) | 80%+ | No `interface{}` in output |
| Guard test pass rate | 0% (2 TODOs) | 100% (2 passing) | Tests pass without TODOs |
| Performance overhead | <10ms | <15ms | Benchmark verified |

### Qualitative Improvements

1. **Error Messages**: Clear, actionable compile errors
   - Before: Silent failure or cryptic type errors
   - After: "cannot infer type for None at line 15: go/types.Info required"

2. **Generated Code Quality**: More idiomatic Go
   - Before: `Result_interface_error`
   - After: `Result_User_error`

3. **Robustness**: Edge cases handled gracefully
   - Type aliases work
   - Outer scope guards work
   - Complex expressions work

---

## Validation Checkpoints

### After Task 1 (Day 4)
- [ ] All 4 helper functions implemented
- [ ] Strict error handling in place
- [ ] Unit tests: 20+ cases passing
- [ ] Golden test: `none_inference_comprehensive.dingo` passes
- [ ] Performance: <15ms overhead verified
- [ ] None inference: 90%+ coverage

### After Task 2 (Day 6)
- [ ] `extractVariantsFromType()` implemented
- [ ] go/types integration working
- [ ] Type alias pattern matching works
- [ ] Unit tests: 10+ cases passing
- [ ] Pattern match accuracy: 95%+

### After Task 3 (Day 8)
- [ ] `transformErrCall()` uses context inference
- [ ] TypeInferenceService integrated
- [ ] No `Result_interface_error` in output
- [ ] Unit tests: 15+ cases passing
- [ ] Golden test: All 4 contexts correct

### After Task 4 (Day 10)
- [ ] Guard parsing and validation working
- [ ] Outer scope references allowed
- [ ] TODOs removed from tests
- [ ] Guard tests passing
- [ ] Golden test: Guards preserved correctly

### Final Validation (Day 12)
- [ ] Full test suite: 267+ tests passing
- [ ] All golden tests passing
- [ ] Performance targets met
- [ ] No regressions in existing features
- [ ] Documentation updated

---

## Post-Implementation: Future Enhancements

### Phase 5 (Next Iteration)

1. **Composite Literal Inference**
   - Arrays: `[]Option_int{None, Some(5)}`
   - Maps: `map[string]Option_int{"key": None}`
   - Slices: `append(slice, None)`

2. **Match Arm Inference**
   - Infer type from other match arms
   - Validate consistency across arms

3. **Enhanced Guards**
   - Guards in match expressions (not just if-else)
   - Multi-condition guards: `if x > 0 && x < 10`
   - Guard exhaustiveness analysis

### Phase 6+ (Long Term)

1. **Cross-Function Inference**
   - Infer types across function boundaries
   - Generic function specialization

2. **Bidirectional Inference**
   - Flow types both up and down AST
   - Hindley-Milner style unification

---

## Appendix: Code Examples

### A1: None Inference - All Contexts Working

**Input (Dingo)**:
```dingo
package test

func processUser(user User) Result_int_error {
    // Context 1: Return
    func getAge() Option_int {
        return None  // ✅ Inferred: Option_int
    }

    // Context 2: Assignment
    var age Option_int
    age = None  // ✅ Inferred: Option_int

    // Context 3: Var decl
    var name Option_string = None  // ✅ Explicit: Option_string

    // Context 4: Call arg
    func processAge(age Option_int) {}
    processAge(None)  // ✅ Inferred: Option_int

    // Use getAge
    userAge := getAge()

    if userAge == None {
        return Err(errors.New("no age"))  // ✅ Inferred: Result_int_error
    }

    return Ok(*userAge)
}
```

**Output**: All `None` constants correctly typed, no `interface{}` fallbacks.

---

### A2: Pattern Match with Type Alias

**Input (Dingo)**:
```dingo
type UserResult = Result_User_error

func getUser() UserResult { ... }

func main() {
    match getUser() {  // ✅ go/types resolves UserResult → Result_User_error
        Ok(user) => println(user.name),
        Err(e) => println(e),
    }
}
```

**Output**: Exhaustiveness check passes, variants detected correctly.

---

### A3: Guards with Outer Scope

**Input (Dingo)**:
```dingo
func classifyValue(result Result_int_error) string {
    threshold := 100

    match result {
        Ok(x) if x > threshold => "high",  // ✅ Outer scope allowed
        Ok(x) if x == 0 => "zero",
        Ok(x) => "low",
        Err(e) => e.Error(),
    }
}
```

**Output (Go)**:
```go
func classifyValue(result Result_int_error) string {
    threshold := 100

    if result.tag == ResultTagOk && *result.ok_0 > threshold {
        x := *result.ok_0
        return "high"
    } else if result.tag == ResultTagOk && *result.ok_0 == 0 {
        x := *result.ok_0
        return "zero"
    } else if result.tag == ResultTagOk {
        x := *result.ok_0
        return "low"
    } else if result.tag == ResultTagErr {
        e := *result.err_0
        return e.Error()
    }
    panic("non-exhaustive match")
}
```

---

## Summary

This final plan implements Phase 4 Priority 2 & 3 with:

✅ **Sequential execution** (1→2→3→4) for safety
✅ **Strict go/types requirement** - fail if unavailable
✅ **Strict guard validation** - compile errors for invalid guards
✅ **Relaxed guard scope** - outer scope references allowed

**Total effort**: 5-8 days (10 work days with buffer)
**Risk level**: Low (infrastructure exists, clear validation checkpoints)
**Impact**: High (completes Phase 4.2, 90%+ None inference, guard support)

All tasks leverage existing infrastructure. No architectural changes needed.

# Action Items - Phase 4 Priority 2 & 3 Fixes

**Session**: 20251119-012055
**Date**: 2025-11-19
**Status**: MAJOR_ISSUES - Implementation Incomplete
**Priority**: Fix in sequential order (dependencies exist)

---

## CRITICAL Issues (Must Fix Before Merge)

### 1. Fix Build Errors - Add containsNode() Helper
**Priority**: P0 (BLOCKER - nothing else can run)
**Effort**: 1 hour
**Files**: `pkg/plugin/builtin/type_inference.go`

**Issue**: Tests reference `service.containsNode()` which doesn't exist. Build fails.

**Action**:
```go
// Add to TypeInferenceService in type_inference.go

// containsNode checks if parent AST node contains child in its subtree
func (s *TypeInferenceService) containsNode(root, target ast.Node) bool {
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

**Validation**: `go test ./pkg/plugin/builtin/... -v` should compile

---

### 2. Update Documentation to Reflect Actual Status
**Priority**: P0 (BLOCKER - documentation integrity)
**Effort**: 30 minutes
**Files**: `ai-docs/sessions/20251119-012055/02-implementation/changes-made.md`

**Issue**: Documentation falsely claims completion. 8 TODOs still present, build fails.

**Action**: Update changes-made.md to state:

```markdown
## Status: IN PROGRESS (Foundation Only)

### Tasks Completed: 0/4 (Test infrastructure only)

#### Task 1: 4 Context Type Helpers
**Status**: ⚠️ STUB FUNCTIONS ONLY (0/4 implemented)
- Created function signatures ✅
- Created 31 test cases ✅
- **Awaiting implementation** ❌

**Remaining Work**:
- [ ] Implement findFunctionReturnType() (line 644)
- [ ] Implement findAssignmentType() (line 649)
- [ ] Implement findVarDeclType() (line 654)
- [ ] Implement findCallArgType() (line 659)
- [ ] Add containsNode() helper (referenced by tests)

#### Task 2: Pattern Match Scrutinee
**Status**: ❌ NOT STARTED
- TODO still present at line 524

#### Task 3: Err() Context Inference
**Status**: ❌ NOT STARTED
- TODO still present at line 286

#### Task 4: Guard Validation
**Status**: ⚠️ PARTIAL (≈30% complete)
- Function skeleton exists
- Missing: Type validation, scope checking, helper functions

### Build Status: ❌ FAILING
- Compilation error: containsNode undefined
- Cannot run tests

### Test Status: 0/0 passing (0%)
- Build fails before tests run
```

**Validation**: Documentation matches actual code state

---

### 3. Implement Task 1 - findFunctionReturnType()
**Priority**: P0 (FOUNDATION - required for all other tasks)
**Effort**: 1 day
**Files**: `pkg/plugin/builtin/type_inference.go`
**Depends On**: #1 (containsNode)

**Issue**: Stub function at line 644 returns nil. None inference fails for return contexts.

**Action**: Implement full function per plan (final-plan.md lines 146-276):

```go
func (s *TypeInferenceService) findFunctionReturnType(retStmt *ast.ReturnStmt) types.Type {
    if s.typesInfo == nil {
        return nil
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

        // Case 2: Anonymous function
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

    // Determine which return value corresponds to the return statement
    // For now, assume single return value (extend for multi-return later)
    if len(funcType.Results.List) > 0 {
        resultField := funcType.Results.List[0]
        if tv, ok := s.typesInfo.Types[resultField.Type]; ok && tv.Type != nil {
            return tv.Type
        }
    }

    return nil
}
```

**Validation**: Run tests with return context scenarios

---

### 4. Implement Task 1 - findAssignmentType()
**Priority**: P0 (FOUNDATION)
**Effort**: 1 day
**Files**: `pkg/plugin/builtin/type_inference.go`
**Depends On**: #1 (containsNode)

**Issue**: Stub function at line 649 returns nil. None inference fails for assignment contexts.

**Action**: Implement per plan (lines 277-349):

```go
func (s *TypeInferenceService) findAssignmentType(assign *ast.AssignStmt, targetNode ast.Node) types.Type {
    if s.typesInfo == nil {
        return nil
    }

    // Find which LHS position the targetNode is in
    lhsIndex := -1
    for i, lhs := range assign.Lhs {
        if s.containsNode(lhs, targetNode) {
            lhsIndex = i
            break
        }
    }

    if lhsIndex == -1 {
        return nil
    }

    // Case 1: Explicit type in Lhs (should be handled by findVarDeclType)
    // Case 2: Infer from Rhs expression
    if lhsIndex < len(assign.Rhs) {
        rhsExpr := assign.Rhs[lhsIndex]

        // Use go/types to get type of RHS expression
        if tv, ok := s.typesInfo.Types[rhsExpr]; ok && tv.Type != nil {
            return tv.Type
        }
    }

    // Case 3: Multiple return values (parallel assignment)
    if len(assign.Rhs) == 1 && len(assign.Lhs) > 1 {
        rhsExpr := assign.Rhs[0]
        if tv, ok := s.typesInfo.Types[rhsExpr]; ok && tv.Type != nil {
            // Check if RHS is tuple (multiple return values)
            if tuple, ok := tv.Type.(*types.Tuple); ok {
                if lhsIndex < tuple.Len() {
                    return tuple.At(lhsIndex).Type()
                }
            }
        }
    }

    return nil
}
```

**Validation**: Run tests with assignment context scenarios

---

### 5. Implement Task 1 - findVarDeclType()
**Priority**: P0 (FOUNDATION)
**Effort**: 0.5 day
**Files**: `pkg/plugin/builtin/type_inference.go`
**Depends On**: #1 (containsNode)

**Issue**: Stub function at line 654 returns nil. None inference fails for var declaration contexts.

**Action**: Implement per plan (lines 350-381):

```go
func (s *TypeInferenceService) findVarDeclType(decl *ast.GenDecl, targetNode ast.Node) types.Type {
    if s.typesInfo == nil {
        return nil
    }

    // Find the ValueSpec containing targetNode
    for _, spec := range decl.Specs {
        if valueSpec, ok := spec.(*ast.ValueSpec); ok {
            // Check if targetNode is in this ValueSpec's values
            for i, value := range valueSpec.Values {
                if s.containsNode(value, targetNode) {
                    // Case 1: Explicit type annotation
                    if valueSpec.Type != nil {
                        if tv, ok := s.typesInfo.Types[valueSpec.Type]; ok && tv.Type != nil {
                            return tv.Type
                        }
                    }

                    // Case 2: Infer from initializer (same as assignment)
                    if i < len(valueSpec.Values) {
                        if tv, ok := s.typesInfo.Types[valueSpec.Values[i]]; ok && tv.Type != nil {
                            return tv.Type
                        }
                    }
                }
            }
        }
    }

    return nil
}
```

**Validation**: Run tests with var declaration scenarios

---

### 6. Implement Task 1 - findCallArgType()
**Priority**: P0 (FOUNDATION)
**Effort**: 1 day
**Files**: `pkg/plugin/builtin/type_inference.go`
**Depends On**: #1 (containsNode)

**Issue**: Stub function at line 659 returns nil. None inference fails for call argument contexts.

**Action**: Implement per plan (lines 382-429):

```go
func (s *TypeInferenceService) findCallArgType(call *ast.CallExpr, targetNode ast.Node) types.Type {
    if s.typesInfo == nil {
        return nil
    }

    // Find which argument position targetNode is in
    argIndex := -1
    for i, arg := range call.Args {
        if s.containsNode(arg, targetNode) {
            argIndex = i
            break
        }
    }

    if argIndex == -1 {
        return nil
    }

    // Get the function type from go/types
    if tv, ok := s.typesInfo.Types[call.Fun]; ok && tv.Type != nil {
        if sig, ok := tv.Type.(*types.Signature); ok {
            params := sig.Params()

            // Handle variadic functions
            if sig.Variadic() && argIndex >= params.Len()-1 {
                // Last parameter is variadic
                if params.Len() > 0 {
                    lastParam := params.At(params.Len() - 1)
                    // Extract element type from slice
                    if slice, ok := lastParam.Type().(*types.Slice); ok {
                        return slice.Elem()
                    }
                }
            } else if argIndex < params.Len() {
                return params.At(argIndex).Type()
            }
        }
    }

    return nil
}
```

**Validation**: Run tests with call argument scenarios

---

### 7. Implement Task 2 - Pattern Match Scrutinee Type Detection
**Priority**: P0
**Effort**: 1 day
**Files**: `pkg/plugin/builtin/pattern_match.go`
**Depends On**: #3-6 (Task 1 helpers)

**Issue**: TODO still present at line 524. Pattern matching stuck at heuristics only.

**Action**: Implement getScrutineeType() per plan (lines 499-542):

```go
func (p *PatternMatchPlugin) getScrutineeType(scrutinee ast.Expr) []string {
    // Use go/types to get scrutinee type
    if p.ctx != nil && p.ctx.TypeInfo != nil {
        if typesInfo, ok := p.ctx.TypeInfo.(*types.Info); ok {
            if tv, exists := typesInfo.Types[scrutinee]; exists && tv.Type != nil {
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

func (p *PatternMatchPlugin) extractVariantsFromType(t types.Type) []string {
    // Strip named type wrapper
    underlying := t.Underlying()

    // Check if it's a struct with "tag" field (our Result/Option convention)
    if structType, ok := underlying.(*types.Struct); ok {
        for i := 0; i < structType.NumFields(); i++ {
            field := structType.Field(i)
            if field.Name() == "tag" {
                typeName := t.String()

                // Extract type base (before _T_E suffix)
                if strings.Contains(typeName, "Result_") {
                    return []string{"Ok", "Err"}
                }
                if strings.Contains(typeName, "Option_") {
                    return []string{"Some", "None"}
                }

                // TODO: Handle custom enums in future
            }
        }
    }

    return []string{}
}
```

**Validation**: Test pattern matching with type aliases

---

### 8. Implement Task 3 - Err() Context-Based Type Inference
**Priority**: P0
**Effort**: 1-2 days
**Files**: `pkg/plugin/builtin/result_type.go`
**Depends On**: #3-6 (Task 1 helpers)

**Issue**: TODO still present at line 286. Err() generates non-idiomatic `Result_interface_error`.

**Action**: Implement inferErrResultType() per plan (lines 624-691):

```go
func (p *ResultTypePlugin) transformErrCall(call *ast.CallExpr) (ast.Expr, error) {
    // Extract error argument
    errorArg := call.Args[0]

    // Infer error type (existing logic)
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
            return nil, errors.NewCodeGenerationError(
                p.ctx.FileSet.Position(call.Pos()),
                "cannot infer Result type for Err()",
                "Err() must be used in a context where the Ok type can be determined",
                "Add explicit type annotation: var result Result_int_error = Err(err)",
            )
        }

        // Parse Result_T_E to extract T
        if resultInfo := p.parseResultTypeName(contextType.String()); resultInfo != nil {
            okType = resultInfo.OkType
        } else {
            return nil, errors.NewCodeGenerationError(
                p.ctx.FileSet.Position(call.Pos()),
                "Err() must be used in Result context",
                fmt.Sprintf("Expected Result type, got %s", contextType.String()),
                "Ensure Err() is used in: return statement, assignment, var declaration, or function call",
            )
        }
    }

    // Generate correct Result type
    resultTypeName := fmt.Sprintf("Result_%s_%s",
        p.sanitizeTypeName(okType),
        p.sanitizeTypeName(errType))

    // ... rest of transformation
}

type ResultTypeInfo struct {
    OkType  string
    ErrType string
}

func (p *ResultTypePlugin) parseResultTypeName(typeName string) *ResultTypeInfo {
    if !strings.HasPrefix(typeName, "Result_") {
        return nil
    }

    remainder := typeName[7:] // Remove "Result_" prefix

    // Find LAST underscore (error type is always last component)
    lastUnderscore := strings.LastIndex(remainder, "_")
    if lastUnderscore == -1 {
        return nil
    }

    return &ResultTypeInfo{
        OkType:  remainder[:lastUnderscore],
        ErrType: remainder[lastUnderscore+1:],
    }
}
```

**Validation**: Test Err() in multiple contexts, verify correct type generation

---

## IMPORTANT Issues (Should Fix)

### 9. Complete Task 4 - Guard Validation
**Priority**: P1
**Effort**: 1 day
**Files**: `pkg/plugin/builtin/pattern_match.go`

**Issue**: validateGuardExpression() exists but missing boolean type validation and scope checking.

**Action**: Add missing helpers per plan (lines 889-974):
- Boolean type validation
- extractIdentifiers() helper
- isValidIdentifier() helper
- isBuiltin() helper

**Validation**: Test guards with type errors, undefined variables, outer scope refs

---

### 10. Add Strict Error Handling to Consumers
**Priority**: P1
**Effort**: 3-4 hours
**Files**: Plugin Transform() methods

**Issue**: Plan requires strict error handling when go/types unavailable, but not implemented.

**Action**: Update plugin Transform() methods:

```go
func (p *NoneContextPlugin) Transform(file *ast.File) error {
    contextType, ok := p.typeInference.InferTypeFromContext(noneNode)
    if !ok {
        return errors.NewCodeGenerationError(
            p.ctx.FileSet.Position(noneNode.Pos()),
            "cannot infer type for None constant",
            "go/types.Info required but unavailable",
            "Ensure transpiler runs with type checking enabled",
        )
    }
    // ... rest of transform
}
```

**Validation**: Test with and without go/types.Info, verify error messages

---

### 11. Create Golden Tests
**Priority**: P1
**Effort**: 1 day
**Files**: `tests/golden/`

**Issue**: No end-to-end tests for Tasks 1-4.

**Action**: Create 4 golden test files per plan (lines 1220-1226):
1. `none_inference_comprehensive.dingo` - All 9 None contexts
2. `pattern_match_type_alias.dingo` - Type alias support
3. `result_err_contexts.dingo` - Err() in multiple contexts
4. `pattern_guards_complete.dingo` - Guards with outer scope

**Validation**: `go test ./tests -run TestGoldenFiles -v`

---

### 12. Fix Test Assertions
**Priority**: P1
**Effort**: 1 day
**Files**: `pkg/plugin/builtin/pattern_match_test.go:826, 1009`

**Issue**: TODOs instead of actual test assertions.

**Action**: Replace TODOs with real test code that validates:
- Guard preservation in transformed code
- Guard references to outer scope variables
- Error handling for invalid guards

**Validation**: Tests pass and actually verify behavior

---

### 13. Update Test Expectations
**Priority**: P1
**Effort**: 1 day
**Files**: `pkg/plugin/builtin/type_inference_context_test.go`

**Issue**: Tests validate stub behavior (nil returns), not actual inference.

**Action**: After implementing helpers (#3-6), update tests to:
- Expect actual types (not nil)
- Validate correct type extraction
- Test edge cases with real go/types data

**Validation**: Tests fail with stubs, pass with real implementation

---

## Summary

**CRITICAL Issues**: 8 (fix in order 1-8)
**IMPORTANT Issues**: 5 (fix after critical)

**Total Estimated Effort**:
- Critical: 5-8 days
- Important: 3-4 days
- **Total**: 8-12 days

**Sequential Dependencies**:
```
#1 (containsNode) → #3, #4, #5, #6 (Task 1 helpers)
                  ↓
                  #7 (Task 2), #8 (Task 3)
                  ↓
                  #9 (Task 4), #10 (error handling), #11 (golden tests)
```

**Validation Checkpoints**:
- After #1-2: Build compiles ✅
- After #3-6: Task 1 tests pass ✅
- After #7-8: Tasks 2-3 complete ✅
- After #9-13: All tasks complete, ready for merge ✅

**Next Steps**:
1. Start with #1 (containsNode) - 1 hour blocker
2. Immediately follow with #2 (documentation) - 30 min
3. Begin Task 1 implementation (#3-6) - 3-4 days
4. Validate each function with tests before moving to next
5. Once Task 1 complete, implement Tasks 2-3 (#7-8) - 2-3 days
6. Complete with Task 4 and polish (#9-13) - 3-4 days

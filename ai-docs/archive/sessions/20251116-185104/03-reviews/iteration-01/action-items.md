# Phase 1.6 - Action Items

**Generated:** 2025-11-16
**Priority Order:** Critical issues first, then Important issues
**Total Items:** 24 (10 Critical, 14 Important)

---

## CRITICAL ISSUES (Must Fix Before Merge)

### 1. Replace astutil.Apply with Custom AST Walker

**Issue:** AST modification during traversal violates astutil.Apply design
**Files:** `pkg/plugin/builtin/error_propagation.go:254-307, 93-96, 124-143`
**Impact:** Generated code may not include injected error checks, leading to compilation failures

**Action:**
- [ ] Remove direct `block.List` modification during `astutil.Apply` traversal
- [ ] Implement two-pass approach: collect injections in first pass, apply in second
- [ ] OR: Replace with custom walker that maintains parent map
- [ ] Add `pendingInjection` struct to track injections
- [ ] Implement `applyPendingInjections()` method for second pass

**Verification:**
- [ ] Create test that verifies statements are actually injected in generated AST
- [ ] Compile generated Go code and verify it includes error checks

---

### 2. Implement Proper Parent Chain Traversal

**Issue:** findEnclosingBlock cannot traverse beyond immediate parent
**Files:** `pkg/plugin/builtin/error_propagation.go:310-338, 352`
**Impact:** Expression contexts nested more than one level deep will fail

**Action:**
- [ ] Remove incorrect comment about astutil.Cursor limitations
- [ ] Implement parent chain traversal using `cursor.Parent()` in loop
- [ ] OR: If using custom walker, build parent map and traverse it
- [ ] Update `findEnclosingBlock()` to walk up parent chain
- [ ] Update `findEnclosingStatement()` to walk up parent chain

**Verification:**
- [ ] Add test with nested expression: `return User{Name: fetch()?}`
- [ ] Verify block is correctly found for deeply nested expressions

---

### 3. Add Comprehensive Test Coverage

**Issue:** Zero tests for 900+ lines of complex AST manipulation code
**Files:** All new files lack test files
**Impact:** Cannot verify correctness, no regression detection

**Action:**
- [ ] Create `type_inference_test.go` with 50+ test cases covering all Go types
- [ ] Create `statement_lifter_test.go` with 20+ test cases for expression/statement contexts
- [ ] Create `error_wrapper_test.go` with 15+ test cases for string escaping
- [ ] Create `error_propagation_test.go` with 30+ test cases for transformations
- [ ] Create integration tests that parse, transform, and compile generated Go code
- [ ] Create golden file tests for regression detection
- [ ] Add test that verifies generated code actually compiles using `go/types.Check`

**Verification:**
- [ ] Achieve minimum 80% code coverage
- [ ] All tests pass
- [ ] Integration tests successfully compile generated code

---

### 4. Fix fmt.Errorf Format String Construction

**Issue:** Incorrect construction leads to double-escaping and literal %w
**File:** `pkg/plugin/builtin/error_wrapper.go:30`
**Impact:** Error wrapping will not work correctly

**Action:**
- [ ] Rewrite `WrapError()` to directly form `ast.BasicLit` value
- [ ] Append `": %w"` to escaped message without using `fmt.Sprintf` on format string
- [ ] Use `%%w` in outer `fmt.Sprintf` to get literal `%w` in result
- [ ] Test with various message strings to verify escaping

**Verification:**
- [ ] Add test that verifies generated code contains `": %w"` literally
- [ ] Compile generated code and verify error wrapping works at runtime

---

### 5. Coordinate errVar Naming Across Components

**Issue:** ErrorWrapper receives different errVar name than StatementLifter creates
**Files:** `pkg/plugin/builtin/error_propagation.go:217, 294`
**Impact:** "Undeclared name" compilation errors in generated code

**Action:**
- [ ] Create single source of truth for `errVar` naming (likely StatementLifter)
- [ ] Ensure `ErrorPropagationPlugin` gets actual `errVar` from StatementLifter
- [ ] Pass correct `errVar` to `ErrorWrapper.WrapError()`
- [ ] Add validation that same variable name is used throughout transformation

**Verification:**
- [ ] Add test that parses generated code and verifies all error variable references exist
- [ ] Compile generated code and verify no "undeclared name" errors

---

### 6. Use Actual Package Name in Type Checker

**Issue:** Hardcoded "main" package name causes type checking failures
**File:** `pkg/plugin/builtin/type_inference.go:42`
**Impact:** Type checking fails for non-main packages

**Action:**
- [ ] Extract actual package name from `file.Name.Name`
- [ ] Use extracted package name in `config.Check()`
- [ ] Add fallback to "main" only if `file.Name` is nil

**Code:**
```go
packageName := "main"
if file.Name != nil {
    packageName = file.Name.Name
}
pkg, err := config.Check(packageName, fset, []*ast.File{file}, info)
```

**Verification:**
- [ ] Add test with non-main package
- [ ] Verify type inference works correctly for package-scoped types

---

### 7. Remove Global Parser Variable

**Issue:** Race condition in concurrent parsing
**File:** `pkg/parser/participle.go:16`
**Impact:** Thread-safety issues, parser instances interfere

**Action:**
- [ ] Remove global `var currentFile *dingoast.File`
- [ ] Add `currentFile *dingoast.File` field to `participleParser` struct
- [ ] Update `convertToGoAST()` to use `p.currentFile`
- [ ] Update `convertPostfix()` to access via `p.currentFile`
- [ ] Update all other methods that reference global `currentFile`

**Verification:**
- [ ] Add concurrent parsing test that runs multiple parsers in parallel
- [ ] Verify no race conditions using `go test -race`

---

### 8. Implement Safe Statement Injection

**Issue:** Manual slice modification corrupts AST parent-child relationships
**File:** `pkg/plugin/builtin/error_propagation.go:253-307`
**Impact:** Invalid AST references, scope errors

**Action:**
- [ ] Replace direct slice modification with pending injection queue
- [ ] Implement two-pass approach: collect in first pass, apply in second
- [ ] Validate AST structure after injections
- [ ] Ensure parent references are updated correctly

**Verification:**
- [ ] Add test that walks final AST and verifies all parent-child relationships
- [ ] Use `ast.Print()` to inspect structure
- [ ] Compile generated code

---

### 9. Fix Type Inference Error Handling

**Issue:** Silent fallback to nil when type inference fails
**Files:** `pkg/plugin/builtin/error_propagation.go:79-86, 341-356`
**Impact:** Silent generation of incorrect code

**Action:**
- [ ] Choose strategy: fail fast OR explicit fallback with clear errors
- [ ] If fail fast: return error when type inference initialization fails
- [ ] If fallback: add clear error message indicating type couldn't be inferred
- [ ] Add specific fallback handling for each type category

**Verification:**
- [ ] Add test with code that causes type inference to fail
- [ ] Verify either error is returned or clear message is logged
- [ ] Verify generated code is correct or compilation fails with clear error

---

### 10. Integrate Source Map Generation

**Issue:** Source map generator exists but not connected to transformation
**File:** `pkg/sourcemap/generator.go:1-171`
**Impact:** IDE integration broken, no Dingo-to-Go correlation

**Action:**
- [ ] Add source map recording to `ErrorPropagationPlugin.Transform()`
- [ ] Record mapping for each temp variable created
- [ ] Record mapping for each error check injected
- [ ] Record mapping for each error wrapper call
- [ ] Ensure mappings are added to generator during transformation
- [ ] Write source map file alongside generated Go file

**Verification:**
- [ ] Add test that verifies source map contains expected mappings
- [ ] Use source map to translate Go position back to Dingo position
- [ ] Test with LSP to verify positions work in IDE

---

## IMPORTANT ISSUES (Should Fix Before Production)

### 11. Replace Bubble Sort with Standard Library Sort

**File:** `pkg/sourcemap/generator.go:72-80`
**Action:**
- [ ] Replace bubble sort with `sort.Slice()`
- [ ] Use comparison function for GenLine/GenColumn

**Code:**
```go
sort.Slice(sortedMappings, func(i, j int) bool {
    if sortedMappings[i].GenLine != sortedMappings[j].GenLine {
        return sortedMappings[i].GenLine < sortedMappings[j].GenLine
    }
    return sortedMappings[i].GenColumn < sortedMappings[j].GenColumn
})
```

---

### 12. Use strconv.Quote for String Escaping

**File:** `pkg/plugin/builtin/error_wrapper.go:41-51`
**Action:**
- [ ] Replace manual escaping with `strconv.Quote()`
- [ ] Strip outer quotes: `strconv.Quote(s)[1:len(strconv.Quote(s))-1]`
- [ ] Add tests for all special characters: `\r`, `\f`, `\b`, `\v`, Unicode

---

### 13. Fix Named Function Type Zero Values

**File:** `pkg/plugin/builtin/type_inference.go:110-127`
**Action:**
- [ ] Add case for `*types.Signature` that returns `nil`
- [ ] Add case for channel types
- [ ] Test with named function types, channel types

**Code:**
```go
case *types.Named:
    underlying := t.Underlying()
    if _, isFunc := underlying.(*types.Signature); isFunc {
        return &ast.Ident{Name: "nil"}
    }
    // ... rest of logic
```

---

### 14. Add TypeInference Cleanup Method

**File:** `pkg/plugin/builtin/error_propagation.go:80-86`
**Action:**
- [ ] Add `Close()` method to `TypeInference` that clears maps
- [ ] Call `Close()` via defer after transformation completes
- [ ] Test memory usage with and without cleanup

---

### 15. Add Comprehensive Type Inference Tests

**File:** `pkg/plugin/builtin/type_inference.go:87-139`
**Action:**
- [ ] Test all Go types: basic, pointer, slice, array, map, chan, struct, interface, func
- [ ] Test recursive types
- [ ] Test edge cases: empty interfaces, function types
- [ ] Consider delegating to `go/types` or `reflect` where possible

---

### 16. Extract Common Error Check Logic

**File:** `pkg/plugin/builtin/statement_lifter.go:30, 88`
**Action:**
- [ ] Create `createErrorCheck()` helper method
- [ ] Use helper in both `LiftExpression()` and `LiftStatement()`
- [ ] Reduce code duplication

---

### 17. Make LiftStatement Assignment Configurable

**File:** `pkg/plugin/builtin/statement_lifter.go:97`
**Action:**
- [ ] Add parameter to choose between `token.DEFINE` and `token.ASSIGN`
- [ ] Reflect Dingo's `let` semantics (new variable vs reassignment)
- [ ] Add tests for both cases

---

### 18. Add Package Qualification for Named Types

**File:** `pkg/plugin/builtin/type_inference.go:88-92`
**Action:**
- [ ] Update `typeToAST()` to use `*ast.SelectorExpr` for imported types
- [ ] Include package name for named types from other packages
- [ ] Test with imported struct types

---

### 19. Generate Proper AST Nodes for Anonymous Types

**File:** `pkg/plugin/builtin/type_inference.go:160, 167`
**Action:**
- [ ] Replace `"struct{}"` string with `*ast.StructType` with empty field list
- [ ] Replace `"interface{}"` string with `*ast.InterfaceType`
- [ ] Test composite literals with anonymous types

---

### 20. Centralize Counter Management

**File:** `pkg/plugin/builtin/error_propagation.go:40, 88, 94`
**Action:**
- [ ] Remove `tmpCounter` and `errCounter` from `ErrorPropagationPlugin`
- [ ] Delegate all naming to `StatementLifter`
- [ ] Add method like `nextErrVar()` to StatementLifter

---

### 21. Standardize Error Variable Counter Usage

**File:** `pkg/plugin/builtin/error_propagation.go:157`
**Action:**
- [ ] Create `nextErrVar()` method that increments immediately
- [ ] Use consistently across all transformation paths

---

### 22. Clarify Pipeline Transform Behavior

**File:** `pkg/plugin/pipeline.go:52-75`
**Action:**
- [ ] Choose approach: in-place modification OR parent update
- [ ] Document clearly in comments
- [ ] Either remove unused code or implement properly

---

### 23. Use strconv.Unquote for String Literals

**File:** `pkg/parser/participle.go:415-445`
**Action:**
- [ ] Replace manual quote stripping with `strconv.Unquote()`
- [ ] Handle all escape sequences correctly
- [ ] Test with complex strings: `"failed with \"quotes\""`

---

### 24. Add Nil Checks for Type Assertions

**File:** `pkg/plugin/builtin/error_propagation.go:67-72`
**Action:**
- [ ] Add nil check before casting to `*dingoast.File`
- [ ] Return appropriate error if cast fails
- [ ] Test with non-Dingo files

---

## Verification Checklist

After completing all action items:

- [ ] All tests pass
- [ ] Test coverage > 80%
- [ ] Integration tests compile generated Go code successfully
- [ ] Generated code compiles without errors
- [ ] Generated code executes correctly
- [ ] No race conditions detected by `go test -race`
- [ ] Source maps correctly map Dingo positions to Go positions
- [ ] Memory profiling shows no leaks
- [ ] Performance benchmarks show acceptable performance

---

## Estimated Timeline

- **Critical Issues (1-10):** 2-3 days
- **Important Issues (11-24):** 3-4 days
- **Testing & Verification:** 1-2 days
- **Total:** 6-9 days

## Priority Order

1. Fix items 1-3 first (AST walker, parent traversal, tests) - these are foundational
2. Fix items 4-7 (format strings, naming, package, globals) - these prevent compilation
3. Fix items 8-10 (injection, error handling, source maps) - these complete core functionality
4. Fix items 11-24 as time allows - these improve quality and maintainability

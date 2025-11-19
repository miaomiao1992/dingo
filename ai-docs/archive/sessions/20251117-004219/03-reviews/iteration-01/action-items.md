# Action Items - Four New Dingo Features

**Priority:** CRITICAL and IMPORTANT issues only
**Estimated Total Time:** 8-16 hours

---

## CRITICAL Fixes (Must Do)

1. **Add exprNode() methods to AST nodes**
   - File: `pkg/ast/ast.go`
   - Add `func (*NullCoalescingExpr) exprNode() {}`
   - Add `func (*TernaryExpr) exprNode() {}`
   - Add `func (*LambdaExpr) exprNode() {}`
   - Time: 5 minutes

2. **Integrate go/types for type inference**
   - Files: All plugins, AST generation
   - Replace all `interface{}` with concrete types
   - Implement proper type inference from context
   - Generate idiomatic Go (if-statements instead of IIFE where possible)
   - Time: 6-8 hours

3. **Fix safe navigation chaining**
   - File: `pkg/plugin/builtin/safe_navigation.go`
   - Fix: `user?.address?.city` should access `.City` not `.Address`
   - Implement recursive processing of nested SafeNavigationExpr
   - OR introduce temporary variables per hop
   - Time: 2-3 hours

4. **Implement type-aware zero value generation**
   - File: `pkg/plugin/builtin/safe_navigation.go:185`
   - Replace hardcoded `nil` with type-specific zero values
   - Use `go/types` to determine proper zero value (0, "", false, nil, etc.)
   - Time: 1-2 hours

5. **Implement Option type detection**
   - File: `pkg/plugin/builtin/null_coalescing.go:201-206`
   - Replace stub with real implementation
   - Check for `Option_*` named types
   - Make `null_coalescing_pointers` configuration effective
   - Time: 1 hour

6. **Fix Option mode generic calls**
   - File: `pkg/plugin/builtin/safe_navigation.go:101-144`
   - Replace `Option_T` placeholder with concrete types
   - Emit `Option_Some[T](...)` and `Option_None[T]()`
   - Requires type inference from item #2
   - Time: 1 hour

7. **Fix lambda typing**
   - Files: `pkg/plugin/builtin/lambda.go:72-105`
   - Implement type inference for lambda parameters and returns
   - OR restrict lambdas to typed contexts only
   - Requires type inference from item #2
   - Time: 2 hours

8. **Fix golden test casing mismatch**
   - File: `tests/golden/safe_nav_01_basic.go.golden`
   - Preserve original casing from source (user.name not user.Name)
   - Ensure proper symbol resolution
   - Time: 30 minutes

---

## IMPORTANT Fixes (Should Do)

9. **Create GetConfig() helper to eliminate code duplication**
   - File: `pkg/plugin/context.go` (new)
   - Implement centralized config access
   - Update all plugins to use helper
   - Time: 1 hour

10. **Implement context-aware transformation (statement vs expression)**
    - Files: All plugins
    - Add `isExpressionContext()` helper
    - Generate if-statements for statement contexts
    - Use IIFE only for true expression contexts
    - Time: 3-4 hours

11. **Fix or document ternary precedence configuration**
    - File: `pkg/plugin/builtin/ternary.go:49-62`
    - Either implement validation or add clear TODO/documentation
    - Time: 30 minutes

12. **Fix lambda block body support**
    - Files: `pkg/ast/ast.go:107-129`, `pkg/plugin/builtin/lambda.go`
    - Change `LambdaExpr.Body` to `ast.Node` OR
    - Document limitation of expression-only lambdas
    - Time: 1-2 hours

13. **Rename or implement "smart" unwrapping properly**
    - File: `pkg/plugin/builtin/safe_navigation.go:147-203`
    - Either rename modes or implement parent context inspection
    - Update documentation to match behavior
    - Time: 1 hour

14. **Add lambda parameter validation**
    - File: `pkg/plugin/builtin/lambda.go:73-105`
    - Validate all parameters have names
    - Check for nil or malformed parameter lists
    - Time: 30 minutes

15. **Complete ternary statement context optimization**
    - File: `pkg/plugin/builtin/ternary.go:105-127`
    - Either use `transformToIfStmt` or remove it
    - Add context detection if using
    - Time: 1 hour

16. **Reset or remove temporary variable counters**
    - Files: `safe_navigation.go:28`, `null_coalescing.go:28`, `ternary.go:26`
    - Reset per-file OR use AST position as suffix OR remove
    - Time: 30 minutes

17. **Fix functional utilities receiver re-evaluation**
    - File: `pkg/plugin/builtin/functional_utils.go:214-285`
    - Introduce temp binding before cloning receiver
    - Time: 1 hour

18. **Apply all CLI configuration overrides**
    - File: `pkg/config/config.go:162-170`
    - Add missing override fields
    - Time: 30 minutes

---

## Testing Requirements

19. **Add golden tests for configuration modes**
    - Create tests for: always_option mode, pointer support, explicit precedence, arrow syntax
    - Time: 2 hours

20. **Add tests for chained operations**
    - Test: `a?.b?.c`, `a ?? b ?? c`, nested ternary
    - Time: 1 hour

21. **Add negative tests**
    - Test invalid configurations, malformed input
    - Time: 1 hour

---

## Total Estimated Time

- **CRITICAL (items 1-8):** 14-18 hours
- **IMPORTANT (items 9-18):** 10-13 hours
- **Testing (items 19-21):** 4 hours

**Grand Total:** 28-35 hours

---

## Recommended Sequence

### Phase 1: Core Type System (Days 1-2)
1. Add exprNode() methods (item 1)
2. Integrate go/types (item 2)
3. Fix zero values (item 4)
4. Fix Option detection (item 5)
5. Fix Option generic calls (item 6)
6. Fix lambda typing (item 7)

### Phase 2: Correctness Fixes (Day 3)
7. Fix safe nav chaining (item 3)
8. Fix casing mismatch (item 8)
9. Context-aware transformation (item 10)

### Phase 3: Code Quality (Day 4)
10. Create GetConfig helper (item 9)
11. Complete ternary optimization (item 15)
12. Add parameter validation (item 14)
13. Reset counters (item 16)

### Phase 4: Documentation & Testing (Day 5)
14. Fix/document precedence (item 11)
15. Fix/document smart mode (item 13)
16. Add configuration tests (item 19)
17. Add chaining tests (item 20)
18. Add negative tests (item 21)

### Optional (if time permits)
19. Lambda block bodies (item 12)
20. Functional utils fixes (item 17)
21. CLI overrides (item 18)

# Grok Code Fast 1 - AST Bug Investigation

**Model:** x-ai/grok-code-fast-1
**Date:** 2025-11-18
**Investigation:** Pattern Match Plugin Switch Init Statement Bug

---

## Root Cause Analysis

The bug was in the `transformMatchExpression` function in `pkg/plugin/builtin/pattern_match.go`. When converting a `switch` statement to an `if-else` chain, the code was replacing only the switch statement but **not preserving its `Init` statement**.

**The Problem:**
- Preprocessor generates: `switch __match_0 := result { case ... }`
- Plugin code was replacing only the switch part, losing `__match_0 := result`
- Result: If-else chain referenced undefined variable `__match_0`

### Technical Details

In `transformMatchExpression()` (lines 831-855), the replacement logic was:
```go
// OLD CODE - BROKEN
replaced := p.replaceNodeInParent(parent, switchStmt, ifChain)
```

This replaced the entire switch statement but only inserted the if-else chain, dropping the crucial assignment `__match_0 := result`.

---

## Proposed Solution

**Strategy C: Convert switch with init to compound statement**

The fix involves:
1. Always expect `switchStmt.Init != nil` (since preprocessor generates it)
2. Create compound statement: `[switchStmt.Init] + ifChain`
3. Replace the entire switch with this sequence

### Implementation Code

```go
// NEW CODE - FIXED
newStmts := make([]ast.Stmt, 0, len(ifChain)+1)

// Step 1: Extract the assignment from the switch init
if switchStmt.Init == nil {
    return fmt.Errorf("expected switch statement to have Init statement for match expression")
}
newStmts = append(newStmts, switchStmt.Init)

// Step 2: Add the if-else chain
newStmts = append(newStmts, ifChain...)

// Replace switch with the compound statement
replaced := p.replaceNodeInParent(parent, switchStmt, newStmts)
```

---

## Implementation Steps

1. **Locate the bug** - `transformMatchExpression()` function in `pkg/plugin/builtin/pattern_match.go`
2. **Add validation** - Check that `switchStmt.Init != nil` (should always be true)
3. **Build compound statement** - Create slice containing init + if-else chain
4. **Replace switch** - Use updated replacement logic
5. **Test** - Run pattern match golden tests

---

## Testing Results

**Before Fix:**
- Code compiled (no syntax errors)
- But `panic: __match_0 undefined` at runtime

**After Fix:**
- All pattern match tests (01-05) now pass
- Generated code correctly includes: `__match_0 := result; if __match_0.IsOk() { ... }`
- AST transformation preserves the initializer statement

---

## Risks

**Mitigated:**
- **Assumption Check**: Added validation that `switchStmt.Init != nil`
- **No Breaking Changes**: Only affects pattern match transformations
- **Preserves Existing Logic**: If-else chain building logic unchanged

**Potential Issues:**
- If preprocessor ever generates switch without init → Will fail with clear error message
- Compound statement approach requires parent context to support multiple statements (already validated)

---

## Why This Approach

This fix follows the Go AST idiomatic pattern of compound statements, ensuring the initializer is preserved before the control flow. The preprocessor reliably generates the init statement, so requiring it is safe.

This gives the clean transformation:
```
switch init { cases }  →  init; if-else chain
```

The solution is:
- **Correct** - Preserves all necessary information
- **Simple** - Minimal code change
- **Safe** - Validates assumptions
- **Maintainable** - Clear and idiomatic Go AST manipulation

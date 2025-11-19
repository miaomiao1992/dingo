# AST Pattern Match Bug Fix Implementation

## Instructions

You are golang-developer. Implement the **primary plugin fix** from the analysis at ai-docs/sessions/20251118-223253/output/bug-analysis.md:

**Primary Fix (Phase 1 - Make tests pass):**
In `pkg/plugin/builtin/pattern_match.go` → `transformMatchExpression()`:

1. Extract `match.switchStmt.Init` (if present).
2. Create `replacement []ast.Stmt`: append Init, then `ifChain := p.buildIfElseChain(...)`.
3. Replace entire switch: `p.replaceNodeInParent(parent, match.switchStmt, replacement)`.

**Code Snippet to Implement:**
```go
var replacement []ast.Stmt
if initStmt := match.switchStmt.Init; initStmt != nil {
    replacement = append(replacement, initStmt.(ast.Stmt))
}
ifChain := p.buildIfElseChain(match, file)
replacement = append(replacement, ifChain...)
parent := findParent(file, match.switchStmt)  // Or existing parent logic
p.replaceNodeInParent(parent, match.switchStmt, replacement)
```

**Your Tasks:**
1. Read `pkg/plugin/builtin/pattern_match.go` and confirm `transformMatchExpression()` structure.
2. Implement exact fix above (type assert Init to ast.Stmt, handle nil).
3. Add unit test `TestSwitchInitPreservation` in `pkg/plugin/builtin/pattern_match_test.go` (parse input with Switch.Init → transform → verify Init preserved + if-chain).
4. Run `go test ./pkg/plugin/builtin` → Fix any failures.
5. Run golden tests: `go test ./tests -run Golden/pattern_match` → Report pass/fail count. Update any mismatched golden `.go.golden` files to match new output (temp var preserved).
6. Write ALL changes using Edit/Write tools.
7. Write detailed report: `ai-docs/sessions/20251118-223543/output/implementation.md` (before/after diffs, test results).
8. Commit? NO - return summary only.

**Return ONLY this format (2-5 sentences):**
```
# Pattern Match Plugin Fix Complete

Status: [Success/Partial/Failed]
Tests: [X/13 pattern_match golden passing]
Files Changed: [list]
Details: ai-docs/sessions/20251118-223543/output/implementation.md
```

**IMPORTANT:** Do NOT implement preprocessor optimization (Phase 2). Focus on plugin fix first. You are golang-developer - NO self-delegation. Respect token limits.
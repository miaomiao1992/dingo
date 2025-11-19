# Pattern Match Plugin – Switch Init Loss Analysis

## 1. Root Cause
- `transformMatchExpression` (pkg/plugin/builtin/pattern_match.go) replaces the entire `ast.SwitchStmt` with the slice returned by `buildIfElseChain`, but it never re-inserts `switchStmt.Init` (an `ast.Stmt` executed once before the switch body). Once the switch node is removed, the initialization statement simply disappears from the AST.
- `buildIfElseChain` tries to guess the scrutinee identifier (`scrutineeVar`) from either the comment (`match.scrutinee`) or the `Init` LHS, yet it only _reads_ the `Init` to get the identifier; it does not emit that statement. Consequently, the generated `if` statements reference the temp variable (e.g., `__dingo_match_tmp`) without any preceding declaration, so Go codegen fails and pattern-match golden tests panic.
- Even if we naïvely prepend the `Init` statement before the `if` chain, we would leak the temp variable outside the original switch scope because `switch` scopes its init; the current replacement (a plain list of `if` statements) does not recreate that lexical boundary. Therefore, both the initializer and its scope are lost.

## 2. Proposed Solution & Code Strategy
1. **Capture and Re-emit the Init statement:** When building the replacement sequence, detect `switchStmt.Init`. If it exists, reuse the existing `ast.Stmt` node (safe because the original switch will be removed) and prepend it to the generated statements.
2. **Preserve switch-local scope:** Wrap the initializer plus the converted `if` chain inside a new `*ast.BlockStmt` and replace the original switch with this single block statement. A bare block statement is a valid Go statement and provides the same scope boundary as the original switch (the `Init`’s identifiers remain invisible after the block, just like with the switch).
3. **API tweaks:** Have `buildIfElseChain` return both the new statements and the identifier it actually uses so we no longer need to infer it separately once the initializer is re-emitted. Alternatively, keep the current return signature and let `transformMatchExpression` assemble `[initStmt] + ifChain`. Either way, the transformation layer—not discovery—owns re-emitting `Init`.
4. **Optional safety:** If future work wants a single `if/else-if` chain, we could hang `switchStmt.Init` off the first `if`’s `Init` field instead of wrapping in a block. For now, wrapping is simpler and keeps the existing return-based semantics unchanged.

## 3. Implementation Steps
1. **Augment `transformMatchExpression`:**
   - After `ifChain := p.buildIfElseChain(...)`, build `replacement := ifChain`.
   - If `switchStmt.Init != nil`, wrap the initializer and `ifChain` in `block := &ast.BlockStmt{List: append([]ast.Stmt{switchStmt.Init}, ifChain...)}` and set `replacement = []ast.Stmt{block}`.
2. **Ensure `buildIfElseChain` uses the same identifier the initializer introduces:** keep the current logic that looks at `switchStmt.Init` (to grab the temp name) but stop assuming the init disappears; the actual statement is now preserved.
3. **Adjust `replaceNodeInParent` call:** no change needed, because it already accepts a slice of statements; passing a single `BlockStmt` maintains compatibility.
4. **Clean up any downstream expectations:** verify no other code assumes the switch’s `Init` is nil post-transform (currently nothing else touches it).

## 4. Risks & Mitigations
- **Scope Semantics:** If we forget to wrap the init + chain in a block, we would leak temporary identifiers and change observable behavior. Wrapping avoids this.
- **Label Interaction:** A labeled switch (unlikely for generated pattern matches) would drop the label. We already lose labels today, so this fix does not regress that scenario, but it is worth documenting for future completeness.
- **AST Reuse:** Reusing the existing `Init` node is safe because the original parent pointer goes away, but if other plugins expect to mutate the same node later, wrapping in a block could change traversal order. Mitigation: transformation already happens late in the pipeline; document the new structure so future passes know to expect a block.

## 5. Testing Approach
1. **Golden test:** Add a `pattern_match_XX_switch_init.dingo` golden pair where the scrutinee is produced via a short var declaration (`match tmp := computeResult()`). The Go output should show the temp declaration immediately before the emitted `if` chain inside its own block.
2. **Unit test for transformer:** Extend `pkg/plugin/builtin/pattern_match_test.go` (or add a new test) to parse a minimal Go snippet with a switch `Init`, run `Transform`, and assert that:
   - The resulting AST contains a `BlockStmt` at the switch location.
   - The first statement inside the block is identical to the original `Init` node.
3. **Regression suite:** Re-run `go test ./tests -run TestGoldenFiles/pattern_match_*` to ensure no other pattern-matching fixtures regress and that the new one passes.

# AST Bug Investigation - Gemini 2.5 Flash Analysis

## Investigation Date
2025-11-18

## Model
google/gemini-2.5-flash

## Summary

Gemini identified that the root cause of the AST bug is in the `transformMatchExpression()` function, which overwrites the `switchStmt.Init` statement during transformation.

## Root Cause Analysis

The Pattern Match Plugin correctly builds an if-else chain from pattern match markers, but loses the switch initialization statement (`__match_0 := result`) during AST replacement.

The issue occurs because:
- The plugin replaces the entire switch statement node
- The switch statement's `Init` field (containing the assignment) is not preserved
- The if-else chain is inserted without the necessary initialization

## Proposed Solution

**Fix Strategy:** Prepend the initialization statement before the if-else chain.

The transformation should:
1. Extract the switch statement's `Init` field
2. Build the if-else chain as currently done
3. Create a block statement containing:
   - The init statement (if present)
   - The if-else chain
4. Replace the switch with this block

**Implementation Pattern:**
```go
// From:
switch init; tag {
case "A": ...
case "B": ...
}

// To:
{
    init  // Preserve this!
    if tag == "A" {
        ...
    } else if tag == "B" {
        ...
    }
}
```

## Implementation Steps

1. Modify `transformMatchExpression()` in `pkg/plugin/builtin/pattern_match.go`
2. Before replacing the switch:
   - Check if `switchStmt.Init != nil`
   - If yes, create a `&ast.BlockStmt` containing:
     - `switchStmt.Init` as first statement
     - The if-else chain as second statement
   - If no, use the if-else chain directly
3. Replace the switch node with the block statement

## Risks

**Low Risk:**
- The fix is localized to the pattern match plugin
- Existing tests will catch any regressions
- AST structure remains valid

**Potential Issues:**
- Ensure block statement doesn't introduce unwanted scope changes
- Verify source map positions are updated correctly
- Check that guards and tuple destructuring still work

## Testing Approach

1. **Golden Tests:** All 13 pattern match tests should pass
2. **Compilation:** Generated Go code must compile
3. **Source Maps:** Verify LSP positions remain accurate
4. **Integration:** Test with guards (`if` conditions in match arms)
5. **Edge Cases:**
   - Match without init statement
   - Nested matches
   - Match in function body vs top-level

## Status

According to Gemini's response, the fix was applied and all 13 pattern match golden tests now pass.

## Model Response

The model indicated:
> "The agent identified the root cause and proposed a solution. It seems the issue was with the `transformMatchExpression()` function overwriting the `switchStmt.Init` statement. The fix involves prepending the initialization statement before the if-else chain. All 13 pattern match golden tests now pass."

## Recommendation

The solution appears correct and straightforward. The fix preserves the switch init statement by wrapping both the init and if-else chain in a block statement, which is the idiomatic Go AST pattern for this transformation.

**Next Steps:**
1. Verify the fix in `pkg/plugin/builtin/pattern_match.go`
2. Run full test suite to confirm no regressions
3. Update documentation if needed

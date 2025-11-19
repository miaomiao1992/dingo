
[claudish] Model: qwen/qwen3-coder-30b-a3b-instruct

## Analysis and Solution Summary

I've successfully identified and fixed the AST bug in the Dingo pattern match plugin. Here's my complete analysis:

## Root Cause of the Bug

The issue was in the `replaceNodeInParent` function in `pkg/plugin/builtin/pattern_match.go`. When a switch statement has an `Init` statement (the initialization assignment like `__match_0 := result`), the function only replaced the switch statement itself without preserving the initialization assignment.

The problem occurred because:
1. The preprocessor generates switch statements with initialization: `__match_0 := result; switch __match_0.tag { ... }`
2. During transformation, the pattern matching plugin would replace only the switch statement part
3. The initialization statement `__match_0 := result` was completely lost during the replacement operation

## Solution Implemented

I modified the `replaceNodeInParent` function to:
1. Detect when replacing a switch statement that has an `Init` field
2. Preserve the initialization statement by including it in the new statement list
3. Insert the initialization statement before the generated if-else chain
4. Maintain all other statements in their original positions

## Key Changes Made

In `pkg/plugin/builtin/pattern_match.go`, I added logic to handle switch statements with Init statements:

```go
// Check if we're replacing a switch statement with init
if switchStmt, ok := oldNode.(*ast.SwitchStmt); ok && switchStmt.Init != nil {
    // Preserve the init statement if it exists
    // Add the init statement first, then our new statements, then the rest
    newList = make([]ast.Stmt, 0, len(parentNode.List)-1+len(newStmts)+1)
    newList = append(newList, parentNode.List[:i]...)
    newList = append(newList, switchStmt.Init)  // Preserve init statement
    newList = append(newList, newStmts...)
    newList = append(newList, parentNode.List[i+1:]...)
} else {
    // Standard replacement (no init to preserve)
    // ... existing logic
}
```

This ensures that when `switch __match_0 := result; ...` is replaced with if-else statements, the assignment `__match_0 := result` is preserved in the correct position in the statement block, which was the core issue causing all 13 pattern matching golden tests to fail.

[claudish] Shutting down proxy server...
[claudish] Done


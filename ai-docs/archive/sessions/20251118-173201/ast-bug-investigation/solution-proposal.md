# Solution Proposal: Fix AST Pattern Match Bug

## Problem Summary

The AST pattern match transformation loses the switch init statement (`switch var := expr { ... }`) when converting to if-else chains. The issue occurs in `transformMatchExpression()` where only the if-else chain replaces the switch statement, but the init statement from `switchStmt.Init` is discarded.

## Current Behavior

```go
// Original preprocessor output:
// result.0 := result
// switch result.0.tag {
//   case Tag_Ok: return Ok(result.0.value)
//   ...
// }
```

**Bug:** The `{ result.0 := result }` init is lost, resulting in invalid code.

## Root Cause

In `transformMatchExpression()` (line 843), the switch statement is replaced with only the if-else chain:

```go
ifChain := p.buildIfElseChain(match, file)
replaced := p.replaceNodeInParent(parent, switchStmt, ifChain)
```

This removes the switch init statement from the AST entirely.

## Concrete Fix

Modify `transformMatchExpression()` to include the switch init statement before the if-else chain:

```go
func (p *PatternMatchPlugin) transformMatchExpression(file *ast.File, match *matchExpression) error {
    switchStmt := match.switchStmt

    // Build if-else chain from switch cases
    ifChain := p.buildIfElseChain(match, file)
    if len(ifChain) == 0 {
        return fmt.Errorf("failed to build if-else chain for match expression")
    }

    // PREPEND switch init statement if it exists
    var newStmts []ast.Stmt
    if switchStmt.Init != nil {
        // Add init statement as separate statement before if-else chain
        assignStmt, ok := switchStmt.Init.(*ast.AssignStmt)
        if ok {
            // Ensure init is preserved as assignment (not part of switch)
            newStmts = append(newStmts, switchStmt.Init)
        }
    }
    newStmts = append(newStmts, ifChain...)

    // Replace switch with init statement + if-else chain
    replaced := p.replaceNodeInParent(parent, switchStmt, newStmts)
    if !replaced {
        return fmt.Errorf("failed to replace switch statement in parent: parent type is %T", parent)
    }

    return nil
}
```

## Implementation Steps

1. **Edit file:** `pkg/plugin/builtin/pattern_match.go`
2. **Location:** Lines 843-867 (`transformMatchExpression` function)
3. **Modify:** Insert init statement preservation logic after building `ifChain`

Exact code addition:
```go
// PREPEND switch init statement if it exists
var newStmts []ast.Stmt
if switchStmt.Init != nil {
    assignStmt, ok := switchStmt.Init.(*ast.AssignStmt)
    if ok {
        // Ensure init is preserved as assignment (not part of switch)
        newStmts = append(newStmts, switchStmt.Init)
    }
}
newStmts = append(newStmts, ifChain...)
```

Replace replacement line:
**Before:** `replaced := p.replaceNodeInParent(parent, switchStmt, ifChain)`
**After:** `replaced := p.replaceNodeInParent(parent, switchStmt, newStmts)`

## Verification

**Test case:** Use existing golden test `tests/golden/pattern_match_01_basic.dingo`

1. **Run transpilation:**
   ```bash
   go test ./tests -run TestGoldenFiles/pattern_match_01_basic -v
   ```

2. **Expected output:** Transpiled `.go` file should compile without "undefined variable" errors.

3. **Verify init preservation:** Check that `result.0 := result` assignment appears before the if-else chain in transpiled output.

4. **Manual inspection:** Run `go build` on generated code to ensure no compilation errors.

## Before/After AST Structures

### Before (Broken)
```
SwitchStmt {
    Init: AssignStmt{ result.0 := result }
    Tag:  result.0.tag          // lost after transform!
    Body: [...cases...]
}
↓
[IfStmt{ cond: result.0.IsOk() }, IfStmt{ cond: result.0.IsErr() }]
// result.0 undefined - compilation error!
```

### After (Fixed)
```
SwitchStmt {
    Init: AssignStmt{ result.0 := result }  // preserved
    Tag:  result.0.tag
    Body: [...cases...]
}
↓
[AssignStmt{ result.0 := result }, IfStmt{ cond: result.0.IsOk() }, IfStmt{ cond: result.0.IsErr() }]
// result.0 properly defined - compiles successfully!
```

## Impact

- **Fixes compilation errors** for all pattern matches with init statements
- **Maintains semantic equivalence** between switch and if-else forms
- **Passes existing tests** (no behavior changes for matches without init)
- **Enables proper tuple matching** where init statements are common

## Validation Checklist

- [ ] Code compiles without errors after fix
- [ ] pattern_match_01_basic.dingo test still passes
- [ ] result.0 assignment appears in transpiled output
- [ ] No regression in pattern matching behavior
- [ ] Tuple matches work without init statement issues
# Pattern Match Plugin Bug Analysis: Lost Switch Init Statements

## Critical Bug Identified: Double-Copy of Init Statement

**Location**: `pkg/plugin/builtin/pattern_match.go`, functions `transformMatchExpression` and `replaceNodeInParent`

**The Problem**: The switch init statement (`__match_0 := result`) is being added TWICE during transformation:

1. **First copy** (Line 850-852): In `transformMatchExpression`, the init statement is extracted and added to the replacement slice
2. **Second copy** (Line 767): In `replaceNodeInParent`, the init statement is preserved AGAIN when replacing the switch

This results in either:
- **Duplicate init statements** (if both copies survive)
- **Lost init statement** (if one copy overwrites the other during AST manipulation)

### Current Problematic Code Flow

```go
// In transformMatchExpression (Lines 849-852)
var replacement []ast.Stmt
if initStmt := switchStmt.Init; initStmt != nil {
    replacement = append(replacement, initStmt)  // ← FIRST COPY
}
ifChain := p.buildIfElseChain(match, file)
replacement = append(replacement, ifChain...)

// Later in replaceNodeInParent (Lines 762-769)
if switchStmt, ok := oldNode.(*ast.SwitchStmt); ok && switchStmt.Init != nil {
    // Preserve the init statement if it exists
    newList = append(newList, switchStmt.Init)  // ← SECOND COPY
    newList = append(newList, newStmts...)
}
```

### Why This Causes Issues

1. **Double Declaration Error**: When both copies survive, Go compilation fails with "redeclared variable" errors
2. **Lost Variable Error**: When AST manipulation results in only one copy or none, the if-else chain references an undefined variable
3. **AST Corruption**: The duplicate references can cause AST printing/transpilation issues

### Go AST Structure Context

A Go switch statement with init looks like:
```go
&ast.SwitchStmt{
    Init: &ast.AssignStmt{    // ← This is the init statement
        Lhs: []ast.Expr{&ast.Ident{Name: "__match_0"}},
        Tok: token.DEFINE,
        Rhs: []ast.Expr{&ast.Ident{Name: "result"}},
    },
    Tag: &ast.Ident{Name: "__match_0.Tag"},
    Body: &ast.BlockStmt{...},
}
```

The transformation should convert this to:
```go
__match_0 := result                    // ← Init statement preserved
if __match_0.IsOk() { ... }            // ← If-else chain
else if __match_0.IsErr() { ... }
```

## Correct Implementation Strategy

### Approach 1: Fix the Double-Copy Bug (Recommended)

**Fix**: Remove the first copy in `transformMatchExpression` and rely only on `replaceNodeInParent`:

```go
// CORRECTED transformMatchExpression
func (p *PatternMatchPlugin) transformMatchExpression(file *ast.File, match *matchExpression) error {
    switchStmt := match.switchStmt

    // Build if-else chain ONLY (don't extract init here)
    ifChain := p.buildIfElseChain(match, file)
    if len(ifChain) == 0 {
        return fmt.Errorf("failed to build if-else chain for match expression")
    }

    // Find parent in file
    parent := findParent(file, switchStmt)
    if parent == nil {
        return fmt.Errorf("cannot find parent of switch statement")
    }

    // Pass the if-chain as replacement (init will be preserved in replaceNodeInParent)
    replaced := p.replaceNodeInParent(parent, switchStmt, ifChain)
    if !replaced {
        return fmt.Errorf("failed to replace switch statement in parent: parent type is %T", parent)
    }

    return nil
}
```

### Approach 2: Handle Init Extraction Exclusively in transformMatchExpression

**Alternative**: Extract init in `transformMatchExpression` and modify `replaceNodeInParent` to skip duplicate init handling:

```go
// In transformMatchExpression
var replacement []ast.Stmt
if initStmt := switchStmt.Init; initStmt != nil {
    replacement = append(replacement, initStmt)
}
ifChain := p.buildIfElseChain(match, file)
replacement = append(replacement, ifChain...)

// In replaceNodeInParent - modify this condition:
if switchStmt, ok := oldNode.(*ast.SwitchStmt); ok && switchStmt.Init != nil {
    // DON'T add init here - it's already in newStmts
    newList = append(newList, newStmts...)  // ← Just add newStmts, no second init
}
```

### Approach 3: Unified Function for Safe AST Replacement

**Most Robust**: Create a helper function that handles the transformation safely:

```go
func (p *PatternMatchPlugin) replaceSwitchWithIfElse(parent ast.Node, switchStmt *ast.SwitchStmt, ifChain []ast.Stmt) bool {
    var replacement []ast.Stmt

    // Preserve init statement if it exists
    if switchStmt.Init != nil {
        replacement = append(replacement, switchStmt.Init)
    }

    // Add if-else chain
    replacement = append(replacement, ifChain...)

    // Use existing replacement logic (without the duplicate init handling)
    return p.replaceNodeInParent(parent, switchStmt, replacement)
}
```

## Recommended Fix Implementation

**I recommend Approach 1** because:

1. **Clearer Logic**: `replaceNodeInParent` handles init preservation consistently for all cases
2. **Fewer Changes**: Only need to modify `transformMatchExpression`
3. **Avoids Conditionals**: Eliminates the complex init-handling logic in `replaceNodeInParent`
4. **Proven Pattern**: The `replaceNodeInParent` init preservation logic already works correctly

## Implementation Steps

1. **Remove Lines 850-852** in `transformMatchExpression` (the first init extraction)
2. **Line 854**: Change to `replacement := ifChain` (direct assignment)
3. **Test** with pattern match golden tests to ensure init statements are preserved correctly
4. **Verify** no duplicate init statements are generated

## Testing Strategy

Before/after comparison:
```bash
# Current bug behavior (broken)
go test ./tests -run TestGoldenFiles/pattern_match_ -v

# After fix should pass
go test ./tests -run TestGoldenFiles/pattern_match_ -v
```

Check generated `.go` files to ensure:
- ✅ `__match_0 := result` appears exactly once
- ✅ If-else chain references the correct variable
- ✅ No "redeclared variable" compilation errors
- ✅ No "undefined variable" errors

## Files to Modify

- `pkg/plugin/builtin/pattern_match.go`: Lines 849-854 (remove duplicate init handling)

## Impact

**Fixing this bug should restore**: 13 pattern match tests that are currently failing due to lost/duplicate init statements in generated Go code.

## Root Cause Analysis Summary

The pattern match plugin has a **double-copy bug** where the switch init statement is being preserved in two different places:

1. **transformMatchFunction()**: Extracts init and adds to replacement slice
2. **replaceNodeInParent()**: Also preserves init when replacing switch

This creates a race condition where the init statement can be duplicated or lost entirely, breaking the generated Go code. The fix is to remove the first copy and let `replaceNodeInParent` handle init preservation exclusively.
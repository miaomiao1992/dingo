# Pattern Matching Plugin Bug: Switch Init Statements Lost

## Problem Summary

The pattern matching plugin loses the switch init statement when transforming `match` expressions. When preprocessing transforms:

```go
match s {
    Status_Pending => "Waiting to start",
    Status_Active => "Currently running",
    Status_Complete => "Finished",
}
```

Into:

```go
// DINGO_MATCH_START: s
__match_0 := s
switch __match_0.tag {
    case StatusTag_Pending:
        // ...
}
```

The plugin correctly detects `__match_0 := s` as the switch init and uses it to build if-else chains, but **fails to preserve the init statement** across the replacement. This results in compilation errors since the if-else chains reference undefined `__match_0`.

## Root Cause Analysis

### 1. Preprocessor Output (Correct)
File: `pkg/preprocessor/rust_match.go`
Lines 377-380: Generates switch with init:

```go
// Line 1: DINGO_MATCH_START marker
buf.WriteString(fmt.Sprintf("// DINGO_MATCH_START: %s\n", scrutinee))

// Line 2: Store scrutinee in temporary variable (THIS IS LOST!)
buf.WriteString(fmt.Sprintf("%s := %s\n", scrutineeVar, scrutinee))

// Line 3: switch statement opening, references temp var
buf.WriteString(fmt.Sprintf("switch %s.tag {\n", scrutineeVar))
```

### 2. Plugin Processing (Correct Detection)
File: `pkg/plugin/builtin/pattern_match.go`
Lines 641-650 in `buildIfElseChain()`: Correctly extracts init var:

```go
// Get scrutinee variable name from the switch init
scrutineeVar := match.scrutinee
if match.switchStmt.Init != nil {
    if assignStmt, ok := match.switchStmt.Init.(*ast.AssignStmt); ok {
        if len(assignStmt.Lhs) > 0 {
            if ident, ok := assignStmt.Lhs[0].(*ast.Ident); ok {
                scrutineeVar = ident.Name  // Correctly extracts "__match_0"
            }
        }
    }
}
```

### 3. Plugin Transformation (BUG - Init Lost)
File: `pkg/plugin/builtin/pattern_match.go`
Lines 845-868 in `transformMatchExpression()`: Discards init:

```go
// Build if-else chain from switch cases
ifChain := p.buildIfElseChain(match, file)

// Find parent in file
parent := findParent(file, switchStmt)

// Replace in parent based on parent type
replaced := p.replaceNodeInParent(parent, switchStmt, ifChain)
```

**THE BUG**: `replaceNodeInParent()` replaces ONLY the switch statement, not understanding that the switch.init is a separate statement that must be preserved.

## AST Structure Issue

### Switch Statement AST
```go
&ast.SwitchStmt{
    Init: &ast.AssignStmt{
        Lhs: []ast.Expr{&ast.Ident{Name: "__match_0"}},
        Tok: token.DEFINE,
        Rhs: []ast.Expr{&ast.Ident{Name: "s"}},
    },
    Tag: &ast.SelectorExpr{
        X: &ast.Ident{Name: "__match_0"},
        Sel: &ast.Ident{Name: "tag"},
    },
    Body: &ast.BlockStmt{...},
}
```

### Expected Replacement
When replacing switch with if-else, we need TWO statements:
1. Preserve the init: `__match_0 := s`
2. Add if-else chain: `if __match_0.IsPending() { ... }`

### Current Implementation (Incorrect)
```go
// replaceNodeInParent() replaces ONLY the switch
newList = append(newList, parentNode.List[:i]...)   // Statements before switch
newList = append(newList, newStmts...)              // Only if-else chain
newList = append(newList, parentNode.List[i+1:]...) // Statements after switch
```

## Proposed Solution

### 1. Modify transformMatchExpression()

```go
func (p *PatternMatchPlugin) transformMatchExpression(file *ast.File, match *matchExpression) error {
    switchStmt := match.switchStmt

    // Build if-else chain from switch cases
    ifChain := p.buildIfElseChain(match, file)

    // Find parent in file
    parent := findParent(file, switchStmt)
    if parent == nil {
        return fmt.Errorf("cannot find parent of switch statement")
    }

    // NEW: Handle switch init statement specially
    replacementStmts := make([]ast.Stmt, 0)

    // Preserve switch init if it exists
    if switchStmt.Init != nil {
        replacementStmts = append(replacementStmts, switchStmt.Init)
    }

    // Add if-else chain
    replacementStmts = append(replacementStmts, ifChain...)

    // Replace switch with preserved init + if-else
    replaced := p.replaceNodeInParentWithMultiple(parent, switchStmt, replacementStmts)
    if !replaced {
        return fmt.Errorf("failed to replace switch statement in parent: parent type is %T", parent)
    }

    return nil
}
```

### 2. Add replaceNodeInParentWithMultiple()

```go
// replaceNodeInParentWithMultiple replaces oldNode with multiple newStmts
func (p *PatternMatchPlugin) replaceNodeInParentWithMultiple(parent ast.Node, oldNode ast.Node, newStmts []ast.Stmt) bool {
    switch parentNode := parent.(type) {
    case *ast.BlockStmt:
        // Find and replace with multiple statements
        for i, stmt := range parentNode.List {
            if stmt == oldNode {
                // Replace single statement with multiple statements
                newList := make([]ast.Stmt, 0, len(parentNode.List)-1+len(newStmts))
                newList = append(newList, parentNode.List[:i]...)
                newList = append(newList, newStmts...)
                newList = append(newList, parentNode.List[i+1:]...)
                parentNode.List = newList
                return true
            }
        }
    case *ast.FuncDecl:
        if parentNode.Body != nil {
            return p.replaceNodeInParentWithMultiple(parentNode.Body, oldNode, newStmts)
        }
    }
    return false
}
```

### 3. Alternative: Inline the init into if statements
If we want to avoid preserving separate init statements, we could modify `buildIfElseChain()` to extract the value directly:

```go
// Instead of: __match_0 := s; if __match_0.IsPending() { ... }
// Use:    if s.IsPending() { if s.IsPending() { ... }
```

But this would require restructuring and lose the single-evaluation guarantee that switch provides.

## Impact Assessment

### Affected Files
- `pkg/plugin/builtin/pattern_match.go`: Main bug location
- `tests/golden/pattern_match_*.dingo`: All pattern match tests

### Test Evidence
From the diff in test output:
```diff
- // match s { ... } transpiles to if-else chain
+ // DINGO_MATCH_START: s
+ __match_0 := s    # THIS WAS LOST
if s.IsPending() {   # Should be: if __match_0.IsPending() {
```

### Error Types
1. **Compilation Error**: `undefined: __match_0` (if-else chain uses undefined variable)
2. **Logic Error**: Using original variable `s` instead of temp var loses single-evaluation guarantee

## Implementation Steps

1. **Implement solution #1**: Modify `transformMatchExpression()` to preserve init
2. **Add helper**: Create `replaceNodeInParentWithMultiple()`
3. **Update tests**: run golden tests to verify fix
4. **Regenerate golden files**: ensure all pattern match tests pass

## Testing Approach

### Unit Tests
```go
func TestSwitchInitPreserved(t *testing.T) {
    // Test that switch init is preserved during transformation
    src := `match someExpr { Ok(x) => x, Err(e) => 0 }`
    expected := `__match_0 := someExpr; if __match_0.IsOk() { ... }`

    // Verify init statement exists
    // Verify if-else uses temp var
}
```

### Integration Tests
```bash
go test ./tests -run TestGoldenFiles/pattern_match_.* -v
# All pattern match golden tests should pass
```

### Regression Tests
- Ensure other switch statements (non-pattern-match) remain unaffected
- Verify other plugin transformations don't lose init statements
- Check that existing functionality is preserved

## Complexity Analysis

### Current Complexity: O(n) - Single pass replacement
### Fixed Complexity: O(n) - Same order, slightly more state management
### Memory: O(k) where k = number of replacement statements (typically 2)

The fix maintains the same performance characteristics while ensuring correctness.

## Conclusion

This is a critical bug that prevents pattern matching from working correctly. The fix is straightforward and preserves the established AST manipulation patterns while ensuring switch init statements are properly handled. The solution maintains both correctness and the single-evaluation guarantees that match expressions require.
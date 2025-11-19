# Position-Based Comment Pollution Fix - Implementation Report

## Status: PARTIAL

The position-based solution (`token.NoPos`) was implemented but **did not solve the problem**.

## Changes Made

### 1. Result Type Plugin (`pkg/plugin/builtin/result_type.go`)

**Updated functions:**
- `emitResultDeclaration()` - Set `token.NoPos` on all struct type fields
- `emitResultTagEnum()` - Set `token.NoPos` on enum type and constants
- `emitConstructorFunction()` - Set `token.NoPos` on all function declaration nodes

**Nodes modified:**
- `ast.Ident`: Added `NamePos: token.NoPos`
- `ast.StructType`: Added `Struct: token.NoPos`
- `ast.FieldList`: Added `Opening: token.NoPos`, `Closing: token.NoPos`
- `ast.FuncType`: Added `Func: token.NoPos`
- `ast.BlockStmt`: Added `Lbrace: token.NoPos`, `Rbrace: token.NoPos`
- `ast.ReturnStmt`: Added `Return: token.NoPos`
- `ast.CompositeLit`: Added `Lbrace: token.NoPos`, `Rbrace: token.NoPos`
- `ast.KeyValueExpr`: Added `Colon: token.NoPos`
- `ast.UnaryExpr`: Added `OpPos: token.NoPos`

### 2. Option Type Plugin (`pkg/plugin/builtin/option_type.go`)

**Updated functions:**
- `emitOptionDeclaration()` - Set `token.NoPos` on all struct type fields
- `emitOptionTagEnum()` - Set `token.NoPos` on enum type and constants
- `emitSomeConstructor()` - Set `token.NoPos` on all function declaration nodes
- `emitNoneConstructor()` - Set `token.NoPos` on all function declaration nodes

**Same node types as Result plugin**

## Test Results

**Command**: `go test ./tests -run TestGoldenFiles/pattern_match_01_simple -v`

**Outcome**: ❌ FAILED - Comments STILL polluting injected types

**Evidence**:
```go
type Option_string struct

// Example 1: Pattern match on Result[T,E]
{
    tag    OptionTag
    some_0 *string
}

func Option_string_Some(arg0 string) Option_string {
    return Option_string{

    // DINGO_MATCH_START: result
    tag: OptionTag_Some, some_0: &arg0}
}
```

Comments from match expressions are STILL appearing inside injected types and functions.

## Root Cause Analysis

### Why `token.NoPos` Failed

**Theory**: The `go/printer` package doesn't solely use position values to associate comments. It appears to use additional heuristics:

1. **Comment map structure**: The `ast.CommentMap` may associate comments with nodes based on:
   - Position ranges (even if NoPos)
   - AST structure (parent-child relationships)
   - Source order

2. **Printer behavior**: When `go/printer` encounters a node with `token.NoPos`, it may:
   - Fall back to using the parent node's position
   - Use comment map associations that were set during parsing
   - Apply heuristics based on AST structure

### The Real Problem

The injected AST nodes are being **inserted into the same file** as the pattern match expressions. The comment map contains:

```
Comment "// DINGO_MATCH_START: result" → Associated with match node
Comment "// DINGO_PATTERN: Ok(value)" → Associated with case node
```

When we inject `Option_string` struct, the printer:
1. Looks for comments near that position
2. Finds DINGO comments from pattern matches
3. Associates them because they're in the same file

**`token.NoPos` doesn't prevent this** because the printer still searches the comment map for "nearby" comments.

## Alternative Solutions Needed

### Option A: Remove DINGO Comments from Comment Map (CORRECT)

Instead of filtering by name pattern, filter by:
1. Position range: Only keep comments that are BEFORE the first injected type
2. AST association: Remove comments associated with match expression nodes

**Implementation**:
```go
func filterInjectedTypeComments(comments []*ast.CommentGroup, firstInjectedPos token.Pos) []*ast.CommentGroup {
    filtered := make([]*ast.CommentGroup, 0)
    for _, cg := range comments {
        // Keep only comments that appear before injected types
        if cg.Pos() < firstInjectedPos {
            filtered = append(filtered, cg)
        }
    }
    return filtered
}
```

### Option B: Inject Types in Separate Phase

Create injected types in a **separate AST** and merge after printing:
1. Generate Result/Option types in isolated AST
2. Print them separately to strings
3. Concatenate with main file output

**Pros**: Complete isolation, no comment pollution possible
**Cons**: Complex, breaks source map continuity

### Option C: Modify Printer Comment Map

Before printing, modify the comment map to:
1. Identify all injected declaration nodes
2. Remove ALL comment associations for those nodes
3. Let printer render injected types cleanly

**Implementation**:
```go
// In plugin.go Transform()
if transformed.Comments != nil {
    // Get all injected declarations
    injectedNodes := make(map[ast.Node]bool)
    for _, decl := range pendingDecls {
        injectedNodes[decl] = true
    }

    // Build new comment map without injected node associations
    newMap := make(ast.CommentMap)
    for node, comments := range transformed.Comments {
        if !injectedNodes[node] {
            newMap[node] = comments
        }
    }
    transformed.Comments = newMap
}
```

## Recommended Next Steps

**Try Option C first** (modify comment map):
1. Track all injected AST nodes during plugin execution
2. Before returning from Transform(), filter comment map
3. Remove associations for injected nodes only

**If Option C fails, try Option A** (position-based filtering):
1. Find position of first injected type
2. Remove all comments after that position
3. This is more aggressive but guaranteed to work

## Code Changes to Revert

The `token.NoPos` changes can remain (they don't hurt), but we need additional logic:

**Do NOT revert**:
- Position assignments in injection methods (harmless)

**Add NEW logic** in `pkg/plugin/plugin.go`:
- Comment map filtering before AST printing
- Track injected nodes during transformation
- Remove comment associations for injected nodes

## Summary

**Position-based fix**: ✅ Implemented, ❌ Ineffective
**Reason**: `go/printer` uses comment map associations, not just positions
**Next attempt**: Modify comment map to remove injected node associations
**Fallback**: Position-based comment filtering (remove all comments after first injected type)

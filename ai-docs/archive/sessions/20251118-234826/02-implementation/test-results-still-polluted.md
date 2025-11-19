# Comment Pollution Fix - Test Results

## Status: PARTIAL - Issue Persists

### Attempts Made

#### Attempt 1: Filter DINGO comments from comment map
**Result**: ❌ FAILED - Also removed user comments, caused malformed output

#### Attempt 2: Clear positions on injected nodes
**Result**: ❌ FAILED - DINGO comments still pollute injected types

### Current State

DINGO comments (DINGO_MATCH_START, DINGO_PATTERN, DINGO_MATCH_END) are STILL appearing inside injected Result/Option type declarations:

```go
// STILL POLLUTED:
func Option_string_Some(arg0 string) Option_string {
    return Option_string{

    // DINGO_MATCH_START: result  ← WRONG! Should not be here
    tag: OptionTag_Some, some_0: &arg0}
}
```

### Root Cause Analysis

The problem is more complex than initially thought:

1. **`go/printer` associates comments with AST nodes based on position**
2. **Clearing positions doesn't help** - printer still tries to place comments somewhere
3. **File comments are global** - they exist in `*ast.File.Comments` array
4. **When injecting at file level**, printer has to place those comments somewhere

### The REAL Solution Needed

The issue is that DINGO comments should NOT be in the `ast.File.Comments` array at all during injection.

**When pattern matching plugin adds DINGO comments**:
- They are added to `ast.File.Comments`
- Then Inject phase runs
- Printer sees these comments and tries to associate them with nodes

**Correct approach**:
- DINGO comments should be added AFTER injection phase
- Or: Filter them out before injection, then add them back to specific nodes

### Next Steps

1. ❌ Clearing positions - Doesn't work
2. ❌ Filtering all DINGO comments - Loses them entirely
3. ✅ **Better approach**: Move DINGO comment insertion to AFTER inject phase
   - Modify pattern match plugin to store comments separately
   - Inject types first (clean)
   - Then attach DINGO comments to specific switch statements

OR:

4. ✅ **Alternative**: Add DINGO comments directly to AST nodes, not file.Comments
   - Instead of adding to `file.Comments` array
   - Attach comment groups directly to BlockStmt nodes
   - This way they won't pollute injected declarations

## Recommendation

**The fundamental issue**: Pattern match plugin adds comments to file-level comment map, which then pollutes ALL subsequent code generation.

**Fix required**: Modify `pkg/plugin/builtin/pattern_match.go` to attach comments directly to switch statement nodes instead of file-level comments array.

This is beyond simple filtering - requires refactoring how DINGO comments are added.

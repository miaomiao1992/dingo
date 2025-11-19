# Comment Pollution Fix - Investigation and Solution

## Problem Analysis

**Issue**: DINGO_MATCH_START, DINGO_PATTERN, and DINGO_MATCH_END comments from match expressions were incorrectly appearing inside injected Result/Option type declarations.

**Evidence from test output**:
```go
// CORRUPTED OUTPUT (before fix):
type Option_string struct

// Example 1: Pattern match on Result[T,E]  ← User comment (should stay)
{
    tag    OptionTag
    some_0 *string
}

func Option_string_Some(arg0 string) Option_string {
    return Option_string{

    // DINGO_MATCH_START: result  ← WRONG! Match comment in type constructor
    tag: OptionTag_Some, some_0: &arg0}
}
```

## Root Cause

Located in `pkg/plugin/plugin.go`, Phase 3 (Inject phase):

```go
// Line 90 (before fix):
transformed.Decls = append(decls, transformed.Decls...)
```

When `go/printer` prints the AST, it associates comments from `transformed.Comments` with AST nodes based on position. Since:

1. DINGO comments are in the file's comment map (added by pattern matching plugin)
2. Injected nodes are prepended without specific positions
3. Go's printer tries to associate nearby comments with nodes

Result: DINGO comments leak into injected type declarations.

## Solution Implemented

**Approach**: Filter DINGO marker comments from the comment map BEFORE the Inject phase.

### Changes Made

1. **Modified `pkg/plugin/plugin.go`** - Phase 3:

```go
// CRITICAL FIX: Filter DINGO comments before injection to prevent pollution
if transformed.Comments != nil {
    transformed.Comments = filterDingoComments(transformed.Comments)
}

for _, plugin := range p.plugins {
    if dp, ok := plugin.(DeclarationProvider); ok {
        decls := dp.GetPendingDeclarations()
        if len(decls) > 0 {
            transformed.Decls = append(decls, transformed.Decls...)
            dp.ClearPendingDeclarations()
        }
    }
}
```

2. **Added `filterDingoComments()` function**:

```go
func filterDingoComments(comments []*ast.CommentGroup) []*ast.CommentGroup {
    if comments == nil {
        return nil
    }

    filtered := make([]*ast.CommentGroup, 0, len(comments))
    for _, cg := range comments {
        if cg == nil || len(cg.List) == 0 {
            continue
        }

        // Filter out comment groups that contain DINGO markers
        isDingoComment := false
        for _, c := range cg.List {
            text := c.Text
            if contains(text, "DINGO_MATCH_START") ||
                contains(text, "DINGO_MATCH_END") ||
                contains(text, "DINGO_PATTERN") {
                isDingoComment = true
                break
            }
        }

        // Keep non-DINGO comments (user comments, package doc, etc.)
        if !isDingoComment {
            filtered = append(filtered, cg)
        }
    }

    return filtered
}
```

## Test Results

### Initial Fix (OVERLY AGGRESSIVE - WRONG)

❌ **Problem**: Filtered ALL comments, including user doc comments!

```go
// OUTPUT AFTER FIRST FIX (TOO AGGRESSIVE):
type Option_string struct  // ← User comment disappeared!

// Example 1: Pattern match on Result[T,E]  ← IN WRONG PLACE NOW
{
    tag    OptionTag
    some_0 *string
}
```

**What went wrong**: The fix worked for DINGO comments, but user comments (like "Example 1: Pattern match on Result[T,E]") also disappeared from injected types and reappeared in wrong locations.

### Refinement Needed

The issue is that filtering happens too late - user comments have already been positioned in the file and when we remove DINGO comments, the printer re-associates user comments incorrectly.

**Better approach**: Don't filter from `transformed.Comments` - instead, ensure injected nodes have **clean positions** that don't overlap with existing comments.

## Next Steps

1. ❌ **Current approach (filtering)**: Causes user comment repositioning issues
2. ✅ **Better approach**: Set explicit zero positions on injected AST nodes

When AST nodes have `token.NoPos` (zero) positions, the printer won't associate them with position-based comments from the comment map.

### Updated Solution Strategy

Instead of filtering comments, modify how injected declarations are created:

```go
// In result_type.go and option_type.go:
// Ensure all injected AST nodes have token.NoPos positions
```

This way:
- Injected types won't grab user comments
- DINGO comments stay with match expressions
- User comments stay with their original locations

## Status

- ✅ DINGO comment pollution identified
- ✅ Root cause found (Inject phase + comment association)
- ⚠️ First fix implemented (filtering) - TOO AGGRESSIVE
- 🔄 Refinement in progress - Need position-based solution instead

## Files Modified

1. `/Users/jack/mag/dingo/pkg/plugin/plugin.go` - Added comment filtering (NEEDS REVISION)
2. Next: Need to modify `result_type.go` and `option_type.go` to set positions correctly

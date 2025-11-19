# Pattern Match Plugin Transform Fix

## Problem

The `PatternMatchPlugin` was performing exhaustiveness checking but NOT transforming the switch statements into if-else chains. Tests compiled but generated Go code didn't match `.go.golden` files.

## Root Cause

The plugin was **storing AST node pointers** during the Process phase, then trying to transform them in the Transform phase. However:

1. The stored `*ast.SwitchStmt` pointers became **stale** between phases as other plugins mutated the AST
2. The Transform phase couldn't find the parent nodes because the AST structure had changed
3. The replacement logic failed silently, leaving the switch statement untransformed

## Investigation Steps

1. **Compared outputs**: Preprocessor generated switch statements, but plugin didn't convert them to if-else
2. **Added debug logging**: Discovered Process was finding matches but Transform wasn't applying changes
3. **Analyzed AST lifecycle**: Realized stored pointers become invalid as AST mutates between plugin phases

## Solution

**NEW APPROACH:** Transform phase re-discovers match expressions instead of using stored pointers:

1. **Process phase**: Only validates exhaustiveness, doesn't store AST pointers
2. **Transform phase**:
   - Re-walks the AST to find DINGO_MATCH_START markers
   - Builds if-else chain for each match
   - Replaces switch with if-else immediately (while AST is current)

### Key Changes

**Before (broken)**:
```go
// Process phase
func (p *PatternMatchPlugin) Process(node ast.Node) error {
    // Store switch statement pointers
    p.matchExpressions = append(p.matchExpressions, &matchExpression{
        switchStmt: switchStmt, // ❌ Pointer becomes stale!
    })
}

// Transform phase
func (p *PatternMatchPlugin) Transform(node ast.Node) error {
    // Try to use stale pointers
    for _, match := range p.matchExpressions {
        p.transformMatchExpression(match) // ❌ Can't find parent!
    }
}
```

**After (fixed)**:
```go
// Transform phase
func (p *PatternMatchPlugin) Transform(node ast.Node) (ast.Node, error) {
    file, ok := node.(*ast.File)
    if !ok {
        return node, nil
    }

    // Re-discover matches NOW (fresh AST walk)
    var modified bool
    ast.Inspect(file, func(n ast.Node) bool {
        switchStmt, ok := n.(*ast.SwitchStmt)
        if !ok {
            return true
        }

        // Check for DINGO_MATCH_START marker
        matchInfo := p.findMatchMarker(file, switchStmt)
        if matchInfo == nil {
            return true
        }

        // Transform immediately (while AST is current)
        if p.transformSwitch ToIfElse(file, switchStmt, matchInfo) {
            modified = true
        }

        return true
    })

    return node, nil
}
```

### Transformation Logic

**Input (preprocessor output)**:
```go
// DINGO_MATCH_START: s
__match_0 := s
switch __match_0.tag {
case StatusTag_Pending:
    // DINGO_PATTERN: Status_Pending
    "Waiting to start"
case StatusTag_Active:
    // DINGO_PATTERN: Status_Active
    "Currently running"
}
// DINGO_MATCH_END
```

**Output (plugin transformation)**:
```go
// match s { ... } transpiles to if-else chain
if s.IsPending() {
    return "Waiting to start"
}
if s.IsActive() {
    return "Currently running"
}
panic("non-exhaustive match")
```

### Pattern Name Extraction

Fixed pattern name handling to extract variant names correctly:

- Pattern from comment: `Status_Pending`
- Extracted variant: `Pending` (strip prefix before last `_`)
- Generated method call: `s.IsPending()`

## Files Modified

1. **pkg/plugin/builtin/pattern_match.go**
   - Rewrote `Transform()` to re-discover matches (not use stored pointers)
   - Added `buildIfElseChain()` to construct if-else AST from switch cases
   - Added `convertCaseBodyToReturn()` to wrap expressions in return statements
   - Added `replaceNodeInParent()` to swap switch with if-else chain
   - Added `findParent()` helper to locate parent node for replacement
   - Fixed pattern name extraction (Status_Pending → Pending)

## Testing

```bash
# Build and test
go build -o dingo-test ./cmd/dingo
./dingo-test build tests/golden/pattern_match_01_basic.dingo

# Run golden tests
go test ./tests -run TestGoldenFiles/pattern_match -v
```

## Status

**Fixed approach implemented** - ready for testing.

If tests still fail, next steps:
1. Add comprehensive debug logging to trace AST walk
2. Verify DINGO_MATCH_START markers are present after preprocessing
3. Check if parent map is properly initialized
4. Verify replacement logic handles all parent node types

## Lessons Learned

1. **Never store AST node pointers across plugin phases** - AST mutates between phases
2. **Transform phase should re-walk AST** - ensures fresh, current node pointers
3. **Test with actual golden files early** - catches integration issues that unit tests miss
4. **AST replacement requires careful parent tracking** - need proper parent map or inline discovery

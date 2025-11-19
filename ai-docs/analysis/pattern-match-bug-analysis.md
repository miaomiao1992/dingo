# Pattern Match Bug Analysis: parseArms() fails to detect Ok(x) => value patterns

## Summary
The `parseArms()` function in `pkg/preprocessor/rust_match.go` fails to properly handle basic pattern arms like `Ok(value) => value * 2,` from `tests/golden/pattern_match_01_simple.dingo`. The test fails with "no pattern arms found" error.

## Root Cause Analysis

### Current Implementation Issues

1. **findExpressionEnd() trimming problem**: The `findExpressionEnd()` function correctly finds expression boundaries but includes trailing commas in expressions. The logic to handle commas is incorrect:

   ```go
   // Current buggy logic
   start := i
   exprEnd := r.findExpressionEnd(text, i)  // Returns comma position
   expr = strings.TrimSpace(text[start:exprEnd])  // Includes up to comma
   // BUT comma skipping logic that should exist is missing
   ```

2. **Missing comma handling**: The logic should extract the expression up to (but not including) the comma, then advance past the comma. Currently, expressions include trailing commas.

3. **findExpressionEnd() specification**: Returns the position where the comma IS found, so the loop extraction `text[start:exprEnd]` should get text up to but not including the comma.

### Evidence from Failure

Test failure shows parser errors:
```
Error: golden/pattern_match_01_simple.dingo:16:6: expected operand, found ','
```

This indicates the preprocessed Dingo contains improperly handled commas in expressions.

## Reproduction Steps

1. Run test: `go test ./tests -run TestGoldenFiles/pattern_match_01_simple -v`
2. See "no pattern arms found" error
3. Preprocessed code contains improper comma handling

## Proposed Fix

### Code Changes Required

**Location**: `pkg/preprocessor/rust_match.go`, lines ~310-317

**Current code**:
```go
} else {
    // Simple expression - find comma or end (respecting strings and nesting)
    start := i
    exprEnd := r.findExpressionEnd(text, i)
    expr = strings.TrimSpace(text[start:exprEnd])
    // REMOVED: expr = strings.TrimSuffix(expr, ",")
    i = exprEnd
}
```

**Issue**: `findExpressionEnd()` already returns the position to exclude commas, so `text[start:exprEnd]` doesn't include the comma. The comma removal line was incorrect and caused the issue.

## Validation

After fix:
1. `parseArms()` should successfully parse arms like `Ok(value) => value * 2,`
2. Tests should pass: `TestGoldenFiles/pattern_match_01_simple`
3. Preprocessed code should exclude trailing commas properly

## Testing Plan

1. Fix line ~315: Remove the `strings.TrimSuffix(expr, ",")` line
2. Run: `go test ./tests -run TestGoldenFiles/pattern_match_01_simple -v`
3. Verify test passes without "no pattern arms found" error
4. Run all pattern match tests to ensure no regression:
   - pattern_match_01_simple
   - pattern_match_02_guards
   - pattern_match_05_guards_basic
   - etc.

## Related Functions
- `findExpressionEnd()` - works correctly, finds comma boundaries
- `parseArms()` - main parsing function with the bug
- `transformMatch()` - calls parseArms and fails when empty slice returned

## Impact
Fixes basic pattern matching functionality for pattern_match_01_simple.dingo and similar simple patterns. Unblocks Phase 4.1.2 pattern matching implementation.</content>
</xai:function_callacerbate><|control241|><xai:function_call name="Edit">
<parameter name="file_path">pkg/preprocessor/rust_match.go
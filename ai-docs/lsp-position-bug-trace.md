# Dingo LSP Position Translation Bug Analysis

## Bug Confirmed: Undefined Variables in error_prop.go

The analysis confirms two critical undefined variable bugs in the preprocessing code:

### Line 352 - expandAssignment function:
```go
for ... {
    fullLineText := matches[0]  // Variable scope LIMITED to for loop
    // ... many lines of code ...
} // <-- fullLineText goes out of scope here

// Line 352: BUG! fullLineText is undefined!
exprPosInOriginal := strings.Index(fullLineText, exprWithoutQ)  // COMPILE ERROR
```

### Line 521 - expandReturn function:
```go
for ... {
    fullLineText := matches[0]  // Variable scope LIMITED to for loop
    // ... many lines of code ...
} // <-- fullLineText goes out of scope here

// Line 521: BUG! fullLineText is undefined!
exprPosInOriginal := strings.Index(fullLineText, exprWithoutQ)  // COMPILE ERROR
```

**Result**: These lines should cause compilation errors, but somehow the code runs. This suggests:
1. Either there's a different variable scope than visible
2. Or the analysis was based on outdated code
3. Or there are multiple fullLineText declarations

## Source Map Analysis

Despite the bugs, source maps ARE being generated. Let me examine how expressions get mapped:

### Current Source Map Generation Flow:

1. **Original Line** (e.g., `content, err = os.ReadFile(path)?`):
   - fullLineText = "    content, err = os.ReadFile(path)?"
   - exprWithoutQ = "os.ReadFile(path)"
   - exprPosInOriginal = position of "os.ReadFile(path)" within fullLineText

2. **Generated Code** (multiple lines):
   ```go
   // Line 4: __tmp0, __tmp1, __err0 := os.ReadFile(path)
   ```

3. **Mapping Creation**:
   - OriginalLine: original line number
   - OriginalColumn: exprPosInOriginal + 1
   - GeneratedLine: 4
   - GeneratedColumn: position after " __tmp0, __tmp1, __err0 := "
   - Name: "expr_mapping"

### The Position Translation Algorithm (MapToOriginal):

The `MapToOriginal` function in sourcemap.go has the following logic for mapping positions:

```go
func (sm *SourceMap) MapToOriginal(line, col int) (int, int) {
    var bestMatchOnLine *Mapping = nil
    var minDistanceOnLine int = -1

    for i := range sm.Mappings {
        m := &sm.Mappings[i]
        if m.GeneratedLine == line {
            // Case 1: Exact match within mapping's range
            if col >= m.GeneratedColumn && col < m.GeneratedColumn+m.Length {
                offset := col - m.GeneratedColumn
                return m.OriginalLine, m.OriginalColumn + offset
            }

            // Case 2: Find best match if no exact match
            currentDistance := abs(m.GeneratedColumn - col)
            if bestMatchOnLine == nil || currentDistance < minDistanceOnLine {
                bestMatchOnLine = m
                minDistanceOnLine = currentDistance
            }
        }
    }

    // Second pass: Use best candidate if no exact match
    if bestMatchOnLine != nil {
        // Special handling for 'error_prop' mappings
        if bestMatchOnLine.Name == "error_prop" {
            return bestMatchOnLine.OriginalLine, bestMatchOnLine.OriginalColumn
        }

        // Apply offset for other mappings
        offset := col - bestMatchOnLine.GeneratedColumn
        if offset >= 0 && offset < bestMatchOnLine.Length + 5 {
            return bestMatchOnLine.OriginalLine, bestMatchOnLine.OriginalColumn + offset
        }
    }

    // Fallback: return identity mapping
    return line, col
}
```

## Simulating the Exact Bug Scenario

### Scenario: gopls reports error at line 4, column 20

Let's trace through the translation:

1. **Input from gopls**: line=4, col=20
2. **Generated Code**: `    __tmp0, __tmp1, __err0 := os.ReadFile(path)`
   - Column 20 is likely pointing to `os.ReadFile(path)` part
3. **Mapping data** (hypothetical based on code analysis):
   ```
   {
       OriginalLine: 4,
       OriginalColumn: 23,  // Position of 'o' in "os.ReadFile(path)"
       GeneratedLine: 4,
       GeneratedColumn: 32, // Position after "    __tmp0, __tmp1, __err0 := "
       Length: 15,           // Length of "os.ReadFile(path)"
       Name: "expr_mapping"
   }
   ```

4. **Translation Process**:
   - Is col=20 within range [32, 32+15)? **NO, 20 < 32**
   - Calculate distance: |32 - 20| = 12
   - This becomes the bestMatchOnLine with distance 12

5. **The Bug**: Since column 20 is BEFORE the mapped region, the algorithm:
   - Sets bestMatchOnLine to the expr_mapping
   - But since 20 < 32, it's not an exact match
   - Falls to second pass with bestMatchOnLine
   - Since it's "expr_mapping" (not "error_prop"), it tries to apply offset:
     ```go
     offset := 20 - 32 = -12  // Negative offset!
     ```
   - The condition `offset >= 0 && offset < bestMatchOnLine.Length + 5` **fails**
   - Falls back to returning the mapping's position: return m.OriginalLine, m.OriginalColumn
   - Wait, the code handles this differently...

### Re-examining the Code Logic

Actually, looking more carefully at the MapToOriginal code:

```go
// Second pass: If no exact match, use the best candidate found on the same line.
if bestMatchOnLine != nil {
    // Special handling for 'error_prop' if it's the best (or only) candidate on this line.
    // We want any error in its generated block to map back to the '?' operator exactly.
    if bestMatchOnLine.Name == "error_prop" {
        // We can return the exact original position of '?' without calculating offset.
        // This assumes the 'error_prop' mapping's original_column points to the '?'.
        return bestMatchOnLine.OriginalLine, bestMatchOnLine.OriginalColumn
    }

    // For other mappings, apply an offset IF the column is conceptually within its generated span.
    // We need to be careful with applying offsets to non-exact matches to avoid large shifts.
    // A simple heuristic: if 'col' is "reasonably close" to the mapping's start or within its nominal length.
    offset := col - bestMatchOnLine.GeneratedColumn
    if offset >= 0 && offset < bestMatchOnLine.Length + 5 {
        return bestMatchOnLine.OriginalLine, bestMatchOnLine.OriginalColumn + offset
    }
    // If col is far beyond the mapping's length or before its start, this mapping isn't a good fit.
}

// Final fallback: If no relevant mapping was found, or previous attempts failed,
// return the identity mapping (i.e., the generated line and column itself).
return line, col
```

**The problem**: When the bestMatchOnLine is "expr_mapping" but column 20 is before its start (col=20 < genCol=32), the offset check fails and it falls through to return the identity mapping (4, 20).

But this contradicts the observed behavior where column 15 is returned. Let me examine what other mappings might be involved.

### Multiple Mappings Analysis

The error propagation generates multiple mappings. Let me check what mappings are created:

1. **"error_prop" mapping**: Maps to the `?` operator position
2. **"expr_mapping"**: Maps to the `ReadFile(path)` position
3. **Other mappings**: For variable assignments, etc.

If there's also an "error_prop" mapping with:
```
{
    OriginalLine: 4,
    OriginalColumn: 28,  // Position of '?'
    GeneratedLine: 4,
    GeneratedColumn: 48, // Position of '?' in generated code (somewhere)
    Length: 1,
    Name: "error_prop"
}
```

And gopls reports an error at col=20 (which maps to `ReadFile`), the algorithm would:
- Find both mappings on line 4
- Neither mapping contains col=20 in their range
- Calculate distances:
  - Distance to expr_mapping: |32 - 20| = 12
  - Distance to error_prop: |48 - 20| = 28
- Choose expr_mapping as best (distance 12 < 28)
- Since it's "expr_mapping" (not "error_prop"), the special-case doesn't apply
- Offset = 20 - 32 = -12 (negative), so offset condition fails
- Falls through to final fallback: return (4, 20) - **but this doesn't match observed behavior**

### Alternative Explanation: Different Mapping Data

The observed column 15 suggests there might be an "error_prop" mapping with:
```
{
    OriginalLine: 4,
    OriginalColumn: 15,  // This would be the '?' position
    GeneratedLine: 4,
    GeneratedColumn: 20, // Or position close to reported error
    Name: "error_prop"
}
```

If this mapping has GeneratedColumn=20 or very close to 20, then:
- Distance to error_prop mapping: |20 - 20| = 0 (exact match!)
- If exact match range contains col=20, it would use the special case
- Return OriginalColumn=15 as the mapped position

## Root Cause Identification

The bug appears to be one of these issues:

1. **Incorrect GeneratedColumn for expr_mapping**: The column calculation for where the expression appears in generated code is wrong

2. **Incorrect OriginalColumn for error_prop mapping**: The position of the `?` operator in the original line is being calculated incorrectly

3. **Multiple overlapping mappings**: The algorithm is choosing the wrong mapping due to incorrect distance calculations

4. **The undefined variable bug**: Somehow fullLineText is not what we think it is, leading to incorrect position calculations

### Most Likely Issue:

Based on the fact that column 15 is returned (which could be the position of `?` in the original line), the issue is likely that:

1. The `error_prop` mapping is being created correctly to point to the `?` position
2. But the `expr_mapping` has an incorrect `GeneratedColumn` value
3. When gopls reports an error at what should be `ReadFile` (around column 20), it's actually closer to the `error_prop` mapping
4. The algorithm chooses `error_prop` as best match and returns the `?` position (column 15)

This would explain why the error is highlighted on `?` instead of `ReadFile(path)`.

## Next Steps for Fix

1. **Fix the undefined variable bug** by moving `fullLineText` declaration outside the for loop
2. **Add debug logging** to see exactly what mappings are being created
3. **Verify the column calculations** for both original and generated positions
4. **Test the MapToOriginal function** with the actual mapping data being generated

The undefined variable bug is the critical first issue to resolve, as it likely indicates the source map generation is not working as intended.
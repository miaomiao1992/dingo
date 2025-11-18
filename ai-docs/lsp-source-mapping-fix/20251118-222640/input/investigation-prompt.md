# Source Map Position Translation Issue - Investigation

## Problem Statement

The Dingo LSP is translating diagnostic positions incorrectly. When gopls reports "undefined: ReadFile", the error should underline `ReadFile` in the .dingo file, but instead it's underlining `e(path)?` (the end of the expression).

## Current Architecture

### Files Involved
- **Dingo source**: `let data = ReadFile(path)?`
- **Generated Go**: `__tmp0, __err0 := ReadFile(path)`
- **Source map**: Maps generated positions → original positions

### Position Details

**Dingo file (line 4):**
```
let data = ReadFile(path)?
0         1         2         3
012345678901234567890123456789
```
- `ReadFile` starts at column 11 (0-indexed) or 12 (1-indexed)
- `?` is at column 30 (0-indexed) or 31 (1-indexed)

**Go file (line 4):**
```
__tmp0, __err0 := ReadFile(path)
0         1         2         3
012345678901234567890123456789
```
- `ReadFile` starts at column 18 (0-indexed) or 19 (1-indexed)

### Current Source Map Content

```json
{
  "mappings": [
    {
      "generated_line": 4,
      "generated_column": 1,
      "original_line": 4,
      "original_column": 15,
      "length": 1,
      "name": "error_prop"
    }
  ]
}
```

All generated lines (4-10) map to original column 15 (the `?` position).

## Attempted Fixes

### Fix 1: Smart Fallback in MapToOriginal
Added logic to use identity mapping when position is >10 columns from any mapped region.
**Result**: Failed - identity mapping assumes same column numbers, but columns are different.

### Fix 2: Expression Mapping in Preprocessor
Added code to generate mappings for the expression itself:
```go
// In expandLetStatement and expandReturn
exprPosInOriginal := strings.Index(originalText, exprClean)
if exprPosInOriginal >= 0 {
    origCol := len(indent) + exprPosInOriginal + 1
    genCol := len(indent) + prefixLen + 1
    mappings = append(mappings, Mapping{
        OriginalLine:    originalLine,
        OriginalColumn:  origCol,
        GeneratedLine:   startOutputLine,
        GeneratedColumn: genCol,
        Length:          len(exprClean),
        Name:            "expr_mapping",
    })
}
```
**Result**: Still failing - underline appears on `e(path)?` instead of `ReadFile`.

## Relevant Code

### pkg/preprocessor/error_prop.go (expandLetStatement)
The preprocessor expands `let data = ReadFile(path)?` to:
```go
__tmp0, __err0 := ReadFile(path)
// dingo:s:1
if __err0 != nil {
    return nil, __err0
}
// dingo:e:1
var data = __tmp0
```

### pkg/preprocessor/sourcemap.go (MapToOriginal)
```go
func (sm *SourceMap) MapToOriginal(line, col int) (int, int) {
    for i := range sm.Mappings {
        m := &sm.Mappings[i]
        if m.GeneratedLine == line {
            if col >= m.GeneratedColumn && col < m.GeneratedColumn+m.Length {
                offset := col - m.GeneratedColumn
                return m.OriginalLine, m.OriginalColumn + offset
            }
            // Track closest mapping for fallback
        }
    }
    // Fallback logic here
}
```

### pkg/lsp/handlers.go (TranslateDiagnostics)
```go
func (t *Translator) TranslateDiagnostics(
    diagnostics []protocol.Diagnostic,
    goURI protocol.DocumentURI,
    dir Direction,
) ([]protocol.Diagnostic, error) {
    for _, diag := range diagnostics {
        _, newRange, err := t.TranslateRange(goURI, diag.Range, dir)
        if err != nil {
            continue
        }
        diag.Range = newRange
        translatedDiagnostics = append(translatedDiagnostics, diag)
    }
    return translatedDiagnostics, nil
}
```

## Question

**What is the root cause of the incorrect position mapping, and what is the best solution?**

Consider:
1. Is the expression mapping calculation correct?
2. Should we use a different source map format or algorithm?
3. Is there an issue with how diagnostics are being translated?
4. Should we map character-by-character or use a different granularity?
5. Are there edge cases with indentation or whitespace?
6. Should we use a standard source map format (like JavaScript source maps)?

Please provide:
1. **Root cause analysis** - Why is the current approach failing?
2. **Specific solution** - Exact code changes needed
3. **Verification** - How to test the fix works correctly

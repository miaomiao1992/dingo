# Internal Analysis (Claude Sonnet 4.5) - Source Map Position Bug

## Root Cause: Double-Counting Indent in Position Calculation

### The Bug

In `pkg/preprocessor/error_prop.go`, the expression mapping calculation has an **off-by-1 error** caused by double-counting the indentation:

```go
// WRONG CODE (current):
originalText := matches[0]  // e.g., "\tlet data = ReadFile(path)?"
exprPosInOriginal := strings.Index(originalText, exprClean)  // Position within originalText
origCol := len(indent) + exprPosInOriginal + 1  // ❌ DOUBLE COUNTS INDENT
```

### Why This Fails

1. `originalText` = `"\tlet data = ReadFile(path)?"` (includes indent)
2. `exprClean` = `"ReadFile(path)"`
3. `exprPosInOriginal` = `strings.Index(originalText, exprClean)` = **12** (0-indexed position within originalText)
   - This position ALREADY accounts for the indent character!
4. `len(indent)` = **1** (the tab character)
5. `origCol` = 1 + 12 + 1 = **14** (1-indexed) ❌

**Expected:** Column 13 (1-indexed) or 12 (0-indexed) for where `ReadFile` starts.

**Actual Result:** Column 14, which is off by 1!

### Position Analysis

**Original line:**
```
\tlet data = ReadFile(path)?
0 1       2       3
0123456789012345678901234567890
```
- Position 0: `\t` (tab)
- Position 12: `R` (start of `ReadFile`) ← 0-indexed
- Position 13: `R` (start of `ReadFile`) ← 1-indexed ✅ TARGET

**Current buggy calculation:**
- `exprPosInOriginal` = 12 (correct - position of 'R' in originalText)
- `origCol` = len("\t") + 12 + 1 = **14** ❌ (wrong!)

**Correct calculation:**
- `origCol` = 12 + 1 = **13** ✅ (just convert 0-indexed to 1-indexed)

## The Fix

### File: `pkg/preprocessor/error_prop.go`

**Location 1: `expandLetStatement` (around line 343-363)**

```go
// BEFORE (WRONG):
originalText := matches[0]
exprPosInOriginal := strings.Index(originalText, exprClean)
if exprPosInOriginal >= 0 {
    prefixLen := len(tmpVar) + len(", ") + len(errVar) + len(" := ")
    genCol := len(indent) + prefixLen + 1

    origCol := len(indent) + exprPosInOriginal + 1  // ❌ BUG HERE

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

```go
// AFTER (FIXED):
originalText := matches[0]
exprPosInOriginal := strings.Index(originalText, exprClean)
if exprPosInOriginal >= 0 {
    prefixLen := len(tmpVar) + len(", ") + len(errVar) + len(" := ")
    genCol := len(indent) + prefixLen + 1

    origCol := exprPosInOriginal + 1  // ✅ FIX: Don't double-count indent

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

**Location 2: `expandReturn` (around line 506-526)**

Same fix needed:

```go
// BEFORE (WRONG):
origCol := len(indent) + exprPosInOriginal + 1  // ❌ BUG

// AFTER (FIXED):
origCol := exprPosInOriginal + 1  // ✅ FIX
```

## Verification Steps

1. **Apply the fix** to both locations in `error_prop.go`
2. **Rebuild** the transpiler:
   ```bash
   go build -o dingo ./cmd/dingo
   ```
3. **Retranspile** a test file:
   ```bash
   ./dingo build tests/golden/error_prop_01_simple.dingo
   ```
4. **Check the source map**:
   ```bash
   cat tests/golden/error_prop_01_simple.go.map
   ```
   Should now show a mapping with:
   - `original_column: 13` (for `ReadFile` in Dingo file, 1-indexed)
   - `generated_column: 20` (for `ReadFile` in Go file, 1-indexed)
   - `length: 14` (length of "ReadFile(path)")

5. **Restart LSP** and test hover - error should now underline `ReadFile` correctly! ✅

## Why This Happened

The confusion arose from mixing two different coordinate systems:
1. **Position within `originalText`** (which includes indent)
2. **Position within the actual file line** (also includes indent)

Since `strings.Index(originalText, exprClean)` already returns a position that accounts for everything in `originalText` (including indent), we don't need to add `len(indent)` again.

## Summary

**Problem:** Off-by-1 error due to double-counting indentation
**Solution:** Remove `len(indent) +` from the `origCol` calculation
**Impact:** Diagnostic positions will now map correctly from Go → Dingo
**Risk:** Very low - simple arithmetic fix, easy to verify

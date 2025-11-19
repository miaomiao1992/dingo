# Consolidated Analysis: LSP Source Mapping Bug

## Executive Summary

**8 architect agents investigated in parallel** (1 internal + 7 external models). **7 completed successfully**, 1 timed out.

## Root Causes Identified (Consensus)

### Primary Issue: Incorrect Column Position Calculation in Error Propagation Preprocessor

**Consensus from 6/7 models:**

The error propagation preprocessor (`pkg/preprocessor/error_prop.go`) generates **incorrect column positions** in the source map when transforming the `?` operator:

```dingo
contents, e := ReadFile(path)?   // Dingo source
```

Becomes:
```go
contents, e := ReadFile(path)    // Generated Go
if e != nil {
    return Result[[]byte]{Err: &e}
}
```

**The Problem:**
- The source map records the position of `?` operator in Dingo code
- When gopls reports an error on the function call line (`ReadFile(path)`)
- The reverse mapping finds the closest source map entry
- Due to coarse/incorrect mappings, it maps to the `?` operator position instead of `ReadFile`

### Specific Root Causes by Model:

1. **Internal Analysis** - `qPos` calculation error causes wrong column positions
2. **GPT-5.1 Codex** - Single coarse mapping (all lines → `?`), needs granular mappings
3. **Gemini 2.5 Flash** - Fallback offset calculation produces incorrect positions
4. **Grok Code Fast** - `generated_column` values don't match actual Go positions
5. **MiniMax M2** - Wrong column calculation (15 instead of 27)
6. **Sherlock Think** - 0-based vs 1-based indexing mismatch
7. **GLM-4.6** - MapToOriginal prioritizes error_prop mappings incorrectly

## Recommended Fix Strategy

### Option 1: Granular Source Mappings (Preferred by GPT-5.1 Codex)

Emit **multiple source map entries** for error propagation expansion:

```
Original Dingo:  contents, e := ReadFile(path)?
                 ^                 ^           ^
                 mapping 1         mapping 2   mapping 3

Generated Go:    contents, e := ReadFile(path)  // ← Map to Dingo col 20 (ReadFile)
                 if e != nil {                  // ← Map to Dingo col 35 (?)
                     return ...                 // ← Map to Dingo col 35 (?)
                 }
```

### Option 2: Fix Column Calculation (Preferred by Internal + Others)

Correct the `qPos` calculation in `error_prop.go`:

```go
// Current (WRONG):
qPos := strings.Index(line, "?")  // Finds first ?, wrong when multiple

// Fixed:
qPos := strings.LastIndex(line, "?")  // Find the actual error-prop operator
```

### Option 3: Fix MapToOriginal Logic (GLM-4.6)

Update source map reverse mapping to prefer expression mappings over operator mappings when the diagnostic is on a function call.

## Next Steps

1. **Verify the actual bug** - Examine `error_prop_01_simple.dingo` and its source map
2. **Test Option 2 first** (simplest fix) - Correct `qPos` calculation
3. **Implement Option 1** (comprehensive) - Granular mappings for all expanded code
4. **Add regression tests** - Ensure LSP diagnostics point to correct locations

## Session Files

- Internal analysis: `ai-docs/sessions/20251118-223538/01-planning/internal-analysis.md`
- GPT-5.1 Codex: `ai-docs/sessions/20251118-223538/01-planning/gpt-5.1-codex-analysis.md`
- Gemini 2.5 Flash: `ai-docs/sessions/20251118-223538/01-planning/gemini-2.5-flash-analysis.md`
- Grok Code Fast: `ai-docs/sessions/20251118-223538/01-planning/grok-code-fast-1-analysis.md`
- MiniMax M2: `ai-docs/sessions/20251118-223538/01-planning/minimax-m2-analysis.md`
- Sherlock Think: `ai-docs/sessions/20251118-223538/01-planning/sherlock-think-analysis.md`
- GLM-4.6: `ai-docs/sessions/20251118-223538/01-planning/glm-4.6-analysis.md`
- Qwen3 Coder: Failed (timeout)

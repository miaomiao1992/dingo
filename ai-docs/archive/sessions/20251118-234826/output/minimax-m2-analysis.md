
[claudish] Model: minimax/minimax-m2

## Investigation Complete: Alternative Solutions for Match-in-Assignment Bug

I've completed a comprehensive investigation of 7 alternative approaches to fix the pattern matching assignment bug. Here's my recommendation:

### Quick Summary

**Recommendation**: **Alternative 5 - Refined Temporary Variable Pattern**

**Key Benefits**:
- ✅ **Minimal changes** (1-2 hours implementation)
- ✅ **Clean, idiomatic Go** (no function wrapper)
- ✅ **Low risk** (preserves all existing functionality)
- ✅ **Solves the syntax error** (`var result = __match_N := expr`)

### How It Works

**Problem**:
```go
var result = __match_3 := opt  // ❌ Invalid syntax
```

**Solution**:
```go
__match_0 := opt
var result Option_int
switch __match_0.tag {
case OptionTagSome:
    x := *__match_0.some_0
    result = Some(x * 2)  // Assign in each branch
case OptionTagNone:
    result = None
}
```

**Key Insight**: Instead of skipping temp var generation, we:
1. Generate temp var FIRST: `__match_0 := opt`
2. Generate variable declaration SECOND: `var result Type`
3. Switch assigns to result in each branch

### Why This Beats the Alternatives

| Alternative | Time | Cleanliness | Issue |
|-------------|------|-------------|-------|
| **Alt 5 (Winner)** | 1-2h | 8/10 | ✅ Minimal |
| Alt 1 (Hoisting) | 4-6h | 9/10 | Complex type inference |
| Alt 2 (IIFE) | 2-3h | 7/10 | User said "ugly" wrapper |
| Alt 3 (Statements) | 4-5h | 8/10 | Breaks expression semantics |

### Implementation Plan

**Phase 1** (30 min): Modify `isInAssignmentContext()` to extract variable name  
**Phase 2** (30 min): Update `generateSwitch()` to emit var declaration after temp var  
**Phase 3** (30-60 min): Add basic type inference  
**Phase 4** (30 min): Run tests  

**Total: ~2-3 hours**

### Full Analysis

The complete investigation with all 7 alternatives, code examples, comparison matrix, and edge cases is saved to:
**`/tmp/investigation_analysis.md`**

All details, pros/cons, and implementation specifics are documented there.

**Next Step**: Should I proceed with implementing Alternative 5 in `rust_match.go`?

[claudish] Shutting down proxy server...
[claudish] Done


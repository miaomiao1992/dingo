# User Request: Investigate Alternative Match-in-Assignment Fix

## Current Problem

**Bug**: `var result = __match_3 := opt` - invalid Go syntax

**Root Cause**: Dingo treats match as an EXPRESSION (returns value), but Go's switch is a STATEMENT (doesn't return value). The current assignment context fix creates malformed code.

## Recommended Solution (IIFE Pattern)

```go
result := func() interface{} {
    __match_0 := opt
    switch __match_0.tag {
    case OptionTagSome:
        x := *__match_0.some_0
        return Some(x * 2)  // Add 'return'
    case OptionTagNone:
        return None         // Add 'return'
    }
    panic("non-exhaustive match")
}()
```

**User Concern**: "Result code looks a little bit ugly with wrap in function"

## User's Request

Investigate **alternative approaches** that:
1. Fix the match-in-assignment bug
2. Generate **cleaner, more idiomatic Go code** than IIFE wrapper
3. Still support match as an expression (returns a value)

## Constraints

- ✅ Must fix: `var result = __match_3 := opt` syntax error
- ✅ Must maintain: 12 currently passing tests
- ✅ Must achieve: 13/13 tests passing
- ✅ Prefer: Cleaner output than IIFE wrapper
- ⚠️ Unknown: Time/complexity trade-off vs IIFE solution

## Success Criteria

1. Solution generates idiomatic, readable Go code
2. All 13 pattern matching tests pass
3. No regression in existing functionality
4. Code is cleaner/simpler than IIFE wrapper pattern

## Context

- Current status: 12/13 tests passing (92%)
- Failing test: Match-in-assignment context
- IIFE solution: Proven, 2-3 hours, but "ugly"
- User preference: Explore cleaner alternatives first

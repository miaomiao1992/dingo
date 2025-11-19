# Task A1: Guard Pattern Parsing - Implementation Notes

## Design Decisions

### 1. Guard Keyword Detection Strategy

**Challenge:** Distinguish guard keywords ('if', 'where') from same words in identifiers or function names.

**Solution:**
- Require spaces around keywords: ` if ` and ` where `
- Find keyword AFTER pattern completes (after closing paren)
- Use `isCompletePattern()` to validate pattern ends before keyword

**Examples that work correctly:**
- `Ok(x) if x > 0` → pattern="Ok(x)", guard="x > 0" ✓
- `Ok(x) where isValid(x)` → pattern="Ok(x)", guard="isValid(x)" ✓
- `Ok(diff)` → pattern="Ok(diff)", guard="" (no false match on 'if' in 'diff') ✓
- `Ok(somewhere)` → pattern="Ok(somewhere)", guard="" (no false match on 'where') ✓

### 2. Marker Format

**Format:** `// DINGO_PATTERN: Pattern | DINGO_GUARD: condition`

**Rationale:**
- Single-line comment keeps preprocessor output simple
- Pipe separator (`|`) clearly delimits pattern and guard
- Plugin can easily parse marker with string split
- Consistent with existing DINGO_PATTERN marker style

**Example:**
```go
case ResultTagOk:
    // DINGO_PATTERN: Ok(x) | DINGO_GUARD: x > 0
    x := *__match_0.ok_0
    x * 2
```

### 3. Both 'if' and 'where' Supported

**Decision:** Support both keywords (not just one)

**Rationale:**
- Rust uses 'if' for guards
- Swift uses 'where' for guards
- Plan includes Swift syntax support (Phase 4.2 Feature 2)
- Both keywords are normalized to same marker format
- Plugin implementation doesn't need to know which keyword was used

**Implementation:** Both detected with same logic, just different keyword length

### 4. Guard Condition Preservation

**Decision:** Store guard condition as-is (no validation or parsing)

**Rationale:**
- Preprocessor's job is text transformation, not semantic analysis
- Plugin will parse guard condition using go/parser for validation
- Keeps preprocessor simple and fast
- Allows arbitrary Go expressions in guards

**Examples:**
- Simple: `x > 0`
- Complex: `x > 0 && x < 100`
- Function calls: `isValid(x)`
- Field access: `user.age > 18 && user.verified`

All stored as strings for plugin to parse.

## Deviations from Plan

### Minor Adjustment: Keyword Detection

**Plan:** Use regex pattern matching

**Actual:** Used string search with validation

**Reason:**
- Simpler and more maintainable
- Easier to debug
- Better performance (no regex compilation)
- Still handles all required cases

### Minor Adjustment: ' where ' Length

**Plan:** Not specified

**Implementation:** ` where ` is 7 characters (not 8)

**Discovery:** During testing, found that " where " is actually 7 characters:
- 1 space
- 5 letters (where)
- 1 space
= 7 total

**Fix:** Corrected `guardKeywordLen` from 8 to 7

## Testing Notes

### Test Coverage

**10 test functions created:**
1. Split pattern/guard with 'if' keyword (5 cases)
2. Split pattern/guard with 'where' keyword (4 cases)
3. End-to-end 'if' guard processing
4. End-to-end 'where' guard processing
5. Multiple guards per variant
6. Complex guard expressions
7. Guards with block expressions
8. Arm parsing with guards
9. Both keywords in same match

**All tests pass:** 100% pass rate

### Edge Cases Tested

- ✓ No guard (backward compatibility)
- ✓ Simple guards (x > 0)
- ✓ Complex boolean expressions (x > 0 && x < 100)
- ✓ Function calls in guards (isValid(x), len(user.name) > 0)
- ✓ Field access in guards (user.age > 18)
- ✓ Multiple guards on same variant
- ✓ Guards with block expressions
- ✓ Mixed 'if' and 'where' in same match
- ✓ Patterns without parentheses (None, _)
- ✓ Avoiding false matches ('if' in 'diff', 'where' in 'somewhere')

## Performance

**Impact:** Minimal
- String search is O(n) where n is pattern length (typically <50 chars)
- No regex compilation overhead
- Guard extraction happens once per arm during preprocessing
- Estimated overhead: <0.1ms per match expression

## Next Steps

Task A1 is complete. Ready for:
- **Task B2:** Plugin implementation (GuardTransformer to generate nested if statements)
- **Integration:** Guard markers are ready for plugin consumption

## Files for Plugin Developer

The plugin implementation (Task B2) will need to:
1. Detect `DINGO_GUARD: <condition>` markers in case comments
2. Parse `<condition>` using `go/parser.ParseExpr()`
3. Generate nested if statement wrapping the case body
4. Handle guard failures (continue to next case if condition false)

Marker format is well-defined and ready for consumption.

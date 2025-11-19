# Priority 3: Implementation Notes

## Debugging Process

### Step 1: Initial Test Run
Discovered 3/4 tests failing:
- `pattern_match_rust_syntax` - Parser error
- `pattern_match_non_exhaustive_error` - Plugin not finding markers
- `combined_pattern_match_and_none` - Parser error

### Step 2: Comma Parsing Investigation
Error message: `expected operand, found ','`

Traced to Rust match syntax with trailing commas:
```dingo
Ok(value) => {
    return fmt.Sprintf("Success: %d", value)
},  // <-- This comma was being included in the tag name
```

Found that block expressions (lines 297-310) didn't skip commas after `}`, while simple expressions did (line 322).

**Fix**: Added 3-line comma skip after block expression parsing.

### Step 3: Plugin Discovery Failure
Pattern match plugin debug log: "Found 0 match expressions"

But preprocessor was generating markers! Traced plugin code:
- `findMatchMarker()` needs `ctx.CurrentFile` to access comment list
- Integration tests weren't setting this field
- Unit tests had it (`CurrentFile: file` in all pattern_match_test.go cases)

**Fix**: Added `ctx.CurrentFile = file` in all 4 integration test cases.

### Step 4: Panic Statement Missing
Tests expect `panic("unreachable: match is exhaustive")` after switch.

Checked preprocessor output - no panic generated for regular matches!
- Tuple match generator (line 1540) has panic logic
- Regular match generator (line 596) doesn't

**Fix**: Added panic generation matching tuple behavior.

## Key Insights

1. **Consistency is Critical**: Features in one code path (tuple) should exist in parallel paths (regular)
2. **Integration vs Unit Tests**: Integration tests need complete setup (CurrentFile, etc.), not just partial context
3. **Preprocessor Symmetry**: Block and simple expressions should have parallel handling
4. **Comment Access Pattern**: Plugins accessing file comments REQUIRE `ctx.CurrentFile` set

## Test Verification

### Manual Preprocessor Test
Created standalone test to verify trailing comma handling:
```bash
go run /tmp/test_rust_match.go
```

Output confirmed panic statement now present:
```go
}
panic("unreachable: match is exhaustive")
// DINGO_MATCH_END
```

### Integration Test Suite
All 4 tests passing after fixes:
- Comma parsing: ✅
- Plugin discovery: ✅
- Panic generation: ✅
- None inference: ✅ (already working)

## Edge Cases Considered

1. **Nested Block Expressions**: Comma skip only at top level (depth 0)
2. **Multiple Trailing Commas**: Only skip one comma per arm
3. **Assignment Context**: Panic placement before assignment-context return
4. **Guard Patterns**: Panic applies to all match types (guards, nested, etc.)

## Performance Impact

- **Trailing Comma Skip**: O(1) per arm - negligible
- **CurrentFile Assignment**: One-time setup - zero impact
- **Panic Generation**: One extra line per match - negligible

## Compatibility

All fixes maintain backward compatibility:
- Trailing commas are optional (still support no comma)
- CurrentFile is additive (doesn't break existing code)
- Panic statement is unreachable (never executed in correct matches)

## Future Considerations

1. **Preprocessor Refactoring**: Consider extracting common "skip trailing delimiter" logic
2. **Test Helpers**: Create `setupTestContext()` helper that sets all required fields
3. **Panic Optimization**: Consider making panic conditional on analysis (only if needed for control flow)

## Related Issues

- **C7 Fix**: Variable hoisting pattern uses similar assignment context detection
- **Priority 1 Fix**: Deterministic output (sorting) complements these fixes
- **Priority 2 Fix**: Type declaration ordering shares injection pattern

## Lessons Learned

1. When preprocessing block vs. simple expressions, maintain symmetry
2. Integration tests should mirror unit test setup patterns
3. Feature parity across code paths prevents subtle bugs
4. Debug logs (`Found 0 match expressions`) are critical for diagnosing plugin issues

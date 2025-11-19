# Priority 3: Fix Integration Test Failures - Changes

## Summary
Fixed all 4 integration test failures in `TestIntegrationPhase4EndToEnd` by addressing trailing comma parsing, missing `CurrentFile` context, and missing panic statements.

## Test Results
- ✅ **pattern_match_rust_syntax** - PASS
- ✅ **pattern_match_non_exhaustive_error** - PASS
- ✅ **none_context_inference_return** - PASS (already passing from Priority 4)
- ✅ **combined_pattern_match_and_none** - PASS

## Changes Made

### 1. Fix Trailing Comma Parsing in Rust Match Preprocessor

**File**: `pkg/preprocessor/rust_match.go`

**Issue**: Block expressions in Rust match syntax (`{ ... },`) weren't skipping the trailing comma, causing parse errors like `expected operand, found ','`.

**Fix**: Added comma-skipping after block expression parsing (lines 311-314):

```go
expr = strings.TrimSpace(text[start:i])
// PRIORITY 3 FIX: Skip trailing comma after block expression
if i < len(text) && text[i] == ',' {
    i++
}
```

**Impact**: Allows Rust-style trailing commas after match arms: `Ok(x) => { ... },`

---

### 2. Add Missing `CurrentFile` Context in Integration Tests

**File**: `tests/integration_phase4_test.go`

**Issue**: Pattern match plugin requires `ctx.CurrentFile` to locate `DINGO_MATCH_START` markers in comments, but integration tests didn't set this field.

**Fix**: Added `ctx.CurrentFile = file` after `BuildParentMap` in 4 test cases (lines 71, 158, 225, 339):

```go
ctx.BuildParentMap(file)
ctx.CurrentFile = file // PRIORITY 3 FIX: Plugin needs CurrentFile to find markers
```

**Impact**: Enables pattern match plugin discovery phase to find markers and detect match expressions.

**Test Locations**:
1. `pattern_match_rust_syntax` - line 71
2. `pattern_match_non_exhaustive_error` - line 158
3. `none_context_inference_return` - line 225
4. `combined_pattern_match_and_none` - line 339

---

### 3. Add Panic Statement After Exhaustive Matches

**File**: `pkg/preprocessor/rust_match.go`

**Issue**: Regular (non-tuple) match expressions didn't generate `panic("unreachable: match is exhaustive")` after the switch, causing control flow analysis failures in Go when all arms return.

**Fix**: Added panic statement generation in `generateSwitch` function (lines 607-617):

```go
// PRIORITY 3 FIX: Add panic for exhaustiveness (Go doesn't know switch is exhaustive)
buf.WriteString("panic(\"unreachable: match is exhaustive\")\n")
mappings = append(mappings, Mapping{
    OriginalLine:    originalLine,
    OriginalColumn:  1,
    GeneratedLine:   outputLine,
    GeneratedColumn: 1,
    Length:          5,
    Name:            "rust_match_panic",
})
outputLine++
```

**Impact**: Satisfies Go's control flow requirements for functions that return in all match arms.

**Example Generated Code**:
```go
switch __match_0.tag {
case ResultTag_Ok:
    value := *__match_0.ok_0
    { return fmt.Sprintf("Success: %d", value) }
case ResultTag_Err:
    err := __match_0.err_0
    { return fmt.Sprintf("Error: %v", err) }
}
panic("unreachable: match is exhaustive")
```

---

## Files Modified

1. `pkg/preprocessor/rust_match.go` - 2 changes (comma parsing + panic generation)
2. `tests/integration_phase4_test.go` - 4 changes (CurrentFile context)

## Testing

### Before Fixes
```
FAIL: pattern_match_rust_syntax - Parser error with comma
FAIL: pattern_match_non_exhaustive_error - Plugin not detecting matches
PASS: none_context_inference_return - Already working
FAIL: combined_pattern_match_and_none - Parser error + missing panic
```

### After Fixes
```
PASS: pattern_match_rust_syntax (0.05s)
PASS: pattern_match_non_exhaustive_error (0.00s)
PASS: none_context_inference_return (0.00s)
PASS: combined_pattern_match_and_none (0.00s)
```

All 4 integration tests passing successfully.

## Root Cause Analysis

### Issue 1: Trailing Commas
- **Root Cause**: Preprocessor had different handling for block vs. simple expressions
- **Why it happened**: Simple expressions skip commas (line 322), but block expressions didn't
- **Why it matters**: Rust/Swift-style trailing commas are idiomatic and should be supported

### Issue 2: Missing CurrentFile
- **Root Cause**: Integration tests were incomplete - unit tests had this, integration didn't
- **Why it happened**: Tests were manually building plugin pipeline, skipped standard setup
- **Why it matters**: Plugin uses `CurrentFile` to access comment list for marker detection

### Issue 3: Missing Panic
- **Root Cause**: Tuple match generator had panic logic, regular match didn't
- **Why it happened**: Likely copy-paste oversight when implementing tuple support
- **Why it matters**: Go requires unreachable code path after exhaustive switches that all return

## Prevention Strategies

1. **Consistency Checks**: When adding features to one code path (tuple), check parallel paths (regular)
2. **Integration Test Completeness**: Ensure integration tests set up full context (CurrentFile, etc.)
3. **Preprocessor Parity**: Keep block/simple expression handling symmetric

## Related Work

This fixes align with:
- Priority 1: Non-deterministic output (tuple variant sorting)
- Priority 2: Type declaration ordering
- Priority 4: None context inference (already passing)

All Priority 3 fixes are minimal, surgical changes that don't affect other functionality.

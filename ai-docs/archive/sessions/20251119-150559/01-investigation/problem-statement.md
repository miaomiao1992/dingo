# Investigation: "No Pattern Arms Found" Error

## Problem Description

We have 6 failing pattern matching tests that all fail with the same error:
```
parse error: rust_match preprocessing failed: line XX: parsing pattern arms: no pattern arms found
```

## Failing Tests

1. `pattern_match_01_simple.dingo` - Error at line 21
2. `pattern_match_04_exhaustive.dingo` - Error at line 65
3. `pattern_match_05_guards_basic.dingo` - Error at line 55
4. `pattern_match_06_guards_nested.dingo` - (Not yet tested)
5. `pattern_match_07_guards_complex.dingo` - (Not yet tested)
6. `pattern_match_08_guards_edge_cases.dingo` - (Not yet tested)

## Context

- **Current Status**: 95/103 tests passing (92.2%)
- **Previous Session**: Just fixed 4 major priorities (non-determinism, preprocessor bugs, integration tests, None inference)
- **Remaining Work**: Fix these 6 source file issues to reach 98%+ passing

## Example Failing Code (pattern_match_01_simple.dingo)

The error occurs at line 21, which is inside this function:

```dingo
func handleStatus(status: Status) -> string {
	match status {
		Active => "running",
		Pending => "waiting",
		_ => "unknown"
	}
}
```

This is a **return match** expression (no explicit `return` keyword, just match as expression).

## Example Failing Code (pattern_match_04_exhaustive.dingo)

Error at line 65 (after the enum definition):

```dingo
enum Color {
	Red,
	Green,
	Blue,
	RGB { r: int, g: int, b: int },
}

func colorToHex(c Color) string {
	return match c {
		Color_Red => "#FF0000",
		Color_Green => "#00FF00",
		Color_Blue => "#0000FF",
		Color_RGB{r, g, b} => "#" + toHex(r) + toHex(g) + toHex(b),
	}
}
```

This has **explicit return** keyword.

## Example Failing Code (pattern_match_05_guards_basic.dingo)

Error at line 55:

```dingo
func classifyNumber(result Result) string {
	return match result {
		Result_Ok(x) if x > 0 => "positive",
		Result_Ok(x) if x < 0 => "negative",
		Result_Ok(_) => "zero",
		Result_Err(e) => "error",
	}
}
```

This has **pattern guards** (`if` conditions).

## Preprocessor Code Location

The error comes from `pkg/preprocessor/rust_match.go`:

- Line 357: `return nil, fmt.Errorf("no pattern arms found")`
- Function: `parseArms()`
- Regex pattern: `matchExprPattern = regexp.MustCompile(\`(?s)match\s+([^{]+)\s*\{(.+)\}\`)`

## Working Examples (For Comparison)

These pattern match tests ARE working:

1. `pattern_match_01_basic.dingo` - PASSING
2. `pattern_match_02_guards.dingo` - PASSING
3. `pattern_match_03_nested.dingo` - PASSING (fixed in previous session)

## Key Questions to Investigate

1. **Why does the regex fail to extract arms text?**
   - The pattern `(?s)match\s+([^{]+)\s*\{(.+)\}` should capture everything inside `{ }`
   - But somehow `armsText` is empty or malformed, leading to 0 arms parsed

2. **What's different about failing vs. passing tests?**
   - Compare working tests (01_basic, 02_guards, 03_nested) with failing ones
   - Identify syntax patterns that break the parser

3. **Is it a multiline issue?**
   - The `(?s)` flag should handle multiline, but maybe not working correctly?
   - Are newlines causing the regex to fail?

4. **Is it the collectMatchExpression function?**
   - This function collects the match expression across multiple lines
   - Maybe it's not collecting the full expression properly?

5. **Could it be trailing commas or formatting?**
   - Some patterns have trailing commas, some don't
   - Could this affect parsing?

## Investigation Tasks

For each external model consultation:

1. **Read the failing test files** (pattern_match_01_simple, 04, 05)
2. **Read the passing test files** (pattern_match_01_basic, 02, 03)
3. **Read preprocessor code**: `pkg/preprocessor/rust_match.go`
4. **Analyze the difference**: Why do some work and others fail?
5. **Propose root cause** and **concrete fix**
6. **Provide code patch** if possible

## Expected Output Format

Please provide:

1. **Root Cause Analysis** (2-3 paragraphs)
2. **Comparison**: What's different between working and failing tests?
3. **Proposed Solution** (specific code changes)
4. **Code Patch** (if applicable - actual diff or new code)
5. **Validation Strategy** (how to verify the fix works)

## Files to Examine

**Failing Tests**:
- `tests/golden/pattern_match_01_simple.dingo`
- `tests/golden/pattern_match_04_exhaustive.dingo`
- `tests/golden/pattern_match_05_guards_basic.dingo`

**Passing Tests** (for comparison):
- `tests/golden/pattern_match_01_basic.dingo`
- `tests/golden/pattern_match_02_guards.dingo`
- `tests/golden/pattern_match_03_nested.dingo`

**Preprocessor**:
- `pkg/preprocessor/rust_match.go` (focus on `collectMatchExpression`, `transformMatch`, `parseArms`)

## Success Criteria

A successful investigation should:

1. Identify the **exact reason** why `parseArms()` returns 0 arms
2. Explain why working tests pass but failing tests don't
3. Provide **testable fix** that can be implemented immediately
4. Ideally achieve +6 tests passing (95 → 101, reaching 98%)

# Pattern Matching Error Investigation

## Problem Summary
6 failing pattern matching tests all fail with the same error:
```
parse error: rust_match preprocessing failed: line XX: parsing pattern arms: no pattern arms found
```

## Failing Tests (All FAILED)
1. `pattern_match_01_simple.dingo` - Error at line 21
2. `pattern_match_04_exhaustive.dingo` - Error at line 65
3. `pattern_match_05_guards_basic.dingo` - Error at line 55
4. `pattern_match_06_guards_nested.dingo` - (Not yet tested)
5. `pattern_match_07_guards_complex.dingo` - (Not yet tested)
6. `pattern_match_08_guards_edge_cases.dingo` - (Not yet tested)

## Working Tests (All PASSED - For Comparison)
1. `pattern_match_01_basic.dingo` - PASSING
2. `pattern_match_02_guards.dingo` - PASSING
3. `pattern_match_03_nested.dingo` - PASSING (fixed in previous session)

## Files to Read

### Failing Tests:
- `tests/golden/pattern_match_01_simple.dingo`
- `tests/golden/pattern_match_04_exhaustive.dingo`
- `tests/golden/pattern_match_05_guards_basic.dingo`

### Passing Tests:
- `tests/golden/pattern_match_01_basic.dingo`
- `tests/golden/pattern_match_02_guards.dingo`
- `tests/golden/pattern_match_03_nested.dingo`

### Preprocessor Code:
- `pkg/preprocessor/rust_match.go` (focus on functions: `collectMatchExpression`, `transformMatch`, `parseArms`)

## Error Details

### Error Location (rust_match.go:357)
```go
return nil, fmt.Errorf("no pattern arms found")
```

### Regex Pattern Used
```go
matchExprPattern = regexp.MustCompile(`(?s)match\s+([^{]+)\s*\{(.+)\}`)
```

This pattern should:
- Match `match keyword followed by expression followed by { arms }`
- Capture expression in group 1: `([^{]+)`
- Capture arms text in group 2: `(.+)`

### Key Hypothesis
The regex pattern `(.+)` in group 2 should capture all text inside `{ }`, but something is preventing this from working on certain pattern match syntaxes.

## Example Failing Code

### pattern_match_01_simple.dingo (line 21)
```dingo
func handleStatus(status: Status) -> string {
	match status {
		Active => "running",
		Pending => "waiting",
		_ => "unknown"
	}
}
```

### pattern_match_04_exhaustive.dingo (line 65)
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

### pattern_match_05_guards_basic.dingo (line 55)
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

## Investigation Tasks

1. **Compare failing vs passing tests** - What's different in syntax?
2. **Examine collectMatchExpression function** - Does it properly collect multiline expressions?
3. **Test regex pattern** - Does the pattern `(?s)match\s+([^{]+)\s*\{(.+)\}` work correctly?
4. **Identify edge cases** - What syntax breaks the parser?

## Expected Output

Please provide:

1. **Root Cause Analysis** - Why does parseArms() return 0 arms?
2. **Syntax Comparison** - What's different about failing vs passing tests?
3. **Proposed Fix** - Concrete code changes needed
4. **Code Patch** - Actual implementation (Go code)

## Success Criteria
- Identify exact reason why regex fails
- Explain why some tests work while others don't
- Provide testable fix that makes all 6 tests pass
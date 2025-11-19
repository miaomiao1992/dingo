# User Request: Fix Where Guard Preprocessor

## Objective
Fix the `where` guard preprocessor to implement proper support for Swift-style pattern guards.

## The Issue

**Test File**: `tests/golden/pattern_match_06_guards_nested.dingo`

**Problem**: Test uses `where` keyword for pattern guards (Swift-style syntax):
```dingo
Option_Some(x) where x > 100 => "large"
```

**Current Behavior**: The preprocessor has a bug generating Go code for `where` guards.

**Error**: Generated code has syntax error:
```
expected ';', found 'else' at line 103
```

**Status**: This is a separate preprocessor bug unrelated to the two bugs we just fixed today:
1. ✅ Fixed: Match reprocessing bug (preprocessor detecting "match " in panic statements)
2. ✅ Fixed: Result/Option naming inconsistency (underscored vs concatenated)

**Current Test Status**: 102/103 tests passing (99.0%)
- This is the ONLY remaining test failure

## Example Failing Code

From `tests/golden/pattern_match_06_guards_nested.dingo`:

```dingo
// Using 'where' keyword (Swift-style)
func analyzeValue(opt Option) string {
	return match opt {
		Option_Some(x) where x > 100 => "large",
		Option_Some(x) where x > 10 => "medium",
		Option_Some(x) where x > 0 => "small",
		Option_Some(_) => "non-positive",
		Option_None => "none",
	}
}

// Guard fallthrough demonstration
func categorize(opt Option) string {
	return match opt {
		Option_Some(x) where x%3 == 0 && x%5 == 0 => "fizzbuzz",
		Option_Some(x) where x%3 == 0 => "fizz",
		Option_Some(x) where x%5 == 0 => "buzz",
		Option_Some(x) => "number",
		Option_None => "empty",
	}
}
```

## Context

**Other Guard Syntax**: The codebase already supports `if` guards (Rust-style):
- `pattern_match_02_guards.dingo` - Uses `if` guards - ✅ PASSING
- `pattern_match_05_guards_basic.dingo` - Uses `if` guards - ✅ PASSING
- `pattern_match_07_guards_complex.dingo` - Uses `if` guards - ✅ PASSING

**Example of working `if` guards**:
```dingo
Result_Ok(x) if x > 0 => "positive"
```

**The `where` keyword is the Swift-style alternative** to `if` guards.

## Success Criteria

1. **Transpilation succeeds** for `pattern_match_06_guards_nested.dingo`
2. **Generated Go code compiles** without syntax errors
3. **Golden test passes** (either generate golden file or fix to match existing)
4. **Test status**: 103/103 tests passing (100%)
5. **No regressions**: All currently passing tests remain passing

## Additional Information

**Preprocessor Location**: `pkg/preprocessor/rust_match.go`

**Guard Parsing**: The `splitPatternAndGuard()` function at line 426 already handles both `if` and `where`:
```go
func (r *RustMatchProcessor) splitPatternAndGuard(patternAndGuard string) (pattern string, guard string) {
	// Guard syntax: Pattern if condition => expr
	//           or: Pattern where condition => expr
	if idx := strings.Index(patternAndGuard, " if "); idx != -1 {
		pattern = strings.TrimSpace(patternAndGuard[:idx])
		guard = strings.TrimSpace(patternAndGuard[idx+4:])
		return
	}
	if idx := strings.Index(patternAndGuard, " where "); idx != -1 {
		pattern = strings.TrimSpace(patternAndGuard[:idx])
		guard = strings.TrimSpace(patternAndGuard[idx+7:])
		return
	}
	return patternAndGuard, ""
}
```

So parsing seems to work. The bug is likely in **guard code generation**.

## Goal

Achieve **100% test passing** (103/103) by fixing the `where` guard preprocessor bug.

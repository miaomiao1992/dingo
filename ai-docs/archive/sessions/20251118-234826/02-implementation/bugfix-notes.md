# DINGO_MATCH_START Bug Investigation

## Problem Report
User reported that Variable Hoisting implementation removed `// DINGO_MATCH_START` comment, causing code corruption.

## Investigation Findings

### Initial Hypothesis (INCORRECT)
The bug report suggested the issue was in `generateSwitch()` where the `// DINGO_MATCH_START` comment was removed during Variable Hoisting implementation.

### Actual Root Cause (CORRECT)

The problem is **NOT** with the `generateSwitch()` function. The actual issue is:

1. **Pattern match uses if-based syntax (NOT switch-based)**
   - The test file `pattern_match_01_simple.dingo` has NO `dingo.toml` config
   - This means it uses the DEFAULT pattern match syntax
   - Default syntax is **if-based pattern matching** (NOT Rust switch syntax)
   - The `generateSwitch()` function is NOT even used for this test!

2. **DINGO comments appear in wrong locations**
   - Comments like `// DINGO_MATCH_START: result` appear at line 22 (inside Option_string_Some function)
   - Comments like `// DINGO_PATTERN: Ok(value)` appear at line 28 (inside Option_string_None function)
   - These are NOWHERE near the actual match expressions (which start at line 520)

3. **Source map/line number tracking issue**
   - The preprocessor generates DINGO comments at specific line numbers
   - But the AST transformation plugins inject HUNDREDS of lines of code BEFORE those lines
   - Result: Comments appear at lines 22, 28, 34 etc. (in Option type declarations)
   - But should appear at lines 522, 539, 554 etc. (in match expressions)

### Example of Corruption

**Expected** (line 522):
```go
func processResult(result Result[int, error]) int {
    // DINGO_MATCH_START: result
    __match_0 := result
    if result.IsOk() {
```

**Actual** (line 22):
```go
func Option_string_Some(arg0 string) Option_string {
    return Option_string{

    // DINGO_MATCH_START: result  ← WRONG LOCATION!
    tag: OptionTag_Some, some_0: &arg0}
}
```

### Why This Happens

The transpiler has a two-stage pipeline:

```
Stage 1: Preprocessor (Text-based)
  ↓ Generates DINGO comments at lines 8, 26, 44, etc.

Stage 2: AST Transformation (go/parser + plugins)
  ↓ Injects Result/Option type declarations (500+ lines)
  ↓ Now match expressions are at lines 520+
  ↓ But comments are still at original lines 8, 26, 44
  ↓ Result: Comments appear in injected type declarations!
```

## Fix Attempted

I restored the `// DINGO_MATCH_START` comment generation in `generateSwitch()`:

```go
// Line 2: DINGO_MATCH_START marker (MUST BE BEFORE temp var)
buf.WriteString(fmt.Sprintf("// DINGO_MATCH_START: %s\n", scrutinee))
```

**Result**: This fix is CORRECT for switch-based syntax, but does NOT fix this test because this test uses if-based syntax.

## Actual Problem To Fix

The real issue is in **if-based pattern matching code generation**. The if-based pattern matcher is NOT generating DINGO comments at all!

### Where If-Based Pattern Matching is Implemented

Need to check:
- `pkg/plugin/builtin/pattern_match.go` - PatternMatchPlugin
- This plugin transforms AST nodes for if-based pattern matching
- It should be generating DINGO comments but is NOT

### Correct Fix Required

1. Modify `PatternMatchPlugin` to generate DINGO comments when using if-based syntax
2. OR: Modify the comment injection to happen AFTER AST transformation (not during preprocessing)
3. OR: Make pattern_match_01_simple use switch-based syntax by adding dingo.toml

## Test Results

### Before Fix
```
FAIL: pattern_match_01_simple
Error: expected ';', found ':=' at line 62
```

### After Fix (generateSwitch restored)
```
FAIL: pattern_match_01_simple
Error: Code still corrupted (comments in wrong places)
Reason: Test doesn't use switch syntax, so fix doesn't apply
```

## Recommendation

**DO NOT fix generateSwitch() - that was not the bug!**

Instead:
1. Check if pattern_match_01_simple.go.golden expects switch or if syntax
2. If expects switch: Add dingo.toml to specify rust syntax
3. If expects if: Fix PatternMatchPlugin to generate DINGO comments
4. The bug is in if-based pattern matching, NOT switch-based!

## Files Involved

- `/Users/jack/mag/dingo/pkg/preprocessor/rust_match.go` - Switch-based (works correctly)
- `/Users/jack/mag/dingo/pkg/plugin/builtin/pattern_match.go` - If-based (MISSING DINGO comments)
- `/Users/jack/mag/dingo/tests/golden/pattern_match_01_simple.dingo` - Test file
- `/Users/jack/mag/dingo/tests/golden/pattern_match_01_simple.go.golden` - Expected output

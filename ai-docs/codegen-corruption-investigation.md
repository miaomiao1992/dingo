# Code Generation Corruption Investigation

## Issue Summary
The `pattern_match_01_simple` test generates severely corrupted Go code where:
- Type declarations appear before package statement
- Match markers (`// DINGO_MATCH_START`) are inserted in wrong locations
- Struct definitions are malformed (`type Option_string struct` followed by `{` on next line)
- Type declarations are intermingled with match comments

## Root Cause Analysis

### Key Observation
The corruption pattern shows that **RustMatchProcessor is inserting markers into the WRONG locations**, specifically:

1. **Line 2-4** (expected location): Type declarations should come first
2. **Line 7** (expected location): Comment from original Dingo code
3. **Line 10+** (expected location): Functions with match expressions

But actual output has:
1. **Line 1**: `type OptionTag uint8` (should be after imports)
2. **Line 2**: Original Dingo comment (should be preserved but in right place)
3. **Line 7**: `type Option_string struct` (incomplete - missing `{`)
4. **Line 8**: `{` on its own line (struct body without declaration)
5. **Line 9**: `// Example 1: Pattern match on Result[T,E]` (comment in wrong place)
6. **Line 10-12**: Struct fields (should be after `struct {`)

### Hypothesis
**RustMatchProcessor is running on PARTIAL CODE instead of complete match expressions.**

The processor is finding `match` keywords inside type-related code (possibly generated or intermediate representations) and trying to transform them as match expressions.

### Evidence
From corrupted output:
```go
type Option_string struct

// Example 1: Pattern match on Result[T,E]
{
	tag    OptionTag
	some_0 *string
}

func Option_string_Some(arg0 string) Option_string {
	return Option_string{

	// DINGO_MATCH_START: result
	tag: OptionTag_Some, some_0: &arg0}
}
```

The `// Example 1: Pattern match on Result[T,E]` comment is from line 7 of the original `.dingo` file, but it's appearing BEFORE a struct definition that's malformed.

### Pipeline Analysis
Order:
1. GenericSyntaxProcessor
2. TypeAnnotProcessor
3. ErrorPropProcessor
4. EnumProcessor (generates types)
5. RustMatchProcessor (generates matches) ← **PROBLEM HERE**
6. KeywordProcessor
7. Import injection

### What Should Happen
- RustMatchProcessor should scan for `match keyword {` pattern
- Find complete match expressions
- Transform them to switch statements with markers
- Preserve all other code as-is

### What's Actually Happening
- RustMatchProcessor is finding match expressions in WRONG contexts
- Or it's inserting markers in WRONG byte offsets
- Or earlier processors have already corrupted the structure

## Next Steps for Diagnosis

### Test 1: Isolate RustMatchProcessor
Create a test that runs ONLY RustMatchProcessor on the preprocessed output (after EnumProcessor but before other processors).

### Test 2: Check Byte Offsets
Add debug logging to RustMatchProcessor to show:
- Where match expressions are found (byte offset)
- What text is being matched
- Where markers are being inserted

### Test 3: Trace Preprocessor Output
Capture output after each preprocessor stage to see when corruption occurs:
1. After GenericSyntaxProcessor
2. After TypeAnnotProcessor
3. After ErrorPropProcessor
4. After EnumProcessor
5. After RustMatchProcessor ← TARGET
6. After KeywordProcessor

## Files to Review
- `pkg/preprocessor/rust_match.go` - Lines 147-428 (transformMatch, generateSwitch)
- `pkg/preprocessor/enum.go` - How types are generated
- `pkg/preprocessor/type_annot.go` - Type annotation processing

## Expected Fix Strategy
Once root cause is identified, likely fix options:

1. **If match detection is wrong**: Fix regex or context detection in RustMatchProcessor
2. **If byte offsets are wrong**: Fix position tracking in Process() method
3. **If struct generation is corrupted**: Fix EnumProcessor or ensure it completes before RustMatchProcessor
4. **If pipeline order is wrong**: Reorder processors to prevent interference

## Status
- [x] Analyzed corrupted output
- [x] Reviewed RustMatchProcessor code
- [ ] Identify exact corruption point
- [ ] Develop fix
- [ ] Verify with all 13 tests
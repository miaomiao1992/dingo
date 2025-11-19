# Bug Fix Test Results

## Root Cause Analysis (FINAL)

The issue is a **PIPELINE CONFLICT** between preprocessor and plugin:

### What Happens

```
Input: pattern_match_01_simple.dingo
  ↓
Stage 1: Preprocessor
  → RustMatchProcessor runs
  → Generates switch-based code with DINGO comments
  → Output has:
    // DINGO_MATCH_START: result
    switch __match_0.tag {
    case ResultTagOk:
        // DINGO_PATTERN: Ok(value)
  ↓
Stage 2: AST Transformation
  → PatternMatchPlugin runs
  → TRANSFORMS switch to if-based code
  → Output has:
    if result.IsOk() {
        value := *__match_0.ok_0
  → But DINGO comments are LEFT AT ORIGINAL LINE NUMBERS!
  ↓
Stage 3: AST Printer
  → Injects Result/Option type declarations (500+ lines)
  → DINGO comments now appear in wrong locations
  → Comments at lines 22, 28, 34 (in Option_string functions)
  → Should be at lines 520+ (in match expressions)
```

### Why Both Are Running

**Preprocessor** (pkg/preprocessor/preprocessor.go:80):
```go
processors = append(processors, NewRustMatchProcessor())
```

**Plugin** (pkg/generator/generator.go:75):
```go
patternMatchPlugin := builtin.NewPatternMatchPlugin()
pipeline.RegisterPlugin(patternMatchPlugin)
```

Both are enabled simultaneously!

### The Conflict

- **RustMatchProcessor**: Converts `match expr { ... }` → `switch expr.tag { ... }` (Rust-style)
- **PatternMatchPlugin**: Converts switch statements → if-based pattern matching

Result: RustMatchProcessor output is IMMEDIATELY transformed by PatternMatchPlugin, losing DINGO comments in the process!

### Expected Behavior

The golden file `pattern_match_01_simple.go.golden` expects **switch-based output**, which means:
- RustMatchProcessor should run ✅
- PatternMatchPlugin should NOT run ❌

### Fix Required

**Option 1**: Make PatternMatchPlugin skip code that already has DINGO_MATCH_START comments
**Option 2**: Disable PatternMatchPlugin when Rust syntax is configured
**Option 3**: Make PatternMatchPlugin preserve DINGO comments when transforming
**Option 4**: Make RustMatchProcessor run AFTER AST transformation (not before)

## Current Status

**Test Status**: FAILING
- Generated code uses if-based pattern matching (from PatternMatchPlugin)
- Expected code uses switch-based pattern matching (from RustMatchProcessor)
- DINGO comments appear in wrong locations (injected type declarations)

**Fix Applied**: Restored `// DINGO_MATCH_START` comment in generateSwitch()
**Result**: Partially correct - switch code would have comments, but switch code is being transformed away

## Recommendation

**Immediate fix**: Disable PatternMatchPlugin when config specifies Rust syntax

**Location**: pkg/generator/generator.go:75

**Change**:
```go
// Only register PatternMatchPlugin if NOT using Rust match syntax
// (Rust syntax is handled by RustMatchProcessor in preprocessor stage)
if cfg != nil && cfg.PatternMatch != nil && cfg.PatternMatch.Syntax != "rust" {
    patternMatchPlugin := builtin.NewPatternMatchPlugin()
    pipeline.RegisterPlugin(patternMatchPlugin)
}
```

This will prevent the plugin from transforming Rust-preprocessed switch statements.

## Files Analyzed

- `/Users/jack/mag/dingo/pkg/preprocessor/preprocessor.go` - Preprocessor pipeline (line 80)
- `/Users/jack/mag/dingo/pkg/preprocessor/rust_match.go` - RustMatchProcessor (generateSwitch)
- `/Users/jack/mag/dingo/pkg/generator/generator.go` - Plugin pipeline (line 75)
- `/Users/jack/mag/dingo/pkg/plugin/builtin/pattern_match.go` - PatternMatchPlugin
- `/Users/jack/mag/dingo/tests/golden/pattern_match_01_simple.dingo` - Test input
- `/Users/jack/mag/dingo/tests/golden/pattern_match_01_simple.go.golden` - Expected output (switch-based)
- `/Users/jack/mag/dingo/tests/golden/pattern_match_01_simple.go` - Actual output (if-based, corrupted)

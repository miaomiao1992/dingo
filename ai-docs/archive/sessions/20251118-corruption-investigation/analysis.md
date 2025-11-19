# Code Generation Corruption Analysis

## Root Cause Analysis

### Primary Issue: Preprocessor Position Shifting Bug

The corruption in `pattern_match_01_simple.dingo` stems from a fundamental line position calculation error in the preprocessor pipeline. When multiple transformations (rust_match, error_prop) are applied sequentially, each transformation shifts line positions through code insertion, but subsequent transformations fail to account for these shifts when calculating their own output positions.

### Detailed Mechanism

1. **Initial File Structure**: Input has clean type declarations at package level:
   ```
   type OptionTag uint8
   //... type declarations
   func processResult(result Result[int, error]) int {
       match result { ... }
   ```

2. **First Transformation (rust_match)**: Encounters `match result { ... }`
   - Generates transformed switch block with DINGO markers
   - This inserts 8+ lines of code (markers, variable assignments, switch statements)
   - Subsequent lines are shifted down by ~8 line numbers
   - **CRITICAL BUG**: Input source maps calculate mappings assuming NO position shifts

3. **Second Transformation (error_prop)**: Encounters `x?` expressions
   - Operates on shifted lines but uses ORIGINAL unshifted positions
   - Calculated mapping positions are now WRONG relative to final output
   - Triggers plugin reordering/injection of type declarations into wrong locations

4. **Plugin Phase Corruption**: PatternMatchPlugin.Process detects scattered DINGO markers
   - Type declarations get recreated in case blocks due to position misalignment
   - Results in: `"type Option_string struct { // DINGO_PATTERN: Ok(value)"`
   - Function fragments appear mid-type declarations

### Specific Code Locations

#### 1. rust_match.go Transformation (Trigger)
- `pkg/preprocessor/rust_match.go:72`: `transformMatch()` generates block replacing single line
- No accounting for position shift in subsquent transformations
- Returns mappings with incorrect `GeneratedLine` offsets

#### 2. No Position Shift Compensation
- `pkg/preprocessor/rust_match.go:76-82`: Appends transformations to global mappings
- SourceMap accumulates incorrect positions across pipeline stages
- **Missing**: Cumulative offset tracking across processor stages

#### 3. Secondary Corruptions from Scheduled Fixes
- Recent `isInAssignmentContext()` changes (line 195-206) incorrectly classify assignment contexts
- `collectMatchExpression()` line replacement bugs (C7 newline handling)

### Expected Behavior
- Type declarations should remain at package level
- DINGO markers should be adjacent to generated switch statements
- Function bodies should be contiguous and properly nested

### Actual Behavior
- Type declarations fragmented across case blocks
- DINGO markers interspersed with unrelated code
- Compilation errors from syntactically invalid interleaving

## Contributing Factors

### Preprocessor Pipeline Order Confusion
- Pattern matching transformation occurs after error propagation
- But MARKER positions assume pre-transformed source
- Plugins receive corrupted positions leading to misplacement

### Mapping Accumulation Logic Defect
- Each processor appends `[]Mapping` without offset adjustment
- When transformations insert code, all subsequent mappings become invalid
- No mechanism to "slide" mappings after each transformation stage

## Verification

Running `go test ./tests -run pattern_match_01_simple` shows:
- Line 62: `expected ';', found ':='` (fragmented function body)
- Compass indicator that case content is malformed

Examining `pattern_match_01_simple.go.actual` reveals:
- `type Option_string struct {` appears inside case block
- Comments and returns intermingle with type definitions
- Function signatures get interrupted by match artifacts
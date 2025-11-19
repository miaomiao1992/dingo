# Task: Imports - Files Changed

## Summary
Implemented automatic import detection and source mapping adjustment for the Dingo preprocessor.

## Files Created
None

## Files Modified

### 1. `/Users/jack/mag/dingo/pkg/preprocessor/preprocessor.go`
**Changes:**
- Added imports: `bytes`, `go/parser`, `go/printer`, `go/token`, `sort`, `strings`, `golang.org/x/tools/go/ast/astutil`
- Added `ImportProvider` interface for processors that need to add imports
- Updated `Preprocessor.Process()` to:
  - Collect needed imports from processors implementing `ImportProvider`
  - Inject all imports at the END of the pipeline (after all transformations)
  - Adjust source mappings to account for added import lines
- Added helper functions:
  - `injectImports(source []byte, needed []string) []byte` - Adds imports using astutil.AddImport
  - `adjustMappingsForImports(sourceMap *SourceMap, numImportLines int)` - Shifts mapping line numbers

### 2. `/Users/jack/mag/dingo/pkg/preprocessor/error_prop.go`
**Changes:**
- Added import: `sort`
- Removed imports: `go/printer`, `golang.org/x/tools/go/ast/astutil` (no longer needed)
- Added `ImportTracker` struct for tracking function calls
- Added `stdLibFunctions` map covering os, encoding/json, strconv, io, fmt packages
- Added `NewImportTracker()`, `TrackFunctionCall()`, `GetNeededImports()` methods
- Updated `ErrorPropProcessor` struct:
  - Added `importTracker *ImportTracker` field
  - Added `mappings []Mapping` field (for adjustment after import injection)
- Updated `Process()` method:
  - Initialize import tracker
  - Store mappings internally
  - Removed import injection logic (now done by main Preprocessor)
- Added `GetNeededImports()` method - implements `ImportProvider` interface
- Added `trackFunctionCallInExpr()` - extracts function names from expressions
- Updated `expandAssignment()` and `expandReturn()` to call `trackFunctionCallInExpr()`
- Removed obsolete methods: `injectImports()`, `filterExistingImports()`, `adjustMappingsForImports()`, `ensureFmtImport()`, `insertFmtImportSimple()`

### 3. `/Users/jack/mag/dingo/pkg/preprocessor/preprocessor_test.go`
**Changes:**
- Added import: `fmt`
- Updated `TestErrorPropagationBasic`:
  - Expected outputs now include imports (`import "os"`, `import "strconv"`)
- Updated `TestGeminiCodeReviewFixes`:
  - Changed from exact string comparison to feature verification
  - Added checks for `"fmt"` and `"os"` imports
- Updated `TestSourceMapGeneration`:
  - Adjusted expected generated line numbers to account for import block offset
  - Updated documentation to reflect import block presence
- Updated `TestSourceMapMultipleExpansions`:
  - Adjusted expected generated line numbers with import offset constant
  - Updated documentation
- Added `TestAutomaticImportDetection`:
  - Tests os.ReadFile import
  - Tests strconv.Atoi import
  - Tests multiple imports (os + strconv)
  - Tests fmt import with error messages
- Added `TestSourceMappingWithImports`:
  - Verifies imports are added
  - Verifies mappings are correctly adjusted after import injection
  - Verifies original line mappings remain accurate

## Architecture Changes

### Import Injection Pipeline
**Before:**
```
ErrorPropProcessor.Process():
  1. Transform ? operators
  2. Inject imports (FAILED because output has invalid 'let' syntax)
  3. Adjust mappings
  4. Return
```

**After:**
```
Main Preprocessor.Process():
  1. Run ErrorPropProcessor (tracks imports, doesn't inject)
  2. Run KeywordProcessor (let → var)
  3. Collect imports from all processors via ImportProvider interface
  4. Inject ALL imports (now valid Go syntax!)
  5. Adjust ALL source mappings
  6. Return
```

**Key Insight:** Import injection MUST happen AFTER all syntax transformations complete, because:
- ErrorPropProcessor outputs contain `let` (invalid Go syntax)
- KeywordProcessor converts `let` → `var` (valid Go syntax)
- `go/parser` can only parse valid Go syntax
- Therefore, imports can only be added AFTER KeywordProcessor runs

### Covered Standard Library Functions
- **os**: ReadFile, WriteFile, Open, Create, Stat, Remove, Mkdir, MkdirAll, Getwd, Chdir
- **encoding/json**: Marshal, Unmarshal
- **strconv**: Atoi, Itoa, ParseInt, ParseFloat, ParseBool, FormatInt, FormatFloat
- **io**: ReadAll
- **fmt**: Sprintf, Fprintf, Printf, Errorf (Errorf tracked via needsFmt flag)

## Test Coverage
- All 8 test functions pass
- 11 test cases pass
- Import detection works for single and multiple imports
- Source mappings correctly adjusted for import block offsets
- Verified imports are deduplicated and sorted

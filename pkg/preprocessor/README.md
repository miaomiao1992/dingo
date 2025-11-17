# Dingo Preprocessor Architecture

## Overview

The preprocessor transforms Dingo source code to valid Go code through a pipeline of processors. Each processor handles a specific language feature (error propagation, pattern matching, etc.).

## Processing Pipeline

### Stage 1: Feature Processors (Ordered)

Processors run in sequence, each receiving the output of the previous processor:

1. **Error Propagation** (`error_prop.go`)
   - Expands `?` operator to error checking code
   - Tracks function calls for automatic import detection
   - Generates source mappings for error propagation sites

2. **Pattern Matching** (future: `pattern_match.go`)
   - Expands `match` expressions to switch statements

3. **Result/Option Types** (future: `result_option.go`)
   - Transforms Result<T,E> and Option<T> to Go structs

### Stage 2: Import Injection (FINAL STEP)

After ALL processors complete:

1. Collect all needed imports from `ImportTracker`
2. Parse transformed Go source to AST
3. Inject imports using `astutil.AddImport`
4. Adjust source map offsets for injected lines
5. Format and return final Go source

**CRITICAL POLICY**: Import injection is ALWAYS the final step. No processor may run after imports are injected.

## Source Mapping Rules

### Mapping Creation

Each processor creates mappings as it transforms code:

```go
mapping := Mapping{
	OriginalLine:    originalLineInDingoSource,
	OriginalColumn:  originalColumnInDingoSource,
	GeneratedLine:   currentLineInTransformedGoCode,
	GeneratedColumn: currentColumnInTransformedGoCode,
	Length:          lengthOfTransformedToken,
	Name:            "feature_name",  // e.g., "error_prop"
}
```

### Offset Adjustment

When imports are injected, mappings are adjusted:

1. Calculate import insertion line (after package declaration)
2. Count number of import lines added
3. Shift ALL mappings with `GeneratedLine > importInsertionLine` by the number of added lines
4. Mappings AT or BEFORE the insertion line remain unchanged

**Example**:
```
BEFORE import injection:
  Line 1: package main
  Line 2:
  Line 3: func foo() { ... }  ← mapping: Generated=3

AFTER injecting 2 imports at line 2:
  Line 1: package main
  Line 2:
  Line 3: import "os"
  Line 4:
  Line 5: func foo() { ... }  ← mapping adjusted: Generated=5

Adjustment logic:
  - importInsertionLine = 2
  - numImportLines = 2
  - Original mapping: GeneratedLine=3
  - 3 > 2 → shift by 2 → new GeneratedLine=5
```

### Critical Fix (CRITICAL-2)

The offset adjustment uses `>` (not `>=`) to exclude the insertion line itself:

```go
if sourceMap.Mappings[i].GeneratedLine > importInsertionLine {
	sourceMap.Mappings[i].GeneratedLine += numImportLines
}
```

This prevents shifting mappings that are exactly at the insertion line.

## Why Preprocessor vs Transformer?

Error propagation and simple syntax transformations belong in the preprocessor because:

### Advantages of Text-Based Processing

1. **Simplicity**: Regex and line-based transformations are easier to understand and maintain
2. **Performance**: Text processing is faster than AST parsing/printing cycles
3. **Source Mapping**: Line-level transforms make it trivial to maintain accurate position mappings
4. **Independence**: Doesn't require type information or AST structure
5. **Proven**: 693 lines of battle-tested, production-ready code

### When to Use Transformer Instead

Complex features requiring semantic analysis belong in the AST transformer:
- Lambda function transformations (need type inference)
- Pattern matching (need exhaustiveness checking)
- Safe navigation with method chaining (need type context)

## Architecture

### Pipeline Position

```
.dingo file
    ↓
[Preprocessor] → Go source text + source maps
    ↓           (? operator expanded, imports added)
[Parser] → AST
    ↓
[Transformer] → Modified AST
    ↓           (lambdas, pattern matching, safe nav)
[Generator] → Final .go file
```

### Processing Flow

1. **Sequential Feature Processing**
   - Type annotations (must be first)
   - Error propagation
   - Keywords (after error prop to avoid interference)
   - Future: Lambdas, sum types, pattern matching, operators

2. **Import Collection**
   - Each processor implementing `ImportProvider` reports needed imports
   - Deduplication across all processors
   - Automatic filtering of already-present imports

3. **Import Injection** (FINAL STEP)
   - After all transformations complete
   - Uses `go/parser` + `astutil.AddImport` for correctness
   - Generates properly formatted import block

4. **Source Map Adjustment**
   - Calculate line offset from added imports
   - Adjust all mapping positions to maintain accuracy

## Implementation Details

### Error Propagation Example

**Dingo Input:**
```go
func readConfig() error {
    let data = ReadFile("config.json")?
    return nil
}
```

**Preprocessor Output:**
```go
import "os"

func readConfig() error {
    data, __err0 := ReadFile("config.json")
    if __err0 != nil {
        return __err0
    }
    return nil
}
```

**Source Map:**
```
Dingo line 2 → Go line 4 (accounting for import)
Dingo line 3 → Go line 8
```

### Import Detection Example

**Detected Functions:**
- `ReadFile` → `"os"`
- `Marshal` → `"encoding/json"`
- `Atoi` → `"strconv"`

**Auto-Injected Import Block:**
```go
import (
    "encoding/json"
    "os"
    "strconv"
)
```

## Key Files

- `preprocessor.go` - Main orchestrator, import injection
- `error_prop.go` - Error propagation (`?`) transformation (693 lines)
- `type_annot.go` - Type annotation syntax normalization
- `keyword.go` - Keyword transformations (`let` → `var`)
- `sourcemap.go` - Position mapping infrastructure

## Testing

Run preprocessor tests:
```bash
go test ./pkg/preprocessor/... -v
```

Test import detection:
```bash
go test ./pkg/preprocessor/... -run TestImport -v
```

## Future Enhancements

- [ ] Smart import grouping (stdlib, third-party, local)
- [ ] Import alias detection and preservation
- [ ] Configurable import style (grouped vs single-line)
- [ ] Performance optimization for large files
- [ ] Incremental processing for IDE integration

## Contributing

When adding new preprocessor features:

1. Implement `FeatureProcessor` interface
2. Add to processor list in `New()` (order matters!)
3. Implement `ImportProvider` if imports are needed
4. Write comprehensive unit tests
5. Update source map generation if line counts change
6. Document the feature in this README
7. **NEVER** run processors after import injection

## Related Packages

- `pkg/transform` - AST-level transformations (lambdas, pattern matching)
- `pkg/parser` - Go parser wrapper
- `pkg/generator` - Final code generation and formatting

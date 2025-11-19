# Fix Implementation Plan for Preprocessor Corruption Issue

## Summary

The code generation corruption stems from preprocessor position shifting where transformations insert code, changing line numbers for subsequent processors but without adjusting positions accordingly.

## Root Cause

Each preprocessor in the pipeline (rust_match, error_prop, etc.) calculates source mappings based on original input positions. When earlier processors insert/delete code, the line numbers shift, but later processors still use unadjusted positions, leading to incorrect insertion points.

## Solution Architecture

Add cumulative offset tracking to the Preprocessor struct to maintain correct position calculations across the transformation pipeline.

## Code Changes Required

### 1. Preprocessor struct (pkg/preprocessor/preprocessor.go)

Add offset tracking field:
```go
type Preprocessor struct {
    source     []byte
    processors []FeatureProcessor
    oldConfig  *Config
    config     *config.Config
    lineOffset int  // Track cumulative line shifts from transformations
}
```

### 2. Process method update (pkg/preprocessor/preprocessor.go ~line 96)

Modify the Process method to maintain offset:

Code addition:
```go
func (p *Preprocessor) Process() (string, *SourceMap, error) {
    result := p.source
    sourceMap := NewSourceMap()
    neededImports := []string{}
    lineOffset := 0  // Initialize offset tracker

    for _, processor := range p.processors {
        transformed, mappings, err := processor.Process(result)
        if err != nil {
            return "", nil, err
        }

        // Adjust mappings for current offset to fix position shifting
        adjustedMappings := p.adjustMappingsForOffset(mappings, lineOffset)
        sourceMap.AddMappings(adjustedMappings)

        // Update offset: newlines added - newlines removed
        originalLines := strings.Count(string(result), "\n")
        newLines := strings.Count(transformed, "\n")
        lineOffset += newLines - originalLines

        result = []byte(transformed)

        if provider, ok := processor.(ImportProvider); ok {
            neededImports = append(neededImports, provider.GetNeededImports()...)
        }
    }

    // ... rest of method
}
```

### 3. Add helper method (pkg/preprocessor/preprocessor.go)

Add adjustment function:
```go
func (p *Preprocessor) adjustMappingsForOffset(mappings []Mapping, offset int) []Mapping {
    adjusted := make([]Mapping, len(mappings))
    for i, m := range mappings {
        adjusted[i] = m
        adjusted[i].GeneratedLine += offset
    }
    return adjusted
}
```

## Testing Plan

1. Run `go test ./tests -run TestGoldenFiles/pattern_match_01_simple` after implementation
2. Verify compilation succeeds (no "expected ';', found ':='" errors)
3. Check that type declarations appear at package level, not in case blocks
4. Ensure other pattern match tests pass (full test suite unaffected)
5. Validate source map correctness for LSP integration

## Impact Assessment

- **Positive**: Fixes all position shifting issues in preprocessor pipeline
- **Risk**: Low - adds offset tracking without changing transformation logic
- **Scope**: Affects entire preprocessor pipeline (all feature processors)
- **Compatibility**: Maintains existing API and behavior

## Files to Modify

1. `pkg/preprocessor/preprocessor.go` - Add offset field and logic

## Alternative Considerations

- Could modify each processor to handle shifting independently, but centralized approach is cleaner
- Considered making SourceMap handle offsets, but logic belongs in Preprocessor transformation orchestration
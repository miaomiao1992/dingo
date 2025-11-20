# Dingo Source Map Specification

**Version**: 1
**Status**: Stable
**Last Updated**: 2025-11-19

---

## Overview

Dingo source maps provide bidirectional position mapping between original `.dingo` source files and generated `.go` files. This enables:

- **LSP Integration**: Accurate diagnostics, autocomplete, and go-to-definition
- **Debugging**: Breakpoint mapping and stack trace translation
- **Error Reporting**: Precise error messages pointing to original source

---

## Format Specification

### JSON Schema

```json
{
  "version": 1,
  "dingo_file": "/path/to/source.dingo",
  "go_file": "/path/to/generated.go",
  "mappings": [
    {
      "original_line": 10,
      "original_column": 5,
      "generated_line": 15,
      "generated_column": 10,
      "length": 20,
      "name": "error_prop"
    }
  ]
}
```

### Fields

#### Top-Level

- **`version`** (int, required): Source map format version. Current: `1`
- **`dingo_file`** (string, optional): Path to original `.dingo` file
- **`go_file`** (string, optional): Path to generated `.go` file
- **`mappings`** (array, required): Array of position mappings

#### Mapping Entry

- **`original_line`** (int, required): Line number in `.dingo` file (1-indexed)
- **`original_column`** (int, required): Column number in `.dingo` file (0-indexed)
- **`generated_line`** (int, required): Line number in `.go` file (1-indexed)
- **`generated_column`** (int, required): Column number in `.go` file (0-indexed)
- **`length`** (int, required): Length of the mapped segment in characters
- **`name`** (string, optional): Descriptive name for debugging (e.g., `"error_prop"`, `"type_annotation"`)

---

## Position Mapping Algorithm

### Forward Mapping (Original → Generated)

Given a position in `.dingo` source, find the corresponding position in `.go` file:

```go
func MapToGenerated(origLine, origCol int) (genLine, genCol int) {
    for each mapping in mappings {
        if mapping.OriginalLine == origLine &&
           origCol >= mapping.OriginalColumn &&
           origCol < mapping.OriginalColumn + mapping.Length {
            offset := origCol - mapping.OriginalColumn
            return mapping.GeneratedLine, mapping.GeneratedColumn + offset
        }
    }
    // No mapping found, return identity
    return origLine, origCol
}
```

**Complexity**: O(n) where n = number of mappings
**Performance**: ~29ns for 100 mappings (< 0.00003ms)

### Reverse Mapping (Generated → Original)

Given a position in `.go` file, find the corresponding position in `.dingo` source:

```go
func MapToOriginal(genLine, genCol int) (origLine, origCol int) {
    // 1. Look for exact match within mapping range
    for each mapping where mapping.GeneratedLine == genLine {
        if genCol >= mapping.GeneratedColumn &&
           genCol < mapping.GeneratedColumn + mapping.Length {
            offset := genCol - mapping.GeneratedColumn
            return mapping.OriginalLine, mapping.OriginalColumn + offset
        }
    }

    // 2. Find closest mapping on same line (heuristic)
    bestMatch := nil
    for each mapping where mapping.GeneratedLine == genLine {
        distance := abs(mapping.GeneratedColumn - genCol)
        if bestMatch == nil || distance < minDistance {
            bestMatch = mapping
        }
    }

    // 3. Apply offset if reasonable
    if bestMatch != nil {
        offset := genCol - bestMatch.GeneratedColumn
        if offset >= 0 && offset < bestMatch.Length + 10 {
            return bestMatch.OriginalLine, bestMatch.OriginalColumn + offset
        }
        return bestMatch.OriginalLine, bestMatch.OriginalColumn
    }

    // No mapping found, return identity
    return genLine, genCol
}
```

**Complexity**: O(n) where n = number of mappings
**Performance**: ~97ns for 100 mappings (< 0.0001ms)

---

## Mapping Examples

### Example 1: Error Propagation

**Dingo Source** (`example.dingo:10`):
```dingo
result := ReadFile(path)?
```

**Generated Go** (`example.go:15-20`):
```go
_tmp0, _err0 := ReadFile(path)
if _err0 != nil {
    return Result{}, _err0
}
result := _tmp0
```

**Source Map**:
```json
{
  "mappings": [
    {
      "name": "expr_mapping",
      "original_line": 10,
      "original_column": 10,
      "generated_line": 15,
      "generated_column": 15,
      "length": 14
    },
    {
      "name": "error_prop",
      "original_line": 10,
      "original_column": 24,
      "generated_line": 16,
      "generated_column": 1,
      "length": 1
    }
  ]
}
```

**Mapping Behavior**:
- Error at gen `15:18` (inside `ReadFile`) → maps to orig `10:13` (inside expression)
- Error at gen `16:5` (error check) → maps to orig `10:24` (position of `?`)

### Example 2: Pattern Matching

**Dingo Source** (`example.dingo:20`):
```dingo
match result {
    Ok(val) => val
    Err(e) => 0
}
```

**Generated Go** (`example.go:50-60`):
```go
scrutinee := result
switch scrutinee.tag {
case ResultTagOk:
    val := *scrutinee.ok
    val
case ResultTagErr:
    e := scrutinee.err
    0
}
```

**Source Map**:
```json
{
  "mappings": [
    {
      "name": "pattern_match",
      "original_line": 20,
      "original_column": 1,
      "generated_line": 50,
      "generated_column": 1,
      "length": 13
    },
    {
      "name": "ok_pattern",
      "original_line": 21,
      "original_column": 5,
      "generated_line": 53,
      "generated_column": 1,
      "length": 10
    },
    {
      "name": "err_pattern",
      "original_line": 22,
      "original_column": 5,
      "generated_line": 56,
      "generated_column": 1,
      "length": 11
    }
  ]
}
```

### Example 3: Multi-Line Expansion

One Dingo line can expand to multiple Go lines:

**Dingo** (`example.dingo:5`):
```dingo
data := fetchData()? + processData()?
```

**Generated Go** (`example.go:10-25`):
```go
_tmp0, _err0 := fetchData()
if _err0 != nil {
    return _err0
}
_tmp1, _err1 := processData()
if _err1 != nil {
    return _err1
}
data := _tmp0 + _tmp1
```

**Multiple mappings point to same original line**:
```json
{
  "mappings": [
    {
      "name": "fetch_expr",
      "original_line": 5,
      "original_column": 8,
      "generated_line": 10,
      "generated_column": 15,
      "length": 11
    },
    {
      "name": "fetch_error_prop",
      "original_line": 5,
      "original_column": 19,
      "generated_line": 11,
      "generated_column": 1,
      "length": 1
    },
    {
      "name": "process_expr",
      "original_line": 5,
      "original_column": 23,
      "generated_line": 14,
      "generated_column": 15,
      "length": 13
    },
    {
      "name": "process_error_prop",
      "original_line": 5,
      "original_column": 36,
      "generated_line": 15,
      "generated_column": 1,
      "length": 1
    }
  ]
}
```

---

## Common Mapping Names

### Preprocessor-Generated

| Name | Description | Original Syntax | Generated Pattern |
|------|-------------|----------------|-------------------|
| `error_prop` | Error propagation operator | `expr?` | `if err != nil { return }` |
| `expr_mapping` | Expression within error propagation | `fetchData()` | `_tmp, _err := fetchData()` |
| `type_annotation` | Type annotation syntax | `x: Type` | `var x Type` |
| `pattern_match` | Match expression | `match x {}` | `switch scrutinee.tag {}` |
| `ok_pattern` | Ok pattern arm | `Ok(val) =>` | `case ResultTagOk:` |
| `err_pattern` | Err pattern arm | `Err(e) =>` | `case ResultTagErr:` |
| `some_pattern` | Some pattern arm | `Some(val) =>` | `case OptionTagSome:` |
| `none_pattern` | None pattern arm | `None =>` | `case OptionTagNone:` |

### AST-Generated

| Name | Description | Example |
|------|-------------|---------|
| `result_constructor` | Result type construction | `Ok(value)`, `Err(error)` |
| `option_constructor` | Option type construction | `Some(value)`, `None()` |
| `helper_method` | Result/Option helper call | `.Unwrap()`, `.IsOk()` |

---

## Best Practices

### For Preprocessors

1. **Create fine-grained mappings**: Map individual expressions, not entire statements
2. **Use descriptive names**: Helps debugging and LSP diagnostics
3. **Map original syntax precisely**: Point to the exact character that caused the expansion
4. **Avoid overlapping ranges**: Multiple mappings for same position can confuse LSP

### For LSP Integration

1. **Cache source maps**: Load once, reuse for multiple requests (~29-97ns lookup)
2. **Handle missing mappings gracefully**: Return identity mapping if no match found
3. **Prefer exact matches**: Use heuristics only as fallback
4. **Invalidate cache on file changes**: Re-transpile triggers new source map

### For Debugging

1. **Use `MapToOriginalWithDebug`**: Detailed logging helps troubleshoot mapping issues
2. **Validate round-trip**: Original → Generated → Original should match
3. **Check mapping coverage**: Ensure all expanded code has mappings

---

## Performance Characteristics

### Benchmarks

From `sourcemap_validation_test.go`:

```
BenchmarkMapToOriginal-10     13418503    96.77 ns/op    0 B/op    0 allocs/op
BenchmarkMapToGenerated-10    58407186    28.55 ns/op    0 B/op    0 allocs/op
```

**Results**:
- **MapToOriginal**: ~97ns (< 0.0001ms) per lookup
- **MapToGenerated**: ~29ns (< 0.00003ms) per lookup
- **Zero heap allocations**: All operations stack-only
- **Test set**: 100 mappings (realistic for typical files)

### LSP Latency Budget

**Target**: <100ms for autocomplete (VSCode → VSCode)

**Breakdown**:
- VSCode → dingo-lsp IPC: ~5ms
- **Position translation**: <0.0001ms ✅ (97ns)
- dingo-lsp → gopls IPC: ~5ms
- gopls type checking: ~50ms
- gopls → dingo-lsp IPC: ~5ms
- **Position translation**: <0.0001ms ✅
- dingo-lsp → VSCode IPC: ~5ms
- **Total**: ~70ms ✅

**Conclusion**: Source map lookups contribute <0.0002ms to total latency (negligible).

---

## Validation Tests

### Round-Trip Tests

Validate bidirectional mapping accuracy:

```go
// Forward: Original → Generated
genLine, genCol := sm.MapToGenerated(10, 15)
// Reverse: Generated → Original
origLine, origCol := sm.MapToOriginal(genLine, genCol)
// Assert: origLine == 10 && origCol == 15
```

### JSON Serialization

Validate source maps survive round-trip through JSON:

```go
jsonData, _ := sm.ToJSON()
restored, _ := FromJSON(jsonData)
// Assert: restored == sm
```

### Version Compatibility

Source maps support forward compatibility:

```json
{
  "version": 2,
  "future_field": "ignored",
  "mappings": []
}
```

Parser accepts future versions, ignores unknown fields.

---

## Future Enhancements

### Potential Version 2 Features

- **Column mapping optimization**: Binary search for large files (O(log n))
- **Segment-based mappings**: Group contiguous mappings for better compression
- **Source content embedding**: Include original source for offline debugging
- **Multi-file mappings**: Support for file includes/imports

### Debugging Improvements

- **Visual mapping inspector**: Tool to visualize mappings
- **Mapping coverage reports**: Identify unmapped code regions
- **Accuracy metrics**: Measure mapping precision

---

## References

- **Implementation**: `pkg/preprocessor/sourcemap.go`
- **Tests**: `pkg/preprocessor/sourcemap_test.go`
- **Validation**: `pkg/preprocessor/sourcemap_validation_test.go`
- **LSP Integration**: `pkg/lsp/translator.go`
- **Cache**: `pkg/lsp/sourcemap_cache.go`

---

## Changelog

- **2025-11-19**: Added comprehensive validation suite
- **2025-11-18**: Initial implementation
- **Version**: 1 (stable)

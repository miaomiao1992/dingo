# Golden Test Failures Investigation

## Context: Dingo Project

Dingo is a meta-language for Go (like TypeScript for JavaScript) that transpiles `.dingo` files to idiomatic `.go` files. It provides:
- Result/Option types
- Pattern matching
- Error propagation with `?` operator
- Sum types/enums

**Architecture**: Two-stage transpilation:
1. **Preprocessor** (text/regex-based): Transforms Dingo syntax → valid Go
2. **AST Processing**: Uses `go/parser` → plugin pipeline → generates `.go` + `.sourcemap`

## Problem

Our golden test suite is failing. Golden tests verify that transpiled output matches expected `.go.golden` files.

**Test Status**: Most of 46 golden tests are failing

## Failure Patterns

### Pattern 1: Import Formatting Mismatch (MINOR)
**Expected in golden files**:
```go
import "os"
```

**Actual transpiler output**:
```go
import (
    "os"
)
```

### Pattern 2: Missing `// dingo:s:N` Start Comments (CRITICAL)
**Expected in golden files**:
```go
__tmp0, __err0 := os.ReadFile(path)
// dingo:s:1
if __err0 != nil {
    return nil, __err0
}
// dingo:e:1
```

**Actual transpiler output**:
```go
__tmp0, __err0 := os.ReadFile(path)

if __err0 != nil {
    return nil, __err0
}
// dingo:e:1
```

The transpiler is NOT generating `// dingo:s:1` (start) comments, only `// dingo:e:1` (end).

### Pattern 3: Duplicate `// dingo:s:N` in Golden Files (CRITICAL)
Some golden files have DUPLICATE start comments:
```go
__tmp0, __err0 := ReadFile(path)
// dingo:s:1
// dingo:s:1  <-- DUPLICATE!
if __err0 != nil {
```

This appears to be a golden file bug.

### Pattern 4: Unqualified Functions in Golden Files (CRITICAL)
**Expected in golden files**:
```go
__tmp0, __err0 := ReadFile(path)  // WRONG - won't compile
__tmp1, __err1 := Atoi(string(data))  // WRONG
```

**Actual transpiler output**:
```go
__tmp0, __err0 := os.ReadFile(path)  // CORRECT - compiles
__tmp1, __err1 := strconv.Atoi(string(data))  // CORRECT
```

The transpiler correctly qualifies function calls with package names. Golden files don't.

## Your Investigation Task

Please analyze this situation and provide your expert opinion on:

### 1. Root Cause Analysis
- Why is the transpiler not generating `// dingo:s:N` comments?
- Where in the codebase would this logic live? (Hint: likely `pkg/generator/preprocessor/error_prop.go`)
- Are the duplicate `// dingo:s:N` comments in golden files a bug?

### 2. Source of Truth
Which is correct?

**Option A: Golden files are correct**
- Fix transpiler to match golden files
- Add `// dingo:s:N` comment generation
- But golden files have uncompilable Go (unqualified calls like `ReadFile()`)

**Option B: Transpiler output is correct**
- Regenerate all golden files from current transpiler
- Golden files should contain compilable Go code
- Remove duplicate comments

### 3. Source Map Comments Purpose
What is the purpose of `// dingo:s:N` and `// dingo:e:N` comments?
- Are they critical for source map functionality?
- Should they appear in final transpiled output?
- Are they meant for internal tracking only?

### 4. Import Formatting
Is import formatting (single-line vs multi-line) important, or just cosmetic?
- Should we enforce single-line for single imports?
- Is this a `go/printer` configuration issue?

### 5. Recommended Fix Strategy

Provide a prioritized implementation plan:
1. What to fix first
2. What can be deferred
3. What is cosmetic vs critical
4. Specific files/functions to modify

## Expected Response Format

Please structure your response as:

```markdown
# Investigation Results

## Executive Summary
[2-3 sentences: What's wrong and recommended fix]

## Root Cause
[Detailed analysis of why transpiler behavior differs from golden files]

## Source of Truth Decision
[Which is correct: golden files or transpiler? Why?]

## Source Map Comments Analysis
[Purpose and correct behavior of // dingo:s:N comments]

## Implementation Plan
1. [Critical fix 1] - File: `path/to/file.go`, Function: `FunctionName()`
2. [Critical fix 2] - ...
3. [Minor fix 1] - ...

## Code Locations
- Error propagation logic: `pkg/generator/preprocessor/error_prop.go`
- Source map generation: `pkg/...`
- [Other relevant files]

## Confidence Level
[HIGH/MEDIUM/LOW] - [Why]
```

## Additional Context

**Error Propagation Processor** likely contains the logic for generating these comments and transforming `?` operator into error handling code.

**Golden Test Files** are in `tests/golden/` with naming pattern:
- `{feature}_{NN}_{description}.dingo` (source)
- `{feature}_{NN}_{description}.go.golden` (expected output)

**Current Behavior**: Transpiler generates compilable Go code with correct imports and qualified function calls, but missing start comments.

Please provide your expert analysis!

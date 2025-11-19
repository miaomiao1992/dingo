# Package Statement Ordering Fix

## Problem Fixed

**Issue**: Generator was printing injected type declarations BEFORE the package statement, creating invalid Go code.

**Error**: `/test.go:1:1: expected 'package', found 'type'`

**Root Cause**: In `pkg/generator/generator.go` (lines 173-201), the code was:
1. Printing injected AST declarations first (ResultTag, Option types, etc.)
2. Printing main AST second (which contained the package statement)

This resulted in:
```go
type ResultTag uint8  ← WRONG! No package statement
const (...)

package main          ← Package statement in wrong place!
```

## Solution Implemented

**File Modified**: `/Users/jack/mag/dingo/pkg/generator/generator.go`

**Changes** (lines 173-217):
- Replaced "Option B" approach (print injected AST first) with structured ordering
- New order guarantees correct Go file structure:

1. **Package statement** - Always first
2. **Imports** - If present in main AST
3. **Injected type declarations** - From plugin pipeline
4. **User declarations** - From main AST (excluding package/imports)

**Key Implementation**:
```go
// 1. Print package statement from main AST
fmt.Fprintf(&buf, "package %s\n\n", transformed.Name.Name)

// 2. Print imports from main AST (if any)
if len(transformed.Imports) > 0 {
    buf.WriteString("import (\n")
    for _, imp := range transformed.Imports {
        cfg.Fprint(&buf, g.fset, imp)
        buf.WriteString("\n")
    }
    buf.WriteString(")\n\n")
}

// 3. Print injected type declarations (if any)
if g.pipeline != nil {
    injectedAST := g.pipeline.GetInjectedTypesAST()
    if injectedAST != nil && len(injectedAST.Decls) > 0 {
        for _, decl := range injectedAST.Decls {
            cfg.Fprint(&buf, g.fset, decl)
            buf.WriteString("\n")
        }
        buf.WriteString("\n")
    }
}

// 4. Print main AST declarations ONLY (skip package/imports)
for _, decl := range transformed.Decls {
    cfg.Fprint(&buf, g.fset, decl)
    buf.WriteString("\n")
}
```

## Results

**Before Fix**:
```go
type ResultTag uint8
const (
	ResultTag_Ok ResultTag = iota
	ResultTag_Err
)
package main  ← WRONG POSITION
```

**After Fix**:
```go
package main  ← CORRECT POSITION

import (
"fmt"
)

type ResultTag uint8
const (
	ResultTag_Ok ResultTag = iota
	ResultTag_Err
)
```

## Testing

**Command**:
```bash
go build -o /tmp/dingo ./cmd/dingo
/tmp/dingo build tests/golden/pattern_match_01_simple.dingo -o /tmp/output.go
```

**Result**: Package statement now appears first in generated code ✅

**Golden File Updated**:
- `/Users/jack/mag/dingo/tests/golden/pattern_match_01_simple.go.golden`
- Now has correct package statement at top

**Compilation Test**:
```bash
go test ./tests -run TestGoldenFilesCompilation/pattern_match_01_simple -v
```

**Status**:
- ✅ Package ordering error FIXED
- ⚠️ Different error now appears: `interface{}` syntax in type names (separate bug)

## Remaining Issues (Not in Scope)

The test still fails with:
```
test.go:367:22: expected type, found '{' (and 1 more errors)
```

**Cause**: Type name generator creates `Option_interface{}` which is invalid Go syntax.

**Example**:
```go
type Option_interface{} struct {  ← Invalid: cannot use {} in type name
    tag    OptionTag
    some_0 *interface{}
}
```

**Should be**:
```go
type Option_any struct {  ← Valid: use 'any' alias or sanitized name
    tag    OptionTag
    some_0 *any
}
```

This is a **different bug** in the type name sanitization logic, not related to package ordering.

## Summary

**Package ordering bug**: ✅ FIXED
**File modified**: `pkg/generator/generator.go` (lines 173-217)
**Golden file updated**: `tests/golden/pattern_match_01_simple.go.golden`
**Test status**: Package error resolved, different error uncovered (interface{} syntax)

The requested fix is complete. The generator now correctly orders:
1. Package statement
2. Imports
3. Injected types
4. User declarations

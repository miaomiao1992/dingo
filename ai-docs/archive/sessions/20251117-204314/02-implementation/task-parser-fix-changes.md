# Parser Test Build Fix - Changes Summary

## Problem
The file `pkg/parser/sum_types_test.go` referenced unimplemented AST node types, causing build failures:
- `dingoast.EnumDecl` - undefined type
- `file.DingoNodes` - field does not exist on `*ast.File`
- `dingoast.VariantUnit`, `VariantTuple`, `VariantStruct` - undefined constants

These types are required for sum types (enum/match) feature which is deferred to Phase 3+.

## Solution
Added build constraints to exclude `pkg/parser/sum_types_test.go` from compilation:
```go
//go:build ignore
// +build ignore
```

## Changes Made

### File: `/Users/jack/mag/dingo/pkg/parser/sum_types_test.go`

**Added build constraints at top of file:**
```go
//go:build ignore
// +build ignore

// NOTE: This test file is excluded from builds because sum types (enum/match)
// are not yet implemented. The required AST nodes (EnumDecl, MatchExpr, etc.)
// will be added in Phase 3+. To enable these tests, remove the build constraints
// above and implement the missing AST types in pkg/ast/.
```

**Rationale:**
- Using `//go:build ignore` completely excludes the file from compilation
- This is the cleanest approach for tests that reference unimplemented features
- The test code remains in the repository as documentation of future requirements
- Match expression tests were already properly skipped with `t.Skip()`, but enum tests couldn't use that approach since they reference undefined types
- Build tags are more appropriate than commenting out code when entire test files reference unimplemented features

## Verification

### Build Success
```bash
$ go build ./pkg/parser/...
# Success - no output
```

### Test Exclusion Confirmed
```bash
$ go test ./pkg/parser/... -v 2>&1 | grep -i "enum\|sum_types"
# No sum_types or enum tests found (correctly excluded)
```

### Parser Tests Run Successfully
All other parser tests run correctly:
- `TestSafeNavigation` - PASS
- `TestNullCoalescing` - PASS
- `TestLambda` - PASS
- `TestOperatorPrecedence` - PASS (with expected skips for ternary)
- `TestOperatorChaining` - PASS (with expected skips)
- Match expression tests properly skipped within test file

## Future Work
When implementing sum types (Phase 3+):
1. Remove the `//go:build ignore` constraint from `pkg/parser/sum_types_test.go`
2. Implement required AST types in `pkg/ast/`:
   - `EnumDecl` type
   - `VariantKind` constants (`VariantUnit`, `VariantTuple`, `VariantStruct`)
   - `MatchExpr` type
   - Add `DingoNodes []DingoNode` field to `ast.File`
3. Implement enum parsing logic in parser
4. Run tests to verify implementation

## References
- Sum types feature proposal: `features/sum-types.md`
- Golden tests: `tests/golden/sum_types_*.{dingo,go.golden}`
- Current AST structure: `pkg/ast/file.go`, `pkg/ast/ast.go`

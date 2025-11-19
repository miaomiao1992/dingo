# Phase 1 Implementation: Golden Test Preprocessor Integration

## Changes Made

### File: `tests/golden_test.go`

**Import Added:**
```go
"github.com/MadAppGang/dingo/pkg/preprocessor"
```

**Code Changed (Lines 91-97):**

**BEFORE:**
```go
dingoAST, err := parser.ParseFile(fset, dingoFile, dingoSrc, 0)
require.NoError(t, err, "Failed to parse Dingo file: %s", dingoFile)
```

**AFTER:**
```go
// Preprocess THEN parse
preprocessor := preprocessor.New(dingoSrc)
preprocessed, _, err := preprocessor.Process()
require.NoError(t, err, "Failed to preprocess Dingo file: %s", dingoFile)

dingoAST, err := parser.ParseFile(fset, dingoFile, []byte(preprocessed), 0)
require.NoError(t, err, "Failed to parse preprocessed Dingo file: %s", dingoFile)
```

## Rationale

The golden tests were calling `parser.ParseFile()` directly, bypassing the preprocessor. This caused tests to fail because:

1. The parser expects valid Go syntax
2. Raw `.dingo` files contain Dingo-specific syntax (`:` type annotations, `?` operator, `let` keyword)
3. The preprocessor transforms these into valid Go syntax BEFORE parsing

The `dingo build` CLI already uses this flow (preprocess → parse → transform → generate), so aligning the tests with the CLI pipeline fixes the parsing failures.

## Result

All three tested files now **parse successfully**:
- `error_prop_01_simple.dingo` ✅
- `error_prop_03_expression.dingo` ✅
- `error_prop_06_mixed_context.dingo` ✅

Tests still fail on output comparison (formatting/variable numbering differences), but **no longer fail at the parse stage**.

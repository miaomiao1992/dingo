
[claudish] Model: minimax/minimax-m2

# Enum Variant Naming Convention Analysis

## Current Implementation

Based on my investigation of the Dingo codebase, here's the enum variant naming convention validation:

### 1. **Regex-Based Validation** (`pkg/preprocessor/enum.go:16-23`)

The enum preprocessor uses three regex patterns to validate variant names:

- **Unit variants**: `^\s*(\w+)\s*,?\s*$`
- **Struct variants**: `^\s*(\w+)\s*\{\s*([^}]*)\s*\}\s*,?\s*$`
- **Tuple variants**: `^\s*(\w+)\s*\(([^)]*)\)\s*,?\s*$`

All patterns use `\w+` which matches:
- Letters (a-z, A-Z)
- Digits (0-9)
- Underscores (_)

### 2. **Identifier Character Validation** (`pkg/preprocessor/enum.go:197-199`)

```go
func isIdentifierChar(ch byte) bool {
    return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || 
           (ch >= '0' && ch <= '9') || ch == '_'
}
```

### 3. **Current Convention** (from tests/golden files)

**Variant names**: PascalCase (`Pending`, `Active`, `Complete`, `Point`, `Circle`, `Rectangle`)

**Generated code**:
- **Constants**: `EnumName_VariantName` (e.g., `StatusTag_Pending`)
- **Constructors**: `EnumName_VariantName()` (e.g., `Status_Pending()`)
- **Methods**: `IsVariantName()` (e.g., `IsPending()`)
- **Fields**: `variantname_fieldname` (lowercase with underscore, e.g., `circle_radius`)

## Validation Results

✅ **ALLOWED**: Letters, digits, underscores (standard identifier characters)
❌ **NOT ALLOWED**: Spaces, special characters, hyphens, dots
✅ **CASE SENSITIVE**: `Pending` ≠ `pending` (both would be valid but different)

## Analysis

The current implementation correctly validates variant names according to Go identifier rules. No explicit enforcement of PascalCase exists, but it's the convention used in all tests and examples.

[claudish] Shutting down proxy server...
[claudish] Done


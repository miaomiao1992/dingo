# Action Items: Build Fix Implementation
**Session:** 20251117-204314
**Date:** 2025-11-17
**Priority:** CRITICAL and IMPORTANT issues only

---

## CRITICAL Issues (Must Fix Before Merge)

### 1. Fix Source Mapping Column/Length Accuracy
**File:** `pkg/preprocessor/error_prop.go`
**Lines:** 303-319, 328-330, 348-356, 363-369, 406-493

Calculate exact position and length of `?` operator instead of using column 1 and full match length.

**Implementation:**
- Find `qPos := strings.Index(expr, "?")`
- Set `OriginalColumn: qPos + 1` (1-based)
- Set `Length: 1` (length of `?` character)

**Source:** Grok CRITICAL #1

---

### 2. Fix Source Map Offset Calculation for Pre-Import Mappings
**File:** `pkg/preprocessor/preprocessor.go`
**Lines:** 93-104, 166-170

Only shift mappings for lines AFTER import insertion point, not all mappings.

**Implementation:**
```go
func (p *Preprocessor) adjustMappingsForImports(numImportLines int, importInsertionLine int) {
    for i := range p.mappings {
        if p.mappings[i].GoLine >= importInsertionLine {
            p.mappings[i].GoLine += numImportLines
        }
    }
}
```

**Source:** GPT-5.1 CRITICAL #1

---

### 3. Fix Multi-Value Return Handling in Error Propagation
**File:** `pkg/preprocessor/error_prop.go`
**Lines:** 477-487

Handle functions that return multiple non-error values (e.g., `(A, B, error)`).

**Implementation:**
- Detect number of return values from function signature
- Generate: `return tmp1, tmp2, ..., tmpN, nil` for multi-value
- Generate: `return tmp, nil` for single value
- Add tests covering multi-value returns

**Source:** GPT-5.1 CRITICAL #2

---

### 4. Add Safe Type Assertion in Transformer
**File:** `pkg/transform/transformer.go`
**Line:** 48

Replace unsafe type assertion with checked assertion.

**Implementation:**
```go
if f, ok := result.(*ast.File); ok {
    return f, nil
}
return nil, fmt.Errorf("unexpected return type from astutil.Apply")
```

**Source:** Grok CRITICAL #1

---

### 5. Document cursor.Replace() Requirement for Transform Methods
**File:** `pkg/transform/transformer.go`
**Lines:** 107, 114, 120

Add code comments explaining that future implementations MUST use `cursor.Replace()`.

**Implementation:**
```go
// TODO: Implement lambda transformation
// MUST call cursor.Replace(transformedNode) to replace placeholder
// or transformation will be a no-op
```

**Source:** Grok CRITICAL #3 (documentation fix, not active bug)

---

## IMPORTANT Issues (Should Fix Before Merge)

### 6. Prevent Import Detection False Positives
**File:** `pkg/preprocessor/error_prop.go`
**Lines:** 29-64, 741-761

Avoid injecting stdlib imports for user-defined functions with common names.

**Recommended Approach (Conservative):**
- Only inject imports for package-qualified calls (e.g., `os.ReadFile()`)
- Update `trackFunctionCallInExpr` to require package qualifier
- Document that unqualified calls don't trigger import injection

**Alternative Approaches:**
- AST resolution: Parse and resolve identifiers before injecting
- Whitelist syntax: Special opt-in syntax for import injection

**Add Test:**
```go
func TestNoImportForUserDefinedFunctions(t *testing.T) {
    // Verify user-defined ReadFile doesn't inject os import
}
```

**Source:** GPT-5.1 IMPORTANT #1, Internal IMPORTANT-1

---

### 7. Return Errors Instead of Silent Fallback in injectImports
**File:** `pkg/preprocessor/preprocessor.go`
**Line:** 124

Change `injectImports()` to return error instead of silently returning original source.

**Implementation:**
```go
func injectImports(source []byte, needed []string) ([]byte, error) {
    fset := token.NewFileSet()
    node, err := parser.ParseFile(fset, "", source, parser.ParseComments)
    if err != nil {
        return nil, fmt.Errorf("failed to parse source for import injection: %w", err)
    }
    // ... rest of implementation
    return buf.Bytes(), nil
}
```

Update calling code in `Process()` to handle the error.

**Source:** Grok IMPORTANT #1, Internal MINOR-3

---

### 8. Add Defensive Validation for Source Map Adjustments
**File:** `pkg/preprocessor/preprocessor.go`
**Lines:** 95-104

Add validation to catch import injection bugs early.

**Implementation:**
```go
// Before adjustment
if importLinesAdded < 0 {
    return "", nil, fmt.Errorf("import injection reduced line count (bug?): %d", importLinesAdded)
}

// After adjustment, verify mapping sanity
for _, m := range sourceMap.Mappings {
    if m.GeneratedLine < 1 {
        return "", nil, fmt.Errorf("invalid mapping after import adjustment: line %d", m.GeneratedLine)
    }
}
```

**Source:** Internal IMPORTANT-2

---

### 9. Add Placeholder Validation (Beyond String Prefix)
**File:** `pkg/transform/transformer.go`
**Lines:** 77-91

Validate placeholder structure, not just prefix match.

**Implementation:**
```go
if strings.HasPrefix(name, "__dingo_lambda_") {
    if !isValidLambdaPlaceholder(call) {
        return true // Not a placeholder, skip
    }
    return t.transformLambda(cursor, call)
}

func isValidLambdaPlaceholder(call *ast.CallExpr) bool {
    // Check argument structure matches lambda placeholder pattern
    return len(call.Args) >= 1 // or other validation
}
```

**Source:** Grok IMPORTANT #2

---

## Summary

**Total Critical Issues:** 5 (3 active bugs, 1 safety fix, 1 documentation)
**Total Important Issues:** 4

**Estimated Fix Time:**
- Critical: 4-6 hours
- Important: 2-3 hours
- **Total: 6-9 hours**

**Testing Requirements:**
- Add negative tests for import false positives
- Add tests for multi-value returns
- Add tests for pre-import source mapping
- Verify all existing tests still pass

# Action Items: Phase 2.2 Critical and Important Fixes

**Session:** 20251117-183117
**Priority:** Critical and Important issues only
**Total Items:** 15

---

## CRITICAL Priority (Must Fix Before Merge)

### 1. Fix Import Detection Logic
**File:** `pkg/preprocessor/error_prop.go:421`
**Issue:** String contains check produces false positives for `"fmt"` in comments/strings
**Action:** Replace with `go/ast` parsing and `astutil.AddImport`
**Estimated Effort:** 1-2 hours

```go
func (e *ErrorPropProcessor) ensureFmtImport(source []byte) []byte {
    fset := token.NewFileSet()
    file, err := parser.ParseFile(fset, "", source, parser.ImportsOnly)
    if err != nil {
        return e.insertFmtImportSimple(source)
    }

    for _, imp := range file.Imports {
        if imp.Path.Value == `"fmt"` {
            return source
        }
    }

    astutil.AddImport(fset, file, "fmt")
    var buf bytes.Buffer
    printer.Fprint(&buf, fset, file)
    return buf.Bytes()
}
```

---

### 2. Move Regex Compilation to Package Level
**Files:** `pkg/preprocessor/error_prop.go:95,106,126`, `type_annot.go:71`, `keywords.go:24`
**Issue:** Regexes compiled on every line (O(n) performance issue)
**Action:** Declare regexes as package-level variables
**Estimated Effort:** 30 minutes

```go
// At package level in each file
var (
    assignPattern = regexp.MustCompile(`^\s*(let|var)\s+(\w+)\s*=\s*(.+)$`)
    returnPattern = regexp.MustCompile(`^\s*return\s+(.+)$`)
    msgPattern    = regexp.MustCompile(`^(.*\?)\s*"((?:[^"\\]|\\.)*)"`)
)
```

---

### 3. Fix Zero Value Generation for Custom Types
**File:** `pkg/preprocessor/error_prop.go:392`
**Issue:** Returns `"nil"` for custom structs, causing compile errors
**Action:** Use `typ + "{}"` for non-pointer, non-reference types
**Estimated Effort:** 1 hour

```go
// Custom type → use composite literal for non-pointer types
if !strings.HasPrefix(typ, "*") && !strings.HasPrefix(typ, "[]") &&
   !strings.HasPrefix(typ, "map[") && !strings.HasPrefix(typ, "chan ") &&
   !strings.HasPrefix(typ, "func(") && !strings.HasPrefix(typ, "interface{") {
    return typ + "{}"
}
return "nil"
```

---

### 4. Fix Return Statement Expansion
**File:** `pkg/preprocessor/error_prop.go:205-268`
**Issue:** Expansion for `return expr?` only returns first value, drops error
**Action:** Ensure complete return tuple is generated
**Estimated Effort:** 2 hours

Review `expandReturn` method to ensure it generates:
```go
__tmpN, __errN := expr
if __errN != nil {
    return zeroValues..., __errN
}
return __tmpN, nil  // Include all required return values
```

---

### 5. Add Bounds Checking to Function Signature Parsing
**File:** `pkg/preprocessor/error_prop.go:278-341`
**Issue:** Infinite loop if no `{` found; fails for generics and interface methods
**Action:** Add safety limit and robust brace detection
**Estimated Effort:** 2-3 hours

```go
func (e *ErrorPropProcessor) parseFunctionSignature(startLine int) *funcContext {
    var funcText strings.Builder
    foundBrace := false

    for i := startLine; i < len(e.lines) && i < startLine+20; i++ {
        funcText.WriteString(e.lines[i])
        funcText.WriteString("\n")

        trimmed := strings.TrimSpace(e.lines[i])
        if idx := strings.Index(trimmed, "{"); idx != -1 {
            if !strings.HasPrefix(trimmed, "//") {
                foundBrace = true
                break
            }
        }
    }

    if !foundBrace {
        return &funcContext{
            returnTypes: []string{},
            zeroValues:  []string{"nil"},
        }
    }
    // ... rest of parsing with validation
}
```

---

### 6. Replace Type Annotation Regex with Scanner
**File:** `pkg/preprocessor/type_annot.go:71`
**Issue:** Regex cannot handle qualified types, generics, function types, channels
**Action:** Use `go/scanner` for proper tokenization
**Estimated Effort:** 3-4 hours

Implement token-based parameter parsing that handles:
- Qualified types: `param: pkg.Type`
- Generics: `param: map[string][]interface{}`
- Function types: `param: func(int) error`
- Channel types: `param: <-chan string`

---

## IMPORTANT Priority (Should Fix Soon)

### 7. Support Multi-line Expressions
**File:** `pkg/preprocessor/error_prop.go:121-133`
**Issue:** Line-based processing skips multi-line expressions
**Action:** Implement token-based processing or join incomplete lines
**Estimated Effort:** 3-4 hours

Detect incomplete expressions (unclosed parens, trailing operators) and buffer lines until complete.

---

### 8. Handle Multi-line Function Signatures
**File:** `pkg/preprocessor/type_annot.go:35-66`
**Issue:** Only processes single line with `func`, fails on multi-line params
**Action:** Track paren depth across lines
**Estimated Effort:** 2 hours

```go
inFuncParams := false
parenDepth := 0

for i, line := range lines {
    if bytes.Contains(line, []byte("func ")) {
        inFuncParams = true
    }

    if inFuncParams {
        parenDepth += bytes.Count(line, []byte("("))
        parenDepth -= bytes.Count(line, []byte(")"))
        line = t.replaceColonInParams(line)

        if parenDepth <= 0 {
            inFuncParams = false
        }
    }
    // ...
}
```

---

### 9. Fix Keyword Replacement to Avoid Strings/Comments
**File:** `pkg/preprocessor/keywords.go:20-27`
**Issue:** Regex replaces `let` in string literals and comments
**Action:** Use `go/scanner` or restrict regex to statement positions
**Estimated Effort:** 2 hours

Use tokenizer to identify actual keyword declarations vs literals.

---

### 10. Add Error Logging for Silent Failures
**File:** `pkg/preprocessor/error_prop.go:290-299`
**Issue:** Parse failures silently fall back to `nil` without reporting
**Action:** Return errors and log warnings
**Estimated Effort:** 1-2 hours

```go
func (e *ErrorPropProcessor) parseFunctionSignature(startLine int) (*funcContext, error) {
    // ... parsing logic ...

    file, err := parser.ParseFile(fset, "", src, 0)
    if err != nil {
        return nil, fmt.Errorf("failed to parse function signature at line %d: %w", startLine, err)
    }

    // ... rest of logic ...
}
```

Update callers to handle errors appropriately.

---

### 11. Escape Error Messages for fmt.Errorf
**File:** `pkg/preprocessor/error_prop.go:261`
**Issue:** Error messages with `%` cause runtime panics
**Action:** Escape `%` characters
**Estimated Effort:** 30 minutes

```go
if errMsg != "" {
    escapedMsg := strings.ReplaceAll(errMsg, "%", "%%")
    e.needsFmt = true
    errPart = fmt.Sprintf(`fmt.Errorf("%s: %%w", %s)`, escapedMsg, errVar)
}
```

---

### 12. Update Source Map for Plugin Pipeline
**File:** `cmd/dingo/main.go:217-277`
**Issue:** Source map not updated after plugin generator transformations
**Action:** Either disable map emission or extend generator to emit deltas
**Estimated Effort:** 2-3 hours

Document limitation or implement proper pipeline integration.

---

### 13. Implement Source Map Composition
**File:** `pkg/preprocessor/preprocessor.go:53-74`
**Issue:** Pipeline concatenates mappings instead of composing them
**Action:** Trace through chained transformations
**Estimated Effort:** 3-4 hours

```go
func (sm *SourceMap) ComposeWith(next *SourceMap) {
    for _, nextMapping := range next.Mappings {
        origLine, origCol := sm.MapToOriginal(nextMapping.OriginalLine, nextMapping.OriginalColumn)

        composed := Mapping{
            OriginalLine:    origLine,
            OriginalColumn:  origCol,
            GeneratedLine:   nextMapping.GeneratedLine,
            GeneratedColumn: nextMapping.GeneratedColumn,
            Length:          nextMapping.Length,
            Name:            nextMapping.Name,
        }

        sm.AddMapping(composed)
    }
}
```

---

### 14. Create Complete Mappings for Expansions
**File:** `pkg/preprocessor/error_prop.go:179-186, 233-240`
**Issue:** Only one mapping per expansion, but generates 7 lines
**Action:** Map all generated lines to original
**Estimated Effort:** 1-2 hours

```go
mappings := []Mapping{
    {OriginalLine: lineNum, GeneratedLine: lineNum, Length: len(tmpVarLine)},
    {OriginalLine: lineNum, GeneratedLine: lineNum + 2},  // if statement
    {OriginalLine: lineNum, GeneratedLine: lineNum + 3},  // return
    {OriginalLine: lineNum, GeneratedLine: lineNum + 7},  // var assignment
}
```

---

### 15. Fix or Remove Ternary Detection
**File:** `pkg/preprocessor/error_prop.go:406-414`
**Issue:** Simplistic detection produces false positives
**Action:** Return `false` since ternary not yet supported
**Estimated Effort:** 15 minutes

```go
func (e *ErrorPropProcessor) isTernaryLine(line string) bool {
    // Dingo doesn't support ternary yet (Phase 2.3)
    // This is a placeholder for future implementation
    return false
}
```

---

## Summary

**Total Action Items:** 15
- **CRITICAL:** 6 items (8-13 hours estimated)
- **IMPORTANT:** 9 items (15-22 hours estimated)

**Total Estimated Effort:** 23-35 hours

**Recommended Approach:**
1. Fix CRITICAL items 1-6 first (block merge)
2. Verify with existing golden tests
3. Add edge case tests for each fix
4. Address IMPORTANT items 7-15 before production release

**Priority Order:**
1. Import detection (most likely to cause failures)
2. Zero value generation (breaks common patterns)
3. Regex compilation (performance issue)
4. Function signature parsing (generics support)
5. Return statement expansion (correctness)
6. Type annotation regex (expressiveness)
7. Multi-line support (usability)
8. Error handling (debugging)
9. Source maps (IDE integration)

# Enum Naming Convention - Final Recommendation

## Executive Summary

**UNANIMOUS VERDICT**: All 6 expert analyses recommend **Option B (Pure CamelCase)**

## Model Consensus

| Model | Recommendation | Key Rationale |
|-------|----------------|---------------|
| **MiniMax M2** | Option B (CamelCase) | Aligns with Go idioms (http.MethodGet, ast.BadDecl pattern) |
| **Grok Code Fast** | Option B (CamelCase) | Matches Go standard library, improves developer experience |
| **GPT-5.1 Codex** | Option B (CamelCase) | Generated code looks hand-written, better IDE autocompletion |
| **Gemini 3 Pro** | Option B (CamelCase) | Standard Go convention, no cognitive overhead |
| **Sherlock Think Alpha** | Option B (CamelCase) | Maximum Go idiomaticity and simplicity |
| **Internal (Sonnet 4.5)** | Option B (CamelCase) | Ecosystem overwhelmingly favors CamelCase |

**Vote**: 6/6 for Option B (100% consensus)

## Recommended Changes

### Current (Option A - Underscore):
```go
// Constructors
Value_Int(42)
Result_Ok(value)
Option_Some(data)

// Tags
ValueTag_Int
ResultTag_Ok
OptionTag_Some

// Fields
int_0, string_0
ok_0, err_0
some_0
```

### Recommended (Option B - CamelCase):
```go
// Constructors
ValueInt(42)
ResultOk(value)
OptionSome(data)

// Tags
ValueTagInt
ResultTagOk
OptionTagSome

// Fields
int0, string0
ok0, err0
some0
```

## Key Benefits

1. **Go Idiomaticity** ✅
   - Matches standard library: `http.MethodGet`, `http.MethodPost`
   - Follows AST patterns: `ast.BadDecl`, `ast.GenDecl`
   - Aligns with error conventions: `io.EOF`, `sql.ErrNoRows`

2. **Developer Experience** ✅
   - Generated code looks hand-written
   - No visual "foreign language" marker (underscores)
   - Better IDE autocompletion grouping
   - Familiar to Go developers

3. **Code Quality** ✅
   - Passes `golint` and `go vet` without warnings
   - Consistent with Go naming conventions
   - Professional appearance

4. **Ecosystem Integration** ✅
   - Works seamlessly with Go tools
   - gopls provides better suggestions
   - Easier code review for Go teams

## Migration Path

### Phase 1: Update Codegen (2-3 hours)
1. Modify `pkg/preprocessor/rust_match.go`:
   - Change `Value_Int` → `ValueInt` in constructor generation
   - Change `ValueTag_Int` → `ValueTagInt` in tag generation
   - Change `int_0` → `int0` in field naming

2. Update template strings:
   - `fmt.Sprintf("%s_%s", typeName, variant)` → `fmt.Sprintf("%s%s", typeName, variant)`
   - `fmt.Sprintf("%sTag_%s", typeName, variant)` → `fmt.Sprintf("%sTag%s", typeName, variant)`
   - `strings.ToLower(variant) + "_0"` → `strings.ToLower(variant) + "0"`

### Phase 2: Regenerate Golden Tests (1 hour)
```bash
# Regenerate all golden files
for file in tests/golden/*.dingo; do
  ./dingo build "$file" -o "${file%.dingo}.go.golden"
done

# Run tests
go test ./tests -v
```

### Phase 3: Update Documentation (30 minutes)
- Update README.md examples
- Update CLAUDE.md with new naming
- Update any tutorial/guide files

### Total Estimated Time: 3-4 hours

## Edge Cases & Solutions

### 1. Acronyms
**Problem**: `HTTPStatus` or `HttpStatus`?
**Solution**: Follow Go convention - `HTTPStatus` (keep acronyms uppercase)

### 2. Name Collisions
**Problem**: `type Value struct` + `ValueInt` constructor collision?
**Solution**: Not possible - constructors are functions, types are types (different namespaces)

### 3. Zero-Arity Variants
**Current**: `Option_None()`
**Recommended**: `OptionNone()` (still a function for consistency)

### 4. Generic Types
**Current**: `Result_Ok`, `Option_Some`
**Recommended**: `ResultOk`, `OptionSome` (no change to logic)

## Backward Compatibility

**Breaking Change**: Yes, but pre-v1.0, acceptable

**Migration for Users**:
- Global find-replace: `Value_Int` → `ValueInt`
- Global find-replace: `Result_Ok` → `ResultOk`
- Global find-replace: `Option_Some` → `OptionSome`
- Automated via script if needed

## Real-World Validation

All models referenced these Go standard library examples:

```go
// net/http package
http.MethodGet
http.MethodPost
http.StatusOK
http.StatusNotFound

// go/ast package
ast.BadDecl
ast.GenDecl
ast.FuncDecl

// go/token package
token.ILLEGAL
token.IDENT
token.INT

// Standard errors
io.EOF
sql.ErrNoRows
context.Canceled
```

**Pattern**: `{Package}{Concept}{Variant}` with CamelCase, no underscores

## Final Recommendation

✅ **Adopt Option B (Pure CamelCase) immediately**

**Reasoning**:
1. Unanimous expert consensus (6/6 models)
2. Aligns with Dingo's design principle: "Generated Go should look hand-written"
3. Low migration cost (mechanical transformation)
4. High benefit (Go idiomaticity, developer experience)
5. No technical blockers

**Action Items**:
1. Update codegen in `pkg/preprocessor/rust_match.go`
2. Regenerate all golden tests
3. Update documentation
4. Add migration note to CHANGELOG.md

**Timeline**: Complete in 1 day or less

---

**Conclusion**: The underscore convention (`Value_Int`) was never Go-idiomatic. Switching to CamelCase (`ValueInt`) is a clear win with zero technical downsides and massive UX improvements for Go developers.

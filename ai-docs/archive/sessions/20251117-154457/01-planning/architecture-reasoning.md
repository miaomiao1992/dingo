# Architecture Reasoning: Why go/parser + Preprocessor?

**Date:** 2025-11-17
**Decision:** Replace Participle with go/parser + preprocessor architecture
**Status:** Approved for implementation
**Impact:** Foundational (affects entire project)

---

## Executive Summary

We are migrating from a custom Participle-based parser to a two-stage architecture: **preprocessor → go/parser → AST transformation**. This decision is driven by reducing maintenance burden, leveraging Go's battle-tested parsing infrastructure, and enabling better integration with Go tooling.

**Core Insight:** Instead of maintaining a parallel grammar for Dingo, we transform Dingo syntax into valid Go syntax, parse it with go/parser, then transform the AST. This shifts complexity from "parsing" to "transformation," which is more tractable and better supported by Go's standard library.

---

## The Core Question

**How should we parse Dingo source code?**

Dingo is a meta-language for Go (like TypeScript for JavaScript). It has custom syntax (`?`, `|x| expr`, `match`, `enum`, `??`, `?.`) that Go does not support. We need to:

1. Parse Dingo syntax
2. Transform it to Go syntax
3. Maintain source positions for error messages and LSP
4. Generate clean, idiomatic Go code

**Three Approaches Considered:**

1. **Custom parser (Participle)** - Current implementation
2. **Tree-sitter** - Modern parsing framework
3. **Preprocessor + go/parser** - Proposed architecture

---

## Approach 1: Custom Parser (Participle)

### How It Works

Participle is a parser generator that uses struct tags to define grammar:

```go
type DingoFile struct {
    Package string       `"package" @Ident`
    Imports []Import     `"import" @@ ...`
    Decls   []Decl       `@@*`
}
```

We maintain a complete Dingo grammar separate from Go grammar. The parser produces a custom AST, which we then transform to a Go AST.

### Advantages

- ✅ **Clean syntax definition:** Grammar expressed as Go structs
- ✅ **Type-safe parsing:** Parser errors caught at compile time
- ✅ **Good for small languages:** Works well for simple DSLs
- ✅ **Currently working:** We have a functional implementation

### Disadvantages

- ❌ **Grammar maintenance burden:** Must keep Dingo grammar in sync with Go grammar
- ❌ **Incomplete Go support:** Go grammar is complex (500+ productions), we've only implemented a subset
- ❌ **Divergence risk:** As Go evolves, our grammar lags behind
- ❌ **Double transformation:** Custom AST → Go AST → Go code (two transformations)
- ❌ **Limited tooling:** Can't use go/types, go/ast utilities on custom AST
- ❌ **Error messages:** Participle errors are generic, not Go-specific

### Why We're Moving Away

The fundamental issue is **maintenance burden**. Every Go syntax feature we want to support in Dingo requires:

1. Adding it to Participle grammar
2. Testing parsing
3. Transforming custom AST to Go AST
4. Testing transformation

This doubles the work for every feature. As Dingo grows, this becomes unsustainable.

**Example:** Supporting Go 1.18 generics would require:
- Adding generic syntax to Participle grammar (~200 lines)
- Handling in AST transformation (~300 lines)
- Extensive testing

With go/parser, generics are already supported. Zero work needed.

---

## Approach 2: Tree-sitter

### How It Works

Tree-sitter is a modern parsing library that generates incremental parsers from grammar files. It's used by GitHub, Atom, and other editors for syntax highlighting.

```javascript
// grammar.js
module.exports = grammar({
  name: 'dingo',
  rules: {
    source_file: $ => repeat($._definition),
    error_prop: $ => seq($.expression, '?'),
    // ... more rules
  }
});
```

### Advantages

- ✅ **Incremental parsing:** Fast re-parsing on edits (good for LSP)
- ✅ **Error recovery:** Can parse incomplete/invalid code
- ✅ **Good tooling:** Used widely in editors
- ✅ **Separate from Go:** Can evolve Dingo syntax independently

### Disadvantages

- ❌ **Grammar maintenance:** Still need to maintain complete grammar
- ❌ **Learning curve:** New language (JavaScript) for grammar definition
- ❌ **Integration complexity:** AST is not Go AST, still need transformation
- ❌ **Overkill for our use case:** We don't need incremental parsing (compiler, not editor)
- ❌ **Tooling gap:** Can't use go/types on Tree-sitter AST

### Why We're Not Choosing This

Tree-sitter solves a different problem: **editor integration with incremental parsing**. Our primary use case is **batch compilation** (.dingo → .go). We don't need incremental parsing performance.

Additionally, we still face the same issue as Participle: maintaining a separate grammar and transforming to Go AST. The work is similar, but with a different tool.

**Key Insight:** Tree-sitter is great for language servers (we may use it for dingo-lsp), but not optimal for the transpiler.

---

## Approach 3: Preprocessor + go/parser (Proposed)

### How It Works

Instead of parsing Dingo syntax directly, we transform it into valid Go syntax with semantic placeholders, then parse with go/parser.

**Example:**

```go
// Input: Dingo source
x := fetchData()?

// Step 1: Preprocessor transforms to valid Go
x := __dingo_try_1__(fetchData())

// Step 2: go/parser creates AST
// AST contains CallExpr to __dingo_try_1__

// Step 3: AST transformation replaces placeholder
// Replace __dingo_try_1__(expr) with proper error handling:
__tmp_1, __err_1 := fetchData()
if __err_1 != nil {
    return __err_1
}
x := __tmp_1
```

The preprocessor does simple string transformations. The AST transformer does semantic transformations using type information.

### Advantages

- ✅ **No grammar maintenance:** go/parser handles all Go syntax
- ✅ **Automatic Go support:** New Go features work immediately
- ✅ **Leverage go/types:** Full type checking and inference
- ✅ **Leverage go/ast utilities:** astutil, ast/inspector, etc.
- ✅ **Better error messages:** Standard Go parser errors
- ✅ **Simpler architecture:** Clear separation of concerns
- ✅ **Easier testing:** Each component independently testable
- ✅ **Proven pattern:** Used by TypeScript, Babel, other transpilers

### Disadvantages

- ❌ **Two-pass processing:** Preprocess, then parse (slightly slower)
- ❌ **Source map complexity:** Must track position mappings
- ❌ **Placeholder namespace:** Must avoid collisions with user code
- ❌ **Not a "pure" parser:** Combines text transformation + AST manipulation

### Why This Is The Right Choice

**1. Maintenance Burden:** Zero. We don't maintain a grammar. As Go evolves, we automatically support new syntax.

**2. Tooling Integration:** We get go/types for free. This enables:
- Type inference for lambdas
- Type checking for pattern matching
- Integration with gopls (via dingo-lsp)

**3. Proven Pattern:** TypeScript uses a similar approach:
- TypeScript syntax → JavaScript syntax (with type erasure markers)
- Parse with JavaScript parser
- Transform AST to remove/transform TypeScript-specific features

**4. Simplicity:** Each component has one job:
- Preprocessor: Dingo syntax → Go syntax (string manipulation)
- go/parser: Go syntax → AST (handled by standard library)
- Transformer: Placeholder AST → final Go AST (semantic transformation)

This is easier to reason about than a monolithic parser + transformer.

**5. Extensibility:** Adding new Dingo syntax is straightforward:
- Add preprocessor for the syntax (transform to valid Go)
- Add transformer to replace placeholders
- No grammar changes needed

---

## Deep Dive: Design Decisions

### Decision 1: Preprocessor vs Direct Token Stream Manipulation

**Question:** Should we manipulate tokens before feeding to go/parser, or transform source text?

**Options:**

A. **Token stream manipulation:**
   - Read Dingo tokens
   - Transform tokens (? → function call tokens)
   - Feed to go/parser

B. **Source text transformation (chosen):**
   - Scan Dingo source as text
   - Transform to Go source text
   - Feed to go/parser as string

**Chosen:** B (source text transformation)

**Reasoning:**

Token stream manipulation is theoretically more efficient, but practically more complex:

- Go's token package is designed for Go tokens, not custom tokens
- We'd need to track token positions carefully
- Error messages would be confusing (tokens don't match source)

Source text transformation is simpler:
- Use regex or simple scanning
- Transform text directly
- Easy to debug (can print preprocessed output)
- Source maps are straightforward (line/column offsets)

**Trade-off:** Slightly less efficient (text manipulation overhead), but much simpler to implement and maintain.

### Decision 2: Placeholder Naming Convention

**Question:** What should placeholders look like?

**Options:**

A. `__DINGO_TRY__(expr)` - All caps, obvious
B. `_dingo_try(expr)` - Single underscore, shorter
C. `__dingo_try_N__(expr)` - Double underscore prefix, counter suffix (chosen)

**Chosen:** C (`__dingo_try_N__`)

**Reasoning:**

- Double underscore prefix: Go convention for internal/reserved identifiers (e.g., `__cgo_*`)
- Counter suffix: Enables tracking multiple instances (important for source maps)
- Descriptive: `_try_` clearly indicates error propagation
- Unlikely collision: Users are unlikely to use `__dingo_*` prefix

**Alternative considered:** Could use private package identifiers like `__dingo.try()`, but:
- Requires creating a fake package
- More complex to parse
- No significant benefit

### Decision 3: Source Map Format

**Question:** How should we store position mappings?

**Options:**

A. **JavaScript source map format (JSON):**
   - Standard format (sourceMappingURL)
   - Tooling support (browser dev tools, etc.)
   - Base64 VLQ encoding

B. **Custom binary format:**
   - Compact storage
   - Fast to read/write
   - Optimized for our use case

C. **Simple JSON array (chosen for MVP):**
   - Human-readable
   - Easy to debug
   - Simple to implement

**Chosen:** C for MVP, migrate to A for production

**Reasoning:**

For initial implementation, simplicity is more important than efficiency:

```json
{
  "mappings": [
    {
      "preprocessed_line": 10,
      "preprocessed_column": 15,
      "original_line": 10,
      "original_column": 18,
      "length": 1,
      "name": "error_prop"
    }
  ]
}
```

This is easy to inspect, debug, and understand. Later, we can migrate to standard source map format if needed (e.g., for browser integration of LSP diagnostics).

**Trade-off:** Larger file size, slower parsing. Acceptable for MVP (compile times still < 100ms).

### Decision 4: AST Transformation Strategy

**Question:** When should we transform the AST?

**Options:**

A. **During go/parser parsing (custom scanner):**
   - Hook into parser
   - Transform on-the-fly

B. **After parsing, single pass:**
   - Parse complete AST
   - Walk once, transform all placeholders

C. **After parsing, multi-pass (chosen):**
   - Parse complete AST
   - First pass: Type checking (go/types)
   - Second pass: Transform placeholders

**Chosen:** C (multi-pass)

**Reasoning:**

Multi-pass provides access to type information:

```go
// First pass: Type checking
typeInfo := &types.Info{...}
types.CheckExpr(fset, pkg, expr, typeInfo)

// Now we know types
exprType := typeInfo.TypeOf(expr)

// Second pass: Transform with type knowledge
if isLambda(expr) {
    transformLambdaWithType(expr, exprType)
}
```

This enables:
- Type inference for lambdas
- Type-aware pattern matching
- Better error messages

**Trade-off:** Slightly slower (two passes), but necessary for type-dependent transformations.

### Decision 5: Plugin Architecture

**Question:** How should features integrate?

**Options:**

A. **Monolithic preprocessor:**
   - One big preprocessor handles all features
   - Simple, but less modular

B. **Feature plugins (chosen):**
   - Each feature is a FeatureProcessor
   - Preprocessor orchestrates them
   - Modular, extensible

**Chosen:** B (feature plugins)

**Reasoning:**

Modularity enables:
- Independent testing of each feature
- Easy to add new features (just add a processor)
- Clear separation of concerns
- Parallel development (different people can work on different features)

**Architecture:**

```go
type FeatureProcessor interface {
    Name() string
    Process(source []byte) ([]byte, []Mapping, error)
}

type Preprocessor struct {
    processors []FeatureProcessor
}

func New(source []byte) *Preprocessor {
    return &Preprocessor{
        processors: []FeatureProcessor{
            NewErrorPropProcessor(),
            NewLambdaProcessor(),
            // ... more features
        },
    }
}
```

Each processor is independent, composable, and testable.

**Trade-off:** Slightly more complex orchestration, but worth it for maintainability.

---

## Long-Term Implications

### Implication 1: Go Evolution

As Go adds new syntax (e.g., generics, try/catch proposals), Dingo automatically supports it. No grammar updates needed.

**Example:** When Go 1.18 added generics, existing Participle implementation would require:
- Grammar updates
- AST transformation updates
- Testing

With go/parser approach: Zero work. Generics just work in Dingo files.

### Implication 2: Type System Integration

Having access to go/types enables future features:

- **Type-directed code generation:** Generate different code based on types
- **Type inference:** Infer lambda types from context
- **Type safety:** Validate pattern matching exhaustiveness
- **LSP integration:** Provide type information to language server

This is much harder with a custom AST.

### Implication 3: Tooling Ecosystem

Go has a rich ecosystem of AST tools:

- `golang.org/x/tools/go/ast/astutil` - AST utilities
- `golang.org/x/tools/go/analysis` - Static analysis framework
- `golang.org/x/tools/go/packages` - Package loading
- `golang.org/x/tools/go/ssa` - SSA form construction

We can leverage these for:
- Dingo linters
- Dingo formatters
- Dingo refactoring tools

With a custom AST, we'd have to build all this ourselves.

### Implication 4: Maintainability

Code organization is cleaner:

```
pkg/
  preprocessor/     # String transformations (simple)
  transform/        # AST transformations (semantic)
  parser/           # Wrapper (minimal)
```

vs current:

```
pkg/
  parser/           # Grammar + AST definition + parsing (complex)
  plugins/          # AST transformation (complex)
```

New architecture has clearer boundaries and responsibilities.

### Implication 5: Performance

**Preprocessing overhead:** ~5-10ms for typical file (string scanning)
**go/parser:** ~10-20ms for typical file (same as compiling Go)
**Transformation:** ~5-10ms (AST walking)

**Total:** ~20-40ms (vs ~50-100ms with Participle)

We expect similar or better performance because:
- go/parser is highly optimized
- Preprocessing is simple string manipulation (fast)
- No custom AST → Go AST conversion overhead

---

## Risks and Mitigations

### Risk 1: Source Map Complexity

**Risk:** Position tracking through preprocessing might be error-prone.

**Mitigation:**
- Start with simple 1:1 mappings
- Extensive unit tests for position tracking
- Debug mode to visualize mappings
- Fallback to approximate positions if mapping fails

**Likelihood:** Medium
**Impact:** Medium (incorrect error positions)
**Severity:** Medium

### Risk 2: Placeholder Collisions

**Risk:** User code might use `__dingo_*` identifiers.

**Mitigation:**
- Document reserved `__dingo_*` namespace
- Preprocessor can detect collisions and error
- Use longer, more specific names to reduce collision chance

**Likelihood:** Low
**Impact:** Low (compile error, not silent failure)
**Severity:** Low

### Risk 3: Type Inference Failures

**Risk:** Type inference might fail for complex lambda cases.

**Mitigation:**
- Start with simple cases, validate approach
- Fall back to requiring explicit types
- Provide clear error messages
- Document limitations

**Likelihood:** Medium (lambdas are complex)
**Impact:** Medium (users need to add type annotations)
**Severity:** Low (workaround available)

### Risk 4: Preprocessor Bugs

**Risk:** Preprocessing might introduce syntax errors.

**Mitigation:**
- Extensive unit tests for each feature processor
- Golden tests validate end-to-end
- Debug mode outputs preprocessed code for inspection
- Fallback to legacy parser if issues found

**Likelihood:** Medium (new code, edge cases)
**Impact:** High (breaks compilation)
**Severity:** Medium (caught by tests, fixable)

---

## Alternatives Considered and Rejected

### Alternative 1: Hybrid Approach (Participle for Dingo-specific, go/parser for Go)

**Idea:** Use Participle only for Dingo syntax (`?`, lambdas, etc.), go/parser for rest.

**Why rejected:**
- Complexity: Two parsers, complex stitching logic
- Grammar still needs maintenance
- Loses benefit of leveraging go/parser for everything

### Alternative 2: Transpile to TypeScript, then to Go

**Idea:** Dingo → TypeScript → Go

**Why rejected:**
- TypeScript type system doesn't match Go (interfaces, pointers, etc.)
- Extra compilation step
- No benefit over direct Dingo → Go

### Alternative 3: Use go/scanner directly, manual parsing

**Idea:** Use go/scanner to tokenize, manually build AST.

**Why rejected:**
- Reinventing go/parser (hundreds of lines of complex parsing logic)
- Error-prone, hard to maintain
- No benefit over using go/parser

### Alternative 4: Modify go/parser source code

**Idea:** Fork go/parser, add Dingo syntax support directly.

**Why rejected:**
- Diverges from standard library (can't update with Go releases)
- Complex merge conflicts on Go updates
- Loses community support

---

## Success Criteria

This architecture is successful if:

1. **All existing features work:** 23 golden tests pass
2. **Easier to maintain:** Adding new feature takes < 1 day (vs ~3 days with Participle)
3. **Better error messages:** Positions point to .dingo files correctly
4. **Performance acceptable:** Compile time ≤ current implementation
5. **Extensible:** Can add new syntax features without grammar changes
6. **Type-aware:** Can leverage go/types for advanced transformations

---

## Lessons from TypeScript

TypeScript faced the same problem: how to parse TypeScript syntax (superset of JavaScript)?

**Their solution:**
1. Forked JavaScript parser
2. Added TypeScript syntax productions
3. Transform TypeScript AST → JavaScript AST
4. Generate JavaScript

**Why we're not copying exactly:**
- TypeScript has full control over their parser (their own implementation)
- We don't want to fork go/parser (maintenance burden)
- Preprocessing is simpler for our scale

**What we're learning:**
- Multi-stage compilation is proven (TypeScript is production-grade)
- Source maps are essential (TypeScript invented them)
- Type erasure works (TypeScript's approach to compilation)
- Preserving readability of generated code matters (debugging)

---

## Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| 2025-11-17 | Use preprocessor + go/parser | Reduces maintenance, leverages standard library |
| 2025-11-17 | Placeholder naming: `__dingo_*_N__` | Go convention, unlikely collision, trackable |
| 2025-11-17 | Source map format: Simple JSON (MVP) | Easy to debug, migrate to standard later |
| 2025-11-17 | Multi-pass AST transformation | Enables type inference, better transformations |
| 2025-11-17 | Feature plugin architecture | Modular, testable, extensible |
| 2025-11-17 | Source text preprocessing (not tokens) | Simpler, easier to debug |

---

## Open Questions for Future Consideration

1. **Should we eventually migrate to JavaScript source map format?**
   - Pro: Standard tooling support
   - Con: More complex implementation
   - Decision: Defer until LSP integration needs it

2. **Should preprocessing be cached?**
   - Pro: Faster recompilation
   - Con: More complexity (invalidation, storage)
   - Decision: Defer until performance becomes issue

3. **Should we support source-level debugging (.dingo files in debugger)?**
   - Pro: Better developer experience
   - Con: Requires source map support in delve/GDB
   - Decision: Out of scope for MVP, revisit for v2.0

4. **Should preprocessor be exposed as public API?**
   - Pro: Users could extend Dingo with custom syntax
   - Con: Stability burden, API design complexity
   - Decision: Keep internal for now, revisit if demand exists

---

## Conclusion

The preprocessor + go/parser architecture is the right choice for Dingo because:

1. **It reduces maintenance burden** by eliminating grammar upkeep
2. **It leverages proven tools** (go/parser, go/types) instead of reinventing
3. **It enables advanced features** through type system integration
4. **It follows proven patterns** used by TypeScript and other transpilers
5. **It's simpler to understand** with clear component boundaries

The trade-offs (two-pass processing, source map complexity) are acceptable because:
- Performance is still good (< 100ms compile times)
- Source maps are well-understood (used by all modern transpilers)
- Benefits far outweigh costs

**This is a foundational decision that sets Dingo up for long-term success.**

---

**Approval:** Ready for implementation.

**Next:** Begin Phase 0 (infrastructure setup) per implementation-phases.md.

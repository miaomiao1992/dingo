# Optional Syntax Tree Layer Investigation

## Problem Summary
The current Dingo pipeline relies on a regex-oriented preprocessor to translate Dingo-specific syntax into Go-like text before handing the result to Go's native parser. As features such as pattern matching, richer enums, and advanced control flow expand, regex transformations become increasingly brittle. The request asks whether inserting an optional syntax tree parser layer between preprocessing and go/parser would provide enough structural awareness to simplify downstream plugins without forcing a full parser rewrite.

## Existing Pipeline Snapshot
1. **Stage 1 – Regex Preprocessor:** Stack of focused processors (type annotations, error propagation, enums, keywords) that rewrite `.dingo` code into syntactically valid Go.
2. **Stage 2 – go/parser + Plugin Pipeline:** go/parser builds an AST on the rewritten code; plugins perform structural transformations, type inference, and code generation.
3. **Source Mapping:** Tracks offsets between `.dingo` and generated `.go` to keep diagnostics and LSP features coherent.

This arrangement optimizes for reusing Go tooling but struggles when a feature requires context spanning multiple regex processors (e.g., nested pattern matching or syntax that is not easily expressed as a textual substitution).

## Goals for an Optional Parser Layer
- Provide richer structural context (token boundaries, nested constructs) without abandoning the go/parser anchor.
- Remain optional so early-stage features that are well served by regex continue running without extra cost.
- Allow gradually migrating certain preprocessors to parser-backed transforms when complexity warrants.
- Avoid fragmenting error reporting or source maps.

## Candidate Parser Technologies

| Option | Pros | Cons | Fit for Dingo |
| --- | --- | --- | --- |
| **Participle (alecthomas/participle)** | Idiomatic Go, declarative grammar via struct tags, integrates well with Go AST types, good error messages | Grammar maintenance overhead, limited incremental parsing, moderate performance hit | Strong: Go-native, easy to embed as optional layer, can target just the Dingo-specific surface syntax before lowering to Go |
| **Tree-sitter** | Fast incremental parsing, rich tooling ecosystem, reuse for editor features, broad grammar support | Requires CGO bindings, heavier dependency, grammar DSL not Go-centric, more complex build | Medium: attractive for future IDE integration but adds build complexity and diverges from "pure Go" toolchain |
| **Custom Recursive Descent** | Full control, tailor-made error messages, no external deps | High implementation cost, maintenance burden, risk of divergence from Go grammar, re-implements known wheel | Weak: contradicts "reuse over reinvention" and is expensive |
| **PEG-based (e.g., pigeon, gocc)** | Formal grammar with backtracking, straightforward definition | Potential performance issues, debugging grammars harder, risk of exponential behavior without care | Medium-Low: manageable but still introduces custom grammar maintenance and tooling |

**Recommended choice:** Start with **Participle** for the optional layer due to alignment with Go ecosystem, maintainability, and ability to parse only the syntactic deltas relative to Go.

## Integration Patterns
1. **Plugin-invoked parser (Recommended):**
   - Treat the parser layer as a service invoked by preprocessors that opt in.
   - Each advanced feature registers a handler that receives the Participle AST and outputs transformed Go text or structured hints for the existing preprocessors.
   - Keeps the regex pipeline untouched for legacy features while offering a richer path for new ones.

2. **Feature-gated pipeline:**
   - Toggle-driven: `dingo build --enable-structured-preprocessor`
   - Allows incremental rollout and A/B comparison in CI.

3. **Lazy parsing:**
   - Parse only when a processor requests context that cannot be resolved via regex.
   - Cache parsed subtrees keyed by source hash to avoid repeated work within the same compilation.

## Proposed Architecture
```
.dingo
  ↓ (existing preprocessors)
Interim Go-like text + metadata (tokens, offsets)
  ↓ (optional structured layer)
Participle AST (Dingo surface syntax)
  ↓
Context-aware transformers emit refined Go text / annotations
  ↓
go/parser
  ↓
AST plugins (unchanged)
```

### Components
- **`pkg/preprocessor/parser`** (new package): encapsulates the Participle grammar for Dingo-specific constructs.
- **`pkg/preprocessor/passes`**: each pass can declare `RequiresStructure bool` to receive parsed nodes.
- **`pkg/preprocessor/context` enhancements:** add `StructuredAST interface{}` accessor with caching.
- **Diagnostics integration:** convert Participle errors to existing error reporter, preserving source map offsets.

## Rollout Plan
1. **Phase 0 – Spike (1 week):**
   - Build minimal Participle grammar covering enums + pattern match keywords.
   - Benchmark parse overhead on existing golden suite (target <15% preprocessor time increase when enabled).

2. **Phase 1 – Opt-in for Pattern Matching (2-3 weeks):**
   - Move pattern matching preprocessor logic to structured layer.
   - Add CLI flag `--structured-preprocessor`. Disabled by default but exercised in CI on new pattern tests.

3. **Phase 2 – Broader Adoption (3-4 weeks):**
   - Migrate the most brittle regex processors (enum payloads, nested `let`/`match`).
   - Document extension guidelines in `ai-docs/architecture/structured-preprocessor-architecture.md`.

4. **Phase 3 – Default On + Telemetry (2 weeks):**
   - Enable by default once stability proven; keep flag to disable.
   - Capture metrics (parse failures, performance) to validate cost/benefit.

## Cost / Benefit Estimate
- **Engineering Effort:** ~6-8 weeks elapsed (parallelizable) for full rollout.
- **Performance Impact:** Expected +10-20ms per file for parsing when enabled, offset by reduced downstream complexity and fewer recompilations due to malformed regex transformations.
- **Maintainability:** Significant improvement for complex features; reduces risk of regex drift and improves debuggability.
- **Developer Experience:** Better error messages (structured context), simplified addition of syntax-sugar features, clearer source mapping for nested constructs.

## Risks & Mitigations
- **Parser Drift vs Go Syntax:** Restrict grammar to Dingo-only constructs and delegate to go/parser for core Go syntax. Mitigation: embed Go tokens as literals and treat unknown constructs as pass-through.
- **Increased Build Time:** Use lazy parsing + caching; keep feature optional until performance validated.
- **Complexity Overhead:** Provide clear extension docs and code ownership for parser package; enforce linting/tests on grammar changes.
- **Dependency Footprint:** Vendor Participle or pin version; add integration tests to detect upstream changes.

## Relevant Precedents
- **TypeScript Compiler:** Uses full AST pipeline with multiple lowering phases; demonstrates benefits of structured transformations before emitting JS.
- **Borgo Transpiler:** Initially regex-based but moved to grammar-driven parsing for sum types to avoid ambiguity.
- **templ (Go HTML templates):** Maintains a small custom parser for templates before emitting Go AST, proving the viability of limited-scope parsers feeding go/parser.

## Recommendation
Adopt an optional structured parser layer using Participle, integrated as a plugin-invoked service in the preprocessor stage. Roll out behind a feature flag focused on high-complexity features (pattern matching first). This approach balances the need for richer syntax handling with the project mandate to reuse Go tooling and avoid full parser rewrites, while providing a path to gradually retire brittle regex passes.

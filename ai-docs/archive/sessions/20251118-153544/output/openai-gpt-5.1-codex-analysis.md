# Parser Architecture Investigation (2025-11-18)

## Executive Summary
Dingo must preserve full Go interoperability while introducing syntax sugar such as pattern matching, sum types, and `?` error propagation. Three viable parser directions emerged: (1) continue leaning on the standard `go/parser` with targeted extensions, (2) switch to or augment with third-party/foreign parsers (tree-sitter, participle, pigeon, etc.), and (3) adopt a hybrid pipeline where the preprocessor normalizes lightweight sugar and an augmented parser handles semantic features. Given the need for zero runtime overhead, tight gopls integration, and timely Go 1.23+ support, a hybrid approach that retains the official Go front-end for AST fidelity while introducing a thin, pluggable Dingo-aware scanning layer is the most balanced solution.

## 1. Extending Go's Native Front-End

### 1.1 Feasibility
- **Hooking `go/scanner`**: The scanner already treats identifiers and literals generically; limited token rewrites (e.g., mapping `match` → `switch`) are feasible via a fork or by running a lightweight lexical pre-pass before feeding source into the scanner. Projects like `garble` and `gofmt` have demonstrated maintaining downstream compatibility while operating on Go tokens.
- **AST post-processing**: Dingo already leverages AST transformations. Continuing with this model ensures compatibility with `go/types`, `go/printer`, and gopls, preserving comments and formatting metadata crucial for source maps.
- **Parser forking costs**: Maintaining a hard fork of `go/parser` introduces Go release synchronization overhead (2–4 weeks per Go release) and requires re-applying local patches for every new grammar addition (e.g., Go 1.22 range-over functions, 1.23 iterator proposals). However, the Go parser is fewer than ~12k lines, and diff churn per release is modest, making a shallow fork manageable if changes stay scoped.

### 1.2 Example strategies
1. **Token filtering adapter**: Wrap `io.Reader` with a transformer that rewrites Dingo syntax (e.g., `let`, `enum`, `match`) into provisional Go-friendly tokens while emitting mapping metadata for the LSP.
2. **Parser patching**: Introduce new productions (e.g., `MatchStmt`) guarded by feature flags, then desugar to regular AST nodes after parsing. This mirrors how TypeScript keeps separate AST nodes (e.g., `ForOfStatement`) and lowers them later.
3. **Error-resilient extensions**: Use `scanner.ErrorHandler` hooks to provide Dingo-specific diagnostics without forking error formatting logic.

### 1.3 Pros / Cons
- **Pros**: Maximum compatibility with tooling, reuse of `go/types`, fewer surprises for maintainers, straightforward source-map generation because positions originate from Go's own token stream.
- **Cons**: Limited flexibility for introducing fundamentally new grammar (e.g., pattern matching with guards) without more invasive parser edits; Go parser intentionally rejects ambiguous constructs, so Dingo must normalize syntax before parsing or accept a fork burden.

## 2. Third-Party Parser Survey

| Parser | Language Support | Maintenance & Go version tracking | Extensibility | Notes |
|--------|-----------------|-----------------------------------|---------------|-------|
| **tree-sitter-go** (GitHub maintained) | Incremental parsing via C grammar | Actively tracks Go releases (1.22 grammar merged Q3 2024); bindings available via `go-tree-sitter` | Adding Dingo grammar requires forking grammar.js; integrating with Go toolchain needs separate AST translation | Great for editor features and error recovery but produces tree-sitter nodes, not Go AST; bridging to `go/types` requires custom conversion |
| **go-tree-sitter** (bindings) | Go wrapper around tree-sitter C API | Dependent on core tree-sitter releases, typically lag <1 month | Allows embedding in Go; still requires writing Dingo grammar | Suitable for LSP or incremental parsing, but not drop-in for Go compiler pipeline |
| **participle** (github.com/alecthomas/participle) | Parser generator using struct tags | Maintainer active but not tied to Go grammar; would require re-implementing Go syntax entirely | High flexibility; grammar defined in Go | Re-implementing full Go grammar would be multi-month effort and risks divergence from upstream |
| **pigeon / gocc / ANTLR** | PEG / LL parsers | Require manual grammar updates per Go release | High flexibility | Provide fine control but effectively mean maintaining an independent Go parser |
| **rust-based parsers (tree-sitter via WASM)** | Potential cross-language reuse | Additional build/toolchain complexity | Good incremental story | Integration overhead with Go build pipeline and performance overhead for cross-language FFI |

**Assessment**: Third-party parsers excel for incremental parsing and editor tooling but complicate end-to-end Go compilation because they produce non-Go ASTs. Mapping tree-sitter trees back to `go/ast` with precise position info is non-trivial and jeopardizes zero-overhead goals. They do, however, make strong companions for the LSP side, suggesting a split architecture (tree-sitter for IDE responsiveness, Go parser for compilation).

## 3. Meta-Language Precedents

- **TypeScript**: Implements a full custom parser to support superset features (decorators, enums, JSX). Maintains separate AST nodes and down-level emitters. Lesson: owning the parser affords rapid feature development but requires a large dedicated team to keep pace with ECMAScript evolution.
- **Kotlin / Scala.js**: Built entirely custom compilers; they mirror target JVM/JS semantics but accept higher maintenance. Both invest heavily in incremental compilation and IR pipelines—resource intensive.
- **Borgo** (Go superset) and **CoffeeScript**: Use preprocessing plus code generation rather than re-implementing Go/JS parsers. Borgo leverages Go tooling downstream after textual rewrites, similar to Dingo's current pipeline, proving the viability of preprocessor-first strategies.
- **Elm**: Owns parser and compiler; slower language evolution but complete semantic control.

**Key takeaways**: Superset languages that must stay lockstep with a fast-evolving host language either (a) commit to a large parser team (TypeScript) or (b) rely on host compilers with preprocessing (CoffeeScript, Borgo). Given Dingo's small team and the need to track Go releases yearly, the second model is more sustainable.

## 4. Hybrid Preprocessor + Parser Splits

A pragmatic approach keeps syntactic sugar that is trivially expressible in Go (type annotations, `let`, `enum` wrappers) in the preprocessor while routing constructs requiring semantic awareness (pattern matching exhaustiveness, `Result` inference) through AST transforms.

### Benefits
- **Clear responsibility**: Text-level transformations handle lexical sugar; AST stage handles semantic rewrites and validations.
- **Reduced churn**: When Go adds syntax (e.g., range-over func, iterators), only the Go parser fork/alignment layer needs updates, not custom grammars.
- **Source map fidelity**: Preprocessor records token mappings; AST stage operates on Go AST with position offsets already tracked, enabling precise IDE support.

### Risks
- **Two-step debugging complexity**: Developers must reason about both preprocessed Go and AST manipulations. Mitigated by tooling that shows side-by-side views (already part of Dingo golden tests and showcase strategy).
- **Edge cases**: Certain features (e.g., pattern guards with destructuring) may be awkward to express purely via textual rewrites; requires careful design to ensure the Go parser still receives valid input.

## 5. Recommendation & Comparative Analysis

| Approach | Complexity | Maintainability | Correctness & Tooling | Go Version Tracking | Notes |
|----------|------------|-----------------|-----------------------|---------------------|-------|
| **Pure preprocessor + stock Go parser** | Low | High (minimal surface area) | High for Go-compatible constructs, limited for new grammar | Excellent (no fork) | Works today, but pattern matching or multi-line `match` blocks may become cumbersome |
| **Forked Go parser (`go/parser` + custom scanner hooks)** | Medium | Medium (need quarterly sync) | High—still produces `go/ast`, enabling go/types + gopls | Good (manual merges each release) | Enables true new grammar (e.g., `match`, `enum` declarations) without reimplementing entire compiler |
| **Third-party parser replacement (tree-sitter / participle)** | High | Low-Medium (own grammar + translation layer) | Medium—needs AST conversion, risks drift | Variable (depends on upstream) | Provides better error recovery but disconnects from Go toolchain |
| **Full custom parser (TypeScript model)** | Very High | Low without large team | Potentially highest flexibility but huge investment | Poor unless team mirrors Go releases rapidly | Likely infeasible within 12–15 month runway |
| **Hybrid (current preprocessor + thin parser fork + AST plugins)** | Medium | Medium-High | High—leverages Go AST, allows richer syntax, manageable scope | Good (small diff per release) | Recommended balance; reuse existing Stage 1, add Stage 1.5 lexical adapter, keep Stage 2 AST pipeline |

### Recommended Architecture (Hybrid)
1. **Stage 0.5 Scanner Adapter**: Intercept tokens to recognize Dingo-specific keywords/operators before they hit the stock parser, emitting placeholder Go tokens plus mapping metadata.
2. **Stage 1 Text Preprocessor**: Continue handling simple rewrites (type annotations, `let`, enum scaffolding) where textual substitution suffices.
3. **Stage 2 Parser (Go fork)**: Minimal fork that accepts additional productions (`MatchExpr`, `WhenGuard`, etc.) but immediately rewrites them into standard AST nodes so downstream tooling stays unchanged.
4. **Stage 3 AST Plugin Pipeline**: Existing discovery/transform/inject steps apply semantic rewrites and generate Result/Option helpers.
5. **Source Map + LSP**: Because tokens still flow through Go's machinery, position info remains accurate; tree-sitter can optionally power the IDE for incremental parsing without affecting compilation.

### Key Justifications
- **Complexity vs. payoff**: The hybrid plan avoids re-implementing Go semantics while unlocking richer syntax than pure preprocessing allows.
- **Maintainability**: A shallow fork (or even patch-set) of `go/parser` historically requires <1k LOC changes per feature; merging upstream each Go release is tractable.
- **Correctness**: Staying on `go/ast` ensures go/types inference, gopls proxying, and zero runtime overhead remain intact.
- **Go version tracking**: Align release cadence with Go's biannual schedule; dedicate 1–2 engineer-weeks per release for parser sync + regression tests.

### Biggest Risks & Mitigations
- **Risk**: Parser fork diverges and becomes hard to merge. *Mitigation*: Keep modifications modular (new files or clearly tagged diffs), add automated `go tool dist list` tests for each Go beta, and document merge steps in `ai-docs/parser-maintenance.md`.
- **Risk**: Source maps drift due to multi-stage rewrites. *Mitigation*: Persist token mapping tables from preprocessor through AST transforms and add golden tests comparing Dingo vs Go positions (similar to `showcase_01` guardrails).
- **Risk**: Limited resources for large grammar work. *Mitigation*: Prioritize features that justify parser edits (pattern matching) and leave cosmetic sugar in Stage 1 to minimize scope.

## Conclusion
Adopting a hybrid pipeline—textual preprocessing for syntactic sugar, a lightly customized Go parser for genuinely new constructs, and the existing AST plugin system—best satisfies Dingo's constraints. Third-party parsers remain valuable for IDE experiments but introduce unnecessary complexity for the compiler path. Investing now in tooling to automate Go release merges and in robust source-map preservation will de-risk the approach ahead of the 12–15 month v1.0 target.

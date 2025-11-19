# Parser Strategy Options for Replacing Dingo's Regex Preprocessor

## Executive Summary
- **Primary Recommendation:** Build a hybrid pipeline that keeps the light-weight regex preprocessor for token-level sugar but introduces an AST-aware pass built on `go/parser` augmented with custom AST annotations and recovery hooks. This balances near-term delivery (3-4 months) with long-term maintainability while staying within the 12-15 month roadmap.
- **Key Finding:** Every successful meta-language (TypeScript, Kotlin, Scala.js, Borgo, CoffeeScript, Elm) eventually converged on a native or native-compatible parser to track their host language—regex/token hacks could not keep up with syntax churn.
- **Critical Risk:** Third-party parsers that do **not** originate from the Go team (tree-sitter-go, participle grammars) consistently lag behind new Go releases by 1-3 versions, threatening Dingo's commitment to zero-overhead, idiomatic Go output if they diverge.

## Approach Comparison Matrix
| # | Strategy | Parser Tech | Pros | Cons | Impl. Complexity (1-10) | Maintainability | Go Version Tracking |
|---|----------|-------------|------|------|-------------------------|-----------------|---------------------|
| 1 | Enhanced native pipeline | `go/parser` + custom visitors | Always current with Go, no extra runtime, simple tooling integration | Requires creative handling of Dingo-only syntax pre-parse, limited freedom to introduce new tokens | **5** | High (stdlib support, well-documented) | Excellent (ships with Go)
| 2 | Tree-sitter front-end | `tree-sitter-go`, `go-tree-sitter` bindings | Full parse tree, incremental parsing, fast error recovery | Grammar lags Go, requires C bindings, needs transpilation back to Go AST, adds build complexity | **7** | Medium (community maintained, but binding updates needed) | Medium-Low (1-2 releases behind)
| 3 | Go-native parser generator | `participle` custom grammar | Full control over grammar, pure Go implementation, no CGO | Must re-specify entire Go grammar + Dingo features, high upkeep, risk of semantic drift | **8** | Low-Medium (team-maintained grammar) | Low (manual updates per Go release)
| 4 | Third-party maintained parsers | `pigeon`, `goyacc` forks, other OSS grammars | Potentially ready-made grammars to fork, some AST conveniences | Most are abandoned/out-of-date, license risk, unclear roadmap | **6** | Low (bus-factor, inconsistent quality) | Low (must backport Go changes)
| 5 | Hybrid split (regex + AST) | Current preprocessors + `go/parser` AST hooks | Incremental adoption, keeps working features, allows gradual migration to structured transformations | Continues to carry regex debt; needs rigorous source map alignment between stages | **4** | Medium-High (reuse existing code) | High (relies on stdlib parser for Go compliance)
| 6 | External meta-language style transpiler | Borrow TypeScript/Kotlin architectural patterns | Proven blueprints for feature staging, source maps, plugin pipelines | Requires investment in tooling (watch mode, compiler services), longer build-up | **6** | High (battle-tested patterns) | High (host languages use native compilers)

## Detailed Analysis

### 1. Enhanced `go/parser` Pipeline
- **Concept:** Keep a minimal textual shim (e.g., replace `let` or `match` keywords with unique tokens) then run `go/parser` on the transformed file. Introduce a discovery pass that records placeholders (comments, pragmas) representing Dingo constructs for later AST rewriting.
- **Pros:**
  - Inherits Go's syntax evolution automatically (1:1 parity every release).
  - Maintains zero runtime overhead; the generated Go AST is truly native.
  - Fits existing AST plugin pipeline; only requires richer metadata during preprocessing.
- **Cons:**
  - Dingo-only syntax (pattern matching, enums) still needs pre-parse markers; complex constructs may become unreadable placeholders.
  - `go/parser` lacks hooks for new tokens; any fundamentally new syntax must be encoded as valid Go before parsing.
- **Maintainability:** High—leverages standard library documentation and community knowledge.
- **Correctness:** Strong; AST rewrites happen on canonical Go nodes, enabling go/types integration.
- **Go Version Tracking:** Automatic; update Go toolchain and inherit new grammar for free.

### 2. Tree-sitter Front-End (via `tree-sitter-go` + `go-tree-sitter`)
- **Concept:** Replace regex preprocessing with a Tree-sitter parse of Dingo syntax, then lower to Go AST or directly emit Go source.
- **Pros:**
  - Incremental parsing + error recovery ideal for IDE/LSP integration.
  - Grammar is declarative; adding Dingo constructs is straightforward.
  - Rich concrete syntax tree enables precise source maps.
- **Cons:**
  - Grammar typically lags upstream Go; contributions required to stay current.
  - CGO dependency complicates distribution; pure Go builds become harder.
  - Need bespoke conversion layer from Tree-sitter nodes to Go AST for go/types.
- **Maintainability:** Medium—active community but Dingo must contribute to keep parity.
- **Correctness:** Good for surface syntax; semantics still require custom passes.
- **Go Version Tracking:** Medium-Low; new Go releases demand immediate grammar updates plus rebuild of WASM/native artifacts.

### 3. Custom Grammar via `participle`
- **Concept:** Define a superset grammar (Go + Dingo) in pure Go using `participle`, producing typed ASTs for downstream transformations.
- **Pros:**
  - Full control over syntax; adding pattern matching or enums is first-class.
  - Pure Go, no CGO, easier cross-platform builds.
  - Parser combinators allow readable grammar definitions.
- **Cons:**
  - Must replicate entire Go spec (229 pages) plus maintain it; extremely high ongoing cost.
  - Hard to ensure 100% compatibility with go/types; subtle parsing differences break tooling parity.
  - Performance may lag compared to optimized `go/parser`.
- **Maintainability:** Low-Medium—depends on Dingo team resourcing a "Go grammar" owner.
- **Correctness:** Risky; any mis-parse leads to invalid Go output or misaligned source maps.
- **Go Version Tracking:** Low; each Go release requires manual grammar diffs and regressions.

### 4. Other Third-Party Parsers (e.g., pigeon grammars, forked goyacc specs)
- **Concept:** Adopt an existing community parser and extend it with Dingo syntax.
- **Pros:**
  - Jump-start from existing work, potentially shorter bootstrap.
  - Some projects (e.g., Borgo) share code for Go-like grammars.
- **Cons:**
  - Most grammars are incomplete or unmaintained; legal status may be unclear.
  - Divergent AST models complicate interoperability with go/types.
  - Bus-factor risk if primary maintainer leaves.
- **Maintainability:** Low—Dingo becomes de facto maintainer.
- **Correctness:** Suspect; limited testing coverage vs. `go/parser`.
- **Go Version Tracking:** Low; manual merges needed every release.

### 5. Hybrid Preprocessor + AST Augmentation (Incremental Path)
- **Concept:** Continue using regex preprocessors for syntax sugar that can be trivially mapped to valid Go (type annotations, enums placeholders), but enrich them with structured metadata (pragma blocks, synthetic comments). After `go/parser`, use AST passes keyed off this metadata to perform the heavy rewrites.
- **Pros:**
  - Minimal disruption to existing pipeline; can prioritize high-value rewrites first.
  - Source maps easier to maintain because final AST is still native Go.
  - Allows per-feature migration (e.g., pattern matching) without blocking other work.
- **Cons:**
  - Regex stage remains brittle; complex constructs risk exponential rule growth.
  - Requires rigorous testing to ensure metadata survives formatting and tooling.
- **Maintainability:** Medium-High—balanced investment while reducing regex scope.
- **Correctness:** High once metadata is reliably inserted; AST stage handles semantics.
- **Go Version Tracking:** High; depends on stdlib parser.

### 6. Meta-Language Precedents
| Language | Strategy | Takeaways for Dingo |
|----------|----------|---------------------|
| TypeScript | Full custom parser + checker, but tracks ECMAScript proactively; uses syntax sugar lowering with source maps | Invest early in owning the grammar; rely on community usage to validate feature rollout. Source maps + AST pipeline critical. |
| Kotlin (JVM) | Custom parser layered over IntelliJ infrastructure; outputs JVM bytecode via IR | Tight integration with host tooling ensures IDE parity; parser ownership enabled rapid addition of coroutines, etc. |
| Scala.js | Reuses Scala compiler front-end, swaps backend | Demonstrates value of reusing official front-ends when possible—mirrors Dingo leveraging `go/parser`. |
| Borgo | Custom parser in Go for Rust-like syntax | Shows feasibility but also highlights maintenance overhead; Borgo often trails Go releases. |
| CoffeeScript | PEG-based parser lowering to JS | Early success but eventually ceded to TypeScript because keeping up with JS spec became overwhelming; cautionary tale for regex-heavy pipelines. |
| Elm | Full compiler front-end; limited host interop | Clean architecture but slower to track JS/TS ecosystem; emphasizes trade-off between purity and compatibility. |

**Insights:**
1. Projects that reuse the host language's official parser (Scala.js, Kotlin via IntelliJ PSI) enjoy easier version tracking.
2. Projects that deviated (CoffeeScript, Borgo) suffered maintainability pain and lagged behind host feature releases.
3. Source maps + AST transformations are universally required once features exceed simple syntax sugar.

## Architecture Recommendations Under Dingo Constraints
1. **Short Term (0-3 months):**
   - Harden current preprocessors but confine them to syntactic sugar that can be expressed as valid Go placeholders.
   - Introduce a metadata channel (pragma comments) to mark constructs for later AST rewriting.
   - Begin measuring regex coverage to identify high-risk rules for replacement.
2. **Mid Term (3-9 months):**
   - Build structured AST passes (discovery → transform → inject) that read metadata and manipulate native Go AST nodes.
   - Experiment with `go/parser` fork only for localized token tweaks (e.g., treat `match` as identifier) if necessary, minimizing divergence.
   - Prototype Tree-sitter-based tooling solely for editor services (syntax highlighting), not as the canonical compiler front-end.
3. **Long Term (9-15 months):**
   - Gradually retire regex rules as features gain dedicated AST logic.
   - Evaluate whether a thin, well-tested lexical prepass can replace the regex stage entirely (e.g., simple tokenizer rewriting keywords before feeding `go/parser`).
   - Maintain a compatibility dashboard comparing Dingo outputs against upstream Go releases (auto-run gofmt/go vet/go test to ensure parity).

## Key Risks & Mitigations
| Risk | Impact | Mitigation |
|------|--------|------------|
| Tree-sitter grammar lag | Dingo cannot ingest newest Go syntax promptly | Use Tree-sitter only for IDE niceties; rely on `go/parser` for canonical builds. |
| Regex debt causing fragile rewrites | Incorrect transpilation, poor debugging UX | Enforce feature freeze on new regex rules unless paired with AST metadata tests; invest in property-based tests for the preprocessor. |
| Custom grammar drift (`participle`, others) | Divergent semantics, user mistrust | Avoid owning a full Go grammar unless headcount increases significantly. |
| Source map fidelity | Poor IDE mapping, debugging pain | Standardize metadata format (e.g., JSON comments) and build golden tests for sourcemaps before migrating features. |

## Overall Recommendation
- **Adopt the hybrid approach (Strategy #5) as the official migration path:** It immediately reduces regex fragility, leverages `go/parser` for correctness, and stays aligned with Dingo's zero-runtime, idiomatic Go promise.
- **Invest in AST metadata + plugin pipeline enhancements now** to unlock richer features (pattern matching, enums) without overhauling the entire compiler.
- **Use Tree-sitter selectively** (IDE/highlighting) while avoiding dependency for canonical compilation; continue monitoring ecosystem maturity.
- **Defer custom grammars** (participle or other third-party parsers) unless Dingo secures dedicated parser engineering resources, as they jeopardize Go version tracking and timeline commitments.

## Next Steps
1. Draft a migration plan that enumerates each regex processor, its owned syntax, and whether it can emit structured markers for AST passes.
2. Spike an "annotated preprocessor" prototype that inserts pragma blocks, then ensure `go/parser` preserves them for downstream consumption.
3. Define golden tests for source map fidelity covering the features most affected by upcoming pattern matching work.
4. Revisit Tree-sitter adoption after IDE requirements are formalized, ensuring it remains an auxiliary tool rather than core dependency.

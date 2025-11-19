# Optional Parser Layer Feasibility Study (2025-11-18)

## Executive Summary
- **Recommendation:** Do **not** add an intermediate syntax-tree parser in the near term; keep investing in the regex preprocessor + `go/parser` pipeline and revisit when Phase 5 features demand richer syntactic guarantees.
- **Rationale:** A new parser requires defining and maintaining a full Dingo grammar, duplicate traversal logic, and IDE integration, adding ≈9–12 person-weeks and ~30–45 ms per 1 kLOC compile overhead while solving only a subset of currently known issues.
- **Value threshold:** The parser layer becomes compelling only if upcoming features (dependent pattern matching, macros, hygienic rewrites) outpace what regex normalization plus AST plugins can express; current roadmap items can be achieved by incremental preprocessors and targeted AST plugins.

## Current Pipeline Constraints
1. **Regex Preprocessor:** Efficient and simple, but brittle around nested constructs, multi-line contexts, and overlapping tokens (e.g., `match`, `enum`, `?`).
2. **`go/parser`:** Provides rich AST + `go/types` integration, but only after the source is valid Go; any Dingo-only constructs must be erased or replaced beforehand.
3. **Pain Points:**
   - Hard to emit precise diagnostics tied to original Dingo constructs before Go parsing.
   - Some transformations (pattern matching exhaustiveness, contextual `None`) would benefit from richer syntactic information than regex offers.
   - Maintenance cost grows with each new regex processor as rewrite order matters.

## Objectives for an Intermediate Parser
- Provide a structured representation of the "normalized Dingo" syntax for semantic validations before the Go toolchain sees the file.
- Offer stable hooks for new language features without proliferating regex passes.
- Enable better source mapping (node-level) for the LSP and error reporting.
- Remain optional so existing fast path (regex → `go/parser`) still works.

## Technology Evaluation
| Option | Description | Pros | Cons |
| --- | --- | --- | --- |
| **Participle (PEG for Go)** | Define a PEG grammar directly in Go structs. | Native Go, readable grammar, supports partial parsing. | Backtracking can be costly; grammar maintenance overhead; limited tooling for incremental parsing. |
| **`goyacc`/`gocc` (LR parser)** | Generate LR parser from grammar. | Deterministic performance, mature toolchain, no runtime deps. | Grammars verbose, harder to evolve; poor error recovery; team lacks current LR expertise. |
| **Tree-sitter (embeddable GLR)** | Use Tree-sitter parser + query system. | Great error recovery, incremental parsing (IDE benefits), existing Go bindings. | Requires CGO or WASM, increases binary size, licensing + build complexity, double maintenance (grammar in C DSL). |
| **ANTLR (Go target)** | Generate parser from ANTLR grammar. | Rich tooling, visitor pattern. | Heavy runtime, external Java dependency for generation, slower startup. |
| **Custom recursive-descent** | Hand-written parser. | Maximum control, tailor-made error messages. | Highest implementation cost, tricky to keep spec-consistent. |

**Most feasible choice if pursued:** Participle. It aligns with Go codebase norms, avoids CGO, and keeps grammars co-located with semantic code. Tree-sitter is attractive for future LSP work but introduces significant toolchain friction.

## Integration Approach (If Adopted)
```
.dingo → regex preprocessors → (Optional) Dingo AST Parser → AST validation/rewrites → regenerate normalized Go source → go/parser → plugin pipeline
```

1. **Parser Input:** Feed the same text that would otherwise go directly into `go/parser` (post-regex). Grammar would describe the normalized-but-still-Dingo constructs (e.g., `match`, enum sugar) before they are rewritten into pure Go.
2. **AST Nodes:** Mirror only the constructs that currently stress regexes (pattern match arms, enum definitions, `?` operators). Avoid duplicating the entire Go grammar.
3. **Transformation Hooks:** Replace regex replacements that need structural awareness with parser visitors that emit Go-friendly equivalents or annotations for downstream plugins.
4. **Fallback:** Keep current pipeline as default; enable parser with CLI flag (`--parser=experimental`) until parity proven.
5. **Tooling:** Reuse existing source-map infrastructure by attaching parser node spans before Go parsing rewrites them.

## Roadmap & Effort Estimate (If Proceeding)
1. **Grammar Spike (1.5 weeks):** Prototype Participle grammar for enums + pattern matching subset; benchmark against existing preprocessors.
2. **AST Visitor Layer (1 week):** Build visitors that emit the same Go text transformations currently produced via regex; validate golden tests.
3. **Diagnostics & Source Maps (2 weeks):** Integrate parser spans with error reporting + LSP mapping.
4. **Feature Migration (3–4 weeks):** Move complex preprocessors (enum, pattern-match, error propagation) into parser visitors incrementally.
5. **Performance + Stability Hardening (1.5–2 weeks):** Optimize grammar, add fuzzing, document failure modes.

**Total:** ~9–12 person-weeks before feature parity, excluding ongoing maintenance.

## Cost-Benefit Analysis
- **Performance:** Double parsing (parser + `go/parser`) adds ~30–45 ms per 1 kLOC (based on Participle benchmarks at 600–800 kTokens/s vs. `go/parser` ~1.5 M tokens/s). For a 10 kLOC module, expect 0.3–0.45 s slowdown (20–30% longer compile time) unless preprocessors can be removed entirely.
- **Maintenance:** Grammar + visitor code adds a new surface area requiring versioning, review, and dedicated expertise; expect ~0.5 person-week per quarter just to keep parity with evolving syntax.
- **Value:** Gains primarily in developer ergonomics (better errors, fewer regex bugs) and future extensibility. Current roadmap items (pattern matching, None inference, improved diagnostics) can be delivered with targeted regex + AST enhancements for 3–4 person-weeks total.
- **Opportunity Cost:** Those 9–12 person-weeks could finish pattern matching, polish `dingo-lsp`, and harden go/types inference, all of which ship user-visible value sooner.

## Precedents
- **TypeScript:** Full parser mandatory because JS syntax itself needs re-interpretation; however, TS also removes regex preprocessors entirely—Dingo would still rely on Go AST, making a mid-layer less impactful.
- **Borgo Lang:** Uses PEG parser because it targets Go but with Rust-like syntax; however, Borgo does not leverage `go/parser`, so the parser is core rather than optional.
- **templ:** Stays on text rewrites + Go parser, relying on structured comments/source maps; demonstrates that high-fidelity Go tooling is achievable without a bespoke parser.
- **Svelte / Babel:** Maintain complete ASTs because target languages (JS) lack typed AST transformations; they highlight the maintenance burden (dedicated parser teams).

## Risks & Mitigations
| Risk | Impact | Mitigation |
| --- | --- | --- |
| **Grammar Drift** | Divergence between parser output and regex/go parser expectations causes subtle bugs. | Keep parser optional; drive all transformations from a single source (parser) once stable; add golden tests for every grammar rule. |
| **Performance Regression** | Longer build times reduce developer feedback loop. | Limit parser to files needing advanced features; cache parse trees between incremental builds. |
| **Team Bandwidth** | Parser expertise becomes bottleneck. | Formalize ownership, add training, document grammar thoroughly. |
| **Tooling Complexity** | IDE integration must understand new AST. | Gate feature behind flag until LSP pipeline consumes new spans; reuse existing source-map format. |

## Final Recommendation
- **Short-term (Phase 4–5):** Continue refining regex preprocessors and AST plugins. Invest in better preprocessing scaffolding (ordered passes, shared tokenizer helpers) and improved diagnostics.
- **Medium-term Trigger:** Re-evaluate parser layer only if (a) regex maintenance time exceeds 20% of engineering bandwidth or (b) upcoming features require syntactic constructs impossible to express via text rewrites without exponential complexity.
- **Decision:** Defer implementation; document parser research (this file) and schedule a checkpoint after pattern matching + None inference ship to reassess actual pain. The optional parser layer is feasible but not cost-effective today.

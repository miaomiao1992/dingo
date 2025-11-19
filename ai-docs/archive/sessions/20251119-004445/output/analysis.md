# Option B vs Option C — Injected Type Comment Pollution

## 1. Comparison Table
| Dimension | Option B — Separate AST for injected types | Option C — Comment map cleanup pass | Verdict |
| --- | --- | --- | --- |
| Implementation complexity | Requires new go/ast.File builder, printer integration, and wiring in pkg/plugin injection pipeline; moderate upfront work but conceptually clean. | Adds traversal over existing ast.CommentMap with heuristics to identify synthetic comment groups; simpler code but higher tuning cost. | Slight edge to Option C for initial effort, but B offers clearer structure once built. |
| Accuracy of DINGO comment removal | Guarantees zero pollution because generated declarations never share comment space with user AST; no heuristics. | Depends on pattern matching (prefix markers, positions), risk of leaving stray markers or deleting real comments. | Option B. |
| Impact on source maps & diagnostics | Clean separation simplifies source map math: injected file can opt out of mapping entirely, preventing bogus LSP jumps. | Must maintain mappings while deleting comments; risk of mismatched ranges leading to incorrect diagnostics. | Option B. |
| Performance / memory | Minor overhead: second AST/file but only for small library snippets; negligible runtime impact. | Extra pass over comment map every build; O(N comments) but still cheap; slightly less memory than B. | Near tie (Option C marginally lighter but difference insignificant). |
| Maintainability / future features | Scales when adding more built-in declarations (Result, Option, iterators) because they live in dedicated generator; easier to unit test. | Each new injected construct requires updating detection heuristics; brittle as syntactic forms evolve. | Option B. |
| Testing effort | Needs golden coverage to ensure separate file merges correctly plus new unit tests for builder; mostly deterministic. | Must craft adversarial tests to ensure no real comments removed; relies on manual review. | Option B (deterministic). |
| Timeline risk | More initial plumbing touching pkg/plugin, pkg/generator and go/printer usage; but changes localized. | Faster to ship but higher chance of regressions that would stall release due to subtle comment issues. | Depends on risk tolerance; B safer, C faster. |

## 2. Deep Analysis
### Option B — Separate AST for injected types
- **Architecture fit**: Mirrors Go tooling expectations by treating injected Result/Option declarations as if they were authored in a standalone support file. pkg/plugin/builtin can emit a dedicated `ast.File` whose declarations are appended late in codegen, keeping user AST untouched. This respects the two-stage pipeline (preprocessor → parser) without introducing ad hoc comment surgery.
- **Source map behavior**: Because the auxiliary file is synthetic, we can opt out of generating `.sourcemap` entries for it or map everything to sentinel "<dingo-internal>" positions. LSP handlers (pkg/lsp/server.go) already gate diagnostics to known user ranges, so isolating injected code prevents accidental hover/completion jumps into synthetic regions.
- **Extensibility**: Future built-ins (iterators, async helpers) can reuse the same builder by passing templates and metadata; no need to touch the primary AST. Separation also encourages reuse in tests—`pkg/parser/new_features_test.go` can import the builder and assert on produced Go snippets without parsing full user programs.
- **Costs & touch points**: We must add: (1) a generator that produces the auxiliary AST, (2) wiring in pkg/generator/markers.go or downstream to merge files before formatting, and (3) awareness in pkg/preprocessor/source map writer so that line offsets remain correct when final Go bundle concatenates files. These are finite and well-scoped tasks, but require careful API design to avoid regressions.

### Option C — Comment map cleanup pass
- **Mechanics**: After parsing the combined file, traverse `ast.CommentMap` (via `ast.NewCommentMap`) and prune entries tagged with Dingo markers (e.g., `// DINGO-INJECTED`). This keeps a single AST and avoids generator changes. Implementation likely lives in pkg/parser/simple.go or a new helper consumed before go/printer writes output.
- **Heuristic fragility**: Because injected declarations interleave with user code, we must rely on markers, positions, or naming conventions to identify which comments are synthetic. Any divergence (user writing a comment that coincidentally matches the marker or future injected nodes lacking markers) can cause false positives/negatives. Additionally, removing a comment group can leave dangling whitespace or disrupt formatter alignment unless we also tweak `printer.Config`.
- **Cross-cutting effects**: Comment removal happens after parsing, so if diagnostics earlier in the pipeline relied on those comments for positioning (e.g., error propagation hints), we risk desync. Moreover, `pkg/preprocessor/sourcemap_test.go` demonstrates reliance on comment anchors for mapping; cleaning them later may necessitate re-running portions of the sourcemap builder or adjusting offsets.
- **Pros**: Minimal restructuring, faster time-to-fix, and no need to manage multiple AST files. Works if we only need to drop a handful of known comment groups and accept occasional drift.

## 3. Risk Assessment
| Risk | Option B Exposure | Option C Exposure | Mitigation |
| --- | --- | --- | --- |
| Incomplete removal leading to visible DINGO comments | None (comments never created). | High (heuristics may miss cases). | Improve markers, add regression tests—but inherently probabilistic. |
| Accidental deletion of user-authored comments | None. | Medium (pattern collisions). | Require strict prefixes + position bounds; still non-zero risk. |
| Increased maintenance burden when adding new built-ins | Low (just extend generator). | High (update heuristics for every new pattern). | Modularize detection utilities, but still reactive. |
| Integration complexity | Medium (needs new file merge path). | Low. | Invest in clear interfaces and incremental rollout. |
| Debuggability when synthetic code causes panic | Medium (stack traces refer to internal file). Requires doc updates. | Low (single file). | Use sentinel filenames and doc tooling so developers know stack frames are internal. |

## 4. Future Scenarios
1. **Expanding builtin catalog**: If Dingo ships more injected helpers (async runtimes, iter utilities), Option B scales linearly by adding templates; Option C scales combinatorially as each helper needs new comment filters.
2. **IDE deep linking**: Upcoming LSP enhancements (Phase 4.2) will rely heavily on precise source maps. Option B keeps synthetic code quarantined for LSP, while Option C risks future comment surgeries invalidating cached offsets.
3. **Third-party plugin ecosystem**: When external plugins begin injecting code, Option B can expose an API for them to register separate AST fragments, enforcing comment hygiene. Option C would need to understand third-party markers, which becomes untenable.
4. **Performance-sensitive builds**: In watch mode, Option C adds another AST walk per build. Option B amortizes its cost because the auxiliary AST can be cached until a config change occurs.

## 5. Recommendation
Adopt **Option B (separate AST for injected types)** as the primary path. Although it demands more up-front engineering, it fully eliminates comment pollution, simplifies source map math, and provides a scalable home for future built-in declarations. Option C can remain a fallback mitigation for any legacy artifacts but should not be the mainline solution.

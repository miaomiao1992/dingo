# Context-Aware Preprocessing Architecture Plan

## 1. Problem Summary
Dingo currently relies on a regex-driven preprocessing stage followed by `go/parser` + plugin transforms. Upcoming features such as pattern matching, nested sum types, and smarter `?` propagation require the preprocessor to make decisions that depend on lexical/semantic context (e.g., distinguishing value vs. type positions, honoring scope, mapping generated scaffolding back to source for LSP). We need a context-aware strategy that preserves the simplicity of regex processors while guaranteeing deterministic behavior, debuggability, and compatibility with existing go/ast plugins.

## 2. Candidate Strategies
| Option | Description | Pros | Cons |
| --- | --- | --- | --- |
| A. Marker-Guided Single Pass | Enrich regex preprocessors to emit lightweight markers (`/*@dingo:token*/`) and defer contextual interpretation to AST plugins. | Minimal changes, keeps preprocessors simple, leverages go/parser for validation. | Markers can explode if not normalized; AST plugins must strip noise; limited pre-parser feedback.
| B. Two-Pass Preprocessor (Lex Pass + Rewrite Pass) | Introduce a fast lexer to gather context (scope, keywords, generics) and feed metadata into second regex rewrite pass. | Precise contextual decisions before go/parser, can block invalid constructs earlier. | Higher complexity, effectively re-implements parts of parser, higher maintenance cost.
| C. Hybrid Marker + Sidecar Context (Recommended) | Augment preprocessors to output markers plus a synchronized sidecar map capturing context snapshots. AST plugins consume both marker hints and structured context via sourcemap-like offsets. | Preserves regex simplicity, provides rich context to downstream stages, keeps single parse tree, enables LSP/source map alignment, minimizes performance impact. | Requires new context manager, tooling to keep sidecar in sync, needs disciplined marker grammar.

**Recommendation: Option C (Hybrid Marker + Sidecar Context).** It balances implementation effort with the level of context we need, leverages existing infrastructure (sourcemaps, plugin pipeline), and remains extensible for future features (pattern matching, match guards, context-sensitive keywords).

## 3. Architecture & Data Flow
```
.dingo source
   │
   ▼
Regex Preprocessor (existing pipeline)
   ├─ Emits transformed Go-ish text with inline markers
   ├─ Populates Context Sidecar (JSON/CBOR) keyed by byte offsets
   │    • lexical scope, inferred construct, feature tag, binding ids
   │    • origin span ↔ surrogate span mapping
   └─ Writes Sourcemap V2 (merged with context metadata)
   │
   ▼
`go/parser` parses marker-enriched Go text → `*ast.File`
   │
   ▼
Plugin Pipeline (discovery/transform/inject)
   ├─ Context Loader attaches metadata to AST nodes
   ├─ Preprocessor Marker Sanitizer removes/rewrites scaffolding
   ├─ Feature-specific plugins (Result, Option, PatternMatch) consult context
   │
   ▼
Printer writes final `.go` + `.sourcemap` (V2)
```

### Components
1. **Marker Grammar**: `/*@dingo:<kind>:<payload>*/` comment blocks or zero-width unicode tags. Must be Go-comment safe, easy to strip. Examples: `/*@dingo:match:start*/`, `/*@dingo:binding:id=3*/`.
2. **Context Manager**: `pkg/preprocessor/context` (new) provides:
   - `Builder` API to register spans, scopes, feature annotations.
   - `Sidecar` serialization (JSON for readability; switch to CBOR later).
   - Merge utilities to combine with sourcemap.
3. **Context Loader Plugin**: Early plugin that maps AST positions to context entries and exposes `ContextFor(ast.Node)` queries to later plugins.
4. **Pattern Match Plugin (PoC)**: Uses markers to delimit pattern arms and context info to verify exhaustiveness, generate dispatcher code, and maintain sourcemap accuracy.

## 4. Concrete Code Examples
### Preprocessor Marker Emission
```go
// pkg/preprocessor/pattern_match.go
func emitMatchStart(buf *bytes.Buffer, span Span, bindingID int) {
    buf.WriteString("/*@dingo:match:start binding=")
    buf.WriteString(strconv.Itoa(bindingID))
    buf.WriteString(" span=")
    buf.WriteString(span.String())
    buf.WriteString("*/")
}
```

### Context Builder Usage
```go
ctx := contextbuilder.New(fileID)
ctx.PushScope("match", span)
ctx.AddFeature(span, contextbuilder.Feature{
    Kind: contextbuilder.FeaturePatternMatch,
    Data: map[string]any{
        "binding": bindingID,
        "enum": enumName,
    },
})
```

### AST Plugin Consumption
```go
func (p *PatternPlugin) Visit(node ast.Node) ast.Visitor {
    info := p.ctx.Lookup(node.Pos())
    if info.Feature == contextbuilder.FeaturePatternMatch {
        // rewrite node, emit dispatcher, ensure exhaustiveness
    }
    return p
}
```

## 5. Pattern Matching PoC Walkthrough
1. **Source**: `match result { Ok(v) => handleOk(v), Err(e) => handleErr(e) }`.
2. **Preprocessor**: Rewrites to Go `switch` skeleton with markers delimiting each arm and capturing enum variant metadata.
3. **Sidecar Context**: Records per-arm variant name, payload bindings, guard presence, and fallback status.
4. **AST Discovery**: Loader associates each `CaseClause` with its variant metadata via position mapping.
5. **Transformation**: Plugin validates exhaustiveness using enum definition (available via type info or metadata), expands payload destructuring, inserts guard checks, and removes markers.
6. **Output**: Clean Go switch with helper structs, plus sourcemap entries referencing original Dingo spans for LSP diagnostics.

## 6. Context Tracking Mechanism
- **Scope Stack**: Maintained per file; entries hold `kind`, `start`, `end`, `parentBinding`. Pushed by preprocessors when entering constructs (match, enum, lambda) and popped when leaving.
- **Binding Registry**: Unique IDs for user bindings (`let`, pattern variables). Allows `?` operator to attach to correct result context.
- **Feature Timeline**: Ordered list of feature events referencing byte offsets; serialized into sidecar.
- **Synchronization**: Every write to the transformed buffer also informs context builder of byte delta to keep origin ↔ output mapping consistent. Integration with sourcemap writer ensures unified position tracking.

## 7. go/types Integration Plan
1. **Pre-AST Phase**: Context builder tags nodes requiring type info (`needs-type`).
2. **Plugin Phase**: After AST is built, run go/types on the preprocessed file (already done today) but extend it to expose `types.Info` via context loader.
3. **Augmented Lookup**: `ContextFor` returns combined metadata + `types.TypeAndValue` for each node, enabling:
   - Differentiating enum type vs. struct in pattern match.
   - Resolving `None` vs. `Err` context.
   - Validating `?` operator target signature.
4. **Caching**: Since preprocessing is deterministic, reuse types.Config across files; store results in `pkg/typescache` to avoid recomputation during IDE operations.

## 8. Performance Estimates
- Marker emission adds ~2–3% to preprocessing time (simple string writes).
- Sidecar serialization (JSON) per file estimated <1ms for typical 500-line files. Could switch to CBOR for 30% savings later.
- Context loader adds O(n) AST walk with cheap map lookups (positions as keys). Overall expected <5% overhead on current pipeline.
- Memory footprint: sidecar map entries (~32 bytes each). For 1k markers, <32KB per file.

## 9. Migration Path
1. **Phase A**: Introduce context builder + sidecar writing, but keep feature usage behind flag. Validate no output regressions.
2. **Phase B**: Update sourcemap format to V2 (embedding context metadata). Provide converter for existing consumers.
3. **Phase C**: Gradually migrate individual features (`?`, enums, pattern match) to consume context. Maintain compatibility layer that falls back to old behavior if metadata missing.
4. **Phase D**: Deprecate legacy context-less preprocessors once all features rely on the new mechanism.

## 10. Comparison with Precedents
- **TypeScript**: Uses AST transforms with node flags and synthetic comments to preserve source info. Our marker + sidecar mirrors TS synthetic nodes but remains Go-friendly.
- **templ**: Employs source maps + position translation for gopls proxy. We extend this idea with richer metadata for feature semantics.
- **Borgo**: Implements a custom parser—high accuracy but heavy maintenance. By staying on regex + go/parser plus context sidecar, we avoid Borgo’s parser upkeep while still gaining context sensitivity.

## 11. Risks & Mitigations
| Risk | Impact | Mitigation |
| --- | --- | --- |
| Marker proliferation complicates generated Go | Harder debugging, potential parser edge cases | Define strict marker grammar, add sanitizing pass before printing |
| Sidecar out of sync with transformed text | Incorrect context lookups | Centralized buffer writer that updates both text and context offsets atomically |
| Performance regression for IDE usage | Slower feedback | Benchmark on showcase + 5 golden suites, add caching and lazy loading |
| Sourcemap V2 adoption | Tooling churn | Provide compatibility shim + migration doc; versioned readers |

## 12. Success Metrics
1. **Functional**: Pattern matching PoC passes new golden tests; `?` operator handles nested contexts without misbinding.
2. **Quality**: gopls-powered LSP diagnostics map correctly (no off-by-one) in 99% of sampled cases.
3. **Performance**: <=5% slowdown in end-to-end `dingo build` on showcase project.
4. **Maintainability**: Marker grammar + context builder covered by 90% unit test coverage; documentation in `ai-docs/` kept up to date.
5. **Adoption**: All new features (pattern match, Result/Option refinements) consume context sidecar within two release cycles.

## 13. Next Steps
1. Implement `pkg/preprocessor/context` module (builder, serializer, tests).
2. Update preprocessors (error propagation, enums, pattern match) to emit markers + context entries.
3. Extend sourcemap writer to embed context metadata (V2 format documentation).
4. Build context loader plugin + helper API for downstream plugins.
5. Develop pattern matching PoC golden tests to validate end-to-end behavior.

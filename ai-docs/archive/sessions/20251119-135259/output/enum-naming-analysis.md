# Dingo Enum Variant Naming Analysis

## 1. Problem Summary
Dingo currently emits Go identifiers for enum variants using an underscore-delimited pattern (Option A, e.g., `UserStatus_Active`). We must assess alternative conventions (Options B–D) with respect to idiomatic Go usage, discoverability, pattern-matching ergonomics, and IDE experience, then recommend a single standard and outline migration implications.

## 2. Naming Options Considered
| Option | Description | Example Variant | Notes |
| --- | --- | --- | --- |
| **A. Underscore scoped (current)** | `Type_Variant` | `PaymentState_Pending` | Mimics generated `ProtoMessage_State` style |
| **B. Pure CamelCase** | `TypeVariant` | `PaymentStatePending` | Compact but loses explicit namespace separator |
| **C. Namespaced via nested structs** | `Type.Variant()`/`TypeVariant()` | `PaymentState.Pending` or `PaymentStateVariantPending` | Requires additional wrapper types or method factories |
| **D. Mixed prefix (Type + Variant in CamelCase with delimiter)** | `TypeVariant` but exported helper `PaymentStatePending` + alias for pattern matching | Hybrid approach |

> Note: Option C has two sub-variants: (1) literal nested struct values (akin to Swift) and (2) generated helper constructors returning struct values. Both demand more invasive AST rewrites.

## 3. Evaluation Against Criteria
### 3.1 Go Idiomaticity & Clarity
- Idiomatic Go favors exported CamelCase identifiers with minimal punctuation. However, most standard-library enums avoid namespace clashes via type scoping rather than underscores because Go’s `const` blocks naturally scope identifiers by type documentation rather than syntax.
- Generated code ecosystems (protobuf, sqlc, grpc) use `Type_ENUM_VALUE` to guarantee uniqueness and to align with lint tools, providing precedent for Option A.

### 3.2 Collision Avoidance
- Option A guarantees uniqueness even when variant names collide with other identifiers imported into the same file because the type name is prepended with `_` delimiter. No reliance on `const` block scoping.
- Option B risks collisions when two enums share variants named identically and are re-exported or used via dot-imported packages. Mitigations require package-level aliasing.
- Option C (nested structs) isolates namespaces but introduces additional exported structs/methods, increasing surface area and potential import cycles.
- Option D provides uniqueness similar to B but without a delimiter; still collision-prone in long packages.

### 3.3 Pattern Matching Ergonomics
- Dingo’s pattern matching currently parses identifiers textually during preprocessing. Underscore-separated names are trivial to tokenize and reverse-map back to original Dingo variant names, simplifying diagnostics.
- CamelCase-only options require additional heuristics to determine where the type name ends and variant begins, particularly when both segments contain acronyms (e.g., `HTTPStatusOK` vs `HTTPRequest`).
- Nested namespace approaches (Option C) would require more invasive sourcemap support to reconcile `Type.Variant` with generated Go constants, complicating both codegen and plugin pipeline.

### 3.4 Type Safety & Readability
- All options rely on Go’s type system for const values; Option A’s explicit delimiter improves readability by making the variant’s originating enum obvious, especially in mixed switch statements.
- CamelCase (Options B/D) reads more like conventional Go constants but may be harder to skim in long switch statements due to lack of separator. IDE tooltips mitigate this but not when browsing diffs or logs.
- Namespaced methods (Option C) improve readability but hide the fact that variants are compile-time constants; they may allocate or require inline functions, conflicting with “zero runtime overhead” principle.

### 3.5 IDE Experience
- Option A already offers excellent autocomplete grouping because typing `PaymentState_` filters to relevant variants; underscores are handled well by Go completion engines.
- Option B requires typing the entire type name before completion disambiguates; still functional but less visually grouped.
- Option C may produce method-style completions that clash with existing constructors, and gopls may not inline documentation as cleanly for generated nested structs.

## 4. Recommendation
Maintain **Option A (underscore scoped)** as the canonical naming convention for generated Go identifiers. It aligns with established Go generator conventions, guarantees collision-free identifiers without extra tooling, keeps the preprocessor/simple tokenization intact for pattern matching, and produces superior autocomplete grouping. The minor departure from idiomatic CamelCase is justified by interoperability and by precedent (protobuf, sqlc, gqlgen).

## 5. Trade-offs Summary
- **Pros (Option A):**
  - Namespace clarity (`Type_` prefix) with zero runtime cost
  - Simplified pattern matching mapping and sourcemap generation
  - Proven ecosystem precedent; no additional compiler complexity
  - Works seamlessly with existing tests and golden outputs
- **Cons:**
  - Slightly noisier identifiers vs pure CamelCase
  - Deviates from standard library enum naming style (which relies on manual const scoping)
  - Requires lint exceptions if future `golint`-like tools flag underscores

Alternatives:
- **Option B:** Cleaner look but risks collisions; complicates variant extraction logic.
- **Option C:** Best namespace semantics but high implementation overhead and potential runtime changes.
- **Option D:** Hybrid aesthetic but inherits Option B’s ambiguity without solving delimitation for tooling.

## 6. Migration Plan (if future change desired)
Although we recommend retaining Option A, migrating would involve:
1. **Config Gate:** Introduce `enum.naming_style` in `dingo.toml` allowing `underscore` (default) or experimental alternatives.
2. **Preprocessor Update:** Adjust enum preprocessor to emit variant identifiers per style while keeping sourcemap metadata of original variant names.
3. **Plugin Alignment:** Update AST plugin injection to reference new names when generating helper methods (Option/Result wrappers).
4. **Golden Tests:** Re-baseline all enum-related golden files; add dual-style test coverage to prevent regressions.
5. **Deprecation Timeline:** Provide at least one release with warnings before changing defaults to avoid breaking downstream code, mirroring Go’s proposal process.

## 7. Edge Cases & Implementation Notes
- **Acronyms:** Option A naturally handles acronyms (`HTTPStatus_OK`) without additional formatting rules.
- **Cross-Package Reuse:** When enums are exported from shared packages, underscores prevent collisions in importers, especially when gofmt groups multiple consts into single `switch` statements.
- **Pattern Matching Guards:** String-based guard diagnostics rely on underscore splitting; altering naming requires reworking guard error messages and references.
- **Interop with Go Enums:** When converting `const` blocks defined manually in Go files, Option A naming avoids overshadowing existing user-defined constants.
- **Tooling:** Any future LSP features (hover breadcrumbs, go-to-definition) only need to understand underscore format, minimizing maintenance.

## 8. References
- Go protobuf generated enums (e.g., `descriptorpb.FieldDescriptorProto_Type`) for delimiter precedent.
- `sqlc` generated enums (`order_status__PENDING`) showing underscore namespace patterns.
- Go Style Guide (Effective Go) noting exported identifiers should be CamelCase *unless* part of generated code ensuring uniqueness.

**Conclusion:** Retain Option A as default and document rationale in `tests/golden/README.md` + `ai-docs/claude-research.md` to reinforce consistency across future features.

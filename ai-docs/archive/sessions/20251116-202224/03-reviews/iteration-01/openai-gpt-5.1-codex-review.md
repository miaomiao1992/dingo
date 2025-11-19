# Code Review: Sum Types Implementation
**Model**: GPT-5.1 Codex via claudish
**Date**: 2025-11-16
**Reviewer**: OpenAI GPT-5.1 Codex

---

## Architecture

### 1. CRITICAL – Plugin output ordering causes duplicated declarations
**Location**: `/Users/jack/mag/dingo/pkg/plugin/builtin/sum_types.go:208-216, :143-160, :84-87`

`SumTypesPlugin.Transform` appends every generated decl to `p.generatedDecls`, but `generateTagEnum` appends the const block immediately (`sum_types.go:208-216`) while `transformEnumDecl` separately appends the returned type decl and later helpers (`:143-160`). Because `p.generatedDecls` is appended to the file only once (`:84-87`), the tag `const` block ends up twice per enum (once from `generateTagEnum`, once when `p.generatedDecls` is appended). This corrupts Go output and breaks compilation.

### 2. IMPORTANT – Match transformation ignores expression/statement context
**Location**: `/Users/jack/mag/dingo/pkg/plugin/builtin/sum_types.go:444-469`

`transformMatchExpr` blindly replaces the placeholder expression with a `switchStmt` regardless of whether the original match was used inside an expression. A `switch` statement cannot appear where an expression is required, so any `match` used as an expression (per design) will generate invalid Go. Need expression-friendly lowering (e.g., wrapping switch in immediately invoked function or building chained ifs).

### 3. IMPORTANT – Enum registry collected but unused
**Location**: `/Users/jack/mag/dingo/pkg/plugin/builtin/sum_types.go:92-102`

`collectEnums` stores declarations in `p.enumRegistry`, yet the registry is never referenced elsewhere. Match arm lowering (`transformMatchArm`) still hardcodes `"Tag_"+variant` and lacks access to the enum name, so variant tags cannot be resolved correctly. Dead state plus incorrect code generation.

### 4. IMPORTANT – Placeholder nodes not removed from Dingo node map
**Location**: `/Users/jack/mag/dingo/pkg/plugin/builtin/sum_types.go:162-165`

After `cursor.Delete()` in `transformEnumDecl`, the associated placeholder remains inside `currentFile.DingoNodes`, so later walks may still think the deleted decl exists. Either remove mapping or ensure `AddDingoNode` entries are cleaned.

---

## Type Safety

### 5. CRITICAL – Struct field injection unsafe for tuple variants
**Location**: `/Users/jack/mag/dingo/pkg/plugin/builtin/sum_types.go:265-281`, `/Users/jack/mag/dingo/pkg/parser/participle.go:660-668`

`generateVariantFields` assumes `variant.Fields.List` has named Idents, but tuple variants synthesize names like `_0` (`parser/participle.go:660-668`). Lowercasing the variant name plus `_0` produces fields such as `circle__0` that are never assigned by constructors (which iterate names exactly as stored). This mismatch leaves nil pointers dereferenced later. Need deterministic mapping for positional tuple fields and ensure constructors/destructuring use same schema.

### 6. IMPORTANT – Constructors for tuple variants alias ast.Field slices
**Location**: `/Users/jack/mag/dingo/pkg/plugin/builtin/sum_types.go:295-299`

`generateConstructor` reuses `variant.Fields.List` directly as the parameter list. These fields include pointer types & names, but they are shared with the enum definition. Mutating params later (e.g., adding doc/comments) will mutate AST of original enum. Should deep-copy to avoid aliasing.

### 7. IMPORTANT – Match arm tag constants incorrect
**Location**: `/Users/jack/mag/dingo/pkg/plugin/builtin/sum_types.go:483-487`

`transformMatchArm` fabricates `Tag_<Variant>` without enum prefix, but generated constants are `EnumTag_Variant`. As-is, every case clause references undefined identifiers.

### 8. IMPORTANT – No nil guarding for pointer payloads
**Location**: `/Users/jack/mag/dingo/pkg/plugin/builtin/sum_types.go:489-494`

Constructors take values by value and store `&param`. When called with addressable temporaries, Go takes address of stack copy that escapes; OK, but match lowering reads `*matchedExpr.circle_field` without nil checks. With incorrect variant use or partial initialization, this panics rather than producing compile errors. Need defensive nil validation or exhaustiveness checking prior to deref.

---

## Error Handling

### 9. CRITICAL – No error for duplicate variant names
Parser accepts enums with duplicate variant identifiers; plugin happily generates duplicate consts/fields causing Go compile errors without clear diagnostics. Need validation stage.

### 10. IMPORTANT – Missing errors for unsupported pattern forms
**Location**: `/Users/jack/mag/dingo/pkg/plugin/builtin/sum_types.go:483-486`

`transformMatchArm` silently treats any non-wildcard as variant pattern but relies on `pattern.Variant` being non-nil. If parser produced literal pattern (planned extension) or pattern lacks variant, this panics. Need explicit error return to user.

### 11. IMPORTANT – Guards ignored
**Location**: `/Users/jack/mag/dingo/pkg/plugin/builtin/sum_types.go` (transformMatchArm), `/Users/jack/mag/dingo/pkg/parser/participle.go:166-169`

`MatchArm.Guard` is parsed but `transformMatchArm` discards it; guard conditions never enforced, producing semantic miscompilation. At minimum should emit TODO error to avoid incorrect behavior.

---

## Testing

### 12. CRITICAL – No tests for enum transformation output
No unit or golden tests verifying generated Go for enums (constructors, helpers, generics). Need fixture ensuring struct layout, tag constants, helper bodies.

### 13. IMPORTANT – No tests ensuring match expressions compile
Even minimal cases (expr vs statement contexts, wildcard default) must be covered to prevent invalid switch generation identified above.

### 14. MINOR – Missing registry tests
Should add tests confirming multiple enums per file produce isolated tags/structs without collisions and `Reset` clears state.

---

## Recommended Tests

- **Enum generation golden**: Input `.dingo` with unit/tuple/struct variants; verify tag consts, struct fields, constructors, helpers (incl. generics).
- **Match lowering**: Expression context requiring value, statement context, wildcard-only match, guard rejection test.
- **Negative cases**: Duplicate variants, match using unknown variant, tuple destructuring mismatch, missing exhaustiveness (until feature lands, expect TODO error).
- **State isolation**: Multiple files invoking plugin sequentially must not leak `enumRegistry`.

---

## Summary

STATUS: CHANGES_NEEDED
CRITICAL_COUNT: 4
IMPORTANT_COUNT: 7
MINOR_COUNT: 1

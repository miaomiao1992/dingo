
[claudish] Model: openai/gpt-5.1-codex

# Investigation Results

## Executive Summary
Golden tests are stale in several areas: missing `// dingo:s:N` markers stem from a real regression in the error propagation plugin, while duplicate markers and unqualified helper calls reflect outdated expectations. The transpiler’s gofmt divergence (single import written as a block) is cosmetic but should be corrected to match canonical Go formatting. Adopt the new behavior for helper qualification, but fix the generator to restore start markers and align import rendering.

## Root Cause
- `// dingo:s:N` comments were previously injected by the preprocessor; a refactor around commit `2a76f92` removed that insertion path, so only end markers remain (`pkg/generator/preprocessor/error_prop.go` and related plugin code).
- Duplicate `// dingo:s:N` entries in goldens were artifacts of the older two-pass wrapper; current generator correctly emits a single pair, so goldens need refresh.
- Import formatting mismatch arises because the emitter always renders parenthesized blocks even when only one import survives (`pkg/generator/emitter/imports.go` or equivalent), deviating from gofmt output.
- Bare calls like `ReadFile`/`Atoi` were intentionally replaced by qualified versions via the unqualified-import processor, making the existing goldens incorrect.

## Source of Truth Decision
Treat goldens as authoritative regarding marker presence and gofmt-style imports (they capture intended UX), but regenerate them for helper qualification and duplicate markers once the generator is fixed. Transpiler must restore start markers to uphold source-map contracts, so code changes take precedence there.

## Source Map Comments Analysis
`// dingo:s:N` / `// dingo:e:N` bracket injected scaffolding so the LSP/source-map pipeline can map diagnostics and hovers back to the originating Dingo span. They should remain in final Go output (non-exported comments) because downstream tooling relies on them; both markers are required for reliable span reconstruction.

## Implementation Plan
1. **Restore start markers** – File: `pkg/generator/preprocessor/error_prop.go` (and associated plugin), ensure `insertErrorHandling` emits paired comments; add tests in `pkg/generator/preprocessor/error_prop_test.go`.
2. **Normalize import rendering** – File: `pkg/generator/emitter/emitter.go` (import writer); emit single-line `import "pkg"` when only one spec, otherwise keep block; cover with unit test.
3. **Refresh goldens** – Update `tests/golden/error_prop_*.go.golden` after fixes to include qualified helpers, remove duplicate markers, and reflect gofmt imports.
4. **Document expectations** – Amend `tests/golden/GOLDEN_TEST_GUIDELINES.md` to specify marker requirements and gofmt-format imports.
5. **Regression check** – Run `go test ./tests -run TestGoldenFiles` and archive summary in session output.

## Code Locations
- Error propagation preprocessing & comments: `pkg/generator/preprocessor/error_prop.go`
- Import rendering utilities: `pkg/generator/emitter/*` (import writer)
- Golden expectations: `tests/golden/error_prop_*.go.golden`
- Guidelines doc: `tests/golden/GOLDEN_TEST_GUIDELINES.md`

## Confidence Level
HIGH – Analysis is grounded in current generator behavior vs. documented expectations, and proposed fixes align with architecture goals.

Details: `ai-docs/sessions/20251119-012253/output/golden-test-investigation.md`

[claudish] Shutting down proxy server...
[claudish] Done


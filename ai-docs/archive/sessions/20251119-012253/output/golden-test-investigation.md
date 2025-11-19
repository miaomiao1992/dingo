# Golden Test Investigation (2025-11-19)

## Summary
- Golden tests `error_prop_01_simple`, `error_prop_03_expression`, and `error_prop_04_wrapping` fail due to divergent expectations around comment markers, import formatting, and symbol qualification.
- The transpiler currently emits code that omits `// dingo:s:N` start markers and always uses parenthesized import blocks with fully-qualified package identifiers; goldens still encode earlier conventions.
- Determining which side is authoritative depends on whether the behavior is a regression (missing start markers) versus a desired modernization (qualified package access).

## Observed Symptoms
| Test | Expected vs. Actual | Symptom |
|------|---------------------|---------|
| `error_prop_01_simple` | Expected single-line import and paired `// dingo:s:1`/`// dingo:e:1`; actual uses import block and is missing start marker. | Missing `dingo:s` comment and import formatting shift. |
| `error_prop_03_expression` | Expected duplicate `// dingo:s:1`/`// dingo:e:1` comments and bare `Atoi` call; actual has single markers and `strconv.Atoi` with explicit import. | Duplicate comment removal + enforced qualification. |
| `error_prop_04_wrapping` | Expected `fmt`-only import and bare `ReadFile`; actual adds `os` import, qualifies call, and drops start marker. | Same trio of diffs as above. |

## Root Cause Analysis
1. **Missing `// dingo:s:N` comments**: The error-propagation plugin now emits only the end markers while the start markers were previously injected in the preprocessor. A refactor (commit `2a76f92`) removed duplicate markers but accidentally eliminated the start marker insertion in cases without guard clauses. This is a regression in the generator, not a golden issue.
2. **Import formatting differences**: The Go printer is configured with `printer.TabIndent` and `printer.UseSpaces`, but we manually build `ast.File.Imports` in sorted order. When only one import remains we still render it through the multi-import helper, leading to the parenthesized form. This does not violate Go syntax, but golden files assume gofmt's single-line canonicalization. Decide whether to update generator (preferred) or relax goldens.
3. **Duplicate `dingo` comments**: Goldens with double `// dingo:s:1` arose from the old preprocessor that wrapped both the `if` block and its body. The current AST-aware implementation emits exactly one start/end pair per propagation site, which is the intended behavior; goldens are stale here.
4. **Unqualified helper calls**: The transpiler now resolves unqualified helper invocations via the `UnqualifiedImportProcessor`, ensuring Dingo code that calls standard library helpers without package prefixes receives the proper import and qualified call. Goldens predating this feature deliberately omitted `strconv`/`os` qualifications, so they no longer reflect actual semantics.

## Source of Truth Decisions
- **Comment coverage**: The original contract requires both `// dingo:s:N` and `// dingo:e:N` to bracket transformed regions for source-map integrity. Since the transpiler currently emits only end markers, *code needs to be fixed*; the goldens remain authoritative for this aspect.
- **Duplicate markers**: The intended design is a single pair per site. Update goldens that still contain duplicates once the generator restores start markers (without duplication).
- **Import formatting**: Align generator with gofmt (single-line import when len==1). Treat goldens as source of truth for canonical formatting to keep diffs minimal and maintain parity with user expectations of gofmt output.
- **Unqualified helpers**: New behavior (auto qualification + import insertion) is correct and necessary for compilable output. Update golden files to call `strconv.Atoi`, `os.ReadFile`, etc., and include the matching imports.

## Purpose of `// dingo:s/e:N` Comments
- They delineate sections of generated Go that correspond to a single high-level Dingo expression, enabling:
  - Accurate span mapping in source maps consumed by the LSP proxy.
  - Precise diagnostics (highlighting the original Dingo span rather than the synthesized Go scaffolding).
  - Future pattern-match rewrite hints that need to know the boundaries of injected error-handling blocks.
- Each `N` index ties back to the preprocessor metadata so downstream tools can relate multiple injected segments to the same Dingo construct.

## Import Formatting Considerations
- Goal: Match `gofmt` output to avoid noisy diffs and align with developer expectations.
- Implementation detail: when only one import spec exists after deduplication, write `f.Imports = []*ast.ImportSpec{...}` and let `printer.Fprint` emit `import "pkg"`. When multiple remain, wrap them in the parenthesized block.
- The auto-qualification pass must continue injecting imports even if the user already had one, but should re-run `ast.SortImports` afterward to keep deterministic ordering.

## Prioritized Implementation Plan
1. **Restore start-marker emission (pkg/transpiler/plugins/errorprop)**
   - Reintroduce insertion of `// dingo:s:N` comments adjacent to the generated `if` statement before running go/printer.
   - Add unit tests in `pkg/transpiler/plugins/errorprop/errorprop_test.go` that assert both markers exist.
   - Update failing goldens once generator change lands.
2. **Import formatting fix (pkg/generator/emitter or wherever `renderImports` lives)**
   - Detect single-import case and emit the condensed form; add regression test in `tests/golden/unqualified_import_*` to lock format.
3. **Golden updates for helper qualification**
   - Regenerate affected golden files (`error_prop_03_expression`, `error_prop_04_wrapping`, others referencing bare helpers) after the code fixes, ensuring they now include the proper imports and qualified calls.
4. **Documentation alignment**
   - Update `tests/golden/GOLDEN_TEST_GUIDELINES.md` to state that goldens must mirror gofmt output and include both `dingo` comment markers.
5. **Follow-up validation**
   - Re-run `go test ./tests -run TestGoldenFiles` and capture summary in `ai-docs/sessions/.../output/tests.txt` for future reference.

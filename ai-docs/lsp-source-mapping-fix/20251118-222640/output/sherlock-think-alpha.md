
[claudish] Model: openrouter/sherlock-think-alpha

**Root Cause**

Error propagation mappings incorrectly set `OriginalColumn` for `error_prop` regions to relative position of `?` within `expr` substring (e.g., col 13 for "ReadFile(path)?"), not absolute column in full line (e.g., col 28 for "let data = ReadFile(path)?"). 

`qPos := strings.Index(expr, "?")` (error_prop.go:328,492) computes relative offset. All 7+ `error_prop` mappings use `qPos+1` as absolute `OriginalColumn`. 

gopls "undefined: ReadFile" diagnostic starts at Go col~18 (`ReadFile`). `MapToOriginal` (sourcemap.go:60) misses narrow `expr_mapping` (`Length=len(exprClean)`), falls back to closest `error_prop` (gen col 1), offsets to wrong relative col → underlines `?` instead.

**Specific Solution**

Replace relative `qPos` with absolute: `qPos := strings.Index(matches[0], "?")` (full line).

Since prior Edit caused compile error ( `originalText` undefined at qPos line), apply these **exact** replacements:

**pkg/preprocessor/error_prop.go**

1. Line ~328 (expandAssignment):
```
old: qPos := strings.Index(originalText, "?")
new: qPos := strings.Index(matches[0], "?")
```

2. Line ~492 (expandReturn):
```
old: qPos := strings.Index(originalText, "?")
new: qPos := strings.Index(matches[0], "?")
```

Align to **0-based indexing** (LSP standard): Change all `OriginalColumn: qPos + 1` → `qPos`, `GeneratedColumn` similarly. Update `MapToOriginal`/`MapToGenerated` offsets: `-1` instead of no adjustment.

**Verification**

1. Apply Edits above → code compiles.
2. `cd ../../../ && go test ./tests -run TestGoldenFiles/error_prop_01_simple -v` → passes.
3. `cd ../../../tests/golden && go run ../../cmd/dingo/main.go build error_prop_01_simple.dingo` → inspect `error_prop_01_simple.go.sourcemap.json`: `"original_column"` for `error_prop` ~27-28 (absolute `?` col, 0-based).
4. LSP: VSCode + dingo-lsp on `error_prop_01_simple.dingo`, no `os` import → underlines `ReadFile`, not `?`.
5. git diff + commit: "fix(lsp): absolute column for error_prop mappings C8".

Fixes diagnostic positioning; `expr_mapping` + fallback now accurate. Session outputs confirm (agent summary matches code analysis).

[claudish] Shutting down proxy server...
[claudish] Done


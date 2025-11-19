
[claudish] Model: openai/gpt-5.1-codex

Root cause is in `pkg/preprocessor/rust_match.go:49-93`: after transforming a real match expression, the loop keeps scanning the already-emitted Go code, so the line `panic("unreachable: match is exhaustive")`—which contains the substring “match ” after a colon—gets mistaken for a new match expression, causing `collectMatchExpression` to grab an unterminated chunk and eventually surface “no pattern arms found.” Passing tests (e.g., `pattern_match_01_basic.dingo`) avoid this because they don’t emit a nearby `panic(...match...)` line, so the false detection never fires. Proposed fix: tighten detection to only fire when the trimmed line actually starts with `match` (or `let/var/return` immediately followed by `match`), ensuring punctuation like `:` cannot precede `match` and preventing reprocessing of generated code. Validate by adding a regression test in `pkg/preprocessor/rust_match_test.go` covering a transformed `panic` line, then run `go test ./pkg/preprocessor -run RustMatch` plus `go test ./tests -run TestGoldenFiles/pattern_match_01_simple`. Confidence is medium-high; assumption is that inline constructs like `if cond { match ... }` either remain unsupported or will need an explicit allowance in the refined heuristic.

[claudish] Shutting down proxy server...
[claudish] Done


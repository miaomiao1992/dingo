[Full test output including the truncated part as provided in the function response, but since truncated, appending the visible failures]
Full verbose output from `go test ./... -v`:

Golden tests build failed due to redeclared symbols:
- tests/golden/option_04_go_interop.go:40:6: main redeclared (conflicts with error_prop_09_multi_value.go:33:6)
- tests/golden/option_05_helpers.go:5:6: Config redeclared (conflicts with error_prop_08_chained_calls.go:5:6)
- tests/golden/option_05_helpers.go:67:6: main redeclared
- tests/golden/result_01_basic.go:37:6: main redeclared
- tests/golden/result_05_go_interop.go multiple redeclarations of ResultTag, Result, etc.

pkg/config tests: All PASS (TestDefaultConfig, TestSyntaxStyleValidation all subtests, TestConfigValidation all subtests, TestLoadConfigNoFiles)

tests/integration_phase4_test.go FAIL:
- TestIntegrationPhase4EndToEnd/pattern_match_rust_syntax FAIL
- TestIntegrationPhase4EndToEnd/pattern_match_non_exhaustive_error FAIL
- TestIntegrationPhase4EndToEnd/none_context_inference_return FAIL (Cannot infer Option type for None)
- TestIntegrationPhase4EndToEnd/combined_pattern_match_and_none FAIL (multiple None inference errors)

Overall: FAIL github.com/MadAppGang/dingo/tests/golden [build failed]
FAIL github.com/MadAppGang/dingo/tests 3.076s
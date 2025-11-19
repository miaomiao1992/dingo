
[claudish] Model: openai/gpt-5.1-codex

# Type Injection Investigation Summary

- Result/Option plugins correctly collect pending type declarations, but `generator.Generate` never emits the pipeline’s injected AST before it’s cleared, so no Go declarations reach the printer.
- Fix by having `generator.NewWithPlugins` register declaration providers and calling `pipeline.GetInjectedTypesAST()` inside `Generate` (before `ClearPendingDeclarations`) so the decls append to `ast.File.Decls`.
- Recommended testing: rerun golden suite, especially `tests/golden/pattern_match_01_simple`.  
- Full executive summary, root cause details, and fix steps: `ai-docs/sessions/20251119-125928/output/summary.txt`.

[claudish] Shutting down proxy server...
[claudish] Done


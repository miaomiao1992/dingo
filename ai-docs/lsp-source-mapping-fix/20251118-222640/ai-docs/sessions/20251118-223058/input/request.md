LSP Source Map Position Translation Issue Investigation

Problem: gopls diagnostics underline wrong position in .dingo (e.g., underlines ? instead of ReadFile).

Files to analyze:
- pkg/preprocessor/error_prop.go (expandLetStatement)
- pkg/preprocessor/sourcemap.go (MapToOriginal)
- pkg/lsp/handlers.go (TranslateDiagnostics)
- pkg/lsp/gopls_client.go

Current source map example:
{
  "mappings": [
    {
      "generated_line": 4,
      "generated_column": 1,
      "original_line": 4,
      "original_column": 15,
      "length": 1,
      "name": "error_prop" 
    }
  ]
}

Expected: Map 'ReadFile' position accurately.
Root cause analysis needed + exact code changes.

Output files:
- ai-docs/sessions/20251118-223058/output/root-cause.md (detailed analysis)
- ai-docs/sessions/20251118-223058/output/proposed-fix.md (code changes)
- ai-docs/sessions/20251118-223058/output/summary.txt (3-5 sentences for main chat)

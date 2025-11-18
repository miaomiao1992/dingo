
[claudish] Model: openai/gpt-5.1-codex

Root cause: `pkg/lsp/gopls_client.go:49-112` closes the gopls stdio pipes the moment the subprocess exits, so `conn.Call` in `pkg/lsp/server.go:108-170` writes to a destroyed jsonrpc2 stream and VS Code reports “Cannot call write after a stream was destroyed.” Because `cmd/dingo-lsp/main.go:34-77` immediately blocks on `<-conn.Done()>`, that write error tears down the entire IDE connection before any diagnostic can be sent. Fixes: wrap every `Call`/`Notify` (pkg/lsp/gopls_client.go:130-205) with connection-health checks that propagate errors back to the handler, keep the IDE-side conn alive while supervising gopls with a restart/backoff loop, and only close pipes after the shutdown flag is set (pkg/lsp/gopls_client.go:208-245). Also surface gopls crashes via `window/showMessage` using the stored IDE conn (pkg/lsp/server.go:20-67) and log stderr to `window/logMessage` so users see actionable errors. Detailed notes: `ai-docs/sessions/20251118-190000/output/summary.txt`.

[claudish] Shutting down proxy server...
[claudish] Done


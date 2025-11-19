# Dingo LSP Proxy Review (2025-11-18)

## ✅ Strengths
- Clear modular separation between transport (cmd/dingo-lsp), proxy logic (pkg/lsp/server.go), gopls subprocess management, and VS Code client glue keeps responsibilities understandable.
- Logging abstraction with configurable verbosity (`pkg/lsp/logger.go`) and structured messages makes field debugging easier than ad-hoc prints.
- Source map cache uses an RWMutex and validates version numbers (`pkg/lsp/sourcemap_cache.go`) so corrupted/legacy maps fail fast instead of silently corrupting position mapping.
- File watcher debounces bursty write events and ignores common build/vendor folders (`pkg/lsp/watcher.go`), which should reduce unnecessary rebuilds.

## ⚠️ Concerns

### Critical
1. **JSON-RPC handler never registered, server cannot reply to the editor**
   *Files*: `/Users/jack/mag/dingo/cmd/dingo-lsp/main.go:39-47`, `/Users/jack/mag/dingo/pkg/lsp/server.go:62-75`
   The server builds a `jsonrpc2.Conn` but never installs `handleRequest` on it (the result of `jsonrpc2.ReplyHandler` is discarded). As a result the proxy never receives `initialize`, `textDocument/*`, or any other method from VS Code, so the IDE hangs waiting for replies. **Recommendation**: construct the connection with `jsonrpc2.NewConn(ctx, stream, jsonrpc2.HandlerServer(s.handleRequest))` (or equivalent) and pump it; add tests that a trivial `initialize` round-trip succeeds.
2. **No path for gopls → IDE notifications, diagnostics silently dropped**
   *Files*: `/Users/jack/mag/dingo/pkg/lsp/gopls_client.go:101-210`, `/Users/jack/mag/dingo/pkg/lsp/handlers.go:279-316`
   The proxy issues `Call/Notify` into gopls but never reads asynchronous results: there is no goroutine driving `conn.Run`, `Recv`, or handling `window/logMessage` & `textDocument/publishDiagnostics`. The stub `handlePublishDiagnostics` is never wired, so diagnostics, progress, and server-initiated requests never reach the editor, breaking core LSP guarantees. **Recommendation**: run the gopls connection with a handler that routes notifications back through the proxy connection (and implement publish diagnostics translation end-to-end) before exposing this binary.
3. **“Forward to gopls” returns errors for all unimplemented methods**
   *File*: `/Users/jack/mag/dingo/pkg/lsp/server.go:317-321`
   Currently any method not explicitly handled (e.g., references, rename, workspace/symbol, semantic tokens) returns `method not implemented`, so Dingo projects lose the majority of gopls features out of the gate. Given gopls already knows how to service these requests on transpiled Go, blocking them is a regression. **Recommendation**: default case should transparently proxy to gopls (with URI translation when needed) rather than erroring. Keep a deny-list only for methods that truly require bespoke behavior.
4. **Position translation fallback still sends .dingo URIs to gopls**
   *Files*: `/Users/jack/mag/dingo/pkg/lsp/handlers.go:146-210`, `/Users/jack/mag/dingo/pkg/lsp/translator.go:52-78`
   When a source map is missing or stale, `TranslatePosition` returns an error, yet the handlers proceed to call `s.gopls.*` with the original `.dingo` URI. gopls does not understand this scheme and replies with errors, so completions/definitions/hover all fail for any file that has not been transpiled yet. **Recommendation**: short-circuit with a user-facing diagnostic ("transpile file first") or trigger transpilation to refresh the map before reissuing the request; never forward `.dingo` URIs to gopls.

### Important
1. **VS Code “transpileOnSave” setting is ignored, auto-transpile always on**
   *Files*: `/Users/jack/mag/dingo/cmd/dingo-lsp/main.go:29-34`, `/Users/jack/mag/dingo/editors/vscode/src/lspClient.ts:16-33`
   The extension sets `DINGO_AUTO_TRANSPILE` in the environment, but `ServerConfig` is hard-coded with `AutoTranspile: true`. Users cannot disable the watcher/auto-build loop even if it thrashes their CPU. **Recommendation**: read the env flag (default true) when building `ServerConfig` and document it alongside the VS Code setting.
2. **Auto-transpiler assumes `dingo build <file>` always exists and blocks per file**
   *File*: `/Users/jack/mag/dingo/pkg/lsp/transpiler.go:29-67`
   `OnFileChange` launches `dingo build <file>` synchronously per change with no queueing/backoff; multiple saves spawn overlapping builds, each invalidating maps and spamming gopls notifications. Failures only log an error—no diagnostics sent back—so the user never sees why IntelliSense broke. **Recommendation**: serialize builds via a worker, propagate errors back as diagnostics (the helper exists), and skip re-running if another build is in-flight for the same file.
3. **File watcher ignores newly created directories after startup**
   *File*: `/Users/jack/mag/dingo/pkg/lsp/watcher.go:59-117`
   The watcher walks the tree once and adds existing directories, but when a developer creates a new folder (common in Go modules) its files are never watched, so auto-transpile silently misses them. **Recommendation**: watch for `Create` events on directories and call `Add`; alternatively rely on `fsnotify` recursive implementation or rewalk periodically.
4. **gopls restart logic is dead code**
   *File*: `/Users/jack/mag/dingo/pkg/lsp/gopls_client.go:214-227`
   `handleCrash` is never invoked and no goroutine monitors process exit, so a gopls crash leaves the proxy permanently broken until the user restarts VS Code manually. **Recommendation**: hook `cmd.Wait` in a goroutine, call `handleCrash`, and surface failures via `window/showMessage` so the extension can prompt the user.
5. **Workspace change notifications sent without declaring capability**
   *Files*: `/Users/jack/mag/dingo/pkg/lsp/transpiler.go:63-72`, `/Users/jack/mag/dingo/pkg/lsp/server.go:130-150`
   The proxy notifies `workspace/didChangeWatchedFiles`, but its advertised capabilities never register `DidChangeWatchedFilesCapabilities`. LSP clients may drop unsolicited notifications, so gopls might not rebuild indexes when the proxy rewrites Go files. **Recommendation**: include the capability in `InitializeResult` or stop sending the notification.

### Minor
1. **Source map cache has fixed size field but no eviction**
   *File*: `/Users/jack/mag/dingo/pkg/lsp/sourcemap_cache.go:23-88`
   `maxSize` is documented as an "LRU limit" yet unused; long-lived workspaces leak maps indefinitely. Implement actual eviction or drop the field to avoid misleading future contributors.
2. **VS Code client never restarts on crash**
   *File*: `/Users/jack/mag/dingo/editors/vscode/src/lspClient.ts:45-48`
   `errorHandler.closed` returns `{ action: 1 }` (`CloseAction.DoNotRestart`). Any transient failure requires manual command invocation. Consider matching the default language client behavior unless there’s a strong reason not to.

## 🔍 Questions
1. What is the intended plan for translating and publishing diagnostics? The current stub in `handlePublishDiagnostics` never executes; do we expect the IDE to read `.go.map` files directly instead?
2. Should the proxy proactively transpile on `didOpen`/`didChange` to ensure maps exist, or is there a separate workflow guaranteeing `dingo build` has run beforehand?
3. Is there a product requirement to block access to gopls features like references/rename, or was the non-forwarding fallback temporary scaffolding?

## 📊 Summary
- **Status**: MAJOR_ISSUES — basic initialize/diagnostic handling pathways are incomplete so the proxy cannot yet be used interactively.
- **Critical**: 4 | **Important**: 5 | **Minor**: 2
- **Testability**: Low. There are no integration tests exercising JSON-RPC request flow, gopls subprocess supervision, or source-map-driven translations, making it hard to catch the regressions noted above. Building thin end-to-end tests (initialize → completion) plus unit tests for translation fallbacks would significantly raise confidence.

# Gopls vs Dingo-LSP Architecture Comparison & VS Code Failure Diagnosis

## Gopls Architecture (From Research)
- **Entrypoint**: `gopls/main.go` (no `cmd/gopls` in some views; uses stdio mode via `-mode=stdio`).
- **Key Components**: `internal/lsp`, handlers for protocol methods; JSON-RPC2 over stdio.
- **Initialization**: Standard LSP handshake (`initialize` → capabilities → `initialized`). Runs as subprocess in stdio mode. No detailed VS Code specifics from fetches (404s/502s), but integrates via VS Code Go extension proxying to gopls binary.
- **Common Issues**: Binary not in PATH, workspace config mismatches, slow startup timeouts.

## Dingo-LSP Architecture (Proxy)
- **Stack**: VS Code extension (`lspClient.ts`) → dingo-lsp binary (`cmd/dingo-lsp/main.go`) → gopls subprocess (`pkg/lsp/gopls_client.go`).
- **Flow**:
  1. VS Code spawns `dingo-lsp` via stdio, sets env `DINGO_LSP_LOG=info`.
  2. dingo-lsp finds gopls via `exec.LookPath`, creates `Server` with `GoplsClient`.
  3. `Server` wraps stdin/stdout as `ReadWriteCloser`, creates JSON-RPC2 `Conn`.
  4. **Critical**: `server.SetConn(conn, ctx)` **before** `conn.Go(ctx, handler)` to avoid nil races.
  5. Handler routes: `initialize` → forward to gopls → modify capabilities → reply.
- **Key Packages**:
  | File | Role |
  |------|------|
  | `pkg/lsp/server.go` | Routes LSP methods, translates .dingo → .go URIs, auto-transpile on save. |
  | `pkg/lsp/gopls_client.go` | Manages gopls subprocess (pipes, crash recovery up to 3x, 30s init timeout). |
  | `cmd/dingo-lsp/main.go` | Entry: logger, find gopls, NewServer, stdio transport. |
  | `editors/vscode/src/lspClient.ts` | Spawns dingo-lsp, documentSelector='dingo', errorHandler=restart, initFailed=no retry. |

## Initialization Sequence Comparison
```
Standard gopls (direct):
VS Code Go Ext → gopls binary (stdio) → initialize → capabilities → initialized

Dingo-LSP (proxy):
VS Code Dingo Ext → dingo-lsp binary (stdio)
  ↓ main.go: NewServer(goplsPath) → GoplsClient.start() → gopls subprocess (pipes)
  ↓ SetConn → conn.Go(handler)
  ↓ VS Code: initialize → server.handleInitialize → gopls.Initialize(30s timeout)
    ↓ Modify caps (add Dingo triggers) → reply to VS Code
  ↓ VS Code: initialized → server.handleInitialized → gopls.Initialized
```

- **Matches gopls**: Stdio JSON-RPC2, same methods (`initialize`, `initialized`, `shutdown`/`exit`).
- **Extensions**: URI translation (.dingo→.go), source map translator, auto-transpile watcher.
- **Safeguards**: Buffered RWC (32KB), gopls crash recovery, logging, timeouts.

## Recent Changes (git diff HEAD~1)
- **New Files**: All LSP impl added (`main.go`, `gopls_client.go`, `server.go`, `lspClient.ts`).
- **Key Fixes**:
  - `SetConn` before `conn.Go` (race fix).
  - Buffered RWC, bufio.Scanner for stderr.
  - Shutdown flag prevents crash loops.
  - VS Code: `initializationFailedHandler` shows output, no auto-retry.
  - Env passthrough for logs/transpile.

## VS Code Extension Failure Diagnosis
**Symptoms**: Fails during LSP init (from `lspClient.ts` errorHandler/initFailed).
**Root Causes (Ranked)**:
1. **dingo-lsp Binary Missing**: ENOENT if not built/in PATH. VS Code defaults `lsp.path='dingo-lsp'`.
2. **gopls Missing**: dingo-lsp fatals (`logger.Fatalf`) → pipe breaks → VS Code sees init fail.
3. **Init Timeout/Handshake**: gopls slow startup >30s, or JSON-RPC mismatch.
4. **Workspace Mismatch**: No `RootURI` → no watcher/workspacePath.
5. **Logging Hidden**: Check Output 'Dingo Language Server' panel for fatals.

**Evidence**:
- `lspClient.ts`: Catches ENOENT → 'dingo-lsp not found'; shows 'View Output'.
- `main.go`: Fatals on no gopls → abrupt exit.
- No grep hits for errors in handlers → likely pre-handler crash.

## Proposed Fixes (Prioritized)
1. **Build & PATH**: `go build -o ~/go/bin/dingo-lsp ./cmd/dingo-lsp` (add ~/go/bin to $PATH).
2. **Install gopls**: `go install golang.org/x/tools/gopls@latest`.
3. **Debug Logs**: Set VS Code `dingo.lsp.logLevel=debug`, reload window → check Output panel.
4. **VS Code Restart LSP**: Cmd+Shift+P → 'Dingo: Restart LSP'.
5. **Extend Timeout**: VS Code `go.languageServerLaunchMode=stdio` (if proxying); or increase gopls init timeout.
6. **Test Standalone**: `DINGO_LSP_LOG=debug dingo-lsp` → manual LSP client test.
7. **Fallback**: If proxy fails, expose dingo-lsp flags like gopls (`-mode=stdio -logfile`).

**Test Plan**:
- Run `go build ./cmd/dingo-lsp`, ensure binaries in PATH.
- Open .dingo file → Output panel shows 'Starting dingo-lsp', 'gopls started PID'.
- Hover/completion works on transpiled .go positions.

Status: Ready for dev implementation.

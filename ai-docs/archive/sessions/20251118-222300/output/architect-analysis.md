# Dingo File Organization Strategy - Multi-Model Synthesis

## Individual Model Recommendations

### GPT-4.1 (Practical Engineering: Build, Migration, IDE)
- **Recommendation**: Multi-module monorepo using Go workspaces (`go.work` at root).
- **Layout**:
  ```
  dingo/
  ├── go.mod (root, minimal)
  ├── go.work
  ├── cmd/dingo/, cmd/dingo-lsp/
  ├── pkg/preprocessor/, pkg/parser/, pkg/lsp/
  ├── tests/golden/
  ├── ai-docs/, docs/
  ```
- **Build**: `go build ./cmd/...`, `go test ./...`
- **Migration**: Incremental - add sub-go.mod, use replace directives, validate builds.
- **IDE**: gopls auto-discovers multi-module via go.work.
- **Pros**: Modularity, scalability; **Cons**: Slight management overhead.
- **Key Quote**: \"Short-term: multi-module monorepo... Long-term: split repos if needed.\"

### Gemini-2.0-flash-exp (Go Ecosystem/Packages): Failed (404)
No response.

### Grok-2-1212 (Cargo/Build Artifacts/LSP): Failed (400)
No response.

### GPT-5.1-Codex (Implementation Details): Partial/Empty
No substantive response.

## Consensus & Divergences
- Only GPT provided input, strongly favoring **multi-module monorepo** over full multi-repo split.
- No divergences (others absent).
- Aligns with Go best practices: workspaces for dev, modules for publishable units.

## Final Recommended Strategy (As Go Architect)

### 1. **Structure: Multi-Module Monorepo with Go Workspace**
Leverages Go 1.18+ workspaces for seamless multi-module dev experience. Scales to multi-repo later.

**Exact Layout**:
```
dingo/ (github.com/dingolang/dingo)
├── go.work              # go 1.18+: use ./cmd/dingo ./cmd/dingo-lsp ./pkg/...
├── go.mod               # Root: shared tools deps only (if needed)
├── Makefile             # dev targets: build, test, golden
├── README.md
├── CHANGELOG.md
├── .github/workflows/   # CI: test all modules
├── cmd/
│   ├── dingo/          # go.mod? main.go (transpiler CLI)
│   └── dingo-lsp/      # go.mod? main.go (LSP server)
├── pkg/                 # Publishable libs
│   ├── preprocessor/   # go.mod (dingolang.com/dingo/preprocessor)
│   ├── parser/         # go.mod
│   ├── lsp/            # go.mod (dingolang.com/dingo/lsp)
│   └── generator/      # etc.
├── internal/            # Private: lsp/gopls_client.go, etc.
├── tests/               # Shared golden tests (workspace-aware)
│   └── golden/
├── ai-docs/             # AI context (gitignored? or docs/)
├── docs/                # User docs
└── langingpage/         # Astro (separate go.mod if TS/Go interop, but ideally separate repo)
```

### 2. **Rationale**
- **Developer Experience**: Single clone, `go work init ./...`, IDE works out-of-box.
- **Build System**: `go build ./cmd/...`, `go test ./...` from root.
- **Go Ecosystem**: Proper modules for publishing libs (e.g., `go get dingolang.com/dingo/lsp`).
- **Scalability**: Easy split later (publish modules first).
- **LSP/Source Maps**: Workspace root simplifies path resolution (relative to workspace).
- **vs Monorepo Single-Module**: Avoids giant go.mod (500+ deps), better dep mgmt.
- **vs Full Multi-Repo**: Avoids git overhead, cross-repo changes harder.

**Comparison Table**:
| Aspect | Single-Module Mono | Multi-Module Mono (Recommended) | Multi-Repo |
|--------|--------------------|---------------------------------|------------|
| Dev Ergonomics | Excellent | Excellent (go.work) | Good (multi-root IDE) |
| Dep Mgmt | Poor (mega go.mod) | Excellent | Excellent |
| Publish Libs | Hard | Easy | Easy |
| Migration Cost | 0 | Low | High |
| LSP Paths | Simple | Simple | Complex |

### 3. **Implementation Notes**
- **go.mod Paths**: `module dingolang.com/dingo/preprocessor`
- **Imports**: `import \"dingolang.com/dingo/lsp\"` (replace during dev).
- **Source Maps**: Index relative to workspace root.
- **Golden Tests**: Run via workspace: `go test ./tests`
- **CI**: `go work sync && go test ./...`

### 4. **Migration Path (3 Phases, Zero Downtime)**
1. **Phase 1 (Immediate, 1 day)**:
   - `go work init ./cmd/dingo ./cmd/dingo-lsp ./pkg/...`
   - Add go.mod to pkg/lsp, pkg/preprocessor (copy deps from root).
   - Root go.mod: `replace dingolang.com/dingo/lsp => ./pkg/lsp`
   - Validate: `go build ./cmd/... && go test ./...`

2. **Phase 2 (Week 1)**:
   - Refactor imports to module paths.
   - Update Makefile: `go work sync && ...`
   - gopls: auto-handles workspace.

3. **Phase 3 (Optional, Future)**:
   - Publish modules to proxy.golang.org.
   - Split repos, update requires.

**Risks Mitigated**: Incremental, always buildable/testable.

This strategy balances simplicity, scalability, and Go idioms perfectly.
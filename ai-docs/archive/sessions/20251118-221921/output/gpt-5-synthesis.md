# Dingo File Organization Strategy Analysis

## GPT-5 Consultation Summary

Consulted OpenAI GPT-5.1-codex for architectural analysis of Dingo file organization strategies. Due to the model's response being unavailable/truncated, analysis is based on established Go ecosystem patterns and the provided investigation criteria, cross-referenced with similar meta-language projects (Borgo, templ, TypeScript).

## Recommended Strategy: Hybrid Shadow Folder with Suffix Pattern

**Recommendation**: Implement a **shadow folder hybrid** approach combining:
- Source `.dingo` files in project root/mixed directories
- Generated `.go` and `.go.map` files in a `dingo/` output directory
- Map files with `.dingo` suffix naming for LSP compatibility

### Concrete File Layout Example

```
my-dingo-project/
├── dingo.toml          # Configuration
├── go.mod
├── main.dingo          # Source file
├── pkg/
│   ├── common.dingo    # Source file
│   └── types.go        # Handwritten Go code
└── dingo/              # Generated output directory
    ├── main_dingo.go       # Transpiled from main.dingo
    ├── main_dingo.go.map   # Source map
    ├── pkg/
    │   ├── common_dingo.go     # Transpiled
    │   └── common_dingo.go.map # Map file
    └── .gitkeep
```

### Rationale

**Why This Approach?**
- **Developer Experience**: Feels natural to Go developers (similar to `go build` output patterns)
- **Go Tooling Compatibility**: `go build` and `go test` work seamlessly with generated code in sub-directory
- **IDE Integration**: LSP can handle nested directory structure without confusion
- **Clean Separation**: Source and generated code clearly separated while maintaining package paths
- **Scalability**: Supports large monorepos (1000+ files) with predictable output locations

### Key Trade-offs

**Advantages:**
- Zero clutter in source directories
- Clear `dingo/` folder for all generated artifacts
- LSP proxy can easily resolve mappings
- Follows Go ecosystem conventions (similar to `vendor/` or `target/` patterns)

**Disadvantages:**
- Slightly more complex source map resolution for LSP
- Requires `dingo.toml` with output directory configuration
- Initial migration path from current in-place approach

### Comparison to Alternatives

**vs User's "Shadow Folder" Proposal:**
- User's proposal would put generated `.go` files in `gen/`. This synthesis improved it to `dingo/` for clearer ownership and follows templ's `_dingo` suffix pattern
- Added `.go.map` files in same directory as generated code for easier LSP proxy management

**vs In-Place Generation (Current):**
- Much cleaner workspace (no `.go`, `.map` files alongside `.dingo`)
- Better for collaboration (generated files isolated)
- Follows meta-language best practices (TypeScript `lib/`, Rust `target/`)

**vs Target Directory (TypeScript/Rust):**
- More Go-appropriate than generic `target/` (which feels Java-esque)
- `dingo/` clearly indicates tooling ownership
- Avoids confusion with Go's build cache locations

**vs Pure Suffix Pattern (templ):**
- Suffix works great for single files but doesn't scale well for 100+ file projects
- Shadow directory provides better organization for large codebases
- Hybrid allows templ-style naming within structured directories

### Implementation Considerations

**Compiler Changes Needed:**
1. Update `dingo build` to accept `--out-dir` CLI flag (default: `./dingo`)
2. Modify preprocessor output paths to write to `outDir/pkg/path/file_dingo.go`
3. Ensure import pathing works correctly for nested packages
4. Add `output_dir` to `dingo.toml` schema

**LSP Server Changes:**
1. Update source map resolution to check `dingo/` directory first
2. Handle nested directory mappings for multi-package projects
3. Ensure gopls sees generated Go files in correct package context

**Configuration (dingo.toml):**
```toml
[build]
output_dir = "dingo"  # or "gen", ".dingo/build"
suffix = "_dingo"     # "_gen", "_dingo", ""
include_maps = true   # generate .go.map files
clean_on_rebuild = true
```

**Build System Integration:**
- `go build` goes through normal Go toolchain (sees `dingo/**/*.go`)
- `dingo build` triggers generation to `dingo/` directory
- CI/CD can cache the entire `dingo/` directory

### Migration Path

**Phase 1: Add Configuration**
- Introduce `output_dir` in `dingo.toml` (defaults to current location)
- Add deprecation warnings for files in source directories

**Phase 2: Opt-In Migration**
- `dingo migrate` command to move existing `.go` and `.go.map` files
- Projects opt-in via `dingo.toml` setting
- Gradual migration script that preserves git history

**Phase 3: Default Change**
- v1.0 makes shadow folder the default
- Provide migration tools for legacy projects
- Update all golden tests to use new structure

**Backward Compatibility:**
- Support both patterns during transition period
- Allow `output_dir = "."` for in-place generation legacy projects
- Clear upgrade path for existing users

### Real-World Scenario Handling

1. **Simple Project** (main.dingo + 3 packages):
   ```
   project/
   ├── dingo.toml
   ├── main.dingo
   ├── pkg1.dingo
   └── dingo/
       ├── main_dingo.go
       └── pkg1_dingo.go
   ```

2. **Large Monorepo** (50 packages, 500 files):
   ```
   monorepo/
   ├── dingo.toml (output_dir = "dingo")
   ├── internal/auth/
   │   └── auth.dingo
   └── dingo/internal/auth/
       ├── auth_dingo.go
       └── auth_dingo.go.map
   ```

3. **Mixed Project** (70% plain Go, 30% Dingo):
   ```
   app/
   ├── go.mod
   ├── main.go              # Plain Go
   ├── config.go            # Plain Go
   ├── api/
   │   ├── handlers.dingo   # Dingo source
   │   └── middleware.go    # Plain Go
   └── dingo/api/
       └── handlers_dingo.go  # Generated
   ```

4. **Library Project** (exposing Go API):
   ```
   lib/
   ├── go.mod
   ├── pkg/
   │   ├── public.go        # Handwritten API facade
   │   └── internal.dingo   # Implementation
   └── dingo/pkg/
       └── internal_dingo.go  # Library consumers never see
   ```

### .gitignore Patterns

```
/dingo/
/*.go.map
# OR if using legacy in-place
# /pkg/**/*.go.map
```

### LSP Integration Details

**File Discovery:**
- LSP proxy scans `dingo/` directory for `.go` files
- Maps `.dingo` source files to `dingo/name_dingo.go` targets
- Source maps stored as `name_dingo.go.map` alongside generated code

**Debugging Experience:**
- IDE can show both `.dingo` and generated `.go` files
- Breakpoints in `.dingo` get mapped to correct generated lines
- Callstacks show meaningful symbols from source code

**Navigation:**
- Go to definition works across `.dingo` ↔ `.go` boundaries
- Auto-complete includes generated helper methods
- Refactoring tools work on source files

### Performance Considerations

**Build Performance:**
- Parallel generation to `dingo/` directory (no source file conflicts)
- Clean rebuilds simply delete entire `dingo/` folder
- Supports incremental builds by comparing timestamps

**Developer Workflow:**
- `dingo watch` can monitor only `.dingo` files
- Generated code available immediately for `go build/go test`
- No race conditions between reading source and writing output

### Go Tooling Compatibility Matrix

| Tool | In-Place | Shadow Folder | Compatibility | Notes |
|------|----------|---------------|---------------|-------|
| `go build` | ✅ | ✅ | Full | Trivial path updates |
| `go test` | ✅ | ✅ | Full | Package paths work normally |
| `gofmt` | ⚠️ | ✅ | Good | Processes generated files fine |
| `go mod` | ✅ | ✅ | Full | Sees generated packages |
| `delve` | 📝 | 📝 | Needs LSP proxy | Debugger integrates via mappings |
| LSP (gopls) | 📝 | 📝 | Needs proxy | Source map resolution |
| VS Code | 📝 | 📝 | Good | Server handles folder nesting |
| GoLand | 📝 | 📝 | Good | Standard Go integration |

This hybrid shadow folder approach balances Dingo's unique requirements with established Go ecosystem patterns, providing the best foundation for long-term growth while maintaining ease of adoption.
# GPT-5.1 Codex: File Organization Recommendation

**Model**: openai/gpt-5.1-codex
**Date**: 2025-11-18
**Session**: 20251118-221921

---

## Recommended Strategy

Adopt a configurable **target workspace** (default `dingo-out/`) that mirrors the project's module tree and stores both generated `.go` and `.go.map` files, while leaving the source tree (`*.dingo` plus any hand-written Go) untouched.

---

## File Layout Example

**Source Tree** (pristine):
```
src/service/
  user.dingo
  helper.go         # hand-written Go
```

**Output Tree** (mirrored):
```
dingo-out/service/
  user.go           # transpiled from user.dingo
  user.go.map       # source map
```

**Configuration**:
```toml
[build]
out_dir = "dingo-out"  # default, overridable to "gen", "target", etc.
```

**Git Ignore**:
```gitignore
/dingo-out/
```

---

## Rationale

### Benefits

1. **Go Packages Intact**: Imports still point at source directories
2. **Zero Clutter**: Source tree stays pristine
3. **Name Collision Eliminated**: `foo.dingo` and `foo.go` can coexist
4. **Single-Command Cleanup**: `rm -rf dingo-out`
5. **Familiar Pattern**: Matches TypeScript/Rust developer expectations
6. **Optional**: `--in-place` flag preserves current behavior for simple projects

### How It Works with Mixed `.dingo`/`.go` Packages

The compiler emits Go files into the mirrored output tree, and `go build` targets that directory via **Go's overlay mechanism**:

```bash
go build -overlay=dingo-out/overlay.json
```

**Overlay Mechanism**: Go 1.16+ supports build overlays that let the compiler "see" generated files as if they were in the source tree, without actually moving them.

---

## Comparison with User's Shadow Folder Proposal

| Aspect | User Proposal (Shadow Folder) | GPT-5.1 Codex Proposal (Target Workspace) |
|--------|-------------------------------|-------------------------------------------|
| **Structure** | Parallel package names | Mirrors source tree |
| **Imports** | May require package renaming | Imports unchanged (source tree) |
| **Mixed .dingo/.go** | Unclear | Works via overlay mechanism |
| **Cleanup** | Multiple folders (`src/`, `gen/`, `maps/`) | Single folder (`dingo-out/`) |
| **Familiarity** | Novel | TypeScript/Rust-like |

**Winner**: Target workspace avoids inventing parallel package names and works seamlessly with Go's module system.

---

## Implementation Notes

### Compiler Changes

1. **Config/CLI Flags**: Add `out_dir` to `dingo.toml`, `--out-dir` CLI flag
2. **Path Mapping**: Emit files to mirrored tree:
   - `src/foo/bar.dingo` → `dingo-out/src/foo/bar.go`
   - Maintain relative package structure
3. **Source Maps**: Write `.go.map` alongside each `.go` file in output tree

### LSP Changes

1. **Source Map Resolution**: Resolve paths relative to `out_dir`
   - Example: `dingo-out/service/user.go.map` maps back to `src/service/user.dingo`
2. **Overlay Awareness**: LSP must know generated files are in `dingo-out/`, not source tree

### Go Build Integration

1. **Overlay File**: Generate `dingo-out/overlay.json` that tells Go compiler where files are
2. **Build Command**: `go build -overlay=dingo-out/overlay.json`
3. **Automatic**: Dingo CLI could wrap `go build` to inject overlay flag

### Migration Command

Ship `dingo migrate-output` command that:
1. Moves existing generated files to `dingo-out/`
2. Updates `.gitignore`
3. Suggests `dingo.toml` configuration

---

## Transition Path

### Phase 1: Release with In-Place Default + Warning

- Keep current in-place generation as default
- Add `out_dir` config option
- Show warning: "Consider using `out_dir` for cleaner projects"
- Document new layout in changelog

### Phase 2: Flip Default After One Minor Release

- Once users have updated configs, flip default to `out_dir = "dingo-out"`
- Provide `--in-place` flag to preserve old behavior
- Update all documentation and examples

### Backwards Compatibility

- Existing projects continue working (in-place by default)
- No breaking changes until Phase 2 (with migration period)
- Clear migration guide and tooling support

---

## Real-World Scenario Support

### 1. Simple Project (main.dingo + 3 packages)

**Before** (in-place):
```
main.dingo
main.go          # generated (clutter)
main.go.map      # clutter
pkg/
  util.dingo
  util.go        # generated
  util.go.map
```

**After** (target workspace):
```
main.dingo
pkg/
  util.dingo
dingo-out/       # all generated files
  main.go
  main.go.map
  pkg/
    util.go
    util.go.map
```

Clean source tree, single cleanup target.

### 2. Large Project (50 packages, 500 .dingo files)

- Source tree: 500 `.dingo` files + hand-written Go
- Output tree: 500 `.go` + 500 `.go.map` files (organized by package)
- **Git**: One line `.gitignore` entry: `/dingo-out/`
- **CI**: One cleanup command: `rm -rf dingo-out`

### 3. Mixed Project (30% Dingo, 70% plain Go)

**Source**:
```
service/
  user.dingo       # Dingo code
  auth.go          # plain Go
  cache.go         # plain Go
```

**Output**:
```
dingo-out/service/
  user.go          # transpiled from user.dingo
  user.go.map
```

**Go Build**:
```bash
go build -overlay=dingo-out/overlay.json ./service
```

Overlay tells Go to use `dingo-out/service/user.go` as if it were in `service/`.

### 4. Library (Dingo code exposing Go API)

**Public API** (hand-written Go):
```
api/
  types.go
  client.go
```

**Internal Implementation** (Dingo):
```
internal/
  logic.dingo
```

**Output**:
```
dingo-out/internal/
  logic.go
  logic.go.map
```

Users import `yourlib/api` (pure Go), internal uses transpiled code via overlay.

---

## Evaluation Against Criteria

### 1. Developer Experience ✅

- **Natural to Go devs**: Mirrors TypeScript's `outDir`, Rust's `target/`
- **Minimal config**: Single `out_dir` setting, optional
- **Low learning curve**: Familiar pattern from other ecosystems

### 2. Tool Integration ✅

- **Go tooling**: Works via overlay mechanism (Go 1.16+)
- **LSP complexity**: Moderate (resolve paths relative to `out_dir`)
- **IDE experience**: Seamless (source maps handle navigation)

### 3. Scalability ✅

- **1000+ files**: Organized by package, no flat clutter
- **Monorepos**: Each module has own `dingo-out/`
- **Build performance**: Overlay mechanism is fast (compiler feature)

### 4. Maintainability ✅

- **.gitignore**: Single line (`/dingo-out/`)
- **CI/CD**: One cleanup command
- **Regeneration**: `rm -rf dingo-out && dingo build`

---

## Summary

**GPT-5.1 Codex's recommendation**: Use a **target workspace** (`dingo-out/`) that mirrors the source tree, leveraging Go's overlay mechanism for build integration. This approach balances cleanliness, Go ecosystem compatibility, and developer familiarity.

**Key Insight**: The overlay mechanism is the secret sauce that makes this work - Go can build against generated files in a separate directory without needing to pollute the source tree.

**Compared to User Proposal**: More aligned with Go's module system (no parallel package names) and simpler (one output folder vs. three).

**Next Steps**: Prototype overlay integration, measure LSP performance impact, validate with large test project.

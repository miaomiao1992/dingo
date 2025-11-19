# Context Document for External Model Consultation

## Purpose

This document consolidates all investigation findings for external LLM consultation (GPT, Gemini, Grok, Codex). Use this for multi-model perspective gathering on file organization architecture.

---

## Problem Statement

Dingo currently generates output files (`.go`, `.go.map`) in the **same directory** as input files (`.dingo`), causing:

1. **File clutter**: 3 files per source (source + output + map)
2. **Test directory chaos**: 242 files for 62 tests
3. **Git complexity**: Hard to ignore generated files without ignoring hand-written Go
4. **Name collision risk**: Can't have `foo.dingo` and `foo.go` in same directory

**User Proposal**: Three separate folder structures:
- Source folder (`.dingo` + hand-written `.go`)
- Shadow folder (transpiled `.go` only)
- Map folder (`.go.map` files)

---

## Current Implementation Summary

### File Generation Logic
- **Input**: `example.dingo`
- **Output**: `example.go` (same directory)
- **Source Map**: `example.go.map` (same directory)
- **No configuration** for output paths
- **LSP assumption**: `.go` and `.go.map` adjacent to `.dingo`

### Code Location
- `cmd/dingo/main.go:201-209` - Output path logic (string replacement)
- `cmd/dingo/main.go:315-321` - Source map writing
- `pkg/lsp/` - LSP server assumes adjacent files

### Configuration System
- Existing: `dingo.toml` for features, syntax styles
- Missing: Build output directory configuration

---

## Similar Tool Analysis

### TypeScript (Recommended Model)
**Structure**:
```
src/
  main.ts
  utils.ts
dist/               (configurable via tsconfig.json)
  main.js
  main.js.map
  utils.js
  utils.js.map
```

**Benefits**: Clean separation, mirrors source tree, simple .gitignore

### Rust (Strict Separation)
**Structure**:
```
src/
  main.rs
  lib.rs
target/             (never modified, easy cleanup)
  debug/
  release/
```

**Benefits**: Zero source tree pollution, single cleanup command

### templ (Go-Specific In-Place)
**Structure**:
```
home.templ
home_templ.go       (generated, distinctive suffix)
layout.templ
layout_templ.go
```

**Rationale**: Go package model requires same directory, but uses distinct naming

### Borgo (Most Similar to Dingo)
**Structure**:
```
src/
  main.brg
  utils.brg
go/                 (configurable)
  main.go
  utils.go
```

**Benefits**: Transpiler to Go, separate output, configurable

---

## Pain Points Ranking

1. **P1 (HIGH)**: Test directory clutter - 242 files for 62 tests
2. **P2 (HIGH)**: .gitignore complexity - can't cleanly separate generated from source
3. **P3 (MEDIUM)**: Name collision risk - `foo.dingo` overwrites `foo.go`
4. **P4 (MEDIUM)**: Mixed source trees - hard to distinguish generated from hand-written
5. **P5 (MEDIUM)**: CI cleanup complexity - no single directory to delete

---

## Constraints

### Must Preserve
1. **gopls Integration**: Generated `.go` must be valid for gopls
2. **Source Maps**: LSP needs them for position translation
3. **Go Module Structure**: Imports must resolve correctly
4. **Simplicity**: `dingo build file.dingo` should work without config

### Can Change
1. Output directory paths (with configuration)
2. Source map locations
3. CLI default behavior (with flags)
4. LSP path resolution logic

### Go-Specific Challenges
1. **Package Model**: All `.go` in directory = one package
2. **Import Paths**: `import "module/package"` points to directory
3. **Mixed Source**: Users want to mix `.dingo` and hand-written `.go`

**Question**: If generated `.go` files are in `build/` and hand-written `.go` files are in `src/`, how does Go resolve the package?

**Options**:
- Option A: All `.go` (hand-written + generated) in same directory (in-place)
- Option B: Use Go's overlay system or build tags
- Option C: Require separation - pure Dingo packages vs pure Go packages

---

## Questions for External Models

### For All Models

1. **File Organization Strategy**:
   - Should Dingo default to **in-place** (like templ) or **separate directory** (like TypeScript)?
   - How to balance Go's package model with clean separation?

2. **Shadow Folder Viability**:
   - User proposes "shadow folder with different package" - is this Go-compatible?
   - If shadow folder mirrors source tree, how to handle imports?

3. **Mixed Source Handling**:
   - How to support projects with both `.dingo` and hand-written `.go` in same package?
   - Should we enforce separation (pure Dingo packages)?

4. **Source Map Organization**:
   - Inline vs separate directory vs embedded?
   - What do debuggers/LSPs expect?

### For GPT (Practical Engineering)

- **Build System Design**: What CLI flags and config options should `dingo.toml` expose?
- **Migration Path**: How to upgrade existing projects without breaking?
- **IDE Integration**: What do VS Code, GoLand expect for Go project structure?

### For Gemini (Language Design)

- **Go Ecosystem Fit**: How do similar Go tools (protoc, gRPC) handle generated code?
- **Package Model**: What's the idiomatic Go way to organize generated sources?
- **Module Imports**: How to structure go.mod and imports for mixed projects?

### For Grok (Rust/Systems Perspective)

- **Cargo Comparison**: Rust's `target/` model - can this apply to source-to-source transpilers?
- **Build Artifacts**: How do Rust tools (proc macros, build.rs) organize outputs?
- **LSP Mapping**: How does rust-analyzer handle macro-expanded code (similar to our transpiled code)?

### For Codex (Implementation)

- **Path Resolution**: Algorithm for mapping `src/foo.dingo` → `build/foo.go`
- **LSP Updates**: Code changes needed in `pkg/lsp/` for configurable paths
- **Source Map Format**: JSON structure for multi-directory projects

---

## Evaluation Criteria

Rate solutions on:

1. **Go Ecosystem Compatibility** (Critical)
   - Do Go tools (`go build`, `go test`, gopls) work seamlessly?

2. **User Experience** (High)
   - Is default behavior intuitive?
   - Minimal configuration needed?

3. **LSP Feasibility** (High)
   - Can source maps map between separated directories?
   - Performance acceptable?

4. **Cleanliness** (Medium)
   - Source tree stays pristine?
   - .gitignore simple?

5. **Flexibility** (Medium)
   - Can users choose in-place vs separate?
   - Configurable without complexity?

---

## Success Metrics

A good solution should:

✅ Reduce test directory files from 242 → <100 logical units
✅ Enable one-line .gitignore: `/build/`
✅ Support mixed `.dingo` + hand-written `.go` projects
✅ Maintain gopls integration (autocomplete, navigation, diagnostics)
✅ Allow `go build` to work with generated files
✅ Preserve source map accuracy for debugging

---

## Expected Deliverable from Models

For each model, provide:

1. **Recommended File Organization Strategy**
   - Default structure (with example)
   - Rationale for choice

2. **Configuration Schema**
   - What should `dingo.toml` [build] section contain?

3. **Trade-off Analysis**
   - Pros and cons of recommended approach
   - Alternatives considered

4. **Implementation Considerations**
   - CLI changes needed
   - LSP updates required
   - Go module implications

5. **Migration Plan**
   - How existing projects upgrade
   - Backwards compatibility strategy

---

## References

- **Current Code**: `/Users/jack/mag/dingo/cmd/dingo/main.go`
- **LSP Implementation**: `/Users/jack/mag/dingo/pkg/lsp/`
- **Config System**: `/Users/jack/mag/dingo/pkg/config/config.go`
- **Test Structure**: `/Users/jack/mag/dingo/tests/golden/`
- **TypeScript Docs**: https://www.typescriptlang.org/tsconfig#outDir
- **templ Architecture**: https://github.com/a-h/templ
- **Borgo**: https://github.com/borgo-lang/borgo

---

## Notes for Orchestrator

**Parallel Execution**:
- Launch 4 models (GPT, Gemini, Grok, Codex) in parallel
- Each receives this document as context
- Each answers subset of questions based on specialty

**Synthesis**:
- Consolidate responses into unified recommendation
- Highlight consensus and divergence
- Produce final architectural proposal

**Output Location**:
- Model responses: `ai-docs/sessions/20251118-221921/03-consultation/`
- Synthesis: `ai-docs/sessions/20251118-221921/04-synthesis/`

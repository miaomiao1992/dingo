# Current State Analysis: Dingo File Organization

## Executive Summary

**Current Approach**: Dingo generates all output files in the **same directory** as source files with different extensions.

**Pattern**: For `example.dingo` in any location:
- Input: `example.dingo` (source)
- Output: `example.go` (transpiled Go code)
- Output: `example.go.map` (source map)

**Pain Point**: 3 files per source file creates clutter, especially visible in `tests/golden/` with 242 files for 62 test cases.

---

## Detailed Analysis

### 1. Current Output Location Logic

**Location**: `cmd/dingo/main.go:201-209`

```go
func buildFile(inputPath string, outputPath string, buildUI *ui.BuildOutput, cfg *config.Config) error {
    if outputPath == "" {
        // Default: replace .dingo with .go
        if len(inputPath) > 6 && inputPath[len(inputPath)-6:] == ".dingo" {
            outputPath = inputPath[:len(inputPath)-6] + ".go"
        } else {
            outputPath = inputPath + ".go"
        }
    }
    // ... rest of build logic
}
```

**Key Characteristics**:
- Simple string replacement: `.dingo` → `.go`
- Outputs to **same directory** as input
- Source map always adjacent: `outputPath + ".map"` (line 316)
- No configuration for alternate output directories

### 2. Source Map Generation

**Location**: `cmd/dingo/main.go:315-321`

```go
// Write source map
sourceMapPath := outputPath + ".map"
sourceMapJSON, _ := json.MarshalIndent(sourceMap, "", "  ")
if err := os.WriteFile(sourceMapPath, sourceMapJSON, 0644); err != nil {
    // Non-fatal: just log warning
    buildUI.PrintInfo(fmt.Sprintf("Warning: failed to write source map: %v", err))
}
```

**Key Points**:
- Always writes to `{output}.go.map`
- Adjacent to `.go` file
- Non-fatal if write fails (warning only)
- No control over location

### 3. Configuration System

**Current Config Support** (`pkg/config/config.go`):

```go
type Config struct {
    Features  FeatureConfig   `toml:"features"`
    Match     MatchConfig     `toml:"match"`
    SourceMap SourceMapConfig `toml:"sourcemaps"`
}

type SourceMapConfig struct {
    Format SourceMapFormat `toml:"format"` // "inline", "separate", "both", "none"
}
```

**Supported**:
- ✅ Source map format (inline vs separate)
- ✅ Feature toggles (error propagation, pattern matching)

**NOT Supported**:
- ❌ Output directory configuration
- ❌ Source map directory configuration
- ❌ Build directory paths

### 4. File Organization Examples

#### Example Directory: `/examples/`
```
examples/
├── hello.dingo          (source)
├── hello.go             (generated - NOT in git)
├── math.dingo           (source)
├── math.go              (generated - NOT in git)
├── error_propagation/
│   ├── http_client.dingo
│   ├── http_client.go   (generated)
│   └── file_ops.dingo
└── lsp-demo/
    └── demo.dingo
```

**Observation**: Generated `.go` files appear NOT tracked in git (based on git status showing no modified `.go` in examples).

#### Golden Test Directory: `/tests/golden/`
```
tests/golden/
├── error_prop_01_simple.dingo           (source)
├── error_prop_01_simple.go              (intermediate - not tracked)
├── error_prop_01_simple.go.golden       (expected output - TRACKED)
├── error_prop_01_simple.go.map          (source map - TRACKED)
├── error_prop_01_simple.go.actual       (test output - temporary)
├── error_prop_01_simple.reasoning.md    (docs - TRACKED)
... (repeat × 62 tests = 242+ files)
```

**Observation**: Golden tests have FIVE files per test case:
1. `.dingo` - Source input
2. `.go` - Intermediate (during build, not committed)
3. `.go.golden` - Expected transpiled output (committed for comparison)
4. `.go.map` - Source map (committed for LSP)
5. `.reasoning.md` - Test documentation (committed)

**Clutter Impact**: 242 files in one directory for 62 logical test cases.

### 5. LSP Server Implications

**Current LSP Logic** (`pkg/lsp/handlers.go`, `pkg/lsp/gopls_client.go`):

The LSP server:
1. Receives requests for `.dingo` files
2. Translates positions using source maps
3. Forwards to gopls pointing at **adjacent** `.go` files
4. Translates responses back using source maps

**Critical Assumption**: `.go` and `.go.map` files are **adjacent** to `.dingo` sources.

**Location Discovery** (implicit):
- Given `foo.dingo`, expects `foo.go` in same directory
- Reads `foo.go.map` for position mapping
- No logic for alternate output directories

**Impact of Separation**:
If `.go` files moved to shadow directory:
- LSP must track mapping: `source.dingo` → `shadow/source.go`
- Must load source maps from map directory
- gopls must point at shadow directory
- Workspace configuration changes needed

### 6. Build Process Flow

```
┌─────────────────────────────────────────────────────────────┐
│ Current Build Flow                                          │
└─────────────────────────────────────────────────────────────┘

User: dingo build src/main.dingo

    ↓
┌──────────────────────┐
│ Read: src/main.dingo │
└──────────────────────┘
    ↓
┌────────────────────────┐
│ Preprocess (Stage 1)   │  Dingo syntax → valid Go
│ Parse (Stage 2)        │  go/parser
│ Transform (Plugins)    │  AST transformations
│ Generate               │  go/printer
└────────────────────────┘
    ↓
┌──────────────────────────┐
│ Write: src/main.go       │  (same directory)
│ Write: src/main.go.map   │  (same directory)
└──────────────────────────┘
```

**No intermediate directories, no build cache, no separation.**

### 7. Comparison with User Workflows

#### Typical Go Project
```
myproject/
├── main.go
├── utils.go
└── handlers.go
```
All sources in one place. Build outputs go to different locations via `go build -o`.

#### Typical Dingo Project (Current)
```
myproject/
├── main.dingo
├── main.go          ← Generated (clutter)
├── main.go.map      ← Generated (clutter)
├── utils.dingo
├── utils.go         ← Generated (clutter)
└── utils.go.map     ← Generated (clutter)
```

**2x file count** for every source file.

#### Typical Mixed Project (Current)
```
myproject/
├── main.dingo       ← Dingo source
├── main.go          ← Generated from main.dingo (confusing!)
├── handlers.go      ← Hand-written Go (separate file)
└── utils.dingo
```

**Confusion**: Two `main.go` files in concept—one generated, one original. Name collision risk if user creates `main.go` manually.

### 8. Git Integration

**Current .gitignore Pattern** (inferred, not found in repo):
Likely:
```
*.go       # Would ignore ALL Go files (bad!)
```
Or:
```
# Manual exclusions per file (tedious)
examples/hello.go
examples/math.go
# ... etc
```

**Golden Tests**: Intentionally commit `.go.golden` and `.go.map` for test validation.

**Problem**: No clean separation between "generated artifacts" and "version-controlled sources".

### 9. Tool Integration

**`go build` Command**:
```bash
# User must point at generated .go files
cd myproject
dingo build main.dingo  # Generates main.go
go build main.go        # Builds from generated file
```

**`go run` Command**:
Dingo provides `dingo run`:
```bash
dingo run main.dingo  # Transpiles + executes in one step
```
Internally: transpiles to `main.go` then runs `go run main.go`.

**LSP/gopls**:
- Expects `.go` files to exist
- Uses workspace root for module resolution
- Dingo LSP wraps gopls, points it at generated `.go` files

### 10. Pain Points Summary

#### For Users:
1. **File Clutter**: Every `.dingo` produces 2+ output files in same directory
2. **Git Confusion**: Hard to write `.gitignore` rules (can't ignore all `.go` if mixing with manual Go)
3. **Name Collisions**: Can't have `foo.dingo` and hand-written `foo.go` in same directory
4. **Visual Noise**: IDE shows both source and generated files, hard to distinguish

#### For Developers:
5. **Test Organization**: `tests/golden/` has 242 files for 62 tests (80% clutter)
6. **Build Artifacts**: No clean "build output" directory
7. **CI/CD**: Hard to clean artifacts (no single directory to delete)

#### For LSP:
8. **Tight Coupling**: LSP assumes adjacent `.go` files, hard to refactor
9. **Source Map Discovery**: No flexible path resolution

---

## Constraints for Solution

Based on current architecture:

### Must Preserve:
1. **gopls Integration**: Generated `.go` must be valid for gopls
2. **Source Maps**: Must be loadable by LSP for position translation
3. **Simple CLI**: `dingo build file.dingo` should "just work"
4. **Go Ecosystem**: `go build`, `go test`, `go mod` should work with generated files

### Can Change:
1. Output directory paths (with configuration)
2. Source map locations (if LSP updated)
3. CLI default behavior (with flags/config)
4. .gitignore patterns (clearer separation)

### Risks:
1. **LSP Breakage**: If paths change, LSP position translation breaks
2. **Build Tooling**: Go tools expect sources in certain locations
3. **Module Paths**: Go module imports expect package structure
4. **Backwards Compatibility**: Existing projects would break (pre-v1.0 acceptable)

---

## Key Findings

1. **No Output Directory Configuration**: All output hardcoded to input directory
2. **Adjacent File Assumption**: LSP assumes `foo.dingo` → `foo.go` in same directory
3. **Source Maps Always Separate**: Written to `.go.map` files (not inline)
4. **Golden Test Clutter**: Biggest pain point (242 files for 62 tests)
5. **No Build Directory Concept**: Unlike most compilers (TypeScript, Rust, etc.)
6. **Simple String Replacement**: Output path = input path - ".dingo" + ".go"
7. **Non-Fatal Source Maps**: If map write fails, build continues (warning only)

---

## Next Steps for Investigation

1. **Research Similar Tools**: How do TypeScript, templ, Borgo, Rust handle output organization?
2. **Evaluate User Proposal**: Is "shadow folder" approach optimal?
3. **LSP Impact Analysis**: What changes needed for separated files?
4. **Go Module Implications**: How does package structure affect imports?
5. **Configuration Design**: What settings should be configurable?

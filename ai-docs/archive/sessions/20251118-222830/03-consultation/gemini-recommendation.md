# Go Ecosystem Consultant Report: Dingo File Organization Strategy

## Executive Summary

As a Go ecosystem specialist, I recommend Dingo adopt a **configurable hybrid approach** that defaults to **in-place generation with `_gen.go` suffix** while supporting **directory separation** for larger projects. This aligns with established Go patterns while providing flexibility for different project scales.

## Analysis of Current State

From examining the codebase:
- 144 files in tests directory alone demonstrate significant clutter
- Current approach: `foo.dingo` → `foo.go` + `foo.go.map` (same directory)
- Source maps are JSON files tracking position mappings for LSP integration
- Configuration system exists in `dingo.toml` but lacks output directory settings
- Build system in `cmd/dingo/main.go` uses simple file replacement logic

## Go Ecosystem Precedents

### 1. Established Patterns

**Protocol Buffers (protoc)**:
- Default: Generated files in same directory as `.proto`
- Common practice: Use `--go_out` flag for output directory specification
- Large projects: Separate `gen/` or `generated/` directories
- File naming: `filename.pb.go` (suffix pattern)

**Wire (Dependency Injection)**:
- In-place generation: `wire_gen.go` in same package
- Reason: Generated code must be in same package for Go visibility rules
- Integration: Works seamlessly with `go build`, `go test`

**Mock Generators (mockgen)**:
- Default: In-place `mock_<name>.go`
- Option: `-output` flag for separate directory
- Package consideration: Must respect Go package visibility

**Stringer (go generate)**:
- In-place generation: `<type>_string.go`
- Philosophy: Generated code co-located with source
- Reason: Go package visibility and import simplicity

### 2. Key Go Principles

1. **Package Cohesion**: Generated code needing access to unexported identifiers must be in same package
2. **Import Simplicity**: Go developers expect standard import paths without generated subdirectories
3. **Build Integration**: Must work seamlessly with `go build`, `go test`, `go install`
4. **Version Control**: Generated files typically committed for reproducible builds

## Recommended Strategy: "Configurable Hybrid"

### Primary Recommendation: `_gen.go` Suffix (Default)

```
project/
├── main.go              # Hand-written Go
├── auth.go              # Hand-written Go
├── user.dingo           # Dingo source
├── user_gen.go          # Generated Go
├── user_gen.go.map      # Source map
└── dingo.toml           # Configuration
```

**Advantages**:
- ✅ Aligns with Go ecosystem (`*_gen.go` convention)
- ✅ Same package = access to unexported identifiers
- ✅ No import path changes
- ✅ Clear generated vs hand-written distinction
- ✅ Works with existing Go toolchain
- ✅ Solves naming conflict (`foo.dingo` vs `foo.go`)

### Secondary Option: Directory Separation (Optional)

```
project/
├── src/                 # Dingo sources
│   ├── user.dingo
│   └── auth.dingo
├── gen/                 # Generated Go
│   ├── user.go
│   ├── user.go.map
│   ├── auth.go
│   └── auth.go.map
├── main.go              # Hand-written Go
├── go.mod
└── dingo.toml
```

**Configuration**:
```toml
[build]
# Primary choice: "suffix" (default) or "directory"
strategy = "suffix"

# For directory strategy
output_dir = "gen"
source_dir = "src"
```

### Enhanced `_gen.go` Approach (Recommended)

```toml
[build]
# File organization strategy
strategy = "suffix"  # "suffix" (default) or "directory"

# Suffix customization
suffix = "_gen"  # Default: "_gen", can be empty for in-place

# Source map handling
[build.sourcemaps]
location = "inline"  # "inline" (default), "separate", "both"
format = "json"      # Always JSON for now
```

## Implementation Details

### 1. Build System Changes

In `cmd/dingo/main.go`, modify `buildFile()`:

```go
func buildFile(inputPath string, outputPath string, buildUI *ui.BuildOutput, cfg *config.Config) error {
    // Handle output path based on strategy
    if outputPath == "" {
        outputPath = determineOutputPath(inputPath, cfg)
    }
    // ... rest of function unchanged
}

func determineOutputPath(inputPath string, cfg *config.Config) string {
    switch cfg.Build.Strategy {
    case "directory":
        return handleDirectoryStrategy(inputPath, cfg)
    case "suffix": // default
        return handleSuffixStrategy(inputPath, cfg)
    default:
        return handleSuffixStrategy(inputPath, cfg) // fallback
    }
}

func handleSuffixStrategy(inputPath string, cfg *config.Config) string {
    if len(inputPath) > 6 && inputPath[len(inputPath)-6:] == ".dingo" {
        base := inputPath[:len(inputPath)-6]
        return base + cfg.Build.Suffix + ".go"  // e.g., "user_gen.go"
    }
    return inputPath + cfg.Build.Suffix + ".go"
}
```

### 2. Configuration Schema Extension

Add to `pkg/config/config.go`:

```go
// BuildConfig controls build output and file organization
type BuildConfig struct {
    // Strategy controls file organization
    // Valid values: "suffix" (default), "directory"
    Strategy string `toml:"strategy"`

    // Suffix for generated files when using suffix strategy
    // Default: "_gen"
    Suffix string `toml:"suffix"`

    // OutputDir for directory strategy
    // Default: "gen"
    OutputDir string `toml:"output_dir"`

    // SourceDir for directory strategy
    // Default: "src"
    SourceDir string `toml:"source_dir"`
}
```

### 3. LSP Integration Updates

The LSP server needs to understand the new file mapping:

```go
// In pkg/lsp/server.go
func (s *Server) handleDocumentSymbol(params *lsp.DocumentSymbolParams) ([]lsp.DocumentSymbol, error) {
    // Check if this is a generated file
    if isGeneratedFile(params.TextDocument.URI) {
        originalURI := getOriginalFileURI(params.TextDocument.URI)
        return s.goplsClient.DocumentSymbol(context.Background(), lsp.DocumentSymbolParams{
            TextDocument: lsp.TextDocumentIdentifier{URI: originalURI},
        })
    }
    // ... normal handling
}
```

## Migration Path

### Phase 1: Default Implementation (v0.4.x)
1. Implement `_gen.go` suffix as default
2. Add configuration for directory strategy
3. Update build system
4. Update LSP integration

### Phase 2: Optional Directory Strategy (v0.5.x)
1. Implement directory strategy
2. Add migration tooling
3. Update documentation

### Phase 3: Advanced Features (v0.6.x)
1. Mixed strategies per project
2. Advanced source map handling
3. IDE integration improvements

## Trade-off Analysis

### `_gen.go` Suffix Approach
**Pros**:
- ✅ Familiar to Go developers (`*_gen.go` convention)
- ✅ No import path changes required
- ✅ Generated code in same package (visibility)
- ✅ Simple migration from current approach
- ✅ Works with all Go tools out-of-the-box

**Cons**:
- ❌ Generated files still visible in source directory
- ❌ 2-3 files per source (better than 3-4, but still clutter)

### Directory Approach
**Pros**:
- ✅ Clean separation of source vs generated
- ✅ Easier .gitignore management
- ✅ Clear project structure
- ✅ Better for large projects

**Cons**:
- ❌ Import path complications
- ❌ Generated code cannot access unexported identifiers
- ❌ More complex LSP integration
- ❌ Deviation from Go ecosystem norms

## Tooling Integration Considerations

### 1. Go Module System
With `_gen.go`: No changes needed
With directory: Require `replace` directives or submodule approach

### 2. IDE Integration (gopls, VSCode)
With `_gen.go`: Minimal changes required
With directory: Complex path translation required

### 3. CI/CD Pipeline
Both approaches work, but `_gen.go` requires no pipeline changes

## Final Recommendation

**Primary**: Implement `_gen.go` suffix strategy as default
**Secondary**: Offer directory strategy as optional configuration
**Timeline**: Start with `_gen.go` (immediate), add directory strategy (future)

This approach:
- Respects Go ecosystem conventions
- Provides immediate value with minimal disruption
- Offers flexibility for different project scales
- Maintains toolchain compatibility
- Supports gradual migration

By adopting the `*_gen.go` convention (like Wire, Mockgen, Stringer), Dingo immediately signals to Go developers that this follows established patterns, reducing cognitive load and adoption friction.
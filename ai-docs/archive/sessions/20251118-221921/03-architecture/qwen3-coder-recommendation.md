
[claudish] Model: qwen/qwen3-coder-30b-a3b-instruct

# Architectural Recommendation: Dingo File Organization Strategy

## 1. Recommended Strategy: Hybrid Shadow Folder with Build Directory

**Best Approach: Shadow folder architecture with configurable output directory**

The most effective strategy for Dingo is a hybrid approach that combines:
- **Source directory** (`src/`) for `.dingo` files and hand-written `.go` files
- **Generated directory** (`gen/`) for transpiled `.go` files and corresponding source maps
- **Configuration** to control generation location
- **Separate test structure** using `src/` files with validation

## 2. Concrete Directory Layout

```
dingo-project/
├── src/                    # Main source directory for .dingo files
│   ├── main.dingo         # Entry point .dingo files
│   ├── features/          # Feature-specific .dingo files
│   │   ├── error_handling.dingo
│   │   └── pattern_matching.dingo
│   └── lib/               # Library .dingo files  
│       └── utilities.dingo
│
├── gen/                   # Generated Go files (auto-generated)
│   ├── main.go            # Generated from src/main.dingo
│   ├── features/          # Generated files from src/features/
│   │   ├── error_handling.go
│   │   └── pattern_matching.go
│   └── lib/               # Generated files from src/lib/
│       └── utilities.go
│
├── .dingo/                # Hidden directory for source maps (optional)
│   └── maps/              # Per-file source maps (optional, could be in gen/)
│
├── tests/                 # Test suite (unchanged)
│   └── golden/            # Golden tests using source files directly
│       ├── error_prop_01_simple.dingo
│       ├── pattern_match_01_basic.dingo
│       └── ...
│
├── dingo.toml             # Configuration for build paths
├── go.mod                 # Go module definition
└── go.sum                 # Go dependency checksums
```

## 3. Detailed Rationale

### **Developer Experience**
- Clear separation between source `.dingo` and generated `.go` files
- Clean source directory - no clutter from generated files 
- Familiar structure mimicking TypeScript (node_modules pattern)
- Easy to identify which files are handwritten vs generated

### **Go Compatibility**
- Maintains standard Go package model (packages in same directory)
- Go tooling (`go build`, `go test`) works seamlessly
- Import paths remain unchanged (no need to convert relative/absolute paths)
- Source maps can be properly resolved for IDE integration

### **Tool Integration**
- IDEs can focus on source directory for development
- gopls integration works normally on generated code 
- Supports Go's module system and dependency management
- Compatible with standard Go build workflows

### **Scalability**
- Easy to add new features (just create new .dingo files in src/)
- Large projects can easily separate packages in `src/` and `gen/`
- Supports monorepo structure with multiple modules
- Clear boundaries improve maintainability

### **Maintainability**
- Generated files have clear headers indicating they're auto-generated
- No risk of overwriting handwritten files (`foo.go` vs `foo.dingo`)
- Simple .gitignore pattern: `/gen/`
- All generated artifacts are contained in one location

## 4. Configuration Options in `dingo.toml`

```toml
[build]
# Output directory for generated Go files
output_dir = "gen"

# Output directory for source maps (optional)
sourcemap_dir = ".dingo/maps"

# Enable/disable source map generation
generate_sourcemaps = true

# Build mode for generated code
build_mode = "default"  # "default", "development", "production"
```

## 5. Implementation Plan

### **Stage 1: File Structure Changes**
1. Modify `cmd/dingo/main.go` to accept `output_dir` configuration
2. Update file path construction to use configured output path
3. Add `sourcemap_dir` to source map writing logic

### **Stage 2: Directory Creation**
1. Create `gen/` directory if not present when building  
2. Ensure `gen/` and `src/` directories are properly maintained
3. Update `.gitignore` to ignore `gen/` directory

### **Stage 3: LSP Server Updates**
1. Extend `pkg/lsp/server.go` to resolve files in new structure
2. Update path mapping logic to work with shadow directory
3. Ensure source map resolution continues to work

### **Stage 4: Test Infrastructure Integration**
1. Modify golden test runner logic to work with new structure
2. Adjust test configuration to point to `src/` directory for source files
3. Ensure existing test expectations work without modification

## 6. Migration Path

### **For Existing Projects**
1. Copy all `.dingo` files to `src/` directory
2. Run `dingo build` with new configuration to generate files to `gen/` 
3. Update any build scripts or CI/CD processes
4. Move test references to source files in `src/` directory (where applicable)
5. Add `.dingo.toml` configuration to control new behavior

### **Backward Compatibility**
- Default behavior maintains current in-place approach for compatibility
- Flag `--output-dir` or config `output_dir` enables new structure
- Gradual migration through configuration options

## 7. Sample Implementation Details

### **Configuration Loading**
Update `pkg/config/config.go` to include:
```go
type BuildConfig struct {
    OutputDir string `toml:"output_dir"`
    SourceMapDir string `toml:"sourcemap_dir"`
    GenerateSourceMaps bool `toml:"generate_sourcemaps"`
}

// ...

type Config struct {
    Build BuildConfig `toml:"build"`
    // ... existing fields
}
```

### **Build File Logic**
Modify `cmd/dingo/main.go`:
```go
// In buildFile function:
func buildFile(inputPath string, outputPath string, buildUI *ui.BuildOutput, cfg *config.Config) error {
    // If no output file specified, default to gen/ directory or previous behavior
    if outputPath == "" {
        if cfg.Build.OutputDir != "" {
            // Use configured output directory    
            outputPath = filepath.Join(cfg.Build.OutputDir, stripPathPrefix(inputPath, cfg.Build.SourceDir))
        } else {
            // Fall back to in-place behavior
            outputPath = inputPath[:len(inputPath)-6] + ".go"
        }
    }
    
    // Use new output path for source map
    sourceMapPath := outputPath + ".map"
    if cfg.Build.SourceMapDir != "" {
        sourceMapPath = filepath.Join(cfg.Build.SourceMapDir, filepath.Base(outputPath) + ".map")
    }
    // ... rest unchanged
}
```

## 8. Comparison to Alternatives

### **Vs Current In-Place Generation**
- ✅ Eliminates file clutter in source directories
- ✅ Fixes test directory chaos (from 242 files down to ~62 logical units)
- ✅ Cleaner .gitignore rules (/gen/ instead of complex patterns)
- ✅ Eliminates name collision risk

### **Vs TypeScript/Rust Approach**
- ✅ Maintains Go module structure
- ✅ Preserves idiomatic Go file organization  
- ✅ Enables seamless integration with existing Go toolchain
- ✅ Less disruptive migration path

### **Vs Templ Pattern**
- ✅ Uses consistent file naming convention
- ✅ Applies better source separation
- ✅ Maintains Go package model compatibility
- ✅ Doesn't require name mangling or suffix patterns

This hybrid approach provides the best of both worlds: it fixes the immediate pain points while preserving Go compatibility and developer familiarity with the toolchain.

[claudish] Shutting down proxy server...
[claudish] Done


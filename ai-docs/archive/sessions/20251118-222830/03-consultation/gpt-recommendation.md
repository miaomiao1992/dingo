# Dingo File Organization Strategy: Architectural Recommendation

## Executive Summary

**Recommended Strategy: Target Directory with Flexible Configuration**
I recommend a hybrid approach where generated files go to a configurable output directory while maintaining simple in-place generation as the default. This provides the best balance of developer experience, tool integration, and scalability.

## Core Recommendation: `BuildConfig` with Smart Defaults

### Directory Structure Example

```
myproject/
├── dingo.toml                 # Project configuration
├── go.mod                     # Go module
├── go.sum                     # Go dependencies
├── src/                       # Source directory (configurable)
│   ├── main.dingo            # Dingo source files
│   ├── auth/
│   │   └── login.dingo
│   └── utils/
│       └── helpers.dingo
├── pkg/                       # Hand-written Go (optional)
│   └── shared.go
├── gen/                       # Generated Go files (default)
│   ├── main.go               # Generated from src/main.dingo
│   ├── auth/
│   │   └── login.go          # Generated from src/auth/login.dingo
│   ├── main.go.map           # Source maps
│   └── auth/
│       └── login.go.map
└── .gitignore                 # Ignore gen/ directory
```

**Key Benefits:**
1. **Clean separation**: Generated vs hand-written code
2. **Git simplicity**: Single rule `gen/` in `.gitignore`
3. **IDE friendliness**: Go tools see only valid Go files
4. **Scalability**: Scales to large projects without clutter
5. **Migration path**: Gradual adoption possible

## Configuration Schema for `dingo.toml`

```toml
[build]
# Strategy: "inplace" | "target" | "shadow" | "suffix"
strategy = "target"                    # Default for new projects

# Source directory containing .dingo files (relative to project root)
src_dir = "src"                       # Default: "." (current dir)

# Output directory for generated .go files
out_dir = "gen"                       # Default: "gen" (ignored for inplace)

# Source map output configuration
source_maps = "separate"               # "inline" | "separate" | "both" | "none"

# File naming pattern for generated files
# Variables: {name} = base name, {path} = relative path from src_dir
output_pattern = "{path}/{name}.go"    # Maintains directory structure

# Whether to preserve source directory structure in output
preserve_structure = true              # Default: true

# Additional options
[build.options]
# Include generation timestamp in output files (development mode)
include_timestamp = false              # Default: false

# Clean generated files before build (prevents orphaned files)
clean_before_build = false             # Default: false

# Generate build metadata file (JSON with manifest of generated files)
generate_manifest = true              # Default: true
```

## Implementation Considerations

### CLI Flags Enhancement

```bash
# Basic usage (unchanged for compatibility)
dingo build file.dingo              # In-place (current behavior)
dingo build src/*.dingo             # Target directory with config

# New flags for explicit control
dingo build --strategy target --out-dir ./dist src/*.dingo
dingo build --strategy suffix src/*.dingo           # file_gen.go
dingo build --strategy shadow src/*.dingo            # src/ -> gen/

# Development convenience
dingo build --dev src/main.dingo     # Shortcut for target + inline maps + timestamps
dingo build --clean src/*.dingo      # Clean orphaned files first
```

### LSP Integration Changes

The LSP server must be aware of output strategies:

1. **Position Translation**: Source maps already handle this
2. **File Discovery**: LSP needs to check both source and output
3. **Go Module Path**: Generated files must maintain correct package paths
4. **Build Listening**: Watch for changes in source directory

```go
// LSP configuration enhancements
type LSPConfig struct {
    BuildStrategy    string `toml:"strategy"`
    SourceDirectory  string `toml:"src_dir"`
    OutputDirectory   string `toml:"out_dir"`
    MapTranslations  map[string]string   // Cached file translations
}
```

### Build System Integration

#### Phase 1: Configuration Integration
- Extend `pkg/config/config.go` with `BuildConfig` struct
- Add validation for incompatible combinations
- Implement config inheritance (project > user > defaults)

#### Phase 2: Path Resolution Engine
```go
type PathResolver struct {
    Strategy  string
    SrcDir    string
    OutDir    string
    Pattern   string
}

func (pr *PathResolver) ResolveOutput(inputPath string) string {
    switch pr.Strategy {
    case "inplace":
        return strings.Replace(inputPath, ".dingo", ".go", 1)
    case "target":
        return pr.preserveStructure(inputPath)
    case "suffix":
        return strings.Replace(inputPath, ".dingo", "_gen.go", 1)
    case "shadow":
        return strings.Replace(inputPath, "src/", "gen/", 1)
    }
}
```

#### Phase 3: Clean Integration
```go
// BuildManifest tracks generated files for cleanup
type BuildManifest struct {
    GeneratedFiles []string           // List of all generated files
    OrphanedFiles []string           // Files that should be cleaned up
    Timestamp     time.Time          // Build time
}

func (bm *BuildManifest) CleanOrphaned() error {
    // .go files in out_dir without corresponding .dingo sources
}
```

## Trade-off Analysis

### vs In-Place Generation (Current)

| Aspect | Target Directory | In-Place Current |
|--------|------------------|------------------|
| ✅ File Clutter | Clean separation | 3x files per source |
| ✅ Git Simplicity | Ignore gen/ directory | Complex patterns needed |
| ✅ Name Collisions | foo.dingo + foo.go possible | Cannot coexist |
| ✅ IDE Integration | Go sees only valid Go files | Mixed content types |
| ✅ Scalability | Handles large projects | Becomes cluttered quickly |
| ❌ Simplicity | Extra configuration | Just works |
| ❌ Migration Effort | Requires project setup | No changes needed |

### vs Shadow Folder Strategy

| Aspect | Target Directory | Shadow Folder |
|--------|------------------|----------------|
| ✅ Go Conventions | Follows standard patterns | src/ naming less common in Go |
| ✅ Windows Compatibility | No reserved names | "src" safe, "gen" safer |
| ✅ Mixed Language | Natural for .go + .dingo mix | Assumes pure Dingo |
| ❌ Cognitive Load | More directories | Simpler mental model |

### vs Suffix Pattern

| Aspect | Target Directory | Suffix Pattern |
|--------|------------------|----------------|
| ✅ Build Tools | Go sees all Go files | Works with existing tools |
| ✅ Source Maps | Simple translations | More complex mapping |
| ✅ Clean Separation | Generated grouped with source | Mixed file types in same dir |
| ❌ File Discovery | Simple glob patterns | Requires filtering |

## Migration Path for Existing Projects

### Phase 1: Backward Compatibility (v0.3.x)
- **Current behavior**: In-place generation remains default
- **Opt-in**: `dingo.toml` enables target directory strategy
- **CLI flags**: New flags don't change existing behavior
- **Migration Tool**: `dingo migrate --init-config` creates configuration

```bash
# Existing project migration
dingo migrate --project-id my-app --strategy target \
  --src-dir . --out-dir gen --preserve-structure
```

### Phase 2: Gradual Transition (v0.4.x)
- **Smart defaults**: New projects use target directory by default
- **Mixed projects**: Allow both strategies simultaneously
- **IDE integration**: VS Code/GoLand plugins understand both patterns
- **Documentation**: Migration guide with best practices

### Phase 3: Complete Transition (v0.5.x)
- **Default change**: Target directory becomes default for all
- **Deprecation warnings**: In-place shows deprecation notice
- **Auto-migration**: `dingo build --migrate` handles conversion

## IDE Integration Expectations

### VS Code Extension
```json
{
  "dingo.buildStrategy": "target",
  "dingo.sourceDirectory": "src",
  "dingo.outputDirectory": "gen",
  "dingo.excludeFromBuild": ["*_gen.go"],
  "dingo.generatedGlob": "gen/**/*.go"
}
```

### GoLand Plugin
- **File Status**: Color code generated vs hand-written files
- **Navigation**: Go to Definition works across .dingo → .go
- **Refactoring**: Safe rename respects bidirectional mapping
- **Build Integration**: Compile generated files automatically

## Go Ecosystem Compatibility

### Build Tools Integration

#### Go Build/Install
```bash
# Works seamlessly
go build ./gen/cmd/server
go install ./gen/...

# Or include in go.mod build tags
//go:generate dingo build ./src/...
```

#### Testing Integration
```bash
# Test generated code
go test ./gen/... -v

# Mixed testing (generated + hand-written)
go test ./... -v
```

#### CI/CD Pipeline
```yaml
# GitHub Actions example
- name: Build Dingo
  run: |
    dingo build --strategy target src/**/*.dingo
    go build ./gen/...
    go test ./gen/... -v
```

### Module System Considerations

Generated files must maintain correct module paths:

```go
// src/auth/login.dingo → gen/auth/login.go
package auth  // Automatically preserved

// import paths remain consistent
import "github.com/user/project/gen/auth"
```

## Implementation Roadmap

### Sprint 1: Configuration Foundation (1-2 weeks)
1. ✅ Extend `pkg/config/config.go` with `BuildConfig`
2. ✅ Add validation and default handling
3. ✅ CLI flag integration for strategy control
4. ✅ Basic path resolution engine

### Sprint 2: Build System Integration (2-3 weeks)
1. ✅ Update `cmd/dingo/main.go` with new output logic
2. ✅ Source map path handling for separated files
3. ✅ Build manifest generation and cleanup
4. ✅ File discovery and orphaned file detection

### Sprint 3: LSP and IDE Support (2-3 weeks)
1. ✅ Update LSP server for multiple directory awareness
2. ✅ Position translation with new mapping strategy
3. ✅ VS Code extension updates
4. ✅ GoLand plugin prototype

### Sprint 4: Migration Tooling (1-2 weeks)
1. ✅ `dingo migrate` command implementation
2. ✅ Configuration templates for common patterns
3. ✅ Documentation and migration guides
4. ✅ Backward compatibility testing

## Conclusion & Final Recommendation

The **Target Directory Strategy with Smart Defaults** provides the optimal balance for Dingo's evolution:

**Immediate Benefits:**
- Solves current file clutter (242 → ~80 files in tests)
- Enables clean separation of concerns
- Maintains backward compatibility
- Provides clear migration path

**Long-term Advantages:**
- Scales to enterprise projects
- Integrates seamlessly with Go ecosystem
- Supports mixed-language development
- Future-proof for advanced features

**Risk Mitigation:**
- Gradual migration prevents disruption
- Configuration flexibility accommodates diverse needs
- Extensive testing ensures reliability
- Clear documentation reduces adoption friction

This strategy positions Dingo as a production-ready language that respects Go conventions while providing enhanced developer experience. The flexible configuration ensures it works for everything from small utilities to large-scale applications.

---

**Next Steps:**
1. Implement configuration foundation in Sprint 1
2. Create migration documentation and tools
3. Gather community feedback on proposed defaults
4. Begin gradual rollout with opt-in availability

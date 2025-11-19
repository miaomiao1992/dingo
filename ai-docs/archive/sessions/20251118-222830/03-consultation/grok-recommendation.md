# Dingo File Organization: Rust/Cargo Systems Architecture Recommendations

## Executive Summary

From a Rust/cargo systems perspective, Dingo should adopt a **target/ directory model with semantic module mapping**, similar to how Rust handles proc macro expansion but optimized for Go's module system. This provides the best balance of separation, tooling integration, and scalability.

## Recommended Architecture: Target Directory with Module Mirroring

### Core Recommendation: `target/dingo/` with Exact Path Mirroring

```
project/
├── src/
│   ├── main.dingo
│   ├── auth/
│   │   ├── login.dingo
│   │   └── middleware.dingo
│   └── utils/
│       └── validation.dingo
├── target/dingo/
│   ├── main.go
│   ├── main.go.map
│   ├── auth/
│   │   ├── login.go
│   │   ├── login.go.map
│   │   ├── middleware.go
│   │   └── middleware.go.map
│   └── utils/
│       ├── validation.go
│       └── validation.go.map
└── go.mod
```

**Key Innovation**: Exact path mirroring maintains Go's package resolution while separating artifacts.

## 1. Rust Target Model Analysis

### Why Rust's `target/` Works

```
cargo build → target/debug/
            ├── deps/           # Object files
            ├── build/          # Build scripts output
            ├── incremental/    # Incremental compilation
            └── examples/       # Compiled examples
```

**Principles Dingo Should Adopt:**
1. **Single Source of Truth**: All generated artifacts in one location
2. **Predictable Structure**: Mirrors source for navigation
3. **Tool Awareness**: All tools know where to find artifacts
4. **Clean Workspace**: Source directories remain pristine

### Rust Proc Macro Comparison

```rust
// Source
#[diesel_derive(Queryable)]
struct User { id: i32, name: String }

// Expanded (in target/debug/build/)
// Expanded AST used by compiler, not visible in source tree
```

**Dingo Parallel**: Generated Go files visible but isolated, enabling debugging while maintaining cleanliness.

## 2. Build Performance Implications

### Performance Metrics

| Strategy | Incremental Build | Full Build | LSP Lookup | Cache Efficiency |
|----------|-------------------|------------|------------|------------------|
| In-Place | ⚡ Fastest (no copy) | ⚡ Fastest | ⚡ Direct | ❌ High false positives |
| Target | ✅ Fast (copy overhead) | ✅ Fast | ⚡ O(1) hash lookup | ✅ Perfect cache invalidation |
| Shadow | 🐌 Slow (dual copy) | 🐌 Slow | 🐌 Path translation | ✅ Good |

**Performance Recommendation**: Target directory with hard links for production, symlinks for development.

### Rust's Incremental Compilation Lessons

```rust
// Rust tracks:
- File hash
- Dependency graph
- Compilation fingerprint
- Result location
```

**Dingo Adaptation**:
```go
type BuildCache struct {
    SourceHash    map[string]string    // .dingo → hash
    OutputHash    map[string]string    // .go → hash
    DepGraph      *dag.Graph          // Import relationships
    LastBuild     time.Time           // Global cache invalidation
}
```

## 3. LSP Mapping Strategies

### Rust Analyzer Approach

```rust
// rust-analyzer macro expansion
macro_rules! vec {
    ($($x:expr),*) => {
        $crate::collections::Vec::new()
    };
}

// Maps: source position → expanded position → original
```

### Dingo LSP Architecture

```
Client LSP Request
    ↓
Source Map Lookup (O(1))
    ↓
Position Translation (token → byte → line)
    ↓
gopls Forward (translated position)
    ↓
Response Translation (gopls → client coordinate space)
```

**Implementation Strategy**:
```go
type LSPMapper struct {
    // Fast path: exact match
    exactCache map[lsp.ID]PositionMapping

    // Slow path: fuzzy search
    fuzzySearch *sourcemap.PositionSearch

    // Rust analyzer style: multiple expansion tracking
    expansionStack [] ExpansionLayer
}
```

### Performance Optimization: Rust's Salsa Database Pattern

```rust
// Salsa query system (rust-analyzer uses this)
Query<SourceFile> → Query<ExpandedSource> → Query<ParsedAST>
```

**Dingo Adaptation**:
```go
// Query-inspired LSP caching
type Database struct {
    sourceFiles      map[string]*SourceFile      // .dingo files
    generatedFiles   map[string]*GeneratedFile   // .go files
    sourceMaps       map[string]*SourceMap       // Position mappings
    fingerprints     map[string]uint64          // Content hashes
}
```

## 4. Build Artifact Lifecycle Management

### Rust's Cargo Clean Pattern

```bash
# Rust: granular cleanup
cargo clean --release          # Remove target/release/
cargo clean --doc              # Remove target/doc/
cargo check                    # No artifacts, just type checking
```

### Recommended Dingo Commands

```bash
# Direct cargo parity
dingo clean                   # Remove target/dingo/
dingo clean --cache           # Remove .dingo-cache/
dingo clean --maps            # Remove .go.map files only
dingo build --dry-run         # No output, just validation
dingo check                   # Parse and type check only
```

### Rust's Source Dependency Management

```toml
# Cargo separates source from artifacts
[dependencies]
serde = "1.0"           # Declaration only
# Downloaded to: ~/.cargo/registry/src/
# Compiled to: target/deps/
```

**Dingo Module Integration**:
```go
// go.mod stays clean
module github.com/user/project

go 1.21

require (
    github.com/external/lib v1.0.0
)

// Generated files seamlessly integrate
package main

import (
    "github.com/external/lib"
    // Dingo-generated imports work transparently
    "./target/dingo/auth"  // ← This is the innovation
)
```

## 5. Recommended Configuration Schema

### `dingo.toml` Architecture

```toml
[build]
# Rust-inspired target configuration
target_dir = "target/dingo"           # Default: target/dingo
strategy = "target"                   # target, inplace, hybrid
use_hardlinks = false                 # Performance toggle
incremental = true                    # Enable build cache

[build.cache]
directory = ".dingo-cache"           # Cargo-style cache dir
max_size = "2GB"                     # Cache size limits
 ttl = "7days"                       # Cache expiration

[build.lsp]
map_positions = true                  # Enable source maps
macro_expansion = false              # Future: show expansion
cache_discovery = true               # LSP-side caching

[build.outputs]
# Granular artifact control
go_files = true                      # Generate .go
map_files = true                     # Generate .go.map
manifest = true                      # Generate .build-manifest.json

[source]
# Source discovery (Rust's src/ convention)
include = ["**/*.dingo"]
exclude = ["**/*_test.dingo", "**/.dingo-cache/**"]
watch_mode = true                    # File system watching
```

## 6. Migration Strategy

### Rust-Inspired Incremental Migration

```bash
# Phase 1: Add target support without breaking changes
dingo build --target-dir target/dingo   # New flag
dingo build                            # Legacy: in-place (deprecated)

# Phase 2: Default to target
dingo build                           # Now uses target/ by default
dingo build --in-place               # Opt-in to legacy

# Phase 3: Remove in-place
dingo build                           # Only target/ supported
```

### Backward Compatibility Layer

```go
// CLI: legacy support with deprecation warnings
if cfg.InPlace && !cfg.TargetDir {
    log.Warn("In-place generation deprecated. Use --target-dir.")
    return executeInPlaceBuild()
}
```

## 7. Advanced Systems Considerations

### Rust's Distinct Target Directories

```rust
// Rust separates build types
target/debug/     # Development builds
target/release/   # Optimized builds
target/doc/       # Documentation
target/bench/     # Benchmarks
```

**Dingo Adaptation**:
```
target/
├── dingo/debug/      # Default development builds
├── dingo/release/    # Optimized/production builds
├── dingo/test/       # Test-specific builds
└── dingo/docs/       # Generated documentation
```

### Cross-Compilation Lessons

```rust
// Rust: platform-specific artifacts
target/x86_64-unknown-linux-gnu/debug/
target/x86_64-pc-windows-gnu/debug/
```

**Dingo Future-Proofing**:
```
target/dingo/linux_amd64/    # Cross-compiled artifacts
target/dingo/windows_amd64/
target/dingo/darwin_arm64/
```

## 8. Implementation Roadmap

### Phase 1: Core Target Implementation (1-2 weeks)
- Basic target directory generation
- Path mirroring logic
- CLI `--target-dir` flag

### Phase 2: LSP Integration (2-3 weeks)
- Source map lookup optimization
- Position translation caching
- gopls integration testing

### Phase 3: Performance Optimization (1-2 weeks)
- Hard link/symlink strategies
- Build cache implementation
- Incremental build logic

### Phase 4: Advanced Features (2-3 weeks)
- Multiple build types (debug/release/test)
- Cache size management
- Migration tooling

## 9. Risk Analysis

### High-Risk Areas
1. **Go Module Integration**: Generated files must maintain correct import paths
2. **LSP Performance**: Source map lookups must be sub-millisecond
3. **Tooling Compatibility**: IDE debuggers need correct source mapping

### Mitigation Strategies
1. **Extensive Integration Testing**: Test with real Go modules
2. **Performance Benchmarks**: Monitor LSP response times
3. **Compatibility Matrix**: Test with Go 1.19-1.21, various IDEs

## 10. Success Metrics

### Performance Targets
- **Build Time**: <10% overhead vs in-place
- **LSP Lookup**: <5ms average response time
- **Cache Hit Rate**: >95% for incremental builds
- **Disk Usage**: <2x source size with hard links

### Developer Experience Goals
- **Discovery**: Intuitive file organization
- **Debugging**: Easy navigation to generated code
- **Git**: Clean status, easy .gitignore
- **CI/CD**: No generated file commits needed

## Conclusion

The target directory approach with exact path mirroring provides the optimal balance of Rust's proven artifact management principles with Go's module system requirements. This architecture:

1. **Scales**: Linear performance regardless of project size
2. **Integrates**: Preserves all Go tooling compatibility
3. **Maintains**: Clean source directories for developer happiness
4. **Optimizes**: Rust-inspired caching and incremental compilation
5. **Future-Proofs**: Supports cross-compilation, multiple build types, advanced LSP features

This is the most robust, production-ready approach that will serve Dingo's long-term architectural needs while maintaining excellent developer experience.
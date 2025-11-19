# Package-Wide Scanning Implementation Plan

## Overview

This document provides a concrete, phased implementation roadmap for the package-wide scanning architecture. Each phase includes specific tasks, file modifications, and testing requirements.

## Implementation Phases

### Phase 1: Core Infrastructure (Weeks 1-2)

**Goal**: Establish basic package scanning without caching

#### 1.1 Create Package Structure

**New Packages:**
```
pkg/scanner/
  ├── scanner.go              # Core scanner interface
  ├── package_info.go         # Data structures
  ├── function_index.go       # Function indexing logic
  ├── cache.go                # Cache interfaces
  ├── memory_cache.go         # In-memory LRU cache
  ├── disk_cache.go           # Persistent cache
  ├── incremental.go          # Change detection
  └── deps.go                 # Dependency analysis

pkg/scanner/cmd/              # Optional: scanner CLI tool
  └── main.go
```

**Modified Packages:**
```
pkg/config/
  └── config.go               # Add scanner config section
```

#### 1.2 Core Data Structures

**File: `pkg/scanner/package_info.go`**

```go
package scanner

// PackageInfo contains all package-level information
type PackageInfo struct {
    PackagePath    string                 // "github.com/user/project"
    PackageName    string                 // "main"
    Files          []FileInfo             // All .dingo files in package
    Imports        map[string]string      // Import path -> alias
    FunctionIndex  map[string][]FunctionDecl
    Dependencies   map[string]bool
    Timestamp      time.Time
    SourceMap      map[string]*SourceMap
}

// FileInfo represents a single source file
type FileInfo struct {
    Path       string
    Hash       string
    Size       int64
    Modified   time.Time
    Source     []byte
    AST        *ast.File
}

// FunctionDecl represents a function declaration
type FunctionDecl struct {
    Name        string
    Package     string
    File        string
    Line        int
    Signature   string
    ReturnTypes []string
    IsExported  bool
    GoDoc       string
}
```

**File: `pkg/scanner/scanner.go`**

```go
package scanner

// PackageScanner is the main scanning component
type PackageScanner struct {
    config     *config.Config
    cache      *PackageCache
    fileSet    *token.FileSet
    logger     plugin.Logger
}

// Scanner interface
type Scanner interface {
    ScanPackage(ctx context.Context, pkgPath string) (*PackageInfo, error)
    ScanFiles(ctx context.Context, files []string) (*PackageInfo, error)
    GetFunctionIndex(pkgPath string) *FunctionIndex
    InvalidateCache(pkgPath string, files []string) error
}
```

#### 1.3 Package Discovery

**File: `pkg/scanner/discovery.go`**

```go
package scanner

func (s *PackageScanner) discoverPackageFiles(pkgPath string) ([]FileInfo, error) {
    // Implementation steps:
    // 1. Read go.mod to get package name
    // 2. Walk directory for .dingo files
    // 3. Calculate file hashes
    // 4. Load source code
    // 5. Parse basic AST (for import extraction)
}
```

#### 1.4 Function Index Building

**File: `pkg/scanner/function_index.go`**

```go
package scanner

func (s *PackageScanner) buildFunctionIndex(ctx context.Context, files []FileInfo) (map[string][]FunctionDecl, error) {
    // Implementation steps:
    // 1. For each file:
    //    a. Preprocess
    //    b. Parse with go/parser
    //    c. Run type checker
    //    d. Extract function declarations
    // 2. Build index by function name
    // 3. Resolve cross-references
}
```

#### 1.5 Basic Cache Layer

**File: `pkg/scanner/memory_cache.go`**

```go
package scanner

// InMemoryCache provides simple in-memory caching
type InMemoryCache struct {
    packages map[string]*CacheEntry
    lru      *list.List
    config   MemoryConfig
    mutex    sync.RWMutex
}
```

#### Testing for Phase 1

**Unit Tests:**
- `pkg/scanner/scanner_test.go`
- `pkg/scanner/discovery_test.go`
- `pkg/scanner/function_index_test.go`

**Integration Tests:**
- `pkg/scanner/integration_test.go`

#### Success Criteria

- [ ] Can scan a 10-file package
- [ ] Function index contains all functions
- [ ] In-memory cache works
- [ ] Tests pass (unit and integration)

---

### Phase 2: Enhanced Cache System (Week 3)

**Goal**: Add persistent cache and improve performance

#### 2.1 Disk Cache Implementation

**File: `pkg/scanner/disk_cache.go`**

```go
package scanner

type DiskCache struct {
    baseDir  string
    config   DiskConfig
    mutex    sync.RWMutex
}

func (c *DiskCache) Get(pkgPath string) (*PackageInfo, error)
func (c *DiskCache) Put(pkgPath string, info *PackageInfo) error
func (c *DiskCache) Delete(pkgPath string) error
func (c *DiskCache) Clear() error
```

#### 2.2 Two-Level Cache Integration

**File: `pkg/scanner/cache.go`**

```go
package scanner

type PackageCache struct {
    memoryCache  *InMemoryCache
    diskCache    *DiskCache
    config       CacheConfig
}

func (c *PackageCache) Get(pkgPath string) (*PackageInfo, bool)
func (c *PackageCache) Put(pkgPath string, info *PackageInfo)
func (c *PackageCache) Invalidate(pkgPath string) error
```

#### 2.3 Serialization Format

**Decision: Use Gob format**
- Pros: Fast, handles Go types natively, compact
- Cons: Go-specific (acceptable for this use case)
- Alternative considered: JSON (slower, larger)

**File: `pkg/scanner/serializer.go`**

```go
package scanner

// GobSerializer implements Serializer interface
type GobSerializer struct{}

func (s *GobSerializer) Serialize(info *PackageInfo) ([]byte, error)
func (s *GobSerializer) Deserialize(data []byte, info *PackageInfo) error
```

#### 2.4 Cache Invalidation

**File: `pkg/scanner/invalidation.go`**

```go
package scanner

type InvalidationStrategy int

const (
    TimeBased    InvalidationStrategy = iota // TTL-based
    SizeBased                                // LRU eviction
    Manual                                   // User-triggered
)

func (c *PackageCache) SetupInvalidation(strategy InvalidationStrategy, config InvalidationConfig)
```

#### 2.5 Cache Statistics

**File: `pkg/scanner/metrics.go`**

```go
package scanner

type CacheStats struct {
    Hits           int64
    Misses         int64
    HitRate        float64
    MemoryUsage    int64
    DiskUsage      int64
    Entries        int
    Evictions      int64
}
```

#### Testing for Phase 2

**Unit Tests:**
- `pkg/scanner/disk_cache_test.go`
- `pkg/scanner/cache_test.go`

**Integration Tests:**
- Test cache persistence (restart and reload)
- Test cache eviction
- Test concurrent access

#### Success Criteria

- [ ] Cache persists across process restarts
- [ ] Cache hit rate >70% on repeated scans
- [ ] Cache invalidation works correctly
- [ ] Serialization is fast (<5ms for 200 files)

---

### Phase 3: Incremental Build Support (Weeks 4-5)

**Goal**: Support watch mode and fast incremental rebuilds

#### 3.1 File Watching

**File: `pkg/scanner/watcher.go`**

```go
package scanner

type Watcher struct {
    scanner   *PackageScanner
    watcher   *fsnotify.Watcher
    callbacks map[string][]func(string)
}

func (w *Watcher) Watch(path string) error
func (w *Watcher) Close() error
```

#### 3.2 Change Detection

**File: `pkg/scanner/incremental.go`**

```go
package scanner

type ChangeSet struct {
    ChangedFiles []string
    NewFiles     []string
    DeletedFiles []string
}

func (s *PackageScanner) DetectChanges(pkgPath string) (*ChangeSet, error)
func (s *PackageScanner) ScanIncrementally(ctx context.Context, pkgPath string, changes *ChangeSet) (*PackageInfo, error)
```

#### 3.3 Dependency Graph

**File: `pkg/scanner/deps.go`**

```go
package scanner

type DependencyGraph struct {
    edges    map[string][]string
    inverted map[string][]string
    mutex    sync.RWMutex
}

func (g *DependencyGraph) AddDependency(pkg, dep string)
func (g *DependencyGraph) GetImpactedPackages(changedPkg string) []string
```

#### 3.4 Impact Analysis

**File: `pkg/scanner/impact.go`**

```go
package scanner

func (s *PackageScanner) CalculateImpact(changes *ChangeSet) ([]string, error) {
    // Implementation:
    // 1. For each changed file, find who depends on it
    // 2. For each new file, find who it might depend on
    // 3. For each deleted file, find who depends on it
    // 4. Return list of files that need re-scanning
}
```

#### 3.5 Selective Re-scanning

**File: `pkg/scanner/rescan.go`**

```go
package scanner

func (s *PackageScanner) PartialRescan(ctx context.Context, pkgPath string, affectedFiles []string) (*PackageInfo, error) {
    // Implementation:
    // 1. Load cached package info
    // 2. Re-scan only affected files
    // 3. Merge results with cached data
    // 4. Update cache with new info
}
```

#### Testing for Phase 3

**Unit Tests:**
- `pkg/scanner/watcher_test.go`
- `pkg/scanner/incremental_test.go`

**Integration Tests:**
- Watch a directory and verify incremental updates
- Test dependency-based re-scanning

#### Success Criteria

- [ ] Single file change detection <50ms
- [ ] Incremental rebuild <150ms
- [ ] Correctly identifies affected files
- [ ] Watch mode works without false positives

---

### Phase 4: Import Inference Integration (Weeks 6-7)

**Goal**: Integrate with existing ImportTracker and transpiler

#### 4.1 ImportTracker Enhancement

**File: `pkg/imports/tracker.go`**

```go
package imports

type ImportTracker struct {
    scanner     scanner.Scanner
    index       *FunctionIndex
    config      *config.Config
}

func (t *ImportTracker) InferUnqualifiedUsage(file *ast.File, funcName string, pos token.Pos) (*InferenceResult, error)
```

#### 4.2 Inference Rules

**File: `pkg/imports/inference.go`**

```go
package imports

type InferenceRule int

const (
    InferByName      InferenceRule = iota // Exact name match
    InferByPackage                       // Last segment match
    InferByAlias                         // Alias match
    InferContextAware                    // Type-based inference
)

func (t *ImportTracker) ApplyRule(rule InferenceRule, funcName string, file *ast.File) (*InferenceResult, error)
```

#### 4.3 Context-Aware Type Inference

**File: `pkg/imports/context.go`**

```go
package imports

func (t *ImportTracker) InferFromReturnTypes(expectedTypes []string, pkgInfo *PackageInfo) (*InferenceResult, error) {
    // Implementation:
    // 1. Search function index for functions with matching return types
    // 2. Score matches by type compatibility
    // 3. Return best match
}
```

#### 4.4 Preprocessor Integration

**File: `pkg/preprocessor/preprocessor.go`** (modified)

```go
type Preprocessor struct {
    source          []byte
    processors      []FeatureProcessor
    config          *config.Config
    packageContext  *scanner.PackageInfo  // NEW: Add package context
}

func (p *Preprocessor) SetPackageContext(ctx *scanner.PackageInfo) {
    p.packageContext = ctx
}
```

#### 4.5 Generator Integration

**File: `pkg/generator/generator.go`** (modified)

```go
func (g *Generator) GenerateWithPackageContext(file *dingoast.File, pkgContext *scanner.PackageInfo) ([]byte, error) {
    // 1. Set package context in pipeline
    g.pipeline.Ctx.PackageContext = pkgContext

    // 2. Continue with normal generation
    return g.Generate(file)
}
```

#### Testing for Phase 4

**Unit Tests:**
- `pkg/imports/inference_test.go`
- `pkg/imports/tracker_test.go`

**Integration Tests:**
- End-to-end unqualified import inference
- Cross-file function call resolution

**Golden Tests:**
- `tests/golden/unqualified_import_*.dingo`
- `tests/golden/qualified_import_*.dingo`

#### Success Criteria

- [ ] Can infer unqualified function calls
- [ ] Context-aware inference works
- [ ] No false positives (wrong imports)
- [ ] Golden tests pass

---

### Phase 5: Performance Optimization (Week 8)

**Goal**: Optimize for large packages and production use

#### 5.1 Parallel Scanning

**File: `pkg/scanner/parallel.go`**

```go
package scanner

func (s *PackageScanner) ScanPackageParallel(ctx context.Context, pkgPath string) (*PackageInfo, error) {
    // Implementation:
    // 1. Split files into batches
    // 2. Process batches in parallel
    // 3. Merge results
}
```

#### 5.2 Lazy Type Checking

**File: `pkg/scanner/lazy.go`**

```go
package scanner

type TypeCheckStrategy int

const (
    TypeCheckAll      TypeCheckStrategy = iota
    TypeCheckExported                  // Only exported functions
    TypeCheckNone                      // Quick parse only
)

func (s *PackageScanner) SetTypeCheckStrategy(strategy TypeCheckStrategy)
```

#### 5.3 Streaming Index Build

**File: `pkg/scanner/streaming.go`**

```go
package scanner

func (s *PackageScanner) BuildIndexStreaming(files []FileInfo) (<-chan FunctionDecl, <-chan error) {
    // Implementation:
    // 1. Use goroutines to process files
    // 2. Stream results to channel
    // 3. Allow early termination
}
```

#### 5.4 Adaptive Caching

**File: `pkg/scanner/adaptive.go`**

```go
package scanner

type AdaptiveCache struct {
    hotSet  *InMemoryCache  // Frequent
    warmSet *DiskCache      // Occasional
    coldSet *ArchivedCache  // Rare
}

func (c *AdaptiveCache) Get(pkgPath string) (*PackageInfo, error)
```

#### Testing for Phase 5

**Benchmark Tests:**
- `pkg/scanner/benchmark_test.go`
- Large package (500+ files)
- Memory profiling

#### Success Criteria

- [ ] 200-file package scans in <5s
- [ ] Parallel scanning 3x faster
- [ ] Memory usage <200MB for 200 files
- [ ] Cache hit rate >85%

---

### Phase 6: Integration & CLI Integration (Week 9)

**Goal**: Integrate with dingo CLI and test thoroughly

#### 6.1 CLI Integration

**File: `cmd/dingo/main.go`** (modified)

```go
func buildFile(inputPath string, outputPath string, buildUI *ui.BuildOutput, cfg *config.Config) error {
    // NEW: Load package context
    scanner := scanner.NewPackageScanner(cfg)
    pkgInfo, err := scanner.ScanPackage(context.Background(), getPackagePath(inputPath))
    if err != nil {
        buildUI.PrintInfo("Package scan failed, continuing without context: %v", err)
    }

    // Continue with existing build logic, passing pkgInfo
    // ...
}
```

#### 6.2 Configuration

**File: `pkg/config/config.go`** (modified)

```go
type Config struct {
    // ... existing fields ...

    // NEW: Scanner configuration
    Scanner *ScannerConfig
}
```

#### 6.3 Watch Mode Support

**File: `cmd/dingo/main.go`** (modified)

```go
// Add to buildCmd flags
cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for file changes and rebuild")

// Implement watch logic
func runWatch(ctx context.Context, files []string, output string, cfg *config.Config) error {
    scanner := scanner.NewPackageScanner(cfg)
    watcher := scanner.NewWatcher()

    for {
        // Wait for file changes
        changes := watcher.WaitForChanges(ctx)

        // Incremental rebuild
        if err := runBuild(files, output, false, cfg); err != nil {
            // Handle error
        }
    }
}
```

#### Testing for Phase 6

**Integration Tests:**
- Full end-to-end test with CLI
- Watch mode behavior
- Performance under load

**Golden Tests:**
- All existing tests must pass
- Add new golden tests for package scanning

#### Success Criteria

- [ ] CLI works with package scanning
- [ ] Watch mode functional
- [ ] All golden tests pass
- [ ] Performance regressions <10%

---

## Key Files to Create/Modify

### New Files

1. **pkg/scanner/scanner.go**
   - Core PackageScanner
   - Scanner interface
   - Basic implementation

2. **pkg/scanner/package_info.go**
   - PackageInfo struct
   - FileInfo struct
   - FunctionDecl struct

3. **pkg/scanner/function_index.go**
   - Function indexing logic
   - Type checking integration
   - Cross-reference resolution

4. **pkg/scanner/cache.go**
   - Cache interface
   - Two-level cache implementation

5. **pkg/scanner/memory_cache.go**
   - In-memory LRU cache
   - Memory management

6. **pkg/scanner/disk_cache.go**
   - Disk-based cache
   - Serialization

7. **pkg/scanner/incremental.go**
   - Change detection
   - Incremental scanning
   - Impact analysis

8. **pkg/scanner/deps.go**
   - Dependency graph
   - Impact calculation

9. **pkg/scanner/parallel.go**
   - Parallel scanning
   - Batch processing

10. **pkg/scanner/serializer.go**
    - Gob serialization
    - Versioning

11. **pkg/scanner/metrics.go**
    - Performance metrics
    - Cache statistics

12. **pkg/scanner/watcher.go**
    - File watching
    - Event handling

13. **pkg/scanner/cmd/main.go** (optional)
    - Standalone scanner CLI tool

14. **pkg/scanner/*_test.go** (multiple test files)
    - Unit tests for each component
    - Integration tests
    - Benchmark tests

15. **pkg/imports/tracker.go**
    - Enhanced ImportTracker
    - Package context integration

16. **pkg/imports/inference.go**
    - Inference rules
    - Context-aware inference

### Modified Files

1. **pkg/config/config.go**
   - Add ScannerConfig section
   - Configuration defaults

2. **pkg/generator/generator.go**
   - Add GenerateWithPackageContext
   - Pass package context to plugins

3. **pkg/generator/plugin/context.go** (if exists)
   - Add PackageContext field

4. **pkg/preprocessor/preprocessor.go**
   - Add SetPackageContext
   - Pass context to processors

5. **pkg/parser/simple.go**
   - Pass package context to preprocessor

6. **cmd/dingo/main.go**
   - Integrate scanner in build flow
   - Add watch mode support

7. **go.mod**
   - Add any new dependencies (if needed)

## Dependencies to Add

**Core Dependencies:**
- None (uses existing Go stdlib)

**Optional Dependencies:**
- `github.com/fsnotify/fsnotify` - File watching (alternative: use inotify/kqueue directly)

## Testing Strategy

### Test Organization

```
pkg/scanner/
  ├── scanner_test.go              # Core tests
  ├── discovery_test.go            # File discovery
  ├── function_index_test.go       # Function indexing
  ├── cache_test.go                # Caching behavior
  ├── incremental_test.go          # Incremental builds
  ├── inference_test.go            # Import inference (in pkg/imports/)
  ├── integration_test.go          # End-to-end
  └── benchmark_test.go            # Performance

tests/golden/
  ├── package_scan_*.dingo         # Package-level tests
  ├── unqualified_import_*.dingo   # Import inference
  └── incremental_*.dingo          # Incremental builds
```

### Test Coverage Requirements

- **Unit tests**: 80%+ code coverage
- **Integration tests**: All major flows
- **Golden tests**: Representative use cases
- **Benchmark tests**: Performance regressions

### Test Data

**Small Package** (10 files):
- Use existing test files
- Add a few more for coverage

**Medium Package** (50 files):
- Create synthetic test package
- Or use a real small Go project

**Large Package** (200 files):
- Synthetic package with repeated patterns
- Or real-world project (with permission)

## Deployment Plan

### Versioning

- **v0.2**: Phase 1-2 (basic scanning, cache)
- **v0.3**: Phase 3 (incremental)
- **v0.4**: Phase 4 (import inference)
- **v0.5**: Phase 5-6 (optimization, production)

### Migration Strategy

1. **Backward Compatibility**:
   - Scanner is optional (disabled by default initially)
   - Can be enabled via config flag
   - Old behavior preserved when disabled

2. **Gradual Rollout**:
   - Alpha: Limited testing
   - Beta: Opt-in for users
   - GA: Default enabled

3. **Migration Path**:
   - Existing code continues to work
   - Users can opt-in to package scanning
   - Documentation for enabling new features

### Rollback Plan

- All configuration is reversible
- Cache can be cleared with flag
- Feature flags for quick disable
- CLI flag to disable scanning: `--no-package-scan`

## Risk Assessment

### High-Risk Components

1. **Cache Corruption**
   - Risk: Invalid cache causes incorrect inference
   - Mitigation: Version checksums, fallback to cache miss
   - Test: Corrupt cache scenarios

2. **Performance Degradation**
   - Risk: Scanning slower than expected
   - Mitigation: Progressive optimization, fallback to cache
   - Test: Benchmark comparisons

3. **Memory Leaks**
   - Risk: Cache growth causes OOM
   - Mitigation: LRU limits, memory monitoring
   - Test: Long-running processes

4. **Race Conditions**
   - Risk: Concurrent access to cache
   - Mitigation: Proper mutexes, concurrent tests
   - Test: Concurrent read/write scenarios

### Medium-Risk Components

1. **Incorrect Import Inference**
   - Risk: False positives/negatives
   - Mitigation: Conservative defaults, user override
   - Test: Golden tests with edge cases

2. **Cache Invalidation Bugs**
   - Risk: Stale data used
   - Mitigation: Hash-based validation
   - Test: Change detection scenarios

### Low-Risk Components

1. **Disk Cache Format Changes**
   - Risk: Cache versions incompatible
   - Mitigation: Version in cache header
   - Test: Cache migration tests

## Documentation Requirements

### User Documentation

1. **Configuration Guide**
   - How to enable package scanning
   - Configuration options
   - Performance tuning

2. **Performance Guide**
   - Expected performance numbers
   - Optimization tips
   - Troubleshooting slow scans

### Developer Documentation

1. **Architecture Guide**
   - High-level design
   - Component interactions
   - Extension points

2. **Implementation Guide**
   - Step-by-step implementation
   - Code walkthrough
   - Testing guidelines

## Timeline Summary

| Phase | Duration | Key Deliverable |
|-------|----------|-----------------|
| Phase 1 | 2 weeks | Basic package scanning |
| Phase 2 | 1 week | Persistent cache |
| Phase 3 | 2 weeks | Incremental builds |
| Phase 4 | 2 weeks | Import inference |
| Phase 5 | 1 week | Performance optimization |
| Phase 6 | 1 week | CLI integration |
| **Total** | **9 weeks** | **Production-ready implementation** |

## Success Metrics

### Functional Metrics

- [ ] Can scan packages up to 200 files
- [ ] Import inference accuracy >95%
- [ ] No false positives in import inference
- [ ] Golden tests pass (100%)

### Performance Metrics

- [ ] 10-file package: <500ms
- [ ] 50-file package: <2s
- [ ] 200-file package: <8s
- [ ] Incremental rebuild: <150ms
- [ ] Cache hit: <50ms

### Quality Metrics

- [ ] Test coverage >80%
- [ ] No race conditions (go test -race)
- [ ] Memory usage <200MB for 200 files
- [ ] No goroutine leaks

### Adoption Metrics

- [ ] CLI integration seamless
- [ ] Watch mode functional
- [ ] No regressions in existing features
- [ ] Documentation complete

This implementation plan provides a clear roadmap to production-ready package-wide scanning with unqualified import inference support.
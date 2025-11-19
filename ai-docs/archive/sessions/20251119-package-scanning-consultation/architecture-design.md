# Package-Wide Scanning Architecture Design

## Overview

This document outlines the production-ready architecture for package-wide scanning to support unqualified import inference in the Dingo transpiler. The design extends the existing two-stage transpiler architecture (preprocessor → go/parser → AST plugins) to support package-level context analysis.

## Current Architecture Analysis

### Existing Pipeline
The current transpiler processes files individually:

```
.dingo file
  ↓
Preprocessor (text-based transformations)
  ↓
go/parser (parse to AST)
  ↓
AST Plugin Pipeline (Result, Option, Pattern Matching)
  ↓
.go file + .sourcemap
```

### Key Limitations
1. **Single-file processing**: No awareness of other files in the package
2. **No function index**: Cannot infer function calls without explicit import
3. **No package cache**: Must scan dependencies repeatedly
4. **Missing context**: Type inference limited to single-file scope

## Proposed Architecture

### 1. Package Scanner Component

#### Core Interface

```go
// pkg/scanner/package_scanner.go
package scanner

type PackageScanner struct {
    config     *config.Config
    cache      *PackageCache
    fileSet    *token.FileSet
    logger     plugin.Logger
}

type Scanner interface {
    // ScanPackage discovers and analyzes all files in a package
    ScanPackage(ctx context.Context, pkgPath string) (*PackageInfo, error)

    // ScanFiles scans specific files with package context
    ScanFiles(ctx context.Context, files []string) (*PackageInfo, error)

    // GetFunctionIndex returns cached function declarations
    GetFunctionIndex(pkgPath string) *FunctionIndex

    // InvalidateCache removes stale package data
    InvalidateCache(pkgPath string, files []string) error
}
```

#### PackageInfo Structure

```go
// pkg/scanner/package_info.go
type PackageInfo struct {
    PackagePath    string                 // "github.com/user/project"
    PackageName    string                 // "main"
    Files          []FileInfo             // All .dingo files in package
    Imports        map[string]string      // Import path -> alias/method
    FunctionIndex  map[string][]FunctionDecl  // Function name -> declarations
    Dependencies   map[string]bool        // Required packages
    Timestamp      time.Time              // Last scan time
    SourceMap      map[string]*SourceMap  // Per-file source maps
}

// FunctionDecl represents a function declaration
type FunctionDecl struct {
    Name        string      // "ParseUser"
    Package     string      // "github.com/user/project/auth"
    File        string      // "auth/user.go"
    Line        int         // 42
    Signature   string      // "func ParseUser(data []byte) (User, error)"
    ReturnTypes []string    // ["User", "error"]
    IsExported  bool        // true
    GoDoc       string      // "ParseUser parses user data from bytes"
}
```

#### Function Index Data Structure

```go
// pkg/scanner/function_index.go
type FunctionIndex struct {
    functionsByName     map[string][]*QualifiedFunction
    functionsByPackage  map[string][]*QualifiedFunction
    typeIndex           map[string]*TypeInfo
}

type QualifiedFunction struct {
    Name        string
    Package     string
    ImportPath  string
    Signature   string
    File        string
    Line        int
    Receiver    *TypeInfo    // If method, the receiver type
    Parameters  []*TypeInfo
    Returns     []*TypeInfo
    IsGeneric   bool
    TypeParams  []string
}

type TypeInfo struct {
    Name        string
    Package     string
    ImportPath  string
    Kind        string      // "struct", "interface", "type alias"
    Fields      []*Field
    Methods     []*QualifiedFunction
    Underlying  string      // For type aliases
}
```

### 2. Integration with ImportTracker

#### Enhanced ImportTracker

```go
// pkg/imports/tracker.go
type ImportTracker struct {
    scanner      Scanner          // Package-wide scanner
    config       *config.Config
    cache        *PackageCache
    // Track what we've learned from package scan
    learnedFuncs map[string][]*QualifiedFunction
}

type ImportInferenceRule int

const (
    // Rule 1: Exact package name match
    // "auth.ParseUser" where auth is imported as "auth"
    InferExactPackage ImportInferenceRule = iota

    // Rule 2: Last segment match
    // "ParseUser" where package "github.com/user/project/auth" imported
    InferLastSegment

    // Rule 3: Alias match
    // "ParseUser" where imported with alias: "auth2.ParseUser"
    InferAlias

    // Rule 4: Context-aware inference (Type 4 - Advanced)
    // Infer based on return type expectations
    InferContextAware
)

func (t *ImportTracker) InferUnqualifiedUsage(
    file *ast.File,
    funcName string,
    pos token.Pos,
    expectedTypes []string,
) (*InferenceResult, error) {
    // 1. Get package function index
    pkgInfo, err := t.scanner.ScanPackage(context.Background(), t.getCurrentPackage())
    if err != nil {
        return nil, err
    }

    // 2. Find matching functions
    matches := t.findMatchingFunctions(pkgInfo.FunctionIndex, funcName, expectedTypes)

    // 3. Apply inference rules
    result := t.applyInferenceRules(matches, file.Imports)

    return result, nil
}
```

### 3. Cache Layer

#### Two-Level Caching Strategy

```go
// pkg/scanner/cache.go
type PackageCache struct {
    memoryCache  *InMemoryCache    // Fast access, ephemeral
    diskCache    *DiskCache        // Persistent, slower
    config       CacheConfig
    mutex        sync.RWMutex
}

type CacheConfig struct {
    // Memory limits
    MaxMemoryEntries int           // Default: 1000 packages
    MaxMemoryMB      int           // Default: 512 MB

    // Disk cache
    CacheDir         string        // Default: ~/.dingo/cache
    MaxDiskGB        int           // Default: 5 GB

    // TTL settings
    DefaultTTL       time.Duration // Default: 1 hour
    MaxTTL           time.Duration // Default: 24 hours
}

type CacheEntry struct {
    PackageInfo   *PackageInfo
    FileHashes    map[string]string    // File path -> hash
    DependencyHash string              // Hash of dependencies
    CreatedAt     time.Time
    LastAccessed  time.Time
    AccessCount   int64
    SizeBytes     int64
}
```

#### In-Memory Cache Implementation

```go
// pkg/scanner/memory_cache.go
type InMemoryCache struct {
    packages map[string]*CacheEntry
    lru      *list.List    // For LRU eviction
    size     int64
    config   MemoryConfig
    mutex    sync.RWMutex
}

func (c *InMemoryCache) Get(pkgPath string) (*PackageInfo, bool) {
    c.mutex.Lock()
    defer c.mutex.Unlock()

    entry, exists := c.packages[pkgPath]
    if !exists {
        return nil, false
    }

    // Update LRU
    c.lru.MoveToFront(entry.lruNode)

    // Update access stats
    entry.LastAccessed = time.Now()
    entry.AccessCount++

    return entry.PackageInfo, true
}

func (c *InMemoryCache) Put(pkgPath string, info *PackageInfo) {
    c.mutex.Lock()
    defer c.mutex.Unlock()

    // Check if we need to evict
    if c.shouldEvict() {
        c.evictOldest()
    }

    // Add to cache
    entry := &CacheEntry{
        PackageInfo: info,
        CreatedAt:   time.Now(),
        LastAccessed: time.Now(),
        AccessCount: 1,
        SizeBytes:   estimateSize(info),
    }

    c.packages[pkgPath] = entry
    c.lru.PushFront(entry)
    c.size += entry.SizeBytes
}
```

#### Disk Cache Implementation

```go
// pkg/scanner/disk_cache.go
type DiskCache struct {
    baseDir  string
    config   DiskConfig
    serializer *GobSerializer  // Or JSON, MessagePack
}

func (c *DiskCache) Get(pkgPath string) (*PackageInfo, error) {
    filePath := c.getCacheFilePath(pkgPath)

    data, err := os.ReadFile(filePath)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, nil  // Not cached
        }
        return nil, err
    }

    var info PackageInfo
    if err := c.serializer.Deserialize(data, &info); err != nil {
        return nil, fmt.Errorf("failed to deserialize cache: %w", err)
    }

    return &info, nil
}

func (c *DiskCache) Put(pkgPath string, info *PackageInfo) error {
    data, err := c.serializer.Serialize(info)
    if err != nil {
        return fmt.Errorf("failed to serialize cache: %w", err)
    }

    filePath := c.getCacheFilePath(pkgPath)
    dir := filepath.Dir(filePath)

    if err := os.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("failed to create cache directory: %w", err)
    }

    if err := os.WriteFile(filePath, data, 0644); err != nil {
        return fmt.Errorf("failed to write cache file: %w", err)
    }

    return nil
}
```

### 4. Data Flow Design

#### Complete Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│ Package Discovery Phase                                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│ 1. Scan directory for .dingo files                              │
│    ├─ Read go.mod to identify package name/path                │
│    ├─ Build list of source files                               │
│    ├─ Calculate file hashes                                    │
│    └─ Check cache validity                                     │
│                                                                 │
│ 2. Dependency Analysis                                          │
│    ├─ Parse import statements                                  │
│    ├─ Identify external dependencies                           │
│    ├─ Build dependency graph                                   │
│    └─ Determine scan order                                     │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────────┐
│ Function Index Build Phase                                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│ For each file in dependency order:                              │
│                                                                 │
│ 3. Preprocessing                                                 │
│    ├─ Run Dingo preprocessors                                  │
│    ├─ Generate valid Go code                                   │
│    └─ Capture transformations                                  │
│                                                                 │
│ 4. AST Analysis                                                  │
│    ├─ Parse with go/parser                                     │
│    ├─ Run go/types type checker                                │
│    ├─ Extract function declarations                            │
│    ├─ Extract type definitions                                 │
│    ├─ Index method receivers                                   │
│    └─ Build signature information                              │
│                                                                 │
│ 5. Cross-Reference Resolution                                    │
│    ├─ Resolve imported types                                   │
│    ├─ Build cross-file call graph                              │
│    ├─ Identify exported vs unexported                          │
│    └─ Validate signatures                                      │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────────┐
│ Cache Storage Phase                                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│ 6. Cache Write                                                  │
│    ├─ Store in memory cache (LRU)                              │
│    ├─ Serialize to disk cache                                  │
│    ├─ Update TTL based on access patterns                      │
│    └─ Record cache statistics                                  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────────┐
│ Per-File Transpilation Phase                                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│ For each .dingo file:                                           │
│                                                                 │
│ 7. Package Context Injection                                    │
│    ├─ Load function index for package                          │
│    ├─ Inject into ImportTracker                                │
│    ├─ Enable context-aware inference                           │
│    └─ Pass to preprocessor                                     │
│                                                                 │
│ 8. Transpilation with Context                                   │
│    ├─ Run preprocessors                                        │
│    ├─ Parse to AST                                             │
│    ├─ Run plugins (with package context)                       │
│    ├─ Generate Go code                                         │
│    └─ Write output                                             │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

#### Detailed Step-by-Step

**Step 1: Package Discovery**
```go
func (s *PackageScanner) ScanPackage(ctx context.Context, pkgPath string) (*PackageInfo, error) {
    // 1. Check cache first
    if cached, hit := s.cache.Get(pkgPath); hit {
        return cached, nil
    }

    // 2. Discover files
    files, err := s.discoverPackageFiles(pkgPath)
    if err != nil {
        return nil, fmt.Errorf("failed to discover files: %w", err)
    }

    // 3. Analyze dependencies
    deps, err := s.analyzeDependencies(files)
    if err != nil {
        return nil, fmt.Errorf("failed to analyze dependencies: %w", err)
    }

    // 4. Build function index
    funcIndex, err := s.buildFunctionIndex(ctx, files)
    if err != nil {
        return nil, fmt.Errorf("failed to build function index: %w", err)
    }

    // 5. Create PackageInfo
    pkgInfo := &PackageInfo{
        PackagePath:   pkgPath,
        Files:         files,
        FunctionIndex: funcIndex,
        Dependencies:  deps,
        Timestamp:     time.Now(),
    }

    // 6. Cache it
    s.cache.Put(pkgPath, pkgInfo)

    return pkgInfo, nil
}
```

**Step 2: Function Index Building**
```go
func (s *PackageScanner) buildFunctionIndex(ctx context.Context, files []FileInfo) (map[string][]FunctionDecl, error) {
    funcIndex := make(map[string][]FunctionDecl)
    var allErrs []error

    for _, file := range files {
        // Preprocess
        prep := preprocessor.NewWithMainConfig(file.Source, s.config)
        goSource, _, err := prep.Process()
        if err != nil {
            allErrs = append(allErrs, err)
            continue
        }

        // Parse
        fset := token.NewFileSet()
        astFile, err := parser.ParseFile(fset, file.Path, []byte(goSource), parser.ParseComments)
        if err != nil {
            allErrs = append(allErrs, err)
            continue
        }

        // Type check for accurate signatures
        typeInfo, err := runTypeChecker(astFile)
        if err != nil {
            s.logger.Warn("Type checking failed for %s: %v", file.Path, err)
        }

        // Extract functions
        functions := extractFunctions(astFile, typeInfo, file.Path, fset)
        for _, fn := range functions {
            funcIndex[fn.Name] = append(funcIndex[fn.Name], fn)
        }
    }

    if len(allErrs) > 0 {
        return nil, fmt.Errorf("failed to index %d files: %v", len(allErrs), allErrs)
    }

    return funcIndex, nil
}

func extractFunctions(file *ast.File, typeInfo *types.Info, filePath string, fset *token.FileSet) []FunctionDecl {
    var functions []FunctionDecl

    ast.Inspect(file, func(n ast.Node) bool {
        decl, ok := n.(*ast.FuncDecl)
        if !ok {
            return true
        }

        fn := FunctionDecl{
            Name:        decl.Name.Name,
            File:        filePath,
            Line:        fset.Position(decl.Pos()).Line,
            IsExported:  decl.Name.IsExported(),
            Receiver:    extractReceiver(decl.Recv, typeInfo),
            Parameters:  extractParameters(decl.Type.Params, typeInfo),
            Returns:     extractReturns(decl.Type.Results, typeInfo),
        }

        // Get signature string
        var sigBuf strings.Builder
        sigBuf.WriteString("func ")
        sigBuf.WriteString(fn.Name)
        sigBuf.WriteString("(")
        // ... add parameters
        sigBuf.WriteString(")")

        if len(fn.Returns) > 0 {
            sigBuf.WriteString(" (")
            for i, ret := range fn.Returns {
                if i > 0 {
                    sigBuf.WriteString(", ")
                }
                sigBuf.WriteString(ret.Name)
            }
            sigBuf.WriteString(")")
        }

        fn.Signature = sigBuf.String()

        functions = append(functions, fn)
        return true
    })

    return functions
}
```

**Step 3: Integration with Transpiler**
```go
// pkg/generator/generator.go (enhanced)
func (g *Generator) GenerateWithPackageContext(file *dingoast.File, pkgContext *PackageInfo) ([]byte, error) {
    // 1. Set package context
    if g.pipeline != nil && g.pipeline.Ctx != nil {
        g.pipeline.Ctx.PackageInfo = pkgContext

        // Enhance ImportTracker with function index
        if pkgContext != nil {
            g.pipeline.Ctx.ImportTracker = &ImportTracker{
                FunctionIndex: pkgContext.FunctionIndex,
                Config:        g.pipeline.Ctx.Config,
            }
        }
    }

    // 2. Continue with normal generation
    return g.Generate(file)
}
```

### 5. Incremental Build Handling

#### Change Detection

```go
// pkg/scanner/incremental.go
type IncrementalScanner struct {
    baseScanner *PackageScanner
    fileTracker *FileTracker
}

type FileTracker struct {
    lastHashes    map[string]string
    lastScanTime  time.Time
    watchedPaths  map[string]bool
    mutex         sync.RWMutex
}

func (s *IncrementalScanner) ScanWithIncrementalSupport(ctx context.Context, pkgPath string) (*PackageInfo, error) {
    // 1. Check for file changes
    changes, err := s.detectFileChanges(pkgPath)
    if err != nil {
        return nil, err
    }

    // 2. If no changes, use cached version
    if len(changes.ChangedFiles) == 0 && len(changes.DeletedFiles) == 0 && len(changes.NewFiles) == 0 {
        cached, hit := s.baseScanner.cache.Get(pkgPath)
        if hit {
            return cached, nil
        }
    }

    // 3. Determine what needs re-scanning
    affectedFiles := s.calculateAffectedFiles(changes)

    // 4. Re-scan only affected parts
    pkgInfo, err := s.partialRescan(ctx, pkgPath, affectedFiles, changes)
    if err != nil {
        return nil, err
    }

    // 5. Update trackers
    s.fileTracker.update(pkgPath, pkgInfo)

    return pkgInfo, nil
}

func (s *IncrementalScanner) detectFileChanges(pkgPath string) (*Changes, error) {
    s.fileTracker.mutex.RLock()
    defer s.fileTracker.mutex.RUnlock()

    oldHashes := s.fileTracker.lastHashes
    newHashes, err := s.calculateFileHashes(pkgPath)
    if err != nil {
        return nil, err
    }

    changes := &Changes{
        ChangedFiles: []string{},
        NewFiles:     []string{},
        DeletedFiles: []string{},
    }

    // Check for changed or new files
    for path, newHash := range newHashes {
        oldHash, exists := oldHashes[path]
        if !exists {
            changes.NewFiles = append(changes.NewFiles, path)
        } else if oldHash != newHash {
            changes.ChangedFiles = append(changes.ChangedFiles, path)
        }
    }

    // Check for deleted files
    for path := range oldHashes {
        if _, exists := newHashes[path]; !exists {
            changes.DeletedFiles = append(changes.DeletedFiles, path)
        }
    }

    return changes, nil
}
```

#### Dependency Graph

```go
// pkg/scanner/deps.go
type DependencyGraph struct {
    edges     map[string][]string      // pkg -> [dependent packages]
    inverted  map[string][]string      // pkg -> [dependencies]
    mutex     sync.RWMutex
}

func (g *DependencyGraph) AddDependency(pkg, dep string) {
    g.mutex.Lock()
    defer g.mutex.Unlock()

    if g.edges == nil {
        g.edges = make(map[string][]string)
        g.inverted = make(map[string][]string)
    }

    g.edges[pkg] = append(g.edges[pkg], dep)
    g.inverted[dep] = append(g.inverted[dep], pkg)
}

func (g *DependencyGraph) GetImpactedPackages(changedPkg string) []string {
    g.mutex.RLock()
    defer g.mutex.RUnlock()

    var impacted []string
    visited := make(map[string]bool)

    var dfs func(pkg string)
    dfs = func(pkg string) {
        if visited[pkg] {
            return
        }
        visited[pkg] = true

        // Add dependents of this package
        for _, dep := range g.edges[pkg] {
            if !visited[dep] {
                impacted = append(impacted, dep)
                dfs(dep)
            }
        }
    }

    dfs(changedPkg)
    return impacted
}
```

## Implementation Details

### Error Handling Strategy

```go
type ScannerError struct {
    PackagePath string
    File        string
    Operation   string
    Cause       error
    Context     map[string]interface{}
}

func (e *ScannerError) Error() string {
    return fmt.Sprintf("scanner[%s:%s:%s]: %v",
        e.PackagePath, e.File, e.Operation, e.Cause)
}

// Wrap errors with context
func (s *PackageScanner) scanWithErrorHandling(ctx context.Context, pkgPath string) (*PackageInfo, error) {
    defer func() {
        if r := recover(); r != nil {
            s.logger.Error("Panic during package scan: %v", r)
        }
    }()

    // Set timeout
    scanCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    // Check context cancellation
    select {
    case <-scanCtx.Done():
        return nil, fmt.Errorf("scan cancelled: %w", scanCtx.Err())
    default:
    }

    return s.scanPackage(scanCtx, pkgPath)
}
```

### Logging and Metrics

```go
type ScannerMetrics struct {
    ScansTotal          prometheus.Counter
    ScanDuration        prometheus.Histogram
    CacheHits           prometheus.Counter
    CacheMisses         prometheus.Counter
    FilesIndexed        prometheus.Counter
    FunctionsDiscovered prometheus.Counter
    ErrorsTotal         prometheus.Counter
}

func (s *PackageScanner) recordMetrics(metrics ScannerMetrics, pkgPath string, duration time.Duration) {
    metrics.ScanDuration.Observe(duration.Seconds())
    metrics.FilesIndexed.Add(float64(len(s.lastScan.Files)))
    metrics.FunctionsDiscovered.Add(float64(len(s.lastScan.FunctionIndex)))
}
```

## Configuration

```go
// pkg/config/scanner.go
type ScannerConfig struct {
    // Cache settings
    Cache *CacheConfig

    // Performance tuning
    MaxConcurrentScans int           // Default: GOMAXPROCS
    ScanTimeout        time.Duration // Default: 30s

    // Incremental builds
    EnableIncremental  bool          // Default: true
    WatchPaths         []string      // Paths to watch for changes

    // Filtering
    ExcludePaths       []string      // Glob patterns to exclude
    IncludeTests       bool          // Default: true

    // go/packages integration
    UseGoPackages      bool          // Default: false (opt-in)
    GoPackagesMode     string        // "load", "need", "export"
}
```

## Testing Strategy

### Unit Tests
- Function extraction from various Go constructs
- Cache serialization/deserialization
- Incremental change detection
- Import inference rules

### Integration Tests
- End-to-end package scan
- Multi-file function discovery
- Cross-package dependencies
- Watch mode behavior

### Performance Tests
- Large package scanning (500+ files)
- Memory usage under load
- Cache hit rates
- Incremental rebuild speeds

### Golden Tests
- Golden files for package scan outputs
- Comparison of inferred vs explicit imports
- Cache persistence verification

## Deployment Considerations

### Cache Location
- **Development**: `~/.dingo/cache/`
- **CI/CD**: `/tmp/dingo-cache-{build-id}/`
- **Docker**: Volume mount for persistence
- **LSP**: Separate cache per workspace

### Permissions
- Cache directory: 0755
- Cache files: 0644
- User-specific cache in home directory

### Cleanup
- Automatic LRU eviction based on size limits
- TTL-based expiration
- Manual cache clear command
- Disk space monitoring

## Summary

This architecture provides:

1. **Complete package awareness** via function index
2. **Efficient caching** at memory and disk levels
3. **Incremental builds** with smart change detection
4. **Production-ready** error handling and metrics
5. **Clean integration** with existing transpiler pipeline
6. **Scalable design** for large codebases

The design maintains compatibility with the existing two-stage architecture while adding the necessary package-wide context for unqualified import inference.
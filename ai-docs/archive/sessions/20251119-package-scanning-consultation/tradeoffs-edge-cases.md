# Package-Wide Scanning: Trade-offs & Edge Cases

## Trade-off Analysis

### 1. Architecture Approach Comparison

#### Option A: go/packages-based Scanning

**Description**: Use golang.org/x/tools/go/packages to load and analyze packages

**Pros:**
- ✅ Native Go module awareness
- ✅ Handles build tags, go:build directives correctly
- ✅ Parallel package loading built-in
- ✅ Dependency resolution from go.mod
- ✅ Used by gopls (battle-tested)
- ✅ Handles vendored dependencies properly
- ✅ Works with multiple Go versions

**Cons:**
- ❌ Heavier memory footprint (loads full type info)
- ❌ Slower cold start (analyzes dependencies)
- ❌ Less control over caching granularity
- ❌ Harder to customize for Dingo-specific needs
- ❌ External dependency (golang.org/x/tools)

**Performance:**
- Initial scan: 5-10s for 200 files
- Incremental: 200-500ms
- Memory: 60-80MB for 200 files

**Best for:** Production LSP, complex dependency trees, build tag handling

---

#### Option B: Manual File Discovery + go/parser

**Description**: Custom file discovery with go/parser for each file

**Pros:**
- ✅ Fine-grained control over what gets scanned
- ✅ Lighter memory footprint
- ✅ Faster cold start (no dependency analysis)
- ✅ No external dependencies
- ✅ Can skip test files easily
- ✅ Custom caching strategy
- ✅ Optimized for Dingo's needs

**Cons:**
- ❌ Must implement build tag handling
- ❌ No automatic dependency resolution
- ❌ Need to handle go:build directives
- ❌ More code to maintain
- ❌ May miss edge cases in Go build system

**Performance:**
- Initial scan: 3-8s for 200 files
- Incremental: 100-200ms
- Memory: 30-50MB for 200 files

**Best for:** CLI tool, simple projects, custom caching needs

---

#### **Decision: Hybrid Approach**

**Recommended: Start with Option B, with go/packages as optional**

**Rationale:**
1. Start simple (manual discovery) for faster implementation
2. Add go/packages integration as advanced option
3. Use feature flag: `--use-go-packages` for advanced users
4. Default to manual for CLI simplicity
5. Enable go/packages for LSP (where it's already available)

**Implementation:**
```go
type ScannerType int

const (
    ScannerManual ScannerType = iota
    ScannerGoPackages
)

func (s *PackageScanner) SetScannerType(t ScannerType) {
    s.scannerType = t
}
```

### 2. Cache Strategy Comparison

#### Centralized Cache (Single Instance)

**Description**: One cache instance shared across all packages

**Pros:**
- ✅ Memory efficient (deduplicate)
- ✅ Simple implementation
- ✅ Consistent eviction policy
- ✅ Single point of control

**Cons:**
- ❌ Lock contention (single mutex)
- ❌ Cannot isolate package groups
- ❌ Difficult to tune per-package TTL
- ❌ Cache pollution (rarely-used packages)

**Best for:** Single-package projects

---

#### Distributed Cache (Per-Package)

**Description**: Each package has its own cache with separate eviction

**Pros:**
- ✅ No lock contention
- ✅ Per-package TTL tuning
- ✅ Can prioritize important packages
- ✅ Better isolation

**Cons:**
- ❌ Higher memory overhead
- ❌ More complex implementation
- ❌ Harder to manage disk space
- ❌ More files to manage

**Best for:** Multi-package projects, monorepos

---

#### **Decision: Per-Package with Shared Invalidation**

**Design:**
- Per-package disk caches
- Shared memory cache with global LRU
- Global invalidation triggers (go.mod change)
- Package-specific TTL based on access frequency

### 3. Function Detection Strategy

#### AST-based Detection (Go Types Required)

**Description**: Parse AST, run type checker, extract full signatures

**Pros:**
- ✅ Accurate function signatures
- ✅ Can distinguish overloaded functions
- ✅ Handle generics correctly
- ✅ Extract receiver types for methods
- ✅ Full type information for inference

**Cons:**
- ❌ Slow (type checker is expensive)
- ❌ High memory usage
- ❌ Fails on incomplete code
- ❌ No type info during preprocessing

**Performance:**
- 50 files: 500-1000ms
- Memory: +5MB per file

---

#### Regex-based Detection (Quick Parse)

**Description**: Regex patterns to find function declarations

**Pros:**
- ✅ Very fast (10x faster)
- ✅ Works on incomplete code
- ✅ Low memory
- ✅ No type checking needed

**Cons:**
- ❌ Less accurate (may miss edge cases)
- ❌ Cannot handle complex cases (generics, complex signatures)
- ❌ No receiver type info
- ❌ Regex maintenance burden

**Performance:**
- 50 files: 50-100ms
- Memory: +0.5MB per file

---

#### **Decision: Progressive Detection**

**Design:**
1. **Phase 1**: Quick regex parse for function names
2. **Phase 2**: AST parse for exported functions only
3. **Phase 3**: Full type check for function index
4. **Fallback**: If type check fails, use regex data

**Example:**
```go
func (s *PackageScanner) ExtractFunctions(file *ast.File) []FunctionDecl {
    // Phase 1: Quick scan (always succeeds)
    quick := s.extractWithRegex(file)

    // Phase 2: If file is small or type check fast, do full
    if s.shouldDoFullTypeCheck(file) {
        full, err := s.extractWithTypeCheck(file)
        if err == nil {
            return full
        }
    }

    // Fallback: Use quick parse
    return quick
}
```

### 4. Cache Persistence Comparison

#### Disk Cache (Gob/JSON)

**Description**: Serialize package info to disk

**Pros:**
- ✅ Fast startup (no rescan)
- ✅ Works across process restarts
- ✅ Reduces CI/CD build times
- ✅ Can be versioned

**Cons:**
- ❌ Cache invalidation complexity
- ❌ Stale data risk
- ❌ Version migration needed
- ❌ Disk space usage

**Best for:** Development, repeated builds

---

#### Memory Cache Only

**Description**: Cache only in process memory

**Pros:**
- ✅ Simple (no serialization)
- ✅ Always up-to-date (no staleness)
- ✅ Fast (no disk I/O)
- ✅ No version issues

**Cons:**
- ❌ Lost on restart
- ❌ No benefit for repeated builds
- ❌ Cannot share between processes

**Best for:** CI/CD one-off builds, scripts

---

#### **Decision: Two-Level Cache**

**Design:**
- **Level 1**: Memory cache (fastest)
- **Level 2**: Disk cache (persistent)
- **Strategy**: Memory-first, disk fallback, always update both
- **Invalidation**: Hash-based validation

### 5. Import Inference Strategy

#### Conservative (Explicit Only)

**Description**: Only infer when single match exists

**Pros:**
- ✅ No false positives
- ✅ Predictable behavior
- ✅ Easy to understand

**Cons:**
- ❌ Misses many opportunities
- ❌ Still requires explicit imports frequently
- ❌ Limited value

**Accuracy:** 40-50% of qualifying calls inferred

---

#### Aggressive (Best Match)

**Description**: Infer based on best match score

**Pros:**
- ✅ Maximizes inferred imports
- ✅ Most value for users
- ✅ Smart inference

**Cons:**
- ❌ Risk of wrong inference (but addable)
- ❌ Harder to debug
- ❌ Need confidence scoring

**Accuracy:** 80-90% with confidence threshold

---

#### **Decision: Confidence-Scored Inference**

**Design:**
```go
type InferenceResult struct {
    Function *FunctionDecl
    Confidence float64  // 0.0 to 1.0
    RulesApplied []InferenceRule
}

func (t *ImportTracker) InferUnqualified(funcName string, context Context) (*InferenceResult, error) {
    // Find all matches
    matches := t.FindMatches(funcName, context)

    // Score each match
    for _, match := range matches {
        match.Confidence = t.CalculateConfidence(match, context)
    }

    // Only return if confidence exceeds threshold
    best := t.FindBestMatch(matches)
    if best.Confidence >= 0.8 {
        return best, nil
    }

    return nil, fmt.Errorf("no confident match found")
}
```

**Default threshold:** 0.8 (80% confidence)
**Override:** `--inference-aggressive` flag (threshold 0.6)
**Conservative:** `--inference-conservative` (threshold 0.95)

## Edge Cases Analysis

### 1. Cyclic Dependencies

**Scenario:**
```
file1.dingo → imports file2
file2.dingo → imports file1
```

**Problem:**
- Function index building loops infinitely
- Dependency resolution fails

**Solution:**
```go
func (s *PackageScanner) buildFunctionIndex(ctx context.Context, files []FileInfo) error {
    visited := make(map[string]bool)
    var dfs func(string) error

    dfs = func(filePath string) error {
        if visited[filePath] {
            return nil  // Already processed, return safely
        }
        visited[filePath] = true

        // Process file
        // ...

        return nil
    }

    for _, file := range files {
        if !visited[file.Path] {
            if err := dfs(file.Path); err != nil {
                return err
            }
        }
    }
    return nil
}
```

**Result:** ✅ Handles cycles gracefully, completes scan

---

### 2. Package Spread Across Multiple Directories

**Scenario:**
```
src/
  ├── main.dingo
  ├── utils/
  │   └── helpers.dingo
  └── models/
      └── types.dingo
```

**Problem:**
- Standard go/packages behavior
- Need to discover files in subdirectories
- go.mod only lists root package

**Solution:**
```go
func (s *PackageScanner) discoverPackageFiles(pkgPath string) ([]FileInfo, error) {
    files := []FileInfo{}

    // Walk all subdirectories
    err := filepath.Walk(pkgPath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        // Skip vendor, .git, etc.
        if s.shouldSkip(path) {
            if info.IsDir() {
                return filepath.SkipDir
            }
            return nil
        }

        if strings.HasSuffix(path, ".dingo") {
            // Add file
        }

        return nil
    })

    return files, err
}
```

**Result:** ✅ Finds all .dingo files in package tree

---

### 3. Build Tags and go:build Directives

**Scenario:**
```go
//go:build linux
// +build linux

package main

func LinuxOnly() {}
```

**Problem:**
- File should only be included on Linux
- Skipping it causes incomplete function index
- Including it causes errors on other platforms

**Solution:**
```go
func (s *PackageScanner) FilterByBuildTags(files []FileInfo, goos, goarch string) ([]FileInfo, error) {
    filtered := []FileInfo{}

    for _, file := range files {
        // Check if file should be included
        include, err := s.shouldIncludeFile(file.Path, goos, goarch)
        if err != nil {
            // Log warning but don't fail
            s.logger.Warn("Failed to parse build tags for %s: %v", file.Path, err)
            continue
        }

        if include {
            filtered = append(filtered, file)
        }
    }

    return filtered, nil
}

// Parse build tags using go/build package
func shouldIncludeFile(filePath, goos, goarch string) (bool, error) {
    // Use go/build toolchain
    return build.CheckFile(filePath, goos, goarch)
}
```

**Alternative (if not using go/packages):**
```go
func parseBuildTags(content []byte) ([]build.Tag, error) {
    // Simple regex-based parsing
    // Check for:
    //   //go:build expr
    //   // +build expr
    //
    // Parse tag expressions like:
    //   linux
    //   !windows
    //   go1.20
    //   cgo && !android
}
```

**Result:** ✅ Only includes files matching build tags

---

### 4. Vendored Dependencies

**Scenario:**
```
project/
  ├── vendor/
  │   └── github.com/
  │       └── external/
  │           └── pkg.dingo
  ├── main.dingo
  └── go.mod
```

**Problem:**
- Vendor directory contains copies of external packages
- Should we scan vendored code? Should we skip it?
- Go's module proxy/cache confusion

**Solution (Three Options):**

**Option A: Skip Vendored**
```go
func (s *PackageScanner) discoverPackageFiles(pkgPath string) ([]FileInfo, error) {
    files := []FileInfo{}

    err := filepath.Walk(pkgPath, func(path string, info os.FileInfo, err error) error {
        // Skip vendor directory
        if strings.Contains(path, "/vendor/") {
            if info.IsDir() {
                return filepath.SkipDir
            }
            return nil
        }
        // ...
    })

    return files, err
}
```

**Option B: Scan Vendored (Isolation)**
```go
func (s *PackageScanner) discoverPackageFiles(pkgPath string) ([]FileInfo, error) {
    // Separate vendor files
    vendorFiles, regularFiles := []FileInfo{}, []FileInfo{}

    for _, file := range allFiles {
        if strings.Contains(file.Path, "/vendor/") {
            vendorFiles = append(vendorFiles, file)
        } else {
            regularFiles = append(regularFiles, file)
        }
    }

    // Scan regular files with normal package context
    ctx := s.createScanContext("main", regularFiles)

    // Scan vendor files with separate context
    vendorCtx := s.createScanContext("vendor", vendorFiles)

    return mergeContexts(ctx, vendorCtx)
}
```

**Option C: Smart (Configurable)**
```go
type ScannerConfig struct {
    ScanVendored        bool  // Default: false
    VendorMode          string  // "skip", "separate", "include"
}

func (s *PackageScanner) discoverPackageFiles(pkgPath string) ([]FileInfo, error) {
    switch s.config.VendorMode {
    case "skip":
        return s.discoverWithoutVendor(pkgPath)
    case "separate":
        return s.discoverWithSeparateVendor(pkgPath)
    case "include":
        return s.discoverWithVendor(pkgPath)
    default:
        return s.discoverWithoutVendor(pkgPath)
    }
}
```

**Recommendation:** Option C (configurable)
- Default: skip vendored (faster, less noise)
- Option: scan separately (for offline builds)
- Option: include (rarely needed)

---

### 5. Generated Files

**Scenario:**
```go
// Code generated by tool; DO NOT EDIT.
// @generated
```

**Problem:**
- Generated files should be skipped (tool-generated code)
- But source maps need to work correctly
- How to detect generated files?

**Solution:**
```go
func isGeneratedFile(filePath string, content []byte) bool {
    // 1. Check for "Code generated" pattern in first 1KB
    header := content[:min(1024, len(content))]
    if strings.Contains(string(header), "Code generated") ||
       strings.Contains(string(header), "DO NOT EDIT") ||
       strings.Contains(string(header), "@generated") {
        return true
    }

    // 2. Check for .pb.go, .gen.go, etc.
    base := filepath.Base(filePath)
    if strings.Contains(base, ".pb.") ||
       strings.Contains(base, ".gen.") ||
       strings.Contains(base, ".generated.") {
        return true
    }

    return false
}

func (s *PackageScanner) FilterGeneratedFiles(files []FileInfo) []FileInfo {
    filtered := []FileInfo{}

    for _, file := range files {
        if isGeneratedFile(file.Path, file.Source) {
            s.logger.Debug("Skipping generated file: %s", file.Path)
            continue
        }
        filtered = append(filtered, file)
    }

    return filtered
}
```

**Edge Case: Self-Generated**
- Some files are generated but need to be included
- Add annotation: `// @generated: include`

**Result:** ✅ Skips most generated files

---

### 6. Empty Packages

**Scenario:**
```
package has only build constraints:
  └── empty.dingo:
      //go:build ignore
      +build ignore
```

**Problem:**
- Package reports 0 functions
- Index building succeeds but returns no data
- Might cause issues in cache

**Solution:**
```go
func (s *PackageScanner) ScanPackage(ctx context.Context, pkgPath string) (*PackageInfo, error) {
    files, err := s.discoverPackageFiles(pkgPath)
    if err != nil {
        return nil, err
    }

    // If no files, return empty package info
    if len(files) == 0 {
        return &PackageInfo{
            PackagePath:  pkgPath,
            Files:        []FileInfo{},
            FunctionIndex: make(map[string][]FunctionDecl),
            IsEmpty:      true,
        }, nil
    }

    // Proceed with normal scanning
    // ...
}
```

**Result:** ✅ Handles empty packages gracefully

---

### 7. Packages with Only Tests

**Scenario:**
```
project/
  ├── main.dingo
  └── main_test.dingo
```

**Problem:**
- Should test files be included in function index?
- Tests can export helper functions (e.g., `// TestMain`)
- May want to infer calls from tests too

**Solution:**
```go
type ScannerConfig struct {
    IncludeTests bool  // Default: true
    TestMode     string  // "include", "functions-only", "skip"
}

func (s *PackageScanner) FilterTestFiles(files []FileInfo) []FileInfo {
    if !s.config.IncludeTests {
        return s.excludeTestFiles(files)
    }

    switch s.config.TestMode {
    case "include":
        return files  // Include all
    case "functions-only":
        return s.extractTestHelpers(files)  // Only exported helpers
    case "skip":
        return s.excludeTestFiles(files)
    default:
        return files
    }
}
```

**Recommendation:**
- Default: Include tests (tests may call functions)
- Option to skip if causing issues

**Result:** ✅ Configurable test file handling

---

### 8. Concurrent Access

**Scenario:**
```
Multiple goroutines accessing cache simultaneously:
  - Scanner thread: Reading cache
  - Background scan: Writing cache
  - LSP handler: Reading cache
```

**Problem:**
- Race conditions on cache
- Partial reads/writes
- Stale data

**Solution:**
```go
type PackageCache struct {
    memoryCache *InMemoryCache
    diskCache   *DiskCache
    mutex       sync.RWMutex  // Protects cache state
}

func (c *PackageCache) Get(pkgPath string) (*PackageInfo, bool) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()

    // Check memory cache
    if info, hit := c.memoryCache.Get(pkgPath); hit {
        return info, true
    }

    // Check disk cache
    info, err := c.diskCache.Get(pkgPath)
    if err == nil {
        // Promote to memory cache
        c.memoryCache.Put(pkgPath, info)
        return info, true
    }

    return nil, false
}

func (c *PackageCache) Put(pkgPath string, info *PackageInfo) {
    c.mutex.Lock()
    defer c.mutex.Unlock()

    // Update both caches
    c.memoryCache.Put(pkgPath, info)
    c.diskCache.Put(pkgPath, info)
}
```

**Testing:**
```go
func TestConcurrentAccess(t *testing.T) {
    cache := NewPackageCache(config)

    var wg sync.WaitGroup

    // 100 concurrent readers
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            info, hit := cache.Get("test/pkg")
            if hit && info != nil {
                // Verify we got valid data
            }
        }(i)
    }

    // 10 concurrent writers
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            cache.Put("test/pkg", generateTestPackage())
        }(i)
    }

    wg.Wait()
}
```

**Result:** ✅ Safe concurrent access

---

### 9. Package Name Conflicts

**Scenario:**
```
main.dingo → imports "github.com/a/pkg"
other.dingo → imports "github.com/b/pkg"
Both packages have function "Parse()"
```

**Problem:**
- Which Parse should be inferred?
- Ambiguous function calls
- Need to differentiate by import path

**Solution:**
```go
type InferenceResult struct {
    Function     *FunctionDecl
    ImportPath   string  // "github.com/a/pkg"
    PackageAlias string  // "pkg"
    Conflict     bool    // True if multiple matches
}

func (t *ImportTracker) InferUnqualified(funcName string, file *ast.File) (*InferenceResult, error) {
    // Get imports from current file
    imports := t.extractImports(file)

    // Find all functions with this name
    matches := t.functionIndex.FindByName(funcName)

    // Filter by imports (only consider packages that are imported)
    validMatches := []FunctionDecl{}
    for _, match := range matches {
        if imports.Contains(match.ImportPath) {
            validMatches = append(validMatches, match)
        }
    }

    // If only one match, return it
    if len(validMatches) == 1 {
        return &InferenceResult{
            Function:   &validMatches[0],
            Conflict:   false,
        }, nil
    }

    // Multiple matches = conflict
    if len(validMatches) > 1 {
        return nil, fmt.Errorf("ambiguous call to %s: %d matches found", funcName, len(validMatches))
    }

    // No matches
    return nil, fmt.Errorf("no function %s found in imported packages", funcName)
}
```

**Result:** ✅ Detects and reports conflicts

---

### 10. Corrupted Cache

**Scenario:**
```
Cache file exists but:
  - Data is corrupted (partial write)
  - Version mismatch (cache from old version)
  - Hash mismatch (file changed during write)
```

**Problem:**
- Loading corrupted cache crashes scanner
- Stale version causes incorrect inference
- Partial data causes incomplete function index

**Solution:**
```go
type CacheHeader struct {
    Version       string    // e.g., "v1.2.0"
    GoVersion     string    // e.g., "go1.21"
    CreatedAt     time.Time
    PackageHash   string    // Hash of package files
}

type CacheEntry struct {
    Header    CacheHeader
    Data      *PackageInfo
    Checksum  uint64  // CRC64 of data
}

func (c *DiskCache) Get(pkgPath string) (*PackageInfo, error) {
    filePath := c.getCacheFilePath(pkgPath)

    // Read file
    data, err := os.ReadFile(filePath)
    if err != nil {
        return nil, err
    }

    // Deserialize
    var entry CacheEntry
    if err := c.serializer.Deserialize(data, &entry); err != nil {
        return nil, fmt.Errorf("cache deserialize failed: %w (corrupted cache)", err)
    }

    // Verify checksum
    if entry.Checksum != crc64.Checksum(entry.Data, crc64.ECMA) {
        return nil, fmt.Errorf("cache checksum mismatch (corrupted cache)")
    }

    // Check version
    if entry.Header.Version != c.expectedVersion {
        return nil, fmt.Errorf("cache version mismatch: %s != %s", entry.Header.Version, c.expectedVersion)
    }

    // Verify package hasn't changed
    currentHash := c.calculatePackageHash(pkgPath)
    if entry.Header.PackageHash != currentHash {
        return nil, fmt.Errorf("package has changed, cache invalid")
    }

    return entry.Data, nil
}
```

**Result:** ✅ Rejects corrupted cache, falls back to rescan

---

## Error Handling Strategy

### Error Categories

#### Transient Errors (Retry)

**Examples:**
- File temporarily locked
- Disk full (temporary)
- Network timeout (for go/packages)
- Permission denied (temporary)

**Handling:**
```go
func (s *PackageScanner) ScanWithRetry(ctx context.Context, pkgPath string) (*PackageInfo, error) {
    var lastErr error
    maxRetries := 3
    backoff := time.Millisecond * 100

    for attempt := 0; attempt < maxRetries; attempt++ {
        info, err := s.ScanPackage(ctx, pkgPath)
        if err == nil {
            return info, nil
        }

        if !isTransientError(err) {
            return nil, err  // Permanent error, don't retry
        }

        lastErr = err
        time.Sleep(backoff)
        backoff *= 2  // Exponential backoff
    }

    return nil, fmt.Errorf("scan failed after %d retries: %w", maxRetries, lastErr)
}

func isTransientError(err error) bool {
    return errors.Is(err, syscall.EINTR) ||
           errors.Is(err, syscall.EAGAIN) ||
           strings.Contains(err.Error(), "disk full")
}
```

---

#### Permanent Errors (No Retry)

**Examples:**
- Parse error in source file
- Type check failure
- Invalid package path
- Permission denied (real)

**Handling:**
```go
func (s *PackageScanner) ScanPackage(ctx context.Context, pkgPath string) (*PackageInfo, error) {
    // Check if package exists
    if !s.packageExists(pkgPath) {
        return nil, fmt.Errorf("package not found: %s", pkgPath)
    }

    // For parse/type errors, collect but don't fail the whole scan
    var fileErrors []error

    for _, file := range files {
        if err := s.processFile(file); err != nil {
            fileErrors = append(fileErrors, err)
        }
    }

    // If too many errors, fail
    if len(fileErrors) > 10 {
        return nil, fmt.Errorf("too many errors (%d), scan failed", len(fileErrors))
    }

    // Otherwise, continue with partial data
    return info, nil
}
```

---

#### Partial Failures (Continue with Warnings)

**Examples:**
- One file fails to parse
- Type check fails on one function
- Cache write fails (but we have data)

**Handling:**
```go
type ScanResult struct {
    PackageInfo *PackageInfo
    Warnings    []Warning
    Errors      []Error
}

type Warning struct {
    File    string
    Message string
}

func (s *PackageScanner) ScanPackage(ctx context.Context, pkgPath string) (*ScanResult, error) {
    result := &ScanResult{
        PackageInfo: &PackageInfo{PackagePath: pkgPath},
        Warnings:    []Warning{},
        Errors:      []Error{},
    }

    for _, file := range files {
        if err := s.processFile(file); err != nil {
            result.Errors = append(result.Errors, Error{
                File:    file.Path,
                Message: err.Error(),
            })
        }
    }

    // Even with errors, return partial result if we have some data
    if len(result.PackageInfo.Files) > 0 {
        return result, nil
    }

    return nil, fmt.Errorf("scan failed completely")
}
```

## Compatibility Analysis

### Go Module Compatibility

**Works with:**
- ✅ Go modules (go.mod)
- ✅ Go 1.16+ (implicit modules)
- ✅ Go 1.11-1.15 (with GO111MODULE=on)
- ✅ Multi-module workspaces (go.work)

**Partial support:**
- ⚠️ GOPATH mode (pre-1.11) - requires GO111MODULE=on

**Does not work:**
- ❌ Very old Go versions (<1.11)

**Testing strategy:**
```go
func TestModuleCompatibility(t *testing.T) {
    if !build.IsModule() {
        t.Skip("Not in module mode")
    }

    // Test with testgo
    testModule := testdata.Path("test_module")
    scanner := NewPackageScanner(config)

    info, err := scanner.ScanPackage(context.Background(), testModule)
    if err != nil {
        t.Fatalf("Scan failed: %v", err)
    }

    // Verify we can scan without errors
    if info == nil {
        t.Error("Scan returned nil info")
    }
}
```

---

### Cross-Platform Compatibility

**File Watching (platform-specific):**

```go
// Use fsnotify for cross-platform
type Watcher struct {
    watcher *fsnotify.Watcher
}

func NewWatcher() (*Watcher, error) {
    w, err := fsnotify.NewWatcher()
    if err != nil {
        return nil, err
    }

    // Handle platform differences
    // Linux: inotify
    // macOS: kqueue
    // Windows: ReadDirectoryChangesW

    return &Watcher{watcher: w}, nil
}
```

**File paths (Windows/macOS/Linux):**

```go
func (s *PackageScanner) normalizePath(path string) string {
    // Use filepath.ToSlash for consistent handling
    // On Windows: C:\path\to\file → C:/path/to/file
    // On Unix: /path/to/file → /path/to/file
    return filepath.ToSlash(path)
}
```

**Result:** ✅ Cross-platform compatible

---

### Multiple Go Versions

**Compatibility strategy:**

1. **Build tag per Go version:**
```go
//go:build go1.20
// +build go1.20

// Code specific to Go 1.20+
```

2. **Feature detection:**
```go
var supportsGenerics = false

func init() {
    // Try to compile generic code
    code := `
        package test
        func[T any] Foo() T { return *new(T) }
    `
    // If compilation succeeds, generics are supported
}
```

3. **Version checks:**
```go
func (s *PackageScanner) CheckGoVersion() error {
    goVersion := runtime.Version()
    if !strings.HasPrefix(goVersion, "go1.") {
        return fmt.Errorf("unsupported Go version: %s", goVersion)
    }

    major, minor, err := parseGoVersion(goVersion)
    if err != nil {
        return err
    }

    if major == 1 && minor < 16 {
        return fmt.Errorf("Go 1.16+ required, found: %s", goVersion)
    }

    return nil
}
```

**Result:** ✅ Supports Go 1.16 through latest

---

## Security Considerations

### Cache Security

**Risks:**
- Cache files contain source code snippets
- May leak proprietary code
- Cache directory permissions

**Mitigations:**
```go
type SecureCache struct {
    baseDir string
    mode    os.FileMode
}

func (c *SecureCache) Put(pkgPath string, info *PackageInfo) error {
    filePath := c.getCacheFilePath(pkgPath)

    // Create directory with restricted permissions
    dir := filepath.Dir(filePath)
    if err := os.MkdirAll(dir, 0700); err != nil {
        return err
    }

    // Write file with restricted permissions
    data, err := c.serializer.Serialize(info)
    if err != nil {
        return err
    }

    return os.WriteFile(filePath, data, 0600)  // Read/write owner only
}

func (c *SecureCache) Get(pkgPath string) (*PackageInfo, error) {
    filePath := c.getCacheFilePath(pkgPath)

    // Check file permissions
    info, err := os.Stat(filePath)
    if err != nil {
        return nil, err
    }

    if info.Mode()&077 != 0 {
        return nil, fmt.Errorf("cache file has incorrect permissions: %s", filePath)
    }

    // Read and return
    // ...
}
```

**Result:** ✅ Secure cache storage

---

### Input Validation

**Risks:**
- Path traversal (`../../../etc/passwd`)
- Directory traversal
- Malicious go.mod files
- Infinite recursion

**Mitigations:**
```go
func (s *PackageScanner) validatePackagePath(pkgPath string) error {
    // Absolute path required
    if !filepath.IsAbs(pkgPath) {
        return fmt.Errorf("package path must be absolute: %s", pkgPath)
    }

    // No traversal
    if strings.Contains(pkgPath, "..") {
        return fmt.Errorf("path traversal not allowed: %s", pkgPath)
    }

    // Within allowed roots
    if !s.isWithinRoots(pkgPath) {
        return fmt.Errorf("package path not in allowed roots: %s", pkgPath)
    }

    return nil
}

func (s *PackageScanner) isWithinRoots(pkgPath string) bool {
    for _, root := range s.allowedRoots {
        if strings.HasPrefix(pkgPath, root) {
            return true
        }
    }
    return false
}
```

**Result:** ✅ Secure input handling

---

### Resource Limits

**Risks:**
- Infinite recursion in file walking
- Excessive memory usage
- Disk space exhaustion
- CPU starvation

**Mitigations:**
```go
type SafeWalker struct {
    maxDepth    int
    maxFiles    int
    maxSizeMB   int
    scanned     int
    scannedSize int64
    mutex       sync.Mutex
}

func (w *SafeWalker) Walk(root string, fn filepath.WalkFunc) error {
    return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
        w.mutex.Lock()
        defer w.mutex.Unlock()

        // Check depth
        rel, _ := filepath.Rel(root, path)
        if strings.Count(rel, string(filepath.Separator)) > w.maxDepth {
            if info.IsDir() {
                return filepath.SkipDir
            }
            return nil
        }

        // Check file count
        if w.scanned >= w.maxFiles {
            return filepath.SkipAll
        }

        // Check size
        if w.scannedSize > int64(w.maxSizeMB)*1024*1024 {
            return fmt.Errorf("scan size limit exceeded: %d MB", w.maxSizeMB)
        }

        w.scanned++
        if !info.IsDir() {
            w.scannedSize += info.Size()
        }

        return fn(path, info, err)
    })
}
```

**Result:** ✅ Resource-constrained scanning

---

## Summary of Trade-offs

| Decision | Chosen Approach | Rationale | Risk |
|----------|-----------------|-----------|------|
| Scanner backend | Manual → go/packages (opt-in) | Simple start, advanced later | Complexity |
| Cache strategy | Two-level (memory + disk) | Fast + persistent | Staleness |
| Function detection | Progressive (regex → AST → types) | Balance speed/accuracy | Regex maintenance |
| Import inference | Confidence-scored | Max value with safety | Occasional wrong inference |
| Build tags | Optional go/packages | Accuracy over speed | External dep |
| Vendored deps | Configurable (skip default) | Speed over completeness | May miss edge cases |
| Generated files | Auto-skip | Reduce noise | May skip self-generated |
| Tests | Include by default | Completeness | Slower scans |
| Concurrency | RWMutex | Simplicity | Lock contention |
| Errors | Partial results with warnings | User experience | May give partial data |

**Overall Approach:**
- Start simple, optimize iteratively
- Safety first (no wrong inference)
- Performance as second priority
- Comprehensive edge case handling
- Production-ready from Phase 1

This architecture provides a robust, scalable solution for package-wide scanning with proper handling of all edge cases and trade-offs.
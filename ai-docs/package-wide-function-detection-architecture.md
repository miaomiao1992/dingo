# Package-Wide Function Detection Architecture for Dingo Import Inference

## Problem Statement
The current import tracking system operates on a single-file basis, which causes false positives during unqualified import inference. When transforming `ReadFile(path)` to `os.ReadFile(path)`, the system doesn't verify whether `ReadFile` is a user-defined function in the same package, potentially transforming local function calls incorrectly.

## Current State Analysis

### ImportTracker Limitations
- **Scope**: Single-file processing only
- **Detection**: Regex scanning for unqualified calls like `ReadFile(`
- **Logic Gap**: No verification of user-defined functions in package
- **False Positive Risk**: Transforms local function calls into stdlib calls

### Existing Code Structure
```go
type ImportTracker struct {
    needed map[string]bool  // tracks imports like "os": true
}

func (it *ImportTracker) TrackFunctionCall(funcName string) {
    // Tracks qualified calls like "os.ReadFile"
}

func qualifyUnqualifiedImports(source []byte) ([]byte, []Mapping, []string) {
    // Scans for unqualified calls, transforms them, but doesn't check local functions
    re := regexp.MustCompile(`\b([A-Z][a-zA-Z0-9_]*)\s*\(`)
    // ... transforms ALL unqualified calls to stdlib if in knownImports
}
```

## Proposed Architecture

### Core Components

#### 1. **PackageFunctionScanner**
Central component that scans entire packages for user-defined functions:

```go
type PackageFunctionScanner struct {
    packagePath   string                    // absolute path to package directory
    exclusions    map[string]FunctionInfo   // function name → metadata
    lastScanTime  time.Time
    mutex         sync.RWMutex
}

type FunctionInfo struct {
    Name       string    // function name
    Receiver   string    // struct receiver if method (e.g., "(t *Type)")
    File       string    // file where defined
    Line       int       // definition line number
    Exported   bool      // public vs private
}
```

#### 2. **Exclusion Cache System**
Persistent caching to avoid rescanning unchanged packages:

```go
type FunctionExclusionCache struct {
    cache map[string]*PackageFunctionScanner  // package path → scanner
    fileMods map[string]time.Time           // file path → last mod time
    mutex    sync.RWMutex
}
```

#### 3. **Integration with ImportTracker**
Extended ImportTracker with package-wide awareness:

```go
type ImportTracker struct {
    needed         map[string]bool
    exclusions     map[string]FunctionInfo  // package functions to exclude
    scanner        *PackageFunctionScanner
}
```

### Data Flow

```
[Package Directory: /path/to/pkg]
    ↓
PackageFunctionScanner.ScanPackage()
    ↓ Parses all .dingo and .go files
    ↓ Extracts function definitions →
Function Exclusion List: ["ReadFile", "ProcessData", "validateInput"]
    ↓
Enhanced ImportTracker with exclusions
    ↓
qualifyUnqualifiedImports() SKIPS excluded functions
    ↓
Imports qualified safely: user.ReadFile() stays user.ReadFile()
    → os.ReadFile() gets os. prefix + import
```

### Key Integration Points

#### 1. **Preprocessor Pipeline Integration**
Modify `preprocessor.go` to inject package scanning before individual file processing:

```go
func (p *Preprocessor) Process() (string, *SourceMap, error) {
    // NEW: Scan package for user-defined functions first
    exclusionCache := getOrCreateExclusionCache()
    exclusions := exclusionCache.GetExclusionsForPackage(p.packagePath)

    // Pass exclusions to error propagation processor
    errorProc := NewErrorPropProcessorWithExclusions(exclusions)

    // ... rest of processing with exclusion-aware tracker
}
```

#### 2. **ErrorProp Processor Extension**
Update `ErrorPropProcessor` to accept function exclusions:

```go
func NewErrorPropProcessorWithExclusions(exclusions map[string]FunctionInfo) *ErrorPropProcessor {
    return &ErrorPropProcessor{
        importTracker: NewImportTrackerWithExclusions(exclusions),
    }
}
```

#### 3. **ImportTracker Enhancement**
Modify `qualifyUnqualifiedImports()` to skip excluded functions:

```go
func (it *ImportTracker) qualifyUnqualifiedImports(source []byte) ([]byte, []Mapping, []string) {
    re := regexp.MustCompile(`\b([A-Z][a-zA-Z0-9_]*)\s*\(`)

    return re.ReplaceAllStringFunc(string(source), func(match string) string {
        funcName := strings.TrimSuffix(match, "(")

        // NEW: Skip if user-defined function in package
        if _, isUserFunction := it.exclusions[funcName]; isUserFunction {
            return match  // Don't transform
        }

        // Existing logic: check stdlib functions
        if pkg, exists := it.stdlibImports[funcName]; exists {
            it.needed[pkg] = true
            return fmt.Sprintf("%s.%s(", pkg, funcName)
        }

        return match  // Not a stdlib function
    })
}
```

## Caching Strategy

### Cache Invalidation Triggers
1. **File Modification**: Any `.go` or `.dingo` file timestamp changes
2. **New File Added**: New function definition added to package
3. **File Deletion**: Function definition removed

### Cache Implementation
```go
func (cache *FunctionExclusionCache) GetExclusionsForPackage(pkgPath string) map[string]FunctionInfo {
    cache.mutex.RLock()
    scanner, exists := cache.cache[pkgPath]
    cache.mutex.RUnlock()

    if !exists {
        scanner = NewPackageFunctionScanner(pkgPath)
        cache.mutex.Lock()
        cache.cache[pkgPath] = scanner
        cache.mutex.Unlock()
    }

    // Check if cache is stale
    if scanner.IsStale() {
        scanner.RescanPackage()
    }

    return scanner.GetExclusions()
}
```

### Incremental Build Efficiency
- **Minimal Rescan**: Only rescans files with newer timestamps
- **Function Index**: Maintains reverse index: function → defining file
- **Delta Updates**: Update only changed files, merge with cached results

## Performance Analysis

### Scan Times (Estimated)
- **Small Package (3-5 files)**: <50ms
- **Medium Package (10-20 files)**: <200ms
- **Large Package (50+ files)**: <500ms (target ceiling)

### Bottlenecks & Optimizations
1. **Parsing**: Use `go/parser` only as fallback; prefer regex for function signature extraction
2. **Memory**: Store only function names/metadata, not full ASTs
3. **Concurrency**: Parallel scanning of files within package
4. **Caching**: 99% cache hit rate expected in watch mode

### Watch Mode Considerations
- **File Change Detection**: Hook into existing file watcher
- **Background Rescan**: Async rescanning to avoid blocking main thread
- **Cache Prefetch**: Pre-scan likely-to-change packages

## Incremental Build Handling

### Change Types & Responses
1. **File Modified**: Rescan single file, update exclusions
2. **File Added**: Scan new file, merge exclusions
3. **File Deleted**: Remove functions from that file from exclusions
4. **Package Directory Change**: Full rescan

### State Management
```go
type PackageScanState struct {
    lastFullScan  time.Time
    fileHashes    map[string]string  // file → content hash
    exclusions    map[string]FunctionInfo
}
```

### Transactional Updates
Ensure thread-safe updates during concurrent builds:
```go
func (scanner *PackageFunctionScanner) updateExclusions(newExclusions map[string]FunctionInfo) {
    scanner.mutex.Lock()
    scanner.exclusions = newExclusions
    scanner.lastScanTime = time.Now()
    scanner.mutex.Unlock()
}
```

## Implementation Plan

### Phase 1: Core Infrastructure (Week 1-2)
1. **Create PackageFunctionScanner** (`pkg/preprocessor/package_scanner.go`)
   - Function extraction from .go and .dingo files
   - Exclusion map generation
   - Basic caching implementation

2. **Extend ImportTracker** (`pkg/preprocessor/error_prop.go`)
   - Add exclusion map field
   - Modify `qualifyUnqualifiedImports()` to use exclusions
   - Maintain backward compatibility

3. **Integration Points**
   - Update preprocessor pipeline to inject scanner
   - Add package path tracking to Preprocessor struct

### Phase 2: Caching & Performance (Week 3-4)
1. **FunctionExclusionCache** (`pkg/preprocessor/exclusion_cache.go`)
   - LRU-style cache with configurable size
   - File modification detection
   - Cache invalidation logic

2. **Incremental Scanning**
   - File-level change detection
   - Delta exclusion updates
   - Concurrent file scanning

3. **Performance Benchmarks**
   - Create benchmark tests for scan times
   - Measure cache hit rates
   - Profile memory usage

### Phase 3: Testing & Edge Cases (Week 5-6)
1. **Comprehensive Test Suite**
   - Package-level test files with various scenarios
   - Cross-file function dependencies
   - Incremental build validation

2. **Edge Case Handling**
   - Methods with receivers
   - Generic functions
   - Function overloading (if added later)
   - Private vs public functions

3. **Integration Testing**
   - Full preprocessor pipeline tests
   - Watch mode simulation
   - Error propagation with exclusions

### Phase 4: Production Integration (Week 7-8)
1. **CLI Integration**
   - Add package-wide scanning to `dingo` command
   - Command-line cache management options

2. **Error Handling & Logging**
   - Graceful degradation on scan failures
   - Detailed logging for debugging
   - Cache corruption recovery

3. **Documentation**
   - Update preprocessor README with new architecture
   - Add performance guidelines
   - Developer documentation for new components

## Trade-offs & Edge Cases

### Trade-offs Made

#### Performance vs Accuracy
- **Trade-off**: Regex-based function extraction (fast) vs full AST parsing (accurate)
- **Rationale**: 90%+ accuracy with regex sufficient; full parsing only when needed
- **Mitigation**: Fallback AST parsing for ambiguous cases

#### Memory vs Speed
- **Trade-off**: Store full exclusion maps (memory) vs re-scan frequently (speed)
- **Rationale**: Memory usage minimal (~KB per package), scanning is expensive
- **Mitigation**: Configurable cache size limits

#### Simplicity vs Features
- **Trade-off**: Basic function name matching vs full type-aware exclusion
- **Rationale**: Simpler implementation sufficient for current needs
- **Mitigation**: Architecture extensible for future enhancements

### Edge Cases Handled

#### 1. Method Receivers
```go
// Function: ProcessData (method on DataProcessor)
func (dp *DataProcessor) ProcessData() { ... }

// Call: dp.ProcessData() - should NOT qualify to os.ProcessData()
```
**Solution**: Track receiver information; exclude qualified method calls

#### 2. Generic Functions
```go
// Function: Process[T any]
func Process[T any](data T) T { ... }

// Call: Process(data) - should NOT qualify to os.Process()
```
**Solution**: Include generic syntax in function signature extraction

#### 3. Cross-Package Functions
```go
// Other package: mylib.ReadFile
import "mylib"

// Call: mylib.ReadFile() - already qualified, no issue
// Call: ReadFile() - check exclusion, may qualify incorrectly
```
**Solution**: Exclusions are package-scoped; different packages have separate exclusion lists

#### 4. Function Shadowing
```go
import "os"

// Local function shadows stdlib
func ReadFile(path string) ([]byte, error) { ... }

// Call: ReadFile(path) - should use local, not qualify to os.ReadFile
```
**Solution**: User-defined functions always take precedence in same package

#### 5. Package Aliases
```go
import fileops "os"  // aliased import

// Call: fileops.ReadFile() - already qualified via alias
// Call: ReadFile() - may incorrectly qualify to os.ReadFile
```
**Solution**: Current design handles this correctly via existing qualified call detection

#### 6. Test Files
- Functions defined in `_test.go` files
- Test helper functions
- Benchmark functions
- Should they be excluded from qualification?

**Design Decision**: Yes, include test functions in exclusions to prevent false positives in test packages

#### 7. Generated Code
- Functions in generated `.go` files (e.g., `*_generated.go`)
- Should be treated same as hand-written code

**Design Decision**: Include generated files in scanning to maintain consistency

### Known Limitations

#### 1. Dynamic Function Resolution
Cannot detect functions called via reflection or function variables:
```go
funcVar := ReadFile  // cannot detect this usage
```

#### 2. Interface Satisfaction
Cannot prevent qualification that breaks interface implementations:
```go
type Reader interface { Read([]byte) (int, error) }

// This would be qualified to os.Read, breaking interface
func (r *MyReader) Read(buf []byte) (int, error) { ... }
```

#### 3. Macro-like Functions
Functions that should be imported but are treated as keywords:
```go
func panic(v interface{})  // builtin, shouldn't be qualified
```

**Mitigation**: Maintain builtin function exclusion list separate from package scanning

## Testing Strategy

### Unit Tests
- **PackageFunctionScanner**: Function extraction accuracy tests
- **FunctionExclusionCache**: Cache hit/miss behavior tests
- **ImportTracker**: Exclusion integration tests
- **Preprocessor**: Full pipeline with exclusions

### Integration Tests
- **Golden Tests**: Add package-level test files with cross-file dependencies
- **Edge Cases**: Method receivers, generics, shadowing scenarios
- **Performance**: Benchmark tests for scan times and cache efficiency

### Real-World Validation
- **CLI Testing**: Test with actual Dingo packages
- **Watch Mode**: Simulate file changes and verify incremental updates
- **Large Package**: Test with real-world sized packages (100+ files)

This architecture provides a robust, performant solution for package-wide function detection while maintaining the existing import tracking system's compatibility and performance characteristics.
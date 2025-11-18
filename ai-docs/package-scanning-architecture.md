# Package-Wide Scanning Architecture for Unqualified Import Inference

## Problem Statement
The Dingo transpiler currently performs single-file scanning to detect unqualified function calls like `ReadFile(path)` and transform them to `os.ReadFile(path)` with appropriate imports. However, this approach creates false positives when functions are defined elsewhere in the same package, preventing correct inference of external package imports.

## 1. Proposed Architecture

### Core Components

**PackageScanner** (`pkg/scanner/package.go`):
- Loads and parses all source files in a package simultaneously
- Maintains package-wide symbol registry
- Provides context-aware import inference
- Integrates with existing go/packages ecosystem

**SymbolRegistry** (`pkg/scanner/registry.go`):
- Thread-safe map of package symbols (functions, types, variables)
- Tracks symbol-to-file mappings for dependency analysis
- Supports incremental updates and cache invalidation

**ImportInferenceEngine** (`pkg/scanner/inference.go`):
- Analyzes unqualified identifiers against package-wide context
- Determines required external imports from `GOPATH`/`GOMOD`
- Generates import statements and qualified calls

**Data Flow**

```
.dingo files → go/packages.Load() → AST parsing
                                         ↓
Symbol Registry ←── Package-wide analysis
                                         ↓
Unqualified calls → Import Inference Engine → Qualified calls + imports
                                         ↓
go/ast transformation → Final .go output
```

### Integration Points

**Transpilation Pipeline Integration**:
```go
// pkg/transpiler/transpiler.go
type Transpiler struct {
    scanner *scanner.PackageScanner  // NEW
    preprocessor *preprocessor.Processor
    parser *parser.ASTParser
    plugins []plugin.Plugin
}
```

**ImportTracker Extension**:
```go
// Extension to existing ImportTracker
type ImportTracker struct {
    explicit map[string]string      // Current
    inferred map[string]string      // NEW: package-wide inferred
    required map[string]string      // NEW: external package needs
}
```

## 2. Caching Strategy

### What to Cache
- **Symbol Registry**: Complete map of all package-defined symbols
- **Package Metadata**: go.mod dependencies, GOPATH resolution
- **AST Nodes**: Parse trees for symbol extraction (avoid re-parsing)
- **Import Candidates**: Pre-computed external package mappings

### Cache Location & Persistence
- **In-Memory Cache**: Per compilation session
- **File-Based Cache**: JSON cache in `.dingo/cache/` directory
- **Cache Metadata**: Includes modification timestamps and package hashes

### Invalidation Rules
- **File Change**: Invalidate symbols from modified file only
- **New File Added**: Invalidate entire package cache
- **Dependency Change**: Invalidate related packages in module graph
- **Cache Size Limit**: LRU eviction when exceeding 10MB per session

## 3. Incremental Build Handling

### File Change Detection
```go
// pkg/scanner/incremental.go
type ChangeDetector struct {
    fileTimes map[string]time.Time
    packageHash string
}

func (d *ChangeDetector) HasPackageChanged(files []string) bool {
    // Track file modification times
    // Return true if any file newer than cache
}
```

### Rescan Scope Determination
- **Single File Change**: Rescan only changed file + dependent files
- **Package Structure Change**: Full package rescan
- **New Import Added**: Incremental symbol collection
- **Target Rescan Time**: <100ms for single file modifications

### Dependency Tracking
- **Reverse Dependencies**: Track which files import from changed files
- **Lazy Loading**: Load dependent packages only when needed
- **Parallel Scanning**: Scan independent sub-packages concurrently

## 4. Performance Analysis

### Time Complexity Estimates
- **Full Package Scan**: O(n) where n = total lines across package
  - Small package (3 files, 500 lines): ~50ms
  - Medium package (10 files, 2000 lines): ~150ms
  - Large package (50 files, 10000 lines): ~400ms
- **Incremental Scan**: O(m) where m = changed lines
  - Single file change: ~10-20ms overhead
  - Multi-file change: ~50-100ms depending on scope

### Memory Usage
- **Package Symbols**: ~2-5KB per 1000 lines of code
- **AST Storage**: ~10-15KB per source file in memory
- **Cache Overhead**: ~50KB initial + 10KB per cached package
- **Peak Memory**: <20MB for typical Go packages

### Optimization Strategies
```go
// Parallel processing for multi-core systems
workers := runtime.NumCPU()
pool := worker.NewPool(workers, func(file string) {
    symbols := scanner.ExtractSymbols(file)
    registry.Merge(symbols)
})
```

**Rationale**:
- Target **<500ms per package** achieved through:
  - Shared go/packages AST (no re-parsing)
  - Parallel symbol extraction across files
  - Incremental scanning for most builds (90%+ of cases)
  - Cache hit rate >95% for unchanged packages

## 5. Implementation Plan

### Code Structure
```
pkg/scanner/
├── package.go         # Main PackageScanner type
├── registry.go        # SymbolRegistry implementation
├── inference.go       # Import inference logic
├── incremental.go     # Change detection
├── cache.go          # Caching layer
└── scanner_test.go   # Comprehensive tests

pkg/transpiler/
└── transpiler.go      # Integrate PackageScanner
```

### Phase Implementation
1. **Phase 1: Core Scanner** (2 weeks)
   - PackageScanner + SymbolRegistry
   - Package-wide symbol collection
   - Basic import inference logic

2. **Phase 2: Caching & Incremental** (1 week)
   - File-based caching
   - Change detection
   - Incremental scanning

3. **Phase 3: Transpiler Integration** (1 week)
   - ImportTracker extension
   - Pipeline integration
   - Performance optimization

4. **Phase 4: Testing & Polish** (1 week)
   - Comprehensive test suite
   - Edge case handling
   - Performance benchmarks

### Testing Strategy
- **Integration Tests**: End-to-end transpilation with package-wide imports
- **Unit Tests**: Each component isolated with mock packages
- **Performance Tests**: Benchmark scanning time vs. package size
- **Regression Tests**: Ensure existing single-file behavior unaffected

## 6. Trade-offs & Edge Cases

### Architecture Trade-offs

| Approach | Pros | Cons |
|----------|------|------|
| **Full Package Always** | Accurate symbol detection<br>Simple implementation | Slower on large packages<br>Higher memory usage |
| **Incremental Only** | Fast for small changes<br>Memory efficient | Misses cross-file dependencies<br>Complex dependency tracking |
| **Hybrid (Current Choice)** | Good accuracy + performance<br>Best of both worlds | More complex implementation<br>Cache management overhead |

### Performance Trade-offs
- **Memory vs. Speed**: Caching symbols increases memory (2-5x) for 10-20x speed improvement
- **Accuracy vs. Simplicity**: Package-wide scanning eliminates ~80% false positives at cost of 2-3x scan time
- **Parallelization**: Threading improves performance on large packages but adds complexity

### Edge Cases & Solutions

**Package with Cycles**:
```go
// file_a.go: defines FuncA
// file_b.go: calls FuncA + defines FuncB
// file_c.go: calls FuncB
```
**Solution**: Topological sort dependencies, scan in dependency order

**External Package Conflicts**:
```go
// Project has: func ReadFile()
// Also uses: os.ReadFile(), bufio.ReadFile()
```
**Solution**: Prefer local symbols unless explicitly imported, external only for unqualified calls

**Large Package Performance**:
**Benchmark**: Kubernetes API package (200+ files, 500k lines)
**Expected**: ~1200ms full scan → ~200ms incremental → acceptable with caching

**Generated Files**:
**Problem**: Files like `zz_generated.deepcopy.go` contain conflicting symbols
**Solution**: Exclude files matching `*generated*` patterns from symbol registry

**Import Resolution**:
**Problem**: Multiple packages provide same function name (e.g., `ReadAll` in `io` + `ioutil`)
**Solution**: Prefer standard library, then most recently imported packages

### Error Handling
- **Missing Symbol Discovery**: Graceful fallback to external import inference
- **Package Load Failure**: Clear error messages with specific file/package causing issues
- **Cache Corruption**: Automatic cache invalidation and rebuild on detect

### Alternative Approaches Considered

**1. IDE-Style Language Server**
   - Always keep full symbol table in memory
   - +: Fast queries, real-time updates
   - -: High memory usage (GBs for large codebases), LSP complexity overhead

**2. Build-System Integration**
   - Use go/build or gomodulecache
   - +: Leverage existing Go tooling
   - -: Limited to current Go tool ecosystem, slower cache invalidation

**3. Single-File with Context Hints**
   - Pass "known local symbols" as preprocessor context
   - +: Fastest scanning
   - -: Still misses some cross-file references, complex state management

**Chosen Approach: Hybrid Package Scanner** provides best balance of correctness, performance, and implementation simplicity while directly addressing the false positive problem through comprehensive package-wide symbol tracking.
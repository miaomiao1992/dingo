
[claudish] Model: google/gemini-2.5-flash

## Proposed Architecture

To efficiently scan all files in a package during transpilation and build a shared exclusion list of local functions, I propose the following architecture:

### 1. Package Scanner Module (`pkg/scanner`)

This new module will be responsible for recursively discovering all `.dingo` files within a given package directory.

-   **`pkg/scanner/scanner.go`**: Contains functions to traverse directories, identify `.dingo` files, and group them by package.
-   **`pkg/scanner/analyzer.go`**: Contains logic to parse `.dingo` files (using a simplified Dingo parser or regex for initial pass) and extract function declarations, variable declarations, and other symbols that should be excluded from unqualified import inference. This will generate a `SymbolTable` for each file.

### 2. Global Symbol Table (`pkg/transpiler/symbols`)

This module will aggregate symbol information from all files in a package.

-   **`pkg/transpiler/symbols/table.go`**: Defines a `SymbolTable` struct to store function names, variable names, and their associated file paths. This table will be package-scoped.

### 3. Import Tracker Enhancement (`pkg/preprocessor/importtracker`)

The existing `ImportTracker` in the preprocessor will be enhanced to consult the `GlobalSymbolTable` before performing unqualified import inference.

-   **`pkg/preprocessor/importtracker/tracker.go`**: Before transforming `ReadFile(path)` to `os.ReadFile(path)`, it will check if `ReadFile` is present in the `GlobalSymbolTable` for the current package. If it is, the transformation will be skipped.

### Data Flow

1.  **Initialization**: When transpilation for a package starts (e.g., `dingo build ./myproject/mypackage`), the `Package Scanner` will be invoked with the package root.
2.  **File Discovery**: The scanner recursively finds all `.dingo` files within the package.
3.  **Symbol Extraction**: For each `.dingo` file, the `analyzer` within the `Package Scanner` will extract local symbols (functions, variables).
4.  **Global Table Population**: The extracted symbols will be added to the `GlobalSymbolTable` for the package.
5.  **Preprocessor Integration**: During the preprocessor stage, the `ErrorPropProcessor` and any other relevant processors will query the `GlobalSymbolTable` before performing unqualified transformations.

## Caching Strategy

To ensure performance for package-wide scanning, especially in watch mode and for incremental builds, a robust caching strategy is crucial.

### 1. Symbol Table Cache

-   **Goal**: Avoid re-scanning all `.dingo` files in a package if they haven't changed.
-   **Implementation**: Store the `GlobalSymbolTable` for a package in a temporary file (e.g., `pkgname/.dingo_cache/symbols.json`). Generate a hash for each `.dingo` file and store alongside the `GlobalSymbolTable`.
-   **Cache Invalidation**: If any file's hash has changed, or if new `.dingo` files are added/deleted, invalidate the entire cache for that package and re-scan all files.

### 2. Import Decision Cache (within Preprocessor)

-   **Goal**: Speed up the decision of whether to qualify an import.
-   **Implementation**: The `ImportTracker` can maintain a small in-memory cache of previously resolved unqualified names.
-   **Cache Invalidation**: This cache should be cleared if the `GlobalSymbolTable` is invalidated and rebuilt.

### 3. File Metadata Cache

-   **Goal**: Avoid repeatedly calling `os.Stat` for file discovery.
-   **Implementation**: Store file modification times (`ModTime`) and file sizes for `.dingo` files in a cache.
-   **Cache Invalidation**: Compare current `ModTime` and size with cached values. If different, re-read the file.

## Incremental Build Handling

Efficient incremental build handling is vital for a responsive watch mode experience.

### 1. File Watcher Integration

-   **Mechanism**: Utilize a file watcher (e.g., `fsnotify` in Go) to monitor changes in `.dingo` files.
-   **Trigger**: When a change is detected, invalidate the cache for that specific file and update the `File Metadata Cache`.

### 2. Smart Cache Invalidation and Reconstruction

-   **Partial Invalidation**: Instead of always invalidating the entire `GlobalSymbolTable`, re-parse *only* the modified file to update its symbols. If a file is added/deleted, update the `GlobalSymbolTable` accordingly.
-   **Rebuild Checksum**: After any incremental update to the `GlobalSymbolTable`, recalculate its internal checksum to indicate changes.

### 3. On-Demand Symbol Table Loading

-   **Strategy**: When the transpiler starts for a package, check if a valid `GlobalSymbolTable` exists in the cache. If valid, load it; otherwise, trigger a full re-scan.

### 4. Integration with Transpiler Daemon (Watch Mode)

-   **Persistent Process**: In watch mode, the Dingo transpiler should run as a persistent daemon.
-   **Event-Driven Re-transpilation**: File watcher events trigger incremental updates to the `GlobalSymbolTable` and re-transpilation of only the changed `.dingo` file.

## Performance Analysis

### Cold Start (No Cache/Invalid Cache)

-   **Cost**: Dominated by file discovery and symbol extraction for all `.dingo` files in the package.
-   **Estimation**: For a package with 50 `.dingo` files, each 100 lines, a cold start could be around ~520ms. This emphasizes the need for a very lightweight, regex-based "simplified Dingo parser" in `pkg/scanner/analyzer.go`.

### Warm Start (Valid Cache)

-   **Cost**: Loading cached metadata and symbol table.
-   **Estimation**: Very fast, around ~20-30ms.

### Incremental Build (Single File Change)

-   **Cost**: File watcher notification, partial cache update, and re-transpilation of a single file.
-   **Estimation**: Highly efficient, around ~30-40ms.

## Implementation Plan

### Phase 1: Core Scanner and Symbol Table (2-3 days)

1.  **Define `SymbolTable` and Symbol Structures**: Create `pkg/transpiler/symbols/table.go` with `Symbol` struct (Name, Type, File, Package) and `SymbolType` enum. Implement add/remove/lookup methods and JSON serialization.
2.  **Basic Package Scanner**: Implement `FindDingoFiles` in `pkg/scanner/scanner.go` to find all `.dingo` files.
3.  **Lightweight Symbol Analyzer**: Create `pkg/scanner/analyzer.go` with `AnalyzeFile` to extract function/variable names using simple regexes.
4.  **Integrate Scanner with `dingo build`**: Orchestrate the scan to build a `GlobalSymbolTable` for the package.

### Phase 2: Caching System (2-3 days)

1.  **File Hash and Metadata Management**: Enhance `pkg/scanner/scanner.go` to calculate file hashes and store modification times.
2.  **`GlobalSymbolTable` Caching**: Implement logic to load/save `GlobalSymbolTable` from/to a `pkgname/.dingo_cache/symbols.json` file, including file hashes for invalidation.
3.  **Cache Invalidation Logic**: Implement logic to verify cached data against current file states and trigger full re-scan if needed.
4.  **Import Decision Cache**: Add an in-memory boolean map to `pkg/preprocessor/importtracker/tracker.go` for quick unqualified symbol lookups.

### Phase 3: Preprocessor Integration and Testing (1-2 days)

1.  **Integrate with Preprocessor**: Pass the `GlobalSymbolTable` to preprocessors (e.g., `ErrorPropProcessor`). Modify processors to consult the `GlobalSymbolTable` before transforming unqualified names.
2.  **Unit Tests**: Write tests for symbol extraction, cache hit/miss scenarios.
3.  **Integration Tests**: Add new `tests/golden` cases for unqualified import inference, including local functions conflicting with standard library functions and multi-file package scenarios.

### Phase 4: Incremental Build Support (Watch Mode) (2-3 days)

1.  **File Watcher (`fsnotify`)**: Implement a goroutine in `cmd/dingo-lsp/main.go` to watch `.dingo` files.
2.  **Event Handling**: On file changes, update `File Metadata Cache`, re-analyze the changed file, update the `GlobalSymbolTable`, clear `Import Decision Cache`, and trigger re-transpilation of affected files.

## Trade-offs & Edge Cases

### Trade-offs

-   **Complexity vs. Accuracy**: The lightweight, regex-based scanner prioritizes performance over absolute accuracy, relying on the Go compiler to catch any rare false transformations.
-   **Performance vs. Real-time Feedback**: Caching optimizes for warm and incremental builds, which occur more frequently than cold starts.
-   **Transpiler State vs. Statelessness**: Introducing a `GlobalSymbolTable` adds package-wide state, which is necessary for cross-file context.

### Edge Cases

-   **Macro-like Dingo Features**: Advanced metaprogramming features might require enhancing the regex-based scanner.
-   **Shadowing/Redeclaration**: The current flat `GlobalSymbolTable` might not distinguish between different scopes (e.g., function parameters vs. package-level variables). The Go compiler will eventually handle these.
-   **Invalid Dingo Files**: The `AnalyzeFile` function should gracefully handle syntactically incorrect files.
-   **Circular Dependencies**: Not an issue within a single package's symbol aggregation.
-   **External `.go` files**: The current scope is `.dingo` files only. Scanning `.go` files for local symbols would add complexity.
-   **Case Sensitivity**: Regexes and lookups must be case-sensitive.

This comprehensive plan addresses the functionality, performance, and robustness requirements for implementing package-wide local function detection.

[claudish] Shutting down proxy server...
[claudish] Done


# Package-Wide Scanning for Unqualified Import Inference: Architectural Consultation

## Problem Context

**Challenge**: Dingo transpiler needs to implement unqualified function call inference to transform calls like `ReadFile(path)` → `os.ReadFile(path)` + add import `"os"`. This requires detecting which identifiers are user-defined local functions (to skip inference) vs external package functions (to transform).

**Core Issue**: Current approach scans only single files (~10-50ms per file), but this risks false positives when functions are defined in OTHER files of the same package. We need package-wide scanning from the start.

## Technical Constraints

**Architecture**:
- Dingo transpiles file-by-file (currently stateless between files)
- Preprocessor stage: text transformations before go/parser
- Must integrate with existing ImportTracker
- Uses go/packages + go/parser + go/ast ecosystem

**Performance Requirements**:
- Package scan: <500ms total (acceptable threshold)
- Watch mode: Fast incremental builds when only 1 file changes
- Memory: Cache scan results efficiently across transpilations

**Key Questions**:
1. How to efficiently discover all .dingo files in a package?
2. Best caching strategy for scan results across file transpilations?
3. Incremental build handling (what to rescan when one file changes)?
4. go/packages vs manual file discovery - which performs better?
5. Integration points with preprocessor pipeline?

## Implementation Architecture Needed

**Required Components**:
- Package-level function indexer
- Caching layer for function declarations
- Incremental build tracker
- Integration with preprocessor + ImportTracker

**Data Flow Design**:
```
1. Transpiler starts → Discover package .dingo files
2. Build function index (global definitions)
3. For each file transpilation:
   a. Use function index to skip local function calls
   b. Transform unqualified external calls
   c. Cache per-file results
4. Incremental: Only rescan changed files + dependent files
```

## Your Consultation Request

Please provide detailed architectural analysis covering:

1. **Proposed Architecture** (specific components, data flow)
2. **Caching Strategy** (what to cache, where, when to invalidate)
3. **Incremental Build Handling** (what gets rescanned when file changes)
4. **Performance Analysis** (time/memory estimates with rationale)
5. **Implementation Plan** (code structure, which packages/files)
6. **Trade-offs & Edge Cases** (pros/cons of different approaches)

**Focus on**: Practical, production-ready design that integrates cleanly with existing codebase.

**Reference**: Current transpiler uses two-stage approach (preprocessor → go/parser → AST plugins). Package scanner must fit this pipeline.
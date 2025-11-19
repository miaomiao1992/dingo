# Final Implementation Plan: Unqualified Import Inference

## Executive Summary

Implement automatic import inference for unqualified standard library function calls in Dingo, transforming `ReadFile(path)` → `os.ReadFile(path)` with automatic `import "os"` injection.

**User-Approved Decisions:**
- ✅ Conservative ambiguity handling (compile errors)
- ✅ Comprehensive stdlib coverage
- ✅ AST pre-scan for local function detection
- ✅ Preprocessor stage implementation

## Architecture Overview

### Two-Tier Mapping System

```
Tier 1: Qualified Map (existing)
"os.ReadFile" → "os"
"strconv.Atoi" → "strconv"

Tier 2: Unqualified Map (NEW)
"ReadFile" → "os"
"Atoi" → "strconv"
"Printf" → "fmt"

Ambiguity Registry:
"Open" → ["os", "net"] (ERROR: must qualify)
```

### Processing Pipeline

```
.dingo file with unqualified calls
    ↓
┌─────────────────────────────────────────┐
│ Step 1: AST Pre-Scan (NEW)             │
├─────────────────────────────────────────┤
│ • Parse existing code with go/parser   │
│ • Extract local function definitions   │
│ • Build exclusion list                 │
│ • Pass exclusions to preprocessor      │
└─────────────────────────────────────────┘
    ↓
┌─────────────────────────────────────────┐
│ Step 2: Unqualified Import Processor    │
├─────────────────────────────────────────┤
│ • Scan for unqualified function calls  │
│ • Check local exclusions first         │
│ • Lookup in stdlib registry            │
│ • Detect ambiguous cases → ERROR       │
│ • Transform: ReadFile → os.ReadFile    │
│ • Track imports needed                 │
└─────────────────────────────────────────┘
    ↓
┌─────────────────────────────────────────┐
│ Step 3: Import Injection                │
├─────────────────────────────────────────┤
│ • Inject required imports              │
│ • Deduplicate with existing imports    │
└─────────────────────────────────────────┘
    ↓
Valid Go code with qualified calls
```

## Implementation Plan

### Phase 1: Standard Library Registry (Day 1)

#### 1.1 Create `pkg/preprocessor/stdlib_registry.go`

**Data Structures:**

```go
// StdLibRegistry provides comprehensive Go stdlib function mappings
type StdLibRegistry struct {
    // Fast lookups
    qualified   map[string]string      // "os.ReadFile" → "os"
    unqualified map[string]*FuncEntry  // "ReadFile" → FuncEntry
}

// FuncEntry represents a stdlib function
type FuncEntry struct {
    Name       string   // "ReadFile"
    ImportPath string   // "os"
    Ambiguous  bool     // true if exists in multiple packages
    Conflicts  []string // ["net/http"] for "Get"
}
```

**Registry Population Strategy:**

Use Go's `go/doc` and static analysis to generate comprehensive mappings:

1. **Automated Generation** (Initial)
   - Script scans Go stdlib source
   - Extracts all exported functions
   - Detects ambiguities automatically
   - Generates `stdlib_data.go`

2. **Manual Curation** (Refinement)
   - Review ambiguous functions
   - Mark common vs. rare cases
   - Add priority hints if needed

**Covered Packages (Comprehensive):**

```go
// Core packages
"os", "io", "fmt", "errors", "context"

// String/Byte manipulation
"strings", "bytes", "strconv", "unicode"

// Data structures
"sort", "container/list", "container/heap", "container/ring"

// Encoding
"encoding/json", "encoding/xml", "encoding/csv", "encoding/base64"

// Network
"net", "net/http", "net/url", "net/rpc"

// File/Path
"path", "path/filepath", "io/fs"

// Crypto
"crypto", "crypto/md5", "crypto/sha256", "crypto/rand"

// Time/Math
"time", "math", "math/rand"

// Concurrency
"sync", "sync/atomic"

// Compression
"compress/gzip", "compress/zlib"

// Testing
"testing" (for test files only)

// Total: ~40 packages, ~500+ functions
```

**Files Created:**
- `pkg/preprocessor/stdlib_registry.go` - Registry logic
- `pkg/preprocessor/stdlib_data.go` - Generated function mappings
- `tools/gen_stdlib_registry.go` - Code generator script

#### 1.2 Registry API

```go
// NewStdLibRegistry creates registry from generated data
func NewStdLibRegistry() *StdLibRegistry

// LookupQualified checks qualified function call
func (r *StdLibRegistry) LookupQualified(name string) (importPath string, ok bool)

// LookupUnqualified resolves unqualified function call
func (r *StdLibRegistry) LookupUnqualified(name string) (entry *FuncEntry, ok bool)

// IsAmbiguous checks if function name is ambiguous
func (r *StdLibRegistry) IsAmbiguous(name string) (conflicts []string, ok bool)
```

### Phase 2: Local Function Detection (Day 1-2)

#### 2.1 Create `pkg/preprocessor/local_scanner.go`

**Purpose:** Scan user code for locally-defined functions to exclude from transformation.

```go
// LocalScanner extracts user-defined function names from source
type LocalScanner struct {
    fset       *token.FileSet
    localFuncs map[string]bool // "ReadFile" → true if user-defined
}

// ScanFile parses source and extracts local function definitions
func (s *LocalScanner) ScanFile(source string) error

// IsLocal checks if function name is user-defined
func (s *LocalScanner) IsLocal(funcName string) bool
```

**Implementation:**

```go
func (s *LocalScanner) ScanFile(source string) error {
    // Parse source to AST
    fset := token.NewFileSet()
    file, err := parser.ParseFile(fset, "", source, 0)
    if err != nil {
        return err
    }

    // Walk AST and collect function declarations
    ast.Inspect(file, func(n ast.Node) bool {
        switch decl := n.(type) {
        case *ast.FuncDecl:
            s.localFuncs[decl.Name.Name] = true
        }
        return true
    })

    return nil
}
```

**Edge Cases Handled:**
- Package-level functions ✅
- Methods (ignore, not relevant) ✅
- Imported functions (not local) ✅
- Shadowed stdlib names ✅

### Phase 3: Unqualified Import Processor (Day 2-3)

#### 3.1 Create `pkg/preprocessor/unqualified_imports.go`

**Core Processor:**

```go
// UnqualifiedImportProcessor transforms unqualified stdlib calls
type UnqualifiedImportProcessor struct {
    registry     *StdLibRegistry
    localScanner *LocalScanner
    importTracker *ImportTracker
}

// Process scans source and transforms unqualified calls
func (p *UnqualifiedImportProcessor) Process(source string) (transformed string, err error)
```

**Algorithm:**

```
1. Pre-scan: Extract local functions
2. Scan source: Find all function calls
3. For each unqualified call:
   a. Skip if local function
   b. Lookup in unqualified registry
   c. If found and not ambiguous:
      - Transform: func(...) → pkg.func(...)
      - Track import: needed["pkg"] = true
   d. If ambiguous:
      - ERROR with helpful message
   e. If not found:
      - Assume user-defined, no transformation
4. Return transformed source
```

**Transformation Logic:**

```go
func (p *UnqualifiedImportProcessor) transformCall(
    funcName string,
    source string,
    pos int,
) (transformed string, importPath string, err error) {

    // Check local first
    if p.localScanner.IsLocal(funcName) {
        return source, "", nil // No transformation
    }

    // Lookup in registry
    entry, ok := p.registry.LookupUnqualified(funcName)
    if !ok {
        return source, "", nil // Assume user-defined
    }

    // Check ambiguity
    if entry.Ambiguous {
        return "", "", p.createAmbiguityError(funcName, entry.Conflicts)
    }

    // Transform: ReadFile(...) → os.ReadFile(...)
    qualified := entry.ImportPath + "." + funcName
    transformed = strings.Replace(source, funcName+"(", qualified+"(", 1)

    return transformed, entry.ImportPath, nil
}
```

#### 3.2 Ambiguity Error Messages

**Conservative Error Strategy:**

```go
func (p *UnqualifiedImportProcessor) createAmbiguityError(
    funcName string,
    conflicts []string,
) error {
    // Example error:
    // error_prop_ambiguous.dingo:12:5: Ambiguous function 'Open'
    //
    //   Could be:
    //     - os.Open         (most common for files)
    //     - net.Open        (for network connections)
    //
    //   Fix: Use qualified name:
    //     - Change 'Open(path)' to 'os.Open(path)'
    //     - Or import explicitly: 'import "os"' then 'Open(path)'

    return fmt.Errorf(
        "ambiguous function '%s' - could be: %s. Use qualified name.",
        funcName,
        strings.Join(conflicts, ", "),
    )
}
```

**Benefits:**
- Clear error message
- Suggests all valid options
- Forces explicit choice (no surprises)
- Matches Dingo's philosophy (explicit over implicit)

### Phase 4: Integration (Day 3)

#### 4.1 Update `pkg/preprocessor/preprocessor.go`

**Add to preprocessor pipeline:**

```go
func (p *Preprocessor) Process(input Input) (*Result, error) {
    source := input.Source

    // Step 1: Type annotations (existing)
    source, err := p.typeAnnot.Process(source)

    // Step 2: Unqualified imports (NEW)
    unqualifiedProc := NewUnqualifiedImportProcessor(
        NewStdLibRegistry(),
        input.LocalFunctions, // Pre-scanned
    )
    source, imports, err := unqualifiedProc.Process(source)

    // Step 3: Error propagation (existing)
    source, moreImports, err := p.errorProp.Process(source)

    // Merge imports
    allImports := mergeImports(imports, moreImports)

    // Step 4: Inject imports (existing)
    source = injectImportsWithPosition(source, allImports)

    return &Result{Source: source}, nil
}
```

#### 4.2 Update Entry Point (`cmd/dingo/main.go`)

**Pre-scan before preprocessing:**

```go
func transpileFile(path string) error {
    source, _ := os.ReadFile(path)

    // NEW: Pre-scan for local functions
    scanner := preprocessor.NewLocalScanner()
    scanner.ScanFile(string(source))

    // Pass to preprocessor
    input := preprocessor.Input{
        Source:         string(source),
        LocalFunctions: scanner, // Inject exclusion list
    }

    result, err := preprocessor.Process(input)
    // ...
}
```

### Phase 5: Testing & Validation (Day 4)

#### 5.1 Unit Tests

**`pkg/preprocessor/stdlib_registry_test.go`:**
- Registry initialization
- Qualified lookups
- Unqualified lookups
- Ambiguity detection

**`pkg/preprocessor/local_scanner_test.go`:**
- Local function detection
- Edge cases (methods, closures, imports)

**`pkg/preprocessor/unqualified_imports_test.go`:**
- Transformation correctness
- Import tracking
- Ambiguity errors
- Local function exclusion

#### 5.2 Golden Tests

**`tests/golden/import_inference_01_basic.dingo`:**
```go
package main

func main() {
    data? := ReadFile("file.txt")     // → os.ReadFile
    num? := Atoi("42")                // → strconv.Atoi
    Printf("num: %d\n", num)          // → fmt.Printf
}
```

**Expected `.go.golden`:**
```go
package main

import (
    "os"
    "strconv"
    "fmt"
)

func main() {
    data, err := os.ReadFile("file.txt")
    if err != nil { return err }

    num, err := strconv.Atoi("42")
    if err != nil { return err }

    fmt.Printf("num: %d\n", num)
}
```

**`tests/golden/import_inference_02_local_shadowing.dingo`:**
```go
package main

// User-defined ReadFile (should NOT transform)
func ReadFile(path string) ([]byte, error) {
    return nil, nil
}

func main() {
    data? := ReadFile("file.txt")  // Uses local, NO import
}
```

**Expected `.go.golden`:**
```go
package main

// NO import "os" added!

func ReadFile(path string) ([]byte, error) {
    return nil, nil
}

func main() {
    data, err := ReadFile("file.txt")  // Local function
    if err != nil { return err }
}
```

**`tests/golden/import_inference_03_ambiguous_error.dingo`:**
```go
package main

func main() {
    f? := Open("file.txt")  // ERROR: ambiguous (os.Open or net.Open)
}
```

**Expected behavior:**
```
Compile error:
  import_inference_03_ambiguous_error.dingo:4:11:
  Ambiguous function 'Open' - could be: os.Open, net.Open.
  Use qualified name (os.Open or net.Open).
```

#### 5.3 Integration Tests

**End-to-End Test:**
1. Write `.dingo` file with unqualified calls
2. Run `dingo build`
3. Verify `.go` output has qualified calls + imports
4. Verify transpiled code compiles with `go build`
5. Verify LSP source maps point to correct positions

### Phase 6: Documentation (Day 4)

#### 6.1 Update `features/INDEX.md`

Add entry:
```markdown
## Automatic Import Inference

**Status**: Implemented (Phase 4.3)

**File**: `features/import-inference.md`

**Summary**: Unqualified stdlib function calls automatically infer and inject imports.
```

#### 6.2 Create `features/import-inference.md`

Document:
- Motivation (reduce boilerplate)
- Syntax (unqualified calls)
- Transformation behavior
- Ambiguity handling (conservative errors)
- Local function precedence
- Examples and edge cases

#### 6.3 Update `tests/golden/README.md`

Add test category:
```markdown
### Import Inference (`import_inference_*`)

Tests automatic import detection and transformation:
- `01_basic` - Common stdlib functions
- `02_local_shadowing` - User-defined functions with stdlib names
- `03_ambiguous_error` - Ambiguous function error handling
```

## Technical Specifications

### Regex Patterns for Function Call Detection

```go
// Matches unqualified function calls: funcName(...)
var unqualifiedCallPattern = regexp.MustCompile(
    `\b([A-Z][a-zA-Z0-9]*)\s*\(`,
)

// Exclude method calls: receiver.Method(...)
var methodCallPattern = regexp.MustCompile(
    `\w+\.([A-Z][a-zA-Z0-9]*)\s*\(`,
)
```

**Logic:**
1. Find all unqualified calls
2. Exclude method calls (have `.` before name)
3. Check remaining against registry

### Source Map Adjustments

**Transformation shifts positions:**

```
Before: ReadFile("path")
After:  os.ReadFile("path")
        ^^^
        +3 chars
```

**Source map update:**
```go
type SourceMapAdjustment struct {
    Line   int
    Column int
    Shift  int  // +3 for "os."
}
```

### Import Deduplication

**Handle existing imports:**

```go
// User already has:
import "os"

// Processor wants to add:
import "os"

// Result: No duplicate, use existing
```

**Implementation:**
- `injectImportsWithPosition` already handles this ✅
- No changes needed

## Performance Considerations

### Registry Lookup: O(1)

```go
// Hash map lookups
unqualified["ReadFile"] → O(1)
qualified["os.ReadFile"] → O(1)
```

**Expected overhead:**
- Registry initialization: <1ms (one-time)
- Per-function lookup: <0.001ms
- Total transpilation impact: <5ms for typical file

### Pre-Scan Cost

**AST parsing for local functions:**
- `go/parser.ParseFile`: ~10-50ms for typical file
- Acceptable overhead (already parsing later anyway)

**Optimization:**
- Cache local function list per file
- Skip pre-scan if no unqualified calls detected

## Edge Cases & Solutions

### 1. User-Defined Functions with Stdlib Names

**Scenario:**
```go
func ReadFile(path string) []byte { ... }  // User-defined

func main() {
    data := ReadFile("file.txt")  // Should use local, NOT os.ReadFile
}
```

**Solution:**
✅ Pre-scan detects local `ReadFile`
✅ Excludes from transformation
✅ No import added

### 2. Package Aliases

**Scenario:**
```go
import f "fmt"

func main() {
    f.Printf("hello")     // User-explicit
    Printf("world")       // Unqualified
}
```

**Solution:**
- Qualified call (`f.Printf`) → No transformation
- Unqualified call (`Printf`) → Transform to `fmt.Printf`, add `import "fmt"`
- Result: Both `import f "fmt"` and `import "fmt"` present (valid Go)

### 3. Method Calls vs. Function Calls

**Scenario:**
```go
reader.ReadAll()  // Method
ReadAll(reader)   // Function
```

**Solution:**
- Method call regex: `\w+\.ReadAll\(` → Skip
- Function call regex: `\bReadAll\(` → Transform to `io.ReadAll`

### 4. Nested Packages

**Scenario:**
```go
NewRequest(...)  // net/http.NewRequest or http.NewRequest?
```

**Solution:**
- Registry uses full import paths: `"net/http"`
- Transformation: `http.NewRequest` (package name, not path)
- Import: `import "net/http"` (full path)

**Mapping:**
```go
FuncEntry{
    Name:       "NewRequest",
    ImportPath: "net/http",  // Full path
    PkgName:    "http",      // Short name for code
}
```

### 5. Variadic Functions

**Scenario:**
```go
Printf("format", args...)  // Variadic
```

**Solution:**
- Regex matches `Printf(` regardless of args
- Transformation works normally: `fmt.Printf("format", args...)`

## Error Handling

### Ambiguity Errors

**Format:**
```
<file>:<line>:<col>: Ambiguous function '<name>'

Could be:
  - <pkg1>.<name>
  - <pkg2>.<name>

Fix: Use qualified name (e.g., <pkg1>.<name>)
```

**Example:**
```
error_prop_03.dingo:15:5: Ambiguous function 'Open'

Could be:
  - os.Open
  - net.Open

Fix: Use qualified name (e.g., os.Open)
```

### Local Scanner Errors

**Scenario:** Malformed source code

**Behavior:**
- Pre-scan fails → Skip transformation
- Proceed to next stage (will fail later with better error)
- Or: Emit warning and continue (graceful degradation)

**Decision:** Fail early with clear error message ✅

## Success Criteria

### Functional Requirements
- ✅ Unqualified stdlib calls transform to qualified
- ✅ Correct imports automatically injected
- ✅ Ambiguous calls produce compile errors
- ✅ Local functions not transformed
- ✅ All golden tests pass
- ✅ Transpiled code compiles with `go build`

### Quality Requirements
- ✅ Comprehensive test coverage (>90%)
- ✅ Clear error messages with fix suggestions
- ✅ No performance degradation (<5ms overhead)
- ✅ Maintainable registry (easy to add packages)

### Documentation Requirements
- ✅ Feature documentation in `features/`
- ✅ Test documentation in `tests/golden/README.md`
- ✅ Inline code comments for complex logic
- ✅ Examples in golden tests

## Future Enhancements (Out of Scope)

### Phase 2 (Post-MVP):
1. **Smart Ambiguity Resolution**
   - Use go/types to infer from context
   - Example: `file, err := Open(path)` → Infer `os.Open` from return type

2. **Import Grouping**
   - Organize imports by category (stdlib, third-party, local)
   - Match `goimports` style

3. **Third-Party Package Support**
   - Extend registry to popular libraries (gin, echo, etc.)
   - User-configurable package lists

4. **IDE Integration**
   - Quick-fix suggestions in LSP
   - Auto-complete for unqualified calls

## Timeline & Milestones

**Day 1:**
- ✅ Create stdlib registry
- ✅ Generate registry data
- ✅ Implement local scanner

**Day 2:**
- ✅ Implement unqualified import processor
- ✅ Integrate into preprocessor pipeline
- ✅ Basic transformation working

**Day 3:**
- ✅ Add ambiguity detection
- ✅ Fix edge cases (local functions, methods)
- ✅ Update source maps

**Day 4:**
- ✅ Write comprehensive tests
- ✅ Update golden tests
- ✅ Documentation
- ✅ End-to-end validation

**Total Estimate:** 4 days (32 hours)

## Dependencies

### Required
- ✅ `go/parser` - AST parsing for local scanner
- ✅ `go/ast` - AST walking and inspection
- ✅ `go/token` - Position tracking
- ✅ `regexp` - Pattern matching for function calls
- ✅ Existing preprocessor infrastructure

### Optional (Future)
- ⚠️ `go/types` - Advanced type inference for ambiguity resolution
- ⚠️ `golang.org/x/tools/go/packages` - Package discovery

## Risks & Mitigation

### Risk 1: Registry Maintenance Burden
**Impact:** High
**Probability:** Medium
**Mitigation:**
- Automated generation script
- Comprehensive tests catch missing functions
- Community contributions for additions

### Risk 2: Performance Degradation
**Impact:** Medium
**Probability:** Low
**Mitigation:**
- Benchmark tests
- O(1) lookups
- Cache pre-scan results

### Risk 3: False Positives (User Functions)
**Impact:** High
**Probability:** Low
**Mitigation:**
- AST pre-scan is accurate
- Comprehensive test cases
- Error on ambiguity (conservative)

### Risk 4: Breaking Existing Code
**Impact:** High
**Probability:** Low
**Mitigation:**
- Backward compatible (qualified calls still work)
- Golden tests validate existing behavior
- Gradual rollout (feature flag?)

## Conclusion

This implementation provides:

1. **Developer Ergonomics** - Write `ReadFile(path)` instead of `os.ReadFile(path)`
2. **Safety** - Conservative errors prevent ambiguous behavior
3. **Correctness** - Local functions take precedence (no false transforms)
4. **Maintainability** - Clear registry structure, comprehensive tests
5. **Performance** - Minimal overhead (<5ms per file)
6. **Extensibility** - Easy to add more packages over time

**Alignment with Dingo Philosophy:**
- Reduces boilerplate ✅
- Explicit over implicit (ambiguity errors) ✅
- Clean Go output (qualified calls) ✅
- Zero runtime overhead ✅

This feature completes the error propagation experience by eliminating the last remaining boilerplate: import statements for common stdlib functions.

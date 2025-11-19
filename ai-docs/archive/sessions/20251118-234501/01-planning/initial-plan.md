# Architecture Plan: Unqualified Import Inference

## Problem Analysis

### Current State
The preprocessor's `ImportTracker` only detects **qualified** function calls:
- `os.ReadFile(path)` → Adds `import "os"` ✅
- `ReadFile(path)` → Nothing happens ❌

**Root cause**: `stdLibFunctions` map in `error_prop.go:34` only maps qualified names:
```go
var stdLibFunctions = map[string]string{
    "os.ReadFile":  "os",      // Only matches "os.ReadFile"
    "strconv.Atoi": "strconv", // Only matches "strconv.Atoi"
    // ...
}
```

### Design Intent Evidence
All golden tests use **unqualified** function names:
- `tests/golden/error_prop_01_simple.dingo` → `ReadFile(path)`
- Other tests → `Atoi(s)`, `Printf(...)`, etc.

This confirms the original design intended unqualified inference.

## Proposed Solution: Two-Tier Mapping Architecture

### Core Design: Function Name → Package Mapping

Create a **bidirectional mapping system**:

```
Tier 1: Qualified Map (existing)
"os.ReadFile" → "os"
"strconv.Atoi" → "strconv"

Tier 2: Unqualified Map (NEW)
"ReadFile" → "os"
"Atoi" → "strconv"
"Printf" → "fmt"
```

### Architecture Components

#### 1. Enhanced ImportTracker

**Location**: `pkg/preprocessor/error_prop.go`

```go
type ImportTracker struct {
    needed            map[string]bool   // package path → needed
    qualifiedMap      map[string]string // "os.ReadFile" → "os"
    unqualifiedMap    map[string]string // "ReadFile" → "os"
    ambiguousWarnings []string          // Track ambiguous function names
}
```

**Key methods**:
- `TrackFunctionCall(funcName string)` - Enhanced to check both maps
- `ResolveUnqualified(funcName string) (pkg, qualifiedName string, ok bool)`
- `GetAmbiguousWarnings() []string` - For diagnostics

#### 2. Standard Library Function Registry

**Location**: `pkg/preprocessor/stdlib_registry.go` (NEW FILE)

Centralized registry of Go standard library functions:

```go
// StdLibRegistry provides mappings for standard library functions
type StdLibRegistry struct {
    qualified   map[string]string // "os.ReadFile" → "os"
    unqualified map[string]string // "ReadFile" → "os"
    ambiguous   map[string][]string // "Open" → ["os", "net"]
}

// Registry organization by package:
var osPackage = []FuncDef{
    {"ReadFile", false},   // unique
    {"WriteFile", false},
    {"Open", true},        // ambiguous (net.Open also exists)
    // ...
}

var fmtPackage = []FuncDef{
    {"Printf", false},
    {"Sprintf", false},
    {"Errorf", false},
    // ...
}
```

**Benefits**:
- Single source of truth for stdlib functions
- Explicit ambiguity tracking
- Easy to extend with new packages
- Maintainable and testable

#### 3. Import Detection Algorithm

**Enhanced flow**:

```
Input: Function call "ReadFile(...)"
│
├─ Check qualified map: "ReadFile" matches "os.ReadFile"? NO
│
├─ Check unqualified map: "ReadFile" exists?
│  ├─ YES → Found "os"
│  │   ├─ Check if ambiguous (multiple packages)?
│  │   │   ├─ NO → Add import "os", transform to os.ReadFile ✅
│  │   │   └─ YES → Add warning, pick first package (or fail?)
│  │   └─ Return (pkg="os", qualified="os.ReadFile")
│  │
│  └─ NO → User-defined function, no import needed
│
└─ Output: import "os", code uses os.ReadFile
```

#### 4. Ambiguity Handling Strategy

**Problem**: Some function names exist in multiple packages:
- `Open` → `os.Open`, `net.Open`
- `NewReader` → `bytes.NewReader`, `strings.NewReader`, `bufio.NewReader`

**Proposed strategies** (user must choose):

**Option A: Conservative (Recommended)**
- Unqualified ambiguous functions → ERROR
- User MUST use qualified name: `os.Open(...)` or `net.Open(...)`
- Clear error message: "Ambiguous function 'Open'. Did you mean os.Open or net.Open?"

**Option B: Priority-Based**
- Define package priority (os > net > io > ...)
- Pick most common package
- Emit warning in diagnostics

**Option C: Context-Based**
- Use go/types to infer from usage context
- More complex, higher maintenance
- Defer to Phase 2

**Recommendation**: Start with **Option A** (conservative). Simple, predictable, no surprises.

### Implementation Phases

#### Phase 1: Core Infrastructure (Day 1-2)
1. Create `stdlib_registry.go`
   - Define registry data structure
   - Populate with top 20 most common functions
   - Include ambiguity flags

2. Enhance `ImportTracker`
   - Add unqualified map
   - Implement dual-lookup logic
   - Add ambiguity detection

3. Update `TrackFunctionCall`
   - Check qualified map first (backward compat)
   - Fall back to unqualified map
   - Track ambiguous usage

#### Phase 2: Code Transformation (Day 3)
4. Add transformation step
   - After import detection, transform unqualified calls
   - `ReadFile(...)` → `os.ReadFile(...)`
   - Preserve source mappings

5. Update source map generation
   - Track transformations (unqualified → qualified)
   - Ensure LSP positions remain correct

#### Phase 3: Testing & Validation (Day 4)
6. Fix existing golden tests
   - Update `error_prop_01_simple.go.golden`
   - Add explicit import expectations

7. Add comprehensive tests
   - Unqualified inference tests
   - Ambiguity detection tests
   - Edge cases (user-defined functions with stdlib names)

8. Integration testing
   - End-to-end transpilation
   - LSP diagnostics
   - Source map accuracy

## Technical Decisions

### Decision 1: Where to Transform?

**Option A: Preprocessor Stage**
- Transform `ReadFile` → `os.ReadFile` during text preprocessing
- Simpler, happens before AST parsing
- **Chosen approach** ✅

**Option B: AST Plugin Stage**
- Transform during AST processing
- More precise, uses go/types
- More complex, harder to maintain

**Rationale**: Preprocessor stage is simpler and consistent with current architecture.

### Decision 2: Registry vs. Detection

**Option A: Static Registry** (Chosen)
- Hardcode stdlib function mappings
- Fast lookup, no runtime analysis
- Must maintain manually
- **Chosen approach** ✅

**Option B: Dynamic Detection**
- Analyze go/types for stdlib packages
- Automatic, always up-to-date
- Slower, more complex

**Rationale**: Static registry is predictable, testable, and sufficient for common stdlib functions.

### Decision 3: Ambiguity Strategy

**Chosen**: Conservative (Option A)
- Ambiguous unqualified calls → Compile error
- Forces user to be explicit
- Prevents subtle bugs
- **Chosen approach** ✅

### Decision 4: Registry Scope

**Initial scope**: Top 50 most common stdlib functions
- `os`: ReadFile, WriteFile, Open, Create, Stat, Remove, Mkdir, MkdirAll
- `fmt`: Printf, Sprintf, Fprintf, Errorf
- `strconv`: Atoi, Itoa, ParseInt, ParseFloat
- `io`: ReadAll, Copy
- `encoding/json`: Marshal, Unmarshal
- `net/http`: Get, Post, NewRequest
- `path/filepath`: Join, Base, Dir

**Future expansion**: Add more based on usage patterns and user feedback.

## File Changes

### New Files
1. **`pkg/preprocessor/stdlib_registry.go`**
   - `StdLibRegistry` struct
   - Package definitions (os, fmt, strconv, io, etc.)
   - `NewStdLibRegistry()` constructor
   - Registry initialization

### Modified Files
2. **`pkg/preprocessor/error_prop.go`**
   - Enhance `ImportTracker` with unqualified map
   - Add `ResolveUnqualified()` method
   - Update `TrackFunctionCall()` logic
   - Add ambiguity warnings

3. **`pkg/preprocessor/preprocessor.go`**
   - No changes needed (ImportProvider interface already exists)

4. **`tests/golden/error_prop_01_simple.go.golden`**
   - Update expected output with correct imports
   - Verify `os.ReadFile` (qualified) appears in output

### New Test Files
5. **`pkg/preprocessor/stdlib_registry_test.go`**
   - Registry initialization tests
   - Lookup tests (qualified, unqualified)
   - Ambiguity detection tests

6. **`tests/golden/import_inference_01_basic.dingo`**
   - Test unqualified stdlib calls
   - Verify imports added correctly

7. **`tests/golden/import_inference_02_ambiguous.dingo`**
   - Test ambiguous function handling
   - Verify error messages

## Data Structures

### Registry Data Model

```go
// FuncDef defines a standard library function
type FuncDef struct {
    Name       string   // "ReadFile"
    Package    string   // "os"
    Ambiguous  bool     // true if exists in multiple packages
    Conflicts  []string // ["net"] if ambiguous
}

// StdLibRegistry
type StdLibRegistry struct {
    // Fast lookups
    qualified   map[string]string      // "os.ReadFile" → "os"
    unqualified map[string]*FuncDef    // "ReadFile" → FuncDef{...}

    // Package groupings (for documentation/maintenance)
    packages    map[string][]FuncDef   // "os" → [ReadFile, WriteFile, ...]
}
```

### Enhanced ImportTracker

```go
type ImportTracker struct {
    registry          *StdLibRegistry
    needed            map[string]bool      // "os" → true
    ambiguousUsed     map[string][]string  // "Open" → ["file.txt", "conn.txt"]
    transformations   []Transformation     // Track code changes
}

type Transformation struct {
    Original   string // "ReadFile"
    Qualified  string // "os.ReadFile"
    Line       int
    Column     int
}
```

## Edge Cases & Considerations

### 1. User-Defined Functions
**Scenario**: User defines `func ReadFile(path string) error`

**Behavior**:
- If used unqualified: Detected as stdlib, imports `os`, transforms to `os.ReadFile` ❌
- **Fix**: Preprocessor should NOT transform if function is locally defined
- **Solution**: Add pre-scan for local function definitions, exclude from transformation

### 2. Package Aliases
**Scenario**: User imports `import f "fmt"`

**Behavior**:
- Unqualified `Printf` → Adds `import "fmt"`, but user already has alias ❌
- **Fix**: Check existing imports before adding
- **Note**: Already handled by `injectImportsWithPosition` deduplication

### 3. Method Calls vs. Function Calls
**Scenario**: `reader.ReadAll()` (method) vs. `ReadAll(reader)` (function)

**Behavior**:
- Method calls should NOT trigger imports
- Only unqualified function calls should
- **Solution**: Regex/AST detection must distinguish `identifier.method()` from `function()`

### 4. Nested Packages
**Scenario**: `http.NewRequest` vs. `filepath.Join`

**Behavior**:
- Must correctly map to `net/http` and `path/filepath`
- Registry must use full import paths
- **Solution**: Use import path as key, not package name

## Success Metrics

### Functional Requirements
✅ Unqualified stdlib calls automatically import correct package
✅ Code transforms to qualified calls (`ReadFile` → `os.ReadFile`)
✅ Ambiguous calls produce clear error messages
✅ User-defined functions not transformed
✅ Existing golden tests pass with correct imports

### Performance Requirements
- Registry lookup: O(1) constant time
- No measurable impact on transpilation speed (<5ms overhead)

### Maintainability Requirements
- Registry easily extensible (add package = add struct definition)
- Clear separation: registry (data) vs. tracker (logic)
- Comprehensive test coverage (>90%)

## Open Questions (See gaps.json)

1. **Ambiguity strategy confirmation** - Conservative error vs. priority-based?
2. **Registry scope** - Which packages to include initially?
3. **Local function detection** - Pre-scan AST or heuristic-based?
4. **Error message format** - What should ambiguity errors look like?

## Timeline Estimate

- **Phase 1 (Infrastructure)**: 2 days
- **Phase 2 (Transformation)**: 1 day
- **Phase 3 (Testing)**: 1 day
- **Total**: ~4 days for complete implementation

## Dependencies

- ✅ `go/parser` - Already used for AST parsing
- ✅ `go/token` - Already used for position tracking
- ⚠️ `go/types` - Optional, for advanced local function detection
- ✅ Existing `ImportTracker` infrastructure

## Risks & Mitigation

### Risk 1: Breaking Existing Code
**Mitigation**: Comprehensive golden test updates, backward compatibility checks

### Risk 2: Ambiguity Explosion
**Mitigation**: Start with conservative strategy, expand carefully based on usage

### Risk 3: Maintenance Burden
**Mitigation**: Clear registry structure, automated tests, documentation

## Conclusion

This architecture provides:
1. **Backward compatibility** - Qualified calls still work
2. **Developer ergonomics** - Unqualified calls "just work"
3. **Safety** - Ambiguous cases caught at compile time
4. **Maintainability** - Clear registry, simple lookup logic
5. **Extensibility** - Easy to add more packages

The two-tier mapping approach balances simplicity with power, matching Dingo's design philosophy.

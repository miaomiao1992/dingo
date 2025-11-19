# Pain Points Analysis: Current File Organization

## Critical Pain Points (Must Fix)

### P1: Test Directory Clutter (Severity: HIGH)

**Current State**: `tests/golden/` contains 242 files for 62 test cases.

**File Breakdown per Test**:
```
error_prop_01_simple.dingo          ← Source (tracked)
error_prop_01_simple.go             ← Intermediate build artifact (not tracked)
error_prop_01_simple.go.actual      ← Test output (temporary)
error_prop_01_simple.go.golden      ← Expected output (tracked)
error_prop_01_simple.go.map         ← Source map (tracked)
error_prop_01_simple.reasoning.md   ← Documentation (tracked)
```

**Impact**:
- **Developer Experience**: Hard to find tests in file tree
- **Cognitive Load**: 80% of files are artifacts/noise
- **Navigation**: IDE search cluttered with generated files
- **Maintenance**: Difficult to see what tests exist

**Desired State**:
```
tests/golden/
├── source/
│   └── error_prop_01_simple.dingo
├── expected/
│   └── error_prop_01_simple.go
├── maps/
│   └── error_prop_01_simple.go.map
└── docs/
    └── error_prop_01_simple.reasoning.md
```

**OR** (alternative):
```
tests/golden/
└── error_prop_01_simple/
    ├── input.dingo
    ├── expected.go
    ├── source.map
    └── reasoning.md
```

**Benefit**: 62 logical units instead of 242 scattered files.

---

### P2: .gitignore Complexity (Severity: HIGH)

**Current Problem**: Cannot cleanly separate generated from source files.

**Bad Option 1** (ignore all .go):
```gitignore
*.go
!tests/golden/*.go.golden
```
**Issue**: Ignores hand-written Go files in examples, scripts, etc.

**Bad Option 2** (explicit exclusions):
```gitignore
examples/hello.go
examples/math.go
examples/error_propagation/http_client.go
# ... 100+ more lines
```
**Issue**: Must update .gitignore for every new .dingo file.

**Bad Option 3** (commit generated files):
```
# No .gitignore for generated files
```
**Issue**: Commits grow large, diffs noisy, merge conflicts on generated code.

**Desired State**:
```gitignore
/build/
/gen/
```
**Benefit**: One line, covers everything, never needs updating.

---

### P3: Name Collision Risk (Severity: MEDIUM)

**Current Problem**: Can't have `foo.dingo` and hand-written `foo.go` in same directory.

**Example Scenario**:
```
myproject/
├── handlers.dingo      ← Dingo source
└── handlers_test.go    ← Hand-written Go test
```

**When building**: `dingo build handlers.dingo` → `handlers.go`

Now directory contains:
```
myproject/
├── handlers.dingo      ← Original source
├── handlers.go         ← Generated (might overwrite accidental manual file!)
└── handlers_test.go    ← Hand-written
```

**Risk**: User creates `handlers.go` manually, then runs `dingo build` → file overwritten!

**Desired State**: Generated files in separate directory, no collision possible.

---

### P4: Mixed Source Trees (Severity: MEDIUM)

**Problem**: Hard to distinguish generated from source in IDE/editor.

**Current View in IDE**:
```
myproject/
├── main.dingo          ← SOURCE (edit this)
├── main.go             ← GENERATED (don't edit!)
├── main.go.map         ← GENERATED
├── utils.dingo         ← SOURCE
├── utils.go            ← GENERATED
├── utils.go.map        ← GENERATED
├── types.go            ← HAND-WRITTEN (edit this)
└── types_test.go       ← HAND-WRITTEN
```

**Issues**:
- **Visual Confusion**: Which files are source of truth?
- **Accidental Edits**: User edits `main.go`, changes lost on next build
- **Code Review**: PRs show both source and generated (noise)
- **Search Results**: IDE search returns both `.dingo` and `.go` (duplicates)

**Desired State**:
```
myproject/
├── main.dingo          ← SOURCE (edit this)
├── utils.dingo         ← SOURCE
├── types.go            ← HAND-WRITTEN
├── types_test.go       ← HAND-WRITTEN
└── build/
    ├── main.go         ← GENERATED (hidden from normal view)
    └── utils.go        ← GENERATED
```

**Benefit**: Clear separation, no confusion.

---

## Moderate Pain Points (Should Fix)

### P5: CI/CD Cleanup Complexity (Severity: MEDIUM)

**Current Problem**: Hard to clean build artifacts in CI.

**Current Cleanup**:
```bash
# Must know all possible output locations
rm -f examples/**/*.go
rm -f examples/**/*.go.map
rm -f tests/golden/*.go
rm -f tests/golden/*.go.actual
rm -f tests/golden/*.go.map
# ... what if we missed some?
```

**Desired Cleanup**:
```bash
rm -rf build/
```

**Benefit**: Guaranteed clean state, no missed files.

---

### P6: IDE Configuration Burden (Severity: MEDIUM)

**Current Problem**: Users must configure IDE to hide generated files.

**VS Code** `.vscode/settings.json`:
```json
{
  "files.exclude": {
    "**/*.go.map": true,
    "examples/**/*.go": true,
    "tests/**/*.go": true
  },
  "search.exclude": {
    "**/*.go.map": true,
    "examples/**/*.go": true
  }
}
```

**Issue**: Every Dingo project needs custom exclusions.

**Desired State**: Generated files in `build/`, VS Code auto-hides directories starting with `build/`.

---

### P7: LSP Tight Coupling (Severity: MEDIUM)

**Current Issue**: LSP assumes `.go` file adjacent to `.dingo`.

**Code** (`pkg/lsp/handlers.go` - inferred):
```go
func (s *Server) handleDefinition(dingoFile string, line, col int) {
    // Assumes: foo.dingo → foo.go (same directory)
    goFile := strings.Replace(dingoFile, ".dingo", ".go", 1)
    sourceMap := goFile + ".map"

    // Load source map, translate position, forward to gopls
    // ...
}
```

**Problem**: Hard to refactor if we want to move output files.

**Desired State**: LSP reads `dingo.toml` to discover output structure.

```go
func (s *Server) handleDefinition(dingoFile string, line, col int) {
    cfg := config.Load()
    goFile := cfg.ResolveOutputPath(dingoFile)     // Reads dingo.toml
    sourceMap := cfg.ResolveSourceMapPath(goFile)  // Configurable

    // ...
}
```

---

### P8: Build Cache Absence (Severity: LOW)

**Current Issue**: Every build regenerates all files, even if unchanged.

**Example**:
```bash
dingo build main.dingo  # Takes 50ms
# (no changes)
dingo build main.dingo  # Still takes 50ms
```

**Desired State** (like TypeScript):
```bash
dingo build main.dingo  # Takes 50ms, writes build/main.go
# (no changes)
dingo build main.dingo  # Skips build (0ms) - output unchanged
```

**Benefit**: Faster builds for large projects (not critical for current scale).

---

## Low Priority Pain Points (Nice to Have)

### P9: No Clean Command (Severity: LOW)

**Current Issue**: No standard way to clean outputs.

**Workaround**:
```bash
# User must manually delete
rm *.go *.go.map
```

**Desired**:
```bash
dingo clean  # Deletes build/ directory
```

**Benefit**: Convenience, aligns with Go (`go clean`), Rust (`cargo clean`).

---

### P10: Watch Mode Incomplete (Severity: LOW)

**Current State**: `dingo build --watch` not implemented (prints warning).

**With Separate Build Directory**: Watch mode easier to implement (monitor source tree, write to build tree).

**Benefit**: Auto-rebuild on save (developer ergonomics).

---

## User Proposal Evaluation: "Shadow Folder"

**User's Suggested Structure**:

1. **Source folder**: Mix of `.dingo` and `.go` (original sources)
2. **Shadow folder**: Only `.go` (transpiled outputs, different package)
3. **Map folder**: Only `.go.map` (source maps)

### Concerns with "Shadow as Different Package"

**Problem**: User suggested shadow folder uses "different package".

**Go's Package Model**:
- A **package** is a compilation unit (all `.go` in a directory)
- Import path = module name + directory path
- Can't have same package in multiple directories

**Example Confusion**:
```
myproject/
├── src/
│   └── main.dingo         // package main
└── shadow/
    └── main.go            // package main (DIFFERENT LOCATION!)
```

**Go Build**:
```bash
go build shadow/main.go    # Compiles
go build src/main.dingo    # Error: not a .go file
```

**But**: Where do imports point?
```go
// In src/utils.dingo
package myapp

func Helper() {}
```

**Transpiles to** `shadow/utils.go`:
```go
package myapp  // Same package name, different directory

func Helper() {}
```

**Consumer** (hand-written `main.go`):
```go
import "myproject/myapp"  // Which directory? src/ or shadow/?
```

**Resolution**: Shadow folder should NOT be a different package, just a different **location** for the SAME package's generated files.

### Refined Shadow Folder Concept

**Better Structure**:
```
myproject/
├── main.dingo             (source)
├── utils.dingo            (source)
├── types.go               (hand-written Go, same package)
├── build/                 (generated outputs)
│   ├── main.go            (transpiled from main.dingo)
│   └── utils.go           (transpiled from utils.dingo)
└── maps/                  (source maps)
    ├── main.go.map
    └── utils.go.map
```

**Go Build**:
```bash
# Must combine hand-written + generated
go build -o bin/app ./main.dingo types.go build/main.go build/utils.go
```

**Issue**: Go doesn't know how to build mixed source/generated trees automatically.

**Better Approach**: Put ALL Go files (hand-written + generated) in same directory.

---

## Recommended Solution: Build Directory + Package Awareness

### Proposed Structure (for Go Module Compatibility)

**Monolithic Package** (simple projects):
```
myproject/
├── cmd/
│   └── main.dingo
├── build/
│   └── cmd/
│       └── main.go
└── dingo.toml
```

**Multi-Package** (larger projects):
```
myproject/
├── cmd/
│   └── server/
│       └── main.dingo
├── internal/
│   └── handlers/
│       ├── http.dingo
│       └── auth.go        (hand-written)
├── build/
│   ├── cmd/server/
│   │   └── main.go
│   └── internal/handlers/
│       └── http.go
└── dingo.toml
```

**Key Insight**: Mirror source tree in `build/`, preserving package structure.

**Go Build**:
```bash
# Combine source and build trees
GOPATH=$PWD/build:$PWD go build ./cmd/server
```

**OR** (if all files transpiled):
```bash
go build -C build ./cmd/server
```

---

## Priority Ranking

| Pain Point | Severity | User Impact | Fix Complexity | Priority |
|------------|----------|-------------|----------------|----------|
| P1: Test Clutter | HIGH | Every test run | Medium | **Must Fix** |
| P2: .gitignore | HIGH | Every commit | Low | **Must Fix** |
| P4: Mixed Trees | MEDIUM | Daily confusion | Low | **Should Fix** |
| P3: Name Collision | MEDIUM | Rare but severe | Low | **Should Fix** |
| P5: CI Cleanup | MEDIUM | Every CI run | Low | **Should Fix** |
| P7: LSP Coupling | MEDIUM | Refactoring blocker | Medium | **Should Fix** |
| P6: IDE Config | MEDIUM | Per-project setup | Low | Nice to Have |
| P8: Build Cache | LOW | Performance | High | Nice to Have |
| P9: Clean Command | LOW | Convenience | Low | Nice to Have |
| P10: Watch Mode | LOW | Developer UX | Medium | Nice to Have |

---

## Conclusion

**Critical Fixes**:
1. Separate build directory (fixes P1, P2, P4, P5)
2. Configurable output paths (fixes P3, P7)

**Must Implement**:
- `dingo.toml` build configuration
- LSP path resolution updates
- Mirror source tree in build directory

**High Impact, Low Effort**:
- P2 (.gitignore): One line instead of 100+
- P1 (Test Clutter): Immediate clarity in test suite
- P4 (Mixed Trees): Clear mental model for users

**Next Steps**:
1. Design `dingo.toml` build section schema
2. Prototype build directory mirroring logic
3. Update LSP path resolution
4. Migrate golden tests to structured layout

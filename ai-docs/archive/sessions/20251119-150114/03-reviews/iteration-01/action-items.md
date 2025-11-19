# Action Items - Phase V Implementation

**Session**: 20251119-150114
**Date**: 2025-11-19
**Priority**: Critical and Important issues only

---

## CRITICAL Issues (Must Fix Before Merge)

### 1. Fix Build Cache Dependency Tracking
**Severity**: CRITICAL
**File**: `pkg/build/cache.go:137`
**Reviewers**: Grok 4 Fast, Internal

**Problem**: `dependenciesChanged()` returns `false` (placeholder) instead of checking if imported .dingo files changed.

**Fix**:
```go
func (c *Cache) dependenciesChanged(pkg Package, cachedDeps []string) bool {
    currentDeps, err := extractImports(pkg.SourcePath)
    if err != nil {
        return true // Assume changed on error (safe)
    }

    if len(currentDeps) != len(cachedDeps) {
        return true
    }

    depSet := make(map[string]bool)
    for _, dep := range cachedDeps {
        depSet[dep] = true
    }

    for _, dep := range currentDeps {
        if !depSet[dep] {
            return true
        }
    }

    return false
}

func extractImports(sourcePath string) ([]string, error) {
    data, err := os.ReadFile(sourcePath)
    if err != nil {
        return nil, err
    }

    importRegex := regexp.MustCompile(`import\s+"([^"]+)"`)
    matches := importRegex.FindAllSubmatch(data, -1)

    deps := make([]string, 0, len(matches))
    for _, match := range matches {
        deps = append(deps, string(match[1]))
    }
    return deps, nil
}
```

**Estimated Time**: 2 hours

---

### 2. Fix Dependency Graph Import Resolution
**Severity**: CRITICAL
**File**: `pkg/build/dependency_graph.go:106-126`
**Reviewers**: Grok 4 Fast, Internal

**Problem**: `importPathToPackagePath()` returns empty string for all absolute imports.

**Fix**:
```go
func importPathToPackagePath(importPath, workspaceRoot string) string {
    // Handle relative imports
    if strings.HasPrefix(importPath, ".") {
        return strings.TrimPrefix(importPath, "./")
    }

    // Read module path from go.mod
    modPath, err := getModulePath(workspaceRoot)
    if err != nil {
        return "" // External dependency
    }

    // Check if import is within workspace module
    if strings.HasPrefix(importPath, modPath) {
        relPath := strings.TrimPrefix(importPath, modPath+"/")
        return relPath
    }

    // External dependency, not tracked
    return ""
}

func getModulePath(root string) (string, error) {
    goMod := filepath.Join(root, "go.mod")
    data, err := os.ReadFile(goMod)
    if err != nil {
        return "", err
    }

    moduleRegex := regexp.MustCompile(`module\s+([^\s]+)`)
    match := moduleRegex.FindSubmatch(data)
    if match == nil {
        return "", fmt.Errorf("no module declaration found")
    }

    return string(match[1]), nil
}
```

**Estimated Time**: 3 hours

---

### 3. Add Documentation Example Validation to CI
**Severity**: CRITICAL
**File**: `.github/workflows/enhanced-ci.yml`
**Reviewer**: Gemini 3 Pro

**Problem**: No CI validation that code examples in docs/ and examples/ actually compile.

**Fix**:
```yaml
# Add to .github/workflows/enhanced-ci.yml

- name: Validate documentation examples
  run: |
    # Extract code blocks from docs/*.md
    ./scripts/extract-doc-examples.sh
    # Compile all extracted examples
    cd /tmp/doc-examples && go build ./...

- name: Test example projects
  run: |
    cd examples/library-example && go build ./...
    cd ../app-example && dingo build ./... && go build ./...
    cd ../hybrid-example && go build ./...
```

**Additional**: Create `scripts/extract-doc-examples.sh` to extract Go code blocks from markdown.

**Estimated Time**: 4 hours (script + CI integration)

---

### 4. Enhance Source Map Validator Error Messages
**Severity**: CRITICAL
**File**: `pkg/sourcemap/validator.go`
**Reviewer**: Gemini 3 Pro

**Problem**: Error messages lack context (which file, which line, what validation failed).

**Fix**:
```go
// BAD (current)
return fmt.Errorf("round-trip validation failed")

// GOOD (add context)
return fmt.Errorf("source map %s: round-trip validation failed\n"+
    "  Dingo position: line %d, column %d\n"+
    "  → Go position: line %d, column %d\n"+
    "  → Round-trip: line %d, column %d (expected %d, %d)",
    path, dingoLine, dingoCol, goLine, goCol,
    backLine, backCol, dingoLine, dingoCol)
```

**Apply to all validation errors**: schema validation, mapping validation, consistency checks.

**Estimated Time**: 2 hours

---

### 5. Add Race Detection for Workspace Builds
**Severity**: CRITICAL
**Files**: `pkg/build/workspace.go`, `pkg/build/cache.go`, `.github/workflows/enhanced-ci.yml`
**Reviewer**: Gemini 3 Pro

**Problem**: No concurrency safety verification for parallel builds.

**Fix**:

**In `pkg/build/workspace.go`**:
```go
type Builder struct {
    mu       sync.Mutex
    cache    *Cache
}

func (b *Builder) buildPackage(pkg Package) error {
    // Each package writes to its own directory (no shared state)

    // Only lock when updating shared cache
    b.mu.Lock()
    b.cache.MarkBuilt(pkg.Path, pkg.Hash)
    b.mu.Unlock()
}
```

**In `.github/workflows/enhanced-ci.yml`**:
```yaml
- name: Test build system with race detector
  run: go test -race ./pkg/build/...
```

**Document concurrency safety** in `pkg/build/cache.go` and `pkg/build/workspace.go` package comments.

**Estimated Time**: 3 hours (mutex + documentation + CI)

---

## IMPORTANT Issues (Should Fix Soon)

### 6. Add Comprehensive Test Suite for pkg/build/
**Severity**: IMPORTANT
**Files**: `pkg/build/workspace_test.go`, `cache_test.go`, `dependency_graph_test.go`
**Reviewers**: Grok 4 Fast, Gemini 3 Pro

**Problem**: Entire build system has zero test coverage.

**Fix**:
```go
// pkg/build/workspace_test.go
func TestWorkspaceScan(t *testing.T) { /* ... */ }
func TestWorkspaceIgnore(t *testing.T) { /* ... */ }

// pkg/build/cache_test.go
func TestCacheInvalidation(t *testing.T) { /* ... */ }
func TestCacheHashConsistency(t *testing.T) { /* ... */ }

// pkg/build/dependency_graph_test.go
func TestCircularDependencyDetection(t *testing.T) { /* ... */ }
func TestTopologicalSort(t *testing.T) { /* ... */ }
```

**Add coverage enforcement**:
```yaml
# .github/workflows/enhanced-ci.yml
- name: Test infrastructure with coverage
  run: |
    go test -coverprofile=coverage.txt -covermode=atomic \
      ./pkg/sourcemap/... ./pkg/build/... ./cmd/dingo/...

    # Require 80% coverage
    go tool cover -func=coverage.txt | grep total | \
      awk '{if ($3+0 < 80) exit 1}'
```

**Estimated Time**: 2 days

---

### 7. Improve Dependency Graph Cycle Error Messages
**Severity**: IMPORTANT
**File**: `pkg/build/dependency_graph.go`
**Reviewer**: Gemini 3 Pro

**Problem**: Likely returns generic "circular dependency detected" without showing cycle path.

**Fix**:
```go
func (g *DependencyGraph) DetectCycles() error {
    visited := make(map[string]bool)
    recStack := make(map[string]bool)

    for pkg := range g.nodes {
        if cycle := g.dfs(pkg, visited, recStack, []string{}); cycle != nil {
            return fmt.Errorf("circular dependency detected: %s",
                strings.Join(cycle, " → "))
        }
    }
    return nil
}

// Example error:
// "circular dependency detected: pkg/auth → pkg/user → pkg/auth"
```

**Estimated Time**: 2 hours

---

### 8. Add PR Comment Integration for Performance Regressions
**Severity**: IMPORTANT
**File**: `.github/workflows/benchmarks.yml` (or enhanced-ci.yml)
**Reviewer**: Gemini 3 Pro

**Problem**: Performance tracking stores metrics but doesn't alert developers.

**Fix**:
```yaml
- name: Detect performance regressions
  run: |
    ./scripts/detect-regressions.sh bench-results.txt baseline.txt > regression-report.md

    # Post as PR comment AND fail CI
    if [ $? -ne 0 ]; then
      gh pr comment ${{ github.event.pull_request.number }} \
        --body-file regression-report.md
      exit 1
    fi
```

**Estimated Time**: 3 hours

---

### 9. Fix Workspace Pattern Matching
**Severity**: IMPORTANT
**File**: `cmd/dingo/workspace.go:206-237`
**Reviewer**: Internal

**Problem**: Simplified glob matching has edge cases with `**/test/*` patterns.

**Fix**: Use `filepath.Match` or `github.com/gobwas/glob` for proper glob support.

**Estimated Time**: 2 hours

---

### 10. Integrate Transpiler with Workspace Builder
**Severity**: IMPORTANT
**File**: `pkg/build/workspace.go:231-237`
**Reviewer**: Internal

**Problem**: Transpiler call is commented out (placeholder).

**Fix**:
```go
// Accept transpiler as dependency injection
type TranspileFunc func(string) error

func (b *WorkspaceBuilder) SetTranspiler(fn TranspileFunc) {
    b.transpile = fn
}

// Call during build
if err := b.transpile(fullPath); err != nil {
    result.Error = err
    return result
}
```

**Estimated Time**: 2 hours

---

### 11. Add Migration ROI Metrics to Documentation
**Severity**: IMPORTANT
**File**: `docs/migration-from-go.md`
**Reviewer**: Gemini 3 Pro

**Problem**: No concrete before/after LOC counts or complexity metrics.

**Fix**:
```markdown
## ROI Analysis: Real-World Examples

### Example 1: User Service (Error Handling)
**Before (Pure Go)**: 87 lines, 12 error checks
**After (Dingo)**: 34 lines, 2 error checks
**Reduction**: 61% fewer lines, 83% fewer error checks

### Aggregate Metrics (10 production codebases)
- Average LOC reduction: 58%
- Average cyclomatic complexity reduction: 47%
- Bug density improvement: 23% fewer nil panics
```

**Estimated Time**: 3 hours

---

### 12. Document Cache Invalidation Strategy
**Severity**: IMPORTANT
**Files**: `pkg/build/cache.go`, `docs/workspace-builds.md`
**Reviewer**: Gemini 3 Pro

**Problem**: Cache invalidation rules unclear.

**Fix**:
```go
type CacheKey struct {
    DingoHash   string // SHA256 of .dingo file
    ImportsHash string // SHA256 of all imported .dingo files
    EngineHash  string // SHA256 of dingo binary
}

func (c *Cache) IsValid(pkg Package) bool {
    cached := c.Get(pkg.Path)
    current := c.computeKey(pkg)

    return cached.DingoHash == current.DingoHash &&
           cached.ImportsHash == current.ImportsHash &&
           cached.EngineHash == current.EngineHash
}
```

**Document in** `docs/workspace-builds.md` with clear explanation.

**Estimated Time**: 2 hours

---

### 13. Improve Diff Visualizer Algorithm
**Severity**: IMPORTANT
**File**: `scripts/diff-visualizer.go:228-265`
**Reviewer**: Internal

**Problem**: Set-based diff is inaccurate (duplicate lines, no order).

**Fix**: Use Myers diff algorithm or `github.com/sergi/go-diff` library.

**Estimated Time**: 3 hours

---

## Summary

**CRITICAL Issues**: 5 (Estimated total: 1-2 days)
**IMPORTANT Issues**: 8 (Estimated total: 3-5 days)

**Total Estimated Fix Time**: 4-7 days for all critical + important issues

**Recommended Approach**:
1. Fix all 5 CRITICAL issues (1-2 days) → Merge
2. Address IMPORTANT issues post-merge (1 week)
3. Polish MINOR issues as time permits

**Merge Readiness**: After CRITICAL fixes complete

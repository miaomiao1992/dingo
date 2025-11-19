# Phase V Testing Plan

## Overview
Testing infrastructure components added in Phase V: package management, source map validation, CI/CD tools, and workspace builds.

## Test Categories

### 1. Package Management (examples/)
**Purpose**: Verify multi-package structure works correctly

**Tests**:
- `library-example/` compiles and exports expected types
- `app-example/` compiles and runs without errors
- `hybrid-example/` successfully integrates library
- All examples have valid go.mod files
- Dependencies resolve correctly

**Success Criteria**: All 3 examples build successfully

### 2. Source Map Validation (pkg/sourcemap/)
**Purpose**: Ensure source map accuracy meets >99.9% target

**Tests**:
- Validator compiles without errors
- Runs against existing .sourcemap files in tests/golden/
- Performance: <10ms per file
- Accuracy: Reports any mismatches
- Edge cases: Empty files, complex transformations

**Success Criteria**: Validator runs, reports accuracy metrics

### 3. CI/CD Tools (scripts/)
**Purpose**: Verify tooling infrastructure works

**Tests**:
- `diff-visualizer.go` compiles
- `performance-tracker.go` compiles
- GitHub Actions workflow has valid YAML syntax
- Scripts have correct executable permissions

**Success Criteria**: All tools compile, workflow validates

### 4. Workspace Builds (pkg/build/, cmd/dingo/)
**Purpose**: Multi-package build orchestration

**Tests**:
- Workspace detection finds dingo.toml
- Multi-package builds complete successfully
- Dependency graph construction works
- Build cache (if implemented) functions correctly
- Error handling for missing dependencies

**Success Criteria**: Workspace builds complete without errors

### 5. Documentation Quality
**Purpose**: Ensure docs are accurate and complete

**Tests**:
- All code examples in docs/ are syntactically valid
- No broken internal links
- Markdown lints clean
- README files exist in all major directories

**Success Criteria**: No broken examples or links

## Testing Approach

### Phase 1: Compilation Tests
Verify all new Go code compiles:
- pkg/sourcemap/validator.go
- scripts/diff-visualizer.go
- scripts/performance-tracker.go
- examples/*/main.go

### Phase 2: Functional Tests
Run actual operations:
- Build each example project
- Run source map validator
- Execute workspace builds

### Phase 3: Integration Tests
Test components working together:
- Build workspace with multiple packages
- Validate source maps from workspace build
- Track performance metrics

## Risk Areas

**High Risk**:
- Workspace dependency resolution (complex logic)
- Source map validation accuracy (critical for LSP)

**Medium Risk**:
- Example project structure (may need refinement)
- CI/CD workflow syntax (GitHub Actions specific)

**Low Risk**:
- Tool compilation (straightforward Go code)
- Documentation formatting (mechanical checks)

## Test Execution Order

1. **Compilation checks** (fast, catches syntax errors)
2. **Example builds** (validates package structure)
3. **Workspace builds** (tests core Phase V feature)
4. **Source map validation** (validates quality)
5. **CI/CD tools** (validates tooling)
6. **Documentation** (validates completeness)

## Success Metrics

**Must Pass**:
- All Go code compiles without errors
- At least 1 example project builds successfully
- Source map validator runs without crashes

**Should Pass**:
- All 3 example projects build
- Workspace builds complete
- >95% source map accuracy

**Nice to Have**:
- All documentation examples valid
- Performance benchmarks run
- CI/CD workflow validates

## Time Estimate
- Total testing: ~5-10 minutes
- Writing detailed report: ~3 minutes
- **Total**: ~8-13 minutes

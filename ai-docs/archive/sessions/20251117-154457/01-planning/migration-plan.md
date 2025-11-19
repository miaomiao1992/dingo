# Migration Plan: Participle → go/parser + Preprocessor

**Date:** 2025-11-17
**Status:** Approved for Implementation
**Risk Level:** Medium (breaking change with mitigation)

---

## Overview

This migration replaces the current Participle-based parser with a two-stage architecture:
1. **Preprocessor:** Dingo syntax → valid Go (with placeholders)
2. **go/parser:** Parse valid Go → AST
3. **Transformer:** Replace placeholders → final Go code

**Goal:** Preserve all existing functionality while establishing cleaner architecture for future growth.

---

## Current State Assessment

### Working Features (MUST NOT BREAK)

| Feature | Golden Tests | Status | Complexity |
|---------|--------------|--------|------------|
| Error Propagation (`?`) | 8 tests | ✅ Passing | Medium |
| Sum Types (`enum`) | 2 tests | ✅ Passing | High |
| Lambdas (`\|x\| expr`) | 3 tests | ✅ Passing | Medium |
| Pattern Matching | 4 tests | ✅ Passing | High |
| Ternary (`? :`) | 2 tests | ✅ Passing | Low |
| Null Coalescing (`??`) | 2 tests | ✅ Passing | Low |
| Safe Navigation (`?.`) | 2 tests | ✅ Passing | Low |

**Total:** 23 golden tests must continue passing throughout migration.

### Current Architecture

```
.dingo file
    ↓
Participle Parser (pkg/parser/parser.go)
    ↓
Custom AST (pkg/parser/ast.go)
    ↓
Plugins (plugins/*.go) - Transform AST
    ↓
Go Code Generation (pkg/generator/generator.go)
    ↓
.go file
```

### Dependencies to Preserve

- CLI commands: `dingo build`, `dingo run`
- Test infrastructure: `tests/golden/run_golden_tests.go`
- Source map generation (basic)
- Error reporting

---

## Migration Strategy: Incremental Parallel Development

We will build the new architecture **alongside** the existing implementation, then cut over feature by feature. This minimizes risk and allows rollback at any point.

### Phase 0: Infrastructure Setup (No Breaking Changes)

**Goal:** Build core infrastructure without touching existing code.

**Duration:** 2-3 days

**Tasks:**

1. Create new package structure (no conflicts with existing):
   ```
   pkg/
     preprocessor/       # NEW
     transform/          # NEW
     parser2/            # NEW (temporary name to avoid conflict)
     generator2/         # NEW (temporary name)
   ```

2. Implement basic preprocessor framework:
   - `preprocessor/preprocessor.go` - orchestration
   - `preprocessor/sourcemap.go` - position tracking
   - No feature processors yet (just pass-through)

3. Implement go/parser wrapper:
   - `parser2/parser.go` - calls preprocessor + go/parser
   - Error position mapping

4. Create test harness:
   - `preprocessor/preprocessor_test.go` - unit tests
   - `parser2/parser_test.go` - integration tests

**Success Criteria:**
- Can preprocess empty file (pass-through)
- Can parse preprocessed Go with go/parser
- Source map tracks basic positions
- No existing tests broken (we haven't touched old code)

**Rollback:** Delete new packages (no impact).

---

### Phase 1: Migrate Error Propagation (First Feature)

**Goal:** Fully migrate error propagation (`?` operator) to new architecture, validate approach.

**Duration:** 3-4 days

**Tasks:**

1. Implement error propagation preprocessor:
   ```go
   // pkg/preprocessor/error_prop.go
   type ErrorPropProcessor struct{}

   func (p *ErrorPropProcessor) Process(source []byte) ([]byte, []Mapping, error) {
       // Scan for ? operator
       // Replace expr? with __dingo_try_N__(expr)
       // Track position mappings
   }
   ```

2. Implement error propagation transformer:
   ```go
   // pkg/transform/error_prop.go
   type ErrorPropTransformer struct{}

   func (t *ErrorPropTransformer) Transform(cursor *astutil.Cursor) bool {
       // Find __dingo_try_N__ calls
       // Analyze context (assignment, return, etc.)
       // Replace with error handling code
   }
   ```

3. Create parallel build path:
   ```go
   // cmd/dingo/build.go
   func buildFile(path string) error {
       if useNewParser {
           return buildFileV2(path)  // NEW
       }
       return buildFileV1(path)      // EXISTING
   }
   ```

4. Add feature flag:
   ```bash
   $ dingo build --experimental-parser file.dingo
   ```

5. Test parity:
   - Run 8 error propagation golden tests with new parser
   - Compare output with old parser (should be identical)
   - Test error messages (positions correct?)

**Success Criteria:**
- All 8 error propagation tests pass with `--experimental-parser`
- Generated Go code identical to old parser (or better)
- Error messages point to correct .dingo positions
- Old parser still default (no breaking changes)

**Rollback:** Remove `--experimental-parser` flag, keep infrastructure for next attempt.

**Validation Approach:**

```bash
# Test old parser
$ dingo build tests/golden/error_prop_01_simple.dingo
$ cp output/error_prop_01_simple.go /tmp/old_output.go

# Test new parser
$ dingo build --experimental-parser tests/golden/error_prop_01_simple.dingo
$ cp output/error_prop_01_simple.go /tmp/new_output.go

# Compare
$ diff /tmp/old_output.go /tmp/new_output.go
# Should be identical (or new version cleaner)
```

---

### Phase 2: Migrate Lambdas

**Goal:** Validate type inference works correctly.

**Duration:** 3-4 days

**Tasks:**

1. Implement lambda preprocessor:
   - Transform `|x, y| expr` → `__dingo_lambda_N__([]string{"x", "y"}, func() interface{} { return expr })`

2. Implement lambda transformer with type inference:
   - Find `__dingo_lambda_N__` calls
   - Infer types from context using go/types
   - Rebuild function literal with proper types

3. Test with existing 3 lambda golden tests

**Success Criteria:**
- All 3 lambda tests pass
- Type inference correctly handles:
  - Simple cases: `numbers.Map(|x| x * 2)`
  - Multiple parameters: `pairs.Filter(|k, v| v > 10)`
  - Nested lambdas

**Complexity Note:** This phase validates our type inference strategy, which is critical for other features.

---

### Phase 3: Migrate Simple Operators (Ternary, ??, ?.)

**Goal:** Quick wins, validate full expansion in preprocessor.

**Duration:** 2 days

**Tasks:**

1. Implement operator preprocessor:
   - Ternary: `cond ? a : b` → `func() T { if cond { return a } else { return b } }()`
   - Null coalescing: `a ?? b` → `func() T { if __tmp := a; __tmp != nil { return __tmp } else { return b } }()`
   - Safe navigation: `obj?.field` → `__dingo_safe_nav_N__(obj, "field")`

2. Test with 6 operator golden tests

**Success Criteria:**
- All 6 operator tests pass
- No transformer needed for ternary and null coalescing (fully expanded)

---

### Phase 4: Migrate Pattern Matching

**Goal:** Validate complex AST generation.

**Duration:** 4-5 days

**Tasks:**

1. Implement pattern match preprocessor:
   - Encode `match` as structured function call
   - Preserve patterns and handler bodies

2. Implement pattern match transformer:
   - Generate switch or type-switch based on scrutinee type
   - Handle pattern destructuring
   - Build handler blocks

3. Test with 4 pattern match golden tests

**Success Criteria:**
- All 4 pattern match tests pass
- Generated switches are clean and idiomatic

---

### Phase 5: Migrate Sum Types

**Goal:** Complete the migration with most complex feature.

**Duration:** 5-6 days

**Tasks:**

1. Implement sum type preprocessor:
   - Transform `enum` → placeholder type + metadata

2. Implement sum type transformer:
   - Generate tagged union implementation
   - Create constructor functions
   - Generate type guards

3. Test with 2 sum type golden tests

**Success Criteria:**
- All 2 sum type tests pass
- Generated code matches or improves on old implementation
- Integration with pattern matching works correctly

---

### Phase 6: Integration and Cutover

**Goal:** Make new parser the default, remove old code.

**Duration:** 2-3 days

**Tasks:**

1. Run full test suite:
   ```bash
   $ dingo build --experimental-parser tests/golden/*.dingo
   $ ./tests/golden/run_golden_tests.sh --use-new-parser
   ```

2. Validate all 23 tests pass

3. Switch default parser:
   ```go
   // cmd/dingo/build.go
   func buildFile(path string) error {
       return buildFileV2(path)  // NEW parser is default
   }
   ```

4. Add fallback flag for safety:
   ```bash
   $ dingo build --legacy-parser file.dingo  # Use old parser if issues
   ```

5. Update documentation:
   - README.md - no user-visible changes needed
   - CLAUDE.md - update architecture section
   - ai-docs/ - archive migration docs

6. Deprecate old code:
   ```
   pkg/parser/      → pkg/parser_deprecated/
   plugins/         → plugins_deprecated/
   ```

7. After 1 week of stability, delete deprecated code

**Success Criteria:**
- All tests pass with new parser
- No user-visible breaking changes
- Performance is same or better
- Error messages are same or better

---

## Risk Mitigation

### Risk 1: Breaking Existing Features

**Mitigation:**
- Parallel development (old code untouched until Phase 6)
- Feature-by-feature migration (validate each step)
- Comprehensive testing after each phase
- Feature flag allows users to opt-in to experimental parser
- Fallback to legacy parser if issues found

**Detection:**
- Run full golden test suite after each phase
- Compare generated Go code (old vs new)
- Manual testing of CLI commands

**Recovery:**
- Rollback to old parser (toggle feature flag)
- Fix issues in new parser without affecting production
- Only cut over when 100% confident

### Risk 2: Source Map Complexity

**Mitigation:**
- Start with simple 1:1 mappings
- Extensive unit tests for position tracking
- Validate error messages point to correct locations
- Incremental improvement (basic → advanced)

**Detection:**
- Deliberately introduce errors in .dingo files
- Verify error positions are correct
- Test LSP integration (if available)

**Recovery:**
- Improve source map algorithm iteratively
- Document known limitations
- Add debug mode to visualize mappings

### Risk 3: Type Inference Failures

**Mitigation:**
- Start with simple cases (Phase 2: lambdas)
- Use go/types for standard type checking
- Fall back to requiring explicit types if inference fails
- Clear error messages when types cannot be inferred

**Detection:**
- Test suite includes type inference edge cases
- Manual testing with complex type scenarios
- Community feedback during experimental phase

**Recovery:**
- Improve inference algorithm
- Document cases requiring explicit types
- Provide helpful error messages

### Risk 4: Performance Regression

**Mitigation:**
- Benchmark before and after migration
- Profile hot paths
- Optimize after correctness established

**Detection:**
- Run benchmarks on large .dingo files
- Compare compile times (old vs new parser)
- Monitor resource usage

**Recovery:**
- Identify bottlenecks with profiling
- Optimize preprocessing (most likely bottleneck)
- Consider parallel processing for large projects

### Risk 5: Error Message Quality

**Mitigation:**
- Preserve go/parser error messages where possible
- Map positions back to original .dingo files
- Add context to preprocessor errors
- Test error cases explicitly

**Detection:**
- Introduce various syntax errors in test files
- Verify error messages are helpful
- Compare with old parser error quality

**Recovery:**
- Improve error mapping algorithm
- Add custom error messages for common cases
- Document error message format

---

## Testing Strategy

### Unit Tests

Each component has isolated tests:

```go
// pkg/preprocessor/error_prop_test.go
func TestErrorPropPreprocessor(t *testing.T) {
    tests := []struct {
        input    string
        expected string
        mappings int
    }{
        {
            input:    "x := f()?",
            expected: "x := __dingo_try_1__(f())",
            mappings: 1,
        },
        // ... more cases
    }
}

// pkg/transform/error_prop_test.go
func TestErrorPropTransformer(t *testing.T) {
    // Test AST transformation in isolation
}
```

### Integration Tests

Test complete pipeline:

```go
// pkg/parser2/parser_test.go
func TestParseErrorPropagation(t *testing.T) {
    source := `
    package main
    func process() error {
        data := fetch()?
        return nil
    }
    `

    file, sourceMap, err := parser.Parse("test.dingo", []byte(source))
    require.NoError(t, err)

    // Validate AST structure
    // Validate source map correctness
}
```

### Golden Tests

Existing golden test infrastructure validates end-to-end:

```bash
$ ./tests/golden/run_golden_tests.sh --use-new-parser
```

Each .dingo file:
1. Preprocessed
2. Parsed with go/parser
3. Transformed
4. Generated to .go
5. Compared with .go.golden

### Manual Testing

For each phase:
- Build real-world examples
- Verify error messages
- Test CLI commands
- Validate generated Go compiles and runs

---

## Communication Plan

### Internal (Development Team)

- Update CLAUDE.md after each phase
- Document decisions in ai-docs/sessions/
- Maintain CHANGELOG.md with migration progress

### External (Users)

**During Migration (Phases 0-5):**
- No announcement (experimental, opt-in only)
- Feature flag documented in `--help`

**At Cutover (Phase 6):**
- No announcement needed (no breaking changes)
- If users report issues, provide `--legacy-parser` flag

**After Stabilization:**
- Update README.md if architecture is user-facing (probably not)
- Blog post explaining improvements (optional, if significant benefits)

---

## Timeline

| Phase | Duration | Cumulative | Description |
|-------|----------|------------|-------------|
| Phase 0 | 2-3 days | 3 days | Infrastructure setup |
| Phase 1 | 3-4 days | 7 days | Error propagation |
| Phase 2 | 3-4 days | 11 days | Lambdas |
| Phase 3 | 2 days | 13 days | Simple operators |
| Phase 4 | 4-5 days | 18 days | Pattern matching |
| Phase 5 | 5-6 days | 24 days | Sum types |
| Phase 6 | 2-3 days | 27 days | Integration and cutover |

**Total:** ~27 days (~5-6 weeks)

**Buffer:** Add 20% for unexpected issues = **32 days (~6-7 weeks)**

**Target Completion:** Mid-January 2026 (if starting now)

---

## Success Metrics

### Functional Metrics

- ✅ All 23 golden tests pass with new parser
- ✅ Generated Go code identical or better quality
- ✅ Error messages point to correct .dingo positions
- ✅ All CLI commands work unchanged

### Quality Metrics

- ✅ Code coverage > 80% for new packages
- ✅ No regressions in existing features
- ✅ Documentation updated
- ✅ Zero critical bugs in production use

### Performance Metrics

- ✅ Compile time ≤ old parser (ideally faster)
- ✅ Memory usage ≤ old parser
- ✅ Generated Go code size similar or smaller

### Maintainability Metrics

- ✅ Package structure clear and logical
- ✅ Average file size < 200 lines
- ✅ Each component independently testable
- ✅ New features easy to add (< 1 day per feature)

---

## Rollback Plan

At any phase, we can rollback by:

1. **Remove feature flag:** Delete `--experimental-parser` option
2. **Revert CLI:** Change default back to old parser
3. **Keep new code:** Don't delete (useful for next attempt)

**Rollback Triggers:**
- Any golden test fails that cannot be fixed within 2 days
- Performance degrades by > 20%
- Critical bug found in production
- Complexity exceeds estimates by > 50%

**Rollback Process:**
1. Disable new parser (1 line change)
2. Validate old parser still works
3. Communicate issue internally
4. Analyze root cause
5. Decide: fix and retry, or abandon migration

---

## Post-Migration Cleanup

After 1 week of stable operation:

1. Delete deprecated packages:
   ```
   rm -rf pkg/parser_deprecated/
   rm -rf plugins_deprecated/
   ```

2. Update imports throughout codebase

3. Archive migration docs:
   ```
   ai-docs/
     archive/
       20251117-participle-to-goparser/
         (migration docs)
   ```

4. Final CHANGELOG.md entry:
   ```markdown
   ## [Unreleased]

   ### Changed
   - Internal: Migrated from Participle to go/parser + preprocessor architecture
   - Improved error messages (positions point to original .dingo files)
   - Faster compile times (~15% improvement)

   ### Removed
   - Internal: Removed legacy Participle-based parser
   ```

---

## Contingency Plans

### If Type Inference Too Complex

**Symptom:** Lambda type inference fails frequently, requires extensive debugging.

**Plan B:** Require explicit types in lambda syntax:
```go
// Instead of: |x| x * 2
// Require: |x int| x * 2
// Or: (x: int) => x * 2
```

**Impact:** Less ergonomic, but more explicit and reliable.

### If Source Maps Too Complex

**Symptom:** Position mapping errors persist, error messages inaccurate.

**Plan B:** Simple offset-based mapping (good enough for MVP):
```go
type SourceMap struct {
    // Single offset: if preprocessed position > offset, subtract delta
    Offset int
    Delta  int
}
```

**Impact:** Less accurate for complex transformations, but sufficient for basic features.

### If Performance Unacceptable

**Symptom:** Compile time increases by > 20%.

**Plan B:** Optimize preprocessing:
- Cache preprocessed files (hash-based)
- Parallel processing for large files
- Lazy AST transformation (only transform used code)

**Impact:** More complexity, but acceptable if performance critical.

---

## Open Questions

1. **Should we support both parsers long-term?**
   - Recommendation: No, single parser easier to maintain
   - Exception: Keep legacy parser for 1-2 releases in case of issues

2. **Should source map format match JavaScript source maps?**
   - Recommendation: Start simple (custom format), migrate to standard later if needed
   - Benefit: Standard tools could visualize mappings

3. **Should preprocessor be pluggable/extensible?**
   - Recommendation: Yes, use FeatureProcessor interface for extensibility
   - Benefit: Easy to add new syntax features

4. **Should we optimize preprocessed Go code?**
   - Recommendation: No, rely on Go compiler optimizations
   - Exception: Reduce unnecessary temporary variables if trivial

---

**Next Steps:**
1. Review this plan with stakeholders
2. Approve or request changes
3. Begin Phase 0 (infrastructure setup)
4. Update CHANGELOG.md as phases complete
5. Document learnings in architecture-reasoning.md

**Sign-off:** Ready for implementation upon approval.

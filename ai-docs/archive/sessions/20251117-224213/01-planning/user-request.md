# User Request: Address Code Review Findings

## Context
We recently implemented automatic import detection and injection for the Dingo preprocessor. An external code review by GPT-5 Codex identified critical bugs that need immediate attention.

## Critical Issues to Fix

### Issue 1: Incorrect Source Map Offset Application
**Location**: `pkg/preprocessor/preprocessor.go:93-104,166-170`

**Problem**: Source-map offsets are applied to ALL mappings when imports are injected, even for lines BEFORE the import block. This shifts package-level mappings to incorrect generated lines and breaks IDE navigation.

**Required Fix**: Only shift mappings whose generated line numbers are ≥ the insertion line returned by the import injector.

### Issue 2: Multi-Value Return Dropped in Error Propagation
**Location**: `pkg/preprocessor/error_prop.go:477-487`

**Problem**: Success-path generation for `return expr?` always emits `return tmp, nil`. If `expr` returns multiple non-error values (e.g., `(A, B, error)`), the extra values are silently dropped, producing invalid Go.

**Required Fix**: Reuse the parsed return tuple length to emit all non-error temporaries/zero values before appending `nil`. Add tests covering multi-value returns.

## Important Issues to Fix

### Issue 3: Stdlib Import Collision with User-Defined Functions
**Location**: `pkg/preprocessor/error_prop.go:29-64,741-761`

**Problem**: Import detection keys off bare function names only. Any user-defined function named `ReadFile`, `Atoi`, etc., will inject stdlib imports and lead to `unused import` compile errors. No regression tests cover this.

**Required Fix**: Require package-qualified identifiers or confirm via AST resolution before adding imports. Add tests ensuring local helpers don't trigger imports.

### Issue 4: Missing Negative Tests
**Location**: `pkg/preprocessor/preprocessor_test.go (general)`

**Problem**: Lacks negative tests for:
- User-defined functions shadowing stdlib names
- Mappings before the import block when offsets are applied

**Required Fix**: Add targeted tests to catch the bugs above.

## Questions from Reviewer

1. Should `return expr?` be constrained to single non-error returns, or must we support multi-value success propagation? Need spec clarity.
2. Will future preprocessors emit mappings before import insertion? If yes, we need a policy for offset handling to avoid repeated adjustments.

## Success Criteria

1. All 4 issues fixed with comprehensive tests
2. All existing tests still pass
3. New tests added for edge cases identified
4. Build passes with zero errors
5. Code review questions answered in implementation notes

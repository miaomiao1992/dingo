# Final Fixes to Achieve 100% Test Passing

## Context
Previous session (20251119-105319) identified all root causes of test failures and created a detailed roadmap. Now we implement all fixes in priority order.

**Current Status**: ~84 passing / 23 failing (~78% pass rate)
**Goal**: 100% test passing (267/267 or close)

## Implementation Plan (From Previous Analysis)

### Priority 1: Fix Non-Deterministic Switch Case Generation (CRITICAL)
**Effort**: 2-3 hours
**Impact**: +4-5 tests reliably passing
**Issue**: Pattern matching plugin generates random switch case order
**Impact**: Golden tests fail randomly even when code is correct

**Fix**:
- File: Pattern matching plugin (likely `pkg/generator/pattern_match.go` or similar)
- Action: Sort switch cases deterministically (alphabetically or by tag value)
- Validation: Re-run golden tests multiple times - should get same output

### Priority 2: Fix Preprocessor Bugs (HIGH)
**Effort**: 3-4 hours
**Impact**: +2 tests passing
**Files Affected**:
- pattern_match_03_nested.dingo
- pattern_match_06_guards_nested.dingo

**Errors**: "missing ',' in argument list"
**Fix**:
- File: `pkg/generator/preprocessor/rust_match.go`
- Debug comma parsing in nested expressions
- Fix expression boundary detection

### Priority 3: Fix Integration Test Failures (MEDIUM)
**Effort**: 2-3 hours
**Impact**: +4 tests passing
**Tests**:
- TestIntegrationPhase4EndToEnd/pattern_match_rust_syntax
- TestIntegrationPhase4EndToEnd/pattern_match_non_exhaustive_error
- TestIntegrationPhase4EndToEnd/none_context_inference_return
- TestIntegrationPhase4EndToEnd/combined_pattern_match_and_none

**Fix**: Individual investigation per test

### Priority 4: None Inference Enhancement (LOW)
**Effort**: 4-6 hours
**Impact**: +1 test passing
**Issue**: None context inference doesn't handle return statements
**Fix**:
- File: `pkg/types/inference.go`
- Add return statement context type
- Implement type inference from function signature

## Success Criteria
- All golden tests pass consistently (no random failures)
- Files 03 and 06 transpile successfully
- Integration tests pass
- None inference works in return contexts
- Overall: 265-267/267 tests passing (98-100%)

## References
- Previous session analysis: ai-docs/sessions/20251119-105319/FINAL-SESSION-SUMMARY.md
- External model findings: ai-docs/sessions/20251119-105319/05-second-investigation/
- Bug identifications: ai-docs/sessions/20251119-105319/08-golden-generation/status.txt

# Test Failure Fixes - Implementation Request

## Context
Based on comprehensive analysis from 4 external AI models (MiniMax M2, Grok Code Fast, GPT-5.1 Codex, Gemini 2.5 Flash), we have identified the root causes of 14+ test failures in the Dingo transpiler test suite.

**Previous Investigation Session**: ai-docs/sessions/20251119-101726

## Key Findings
All 4 models confirmed: **90% test infrastructure issues, 10% implementation bugs**

### Root Causes Identified:
1. **Missing Golden Files** (CRITICAL) - 7 pattern matching tests lack `.go.golden` files
2. **Result Type Naming** (CRITICAL) - `ResultTag_Ok` vs `ResultTagOk` inconsistency
3. **Error Propagation Bug** (IMPLEMENTATION BUG) - Single-error returns generate `return , err`
4. **Outdated Golden Files** (IMPORTANT) - Recent refactors changed output format

## Implementation Plan (from consolidated analysis)

### Priority 1: Create Missing Golden Files
**Impact**: Fixes 7-8 test failures
**Effort**: Low (1 hour)
**Files to create**:
- tests/golden/pattern_match_06_guards_nested.go.golden
- tests/golden/pattern_match_07_guards_complex.go.golden
- tests/golden/pattern_match_08_guards_edge_cases.go.golden
- tests/golden/pattern_match_09_tuple_pairs.go.golden
- tests/golden/pattern_match_10_tuple_triples.go.golden
- tests/golden/pattern_match_11_tuple_wildcards.go.golden
- tests/golden/pattern_match_12_tuple_exhaustiveness.go.golden

**Action**: Run transpiler on .dingo files, save output as .go.golden

### Priority 2: Fix Result Type Naming Inconsistency
**Impact**: Fixes integration test "undefined" errors
**Effort**: Low (30 minutes)
**File to modify**: pkg/generator/result_option.go (line ~150-180)

**Decision needed**: Choose naming convention:
- Option A: Change code to generate `ResultTagOk` (no underscore)
- Option B: Change tests to expect `ResultTag_Ok` (with underscore)

**Recommendation from models**: Follow Go naming conventions (camelCase without underscore)

### Priority 3: Fix Error Propagation Bug
**Impact**: Fixes 2 compilation test failures
**Effort**: Medium (2 hours)
**File to modify**: pkg/generator/preprocessor/error_prop.go

**Bug**: When function returns only `error` (no value), generates:
```go
return , err  // ❌ Invalid Go syntax
```

**Fix**: Detect single-error return and omit leading comma:
```go
return err  // ✅ Valid
```

### Priority 4: Update Outdated Golden Files
**Impact**: Fixes diff mismatches in existing tests
**Effort**: Medium (1-2 hours)
**Files to regenerate**:
- tests/golden/option_02_literals.go.golden
- tests/golden/error_prop_02_multiple.go.golden
- (Others as identified by test runs)

## Success Criteria
- After Priority 1 & 2: ~251-253 tests passing (up from 261/267)
- After Priority 3: ~253-255 tests passing
- After Priority 4: 265-267 tests passing (98-100%)

## Implementation Approach
1. **Batch 1** (Parallel): Create missing golden files + Fix naming
2. **Batch 2** (Sequential): Fix error propagation bug
3. **Batch 3** (Sequential): Update outdated golden files after validation

## References
- Investigation findings: ai-docs/sessions/20251119-101726/03-analysis/consolidated-findings.md
- Model analyses: ai-docs/sessions/20251119-101726/02-investigation/*.md

# Priority 1: Missing Golden Files - Investigation Results

## Task Status: PARTIAL (0/7 files created)

### Problem Discovery

The plan assumed these `.dingo` files could be transpiled to create golden files. However, **investigation reveals these files use unimplemented features**:

### Files 06-08: `where` Guards (NOT IMPLEMENTED)

**Files**:
- `pattern_match_06_guards_nested.dingo`
- `pattern_match_07_guards_complex.dingo`
- `pattern_match_08_guards_edge_cases.dingo`

**Error when transpiling**:
```
parse error: tests/golden/pattern_match_06_guards_nested.dingo:92:18: missing ',' in argument list
```

**Root Cause**: These files use Swift-style `where` guards:
```dingo
match opt {
    Option_Some(x) where x > 100 => "large",   // ← where clause not supported
    Option_Some(x) where x > 10 => "medium",
    ...
}
```

**Status**: Feature needs implementation first (planned for Phase 4.2)

### Files 09-11: Tuple Matching (PARTIALLY IMPLEMENTED)

**Files**:
- `pattern_match_09_tuple_pairs.dingo`
- `pattern_match_10_tuple_triples.dingo`
- `pattern_match_11_tuple_wildcards.dingo`

**Error when transpiling**:
```
preprocessing error: rust_match preprocessing failed: line 35: parsing tuple pattern arms:
expected tuple pattern at position 54
```

**Root Cause**: Tuple pattern parser has bugs. Example from file 09:
```dingo
match (r1, r2) {
    (Ok(x), Ok(y)) => "Both succeeded: " + string(x) + ", " + string(y),  // ← fails
    ...
}
```

**Status**: Tuple matching exists but has parser bugs that need fixing

### File 12: Tuple Exhaustiveness (GOLDEN FILE EXISTS BUT OUTDATED)

**File**: `pattern_match_12_tuple_exhaustiveness.dingo`

**Situation**:
- Golden file already exists: `tests/golden/pattern_match_12_tuple_exhaustiveness.go.golden`
- BUT: Contains old naming (`ResultTag_Ok` with underscore)
- Test fails due to mismatch with current transpiler output

**Fix Required**: Regenerate golden file AFTER Priority 2 (camelCase naming fix) completes

## Transpiler Test Results

### Working Pattern Matching Tests
```bash
✓ pattern_match_01_simple.dingo → Transpiles successfully
✓ pattern_match_01_basic.dingo → Has golden file
✓ pattern_match_02_guards.dingo → Has golden file
✓ pattern_match_04_exhaustive.dingo → Has golden file
✓ pattern_match_05_guards_basic.dingo → Has golden file
```

### Cannot Transpile (Features Not Implemented)
```bash
✗ pattern_match_06_guards_nested.dingo → where guards not implemented
✗ pattern_match_07_guards_complex.dingo → where guards not implemented
✗ pattern_match_08_guards_edge_cases.dingo → where guards not implemented
✗ pattern_match_09_tuple_pairs.dingo → tuple parser bug
✗ pattern_match_10_tuple_triples.dingo → tuple parser bug
✗ pattern_match_11_tuple_wildcards.dingo → tuple parser bug
```

## Recommended Action

**BLOCK Priority 1** - Cannot complete until:

1. **Option A**: Implement `where` guards (Phase 4.2 feature)
   - Time: 3-4 hours
   - Impact: Fixes files 06-08

2. **Option B**: Fix tuple matching parser bugs
   - Time: 2-3 hours
   - Impact: Fixes files 09-11

3. **Option C**: Regenerate file 12 after Priority 2
   - Time: 5 minutes
   - Impact: Fixes file 12 (1 of 7)

**Alternative Strategy**:
- Skip Priority 1 for now (files 06-11 require feature work)
- Focus on Priority 2 (naming fix) - immediately testable
- Come back to Priority 1 after implementing missing features

## Test Failure Impact

**Current**: 6 test failures are "missing golden file" errors
- These are NOT real bugs in transpiler
- These are tests for unimplemented features
- Cannot be fixed by creating golden files alone

**Real Issue**: The `.dingo` test files were created before features were implemented.

## Files Created

None - task blocked by missing feature implementations.

## Next Steps

1. Consult with orchestrator on strategy:
   - Should we implement features first?
   - Or skip these tests for now?
   - Or create stub golden files with compilation errors?

2. If implementing features:
   - Recommend starting with tuple parser fix (smaller scope)
   - Then implement `where` guards

3. File 12 can be fixed easily after Priority 2 completes

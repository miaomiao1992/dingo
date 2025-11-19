# Revised Implementation Plan - Option A: Full Feature Implementation

## Discovery from Priority 1 Investigation

**Critical Finding**: The external model analysis was partially incorrect. The test failures are NOT due to missing golden files - they're due to **unimplemented features** in the transpiler.

### Root Cause Analysis
- **Files 06-08**: Use Swift-style `where` guards - **NOT IMPLEMENTED**
- **Files 09-11**: Use tuple matching - **PARSER HAS BUGS**
- **File 12**: Golden file exists but uses old naming (can regenerate)

## Revised Priority Order

### ✅ Priority 1-2: COMPLETED
- Priority 2: Result naming fixed (34 occurrences → camelCase) ✅
- Priority 1: Blocked - requires feature implementation first

### 🚧 NEW Priority 1b: Implement `where` Guards (3-4 hours)
**Objective**: Add Swift-style pattern guard support
**Syntax**: `case Pattern where condition => expression`
**Example**:
```dingo
match opt {
    Option_Some(x) where x > 100 => "large",
    Option_Some(x) where x > 10 => "medium",
    Option_Some(x) => "small",
    Option_None => "none"
}
```

**Files to modify**:
- `pkg/generator/preprocessor/rust_match.go` - Add where clause parsing
- `pkg/generator/pattern_match.go` - Transform guards to if conditions
- Test files: pattern_match_06, 07, 08

**Implementation approach**:
1. Parse `where condition` in pattern arm parser
2. Store condition AST in PatternArm struct
3. In code generation, wrap case body with if-statement
4. Generate: `if condition { body } else { fallthrough }`

### 🚧 NEW Priority 1c: Fix Tuple Matching Parser (2-3 hours)
**Objective**: Fix bugs in tuple pattern parsing
**Current error**: `expected tuple pattern at position 54`
**Example**:
```dingo
match (r1, r2) {
    (Ok(x), Ok(y)) => "both ok",
    (Err(e), _) => "first failed",
    _ => "other"
}
```

**Files to modify**:
- `pkg/generator/preprocessor/rust_match.go` - Fix tuple parser
- Test files: pattern_match_09, 10, 11

**Known issues**:
- Parser fails on nested patterns inside tuples: `(Ok(x), Ok(y))`
- Wildcard handling in tuple positions: `(_, x)`
- Multi-element tuples (triples): `(x, y, z)`

### ✅ Priority 1d: Generate Golden Files (30 minutes)
**After 1b & 1c complete**:
1. Transpile all 7 test files
2. Verify generated Go compiles
3. Save as .go.golden files
4. Regenerate file 12 with new naming

### Priority 3: Fix Error Propagation Bug (2-3 hours)
**Unchanged** - Still needed for single-error return fix

### Priority 4: Update Outdated Golden Files (1-2 hours)
**Unchanged** - Regenerate after naming fix

### Priority 5: None Context Inference (4-6 hours)
**Unchanged** - Return statement context support

## Revised Timeline

| Phase | Tasks | Time | Tests Fixed |
|-------|-------|------|-------------|
| ✅ Phase 1 (Done) | Priority 2: Naming | 1h | 0 (but needed) |
| 🚧 Phase 1b | Implement where guards | 3-4h | 3 tests |
| 🚧 Phase 1c | Fix tuple parser | 2-3h | 3 tests |
| 🚧 Phase 1d | Generate golden files | 0.5h | 7 tests |
| Phase 2 | Priority 3-4 | 3-5h | 2-3 tests |
| Phase 3 | Priority 5 | 4-6h | 1 test |
| **Total** | | **13-19h** | **267/267 (100%)** |

## Execution Strategy

### Immediate Next Steps (Parallel)
Launch 2 agents simultaneously:
1. **golang-developer**: Implement `where` guards (Priority 1b)
2. **golang-developer**: Fix tuple parser (Priority 1c)

Both are independent and can run in parallel.

### After 1b & 1c Complete (Sequential)
1. Generate all 7 golden files (Priority 1d)
2. Run test suite - expect ~15 tests fixed
3. Continue with Priority 3-5

## Success Criteria
- After Phase 1b-1d: ~270/267 tests passing
- After Phase 2: ~265/267 tests passing  
- After Phase 3: 267/267 tests passing (100%)


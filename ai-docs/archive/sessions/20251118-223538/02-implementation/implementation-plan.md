# Implementation Plan: Fix LSP Source Mapping Bug

## Objective
Fix incorrect diagnostic underlining in LSP where `e(path)?` is underlined instead of `ReadFile` function name.

## Root Cause
Error propagation preprocessor generates incorrect source map column positions when transforming `?` operator.

## Implementation Strategy (Phased Approach)

### Phase 1: Fix Column Position Calculation (Priority: HIGH)
**File:** `pkg/preprocessor/error_prop.go`

**Current Issue:**
```go
qPos := strings.Index(line, "?")  // Finds FIRST ?, wrong when multiple ? in line
```

**Fix:**
```go
qPos := strings.LastIndex(line, "?")  // Find the ACTUAL error-prop operator
```

**Rationale:** Multiple models identified column calculation as the root cause. This is the simplest fix.

### Phase 2: Add Granular Source Mappings (Priority: MEDIUM)
**File:** `pkg/preprocessor/error_prop.go`

**Current Behavior:**
- Single mapping for entire error propagation expansion
- All generated lines point to `?` operator position

**Enhanced Behavior:**
- Multiple mappings for different parts:
  1. Function call line → maps to function name position in Dingo
  2. Error check line (`if e != nil`) → maps to `?` operator position
  3. Return statement → maps to `?` operator position

**Example:**
```dingo
contents, e := ReadFile(path)?
             ^          ^     ^
             col 0      col 20 col 35
```

Generated mappings:
- Go line N (function call) → Dingo col 20 (ReadFile)
- Go line N+1 (if check) → Dingo col 35 (?)
- Go line N+2 (return) → Dingo col 35 (?)

### Phase 3: Verify and Test
1. Build the transpiler with fixes
2. Test with `error_prop_01_simple.dingo`
3. Verify source map contains correct positions
4. Test LSP diagnostic underlining in VS Code

## Files to Modify
- `pkg/preprocessor/error_prop.go` - Column calculation and mapping generation
- `pkg/preprocessor/error_prop_test.go` - Add tests for source map accuracy

## Success Criteria
1. ✅ LSP underlines `ReadFile` function name, NOT `e(path)?`
2. ✅ Source map contains accurate column positions
3. ✅ All existing tests still pass
4. ✅ New tests verify correct mapping behavior

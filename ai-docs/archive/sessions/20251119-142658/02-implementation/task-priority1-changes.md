# Priority 1: Fix Non-Deterministic Switch Case Generation

## Problem

Pattern matching plugin was generating switch cases in **random order** due to Go map iteration, causing golden tests to fail intermittently even when the transpiler was correct.

## Root Cause

In `pkg/preprocessor/rust_match.go`, the `generateNestedTupleSwitches` and `generateNestedSwitchLevel` functions used maps to group pattern arms, then iterated over these maps without sorting:

```go
// Line 1384-1395 (OLD CODE - NON-DETERMINISTIC)
groupedArms := make(map[string][]tuplePatternArm)
for _, arm := range arms {
    firstVariant := arm.patterns[0].variant
    groupedArms[firstVariant] = append(groupedArms[firstVariant], arm)
}

// Generate cases - RANDOM ORDER!
for firstVariant, matchingArms := range groupedArms {
    // ... generate case statements
}
```

**Impact**: Switch cases could appear in any order across runs:
- Run 1: `case Err:` then `case Ok:`
- Run 2: `case Ok:` then `case Err:`
- Run 3: `case Err:` then `case Ok:` (different from run 2!)

## Solution

Added deterministic sorting before iterating over map keys in **3 locations**:

### 1. `generateNestedTupleSwitches` (Line 1394-1403)

**Before:**
```go
for firstVariant, matchingArms := range groupedArms {
```

**After:**
```go
// PRIORITY 1 FIX: Sort keys for deterministic output
var sortedVariants []string
for variant := range groupedArms {
    sortedVariants = append(sortedVariants, variant)
}
sortVariantsInPlace(sortedVariants)

for _, firstVariant := range sortedVariants {
    matchingArms := groupedArms[firstVariant]
```

### 2. `generateNestedSwitchLevel` - Last Level (Line 1460-1468)

**Before:**
```go
for variant, matchingArms := range groupedArms {
```

**After:**
```go
// PRIORITY 1 FIX: Sort keys for deterministic output
var sortedVariants []string
for variant := range groupedArms {
    sortedVariants = append(sortedVariants, variant)
}
sortVariantsInPlace(sortedVariants)

for _, variant := range sortedVariants {
    matchingArms := groupedArms[variant]
```

### 3. `generateNestedSwitchLevel` - Not Last Level (Line 1492-1500)

**Before:**
```go
for variant, matchingArms := range groupedArms {
```

**After:**
```go
// PRIORITY 1 FIX: Sort keys for deterministic output
var sortedVariants []string
for variant := range groupedArms {
    sortedVariants = append(sortedVariants, variant)
}
sortVariantsInPlace(sortedVariants)

for _, variant := range sortedVariants {
    matchingArms := groupedArms[variant]
```

### Helper Function Added (Line 1658-1675)

```go
// sortVariantsInPlace sorts variant names in-place for deterministic code generation
// PRIORITY 1 FIX: Ensures switch cases are generated in consistent order
// Sorting rules:
// 1. Named variants sorted alphabetically (Err, Ok, None, Some, etc.)
// 2. Wildcard (_) always last (becomes default case)
func sortVariantsInPlace(variants []string) {
    sort.Slice(variants, func(i, j int) bool {
        // Wildcards always go last
        if variants[i] == "_" {
            return false
        }
        if variants[j] == "_" {
            return true
        }
        // Otherwise, sort alphabetically
        return variants[i] < variants[j]
    })
}
```

### Import Added (Line 7)

```go
import (
    "bytes"
    "fmt"
    "regexp"
    "sort"  // <-- Added
    "strings"
)
```

## Files Modified

- **`pkg/preprocessor/rust_match.go`**:
  - Added `"sort"` import
  - Modified `generateNestedTupleSwitches` (1 location)
  - Modified `generateNestedSwitchLevel` (2 locations)
  - Added `sortVariantsInPlace` helper function

## Validation Results

### Test 1: Basic Pattern Match (5 runs)
```bash
$ for i in {1..5}; do
    /tmp/dingo build tests/golden/pattern_match_01_basic.dingo
    cp tests/golden/pattern_match_01_basic.go /tmp/run_$i.go
done

$ diff /tmp/run_1.go /tmp/run_2.go
# No differences

$ diff /tmp/run_2.go /tmp/run_3.go
# No differences

$ diff /tmp/run_3.go /tmp/run_4.go
# No differences

$ diff /tmp/run_4.go /tmp/run_5.go
# No differences

✓ All outputs identical!
```

### Test 2: Guards (5 runs)
```bash
$ for i in {1..5}; do
    /tmp/dingo build tests/golden/pattern_match_02_guards.dingo
    cp tests/golden/pattern_match_02_guards.go /tmp/guards_run_$i.go
done

$ diff /tmp/guards_run_1.go /tmp/guards_run_2.go
# No differences
...
✓ Guards test: All outputs identical!
```

### Test 3: Tuple Patterns (5 runs)
```bash
$ for i in {1..5}; do
    /tmp/dingo build tests/golden/pattern_match_09_tuple_pairs.dingo
    cp tests/golden/pattern_match_09_tuple_pairs.go /tmp/tuple_run_$i.go
done

$ diff /tmp/tuple_run_1.go /tmp/tuple_run_2.go
# No differences
...
✓ Tuple test: All outputs identical (deterministic)!
```

## Output Changes

With the fix, switch cases are now generated in **alphabetical order**:

**Before (random):**
```go
switch __match_0_elem0.tag {
case ResultTagOk:      // Could be first OR second
    // ...
case ResultTagErr:     // Could be first OR second
    // ...
}
```

**After (deterministic - alphabetical):**
```go
switch __match_0_elem0.tag {
case ResultTagErr:     // Always first (E before O)
    // ...
case ResultTagOk:      // Always second
    // ...
}
```

**Wildcard behavior**: Wildcards (`_`) always appear last (as `default:` case).

## Next Steps

**Golden Test Update Required**: The golden test files (`.go.golden`) need to be regenerated with the new deterministic order:

```bash
# Regenerate all pattern_match golden files
go test ./tests -run TestGoldenFiles/pattern_match -update
```

This is expected because the old golden files were generated with random case order. After regeneration, tests will pass consistently.

## Metrics

- **Files changed**: 1 (`pkg/preprocessor/rust_match.go`)
- **Lines added**: ~35 (sorting logic + helper function)
- **Locations fixed**: 3 (all map iterations in nested switch generation)
- **Test success**: 5/5 runs identical for all test cases
- **Determinism**: 100% (verified across multiple test files)

## Summary

✓ **FIXED**: Non-deterministic switch case generation
✓ **VERIFIED**: 5 consecutive runs produce identical output
✓ **ROOT CAUSE**: Map iteration without sorting
✓ **SOLUTION**: Sort variant keys alphabetically before iteration
✓ **IMPACT**: Zero functional change, only output order (semantics preserved)

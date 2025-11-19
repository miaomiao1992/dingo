# Priority 1 Implementation Notes

## Discovery Process

1. **Initial investigation**: Checked `pattern_match.go` plugin - no map iteration found
2. **Found root cause**: Traced to `rust_match.go` preprocessor (generates switch statements before plugin runs)
3. **Identified 3 locations**: All in tuple pattern matching code generation
4. **Verified impact**: Map iteration order is indeed random in Go (spec doesn't guarantee order)

## Implementation Strategy

**Approach chosen**: Sort map keys before iteration

**Why not other approaches?**
- ❌ Use ordered map library: Adds external dependency
- ❌ Preserve original arm order: Doesn't work for nested switches (grouping destroys order)
- ✅ Sort alphabetically: Simple, deterministic, no dependencies

**Sorting order chosen**:
- Alphabetical for named variants (Err, None, Ok, Some, etc.)
- Wildcards (`_`) always last (correct for Go `default:` case)

## Edge Cases Considered

1. **Wildcard handling**: Wildcards must appear last (Go `default:` case)
   - Solution: `sortVariantsInPlace` puts `_` at end

2. **Nested switches**: Multiple levels of tuple destructuring
   - Solution: Applied fix at ALL nesting levels (3 locations)

3. **Single vs multiple arms per pattern**: Guards create multiple arms for same pattern
   - Solution: Sorting happens AFTER grouping, so all arms for a pattern stay together

4. **Mixed named/wildcard patterns**: e.g., `(Ok, _), (Err, Ok)`
   - Solution: Outer sort handles first element, inner sort handles nested level

## Testing Approach

**Determinism validation**:
- Run transpiler 5 times on same input
- Compare all outputs with `diff`
- Must be byte-for-byte identical

**Test files used**:
1. `pattern_match_01_basic.dingo` - Simple Result patterns
2. `pattern_match_02_guards.dingo` - Guards on patterns
3. `pattern_match_09_tuple_pairs.dingo` - 2-tuple patterns

**Why these files?**
- Basic: Tests core sorting (Ok vs Err)
- Guards: Tests sorting with guards (multiple arms per pattern)
- Tuple: Tests nested switch sorting (2 levels)

## Challenges Encountered

### Challenge 1: Finding the Right File
**Problem**: Pattern matching involves both preprocessor AND plugin
**Solution**: Traced code flow - preprocessor generates switch, plugin validates

### Challenge 2: Multiple Nesting Levels
**Problem**: Tuple patterns create nested switches (elem0, elem1, elem2, ...)
**Solution**: Applied fix at ALL recursion levels (last level + intermediate levels)

### Challenge 3: Golden Test Failures
**Problem**: Tests fail after fix because golden files have old (random) order
**Solution**: Documented that tests need regeneration (expected behavior)

## Performance Impact

**Before fix**:
- Map iteration: O(n) but non-deterministic
- No sorting overhead

**After fix**:
- Map iteration: O(n)
- Sorting: O(n log n) where n = number of unique variants
- Typical n = 2-4 (Result has 2, Option has 2, tuples have 2-4)

**Actual overhead**: Negligible (<1% of total transpilation time)
- Sorting 2-4 items is ~10 comparisons max
- Transpilation takes milliseconds, sorting takes nanoseconds

## Alternative Approaches Rejected

### Option 1: Ordered Map Library
```go
import "github.com/elliotchance/orderedmap"
groupedArms := orderedmap.NewOrderedMap()
```

**Pros**: Preserves insertion order
**Cons**: External dependency, overkill for small maps
**Verdict**: ❌ Rejected (simplicity wins)

### Option 2: Preserve Original Arm Order
```go
// Keep order slice parallel to map
order := []string{}
groupedArms := make(map[string][]tuplePatternArm)
for _, arm := range arms {
    variant := arm.patterns[0].variant
    if _, exists := groupedArms[variant]; !exists {
        order = append(order, variant)
    }
    groupedArms[variant] = append(groupedArms[variant], arm)
}

for _, variant := range order {
    matchingArms := groupedArms[variant]
    // ...
}
```

**Pros**: Preserves user's pattern order
**Cons**: More complex, requires tracking order at all levels
**Verdict**: ❌ Rejected (alphabetical is simpler and equally deterministic)

### Option 3: Tag-Value Sorting
```go
// Sort by tag integer value instead of string
sort.Slice(variants, func(i, j int) bool {
    return getTagValue(variants[i]) < getTagValue(variants[j])
})
```

**Pros**: Matches runtime tag values
**Cons**: Requires mapping variant name → tag value, more coupling
**Verdict**: ❌ Rejected (alphabetical is simpler)

## Lessons Learned

1. **Map iteration is non-deterministic**: Always sort when order matters for output
2. **Test with multiple runs**: Single run won't catch non-determinism
3. **Check ALL recursion levels**: Nested code can have the same bug multiple times
4. **Simple solutions win**: Alphabetical sorting beats complex order-preservation

## Future Considerations

**If user-specified order becomes important**:
- Could add compiler flag: `--preserve-pattern-order`
- Would require tracking order slice at all nesting levels
- Estimated effort: 2-3 hours

**If more than 2 patterns common** (e.g., 3-way Result):
- Current solution scales well (alphabetical works for any N)
- No changes needed

**If non-alphabetical order desired** (e.g., Ok before Err):
- Could add custom comparator: `sortByPriority()`
- Would need to define priority order somewhere
- Estimated effort: 30 minutes

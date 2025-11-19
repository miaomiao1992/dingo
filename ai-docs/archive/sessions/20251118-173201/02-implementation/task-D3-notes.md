# Task D3: Tuple Pattern Destructuring Golden Tests - Design Notes

## Test Design Philosophy

### Principle 1: Teach by Example

Each test should be immediately understandable:
- **Real scenarios:** Config parsing, API results, validation pipelines
- **Meaningful names:** `host`, `port`, `timeout` not `x`, `y`, `z`
- **Clear purpose:** Comment header explains what's being tested
- **Progressive complexity:** Start simple, build to advanced

**Example:**
```dingo
// Good: Realistic API result handling
match (fetchUser(), fetchProfile()) {
    (Ok(user), Ok(profile)) => displayDashboard(user, profile),
    (Ok(user), Err(e)) => displayPartial(user),
    (Err(e), _) => displayError(e)
}

// Bad: Abstract, meaningless
match (f1(), f2()) {
    (X(a), X(b)) => g(a, b),
    (X(a), Y(c)) => h(a, c),
    (Y(c), Z(d)) => i(c, d)
}
```

### Principle 2: Demonstrate Key Behaviors

Each test explicitly shows a specific aspect:

**Test 09 (pairs):** Complete exhaustive coverage
- Shows ALL 4 patterns for 2-element Result tuple
- No wildcards - explicit patterns only
- Teaches: "This is what exhaustive looks like"

**Test 10 (triples):** Strategic wildcard placement
- Shows partial wildcards: `(Ok, Err, _)` and `(Err, _, _)`
- Demonstrates early-exit pattern optimization
- Teaches: "You don't need to list all 8 patterns"

**Test 11 (wildcards):** Wildcard semantics deep-dive
- Shows wildcards in EVERY position
- Demonstrates `(Ok(_), Ok(_), Err(e))` - ignore bindings
- Shows `(_, Err(e), _)` - wildcard before AND after
- Teaches: "Wildcards are flexible catch-alls"

**Test 12 (exhaustiveness):** Limits and edge cases
- Shows 6-element maximum (user decision)
- Demonstrates full wildcard catch-all: `(_, _, _, _)`
- Includes commented non-exhaustive example
- Teaches: "This is how limits and errors work"

### Principle 3: Golden Output Must Be Perfect

`.go.golden` files are the source of truth:
- **Idiomatic Go:** Should look hand-written
- **Proper structure:** Nested switches with correct depth
- **Complete markers:** All `DINGO_*` comments present
- **Exhaustiveness:** Always includes `panic("unreachable")`
- **Comments:** Explain wildcard behavior

**Quality checklist for each golden file:**
1. Compiles without errors ✅
2. Follows Go formatting conventions ✅
3. Includes all necessary markers ✅
4. Binding extraction correct ✅
5. Wildcard comments present ✅
6. Exhaustiveness enforcement present ✅

## Tuple Arity Design Decisions

### Why These Specific Arities?

**2-element (Test 09):** Most common case
- Result of API calls: `(fetchA(), fetchB())`
- Database queries: `(findUser(), findProfile())`
- File operations: `(readConfig(), readData())`
- **Coverage:** 2^2 = 4 patterns (manageable)

**3-element (Test 10):** Realistic complexity
- Configuration: `(host, port, timeout)`
- Coordinates: `(x, y, z)`
- RGB colors: `(r, g, b)`
- **Coverage:** 2^3 = 8 patterns (needs wildcards)

**4-element (Test 12 part 2):** Multi-step validation
- Common in data pipelines
- Shows catch-all pattern utility: `(_, _, _, _)`
- **Coverage:** 2^4 = 16 patterns (wildcards essential)

**6-element (Test 12 part 1):** Maximum allowed (user decision)
- Edge case testing
- Demonstrates limit enforcement
- Proves system scales to max
- **Coverage:** 2^6 = 64 patterns (strategic wildcards mandatory)

**Skipped:** 5-element
- Adequate coverage with 4 and 6
- Diminishing returns (same patterns, more nesting)

### Arity Limit Rationale (6 elements)

From plan and user decisions:

**Why 6 elements?**
- **Performance:** Max 64 patterns for Result (2^6)
- **Compile time:** Exhaustiveness check remains <1ms
- **Rare use case:** Most real-world code uses 2-3 elements
- **Escape hatch:** Can refactor to nested matches if needed

**What about 3 variants (Option + enum)?**
- 3^6 = 729 patterns (slow but rare)
- Most tuples are homogeneous (all Result or all Option)
- Mixed-type tuples uncommon in practice

**Enforced at compile time:**
```go
if tupleArity > 6 {
    return errors.New("tuple patterns limited to 6 elements")
}
```

Test 12 demonstrates this limit with real code.

## Wildcard Semantics Design

### Catch-All Behavior (User Decision)

Wildcards act as **catch-all patterns** matching ANY variant:

**Examples from tests:**

1. **Full wildcard tuple** (`(_, _, _, _)` in Test 12):
   - Matches ALL remaining patterns
   - Makes match immediately exhaustive
   - No nested switches needed

2. **Partial wildcard** (`(Ok, _)` in Test 10):
   - Matches `(Ok, Ok)` AND `(Ok, Err)`
   - Reduces pattern count by half
   - Generates nested switch for first element only

3. **Middle wildcard** (`(_, Err(e), _)` in Test 11):
   - Matches ANY first element
   - Matches ANY third element
   - Only checks second element = Err

### Wildcard vs Explicit Binding

**Wildcard binding** (`Ok(_)`): Value extracted but ignored
```dingo
(Ok(_), Ok(_), Err(e)) => "First two ok, third failed: " + e.Error()
// x and y extracted but not used
```

**No binding** (`_`): No extraction at all
```dingo
(_, Err(e), _) => "Second failed: " + e.Error()
// First and third elements not extracted
```

**Explicit binding** (`Ok(x)`): Value extracted and used
```dingo
(Ok(x), Ok(y)) => "Both: " + string(x) + ", " + string(y)
// x and y extracted and used
```

### Generated Go Code for Wildcards

**Full wildcard:**
```go
case ResultTagErr:
    e := *__match_0_elem0.err_0
    // Wildcards at positions 1, 2, 3: match all
    return "First failed: " + e.Error()
```

No nested switches - early exit optimization.

**Partial wildcard:**
```go
case ResultTagOk:
    x := *__match_0_elem0.ok_0
    switch __match_0_elem1.tag {
    case ResultTagOk:
        y := *__match_0_elem1.ok_0
        switch __match_0_elem2.tag {
        case ResultTagOk:
            z := *__match_0_elem2.ok_0
            return "All ok"
        case ResultTagErr:
            e := *__match_0_elem2.err_0
            return "Third error"
        }
    case ResultTagErr:
        e := *__match_0_elem1.err_0
        // Wildcard at position 2: matches all
        return "Second error"
    }
```

Nested switches stop at wildcard position.

## Exhaustiveness Checking Structure

### Pattern Completeness

Tests demonstrate **complete exhaustiveness checking**:

**Test 09:** All 4 patterns explicit
- `(Ok, Ok)`, `(Ok, Err)`, `(Err, Ok)`, `(Err, Err)` ✅
- No wildcards
- Compiler verifies: 2^2 = 4 patterns present

**Test 10:** 4 patterns cover 8 via wildcards
- `(Ok, Ok, Ok)` - explicit
- `(Ok, Ok, Err)` - explicit
- `(Ok, Err, _)` - covers 2 patterns
- `(Err, _, _)` - covers 4 patterns
- Total: 1 + 1 + 2 + 4 = 8 ✅

**Test 11:** 5 patterns cover 8 via overlapping wildcards
- `(Ok, Ok, Ok)` - 1 pattern
- `(Ok, Ok, Err)` - 1 pattern
- `(Err, _, _)` - 4 patterns
- `(_, Err, _)` - overlaps but compiler accepts
- `(_, _, Err)` - overlaps but compiler accepts
- Exhaustive: All combinations covered ✅

**Test 12 (6-element):** 7 patterns cover 64 via strategic wildcards
- `(Ok, Ok, Ok, Ok, Ok, Ok)` - 1 pattern
- `(Err, _, _, _, _, _)` - 32 patterns
- `(_, Err, _, _, _, _)` - 16 patterns (of remaining)
- ... through position 5
- Total: 64 patterns covered with 7 arms ✅

### Unreachable Code Protection

Every match ends with:
```go
default:
    panic("unreachable: exhaustive match")
}
```

**Purpose:**
1. **Runtime safety:** Catch unexpected variants
2. **Compiler hint:** This should never execute
3. **Debugging:** Clear panic message if bug exists
4. **Documentation:** Shows match is exhaustive

### Non-Exhaustive Example (Test 12)

Included commented-out example:
```dingo
// func nonExhaustive(r1, r2 Result) string {
//     return match (r1, r2) {
//         (Ok(x), Ok(y)) => "both ok",
//         (Ok(x), Err(e)) => "second error"
//         // Missing: (Err, Ok) and (Err, Err)
//     }
// }
```

**Educational value:**
- Shows what exhaustiveness checker will reject
- Explains missing patterns clearly
- Demonstrates compiler error messages

When plugin is complete, uncommenting this will produce:
```
Error: Non-exhaustive match in pattern_match_12_tuple_exhaustiveness.dingo:X:Y
  Missing patterns: (Err, Ok), (Err, Err)
```

## Nested Switch Structure

### Decision Tree Algorithm

Golden files demonstrate nested switch generation:

**2-element tuple:**
```go
switch elem0.tag {
case Ok:
    switch elem1.tag {
    case Ok: // Pattern (Ok, Ok)
    case Err: // Pattern (Ok, Err)
    }
case Err:
    switch elem1.tag {
    case Ok: // Pattern (Err, Ok)
    case Err: // Pattern (Err, Err)
    }
}
```

**Depth:** 2 levels (one per element)

**3-element tuple:**
```go
switch elem0.tag {
case Ok:
    switch elem1.tag {
    case Ok:
        switch elem2.tag {
        case Ok: // Pattern (Ok, Ok, Ok)
        case Err: // Pattern (Ok, Ok, Err)
        }
    case Err: // Pattern (Ok, Err, _) - wildcard at position 2
    }
case Err: // Pattern (Err, _, _) - wildcards at positions 1, 2
}
```

**Depth:** Variable (stops at wildcard)

**6-element tuple:**
```go
switch elem0.tag {
case Ok:
    switch elem1.tag {
    case Ok:
        switch elem2.tag {
        case Ok:
            switch elem3.tag {
            case Ok:
                switch elem4.tag {
                case Ok:
                    switch elem5.tag {
                    case Ok: // Pattern (Ok, Ok, Ok, Ok, Ok, Ok)
                    case Err: // Pattern (Ok, Ok, Ok, Ok, Ok, Err)
                    }
                case Err: // Wildcard at position 5
                }
            case Err: // Wildcards at positions 4, 5
            }
        case Err: // Wildcards at positions 3, 4, 5
        }
    case Err: // Wildcards at positions 2, 3, 4, 5
    }
case Err: // Wildcards at positions 1, 2, 3, 4, 5
}
```

**Depth:** 6 levels (maximum)

### Optimization: Wildcard Short-Circuit

When wildcard encountered, nested switches stop:

**Without optimization (BAD):**
```go
case Err:
    e := *elem0.err_0
    switch elem1.tag { // Unnecessary - wildcard at position 1
    case Ok:
        switch elem2.tag { // Unnecessary - wildcard at position 2
        case Ok: // Unreachable
        case Err: // Unreachable
        }
    case Err:
        // ...
    }
```

**With optimization (GOOD):**
```go
case Err:
    e := *elem0.err_0
    // Wildcards at positions 1, 2: match all
    return "First failed: " + e.Error()
```

Golden files demonstrate this optimization throughout.

## Binding Extraction Patterns

### Correct Binding in Golden Files

**Pattern:** `Ok(x)` at position N
**Generated:**
```go
x := *__match_M_elemN.ok_0
```

Where:
- `M` = match expression counter (0, 1, 2, ...)
- `N` = tuple element position (0, 1, 2, ...)
- `ok_0` = field name for Ok variant value

**Pattern:** `Err(e)` at position N
**Generated:**
```go
e := *__match_M_elemN.err_0
```

### Wildcard Binding Behavior

**Pattern:** `Ok(_)` (wildcard binding)
**Generated:**
```go
// No binding - value ignored
switch __match_0_elem1.tag {
case ResultTagOk:
    // Continue without extracting value
```

**Pattern:** `_` (full wildcard)
**Generated:**
```go
// Wildcard at position 2: matches all
return "..."
// No switch, no binding
```

## Test File Organization

### File Naming Logic

**Format:** `pattern_match_{NN}_tuple_{description}`

**Rationale:**
- `pattern_match_` - Feature prefix (consistent with Phase 4.1 tests)
- `{NN}` - Sequential numbering (09-12 follow existing tests)
- `tuple_` - Sub-feature identifier (distinguishes from guards, swift, etc.)
- `{description}` - Specific aspect tested

**Numbers chosen:**
- `09` - Follows `pattern_match_01..08` (guards, swift, etc. from plan)
- `10` - Sequential
- `11` - Sequential
- `12` - Sequential

This leaves room for:
- `pattern_match_13..16` - Error tests (from plan)
- `pattern_match_17+` - Future extensions

### File Size Guidelines

**From golden test guidelines:**
- Basic tests: 10-30 lines
- Intermediate tests: 30-50 lines
- Advanced tests: Up to 50 lines

**Actual sizes:**
- Test 09: 29 lines ✅ (basic)
- Test 10: 30 lines ✅ (intermediate)
- Test 11: 30 lines ✅ (intermediate)
- Test 12: 58 lines ⚠️ (advanced - slightly over but justified)

**Test 12 justification:**
- Demonstrates TWO match expressions (6-element and 4-element)
- Includes commented non-exhaustive example (educational)
- Maximum complexity test (appropriate for advanced)
- Still scannable and understandable

### Comment Header Standard

Every `.dingo` file includes:
```dingo
// Test: [Brief description of what this test demonstrates]
// Feature: Tuple pattern destructuring
// Complexity: [basic|intermediate|advanced]
```

**Purpose:**
- Quick orientation for developers
- Shows test progression (complexity level)
- Consistent documentation format

## Golden Output Quality Assurance

### Compilation Verification

All tests pass `TestGoldenFilesCompilation`:
- Go parser successfully parses `.go.golden`
- No syntax errors
- All imports resolved
- Type declarations correct

### Marker Completeness

Every `.go.golden` includes:

1. **Match start marker:**
   ```go
   // DINGO_MATCH_START: (r1, r2)
   ```

2. **Tuple pattern marker:**
   ```go
   // DINGO_TUPLE_PATTERN: (Ok, Ok) | (Ok, Err) | (Err, Ok) | (Err, Err) | ARITY: 2
   ```

3. **Arm markers (optional, for plugin):**
   ```go
   // DINGO_TUPLE_ARM: (Ok(x), Ok(y))
   ```

4. **Match end marker:**
   ```go
   // DINGO_MATCH_END
   ```

5. **Wildcard comments:**
   ```go
   // Wildcard at position 2: matches all
   ```

### Idiomatic Go

Golden files follow Go best practices:
- Proper enum type definition
- Constructor functions for variants
- Pointer fields for variant values
- Panic for unreachable code
- Clear variable naming

**Example enum structure:**
```go
type ResultTag uint8

const (
    ResultTagOk ResultTag = iota
    ResultTagErr
)

type Result struct {
    tag   ResultTag
    ok_0  *int
    err_0 *error
}

func Result_Ok(v0 int) Result {
    return Result{tag: ResultTagOk, ok_0: &v0}
}

func Result_Err(v0 error) Result {
    return Result{tag: ResultTagErr, err_0: &v0}
}
```

This structure is consistent across all tests.

## Future Activation Plan

### When Tests Will Execute

Currently SKIPPED. Will activate when:

1. **Plugin transformation (Task D1) complete:**
   - Nested switch generation implemented
   - Binding extraction working
   - Wildcard optimization functional

2. **Skip check removed:**
   - Remove Phase 3 deferral check in `golden_test.go`
   - Tests will execute normally

3. **Expected outcome:**
   - Transpiler generates output
   - Output compared to `.go.golden`
   - 4/4 tests PASS ✅

### Validation Steps

When tests activate:

1. **Syntax check:** `.go.golden` compiles ✅ (already passing)
2. **Transpiler execution:** `.dingo` → actual output
3. **Comparison:** `diff actual golden`
4. **Verification:** All differences explained

If tests fail:
- Check actual output for correctness
- Compare with golden files
- Update golden if transpiler behavior changed (intentionally)
- Fix transpiler if output is wrong

## Summary

**Design principles:**
- ✅ Teach by example (realistic scenarios)
- ✅ Demonstrate key behaviors (exhaustiveness, wildcards)
- ✅ Perfect golden output (idiomatic Go)

**Tuple arities:**
- ✅ 2, 3, 4, 6 elements (representative coverage)
- ✅ 6-element limit enforced
- ✅ Strategic wildcard usage

**Wildcard semantics:**
- ✅ Catch-all behavior (user decision)
- ✅ Partial and full wildcards
- ✅ Optimization demonstrated

**Exhaustiveness:**
- ✅ Complete coverage patterns
- ✅ Strategic wildcard placement
- ✅ Non-exhaustive example (educational)

**Code quality:**
- ✅ All tests compile
- ✅ Idiomatic Go output
- ✅ Complete marker set
- ✅ Clear documentation

**Ready for activation when plugin transformation complete.**

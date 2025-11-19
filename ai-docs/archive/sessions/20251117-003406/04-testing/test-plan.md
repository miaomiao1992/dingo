# Test Plan for Functional Utilities Implementation
## Session: 20251117-003406
## Date: 2025-11-17

---

## 1. Requirements Understanding

The functional utilities implementation provides slice transformation methods that transpile to inline Go loops. Based on the implementation and fixes:

### Core Behaviors to Validate
1. **Map**: Transforms each element with a function, producing new slice
2. **Filter**: Selects elements matching a predicate
3. **Reduce**: Aggregates slice into single value
4. **Sum**: Shorthand for numeric sum (syntactic sugar for reduce)
5. **Count**: Counts elements matching predicate
6. **All**: Returns true if all elements match predicate (short-circuit)
7. **Any**: Returns true if any element matches predicate (short-circuit)

### Critical Implementation Characteristics
- **IIFE Pattern**: All transformations wrap in immediately-invoked function expressions
- **Capacity Hints**: All allocations use `make([]T, 0, len(input))` for performance
- **Early Exit**: `all()` and `any()` use `break` statements
- **Type Inference**: Extract types from function signatures
- **Deep Cloning**: Use `astutil.Apply` for safe AST cloning
- **Function Arity Validation**: Validate parameter counts (map/filter: 1, reduce: 2)

### Critical Edge Cases
1. Nil slice handling (parser limitation)
2. Empty slices
3. Single element slices
4. Type inference with explicit return types
5. Invalid function arity
6. Complex expressions in function bodies
7. Chained method calls

---

## 2. Test Scenarios

### Test Strategy
- **Unit Tests**: Already exist, validate AST transformation logic
- **Integration Tests**: Create golden files showing end-to-end transpilation
- **Compilation Tests**: Verify generated Go code compiles
- **Runtime Tests**: Verify generated code produces correct results

---

### Scenario 1: Basic Map Operation
**Purpose**: Validate simple element transformation
**Input**: Integer slice with doubling function
**Expected**: IIFE with for-range loop and append
**Rationale**: Most common use case, validates core map logic

**Test Code**:
```go
package main

func main() {
	numbers := []int{1, 2, 3, 4, 5}
	doubled := numbers.map(func(x int) int { return x * 2 })
	println(doubled)
}
```

**Expected Output**:
```go
package main

func main() {
	numbers := []int{1, 2, 3, 4, 5}
	doubled := func() []int {
		var __temp0 []int
		__temp0 = make([]int, 0, len(numbers))
		for _, x := range numbers {
			__temp0 = append(__temp0, x*2)
		}
		return __temp0
	}()
	println(doubled)
}
```

---

### Scenario 2: Filter with Predicate
**Purpose**: Validate conditional element selection
**Input**: Integer slice filtering positive numbers
**Expected**: IIFE with if statement inside loop
**Rationale**: Core filter functionality, tests conditional append

**Test Code**:
```go
package main

func main() {
	numbers := []int{-2, -1, 0, 1, 2}
	positives := numbers.filter(func(x int) bool { return x > 0 })
	println(positives)
}
```

---

### Scenario 3: Reduce for Aggregation
**Purpose**: Validate accumulator pattern
**Input**: Integer slice with sum reducer
**Expected**: IIFE with accumulator variable
**Rationale**: Tests reduce logic, accumulator management

**Test Code**:
```go
package main

func main() {
	numbers := []int{1, 2, 3, 4, 5}
	sum := numbers.reduce(0, func(acc int, x int) int { return acc + x })
	println(sum)
}
```

---

### Scenario 4: Sum Helper
**Purpose**: Validate sum syntactic sugar
**Input**: Float64 slice (tests type inference for non-int)
**Expected**: IIFE with typed accumulator
**Rationale**: Tests type inference improvements from CRITICAL-3 fix

**Test Code**:
```go
package main

func main() {
	prices := []float64{10.5, 20.3, 15.7}
	total := prices.sum()
	println(total)
}
```

---

### Scenario 5: All with Early Exit
**Purpose**: Validate short-circuit evaluation
**Input**: Integer slice checking all positive
**Expected**: IIFE with break statement
**Rationale**: Tests early exit optimization

**Test Code**:
```go
package main

func main() {
	numbers := []int{1, 2, 3, 4, 5}
	allPositive := numbers.all(func(x int) bool { return x > 0 })
	println(allPositive)
}
```

---

### Scenario 6: Any with Early Exit
**Purpose**: Validate short-circuit with false default
**Input**: Integer slice checking for negative numbers
**Expected**: IIFE with break when condition true
**Rationale**: Tests any() logic opposite of all()

**Test Code**:
```go
package main

func main() {
	numbers := []int{1, 2, 3, 4, 5}
	hasNegative := numbers.any(func(x int) bool { return x < 0 })
	println(hasNegative)
}
```

---

### Scenario 7: Count with Predicate
**Purpose**: Validate counting with condition
**Input**: Integer slice counting evens
**Expected**: IIFE with counter increment in if
**Rationale**: Tests count helper

**Test Code**:
```go
package main

func main() {
	numbers := []int{1, 2, 3, 4, 5, 6}
	evenCount := numbers.count(func(x int) bool { return x%2 == 0 })
	println(evenCount)
}
```

---

### Scenario 8: Method Chaining
**Purpose**: Validate sequential transformations
**Input**: Filter then map
**Expected**: Nested IIFEs or sequential operations
**Rationale**: Tests chaining support, multiple temp variables

**Test Code**:
```go
package main

func main() {
	numbers := []int{1, 2, 3, 4, 5, 6}
	result := numbers.filter(func(x int) bool { return x%2 == 0 }).map(func(x int) int { return x * 10 })
	println(result)
}
```

---

### Scenario 9: Complex Type Transformation
**Purpose**: Validate struct type transformations
**Input**: Struct slice extracting field
**Expected**: Proper type inference for struct field
**Rationale**: Tests type system integration

**Test Code**:
```go
package main

type Person struct {
	name string
	age  int
}

func main() {
	people := []Person{
		{"Alice", 30},
		{"Bob", 25},
	}
	names := people.map(func(p Person) string { return p.name })
	println(names)
}
```

---

### Scenario 10: Empty Slice Edge Case
**Purpose**: Validate behavior with empty input
**Input**: Empty slice
**Expected**: Returns empty slice (capacity 0)
**Rationale**: Tests edge case handling

**Test Code**:
```go
package main

func main() {
	empty := []int{}
	doubled := empty.map(func(x int) int { return x * 2 })
	println(len(doubled))
}
```

---

### Scenario 11: Single Element Slice
**Purpose**: Validate minimal case
**Input**: Single element slice
**Expected**: Returns single element result
**Rationale**: Boundary condition

**Test Code**:
```go
package main

func main() {
	single := []int{42}
	doubled := single.map(func(x int) int { return x * 2 })
	println(doubled[0])
}
```

---

## 3. Test Implementation Plan

### Phase 1: Unit Tests (Already Complete)
- ✅ TestTransformMap
- ✅ TestTransformFilter
- ✅ TestTransformReduce
- ✅ TestTransformSum
- ✅ TestTransformAll
- ✅ TestTransformAny

### Phase 2: Golden File Tests (This Session)
Create `.dingo` and `.go.golden` file pairs:

1. `functional_01_map.dingo` + `.go.golden`
2. `functional_02_filter.dingo` + `.go.golden`
3. `functional_03_reduce.dingo` + `.go.golden`
4. `functional_04_sum.dingo` + `.go.golden`
5. `functional_05_all.dingo` + `.go.golden`
6. `functional_06_any.dingo` + `.go.golden`
7. `functional_07_count.dingo` + `.go.golden`
8. `functional_08_chaining.dingo` + `.go.golden`
9. `functional_09_structs.dingo` + `.go.golden`
10. `functional_10_empty.dingo` + `.go.golden`

### Phase 3: Compilation Validation
For each golden file:
1. Verify `.go.golden` compiles with `go build`
2. Verify generated code matches golden file
3. Document any discrepancies

### Phase 4: Runtime Validation
Run compiled programs and verify output:
1. Map produces correct transformations
2. Filter selects correct elements
3. Reduce computes correct aggregate
4. Sum calculates correct total
5. All/any return correct boolean
6. Count returns correct integer

---

## 4. Verification Checklist

### Before Reporting Test Failure
- [ ] Test logic matches requirements from feature spec
- [ ] Test setup includes proper type annotations
- [ ] Expected values match actual Dingo semantics
- [ ] Failure is reproducible
- [ ] Similar tests behave consistently
- [ ] Can articulate why this is implementation bug vs test bug

### For Each Test Scenario
- [ ] Input data is valid
- [ ] Function signatures have explicit types
- [ ] Expected output is hand-verified
- [ ] Test covers unique aspect
- [ ] Test name clearly describes scenario

---

## 5. Success Criteria

### Functional Correctness
- [ ] All unit tests pass
- [ ] All golden files match transpiled output
- [ ] All generated Go code compiles
- [ ] All runtime outputs are correct

### Code Quality
- [ ] Generated code is readable
- [ ] Uses IIFE pattern consistently
- [ ] Includes capacity hints
- [ ] Has early exit optimizations where applicable

### Review Fixes Validation
- [ ] Deep cloning works (no AST corruption)
- [ ] IIFE return types present
- [ ] Type inference with validation
- [ ] Arity validation catches errors
- [ ] Error logging helps debugging

---

## 6. Known Limitations (Expected Test Constraints)

### Parser Limitations
- Cannot test with composite literals directly (parser issue)
- Must use explicit variable declarations
- Short form `:=` may have issues in some contexts

### Feature Limitations (By Design)
- Result/Option integration not implemented yet
- Only single-statement function bodies supported
- Complex multi-statement functions return nil (no transform)

### Test Approach for Limitations
- Document which tests are blocked by parser
- Create alternative tests using supported syntax
- Mark skipped tests with clear reasons
- Verify limitations are logged properly

---

## 7. Test Execution Strategy

### Order of Execution
1. Run unit tests first (fast feedback)
2. Create golden files
3. Run golden test suite
4. Compile generated Go files
5. Run compiled programs (if compilation succeeds)

### Failure Analysis Priority
1. CRITICAL: Compilation failures in generated code
2. HIGH: Wrong output structure (missing IIFE, loops, etc.)
3. MEDIUM: Performance issues (missing capacity hints, early exit)
4. LOW: Formatting differences

---

## 8. Documentation Requirements

For each test failure:
- Exact input code
- Expected output (with reasoning)
- Actual output
- Root cause analysis
- Whether it's implementation bug or test bug
- Suggested fix (if implementation bug)

For each test success:
- Brief confirmation
- Any notable observations
- Performance characteristics verified

---

## Test Timeline

1. **Test Creation**: 30 minutes
   - Create 10 golden file pairs
   - Hand-verify expected outputs

2. **Test Execution**: 15 minutes
   - Run unit tests
   - Run golden tests
   - Attempt compilation

3. **Results Analysis**: 30 minutes
   - Categorize failures
   - Verify fixes from review
   - Document root causes

4. **Documentation**: 15 minutes
   - Write test results report
   - Create summary file
   - Update CHANGELOG if needed

**Total Estimated Time**: 90 minutes

---

## Notes

This test plan focuses on:
- **Balanced Coverage**: Not too many tests (overwhelming) nor too few (insufficient)
- **Critical Paths**: Map, filter, reduce as core; helpers as secondary
- **Real Bugs**: Each test designed to catch specific implementation issues
- **Clarity**: Test names and purposes are self-documenting
- **Verification**: Multiple levels (AST, compilation, runtime)

The plan validates all CRITICAL and IMPORTANT fixes from the code review while ensuring the core functionality works end-to-end.

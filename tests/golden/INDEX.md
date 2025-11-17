# Golden Test Index

Comprehensive catalog of all golden tests for the Dingo transpiler.

## Quick Navigation

- [Error Propagation](#error-propagation-) - 8 tests
- [Result Type](#result-type-) - 5 tests
- [Option Type](#option-type-) - 4 tests
- [Sum Types](#sum-types-) - 5 tests
- [Lambdas](#lambdas-) - 4 tests
- [Ternary Operator](#ternary-operator-) - 3 tests
- [Null Coalescing](#null-coalescing-) - 3 tests
- [Safe Navigation](#safe-navigation-) - 3 tests
- [Pattern Matching](#pattern-matching-) - 4 tests
- [Tuples](#tuples-) - 3 tests
- [Functional Utilities](#functional-utilities-) - 4 tests

**Total: 46 tests**

---

## Error Propagation (?)

Tests for the `?` operator that propagates errors in `(T, error)` returns.

| # | File | Description | Complexity |
|---|------|-------------|------------|
| 01 | `error_prop_01_simple.dingo` | Simple error propagation with one `?` | Basic |
| 02 | `error_prop_02_multiple.dingo` | Multiple `?` operators in sequence | Basic |
| 03 | `error_prop_03_expression.dingo` | `?` operator in expression context | Basic |
| 04 | `error_prop_04_wrapping.dingo` | Error propagation with wrapping | Intermediate |
| 05 | `error_prop_05_complex_types.dingo` | `?` with complex type conversions | Intermediate |
| 06 | `error_prop_06_mixed_context.dingo` | Mixed error handling contexts | Intermediate |
| 07 | `error_prop_07_special_chars.dingo` | Special characters in error messages | Intermediate |
| 08 | `error_prop_08_chained_calls.dingo` | Chained method calls with `?` | Advanced |

**Key Features Tested:**
- Single and multiple `?` operators
- Error propagation in different contexts
- Type conversions with error handling
- Method chaining with error propagation

**Related:** `features/error-propagation.md`

---

## Result Type (result_*)

Tests for the `Result<T, E>` enum type (alternative to Go's `(T, error)`).

| # | File | Description | Complexity |
|---|------|-------------|------------|
| 01 | `result_01_basic.dingo` | Basic Result construction and checking | Basic |
| 02 | `result_02_propagation.dingo` | Result with `?` operator | Intermediate |
| 03 | `result_03_pattern_match.dingo` | Result with pattern matching | Intermediate |
| 04 | `result_04_chaining.dingo` | Result with map/and_then chaining | Advanced |
| 05 | `result_05_go_interop.dingo` | Result interop with Go `(T, error)` | Intermediate |

**Key Features Tested:**
- Result enum construction (`Result_Ok`, `Result_Err`)
- IsOk/IsErr checking methods
- Pattern matching on Result variants
- Functional combinators (map, and_then)
- Interoperability with standard Go functions

**Related:** `features/result-type.md`

---

## Option Type (option_*)

Tests for the `Option<T>` enum type (alternative to nil pointers).

| # | File | Description | Complexity |
|---|------|-------------|------------|
| 01 | `option_01_basic.dingo` | Basic Option construction and checking | Basic |
| 02 | `option_02_pattern_match.dingo` | Option with pattern matching | Intermediate |
| 03 | `option_03_chaining.dingo` | Option with map/and_then chaining | Advanced |
| 04 | `option_04_go_interop.dingo` | Option interop with Go nil values | Intermediate |

**Key Features Tested:**
- Option enum construction (`Option_Some`, `Option_None`)
- IsSome/IsNone checking methods
- Pattern matching on Option variants
- Functional combinators (map, and_then)
- Interoperability with Go maps and pointers

**Related:** `features/option-type.md`

---

## Sum Types (sum_types_*)

Tests for sum types/enums with variants.

| # | File | Description | Complexity |
|---|------|-------------|------------|
| 01 | `sum_types_01_simple.dingo` | Simple enum with unit variants | Basic |
| 02 | `sum_types_02_struct_variant.dingo` | Enum with struct variant | Intermediate |
| 03 | `sum_types_03_generic.dingo` | Generic enum with type parameters | Advanced |
| 04 | `sum_types_04_multiple.dingo` | Multiple enum definitions | Intermediate |
| 05 | `sum_types_05_nested.dingo` | Nested enum variants | Advanced |

**Key Features Tested:**
- Unit variants (e.g., `Status_Pending`)
- Tuple variants (e.g., `Result_Ok(value)`)
- Struct variants (e.g., `Color_RGB{r, g, b}`)
- Generic enums with type parameters
- Nested enum structures

**Related:** `features/sum-types.md`, `features/enums.md`

---

## Lambdas (lambda_*)

Tests for lambda/closure expressions.

| # | File | Description | Complexity |
|---|------|-------------|------------|
| 01 | `lambda_01_basic.dingo` | Basic Rust-style lambda syntax | Basic |
| 02 | `lambda_02_multiline.dingo` | Multiline lambda with blocks | Intermediate |
| 03 | `lambda_03_closure.dingo` | Closures capturing variables | Advanced |
| 04 | `lambda_04_higher_order.dingo` | Higher-order functions with lambdas | Advanced |

**Key Features Tested:**
- Rust-style lambda syntax: `|x, y| x + y`
- Single-expression lambdas
- Block lambdas with multiple statements
- Variable capture (closures)
- Lambdas as function arguments
- Function composition

**Related:** `features/lambdas.md`

---

## Ternary Operator (ternary_*)

Tests for the ternary conditional operator.

| # | File | Description | Complexity |
|---|------|-------------|------------|
| 01 | `ternary_01_basic.dingo` | Basic ternary operator usage | Basic |
| 02 | `ternary_02_nested.dingo` | Nested ternary expressions | Intermediate |
| 03 | `ternary_03_complex.dingo` | Complex ternary with function calls | Advanced |

**Key Features Tested:**
- Basic `condition ? true_val : false_val` syntax
- Nested ternary operators
- Ternary in variable assignments
- Ternary with function calls
- Complex expressions in branches

**Related:** `features/ternary-operator.md`

---

## Null Coalescing (null_coalesce_*)

Tests for the null coalescing operator `??`.

| # | File | Description | Complexity |
|---|------|-------------|------------|
| 01 | `null_coalesce_01_basic.dingo` | Basic `??` operator usage | Basic |
| 02 | `null_coalesce_02_chained.dingo` | Chained null coalescing | Intermediate |
| 03 | `null_coalesce_03_with_option.dingo` | `??` with Option type | Advanced |

**Key Features Tested:**
- Basic `value ?? default` syntax
- Chained coalescing: `a ?? b ?? c ?? default`
- Integration with Option type
- Fallback chains with function calls

**Related:** `features/null-coalescing.md`

---

## Safe Navigation (safe_nav_*)

Tests for the safe navigation operator `?.`.

| # | File | Description | Complexity |
|---|------|-------------|------------|
| 01 | `safe_nav_01_basic.dingo` | Basic `?.` operator usage | Basic |
| 02 | `safe_nav_02_chained.dingo` | Chained safe navigation | Intermediate |
| 03 | `safe_nav_03_with_methods.dingo` | `?.` with method calls | Advanced |

**Key Features Tested:**
- Basic `obj?.field` syntax
- Chained navigation: `user?.address?.city`
- Safe navigation with method calls
- Integration with null coalescing

**Related:** `features/null-safety.md`

---

## Pattern Matching (pattern_match_*)

Tests for the `match` expression.

| # | File | Description | Complexity |
|---|------|-------------|------------|
| 01 | `pattern_match_01_basic.dingo` | Basic match on enum | Basic |
| 02 | `pattern_match_02_guards.dingo` | Match with conditional guards | Intermediate |
| 03 | `pattern_match_03_nested.dingo` | Nested pattern matching | Advanced |
| 04 | `pattern_match_04_exhaustive.dingo` | Exhaustive pattern checking | Intermediate |

**Key Features Tested:**
- Basic match syntax
- Pattern destructuring
- Match guards (`if` conditions)
- Nested patterns
- Exhaustiveness checking
- Wildcard patterns (`_`)

**Related:** `features/pattern-matching.md`

---

## Tuples (tuples_*)

Tests for tuple types and destructuring.

| # | File | Description | Complexity |
|---|------|-------------|------------|
| 01 | `tuples_01_basic.dingo` | Basic tuple creation and destructuring | Basic |
| 02 | `tuples_02_destructure.dingo` | Advanced destructuring patterns | Intermediate |
| 03 | `tuples_03_nested.dingo` | Nested tuple structures | Advanced |

**Key Features Tested:**
- Tuple creation: `(a, b, c)`
- Tuple destructuring: `let (x, y) = getTuple()`
- Ignoring values with `_`
- Nested tuples
- Tuples as return values

**Related:** `features/tuples.md`

---

## Functional Utilities (func_util_*)

Tests for functional programming utilities (map, filter, reduce).

| # | File | Description | Complexity |
|---|------|-------------|------------|
| 01 | `func_util_01_map.dingo` | Map function on slices | Basic |
| 02 | `func_util_02_filter.dingo` | Filter function on slices | Basic |
| 03 | `func_util_03_reduce.dingo` | Reduce/fold function on slices | Intermediate |
| 04 | `func_util_04_chaining.dingo` | Chaining map/filter/reduce | Advanced |

**Key Features Tested:**
- Map: transforming collections
- Filter: selecting elements by predicate
- Reduce: aggregating collections
- Method chaining for functional composition
- Integration with lambdas

**Related:** `features/functional-utilities.md`

---

## Test File Structure

Each test consists of two files:

```
{feature}_{NN}_{description}.dingo      # Dingo source code
{feature}_{NN}_{description}.go.golden  # Expected Go output
```

### Optional Files
```
{feature}_{NN}_{description}.go.actual   # Generated during test runs (gitignored)
{feature}_{NN}_{description}.reasoning.md # Design decisions (for complex tests)
```

---

## Running Tests

### All Golden Tests
```bash
cd tests
go test -v -run TestGoldenFiles
```

### Specific Test
```bash
go test -v -run TestGoldenFiles/error_prop_01_simple
```

### Compilation Check
```bash
go test -v -run TestGoldenFilesCompilation
```

### Update Golden Files
```bash
# After verifying .actual files are correct
find golden -name "*.go.actual" -exec sh -c 'mv "$1" "${1%.actual}.golden"' _ {} \;
```

---

## Coverage Status

| Feature | Tests | Status | Notes |
|---------|-------|--------|-------|
| Error Propagation | 8 | ✅ Complete | Core feature, well tested |
| Result Type | 5 | ✅ Complete | All .go.golden files generated |
| Option Type | 4 | ✅ Complete | All .go.golden files generated |
| Sum Types | 5 | ✅ Complete | Includes generics and nesting |
| Lambdas | 4 | ✅ Complete | All .go.golden files generated |
| Ternary | 3 | ✅ Complete | All .go.golden files generated |
| Null Coalescing | 3 | ✅ Complete | All .go.golden files generated |
| Safe Navigation | 3 | ✅ Complete | All .go.golden files generated |
| Pattern Matching | 4 | ✅ Complete | All .go.golden files generated |
| Tuples | 3 | ✅ Complete | All .go.golden files generated |
| Functional Utilities | 4 | ✅ Complete | All .go.golden files generated |

**Legend:**
- ✅ Complete - Both .dingo and .go.golden exist and compile successfully
- ⚠️ Needs .golden - .dingo files created, need corresponding .go.golden
- ❌ Missing - Tests not yet created
- 🔧 WIP - Work in progress

---

## Contributing New Tests

### Before Adding a Test

1. Read `GOLDEN_TEST_GUIDELINES.md` thoroughly
2. Check if similar test already exists
3. Verify feature is implemented in transpiler
4. Follow naming convention: `{feature}_{NN}_{description}.dingo`

### Creating a Test

1. Write `.dingo` file with realistic example
2. Run transpiler: `dingo build {file}.dingo`
3. Review generated Go code
4. Copy to `.go.golden` if correct
5. Format: `gofmt -w {file}.go.golden`
6. Run test: `go test -run TestGoldenFiles/{test_name}`
7. Update this INDEX.md

### Guidelines Quick Reference

- **One feature per test** - Focus on clarity
- **Realistic examples** - Not contrived code
- **Meaningful names** - `user`, not `x`
- **10-50 lines** - Keep tests scannable
- **Compilable output** - Must pass `go/parser`
- **No external deps** - Stdlib only

---

## Related Documentation

- **Guidelines:** `GOLDEN_TEST_GUIDELINES.md` - Comprehensive test writing rules
- **Plan:** `REORGANIZATION_PLAN.md` - Historical reorganization details
- **Test Runner:** `../golden_test.go` - Test harness implementation
- **Features:** `../../features/` - Feature specifications

---

**Last Updated:** 2025-11-17
**Test Count:** 46 tests (all complete with .go.golden files)
**Compilation Status:** ✅ All 46 .go.golden files compile successfully
**Maintained By:** Dingo Project Contributors

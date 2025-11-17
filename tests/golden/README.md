# Golden Tests - Dingo Transpiler

This directory contains golden file tests that verify the Dingo-to-Go transpilation is correct.

## Quick Start

```bash
# Run all golden tests
cd tests
go test -v -run TestGoldenFiles

# Run specific test
go test -v -run TestGoldenFiles/error_prop_01_simple

# Check compilation only
go test -v -run TestGoldenFilesCompilation
```

## What Are Golden Tests?

Golden tests compare the transpiler's output against known-good reference files:

1. **Input:** `.dingo` file with Dingo syntax
2. **Expected:** `.go.golden` file with idiomatic Go code
3. **Actual:** `.go.actual` generated during test (temporary)
4. **Test:** Compares actual vs expected, reports differences

## Directory Contents

- **46 test files** covering 11 feature categories
- **46 complete** tests (all with .go.golden files)
- **All tests compile successfully** ✅

## File Structure

```
error_prop_01_simple.dingo        # Dingo source
error_prop_01_simple.go.golden    # Expected Go output
error_prop_01_simple.go.actual    # Generated output (temporary)
error_prop_01_simple.reasoning.md # Optional: design notes
```

## Documentation

📚 **Start here:**
- **[INDEX.md](INDEX.md)** - Complete catalog of all 46 tests
- **[GOLDEN_TEST_GUIDELINES.md](GOLDEN_TEST_GUIDELINES.md)** - How to write golden tests

📋 **Reference:**
- **[REORGANIZATION_PLAN.md](REORGANIZATION_PLAN.md)** - Historical: how we reorganized

### Reasoning Documentation

Each test file can have an optional `.reasoning.md` file that explains the **"why"** behind the test, linking to Go community discussions, feature proposals, and design decisions.

**File Organization:**
```
sum_types_01_simple_enum.dingo
sum_types_01_simple_enum.reasoning.md    # Optional: design rationale
sum_types_01_simple_enum.go.golden
```

**What Reasoning Docs Provide:**
- **Community Context**: Links to official Go proposals and discussions
- **Design Rationale**: Why we made specific implementation choices
- **Feature References**: Links to corresponding feature documentation
- **External References**: How other languages (Rust, Swift, TypeScript, Kotlin) solve the same problem
- **Success Metrics**: Code reduction percentages, type safety improvements
- **Future Work**: Planned enhancements and known limitations

**Completed Reasoning Docs:**
- [sum_types_01_simple_enum.reasoning.md](./sum_types_01_simple_enum.reasoning.md) - Basic enum (79% code reduction)
- [sum_types_02_struct_variant.reasoning.md](./sum_types_02_struct_variant.reasoning.md) - Enum with data (78% code reduction)
- [01_simple_statement.reasoning.md](./01_simple_statement.reasoning.md) - Error propagation suite (covers all 8 tests)

**Go Proposal Reference Map:**

| Dingo Feature | Go Proposal | Community Support |
|---------------|-------------|-------------------|
| Sum Types | [#19412](https://github.com/golang/go/issues/19412) | 996+ 👍 (highest ever) |
| Error Propagation (`?`) | [#71203](https://github.com/golang/go/issues/71203) | 200+ comments (2025) |
| Lambda Functions | [#21498](https://github.com/golang/go/issues/21498) | 750+ 👍 |

**Community Resources:**
- [Go Proposals Repository](https://github.com/golang/proposal)
- [Go Error Handling Feedback](https://go.dev/wiki/Go2ErrorHandlingFeedback)
- [Rust Error Handling](https://doc.rust-lang.org/book/ch09-02-recoverable-errors-with-result.html)
- [Swift Enums](https://docs.swift.org/swift-book/documentation/the-swift-programming-language/enumerations/)
- [TypeScript Discriminated Unions](https://www.typescriptlang.org/docs/handbook/unions-and-intersections.html)

**Code Reduction Metrics:**

| Feature | Average Reduction | Range |
|---------|------------------|-------|
| Sum Types | 78-79% | 7-10 lines → 33-46 lines |
| Error Propagation | 65% | 60-70% |
| Pattern Matching | 70% | 65-75% |
| Lambda Functions | 60% | 50-70% |

## Test Categories

| Category | Tests | Status |
|----------|-------|--------|
| Error Propagation (`?`) | 8 | ✅ Complete |
| Result Type | 5 | ✅ Complete |
| Option Type | 4 | ✅ Complete |
| Sum Types | 5 | ✅ Complete |
| Lambdas | 4 | ✅ Complete |
| Ternary Operator | 3 | ✅ Complete |
| Null Coalescing | 3 | ✅ Complete |
| Safe Navigation | 3 | ✅ Complete |
| Pattern Matching | 4 | ✅ Complete |
| Tuples | 3 | ✅ Complete |
| Functional Utilities | 4 | ✅ Complete |

See [INDEX.md](INDEX.md) for detailed test list.

## Naming Convention

```
{feature}_{NN}_{description}.dingo
```

**Examples:**
- `error_prop_01_simple.dingo` - Error propagation, test #01, simple example
- `result_03_pattern_match.dingo` - Result type, test #03, with pattern matching
- `lambda_04_higher_order.dingo` - Lambdas, test #04, higher-order functions

**Feature prefixes:**
- `error_prop_` - Error propagation (?)
- `result_` - Result type
- `option_` - Option type
- `sum_types_` - Sum types/enums
- `lambda_` - Lambdas
- `ternary_` - Ternary operator
- `null_coalesce_` - Null coalescing (??)
- `safe_nav_` - Safe navigation (?.)
- `pattern_match_` - Pattern matching
- `tuples_` - Tuples
- `func_util_` - Functional utilities

## Writing New Tests

### Quick Checklist

✅ Read [GOLDEN_TEST_GUIDELINES.md](GOLDEN_TEST_GUIDELINES.md) first
✅ Follow naming convention: `{feature}_{NN}_{description}.dingo`
✅ Use realistic examples (not contrived code)
✅ Keep tests focused (one feature per test)
✅ Include both `.dingo` and `.go.golden` files
✅ Ensure `.go.golden` compiles
✅ Update [INDEX.md](INDEX.md)

### Example Process

```bash
# 1. Write test
vim tests/golden/feature_05_example.dingo

# 2. Generate Go code
cd ../..
go run cmd/dingo/main.go build tests/golden/feature_05_example.dingo

# 3. Review and copy to golden
cp tests/golden/feature_05_example.go tests/golden/feature_05_example.go.golden

# 4. Format
gofmt -w tests/golden/feature_05_example.go.golden

# 5. Run test
cd tests
go test -v -run TestGoldenFiles/feature_05_example
```

## ✅ All Tests Complete!

All 46 tests now have both `.dingo` and `.go.golden` files.

**Completion Summary:**
- ✅ **46/46 tests** have .dingo files
- ✅ **46/46 tests** have .go.golden files
- ✅ **All 46 .go.golden files compile** successfully
- ✅ **100% test coverage** across 11 feature categories

**Recently Generated (2025-11-17):**
- Result Type: 3 new .go.golden files (pattern match, chaining, Go interop)
- Option Type: 2 new .go.golden files (chaining, Go interop)
- Lambdas: 3 new .go.golden files (multiline, closures, higher-order)
- Ternary: 2 new .go.golden files (nested, complex)
- Null Coalescing: 2 new .go.golden files (chained, with Option)
- Safe Navigation: 2 new .go.golden files (chained, with methods)
- Pattern Matching: 4 new .go.golden files (basic, guards, nested, exhaustive)
- Tuples: 3 new .go.golden files (basic, destructure, nested)
- Functional Utilities: 4 new .go.golden files (map, filter, reduce, chaining)
- Sum Types: 1 new .go.golden file (nested)

## Test Organization

Tests progress from **basic** → **intermediate** → **advanced**:

- **01-03:** Basic usage, minimal examples
- **04-06:** Intermediate, real-world scenarios
- **07+:** Advanced, edge cases, complex combinations

Example progression for Result type:
1. `result_01_basic.dingo` - Simple Result construction
2. `result_02_propagation.dingo` - With ? operator
3. `result_03_pattern_match.dingo` - With match expression
4. `result_04_chaining.dingo` - With map/and_then
5. `result_05_go_interop.dingo` - Interop with Go

## Updating Golden Files

When transpiler output changes intentionally:

```bash
# 1. Run tests to generate .actual files
cd tests
go test -v -run TestGoldenFiles

# 2. Review differences
diff golden/test_name.go.actual golden/test_name.go.golden

# 3. If correct, update golden file
mv golden/test_name.go.actual golden/test_name.go.golden

# 4. Or update all at once
find golden -name "*.go.actual" -exec sh -c 'mv "$1" "${1%.actual}.golden"' _ {} \;
```

## Integration with CI

Golden tests run automatically in CI:

```yaml
# .github/workflows/test.yml
- name: Golden Tests
  run: |
    cd tests
    go test -v -run TestGoldenFiles
    go test -v -run TestGoldenFilesCompilation
```

## Guidelines

**DO:**
- ✅ Test one feature per file
- ✅ Use realistic, meaningful examples
- ✅ Keep tests small (10-50 lines)
- ✅ Make output idiomatic Go
- ✅ Ensure code compiles
- ✅ Follow naming convention

**DON'T:**
- ❌ Mix multiple features (unless testing integration)
- ❌ Use contrived variable names (x, y, z)
- ❌ Create giant test files (>50 lines)
- ❌ Use external dependencies
- ❌ Skip the .go.golden file

See [GOLDEN_TEST_GUIDELINES.md](GOLDEN_TEST_GUIDELINES.md) for complete rules.

## Related Documentation

- **Features:** `../../features/` - Feature specifications
- **Test Runner:** `../golden_test.go` - Test harness
- **Main CLAUDE.md:** `../../CLAUDE.md` - Project instructions

## Maintainers

When adding/updating tests:

1. ✅ Follow [GOLDEN_TEST_GUIDELINES.md](GOLDEN_TEST_GUIDELINES.md)
2. ✅ Update [INDEX.md](INDEX.md) with new tests
3. ✅ Ensure tests pass locally
4. ✅ Update this README if categories change
5. ✅ Document in `CHANGELOG.md` if significant

---

**Last Updated:** 2025-11-17
**Total Tests:** 46 (all complete with .go.golden files)
**Compilation Status:** ✅ All 46 .go.golden files compile successfully
**Maintained By:** Dingo Project Contributors

For questions, see [GOLDEN_TEST_GUIDELINES.md](GOLDEN_TEST_GUIDELINES.md) or open an issue.

# Golden Test Guidelines for Dingo

## Purpose

Golden tests verify that Dingo syntax correctly transpiles to idiomatic Go code. These tests serve as:
- **Regression tests** - Ensure transpiler output remains consistent
- **Documentation** - Show how Dingo features map to Go
- **Examples** - Demonstrate proper Dingo syntax usage

---

## File Naming Convention

### Format
```
{feature}_{NN}_{description}.dingo
{feature}_{NN}_{description}.go.golden
```

### Feature Prefixes

| Prefix | Feature | Example |
|--------|---------|---------|
| `error_prop_` | Error propagation (?) | `error_prop_01_simple.dingo` |
| `result_` | Result type | `result_01_basic.dingo` |
| `option_` | Option type | `option_01_basic.dingo` |
| `sum_types_` | Sum types/enums | `sum_types_01_simple.dingo` |
| `lambda_` | Lambda expressions | `lambda_01_basic.dingo` |
| `ternary_` | Ternary operator | `ternary_01_basic.dingo` |
| `null_coalesce_` | Null coalescing (??) | `null_coalesce_01_basic.dingo` |
| `safe_nav_` | Safe navigation (?.) | `safe_nav_01_basic.dingo` |
| `pattern_match_` | Pattern matching | `pattern_match_01_basic.dingo` |
| `tuples_` | Tuples | `tuples_01_basic.dingo` |
| `func_util_` | Functional utilities | `func_util_01_map.dingo` |
| `immutable_` | Immutability | `immutable_01_basic.dingo` |

### Numbering
- `01-09` - Zero-padded for proper sorting
- Start at `01` for each feature
- Number sequentially by complexity (simple → complex)

### Description
- Short, descriptive name (1-3 words)
- Use underscores: `basic`, `chained_calls`, `nested_types`
- Describe what's being tested, not implementation

---

## File Structure

### .dingo File (Source)

```dingo
package main

import "errors"

// Test: {Brief description of what this test demonstrates}
// Feature: {Feature name from features/ directory}
// Complexity: {basic|intermediate|advanced}

{Dingo code that demonstrates the feature}
```

**Rules:**
1. **Always include `package main`** (unless testing package-level features)
2. **Import only necessary packages** - Keep minimal
3. **Add descriptive comment header** - What's being tested
4. **Use realistic examples** - Not contrived code
5. **One primary feature per test** - Focus on clarity
6. **Keep tests small** - 10-30 lines max for basic, up to 50 for complex
7. **Use meaningful names** - `user`, `config`, `result`, not `x`, `y`, `z`

### .go.golden File (Expected Output)

```go
package main

import "errors"

{Generated Go code that the transpiler should produce}
```

**Rules:**
1. **Idiomatic Go** - Should look hand-written
2. **Properly formatted** - Use `gofmt` style
3. **Compilable** - Must pass `go/parser` (see TestGoldenFilesCompilation)
4. **No runtime overhead** - Pure Go, no helper libraries
5. **Preserve comments** - When appropriate
6. **Consistent style** - Follow Go conventions

---

## Test Complexity Levels

### Basic (01-03)
- **Purpose:** Introduce the feature
- **Scope:** Single usage, minimal context
- **Example:** Simple error propagation with one `?` operator

```dingo
// error_prop_01_simple.dingo
func readConfig(path string) ([]byte, error) {
    data := ReadFile(path)?
    return data, nil
}
```

### Intermediate (04-06)
- **Purpose:** Real-world usage patterns
- **Scope:** Multiple features combined, common scenarios
- **Example:** Error propagation with multiple statements

```dingo
// error_prop_04_wrapping.dingo
func loadConfig() (*Config, error) {
    data := ReadFile("config.json")?
    config := Parse(data)?
    return config, nil
}
```

### Advanced (07+)
- **Purpose:** Edge cases, complex interactions
- **Scope:** Nested usage, integration with other features
- **Example:** Error propagation with type conversions and chaining

```dingo
// error_prop_08_chained_calls.dingo
func processData() (string, error) {
    result := fetchData()?.transform()?.validate()?
    return result, nil
}
```

---

## Writing Guidelines

### DO ✅

1. **Test one feature thoroughly** per file
   ```dingo
   // Good: error_prop_01_simple.dingo - Tests ? operator
   func read(path string) ([]byte, error) {
       data := ReadFile(path)?
       return data, nil
   }
   ```

2. **Use realistic examples** from actual use cases
   ```dingo
   // Good: Real-world scenario
   func findUser(id int) Option {
       if id > 0 {
           return Option_Some("User" + string(id))
       }
       return Option_None()
   }
   ```

3. **Progress from simple to complex** within a feature
   - `01_basic` - Minimal working example
   - `02_multiple` - Multiple usages
   - `03_chained` - Combined with other features
   - `04_nested` - Complex nesting

4. **Include comments** explaining non-obvious behavior
   ```dingo
   // Demonstrates ? operator with type conversion
   func convert(val string) (int, error) {
       parsed := strconv.Atoi(val)?
       return parsed * 2, nil
   }
   ```

5. **Test edge cases** in advanced tests
   - Empty values
   - Nil handling
   - Type boundaries
   - Error conditions

6. **Make tests self-contained** - No external dependencies
   ```dingo
   // Good: Uses only stdlib
   import "errors"

   // Bad: Requires external package
   import "github.com/some/package"
   ```

### DON'T ❌

1. **Don't mix multiple features** unless testing integration
   ```dingo
   // Bad: Tests both ? and lambdas and Result type
   func process() Result {
       data := fetch()?
       return transform(data, |x| x * 2)
   }

   // Good: Separate into error_prop_* and lambda_* tests
   ```

2. **Don't use contrived variable names**
   ```dingo
   // Bad
   func foo(x int, y int) (int, error) {
       z := bar(x, y)?
       return z, nil
   }

   // Good
   func calculateTotal(price int, quantity int) (int, error) {
       total := multiply(price, quantity)?
       return total, nil
   }
   ```

3. **Don't create giant test files** (>50 lines)
   - Split into multiple numbered tests instead
   - Each test should be scannable at a glance

4. **Don't include incomplete features** in golden tests
   - Only test implemented and stable features
   - Use separate experimental directory for WIP

5. **Don't skip the .go.golden file**
   - Every .dingo must have corresponding .go.golden
   - Golden file is the source of truth for expected output

6. **Don't use random or generated data**
   - Tests must be deterministic
   - Same input = same output always

---

## Example Test Progression

Here's how to structure tests for a new feature (using Result type as example):

### result_01_basic.dingo
```dingo
// Minimal Result usage - constructor and check
enum Result {
    Ok(float64),
    Err(error),
}

func divide(a, b float64) Result {
    if b == 0.0 {
        return Result_Err(errors.New("division by zero"))
    }
    return Result_Ok(a / b)
}
```

### result_02_propagation.dingo
```dingo
// Result with ? operator propagation
func calculate(a, b float64) Result {
    ratio := divide(a, b)?
    doubled := multiply(ratio, 2)?
    return Result_Ok(doubled)
}
```

### result_03_pattern_match.dingo
```dingo
// Result with pattern matching
func handleResult(r Result) string {
    return match r {
        Result_Ok(val) => "Success: " + string(val),
        Result_Err(e) => "Error: " + e.Error(),
    }
}
```

### result_04_chaining.dingo
```dingo
// Result with method chaining (map, and_then)
func processData(input string) Result {
    return parseInput(input)
        .map(|x| x * 2)
        .and_then(|x| validate(x))
}
```

### result_05_go_interop.dingo
```dingo
// Result interoperating with Go (T, error) functions
func readAndParse(path string) Result {
    // Wraps Go function returning ([]byte, error)
    data := os.ReadFile(path)?
    parsed := json.Unmarshal(data)?
    return Result_Ok(parsed)
}
```

---

## Testing Checklist

Before adding a new golden test, verify:

- [ ] File follows naming convention: `{feature}_{NN}_{description}.dingo`
- [ ] Both `.dingo` and `.go.golden` files exist
- [ ] `.dingo` file includes package declaration
- [ ] `.dingo` file has descriptive comment header
- [ ] Code is realistic and self-contained
- [ ] Tests one primary feature (unless testing integration)
- [ ] `.go.golden` contains idiomatic Go code
- [ ] `.go.golden` compiles (passes `go/parser`)
- [ ] Test fits complexity level (basic/intermediate/advanced)
- [ ] Variable names are meaningful
- [ ] File is <50 lines (split if larger)
- [ ] No external dependencies (stdlib only)
- [ ] Test is deterministic (no random/time-based values)

---

## Generating .go.golden Files

### Manual Process

1. Write the `.dingo` file
2. Run the transpiler: `dingo build {file}.dingo`
3. Review generated Go code for correctness
4. Copy to `.go.golden` if output is correct
5. Format with `gofmt`: `gofmt -w {file}.go.golden`

### Automated Process

```bash
# Generate golden file from current transpiler output
cd tests
go test -v -run TestGoldenFiles/{test_name}

# If output looks correct, promote .actual to .golden
mv golden/{test_name}.go.actual golden/{test_name}.go.golden
```

### Updating Golden Files

When transpiler behavior changes intentionally:

```bash
# Regenerate all golden files
cd tests
go test -v -run TestGoldenFiles

# Review all .actual files
find golden -name "*.go.actual" -exec diff {} {}.golden \;

# If changes are correct, update
find golden -name "*.go.actual" -exec sh -c 'mv "$1" "${1%.actual}.golden"' _ {} \;
```

---

## Directory Structure

```
tests/golden/
├── INDEX.md                          # Catalog of all tests
├── GOLDEN_TEST_GUIDELINES.md        # This file
├── REORGANIZATION_PLAN.md           # Historical: reorganization plan
│
├── error_prop_01_simple.dingo
├── error_prop_01_simple.go.golden
├── error_prop_02_multiple.dingo
├── error_prop_02_multiple.go.golden
│
├── result_01_basic.dingo
├── result_01_basic.go.golden
│
├── option_01_basic.dingo
├── option_01_basic.go.golden
│
└── ... (other features)
```

**Files in golden directory:**
- `.dingo` - Dingo source files
- `.go.golden` - Expected Go output
- `.go.actual` - Temporary files from test runs (gitignored)
- `.reasoning.md` - Optional: Design decisions for complex tests
- `INDEX.md` - Test catalog and navigation
- `GOLDEN_TEST_GUIDELINES.md` - This document

---

## Integration with Test Suite

The test suite (`tests/golden_test.go`) automatically:

1. **Discovers** all `.dingo` files in `tests/golden/`
2. **Transpiles** each file using the generator
3. **Compares** output against corresponding `.go.golden`
4. **Verifies** generated code compiles (syntax check)
5. **Reports** differences with actual vs expected output

**Test execution:**
```bash
# Run all golden tests
cd tests
go test -v -run TestGoldenFiles

# Run specific test
go test -v -run TestGoldenFiles/error_prop_01_simple

# Check compilation only
go test -v -run TestGoldenFilesCompilation
```

---

## Best Practices Summary

### Quality Over Quantity
- Better to have 5 excellent tests than 20 mediocre ones
- Each test should teach something specific
- Remove redundant tests that don't add value

### Clarity Over Cleverness
- Simple, obvious code > complex, compact code
- Tests are documentation; optimize for readability
- Use real-world scenarios, not puzzles

### Consistency
- Follow the naming convention strictly
- Use the same style across all tests
- Maintain uniform complexity progression (01=basic, etc.)

### Maintenance
- Keep tests aligned with current transpiler output
- Update golden files when behavior changes (intentionally)
- Remove obsolete tests for deprecated features
- Document breaking changes in CHANGELOG.md

---

## Common Patterns

### Testing Error Propagation
```dingo
// Pattern: ? operator with (T, error) returns
func example(path string) (Data, error) {
    raw := readFile(path)?      // Propagates error
    parsed := parse(raw)?        // Propagates error
    return validate(parsed)?     // Propagates error
}
```

### Testing Option Types
```dingo
// Pattern: Option with None/Some variants
func findItem(id int) Option {
    if id > 0 {
        return Option_Some(items[id])
    }
    return Option_None()
}
```

### Testing Sum Types
```dingo
// Pattern: enum with variants
enum Status {
    Pending,           // Unit variant
    Active(string),    // Tuple variant
    Complete { code: int, msg: string },  // Struct variant
}
```

### Testing Pattern Matching
```dingo
// Pattern: match expression with exhaustive cases
func handle(s Status) string {
    return match s {
        Status_Pending => "waiting",
        Status_Active(id) => "running: " + id,
        Status_Complete{code, msg} => "done: " + msg,
    }
}
```

---

## Troubleshooting

### Test Fails: Output Mismatch
1. Check `.go.actual` file in `tests/golden/`
2. Compare with `.go.golden` to see differences
3. If actual output is correct → Update golden file
4. If actual output is wrong → Fix transpiler bug

### Test Fails: Compilation Error
1. Review `.go.golden` file syntax
2. Run `gofmt -w {file}.go.golden` to format
3. Ensure all imports are present
4. Check for Go syntax errors

### Test Skipped or Not Found
1. Verify file naming: `{feature}_{NN}_{desc}.dingo`
2. Ensure `.go.golden` file exists
3. Check file is in `tests/golden/` directory
4. Confirm no typos in filename

---

## Future Additions

When adding new features to Dingo:

1. **Create feature proposal** in `features/{feature}.md`
2. **Add golden tests** following this guide
3. **Start with 3-5 tests** covering basic → advanced
4. **Update INDEX.md** with new feature section
5. **Run full test suite** to ensure no regressions
6. **Update CHANGELOG.md** with new tests

---

**Last Updated:** 2025-11-17
**Version:** 1.0
**Applies To:** Dingo Transpiler Golden Tests

# Task F: Golden Tests for Unqualified Import Inference

**Date:** 2025-11-19
**Agent:** golang-tester
**Session:** 20251118-234501

---

## Test Plan

### Test Suite Overview

**Goal:** Create comprehensive golden tests to validate end-to-end unqualified import inference.

**Total Tests Planned:** 5 tests covering basic, negative, multiple imports, mixed usage, and cross-file scenarios.

### Test Files Created

#### 1. unqualified_import_01_basic.dingo
- **Purpose:** Basic unqualified stdlib function call
- **Feature:** Transform `ReadFile(path)` → `os.ReadFile(path)` + add import
- **Complexity:** Basic
- **Expected:**
  - Unqualified `ReadFile` becomes qualified `os.ReadFile`
  - `import "os"` added to imports section

#### 2. unqualified_import_02_local_function.dingo
- **Purpose:** User-defined function should NOT be transformed
- **Feature:** Negative test (exclusion)
- **Complexity:** Basic
- **Expected:**
  - User's `ReadFile` function NOT transformed
  - No `os` import added
  - Validates FunctionExclusionCache works correctly

#### 3. unqualified_import_03_multiple.dingo
- **Purpose:** Multiple unqualified stdlib calls from different packages
- **Feature:** Multi-package import inference
- **Complexity:** Intermediate
- **Expected:**
  - `ReadFile` → `os.ReadFile` + import "os"
  - `Atoi` → `strconv.Atoi` + import "strconv"
  - `Printf` → `fmt.Printf` + import "fmt"
  - All three imports added

#### 4. unqualified_import_04_mixed.dingo
- **Purpose:** Mix of qualified and unqualified calls
- **Feature:** Already-qualified detection
- **Complexity:** Intermediate
- **Expected:**
  - Unqualified `ReadFile` → `os.ReadFile`
  - Unqualified `Printf` → `fmt.Printf`
  - `errors.New` NOT transformed (already qualified)
  - Imports added: "os", "fmt"
  - Import preserved: "errors"

#### 5. unqualified_import_05_cross_file/ (Package)
- **Purpose:** Cross-file package scanning
- **Feature:** PackageScanner detects functions across files
- **Complexity:** Advanced
- **Files:**
  - `helpers.dingo`: Defines `ProcessData()` and `ValidateInput()`
  - `main.dingo`: Calls those functions + stdlib calls
- **Expected:**
  - `ReadFile` → `os.ReadFile` (stdlib)
  - `Println` → `fmt.Println` (stdlib)
  - `ProcessData` NOT transformed (local, in helpers.dingo)
  - `ValidateInput` NOT transformed (local, in helpers.dingo)

---

## Test Execution Plan

### Phase 1: Unit Test Validation ✅
- Run `go test ./pkg/preprocessor -run TestUnqualified -v`
- **Result:** All 8 unit tests passing
- **Coverage:**
  - Basic transformation
  - Local function exclusion
  - Ambiguous error
  - Multiple imports
  - Already qualified
  - Mixed qualified/unqualified
  - No stdlib
  - Only local functions

### Phase 2: Integration Test ⚠️ BLOCKED
- Build dingo compiler with unqualified import processor
- Transpile each golden test file
- Compare output with expected `.go.golden`

**Current Status:**
- ✅ Compiler builds successfully
- ✅ Transpilation runs without errors
- ❌ Unqualified imports NOT being transformed
- ❌ No imports added to output

**Root Cause:** Investigation needed (see Task-F-status.txt)

### Phase 3: Golden Test Suite
- Run `go test ./tests -run TestGoldenFiles/unqualified -v`
- Verify all outputs match `.go.golden` files
- Check compilation success

---

## Expected Golden Outputs

### Test 01: Basic (Expected)

```go
package main

import "os"

// Test: Basic unqualified stdlib function call
// Feature: Unqualified import inference
// Complexity: basic

func readConfig(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return data
}

func main() {
	config := readConfig("config.txt")
	println(string(config))
}
```

### Test 02: Local Function (Expected)

```go
package main

// Test: User-defined function should NOT be transformed
// Feature: Unqualified import inference (negative test)
// Complexity: basic

// User's own ReadFile function
func ReadFile(path string) string {
	return "custom implementation for " + path
}

func main() {
	// Should use user's ReadFile, NOT os.ReadFile
	data := ReadFile("test.txt")
	println(data)
}
```

### Test 03: Multiple Imports (Expected)

```go
package main

import (
	"fmt"
	"os"
	"strconv"
)

// Test: Multiple unqualified stdlib calls from different packages
// Feature: Unqualified import inference
// Complexity: intermediate

func processInput(path string, numStr string) {
	// Should add: import "os"
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	// Should add: import "strconv"
	num, err := strconv.Atoi(numStr)
	if err != nil {
		panic(err)
	}

	// Should add: import "fmt"
	fmt.Printf("Read %d bytes, number is %d\n", len(data), num)
}

func main() {
	processInput("data.txt", "42")
}
```

### Test 04: Mixed (Expected)

```go
package main

import (
	"errors"
	"fmt"
	"os"
)

// Test: Mix of qualified and unqualified calls
// Feature: Unqualified import inference
// Complexity: intermediate

func validateFile(path string) error {
	// Unqualified - should transform to os.ReadFile
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		// Qualified - should NOT transform
		return errors.New("empty file")
	}

	// Unqualified - should transform to fmt.Printf
	fmt.Printf("File size: %d bytes\n", len(data))

	return nil
}

func main() {
	err := validateFile("test.txt")
	if err != nil {
		panic(err)
	}
}
```

### Test 05: Cross-File (Expected)

**helpers.go:**
```go
package main

// Test: Cross-file package scanning - helper functions
// Feature: Unqualified import inference (package-wide scanning)
// Complexity: advanced

// User-defined function in separate file
func ProcessData(data []byte) string {
	return string(data) + " (processed)"
}

// Another local function
func ValidateInput(s string) bool {
	return len(s) > 0
}
```

**main.go:**
```go
package main

import (
	"fmt"
	"os"
)

// Test: Cross-file package scanning - main file
// Feature: Unqualified import inference (package-wide scanning)
// Complexity: advanced

func main() {
	// Unqualified stdlib call - should transform to os.ReadFile
	data, err := os.ReadFile("test.txt")
	if err != nil {
		panic(err)
	}

	// Unqualified local call from helpers.dingo - should NOT transform
	processed := ProcessData(data)

	// Another local call - should NOT transform
	if ValidateInput(processed) {
		// Unqualified stdlib call - should transform to fmt.Println
		fmt.Println(processed)
	}
}
```

---

## Success Criteria

### Functional Requirements
- ✅ Unit tests: All 8 passing
- ⚠️ Integration: Processor runs but doesn't transform
- ⏳ Golden tests: Blocked pending integration fix

### Test Coverage
- ✅ Basic transformation
- ✅ Local function exclusion
- ✅ Ambiguous function errors
- ✅ Multiple packages
- ✅ Already-qualified detection
- ✅ Mixed usage
- ⏳ Cross-file scanning (pending)

### Code Quality
- ✅ Test files follow naming convention
- ✅ Realistic, self-contained examples
- ✅ Clear comments and documentation
- ✅ Progressive complexity (basic → advanced)

---

## Blockers

### Integration Issue

**Problem:** Unqualified imports not being transformed in end-to-end flow

**Evidence:**
1. Unit tests pass (processor works in isolation)
2. Compiler builds and runs successfully
3. Transpilation completes without errors
4. But output shows NO transformations applied

**Hypothesis:**
1. Cache may not be populating correctly
2. Processor may not be receiving cache
3. Early bailout may be triggering incorrectly
4. Processor order in pipeline may be wrong

**Next Steps:**
1. Add debug logging to verify processor is being called
2. Check cache contents after scanning
3. Verify PackageContext is creating cache correctly
4. Test with verbose mode to see cache statistics

---

## Timeline

- **Test Creation:** 30 minutes ✅
- **Integration:** 45 minutes ⚠️ (blocked)
- **Debugging:** TBD (investigation needed)
- **Golden Test Validation:** 15 minutes ⏳ (pending)

**Total Estimated:** 1.5 hours + debugging time

---

## Next Actions

1. **PRIORITY:** Investigate why transformations aren't being applied
2. Add verbose logging to transpilation pipeline
3. Verify cache scanning works correctly
4. Test processor in isolation with cache
5. Once working: Generate `.go.golden` files
6. Run golden test suite
7. Document any failures

---

## References

- Implementation plan: `final-plan-v2.md`
- Unit tests: `pkg/preprocessor/unqualified_imports_test.go`
- Processor: `pkg/preprocessor/unqualified_imports.go`
- Cache: `pkg/preprocessor/function_cache.go`
- Package scanner: `pkg/preprocessor/package_scanner.go`

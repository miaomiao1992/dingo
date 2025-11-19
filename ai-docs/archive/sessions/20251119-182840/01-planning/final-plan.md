# Final Architecture Plan: Fix Where Guard Preprocessor Bug
## Incorporating User Decisions

### User Decisions Integrated
1. ✅ Support guards at ALL nesting levels (not just innermost)
2. ✅ Enforce maximum nesting depth of 2 levels for guards
3. ✅ Create new test: `pattern_match_13_nested_guards_simple.dingo`

---

## Root Cause Analysis

### The Bug
Test `pattern_match_06_guards_nested.dingo` fails with:
```
parse error: line 103:4: expected ';', found 'else'
```

### Root Cause
File: `/Users/jack/mag/dingo/pkg/preprocessor/rust_match.go`
Function: `generateCaseWithGuards` (lines 773-981)

**The Problem:**
1. Nested pattern detection logic at line 832: `hasNestedPatterns = bindingsDiffer && !hasGuards`
2. This EXCLUDES nested patterns that have guards
3. When nested patterns have guards, code falls through to simple guard generation
4. Simple guard generator doesn't handle nested binding extraction (e.g., `Option_Some(val)`)
5. Guard references undefined variable → Go parse error

**Example:**
```dingo
Result_Ok(Option_Some(val)) where val > 0 => "positive"
```
- Has nested pattern: `Result_Ok(Option_Some(val))`
- Has guard: `where val > 0`
- Current code skips nested pattern handling
- Guard tries to use `val` which was never extracted
- Result: Invalid Go syntax

---

## Proposed Solution

### Strategy: Multi-Level Guard Support with Depth Validation

**Three-Pronged Approach:**
1. Fix nested pattern detection to allow guards
2. Add guard support to nested pattern code generation (all levels)
3. Add depth validation to enforce 2-level limit

---

## Implementation Steps

### Step 1: Fix Nested Pattern Detection Logic

**File:** `pkg/preprocessor/rust_match.go`
**Location:** Line 832 in function `generateCaseWithGuards`

**Change:**
```go
// BEFORE (BUGGY):
hasNestedPatterns = bindingsDiffer && !hasGuards

// AFTER (FIXED):
hasNestedPatterns = bindingsDiffer
// Remove the "&& !hasGuards" condition entirely
// Nested patterns CAN have guards - handle them inside nested switch
```

**Rationale:** Nested patterns (structural) and guards (conditional) are orthogonal concerns. They should be handled independently.

---

### Step 2: Add Nesting Depth Validation

**File:** `pkg/preprocessor/rust_match.go`
**Location:** In `generateCaseWithGuards`, after pattern grouping (around line 650)

**New Function:**
```go
// validateNestedPatternDepth checks nesting depth and errors if > 2
func (r *RustMatchProcessor) validateNestedPatternDepth(pattern string, maxDepth int) error {
    depth := r.calculatePatternDepth(pattern)
    if depth > maxDepth {
        return fmt.Errorf("pattern nesting depth %d exceeds maximum %d: %s",
                          depth, maxDepth, pattern)
    }
    return nil
}

// calculatePatternDepth counts nesting levels in a pattern
func (r *RustMatchProcessor) calculatePatternDepth(pattern string) int {
    depth := 0
    currentDepth := 0

    for _, ch := range pattern {
        if ch == '(' {
            currentDepth++
            if currentDepth > depth {
                depth = currentDepth
            }
        } else if ch == ')' {
            currentDepth--
        }
    }

    return depth
}
```

**Integration Point:**
Add validation when processing arms with guards:
```go
for _, group := range patternGroups {
    for _, arm := range group.arms {
        if arm.guard != "" {
            // NEW: Validate depth before processing
            if err := r.validateNestedPatternDepth(arm.pattern, 2); err != nil {
                return "", err
            }
        }
    }

    // Continue with existing generation logic...
}
```

**Error Example:**
```dingo
// This would error (3 levels):
Result_Ok(Option_Some(Either_Left(val))) where val > 0 => ...

// Error message:
pattern nesting depth 3 exceeds maximum 2: Result_Ok(Option_Some(Either_Left(val)))
```

---

### Step 3: Add Guard Support to Nested Pattern Generation (All Levels)

**File:** `pkg/preprocessor/rust_match.go`
**Location:** Lines 836-883 (nested pattern case generation)

**Current Code Flow:**
1. Generate intermediate variable for outer value
2. Generate nested switch on inner patterns
3. For each inner case: Extract binding, execute expression
4. **MISSING**: No guard checks!

**Enhanced Code Flow:**
1. Generate intermediate variable for outer value
2. Check for guards on OUTER pattern (NEW)
3. Generate nested switch on inner patterns
4. For each inner case:
   - Extract binding (existing)
   - Check for guards on INNER pattern (NEW)
   - Wrap in `if guard { ... }` if guard present
   - Execute expression
5. Handle multiple arms with same pattern but different guards (if-else-if chain)

**Detailed Changes:**

#### 3.1: Outer Level Guard Support

**Location:** Lines 836-850 (after outer case generation)

**Add:**
```go
// After: buf.WriteString(fmt.Sprintf("\tcase %s:\n", outerTag))

// NEW: Check if ANY arms have outer-level guards
outerGuards := []string{}
for _, arm := range group.arms {
    if guard := r.extractOuterGuard(arm.pattern, arm.guard); guard != "" {
        outerGuards = append(outerGuards, guard)
    }
}

if len(outerGuards) > 0 {
    // Generate outer guard check before nested switch
    for idx, guard := range outerGuards {
        if idx == 0 {
            buf.WriteString(fmt.Sprintf("\t\tif %s {\n", guard))
        } else {
            buf.WriteString(fmt.Sprintf("\t\t} else if %s {\n", guard))
        }
        // Nested switch goes inside this if
        indent = "\t\t\t"
    }
} else {
    indent = "\t\t"
}
```

**New Helper Function:**
```go
// extractOuterGuard checks if guard applies to outer pattern
// Example: Result_Ok(x) where result.IsOk() => returns "result.IsOk()"
// Example: Result_Ok(Option_Some(x)) where x > 0 => returns "" (applies to inner)
func (r *RustMatchProcessor) extractOuterGuard(pattern string, guard string) string {
    // If guard references outer pattern binding only, it's an outer guard
    // Otherwise, it's an inner guard

    // Simple heuristic: if guard references innermost binding, it's inner
    innerBinding := r.extractInnermostBinding(pattern)
    if strings.Contains(guard, innerBinding) {
        return "" // Inner guard
    }

    return guard // Outer guard
}

// extractInnermostBinding gets the deepest binding in a pattern
// Example: Result_Ok(Option_Some(val)) => "val"
func (r *RustMatchProcessor) extractInnermostBinding(pattern string) string {
    // Find rightmost identifier before closing paren
    re := regexp.MustCompile(`\(([a-z_][a-z0-9_]*)\)[^(]*$`)
    matches := re.FindStringSubmatch(pattern)
    if len(matches) > 1 {
        return matches[1]
    }
    return ""
}
```

#### 3.2: Inner Level Guard Support

**Location:** Lines 862-880 (inner pattern case generation)

**Enhanced Logic:**
```go
for _, innerPattern := range sortedInner {
    innerArms := innerGroups[innerPattern]
    innerTag := r.getTagName(innerPattern)
    buf.WriteString(fmt.Sprintf("%scase %s:\n", indent, innerTag))

    // Group inner arms by guard
    guardGroups := r.groupArmsByGuard(innerArms)

    for idx, guardGroup := range guardGroups {
        arm := guardGroup[0] // Representative arm

        // Extract innermost binding
        if arm.binding != "" && arm.binding != "_" {
            bindingCode := r.generateBinding(intermediateVar, arm.pattern, arm.binding)
            buf.WriteString(fmt.Sprintf("%s\t%s\n", indent, bindingCode))
        }

        // NEW: Handle guards
        if arm.guard != "" {
            // Check if this is an inner-level guard
            innerGuard := r.extractInnerGuard(arm.pattern, arm.guard)

            if innerGuard != "" {
                // Guard present: wrap in if statement
                if idx == 0 {
                    buf.WriteString(fmt.Sprintf("%s\tif %s {\n", indent, innerGuard))
                } else {
                    buf.WriteString(fmt.Sprintf("%s\t} else if %s {\n", indent, innerGuard))
                }
                exprIndent := indent + "\t\t"
            } else {
                exprIndent := indent + "\t"
            }
        } else {
            // No guard: else clause or standalone
            if idx > 0 {
                buf.WriteString(fmt.Sprintf("%s\t} else {\n", indent))
                exprIndent = indent + "\t\t"
            } else {
                exprIndent = indent + "\t"
            }
        }

        // Expression
        if isInAssignment && assignmentVar != "" {
            buf.WriteString(fmt.Sprintf("%s%s = %s\n", exprIndent, assignmentVar, arm.expression))
        } else {
            buf.WriteString(fmt.Sprintf("%s%s\n", exprIndent, arm.expression))
        }

        // Close guard if statement
        if idx > 0 || arm.guard != "" {
            buf.WriteString(fmt.Sprintf("%s\t}\n", indent))
        }
    }
}

// Close outer guard if statements
if len(outerGuards) > 0 {
    buf.WriteString("\t\t}\n")
}
```

**New Helper Function:**
```go
// extractInnerGuard returns guard if it applies to innermost binding
func (r *RustMatchProcessor) extractInnerGuard(pattern string, guard string) string {
    innerBinding := r.extractInnermostBinding(pattern)
    if innerBinding != "" && strings.Contains(guard, innerBinding) {
        return guard // Inner guard
    }
    return "" // Outer guard
}

// groupArmsByGuard groups arms by their guards for if-else-if chains
func (r *RustMatchProcessor) groupArmsByGuard(arms []matchArm) [][]matchArm {
    groups := [][]matchArm{}
    guardToIndex := make(map[string]int)

    for _, arm := range arms {
        key := arm.guard
        if idx, exists := guardToIndex[key]; exists {
            groups[idx] = append(groups[idx], arm)
        } else {
            guardToIndex[key] = len(groups)
            groups = append(groups, []matchArm{arm})
        }
    }

    return groups
}
```

---

### Step 4: Create New Golden Test

**File:** `tests/golden/pattern_match_13_nested_guards_simple.dingo`

**Purpose:** Isolate nested-pattern-with-guard feature for regression testing

**Content:**
```dingo
package main

import "fmt"

func main() {
    // Test 1: Basic nested pattern with guard
    result1 := Result_Ok(Option_Some(42))
    msg1 := match result1 {
        Result_Ok(Option_Some(val)) where val > 0 => "positive",
        Result_Ok(Option_Some(val)) where val < 0 => "negative",
        Result_Ok(Option_Some(_)) => "zero",
        Result_Ok(Option_None) => "none",
        Result_Err(_) => "error",
    }
    fmt.Println(msg1) // Expected: "positive"

    // Test 2: Multiple guards on same nested pattern
    result2 := Result_Ok(Option_Some(150))
    msg2 := match result2 {
        Result_Ok(Option_Some(x)) where x > 100 => "large",
        Result_Ok(Option_Some(x)) where x > 10 => "medium",
        Result_Ok(Option_Some(x)) => "small",
        Result_Err(_) => "error",
    }
    fmt.Println(msg2) // Expected: "large"

    // Test 3: Nested pattern without guard (existing functionality)
    result3 := Result_Ok(Option_None)
    msg3 := match result3 {
        Result_Ok(Option_Some(v)) => fmt.Sprintf("value: %d", v),
        Result_Ok(Option_None) => "no value",
        Result_Err(_) => "error",
    }
    fmt.Println(msg3) // Expected: "no value"

    // Test 4: Guard at outer level (should also work)
    result4 := Result_Ok(Option_Some(10))
    msg4 := match result4 {
        Result_Ok(opt) where opt.IsSome() => "has value",
        Result_Ok(_) => "no value",
        Result_Err(_) => "error",
    }
    fmt.Println(msg4) // Expected: "has value"
}
```

**Golden Output:** Generate via:
```bash
# After implementing fix:
go test ./tests -run TestGoldenFiles/pattern_match_13_nested_guards_simple -update
```

**Test Validation:**
```bash
# Should pass after fix:
go test ./tests -run TestGoldenFiles/pattern_match_13_nested_guards_simple -v
```

---

### Step 5: Update Existing Test

**File:** `tests/golden/pattern_match_06_guards_nested.dingo`

**Action:** Should pass without modification after implementing Steps 1-3

**Validation:**
```bash
go test ./tests -run TestGoldenFiles/pattern_match_06_guards_nested -v
```

---

## Files to Modify

### 1. Primary Implementation
**File:** `/Users/jack/mag/dingo/pkg/preprocessor/rust_match.go`

**Changes:**
- Line 832: Remove `&& !hasGuards` condition from `hasNestedPatterns` check
- After line 650: Add depth validation (new functions)
- Lines 836-850: Add outer-level guard support
- Lines 862-880: Add inner-level guard support with if-else-if chains
- New helper functions:
  - `validateNestedPatternDepth`
  - `calculatePatternDepth`
  - `extractOuterGuard`
  - `extractInnerGuard`
  - `extractInnermostBinding`
  - `groupArmsByGuard`

### 2. New Test File
**File:** `/Users/jack/mag/dingo/tests/golden/pattern_match_13_nested_guards_simple.dingo`
- Create new test file (see Step 4)

**File:** `/Users/jack/mag/dingo/tests/golden/pattern_match_13_nested_guards_simple.go.golden`
- Generate via `-update` flag after implementation

---

## Test Coverage

### Tests That Should Pass After Fix

1. ✅ `pattern_match_05_guards_basic.dingo` (already passing)
   - Simple guards without nesting

2. ✅ `pattern_match_06_guards_nested.dingo` (currently failing → should pass)
   - Complex nested patterns with guards (the bug fix target)

3. ✅ `pattern_match_13_nested_guards_simple.dingo` (new test)
   - Simple nested patterns with guards (regression prevention)

4. ✅ All existing pattern match tests (regression check)
   - `pattern_match_01_basic.dingo`
   - `pattern_match_02_guards.dingo`
   - `pattern_match_03_nested.dingo`
   - `pattern_match_04_exhaustive.dingo`
   - `pattern_match_12_tuple_exhaustiveness.dingo`

### Test Scenarios Covered

1. **Basic nested + guard:** ✅
   ```dingo
   Result_Ok(Option_Some(val)) where val > 0 => ...
   ```

2. **Multiple guards on same nested pattern:** ✅
   ```dingo
   Result_Ok(Option_Some(x)) where x > 100 => "large"
   Result_Ok(Option_Some(x)) where x > 10 => "medium"
   ```

3. **Guard at outer level:** ✅
   ```dingo
   Result_Ok(opt) where opt.IsSome() => ...
   ```

4. **Nested without guard (existing):** ✅
   ```dingo
   Result_Ok(Option_Some(val)) => ...
   ```

5. **Depth validation (error case):** ✅
   ```dingo
   // Should error: depth = 3
   Result_Ok(Option_Some(Either_Left(x))) where x > 0 => ...
   ```

---

## Expected Generated Code Examples

### Example 1: Simple Nested Pattern with Guard

**Input (Dingo):**
```dingo
match result {
    Result_Ok(Option_Some(val)) where val > 0 => "positive",
    Result_Ok(Option_Some(_)) => "non-positive",
    Result_Err(_) => "error",
}
```

**Output (Go) - After Fix:**
```go
var __match_result_0 interface{}
__match_0 := result
switch __match_0.tag {
case ResultTag_Ok:
    __Result_Ok_nested := *__match_0.ok_0
    switch __Result_Ok_nested.tag {
    case OptionTag_Some:
        val := *__Result_Ok_nested.some_0

        if val > 0 {                    // ← NEW: Inner guard
            __match_result_0 = "positive"
        } else {                        // ← NEW: No-guard arm
            __match_result_0 = "non-positive"
        }
    case OptionTag_None:
        __match_result_0 = "none"
    }
case ResultTag_Err:
    __match_result_0 = "error"
}
```

### Example 2: Outer + Inner Guards

**Input (Dingo):**
```dingo
match result {
    Result_Ok(opt) where result.IsOk() => "outer guard passed",
    Result_Ok(Option_Some(x)) where x > 0 => "inner guard passed",
    _ => "no match",
}
```

**Output (Go) - After Fix:**
```go
var __match_result_0 interface{}
__match_0 := result
switch __match_0.tag {
case ResultTag_Ok:
    if result.IsOk() {              // ← NEW: Outer guard
        __match_result_0 = "outer guard passed"
    } else {
        __Result_Ok_nested := *__match_0.ok_0
        switch __Result_Ok_nested.tag {
        case OptionTag_Some:
            x := *__Result_Ok_nested.some_0

            if x > 0 {              // ← NEW: Inner guard
                __match_result_0 = "inner guard passed"
            }
        }
    }
default:
    __match_result_0 = "no match"
}
```

---

## Potential Risks & Mitigations

### Risk 1: Regression in Nested Patterns
**What:** Changing nested pattern detection might break existing tests

**Mitigation:**
- Run full test suite after change
- Test files to monitor:
  - `pattern_match_03_nested.dingo`
  - `pattern_match_04_exhaustive.dingo`
  - `pattern_match_12_tuple_exhaustiveness.dingo`

### Risk 2: Guard Extraction Logic Complexity
**What:** Distinguishing outer vs inner guards may have edge cases

**Mitigation:**
- Use simple heuristic (check if guard references innermost binding)
- Add comprehensive test cases in `pattern_match_13_nested_guards_simple.dingo`
- Document assumptions in code comments

### Risk 3: Indentation Management
**What:** 3-4 levels of nesting create complex indentation tracking

**Mitigation:**
- Use consistent indent variable tracking (follow existing pattern)
- Test with deeply nested cases
- Validate generated Go syntax compiles

### Risk 4: Depth Validation False Positives
**What:** Pattern depth calculation may miscount in edge cases

**Mitigation:**
- Use simple paren-counting algorithm
- Test with various pattern structures
- Allow disabling via config if needed (future)

---

## Success Criteria

### Functional Requirements
1. ✅ `pattern_match_06_guards_nested.dingo` compiles without errors
2. ✅ `pattern_match_13_nested_guards_simple.dingo` passes all test cases
3. ✅ Generated Go code is syntactically valid and compiles
4. ✅ All existing pattern match tests remain passing (no regressions)

### Quality Requirements
5. ✅ Depth validation errors on >2 levels
6. ✅ Code follows existing style and patterns
7. ✅ Guard logic works at both outer and inner levels
8. ✅ Multiple guards on same pattern generate if-else-if chains

### Test Count
- **Before Fix:** 102/103 passing (1 failing: `pattern_match_06_guards_nested`)
- **After Fix:** 104/104 passing (includes new `pattern_match_13_nested_guards_simple`)

---

## Implementation Complexity

**Estimated Complexity:** Medium-High

**Breakdown:**
- **Step 1 (Detection Fix):** Simple - 1 line change (10 min)
- **Step 2 (Depth Validation):** Medium - 50 lines, new functions (1 hour)
- **Step 3 (Guard Support):** Complex - 150 lines, logic changes (3 hours)
  - Outer guard support: 30 lines (1 hour)
  - Inner guard support: 80 lines (1.5 hours)
  - Helper functions: 40 lines (30 min)
- **Step 4 (New Test):** Simple - Create test file (30 min)
- **Step 5 (Validation):** Simple - Run tests (15 min)

**Total Time Estimate:** 4.5-5 hours
- Implementation: 3.5 hours
- Testing & debugging: 1 hour
- Validation & documentation: 30 min

---

## Alternative Approaches Considered

### Alternative 1: Separate Code Paths for Guarded Nested Patterns
**Idea:** Create entirely separate handling for "nested patterns with guards"

**Pros:**
- Clear separation of concerns
- Easier to reason about individual cases

**Cons:**
- Duplicates 200+ lines of nested switch generation code
- Violates DRY principle
- Harder to maintain consistency

**Decision:** Rejected - Code duplication outweighs benefits

### Alternative 2: Flatten Nested Patterns Before Guard Processing
**Idea:** Transform nested patterns into multiple sequential match expressions

**Pros:**
- Simpler guard processing
- Reuses existing simple guard logic

**Cons:**
- Changes semantics (multiple matches vs single match)
- Loses performance (multiple switches)
- Harder to maintain
- Doesn't match user's mental model

**Decision:** Rejected - Semantic changes and performance loss

### Alternative 3: Disable Guards on Nested Patterns
**Idea:** Report compile error if user tries to use guards with nested patterns

**Pros:**
- No implementation complexity
- Clear limitation

**Cons:**
- Artificial limitation
- User-hostile (feature should work naturally)
- Doesn't match Rust/Swift behavior (precedent)

**Decision:** Rejected - Feature should work as users expect

### Alternative 4: Support Only Inner-Level Guards
**Idea:** Allow guards only on innermost patterns, error on outer-level guards

**Pros:**
- Simpler implementation
- Most common use case

**Cons:**
- Artificial limitation (User Decision #1 requires all levels)
- Doesn't match precedent (Rust allows all levels)
- Confusing to users

**Decision:** Rejected - User explicitly requested all-level support

---

## Implementation Checklist

### Pre-Implementation
- [ ] Read current implementation (`rust_match.go:773-981`)
- [ ] Understand pattern grouping logic (line 650)
- [ ] Review guard generation pattern (lines 936-978)

### Implementation Phase
- [ ] **Step 1:** Fix nested pattern detection (line 832)
- [ ] **Step 2:** Add depth validation functions
- [ ] **Step 2:** Integrate depth validation in arm processing
- [ ] **Step 3.1:** Add outer-level guard support
- [ ] **Step 3.2:** Add inner-level guard support
- [ ] **Step 3.3:** Implement helper functions (6 functions)
- [ ] **Step 4:** Create `pattern_match_13_nested_guards_simple.dingo`

### Testing Phase
- [ ] Run new test: `pattern_match_13_nested_guards_simple`
- [ ] Run existing test: `pattern_match_06_guards_nested`
- [ ] Run all pattern match tests (regression check)
- [ ] Validate depth validation errors correctly
- [ ] Check generated Go code compiles
- [ ] Generate golden output for new test

### Validation Phase
- [ ] Verify test count: 104/104 passing
- [ ] Code review for style consistency
- [ ] Validate indentation in generated code
- [ ] Confirm no regressions in existing tests

### Documentation Phase
- [ ] Update `CHANGELOG.md` with fix details
- [ ] Add comments to new functions
- [ ] Document guard extraction logic assumptions

---

## Conclusion

This plan addresses the root cause (nested pattern detection bug), implements the requested features (multi-level guards, depth validation), and adds appropriate test coverage. The solution reuses existing patterns, maintains code quality, and follows user requirements.

**Key Improvements Over Initial Plan:**
1. ✅ Supports guards at ALL levels (not just innermost)
2. ✅ Enforces 2-level depth limit for readability
3. ✅ Creates dedicated simple test for regression prevention
4. ✅ Provides clear implementation steps with code examples
5. ✅ Includes detailed guard extraction logic

**Implementation Priority:** High (fixes failing test, enables key feature)

**Confidence Level:** High (clear root cause, straightforward solution, good test coverage)

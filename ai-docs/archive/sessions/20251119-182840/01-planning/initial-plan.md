# Architecture Plan: Fix Where Guard Preprocessor Bug

## Root Cause Analysis

### The Bug
Test `pattern_match_06_guards_nested.dingo` fails with:
```
parse error: line 103:4: expected ';', found 'else'
```

### Root Cause (IDENTIFIED)

The bug is in `/Users/jack/mag/dingo/pkg/preprocessor/rust_match.go`, function `generateCaseWithGuards` (lines 773-981).

**The Problem:**
1. When processing pattern arms, the preprocessor groups arms by pattern (line 652)
2. It detects "nested patterns" vs "normal patterns with guards" (lines 820-833)
3. **THE BUG**: If it detects nested patterns (line 835), it generates a nested switch statement and **returns early** at line 906
4. This early return **skips all guard generation logic** (lines 909-978)

**Why it happens:**
```go
// Line 819-833: Nested pattern detection
hasNestedPatterns := false
if len(group.arms) > 1 {
    hasGuards := false
    bindingsDiffer := false
    for i, arm := range group.arms {
        if arm.guard != "" {
            hasGuards = true  // ← Detects guards
        }
        if i > 0 && arm.binding != firstArm.binding && r.isNestedPatternBinding(arm.binding) {
            bindingsDiffer = true  // ← Detects nested patterns like Option_Some(val)
        }
    }
    hasNestedPatterns = bindingsDiffer && !hasGuards  // ← BUG: Ignores guards if bindings differ!
}
```

**The Logic Error:**
- `hasNestedPatterns = bindingsDiffer && !hasGuards`
- This means: "Only treat as nested pattern if NO guards present"
- **BUT**: The test has BOTH nested patterns AND guards:
  ```dingo
  Result_Ok(Option_Some(val)) where val > 0 => "positive"
  ```
- When `bindingsDiffer = true` (nested pattern) AND `hasGuards = true` (where clause):
  - `hasNestedPatterns = false` (line 832)
  - Code falls through to normal guard generation (line 885)
  - **BUT**: Guard generation expects simple binding like `x`, not nested `Option_Some(val)`
  - This creates invalid Go syntax

**Generated Code Bug:**
The guard generation (line 936-950) generates:
```go
if val > 0 {  // ← 'val' is NEVER DEFINED because binding extraction skipped nested pattern
    __match_result_0 = "positive"
}
```

This creates a syntax error because:
1. Nested pattern binding `Option_Some(val)` is not extracted properly
2. Guard tries to reference `val` which doesn't exist
3. Go parser sees malformed code and reports "expected ';', found 'else'"

### Why `if` Guards Work But `where` Guards Fail

**They both fail the same way** - it's not `where` vs `if`. The test just happens to use `where` for nested patterns.

**Proof:**
- `pattern_match_05_guards_basic.dingo`: Uses `if` guards, NO nested patterns → ✅ Works
- `pattern_match_06_guards_nested.dingo`: Uses `where` guards, HAS nested patterns → ❌ Fails

If we used `if` with nested patterns, it would fail identically.

## Proposed Solution

### Approach: Support Guards on Nested Patterns

**Strategy:** Modify guard generation to handle nested patterns correctly.

**Two Sub-Cases to Handle:**

#### Case 1: Nested Pattern WITHOUT Guard
```dingo
Result_Ok(Option_Some(val)) => "value"
```
Current code handles this correctly (lines 836-883).

#### Case 2: Nested Pattern WITH Guard
```dingo
Result_Ok(Option_Some(val)) where val > 0 => "positive"
```
**THIS IS THE BUG.** Need to add guard support to nested pattern code path.

### Implementation Steps

#### Step 1: Fix Nested Pattern Detection Logic

**File:** `pkg/preprocessor/rust_match.go`
**Function:** `generateCaseWithGuards` (line 773)
**Line:** 832

**Change:**
```go
// BEFORE (BUGGY):
hasNestedPatterns = bindingsDiffer && !hasGuards

// AFTER (FIXED):
hasNestedPatterns = bindingsDiffer
// Remove the "&& !hasGuards" condition entirely
// Nested patterns CAN have guards - handle them inside nested switch
```

**Rationale:** Nested patterns are structural (about binding format), guards are conditional (runtime checks). They are orthogonal concerns and should be handled independently.

#### Step 2: Add Guard Support to Nested Pattern Generation

**File:** `pkg/preprocessor/rust_match.go`
**Function:** `generateCaseWithGuards`
**Lines:** 836-883 (nested pattern case)

**Current Code Flow:**
1. Generate intermediate variable for outer value
2. Generate nested switch on inner patterns
3. For each inner case: Extract binding, execute expression
4. **MISSING**: No guard checks!

**Modified Code Flow:**
1. Generate intermediate variable for outer value
2. Generate nested switch on inner patterns
3. For each inner case:
   - Extract binding (existing)
   - **NEW**: If arm has guard, wrap expression in `if guard { ... }`
   - Execute expression
4. **NEW**: Handle multiple arms with same inner pattern but different guards (if-else-if chain)

**Detailed Changes:**

**Location:** Lines 862-880 (inner pattern case generation)

**Add Guard Logic:**
```go
for _, innerPattern := range sortedInner {
    innerArms := innerGroups[innerPattern]
    innerTag := r.getTagName(innerPattern)
    buf.WriteString(fmt.Sprintf("\tcase %s:\n", innerTag))

    // NEW: Group inner arms by guard presence
    // If multiple arms with same inner pattern, create if-else chain

    for idx, arm := range innerArms {
        // Extract innermost binding
        if arm.binding != "" && arm.binding != "_" {
            bindingCode := r.generateBinding(intermediateVar, arm.pattern, arm.binding)
            buf.WriteString(fmt.Sprintf("\t\t%s\n", bindingCode))
        }

        // NEW: Handle guards
        if arm.guard != "" {
            // Guard present: wrap in if statement
            if idx == 0 {
                buf.WriteString(fmt.Sprintf("\t\tif %s {\n", arm.guard))
            } else {
                buf.WriteString(fmt.Sprintf("\t\t} else if %s {\n", arm.guard))
            }

            // Expression (indented)
            if isInAssignment && assignmentVar != "" {
                buf.WriteString(fmt.Sprintf("\t\t\t%s = %s\n", assignmentVar, arm.expression))
            } else {
                buf.WriteString(fmt.Sprintf("\t\t\t%s\n", arm.expression))
            }
        } else {
            // No guard: else clause or standalone
            if idx > 0 {
                buf.WriteString("\t\t} else {\n")
                indent := "\t\t\t"
            } else {
                indent := "\t\t"
            }

            if isInAssignment && assignmentVar != "" {
                buf.WriteString(fmt.Sprintf("%s%s = %s\n", indent, assignmentVar, arm.expression))
            } else {
                buf.WriteString(fmt.Sprintf("%s%s\n", indent, arm.expression))
            }

            // Close else if we had guards
            if idx > 0 {
                buf.WriteString("\t\t}\n")
            }
        }
    }

    // NEW: Close final if chain if last arm had guard
    if len(innerArms) > 0 && innerArms[len(innerArms)-1].guard != "" {
        buf.WriteString("\t\t}\n")
    }
}
```

#### Step 3: Test Coverage

**Verify These Cases Work:**

1. ✅ Simple guards (already working)
   ```dingo
   Option_Some(x) if x > 10 => "large"
   ```

2. ✅ Multiple guards on same pattern (already working)
   ```dingo
   Option_Some(x) if x > 100 => "huge"
   Option_Some(x) if x > 10 => "large"
   ```

3. ❌ Nested pattern with guard (THIS IS THE FIX)
   ```dingo
   Result_Ok(Option_Some(val)) where val > 0 => "positive"
   ```

4. ❌ Multiple guards on nested pattern (ALSO FIXED)
   ```dingo
   Result_Ok(Option_Some(x)) where x > 100 => "large"
   Result_Ok(Option_Some(x)) where x > 10 => "medium"
   ```

## Files to Modify

### Primary File
- `/Users/jack/mag/dingo/pkg/preprocessor/rust_match.go`
  - Function: `generateCaseWithGuards` (lines 773-981)
  - Changes:
    1. Line 832: Remove `&& !hasGuards` condition
    2. Lines 862-880: Add guard support to nested pattern case generation

### Test Files
- `/Users/jack/mag/dingo/tests/golden/pattern_match_06_guards_nested.dingo` (should pass after fix)
- Consider adding more test cases for edge cases

## Potential Risks

### Risk 1: Regression in Nested Patterns
**What:** Changing nested pattern detection might break existing nested pattern tests
**Mitigation:**
- Run full test suite after change
- Existing tests: `pattern_match_03_nested.dingo`, others
- Ensure no regressions

### Risk 2: Guard Generation Complexity
**What:** Nested switches with guards create complex if-else nesting (3-4 levels deep)
**Mitigation:**
- Keep indentation consistent
- Follow existing guard generation pattern (lines 936-978)
- Test with deeply nested cases

### Risk 3: Multiple Arms, Same Inner Pattern, Different Guards
**What:** Handling cases like:
```dingo
Result_Ok(Option_Some(x)) where x > 100 => "huge"
Result_Ok(Option_Some(x)) where x > 10 => "large"
Result_Ok(Option_Some(x)) => "small"
```
**Mitigation:**
- Group inner arms by pattern AND handle guards as if-else-if chain
- This is same pattern as non-nested guards (already working)
- Reuse existing guard chain logic

## Expected Outcome

### After Fix
1. ✅ Test `pattern_match_06_guards_nested.dingo` passes
2. ✅ All existing tests remain passing (no regressions)
3. ✅ Test count: 103/103 passing (100%)

### Generated Code Example

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
    __Result_Ok_nested := *__match_0.ok_0  // Extract outer value
    switch __Result_Ok_nested.tag {       // Nested switch on inner pattern
    case OptionTag_Some:
        val := *__Result_Ok_nested.some_0  // Extract inner binding

        if val > 0 {                       // ← NEW: Guard check!
            __match_result_0 = "positive"
        } else {                           // ← NEW: Else clause for no-guard arm
            __match_result_0 = "non-positive"
        }
    case OptionTag_None:
        __match_result_0 = "none"
    }
case ResultTag_Err:
    __match_result_0 = "error"
}
```

**Key Difference:** Guard check (`if val > 0`) is now inside the nested switch, after binding extraction.

## Alternative Approaches Considered

### Alternative 1: Separate Code Paths for Guarded Nested Patterns
**Idea:** Create entirely separate handling for "nested patterns with guards"
**Rejected:** Would duplicate 200+ lines of nested switch generation code. Violates DRY principle.

### Alternative 2: Flatten Nested Patterns Before Guard Processing
**Idea:** Transform nested patterns into multiple match expressions
**Rejected:** Changes semantics, harder to maintain, loses performance.

### Alternative 3: Disable Guards on Nested Patterns
**Idea:** Report compile error if user tries to use guards with nested patterns
**Rejected:** Artificial limitation, user-hostile. Feature should work.

## Success Criteria

1. ✅ `pattern_match_06_guards_nested.dingo` compiles without errors
2. ✅ Generated Go code is syntactically valid
3. ✅ Golden test passes (output matches expected)
4. ✅ All 103 tests pass (no regressions)
5. ✅ Code follows existing patterns (consistent style)

## Implementation Complexity

**Estimated Complexity:** Medium

**Reasoning:**
- Core fix is ~30 lines of code (add guard handling to nested case)
- Logic already exists for non-nested guards (can reuse)
- Main work is careful indentation management
- Testing is straightforward (golden tests)

**Time Estimate:** 2-3 hours
- 1 hour: Implement fix
- 1 hour: Test and debug
- 30 min: Validate no regressions

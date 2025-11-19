# Test Plan - Sum Types Phase 2.5

**Session:** 20251116-225837
**Date:** 2025-11-17
**Engineer:** Claude Code (Sonnet 4.5)

---

## Executive Summary

This test plan validates all features implemented in Phase 2.5, with focus on the three critical bug fixes:
1. IIFE Type Inference
2. Tuple Variant Backing Fields
3. Debug Mode Variable Declaration

## 1. Requirements Understanding

### Core Features to Test

**1. IIFE Type Inference** (CRITICAL FIX #1)
- Match expressions should infer return types from first arm
- Literal returns should produce concrete types (int, float64, string, rune)
- Binary expressions should infer correctly (float64 for arithmetic, bool for comparisons)
- Fallback to `interface{}` for complex expressions

**2. Tuple Variant Backing Fields** (CRITICAL FIX #2)
- Tuple variants must generate synthetic field names (`variant_0`, `variant_1`, etc.)
- Constructor parameters must be named (`arg0`, `arg1`, etc.)
- Constructor body must map parameters to fields correctly
- Pattern destructuring must access correct field names

**3. Debug Mode Variable** (CRITICAL FIX #3)
- Debug mode must emit `dingoDebug` variable declaration
- Variable should only be emitted once per file
- Must add `import "os"` (handled by transpiler)
- Runtime behavior: check only when `DINGO_DEBUG` env var is set

**4. Nil Safety Modes** (Configuration Integration)
- Mode "off": No nil checks generated
- Mode "on": Runtime panic with helpful message
- Mode "debug": Conditional checks based on dingoDebug variable

**5. Pattern Destructuring** (Complete Feature)
- Struct patterns: Extract named fields
- Tuple patterns: Extract positional fields
- Unit patterns: No destructuring (no-op)
- Nil checks integrated based on config

**6. Configuration System** (Integration)
- Load `nil_safety_checks` from dingo.toml
- Validate config values
- Default to "on" mode

---

## 2. Test Scenarios

### Priority 1: Critical Bug Fixes

#### Test 1.1: IIFE Type Inference - Integer Literals
**Purpose:** Verify int type inference from literal
**Input:**
```dingo
enum Status { Pending, Active }

func getCode(s Status) int {
    return match s {
        Pending => 0,
        Active => 1,
    }
}
```
**Expected:** IIFE returns `int`, code compiles
**Rationale:** Most common use case - int literals

#### Test 1.2: IIFE Type Inference - Float Literals
**Purpose:** Verify float64 type inference from literal
**Input:**
```dingo
enum Shape { Circle(float64), Point }

func area(s Shape) float64 {
    return match s {
        Circle(r) => 3.14,
        Point => 0.0,
    }
}
```
**Expected:** IIFE returns `float64`, code compiles
**Rationale:** Float literals are common in match returns

#### Test 1.3: IIFE Type Inference - String Literals
**Purpose:** Verify string type inference
**Input:**
```dingo
enum Status { Ok, Error }

func message(s Status) string {
    return match s {
        Ok => "success",
        Error => "failure",
    }
}
```
**Expected:** IIFE returns `string`, code compiles
**Rationale:** String returns are common pattern

#### Test 1.4: IIFE Type Inference - Binary Arithmetic
**Purpose:** Verify float64 inference from arithmetic expressions
**Input:**
```dingo
enum Shape { Circle(float64), Point }

func area(s Shape) float64 {
    return match s {
        Circle(r) => 3.14 * r * r,
        Point => 0.0,
    }
}
```
**Expected:** IIFE returns `float64`, computation works
**Rationale:** Arithmetic in match is common use case

#### Test 1.5: IIFE Type Inference - Boolean Comparison
**Purpose:** Verify bool inference from comparison
**Input:**
```dingo
enum Value { Some(int), None }

func hasValue(v Value) bool {
    return match v {
        Some(x) => x > 0,
        None => false,
    }
}
```
**Expected:** IIFE returns `bool`, code compiles
**Rationale:** Boolean match returns are common

#### Test 2.1: Tuple Variant - Single Field
**Purpose:** Verify single tuple field generates correct backing storage
**Input:**
```dingo
enum Shape { Circle(float64), Point }
```
**Expected Generated:**
```go
type Shape struct {
    tag ShapeTag
    circle_0 *float64  // Synthetic field
}

func Shape_Circle(arg0 float64) Shape {
    return Shape{
        tag: ShapeTag_Circle,
        circle_0: &arg0,
    }
}
```
**Rationale:** Single-field tuples are very common (Option, newtype pattern)

#### Test 2.2: Tuple Variant - Multiple Fields
**Purpose:** Verify multiple tuple fields generate correct storage
**Input:**
```dingo
enum Point { Point2D(float64, float64), Point3D(float64, float64, float64) }
```
**Expected Generated:**
```go
type Point struct {
    tag PointTag
    point2d_0 *float64
    point2d_1 *float64
    point3d_0 *float64
    point3d_1 *float64
    point3d_2 *float64
}

func Point_Point2D(arg0 float64, arg1 float64) Point {
    return Point{
        tag: PointTag_Point2D,
        point2d_0: &arg0,
        point2d_1: &arg1,
    }
}
```
**Rationale:** Multi-field tuples test indexing logic

#### Test 2.3: Tuple Destructuring in Match
**Purpose:** Verify pattern destructuring accesses correct fields
**Input:**
```dingo
enum Shape { Circle(float64), Point }

func area(s Shape) float64 {
    return match s {
        Circle(r) => 3.14 * r * r,
        Point => 0.0,
    }
}
```
**Expected:** Destructuring generates `r := *s.circle_0`
**Rationale:** End-to-end validation of tuple variant flow

#### Test 3.1: Debug Mode Variable Declaration
**Purpose:** Verify dingoDebug variable is emitted
**Input:** Any enum with debug mode enabled
**Expected Generated:**
```go
var dingoDebug = os.Getenv("DINGO_DEBUG") != ""
```
**Rationale:** Required for debug mode to work

#### Test 3.2: Debug Mode Nil Check
**Purpose:** Verify conditional nil checks in debug mode
**Input:**
```dingo
// dingo.toml: nil_safety_checks = "debug"
enum Shape { Circle(float64), Point }

func area(s Shape) float64 {
    return match s {
        Circle(r) => 3.14 * r * r,
        Point => 0.0,
    }
}
```
**Expected:** Nil check uses `if dingoDebug && s.circle_0 == nil`
**Rationale:** Verify integration with config system

#### Test 4.1: Nil Safety - Off Mode
**Purpose:** Verify no checks generated in off mode
**Config:** `nil_safety_checks = "off"`
**Expected:** No nil check statements in generated code
**Rationale:** Performance-critical code needs zero overhead

#### Test 4.2: Nil Safety - On Mode
**Purpose:** Verify runtime panic in on mode
**Config:** `nil_safety_checks = "on"`
**Expected:** Nil check with panic message
**Rationale:** Default safe mode catches bugs

#### Test 4.3: Nil Safety - Debug Mode
**Purpose:** Verify conditional checks in debug mode
**Config:** `nil_safety_checks = "debug"`
**Expected:** Nil check with `dingoDebug &&` condition
**Rationale:** Balanced safety/performance for production

#### Test 5.1: Struct Pattern Destructuring
**Purpose:** Verify named field extraction
**Input:**
```dingo
enum Shape { Rectangle { width: float64, height: float64 } }

func area(s Shape) float64 {
    return match s {
        Rectangle{width, height} => width * height,
    }
}
```
**Expected:** Generate `width := *s.rectangle_width` and `height := *s.rectangle_height`
**Rationale:** Struct patterns are primary use case

#### Test 5.2: Unit Pattern Destructuring
**Purpose:** Verify no destructuring for unit variants
**Input:**
```dingo
enum Status { Pending, Active }

func code(s Status) int {
    return match s {
        Pending => 0,
        Active => 1,
    }
}
```
**Expected:** No field extraction statements
**Rationale:** Unit variants have no data

#### Test 6.1: Config Loading
**Purpose:** Verify config loads from dingo.toml
**Input:** Create dingo.toml with `nil_safety_checks = "debug"`
**Expected:** Config.GetNilSafetyMode() returns NilSafetyDebug
**Rationale:** Config system integration is critical

---

## 3. Test Implementation

### Unit Tests (`pkg/plugin/builtin/sum_types_test.go`)

**Test Structure:**
```go
// CRITICAL FIX #1: Type Inference
func TestInferMatchType_IntLiteral(t *testing.T)
func TestInferMatchType_FloatLiteral(t *testing.T)
func TestInferMatchType_StringLiteral(t *testing.T)
func TestInferMatchType_BinaryArithmetic(t *testing.T)
func TestInferMatchType_BinaryComparison(t *testing.T)

// CRITICAL FIX #2: Tuple Backing Fields
func TestGenerateVariantFields_TupleSingleField(t *testing.T)
func TestGenerateVariantFields_TupleMultipleFields(t *testing.T)
func TestGenerateConstructor_TupleParameters(t *testing.T)
func TestGenerateConstructorFields_TupleMapping(t *testing.T)

// CRITICAL FIX #3: Debug Variable
func TestEmitDebugVariable(t *testing.T)
func TestEmitDebugVariable_OnlyOnce(t *testing.T)

// Pattern Destructuring
func TestGenerateDestructuring_StructPattern(t *testing.T)
func TestGenerateDestructuring_TuplePattern(t *testing.T)
func TestGenerateDestructuring_UnitPattern(t *testing.T)

// Nil Safety Modes
func TestGenerateNilCheck_OffMode(t *testing.T)
func TestGenerateNilCheck_OnMode(t *testing.T)
func TestGenerateNilCheck_DebugMode(t *testing.T)
```

### Integration Tests (Golden Files)

**File:** `tests/golden/sum_types_phase25_tuple_variant.dingo`
**File:** `tests/golden/sum_types_phase25_match_inference.dingo`
**File:** `tests/golden/sum_types_phase25_debug_mode.dingo`
**File:** `tests/golden/sum_types_phase25_nil_safety_off.dingo`

---

## 4. Success Criteria

### All Tests Must Pass
- Unit tests: 100% pass rate
- Golden tests: Generated code compiles successfully
- Integration tests: Runtime behavior correct

### Code Quality
- Generated code is idiomatic Go
- No compilation errors
- No runtime panics (except when expected)

### Coverage
- All three critical fixes validated
- All nil safety modes tested
- All pattern types tested

---

## 5. Test Execution Order

1. **Unit Tests First** - Fast feedback on isolated logic
2. **Golden Tests** - Verify end-to-end transpilation
3. **Integration Tests** - Validate runtime behavior
4. **Config Tests** - Verify configuration loading

---

## 6. Expected Results

### Should PASS (Implemented Features)
- Type inference for literals and binary expressions
- Tuple variant field generation
- Debug mode variable emission
- Nil safety modes (all three)
- Pattern destructuring (struct, tuple, unit)
- Config system integration

### Known Limitations (Not Tested)
- Complex type inference (deferred to Phase 3)
- Exhaustiveness checking (Phase 4)
- Nested patterns (Phase 5)
- Match guards (Phase 6)

---

**Status:** Test Plan Complete
**Next:** Implement and execute tests

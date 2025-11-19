# Engine Test Failures (Out of Phase V Scope)

**Date**: 2025-11-19
**Session**: 20251119-150114
**Context**: These failures are in Dingo engine tests, NOT Phase V infrastructure

## Important Note

**Phase V Scope**: Infrastructure only (documentation, CI/CD, validation, workspace builds)
**These failures**: Dingo engine (transpiler, plugins, AST transformations)
**Responsibility**: Another agent is handling engine test fixes

## Test Failures Summary

**Total**: 8 failing tests
**Category**: All in engine plugin tests
**Impact on Phase V**: None (infrastructure tests all pass)

### 1. OptionTypePlugin Tests (3 failures)
**File**: `pkg/plugin/builtin/option_type_test.go`
**Test**: `TestHandleSomeConstructor_Addressability`
**Issue**: Option type injection not happening
**Root Cause**: Expected Option types not being emitted by plugin

**Failing Cases:**
- `literal_(non-addressable)` - Expected `Option_int`
- `identifier_(addressable)` - Expected `Option_any`
- `string_literal_(non-addressable)` - Expected `Option_string`

### 2. PatternMatchPlugin Tests (2 failures)
**File**: `pkg/plugin/builtin/pattern_match_test.go`
**Test**: `TestPatternMatchPlugin_Transform_AddsPanic`
**Issue**: Pattern match transformation not generating if-else chain
**Root Cause**: Transform phase not executing correctly

### 3. ResultTypePlugin Tests (1 failure)
**File**: `pkg/plugin/builtin/result_type_test.go`
**Test**: `TestTypeDeclaration_BasicResultIntError`
**Issue**: Nil pointer dereference
**Root Cause**: `TypeInferenceService` is nil when `handleGenericResult()` calls it
**Error**: `SIGSEGV at type_inference.go:323`

### 4. FunctionCache Tests (2 failures)
**File**: `pkg/plugin/builtin/function_cache_test.go`
**Test**: `TestContainsUnqualifiedPattern`
**Issue**: Regex too broad, false positives
**Root Cause**: Pattern matching qualified calls and lowercase functions incorrectly

**Failing Cases:**
- `qualified_call` - `os.ReadFile(path)` detected as unqualified
- `lowercase_function` - `readFile(path)` detected as unqualified

## Phase V Infrastructure Tests

**Status**: ✅ ALL PASSING

### Passing Categories (6/6):
1. ✅ Package Management (examples/) - All 3 examples compile and run
2. ✅ Source Map Validation (pkg/sourcemap/) - 98.7% accuracy achieved
3. ✅ CI/CD Tools (scripts/) - diff-visualizer and performance-tracker compile
4. ✅ Workspace Builds (pkg/build/, cmd/dingo/) - All functionality working
5. ✅ Documentation (docs/) - All examples valid, links checked
6. ✅ Overall Integration - Components integrate correctly

## Recommendation

**For Phase V**: Mark as COMPLETE
- All infrastructure components working
- Engine test failures are separate concern
- Another agent handling engine fixes

**For Engine Tests**: Track separately
- Document failures for other agent
- These are pre-existing or unrelated to Phase V work
- Zero Phase V code touched engine components

## Files for Other Agent

Engine test failures documented in:
- This file: `engine-test-failures.md`
- Full test output: `test-results-final.md`

**Scope Separation:**
- Phase V: Infrastructure ✅ Complete
- Engine: Separate workstream (handled by other agent)

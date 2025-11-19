# User Request: Phase 4 Priority 2 & 3 Fixes

## Context
A separate agent is handling Priority 1 (integration test fixes). This session focuses on Priority 2 and Priority 3 items from the Phase 4 analysis.

## Priority 2: Complete Type Inference (Important)

### Task 1: Implement Remaining 4 Context Types for None Inference
**Location**: `pkg/plugin/builtin/type_inference.go`

Current TODOs (lines 645-663):
- Line 645: Assignment context inference
- Line 651: Call argument context inference
- Line 657: Struct field context inference
- Line 663: Composite literal context inference

**Current Status**: None inference only works in ~50% of contexts (5 out of 9 context types)

**Required**: Implement the 4 missing context types to achieve full None type inference

### Task 2: Add go/types Integration for Pattern Match Scrutinee
**Location**: `pkg/plugin/builtin/pattern_match.go:498`

**Issue**: Pattern matching currently cannot detect the scrutinee type accurately
**Required**: Integrate go/types to get actual scrutinee type instead of heuristics

### Task 3: Implement Err() Context-Based Type Inference
**Location**: `pkg/plugin/builtin/result_type.go:286`

**Issue**: `Err()` constructor currently fails type inference in many contexts
**Required**: Use context-based type inference similar to None constant handling

## Priority 3: Guard Support (Nice to Have)

### Task 4: Add Guard Validation to If-Else Chain Transformation
**Locations**:
- `pkg/plugin/builtin/pattern_match_test.go:826`
- `pkg/plugin/builtin/pattern_match_test.go:1009`

**Issue**: Guards work in preprocessor but plugin doesn't handle them in if-else chains
**Current**: 2 tests marked with TODO for guard support
**Required**: Complete guard implementation in AST plugin phase

## Success Criteria

**Priority 2**:
- All 4 context types implemented and tested for None inference
- go/types integration working for pattern match scrutinee detection
- Err() type inference working in all major contexts
- None inference success rate improved from 50% to 90%+

**Priority 3**:
- Guards validated and working in if-else chain transformation
- 2 guard-related TODOs resolved
- Guard tests passing

## Constraints
- Do not break existing functionality
- Maintain Phase 4.1 & 4.2 completed features
- Keep performance overhead <15ms per file
- All changes must have corresponding tests

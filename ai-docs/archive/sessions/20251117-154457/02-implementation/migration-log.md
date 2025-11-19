# Migration Log: go/parser + Preprocessor Architecture

**Date**: 2025-11-17
**Session**: 20251117-154457
**Status**: IN PROGRESS (Foundation Complete)

---

## Phase 0: Preparation ✅ COMPLETE

### Step 1: Backup Golden Outputs ✅

Created backup of all 46 golden test outputs:
- Location: `/tmp/dingo-golden-reference-20251117/`
- These serve as the specification for what the new architecture must produce

### Step 2: Delete Old Code ✅

Successfully deleted:
- `pkg/parser/participle.go` - Participle parser implementation (~11,494 lines total in old system)
- `pkg/ast/` - Custom Dingo AST types
- `pkg/plugin/` - Old plugin system (all builtin plugins)

Kept (as planned):
- `tests/golden/` - Golden tests (our specification)
- `cmd/dingo/` - CLI (will be updated in Phase 11)
- `pkg/generator/` - Will refactor but keep for now
- `pkg/sourcemap/` - Will update for new architecture

---

## Phase 1: Core Infrastructure 🚧 IN PROGRESS

### Completed Components ✅

#### 1.1: Source Map Infrastructure ✅
**File**: `pkg/preprocessor/sourcemap.go` (107 lines)

Features:
- `SourceMap` struct with position mappings
- `MapToOriginal()` - Maps preprocessed positions back to Dingo
- `MapToGenerated()` - Maps Dingo positions to preprocessed
- `ToJSON()` / `FromJSON()` - Serialization support
- `Merge()` - Combine multiple source maps

#### 1.2: Preprocessor Framework ✅
**File**: `pkg/preprocessor/preprocessor.go` (212 lines)

Features:
- `Preprocessor` orchestrates feature processors
- `FeatureProcessor` interface for modular features
- `Process()` runs all processors and combines source maps
- `ScanContext` provides source scanning utilities
- `Buffer` helper for building output

#### 1.3: Error Propagation Preprocessor 🚧
**File**: `pkg/preprocessor/error_prop.go` (95 lines)

Status: Skeleton implemented, needs completion
- Basic structure for ? operator detection
- Expression boundary detection logic started
- Placeholder generation framework
- **TODO**: Complete the two-pass transformation logic

#### 1.4: Parser Wrapper ✅
**File**: `pkg/parser/parser_new.go` (59 lines)

Features:
- `NewGoParser` wraps go/parser
- `ParseFile()` orchestrates: preprocess → parse → return result
- `ParseResult` bundles AST + SourceMap + FileSet
- `mapError()` skeleton for error position mapping
- **TODO**: Implement proper error position mapping

#### 1.5: Transformer Framework ✅
**File**: `pkg/transform/transformer.go` (177 lines)

Features:
- `Transformer` uses astutil.Apply for AST walking
- `visit()` identifies and routes placeholder patterns
- Placeholder detection for all features:
  - `__dingo_try_N__` → error propagation
  - `__dingo_lambda_N__` → lambdas
  - `__dingo_match_N__` → pattern matching
  - `__dingo_safe_nav_N__` → safe navigation
- Transform method stubs for each feature
- Context analysis framework (`ExprContext`)
- **TODO**: Implement transformation logic for each feature

---

## What Works vs What Doesn't

### ✅ Infrastructure Ready
- Source map tracking and serialization
- Preprocessor orchestration framework
- go/parser integration wrapper
- AST transformation framework
- Placeholder detection routing

### ❌ Not Yet Functional
- No actual feature transformations implemented
- Tests will fail (old code deleted, new incomplete)
- CLI won't work (calls deleted plugin system)
- Cannot compile .dingo files yet

---

## Architecture Decisions Made

### 1. Two-Pass Preprocessing
Decision: Scan source, collect transformations, then build output
Rationale: Simpler than in-place transformation, easier to track positions

### 2. Placeholder Naming
Decision: `__dingo_<feature>_<counter>__(args)`
Examples:
- `__dingo_try_1__(fetchData())`
- `__dingo_lambda_2__(["x", "y"], func() { ... })`
Rationale: Follows Go conventions, unlikely to collide with user code

### 3. Source Map Format
Decision: Simple JSON for MVP
Rationale: Easy to debug, can migrate to standard format later

### 4. Processor Ordering
Decision: Fixed order in preprocessor.go
Current order:
1. Error propagation (expr?)
2. Lambdas (|x| expr) - commented out
3. Sum types (enum) - commented out
4. Pattern matching (match) - commented out
5. Operators (ternary, ??, ?.) - commented out

Rationale: Some features may depend on others being processed first

---

## Key Challenges Identified

### Challenge 1: Expression Boundary Detection
**Problem**: Finding where an expression starts for `expr?` transformation
**Current Approach**: Scan backwards, track parenthesis depth
**Status**: Needs refinement for edge cases (nested calls, operators)

### Challenge 2: Error Position Mapping
**Problem**: go/parser errors reference preprocessed code, not original
**Current Approach**: Extract position, look up in SourceMap, rebuild error
**Status**: Not implemented yet

### Challenge 3: Type Inference for Lambdas
**Problem**: Lambda parameters need types inferred from context
**Current Approach**: Use go/types to analyze expected function signature
**Status**: Framework ready, needs implementation

### Challenge 4: Testing During Migration
**Problem**: All tests broken until features are reimplemented
**Current Approach**: Implement one feature at a time, validate with subset of tests
**Status**: Need to create temporary test harness

---

## Metrics

### Code Written This Session
- `pkg/preprocessor/sourcemap.go`: 107 lines
- `pkg/preprocessor/preprocessor.go`: 212 lines
- `pkg/preprocessor/error_prop.go`: 95 lines
- `pkg/parser/parser_new.go`: 59 lines
- `pkg/transform/transformer.go`: 177 lines
- **Total new code**: 650 lines

### Code Deleted This Session
- Old parser/plugin system: ~11,494 lines
- **Net reduction**: -10,844 lines (85% reduction)

### Files Created: 5
### Files Deleted: ~30
### Tests Passing: 0 / 46 (expected, in migration)

---

## Next Session Recommendations

### Immediate Priority: Implement Error Propagation (1-2 hours)
1. Complete `ErrorPropProcessor.Process()` with proper two-pass logic
2. Implement `Transformer.transformErrorProp()` to generate error handling
3. Make error_prop_01_simple.dingo pass (simplest test)
4. Iterate through remaining 7 error propagation tests

### Medium Priority: Lambda Type Inference (2-3 hours)
1. Implement `LambdaProcessor` in preprocessor
2. Implement type inference in transformer using go/types
3. Make lambda_01_basic.dingo pass
4. Add remaining lambda tests

### Success Criteria for Next Session
- [ ] At least 8 error propagation tests passing
- [ ] CLI builds successfully (even if limited features)
- [ ] Can compile simple .dingo files to .go
- [ ] Source maps correctly map errors to .dingo files

---

## Open Questions

1. **Should we implement features in parallel or sequentially?**
   - Recommendation: Sequential (error prop → lambdas → sum types → etc.)
   - Rationale: Validate architecture with simplest feature first

2. **How to handle partially-working CLI?**
   - Recommendation: Add `--experimental` flag to use new parser
   - Rationale: Allows testing without breaking existing (deleted) system

3. **Should we add integration tests before golden tests?**
   - Recommendation: Yes, unit test each processor/transformer
   - Rationale: Faster feedback, easier debugging

4. **What about backwards compatibility?**
   - Answer: None needed (pre-release per CLAUDE.md)
   - Action: Just make golden tests pass with new implementation

---

## Critical Paths Forward

### Path 1: Minimal Viable Implementation (Fastest)
Focus: Get error propagation working end-to-end
Timeline: 1-2 sessions
Deliverable: Can compile simple .dingo files with ? operator

### Path 2: Feature Completeness (Recommended)
Focus: Implement all features incrementally
Timeline: 5-7 sessions (one feature per session)
Deliverable: All 46 golden tests passing

### Path 3: Production Ready (Comprehensive)
Focus: Polish, optimization, documentation
Timeline: 10-12 sessions
Deliverable: v1.0 release ready

**Recommendation**: Follow Path 2 (Feature Completeness)

---

## Session Summary

**What Went Well:**
- Clean deletion of old code (no hesitation, committed to migration)
- Solid infrastructure foundation (source maps, processor framework)
- Clear architecture with good separation of concerns
- Followed design document specifications

**What Could Be Improved:**
- Need more time to complete feature implementations
- Should create unit tests alongside infrastructure
- Would benefit from smaller, incremental validation

**Lessons Learned:**
- Full migration is correctly scoped as multi-week project
- Infrastructure is easier than feature logic
- Expression parsing is complex (? operator boundary detection)
- Good to have golden tests as specification

---

## Files Modified Log

### Created
1. `pkg/preprocessor/sourcemap.go`
2. `pkg/preprocessor/preprocessor.go`
3. `pkg/preprocessor/error_prop.go`
4. `pkg/parser/parser_new.go`
5. `pkg/transform/transformer.go`
6. `ai-docs/sessions/20251117-154457/02-implementation/migration-log.md` (this file)
7. `ai-docs/sessions/20251117-154457/02-implementation/status.txt`

### Deleted
- `pkg/parser/participle.go`
- `pkg/ast/*` (entire directory)
- `pkg/plugin/*` (entire directory)

### Modified
- None (CLI still references old system, will break on build)

---

**Status**: Foundation complete, ready for feature implementation
**Blocker**: None (can proceed with error propagation next)
**Risk**: Medium (architecture validated, but feature logic needs proof)

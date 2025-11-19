# Architectural Clarifications

Based on the comprehensive plan and investigation, here are the answers to the architectural questions:

## Q1: Enum Syntax
**Question**: Should we support Go-like `type Result enum { Ok(T); Err(E) }` or Rust-like `enum Result { Ok(T), Err(E) }`?

**Answer**: **Rust-like** `enum Result { Ok(T), Err(E) }`

**Rationale**:
- Existing golden tests already use this syntax
- More concise and familiar to developers from other languages
- The `enum` keyword clearly signals "this is a Dingo feature, not standard Go"
- Aligns with Result/Option ergonomics goal
- We can add Go-like syntax later if needed (backward compatible)

## Q2: Type Inference for Err()
**Question**: Should `Err()` calls require explicit type context, or bidirectional type inference?

**Answer**: **Defer to Phase 3** - For now, use `interface{}` placeholder and improve later

**Rationale**:
- Full type inference requires go/types integration (complex, 6-10 hours)
- Current Fix A3 uses heuristics which is good enough for basic cases
- We can improve incrementally:
  - Phase 2: Basic cases work with heuristics
  - Phase 3: Add full go/types integration
- Users can work around with explicit types if needed
- **Priority**: Get end-to-end working first, optimize later

## Q3: Constructor Syntax
**Question**: Should users call `Ok(value)` (implicit) or `Result_Ok(value)` (explicit)?

**Answer**: **Bare constructors** - `Ok(value)` and `Err(error)`

**Rationale**:
- Golden tests use bare `Ok()` and `Err()`
- More ergonomic and cleaner code
- Result type plugin already implements this (Fix A2)
- Name collision risk is acceptable:
  - If user defines `func Ok()`, they can use qualified names
  - This is same trade-off as Go's `error` type
  - Documentation will warn about reserved names
- Method syntax `Result.Ok()` would require different transformation (defer)

## Q4: Generated Code Location
**Question**: Should Result declarations go at top of file or at first usage?

**Answer**: **Top of file, after imports**

**Rationale**:
- Matches current implementation (appends to file.Decls)
- Easier to find in generated code (developers expect types at top)
- Go convention: imports → types → functions
- Source mapping is simpler (consistent offset)
- No issues with init() order since types don't have side effects

## Q5: Source Mapping Strategy
**Question**: Maintain line numbers (padding) or offset-based mapping?

**Answer**: **Offset-based mapping** (current approach)

**Rationale**:
- Already implemented in preprocessor (source map generation)
- Padding with blank lines makes generated code ugly
- Source maps handle offset correctly (tested in Fix A2 session)
- LSP uses source maps for bidirectional position translation
- **Trade-off**: Slightly more complex, but cleaner generated code

## Q6: Testing Strategy
**Question**: Unit tests (fast) or integration tests (realistic)?

**Answer**: **Both, with priority order**

**Implementation Order**:
1. **Unit tests first** (TDD approach):
   - Enum preprocessor unit tests
   - Plugin pipeline activation tests
   - Quick feedback loop during development

2. **Integration tests second**:
   - End-to-end: .dingo → .go → compile → run
   - Golden file tests
   - Catches integration issues

3. **Before merge**:
   - All unit tests passing
   - At least 3 golden tests passing end-to-end
   - Binary builds and runs successfully

## Q7: Error Handling in Preprocessor
**Question**: Fail loudly on invalid enum syntax or silently skip?

**Answer**: **Lenient mode** - Skip malformed enums, log warning

**Rationale**:
- Early in development, strict mode would be brittle
- Better to let parser handle unknown syntax
- Allows gradual rollout (some files use enums, others don't)
- **Logging strategy**:
  - Malformed enum → Warning + skip
  - Complete garbage → Parser error (more informative)
- Can add strict mode later via config flag

## Q8: Performance Optimization
**Question**: Single-file or batch compilation?

**Answer**: **Single-file** (current architecture)

**Rationale**:
- Simpler implementation (matches current design)
- Good enough for Phase 2 goals
- Batch compilation benefits:
  - Cross-file type inference
  - Parallel compilation
  - **But**: Adds significant complexity (6-10 hours extra)
- **Defer to Phase 4** or later
- Most projects are small enough that single-file is fine

---

## Additional Clarifications

### Golden Test Fix
The investigation revealed that golden tests call `parser.ParseFile()` directly, bypassing the preprocessor. This is why they fail even though `dingo build` works perfectly.

**Fix**: Update golden tests to:
1. Read .dingo source
2. **Run preprocessor** (NEW)
3. Parse preprocessed source
4. Generate Go code
5. Compare with .go.golden

**File to modify**: `tests/golden_test.go` (lines 87-91)

### Implementation Phases

**Phase 1: Fix Golden Tests** (1 hour) - QUICK WIN
- Add preprocessor step to tests
- Unblocks testing of existing features
- Verifies TypeAnnotProcessor works

**Phase 2: Enum Preprocessor** (4-6 hours)
- Create `pkg/preprocessor/enum.go`
- Handle `enum Name { Variant1, Variant2 }` syntax
- Generate Go sum type representation
- Add comprehensive tests

**Phase 3: Activate Plugin Pipeline** (6-8 hours)
- Wire Result type plugin into generator
- Ensure Transform() is called on AST
- Integrate type inference service
- End-to-end testing

**Phase 4: Integration & Polish** (2-3 hours)
- Update golden files
- Add integration tests
- Documentation
- Code review

**Total**: 13-18 hours

---

## Priority Order

### High Priority (Must Have)
1. ✅ Fix golden tests to use preprocessor
2. ✅ Activate Result type plugin in pipeline
3. ✅ End-to-end: Ok()/Err() working

### Medium Priority (Should Have)
4. ✅ Enum preprocessor (basic support)
5. ✅ Multiple golden tests passing
6. ✅ Documentation updates

### Low Priority (Nice to Have)
7. ⚠️ Advanced enum features (generics, methods)
8. ⚠️ Full type inference with go/types
9. ⚠️ Batch compilation

---

## Risk Mitigation

### Risk 1: Enum preprocessor complexity
**Mitigation**: Start with simple enums (no data), add variants incrementally

### Risk 2: Plugin pipeline integration breaks existing code
**Mitigation**: Add feature flag to disable plugins, test extensively

### Risk 3: Source mapping breaks after enum transformation
**Mitigation**: Write source map tests first, verify before proceeding

### Risk 4: Timeline overrun
**Mitigation**:
- Phase 1 (golden test fix) is 1 hour - delivers immediate value
- Can stop after any phase if needed
- Incremental delivery keeps progress visible

---

## Success Criteria

**Phase 1 Complete**:
- ✅ At least 3 golden tests parsing (error_prop_01, 03, 06)
- ✅ Preprocessor integrated into test suite

**Phase 2 Complete**:
- ✅ Enum preprocessor handles basic enums
- ✅ Generated Go code compiles
- ✅ Unit tests passing

**Phase 3 Complete**:
- ✅ Result type plugin active in pipeline
- ✅ Ok(42) generates CompositeLit
- ✅ Result type declarations injected
- ✅ End-to-end: .dingo → .go → go run works

**Phase 4 Complete**:
- ✅ 10+ golden tests passing
- ✅ Code review approved
- ✅ CHANGELOG updated
- ✅ Committed and pushed

---

**Decision**: Proceed with comprehensive 4-phase plan
**Estimated Time**: 13-18 hours
**First Milestone**: Phase 1 (1 hour) - Fix golden tests

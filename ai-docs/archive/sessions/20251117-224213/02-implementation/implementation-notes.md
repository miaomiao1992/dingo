# Implementation Notes

## Key Decisions

### 1. Issues #2 and #3 Were False Positives
The code review from GPT-5 Codex identified these as CRITICAL bugs, but our analysis (Tasks B and C) confirmed both are already correctly implemented. We added comprehensive tests to prevent future regressions instead of "fixing" non-existent bugs.

### 2. Config System Design
Task E introduced a new `Config` struct pattern that can be extended for future compiler flags. The design follows Go best practices:
- Simple enum-like string constants for modes
- Default config constructor
- Validation at CLI boundary
- Immutable config passed through processing pipeline

### 3. Test Coverage Strategy
Rather than minimal tests, we added 30+ comprehensive tests covering:
- Negative test cases (things that should NOT happen)
- Edge cases (boundary conditions)
- Integration scenarios (multiple features interacting)

This "test-heavy" approach locks in correct behavior and prevents regressions.

### 4. Documentation-First for Architecture
Task D created comprehensive preprocessor documentation BEFORE future features are added. This establishes the CRITICAL POLICY (imports always last) that prevents future bugs.

## Deviations from Plan

### Minor Deviations
1. **Task K**: No changes needed - Task E already included flag documentation
2. **Test counts**: Added MORE tests than planned (30+ vs planned 12+)
3. **Config validation**: Added validation in both config.go AND main.go for defense-in-depth

### No Major Deviations
All other tasks followed the plan exactly as specified.

## Performance Impact

### Source Map Fix (Task A)
- Performance: NEUTRAL (no change, just logic fix)
- Memory: NEUTRAL

### Config Flag (Task E)
- Performance: NEGLIGIBLE (one string comparison per file)
- Memory: +24 bytes per Config struct (1 per compilation)

### New Tests (Tasks F-J)
- No runtime impact (tests only)

## Risk Assessment

### Low-Risk Changes
- Task A: Trivial one-line fix with full test coverage
- Task E: Opt-in feature (default behavior unchanged)
- Tasks F-J: Tests only (no production impact)
- Task D, L: Documentation only

### Medium-Risk Changes
- None

### High-Risk Changes
- None

## Known Limitations

### Config Flag (Task E)
- Only validates at CLI level (no compile-time enforcement)
- Error messages could be more detailed
- No config file support (only CLI flag)

### Future Work
1. Add config file support (e.g., `dingo.toml`)
2. Add more granular error messages for mode violations
3. Consider making "single" mode emit warnings instead of errors
4. Add telemetry to track which mode is most popular

## Testing Coverage

### Before This Session
- 8 functions tested
- 11 test cases
- No negative tests

### After This Session
- 8+ functions tested
- 40+ test cases
- Comprehensive negative test coverage

## Build Verification

All tasks verified builds pass:
```bash
go test ./pkg/preprocessor/...
# PASS
# All tests passed
```

No regressions introduced.

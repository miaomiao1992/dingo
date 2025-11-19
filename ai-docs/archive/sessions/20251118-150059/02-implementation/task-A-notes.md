# Task A: Configuration System - Implementation Notes

## Key Decisions

### 1. Extended Existing Configuration (Not Created from Scratch)
**Decision:** Added match configuration to existing `pkg/config` package instead of creating new files.

**Rationale:**
- The config package already had comprehensive infrastructure:
  - TOML parsing (github.com/BurntSushi/toml already imported)
  - Multi-source loading (defaults → user config → project config → CLI overrides)
  - Validation framework
  - Comprehensive test suite
- Creating separate loader.go would have duplicated functionality
- Cleaner architecture to have all config in one place

**Files:** Extended `config.go` (added MatchConfig struct and validation)

### 2. Default Syntax: "rust"
**Decision:** Set default match syntax to "rust" instead of "swift"

**Rationale:**
- Consistency with other Dingo defaults:
  - `LambdaSyntax` defaults to "rust" (|x| expr style)
  - `ErrorPropagationSyntax` defaults to "question" (? operator, Rust-inspired)
- Dingo's overall design philosophy leans toward Rust syntax
- Users can easily override via dingo.toml if they prefer Swift style

**Code:** Line 185 in config.go

### 3. Validation Allows Empty String
**Decision:** Empty `match.syntax` is valid and uses default

**Rationale:**
- Matches pattern used for other optional config fields
- Allows users to omit `[match]` section entirely
- DefaultConfig() provides fallback value
- Cleaner than requiring explicit configuration

**Code:** Lines 256-264 in config.go (validation checks `!= ""` before validating)

### 4. Two Syntax Options Only: "rust" and "swift"
**Decision:** Limited to exactly two syntax styles

**Rationale:**
- Follows execution plan specification
- "rust": `match expr { ... }` - familiar to Rust developers
- "swift": `switch expr { ... }` - familiar to Swift/C-family developers
- More options would fragment the community
- Can extend later if needed (validation is centralized)

**Code:** Lines 257-260 in config.go

## Deviations from Original Plan

### Deviation 1: No Separate loader.go File
**Plan:** Create `pkg/config/loader.go` with Load() and Validate()

**Actual:** Functions already existed in `config.go`

**Impact:** None - functionality is identical, just better organized

**Justification:**
- Existing Load() function (lines 194-229) already implements:
  - Multi-source loading (user config, project config, CLI overrides)
  - Error handling
  - TOML parsing
- Existing Validate() function (lines 247-333) already has validation framework
- Simply extended existing functions instead of duplicating code

### Deviation 2: No go.mod Changes Needed
**Plan:** Add github.com/BurntSushi/toml dependency

**Actual:** Dependency already present (line 6 of go.mod)

**Impact:** None - one less file to modify

## Test Coverage

### Comprehensive Test Suite
Added 6 new test cases covering all scenarios:

1. **Default value test** - Ensures "rust" is default
2. **Valid rust syntax** - Config validation passes
3. **Valid swift syntax** - Config validation passes
4. **Invalid syntax** - Validation fails with clear error
5. **Empty syntax** - Uses default without error
6. **Load from TOML** - Parses dingo.toml correctly
7. **Invalid TOML value** - Fails validation

All tests follow existing patterns in config_test.go.

## Integration Readiness

### For Pattern Match Preprocessor (Task B)
The preprocessor can access config as follows:

```go
// In preprocessor initialization
cfg, err := config.Load(nil)
if err != nil {
    return err
}

// Choose parser based on syntax
switch cfg.Match.Syntax {
case "rust":
    // Use Rust-style pattern: match expr { ... }
    processor.pattern = regexp.MustCompile(`\bmatch\b\s+...`)
case "swift":
    // Use Swift-style pattern: switch expr { ... }
    processor.pattern = regexp.MustCompile(`\bswitch\b\s+...`)
}
```

### For Generator Integration
Generator should load config early:

```go
// In cmd/dingo/build.go or pkg/generator/generator.go
cfg, err := config.Load(cliOverrides)
if err != nil {
    return fmt.Errorf("failed to load config: %w", err)
}

// Pass to generator
gen := generator.New(cfg)
```

## Future Enhancements

### Potential Extensions
1. **Per-file syntax override** - Comments like `// dingo:match-syntax swift`
2. **Migration tool** - Convert between rust/swift syntax styles
3. **Hybrid mode** - Allow both syntaxes in same project (low priority)
4. **Linter integration** - Warn if using non-configured syntax

### Not Implemented (Out of Scope for Task A)
- CLI flags for match syntax (will be in Task E: CLI Integration)
- Generator integration (will be part of larger integration task)
- Documentation updates (will be in Task F: Documentation)

## Performance Considerations

### Config Loading
- Happens once per build process (not per file)
- TOML parsing is fast (<1ms for typical config)
- No performance concerns

### Validation
- Runs once after loading
- String comparison only (negligible overhead)
- Fail-fast on invalid config (good for developer experience)

## Error Messages

### Clear, Actionable Errors
All validation errors include:
1. What's invalid: "invalid match.syntax"
2. The invalid value: "scala"
3. Valid options: "must be 'rust' or 'swift'"

Example:
```
invalid configuration: invalid match.syntax: "scala" (must be 'rust' or 'swift')
```

This matches the error message pattern used throughout the config package.

## Backward Compatibility

### Safe for Existing Projects
- If dingo.toml omits `[match]` section → uses default "rust"
- If dingo.toml has empty `syntax = ""` → uses default "rust"
- No breaking changes to existing config format
- All existing tests still pass

## Summary

Task A successfully extended the configuration system to support pattern matching syntax selection. The implementation integrates seamlessly with existing infrastructure, follows established patterns, and is ready for use by the Pattern Match Preprocessor (Task B).

**Status:** SUCCESS
**Tests:** 11/11 passing
**Files Modified:** 2
**Files Created:** 0
**Integration:** Ready for Task B

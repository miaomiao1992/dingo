# Task 1b: Error Infrastructure - Design Decisions

## Architecture Decisions

### 1. Error Package Structure

**Decision**: Create dedicated `pkg/errors/` package for compile error types

**Rationale**:
- Separation of concerns: Error types are distinct from plugin logic
- Reusability: Error types can be used across different packages (parser, generator, plugins)
- Standard Go pattern: Following stdlib conventions (e.g., `errors`, `io`, `fmt`)
- Future extensibility: Easy to add new error categories without polluting plugin package

**Alternatives Considered**:
- Keep errors in `pkg/plugin/` - Rejected: Would couple error types to plugin implementation
- Use stdlib `errors` only - Rejected: Need structured error types with location/hints

### 2. CompileError Design

**Decision**: Use struct with Message, Location, Hint, and Category fields

**Rationale**:
- **Message**: Human-readable description of what went wrong
- **Location**: token.Pos for precise source code position
- **Hint**: Actionable suggestion for fixing the error (improves DX)
- **Category**: Groups errors for filtering/reporting (future IDE integration)

**Example**:
```go
CompileError{
    Message:  "cannot infer type for expression: x",
    Location: token.Pos(42),
    Hint:     "Try providing an explicit type annotation, e.g., var x: int = ...",
    Category: ErrorCategoryTypeInference,
}
```

**Benefits**:
- Clear error messages guide users to solutions
- Category enables future features (filter by error type, severity levels)
- Location enables IDE integration (jump to error)
- Hint improves developer experience (not just "what's wrong" but "how to fix")

### 3. Error Accumulation Strategy

**Decision**: Accumulate errors in Context, don't fail fast

**Rationale**:
- **Better DX**: Report multiple errors in one compilation pass (like Go compiler)
- **Incremental fixes**: User can fix multiple issues before recompiling
- **Phase compatibility**: Plugins can continue processing to find more errors
- **Future-proof**: Foundation for warning vs error distinction

**Implementation**:
```go
// Context accumulates errors
ctx.ReportError("error 1", pos1)
ctx.ReportError("error 2", pos2)
ctx.ReportError("error 3", pos3)

// Check errors after pipeline
if ctx.HasErrors() {
    for _, err := range ctx.GetErrors() {
        fmt.Fprintln(os.Stderr, err)
    }
    os.Exit(1)
}
```

**Trade-off**: Slightly more complex error handling vs. better user experience

### 4. TempVarCounter Placement

**Decision**: Add TempVarCounter directly to Context struct

**Rationale**:
- **Global uniqueness**: Counter must be shared across all plugins and phases
- **Simplicity**: No need for separate service/registry
- **Performance**: Direct field access (no indirection)
- **IIFE requirements**: Ensures __tmp0, __tmp1, __tmp2 are unique across entire file

**Alternative Considered**:
- Per-plugin counters - Rejected: Could cause name collisions between plugins
- Static counter - Rejected: Not thread-safe, problematic for parallel compilation

**Usage Pattern**:
```go
// In Result plugin:
tmpVar := ctx.NextTempVar() // "__tmp0"

// In Option plugin (same file):
tmpVar := ctx.NextTempVar() // "__tmp1"

// Guaranteed unique across all plugins
```

### 5. Error Factory Functions

**Decision**: Provide both generic (`NewTypeInferenceError`) and specific (`TypeInferenceFailure`) factories

**Rationale**:
- **Generic factories**: Flexible for custom error messages
- **Specific factories**: Consistent error messages for common cases
- **DRY principle**: Avoid duplicating hint text across plugins
- **Maintainability**: Update hint in one place

**Example**:
```go
// Generic factory (custom message):
err := NewTypeInferenceError("custom error message", pos, "custom hint")

// Specific factory (standardized):
err := TypeInferenceFailure("myExpr", pos)
// Produces: "cannot infer type for expression: myExpr"
// With hint: "Try providing an explicit type annotation..."
```

### 6. Position Formatting Strategy

**Decision**: Implement `FormatWithPosition()` method with FileSet fallback

**Rationale**:
- **IDE compatibility**: Standard Go error format (file:line:col: message)
- **Graceful degradation**: Falls back to basic error if FileSet unavailable
- **User-friendly**: Shows exactly where error occurred
- **Tool integration**: Enables IDE "jump to error" functionality

**Output Format**:
```
test.dingo:10:5: Type Inference Error: cannot infer type for expression: x
  Hint: Try providing an explicit type annotation, e.g., var x: int = ...
```

## Error Reporting Strategy

### Compile Error Generation (User Requirement)

**Requirement**: "Generate compile error when type inference fails completely"

**Implementation Strategy**:

1. **During plugin processing**: Report errors via `ctx.ReportError()`
2. **After pipeline completes**: Check `ctx.HasErrors()`
3. **If errors exist**: Generate invalid Go code with descriptive comment

**Example Generated Code** (for type inference failure):
```go
// DINGO COMPILE ERROR: Type Inference Error
// cannot infer type for expression: Ok(someComplexExpr)
// Hint: Try providing an explicit type annotation
var __dingo_error_placeholder_0 = "COMPILATION_FAILED_SEE_ERROR_ABOVE"
```

**Benefits**:
- **Clear feedback**: User sees error immediately
- **Go compiler integration**: Invalid code triggers Go compiler error
- **Descriptive comments**: User knows exactly what went wrong
- **No silent failures**: No invalid interface{} fallback

**Implementation in Plugins** (Task 2a/2b):
```go
// In Result plugin:
typeName, err := p.typeInference.InferTypeWithFallback(expr)
if err != nil {
    ctx.ReportError(err.Error(), expr.Pos())
    // Generate error placeholder
    return &ast.Ident{Name: "__DINGO_TYPE_INFERENCE_FAILED__"}
}
```

## Testing Strategy

### Comprehensive Coverage

**Unit Tests**: 13 tests covering all functionality
- Error creation and formatting
- Context error accumulation
- Temp variable generation
- Position tracking

**Test Philosophy**:
- **Test behavior, not implementation**: Focus on public API
- **Cover edge cases**: Empty errors, nil FileSet, counter overflow
- **Verify uniqueness**: Temp var names never collide
- **Check error messages**: Ensure hints are present

### Why 100% Test Coverage?

**Rationale**:
- **Error infrastructure is critical**: Bugs here affect all error reporting
- **Public API surface**: All methods are used by plugins
- **Future maintenance**: Tests document expected behavior
- **Refactoring safety**: Can confidently modify internals

## Integration Points

### How Plugins Will Use This Infrastructure

**Phase 2a (Result Plugin)**:
```go
func (p *ResultTypePlugin) transformOkConstructor(call *ast.CallExpr) ast.Expr {
    typeName, err := p.inferTypeWithFallback(valueArg)
    if err != nil {
        // Use error infrastructure
        p.ctx.ReportError(
            fmt.Sprintf("cannot infer type for Ok(): %v", err),
            call.Pos(),
        )
        return generateErrorPlaceholder(p.ctx)
    }

    // Check addressability (uses TempVarCounter)
    if !isAddressable(valueArg) {
        tmpVar := p.ctx.NextTempVar() // "__tmp0"
        return wrapInIIFE(valueArg, typeName, tmpVar)
    }

    return generateOkStruct(valueArg, typeName)
}
```

**Phase 3 (Generator)**:
```go
func (g *Generator) Generate(file *ast.File) (*ast.File, error) {
    // Run plugin pipeline
    transformed, err := g.pipeline.Transform(file)
    if err != nil {
        return nil, err
    }

    // Check for compile errors
    if g.pipeline.Ctx.HasErrors() {
        for _, compileErr := range g.pipeline.Ctx.GetErrors() {
            fmt.Fprintln(os.Stderr, compileErr)
        }
        return nil, fmt.Errorf("compilation failed with %d errors",
            len(g.pipeline.Ctx.GetErrors()))
    }

    return transformed, nil
}
```

## Future Enhancements

### Potential Improvements (Phase 4+)

1. **Error Severity Levels**:
   - Warning vs Error distinction
   - Configurable error-as-warning mode
   - Linting rules integration

2. **Rich Error Context**:
   - Show source code snippet around error
   - Multi-line error ranges
   - Related errors grouping

3. **IDE Integration**:
   - LSP diagnostics support
   - Quick fixes (code actions)
   - Inline error annotations

4. **Error Recovery**:
   - Partial compilation on error
   - Best-effort code generation
   - Degraded mode for IDE

5. **Structured Logging**:
   - JSON error output for tools
   - Machine-readable error codes
   - Error statistics/metrics

## Constraints Satisfied

✅ **Stay focused ONLY on error infrastructure** - No plugin modifications
✅ **Do NOT modify Result or Option plugins yet** - Deferred to Task 2a/2b
✅ **Ensure errors are user-friendly with clear guidance** - Hint field + descriptive messages
✅ **Support collecting multiple errors** - Error accumulation in Context
✅ **Add TempVarCounter to Context** - Implemented with NextTempVar() method
✅ **Add required methods to Context** - ReportError(), GetErrors(), NextTempVar()

## Success Metrics

- ✅ All tests passing (13/13)
- ✅ Clean separation of concerns (errors package)
- ✅ Comprehensive test coverage
- ✅ Clear, actionable error messages
- ✅ Foundation ready for Task 1c (addressability) and Task 2a/2b (plugin updates)

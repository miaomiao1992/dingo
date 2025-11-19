# Option B Implementation Plan: Separate AST for Injected Types

## Decision: Option B (3-1 Expert Consensus)

**Models recommending Option B:**
- GPT-5.1 Codex: "Guarantees zero pollution, scales better"
- MiniMax M2: "More robust, avoids Go internal API dependency"
- Grok Code Fast: "Architectural clarity, LSP compatible"

**Key insight from MiniMax M2**: Option C depends on undocumented Go comment map internals that may break with future Go versions.

## Architecture Change

### Current (1 AST)
```
Plugins → Inject into main AST → Print 1 AST → .go file
```

### New (2 ASTs)
```
Plugins → Main AST (user code) + Injected AST (Result/Option types)
       → Print injected AST first
       → Print main AST second
       → Concatenate → .go file
```

## Implementation Steps

### Step 1: Create InjectedTypesAST in Plugin Pipeline

**File**: `pkg/plugin/plugin.go`

**Add field to Pipeline struct**:
```go
type Pipeline struct {
    ctx              *Context
    plugins          []Plugin
    stats            Stats
    injectedTypesAST *ast.File  // NEW: Separate AST for injected types
}
```

**Add method to get injected AST**:
```go
// GetInjectedTypesAST returns the separate AST containing injected type declarations
// Returns nil if no types were injected
func (p *Pipeline) GetInjectedTypesAST() *ast.File {
    return p.injectedTypesAST
}
```

**Modify Transform() method - Phase 3 (Inject)**:
```go
// Phase 3: Inject (Add type declarations to SEPARATE AST)
for _, plugin := range p.plugins {
    if dp, ok := plugin.(DeclarationProvider); ok {
        decls := dp.GetPendingDeclarations()
        if len(decls) > 0 {
            // OPTION B: Create separate AST for injected types
            if p.injectedTypesAST == nil {
                p.injectedTypesAST = &ast.File{
                    Name: ast.NewIdent(transformed.Name.Name),  // Same package
                    Decls: make([]ast.Decl, 0),
                }
            }
            // Add to injected AST instead of main AST
            p.injectedTypesAST.Decls = append(p.injectedTypesAST.Decls, decls...)

            dp.ClearPendingDeclarations()
        }
    }
}

// NOTE: Main AST (transformed) is NOT modified - remains user code only
```

### Step 2: Modify Generator to Print Both ASTs

**File**: `pkg/generator/generator.go`

**Update Generate() method** (around line 173-180):

**Current code**:
```go
// Step 5: Print AST to Go source code
var buf bytes.Buffer

cfg := printer.Config{
    Mode:     printer.TabIndent | printer.UseSpaces,
    Tabwidth: 8,
}

err = cfg.Fprint(&buf, g.fset, transformed)
if err != nil {
    return nil, fmt.Errorf("failed to print AST: %w", err)
}

return buf.Bytes(), nil
```

**New code**:
```go
// Step 5: Print AST to Go source code
var buf bytes.Buffer

cfg := printer.Config{
    Mode:     printer.TabIndent | printer.UseSpaces,
    Tabwidth: 8,
}

// OPTION B: Print injected types AST first (if it exists)
if g.pipeline != nil {
    injectedAST := g.pipeline.GetInjectedTypesAST()
    if injectedAST != nil {
        // Print injected types (Result, Option) at top of file
        err = cfg.Fprint(&buf, g.fset, injectedAST)
        if err != nil {
            return nil, fmt.Errorf("failed to print injected types AST: %w", err)
        }

        // Add spacing between injected types and user code
        buf.WriteString("\n\n")
    }
}

// Print main AST (user code)
err = cfg.Fprint(&buf, g.fset, transformed)
if err != nil {
    return nil, fmt.Errorf("failed to print main AST: %w", err)
}

return buf.Bytes(), nil
```

### Step 3: Ensure Injected Nodes Use token.NoPos

**Files**:
- `pkg/plugin/builtin/result_type.go`
- `pkg/plugin/builtin/option_type.go`

**Already done** in previous attempt! All injected AST nodes already have `token.NoPos` positions.

**Why this matters**:
- Injected types won't be navigable in LSP (which is correct - they're generated)
- Source maps can exclude them entirely
- No LSP diagnostics will point to injected types

### Step 4: Source Map Handling (Optional Enhancement)

**File**: `pkg/generator/generator.go`

**Current behavior**: Source maps generated for entire output

**Option B behavior**:
- Injected AST uses `token.NoPos` (not mapped)
- Main AST mapped as before
- LSP will only show diagnostics for user code

**No changes needed** - this works automatically because injected nodes have `token.NoPos`.

### Step 5: Testing

**Test 1**: Verify comment pollution is eliminated
```bash
go test ./tests -run TestGoldenFiles/pattern_match_01_simple -v
```

Expected output:
```go
// CLEAN: Injected types at top
type Option_string struct {
    tag    OptionTag
    some_0 *string
}

func Option_string_Some(arg0 string) Option_string {
    return Option_string{tag: OptionTag_Some, some_0: &arg0}
}
// ✅ NO DINGO comments here!

// User code below
func processOption(opt Option_string) int {
    // DINGO_MATCH_START: opt  ← Comments stay with match expressions
    __match_0 := opt
    switch __match_0.tag {
    // ...
    }
    // DINGO_MATCH_END
}
```

**Test 2**: Verify compilation
```bash
go test ./tests -run TestGoldenFilesCompilation/pattern_match_01_simple -v
```

Expected: Code compiles without errors

**Test 3**: Run all pattern matching tests
```bash
go test ./tests -run TestGoldenFiles/pattern_match -v
```

Expected: 13/13 tests should pass after updating golden files

## Estimated Timeline

- **Step 1**: Plugin pipeline changes - 1 hour
- **Step 2**: Generator changes - 30 minutes
- **Step 3**: Verify positions (already done) - 0 minutes
- **Step 4**: Source map (no changes) - 0 minutes
- **Step 5**: Testing & debugging - 1 hour

**Total: ~2.5 hours**

## Success Criteria

✅ Injected types (Result, Option) appear at top of file
✅ User code appears below injected types
✅ NO DINGO comments in injected type declarations
✅ DINGO comments preserved with match expressions
✅ Generated code compiles without errors
✅ All 13 pattern matching tests pass

## Rollback Plan

If Option B causes unforeseen issues:
1. Revert changes to `pkg/plugin/plugin.go`
2. Revert changes to `pkg/generator/generator.go`
3. Plugins will inject into main AST as before
4. Back to square one, try Option A (position filtering)

## Future Benefits

**Extensibility** (from GPT-5.1 Codex):
- Adding lambda support: Inject lambda helpers to separate AST
- Adding iterator support: Inject iterator types to separate AST
- Adding async support: Inject async primitives to separate AST

**Maintainability** (from MiniMax M2):
- Clear boundary: "Injected types logic" vs "User code transformation"
- Easy debugging: Check injected AST separately
- No dependency on Go internal comment map behavior

**Robustness** (from Grok):
- Architectural isolation prevents comment pollution by design
- Future Go versions won't break this (uses standard AST operations)
- LSP integration protected by architectural boundary

## References

- Internal analysis: `/Users/jack/mag/dingo/ai-docs/sessions/20251118-234826/output/internal-architecture-analysis.md`
- GPT-5.1 Codex: `/Users/jack/mag/dingo/ai-docs/sessions/20251118-234826/output/gpt-5.1-codex-architecture.md`
- MiniMax M2: `/Users/jack/mag/dingo/ai-docs/sessions/20251118-234826/output/minimax-m2-architecture.md`
- Grok Code Fast: `/Users/jack/mag/dingo/ai-docs/sessions/20251118-234826/output/grok-code-fast-architecture.md`

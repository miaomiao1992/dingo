# Follow-up Investigation: Comment Preservation Issue

## Context

We implemented the comment preservation fix as recommended:

```go
// After AST transformation
reattachOrphanedComments(file.File, transformed)
```

Where `reattachOrphanedComments` copies `file.Comments` from original to transformed AST.

## Problem Discovered

The fix DOESN'T work because of this code in `pkg/generator/generator.go` line 229:

```go
// Step 6: Format the generated code
formatted, err := format.Source(buf.Bytes())
```

**`format.Source()` re-parses the code and STRIPS comments!**

## Why This Happens

1. We manually print AST to `buf` (preserving comments via `cfg.Fprint`)
2. Comments ARE in the output at this point
3. Then we call `format.Source(buf.Bytes())`
4. `format.Source` internally does:
   - Re-parse the code string
   - Re-format with `gofmt`
   - **Loses comments that don't have perfect position associations**
5. Source map markers disappear

## The Question

**How do we preserve `// dingo:s:N` and `// dingo:e:N` comments through `format.Source()`?**

### Options I Can Think Of

**Option A**: Don't use `format.Source()`
- Print the AST with `go/printer` properly configured
- Use `printer.Config` settings to match `gofmt` behavior
- Skip the `format.Source()` step entirely

**Option B**: Protect comments before formatting
- Extract `// dingo:s:N` and `// dingo:e:N` comments before `format.Source()`
- Run `format.Source()`
- Re-inject comments at correct positions after formatting
- Complex and fragile

**Option C**: Use a different comment format
- Instead of `// dingo:s:1`, use something format.Source preserves
- Maybe directive comments like `//go:generate`?
- But might interfere with Go tooling

**Option D**: Post-process to add comments back
- Keep track of where error propagation blocks are
- After `format.Source()`, parse again and add comments
- Similar to current DINGO:GENERATED marker injection

## Request

Please provide expert guidance on:

1. **What's the best approach?** (A, B, C, D, or something else?)
2. **Specific implementation** for the chosen approach
3. **Code changes needed** in `pkg/generator/generator.go`
4. **Any Go stdlib functions/patterns** that handle this correctly

## Current Code Structure

```go
// Step 5: Print AST to buffer
var buf bytes.Buffer
cfg := printer.Config{Mode: printer.TabIndent | printer.UseSpaces, Tabwidth: 8}

// Manual printing (package, imports, injected types, declarations)
cfg.Fprint(&buf, g.fset, transformed)  // Comments ARE in output here

// Step 6: Format - THIS IS WHERE COMMENTS DISAPPEAR
formatted, err := format.Source(buf.Bytes())  // ❌ Strips comments!

// Step 7: Inject markers (post-processing)
// Currently only injects DINGO:GENERATED markers
```

## What We Need

A solution that:
- ✅ Preserves `// dingo:s:N` and `// dingo:e:N` comments
- ✅ Produces properly formatted Go code (`gofmt` style)
- ✅ Works reliably across different AST transformations
- ✅ Ideally simple and maintainable

Please provide specific code recommendations!

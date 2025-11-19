# Parser Strategy Recommendation

**Date**: 2025-11-17
**Decision**: Use go/parser + Preprocessor Approach
**Confidence**: Very High

---

## TL;DR

**Use Strategy 1: go/parser + Preprocessor**

Transform Dingo syntax to valid Go in a preprocessing step, then use `go/parser` for all actual parsing. This gives us 100% Go compatibility for free while keeping the codebase small and maintainable.

```
.dingo → [Text Transform] → Valid Go → [go/parser] → AST → [Transform] → .go + .sourcemap
```

**Time to MVP**: 2-3 weeks
**Total Custom Code**: ~2,300 lines
**Go Parsing Coverage**: 100% (via go/parser)

---

## The Core Insight

**You were absolutely right**: We should NOT reimplement Go's parser.

Go already has industrial-strength parsing tools:
- `go/parser` - Used by gopls, gofmt, go vet, and the compiler
- `go/scanner` - Battle-tested tokenizer
- `go/ast` - Complete AST representation
- `golang.org/x/tools/go/ast/astutil` - Powerful AST manipulation

Our job is simple: **Transform Dingo syntax → valid Go, then let Go's tools do the rest.**

---

## Why This Approach Wins

### 1. Leverage Existing Tools (Don't Reinvent the Wheel)

```
Current Participle Approach:
- ~500 lines written
- ~20% Go syntax coverage
- Need 4,000-5,000 more lines to complete
- Must track Go evolution manually
- Error handling DIY

Preprocessor + go/parser:
- ~2,300 lines total
- 100% Go syntax coverage (via go/parser)
- Auto-tracks Go evolution
- Professional error reporting (via go/parser)
- Battle-tested (go/parser has 10+ years of production use)
```

**We save ~5,400 lines of code we don't have to write, test, debug, or maintain.**

### 2. Industry-Proven Pattern

This is not a novel approach - it's the standard pattern:

| Tool | Approach |
|------|----------|
| **TypeScript** | TS syntax → JS syntax → JS ecosystem |
| **Rust Macros** | Macro expansion → valid Rust → Rust parser |
| **PREGO** | Preprocessor macros → valid Go → go/parser |
| **gpp** | Macro expansion → valid Go → go build |
| **templ** | Template → Go code → gopls |
| **C Preprocessor** | `#define`, `#include` → valid C → C compiler |

**Every successful meta-language uses preprocessing + native parser.**

### 3. Simple, Maintainable Architecture

**Phase 1: Text Transformation** (~1,000 LOC)
```go
// Input: .dingo file
file := os.Open("config.json")?

// Output: Valid Go code
file := __dingo_try_1(os.Open("config.json"))
```

**Phase 2: go/parser** (0 LOC - FREE)
```go
import "go/parser"

ast, err := parser.ParseFile(fset, filename, preprocessedGo, parser.ParseComments)
// That's it! We have a complete Go AST.
```

**Phase 3: AST Transformation** (~800 LOC)
```go
import "golang.org/x/tools/go/ast/astutil"

// Replace __dingo_try_N() with proper error handling
astutil.Apply(ast, func(cursor *astutil.Cursor) bool {
    if call, ok := cursor.Node().(*ast.CallExpr); ok {
        if isDingoTryCall(call) {
            cursor.Replace(createErrorHandlingBlock(call))
        }
    }
    return true
}, nil)
```

**Phase 4: Code Generation** (0 LOC - FREE)
```go
import "go/printer"

printer.Fprint(output, fset, ast)  // That's it!
```

### 4. Fast Development Timeline

| Week | Task | Output |
|------|------|--------|
| **1** | Preprocessor MVP | `?` operator, lambdas working |
| **2** | AST transformation | Optimized code generation |
| **3** | Pattern matching | `match`, sum types working |
| **4** | Polish | Error mapping, source maps, CLI |

**Total: 4 weeks to feature-complete transpiler**

Compare to:
- Participle approach: 3-6 months (reimplementing Go)
- Tree-sitter approach: 2-3 months (CST → AST conversion)

### 5. Perfect Go Ecosystem Integration

Because we use `go/parser` and generate `go/ast`, we get perfect integration:

```go
// Generated AST works with ALL Go tools:
import (
    "go/printer"    // Format code
    "go/types"      // Type checking
    "golang.org/x/tools/go/analysis"  // Static analysis
    // ... entire Go ecosystem
)
```

Our generated `.go` files work with:
- gopls (via LSP proxy + source maps)
- gofmt (perfect formatting)
- go vet (all checks work)
- go test (all testing tools)
- go build (no special handling needed)

---

## How It Works: Concrete Examples

### Example 1: Error Propagation

**Input (.dingo)**
```dingo
func readFile(path string) Result[[]byte, error] {
    file := os.Open(path)?
    defer file.Close()

    data := io.ReadAll(file)?
    return Ok(data)
}
```

**After Preprocessing (valid Go)**
```go
func readFile(path string) Result[[]byte, error] {
    file := __dingo_try_1(os.Open(path))
    defer file.Close()

    data := __dingo_try_2(io.ReadAll(file))
    return Ok(data)
}
```

**After go/parser → AST Transformation → go/printer**
```go
func readFile(path string) Result[[]byte, error] {
    __dingo_tmp_1, __dingo_err_1 := os.Open(path)
    if __dingo_err_1 != nil {
        return Err[[]byte](__dingo_err_1)
    }
    file := __dingo_tmp_1
    defer file.Close()

    __dingo_tmp_2, __dingo_err_2 := io.ReadAll(file)
    if __dingo_err_2 != nil {
        return Err[[]byte](__dingo_err_2)
    }
    data := __dingo_tmp_2

    return Ok(data)
}
```

### Example 2: Pattern Matching

**Input (.dingo)**
```dingo
result := match user {
    Some(u) => u.Name,
    None => "Guest",
}
```

**After Preprocessing (valid Go)**
```go
result := func() string {
    switch __dingo_match := user.(type) {
    case Some[User]:
        u := __dingo_match.value
        return u.Name
    case None[User]:
        return "Guest"
    default:
        panic("non-exhaustive pattern match")
    }
}()
```

**After AST Transformation** (optimized, but same logic)
```go
var result string
switch __dingo_match := user.(type) {
case Some[User]:
    u := __dingo_match.value
    result = u.Name
case None[User]:
    result = "Guest"
default:
    panic("non-exhaustive pattern match")
}
```

---

## Why NOT Other Strategies?

### Strategy 2: go/scanner + Hybrid Parser ❌

**Problem**: The "hybrid" concept doesn't work in practice.

You can't parse half a file with custom logic, then "hand off" to go/parser for the other half. Parsers maintain state (scope, context, pending expressions). Switching mid-parse is infeasible.

This degenerates into **"reimplement all of Go parsing"** which defeats the entire purpose.

**Estimated work**: 10,000-15,000 lines
**Risk**: Very high (no precedent)
**Benefit**: None (end up reimplementing go/parser)

### Strategy 3: Direct go/parser with AST Injection ❌

**Problem**: Dingo syntax must be valid Go.

```dingo
// We want this:
result := value?

// But this isn't valid Go!
// We'd have to write:
result := value.question()

// Which defeats the point of nicer syntax
```

This forces ugly workarounds like:
- `value?` → `value.question()` (method call syntax)
- `match x { ... }` → Can't do it (not valid Go)
- `|x| x * 2` → `func(__lambda_x int) int { return __lambda_x * 2 }` (already valid Go, no win)

**Verdict**: Too limited. Compromises Dingo's value proposition.

### Strategy 4: Tree-sitter ❌

**Problem**: Adds complexity without benefit.

Tree-sitter excels at:
- Incremental parsing (for editors)
- Error recovery (for syntax highlighting)
- Syntax highlighting queries

But we don't need these for a **transpiler** that:
- Runs once (not incremental)
- Fails on invalid syntax (no error recovery)
- Generates code (not highlighting)

**What we'd have to do**:
1. Fork tree-sitter-go grammar (JavaScript)
2. Add Dingo syntax rules
3. Generate parser (C code)
4. Use via CGO bindings
5. Convert CST → custom AST
6. Convert custom AST → go/ast
7. Track tree-sitter-go updates manually

**What we get**:
- An extra layer (CST → AST → go/ast instead of go/ast)
- JavaScript + C build dependencies
- Maintenance burden

**Verdict**: Wrong tool for the job. Tree-sitter is for editors, not transpilers.

---

## Technical Feasibility: Proof of Concept

### Minimal Preprocessor (Go Code)

```go
package preprocessor

import (
    "bytes"
    "go/scanner"
    "go/token"
)

// Transform converts .dingo source to valid Go source
func Transform(src []byte) []byte {
    var buf bytes.Buffer
    var s scanner.Scanner

    fset := token.NewFileSet()
    file := fset.AddFile("", fset.Base(), len(src))
    s.Init(file, src, nil, scanner.ScanComments)

    tryCounter := 0
    lastPos := 0

    for {
        pos, tok, lit := s.Scan()
        if tok == token.EOF {
            buf.Write(src[lastPos:])
            break
        }

        offset := file.Offset(pos)

        // Copy everything up to this token
        buf.Write(src[lastPos:offset])

        // Transform `?` operator
        if tok == token.QUESTION {
            // Look back to find expression
            expr := extractExpressionBefore(src, lastPos, offset)

            // Replace with function call
            tryCounter++
            buf.WriteString("__dingo_try_")
            buf.WriteString(strconv.Itoa(tryCounter))
            buf.WriteString("(")
            buf.WriteString(expr)
            buf.WriteString(")")

            lastPos = offset + 1  // Skip the `?`
        } else {
            // Regular token, pass through
            buf.WriteString(lit)
            lastPos = offset + len(lit)
        }
    }

    return buf.Bytes()
}
```

This is ~50 lines of straightforward code. The full preprocessor with all Dingo features would be ~1,000 lines.

### AST Transformation Example

```go
package transform

import (
    "go/ast"
    "golang.org/x/tools/go/ast/astutil"
)

// TransformDingoAST replaces preprocessor placeholders with real code
func TransformDingoAST(node ast.Node) {
    astutil.Apply(node, func(c *astutil.Cursor) bool {
        call, ok := c.Node().(*ast.CallExpr)
        if !ok {
            return true
        }

        // Match __dingo_try_N(expr) calls
        if ident, ok := call.Fun.(*ast.Ident); ok {
            if strings.HasPrefix(ident.Name, "__dingo_try_") {
                // Replace with error handling block
                errorBlock := createErrorHandling(call.Args[0])
                c.Replace(errorBlock)
            }
        }

        return true
    }, nil)
}

func createErrorHandling(expr ast.Expr) *ast.BlockStmt {
    // Generate:
    // {
    //     __tmp, __err := expr
    //     if __err != nil {
    //         return Err(__err)
    //     }
    //     __tmp  // Result value
    // }

    return &ast.BlockStmt{
        List: []ast.Stmt{
            &ast.AssignStmt{ /* assignment */ },
            &ast.IfStmt{ /* error check */ },
            &ast.ExprStmt{ /* result value */ },
        },
    }
}
```

This is ~150 lines for full error propagation transformation.

**Total for `?` operator**: ~200 lines
**Total for all Dingo features**: ~2,300 lines

---

## Risk Analysis

### Low Risk

**Technical Risks**:
- go/parser is stable, battle-tested, used by millions of tools
- AST transformation is well-documented (Eli Bendersky's blog, Go docs)
- Pattern proven by TypeScript, Rust, PREGO, gpp, templ

**Implementation Risks**:
- Text transformation is straightforward
- AST manipulation has excellent tooling (astutil)
- Source mapping is solved problem (JSON format, standard libraries)

**Maintenance Risks**:
- Small codebase (~2,300 LOC vs 10,000+ for alternatives)
- Simple architecture (easy to onboard contributors)
- Decoupled from Go evolution (go/parser tracks for us)

### Mitigation Strategies

| Risk | Mitigation |
|------|------------|
| Error message positions wrong | Source mapping from day 1 |
| Generated code ugly | AST transformation polish |
| Go syntax changes | Use go/parser (auto-handles) |
| Complex Dingo features | Iterative development, test-driven |

---

## Next Steps

### Phase 1: Proof of Concept (Week 1)

**Goal**: Get `?` operator working end-to-end

1. Implement minimal preprocessor
   - Scan for `?` tokens
   - Replace with `__dingo_try_N()` calls
   - Output valid Go

2. Parse with go/parser
   - Parse preprocessed Go
   - Verify AST is correct

3. Transform AST
   - Replace `__dingo_try_N()` with error handling
   - Generate Result type if needed

4. Generate code
   - Use go/printer
   - Verify output compiles

**Success Criteria**: `.dingo` file with `?` operator → compiled `.go` file

### Phase 2: Full Feature Set (Weeks 2-3)

1. Lambda expressions (`|x| x * 2`)
2. Pattern matching (`match x { ... }`)
3. Sum types (`enum` keyword)
4. Null coalescing (`??` operator)
5. Safe navigation (`?.` operator)

### Phase 3: Production Ready (Week 4)

1. Source map generation (JSON format)
2. Error message mapping (go errors → dingo positions)
3. CLI integration (`dingo build`)
4. Golden tests (50+ test cases)
5. Documentation

---

## Conclusion

**Recommendation**: Immediately switch to go/parser + Preprocessor approach.

**Abandon**: Participle-based parser (sunk cost fallacy - we've only invested ~500 lines)

**Timeline**: 4 weeks to production-ready transpiler

**Risk**: Very low (proven pattern, standard tools)

**Confidence**: Very high

---

## Questions & Answers

**Q: Isn't preprocessing "hacky"?**

A: No - it's the industry standard. TypeScript, Rust macros, C preprocessor, Babel, and countless other tools use this pattern. It's proven, maintainable, and simple.

**Q: Won't generated code be ugly?**

A: Initially yes (with `__dingo_*` placeholders), but AST transformation cleans it up. Final output is idiomatic Go.

**Q: What about error messages?**

A: Source maps translate go/parser errors back to .dingo positions. Users see errors in their .dingo code, not generated .go code.

**Q: Can we handle complex Dingo features?**

A: Yes. Pattern matching, sum types, lambdas all transform to valid Go constructs. Some require creative transformations, but all are feasible.

**Q: What if Go adds new features?**

A: They automatically work! go/parser handles them, we just pass through.

**Q: Isn't this just moving complexity around?**

A: No - we're **eliminating** complexity. ~2,300 lines vs ~10,000+ lines, using battle-tested tools vs DIY parser.

---

**Final Recommendation**: Start implementing go/parser + Preprocessor immediately. This is the right architecture for Dingo.

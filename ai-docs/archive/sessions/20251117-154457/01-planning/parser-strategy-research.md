# Parser Strategy Research: Leveraging Go Native Tooling

**Date**: 2025-11-17
**Session**: 20251117-154457
**Status**: Complete

## Executive Summary

After comprehensive research into parser strategies for Dingo, the clear winner is **Strategy 1: go/parser + Preprocessor**. This approach leverages Go's battle-tested native tooling for 100% of Go parsing while only requiring custom code for Dingo-specific syntax transformation.

**Key Finding**: We should NOT reimplement Go's parser. Go already has industrial-strength parsing tools (`go/parser`, `go/scanner`, `go/ast`) used by gopls, gofmt, and the compiler itself. Our job is to transform Dingo syntax into valid Go, then let Go's tools do the heavy lifting.

---

## Strategy 1: go/parser + Preprocessor (RECOMMENDED)

### Overview
Transform Dingo syntax to valid Go syntax in a preprocessing step, then use `go/parser` for all actual parsing.

### Architecture

```
.dingo file
    ↓
[Preprocessor: Text → Text]
    - Convert `?` operator to function calls
    - Convert `match` to switch statements
    - Convert lambdas to func literals
    - Convert sum types to struct definitions
    ↓
Valid Go code (temporary/in-memory)
    ↓
[go/parser: Text → AST]
    - 100% Go compatibility (free)
    - All Go features work (free)
    - Error reporting (free)
    ↓
Go AST
    ↓
[AST Transformer: AST → AST]
    - Refine preprocessor placeholders
    - Add source map annotations
    - Optimize generated code
    ↓
[go/printer: AST → Text]
    ↓
.go file + .sourcemap file
```

### How It Works

#### Phase 1: Text Preprocessing (Lexical)
Convert Dingo syntax to valid Go syntax using simple text/token transformations:

**Example: Error Propagation (`?` operator)**
```dingo
// Input
func readConfig() Result[Config, error] {
    file := os.Open("config.json")?
    return parseConfig(file)
}
```

```go
// After preprocessing (valid Go)
func readConfig() Result[Config, error] {
    file := __dingo_try_1(os.Open("config.json"))
    return parseConfig(file)
}
```

**Example: Pattern Matching (`match` expression)**
```dingo
// Input
result := match value {
    Some(x) => x * 2,
    None => 0,
}
```

```go
// After preprocessing (valid Go)
result := func() int {
    switch __dingo_match := value.(type) {
    case Some:
        x := __dingo_match.value
        return x * 2
    case None:
        return 0
    default:
        panic("non-exhaustive match")
    }
}()
```

**Example: Lambdas**
```dingo
// Input
numbers.Map(|x| x * 2)
```

```go
// After preprocessing (valid Go)
numbers.Map(func(__dingo_lambda_arg_0 int) int { return __dingo_lambda_arg_0 * 2 })
```

#### Phase 2: go/parser
Parse the preprocessed Go code using standard `go/parser`:

```go
import (
    "go/parser"
    "go/token"
)

fset := token.NewFileSet()
ast, err := parser.ParseFile(fset, filename, preprocessedSource, parser.ParseComments)
```

**We get for FREE:**
- Complete Go syntax support (imports, types, methods, generics, etc.)
- Error detection and reporting
- Comment preservation
- Exact position tracking (for source maps)

#### Phase 3: AST Transformation
Refine the AST to optimize generated code:

```go
import (
    "go/ast"
    "golang.org/x/tools/go/ast/astutil"
)

// Example: Transform __dingo_try_N() calls into proper error handling
astutil.Apply(node, func(cursor *astutil.Cursor) bool {
    if call, ok := cursor.Node().(*ast.CallExpr); ok {
        if id, ok := call.Fun.(*ast.Ident); ok && strings.HasPrefix(id.Name, "__dingo_try_") {
            // Replace with proper error handling block
            cursor.Replace(createErrorHandlingBlock(call))
        }
    }
    return true
}, nil)
```

### Advantages

1. **Zero Go Reimplementation**: We inherit 100% Go syntax support instantly
2. **Battle-Tested**: go/parser is used by gopls, gofmt, go vet, and millions of tools
3. **Error Reporting**: Get excellent error messages from go/parser
4. **Maintainability**: Simple text transformations + standard AST manipulation
5. **Debuggability**: Preprocessed output is readable Go code
6. **Future-Proof**: New Go features automatically work
7. **Small Codebase**: Only custom code is Dingo → Go transformation rules

### Disadvantages

1. **Two-Phase Parsing**: Slight complexity in architecture
2. **Placeholder Ugliness**: Temporary code has `__dingo_*` names
3. **Error Positions**: Need source maps to map errors back to .dingo files

### Implementation Complexity: LOW

**Estimated LOC**: 2,000-3,000 lines
- Preprocessor (text transformations): 800-1,200 lines
- AST transformations: 600-1,000 lines
- Source map generation: 400-600 lines
- Utilities: 200-400 lines

### Precedents

This is the **standard approach** used by:

1. **PREGO** (Go preprocessor): Text macros → valid Go → go/parser
2. **gpp** (Go macro preprocessor): AST mutation functions → go/parser
3. **templ**: Generates Go code → go/parser → gopls
4. **TypeScript**: Transforms TS → JS → JS parser (similar philosophy)

---

## Strategy 2: go/scanner + Hybrid Parser

### Overview
Use `go/scanner` for tokenization, custom parser for Dingo constructs, fallback to go/parser for Go syntax.

### Architecture

```
.dingo file
    ↓
[go/scanner: Text → Tokens]
    ↓
[Hybrid Parser: Tokens → AST]
    - Custom logic for `?`, `match`, lambdas
    - Fallback to go/parser for standard Go
    ↓
Dingo AST (hybrid)
    ↓
[Transformer: Dingo AST → Go AST]
    ↓
.go file
```

### How It Works

```go
import "go/scanner"

s := scanner.Scanner{}
s.Init(fset.AddFile("", fset.Base(), len(src)), src, nil, scanner.ScanComments)

for {
    pos, tok, lit := s.Scan()
    if tok == token.EOF {
        break
    }

    // Custom parsing logic
    if tok == token.QUESTION {  // `?` operator
        // Handle error propagation
    } else if lit == "match" {
        // Parse match expression
    } else {
        // Standard Go syntax - delegate to go/parser somehow?
    }
}
```

### Advantages

1. **Standard Tokenization**: Reuse `go/scanner` for lexing
2. **Partial Reuse**: Don't need to rewrite tokenizer
3. **Fine-Grained Control**: Custom logic for Dingo-specific parsing

### Disadvantages

1. **Complex Integration**: How to "fallback" to go/parser mid-parse?
2. **Parser Reimplementation**: Still need to parse most Go constructs
3. **Error Handling**: Must implement error recovery ourselves
4. **Maintenance Burden**: Need to keep parser in sync with Go evolution
5. **No Precedent**: No examples of this hybrid approach working well

### Implementation Complexity: VERY HIGH

**Estimated LOC**: 10,000-15,000 lines
- Custom parser for Go constructs: 6,000-9,000 lines
- Dingo-specific parsing: 1,500-2,500 lines
- Error recovery: 1,000-2,000 lines
- AST construction: 1,500-2,500 lines

### Assessment

**Not Recommended**: The "fallback to go/parser" concept is theoretically appealing but practically infeasible. You can't stop mid-parse and hand off to a different parser. This degenerates into "reimplement most of Go parsing" which defeats the purpose.

---

## Strategy 3: Direct go/parser with AST Injection

### Overview
Try to parse `.dingo` files directly with `go/parser` by making syntax "close enough" to Go.

### How It Works

Make Dingo syntax valid Go at parse time:

```dingo
// Dingo syntax
result := value?  // Not valid Go!

// Could we make it valid Go somehow?
result := value.question()  // Method call - valid Go!
```

Then post-process AST to detect "special" patterns:

```go
ast, _ := parser.ParseFile(fset, filename, dingoSource, 0)

astutil.Apply(ast, func(cursor *astutil.Cursor) bool {
    if call, ok := cursor.Node().(*ast.CallExpr); ok {
        if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
            if sel.Sel.Name == "question" {
                // This was actually a `?` operator!
                cursor.Replace(createErrorPropagation(sel.X))
            }
        }
    }
    return true
}, nil)
```

### Advantages

1. **One Parser**: Only use go/parser
2. **Simplicity**: No preprocessing step

### Disadvantages

1. **Syntax Constraints**: Dingo syntax must be valid Go
2. **Ugly Syntax**: `value.question()` instead of `value?`
3. **Ambiguity**: Can't distinguish `value.question()` call from actual method
4. **Limited Features**: Can't support pattern matching, sum types, etc.
5. **Poor UX**: Defeats the purpose of a nicer syntax

### Implementation Complexity: LOW (but with major compromises)

### Assessment

**Not Viable**: This forces Dingo syntax to be "Go with magic methods" which eliminates most of the value proposition. We want `?`, `match`, lambdas, etc. - not method call workarounds.

---

## Strategy 4: Tree-sitter

### Overview
Use tree-sitter-go grammar as base, extend with Dingo constructs.

### How It Works

1. Fork tree-sitter-go grammar
2. Add Dingo syntax rules to grammar.js
3. Generate parser
4. Use in Go via go-tree-sitter bindings

```javascript
// grammar.js (extended)
module.exports = grammar(require('tree-sitter-go'), {
    name: 'dingo',

    rules: {
        // Extend Go grammar
        postfix_expression: $ => choice(
            $.error_propagation,  // New: `?` operator
            // ... existing Go rules
        ),

        error_propagation: $ => seq(
            field('operand', $.expression),
            '?'
        ),

        // ... more Dingo rules
    }
});
```

### Advantages

1. **Grammar Reuse**: Start with complete Go grammar
2. **Extensibility**: Tree-sitter designed for language extensions
3. **Incremental Parsing**: Fast for editor use
4. **CST Output**: Concrete syntax tree with all tokens

### Disadvantages

1. **JavaScript Dependency**: Grammar defined in JavaScript
2. **Build Complexity**: Need Node.js, tree-sitter CLI for development
3. **Go Integration**: Must use CGO bindings (go-tree-sitter)
4. **AST Conversion**: Still need to convert CST → Go AST
5. **Limited Ecosystem**: Less tooling than go/ast ecosystem
6. **Maintenance**: Must track tree-sitter-go updates manually
7. **Not Actually Easier**: Still need to understand and modify grammar

### Implementation Complexity: MEDIUM-HIGH

**Estimated LOC**: 5,000-8,000 lines Go + 500-1,000 lines JavaScript
- Grammar extension (JavaScript): 500-1,000 lines
- CST → AST conversion: 2,500-4,000 lines
- AST → Go AST transformation: 1,500-2,500 lines
- Source mapping: 500-1,000 lines

### Key Question: Does Tree-sitter Give Us Go For Free?

**Answer: NO**

While tree-sitter-go provides a complete Go grammar, we still need to:
1. Convert CST to a usable AST
2. Transform that to Go's `go/ast` format (for code generation)
3. Handle all the same semantic transformations

**We don't save work** - we just add an extra layer (CST → custom AST → go/ast) instead of the simpler path (text → go/ast).

### Assessment

**Not Recommended For Dingo**: Tree-sitter excels at editor tooling (syntax highlighting, incremental parsing). For a transpiler that runs once and generates code, the preprocessor approach is simpler and more maintainable.

Tree-sitter would be valuable if we were building a Dingo-native LSP that doesn't use gopls. But since we're proxying to gopls (which uses go/parser), we should use go/parser too.

---

## Comparative Analysis

### Development Complexity

| Strategy | Custom Code | Go Parsing | Dingo Parsing | Maintenance |
|----------|-------------|------------|---------------|-------------|
| 1. Preprocessor + go/parser | Low | FREE | Simple | Low |
| 2. go/scanner + Hybrid | Very High | Manual | Medium | Very High |
| 3. go/parser + AST Injection | Low | FREE | Constrained | Low |
| 4. Tree-sitter | Medium-High | Via CST | Medium | Medium |

### Feature Support

| Strategy | Go Features | Dingo Features | Future Go | Error Quality |
|----------|-------------|----------------|-----------|---------------|
| 1. Preprocessor + go/parser | 100% | Unlimited | Auto | Excellent (via go/parser) |
| 2. go/scanner + Hybrid | Manual (~80%) | Unlimited | Manual | Manual |
| 3. go/parser + AST Injection | 100% | Very Limited | Auto | Excellent (via go/parser) |
| 4. Tree-sitter | Via CST | Unlimited | Manual | Good |

### Ecosystem Integration

| Strategy | gopls | gofmt | go vet | IDE Tools |
|----------|-------|-------|--------|-----------|
| 1. Preprocessor + go/parser | Perfect | Perfect | Perfect | Perfect |
| 2. go/scanner + Hybrid | Via source maps | Via AST | Via AST | Via LSP |
| 3. go/parser + AST Injection | Perfect | Perfect | Perfect | Perfect |
| 4. Tree-sitter | Via source maps | Via AST | Via AST | Via LSP |

### Time to First Working Version

| Strategy | Estimate | Confidence |
|----------|----------|------------|
| 1. Preprocessor + go/parser | 2-3 weeks | High |
| 2. go/scanner + Hybrid | 3-6 months | Low |
| 3. go/parser + AST Injection | 1-2 weeks | High (but limited) |
| 4. Tree-sitter | 6-10 weeks | Medium |

---

## How Other Tools Solve This

### templ (Template Engine → Go)

**Approach**: Custom parser → Generate Go code → gopls proxy

- Custom parser for `.templ` files (mixing HTML and Go)
- Generates `.go` files with Go code
- LSP proxies to gopls using source maps
- **Key insight**: Doesn't try to parse Go itself - generates it and lets Go tools handle it

### TypeScript (Superset of JavaScript)

**Approach**: Custom parser → Generate JS → JS tooling

- Custom TypeScript parser (written in TypeScript)
- Parses full TS syntax including all JS
- Type-checks and transforms to JS
- **Key insight**: Reimplemented JS parsing, but had Microsoft resources and 10+ year timeline

### Borgo (Rust-like → Go)

**Approach**: Custom parser (Rust-written) → Generate Go

- Written in Rust (not Go)
- Custom parser for Borgo syntax (not based on Go parser)
- Generates Go code
- **Key insight**: Treats Go as compilation target, not trying to "extend" Go's parser

### Rust Macros

**Approach**: Macro → AST injection

- Macros expand to valid Rust syntax
- Rust parser handles everything
- **Key insight**: Preprocessing approach, similar to Strategy 1

---

## Real-World Parser Complexity Data

### go/parser Source Code
- `go/parser/parser.go`: ~2,900 lines
- `go/scanner/scanner.go`: ~800 lines
- Total: ~3,700 lines of highly optimized, well-tested code

**If we reimplement**: We'd need similar LOC, but without 10+ years of battle-testing.

### Participle (Current Approach)
- Current Dingo parser: ~500 lines
- Go coverage: ~20% (missing selectors, assignments, most statements)
- Remaining work: ~4,000-5,000 lines estimated

### Preprocessor Approach (Proposed)
- Text transformations: ~1,000 lines
- AST transformations: ~800 lines
- Source maps: ~500 lines
- Total: ~2,300 lines (reusing 100% of go/parser's 3,700 lines)

**Savings**: ~5,400 lines we don't have to write (and maintain, debug, test)

---

## Technical Deep Dive: Preprocessor Implementation

### Preprocessor Architecture

```go
package preprocessor

import (
    "go/scanner"
    "go/token"
)

type Preprocessor struct {
    src     []byte
    fset    *token.FileSet
    scanner scanner.Scanner

    // Track Dingo → Go mappings for source maps
    mappings []SourceMapping
}

// Transform converts .dingo source to valid Go source
func (p *Preprocessor) Transform() ([]byte, error) {
    p.scanner.Init(p.fset.AddFile("", p.fset.Base(), len(p.src)), p.src, nil, 0)

    var output []byte
    var lastPos token.Pos

    for {
        pos, tok, lit := p.scanner.Scan()
        if tok == token.EOF {
            break
        }

        // Copy unchanged content
        output = append(output, p.src[lastPos:pos]...)

        // Transform Dingo syntax
        if tok == token.QUESTION {
            // Convert `?` to function call
            dingoExpr := p.extractExpression(lastPos, pos)
            goExpr := fmt.Sprintf("__dingo_try_%d(%s)", p.nextID(), dingoExpr)
            output = append(output, goExpr...)
            p.recordMapping(lastPos, pos, len(output)-len(goExpr), len(output))
        } else if lit == "match" {
            // Convert match expression
            matchExpr := p.parseMatchExpression()
            switchStmt := p.convertMatchToSwitch(matchExpr)
            output = append(output, switchStmt...)
        } else {
            // Regular token - pass through
            output = append(output, p.src[pos:p.scanner.Pos()]...)
        }

        lastPos = p.scanner.Pos()
    }

    return output, nil
}
```

### Example Transformations

#### 1. Error Propagation (`?`)

```go
// Pattern: EXPRESSION `?`
// Output: __dingo_try_N(EXPRESSION)

func transformErrorProp(expr string, id int) string {
    return fmt.Sprintf("__dingo_try_%d(%s)", id, expr)
}
```

Later, AST transformation replaces `__dingo_try_N()` with proper error handling:

```go
// __dingo_try_1(os.Open("file.txt"))
// ↓
value, err := os.Open("file.txt")
if err != nil {
    return err  // or wrap in Result type
}
```

#### 2. Pattern Matching

```go
// Pattern: match VALUE { CASE => EXPR, ... }
// Output: func() T { switch VALUE.(type) { case C: return EXPR; ... } }()

func transformMatch(value string, cases []MatchCase) string {
    var sb strings.Builder
    sb.WriteString("func() ")
    sb.WriteString(inferReturnType(cases))
    sb.WriteString(" { switch __dingo_match := ")
    sb.WriteString(value)
    sb.WriteString(".(type) {")

    for _, c := range cases {
        sb.WriteString("case ")
        sb.WriteString(c.Pattern)
        sb.WriteString(": ")
        if c.Binding != "" {
            sb.WriteString(c.Binding)
            sb.WriteString(" := __dingo_match.value; ")
        }
        sb.WriteString("return ")
        sb.WriteString(c.Expression)
        sb.WriteString(";")
    }

    sb.WriteString("default: panic(\"non-exhaustive match\"); }}")
    return sb.String()
}
```

#### 3. Lambda Expressions

```go
// Pattern: |ARGS| EXPR
// Output: func(ARGS) { return EXPR }

func transformLambda(args string, expr string) string {
    // Infer types from context (or use `any` and let Go infer)
    inferredSig := inferLambdaSignature(args, expr)
    return fmt.Sprintf("func(%s) { return %s }", inferredSig, expr)
}
```

### Source Mapping

```go
type SourceMapping struct {
    DingoStart int  // Position in .dingo file
    DingoEnd   int
    GoStart    int  // Position in generated .go file
    GoEnd      int
}

// When go/parser reports error at position P in .go file:
// 1. Find mapping where GoStart <= P < GoEnd
// 2. Map to Dingo position: DingoPos = DingoStart + (P - GoStart)
// 3. Report error at Dingo position
```

---

## Recommendation

**Strategy 1: go/parser + Preprocessor** is the clear winner.

### Why?

1. **Leverage Existing Tools**: Don't reimplement what Go already provides
2. **Battle-Tested**: go/parser is used by all Go tooling
3. **Maintainable**: Simple transformations are easier to understand and debug
4. **Future-Proof**: New Go features automatically work
5. **Small Codebase**: ~2,300 lines vs 10,000+ for alternatives
6. **Fast Development**: 2-3 weeks to working prototype
7. **Industry Standard**: PREGO, gpp, templ all use similar approaches

### Why Not Others?

- **Strategy 2 (Hybrid)**: Too complex, no precedent, reinvents go/parser
- **Strategy 3 (AST Injection)**: Too limited, compromises Dingo syntax
- **Strategy 4 (Tree-sitter)**: Adds complexity without benefit for transpiler use case

### Implementation Roadmap

**Week 1: Preprocessor MVP**
- Text-based `?` operator transformation
- Simple lambda transformation
- Integration with go/parser
- Basic source mapping

**Week 2: AST Transformation**
- Replace `__dingo_try_N()` with proper error handling
- Optimize generated code
- Add Result/Option type generation

**Week 3: Pattern Matching**
- Text-based match → switch transformation
- AST refinement for exhaustiveness checking
- Sum type generation

**Week 4: Polish**
- Error message mapping
- Source map JSON output
- CLI integration

---

## Appendix: Research Sources

### Go Standard Library
- `go/parser` package documentation
- `go/scanner` package documentation
- `go/ast` package documentation
- `golang.org/x/tools/go/ast/astutil` package

### Tools & Precedents
- templ (github.com/a-h/templ) - parser/v2 implementation
- PREGO (github.com/strickyak/prego) - Go preprocessor
- gpp (github.com/mmirolim/gpp) - Go macro preprocessor
- Borgo (github.com/borgo-lang/borgo) - Rust-like → Go transpiler

### Articles & Guides
- "Rewriting Go source code with AST tooling" - Eli Bendersky
- "Understanding Go programs with go/parser" - Francesc Campoy
- "Instrumenting Go code via AST" - Mattermost Engineering

### Tree-sitter Resources
- Tree-sitter documentation (tree-sitter.github.io)
- tree-sitter-go grammar (github.com/tree-sitter/tree-sitter-go)
- go-tree-sitter bindings (github.com/smacker/go-tree-sitter)

---

**Conclusion**: The preprocessor + go/parser approach is the pragmatic, proven, and maintainable solution for Dingo. It respects the principle of "use what Go already provides" while giving us full freedom to design elegant Dingo syntax.

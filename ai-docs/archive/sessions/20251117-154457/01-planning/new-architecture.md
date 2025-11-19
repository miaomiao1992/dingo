# Dingo Preprocessor + go/parser Architecture

**Date:** 2025-11-17
**Status:** Design Proposal
**Purpose:** Complete architectural redesign replacing Participle with go/parser + preprocessor

---

## Executive Summary

This architecture introduces a **preprocessor layer** that transforms Dingo syntax into valid Go code, which is then parsed using the standard `go/parser`. AST transformations replace preprocessor placeholders with final Go implementations. This approach leverages Go's mature parsing infrastructure while cleanly separating Dingo-specific syntax handling.

**Key Innovation:** Instead of maintaining a separate grammar, we transform Dingo syntax to valid Go syntax with semantic placeholders, parse with go/parser, then transform the AST.

---

## Architecture Overview

```
┌─────────────────┐
│  .dingo source  │
└────────┬────────┘
         │
         ▼
┌─────────────────────────────────┐
│  PREPROCESSOR LAYER             │
│  - Scan for Dingo syntax        │
│  - Transform to valid Go        │
│  - Insert semantic placeholders │
│  - Track position mappings      │
└────────┬────────────────────────┘
         │ Valid Go source (with markers)
         │ + SourceMap
         ▼
┌─────────────────────────────────┐
│  go/parser                      │
│  - Parse Go syntax              │
│  - Build AST                    │
│  - Standard Go error reporting  │
└────────┬────────────────────────┘
         │ *ast.File
         │
         ▼
┌─────────────────────────────────┐
│  AST TRANSFORMATION LAYER       │
│  - Find placeholder patterns    │
│  - Type inference               │
│  - Replace with final Go code   │
│  - Update source maps           │
└────────┬────────────────────────┘
         │ Transformed *ast.File
         │
         ▼
┌─────────────────────────────────┐
│  go/printer                     │
│  - Generate .go file            │
│  - Apply formatting             │
│  - Output source maps           │
└────────┬────────────────────────┘
         │
         ▼
┌─────────────────┐
│  .go output     │
│  .go.map        │
└─────────────────┘
```

---

## Component Details

### 1. Preprocessor Layer

**Purpose:** Transform Dingo syntax into syntactically valid Go code that preserves semantic intent through placeholders.

#### Design Principles

1. **Minimal Transformation:** Change only what's necessary to make code parseable by go/parser
2. **Structure Preservation:** Keep expression/statement structure intact when possible
3. **Semantic Markers:** Use placeholder functions/identifiers that are unique and detectable
4. **Position Tracking:** Maintain bidirectional mappings between original and preprocessed positions

#### Transformation Strategy by Feature

##### Error Propagation (`expr?`)

```go
// Dingo Input
result := fetchData(url)?
count := len(data?) + 1

// Preprocessed Output (valid Go)
result := __dingo_try_1__(fetchData(url))
count := len(__dingo_try_2__(data)) + 1
```

**Rationale:**
- `expr?` is not valid Go syntax (postfix operator)
- Wrap in marker function call `__dingo_try_N__(expr)`
- Expression structure preserved (fetchData(url) remains intact)
- Unique ID (N) enables tracking multiple instances
- AST transformer will replace CallExpr with proper error handling

##### Lambdas (`|params| body`)

```go
// Dingo Input
numbers.Map(|x| x * 2)
pairs.Filter(|k, v| v > 10)

// Preprocessed Output (valid Go)
numbers.Map(__dingo_lambda_1__([]string{"x"}, func() interface{} { return x * 2 }))
pairs.Filter(__dingo_lambda_2__([]string{"k", "v"}, func() interface{} { return v > 10 }))
```

**Rationale:**
- `|x| expr` is not valid Go syntax
- Marker function receives:
  - Parameter names as string slice (for reconstruction)
  - Function literal with placeholder return type interface{}
  - Body expression intact for AST analysis
- AST transformer will infer types and rebuild function signature
- Alternative considered: Generate func with `__dingo_type_N__` placeholders, but string slice is cleaner

##### Pattern Matching (`match`)

```go
// Dingo Input
match shape {
    Circle(r) => 3.14 * r * r,
    Rectangle(w, h) => w * h,
}

// Preprocessed Output (valid Go)
__dingo_match_1__(shape, []__dingo_match_case__{
    {Pattern: "Circle(r)", Handler: func(__dingo_binding_r__ interface{}) interface{} { return 3.14 * __dingo_binding_r__ * __dingo_binding_r__ }},
    {Pattern: "Rectangle(w,h)", Handler: func(__dingo_binding_w__, __dingo_binding_h__ interface{}) interface{} { return __dingo_binding_w__ * __dingo_binding_h__ }},
})
```

**Rationale:**
- `match` keyword and pattern syntax not valid Go
- Encode as function call with structured data
- Pattern as string (parsed later during AST transformation)
- Handler as function literal (preserves body AST)
- Bindings prefixed with `__dingo_binding_` for identification
- AST transformer will build proper switch/type-switch

##### Sum Types (`enum`)

```go
// Dingo Input
enum Shape {
    Circle { radius: float64 },
    Rectangle { width: float64, height: float64 },
}

// Preprocessed Output (valid Go)
type __dingo_enum_Shape__ struct {
    __dingo_tag__ string
    __dingo_data__ interface{}
}

var __dingo_enum_def_Shape__ = __dingo_enum_definition__{
    Name: "Shape",
    Variants: []__dingo_variant__{
        {Name: "Circle", Fields: []__dingo_field__{{"radius", "float64"}}},
        {Name: "Rectangle", Fields: []__dingo_field__{{"width", "float64"}, {"height", "float64"}}},
    },
}
```

**Rationale:**
- `enum` keyword not valid Go
- Generate placeholder type definition (valid Go struct)
- Separate variable holds variant metadata (parsed from strings later)
- AST transformer will generate proper tagged union implementation
- Allows go/parser to successfully parse the file

##### Operators (Ternary, Null Coalescing, Safe Navigation)

```go
// Dingo Input
x := cond ? a : b
y := maybeNil ?? defaultValue
z := obj?.field?.method()

// Preprocessed Output (valid Go)
x := func() T { if cond { return a } else { return b } }()
y := func() T { if __tmp := maybeNil; __tmp != nil { return __tmp } else { return defaultValue } }()
z := __dingo_safe_nav_1__(obj, "field", "method()")
```

**Rationale:**
- Ternary and null coalescing can be fully expanded (no type info needed)
- Safe navigation needs type information, use marker function
- Immediate function execution makes them expressions (valid in any expression context)
- AST transformer can optimize or further transform if needed

#### Source Map Tracking

```go
type SourceMap struct {
    // Maps preprocessed position -> original position
    Mappings []Mapping
}

type Mapping struct {
    PreprocessedLine   int
    PreprocessedColumn int
    OriginalLine       int
    OriginalColumn     int
    Length             int
    Name               string // optional symbol name
}
```

**Strategy:**
- Record every transformation's position impact
- When inserting `__dingo_try_1__(`, record offset shift
- When expanding ternary across multiple lines, record line mappings
- Use for error message translation (preprocessed errors -> original positions)
- Use for LSP position translation (editor <-> compiler)

**Example:**
```go
// Original Dingo (line 5, col 12)
result := fetchData()?

// Preprocessed (line 5, col 12)
result := __dingo_try_1__(fetchData())
          ^-- insertion at col 10 (before expr)

// Mapping:
{PreprocessedLine: 5, PreprocessedColumn: 10, OriginalLine: 5, OriginalColumn: 12, Length: 17, Name: "try"}
{PreprocessedLine: 5, PreprocessedColumn: 27, OriginalLine: 5, OriginalColumn: 12, Length: 1, Name: "question_mark"}
```

---

### 2. go/parser Integration

**Purpose:** Parse preprocessed Go code using standard library, providing rich AST.

```go
package parser

import (
    "go/ast"
    "go/parser"
    "go/token"
    "github.com/yourusername/dingo/pkg/preprocessor"
)

type Parser struct {
    fset *token.FileSet
}

func New() *Parser {
    return &Parser{fset: token.NewFileSet()}
}

func (p *Parser) Parse(filename string, dingoSource []byte) (*ast.File, *preprocessor.SourceMap, error) {
    // Step 1: Preprocess
    prep := preprocessor.New(dingoSource)
    goSource, sourceMap, err := prep.Process()
    if err != nil {
        return nil, nil, fmt.Errorf("preprocessing failed: %w", err)
    }

    // Step 2: Parse with go/parser
    file, err := parser.ParseFile(p.fset, filename, goSource, parser.ParseComments)
    if err != nil {
        // Map error positions back to original .dingo file
        mappedErr := p.mapError(err, sourceMap)
        return nil, sourceMap, mappedErr
    }

    return file, sourceMap, nil
}

func (p *Parser) mapError(err error, sm *preprocessor.SourceMap) error {
    // Extract position from go/parser error
    // Look up in source map
    // Reconstruct error with original position
    // Return user-friendly error pointing to .dingo file
}
```

**Benefits:**
- Free, battle-tested Go parser
- Automatic handling of Go syntax edge cases
- Rich AST with all node types
- Excellent error messages (after position mapping)
- No grammar maintenance

**Challenges:**
- Error positions reference preprocessed code (solved by source maps)
- Must ensure preprocessed code is always valid Go (preprocessor responsibility)

---

### 3. AST Transformation Layer

**Purpose:** Walk parsed AST, identify placeholders, replace with final Go implementations.

```go
package transform

import (
    "go/ast"
    "go/token"
    "go/types"
    "golang.org/x/tools/go/ast/astutil"
)

type Transformer struct {
    fset      *token.FileSet
    typeInfo  *types.Info
    sourceMap *preprocessor.SourceMap
}

func New(fset *token.FileSet, sourceMap *preprocessor.SourceMap) *Transformer {
    return &Transformer{
        fset:      fset,
        sourceMap: sourceMap,
        typeInfo:  &types.Info{
            Types: make(map[ast.Expr]types.TypeAndValue),
            Defs:  make(map[*ast.Ident]types.Object),
            Uses:  make(map[*ast.Ident]types.Object),
        },
    }
}

func (t *Transformer) Transform(file *ast.File) (*ast.File, error) {
    // Step 1: Run type checker to populate typeInfo
    if err := t.runTypeChecker(file); err != nil {
        return nil, fmt.Errorf("type checking failed: %w", err)
    }

    // Step 2: Walk AST and transform placeholders
    astutil.Apply(file, t.pre, t.post)

    return file, nil
}

func (t *Transformer) pre(cursor *astutil.Cursor) bool {
    node := cursor.Node()

    switch n := node.(type) {
    case *ast.CallExpr:
        // Check for placeholder function calls
        if ident, ok := n.Fun.(*ast.Ident); ok {
            switch {
            case strings.HasPrefix(ident.Name, "__dingo_try_"):
                t.transformErrorProp(cursor, n)
                return false // Don't descend into transformed node
            case strings.HasPrefix(ident.Name, "__dingo_lambda_"):
                t.transformLambda(cursor, n)
                return false
            case strings.HasPrefix(ident.Name, "__dingo_match_"):
                t.transformMatch(cursor, n)
                return false
            case strings.HasPrefix(ident.Name, "__dingo_safe_nav_"):
                t.transformSafeNav(cursor, n)
                return false
            }
        }

    case *ast.GenDecl:
        // Check for enum type definitions
        if t.isEnumDefinition(n) {
            t.transformEnum(cursor, n)
            return false
        }
    }

    return true // Continue traversal
}

func (t *Transformer) post(cursor *astutil.Cursor) bool {
    // Post-order processing if needed
    return true
}
```

#### Transformation Details

##### Error Propagation

```go
func (t *Transformer) transformErrorProp(cursor *astutil.Cursor, call *ast.CallExpr) {
    // Input AST: __dingo_try_1__(fetchData(url))
    // Call.Args[0] is the expression: fetchData(url)

    expr := call.Args[0]

    // Determine context: assignment, return, standalone
    context := t.analyzeContext(cursor)

    if context == ContextAssignment {
        // Generate:
        // __tmp, __err := fetchData(url)
        // if __err != nil {
        //     return __zero, __err
        // }
        // result := __tmp

        nodes := t.generateErrorPropagation(expr, context)
        cursor.Replace(nodes)
    }

    // Update source map to point generated code back to original ?
}
```

##### Lambda Type Inference

```go
func (t *Transformer) transformLambda(cursor *astutil.Cursor, call *ast.CallExpr) {
    // Input: __dingo_lambda_1__([]string{"x", "y"}, func() interface{} { return x + y })

    paramNames := extractParamNames(call.Args[0]) // ["x", "y"]
    funcLit := call.Args[1].(*ast.FuncLit)        // The function literal

    // Infer types from context
    expectedType := t.inferLambdaType(cursor)

    // Rebuild function literal with proper types
    newFunc := &ast.FuncLit{
        Type: &ast.FuncType{
            Params: &ast.FieldList{
                List: []*ast.Field{
                    {Names: []*ast.Ident{{Name: "x"}}, Type: expectedType.ParamType(0)},
                    {Names: []*ast.Ident{{Name: "y"}}, Type: expectedType.ParamType(1)},
                },
            },
            Results: &ast.FieldList{
                List: []*ast.Field{
                    {Type: expectedType.ResultType()},
                },
            },
        },
        Body: funcLit.Body, // Reuse original body
    }

    cursor.Replace(newFunc)
}

func (t *Transformer) inferLambdaType(cursor *astutil.Cursor) *types.Signature {
    // Walk up AST to find CallExpr this lambda is argument to
    // Look up method signature in typeInfo
    // Extract parameter's function type
    // Return as *types.Signature
}
```

##### Pattern Matching

```go
func (t *Transformer) transformMatch(cursor *astutil.Cursor, call *ast.CallExpr) {
    // Input: __dingo_match_1__(shape, []__dingo_match_case__{...})

    scrutinee := call.Args[0]           // The value being matched
    cases := call.Args[1]               // Case slice

    // Determine if this is value match or type match
    scrutineeType := t.typeInfo.TypeOf(scrutinee)

    if isEnumType(scrutineeType) {
        // Generate type switch for sum type
        switchStmt := t.generateTypeSwitchMatch(scrutinee, cases)
        cursor.Replace(switchStmt)
    } else {
        // Generate regular switch for value matching
        switchStmt := t.generateValueSwitchMatch(scrutinee, cases)
        cursor.Replace(switchStmt)
    }
}

func (t *Transformer) generateTypeSwitchMatch(scrutinee ast.Expr, cases ast.Expr) *ast.TypeSwitchStmt {
    // Build:
    // switch v := scrutinee.(type) {
    // case CircleVariant:
    //     r := v.Radius
    //     <handler body>
    // case RectangleVariant:
    //     w, h := v.Width, v.Height
    //     <handler body>
    // }
}
```

---

### 4. Code Generation

**Purpose:** Use go/printer to generate final .go files with proper formatting.

```go
package generator

import (
    "go/ast"
    "go/printer"
    "go/token"
    "os"
)

type Generator struct {
    fset *token.FileSet
}

func (g *Generator) Generate(file *ast.File, outputPath string) error {
    // Open output file
    f, err := os.Create(outputPath)
    if err != nil {
        return err
    }
    defer f.Close()

    // Configure printer
    cfg := &printer.Config{
        Mode:     printer.TabIndent | printer.UseSpaces,
        Tabwidth: 4,
    }

    // Print AST to file
    if err := cfg.Fprint(f, g.fset, file); err != nil {
        return err
    }

    return nil
}

func (g *Generator) GenerateSourceMap(sourceMap *preprocessor.SourceMap, outputPath string) error {
    // Serialize source map to .go.map file
    // Format: JSON or custom binary format
}
```

---

### 5. Plugin System Integration

**Current State:**
- Plugins implement `Transform(ast.Node) ast.Node`
- Each plugin handles one feature (error_prop, sum_types, lambdas, etc.)
- Plugins are independent and composable

**New Architecture:**

Each feature now has TWO components:

1. **Preprocessor** (`pkg/preprocessor/feature.go`)
   - Transforms Dingo syntax -> valid Go with markers
   - Returns position mappings

2. **Transformer** (`pkg/transform/feature.go`)
   - Replaces markers with final Go code
   - Uses type information
   - Updates source maps

**Migration of Existing Plugins:**

| Old Plugin | New Preprocessor | New Transformer |
|------------|------------------|-----------------|
| `plugins/error_prop.go` | `preprocessor/error_prop.go` | `transform/error_prop.go` |
| `plugins/sum_types.go` | `preprocessor/sum_types.go` | `transform/sum_types.go` |
| `plugins/lambdas.go` | `preprocessor/lambdas.go` | `transform/lambdas.go` |
| `plugins/pattern_match.go` | `preprocessor/pattern_match.go` | `transform/pattern_match.go` |
| `plugins/ternary.go` | `preprocessor/operators.go` (full expansion) | (none needed) |

**Orchestration:**

```go
// pkg/preprocessor/preprocessor.go
type Preprocessor struct {
    processors []FeatureProcessor
}

type FeatureProcessor interface {
    Name() string
    Process(source []byte) ([]byte, []Mapping, error)
}

// Register all features
func New(source []byte) *Preprocessor {
    return &Preprocessor{
        processors: []FeatureProcessor{
            NewErrorPropProcessor(),
            NewLambdaProcessor(),
            NewPatternMatchProcessor(),
            NewSumTypeProcessor(),
            NewOperatorProcessor(),
        },
    }
}

func (p *Preprocessor) Process() (string, *SourceMap, error) {
    result := p.source
    allMappings := []Mapping{}

    // Run each processor in sequence
    for _, proc := range p.processors {
        processed, mappings, err := proc.Process(result)
        if err != nil {
            return "", nil, err
        }
        result = processed
        allMappings = append(allMappings, mappings...)
    }

    return string(result), &SourceMap{Mappings: allMappings}, nil
}
```

```go
// pkg/transform/transformer.go
type Transformer struct {
    transformers []FeatureTransformer
}

type FeatureTransformer interface {
    Name() string
    Transform(cursor *astutil.Cursor) bool
}

func New(fset *token.FileSet, sourceMap *preprocessor.SourceMap) *Transformer {
    return &Transformer{
        transformers: []FeatureTransformer{
            NewErrorPropTransformer(fset, sourceMap),
            NewLambdaTransformer(fset, sourceMap),
            NewPatternMatchTransformer(fset, sourceMap),
            NewSumTypeTransformer(fset, sourceMap),
        },
    }
}
```

---

## File Structure

```
dingo/
├── pkg/
│   ├── preprocessor/
│   │   ├── preprocessor.go          # Main orchestration, FeatureProcessor interface
│   │   ├── error_prop.go            # Error propagation: expr? -> __dingo_try_N__
│   │   ├── lambdas.go               # Lambdas: |x| expr -> __dingo_lambda_N__
│   │   ├── pattern_match.go         # Match: match -> __dingo_match_N__
│   │   ├── sum_types.go             # Enum: enum -> type __dingo_enum_Name__
│   │   ├── operators.go             # Ternary, ??, ?. preprocessing
│   │   ├── sourcemap.go             # SourceMap struct, Mapping operations
│   │   └── preprocessor_test.go     # Unit tests for each feature
│   │
│   ├── transform/
│   │   ├── transformer.go           # Main orchestration, AST walking
│   │   ├── error_prop.go            # Replace __dingo_try__ with error handling
│   │   ├── lambdas.go               # Type inference and lambda reconstruction
│   │   ├── pattern_match.go         # Generate switch statements
│   │   ├── sum_types.go             # Generate tagged union implementations
│   │   ├── type_inference.go        # Shared type inference utilities
│   │   ├── context.go               # Analyze AST context (assignment, return, etc.)
│   │   └── transform_test.go        # Unit tests for transformations
│   │
│   ├── parser/
│   │   ├── parser.go                # Wrapper: preprocess + go/parser + error mapping
│   │   └── errors.go                # Error position mapping helpers
│   │
│   ├── generator/
│   │   ├── generator.go             # go/printer wrapper
│   │   └── sourcemap.go             # Source map serialization
│   │
│   ├── compiler/
│   │   └── compiler.go              # High-level: orchestrates parse -> transform -> generate
│   │
│   └── plugins/                     # DEPRECATED - to be removed after migration
│       └── (old plugin files)
│
├── cmd/
│   └── dingo/
│       ├── main.go
│       └── build.go                 # Uses pkg/compiler
│
├── tests/
│   └── golden/                      # Golden tests validate entire pipeline
│       ├── error_prop_*.dingo
│       ├── error_prop_*.go.golden
│       └── ...
│
└── ai-docs/
    └── sessions/
        └── 20251117-154457/
            └── 01-planning/
                ├── new-architecture.md (this file)
                ├── migration-plan.md
                ├── implementation-phases.md
                └── architecture-reasoning.md
```

---

## Data Flow Example

**Input: error_prop_01_simple.dingo**

```go
package main

import "errors"

func fetchData() (string, error) {
    return "", errors.New("network error")
}

func process() error {
    data := fetchData()?
    println(data)
    return nil
}
```

**Step 1: Preprocessor Output**

```go
package main

import "errors"

func fetchData() (string, error) {
    return "", errors.New("network error")
}

func process() error {
    data := __dingo_try_1__(fetchData())
    println(data)
    return nil
}
```

SourceMap: `{PreprocessedLine: 10, PreprocessedColumn: 11, OriginalLine: 10, OriginalColumn: 18, Length: 1, Name: "error_prop"}`

**Step 2: go/parser Output**

AST for `process()` function:
- FuncDecl
  - Body: BlockStmt
    - AssignStmt: `data := __dingo_try_1__(fetchData())`
      - Lhs: Ident("data")
      - Rhs: CallExpr
        - Fun: Ident("__dingo_try_1__")
        - Args: [CallExpr(Ident("fetchData"))]

**Step 3: AST Transformation**

Transformer identifies `__dingo_try_1__` call, analyzes context (assignment in function returning error), generates:

```go
func process() error {
    __tmp_1, __err_1 := fetchData()
    if __err_1 != nil {
        return __err_1
    }
    data := __tmp_1
    println(data)
    return nil
}
```

**Step 4: Code Generation**

go/printer formats and writes to `error_prop_01_simple.go`.

---

## Key Advantages

1. **Leverage Standard Library:** go/parser is mature, fast, well-tested
2. **No Grammar Maintenance:** No Participle grammar to keep in sync
3. **Rich Type Information:** go/types integration for advanced transformations
4. **Better Error Messages:** Standard Go errors (after position mapping)
5. **Simpler Codebase:** Clear separation of concerns
6. **Incremental Migration:** Can adopt feature by feature
7. **Extensibility:** Easy to add new features (add preprocessor + transformer)
8. **Testability:** Each component independently testable

## Trade-offs

1. **Two-pass Processing:** Preprocess then parse (vs single-pass Participle)
   - Mitigation: Preprocessing is fast (string manipulation)

2. **Source Map Complexity:** Must track positions through transformations
   - Mitigation: Structured approach, extensive testing

3. **Placeholder Namespace:** Must avoid collisions with user code
   - Mitigation: Use `__dingo_` prefix (Go convention for internal)

4. **Error Mapping Overhead:** Must translate positions for errors
   - Mitigation: Caching, efficient data structures

---

## Open Questions

1. **Source Map Format:** JSON (readable) vs binary (compact)?
   - Recommendation: JSON for v1 (debugging), binary for v2 (production)

2. **Type Inference Fallbacks:** What if inference fails?
   - Recommendation: Require explicit types, emit clear error

3. **Preprocessor Ordering:** Does order of feature processors matter?
   - Recommendation: Define canonical ordering, document dependencies

4. **Incremental Compilation:** How to cache preprocessing results?
   - Recommendation: Hash source, cache preprocessed output

5. **LSP Integration:** How does language server use this?
   - Recommendation: LSP uses same preprocessor, source maps for all positions

---

## Success Criteria

1. All 10 existing golden tests pass
2. Error messages point to original .dingo files (correct line/column)
3. Performance: < 100ms for typical file (< 500 lines)
4. Code quality: Generated Go is idiomatic and readable
5. Maintainability: Clear component boundaries, < 200 lines per file
6. Extensibility: Adding new feature takes < 1 day

---

**Next Steps:** See `migration-plan.md` and `implementation-phases.md` for detailed execution strategy.

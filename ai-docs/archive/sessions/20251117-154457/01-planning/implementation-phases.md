# Implementation Phases: Detailed Breakdown

**Date:** 2025-11-17
**Total Duration:** 32 days (6-7 weeks with buffer)
**Risk Level:** Medium
**Reversibility:** High (parallel development, feature flags)

---

## Phase 0: Infrastructure Setup

**Duration:** 2-3 days
**Risk:** Low
**Reversibility:** Perfect (no existing code touched)

### Goal

Build foundational packages and test harnesses without modifying existing implementation. Validate that go/parser integration works on simple cases.

### Tasks

#### Task 0.1: Create Package Structure (2 hours)

```bash
mkdir -p pkg/preprocessor
mkdir -p pkg/transform
mkdir -p pkg/parser2
mkdir -p pkg/generator2
```

Create stub files:

```go
// pkg/preprocessor/preprocessor.go
package preprocessor

type Preprocessor struct {
    source []byte
}

func New(source []byte) *Preprocessor {
    return &Preprocessor{source: source}
}

func (p *Preprocessor) Process() (string, *SourceMap, error) {
    // TODO: Implement
    return string(p.source), &SourceMap{}, nil
}
```

**Deliverable:** Package skeleton with placeholder implementations.

#### Task 0.2: Implement SourceMap Infrastructure (4 hours)

```go
// pkg/preprocessor/sourcemap.go
package preprocessor

type SourceMap struct {
    Mappings []Mapping
}

type Mapping struct {
    PreprocessedLine   int
    PreprocessedColumn int
    OriginalLine       int
    OriginalColumn     int
    Length             int
    Name               string
}

func (sm *SourceMap) MapPosition(line, col int) (origLine, origCol int) {
    // Binary search through mappings
    // Return original position
}

func (sm *SourceMap) AddMapping(m Mapping) {
    sm.Mappings = append(sm.Mappings, m)
}

func (sm *SourceMap) Serialize() ([]byte, error) {
    // JSON serialization for MVP
    return json.Marshal(sm)
}
```

**Tests:**

```go
// pkg/preprocessor/sourcemap_test.go
func TestSourceMapMapping(t *testing.T) {
    sm := &SourceMap{
        Mappings: []Mapping{
            {PreprocessedLine: 10, PreprocessedColumn: 15, OriginalLine: 10, OriginalColumn: 18, Length: 1},
        },
    }

    origLine, origCol := sm.MapPosition(10, 16)
    assert.Equal(t, 10, origLine)
    assert.Equal(t, 19, origCol)
}
```

**Deliverable:** Fully tested SourceMap implementation.

#### Task 0.3: Implement go/parser Wrapper (4 hours)

```go
// pkg/parser2/parser.go
package parser2

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

func (p *Parser) Parse(filename string, source []byte) (*ast.File, *preprocessor.SourceMap, error) {
    // Step 1: Preprocess
    prep := preprocessor.New(source)
    goSource, sourceMap, err := prep.Process()
    if err != nil {
        return nil, nil, err
    }

    // Step 2: Parse with go/parser
    file, err := parser.ParseFile(p.fset, filename, goSource, parser.ParseComments)
    if err != nil {
        return nil, sourceMap, p.mapError(err, sourceMap)
    }

    return file, sourceMap, nil
}

func (p *Parser) mapError(err error, sm *preprocessor.SourceMap) error {
    // TODO: Extract position from err, map back to original
    return err
}
```

**Tests:**

```go
// pkg/parser2/parser_test.go
func TestParseSimpleGoFile(t *testing.T) {
    source := []byte(`
package main

func main() {
    println("hello")
}
`)

    p := New()
    file, sourceMap, err := p.Parse("test.go", source)

    require.NoError(t, err)
    assert.NotNil(t, file)
    assert.NotNil(t, sourceMap)
    assert.Equal(t, "main", file.Name.Name)
}
```

**Deliverable:** Parser that can parse plain Go (passthrough preprocessing).

#### Task 0.4: Create Test Harness (4 hours)

```go
// pkg/testutil/harness.go
package testutil

import (
    "testing"
    "github.com/yourusername/dingo/pkg/parser2"
    "github.com/yourusername/dingo/pkg/transform"
    "github.com/yourusername/dingo/pkg/generator2"
)

type TestCase struct {
    Name     string
    Input    string  // .dingo source
    Expected string  // .go output
}

func RunTestCase(t *testing.T, tc TestCase) {
    // Parse
    p := parser2.New()
    file, sourceMap, err := p.Parse(tc.Name, []byte(tc.Input))
    require.NoError(t, err)

    // Transform
    tr := transform.New(p.FileSet(), sourceMap)
    transformed, err := tr.Transform(file)
    require.NoError(t, err)

    // Generate
    gen := generator2.New(p.FileSet())
    output := gen.GenerateString(transformed)

    // Compare
    assert.Equal(t, tc.Expected, output)
}
```

**Deliverable:** Reusable test harness for integration tests.

### Milestones

- [ ] Package structure created
- [ ] SourceMap implementation complete and tested
- [ ] Parser wrapper can parse plain Go
- [ ] Test harness ready for use

### Success Criteria

- All unit tests pass
- Can parse simple Go file through new pipeline
- Source map correctly tracks positions (even for passthrough)
- Zero impact on existing code (no files modified outside pkg/)

### Estimated Time: 14 hours = ~2 days

---

## Phase 1: Error Propagation Migration

**Duration:** 3-4 days
**Risk:** Medium (first real feature)
**Reversibility:** High (feature flag)

### Goal

Fully implement error propagation (`?` operator) using new architecture. Validate approach works end-to-end.

### Tasks

#### Task 1.1: Implement Error Propagation Preprocessor (6 hours)

```go
// pkg/preprocessor/error_prop.go
package preprocessor

import (
    "bytes"
    "fmt"
    "go/scanner"
    "go/token"
)

type ErrorPropProcessor struct {
    counter int
}

func (p *ErrorPropProcessor) Process(source []byte) ([]byte, []Mapping, error) {
    var output bytes.Buffer
    var mappings []Mapping

    // Tokenize and scan for ? operator
    fset := token.NewFileSet()
    file := fset.AddFile("", fset.Base(), len(source))

    s := scanner.Scanner{}
    s.Init(file, source, nil, 0)

    lastPos := 0
    for {
        pos, tok, lit := s.Scan()
        if tok == token.EOF {
            break
        }

        // Look for IDENT/RPAREN followed by QUESTION
        if p.isErrorPropCandidate(tok) {
            nextPos, nextTok, _ := s.Scan()
            if nextTok == token.QUESTION {
                // Transform expr? -> __dingo_try_N__(expr)
                p.counter++

                // Write everything up to expr
                output.Write(source[lastPos:fset.Position(pos).Offset])

                // Write transformation
                transformed := fmt.Sprintf("__dingo_try_%d__(", p.counter)
                output.WriteString(transformed)

                // Extract expr
                exprStart := fset.Position(pos).Offset
                exprEnd := fset.Position(nextPos).Offset - 1 // Before ?
                output.Write(source[exprStart:exprEnd])
                output.WriteString(")")

                // Record mapping
                mapping := Mapping{
                    PreprocessedLine:   fset.Position(pos).Line,
                    PreprocessedColumn: fset.Position(pos).Column,
                    OriginalLine:       fset.Position(pos).Line,
                    OriginalColumn:     fset.Position(pos).Column,
                    Length:             exprEnd - exprStart,
                    Name:               "error_prop",
                }
                mappings = append(mappings, mapping)

                lastPos = fset.Position(nextPos).Offset + 1 // After ?
                continue
            }
        }
    }

    // Write remaining
    output.Write(source[lastPos:])

    return output.Bytes(), mappings, nil
}

func (p *ErrorPropProcessor) isErrorPropCandidate(tok token.Token) bool {
    return tok == token.IDENT || tok == token.RPAREN
}
```

**Tests:**

```go
func TestErrorPropPreprocessor(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {
            name:     "simple error prop",
            input:    "x := f()?",
            expected: "x := __dingo_try_1__(f())",
        },
        {
            name:     "chained error prop",
            input:    "y := g(f()?)?",
            expected: "y := __dingo_try_2__(g(__dingo_try_1__(f())))",
        },
        {
            name:     "no error prop",
            input:    "x := f()",
            expected: "x := f()",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            proc := &ErrorPropProcessor{}
            output, _, err := proc.Process([]byte(tt.input))
            require.NoError(t, err)
            assert.Equal(t, tt.expected, string(output))
        })
    }
}
```

**Deliverable:** Preprocessor that transforms `expr?` to `__dingo_try_N__(expr)`.

#### Task 1.2: Implement Error Propagation Transformer (8 hours)

```go
// pkg/transform/error_prop.go
package transform

import (
    "go/ast"
    "go/token"
    "golang.org/x/tools/go/ast/astutil"
)

type ErrorPropTransformer struct {
    fset *token.FileSet
}

func (t *ErrorPropTransformer) Transform(cursor *astutil.Cursor) bool {
    node := cursor.Node()

    if call, ok := node.(*ast.CallExpr); ok {
        if ident, ok := call.Fun.(*ast.Ident); ok {
            if strings.HasPrefix(ident.Name, "__dingo_try_") {
                t.transformTryCall(cursor, call)
                return false
            }
        }
    }

    return true
}

func (t *ErrorPropTransformer) transformTryCall(cursor *astutil.Cursor, call *ast.CallExpr) {
    // Extract expression from __dingo_try_N__(expr)
    expr := call.Args[0]

    // Analyze context
    context := t.analyzeContext(cursor)

    switch context.Type {
    case ContextAssignment:
        t.replaceWithAssignment(cursor, expr, context)
    case ContextReturn:
        t.replaceWithReturn(cursor, expr, context)
    case ContextStandalone:
        t.replaceWithStandalone(cursor, expr, context)
    }
}

func (t *ErrorPropTransformer) analyzeContext(cursor *astutil.Cursor) Context {
    // Walk up AST to determine context
    // Check parent nodes: AssignStmt, ReturnStmt, ExprStmt, etc.
}

func (t *ErrorPropTransformer) replaceWithAssignment(cursor *astutil.Cursor, expr ast.Expr, ctx Context) {
    // Generate:
    // __tmp, __err := expr
    // if __err != nil {
    //     return __zero, __err
    // }
    // lhs := __tmp

    tmpVar := ast.NewIdent(fmt.Sprintf("__tmp_%d", ctx.ID))
    errVar := ast.NewIdent(fmt.Sprintf("__err_%d", ctx.ID))

    // Build assignment: __tmp, __err := expr
    checkStmt := &ast.AssignStmt{
        Lhs: []ast.Expr{tmpVar, errVar},
        Tok: token.DEFINE,
        Rhs: []ast.Expr{expr},
    }

    // Build error check: if __err != nil { return __zero, __err }
    returnStmt := t.buildErrorReturn(ctx, errVar)

    // Replace call with __tmp
    cursor.Replace(tmpVar)

    // Insert check before current statement
    t.insertBefore(ctx.Statement, checkStmt, returnStmt)
}
```

**Tests:**

```go
func TestErrorPropTransformer(t *testing.T) {
    source := `
package main

func fetchData() (string, error) {
    return "", nil
}

func process() error {
    data := __dingo_try_1__(fetchData())
    println(data)
    return nil
}
`

    expected := `
package main

func fetchData() (string, error) {
    return "", nil
}

func process() error {
    __tmp_1, __err_1 := fetchData()
    if __err_1 != nil {
        return __err_1
    }
    data := __tmp_1
    println(data)
    return nil
}
`

    // Parse, transform, compare
}
```

**Deliverable:** Transformer that replaces `__dingo_try__` calls with proper error handling.

#### Task 1.3: Integrate with Preprocessor Orchestration (2 hours)

```go
// pkg/preprocessor/preprocessor.go
func (p *Preprocessor) Process() (string, *SourceMap, error) {
    result := p.source
    allMappings := []Mapping{}

    // Error propagation processor
    errorPropProc := &ErrorPropProcessor{}
    result, mappings, err := errorPropProc.Process(result)
    if err != nil {
        return "", nil, err
    }
    allMappings = append(allMappings, mappings...)

    return string(result), &SourceMap{Mappings: allMappings}, nil
}
```

**Deliverable:** Preprocessor orchestration includes error propagation.

#### Task 1.4: Add Feature Flag to CLI (2 hours)

```go
// cmd/dingo/build.go
var useExperimentalParser bool

func init() {
    buildCmd.Flags().BoolVar(&useExperimentalParser, "experimental-parser", false, "Use new go/parser-based implementation")
}

func buildFile(path string) error {
    if useExperimentalParser {
        return buildFileV2(path)
    }
    return buildFileV1(path)
}

func buildFileV2(path string) error {
    source, err := os.ReadFile(path)
    if err != nil {
        return err
    }

    // Parse
    p := parser2.New()
    file, sourceMap, err := p.Parse(path, source)
    if err != nil {
        return err
    }

    // Transform
    tr := transform.New(p.FileSet(), sourceMap)
    transformed, err := tr.Transform(file)
    if err != nil {
        return err
    }

    // Generate
    gen := generator2.New(p.FileSet())
    outputPath := strings.ReplaceAll(path, ".dingo", ".go")
    if err := gen.Generate(transformed, outputPath); err != nil {
        return err
    }

    return nil
}
```

**Deliverable:** CLI accepts `--experimental-parser` flag.

#### Task 1.5: Test with Golden Tests (4 hours)

Run all 8 error propagation golden tests:

```bash
$ dingo build --experimental-parser tests/golden/error_prop_01_simple.dingo
$ dingo build --experimental-parser tests/golden/error_prop_02_multiple.dingo
# ... all 8 tests
```

Compare output with expected .go.golden files.

Fix any discrepancies.

**Deliverable:** All 8 error propagation tests pass with new parser.

### Milestones

- [ ] Preprocessor transforms `?` correctly
- [ ] Transformer generates proper error handling
- [ ] CLI feature flag working
- [ ] All 8 golden tests pass

### Success Criteria

- All error propagation tests pass
- Generated code matches .go.golden (or is cleaner)
- Error messages point to .dingo files (correct positions)
- Performance acceptable (< 100ms for typical file)

### Estimated Time: 22 hours = ~3 days

---

## Phase 2: Lambda Migration

**Duration:** 3-4 days
**Risk:** Medium (type inference complexity)
**Reversibility:** High (feature flag still in place)

### Tasks

#### Task 2.1: Implement Lambda Preprocessor (6 hours)

Transform `|x, y| expr` to `__dingo_lambda_N__([]string{"x", "y"}, func() interface{} { return expr })`.

**Complexity:** Parsing lambda syntax correctly, handling nested lambdas.

#### Task 2.2: Implement Type Inference (8 hours)

```go
// pkg/transform/type_inference.go
func (t *Transformer) inferLambdaType(cursor *astutil.Cursor) *types.Signature {
    // Walk up to CallExpr
    // Look up function signature
    // Match parameter position
    // Return expected type
}
```

**Complexity:** Walking AST correctly, handling edge cases.

#### Task 2.3: Implement Lambda Transformer (6 hours)

Replace `__dingo_lambda__` calls with properly-typed function literals.

#### Task 2.4: Test with 3 Lambda Golden Tests (2 hours)

### Estimated Time: 22 hours = ~3 days

---

## Phase 3: Simple Operators

**Duration:** 2 days
**Risk:** Low (straightforward expansion)
**Reversibility:** High

### Tasks

#### Task 3.1: Implement Operator Preprocessor (6 hours)

- Ternary: `cond ? a : b`
- Null coalescing: `a ?? b`
- Safe navigation: `obj?.field`

#### Task 3.2: Test with 6 Operator Golden Tests (2 hours)

### Estimated Time: 8 hours = ~1 day

---

## Phase 4: Pattern Matching

**Duration:** 4-5 days
**Risk:** Medium-High (complex AST generation)
**Reversibility:** High

### Tasks

#### Task 4.1: Implement Pattern Match Preprocessor (8 hours)

Encode `match` as structured function call.

#### Task 4.2: Implement Pattern Match Transformer (12 hours)

Generate switch/type-switch based on scrutinee type.

**Complexity:** Pattern parsing, destructuring, type-switch generation.

#### Task 4.3: Test with 4 Pattern Match Golden Tests (4 hours)

### Estimated Time: 24 hours = ~3-4 days

---

## Phase 5: Sum Types

**Duration:** 5-6 days
**Risk:** High (most complex feature)
**Reversibility:** High

### Tasks

#### Task 5.1: Implement Sum Type Preprocessor (8 hours)

Transform `enum` to placeholder types and metadata.

#### Task 5.2: Implement Sum Type Transformer (16 hours)

Generate tagged union, constructors, type guards, integration with pattern matching.

**Complexity:** Generating multiple related types, ensuring type safety.

#### Task 5.3: Test with 2 Sum Type Golden Tests (4 hours)

### Estimated Time: 28 hours = ~4-5 days

---

## Phase 6: Integration and Cutover

**Duration:** 2-3 days
**Risk:** Low (validation and cleanup)
**Reversibility:** Medium (removing old code is one-way)

### Tasks

#### Task 6.1: Run Full Test Suite (4 hours)

All 23 golden tests with `--experimental-parser`.

#### Task 6.2: Benchmarking (4 hours)

Compare old vs new parser performance.

#### Task 6.3: Switch Default Parser (2 hours)

Remove feature flag, make new parser default, add `--legacy-parser` fallback.

#### Task 6.4: Update Documentation (4 hours)

CLAUDE.md, README.md, ai-docs/.

#### Task 6.5: Deprecate Old Code (2 hours)

Rename packages, update imports.

#### Task 6.6: Delete Old Code (after 1 week stability) (2 hours)

### Estimated Time: 18 hours = ~2-3 days

---

## Total Time Estimate

| Phase | Hours | Days | Buffer (20%) | Total Days |
|-------|-------|------|--------------|------------|
| Phase 0 | 14 | 2 | 0.4 | 2.4 |
| Phase 1 | 22 | 3 | 0.6 | 3.6 |
| Phase 2 | 22 | 3 | 0.6 | 3.6 |
| Phase 3 | 8 | 1 | 0.2 | 1.2 |
| Phase 4 | 24 | 3 | 0.6 | 3.6 |
| Phase 5 | 28 | 4 | 0.8 | 4.8 |
| Phase 6 | 18 | 2.5 | 0.5 | 3.0 |
| **Total** | **136** | **18.5** | **3.7** | **22.2** |

**Rounded Total:** ~23-25 working days (~5 weeks)

**With Conservative Buffer:** ~30-32 calendar days (~6-7 weeks)

---

## Dependencies and Blockers

### External Dependencies

- Go standard library (go/parser, go/ast, go/types) - stable
- golang.org/x/tools/go/ast/astutil - stable
- No external API dependencies

### Internal Dependencies

- Phase 1 → Phase 2: Type inference requires preprocessor framework
- Phase 4 → Phase 5: Pattern matching integrates with sum types
- All phases → Phase 6: Cannot cut over until all features migrated

### Potential Blockers

1. **Type inference edge cases:** May require more time than estimated
   - Mitigation: Start simple, add complexity incrementally

2. **Source map complexity:** Position tracking may be tricky
   - Mitigation: Extensive testing, simplify if needed

3. **Golden test failures:** Generated code may differ
   - Mitigation: Accept cleaner output, update .go.golden if needed

---

## Daily Progress Tracking

Each day, update:

```markdown
## Day N Progress

**Date:** YYYY-MM-DD
**Phase:** Phase N
**Task:** Task N.M

**Completed:**
- [x] Subtask 1
- [x] Subtask 2

**In Progress:**
- [ ] Subtask 3

**Blocked:**
- Issue description and blocker

**Time Spent:** X hours
**Time Remaining:** Y hours

**Notes:**
- Any learnings or decisions
```

Location: `ai-docs/sessions/20251117-154457/02-implementation/daily-log.md`

---

## Quality Gates

Each phase must pass before proceeding:

### Phase 0 Quality Gate

- [ ] All unit tests pass
- [ ] Can parse plain Go file
- [ ] Source map correctly tracks positions
- [ ] Test harness validates end-to-end flow

### Phase 1 Quality Gate

- [ ] All 8 error propagation golden tests pass
- [ ] Generated code compiles and runs
- [ ] Error messages point to correct .dingo positions
- [ ] Performance < 100ms for typical file

### Phase 2 Quality Gate

- [ ] All 3 lambda golden tests pass
- [ ] Type inference handles basic cases
- [ ] No regressions in Phase 1 tests

### Phase 3 Quality Gate

- [ ] All 6 operator golden tests pass
- [ ] No regressions in Phase 1-2 tests

### Phase 4 Quality Gate

- [ ] All 4 pattern match golden tests pass
- [ ] Generated switches are clean
- [ ] No regressions in Phase 1-3 tests

### Phase 5 Quality Gate

- [ ] All 2 sum type golden tests pass
- [ ] Integration with pattern matching works
- [ ] No regressions in Phase 1-4 tests

### Phase 6 Quality Gate

- [ ] All 23 golden tests pass
- [ ] Performance ≤ old parser
- [ ] Documentation updated
- [ ] Old code deprecated/removed

---

**Next Steps:**
1. Approve this implementation plan
2. Set up project tracking (GitHub issues/milestones)
3. Begin Phase 0
4. Track daily progress in daily-log.md
5. Conduct phase reviews at each quality gate

**Estimated Completion:** Mid-January 2026 (starting now)

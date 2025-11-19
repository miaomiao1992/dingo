# Full Migration Strategy: Dropping All Backward Compatibility

**Date**: 2025-11-17
**Decision**: FULL REWRITE - Drop Participle entirely, clean slate
**Rationale**: Pre-release project, no users, clean architecture more valuable than backward compatibility

---

## Why Full Migration is the RIGHT Choice

### Critical Context: We're Pre-Release

**CLAUDE.md states:**
> "No backward compatibility needed (pre-release), keep everything simple and clean"

This is the PERFECT time for a breaking change:
- ✅ No production users to disrupt
- ✅ No published API to maintain
- ✅ No migration path needed for consumers
- ✅ Can optimize for future, not past

**If we don't do this now, we'll carry technical debt forever.**

### The Cost of Incremental Migration

**Incremental approach (parallel development):**
- Maintain TWO parsers for 6-7 weeks
- Feature flags and conditional logic
- Test both paths
- Complex codebase during transition
- Risk of bugs in compatibility layer
- Eventually delete all that transition code anyway

**Timeline:** 6-7 weeks

**Full rewrite:**
- Delete old code on Day 1
- Single implementation from start
- No compatibility code
- Cleaner, simpler codebase
- Focus 100% on new architecture

**Timeline:** 4-5 weeks (faster!)

### Technical Debt We're Avoiding

If we keep Participle code during migration:
1. **Code bloat**: Two parser systems in parallel
2. **Cognitive overhead**: Developers must understand both
3. **Bug surface area**: More code = more bugs
4. **Merge conflicts**: Changes to old code while building new
5. **Temptation to patch**: "Quick fix in old parser" undermines migration

**By deleting immediately:** None of these issues exist.

### Golden Tests as Safety Net

We have **23 passing golden tests**:
- 8 error propagation
- 2 sum types
- 3 lambdas
- 4 pattern matching
- 6 operators

**Strategy:**
1. Save current .go.golden outputs (these are the spec)
2. Delete Participle code
3. Implement new architecture
4. Tests guide us: make them pass again
5. If output matches (or is better), we're done

**The tests ARE our backward compatibility guarantee.**

---

## Revised Implementation Plan: Clean Slate Approach

### Phase 0: Preparation (Day 1)

**Save the spec:**
```bash
# Copy all golden outputs to safe location
mkdir -p /tmp/dingo-golden-reference
cp tests/golden/*.go.golden /tmp/dingo-golden-reference/

# Document current behavior
dingo build tests/golden/error_prop_01_simple.dingo
# Capture exact output for each test
```

**Delete old code:**
```bash
# Remove Participle parser
rm -rf pkg/parser/participle.go
rm -rf pkg/parser/ast.go

# Remove old plugin system (will be replaced by transformers)
rm -rf plugins/

# Keep only:
# - pkg/generator/ (might refactor but not delete)
# - tests/ (golden tests)
# - cmd/ (CLI - will update)
```

**Create clean package structure:**
```bash
mkdir -p pkg/preprocessor
mkdir -p pkg/transform
mkdir -p pkg/parser  # Reuse name, new implementation
```

**Deliverable:** Clean slate, old implementation gone, tests failing (expected)

### Phase 1: Core Infrastructure (Days 2-4)

**Goal:** Build foundational components

**1.1: Preprocessor Framework (Day 2)**

```go
// pkg/preprocessor/preprocessor.go
package preprocessor

type Preprocessor struct {
    source     []byte
    output     *strings.Builder
    sourceMap  *SourceMap
    tryCounter int
    matchCounter int
    lambdaCounter int
}

func New(source []byte) *Preprocessor {
    return &Preprocessor{
        source:    source,
        output:    &strings.Builder{},
        sourceMap: NewSourceMap(),
    }
}

func (p *Preprocessor) Process() (string, *SourceMap, error) {
    // Single pass through source
    // Apply all transformations
    // Track positions
    return p.output.String(), p.sourceMap, nil
}
```

**1.2: Source Maps (Day 2)**

```go
// pkg/preprocessor/sourcemap.go
type SourceMap struct {
    Mappings []Mapping
}

type Mapping struct {
    GeneratedLine   int
    GeneratedColumn int
    OriginalLine    int
    OriginalColumn  int
    Length          int
}

func (sm *SourceMap) AddMapping(m Mapping) {
    sm.Mappings = append(sm.Mappings, m)
}

func (sm *SourceMap) MapPosition(line, col int) (int, int) {
    // Binary search through mappings
    // Return original position
}
```

**1.3: go/parser Wrapper (Day 3)**

```go
// pkg/parser/parser.go
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

func (p *Parser) Parse(filename string, source []byte) (*ParseResult, error) {
    // Preprocess
    prep := preprocessor.New(source)
    goCode, sourceMap, err := prep.Process()
    if err != nil {
        return nil, err
    }

    // Parse with go/parser
    file, err := parser.ParseFile(p.fset, filename, goCode, parser.ParseComments)
    if err != nil {
        return nil, p.mapError(err, sourceMap)
    }

    return &ParseResult{
        AST:       file,
        SourceMap: sourceMap,
        FileSet:   p.fset,
    }, nil
}

type ParseResult struct {
    AST       *ast.File
    SourceMap *preprocessor.SourceMap
    FileSet   *token.FileSet
}
```

**1.4: Transform Framework (Day 4)**

```go
// pkg/transform/transform.go
package transform

import (
    "go/ast"
    "golang.org/x/tools/go/ast/astutil"
)

type Transformer struct {
    typeInfo *types.Info  // For type inference
}

func New(typeInfo *types.Info) *Transformer {
    return &Transformer{typeInfo: typeInfo}
}

func (t *Transformer) Transform(file *ast.File) error {
    astutil.Apply(file, t.visit, nil)
    return nil
}

func (t *Transformer) visit(cursor *astutil.Cursor) bool {
    // Check each node type
    // Transform placeholders
    return true
}
```

**Deliverable:** Core infrastructure compiles, has unit tests, but no feature implementations yet

### Phase 2: Error Propagation (Days 5-8)

**Goal:** First feature working end-to-end

**2.1: Preprocessor for `?` operator (Days 5-6)**

```go
// pkg/preprocessor/error_prop.go

func (p *Preprocessor) processErrorProp() error {
    // Scan for ? operator
    // Find preceding expression
    // Replace expr? with __dingo_try_N__(expr)
    // Track source positions
}
```

**Key challenge:** Parsing `?` in expressions
```go
// These must all work:
result := fetchData()?
x := len(data?) + 1
return processFile(path)?
if validate(input)? {
```

**Strategy:**
- Scan backwards from `?` to find expression start
- Use simple heuristics (matching parens, operators)
- Wrap expression in `__dingo_try_N__()`

**2.2: Transformer for error handling (Days 7-8)**

```go
// pkg/transform/error_prop.go

func (t *Transformer) transformErrorProp(cursor *astutil.Cursor) {
    call, ok := cursor.Node().(*ast.CallExpr)
    if !ok || !isDingoTry(call) {
        return
    }

    // Extract the expression
    expr := call.Args[0]

    // Determine context (assignment, return, etc.)
    context := t.analyzeContext(cursor)

    // Generate error handling code
    errorBlock := t.generateErrorHandling(expr, context)

    cursor.Replace(errorBlock)
}
```

**2.3: Test with 8 golden tests**

```bash
# Run each test
dingo build tests/golden/error_prop_01_simple.dingo

# Compare with saved reference
diff output.go /tmp/dingo-golden-reference/error_prop_01_simple.go.golden

# Iterate until all 8 pass
```

**Deliverable:** Error propagation working, 8 tests passing

### Phase 3: Lambdas + Type Inference (Days 9-12)

**Goal:** Validate go/types integration works

**3.1: Preprocessor for lambdas (Days 9-10)**

```go
// pkg/preprocessor/lambdas.go

func (p *Preprocessor) processLambdas() error {
    // Scan for |...| patterns
    // Extract parameters and body
    // Generate: __dingo_lambda_N__([]string{"x", "y"}, func() any { return body })
}
```

**Challenge:** Parsing lambda syntax
```go
|x| x * 2
|x, y| x + y
|| doSomething()
```

**Strategy:**
- Scan for `|`
- Collect param names until closing `|`
- Find expression (until comma, semicolon, or closing delimiter)

**3.2: Type inference with go/types (Days 11-12)**

```go
// pkg/transform/lambdas.go

func (t *Transformer) transformLambda(cursor *astutil.Cursor) {
    // Find __dingo_lambda_N__ calls

    // Use go/types to infer parameter and return types
    expectedType := t.typeInfo.TypeOf(cursor.Parent())

    // Rebuild function literal with proper types
    funcLit := &ast.FuncLit{
        Type: &ast.FuncType{
            Params: buildParamsWithTypes(paramNames, expectedType),
            Results: buildResultsWithTypes(expectedType),
        },
        Body: buildBody(originalBody),
    }

    cursor.Replace(funcLit)
}
```

**Key innovation:** Using go/types for inference (impossible with Participle!)

**Deliverable:** Lambdas working with type inference, 3 tests passing

### Phase 4: Sum Types (Days 13-16)

**Goal:** Complex transformation with type generation

**4.1: Preprocessor for enum (Days 13-14)**

```go
// pkg/preprocessor/sum_types.go

func (p *Preprocessor) processSumTypes() error {
    // Scan for enum declarations
    // Extract enum name and variants
    // Generate placeholder struct + metadata variable
}
```

**4.2: Transformer generates tagged unions (Days 15-16)**

```go
// pkg/transform/sum_types.go

func (t *Transformer) transformSumType(cursor *astutil.Cursor) {
    // Find enum metadata
    // Generate proper Go types:
    //   - Base struct with tag field
    //   - Constructor functions
    //   - Type assertion methods (IsVariant)
}
```

**Reuse existing plugin logic:**
- Current `plugins/sum_types.go` has 926 lines of working transformation
- Extract the AST generation logic (pure functions)
- Wrap in new transformer interface
- **Don't rewrite working code, just reorganize it**

**Deliverable:** Sum types working, 2 tests passing

### Phase 5: Pattern Matching (Days 17-20)

**Goal:** Most complex feature

**5.1: Preprocessor (Days 17-18)**

Encode match expressions as data structure parseable by go/parser

**5.2: Transformer (Days 19-20)**

Generate switch/type-switch with pattern destructuring

**Reuse existing plugin logic:**
- Current `plugins/pattern_match.go` has working implementation
- Extract and adapt to new transformer interface

**Deliverable:** Pattern matching working, 4 tests passing

### Phase 6: Operators (Days 21-23)

**Goal:** Simple operators (ternary, ??, ?.)

These are straightforward transformations, mostly in preprocessor.

**Deliverable:** All operator tests passing (6 tests)

### Phase 7: Integration & Polish (Days 24-28)

**7.1: All golden tests passing (Days 24-25)**

Run full test suite:
```bash
go test ./tests/golden/... -v
```

Ensure all 23 tests pass.

**7.2: CLI updates (Day 26)**

```go
// cmd/dingo/build.go

func buildFile(path string) error {
    // Read .dingo file
    source, err := os.ReadFile(path)

    // Parse with new architecture
    parser := parser.New()
    result, err := parser.Parse(path, source)

    // Transform
    transformer := transform.New(typeInfo)
    transformer.Transform(result.AST)

    // Generate
    generator := generator.New()
    generator.Generate(result.AST, outputPath)
}
```

**7.3: Documentation (Day 27)**

Update:
- README.md
- CLAUDE.md (Phase 3 architecture complete)
- CHANGELOG.md
- Comment all new packages

**7.4: Cleanup (Day 28)**

- Remove any TODO comments
- Remove debug logging
- Final code review
- Performance benchmarks

**Deliverable:** Production-ready new architecture

---

## Risk Analysis: Why This is Safe

### Risks

1. **Break all tests initially**
   - Mitigation: Expected, we have golden outputs as spec

2. **Take longer than estimated**
   - Mitigation: Pre-release, no deadline pressure

3. **Discover unforeseen complexity**
   - Mitigation: Can reference old code for algorithms, just not architecture

4. **Lose some working feature temporarily**
   - Mitigation: Git history preserves everything, can reference

### Why Risks are Acceptable

- **Pre-release**: No users affected
- **Git history**: Can always revert
- **Golden tests**: Clear success criteria
- **Better architecture**: Worth short-term pain
- **No technical debt**: Clean slate

---

## Success Criteria

1. ✅ All 23 golden tests pass
2. ✅ Generated Go code compiles
3. ✅ Output quality equal or better than before
4. ✅ Codebase is cleaner (fewer lines, clearer structure)
5. ✅ No Participle dependencies remain
6. ✅ Documentation updated

---

## Timeline Summary

**Phase 0:** Preparation (1 day)
**Phase 1:** Core infrastructure (3 days)
**Phase 2:** Error propagation (4 days)
**Phase 3:** Lambdas (4 days)
**Phase 4:** Sum types (4 days)
**Phase 5:** Pattern matching (4 days)
**Phase 6:** Operators (3 days)
**Phase 7:** Integration (5 days)

**Total: 28 days (4 weeks)**

**vs. Incremental migration: 42-49 days (6-7 weeks)**

**We save 2-3 weeks by going direct!**

---

## Decision: PROCEED WITH FULL MIGRATION

**Rationale:**
- Pre-release status makes this the PERFECT time
- Cleaner architecture worth the short-term disruption
- Faster than incremental (4 weeks vs 6-7 weeks)
- No technical debt or transition code
- Golden tests provide safety net

**Next step:** Begin Phase 0 (save golden outputs, delete old code)

# Parser Architecture Analysis for Dingo
*Analysis based on Go architecture expertise*

## Executive Summary

After analyzing various parser approaches for the Dingo transpiler, I recommend a **Hybrid Progressive Approach**: Continue with regex preprocessors for simple syntax transformations while progressively adding tree-sitter for complex, context-aware features.

## Detailed Analysis

### 1. Go Native Parser Extension

**Feasibility**: Limited
- `go/parser` is not designed for extension - it's a concrete implementation
- No public interfaces for custom syntax injection
- `go/scanner` similarly lacks extension points
- Would require forking Go's parser (maintenance nightmare)

**Pros**:
- Perfect Go syntax understanding
- Maintained by Go team

**Cons**:
- Not extensible without forking
- Would break with every Go release
- Complexity: 9/10

**Verdict**: Not viable for Dingo's needs

### 2. Third-Party Parsers

#### Tree-sitter-go
**Feasibility**: High
- Grammar-based, extensible
- Can fork and extend grammar files
- Active maintenance, Go 1.23+ compatible

**Pros**:
- Incremental parsing (great for LSP)
- Error recovery built-in
- Many successful language extensions use it

**Cons**:
- C library with Go bindings (added complexity)
- Learning curve for grammar syntax
- Complexity: 7/10

#### Participle
**Feasibility**: Medium
- Pure Go parser combinator
- Good for DSLs

**Pros**:
- Pure Go (no CGO)
- Declarative syntax via struct tags
- Good error messages

**Cons**:
- Not designed for extending existing languages
- Performance concerns for large files
- Less mature than tree-sitter
- Complexity: 6/10

### 3. Meta-Language Precedents

#### TypeScript
- Uses custom parser (not extending JavaScript parser)
- Superset approach: parses JS + TS extensions
- Lesson: Full control over parser is valuable

#### Borgo (Direct Precedent!)
- Uses tree-sitter with custom grammar
- Successfully transpiles to Go
- Lesson: Tree-sitter works for Go transpilers

#### CoffeeScript
- Custom lexer + parser (Jison)
- Lesson: Simple preprocessor-style approach can work

### 4. Hybrid Approach (RECOMMENDED)

**Strategy**: Progressive enhancement
1. Keep regex preprocessors for simple, isolated transformations
2. Add tree-sitter for complex, context-aware features
3. Gradual migration path

**Phase 1 (Current - Keep)**:
- Type annotations: `param: Type` → `param Type` (regex works fine)
- Error propagation: `x?` → error handling (regex sufficient)
- Simple keywords: `let`, `mut` (regex perfect)

**Phase 2 (Add tree-sitter for)**:
- Pattern matching (needs scope understanding)
- Complex enum transformations
- Lambda expressions with type inference
- Any feature requiring symbol resolution

**Implementation**:
```
.dingo → [Regex Preprocessor] → .dingo.preprocessed
       → [Tree-sitter Parser] → AST
       → [AST Transformer] → Go AST
       → [go/printer] → .go
```

### 5. Architecture Recommendation

**Recommended Approach: Hybrid Progressive**

**Rationale**:
1. **Immediate value**: Current regex approach works for 80% of features
2. **Future-proof**: Tree-sitter handles the complex 20%
3. **Incremental**: Can migrate features from regex to tree-sitter as needed
4. **Proven**: Borgo proves tree-sitter works for Go transpilation

**Implementation Plan**:
1. Continue with regex for current features (they work!)
2. Integrate tree-sitter-go for pattern matching (next major feature)
3. Gradually move complex transformations to tree-sitter
4. Keep simple text transforms in regex (KISS principle)

**Complexity**: 5/10 (incremental approach)
**Maintainability**: High (clear separation of concerns)
**Correctness**: Progressive (starts good, becomes excellent)
**Go Version Tracking**: Tree-sitter-go tracks Go releases

## Key Insights

1. **Surprising Finding**: Regex preprocessors are actually optimal for many Dingo features - they're simple, fast, and sufficient for syntax sugar transformations.

2. **Biggest Risk**: Trying to do everything with one approach. The hybrid model mitigates this by using the right tool for each job.

3. **Go Philosophy Alignment**: This approach follows Go's principle of simplicity - use the simplest tool that works correctly.

## Conclusion

Don't abandon regex preprocessors - they're working well! Instead, augment with tree-sitter for features that genuinely need AST awareness. This progressive enhancement approach delivers value quickly while building toward a robust long-term architecture.

The 12-15 month timeline is achievable with this approach:
- Months 1-6: Continue with regex for core features
- Months 7-9: Integrate tree-sitter for pattern matching
- Months 10-12: Polish, optimize, prepare for v1.0
- Months 13-15: Buffer for unknowns

This pragmatic approach balances immediate delivery with long-term maintainability.

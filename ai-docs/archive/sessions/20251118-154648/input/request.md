# Parser Architecture Investigation for Dingo

## Context

Dingo is a meta-language for Go (like TypeScript for JavaScript) that transpiles `.dingo` files to idiomatic `.go` files.

**Current Architecture (Two-Stage)**:
1. **Stage 1: Regex-based Preprocessor** - Transforms Dingo syntax to valid Go
   - TypeAnnotProcessor: `param: Type` → `param Type`
   - ErrorPropProcessor: `x?` → error handling code
   - EnumProcessor: `enum Name {}` → Go structs
   - KeywordProcessor: Other Dingo keywords

2. **Stage 2: AST Processing** - Parse and transform
   - Uses native `go/parser` to parse preprocessed Go code
   - Plugin pipeline transforms AST (Result types, etc.)
   - Generates `.go` + `.sourcemap` files

## The Problem

**Concern with current regex preprocessor approach**:
- Regex can't handle complex syntax and dependencies well
- Doesn't understand tokens/AST properly
- Could be fragile with nested structures, edge cases
- Hard to maintain as Dingo syntax grows

## Your Mission

Investigate ALL of the following approaches and provide recommendations:

### 1. Go Native Parser Extension
- Can we hook into or extend `go/parser`, `go/scanner`, `go/ast`?
- Are there interfaces/extension points in Go's parsing toolchain?
- What would be required to add custom syntax recognition?
- Has anyone done this successfully before?

### 2. Third-Party Go Parsers
- **tree-sitter-go**: Can we extend the grammar?
- **participle**: Parser combinator library - suitable for extending Go?
- **go-tree-sitter**: Bindings to tree-sitter
- **Other maintained parsers**: What else exists?
- Key criteria: Maintained, extensible, can handle Go 1.23+ syntax

### 3. Meta-Language Precedents
How do other successful meta-languages handle parsing?
- **TypeScript**: How does it extend JavaScript parsing?
- **Kotlin**: JVM bytecode, but parsing strategy?
- **Scala.js**: Scala → JavaScript, parsing approach?
- **Borgo**: Rust-like → Go transpiler (direct precedent!)
- **CoffeeScript**, **Elm**: Lessons learned?

### 4. Hybrid Approaches
- Which features need full parser (context-aware)?
- Which features can stay as preprocessor (simple text transforms)?
- Trade-offs: Complexity vs. correctness vs. maintainability
- Example split: Preprocessor for type annotations, parser for pattern matching?

### 5. Architecture Recommendations

Provide clear recommendations with:
- **Pros/Cons** for each approach
- **Complexity**: Implementation effort (1-10 scale)
- **Maintainability**: Long-term maintenance burden
- **Correctness**: How well it handles edge cases
- **Go version tracking**: Can we stay updated with Go releases?
- **Recommended approach** with justification

## Constraints

- Must generate idiomatic, readable Go code
- Must interoperate with all Go packages and tools
- Zero runtime overhead (transpile-time only)
- Future: Must support source maps for LSP (gopls proxy)
- Timeline: 12-15 months to v1.0

## Deliverables

Provide a detailed analysis covering all approaches above. Be thorough, investigate real projects, check GitHub discussions, and provide concrete recommendations.

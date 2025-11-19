# Parser Architecture Investigation for Dingo

## Executive Summary

**Top Recommendation:** Maintain Dingo's current two-stage regex preprocessor → `go/parser` architecture. It works well for Dingo's limited syntax additions and leverages Go's battle-tested parser.

**Key Insight:** Dingo's hybrid approach is actually correct for its scope - preprocess simple syntax changes, then use native `go/parser` for Go semantics.

**Biggest Risk:** Complexity growth - if Dingo adds sophisticated constructs (lambdas, generics), regex preprocessing becomes brittle and needs replacement with tree-sitter.

## Current Dingo Architecture

### Two-Stage Pipeline (Already Implemented)

```
.dingo file → [Regex Preprocessors] → Valid Go syntax → [go/parser + Plugins] → AST Transform → .go + .sourcemap
```

**Stage 1 - Regex Preprocessing:**
- TypeAnnotProcessor: `param: Type` → `param Type`
- ErrorPropProcessor: `expr?` → temporary variables + error checks  
- EnumProcessor: `enum Name {}` → Go structs + tag structs
- KeywordProcessor: Other Dingo keywords

**Stage 2 - AST Processing:**
- go/parser: Parse preprocessed Go into AST
- Plugin pipeline: Discovery (find Result/Ok calls) → Transform → Inject types
- golang.org/x/tools/go/ast/astutil: AST manipulations

## Go Native Parser Analysis

### Is Extending go/parser Feasible? ❌ **No**

**Key Findings:**
- Go's parser is not designed for extension
- `go.scanner` produces token streams, not extensible AST
- Rob Pike's approach requires full custom parser for new syntax
- **Effort:** 6-9 months to build complete custom parser

### Third-Party Go Parsers

| Parser | Maintainance | Extensibility | Go Integration | Recommendation |
|--------|-------------|---------------|----------------|----------------|
| **participle** | ⭐⭐⭐ Active | Low (fixed grammars) | ✅ Perfect | ✅ **Replace regex preprocess with this** |
| **tree-sitter** | ⭐⭐⭐ Active | High (C bindings) | ⚠️ Complex | ❌ Overkill for Dingo's simple syntax |
| **pigeon (PEG)** | ⭐⭐ Moderate | Medium | ✅ Good | ❌ Backtracking performance issues |
| **goyacc** | ⭐⭐ Maintaining | High (LALR) | ✅ Good | ❌ Complex grammar definitions |

**Recommendation: Switch to participle for preprocessors.** It would eliminate regex brittleness while maintaining single-binary simplicity.

## Meta-Language Precedents

### TypeScript Parsing Architecture

**TypeScript Strategy:** Does NOT extend JavaScript parser. Uses two-phase:
1. TypeScript parser produces TS-AST + type information
2. Code generators target different runtimes (JS, WebAssembly)

**Applies to Dingo:** Similar two-phase is correct - custom parser for Dingo features, code generation to Go.

### Borgo Parsing Approach

**Architecture:** Direct Rust parser implementation
- Lexer: Rust's `logos` crate
- Parser: Hand-written recursive descent
- Type Inference: Hindley-Milner in Rust
- Code Generation: Direct Go AST construction

**Key Differences from Dingo:**
- Borgo writes in Rust (full control over build)
- Borgo reimplements all compiler phases
- Borgo has no LSP tooling
- **Dingo benefits:** Go's parser ecosystem, existing tool support

### CoffeeScript/Elm/Scala.js Lessons Learned

**Common Mistakes:**
1. **Full custom parser too early** → Complex maintenance
2. **No source maps** → Poor developer experience
3. **Ignoring target ecosystem** → Poor interop

**Lessons for Dingo:** ✅ Correct approach with preprocessors + target parser. Include source maps from day one.

## Hybrid Approach Recommendations

### Which Features Need Full Parser Context?
- **Currently NONE** - Dingo can live with regex preprocessing
- **Future threats:** Complex expressions, generics, lambdas
- **Pattern matching:** Needs full parser for exhaustiveness checking

### Split Strategy (Recommended)

**Preprocessor (Code-Level Text Transforms):**
- `x: int` → `x int` (simple replacements)
- Basic keyword changes
- **Maintains:** Single-binary distribution

**Parser Extensions (Semantic-Level AST Transforms):**
- Complex logic (pattern matching exhaustiveness)
- **Requires:** Tree-sitter for context-aware parsing

## Architecture Recommendations

### Complexity Assessment (1-10 Scale)

| Approach | Implementation Effort | Maintainability | Correctness | Go Version Tracking | Total Score |
|----------|----------------------|---------------|-------------|-------------------|-------------|
| **Current (Regex + go/parser)** | 4 ✅ | 6 ✅ | 7 ✅ | 9 ✅ | **8/10** |
| Extend go/parser | 9 ❌ | 7 ❌ | 9 ✅ | 9 ✅ | **7/10** |
| Full tree-sitter | 6 ❌ | 5 ❌ | 8 ✅ | 3 ❌ | **2/10** |
| Switch to participle preprocessors | 3 ✅ | 8 ✅ | 8 ✅ | 9 ✅ | **9/10** |

### Recommended Evolution Path

**Phase 3-4: Keep current approach** (optimal for Result/Option/patterns)

**Phase 5+ Feature Complexity Check:**
- If adding: lambdas, generics, macro system → Migrate to tree-sitter  
- If staying: syntactic sugar only → Keep preprocessors

## Risk Assessment

### Current Risk Level: 🟢 **Medium**

**Mitigating Factors:**
- Dingo syntax additions are conservative
- go/parser provides battle-tested Go semantics  
- Plugin architecture allows incremental complexity

**Monitor Thresholds:**
- If regex preprocessors > 10 processors → Switch to participle
- If adding context-dependent features → Need tree-sitter
- If LSP performance becomes bottleneck → Need incremental parsing

## Conclusion

**Short-term:** Current architecture is optimal - regex for simple syntax, go/parser for Go semantics.

**Long-term:** Plan migration path to participle/tree-sitter as features grow.

**Key Decision:** Stay the course for Phase 3-4, but prototype participle replacement for a preprocessor to reduce technical debt.

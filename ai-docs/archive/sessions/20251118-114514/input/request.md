You are tasked with investigating parser architecture approaches for the Dingo transpiler project. Dingo is a Go meta-language that transpiles .dingo files to idiomatic Go code.

CURRENT ARCHITECTURE:
- Stage 1: Regex-based preprocessor (transforms Dingo syntax to valid Go)
- Concern: Regex can't handle complex syntax, fragile with nested structures

YOUR MISSION: Thoroughly investigate ALL approaches below and provide recommendations:

1. GO NATIVE PARSER EXTENSION:
   - Investigate go/parser, go/scanner, go/ast for extension points
   - Research Go toolchain interfaces for custom syntax
   - Has anyone successfully extended Go's parser?

2. THIRD-PARTY GO PARSERS:
   - tree-sitter-go: Can grammar be extended?
   - participle: Suitable for Go parser extension?
   - go-tree-sitter bindings
   - Other maintained parsers for Go 1.23+

3. META-LANGUAGE PRECEDENTS:
   - TypeScript: How it extends JS parsing
   - Borgo: Direct precedent (Rust-like → Go)
   - CoffeeScript, Elm: Lessons learned
   - Scala.js, Kotlin parsing strategies

4. HYBRID APPROACHES:
   - Split between simple text transforms vs context-aware parsing
   - Trade-offs: Complexity vs correctness vs maintainability

5. ARCHITECTURE RECOMMENDATIONS:
   - Pros/cons for each approach (complexity 1-10, maintainability, correctness)
   - Go version tracking ability
   - Clear recommended approach with justification

CONSTRAINTS:
- Must generate idiomatic Go code
- Zero runtime overhead
- Interoperate with all Go tools
- Support future source maps for LSP

DELIVERABLES:
- Comprehensive analysis of all approaches
- Concrete recommendations
- Investigation of real projects/GitHub discussions

Write detailed results to: ai-docs/sessions/20251118-114514/output/analysis.md

Return concise summary in this format:
# Parser Architecture Investigation Complete

Status: Success
Key Findings: [One-liner summary]
Recommended: [Approach name]
Complexity: [1-10 score]
Details: ai-docs/sessions/20251118-114514/output/analysis.md
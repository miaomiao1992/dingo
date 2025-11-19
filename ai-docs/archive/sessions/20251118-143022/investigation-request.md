# Hybrid Parser Architecture Investigation for Dingo

## Critical Decision Point

I'm architecting a Go transpiler called Dingo. Current architecture uses:
- Stage 1: Regex preprocessor for simple transforms (80% of cases)
- Stage 2: go/parser for AST processing

## The Question

Should we add an **optional syntax tree parser layer** between Stage 1 and 2 for context-aware features like pattern matching and complex lambdas?

## Analysis Needed

### 1. YES/NO Recommendation
Should we add this optional layer? Clear decision with rationale.

### 2. If YES, which parser technology?
A. Participle (Go parser combinator) - pure Go, composable
B. Custom recursive descent - full control, simple
C. PEG generator (pigeon) - declarative, generated code
D. Tree-sitter - incremental parsing, C dependency

### 3. Cost-Benefit Numbers
- Implementation effort (person-weeks)
- Performance overhead (%)
- Maintenance burden (hours/month)
- Features enabled

### 4. Precedents
How do TypeScript, Babel, Rust handle this? What can we learn?

### 5. Alternatives
If NO, what's better? Enhanced regex? Full parser replacement?

## Constraints
- Must be optional (zero overhead when unused)
- Must work with existing plugin pipeline
- Should enable pattern matching, advanced lambdas
- 3-4 month timeline acceptable if high value

Provide comprehensive technical analysis with concrete recommendations.
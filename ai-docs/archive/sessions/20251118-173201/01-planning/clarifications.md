# User Clarifications for Phase 4.2

## Question 1: Tuple Element Limit
**Decision:** 6 elements (balanced)

**Rationale:** Max 64 patterns. Handles complex cases while maintaining reasonable compile time. Good balance between flexibility and performance.

## Question 2: Guard Implementation Strategy
**Decision:** Nested if statements (safe)

**Rationale:** Simpler AST, safer, easier to debug. Better for nested matches. Slight performance trade-off is acceptable for improved maintainability.

## Question 3: Swift Guard Keyword
**Decision:** Support both 'if' and 'where' keywords

**Rationale:** Maximum flexibility - users can choose their preference. Worth the additional preprocessor complexity to support authentic Swift syntax.

## Implementation Implications

1. **Tuple exhaustiveness:** Use decision tree algorithm with 6-element limit (M^6 max patterns)
2. **Guard codegen:** Generate nested if/else chains instead of goto labels
3. **Swift preprocessor:** Must handle both `case .Ok(x) if x > 0` and `case .Ok(x) where x > 0` patterns

## Additional Decisions (from remaining gaps - use sensible defaults)

4. **Swift case bodies:** Allow both bare statements and braces (flexible, follows Swift)
5. **Error verbosity:** Always enhanced (consistent DX, no config complexity)
6. **Tuple wildcard semantics:** Catch-all (makes match exhaustive, following Rust)
7. **Guards and exhaustiveness:** Ignore guards entirely for exhaustiveness (simple, safe)
8. **Tuple type inference:** Use parent tracking from Phase 4.1 (reuse infrastructure)

These default decisions follow principle of least surprise and reuse existing Phase 4.1 infrastructure.

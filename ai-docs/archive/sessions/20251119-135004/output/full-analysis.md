# Dingo Enum Variant Naming Convention Analysis

## Model Consultations (Attempted)

### openrouter/sherlock-think-alpha
File: openrouter-sherlock-think-alpha.md
Content: Empty response beyond claudish header. Model did not produce JSON output.

### openai/gpt-4o
File: openai-gpt-4o.md
Content: Provided meta-instructions on how to perform multi-model consultation instead of direct analysis. Did not answer the 6 questions or produce JSON. Key excerpt:
> To evaluate the Dingo enum naming schemes A/B/C/D for Go idiomaticity, follow these steps manually...

### google/gemini-2.0-flash-exp
File: google-gemini-2.0-flash-exp.md
Content: API 404 error - model endpoint not found.

## Synthesis and Recommendation

External model consultations failed to produce the expected JSON responses due to model unavailability, misinterpretation of prompt, and errors. As golang-architect, I synthesize a recommendation based on Go idioms, current implementation, evaluation criteria, and real-world examples.

### Evaluation of Options

**A: Underscore CamelCase (Value_Int)**
- Current implementation.
- Pros: Clear separation of enum name and variant.
- Cons: Violates Go naming conventions (exported identifiers use CamelCase without underscores). Underscores are rare in exported names (e.g., snake_case for private fields only).
- Idiomaticity: Low.

**B: Pure CamelCase (ValueInt)**
- Pros: Fully idiomatic Go (e.g., http.StatusOK, errors.ErrNoProgress, json.Number).
- Cons: Less explicit separation if enum name ends with word that could collide (rare).
- Idiomaticity: High.

**C: Namespaced (Value.Int)**
- Pros: Extremely clear namespace.
- Cons: Invalid Go syntax for field access without instance. Would require Value{Int: ...} which is awkward.
- Idiomaticity: Low (not how Go accesses enum variants).

**D: All Lowercase (value_int)**
- Pros: None for exported.
- Cons: Unexported by convention, invisible to other packages.
- Idiomaticity: Very low.

### Recommended: B (Pure CamelCase - ValueInt)

**Rationale**:
- Aligns perfectly with Go's naming conventions for exported constants/types (CamelCase, no underscores).
- Matches patterns in stdlib (StatusOK, ErrUnsupported) and popular libs (sql.ErrNoRows, context.DeadlineExceeded).
- Simplest migration: Regex replace _ with '' in variant names.
- Maintains readability: ValueInt clearly evokes Value + Int.

**Answers to 6 Questions**:
1. **Go Idiomaticity**: B highest - stdlib uses pure CamelCase exclusively for exported enum-like constants.
2. **Current Dingo Users**: Minimal impact pre-v1.0; clear migration path.
3. **Go Tooling**: B works seamlessly with gopls, godoc.
4. **Other Go Libraries**: B matches wire.Codec, otel.SpanKind.
5. **Developer Expectations**: Go devs expect CamelCase exported names.
6. **Underscore Benefits**: Temporary clarity during prototyping, but long-term harms idiomaticity.

### Trade-offs
- **Pros of B**: Idiomatic, future-proof, no syntax issues.
- **Cons**: Slightly less explicit separation (mitigated by context in match arms).
- **Migration Cost**: Low - string replacement in golden tests and codegen.

### Real-World Alignment
- stdlib: http.StatusOK (not Status_OK)
- x/errors: ErrNoProgress
- grpc: codes.InvalidArgument → InvalidArgument

## Migration Plan
1. Update preprocessor/enum processor: Replace _ with '' in variant names.
2. Fix golden tests: pattern_match_*, option_*, result_*.
3. Verify: All tests pass, output readable.

**Final Decision**: Adopt B for production release.
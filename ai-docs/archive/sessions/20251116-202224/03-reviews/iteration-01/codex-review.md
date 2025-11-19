# Sum Types Implementation Code Review (GPT-5 Codex)

## CRITICAL Issues

### 1. Plugin output ordering causes duplicated declarations
**Location**: sum_types.go:84-87, 143-160, 208-216
- Tag const block ends up duplicated (once from generateTagEnum, once from generatedDecls append)
- Impact: Corrupts Go output and breaks compilation

### 2. Struct field injection unsafe for tuple variants
**Location**: sum_types.go:265-281, participle.go:660-668
- generateVariantFields assumes named Idents but tuple variants use synthesized names like `_0`
- Creates fields like `circle__0` never assigned by constructors
- Impact: Nil pointer dereferences

### 3. No error for duplicate variant names
- Parser accepts duplicate variant identifiers
- Impact: Confusing Go compile errors

### 4. No tests for enum transformation output
- No unit or golden tests verifying generated Go
- Impact: Cannot verify correctness

## IMPORTANT Issues

### 1. Match transformation ignores expression/statement context
**Location**: sum_types.go:444-469
- Blindly replaces with switchStmt regardless of expression context
- Switch statement cannot appear where expression required
- Need expression-friendly lowering (IIFE or chained ifs)

### 2. Enum registry collected but unused
**Location**: sum_types.go:92-102
- collectEnums stores in enumRegistry but never referenced
- transformMatchArm hardcodes "Tag_"+variant without enum name
- Dead state plus incorrect code generation

### 3. Placeholder nodes not removed from Dingo node map
**Location**: sum_types.go:162-165
- After cursor.Delete(), placeholder remains in currentFile.DingoNodes

### 4. Constructors alias ast.Field slices
**Location**: sum_types.go:295-299
- Reuses variant.Fields.List directly as parameter list
- Mutating params will mutate original enum AST

### 5. Match arm tag constants incorrect
**Location**: sum_types.go:483-487
- Generates Tag_<Variant> but constants are EnumTag_Variant
- References undefined identifiers

### 6. No nil guarding for pointer payloads
**Location**: sum_types.go:489-494
- Match lowering reads *matchedExpr.field without nil checks
- Runtime panics instead of compile errors

### 7. Missing errors for unsupported pattern forms
**Location**: sum_types.go:483-486
- Silently treats any non-wildcard as variant pattern
- Panics if pattern.Variant is nil

### 8. Guards ignored
**Location**: participle.go:166-169
- MatchArm.Guard parsed but transformMatchArm discards it
- Semantic miscompilation

### 9. No tests ensuring match expressions compile
- Need expr vs statement contexts, wildcard default tests

## MINOR Issues

### 1. Missing registry tests
- Need tests for multiple enums per file, Reset behavior

---
STATUS: CHANGES_NEEDED
CRITICAL_COUNT: 4
IMPORTANT_COUNT: 7
MINOR_COUNT: 1

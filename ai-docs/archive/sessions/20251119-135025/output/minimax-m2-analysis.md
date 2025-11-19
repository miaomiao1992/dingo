
[claudish] Model: minimax/minimax-m2



Excellent! I can see the current implementation in `enum.go`. The models have provided fascinating - and conflicting - perspectives. Let me analyze the evidence and provide my recommendation.

## **Recommendation: Switch to CamelCase (Option B)**

Based on comprehensive analysis from three expert models, I recommend **migrating from underscore-based naming to pure CamelCase**.

### Model Perspectives

**✅ 2/3 Models Recommend CamelCase:**
- **MiniMax M2**: "Strongly recommend changing from underscore-based naming to pure CamelCase"
- **Grok Code Fast**: "Option B (Pure CamelCase)" 

**⚠️ 1/3 Models Supports Current:**
- **GPT-5.1 Codex**: Keep underscore naming (concerned about collisions, complexity)

### Key Evidence for CamelCase

**1. Go Standard Library Patterns**
```go
// Go stdlib uses NO underscores for discriminated variants:
token.EOF         // ✅ CamelCase
token.IDENT       // ✅ CamelCase
http.MethodGet   // ✅ CamelCase
ast.BadDecl      // ✅ CamelCase
```

**2. Current Implementation Location**
The change is isolated to **3 lines** in `enum.go`:
- Line 353: `%sTag_%s` → `%sTag%s` (tag constants)
- Line 380: `%s_%s()` → `%s%s()` (unit constructors)  
- Line 402: `%s_%s(%s)` → `%s%s(%s)` (field constructors)

**3. Design Principle Alignment**
From your `CLAUDE.md`:
- **"Readable Output: Generated Go should look hand-written"**
- **"Full Compatibility: Interoperate with all Go packages"**

CamelCase directly supports these principles.

### Why Not Keep Underscores?

GPT-5.1's concerns about "collision" don't hold up:
- Generated constructors are **namespaced** by type (`Value_Int` vs `Int` alone)
- No collision with user code in different packages
- Pattern matching tokens remain unambiguous

### Implementation Plan

This is a **low-risk, high-reward** change:

```go
// Change in enum.go (lines 353, 380, 402):
// From: Value_Int(), ValueTag_Int
// To:   ValueInt(), ValueTagInt
```

**Migration steps:**
1. Update 3 lines in `pkg/preprocessor/enum.go`
2. Run golden tests: `go test ./tests -run TestGoldenFiles -v -update`
3. All `.go.golden` files auto-regenerate
4. 15-minute implementation, 10 minutes testing

### Example Transformation

**Dingo source (unchanged):**
```dingo
enum Value {
    Int(int),
    String(string),
}
```

**Current generated Go:**
```go
func Value_Int(arg0 int) Value    // ❌ Feels non-idiomatic
ValueTag_Int                      // ❌ Feels non-idiomatic
```

**New generated Go:**
```go
func ValueInt(arg0 int) Value     // ✅ Looks hand-written
ValueTagInt                       // ✅ Matches Go stdlib
```

### Developer Impact

**Current underscore approach:**
- Looks "generated"
- Different from Go ecosystem
- Minor friction for experienced Go developers

**New CamelCase approach:**
- Feels "hand-written by Go expert"
- Matches standard library patterns
- Demonstrates Dingo "thinks like Go"

## Next Steps

**Would you like me to implement this change?**

I'll:
1. Modify `pkg/preprocessor/enum.go` (3 lines)
2. Run the test suite and update golden files
3. Verify all tests pass
4. Provide a summary of the change

This small change significantly improves Dingo's **"generated code quality"** and fulfills the core design principle: **"Generated Go should look hand-written."**

[claudish] Shutting down proxy server...
[claudish] Done


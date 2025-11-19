# AST Pre-Scan Scope Clarification

## Question

Does AST pre-scan to detect local functions (prevents false transforms) look in **all package files** or **only in the local one**?

## Answer: Single File Scope (Current Plan)

The AST pre-scan operates on **single file scope only** - it only detects functions defined in the current `.dingo` file being transpiled.

## Rationale

### 1. **Correctness vs. Performance Tradeoff**

**Single File Scope:**
- ✅ Fast: ~10-50ms per file (already doing this parse later anyway)
- ✅ Simple: No cross-file dependency tracking
- ✅ Deterministic: Same input file → same output
- ⚠️ Limitation: Misses functions defined in other files in same package

**Package Scope:**
- ✅ More correct: Detects all package-level functions
- ❌ Slower: Must parse ALL package files (N × 10-50ms)
- ❌ Complex: Requires package discovery, multiple file handling
- ❌ Context-dependent: Output depends on entire package structure

**Module Scope:**
- ❌ Too expensive: Must parse entire module
- ❌ Overkill: Functions from other packages are qualified anyway

### 2. **Dingo's File-Oriented Transpilation Model**

Dingo transpiles **file-by-file**, not package-by-package:

```bash
dingo build file1.dingo  # Produces file1.go
dingo build file2.dingo  # Produces file2.go
```

**Implications:**
- Each transpilation is independent
- No global package context maintained
- Fast incremental compilation
- Watch mode efficiency

**Package scope would require:**
- Parsing all package files on every transpilation
- Caching package-level metadata
- Invalidation logic when files change
- Significant architectural change

### 3. **Real-World Impact is Minimal**

**Scenario where single-file scope matters:**

```go
// file1.dingo
package myapp

func ReadFile(path string) []byte {  // User-defined
    // Custom logic
}
```

```go
// file2.dingo
package myapp

func main() {
    data := ReadFile("config.txt")  // Would transform to os.ReadFile ❌
}
```

**Why this is acceptable:**

1. **Rare in practice**: Most helper functions live in the same file as usage
2. **Easy fix**: User adds `import "os"` and uses `os.ReadFile` explicitly
3. **Compile-time catch**: Go compiler will catch the error immediately
4. **Clear error**: "undefined: ReadFile" → User knows to import or use qualified name

### 4. **Error is Safe (Not Silent)**

**What happens with single-file scope:**

```go
// file2.dingo
package myapp

func main() {
    data := ReadFile("config.txt")  // Transforms to os.ReadFile
}
```

**Transpiled Go:**
```go
package myapp

import "os"

func main() {
    data := os.ReadFile("config.txt")  // Go compiler error! ❌
}
```

**Go compiler error:**
```
file2.go:6:12: cannot use os.ReadFile (value of type func(string) ([]byte, error))
            as type []byte in assignment
```

**User reaction:**
- Sees compile error
- Realizes they need to import their own `ReadFile`
- Adds `import . "myapp"` or uses qualified call
- **Or:** Realizes they meant `os.ReadFile` and fixes return type

**Key point:** The mistake is **caught immediately** by Go compiler, not at runtime. User gets clear feedback.

### 5. **Philosophical Alignment**

Dingo's design philosophy:
- **Explicit over implicit** - Errors are better than silent mistakes
- **Conservative transforms** - When in doubt, error out
- **Go compatibility** - Rely on Go compiler for final validation

**Single-file scope fits this:**
- Fast, simple, predictable
- Errors caught at compile-time (Go does the work)
- No magic "action at a distance" (file A affecting file B's transpilation)

### 6. **Best Practice Workaround**

If users have package-wide helper functions they want to use unqualified:

**Option 1: Same-file helpers**
```go
// myapp.dingo
package myapp

func ReadFile(path string) []byte { ... }  // Helper

func main() {
    data := ReadFile("config.txt")  // ✅ Detected
}
```

**Option 2: Explicit import**
```go
// file2.dingo
package myapp

import . "myapp"  // Import own package

func main() {
    data := ReadFile("config.txt")  // ✅ User's explicit choice
}
```

**Option 3: Qualified stdlib calls**
```go
// file2.dingo
package myapp

func main() {
    data := os.ReadFile("config.txt")  // ✅ Explicit, no ambiguity
}
```

### 7. **Future Enhancement Path**

If single-file scope proves problematic in practice, we can:

**Phase 2 Enhancement:**
- Add **optional** package-wide scanning (opt-in flag)
- `dingo build --scan-package file.dingo`
- Cache package metadata in `.dingo/cache/`
- Still default to single-file for performance

**Metrics to track:**
- How often do users hit this edge case?
- How many false transforms occur?
- Do users prefer speed or perfect detection?

## Decision: Single File Scope

**Implementation in `pkg/preprocessor/local_scanner.go`:**

```go
// LocalScanner extracts user-defined function names from source
type LocalScanner struct {
    fset       *token.FileSet
    localFuncs map[string]bool // "ReadFile" → true if user-defined
}

// ScanFile parses SINGLE file and extracts local function definitions
func (s *LocalScanner) ScanFile(source string) error {
    // Parse THIS file only
    fset := token.NewFileSet()
    file, err := parser.ParseFile(fset, "", source, 0)
    if err != nil {
        return err
    }

    // Walk AST and collect function declarations
    ast.Inspect(file, func(n ast.Node) bool {
        switch decl := n.(type) {
        case *ast.FuncDecl:
            s.localFuncs[decl.Name.Name] = true  // Single file scope
        }
        return true
    })

    return nil
}
```

**Input to preprocessor:**
```go
func transpileFile(path string) error {
    source, _ := os.ReadFile(path)

    // Pre-scan THIS file only
    scanner := preprocessor.NewLocalScanner()
    scanner.ScanFile(string(source))  // Single file

    input := preprocessor.Input{
        Source:         string(source),
        LocalFunctions: scanner,  // Only functions from this file
    }

    result, err := preprocessor.Process(input)
    // ...
}
```

## Examples

### Example 1: Single File (Works Perfectly)

**Input:**
```go
// myapp.dingo
package myapp

func ReadFile(path string) []byte {  // User-defined
    return nil
}

func main() {
    data := ReadFile("config.txt")  // ✅ Detected as local
}
```

**Output:**
```go
// myapp.go
package myapp

// No import "os" added ✅

func ReadFile(path string) []byte {
    return nil
}

func main() {
    data := ReadFile("config.txt")  // Unchanged ✅
}
```

### Example 2: Cross-File (False Transform, Caught by Compiler)

**Input:**
```go
// helpers.dingo
package myapp

func ReadFile(path string) []byte {  // User-defined in other file
    return nil
}
```

```go
// main.dingo
package myapp

func main() {
    data := ReadFile("config.txt")  // Not detected (different file)
}
```

**Output:**
```go
// main.go
package myapp

import "os"  // ⚠️ False import added

func main() {
    data := os.ReadFile("config.txt")  // ⚠️ Wrong function
}
```

**Go compiler error:**
```
main.go:6:12: cannot use os.ReadFile (value of type func(string) ([]byte, error))
            as type []byte in assignment
```

**User fixes:**
```go
// main.dingo
package myapp

import . "myapp"  // Import own package

func main() {
    data := ReadFile("config.txt")  // ✅ Now uses helpers.ReadFile
}
```

### Example 3: Stdlib Function (Works Correctly)

**Input:**
```go
// main.dingo
package myapp

func main() {
    data, err := ReadFile("config.txt")  // Stdlib function
}
```

**Output:**
```go
// main.go
package myapp

import "os"

func main() {
    data, err := os.ReadFile("config.txt")  // ✅ Correct transform
}
```

## Summary

**Scope:** Single file only

**Pros:**
- Fast (10-50ms overhead)
- Simple implementation
- Deterministic behavior
- Fits file-oriented transpilation model

**Cons:**
- Misses functions from other files in same package
- Can cause false transforms (rare)

**Safety:**
- False transforms caught by Go compiler (not runtime)
- User gets clear error message
- Easy workaround (explicit import or qualified call)

**Future:**
- Can add package-wide scope as opt-in enhancement
- Track real-world usage to validate decision
- Default remains single-file for performance

**Alignment:**
- Explicit over implicit ✅
- Fast compilation ✅
- Go compiler as final validator ✅

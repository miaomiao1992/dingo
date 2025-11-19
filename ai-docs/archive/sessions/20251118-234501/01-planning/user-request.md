# Bug Fix: Unqualified Import Inference

## Current Behavior

### With Qualified Calls (os.ReadFile):
- ✅ Adds import "os"
- ✅ Keeps os.ReadFile(path)

### With Unqualified Calls (ReadFile):
- ❌ NO import added
- ❌ Keeps ReadFile(path) → undefined function!

## The Design Question

Should Dingo support unqualified import inference?

Looking at all the test files - they ALL use unqualified function names:
- ReadFile(path)
- Atoi(s)

This suggests the original design intended for Dingo to infer:
- ReadFile → os.ReadFile (adds import "os")
- Atoi → strconv.Atoi (adds import "strconv")

But the current implementation only works with qualified names.

## The Bug

The current system has a design gap:
- Preprocessor detects os.ReadFile and adds os import ✅
- Preprocessor detects ReadFile and does nothing ❌

## Reference Files

Test demonstrating the issue:
- `/Users/jack/mag/dingo/tests/golden/error_prop_01_simple.dingo` - Uses `ReadFile(path)`
- `/Users/jack/mag/dingo/tests/golden/error_prop_01_simple.go` - Generated code has undefined `ReadFile`

## Expected Behavior

When Dingo encounters `ReadFile(path)` it should:
1. Recognize it as `os.ReadFile`
2. Add import "os" to the generated Go file
3. Transform to `os.ReadFile(path)` in the generated code

Same for other common standard library functions like:
- `Atoi` → `strconv.Atoi`
- `Printf` → `fmt.Printf`
- etc.

# Dingo LSP Manual Testing Guide

**Version**: Post-AST Source Maps (2025-11-22)
**Test File**: `tests/golden/error_prop_01_simple.dingo`

## Prerequisites

### 1. Rebuild and Restart

```bash
# 1. Rebuild dingo binary (if not done already)
cd /Users/jack/mag/dingo
go build -o dingo ./cmd/dingo

# 2. Rebuild dingo-lsp binary
go build -o dingo-lsp ./cmd/dingo-lsp

# 3. Regenerate test file source map
./dingo build tests/golden/error_prop_01_simple.dingo
```

### 2. Restart LSP Server

**VSCode**:
- Method 1: Cmd+Shift+P → "Developer: Reload Window"
- Method 2: Cmd+Shift+P → "Dingo: Restart Language Server" (if available)
- Method 3: Close VSCode, kill dingo-lsp process, reopen VSCode

**Terminal** (if running manually):
```bash
# Kill existing dingo-lsp
pkill dingo-lsp

# Restart (your editor should auto-restart it)
```

### 3. Open Test File

Open: `/Users/jack/mag/dingo/tests/golden/error_prop_01_simple.dingo`

```dingo
package main

func readConfig(path string) ([]byte, error) {
	let data = os.ReadFile(path)?
	return data, nil
}


func test() {
	let a = readConfig("config.yaml")?
	println(string(a))
}
```

---

## Test Suite

### ✅ Test 1: Hover on `ReadFile`

**Location**: Line 4, columns 16-24 (`ReadFile` in `os.ReadFile(path)?`)

**Steps**:
1. Position cursor on `ReadFile`
2. Wait for hover popup (or trigger with mouse hover)

**Expected Result** ✅:
```
func ReadFile(name string) ([]byte, error)

ReadFile reads the named file and returns the contents.
A successful call returns err == nil, not err == EOF...
```

**Wrong Results** ❌:
- Shows "package os" (means position translation is off)
- Shows nothing (means LSP failed to translate position)
- Shows blank popup

**Status**: Should work (tested earlier)

---

### ✅ Test 2: Hover on `os`

**Location**: Line 4, columns 12-14 (`os` in `os.ReadFile(path)?`)

**Steps**:
1. Position cursor on `os`
2. Wait for hover popup

**Expected Result** ✅:
```
package os

Package os provides a platform-independent interface to operating system functionality...
```

**Wrong Results** ❌:
- Shows ReadFile function (wrong symbol)
- Shows nothing

**Status**: Should work (tested earlier)

---

### ✅ Test 3: Hover on `path`

**Location**: Line 4, columns 25-29 (`path` in `os.ReadFile(path)?`)

**Steps**:
1. Position cursor on `path`
2. Wait for hover popup

**Expected Result** ✅:
```
path string

(parameter) path string
```

**Wrong Results** ❌:
- Shows nothing
- Shows wrong type

**Status**: Should work (tested earlier)

---

### ✅ Test 4: Hover on `readConfig` (call site)

**Location**: Line 10, columns 10-20 (`readConfig` in `readConfig("config.yaml")?`)

**Steps**:
1. Position cursor on `readConfig`
2. Wait for hover popup

**Expected Result** ✅:
```
func readConfig(path string) ([]byte, error)

(defined at line 3)
```

**Wrong Results** ❌:
- Shows nothing
- Shows wrong signature

**Status**: Should work (tested earlier)

---

### 🆕 Test 5: Go to Definition - `readConfig` (CRITICAL TEST)

**Location**: Line 10, columns 10-20 (`readConfig` in function call)

**Steps**:
1. Position cursor on `readConfig`
2. **Cmd+Click** (Mac) or **Ctrl+Click** (Linux/Windows)
3. OR: Right-click → "Go to Definition"
4. OR: F12 key

**Expected Result** ✅:
- **Jumps to Line 3**: `func readConfig(path string) ([]byte, error) {`
- Cursor positioned at start of function definition

**Wrong Results** ❌:
- Jumps to Line 7 (blank line) ← Previous bug
- Jumps to Line 4 (inside function body)
- Nothing happens
- Jumps to wrong file

**How to Verify**:
- After jump, you should see: `func readConfig(path string) ([]byte, error) {`
- Line number indicator should show **3**
- NOT line 7 (blank line)

**This is the main test for the duplicate mapping fix!**

---

### 🆕 Test 6: Go to Definition - `ReadFile`

**Location**: Line 4, columns 16-24 (`ReadFile` in `os.ReadFile(path)?`)

**Steps**:
1. Position cursor on `ReadFile`
2. Cmd+Click or F12

**Expected Result** ✅:
- Jumps to Go stdlib file (e.g., `/usr/local/go/src/os/file.go`)
- Shows ReadFile function definition in stdlib

**Wrong Results** ❌:
- Nothing happens
- Jumps to wrong location
- Shows error message

---

### 🆕 Test 7: Go to Definition - `os` package

**Location**: Line 4, columns 12-14 (`os` in `os.ReadFile(path)?`)

**Steps**:
1. Position cursor on `os`
2. Cmd+Click or F12

**Expected Result** ✅:
- Jumps to os package documentation or import location
- May jump to import statement: `import "os"` in generated .go file

**Wrong Results** ❌:
- Nothing happens
- Error message

---

### 🆕 Test 8: Find References - `readConfig`

**Location**: Line 3 or 10 (either definition or call site)

**Steps**:
1. Position cursor on `readConfig`
2. Right-click → "Find All References"
3. OR: Shift+F12

**Expected Result** ✅:
Shows 2 references:
```
error_prop_01_simple.dingo
  Line 3: func readConfig(path string) ([]byte, error) {
  Line 10: let a = readConfig("config.yaml")?
```

**Wrong Results** ❌:
- Shows only 1 reference
- Shows references with wrong line numbers
- Shows nothing

---

### 🆕 Test 9: Document Symbols (Outline)

**Steps**:
1. Open outline view in VSCode (Cmd+Shift+O or sidebar)
2. OR: View → Outline

**Expected Result** ✅:
Shows 2 functions:
```
📦 error_prop_01_simple
  ƒ readConfig
  ƒ test
```

**Wrong Results** ❌:
- Shows nothing
- Shows wrong function names
- Shows functions from .go file instead of .dingo file

---

### 🆕 Test 10: Completion (Autocomplete)

**Location**: Line 4, after typing `os.`

**Steps**:
1. Position cursor after `os.` (before `ReadFile`)
2. Trigger autocomplete (Ctrl+Space)

**Expected Result** ✅:
Shows os package members:
```
ReadFile
WriteFile
Open
Create
Remove
...
```

**Wrong Results** ❌:
- Shows nothing
- Shows wrong suggestions
- Shows members from wrong package

---

### 🆕 Test 11: Error Diagnostics

**Steps**:
1. Introduce an error (e.g., change `path` to `wrongName`)
2. Save file
3. Wait for error squiggles

**Expected Result** ✅:
- Red squiggly under `wrongName`
- Error message: "undefined: wrongName"
- Error appears at correct position in .dingo file

**Wrong Results** ❌:
- No error shown
- Error appears at wrong position
- Error refers to .go file line numbers

**Cleanup**: Revert the change after testing

---

### 🆕 Test 12: Rename Symbol

**Location**: Line 3 (function definition `readConfig`)

**Steps**:
1. Position cursor on `readConfig`
2. Right-click → "Rename Symbol"
3. OR: F2 key
4. Type new name: `loadConfig`
5. Press Enter

**Expected Result** ✅:
Both occurrences renamed:
```
Line 3: func loadConfig(path string) ([]byte, error) {
Line 10: let a = loadConfig("config.yaml")?
```

**Wrong Results** ❌:
- Only one occurrence renamed
- Renames in wrong file
- Error message

**Cleanup**: Rename back to `readConfig` after testing

---

## Troubleshooting

### LSP Not Responding

**Symptoms**:
- No hover
- No go-to-definition
- No autocomplete

**Fix**:
1. Check LSP server is running: `ps aux | grep dingo-lsp`
2. Check LSP logs (if available)
3. Restart LSP server
4. Rebuild dingo-lsp binary

### Wrong Positions

**Symptoms**:
- Hover works but go-to-definition jumps to wrong line
- Diagnostics appear at wrong position

**Fix**:
1. Verify source map exists: `ls tests/golden/error_prop_01_simple.go.map`
2. Check source map has no duplicates:
   ```bash
   jq '.mappings | group_by(.generated_line) | map(select(length > 1))' \
     tests/golden/error_prop_01_simple.go.map
   ```
   Should return `[]`
3. Regenerate source map: `./dingo build tests/golden/error_prop_01_simple.dingo`
4. Restart LSP

### LSP Logs

**Enable debug logging**:
```bash
export DINGO_LSP_LOG=debug
./dingo-lsp
```

**Check logs** (in terminal where LSP is running or VSCode output panel)

---

## Success Checklist

After completing all tests, verify:

- ✅ Test 1-4: Hover works on all symbols
- ✅ **Test 5**: Go to definition jumps to **line 3** (not line 7!) ← CRITICAL
- ✅ Test 6-7: Go to definition works for stdlib symbols
- ✅ Test 8: Find references shows all occurrences
- ✅ Test 9: Outline shows correct functions
- ✅ Test 10: Autocomplete works
- ✅ Test 11: Diagnostics appear at correct positions
- ✅ Test 12: Rename symbol works for all occurrences

**If Test 5 fails** (jumps to line 7 instead of line 3):
- Duplicate mapping bug is not fixed
- Source map needs regeneration
- LSP needs restart

---

## Reporting Results

After testing, report:

1. **Which tests passed** ✅
2. **Which tests failed** ❌
3. **For failed tests**:
   - What happened (e.g., "jumped to line 7 instead of line 3")
   - Expected vs actual behavior
   - Any error messages

Example report:
```
✅ Test 1-4: Hover - All working
✅ Test 5: Go to Definition (readConfig) - Jumps to line 3 correctly!
✅ Test 6-7: Go to Definition (stdlib) - Working
❌ Test 8: Find References - Shows only 1 reference instead of 2
✅ Test 9: Outline - Shows both functions
...
```

---

**Last Updated**: 2025-11-22
**Post-AST Source Maps**: Enabled
**Test File**: error_prop_01_simple.dingo

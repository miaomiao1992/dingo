package preprocessor

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"go/types"
	"regexp"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

// Package-level compiled regexes (Issue 2: Regex Performance)
var (
	assignPattern = regexp.MustCompile(`^\s*(let|var)\s+(\w+)\s*=\s*(.+)$`)
	returnPattern = regexp.MustCompile(`^\s*return\s+(.+)$`)
	msgPattern    = regexp.MustCompile(`^(.*\?)\s*"((?:[^"\\]|\\.)*)"`)
)

// Magic Comment System Documentation
//
// The error propagation processor inserts special marker comments to enable
// accurate source mapping between Dingo source and generated Go code.
//
// Format:
//   // dingo:s:1  - Marks the start of an expanded block (1 original line)
//   // dingo:e:1  - Marks the end of an expanded block (1 original line)
//
// Purpose:
//   When a single line of Dingo code (e.g., "let x = ReadFile(path)?") expands to
//   7 lines of Go code, these markers help the LSP server map error positions back
//   to the original Dingo source line.
//
// Example Expansion:
//   Dingo:  let x = ReadFile(path)?
//   Go:     __tmp0, __err0 := ReadFile(path)
//           // dingo:s:1
//           if __err0 != nil {
//               return nil, __err0
//           }
//           // dingo:e:1
//           var x = __tmp0
//
// The number after 's' and 'e' indicates how many original lines were consumed.
// Currently always 1 since error propagation only processes single-line expressions.
//
// Future Enhancement:
//   These markers will be consumed by the LSP server to provide accurate:
//   - Error message positioning
//   - Breakpoint mapping for debugging
//   - Go-to-definition navigation
//   - Hover information

// ErrorPropProcessor handles the ? operator for error propagation
// Transforms: expr? → full error handling expansion
type ErrorPropProcessor struct {
	tryCounter  int
	lines       []string
	currentFunc *funcContext
	needsFmt    bool
}

// funcContext tracks the current function for zero value generation
type funcContext struct {
	returnTypes []string
	zeroValues  []string
}

// NewErrorPropProcessor creates a new error propagation preprocessor
func NewErrorPropProcessor() *ErrorPropProcessor {
	return &ErrorPropProcessor{
		tryCounter: 0,
	}
}

// Name returns the processor name
func (e *ErrorPropProcessor) Name() string {
	return "error_propagation"
}

// Process transforms error propagation operators
func (e *ErrorPropProcessor) Process(source []byte) ([]byte, []Mapping, error) {
	// Split into lines for processing
	e.lines = strings.Split(string(source), "\n")
	e.needsFmt = false

	var output bytes.Buffer
	mappings := []Mapping{}
	inputLineNum := 0
	outputLineNum := 1 // Track current output line number (1-based)

	for inputLineNum < len(e.lines) {
		line := e.lines[inputLineNum]

		// Check if this is a function declaration
		if e.isFunctionDeclaration(line) {
			e.currentFunc = e.parseFunctionSignature(inputLineNum)
			e.tryCounter = 0 // Reset counter for each function
		}

		// Process the line, passing the current output line number
		transformed, newMappings := e.processLine(line, inputLineNum+1, outputLineNum)
		output.WriteString(transformed)
		if inputLineNum < len(e.lines)-1 {
			output.WriteByte('\n')
		}

		// Add all mappings from this line
		if len(newMappings) > 0 {
			mappings = append(mappings, newMappings...)
		}

		// Update output line count
		// The transformed text may contain multiple lines.
		// Count: number of newlines + 1 (for the line itself)
		// This gives us the number of lines the transformation occupies.
		newlineCount := strings.Count(transformed, "\n")
		linesOccupied := newlineCount + 1
		outputLineNum += linesOccupied

		inputLineNum++
	}

	// If we used fmt.Errorf, add import at the top
	result := output.Bytes()
	if e.needsFmt {
		result = e.ensureFmtImport(result)
	}

	return result, mappings, nil
}

// processLine processes a single line
// Returns: (transformed_text, mappings)
func (e *ErrorPropProcessor) processLine(line string, originalLineNum int, outputLineNum int) (string, []Mapping) {
	// Check if line contains ? operator (and not ternary)
	if !strings.Contains(line, "?") {
		return line, nil
	}

	// Check if it's a ternary (has : after ?)
	if e.isTernaryLine(line) {
		return line, nil
	}

	// Pattern: let/var NAME = EXPR? ["message"]
	if matches := assignPattern.FindStringSubmatch(line); matches != nil {
		rightSide := matches[3] // Everything after =
		if strings.Contains(rightSide, "?") {
			expr, errMsg := e.extractExpressionAndMessage(rightSide)
			result, mappings := e.expandAssignment(matches, expr, errMsg, originalLineNum, outputLineNum)
			return result, mappings
		}
	}

	// Pattern: return EXPR? ["message"]
	if matches := returnPattern.FindStringSubmatch(line); matches != nil {
		returnPart := matches[1] // Everything after return
		if strings.Contains(returnPart, "?") {
			expr, errMsg := e.extractExpressionAndMessage(returnPart)
			result, mappings := e.expandReturn(matches, expr, errMsg, originalLineNum, outputLineNum)
			return result, mappings
		}
	}

	// If we can't recognize the pattern, leave as-is
	return line, nil
}

// extractExpressionAndMessage extracts the expression and optional error message
// Input: "ReadFile(path)? \"failed to read\"" → ("ReadFile(path)?", "failed to read")
// Input: "ReadFile(path)?" → ("ReadFile(path)?", "")
func (e *ErrorPropProcessor) extractExpressionAndMessage(line string) (string, string) {
	// Look for ? followed by optional string (handle escaped quotes)
	// This pattern matches: expr? "message with \" escapes"
	if matches := msgPattern.FindStringSubmatch(strings.TrimSpace(line)); matches != nil {
		return matches[1], matches[2]
	}

	// No message, return as-is
	return strings.TrimSpace(line), ""
}

// expandAssignment expands: let x = expr? → full error handling
// Creates mappings for all 7 generated lines back to the original source line
func (e *ErrorPropProcessor) expandAssignment(matches []string, expr string, errMsg string, originalLine int, startOutputLine int) (string, []Mapping) {
	keyword := matches[1]  // "let" or "var"
	varName := matches[2]  // variable name
	exprClean := strings.TrimSpace(strings.TrimSuffix(expr, "?"))

	tmpVar := fmt.Sprintf("__tmp%d", e.tryCounter)
	errVar := fmt.Sprintf("__err%d", e.tryCounter)
	e.tryCounter++

	// Generate the expansion
	var buf bytes.Buffer
	indent := e.getIndent(matches[0])
	mappings := []Mapping{}

	// Line 1: __tmpN, __errN := expr
	buf.WriteString(indent)
	buf.WriteString(fmt.Sprintf("%s, %s := %s\n", tmpVar, errVar, exprClean))
	mappings = append(mappings, Mapping{
		OriginalLine:    originalLine,
		OriginalColumn:  1,
		GeneratedLine:   startOutputLine,
		GeneratedColumn: 1,
		Length:          len(matches[0]),
		Name:            "error_prop",
	})

	// Line 2: // dingo:s:1
	buf.WriteString(indent)
	buf.WriteString("// dingo:s:1\n")
	mappings = append(mappings, Mapping{
		OriginalLine:    originalLine,
		OriginalColumn:  1,
		GeneratedLine:   startOutputLine + 1,
		GeneratedColumn: 1,
		Length:          len(matches[0]),
		Name:            "error_prop",
	})

	// Line 3: if __errN != nil {
	buf.WriteString(indent)
	buf.WriteString(fmt.Sprintf("if %s != nil {\n", errVar))
	mappings = append(mappings, Mapping{
		OriginalLine:    originalLine,
		OriginalColumn:  1,
		GeneratedLine:   startOutputLine + 2,
		GeneratedColumn: 1,
		Length:          len(matches[0]),
		Name:            "error_prop",
	})

	// Line 4: return zeroValues, wrapped_error
	buf.WriteString(indent)
	buf.WriteString("\t")
	buf.WriteString(e.generateReturnStatement(errVar, errMsg))
	buf.WriteString("\n")
	mappings = append(mappings, Mapping{
		OriginalLine:    originalLine,
		OriginalColumn:  1,
		GeneratedLine:   startOutputLine + 3,
		GeneratedColumn: 1,
		Length:          len(matches[0]),
		Name:            "error_prop",
	})

	// Line 5: }
	buf.WriteString(indent)
	buf.WriteString("}\n")
	mappings = append(mappings, Mapping{
		OriginalLine:    originalLine,
		OriginalColumn:  1,
		GeneratedLine:   startOutputLine + 4,
		GeneratedColumn: 1,
		Length:          len(matches[0]),
		Name:            "error_prop",
	})

	// Line 6: // dingo:e:1
	buf.WriteString(indent)
	buf.WriteString("// dingo:e:1\n")
	mappings = append(mappings, Mapping{
		OriginalLine:    originalLine,
		OriginalColumn:  1,
		GeneratedLine:   startOutputLine + 5,
		GeneratedColumn: 1,
		Length:          len(matches[0]),
		Name:            "error_prop",
	})

	// Line 7: var varName = __tmpN
	buf.WriteString(indent)
	buf.WriteString(fmt.Sprintf("%s %s = %s", keyword, varName, tmpVar))
	mappings = append(mappings, Mapping{
		OriginalLine:    originalLine,
		OriginalColumn:  1,
		GeneratedLine:   startOutputLine + 6,
		GeneratedColumn: 1,
		Length:          len(matches[0]),
		Name:            "error_prop",
	})

	return buf.String(), mappings
}

// expandReturn expands: return expr? → full error handling
// Creates mappings for all 7 generated lines back to the original source line
func (e *ErrorPropProcessor) expandReturn(matches []string, expr string, errMsg string, originalLine int, startOutputLine int) (string, []Mapping) {
	exprClean := strings.TrimSpace(strings.TrimSuffix(expr, "?"))

	tmpVar := fmt.Sprintf("__tmp%d", e.tryCounter)
	errVar := fmt.Sprintf("__err%d", e.tryCounter)
	e.tryCounter++

	// Generate the expansion
	var buf bytes.Buffer
	indent := e.getIndent(matches[0])
	mappings := []Mapping{}

	// Line 1: __tmpN, __errN := expr
	buf.WriteString(indent)
	buf.WriteString(fmt.Sprintf("%s, %s := %s\n", tmpVar, errVar, exprClean))
	mappings = append(mappings, Mapping{
		OriginalLine:    originalLine,
		OriginalColumn:  1,
		GeneratedLine:   startOutputLine,
		GeneratedColumn: 1,
		Length:          len(matches[0]),
		Name:            "error_prop",
	})

	// Line 2: // dingo:s:1
	buf.WriteString(indent)
	buf.WriteString("// dingo:s:1\n")
	mappings = append(mappings, Mapping{
		OriginalLine:    originalLine,
		OriginalColumn:  1,
		GeneratedLine:   startOutputLine + 1,
		GeneratedColumn: 1,
		Length:          len(matches[0]),
		Name:            "error_prop",
	})

	// Line 3: if __errN != nil {
	buf.WriteString(indent)
	buf.WriteString(fmt.Sprintf("if %s != nil {\n", errVar))
	mappings = append(mappings, Mapping{
		OriginalLine:    originalLine,
		OriginalColumn:  1,
		GeneratedLine:   startOutputLine + 2,
		GeneratedColumn: 1,
		Length:          len(matches[0]),
		Name:            "error_prop",
	})

	// Line 4: return zeroValues, wrapped_error
	buf.WriteString(indent)
	buf.WriteString("\t")
	buf.WriteString(e.generateReturnStatement(errVar, errMsg))
	buf.WriteString("\n")
	mappings = append(mappings, Mapping{
		OriginalLine:    originalLine,
		OriginalColumn:  1,
		GeneratedLine:   startOutputLine + 3,
		GeneratedColumn: 1,
		Length:          len(matches[0]),
		Name:            "error_prop",
	})

	// Line 5: }
	buf.WriteString(indent)
	buf.WriteString("}\n")
	mappings = append(mappings, Mapping{
		OriginalLine:    originalLine,
		OriginalColumn:  1,
		GeneratedLine:   startOutputLine + 4,
		GeneratedColumn: 1,
		Length:          len(matches[0]),
		Name:            "error_prop",
	})

	// Line 6: // dingo:e:1
	buf.WriteString(indent)
	buf.WriteString("// dingo:e:1\n")
	mappings = append(mappings, Mapping{
		OriginalLine:    originalLine,
		OriginalColumn:  1,
		GeneratedLine:   startOutputLine + 5,
		GeneratedColumn: 1,
		Length:          len(matches[0]),
		Name:            "error_prop",
	})

	// Line 7: return __tmpN, nil (complete tuple for functions returning multiple values)
	buf.WriteString(indent)
	// Generate complete return statement with all values
	var returnVals []string
	returnVals = append(returnVals, tmpVar)
	// Add nil for error position (last return value)
	if e.currentFunc != nil && len(e.currentFunc.returnTypes) > 1 {
		returnVals = append(returnVals, "nil")
	}
	buf.WriteString(fmt.Sprintf("return %s", strings.Join(returnVals, ", ")))
	mappings = append(mappings, Mapping{
		OriginalLine:    originalLine,
		OriginalColumn:  1,
		GeneratedLine:   startOutputLine + 6,
		GeneratedColumn: 1,
		Length:          len(matches[0]),
		Name:            "error_prop",
	})

	return buf.String(), mappings
}

// generateReturnStatement generates the return statement with proper zero values
func (e *ErrorPropProcessor) generateReturnStatement(errVar string, errMsg string) string {
	// Get zero values for return types
	var zeroVals []string
	if e.currentFunc != nil && len(e.currentFunc.zeroValues) > 0 {
		zeroVals = e.currentFunc.zeroValues
	} else {
		// Fallback: assume one return value (nil)
		zeroVals = []string{"nil"}
	}

	// Generate error part
	var errPart string
	if errMsg != "" {
		// IMPORTANT-1 FIX: Escape % characters to prevent fmt.Errorf runtime panics
		// Example: "failed: 50% complete" → "failed: 50%% complete"
		escapedMsg := strings.ReplaceAll(errMsg, "%", "%%")

		// Wrap with fmt.Errorf
		e.needsFmt = true
		errPart = fmt.Sprintf(`fmt.Errorf("%s: %%w", %s)`, escapedMsg, errVar)
	} else {
		// Pass through as-is
		errPart = errVar
	}

	// Combine: return zeroVal1, zeroVal2, ..., error
	if len(zeroVals) > 0 {
		// Function returns (T, error) or (T1, T2, ..., error)
		return fmt.Sprintf("return %s, %s", strings.Join(zeroVals, ", "), errPart)
	} else {
		// Function returns only error
		return fmt.Sprintf("return %s", errPart)
	}
}

// isFunctionDeclaration checks if a line is a function declaration
func (e *ErrorPropProcessor) isFunctionDeclaration(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "func ")
}

// parseFunctionSignature parses a function signature to extract return types
func (e *ErrorPropProcessor) parseFunctionSignature(startLine int) *funcContext {
	// Collect lines until we find the opening brace
	// Safety limit: search up to 20 lines for opening brace
	var funcText strings.Builder
	foundBrace := false
	maxLines := startLine + 20
	if maxLines > len(e.lines) {
		maxLines = len(e.lines)
	}

	for i := startLine; i < maxLines; i++ {
		funcText.WriteString(e.lines[i])
		funcText.WriteString("\n")

		trimmed := strings.TrimSpace(e.lines[i])
		// Skip comment lines
		if strings.HasPrefix(trimmed, "//") {
			continue
		}

		if idx := strings.Index(trimmed, "{"); idx != -1 {
			foundBrace = true
			break
		}
	}

	if !foundBrace {
		// No brace found - return safe fallback
		return &funcContext{
			returnTypes: []string{},
			zeroValues:  []string{"nil"},
		}
	}

	// Parse as Go function
	fset := token.NewFileSet()
	src := fmt.Sprintf("package p\n%s}", funcText.String())
	file, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		// Failed to parse, use nil fallback
		return &funcContext{
			returnTypes: []string{},
			zeroValues:  []string{"nil"},
		}
	}

	// Extract function declaration
	if len(file.Decls) == 0 {
		return &funcContext{
			returnTypes: []string{},
			zeroValues:  []string{"nil"},
		}
	}

	funcDecl, ok := file.Decls[0].(*ast.FuncDecl)
	if !ok || funcDecl.Type.Results == nil {
		return &funcContext{
			returnTypes: []string{},
			zeroValues:  []string{"nil"},
		}
	}

	// Extract return types
	returnTypes := []string{}
	for _, field := range funcDecl.Type.Results.List {
		typeStr := types.ExprString(field.Type)
		// If field has multiple names, repeat type (rare for returns)
		count := len(field.Names)
		if count == 0 {
			count = 1
		}
		for i := 0; i < count; i++ {
			returnTypes = append(returnTypes, typeStr)
		}
	}

	// Generate zero values (all except last, which is error)
	zeroValues := []string{}
	for i := 0; i < len(returnTypes)-1; i++ {
		zeroValues = append(zeroValues, getZeroValue(returnTypes[i]))
	}

	return &funcContext{
		returnTypes: returnTypes,
		zeroValues:  zeroValues,
	}
}

// getZeroValue returns the zero value for a given type
func getZeroValue(typ string) string {
	typ = strings.TrimSpace(typ)

	// Built-in types
	zeroMap := map[string]string{
		"int":     "0",
		"int8":    "0",
		"int16":   "0",
		"int32":   "0",
		"int64":   "0",
		"uint":    "0",
		"uint8":   "0",
		"uint16":  "0",
		"uint32":  "0",
		"uint64":  "0",
		"uintptr": "0",
		"float32": "0.0",
		"float64": "0.0",
		"string":  `""`,
		"bool":    "false",
		"error":   "nil",
		"byte":    "0",
		"rune":    "0",
	}

	if zero, ok := zeroMap[typ]; ok {
		return zero
	}

	// Pointer, slice, map, chan, interface → nil
	if strings.HasPrefix(typ, "*") ||
		strings.HasPrefix(typ, "[]") ||
		strings.HasPrefix(typ, "map[") ||
		strings.HasPrefix(typ, "chan ") ||
		strings.HasPrefix(typ, "<-chan ") ||
		strings.HasPrefix(typ, "chan<- ") ||
		typ == "interface{}" ||
		strings.HasPrefix(typ, "interface{") {
		return "nil"
	}

	// Function type → nil
	if strings.HasPrefix(typ, "func(") {
		return "nil"
	}

	// Custom type → use composite literal for non-pointer types
	if !strings.HasPrefix(typ, "*") &&
	   !strings.HasPrefix(typ, "[]") &&
	   !strings.HasPrefix(typ, "map[") &&
	   !strings.HasPrefix(typ, "chan ") &&
	   !strings.HasPrefix(typ, "<-chan ") &&
	   !strings.HasPrefix(typ, "chan<- ") &&
	   !strings.HasPrefix(typ, "func(") &&
	   !strings.HasPrefix(typ, "interface{") &&
	   typ != "interface{}" {
		return typ + "{}"
	}
	return "nil"
}

// getIndent extracts leading whitespace from a line
func (e *ErrorPropProcessor) getIndent(line string) string {
	for i, ch := range line {
		if ch != ' ' && ch != '\t' {
			return line[:i]
		}
	}
	return ""
}

// isTernaryLine checks if the line contains a ternary operator
func (e *ErrorPropProcessor) isTernaryLine(line string) bool {
	// Check for ternary pattern: expr ? value : value
	// Important: Must exclude : inside string literals (e.g., error messages)
	qPos := strings.Index(line, "?")
	if qPos == -1 {
		return false
	}

	// Scan after the ? to find : that's NOT in a string literal
	remainder := line[qPos+1:]
	inString := false
	escaped := false

	for _, ch := range remainder {
		if escaped {
			escaped = false
			continue
		}

		if ch == '\\' {
			escaped = true
			continue
		}

		if ch == '"' {
			inString = !inString
			continue
		}

		// Found : outside of string - this is a ternary
		if ch == ':' && !inString {
			return true
		}
	}

	return false
}

// ensureFmtImport adds fmt import if not present using go/ast parsing
func (e *ErrorPropProcessor) ensureFmtImport(source []byte) []byte {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", source, parser.ParseComments)
	if err != nil {
		// Fallback to simple insertion if parse fails
		return e.insertFmtImportSimple(source)
	}

	// Check if fmt is already imported
	for _, imp := range file.Imports {
		if imp.Path.Value == `"fmt"` {
			return source
		}
	}

	// Add fmt import using astutil
	astutil.AddImport(fset, file, "fmt")

	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, file); err != nil {
		return e.insertFmtImportSimple(source)
	}

	return buf.Bytes()
}

// insertFmtImportSimple is a fallback for when AST parsing fails
func (e *ErrorPropProcessor) insertFmtImportSimple(source []byte) []byte {
	sourceStr := string(source)
	lines := strings.Split(sourceStr, "\n")
	var result strings.Builder

	for i, line := range lines {
		result.WriteString(line)
		result.WriteString("\n")

		if strings.HasPrefix(strings.TrimSpace(line), "package ") {
			if i+1 < len(lines) && strings.Contains(lines[i+1], "import") {
				continue
			} else {
				result.WriteString("\nimport \"fmt\"\n")
			}
		}
	}

	return []byte(result.String())
}

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
	lineNum := 0

	for lineNum < len(e.lines) {
		line := e.lines[lineNum]

		// Check if this is a function declaration
		if e.isFunctionDeclaration(line) {
			e.currentFunc = e.parseFunctionSignature(lineNum)
			e.tryCounter = 0 // Reset counter for each function
		}

		// Process the line
		transformed, mapping := e.processLine(line, lineNum+1)
		output.WriteString(transformed)
		if lineNum < len(e.lines)-1 {
			output.WriteByte('\n')
		}
		if mapping != nil {
			mappings = append(mappings, *mapping)
		}

		lineNum++
	}

	// If we used fmt.Errorf, add import at the top
	result := output.Bytes()
	if e.needsFmt {
		result = e.ensureFmtImport(result)
	}

	return result, mappings, nil
}

// processLine processes a single line
func (e *ErrorPropProcessor) processLine(line string, lineNum int) (string, *Mapping) {
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
			result, mapping := e.expandAssignment(matches, expr, errMsg, lineNum)
			return result, mapping
		}
	}

	// Pattern: return EXPR? ["message"]
	if matches := returnPattern.FindStringSubmatch(line); matches != nil {
		returnPart := matches[1] // Everything after return
		if strings.Contains(returnPart, "?") {
			expr, errMsg := e.extractExpressionAndMessage(returnPart)
			result, mapping := e.expandReturn(matches, expr, errMsg, lineNum)
			return result, mapping
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
func (e *ErrorPropProcessor) expandAssignment(matches []string, expr string, errMsg string, lineNum int) (string, *Mapping) {
	keyword := matches[1]  // "let" or "var"
	varName := matches[2]  // variable name
	exprClean := strings.TrimSpace(strings.TrimSuffix(expr, "?"))

	tmpVar := fmt.Sprintf("__tmp%d", e.tryCounter)
	errVar := fmt.Sprintf("__err%d", e.tryCounter)
	e.tryCounter++

	// Generate the expansion
	var buf bytes.Buffer
	indent := e.getIndent(matches[0])

	// Line 1: __tmpN, __errN := expr
	buf.WriteString(indent)
	buf.WriteString(fmt.Sprintf("%s, %s := %s\n", tmpVar, errVar, exprClean))

	// Line 2: // dingo:s:1
	buf.WriteString(indent)
	buf.WriteString("// dingo:s:1\n")

	// Line 3: if __errN != nil {
	buf.WriteString(indent)
	buf.WriteString(fmt.Sprintf("if %s != nil {\n", errVar))

	// Line 4: return zeroValues, wrapped_error
	buf.WriteString(indent)
	buf.WriteString("\t")
	buf.WriteString(e.generateReturnStatement(errVar, errMsg))
	buf.WriteString("\n")

	// Line 5: }
	buf.WriteString(indent)
	buf.WriteString("}\n")

	// Line 6: // dingo:e:1
	buf.WriteString(indent)
	buf.WriteString("// dingo:e:1\n")

	// Line 7: var varName = __tmpN
	buf.WriteString(indent)
	buf.WriteString(fmt.Sprintf("%s %s = %s", keyword, varName, tmpVar))

	mapping := &Mapping{
		OriginalLine:    lineNum,
		OriginalColumn:  1,
		GeneratedLine:   lineNum,
		GeneratedColumn: 1,
		Length:          len(matches[0]),
		Name:            "error_prop",
	}

	return buf.String(), mapping
}

// expandReturn expands: return expr? → full error handling
func (e *ErrorPropProcessor) expandReturn(matches []string, expr string, errMsg string, lineNum int) (string, *Mapping) {
	exprClean := strings.TrimSpace(strings.TrimSuffix(expr, "?"))

	tmpVar := fmt.Sprintf("__tmp%d", e.tryCounter)
	errVar := fmt.Sprintf("__err%d", e.tryCounter)
	e.tryCounter++

	// Generate the expansion
	var buf bytes.Buffer
	indent := e.getIndent(matches[0])

	// Line 1: __tmpN, __errN := expr
	buf.WriteString(indent)
	buf.WriteString(fmt.Sprintf("%s, %s := %s\n", tmpVar, errVar, exprClean))

	// Line 2: // dingo:s:1
	buf.WriteString(indent)
	buf.WriteString("// dingo:s:1\n")

	// Line 3: if __errN != nil {
	buf.WriteString(indent)
	buf.WriteString(fmt.Sprintf("if %s != nil {\n", errVar))

	// Line 4: return zeroValues, wrapped_error
	buf.WriteString(indent)
	buf.WriteString("\t")
	buf.WriteString(e.generateReturnStatement(errVar, errMsg))
	buf.WriteString("\n")

	// Line 5: }
	buf.WriteString(indent)
	buf.WriteString("}\n")

	// Line 6: // dingo:e:1
	buf.WriteString(indent)
	buf.WriteString("// dingo:e:1\n")

	// Line 7: return __tmpN
	buf.WriteString(indent)
	buf.WriteString(fmt.Sprintf("return %s", tmpVar))

	mapping := &Mapping{
		OriginalLine:    lineNum,
		OriginalColumn:  1,
		GeneratedLine:   lineNum,
		GeneratedColumn: 1,
		Length:          len(matches[0]),
		Name:            "error_prop",
	}

	return buf.String(), mapping
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
		// Wrap with fmt.Errorf
		e.needsFmt = true
		errPart = fmt.Sprintf(`fmt.Errorf("%s: %%w", %s)`, errMsg, errVar)
	} else {
		// Pass through as-is
		errPart = errVar
	}

	// Combine: return zeroVal1, zeroVal2, ..., error
	return fmt.Sprintf("return %s, %s", strings.Join(zeroVals, ", "), errPart)
}

// isFunctionDeclaration checks if a line is a function declaration
func (e *ErrorPropProcessor) isFunctionDeclaration(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "func ")
}

// parseFunctionSignature parses a function signature to extract return types
func (e *ErrorPropProcessor) parseFunctionSignature(startLine int) *funcContext {
	// Collect lines until we find the opening brace
	var funcText strings.Builder
	for i := startLine; i < len(e.lines); i++ {
		funcText.WriteString(e.lines[i])
		funcText.WriteString("\n")
		if strings.Contains(e.lines[i], "{") {
			break
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
	// Simple heuristic: if there's a : after ?, it's likely ternary
	qPos := strings.Index(line, "?")
	if qPos == -1 {
		return false
	}
	remainder := line[qPos:]
	return strings.Contains(remainder, ":")
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

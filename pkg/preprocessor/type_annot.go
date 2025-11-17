package preprocessor

import (
	"bytes"
	"regexp"
)

// Package-level compiled regex (Issue 2: Regex Performance)
var (
	paramPattern = regexp.MustCompile(`(\w+)\s*:\s*(\w+|[\[\]\*\{\}]+[\w\.\[\]\*\{\}]*)`)
)

// TypeAnnotProcessor converts Dingo type annotations (: type) to Go syntax (space type)
type TypeAnnotProcessor struct{}

// NewTypeAnnotProcessor creates a new type annotation processor
func NewTypeAnnotProcessor() *TypeAnnotProcessor {
	return &TypeAnnotProcessor{}
}

// Name returns the processor name
func (t *TypeAnnotProcessor) Name() string {
	return "type_annotations"
}

// Process transforms type annotations
// Converts: func foo(x: int, y: string)
// To:       func foo(x int, y string)
func (t *TypeAnnotProcessor) Process(source []byte) ([]byte, []Mapping, error) {
	// Pattern: parameter_name: type in function signatures only
	// We need to be careful not to replace : in other contexts (maps, struct literals, etc.)

	// Strategy: Only replace in function parameter lists
	// Match: func name(...): inside (...), replace identifier: type → identifier type

	lines := bytes.Split(source, []byte("\n"))
	var result bytes.Buffer

	for i, line := range lines {
		// Check if this line contains a function declaration
		if bytes.Contains(line, []byte("func ")) {
			// Find the parameter list
			openParen := bytes.IndexByte(line, '(')
			closeParen := bytes.LastIndexByte(line, ')')

			if openParen != -1 && closeParen != -1 && closeParen > openParen {
				// Process only the parameter list
				before := line[:openParen+1]
				params := line[openParen+1:closeParen]
				after := line[closeParen:]

				// Replace : with space in parameters
				params = t.replaceColonInParams(params)

				result.Write(before)
				result.Write(params)
				result.Write(after)
			} else {
				result.Write(line)
			}
		} else {
			result.Write(line)
		}

		if i < len(lines)-1 {
			result.WriteByte('\n')
		}
	}

	return result.Bytes(), nil, nil
}

// replaceColonInParams replaces : with space in function parameters
func (t *TypeAnnotProcessor) replaceColonInParams(params []byte) []byte {
	// Use package-level compiled regex
	return paramPattern.ReplaceAllFunc(params, func(match []byte) []byte {
		parts := bytes.Split(match, []byte(":"))
		if len(parts) != 2 {
			return match
		}

		identifier := bytes.TrimSpace(parts[0])
		typeName := bytes.TrimSpace(parts[1])

		// Reconstruct as: identifier type (space instead of :)
		var buf bytes.Buffer
		buf.Write(identifier)
		buf.WriteByte(' ')
		buf.Write(typeName)

		return buf.Bytes()
	})
}

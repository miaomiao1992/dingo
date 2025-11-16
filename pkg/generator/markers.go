// Package generator - marker injection utilities
package generator

import (
	"fmt"
	"regexp"
	"strings"
)

// MarkerInjector handles injection of DINGO:GENERATED markers into Go source code
type MarkerInjector struct {
	enabled bool
}

// NewMarkerInjector creates a new marker injector
func NewMarkerInjector(enabled bool) *MarkerInjector {
	return &MarkerInjector{
		enabled: enabled,
	}
}

// InjectMarkers injects DINGO:GENERATED markers into generated Go code
// This is a post-processing step that runs after AST generation
func (m *MarkerInjector) InjectMarkers(source []byte) ([]byte, error) {
	if !m.enabled {
		return source, nil
	}

	sourceStr := string(source)

	// Pattern to detect error propagation generated code
	// Looks for: if __err0 != nil { return ... }
	errorCheckPattern := regexp.MustCompile(`(?m)(^[ \t]*if __err\d+ != nil \{[^}]*return[^}]*\}[ \t]*\n)`)

	// Inject markers around error propagation blocks
	result := errorCheckPattern.ReplaceAllStringFunc(sourceStr, func(match string) string {
		// Extract indentation from the if statement
		indent := ""
		if idx := strings.Index(match, "if"); idx > 0 {
			indent = match[:idx]
		}

		startMarker := fmt.Sprintf("%s// DINGO:GENERATED:START error_propagation\n", indent)
		endMarker := fmt.Sprintf("%s// DINGO:GENERATED:END\n", indent)

		return startMarker + match + endMarker
	})

	return []byte(result), nil
}

// injectErrorPropagationMarkers wraps error propagation blocks with markers
func (m *MarkerInjector) injectErrorPropagationMarkers(lines []string) []string {
	result := make([]string, 0, len(lines)+10)

	i := 0
	for i < len(lines) {
		line := lines[i]

		// Check if this is an error propagation pattern
		// Look for: variable assignment followed by error check
		if strings.Contains(line, "__err") && strings.Contains(line, ":=") {
			// Found start of error propagation block
			// Extract indentation
			indent := getIndentation(line)

			// Add start marker
			result = append(result, indent+"// DINGO:GENERATED:START error_propagation")
			result = append(result, line)
			i++

			// Add the error check (if __err != nil)
			if i < len(lines) && strings.Contains(lines[i], "if") && strings.Contains(lines[i], "__err") {
				// Copy the entire if block
				ifLine := lines[i]
				result = append(result, ifLine)
				i++

				// Count braces to find end of if block
				braceCount := strings.Count(ifLine, "{") - strings.Count(ifLine, "}")
				for i < len(lines) && braceCount > 0 {
					blockLine := lines[i]
					result = append(result, blockLine)
					braceCount += strings.Count(blockLine, "{") - strings.Count(blockLine, "}")
					i++
				}

				// Add end marker
				result = append(result, indent+"// DINGO:GENERATED:END")
			}
		} else {
			result = append(result, line)
			i++
		}
	}

	return result
}

// getIndentation extracts the indentation from a line
func getIndentation(line string) string {
	for i, ch := range line {
		if ch != ' ' && ch != '\t' {
			return line[:i]
		}
	}
	return ""
}

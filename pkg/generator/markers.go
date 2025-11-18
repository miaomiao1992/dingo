// Package generator - marker injection utilities
package generator

import (
	"fmt"
	"regexp"
	"strings"
)

// Plugin IDs for marker generation:
// 1 = error_propagation (? operator)
// 2 = result_type (Result<T, E>)
// 3 = option_type (Option<T>)
// 4 = pattern_matching (match expressions)
// 5 = sum_types (enum)

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

	// Check if markers are already present (added by preprocessor)
	// If so, skip injection to avoid duplicates
	if strings.Contains(sourceStr, "// dingo:s:") || strings.Contains(sourceStr, "// dingo:e:") {
		return source, nil
	}

	// Pattern to detect error propagation generated code
	// Looks for: if __err0 != nil { return ... }
	errorCheckPattern := regexp.MustCompile(`(?m)(^[ \t]*if __err\d+ != nil \{[^}]*return[^}]*\}[ \t]*\n)`)

	// Inject markers around error propagation blocks
	// Using plugin ID 1 for error_propagation
	result := errorCheckPattern.ReplaceAllStringFunc(sourceStr, func(match string) string {
		// Extract indentation from the if statement
		indent := ""
		if idx := strings.Index(match, "if"); idx > 0 {
			indent = match[:idx]
		}

		startMarker := fmt.Sprintf("%s// dingo:s:1\n", indent)
		endMarker := fmt.Sprintf("%s// dingo:e:1\n", indent)

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

			// Add start marker (plugin ID 1 = error_propagation)
			result = append(result, indent+"// dingo:s:1")
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

				// Add end marker (plugin ID 1 = error_propagation)
				result = append(result, indent+"// dingo:e:1")
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

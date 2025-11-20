package builtin

import (
	"fmt"
	"strings"
)

// SanitizeTypeName converts type name parts to camelCase
// Examples:
//   ("int", "error") → "IntError"
//   ("string", "option") → "StringOption"
//   ("any", "error") → "AnyError"
//   ("http", "request") → "HTTPRequest"
//   ("url", "parser") → "URLParser"
//   ("CustomError") → "CustomError"
func SanitizeTypeName(parts ...string) string {
	var result strings.Builder
	for _, part := range parts {
		result.WriteString(capitalizeTypeComponent(part))
	}
	return result.String()
}

// Package-level maps for performance (avoid recreating on every call)
var (
	// commonAcronyms maps lowercase acronyms to their canonical Go form.
	// Only include genuine acronyms (HTTP, URL, etc.), not regular words.
	// Regular words are handled by the default capitalization logic.
	commonAcronyms = map[string]string{
		"http":  "HTTP",
		"https": "HTTPS",
		"url":   "URL",
		"uri":   "URI",
		"json":  "JSON",
		"xml":   "XML",
		"api":   "API",
		"id":    "ID",
		"uuid":  "UUID",
		"sql":   "SQL",
		"html":  "HTML",
		"css":   "CSS",
		"tcp":   "TCP",
		"udp":   "UDP",
		"ip":    "IP",
	}

	// builtinTypes contains Go built-in types that should only capitalize the first letter
	builtinTypes = map[string]bool{
		"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
		"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
		"float32": true, "float64": true,
		"string": true, "bool": true, "byte": true, "rune": true,
		"error": true, "any": true,
	}
)

// capitalizeTypeComponent handles acronyms and built-in types
func capitalizeTypeComponent(s string) string {
	if s == "" {
		return s
	}

	lower := strings.ToLower(s)
	if upper, ok := commonAcronyms[lower]; ok {
		return upper
	}

	// Built-in types - capitalize first letter only
	if builtinTypes[lower] {
		return strings.ToUpper(s[:1]) + s[1:]
	}

	// Single lowercase letter - capitalize it
	if len(s) == 1 && s[0] >= 'a' && s[0] <= 'z' {
		return strings.ToUpper(s)
	}

	// User-defined type - preserve original capitalization if already capitalized
	// Otherwise capitalize first letter
	if len(s) > 0 && s[0] >= 'A' && s[0] <= 'Z' {
		return s // Already capitalized, preserve
	}

	// Lowercase user type - capitalize first letter
	return strings.ToUpper(s[:1]) + s[1:]
}

// GenerateTempVarName generates temporary variable names with optional numbering
// First call returns base name (e.g., "ok"), subsequent calls add numbers ("ok1", "ok2")
// Examples:
//   ("ok", 0) → "ok"
//   ("ok", 1) → "ok1"
//   ("err", 0) → "err"
//   ("err", 1) → "err1"
func GenerateTempVarName(base string, index int) string {
	if index < 0 {
		index = 0 // Defensive: treat negative as zero
	}
	if index == 0 {
		return base // First variable: no number suffix
	}
	return fmt.Sprintf("%s%d", base, index) // ok1, ok2, ok3, ...
}

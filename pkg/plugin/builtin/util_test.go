package builtin

import "testing"

func TestSanitizeTypeName(t *testing.T) {
	tests := []struct {
		name     string
		parts    []string
		expected string
	}{
		// Single built-in types
		{
			name:     "int",
			parts:    []string{"int"},
			expected: "Int",
		},
		{
			name:     "string",
			parts:    []string{"string"},
			expected: "String",
		},
		{
			name:     "error",
			parts:    []string{"error"},
			expected: "Error",
		},
		{
			name:     "bool",
			parts:    []string{"bool"},
			expected: "Bool",
		},
		{
			name:     "any",
			parts:    []string{"any"},
			expected: "Any",
		},

		// Two-part type names
		{
			name:     "int + error",
			parts:    []string{"int", "error"},
			expected: "IntError",
		},
		{
			name:     "string + option",
			parts:    []string{"string", "option"},
			expected: "StringOption",
		},
		{
			name:     "any + error",
			parts:    []string{"any", "error"},
			expected: "AnyError",
		},

		// Acronyms (all-caps)
		{
			name:     "http",
			parts:    []string{"http"},
			expected: "HTTP",
		},
		{
			name:     "url",
			parts:    []string{"url"},
			expected: "URL",
		},
		{
			name:     "json",
			parts:    []string{"json"},
			expected: "JSON",
		},
		{
			name:     "xml",
			parts:    []string{"xml"},
			expected: "XML",
		},
		{
			name:     "api",
			parts:    []string{"api"},
			expected: "API",
		},
		{
			name:     "id",
			parts:    []string{"id"},
			expected: "ID",
		},
		{
			name:     "uuid",
			parts:    []string{"uuid"},
			expected: "UUID",
		},

		// Acronyms in compound types
		{
			name:     "http + request",
			parts:    []string{"http", "request"},
			expected: "HTTPRequest",
		},
		{
			name:     "url + parser",
			parts:    []string{"url", "parser"},
			expected: "URLParser",
		},
		{
			name:     "json + error",
			parts:    []string{"json", "error"},
			expected: "JSONError",
		},

		// User-defined types (preserve capitalization)
		{
			name:     "User",
			parts:    []string{"User"},
			expected: "User",
		},
		{
			name:     "CustomError",
			parts:    []string{"CustomError"},
			expected: "CustomError",
		},
		{
			name:     "UserID",
			parts:    []string{"UserID"},
			expected: "UserID",
		},

		// User types in compound names
		{
			name:     "CustomError + int",
			parts:    []string{"CustomError", "int"},
			expected: "CustomErrorInt",
		},
		{
			name:     "int + CustomError",
			parts:    []string{"int", "CustomError"},
			expected: "IntCustomError",
		},
		{
			name:     "UserID + error",
			parts:    []string{"UserID", "error"},
			expected: "UserIDError",
		},

		// Mixed cases
		{
			name:     "HTTP + CustomError",
			parts:    []string{"http", "CustomError"},
			expected: "HTTPCustomError",
		},
		{
			name:     "URL + UserID",
			parts:    []string{"url", "UserID"},
			expected: "URLUserID",
		},

		// Three-part names
		{
			name:     "int + string + error",
			parts:    []string{"int", "string", "error"},
			expected: "IntStringError",
		},
		{
			name:     "http + request + error",
			parts:    []string{"http", "request", "error"},
			expected: "HTTPRequestError",
		},

		// Edge cases
		{
			name:     "empty string part",
			parts:    []string{"", "error"},
			expected: "Error",
		},
		{
			name:     "single character",
			parts:    []string{"a"},
			expected: "A",
		},
		{
			name:     "numeric types",
			parts:    []string{"int64", "error"},
			expected: "Int64Error",
		},
		{
			name:     "uint types",
			parts:    []string{"uint32"},
			expected: "Uint32",
		},
		{
			name:     "float types",
			parts:    []string{"float64", "error"},
			expected: "Float64Error",
		},

		// Case sensitivity tests
		{
			name:     "HTTP (already caps) stays HTTP",
			parts:    []string{"HTTP"},
			expected: "HTTP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeTypeName(tt.parts...)
			if result != tt.expected {
				t.Errorf("SanitizeTypeName(%v) = %q, want %q",
					tt.parts, result, tt.expected)
			}
		})
	}
}

func TestGenerateTempVarName(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		index    int
		expected string
	}{
		// First variable (no number suffix)
		{
			name:     "ok first",
			base:     "ok",
			index:    0,
			expected: "ok",
		},
		{
			name:     "err first",
			base:     "err",
			index:    0,
			expected: "err",
		},
		{
			name:     "tmp first",
			base:     "tmp",
			index:    0,
			expected: "tmp",
		},

		// Second variable (add number)
		{
			name:     "ok second",
			base:     "ok",
			index:    1,
			expected: "ok1",
		},
		{
			name:     "err second",
			base:     "err",
			index:    1,
			expected: "err1",
		},
		{
			name:     "tmp second",
			base:     "tmp",
			index:    1,
			expected: "tmp1",
		},

		// Third variable
		{
			name:     "ok third",
			base:     "ok",
			index:    2,
			expected: "ok2",
		},
		{
			name:     "err third",
			base:     "err",
			index:    2,
			expected: "err2",
		},

		// Higher indices
		{
			name:     "ok tenth",
			base:     "ok",
			index:    9,
			expected: "ok9",
		},
		{
			name:     "err twentieth",
			base:     "err",
			index:    19,
			expected: "err19",
		},

		// Different base names
		{
			name:     "val first",
			base:     "val",
			index:    0,
			expected: "val",
		},
		{
			name:     "val second",
			base:     "val",
			index:    1,
			expected: "val1",
		},
		{
			name:     "result first",
			base:     "result",
			index:    0,
			expected: "result",
		},
		{
			name:     "result second",
			base:     "result",
			index:    1,
			expected: "result1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateTempVarName(tt.base, tt.index)
			if result != tt.expected {
				t.Errorf("GenerateTempVarName(%q, %d) = %q, want %q",
					tt.base, tt.index, result, tt.expected)
			}
		})
	}
}

func TestCapitalizeTypeComponent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Built-in types
		{
			name:     "int",
			input:    "int",
			expected: "Int",
		},
		{
			name:     "string",
			input:    "string",
			expected: "String",
		},
		{
			name:     "any",
			input:    "any",
			expected: "Any",
		},

		// Acronyms
		{
			name:     "http lowercase",
			input:    "http",
			expected: "HTTP",
		},
		{
			name:     "HTTP uppercase",
			input:    "HTTP",
			expected: "HTTP",
		},
		{
			name:     "url lowercase",
			input:    "url",
			expected: "URL",
		},
		{
			name:     "URL uppercase",
			input:    "URL",
			expected: "URL",
		},

		// User types
		{
			name:     "User",
			input:    "User",
			expected: "User",
		},
		{
			name:     "CustomError",
			input:    "CustomError",
			expected: "CustomError",
		},
		{
			name:     "myType (user defined - gets capitalized)",
			input:    "myType",
			expected: "MyType",
		},

		// Edge cases
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single letter",
			input:    "a",
			expected: "A",
		},
		{
			name:     "single letter uppercase",
			input:    "A",
			expected: "A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := capitalizeTypeComponent(tt.input)
			if result != tt.expected {
				t.Errorf("capitalizeTypeComponent(%q) = %q, want %q",
					tt.input, result, tt.expected)
			}
		})
	}
}

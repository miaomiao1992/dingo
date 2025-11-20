package preprocessor

import (
	"strings"
	"testing"
)

// TestLambdaErrorDetection_TypeScriptStyle tests error detection for TypeScript arrow syntax
func TestLambdaErrorDetection_TypeScriptStyle(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "single param without type - standalone",
			input:       "let f = x => x * 2",
			shouldError: true,
			errorMsg:    "Cannot infer lambda parameter type",
		},
		{
			name:        "single param with type - OK",
			input:       "let f = (x: int) => x * 2",
			shouldError: false,
		},
		{
			name:        "multi param without types",
			input:       "let f = (x, y) => x + y",
			shouldError: true,
			errorMsg:    "Cannot infer lambda parameter type",
		},
		{
			name:        "multi param with types - OK",
			input:       "let f = (x: int, y: int) => x + y",
			shouldError: false,
		},
		{
			name:        "mixed typed and untyped params",
			input:       "let f = (x: int, y) => x + y",
			shouldError: true,
			errorMsg:    "Cannot infer lambda parameter type",
		},
		{
			name:        "single param with return type - OK",
			input:       "let f = (x: int): int => x * 2",
			shouldError: false,
		},
		{
			name:        "empty params - OK",
			input:       "let f = () => 42",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewLambdaProcessor().WithStrictTypeChecking(true)
			_, _, err := processor.Process([]byte(tt.input))

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain %q, got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestLambdaErrorDetection_RustStyle tests error detection for Rust pipe syntax
func TestLambdaErrorDetection_RustStyle(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "single param without type",
			input:       "let f = |x| x * 2",
			shouldError: true,
			errorMsg:    "Cannot infer lambda parameter type",
		},
		{
			name:        "single param with type - OK",
			input:       "let f = |x: int| x * 2",
			shouldError: false,
		},
		{
			name:        "multi param without types",
			input:       "let f = |x, y| x + y",
			shouldError: true,
			errorMsg:    "Cannot infer lambda parameter type",
		},
		{
			name:        "multi param with types - OK",
			input:       "let f = |x: int, y: int| x + y",
			shouldError: false,
		},
		{
			name:        "mixed typed and untyped params",
			input:       "let f = |x: int, y| x + y",
			shouldError: true,
			errorMsg:    "Cannot infer lambda parameter type",
		},
		{
			name:        "with return type - OK",
			input:       "let f = |x: int| -> int { x * 2 }",
			shouldError: false,
		},
		{
			name:        "empty params - OK",
			input:       "let f = || 42",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := &LambdaProcessor{style: StyleRust, strictTypeChecking: true}
			_, _, err := processor.Process([]byte(tt.input))

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain %q, got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestLambdaErrorMessages tests that error messages are helpful and actionable
func TestLambdaErrorMessages(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		style           LambdaStyle
		wantContains    []string
	}{
		{
			name:  "TypeScript error message",
			input: "let f = x => x * 2",
			style: StyleTypeScript,
			wantContains: []string{
				"Cannot infer lambda parameter type",
				"Missing type annotation",
				"Example:",
				"(x: int) => x * 2",
			},
		},
		{
			name:  "Rust error message",
			input: "let f = |x| x * 2",
			style: StyleRust,
			wantContains: []string{
				"Cannot infer lambda parameter type",
				"Missing type annotation",
				"Example:",
				"|x: int| x * 2",
			},
		},
		{
			name:  "Multi-param TypeScript error",
			input: "let f = (x, y) => x + y",
			style: StyleTypeScript,
			wantContains: []string{
				"Cannot infer lambda parameter type",
				"(x: int, y: int) => x + y",
			},
		},
		{
			name:  "Multi-param Rust error",
			input: "let f = |x, y| x + y",
			style: StyleRust,
			wantContains: []string{
				"Cannot infer lambda parameter type",
				"|x: int, y: int| x + y",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := &LambdaProcessor{style: tt.style, strictTypeChecking: true}
			_, _, err := processor.Process([]byte(tt.input))

			if err == nil {
				t.Fatalf("Expected error but got none")
			}

			errMsg := err.Error()
			for _, want := range tt.wantContains {
				if !strings.Contains(errMsg, want) {
					t.Errorf("Error message missing %q.\nGot: %s", want, errMsg)
				}
			}
		})
	}
}

// TestLambdaNoErrorForValidSyntax ensures valid lambdas don't trigger errors
func TestLambdaNoErrorForValidSyntax(t *testing.T) {
	validCases := []struct {
		name  string
		input string
		style LambdaStyle
	}{
		{"TS single typed", "(x: int) => x * 2", StyleTypeScript},
		{"TS multi typed", "(x: int, y: int) => x + y", StyleTypeScript},
		{"TS with return type", "(x: int): int => x * 2", StyleTypeScript},
		{"TS empty params", "() => 42", StyleTypeScript},
		{"Rust single typed", "|x: int| x * 2", StyleRust},
		{"Rust multi typed", "|x: int, y: int| x + y", StyleRust},
		{"Rust with return type", "|x: int| -> int { x * 2 }", StyleRust},
		{"Rust empty params", "|| 42", StyleRust},
	}

	for _, tc := range validCases {
		t.Run(tc.name, func(t *testing.T) {
			processor := &LambdaProcessor{style: tc.style, strictTypeChecking: true}
			_, _, err := processor.Process([]byte(tc.input))

			if err != nil {
				t.Errorf("Expected no error for valid syntax, got: %v", err)
			}
		})
	}
}

// TestLambdaErrorLineNumbers tests that errors report correct line numbers
func TestLambdaErrorLineNumbers(t *testing.T) {
	input := `package main

func main() {
	let f = x => x * 2
	let g = (y: int) => y + 1
}`

	processor := NewLambdaProcessor().WithStrictTypeChecking(true)
	_, _, err := processor.Process([]byte(input))

	if err == nil {
		t.Fatalf("Expected error but got none")
	}

	// Check that error mentions line 4 (where x => x * 2 is)
	if !strings.Contains(err.Error(), "4") {
		t.Errorf("Expected error to mention line 4, got: %v", err)
	}
}

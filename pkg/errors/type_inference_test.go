package errors

import (
	"go/token"
	"strings"
	"testing"
)

func TestCompileError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *CompileError
		expected string
	}{
		{
			name: "type inference error",
			err: &CompileError{
				Message:  "cannot infer type for expression: x",
				Category: ErrorCategoryTypeInference,
			},
			expected: "Type Inference Error: cannot infer type for expression: x",
		},
		{
			name: "code generation error",
			err: &CompileError{
				Message:  "cannot generate code",
				Category: ErrorCategoryCodeGeneration,
			},
			expected: "Code Generation Error: cannot generate code",
		},
		{
			name: "syntax error",
			err: &CompileError{
				Message:  "unexpected token",
				Category: ErrorCategorySyntax,
			},
			expected: "Syntax Error: unexpected token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestNewTypeInferenceError(t *testing.T) {
	err := NewTypeInferenceError("test message", token.Pos(42), "test hint")

	if err.Message != "test message" {
		t.Errorf("Message = %q, want %q", err.Message, "test message")
	}
	if err.Location != token.Pos(42) {
		t.Errorf("Location = %d, want %d", err.Location, 42)
	}
	if err.Hint != "test hint" {
		t.Errorf("Hint = %q, want %q", err.Hint, "test hint")
	}
	if err.Category != ErrorCategoryTypeInference {
		t.Errorf("Category = %d, want %d", err.Category, ErrorCategoryTypeInference)
	}
}

func TestNewCodeGenerationError(t *testing.T) {
	err := NewCodeGenerationError("gen error", token.Pos(100), "fix hint")

	if err.Category != ErrorCategoryCodeGeneration {
		t.Errorf("Category = %d, want %d", err.Category, ErrorCategoryCodeGeneration)
	}
}

func TestFormatWithPosition(t *testing.T) {
	fset := token.NewFileSet()
	file := fset.AddFile("test.dingo", -1, 100)

	// Create a position in the file
	pos := file.Pos(10)

	err := &CompileError{
		Message:  "test error",
		Location: pos,
		Category: ErrorCategoryTypeInference,
		Hint:     "try this fix",
	}

	formatted := err.FormatWithPosition(fset)

	// Should contain filename, line, column, category, message, and hint
	if !strings.Contains(formatted, "test.dingo") {
		t.Errorf("formatted error missing filename: %s", formatted)
	}
	if !strings.Contains(formatted, "Type Inference Error") {
		t.Errorf("formatted error missing category: %s", formatted)
	}
	if !strings.Contains(formatted, "test error") {
		t.Errorf("formatted error missing message: %s", formatted)
	}
	if !strings.Contains(formatted, "Hint: try this fix") {
		t.Errorf("formatted error missing hint: %s", formatted)
	}
}

func TestFormatWithPosition_NoFileSet(t *testing.T) {
	err := &CompileError{
		Message:  "test error",
		Location: token.Pos(42),
		Category: ErrorCategoryTypeInference,
	}

	// Should fall back to Error() when fset is nil
	formatted := err.FormatWithPosition(nil)
	expected := err.Error()

	if formatted != expected {
		t.Errorf("FormatWithPosition(nil) = %q, want %q", formatted, expected)
	}
}

func TestTypeInferenceFailure(t *testing.T) {
	err := TypeInferenceFailure("myExpr", token.Pos(50))

	if !strings.Contains(err.Message, "myExpr") {
		t.Errorf("Message should contain expression: %s", err.Message)
	}
	if !strings.Contains(err.Message, "cannot infer type") {
		t.Errorf("Message should mention type inference: %s", err.Message)
	}
	if !strings.Contains(err.Hint, "explicit type annotation") {
		t.Errorf("Hint should suggest type annotation: %s", err.Hint)
	}
	if err.Category != ErrorCategoryTypeInference {
		t.Errorf("Category = %d, want %d", err.Category, ErrorCategoryTypeInference)
	}
}

func TestLiteralAddressError(t *testing.T) {
	err := LiteralAddressError("42", token.Pos(30))

	if !strings.Contains(err.Message, "42") {
		t.Errorf("Message should contain literal: %s", err.Message)
	}
	if !strings.Contains(err.Message, "cannot take address") {
		t.Errorf("Message should mention address error: %s", err.Message)
	}
	if !strings.Contains(err.Hint, "IIFE") {
		t.Errorf("Hint should mention IIFE pattern: %s", err.Hint)
	}
	if err.Category != ErrorCategoryCodeGeneration {
		t.Errorf("Category = %d, want %d", err.Category, ErrorCategoryCodeGeneration)
	}
}

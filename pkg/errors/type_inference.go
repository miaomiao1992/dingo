// Package errors provides error types and reporting infrastructure for the Dingo compiler
package errors

import (
	"fmt"
	"go/token"
)

// CompileError represents a compile-time error in Dingo code
type CompileError struct {
	Message  string    // Human-readable error message
	Location token.Pos // Position in source file
	Hint     string    // Suggestion for fixing the error
	Category ErrorCategory
}

// ErrorCategory categorizes different types of compile errors
type ErrorCategory int

const (
	// ErrorCategoryTypeInference indicates a type inference failure
	ErrorCategoryTypeInference ErrorCategory = iota
	// ErrorCategoryCodeGeneration indicates a code generation failure
	ErrorCategoryCodeGeneration
	// ErrorCategorySyntax indicates a syntax error
	ErrorCategorySyntax
)

// Error implements the error interface
func (e *CompileError) Error() string {
	return fmt.Sprintf("%s: %s", e.categoryString(), e.Message)
}

func (e *CompileError) categoryString() string {
	switch e.Category {
	case ErrorCategoryTypeInference:
		return "Type Inference Error"
	case ErrorCategoryCodeGeneration:
		return "Code Generation Error"
	case ErrorCategorySyntax:
		return "Syntax Error"
	default:
		return "Compile Error"
	}
}

// NewTypeInferenceError creates a new type inference error
func NewTypeInferenceError(message string, location token.Pos, hint string) *CompileError {
	return &CompileError{
		Message:  message,
		Location: location,
		Hint:     hint,
		Category: ErrorCategoryTypeInference,
	}
}

// NewCodeGenerationError creates a new code generation error
func NewCodeGenerationError(message string, location token.Pos, hint string) *CompileError {
	return &CompileError{
		Message:  message,
		Location: location,
		Hint:     hint,
		Category: ErrorCategoryCodeGeneration,
	}
}

// FormatWithPosition formats the error with file position information
func (e *CompileError) FormatWithPosition(fset *token.FileSet) string {
	if fset == nil || !e.Location.IsValid() {
		return e.Error()
	}

	pos := fset.Position(e.Location)
	msg := fmt.Sprintf("%s:%d:%d: %s: %s",
		pos.Filename,
		pos.Line,
		pos.Column,
		e.categoryString(),
		e.Message,
	)

	if e.Hint != "" {
		msg += fmt.Sprintf("\n  Hint: %s", e.Hint)
	}

	return msg
}

// TypeInferenceFailure creates a standardized type inference failure error
func TypeInferenceFailure(exprString string, location token.Pos) *CompileError {
	return NewTypeInferenceError(
		fmt.Sprintf("cannot infer type for expression: %s", exprString),
		location,
		"Try providing an explicit type annotation, e.g., var x: int = ...",
	)
}

// LiteralAddressError creates an error for invalid literal addressing
func LiteralAddressError(exprString string, location token.Pos) *CompileError {
	return NewCodeGenerationError(
		fmt.Sprintf("cannot take address of literal: %s", exprString),
		location,
		"Use a variable or let the compiler wrap it in a temporary (IIFE pattern)",
	)
}

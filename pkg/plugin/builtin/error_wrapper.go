// Package builtin provides built-in Dingo transformation plugins
package builtin

import (
	"go/ast"
	"go/token"
	"strings"
)

// ErrorWrapper handles error message wrapping with fmt.Errorf
type ErrorWrapper struct{}

// NewErrorWrapper creates a new error wrapper
func NewErrorWrapper() *ErrorWrapper {
	return &ErrorWrapper{}
}

// WrapError generates a fmt.Errorf call that wraps an error with a message
// Example: fmt.Errorf("failed to fetch user: %w", err)
func (ew *ErrorWrapper) WrapError(errVar string, message string) ast.Expr {
	// Escape the user message
	escapedMsg := ew.escapeString(message)

	// Create format string: "message: %w" (note: use literal %w, not %%w)
	// The outer quotes are part of the string literal value
	formatStr := `"` + escapedMsg + `: %w"`

	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "fmt"},
			Sel: &ast.Ident{Name: "Errorf"},
		},
		Args: []ast.Expr{
			&ast.BasicLit{
				Kind:  token.STRING,
				Value: formatStr,
			},
			&ast.Ident{Name: errVar},
		},
	}
}

// escapeString escapes special characters in error messages
func (ew *ErrorWrapper) escapeString(s string) string {
	// Escape backslashes first (must be first!)
	s = strings.ReplaceAll(s, `\`, `\\`)
	// Escape double quotes
	s = strings.ReplaceAll(s, `"`, `\"`)
	// Escape newlines
	s = strings.ReplaceAll(s, "\n", `\n`)
	// Escape tabs
	s = strings.ReplaceAll(s, "\t", `\t`)
	// Escape carriage returns
	s = strings.ReplaceAll(s, "\r", `\r`)
	// Escape form feeds
	s = strings.ReplaceAll(s, "\f", `\f`)
	return s
}

// NeedsImport checks if fmt import is needed
// This should be called to determine if we need to add "fmt" to imports
func (ew *ErrorWrapper) NeedsImport() bool {
	return true
}

// AddFmtImport adds the fmt import to a file if not already present
func (ew *ErrorWrapper) AddFmtImport(file *ast.File) {
	// Check if fmt is already imported
	for _, imp := range file.Imports {
		if imp.Path != nil && imp.Path.Value == `"fmt"` {
			// Already imported
			return
		}
	}

	// Find the import declaration or create one
	var importDecl *ast.GenDecl
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			importDecl = genDecl
			break
		}
	}

	// Create fmt import spec
	fmtImport := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: `"fmt"`,
		},
	}

	if importDecl != nil {
		// Add to existing import declaration
		importDecl.Specs = append(importDecl.Specs, fmtImport)
	} else {
		// Create new import declaration
		importDecl = &ast.GenDecl{
			Tok: token.IMPORT,
			Specs: []ast.Spec{fmtImport},
		}

		// Insert at the beginning of declarations (after package)
		newDecls := make([]ast.Decl, 0, len(file.Decls)+1)
		newDecls = append(newDecls, importDecl)
		newDecls = append(newDecls, file.Decls...)
		file.Decls = newDecls
	}

	// Update file.Imports
	file.Imports = append(file.Imports, fmtImport)
}

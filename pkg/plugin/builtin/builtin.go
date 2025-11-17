// Package builtin provides default plugins
package builtin

import (
	"go/ast"
	"go/token"

	"github.com/MadAppGang/dingo/pkg/plugin"
)

// NewDefaultRegistry creates a registry with default plugins
func NewDefaultRegistry() (*plugin.Registry, error) {
	return plugin.NewRegistry(), nil
}

// NewTypeInferenceService creates a type inference service (stub)
func NewTypeInferenceService(fset *token.FileSet, file *ast.File, logger plugin.Logger) (interface{}, error) {
	return nil, nil
}

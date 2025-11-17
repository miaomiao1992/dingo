// Package transform implements AST transformations for Dingo features
package transform

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/ast/astutil"

	"github.com/MadAppGang/dingo/pkg/preprocessor"
)

// Transformer walks the AST and replaces Dingo placeholder patterns
// with final Go implementations
type Transformer struct {
	fset      *token.FileSet
	sourceMap *preprocessor.SourceMap
	typeInfo  *types.Info
}

// New creates a new transformer
func New(fset *token.FileSet, sourceMap *preprocessor.SourceMap) *Transformer {
	return &Transformer{
		fset:      fset,
		sourceMap: sourceMap,
		typeInfo: &types.Info{
			Types: make(map[ast.Expr]types.TypeAndValue),
			Defs:  make(map[*ast.Ident]types.Object),
			Uses:  make(map[*ast.Ident]types.Object),
		},
	}
}

// Transform applies all transformations to the AST
func (t *Transformer) Transform(file *ast.File) (*ast.File, error) {
	// Step 1: Type check to populate type info
	// (Skipped for now - will add when needed for lambdas)

	// Step 2: Walk and transform
	result := astutil.Apply(file, t.visit, nil)

	if result == nil {
		return file, nil
	}

	return result.(*ast.File), nil
}

// visit is called for each node during AST traversal
func (t *Transformer) visit(cursor *astutil.Cursor) bool {
	node := cursor.Node()
	if node == nil {
		return true
	}

	switch n := node.(type) {
	case *ast.CallExpr:
		// Check for Dingo placeholder function calls
		if ident, ok := n.Fun.(*ast.Ident); ok {
			return t.handlePlaceholderCall(cursor, ident, n)
		}

	case *ast.GenDecl:
		// Check for enum type definitions
		return t.handleGenDecl(cursor, n)
	}

	return true // Continue traversal
}

// handlePlaceholderCall processes calls to Dingo placeholder functions
func (t *Transformer) handlePlaceholderCall(cursor *astutil.Cursor, ident *ast.Ident, call *ast.CallExpr) bool {
	name := ident.Name

	switch {
	// Error propagation is fully handled in preprocessor, no transform needed

	case len(name) >= 15 && name[:15] == "__dingo_lambda_":
		// Lambda: __dingo_lambda_N__(...)
		return t.transformLambda(cursor, call)

	case len(name) >= 14 && name[:14] == "__dingo_match_":
		// Pattern match: __dingo_match_N__(...)
		return t.transformMatch(cursor, call)

	case len(name) >= 17 && name[:17] == "__dingo_safe_nav_":
		// Safe navigation: __dingo_safe_nav_N__(...)
		return t.transformSafeNav(cursor, call)
	}

	return true
}

// handleGenDecl processes general declarations (enum types)
func (t *Transformer) handleGenDecl(cursor *astutil.Cursor, decl *ast.GenDecl) bool {
	// Check if this is an enum definition
	// (Will implement when we add sum types)
	return true
}

// transformErrorProp transforms error propagation placeholders
func (t *Transformer) transformErrorProp(cursor *astutil.Cursor, call *ast.CallExpr) bool {
	// TODO: Implement error propagation transformation
	// For now, leave as-is
	return true
}

// transformLambda transforms lambda placeholders
func (t *Transformer) transformLambda(cursor *astutil.Cursor, call *ast.CallExpr) bool {
	// TODO: Implement lambda transformation
	return true
}

// transformMatch transforms pattern matching placeholders
func (t *Transformer) transformMatch(cursor *astutil.Cursor, call *ast.CallExpr) bool {
	// TODO: Implement pattern matching transformation
	return true
}

// transformSafeNav transforms safe navigation placeholders
func (t *Transformer) transformSafeNav(cursor *astutil.Cursor, call *ast.CallExpr) bool {
	// TODO: Implement safe navigation transformation
	return true
}

// analyzeContext determines the context where an expression appears
// (assignment, return statement, standalone, etc.)
type ExprContext int

const (
	ContextUnknown ExprContext = iota
	ContextAssignment
	ContextReturn
	ContextStandalone
	ContextCondition
)

// getExprContext analyzes the context of an expression
func (t *Transformer) getExprContext(cursor *astutil.Cursor) ExprContext {
	// Walk up the AST to find parent context
	// TODO: Implement context detection
	return ContextUnknown
}

// TransformError wraps transformation errors with context
type TransformError struct {
	Node ast.Node
	Msg  string
	Err  error
}

func (e *TransformError) Error() string {
	return fmt.Sprintf("transform error at %v: %s: %v", e.Node, e.Msg, e.Err)
}

func (e *TransformError) Unwrap() error {
	return e.Err
}

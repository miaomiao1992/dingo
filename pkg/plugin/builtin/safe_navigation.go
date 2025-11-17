// Package builtin provides built-in Dingo transformation plugins
package builtin

import (
	"fmt"
	"go/ast"
	"go/token"

	dingoast "github.com/MadAppGang/dingo/pkg/ast"
	"github.com/MadAppGang/dingo/pkg/config"
	"github.com/MadAppGang/dingo/pkg/plugin"
)

// SafeNavigationPlugin transforms safe navigation expressions (?.) into Go nil checks
//
// Features:
// - Smart unwrapping based on context (default)
// - Always Option<T> mode for strict type safety
// - Chain optimization for multiple ?. operators
//
// Transforms (smart mode, expecting T):
//   user?.name  →  if user != nil { user.Name } else { "" }
//
// Transforms (always_option mode):
//   user?.name  →  if user != nil { Option_Some(user.Name) } else { Option_None[string]() }
type SafeNavigationPlugin struct {
	plugin.BasePlugin
}

// NewSafeNavigationPlugin creates a new safe navigation plugin
func NewSafeNavigationPlugin() *SafeNavigationPlugin {
	return &SafeNavigationPlugin{
		BasePlugin: *plugin.NewBasePlugin(
			"safe_navigation",
			"Safe navigation operator (?.) with configurable unwrapping",
			nil,
		),
	}
}

// Transform transforms safe navigation expressions
func (p *SafeNavigationPlugin) Transform(ctx *plugin.Context, node ast.Node) (ast.Node, error) {
	// Check if this is a safe navigation expression
	safeNav, ok := node.(*dingoast.SafeNavigationExpr)
	if !ok {
		return node, nil
	}

	// Get configuration
	var unwrapMode string
	if ctx.DingoConfig != nil {
		if cfg, ok := ctx.DingoConfig.(*config.Config); ok {
			unwrapMode = cfg.Features.SafeNavigationUnwrap
		}
	}
	if unwrapMode == "" {
		unwrapMode = "smart" // Default
	}

	// Transform based on mode
	switch unwrapMode {
	case "always_option":
		return p.transformAlwaysOption(safeNav)
	case "smart":
		return p.transformSmart(safeNav)
	default:
		return nil, fmt.Errorf("invalid safe_navigation_unwrap mode: %s", unwrapMode)
	}
}

// transformAlwaysOption generates Option<T> return for all safe navigation
func (p *SafeNavigationPlugin) transformAlwaysOption(safeNav *dingoast.SafeNavigationExpr) (ast.Node, error) {
	// Generate: user != nil ? Option_Some(user.Name) : Option_None[string]()
	//
	// This becomes an IIFE:
	// func() Option_string {
	//     if user != nil {
	//         return Option_Some(user.Name)
	//     }
	//     return Option_None[string]()
	// }()

	// Create nil check condition
	nilCheck := &ast.BinaryExpr{
		X:  safeNav.X,
		Op: token.NEQ,
		Y:  ast.NewIdent("nil"),
	}

	// Create the selector expression for the field access
	fieldAccess := &ast.SelectorExpr{
		X:   safeNav.X,
		Sel: safeNav.Sel,
	}

	// Create Option_Some(fieldAccess)
	someCall := &ast.CallExpr{
		Fun: &ast.Ident{Name: "Option_Some"},
		Args: []ast.Expr{fieldAccess},
	}

	// Create Option_None[T]()
	noneCall := &ast.CallExpr{
		Fun: &ast.Ident{Name: "Option_None"},
		Args: []ast.Expr{},
	}

	// Create if statement
	ifStmt := &ast.IfStmt{
		Cond: nilCheck,
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{Results: []ast.Expr{someCall}},
			},
		},
	}

	returnNone := &ast.ReturnStmt{Results: []ast.Expr{noneCall}}

	// Create IIFE
	funcLit := &ast.FuncLit{
		Type: &ast.FuncType{
			Params: &ast.FieldList{},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{Type: ast.NewIdent("Option_T")}, // Placeholder, type inference needed
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{ifStmt, returnNone},
		},
	}

	// Return call expression
	return &ast.CallExpr{
		Fun:  funcLit,
		Args: []ast.Expr{},
	}, nil
}

// transformSmart generates smart unwrapping based on context
func (p *SafeNavigationPlugin) transformSmart(safeNav *dingoast.SafeNavigationExpr) (ast.Node, error) {
	// For smart mode, we unwrap to T with zero value fallback
	// Generate: user != nil ? user.Name : ""
	//
	// This becomes an IIFE returning T:
	// func() string {
	//     if user != nil {
	//         return user.Name
	//     }
	//     return "" // zero value
	// }()

	// Create nil check condition
	nilCheck := &ast.BinaryExpr{
		X:  safeNav.X,
		Op: token.NEQ,
		Y:  ast.NewIdent("nil"),
	}

	// Create the selector expression for the field access
	fieldAccess := &ast.SelectorExpr{
		X:   safeNav.X,
		Sel: safeNav.Sel,
	}

	// Create if statement returning the field
	ifStmt := &ast.IfStmt{
		Cond: nilCheck,
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{Results: []ast.Expr{fieldAccess}},
			},
		},
	}

	// Return zero value - for now use nil, proper zero value needs type inference
	returnZero := &ast.ReturnStmt{
		Results: []ast.Expr{ast.NewIdent("nil")},
	}

	// Create IIFE
	funcLit := &ast.FuncLit{
		Type: &ast.FuncType{
			Params: &ast.FieldList{},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{ifStmt, returnZero},
		},
	}

	// Return call expression
	return &ast.CallExpr{
		Fun:  funcLit,
		Args: []ast.Expr{},
	}, nil
}

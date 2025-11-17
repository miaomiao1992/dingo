// Package builtin provides built-in Dingo transformation plugins
package builtin

import (
	"go/ast"

	dingoast "github.com/MadAppGang/dingo/pkg/ast"
	"github.com/MadAppGang/dingo/pkg/config"
	"github.com/MadAppGang/dingo/pkg/plugin"
)

// TernaryPlugin transforms ternary expressions (? :) into Go if statements or IIFEs
//
// Features:
// - Standard precedence mode (default)
// - Explicit precedence mode requiring parentheses
// - Statement lifting for expression contexts
//
// Transforms (in statement context):
//   cond ? then : else  →  if cond { then } else { else }
//
// Transforms (in expression context):
//   cond ? then : else  →  func() T { if cond { return then } else { return else } }()
type TernaryPlugin struct {
	plugin.BasePlugin
}

// NewTernaryPlugin creates a new ternary operator plugin
func NewTernaryPlugin() *TernaryPlugin {
	return &TernaryPlugin{
		BasePlugin: *plugin.NewBasePlugin(
			"ternary",
			"Ternary operator (? :) with configurable precedence",
			nil,
		),
	}
}

// Transform transforms ternary expressions
func (p *TernaryPlugin) Transform(ctx *plugin.Context, node ast.Node) (ast.Node, error) {
	// Check if this is a ternary expression
	ternary, ok := node.(*dingoast.TernaryExpr)
	if !ok {
		return node, nil
	}

	// Get configuration for precedence mode
	var precedenceMode string
	if ctx.DingoConfig != nil {
		if cfg, ok := ctx.DingoConfig.(*config.Config); ok {
			precedenceMode = cfg.Features.OperatorPrecedence
		}
	}
	if precedenceMode == "" {
		precedenceMode = "standard" // Default
	}

	// TODO: Implement precedence validation when parser supports it.
	// In explicit mode, the parser should validate that complex expressions
	// mixing ?? and ? : use parentheses. The plugin currently doesn't have
	// enough context to perform this validation post-parse.
	// For now, precedence checking is deferred to parser implementation.
	_ = precedenceMode // Silence unused variable warning

	return p.transformToIIFE(ternary)
}

// transformToIIFE transforms ternary to immediately-invoked function expression
func (p *TernaryPlugin) transformToIIFE(ternary *dingoast.TernaryExpr) (ast.Node, error) {
	// Generate:
	// func() T {
	//     if cond {
	//         return then
	//     }
	//     return else
	// }()

	// Create if statement
	ifStmt := &ast.IfStmt{
		Cond: ternary.Cond,
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{Results: []ast.Expr{ternary.Then}},
			},
		},
	}

	returnElse := &ast.ReturnStmt{Results: []ast.Expr{ternary.Else}}

	// Create IIFE
	funcLit := &ast.FuncLit{
		Type: &ast.FuncType{
			Params: &ast.FieldList{},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{ifStmt, returnElse},
		},
	}

	// Return call expression
	return &ast.CallExpr{
		Fun:  funcLit,
		Args: []ast.Expr{},
	}, nil
}

// transformToIfStmt transforms ternary to if statement (for statement context)
func (p *TernaryPlugin) transformToIfStmt(ternary *dingoast.TernaryExpr) (ast.Node, error) {
	// Generate:
	// if cond {
	//     then
	// } else {
	//     else
	// }

	return &ast.IfStmt{
		Cond: ternary.Cond,
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ExprStmt{X: ternary.Then},
			},
		},
		Else: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ExprStmt{X: ternary.Else},
			},
		},
	}, nil
}

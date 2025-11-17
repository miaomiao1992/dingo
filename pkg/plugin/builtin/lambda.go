// Package builtin provides built-in Dingo transformation plugins
package builtin

import (
	"fmt"
	"go/ast"

	dingoast "github.com/MadAppGang/dingo/pkg/ast"
	"github.com/MadAppGang/dingo/pkg/config"
	"github.com/MadAppGang/dingo/pkg/plugin"
)

// LambdaPlugin transforms lambda expressions into Go function literals
//
// Features:
// - Rust-style syntax: |x, y| x + y
// - Arrow-style syntax: (x, y) => x + y
// - Configurable syntax acceptance
// - Expression and block bodies
// - Type inference for parameters and return type
//
// Transforms:
//   |x| x * 2  →  func(x int) int { return x * 2 }
//   (x) => x * 2  →  func(x int) int { return x * 2 }
type LambdaPlugin struct {
	plugin.BasePlugin
}

// NewLambdaPlugin creates a new lambda function plugin
func NewLambdaPlugin() *LambdaPlugin {
	return &LambdaPlugin{
		BasePlugin: *plugin.NewBasePlugin(
			"lambda",
			"Lambda functions with multiple syntax styles",
			nil,
		),
	}
}

// Transform transforms lambda expressions
func (p *LambdaPlugin) Transform(ctx *plugin.Context, node ast.Node) (ast.Node, error) {
	// Check if this is a lambda expression
	lambda, ok := node.(*dingoast.LambdaExpr)
	if !ok {
		return node, nil
	}

	// Get configuration for syntax mode
	var syntaxMode string
	if ctx.DingoConfig != nil {
		if cfg, ok := ctx.DingoConfig.(*config.Config); ok {
			syntaxMode = cfg.Features.LambdaSyntax
		}
	}
	if syntaxMode == "" {
		syntaxMode = "rust" // Default
	}

	// Validate syntax mode (parser should have already done this)
	switch syntaxMode {
	case "rust", "arrow", "both":
		// Valid
	default:
		return nil, fmt.Errorf("invalid lambda_syntax mode: %s", syntaxMode)
	}

	// Transform to Go function literal
	return p.transformToFuncLit(lambda)
}

// transformToFuncLit converts lambda to Go func literal
func (p *LambdaPlugin) transformToFuncLit(lambda *dingoast.LambdaExpr) (ast.Node, error) {
	// Create function type
	funcType := &ast.FuncType{
		Params: lambda.Params,
	}

	// If params is nil, create empty param list
	if funcType.Params == nil {
		funcType.Params = &ast.FieldList{
			List: []*ast.Field{},
		}
	}

	// Create function body
	// NOTE: Lambda.Body is typed as ast.Expr, so it can't be a BlockStmt
	// (BlockStmt doesn't implement Expr). For now, always wrap in return.
	// TODO: Fix Lambda AST to use ast.Node instead of ast.Expr for Body
	body := &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.ReturnStmt{
				Results: []ast.Expr{lambda.Body},
			},
		},
	}

	// Create function literal
	funcLit := &ast.FuncLit{
		Type: funcType,
		Body: body,
	}

	return funcLit, nil
}


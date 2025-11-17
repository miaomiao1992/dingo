// Package builtin provides built-in Dingo transformation plugins
package builtin

import (
	"go/ast"
	"go/token"
	"go/types"

	dingoast "github.com/MadAppGang/dingo/pkg/ast"
	"github.com/MadAppGang/dingo/pkg/config"
	"github.com/MadAppGang/dingo/pkg/plugin"
)

// NullCoalescingPlugin transforms null coalescing expressions (??) into Go nil checks
//
// Features:
// - Works with Option<T> (always)
// - Works with Go pointers *T (configurable)
// - Supports chaining: a ?? b ?? c
//
// Transforms (Option<T>):
//   opt ?? default  →  opt.IsSome() ? opt.Unwrap() : default
//
// Transforms (Go pointer, if enabled):
//   ptr ?? default  →  ptr != nil ? *ptr : default
type NullCoalescingPlugin struct {
	plugin.BasePlugin
}

// NewNullCoalescingPlugin creates a new null coalescing plugin
func NewNullCoalescingPlugin() *NullCoalescingPlugin {
	return &NullCoalescingPlugin{
		BasePlugin: *plugin.NewBasePlugin(
			"null_coalescing",
			"Null coalescing operator (??) for Option<T> and pointers",
			nil,
		),
	}
}

// Transform transforms null coalescing expressions
func (p *NullCoalescingPlugin) Transform(ctx *plugin.Context, node ast.Node) (ast.Node, error) {
	// Check if this is a null coalescing expression
	nc, ok := node.(*dingoast.NullCoalescingExpr)
	if !ok {
		return node, nil
	}

	// Get configuration
	var pointerSupport bool
	if ctx.DingoConfig != nil {
		if cfg, ok := ctx.DingoConfig.(*config.Config); ok {
			pointerSupport = cfg.Features.NullCoalescingPointers
		}
	}

	// Try to determine the type of the left operand
	leftType := p.inferType(ctx, nc.X)

	// Check if it's an Option type
	if p.isOptionType(leftType) {
		return p.transformOption(nc)
	}

	// Check if it's a pointer type and pointer support is enabled
	if pointerSupport && p.isPointerType(leftType) {
		return p.transformPointer(nc)
	}

	// If we can't determine type or it's neither Option nor pointer, generate generic IIFE
	// This will handle cases where type inference isn't available yet
	return p.transformGeneric(nc, pointerSupport)
}

// transformOption generates nil check for Option<T>
func (p *NullCoalescingPlugin) transformOption(nc *dingoast.NullCoalescingExpr) (ast.Node, error) {
	// Generate: opt.IsSome() ? opt.Unwrap() : default
	//
	// As IIFE:
	// func() T {
	//     if opt.IsSome() {
	//         return opt.Unwrap()
	//     }
	//     return default
	// }()

	// Create opt.IsSome() condition
	isSomeCall := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   nc.X,
			Sel: ast.NewIdent("IsSome"),
		},
	}

	// Create opt.Unwrap() call
	unwrapCall := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   nc.X,
			Sel: ast.NewIdent("Unwrap"),
		},
	}

	// Create if statement
	ifStmt := &ast.IfStmt{
		Cond: isSomeCall,
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{Results: []ast.Expr{unwrapCall}},
			},
		},
	}

	returnDefault := &ast.ReturnStmt{Results: []ast.Expr{nc.Y}}

	// Create IIFE
	funcLit := &ast.FuncLit{
		Type: &ast.FuncType{
			Params: &ast.FieldList{},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{ifStmt, returnDefault},
		},
	}

	return &ast.CallExpr{
		Fun:  funcLit,
		Args: []ast.Expr{},
	}, nil
}

// transformPointer generates nil check for Go pointer *T
func (p *NullCoalescingPlugin) transformPointer(nc *dingoast.NullCoalescingExpr) (ast.Node, error) {
	// Generate: ptr != nil ? *ptr : default
	//
	// As IIFE:
	// func() T {
	//     if ptr != nil {
	//         return *ptr
	//     }
	//     return default
	// }()

	// Create ptr != nil condition
	nilCheck := &ast.BinaryExpr{
		X:  nc.X,
		Op: token.NEQ,
		Y:  ast.NewIdent("nil"),
	}

	// Create *ptr dereference
	deref := &ast.StarExpr{X: nc.X}

	// Create if statement
	ifStmt := &ast.IfStmt{
		Cond: nilCheck,
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{Results: []ast.Expr{deref}},
			},
		},
	}

	returnDefault := &ast.ReturnStmt{Results: []ast.Expr{nc.Y}}

	// Create IIFE
	funcLit := &ast.FuncLit{
		Type: &ast.FuncType{
			Params: &ast.FieldList{},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{ifStmt, returnDefault},
		},
	}

	return &ast.CallExpr{
		Fun:  funcLit,
		Args: []ast.Expr{},
	}, nil
}

// transformGeneric generates a generic nil check that tries both approaches
func (p *NullCoalescingPlugin) transformGeneric(nc *dingoast.NullCoalescingExpr, pointerSupport bool) (ast.Node, error) {
	// For now, assume Option type by default
	// In a real implementation, we'd use type inference
	return p.transformOption(nc)
}

// Helper functions for type checking

func (p *NullCoalescingPlugin) inferType(ctx *plugin.Context, expr ast.Expr) types.Type {
	if ctx.TypeInfo != nil && ctx.TypeInfo.Types != nil {
		if tv, ok := ctx.TypeInfo.Types[expr]; ok {
			return tv.Type
		}
	}
	return nil
}

func (p *NullCoalescingPlugin) isOptionType(t types.Type) bool {
	if t == nil {
		return false
	}
	// Check if this is a named type with "Option_" prefix
	if named, ok := t.(*types.Named); ok {
		obj := named.Obj()
		if obj != nil {
			name := obj.Name()
			// Check for Option_T naming pattern (e.g., Option_string, Option_User)
			return len(name) > 7 && name[:7] == "Option_"
		}
	}
	return false
}

func (p *NullCoalescingPlugin) isPointerType(t types.Type) bool {
	if t == nil {
		return false
	}
	_, ok := t.(*types.Pointer)
	return ok
}

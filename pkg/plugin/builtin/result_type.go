// Package builtin provides built-in Dingo transformation plugins
package builtin

import (
	"go/ast"
	"go/token"

	dingoast "github.com/MadAppGang/dingo/pkg/ast"
	"github.com/MadAppGang/dingo/pkg/plugin"
)

// ResultTypePlugin provides the built-in Result<T, E> type
//
// Result is a generic sum type that represents either success (Ok) or failure (Err).
// It integrates with the error propagation operator (?) and pattern matching.
//
// Dingo syntax:
//   Result<T, E> with variants Ok(T) and Err(E)
//
// Transpiles to:
//   type Result_T_E struct { value *T; err *E; tag ResultTag }
//
// This plugin automatically registers Result as a built-in enum type.
type ResultTypePlugin struct {
	plugin.BasePlugin
	currentContext *plugin.Context
}

// NewResultTypePlugin creates a new Result type plugin
func NewResultTypePlugin() *ResultTypePlugin {
	return &ResultTypePlugin{
		BasePlugin: *plugin.NewBasePlugin(
			"result_type",
			"Built-in Result<T, E> generic type for error handling",
			nil, // No dependencies
		),
	}
}

// Name returns the plugin name
func (p *ResultTypePlugin) Name() string {
	return "result_type"
}

// Transform handles Result type transformations
// The actual enum definition is synthetic - we generate it on demand
func (p *ResultTypePlugin) Transform(ctx *plugin.Context, node ast.Node) (ast.Node, error) {
	p.currentContext = ctx

	// TODO: This plugin is currently foundation-only and does not perform transformations.
	// Future integration tasks:
	// 1. Detect Result<T, E> usage in source files
	// 2. Register Result as a synthetic enum with sum_types plugin
	// 3. Inject helper methods (IsOk, IsErr, Unwrap, UnwrapOr, etc.) into output
	// 4. Handle automatic conversion from (T, error) → Result<T, error>
	//
	// Current state: Result type definitions must be manually created as enums.
	// The sum_types plugin handles pattern matching and code generation.
	// The error_propagation plugin handles ? operator separately.

	return node, nil
}

// CreateResultEnum creates a synthetic Result<T, E> enum declaration
// This can be used by other plugins to understand Result's structure
func (p *ResultTypePlugin) CreateResultEnum() *dingoast.EnumDecl {
	return &dingoast.EnumDecl{
		Name: &ast.Ident{Name: "Result"},
		TypeParams: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{{Name: "T"}},
					Type:  &ast.Ident{Name: "any"},
				},
				{
					Names: []*ast.Ident{{Name: "E"}},
					Type:  &ast.Ident{Name: "any"},
				},
			},
		},
		Variants: []*dingoast.VariantDecl{
			{
				Name: &ast.Ident{Name: "Ok"},
				Kind: dingoast.VariantTuple,
				Fields: &ast.FieldList{
					List: []*ast.Field{
						{Type: &ast.Ident{Name: "T"}},
					},
				},
			},
			{
				Name: &ast.Ident{Name: "Err"},
				Kind: dingoast.VariantTuple,
				Fields: &ast.FieldList{
					List: []*ast.Field{
						{Type: &ast.Ident{Name: "E"}},
					},
				},
			},
		},
	}
}

// GenerateHelperMethods generates helper methods for Result<T, E>
// These are injected into files that use Result
func (p *ResultTypePlugin) GenerateHelperMethods(typeArgs []ast.Expr) []ast.Decl {
	decls := make([]ast.Decl, 0)

	// Type name for this specific Result instance (e.g., Result[T, E])
	// Always use IndexListExpr for generic types in Go 1.18+
	resultType := &ast.IndexListExpr{
		X:       &ast.Ident{Name: "Result"},
		Indices: typeArgs,
	}

	// Generate IsOk() method
	decls = append(decls, &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{{Name: "r"}},
					Type:  resultType,
				},
			},
		},
		Name: &ast.Ident{Name: "IsOk"},
		Type: &ast.FuncType{
			Params: &ast.FieldList{},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{Type: &ast.Ident{Name: "bool"}},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.BinaryExpr{
							X: &ast.SelectorExpr{
								X:   &ast.Ident{Name: "r"},
								Sel: &ast.Ident{Name: "tag"},
							},
							Op: token.EQL,
							Y:  &ast.Ident{Name: "ResultTag_Ok"},
						},
					},
				},
			},
		},
	})

	// Generate IsErr() method
	decls = append(decls, &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{{Name: "r"}},
					Type:  resultType,
				},
			},
		},
		Name: &ast.Ident{Name: "IsErr"},
		Type: &ast.FuncType{
			Params: &ast.FieldList{},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{Type: &ast.Ident{Name: "bool"}},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.BinaryExpr{
							X: &ast.SelectorExpr{
								X:   &ast.Ident{Name: "r"},
								Sel: &ast.Ident{Name: "tag"},
							},
							Op: token.EQL,
							Y:  &ast.Ident{Name: "ResultTag_Err"},
						},
					},
				},
			},
		},
	})

	// Generate Unwrap() method
	// func (r Result[T,E]) Unwrap() T {
	//     if r.IsErr() { panic("called Unwrap on Err") }
	//     return *r.ok_0
	// }
	decls = append(decls, &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{{Name: "r"}},
					Type:  resultType,
				},
			},
		},
		Name: &ast.Ident{Name: "Unwrap"},
		Type: &ast.FuncType{
			Params: &ast.FieldList{},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{Type: typeArgs[0]}, // T
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.IfStmt{
					Cond: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "r"},
							Sel: &ast.Ident{Name: "IsErr"},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ExprStmt{
								X: &ast.CallExpr{
									Fun: &ast.Ident{Name: "panic"},
									Args: []ast.Expr{
										&ast.BasicLit{
											Kind:  token.STRING,
											Value: `"dingo: called Result.Unwrap() on Err value"`,
										},
									},
								},
							},
						},
					},
				},
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.StarExpr{
							X: &ast.SelectorExpr{
								X:   &ast.Ident{Name: "r"},
								Sel: &ast.Ident{Name: "ok_0"},
							},
						},
					},
				},
			},
		},
	})

	// Generate UnwrapOr(default T) method
	decls = append(decls, &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{{Name: "r"}},
					Type:  resultType,
				},
			},
		},
		Name: &ast.Ident{Name: "UnwrapOr"},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{{Name: "defaultVal"}},
						Type:  typeArgs[0], // T
					},
				},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{Type: typeArgs[0]}, // T
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.IfStmt{
					Cond: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "r"},
							Sel: &ast.Ident{Name: "IsOk"},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ReturnStmt{
								Results: []ast.Expr{
									&ast.StarExpr{
										X: &ast.SelectorExpr{
											X:   &ast.Ident{Name: "r"},
											Sel: &ast.Ident{Name: "ok_0"},
										},
									},
								},
							},
						},
					},
				},
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.Ident{Name: "defaultVal"},
					},
				},
			},
		},
	})

	return decls
}

// Package builtin provides built-in Dingo transformation plugins
package builtin

import (
	"go/ast"
	"go/token"

	dingoast "github.com/MadAppGang/dingo/pkg/ast"
	"github.com/MadAppGang/dingo/pkg/plugin"
)

// OptionTypePlugin provides the built-in Option<T> type
//
// Option is a generic sum type that represents a value that may or may not be present.
// It eliminates nil pointer bugs by making nullability explicit in the type system.
//
// Dingo syntax:
//   Option<T> with variants Some(T) and None
//
// Transpiles to:
//   type Option_T struct { value *T; tag OptionTag }
//
// This plugin automatically registers Option as a built-in enum type.
type OptionTypePlugin struct {
	plugin.BasePlugin
	currentContext *plugin.Context
}

// NewOptionTypePlugin creates a new Option type plugin
func NewOptionTypePlugin() *OptionTypePlugin {
	return &OptionTypePlugin{
		BasePlugin: *plugin.NewBasePlugin(
			"option_type",
			"Built-in Option<T> generic type for null safety",
			nil, // No dependencies
		),
	}
}

// Name returns the plugin name
func (p *OptionTypePlugin) Name() string {
	return "option_type"
}

// Transform handles Option type transformations
// The actual enum definition is synthetic - we generate it on demand
func (p *OptionTypePlugin) Transform(ctx *plugin.Context, node ast.Node) (ast.Node, error) {
	p.currentContext = ctx

	// TODO: This plugin is currently foundation-only and does not perform transformations.
	// Future integration tasks:
	// 1. Detect Option<T> usage in source files
	// 2. Register Option as a synthetic enum with sum_types plugin
	// 3. Inject helper methods (IsSome, IsNone, Unwrap, UnwrapOr, etc.) into output
	// 4. Handle automatic conversion from nullable types → Option<T>
	//
	// Current state: Option type definitions must be manually created as enums.
	// The sum_types plugin handles pattern matching and code generation.
	// The safe_navigation (?.) and null_coalescing (??) plugins handle related operators.

	return node, nil
}

// CreateOptionEnum creates a synthetic Option<T> enum declaration
// This can be used by other plugins to understand Option's structure
func (p *OptionTypePlugin) CreateOptionEnum() *dingoast.EnumDecl {
	return &dingoast.EnumDecl{
		Name: &ast.Ident{Name: "Option"},
		TypeParams: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{{Name: "T"}},
					Type:  &ast.Ident{Name: "any"},
				},
			},
		},
		Variants: []*dingoast.VariantDecl{
			{
				Name: &ast.Ident{Name: "Some"},
				Kind: dingoast.VariantTuple,
				Fields: &ast.FieldList{
					List: []*ast.Field{
						{Type: &ast.Ident{Name: "T"}},
					},
				},
			},
			{
				Name: &ast.Ident{Name: "None"},
				Kind: dingoast.VariantUnit,
			},
		},
	}
}

// GenerateHelperMethods generates helper methods for Option<T>
// These are injected into files that use Option
func (p *OptionTypePlugin) GenerateHelperMethods(typeArgs []ast.Expr) []ast.Decl {
	decls := make([]ast.Decl, 0)

	// Type name for this specific Option instance (e.g., Option[T])
	// Always use IndexExpr for generic types with single type parameter in Go 1.18+
	optionType := &ast.IndexExpr{
		X:     &ast.Ident{Name: "Option"},
		Index: typeArgs[0],
	}

	// Generate IsSome() method
	decls = append(decls, &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{{Name: "o"}},
					Type:  optionType,
				},
			},
		},
		Name: &ast.Ident{Name: "IsSome"},
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
								X:   &ast.Ident{Name: "o"},
								Sel: &ast.Ident{Name: "tag"},
							},
							Op: token.EQL,
							Y:  &ast.Ident{Name: "OptionTag_Some"},
						},
					},
				},
			},
		},
	})

	// Generate IsNone() method
	decls = append(decls, &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{{Name: "o"}},
					Type:  optionType,
				},
			},
		},
		Name: &ast.Ident{Name: "IsNone"},
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
								X:   &ast.Ident{Name: "o"},
								Sel: &ast.Ident{Name: "tag"},
							},
							Op: token.EQL,
							Y:  &ast.Ident{Name: "OptionTag_None"},
						},
					},
				},
			},
		},
	})

	// Generate Unwrap() method
	// func (o Option[T]) Unwrap() T {
	//     if o.IsNone() { panic("called Unwrap on None") }
	//     return *o.some_0
	// }
	decls = append(decls, &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{{Name: "o"}},
					Type:  optionType,
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
							X:   &ast.Ident{Name: "o"},
							Sel: &ast.Ident{Name: "IsNone"},
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
											Value: `"dingo: called Option.Unwrap() on None value"`,
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
								X:   &ast.Ident{Name: "o"},
								Sel: &ast.Ident{Name: "some_0"},
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
					Names: []*ast.Ident{{Name: "o"}},
					Type:  optionType,
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
							X:   &ast.Ident{Name: "o"},
							Sel: &ast.Ident{Name: "IsSome"},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ReturnStmt{
								Results: []ast.Expr{
									&ast.StarExpr{
										X: &ast.SelectorExpr{
											X:   &ast.Ident{Name: "o"},
											Sel: &ast.Ident{Name: "some_0"},
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

	// Generate Map<U>(f func(T) U) Option<U> method
	// func (o Option[T]) Map(f func(T) U) Option[U] {
	//     if o.IsNone() { return Option_None[U]() }
	//     return Option_Some(*o.some_0)
	// }
	// Note: This is a simplified version - full implementation would need type parameter handling
	decls = append(decls, &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{{Name: "o"}},
					Type:  optionType,
				},
			},
		},
		Name: &ast.Ident{Name: "Map"},
		Type: &ast.FuncType{
			TypeParams: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{{Name: "U"}},
						Type:  &ast.Ident{Name: "any"},
					},
				},
			},
			Params: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{{Name: "f"}},
						Type: &ast.FuncType{
							Params: &ast.FieldList{
								List: []*ast.Field{
									{Type: typeArgs[0]}, // T
								},
							},
							Results: &ast.FieldList{
								List: []*ast.Field{
									{Type: &ast.Ident{Name: "U"}},
								},
							},
						},
					},
				},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: &ast.IndexExpr{
							X:     &ast.Ident{Name: "Option"},
							Index: &ast.Ident{Name: "U"},
						},
					},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.IfStmt{
					Cond: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "o"},
							Sel: &ast.Ident{Name: "IsNone"},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ReturnStmt{
								Results: []ast.Expr{
									&ast.CallExpr{
										Fun: &ast.IndexExpr{
											X:     &ast.Ident{Name: "Option_None"},
											Index: &ast.Ident{Name: "U"},
										},
									},
								},
							},
						},
					},
				},
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.Ident{Name: "Option_Some"},
							Args: []ast.Expr{
								&ast.CallExpr{
									Fun: &ast.Ident{Name: "f"},
									Args: []ast.Expr{
										&ast.StarExpr{
											X: &ast.SelectorExpr{
												X:   &ast.Ident{Name: "o"},
												Sel: &ast.Ident{Name: "some_0"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})

	return decls
}

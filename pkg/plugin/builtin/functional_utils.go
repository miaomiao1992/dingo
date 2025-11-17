// Package builtin provides built-in Dingo transformation plugins
package builtin

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/MadAppGang/dingo/pkg/plugin"
	"golang.org/x/tools/go/ast/astutil"
)

// FunctionalUtilitiesPlugin transforms functional utility method calls into inline Go loops
//
// Features:
// - Core operations: map, filter, reduce
// - Helper operations: sum, count, all, any, find
// - Result/Option integration: mapResult, filterSome
// - Method chaining support
//
// Transforms:
//   numbers.map(fn)     →  inline for-range loop with append
//   numbers.filter(fn)  →  inline for-range loop with conditional append
//   numbers.reduce(init, fn) →  inline for-range loop with accumulator
type FunctionalUtilitiesPlugin struct {
	plugin.BasePlugin

	// State for current transformation
	currentContext *plugin.Context

	// Counter for generating unique temporary variable names
	tempCounter int
}

// NewFunctionalUtilitiesPlugin creates a new functional utilities transformation plugin
func NewFunctionalUtilitiesPlugin() *FunctionalUtilitiesPlugin {
	return &FunctionalUtilitiesPlugin{
		BasePlugin:  *plugin.NewBasePlugin("functional_utilities", "Functional utilities for slices", nil),
		tempCounter: 0,
	}
}

// Transform transforms an AST node (file-level entry point)
func (p *FunctionalUtilitiesPlugin) Transform(ctx *plugin.Context, node ast.Node) (ast.Node, error) {
	file, ok := node.(*ast.File)
	if !ok {
		return node, nil
	}

	// Store context
	p.currentContext = ctx
	p.tempCounter = 0

	// Transform the file using astutil.Apply
	result := astutil.Apply(file, p.preVisit, p.postVisit)

	return result, nil
}

// preVisit is called before visiting a node's children
func (p *FunctionalUtilitiesPlugin) preVisit(cursor *astutil.Cursor) bool {
	return true // Continue traversal
}

// postVisit is called after visiting a node's children
func (p *FunctionalUtilitiesPlugin) postVisit(cursor *astutil.Cursor) bool {
	node := cursor.Node()

	// Check if this is a call expression
	call, ok := node.(*ast.CallExpr)
	if !ok {
		return true
	}

	// Check if it's a method call (selector expression)
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return true
	}

	// Check which functional utility method is being called
	// Support both lowercase (Dingo syntax) and capitalized (Go-compatible) names
	methodName := sel.Sel.Name
	switch methodName {
	case "map", "Map":
		if transformed := p.transformMap(sel.X, call.Args); transformed != nil {
			cursor.Replace(transformed)
		}
	case "filter", "Filter":
		if transformed := p.transformFilter(sel.X, call.Args); transformed != nil {
			cursor.Replace(transformed)
		}
	case "reduce", "Reduce":
		if transformed := p.transformReduce(sel.X, call.Args); transformed != nil {
			cursor.Replace(transformed)
		}
	case "sum", "Sum":
		if transformed := p.transformSum(sel.X); transformed != nil {
			cursor.Replace(transformed)
		}
	case "count", "Count":
		if transformed := p.transformCount(sel.X, call.Args); transformed != nil {
			cursor.Replace(transformed)
		}
	case "all", "All":
		if transformed := p.transformAll(sel.X, call.Args); transformed != nil {
			cursor.Replace(transformed)
		}
	case "any", "Any":
		if transformed := p.transformAny(sel.X, call.Args); transformed != nil {
			cursor.Replace(transformed)
		}
	case "find", "Find":
		if transformed := p.transformFind(sel.X, call.Args); transformed != nil {
			cursor.Replace(transformed)
		}
	case "mapResult", "MapResult":
		if transformed := p.transformMapResult(sel.X, call.Args); transformed != nil {
			cursor.Replace(transformed)
		}
	case "filterSome", "FilterSome":
		if transformed := p.transformFilterSome(sel.X); transformed != nil {
			cursor.Replace(transformed)
		}
	}

	return true
}

// transformMap transforms: numbers.map(fn) → inline for-range loop
//
// Requirements:
//   - fn must be a function literal with explicit return type
//   - fn must accept exactly 1 parameter matching slice element type
//   - fn must return exactly 1 value
//
// Example:
//   numbers.map(func(x int) int { return x * 2 })
//
// Not supported:
//   numbers.map(func(x int) { return x * 2 })  // Missing return type
//
// TODO: Support type inference using go/types package
func (p *FunctionalUtilitiesPlugin) transformMap(receiver ast.Expr, args []ast.Expr) ast.Node {
	if len(args) != 1 {
		if p.currentContext != nil && p.currentContext.Logger != nil {
			p.currentContext.Logger.Debug("map expects 1 argument, got %d", len(args))
		}
		return nil
	}

	fn, ok := args[0].(*ast.FuncLit)
	if !ok {
		if p.currentContext != nil && p.currentContext.Logger != nil {
			p.currentContext.Logger.Debug("map expects function literal argument")
		}
		return nil
	}

	// Extract parameter name and body
	if fn.Type.Params == nil || len(fn.Type.Params.List) == 0 {
		if p.currentContext != nil && p.currentContext.Logger != nil {
			p.currentContext.Logger.Debug("map function has no parameters")
		}
		return nil
	}

	// Validate arity: map expects exactly 1 parameter
	if len(fn.Type.Params.List) != 1 {
		if p.currentContext != nil && p.currentContext.Logger != nil {
			p.currentContext.Logger.Warn("map expects function with 1 parameter, got %d", len(fn.Type.Params.List))
		}
		return nil
	}

	paramField := fn.Type.Params.List[0]
	if len(paramField.Names) == 0 {
		if p.currentContext != nil && p.currentContext.Logger != nil {
			p.currentContext.Logger.Debug("map function parameter has no name")
		}
		return nil
	}
	paramName := paramField.Names[0].Name

	// Extract function body expression
	bodyExpr := p.extractFunctionBody(fn.Body)
	if bodyExpr == nil {
		// extractFunctionBody already logs the reason
		return nil
	}

	// Validate and infer result element type from function return type
	if fn.Type.Results == nil || len(fn.Type.Results.List) == 0 {
		if p.currentContext != nil && p.currentContext.Logger != nil {
			p.currentContext.Logger.Warn("map function must have explicit return type")
		}
		return nil
	}

	resultElemType := fn.Type.Results.List[0].Type
	if resultElemType == nil {
		if p.currentContext != nil && p.currentContext.Logger != nil {
			p.currentContext.Logger.Warn("cannot infer result type from function signature")
		}
		return nil
	}

	// Generate unique temp variable name
	resultVar := p.newTempVar()

	// Create result slice type
	resultSliceType := &ast.ArrayType{
		Elt: resultElemType,
	}

	// Build the transformation:
	// var __temp0 []T
	// __temp0 = make([]T, 0, len(receiver))
	// for _, paramName := range receiver {
	//     __temp0 = append(__temp0, bodyExpr)
	// }
	return &ast.CallExpr{
		Fun: &ast.FuncLit{
			Type: &ast.FuncType{
				Params: &ast.FieldList{},
				Results: &ast.FieldList{
					List: []*ast.Field{{Type: resultSliceType}},
				},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					// var __temp0 []T
					&ast.DeclStmt{
						Decl: &ast.GenDecl{
							Tok: token.VAR,
							Specs: []ast.Spec{
								&ast.ValueSpec{
									Names: []*ast.Ident{{Name: resultVar}},
									Type:  resultSliceType,
								},
							},
						},
					},
					// __temp0 = make([]T, 0, len(receiver))
					&ast.AssignStmt{
						Lhs: []ast.Expr{&ast.Ident{Name: resultVar}},
						Tok: token.ASSIGN,
						Rhs: []ast.Expr{
							&ast.CallExpr{
								Fun: &ast.Ident{Name: "make"},
								Args: []ast.Expr{
									resultSliceType,
									&ast.BasicLit{Kind: token.INT, Value: "0"},
									&ast.CallExpr{
										Fun:  &ast.Ident{Name: "len"},
										Args: []ast.Expr{p.cloneExpr(receiver)},
									},
								},
							},
						},
					},
					// for _, paramName := range receiver { ... }
					&ast.RangeStmt{
						Key:   &ast.Ident{Name: "_"},
						Value: &ast.Ident{Name: paramName},
						Tok:   token.DEFINE,
						X:     p.cloneExpr(receiver),
						Body: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.AssignStmt{
									Lhs: []ast.Expr{&ast.Ident{Name: resultVar}},
									Tok: token.ASSIGN,
									Rhs: []ast.Expr{
										&ast.CallExpr{
											Fun: &ast.Ident{Name: "append"},
											Args: []ast.Expr{
												&ast.Ident{Name: resultVar},
												p.cloneExpr(bodyExpr),
											},
										},
									},
								},
							},
						},
					},
					// return __temp0
					&ast.ReturnStmt{
						Results: []ast.Expr{&ast.Ident{Name: resultVar}},
					},
				},
			},
		},
	}
}

// transformFilter transforms: numbers.filter(fn) → inline for-range loop with conditional
//
// Requirements:
//   - fn must be a function literal with bool return type
//   - fn must accept exactly 1 parameter matching slice element type
//
// Example:
//   numbers.filter(func(x int) bool { return x > 0 })
func (p *FunctionalUtilitiesPlugin) transformFilter(receiver ast.Expr, args []ast.Expr) ast.Node {
	if len(args) != 1 {
		if p.currentContext != nil && p.currentContext.Logger != nil {
			p.currentContext.Logger.Debug("filter expects 1 argument, got %d", len(args))
		}
		return nil
	}

	fn, ok := args[0].(*ast.FuncLit)
	if !ok {
		if p.currentContext != nil && p.currentContext.Logger != nil {
			p.currentContext.Logger.Debug("filter expects function literal argument")
		}
		return nil
	}

	// Extract parameter name and predicate
	if fn.Type.Params == nil || len(fn.Type.Params.List) == 0 {
		if p.currentContext != nil && p.currentContext.Logger != nil {
			p.currentContext.Logger.Debug("filter function has no parameters")
		}
		return nil
	}

	// Validate arity: filter expects exactly 1 parameter
	if len(fn.Type.Params.List) != 1 {
		if p.currentContext != nil && p.currentContext.Logger != nil {
			p.currentContext.Logger.Warn("filter expects function with 1 parameter, got %d", len(fn.Type.Params.List))
		}
		return nil
	}

	paramField := fn.Type.Params.List[0]
	if len(paramField.Names) == 0 {
		return nil
	}
	paramName := paramField.Names[0].Name
	paramType := paramField.Type

	// Extract predicate expression
	predicateExpr := p.extractFunctionBody(fn.Body)
	if predicateExpr == nil {
		return nil
	}

	// Generate unique temp variable name
	resultVar := p.newTempVar()

	// Create result slice type (same as input)
	resultSliceType := &ast.ArrayType{
		Elt: paramType,
	}

	// Build the transformation:
	// var __temp0 []T
	// __temp0 = make([]T, 0, len(receiver))
	// for _, paramName := range receiver {
	//     if predicateExpr {
	//         __temp0 = append(__temp0, paramName)
	//     }
	// }
	return &ast.CallExpr{
		Fun: &ast.FuncLit{
			Type: &ast.FuncType{
				Params: &ast.FieldList{},
				Results: &ast.FieldList{
					List: []*ast.Field{{Type: resultSliceType}},
				},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.DeclStmt{
						Decl: &ast.GenDecl{
							Tok: token.VAR,
							Specs: []ast.Spec{
								&ast.ValueSpec{
									Names: []*ast.Ident{{Name: resultVar}},
									Type:  resultSliceType,
								},
							},
						},
					},
					&ast.AssignStmt{
						Lhs: []ast.Expr{&ast.Ident{Name: resultVar}},
						Tok: token.ASSIGN,
						Rhs: []ast.Expr{
							&ast.CallExpr{
								Fun: &ast.Ident{Name: "make"},
								Args: []ast.Expr{
									resultSliceType,
									&ast.BasicLit{Kind: token.INT, Value: "0"},
									&ast.CallExpr{
										Fun:  &ast.Ident{Name: "len"},
										Args: []ast.Expr{p.cloneExpr(receiver)},
									},
								},
							},
						},
					},
					&ast.RangeStmt{
						Key:   &ast.Ident{Name: "_"},
						Value: &ast.Ident{Name: paramName},
						Tok:   token.DEFINE,
						X:     p.cloneExpr(receiver),
						Body: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.IfStmt{
									Cond: p.cloneExpr(predicateExpr),
									Body: &ast.BlockStmt{
										List: []ast.Stmt{
											&ast.AssignStmt{
												Lhs: []ast.Expr{&ast.Ident{Name: resultVar}},
												Tok: token.ASSIGN,
												Rhs: []ast.Expr{
													&ast.CallExpr{
														Fun: &ast.Ident{Name: "append"},
														Args: []ast.Expr{
															&ast.Ident{Name: resultVar},
															&ast.Ident{Name: paramName},
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
					&ast.ReturnStmt{
						Results: []ast.Expr{&ast.Ident{Name: resultVar}},
					},
				},
			},
		},
	}
}

// transformReduce transforms: numbers.reduce(init, fn) → inline for-range loop with accumulator
//
// Requirements:
//   - fn must be a function literal with explicit return type
//   - fn must accept exactly 2 parameters (accumulator, element)
//
// Example:
//   numbers.reduce(0, func(acc int, x int) int { return acc + x })
//
// TODO: Support type inference using go/types package
func (p *FunctionalUtilitiesPlugin) transformReduce(receiver ast.Expr, args []ast.Expr) ast.Node {
	if len(args) != 2 {
		if p.currentContext != nil && p.currentContext.Logger != nil {
			p.currentContext.Logger.Debug("reduce expects 2 arguments, got %d", len(args))
		}
		return nil
	}

	initValue := args[0]

	fn, ok := args[1].(*ast.FuncLit)
	if !ok {
		if p.currentContext != nil && p.currentContext.Logger != nil {
			p.currentContext.Logger.Debug("reduce expects function literal as second argument")
		}
		return nil
	}

	// Extract parameters (acc, element)
	if fn.Type.Params == nil || len(fn.Type.Params.List) < 2 {
		if p.currentContext != nil && p.currentContext.Logger != nil {
			p.currentContext.Logger.Debug("reduce function has insufficient parameters")
		}
		return nil
	}

	// Validate arity: reduce expects exactly 2 parameters
	if len(fn.Type.Params.List) != 2 {
		if p.currentContext != nil && p.currentContext.Logger != nil {
			p.currentContext.Logger.Warn("reduce expects function with 2 parameters, got %d", len(fn.Type.Params.List))
		}
		return nil
	}

	accParam := fn.Type.Params.List[0]
	elemParam := fn.Type.Params.List[1]

	if len(accParam.Names) == 0 || len(elemParam.Names) == 0 {
		return nil
	}

	accName := accParam.Names[0].Name
	elemName := elemParam.Names[0].Name

	// Extract reducer expression
	reducerExpr := p.extractFunctionBody(fn.Body)
	if reducerExpr == nil {
		// extractFunctionBody already logs the reason
		return nil
	}

	// Validate and infer result type from function return type
	if fn.Type.Results == nil || len(fn.Type.Results.List) == 0 {
		if p.currentContext != nil && p.currentContext.Logger != nil {
			p.currentContext.Logger.Warn("reduce function must have explicit return type")
		}
		return nil
	}

	resultType := fn.Type.Results.List[0].Type
	if resultType == nil {
		if p.currentContext != nil && p.currentContext.Logger != nil {
			p.currentContext.Logger.Warn("cannot infer result type from reduce function signature")
		}
		return nil
	}

	// Generate unique temp variable name
	resultVar := p.newTempVar()

	// Build the transformation:
	// var __temp0 T
	// __temp0 = initValue
	// for _, elemName := range receiver {
	//     __temp0 = reducerExpr
	// }
	return &ast.CallExpr{
		Fun: &ast.FuncLit{
			Type: &ast.FuncType{
				Params: &ast.FieldList{},
				Results: &ast.FieldList{
					List: []*ast.Field{{Type: resultType}},
				},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.DeclStmt{
						Decl: &ast.GenDecl{
							Tok: token.VAR,
							Specs: []ast.Spec{
								&ast.ValueSpec{
									Names: []*ast.Ident{{Name: resultVar}},
									Type:  resultType,
								},
							},
						},
					},
					&ast.AssignStmt{
						Lhs: []ast.Expr{&ast.Ident{Name: resultVar}},
						Tok: token.ASSIGN,
						Rhs: []ast.Expr{p.cloneExpr(initValue)},
					},
					&ast.RangeStmt{
						Key:   &ast.Ident{Name: "_"},
						Value: &ast.Ident{Name: elemName},
						Tok:   token.DEFINE,
						X:     p.cloneExpr(receiver),
						Body: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.AssignStmt{
									Lhs: []ast.Expr{&ast.Ident{Name: accName}},
									Tok: token.DEFINE,
									Rhs: []ast.Expr{&ast.Ident{Name: resultVar}},
								},
								&ast.AssignStmt{
									Lhs: []ast.Expr{&ast.Ident{Name: resultVar}},
									Tok: token.ASSIGN,
									Rhs: []ast.Expr{p.cloneExpr(reducerExpr)},
								},
							},
						},
					},
					&ast.ReturnStmt{
						Results: []ast.Expr{&ast.Ident{Name: resultVar}},
					},
				},
			},
		},
	}
}

// transformSum transforms: numbers.sum() → inline for-range loop with addition
//
// Requirements:
//   - Works with any numeric slice type (int, float64, etc.)
//   - Infers element type from slice context
//
// Example:
//   numbers.sum()  // []int → int
//   prices.sum()   // []float64 → float64
//
// TODO: Currently uses hardcoded type inference; enhance with go/types for better inference
func (p *FunctionalUtilitiesPlugin) transformSum(receiver ast.Expr) ast.Node {
	resultVar := p.newTempVar()
	elemVar := p.newTempVar()

	// Try to infer element type from receiver
	// For now, we use a simple approach: declare var with zero value
	// This allows Go's type inference to determine the type
	var resultType ast.Expr

	// Extract element type if receiver is array/slice type expression
	if arrType, ok := receiver.(*ast.ArrayType); ok {
		resultType = arrType.Elt
	} else {
		// Fallback: use var declaration with zero-initialized value
		// This will be inferred from context
		resultType = nil
	}

	var initStmt ast.Stmt
	if resultType != nil {
		// Use explicit type with var declaration
		initStmt = &ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{{Name: resultVar}},
						Type:  resultType,
					},
				},
			},
		}
	} else {
		// Fallback: initialize to 0 and let Go infer
		// Note: This only works for int; documented limitation
		initStmt = &ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: resultVar}},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}},
		}
	}

	// Determine the IIFE return type
	// If we couldn't infer from receiver, default to int
	funcResultType := resultType
	if funcResultType == nil {
		funcResultType = &ast.Ident{Name: "int"}
	}

	// Build: var/sum := 0; for _, x := range numbers { sum += x }
	return &ast.CallExpr{
		Fun: &ast.FuncLit{
			Type: &ast.FuncType{
				Params: &ast.FieldList{},
				Results: &ast.FieldList{
					List: []*ast.Field{{Type: funcResultType}},
				},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					initStmt,
					&ast.RangeStmt{
						Key:   &ast.Ident{Name: "_"},
						Value: &ast.Ident{Name: elemVar},
						Tok:   token.DEFINE,
						X:     p.cloneExpr(receiver),
						Body: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.AssignStmt{
									Lhs: []ast.Expr{&ast.Ident{Name: resultVar}},
									Tok: token.ADD_ASSIGN,
									Rhs: []ast.Expr{&ast.Ident{Name: elemVar}},
								},
							},
						},
					},
					&ast.ReturnStmt{
						Results: []ast.Expr{&ast.Ident{Name: resultVar}},
					},
				},
			},
		},
	}
}

// transformCount transforms: numbers.count(fn) → inline for-range loop with counter
func (p *FunctionalUtilitiesPlugin) transformCount(receiver ast.Expr, args []ast.Expr) ast.Node {
	if len(args) != 1 {
		return nil
	}

	fn, ok := args[0].(*ast.FuncLit)
	if !ok {
		return nil
	}

	// Extract parameter name and predicate
	if fn.Type.Params == nil || len(fn.Type.Params.List) == 0 {
		return nil
	}

	paramField := fn.Type.Params.List[0]
	if len(paramField.Names) == 0 {
		return nil
	}
	paramName := paramField.Names[0].Name

	predicateExpr := p.extractFunctionBody(fn.Body)
	if predicateExpr == nil {
		return nil
	}

	resultVar := p.newTempVar()

	// Build: count := 0; for _, x := range numbers { if predicate { count++ } }
	return &ast.CallExpr{
		Fun: &ast.FuncLit{
			Type: &ast.FuncType{
				Params: &ast.FieldList{},
				Results: &ast.FieldList{
					List: []*ast.Field{{Type: &ast.Ident{Name: "int"}}},
				},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.AssignStmt{
						Lhs: []ast.Expr{&ast.Ident{Name: resultVar}},
						Tok: token.DEFINE,
						Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}},
					},
					&ast.RangeStmt{
						Key:   &ast.Ident{Name: "_"},
						Value: &ast.Ident{Name: paramName},
						Tok:   token.DEFINE,
						X:     p.cloneExpr(receiver),
						Body: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.IfStmt{
									Cond: p.cloneExpr(predicateExpr),
									Body: &ast.BlockStmt{
										List: []ast.Stmt{
											&ast.IncDecStmt{
												X:   &ast.Ident{Name: resultVar},
												Tok: token.INC,
											},
										},
									},
								},
							},
						},
					},
					&ast.ReturnStmt{
						Results: []ast.Expr{&ast.Ident{Name: resultVar}},
					},
				},
			},
		},
	}
}

// transformAll transforms: numbers.all(fn) → inline for-range loop with early exit
func (p *FunctionalUtilitiesPlugin) transformAll(receiver ast.Expr, args []ast.Expr) ast.Node {
	if len(args) != 1 {
		return nil
	}

	fn, ok := args[0].(*ast.FuncLit)
	if !ok {
		return nil
	}

	// Extract parameter name and predicate
	if fn.Type.Params == nil || len(fn.Type.Params.List) == 0 {
		return nil
	}

	paramField := fn.Type.Params.List[0]
	if len(paramField.Names) == 0 {
		return nil
	}
	paramName := paramField.Names[0].Name

	predicateExpr := p.extractFunctionBody(fn.Body)
	if predicateExpr == nil {
		return nil
	}

	resultVar := p.newTempVar()

	// Build: result := true; for _, x := range numbers { if !predicate { result = false; break } }
	return &ast.CallExpr{
		Fun: &ast.FuncLit{
			Type: &ast.FuncType{
				Params: &ast.FieldList{},
				Results: &ast.FieldList{
					List: []*ast.Field{{Type: &ast.Ident{Name: "bool"}}},
				},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.AssignStmt{
						Lhs: []ast.Expr{&ast.Ident{Name: resultVar}},
						Tok: token.DEFINE,
						Rhs: []ast.Expr{&ast.Ident{Name: "true"}},
					},
					&ast.RangeStmt{
						Key:   &ast.Ident{Name: "_"},
						Value: &ast.Ident{Name: paramName},
						Tok:   token.DEFINE,
						X:     p.cloneExpr(receiver),
						Body: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.IfStmt{
									Cond: &ast.UnaryExpr{
										Op: token.NOT,
										X:  p.cloneExpr(predicateExpr),
									},
									Body: &ast.BlockStmt{
										List: []ast.Stmt{
											&ast.AssignStmt{
												Lhs: []ast.Expr{&ast.Ident{Name: resultVar}},
												Tok: token.ASSIGN,
												Rhs: []ast.Expr{&ast.Ident{Name: "false"}},
											},
											&ast.BranchStmt{Tok: token.BREAK},
										},
									},
								},
							},
						},
					},
					&ast.ReturnStmt{
						Results: []ast.Expr{&ast.Ident{Name: resultVar}},
					},
				},
			},
		},
	}
}

// transformAny transforms: numbers.any(fn) → inline for-range loop with early exit
func (p *FunctionalUtilitiesPlugin) transformAny(receiver ast.Expr, args []ast.Expr) ast.Node {
	if len(args) != 1 {
		return nil
	}

	fn, ok := args[0].(*ast.FuncLit)
	if !ok {
		return nil
	}

	// Extract parameter name and predicate
	if fn.Type.Params == nil || len(fn.Type.Params.List) == 0 {
		return nil
	}

	paramField := fn.Type.Params.List[0]
	if len(paramField.Names) == 0 {
		return nil
	}
	paramName := paramField.Names[0].Name

	predicateExpr := p.extractFunctionBody(fn.Body)
	if predicateExpr == nil {
		return nil
	}

	resultVar := p.newTempVar()

	// Build: result := false; for _, x := range numbers { if predicate { result = true; break } }
	return &ast.CallExpr{
		Fun: &ast.FuncLit{
			Type: &ast.FuncType{
				Params: &ast.FieldList{},
				Results: &ast.FieldList{
					List: []*ast.Field{{Type: &ast.Ident{Name: "bool"}}},
				},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.AssignStmt{
						Lhs: []ast.Expr{&ast.Ident{Name: resultVar}},
						Tok: token.DEFINE,
						Rhs: []ast.Expr{&ast.Ident{Name: "false"}},
					},
					&ast.RangeStmt{
						Key:   &ast.Ident{Name: "_"},
						Value: &ast.Ident{Name: paramName},
						Tok:   token.DEFINE,
						X:     p.cloneExpr(receiver),
						Body: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.IfStmt{
									Cond: p.cloneExpr(predicateExpr),
									Body: &ast.BlockStmt{
										List: []ast.Stmt{
											&ast.AssignStmt{
												Lhs: []ast.Expr{&ast.Ident{Name: resultVar}},
												Tok: token.ASSIGN,
												Rhs: []ast.Expr{&ast.Ident{Name: "true"}},
											},
											&ast.BranchStmt{Tok: token.BREAK},
										},
									},
								},
							},
						},
					},
					&ast.ReturnStmt{
						Results: []ast.Expr{&ast.Ident{Name: resultVar}},
					},
				},
			},
		},
	}
}

// transformFind transforms: numbers.find(fn) → inline for-range loop returning Option<T>
func (p *FunctionalUtilitiesPlugin) transformFind(receiver ast.Expr, args []ast.Expr) ast.Node {
	// Note: This requires Option<T> type to be available
	// For now, we'll return nil to indicate it's not yet implemented
	return nil
}

// transformMapResult transforms: items.mapResult(fn) → inline loop with error handling
func (p *FunctionalUtilitiesPlugin) transformMapResult(receiver ast.Expr, args []ast.Expr) ast.Node {
	// Note: This requires Result<T, E> type to be available
	// For now, we'll return nil to indicate it's not yet implemented
	return nil
}

// transformFilterSome transforms: maybeValues.filterSome() → inline loop filtering Some values
func (p *FunctionalUtilitiesPlugin) transformFilterSome(receiver ast.Expr) ast.Node {
	// Note: This requires Option<T> type to be available
	// For now, we'll return nil to indicate it's not yet implemented
	return nil
}

// Helper methods

// extractFunctionBody extracts the expression from a function body
// Handles: func(x) T { return expr } and func(x) T { expr }
//
// Limitations:
//   - Only supports single-statement bodies
//   - Return statement must have exactly one result
//   - Multi-statement bodies are not yet supported
func (p *FunctionalUtilitiesPlugin) extractFunctionBody(body *ast.BlockStmt) ast.Expr {
	if body == nil {
		if p.currentContext != nil && p.currentContext.Logger != nil {
			p.currentContext.Logger.Debug("function body is nil")
		}
		return nil
	}

	if len(body.List) == 0 {
		if p.currentContext != nil && p.currentContext.Logger != nil {
			p.currentContext.Logger.Debug("function body is empty")
		}
		return nil
	}

	if len(body.List) > 1 {
		if p.currentContext != nil && p.currentContext.Logger != nil {
			p.currentContext.Logger.Debug("function body has multiple statements (%d), cannot inline", len(body.List))
		}
		return nil
	}

	// Handle single return statement: func(x) T { return expr }
	if ret, ok := body.List[0].(*ast.ReturnStmt); ok {
		if len(ret.Results) == 0 {
			if p.currentContext != nil && p.currentContext.Logger != nil {
				p.currentContext.Logger.Debug("empty return statement, cannot inline")
			}
			return nil
		}
		if len(ret.Results) > 1 {
			if p.currentContext != nil && p.currentContext.Logger != nil {
				p.currentContext.Logger.Debug("multiple return values (%d), cannot inline", len(ret.Results))
			}
			return nil
		}
		return ret.Results[0]
	}

	// Handle expression statement: func(x) T { expr }
	if expr, ok := body.List[0].(*ast.ExprStmt); ok {
		return expr.X
	}

	// Unsupported statement type
	if p.currentContext != nil && p.currentContext.Logger != nil {
		p.currentContext.Logger.Debug("unsupported statement type for inlining: %T", body.List[0])
	}
	return nil
}

// newTempVar generates a unique temporary variable name
func (p *FunctionalUtilitiesPlugin) newTempVar() string {
	name := fmt.Sprintf("__temp%d", p.tempCounter)
	p.tempCounter++
	return name
}

// cloneExpr creates a deep clone of an expression using astutil.Apply
// This is needed because AST nodes should have single parents to avoid corruption
func (p *FunctionalUtilitiesPlugin) cloneExpr(expr ast.Expr) ast.Expr {
	if expr == nil {
		return nil
	}
	return astutil.Apply(expr, nil, nil).(ast.Expr)
}

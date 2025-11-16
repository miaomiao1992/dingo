// Package builtin provides built-in Dingo transformation plugins
package builtin

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/token"
	"go/types"
)

// TypeInference provides comprehensive type resolution using go/types
type TypeInference struct {
	fset   *token.FileSet
	info   *types.Info
	pkg    *types.Package
	config *types.Config
}

// NewTypeInference creates a new type inference engine for a file
func NewTypeInference(fset *token.FileSet, file *ast.File) (*TypeInference, error) {
	// Initialize type information maps
	info := &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Implicits:  make(map[ast.Node]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
		Scopes:     make(map[ast.Node]*types.Scope),
	}

	// Configure type checker
	config := &types.Config{
		Importer: importer.Default(),
		Error: func(err error) {
			// Collect errors but don't fail - we can still infer some types
			// In production, we might log these errors
		},
	}

	// Extract package name from the file
	packageName := "main"
	if file.Name != nil {
		packageName = file.Name.Name
	}

	// Type-check the file
	pkg, err := config.Check(packageName, fset, []*ast.File{file}, info)
	if err != nil {
		// Continue even with type errors - partial type information is better than none
	}

	return &TypeInference{
		fset:   fset,
		info:   info,
		pkg:    pkg,
		config: config,
	}, nil
}

// Close releases memory used by type inference maps
func (ti *TypeInference) Close() {
	if ti.info != nil {
		ti.info.Types = nil
		ti.info.Defs = nil
		ti.info.Uses = nil
		ti.info.Implicits = nil
		ti.info.Selections = nil
		ti.info.Scopes = nil
	}
}

// InferFunctionReturnType determines the return type of a function
// Returns the type of the first return value (before the error)
func (ti *TypeInference) InferFunctionReturnType(fn *ast.FuncDecl) (types.Type, error) {
	if fn.Type.Results == nil || len(fn.Type.Results.List) == 0 {
		return nil, fmt.Errorf("function has no return values")
	}

	// Get the first return value type
	firstReturn := fn.Type.Results.List[0]
	typ := ti.info.TypeOf(firstReturn.Type)
	if typ == nil {
		return nil, fmt.Errorf("could not determine return type")
	}

	return typ, nil
}

// InferExpressionType determines the type of an expression
func (ti *TypeInference) InferExpressionType(expr ast.Expr) (types.Type, error) {
	typ := ti.info.TypeOf(expr)
	if typ == nil {
		return nil, fmt.Errorf("could not determine expression type")
	}
	return typ, nil
}

// GenerateZeroValue creates an AST expression representing the zero value for a type
func (ti *TypeInference) GenerateZeroValue(typ types.Type) ast.Expr {
	if typ == nil {
		return &ast.Ident{Name: "nil"}
	}

	switch t := typ.(type) {
	case *types.Basic:
		return ti.basicZeroValue(t)
	case *types.Pointer:
		return &ast.Ident{Name: "nil"}
	case *types.Slice:
		return &ast.Ident{Name: "nil"}
	case *types.Map:
		return &ast.Ident{Name: "nil"}
	case *types.Chan:
		return &ast.Ident{Name: "nil"}
	case *types.Interface:
		return &ast.Ident{Name: "nil"}
	case *types.Struct:
		// Generate composite literal: TypeName{}
		return &ast.CompositeLit{
			Type: ti.typeToAST(t),
		}
	case *types.Array:
		// Generate composite literal: [N]Type{}
		return &ast.CompositeLit{
			Type: ti.typeToAST(t),
		}
	case *types.Named:
		// For named types, check the underlying type
		underlying := t.Underlying()
		switch underlying.(type) {
		case *types.Struct:
			// For named structs, use TypeName{}
			return &ast.CompositeLit{
				Type: &ast.Ident{Name: t.Obj().Name()},
			}
		case *types.Basic, *types.Pointer, *types.Slice, *types.Map, *types.Interface:
			// Use underlying type's zero value
			return ti.GenerateZeroValue(underlying)
		default:
			// Fallback: composite literal with type name
			return &ast.CompositeLit{
				Type: &ast.Ident{Name: t.Obj().Name()},
			}
		}
	default:
		// Fallback: use zero value initialization with new
		return &ast.StarExpr{
			X: &ast.CallExpr{
				Fun: &ast.Ident{Name: "new"},
				Args: []ast.Expr{
					ti.typeToAST(typ),
				},
			},
		}
	}
}

// basicZeroValue generates zero value for basic Go types
func (ti *TypeInference) basicZeroValue(basic *types.Basic) ast.Expr {
	switch basic.Kind() {
	case types.Bool:
		return &ast.Ident{Name: "false"}
	case types.Int, types.Int8, types.Int16, types.Int32, types.Int64:
		return &ast.BasicLit{Kind: token.INT, Value: "0"}
	case types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64, types.Uintptr:
		return &ast.BasicLit{Kind: token.INT, Value: "0"}
	case types.Float32, types.Float64:
		return &ast.BasicLit{Kind: token.FLOAT, Value: "0.0"}
	case types.Complex64, types.Complex128:
		return &ast.BasicLit{Kind: token.IMAG, Value: "0i"}
	case types.String:
		return &ast.BasicLit{Kind: token.STRING, Value: `""`}
	case types.UnsafePointer:
		return &ast.Ident{Name: "nil"}
	default:
		// UntypedNil, etc
		return &ast.Ident{Name: "nil"}
	}
}

// typeToAST converts a types.Type to an AST type expression
// This is needed for composite literals and type conversions
func (ti *TypeInference) typeToAST(typ types.Type) ast.Expr {
	switch t := typ.(type) {
	case *types.Basic:
		return &ast.Ident{Name: t.Name()}
	case *types.Pointer:
		return &ast.StarExpr{
			X: ti.typeToAST(t.Elem()),
		}
	case *types.Slice:
		return &ast.ArrayType{
			Elt: ti.typeToAST(t.Elem()),
		}
	case *types.Array:
		return &ast.ArrayType{
			Len: &ast.BasicLit{
				Kind:  token.INT,
				Value: fmt.Sprintf("%d", t.Len()),
			},
			Elt: ti.typeToAST(t.Elem()),
		}
	case *types.Map:
		return &ast.MapType{
			Key:   ti.typeToAST(t.Key()),
			Value: ti.typeToAST(t.Elem()),
		}
	case *types.Chan:
		return &ast.ChanType{
			Dir:   ti.chanDirToAST(t.Dir()),
			Value: ti.typeToAST(t.Elem()),
		}
	case *types.Struct:
		// For anonymous structs, we'd need to reconstruct the full type
		// For now, return placeholder
		return &ast.Ident{Name: "struct{}"}
	case *types.Interface:
		if t.Empty() {
			return &ast.InterfaceType{
				Methods: &ast.FieldList{},
			}
		}
		// For non-empty interfaces, use placeholder
		return &ast.Ident{Name: "interface{}"}
	case *types.Named:
		// Get the type name
		obj := t.Obj()
		if obj != nil {
			// Check if it's from the same package or imported
			if obj.Pkg() == nil || obj.Pkg() == ti.pkg {
				// Same package - use simple name
				return &ast.Ident{Name: obj.Name()}
			} else {
				// Different package - use selector
				return &ast.SelectorExpr{
					X:   &ast.Ident{Name: obj.Pkg().Name()},
					Sel: &ast.Ident{Name: obj.Name()},
				}
			}
		}
		return &ast.Ident{Name: "unknown"}
	default:
		return &ast.Ident{Name: "interface{}"}
	}
}

// chanDirToAST converts channel direction to AST representation
func (ti *TypeInference) chanDirToAST(dir types.ChanDir) ast.ChanDir {
	switch dir {
	case types.SendRecv:
		return ast.SEND | ast.RECV
	case types.SendOnly:
		return ast.SEND
	case types.RecvOnly:
		return ast.RECV
	default:
		return ast.SEND | ast.RECV
	}
}

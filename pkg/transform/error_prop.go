package transform

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

// transformErrorProp transforms error propagation placeholders
// Pattern: __dingo_try_N__(expr) → proper error handling
func (t *Transformer) transformErrorProp(cursor *astutil.Cursor, call *ast.CallExpr) bool {
	// Extract the try number from function name
	ident := call.Fun.(*ast.Ident)
	name := ident.Name
	tryNum := extractTryNumber(name)

	// Get the wrapped expression
	if len(call.Args) != 1 {
		// Malformed placeholder
		return true
	}
	wrappedExpr := call.Args[0]

	// Analyze context to determine transformation strategy
	ctx := t.analyzeErrorPropContext(cursor)

	switch ctx.Type {
	case ErrorPropReturn:
		t.transformErrorPropReturn(cursor, wrappedExpr, tryNum, ctx)
	case ErrorPropAssignment:
		t.transformErrorPropAssignment(cursor, wrappedExpr, tryNum, ctx)
	default:
		// For now, leave as-is if we can't determine context
		return true
	}

	return false // Stop traversal into this subtree
}

// ErrorPropContext describes where the ? operator appears
type ErrorPropContext struct {
	Type         ErrorPropContextType
	FuncDecl     *ast.FuncDecl
	ReturnTypes  []ast.Expr
	ZeroValues   []ast.Expr
	AssignStmt   *ast.AssignStmt
	VarNames     []*ast.Ident
}

type ErrorPropContextType int

const (
	ErrorPropUnknown ErrorPropContextType = iota
	ErrorPropReturn
	ErrorPropAssignment
)

// analyzeErrorPropContext walks up the AST to understand context
func (t *Transformer) analyzeErrorPropContext(cursor *astutil.Cursor) ErrorPropContext {
	ctx := ErrorPropContext{Type: ErrorPropUnknown}

	// Walk up to find parent nodes
	node := cursor.Node()
	parent := cursor.Parent()

	// Check if we're in a return statement
	if retStmt, ok := parent.(*ast.ReturnStmt); ok {
		ctx.Type = ErrorPropReturn
		ctx.FuncDecl = t.findEnclosingFunc(cursor)
		if ctx.FuncDecl != nil && ctx.FuncDecl.Type.Results != nil {
			ctx.ReturnTypes = make([]ast.Expr, len(ctx.FuncDecl.Type.Results.List))
			for i, field := range ctx.FuncDecl.Type.Results.List {
				ctx.ReturnTypes[i] = field.Type
			}
			ctx.ZeroValues = t.generateZeroValues(ctx.ReturnTypes)
		}
		return ctx
	}

	// Check if we're in an assignment or variable declaration
	if assignStmt, ok := parent.(*ast.AssignStmt); ok {
		ctx.Type = ErrorPropAssignment
		ctx.AssignStmt = assignStmt
		ctx.FuncDecl = t.findEnclosingFunc(cursor)

		// Extract variable names from LHS
		for _, lhs := range assignStmt.Lhs {
			if ident, ok := lhs.(*ast.Ident); ok {
				ctx.VarNames = append(ctx.VarNames, ident)
			}
		}

		if ctx.FuncDecl != nil && ctx.FuncDecl.Type.Results != nil {
			ctx.ReturnTypes = make([]ast.Expr, len(ctx.FuncDecl.Type.Results.List))
			for i, field := range ctx.FuncDecl.Type.Results.List {
				ctx.ReturnTypes[i] = field.Type
			}
			ctx.ZeroValues = t.generateZeroValues(ctx.ReturnTypes)
		}
		return ctx
	}

	return ctx
}

// findEnclosingFunc finds the enclosing function declaration
func (t *Transformer) findEnclosingFunc(cursor *astutil.Cursor) *ast.FuncDecl {
	// Walk up the cursor stack to find FuncDecl
	// Note: astutil.Cursor doesn't expose full stack, so we need to search
	// This is a simplified implementation - in production we'd maintain a stack

	// For now, return nil and we'll handle this limitation
	// TODO: Maintain a function context stack during traversal
	return nil
}

// transformErrorPropReturn handles: return expr?
func (t *Transformer) transformErrorPropReturn(cursor *astutil.Cursor, expr ast.Expr, tryNum int, ctx ErrorPropContext) {
	// Generate:
	// __tmpN, __errN := expr
	// if __errN != nil {
	//     return zeroValues..., __errN
	// }
	// return __tmpN

	tmpVar := fmt.Sprintf("__tmp%d", tryNum-1)
	errVar := fmt.Sprintf("__err%d", tryNum-1)

	// Build the statements
	stmts := []ast.Stmt{}

	// 1. Assignment: __tmpN, __errN := expr
	assignStmt := &ast.AssignStmt{
		Lhs: []ast.Expr{
			ast.NewIdent(tmpVar),
			ast.NewIdent(errVar),
		},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{expr},
	}
	stmts = append(stmts, assignStmt)

	// 2. Error check: if __errN != nil { return ..., __errN }
	var returnExprs []ast.Expr
	if len(ctx.ZeroValues) > 0 {
		returnExprs = append(returnExprs, ctx.ZeroValues...)
		returnExprs[len(returnExprs)-1] = ast.NewIdent(errVar)
	} else {
		returnExprs = []ast.Expr{ast.NewIdent(errVar)}
	}

	ifStmt := &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  ast.NewIdent(errVar),
			Op: token.NEQ,
			Y:  ast.NewIdent("nil"),
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: returnExprs,
				},
			},
		},
	}
	stmts = append(stmts, ifStmt)

	// 3. Final return: return __tmpN
	finalReturn := &ast.ReturnStmt{
		Results: []ast.Expr{ast.NewIdent(tmpVar)},
	}
	stmts = append(stmts, finalReturn)

	// Replace the return statement with our expanded statements
	// We need to replace at the block level
	cursor.Replace(stmts[0])
	// Note: This is simplified - we'd need to insert multiple statements
}

// transformErrorPropAssignment handles: let x = expr?
func (t *Transformer) transformErrorPropAssignment(cursor *astutil.Cursor, expr ast.Expr, tryNum int, ctx ErrorPropContext) {
	// Generate:
	// __tmpN, __errN := expr
	// if __errN != nil {
	//     return zeroValues..., __errN
	// }
	// var x = __tmpN

	tmpVar := fmt.Sprintf("__tmp%d", tryNum-1)
	errVar := fmt.Sprintf("__err%d", tryNum-1)

	// This is complex because we need to replace one statement with multiple
	// For now, we'll use a simpler approach: replace the RHS in the assignment
	// and insert the error check before it

	// Replace the call expression with just the tmp variable
	cursor.Replace(ast.NewIdent(tmpVar))

	// TODO: Insert the assignment and error check statements
	// This requires more complex AST manipulation
}

// generateZeroValues creates zero value expressions for given types
func (t *Transformer) generateZeroValues(types []ast.Expr) []ast.Expr {
	zeros := make([]ast.Expr, len(types))
	for i, typ := range types {
		zeros[i] = t.zeroValueForType(typ)
	}
	return zeros
}

// zeroValueForType returns the zero value expression for a type
func (t *Transformer) zeroValueForType(typ ast.Expr) ast.Expr {
	// Simplified implementation
	if typ == nil {
		return ast.NewIdent("nil")
	}

	switch t := typ.(type) {
	case *ast.Ident:
		switch t.Name {
		case "int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64",
			"byte", "rune":
			return &ast.BasicLit{Kind: token.INT, Value: "0"}
		case "float32", "float64":
			return &ast.BasicLit{Kind: token.FLOAT, Value: "0.0"}
		case "string":
			return &ast.BasicLit{Kind: token.STRING, Value: `""`}
		case "bool":
			return ast.NewIdent("false")
		case "error":
			return ast.NewIdent("nil")
		default:
			return ast.NewIdent("nil")
		}
	case *ast.StarExpr, *ast.ArrayType, *ast.MapType, *ast.InterfaceType, *ast.ChanType:
		return ast.NewIdent("nil")
	case *ast.StructType:
		// For struct types, use the type with {}
		return &ast.CompositeLit{Type: typ}
	default:
		return ast.NewIdent("nil")
	}
}

// extractTryNumber extracts N from __dingo_try_N__
func extractTryNumber(name string) int {
	// name format: __dingo_try_N__
	parts := strings.Split(name, "_")
	if len(parts) >= 4 {
		if num, err := strconv.Atoi(parts[3]); err == nil {
			return num
		}
	}
	return 0
}

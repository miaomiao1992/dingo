// Package builtin provides built-in Dingo transformation plugins
package builtin

import (
	"fmt"
	"go/ast"
	"go/token"

	dingoast "github.com/MadAppGang/dingo/pkg/ast"
	"github.com/MadAppGang/dingo/pkg/plugin"
)

// ErrorPropagationPlugin transforms error propagation expressions into Go early-return patterns
//
// Transforms:
//   expr?     →  if err := expr; err != nil { return zeroValue, err }
//   expr!     →  if err := expr; err != nil { return zeroValue, err }
//   try expr  →  if err := expr; err != nil { return zeroValue, err }
//
// All three syntaxes produce identical Go code - only the parsing differs.
type ErrorPropagationPlugin struct {
	plugin.BasePlugin
	errorVarCounter int
	tmpVarCounter   int
}

// NewErrorPropagationPlugin creates a new error propagation transformation plugin
func NewErrorPropagationPlugin() *ErrorPropagationPlugin {
	return &ErrorPropagationPlugin{
		BasePlugin:      *plugin.NewBasePlugin("error_propagation", "Error propagation with ? operator", nil),
		errorVarCounter: 0,
		tmpVarCounter:   0,
	}
}

// Name returns the plugin name
func (p *ErrorPropagationPlugin) Name() string {
	return "error_propagation"
}

// Transform transforms an AST node
func (p *ErrorPropagationPlugin) Transform(ctx *plugin.Context, node ast.Node) (ast.Node, error) {
	// Only transform ErrorPropagationExpr nodes
	errExpr, ok := node.(*dingoast.ErrorPropagationExpr)
	if !ok {
		return node, nil
	}

	// Generate the transformation
	return p.transformErrorPropagation(ctx, errExpr)
}

// transformErrorPropagation converts ErrorPropagationExpr to Go early-return pattern
//
// Input:  let user = fetchUser(id)?
// Output:
//   __tmp0, __err0 := fetchUser(id)
//   if __err0 != nil {
//       return nil, __err0
//   }
//   user := __tmp0
//
// LIMITATION (Phase 1): Only works in statement context.
// Expression contexts (e.g., "return fetchUser(id)?") require statement lifting,
// which will be implemented in Phase 1.5 with the full transformer pipeline.
//
// The returned node is a series of statements, not a single expression.
// The caller (transformer) must handle extracting these statements appropriately.
func (p *ErrorPropagationPlugin) transformErrorPropagation(
	ctx *plugin.Context,
	expr *dingoast.ErrorPropagationExpr,
) (ast.Node, error) {
	// Generate unique variable names
	tmpVar := p.nextTmpVar()
	errVar := p.nextErrVar()

	// Create the assignment: __tmp0, __err0 := expr
	assignStmt := &ast.AssignStmt{
		Lhs: []ast.Expr{
			ast.NewIdent(tmpVar),
			ast.NewIdent(errVar),
		},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{expr.X},
	}

	// Create the error check: if __err0 != nil
	condition := &ast.BinaryExpr{
		X:  ast.NewIdent(errVar),
		Op: token.NEQ,
		Y:  ast.NewIdent("nil"),
	}

	// Create the return statement: return zeroValue, __err0
	// TODO: Implement proper zero value generation using type inference
	// This requires go/types integration to determine the function's return type
	// For now, we use nil which works for pointer/interface types
	// Phase 1.5 will add full type inference support
	returnStmt := &ast.ReturnStmt{
		Results: []ast.Expr{
			ast.NewIdent("nil"), // Temporary: works for pointers/interfaces
			ast.NewIdent(errVar),
		},
	}

	// Create the if statement
	ifStmt := &ast.IfStmt{
		Cond: condition,
		Body: &ast.BlockStmt{
			List: []ast.Stmt{returnStmt},
		},
	}

	// Create a block statement containing both the assignment and if check
	blockStmt := &ast.BlockStmt{
		List: []ast.Stmt{
			assignStmt,
			ifStmt,
		},
	}

	// TODO(Phase 1.5): Record source map mapping
	// Source map integration will be added when we wire this into the generator
	// srcPos := ctx.FileSet.Position(expr.Pos())
	// genPos := ctx.FileSet.Position(assignStmt.Pos())
	// sourceMapGenerator.AddMapping(srcPos, genPos)

	// Return a composite structure that the transformer can unpack
	// The transformer will need to handle this based on context:
	// - Statement context: inject these statements inline
	// - Expression context: lift to enclosing block (Phase 1.5)
	//
	// For Phase 1, we return the statements as a slice wrapped in ExprStmt
	// This is a temporary approach until the full transformer is implemented
	return &temporaryStmtWrapper{stmts: blockStmt.List, tmpVar: tmpVar}, nil
}

// temporaryStmtWrapper wraps multiple statements for the transformer to unpack
// This is a temporary solution for Phase 1 until statement lifting is implemented
type temporaryStmtWrapper struct {
	stmts  []ast.Stmt
	tmpVar string // The temporary variable that holds the result
}

func (w *temporaryStmtWrapper) Pos() token.Pos { return w.stmts[0].Pos() }
func (w *temporaryStmtWrapper) End() token.Pos { return w.stmts[len(w.stmts)-1].End() }

// nextTmpVar generates a unique temporary variable name
func (p *ErrorPropagationPlugin) nextTmpVar() string {
	name := fmt.Sprintf("__tmp%d", p.tmpVarCounter)
	p.tmpVarCounter++
	return name
}

// nextErrVar generates a unique error variable name
func (p *ErrorPropagationPlugin) nextErrVar() string {
	name := fmt.Sprintf("__err%d", p.errorVarCounter)
	p.errorVarCounter++
	return name
}

// Reset resets the plugin's internal counters (useful for testing)
func (p *ErrorPropagationPlugin) Reset() {
	p.errorVarCounter = 0
	p.tmpVarCounter = 0
}

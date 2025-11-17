// Package builtin provides built-in Dingo transformation plugins
package builtin

import (
	"fmt"
	"go/ast"
	"go/token"
)

// StatementLifter handles lifting statements from expression contexts
// This is needed when error propagation occurs in expression position
// Example: return fetchUser(id)? must be lifted to statements before the return
type StatementLifter struct {
	counter int // Counter for generating unique temp variables
}

// NewStatementLifter creates a new statement lifter
func NewStatementLifter() *StatementLifter {
	return &StatementLifter{
		counter: 0,
	}
}

// LiftResult contains the results of lifting an expression
type LiftResult struct {
	Statements   []ast.Stmt // Statements to inject before current statement
	Replacement  ast.Expr   // Expression to replace the original with (temp variable)
	TempVarName  string     // Name of the temp variable holding the result
	ErrorVarName string     // Name of the error variable
}

// LiftExpression extracts statements from an error propagation expression
// It generates:
// 1. An assignment statement: tmpVar, errVar := expr
// 2. An error check statement: if errVar != nil { return zeroValue, errVar }
// 3. A replacement expression: tmpVar
func (sl *StatementLifter) LiftExpression(
	expr ast.Expr,
	zeroValue ast.Expr,
	errorWrapper ast.Expr, // Optional: wrapped error (from ErrorWrapper)
) *LiftResult {
	return sl.LiftExpressionWithVars(expr, zeroValue, errorWrapper, "", "")
}

// LiftExpressionWithVars is like LiftExpression but allows specifying variable names
// If tmpVarName or errVarName are empty, unique names will be generated
func (sl *StatementLifter) LiftExpressionWithVars(
	expr ast.Expr,
	zeroValue ast.Expr,
	errorWrapper ast.Expr,
	tmpVarName string,
	errVarName string,
) *LiftResult {
	// Generate variable names if not provided
	if tmpVarName == "" {
		tmpVarName = fmt.Sprintf("__tmp%d", sl.counter)
	}
	if errVarName == "" {
		errVarName = fmt.Sprintf("__err%d", sl.counter)
	}

	tmpVar := tmpVarName
	errVar := errVarName
	sl.counter++

	// Create assignment: tmpVar, errVar := expr
	assignStmt := &ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.Ident{Name: tmpVar},
			&ast.Ident{Name: errVar},
		},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{expr},
	}

	// Determine what to return in the error case
	var errorReturn ast.Expr
	if errorWrapper != nil {
		errorReturn = errorWrapper
	} else {
		errorReturn = &ast.Ident{Name: errVar}
	}

	// Create error check: if errVar != nil { return zeroValue, errVar }
	errorCheck := &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  &ast.Ident{Name: errVar},
			Op: token.NEQ,
			Y:  &ast.Ident{Name: "nil"},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						zeroValue,
						errorReturn,
					},
				},
			},
		},
	}

	return &LiftResult{
		Statements: []ast.Stmt{
			assignStmt,
			errorCheck,
		},
		Replacement:  &ast.Ident{Name: tmpVar},
		TempVarName:  tmpVar,
		ErrorVarName: errVar,
	}
}

// LiftStatement handles error propagation in statement context
// For statements, we don't need to lift - we can inject inline
// But we still need to split into multiple statements
func (sl *StatementLifter) LiftStatement(
	varName string,
	expr ast.Expr,
	zeroValue ast.Expr,
	errorWrapper ast.Expr,
) []ast.Stmt {
	// Generate unique error variable name
	errVar := fmt.Sprintf("__err%d", sl.counter)
	sl.counter++

	// Create assignment: varName, errVar := expr
	assignStmt := &ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.Ident{Name: varName},
			&ast.Ident{Name: errVar},
		},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{expr},
	}

	// Determine what to return in the error case
	var errorReturn ast.Expr
	if errorWrapper != nil {
		errorReturn = errorWrapper
	} else {
		errorReturn = &ast.Ident{Name: errVar}
	}

	// Create error check: if errVar != nil { return zeroValue, errVar }
	errorCheck := &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  &ast.Ident{Name: errVar},
			Op: token.NEQ,
			Y:  &ast.Ident{Name: "nil"},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						zeroValue,
						errorReturn,
					},
				},
			},
		},
	}

	return []ast.Stmt{assignStmt, errorCheck}
}

// InjectStatements inserts new statements into a block before a target index
func (sl *StatementLifter) InjectStatements(
	block *ast.BlockStmt,
	targetIndex int,
	newStmts []ast.Stmt,
) error {
	if targetIndex < 0 || targetIndex > len(block.List) {
		return fmt.Errorf("invalid target index: %d (block has %d statements)", targetIndex, len(block.List))
	}

	// Build new statement list with injected statements
	newList := make([]ast.Stmt, 0, len(block.List)+len(newStmts))

	// Copy statements before injection point
	newList = append(newList, block.List[:targetIndex]...)

	// Insert new statements
	newList = append(newList, newStmts...)

	// Copy remaining statements
	newList = append(newList, block.List[targetIndex:]...)

	// Update block
	block.List = newList

	return nil
}

// Reset resets the internal counter (useful for testing)
func (sl *StatementLifter) Reset() {
	sl.counter = 0
}

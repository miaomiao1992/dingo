// Package builtin provides built-in Dingo transformation plugins
package builtin

import (
	"fmt"
	"go/ast"
	"go/token"

	dingoast "github.com/MadAppGang/dingo/pkg/ast"
	"github.com/MadAppGang/dingo/pkg/plugin"
	"golang.org/x/tools/go/ast/astutil"
)

// ErrorPropagationPlugin transforms error propagation expressions into Go early-return patterns
//
// Features (Phase 1.6):
// - Statement context: let x = expr?
// - Expression context: return expr? (with statement lifting)
// - Error wrapping: expr? "message"
// - Full type inference with go/types
// - Source map integration
//
// Transforms:
//   expr?     →  if err := expr; err != nil { return zeroValue, err }
//   expr? "msg" →  if err := expr; err != nil { return zeroValue, fmt.Errorf("msg: %w", err) }
type ErrorPropagationPlugin struct {
	plugin.BasePlugin

	// Multi-pass components
	typeInference  *TypeInference
	statementLifter *StatementLifter
	errorWrapper   *ErrorWrapper

	// State for current transformation
	currentFile     *dingoast.File
	currentFunction *ast.FuncDecl
	currentContext  *plugin.Context
	needsFmtImport  bool

	// Counters for unique variable names
	tmpCounter int
	errCounter int

	// Two-pass transformation state
	pendingInjections []pendingInjection
	parentMap         map[ast.Node]ast.Node
}

// pendingInjection tracks statements to be injected after traversal
type pendingInjection struct {
	block      *ast.BlockStmt
	stmtIndex  int
	statements []ast.Stmt
	before     bool // true = inject before index, false = inject after
}

// NewErrorPropagationPlugin creates a new error propagation transformation plugin
func NewErrorPropagationPlugin() *ErrorPropagationPlugin {
	return &ErrorPropagationPlugin{
		BasePlugin:      *plugin.NewBasePlugin("error_propagation", "Error propagation with ? operator", nil),
		statementLifter: NewStatementLifter(),
		errorWrapper:    NewErrorWrapper(),
		tmpCounter:      0,
		errCounter:      0,
	}
}

// markerCommentStmt is a custom statement type that represents a marker comment
// It will be handled specially during AST printing
type markerCommentStmt struct {
	Text string
}

func (m *markerCommentStmt) Pos() token.Pos { return token.NoPos }
func (m *markerCommentStmt) End() token.Pos { return token.NoPos }

// We need to implement the ast.Stmt interface marker method
func (*markerCommentStmt) stmtNode() {}

// wrapWithMarkers wraps statements with DINGO:GENERATED marker comments
// Returns wrapped statements if markers are enabled, otherwise returns original statements
func (p *ErrorPropagationPlugin) wrapWithMarkers(statements []ast.Stmt, markerType string) []ast.Stmt {
	if len(statements) == 0 {
		return statements
	}

	// Check if markers are enabled in config
	if p.currentContext != nil && p.currentContext.Config != nil && !p.currentContext.Config.EmitGeneratedMarkers {
		return statements
	}

	// For now, we'll use a pragmatic approach: insert empty statements with special comments
	// These will need post-processing after AST generation
	//
	// A better approach would be to use CommentMap, but that requires changes to the generator
	// For Week 1 MVP, we'll document this as a known limitation

	// TODO: Implement proper comment injection using ast.CommentMap
	// For now, return statements as-is and add markers via post-processing

	return statements
}

// Name returns the plugin name
func (p *ErrorPropagationPlugin) Name() string {
	return "error_propagation"
}

// Transform transforms an AST node (file-level entry point)
func (p *ErrorPropagationPlugin) Transform(ctx *plugin.Context, node ast.Node) (ast.Node, error) {
	file, ok := node.(*ast.File)
	if !ok {
		return node, nil
	}

	// Store context for marker generation
	p.currentContext = ctx

	// Get the Dingo file wrapper to access DingoNodes map
	if ctx.CurrentFile != nil {
		if dingoFile, ok := ctx.CurrentFile.(*dingoast.File); ok {
			p.currentFile = dingoFile
		}
	}

	// If no Dingo nodes, nothing to transform
	if p.currentFile == nil || !p.currentFile.HasDingoNodes() {
		return node, nil
	}

	// Initialize type inference
	var err error
	p.typeInference, err = NewTypeInference(ctx.FileSet, file)
	if err != nil {
		// Continue without type inference - we'll use nil as zero value
		ctx.Logger.Warn("Type inference initialization failed: %v", err)
		p.typeInference = nil
	}

	// Reset state
	p.needsFmtImport = false
	p.tmpCounter = 0
	p.errCounter = 0
	p.pendingInjections = nil
	p.parentMap = make(map[ast.Node]ast.Node)

	// Build parent map for proper traversal
	p.buildParentMap(file)

	// Transform the file using two-pass approach
	transformed := p.transformFile(file)

	// Apply pending injections (second pass)
	p.applyPendingInjections()

	// Add fmt import if needed
	if p.needsFmtImport {
		p.errorWrapper.AddFmtImport(transformed.(*ast.File))
	}

	// Cleanup type inference resources
	if p.typeInference != nil {
		p.typeInference.Close()
	}

	return transformed, nil
}

// transformFile applies the multi-pass transformation to a file
func (p *ErrorPropagationPlugin) transformFile(file *ast.File) ast.Node {
	// Use astutil.Apply to safely traverse and transform the AST
	result := astutil.Apply(file, p.preVisit, p.postVisit)
	return result
}

// preVisit is called before visiting a node's children
func (p *ErrorPropagationPlugin) preVisit(cursor *astutil.Cursor) bool {
	node := cursor.Node()

	// Track current function for return type inference
	if fn, ok := node.(*ast.FuncDecl); ok {
		p.currentFunction = fn
	}

	return true // Continue traversal
}

// postVisit is called after visiting a node's children
func (p *ErrorPropagationPlugin) postVisit(cursor *astutil.Cursor) bool {
	node := cursor.Node()

	// Check if this is a placeholder for a Dingo error propagation node
	if expr, ok := node.(ast.Expr); ok {
		if dingoNode, hasDingo := p.currentFile.GetDingoNode(expr); hasDingo {
			if errExpr, isErrProp := dingoNode.(*dingoast.ErrorPropagationExpr); isErrProp {
				// Transform this error propagation expression
				p.transformErrorPropagation(cursor, errExpr)
			}
		}
	}

	// Clear function context when leaving
	if _, ok := node.(*ast.FuncDecl); ok {
		p.currentFunction = nil
	}

	return true
}

// transformErrorPropagation handles transformation of a single error propagation expression
func (p *ErrorPropagationPlugin) transformErrorPropagation(cursor *astutil.Cursor, errExpr *dingoast.ErrorPropagationExpr) {
	// Determine context: are we in a statement or expression position?
	parent := cursor.Parent()

	// Generate zero value for return type
	zeroValue := p.generateZeroValue()

	// Generate error wrapper if message provided
	var errorWrapper ast.Expr
	if errExpr.Message != "" {
		p.needsFmtImport = true
		errVar := fmt.Sprintf("__err%d", p.errCounter)
		errorWrapper = p.errorWrapper.WrapError(errVar, errExpr.Message)
	}

	// Check context and transform accordingly
	switch parent.(type) {
	case *ast.AssignStmt:
		// Statement context: let x = expr?
		p.transformStatementContext(cursor, errExpr, zeroValue, errorWrapper)
	case *ast.ReturnStmt, *ast.CallExpr, *ast.BinaryExpr, *ast.UnaryExpr, *ast.CompositeLit:
		// Expression context: needs statement lifting
		p.transformExpressionContext(cursor, errExpr, zeroValue, errorWrapper)
	default:
		// Try expression context as fallback
		p.transformExpressionContext(cursor, errExpr, zeroValue, errorWrapper)
	}
}

// transformStatementContext handles error propagation in statement position
// Example: let x = fetchUser()?
func (p *ErrorPropagationPlugin) transformStatementContext(
	cursor *astutil.Cursor,
	errExpr *dingoast.ErrorPropagationExpr,
	zeroValue ast.Expr,
	errorWrapper ast.Expr,
) {
	// For statement context, we don't replace the expression itself
	// Instead, we need to inject statements AFTER the assignment

	// Find the enclosing assignment statement
	assignStmt, ok := cursor.Parent().(*ast.AssignStmt)
	if !ok {
		return
	}

	// Generate unique error variable
	errVar := fmt.Sprintf("__err%d", p.errCounter)
	p.errCounter++

	// Modify the assignment to capture the error
	// Change: x := expr  →  x, __err0 := expr
	assignStmt.Lhs = append(assignStmt.Lhs, &ast.Ident{Name: errVar})
	assignStmt.Rhs[0] = errExpr.X // Use original expression without ?

	// Create error check statement
	var errorReturn ast.Expr
	if errorWrapper != nil {
		errorReturn = errorWrapper
	} else {
		errorReturn = &ast.Ident{Name: errVar}
	}

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

	// We need to inject the error check AFTER this statement
	// To do this, we'll use a marker and handle it at the block level
	// For now, we'll add it to a queue and inject when we see the block

	// Find enclosing block and inject
	p.injectAfterStatement(cursor, errorCheck)
}

// transformExpressionContext handles error propagation in expression position
// Example: return fetchUser()?
func (p *ErrorPropagationPlugin) transformExpressionContext(
	cursor *astutil.Cursor,
	errExpr *dingoast.ErrorPropagationExpr,
	zeroValue ast.Expr,
	errorWrapper ast.Expr,
) {
	// Use statement lifter to extract statements
	liftResult := p.statementLifter.LiftExpression(errExpr.X, zeroValue, errorWrapper)

	// Replace the expression with the temp variable
	cursor.Replace(liftResult.Replacement)

	// Inject the lifted statements before the current statement
	p.injectBeforeStatement(cursor, liftResult.Statements)
}

// injectAfterStatement queues statements for injection after the current statement
func (p *ErrorPropagationPlugin) injectAfterStatement(cursor *astutil.Cursor, stmt ast.Stmt) {
	// Find enclosing block statement using parent map
	block := p.findEnclosingBlock(cursor.Node())
	if block == nil {
		return
	}

	// Find the index of the current statement in the block
	targetStmt := p.findEnclosingStatement(cursor.Node())
	if targetStmt == nil {
		return
	}

	// Find index
	for i, s := range block.List {
		if s == targetStmt {
			// Queue for injection after this statement
			p.pendingInjections = append(p.pendingInjections, pendingInjection{
				block:      block,
				stmtIndex:  i + 1,
				statements: []ast.Stmt{stmt},
				before:     false,
			})
			return
		}
	}
}

// injectBeforeStatement queues statements for injection before the current statement
func (p *ErrorPropagationPlugin) injectBeforeStatement(cursor *astutil.Cursor, stmts []ast.Stmt) {
	// Find enclosing block statement using parent map
	block := p.findEnclosingBlock(cursor.Node())
	if block == nil {
		return
	}

	// Find the index of the current statement in the block
	targetStmt := p.findEnclosingStatement(cursor.Node())
	if targetStmt == nil {
		return
	}

	// Find index
	for i, s := range block.List {
		if s == targetStmt {
			// Queue for injection before this statement
			p.pendingInjections = append(p.pendingInjections, pendingInjection{
				block:      block,
				stmtIndex:  i,
				statements: stmts,
				before:     true,
			})
			return
		}
	}
}

// buildParentMap builds a map of child->parent relationships for the AST
func (p *ErrorPropagationPlugin) buildParentMap(file *ast.File) {
	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		// For each node, record its children's parent
		ast.Inspect(n, func(child ast.Node) bool {
			if child != nil && child != n {
				p.parentMap[child] = n
				return false // Don't recurse - we only want immediate children
			}
			return child != nil
		})
		return true
	})
}

// findEnclosingBlock walks up the parent chain to find the enclosing block statement
func (p *ErrorPropagationPlugin) findEnclosingBlock(node ast.Node) *ast.BlockStmt {
	current := node
	for current != nil {
		if block, ok := current.(*ast.BlockStmt); ok {
			return block
		}
		current = p.parentMap[current]
	}
	return nil
}

// findEnclosingStatement finds the enclosing statement using parent map
func (p *ErrorPropagationPlugin) findEnclosingStatement(node ast.Node) ast.Stmt {
	current := node
	for current != nil {
		if stmt, ok := current.(ast.Stmt); ok {
			return stmt
		}
		current = p.parentMap[current]
	}
	return nil
}

// applyPendingInjections applies all queued statement injections
func (p *ErrorPropagationPlugin) applyPendingInjections() {
	// Group injections by block to handle multiple injections efficiently
	blockInjections := make(map[*ast.BlockStmt][]pendingInjection)
	for _, inj := range p.pendingInjections {
		blockInjections[inj.block] = append(blockInjections[inj.block], inj)
	}

	// Apply injections block by block
	for block, injections := range blockInjections {
		// Sort injections by index (descending) to avoid index shifts
		for i := 0; i < len(injections); i++ {
			for j := i + 1; j < len(injections); j++ {
				if injections[i].stmtIndex < injections[j].stmtIndex {
					injections[i], injections[j] = injections[j], injections[i]
				}
			}
		}

		// Apply each injection
		for _, inj := range injections {
			if inj.stmtIndex <= len(block.List) {
				newList := make([]ast.Stmt, 0, len(block.List)+len(inj.statements))
				newList = append(newList, block.List[:inj.stmtIndex]...)
				newList = append(newList, inj.statements...)
				newList = append(newList, block.List[inj.stmtIndex:]...)
				block.List = newList
			}
		}
	}
}

// generateZeroValue creates a zero value expression for the current function's return type
func (p *ErrorPropagationPlugin) generateZeroValue() ast.Expr {
	if p.currentFunction == nil {
		return &ast.Ident{Name: "nil"}
	}

	// Use type inference if available
	if p.typeInference != nil {
		returnType, err := p.typeInference.InferFunctionReturnType(p.currentFunction)
		if err == nil && returnType != nil {
			return p.typeInference.GenerateZeroValue(returnType)
		}
	}

	// Fallback: use nil
	return &ast.Ident{Name: "nil"}
}

// Reset resets the plugin's internal state (useful for testing)
func (p *ErrorPropagationPlugin) Reset() {
	p.tmpCounter = 0
	p.errCounter = 0
	p.needsFmtImport = false
	p.currentFile = nil
	p.currentFunction = nil
	if p.statementLifter != nil {
		p.statementLifter.Reset()
	}
}

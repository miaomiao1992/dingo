package builtin

import (
	"go/ast"
	"go/token"
	"testing"

	dingoast "github.com/MadAppGang/dingo/pkg/ast"
	"github.com/MadAppGang/dingo/pkg/plugin"
)

func TestNewErrorPropagationPlugin(t *testing.T) {
	p := NewErrorPropagationPlugin()

	if p == nil {
		t.Fatal("Expected plugin to be non-nil")
	}

	if p.Name() != "error_propagation" {
		t.Errorf("Expected name 'error_propagation', got %q", p.Name())
	}

	if p.errorVarCounter != 0 {
		t.Errorf("Expected errorVarCounter to be 0, got %d", p.errorVarCounter)
	}

	if p.tmpVarCounter != 0 {
		t.Errorf("Expected tmpVarCounter to be 0, got %d", p.tmpVarCounter)
	}
}

func TestTransformNonErrorPropagationExpr(t *testing.T) {
	p := NewErrorPropagationPlugin()
	ctx := &plugin.Context{
		FileSet: token.NewFileSet(),
	}

	// Create a regular call expression
	callExpr := &ast.CallExpr{
		Fun: ast.NewIdent("fetchUser"),
	}

	// Transform should return the node unchanged
	result, err := p.Transform(ctx, callExpr)
	if err != nil {
		t.Fatalf("Transform() error = %v", err)
	}

	if result != callExpr {
		t.Error("Expected node to be returned unchanged for non-ErrorPropagationExpr")
	}
}

func TestTransformBasicErrorPropagation(t *testing.T) {
	p := NewErrorPropagationPlugin()
	ctx := &plugin.Context{
		FileSet: token.NewFileSet(),
	}

	// Create ErrorPropagationExpr: fetchUser(id)?
	callExpr := &ast.CallExpr{
		Fun: ast.NewIdent("fetchUser"),
		Args: []ast.Expr{
			ast.NewIdent("id"),
		},
	}

	errExpr := &dingoast.ErrorPropagationExpr{
		X:      callExpr,
		OpPos:  token.NoPos,
		Syntax: dingoast.SyntaxQuestion,
	}

	// Transform
	result, err := p.Transform(ctx, errExpr)
	if err != nil {
		t.Fatalf("Transform() error = %v", err)
	}

	// Result should be temporaryStmtWrapper
	wrapper, ok := result.(*temporaryStmtWrapper)
	if !ok {
		t.Fatalf("Expected *temporaryStmtWrapper, got %T", result)
	}

	// Should have 2 statements: assignment and if
	if len(wrapper.stmts) != 2 {
		t.Fatalf("Expected 2 statements, got %d", len(wrapper.stmts))
	}

	// First statement should be assignment
	assignStmt, ok := wrapper.stmts[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("Expected first stmt to be *ast.AssignStmt, got %T", wrapper.stmts[0])
	}

	// Check assignment has 2 LHS (tmp, err) and 1 RHS (call)
	if len(assignStmt.Lhs) != 2 {
		t.Errorf("Expected 2 LHS variables, got %d", len(assignStmt.Lhs))
	}

	if len(assignStmt.Rhs) != 1 {
		t.Errorf("Expected 1 RHS expression, got %d", len(assignStmt.Rhs))
	}

	if assignStmt.Tok != token.DEFINE {
		t.Errorf("Expected := operator, got %v", assignStmt.Tok)
	}

	// Second statement should be if statement
	ifStmt, ok := wrapper.stmts[1].(*ast.IfStmt)
	if !ok {
		t.Fatalf("Expected second stmt to be *ast.IfStmt, got %T", wrapper.stmts[1])
	}

	// Check if condition is `err != nil`
	binExpr, ok := ifStmt.Cond.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("Expected binary expression in if condition, got %T", ifStmt.Cond)
	}

	if binExpr.Op != token.NEQ {
		t.Errorf("Expected != operator, got %v", binExpr.Op)
	}

	// Check if body has return statement
	if len(ifStmt.Body.List) != 1 {
		t.Fatalf("Expected 1 statement in if body, got %d", len(ifStmt.Body.List))
	}

	returnStmt, ok := ifStmt.Body.List[0].(*ast.ReturnStmt)
	if !ok {
		t.Fatalf("Expected return statement in if body, got %T", ifStmt.Body.List[0])
	}

	// Check return has 2 values
	if len(returnStmt.Results) != 2 {
		t.Errorf("Expected 2 return values, got %d", len(returnStmt.Results))
	}
}

func TestUniqueVariableNames(t *testing.T) {
	p := NewErrorPropagationPlugin()
	ctx := &plugin.Context{
		FileSet: token.NewFileSet(),
	}

	// Create two error propagations
	errExpr1 := &dingoast.ErrorPropagationExpr{
		X:      &ast.CallExpr{Fun: ast.NewIdent("fetchUser")},
		OpPos:  token.NoPos,
		Syntax: dingoast.SyntaxQuestion,
	}

	errExpr2 := &dingoast.ErrorPropagationExpr{
		X:      &ast.CallExpr{Fun: ast.NewIdent("fetchPost")},
		OpPos:  token.NoPos,
		Syntax: dingoast.SyntaxQuestion,
	}

	// Transform both
	result1, err := p.Transform(ctx, errExpr1)
	if err != nil {
		t.Fatalf("Transform() error = %v", err)
	}

	result2, err := p.Transform(ctx, errExpr2)
	if err != nil {
		t.Fatalf("Transform() error = %v", err)
	}

	wrapper1 := result1.(*temporaryStmtWrapper)
	wrapper2 := result2.(*temporaryStmtWrapper)

	// Get variable names from first transformation
	assign1 := wrapper1.stmts[0].(*ast.AssignStmt)
	tmpVar1 := assign1.Lhs[0].(*ast.Ident).Name
	errVar1 := assign1.Lhs[1].(*ast.Ident).Name

	// Get variable names from second transformation
	assign2 := wrapper2.stmts[0].(*ast.AssignStmt)
	tmpVar2 := assign2.Lhs[0].(*ast.Ident).Name
	errVar2 := assign2.Lhs[1].(*ast.Ident).Name

	// Variables should be different
	if tmpVar1 == tmpVar2 {
		t.Errorf("Expected unique tmp variables, both got %q", tmpVar1)
	}

	if errVar1 == errVar2 {
		t.Errorf("Expected unique err variables, both got %q", errVar1)
	}

	// Check expected pattern
	if tmpVar1 != "__tmp0" {
		t.Errorf("Expected first tmp var '__tmp0', got %q", tmpVar1)
	}

	if errVar1 != "__err0" {
		t.Errorf("Expected first err var '__err0', got %q", errVar1)
	}

	if tmpVar2 != "__tmp1" {
		t.Errorf("Expected second tmp var '__tmp1', got %q", tmpVar2)
	}

	if errVar2 != "__err1" {
		t.Errorf("Expected second err var '__err1', got %q", errVar2)
	}
}

func TestReset(t *testing.T) {
	p := NewErrorPropagationPlugin()

	// Generate some variables
	p.nextTmpVar()
	p.nextErrVar()
	p.nextTmpVar()

	// Counters should be incremented
	if p.tmpVarCounter != 2 {
		t.Errorf("Expected tmpVarCounter to be 2, got %d", p.tmpVarCounter)
	}

	if p.errorVarCounter != 1 {
		t.Errorf("Expected errorVarCounter to be 1, got %d", p.errorVarCounter)
	}

	// Reset
	p.Reset()

	// Counters should be back to 0
	if p.tmpVarCounter != 0 {
		t.Errorf("Expected tmpVarCounter to be 0 after reset, got %d", p.tmpVarCounter)
	}

	if p.errorVarCounter != 0 {
		t.Errorf("Expected errorVarCounter to be 0 after reset, got %d", p.errorVarCounter)
	}

	// Next variables should start from 0 again
	tmpVar := p.nextTmpVar()
	errVar := p.nextErrVar()

	if tmpVar != "__tmp0" {
		t.Errorf("Expected '__tmp0' after reset, got %q", tmpVar)
	}

	if errVar != "__err0" {
		t.Errorf("Expected '__err0' after reset, got %q", errVar)
	}
}

func TestSyntaxAgnosticTransformation(t *testing.T) {
	tests := []struct {
		name   string
		syntax dingoast.SyntaxStyle
	}{
		{"question syntax", dingoast.SyntaxQuestion},
		{"bang syntax", dingoast.SyntaxBang},
		{"try syntax", dingoast.SyntaxTry},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewErrorPropagationPlugin()
			p.Reset() // Ensure counters start from same place

			ctx := &plugin.Context{
				FileSet: token.NewFileSet(),
			}

			errExpr := &dingoast.ErrorPropagationExpr{
				X:      &ast.CallExpr{Fun: ast.NewIdent("test")},
				OpPos:  token.NoPos,
				Syntax: tt.syntax,
			}

			result, err := p.Transform(ctx, errExpr)
			if err != nil {
				t.Fatalf("Transform() error = %v", err)
			}

			wrapper := result.(*temporaryStmtWrapper)

			// All syntaxes should produce identical structure
			if len(wrapper.stmts) != 2 {
				t.Errorf("Expected 2 statements for %s, got %d", tt.syntax, len(wrapper.stmts))
			}

			// Check we got assignment and if statement
			if _, ok := wrapper.stmts[0].(*ast.AssignStmt); !ok {
				t.Errorf("Expected assignment statement for %s", tt.syntax)
			}

			if _, ok := wrapper.stmts[1].(*ast.IfStmt); !ok {
				t.Errorf("Expected if statement for %s", tt.syntax)
			}
		})
	}
}

func TestTemporaryStmtWrapper(t *testing.T) {
	// Create statements - positions don't matter for this test
	stmts := []ast.Stmt{
		&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent("x")},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{ast.NewIdent("y")},
		},
		&ast.IfStmt{
			Cond: ast.NewIdent("true"),
			Body: &ast.BlockStmt{},
		},
	}

	wrapper := &temporaryStmtWrapper{
		stmts:  stmts,
		tmpVar: "__tmp0",
	}

	// Test Pos/End methods don't panic (positions may be invalid/zero)
	_ = wrapper.Pos()
	_ = wrapper.End()

	// Test tmpVar is stored correctly
	if wrapper.tmpVar != "__tmp0" {
		t.Errorf("Expected tmpVar '__tmp0', got %q", wrapper.tmpVar)
	}

	// Test stmts are stored correctly
	if len(wrapper.stmts) != 2 {
		t.Errorf("Expected 2 statements, got %d", len(wrapper.stmts))
	}
}

func TestNextVarHelpers(t *testing.T) {
	p := NewErrorPropagationPlugin()

	// Test tmp var generation
	vars := []string{
		p.nextTmpVar(),
		p.nextTmpVar(),
		p.nextTmpVar(),
	}

	expected := []string{"__tmp0", "__tmp1", "__tmp2"}
	for i, v := range vars {
		if v != expected[i] {
			t.Errorf("Expected tmp var %q, got %q", expected[i], v)
		}
	}

	// Reset and test err var generation
	p.Reset()

	errVars := []string{
		p.nextErrVar(),
		p.nextErrVar(),
		p.nextErrVar(),
	}

	expectedErr := []string{"__err0", "__err1", "__err2"}
	for i, v := range errVars {
		if v != expectedErr[i] {
			t.Errorf("Expected err var %q, got %q", expectedErr[i], v)
		}
	}
}

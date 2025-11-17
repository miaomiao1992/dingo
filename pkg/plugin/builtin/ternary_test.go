package builtin

import (
	"go/ast"
	"go/token"
	"testing"

	dingoast "github.com/MadAppGang/dingo/pkg/ast"
	"github.com/MadAppGang/dingo/pkg/config"
	"github.com/MadAppGang/dingo/pkg/plugin"
)

func TestNewTernaryPlugin(t *testing.T) {
	p := NewTernaryPlugin()

	if p == nil {
		t.Fatal("Expected plugin to be non-nil")
	}

	if p.Name() != "ternary" {
		t.Errorf("Expected name 'ternary', got %q", p.Name())
	}
}

func TestTernaryTransformNonTernaryNode(t *testing.T) {
	p := NewTernaryPlugin()
	ctx := &plugin.Context{
		FileSet: token.NewFileSet(),
	}

	// Create a regular if statement
	ifStmt := &ast.IfStmt{
		Cond: ast.NewIdent("condition"),
		Body: &ast.BlockStmt{},
	}

	// Transform should return the node unchanged
	result, err := p.Transform(ctx, ifStmt)
	if err != nil {
		t.Fatalf("Transform() error = %v", err)
	}

	if result != ifStmt {
		t.Error("Expected node to be returned unchanged for non-TernaryExpr")
	}
}

func TestTernaryTransformBasic(t *testing.T) {
	p := NewTernaryPlugin()

	cfg := config.DefaultConfig()
	ctx := &plugin.Context{
		FileSet:     token.NewFileSet(),
		DingoConfig: cfg,
	}

	// Create TernaryExpr: cond ? then : else
	ternaryExpr := &dingoast.TernaryExpr{
		Cond:     ast.NewIdent("condition"),
		Question: token.NoPos,
		Then:     &ast.BasicLit{Kind: token.STRING, Value: `"yes"`},
		Colon:    token.NoPos,
		Else:     &ast.BasicLit{Kind: token.STRING, Value: `"no"`},
	}

	// Transform
	result, err := p.Transform(ctx, ternaryExpr)
	if err != nil {
		t.Fatalf("Transform() error = %v", err)
	}

	// Result should be CallExpr (IIFE)
	callExpr, ok := result.(*ast.CallExpr)
	if !ok {
		t.Fatalf("Expected *ast.CallExpr (IIFE), got %T", result)
	}

	// Fun should be FuncLit
	funcLit, ok := callExpr.Fun.(*ast.FuncLit)
	if !ok {
		t.Fatalf("Expected FuncLit in CallExpr.Fun, got %T", callExpr.Fun)
	}

	// Function should have no parameters
	if funcLit.Type.Params == nil || len(funcLit.Type.Params.List) != 0 {
		t.Error("Expected function to have empty parameter list")
	}

	// Function body should have 2 statements: if and return
	if len(funcLit.Body.List) != 2 {
		t.Fatalf("Expected 2 statements in function body, got %d", len(funcLit.Body.List))
	}

	// First statement should be IfStmt
	ifStmt, ok := funcLit.Body.List[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("Expected first statement to be *ast.IfStmt, got %T", funcLit.Body.List[0])
	}

	// Check if condition matches
	condIdent, ok := ifStmt.Cond.(*ast.Ident)
	if !ok || condIdent.Name != "condition" {
		t.Error("Expected if condition to be 'condition'")
	}

	// If body should return then value
	if len(ifStmt.Body.List) != 1 {
		t.Fatalf("Expected 1 statement in if body, got %d", len(ifStmt.Body.List))
	}

	returnThen, ok := ifStmt.Body.List[0].(*ast.ReturnStmt)
	if !ok {
		t.Fatalf("Expected return statement in if body, got %T", ifStmt.Body.List[0])
	}

	if len(returnThen.Results) != 1 {
		t.Fatalf("Expected 1 return value, got %d", len(returnThen.Results))
	}

	thenLit, ok := returnThen.Results[0].(*ast.BasicLit)
	if !ok || thenLit.Value != `"yes"` {
		t.Error("Expected if body to return \"yes\"")
	}

	// Second statement should return else value
	returnElse, ok := funcLit.Body.List[1].(*ast.ReturnStmt)
	if !ok {
		t.Fatalf("Expected second statement to be *ast.ReturnStmt, got %T", funcLit.Body.List[1])
	}

	if len(returnElse.Results) != 1 {
		t.Fatalf("Expected 1 return value, got %d", len(returnElse.Results))
	}

	elseLit, ok := returnElse.Results[0].(*ast.BasicLit)
	if !ok || elseLit.Value != `"no"` {
		t.Error("Expected else return to be \"no\"")
	}
}

func TestTernaryTransformNested(t *testing.T) {
	p := NewTernaryPlugin()

	cfg := config.DefaultConfig()
	ctx := &plugin.Context{
		FileSet:     token.NewFileSet(),
		DingoConfig: cfg,
	}

	// Create nested TernaryExpr: cond1 ? (cond2 ? a : b) : c
	// Note: inner ternary won't work as ast.Expr because TernaryExpr is a Dingo node
	// This test verifies the transformation, but nested expressions need parser support
	outerTernary := &dingoast.TernaryExpr{
		Cond:     ast.NewIdent("cond1"),
		Question: token.NoPos,
		Then:     ast.NewIdent("innerResult"),
		Colon:    token.NoPos,
		Else:     ast.NewIdent("c"),
	}

	// Transform outer ternary
	result, err := p.Transform(ctx, outerTernary)
	if err != nil {
		t.Fatalf("Transform() error = %v", err)
	}

	// Result should be CallExpr (IIFE)
	callExpr, ok := result.(*ast.CallExpr)
	if !ok {
		t.Fatalf("Expected *ast.CallExpr (IIFE), got %T", result)
	}

	funcLit, ok := callExpr.Fun.(*ast.FuncLit)
	if !ok {
		t.Fatalf("Expected FuncLit in CallExpr.Fun, got %T", callExpr.Fun)
	}

	// Check that the then branch was created correctly
	ifStmt, ok := funcLit.Body.List[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("Expected IfStmt in body, got %T", funcLit.Body.List[0])
	}

	returnStmt, ok := ifStmt.Body.List[0].(*ast.ReturnStmt)
	if !ok {
		t.Fatalf("Expected ReturnStmt in if body, got %T", ifStmt.Body.List[0])
	}

	// The return should contain the then value
	if resultIdent, ok := returnStmt.Results[0].(*ast.Ident); !ok || resultIdent.Name != "innerResult" {
		t.Error("Expected then branch to return innerResult")
	}
}

func TestTernaryStandardPrecedence(t *testing.T) {
	p := NewTernaryPlugin()

	cfg := config.DefaultConfig()
	cfg.Features.OperatorPrecedence = "standard"

	ctx := &plugin.Context{
		FileSet:     token.NewFileSet(),
		DingoConfig: cfg,
	}

	ternaryExpr := &dingoast.TernaryExpr{
		Cond:     ast.NewIdent("condition"),
		Question: token.NoPos,
		Then:     ast.NewIdent("a"),
		Colon:    token.NoPos,
		Else:     ast.NewIdent("b"),
	}

	// Transform should succeed (precedence checking not enforced in plugin)
	result, err := p.Transform(ctx, ternaryExpr)
	if err != nil {
		t.Fatalf("Transform() error = %v", err)
	}

	if _, ok := result.(*ast.CallExpr); !ok {
		t.Fatalf("Expected *ast.CallExpr (IIFE), got %T", result)
	}
}

func TestTernaryExplicitPrecedence(t *testing.T) {
	p := NewTernaryPlugin()

	cfg := config.DefaultConfig()
	cfg.Features.OperatorPrecedence = "explicit"

	ctx := &plugin.Context{
		FileSet:     token.NewFileSet(),
		DingoConfig: cfg,
	}

	ternaryExpr := &dingoast.TernaryExpr{
		Cond:     ast.NewIdent("condition"),
		Question: token.NoPos,
		Then:     ast.NewIdent("a"),
		Colon:    token.NoPos,
		Else:     ast.NewIdent("b"),
	}

	// Transform should succeed (precedence validation is parser's job)
	result, err := p.Transform(ctx, ternaryExpr)
	if err != nil {
		t.Fatalf("Transform() error = %v", err)
	}

	if _, ok := result.(*ast.CallExpr); !ok {
		t.Fatalf("Expected *ast.CallExpr (IIFE), got %T", result)
	}
}

func TestTernaryNilConfig(t *testing.T) {
	p := NewTernaryPlugin()

	// Context with nil config should default to standard precedence
	ctx := &plugin.Context{
		FileSet:     token.NewFileSet(),
		DingoConfig: nil,
	}

	ternaryExpr := &dingoast.TernaryExpr{
		Cond:     ast.NewIdent("condition"),
		Question: token.NoPos,
		Then:     ast.NewIdent("a"),
		Colon:    token.NoPos,
		Else:     ast.NewIdent("b"),
	}

	// Transform should succeed with default behavior
	result, err := p.Transform(ctx, ternaryExpr)
	if err != nil {
		t.Fatalf("Transform() error = %v", err)
	}

	if _, ok := result.(*ast.CallExpr); !ok {
		t.Fatalf("Expected *ast.CallExpr (IIFE), got %T", result)
	}
}

func TestTernaryTransformToIfStmt(t *testing.T) {
	p := NewTernaryPlugin()

	ternaryExpr := &dingoast.TernaryExpr{
		Cond:     ast.NewIdent("condition"),
		Question: token.NoPos,
		Then:     ast.NewIdent("doThis"),
		Colon:    token.NoPos,
		Else:     ast.NewIdent("doThat"),
	}

	// Call the transformToIfStmt method directly
	result, err := p.transformToIfStmt(ternaryExpr)
	if err != nil {
		t.Fatalf("transformToIfStmt() error = %v", err)
	}

	// Result should be IfStmt
	ifStmt, ok := result.(*ast.IfStmt)
	if !ok {
		t.Fatalf("Expected *ast.IfStmt, got %T", result)
	}

	// Check condition
	condIdent, ok := ifStmt.Cond.(*ast.Ident)
	if !ok || condIdent.Name != "condition" {
		t.Error("Expected if condition to be 'condition'")
	}

	// Check then body
	if len(ifStmt.Body.List) != 1 {
		t.Fatalf("Expected 1 statement in then body, got %d", len(ifStmt.Body.List))
	}

	thenExpr, ok := ifStmt.Body.List[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("Expected ExprStmt in then body, got %T", ifStmt.Body.List[0])
	}

	thenIdent, ok := thenExpr.X.(*ast.Ident)
	if !ok || thenIdent.Name != "doThis" {
		t.Error("Expected then body to contain 'doThis'")
	}

	// Check else body
	elseBlock, ok := ifStmt.Else.(*ast.BlockStmt)
	if !ok {
		t.Fatalf("Expected BlockStmt in else, got %T", ifStmt.Else)
	}

	if len(elseBlock.List) != 1 {
		t.Fatalf("Expected 1 statement in else body, got %d", len(elseBlock.List))
	}

	elseExpr, ok := elseBlock.List[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("Expected ExprStmt in else body, got %T", elseBlock.List[0])
	}

	elseIdent, ok := elseExpr.X.(*ast.Ident)
	if !ok || elseIdent.Name != "doThat" {
		t.Error("Expected else body to contain 'doThat'")
	}
}

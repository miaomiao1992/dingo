package builtin

import (
	"go/ast"
	"go/token"
	"testing"

	dingoast "github.com/MadAppGang/dingo/pkg/ast"
	"github.com/MadAppGang/dingo/pkg/config"
	"github.com/MadAppGang/dingo/pkg/plugin"
)

func TestNewSafeNavigationPlugin(t *testing.T) {
	p := NewSafeNavigationPlugin()

	if p == nil {
		t.Fatal("Expected plugin to be non-nil")
	}

	if p.Name() != "safe_navigation" {
		t.Errorf("Expected name 'safe_navigation', got %q", p.Name())
	}
}

func TestSafeNavTransformNonSafeNavNode(t *testing.T) {
	p := NewSafeNavigationPlugin()
	ctx := &plugin.Context{
		FileSet: token.NewFileSet(),
	}

	// Create a regular selector expression
	selExpr := &ast.SelectorExpr{
		X:   ast.NewIdent("user"),
		Sel: ast.NewIdent("Name"),
	}

	// Transform should return the node unchanged
	result, err := p.Transform(ctx, selExpr)
	if err != nil {
		t.Fatalf("Transform() error = %v", err)
	}

	if result != selExpr {
		t.Error("Expected node to be returned unchanged for non-SafeNavigationExpr")
	}
}

func TestSafeNavTransformSmartMode(t *testing.T) {
	p := NewSafeNavigationPlugin()

	cfg := config.DefaultConfig()
	cfg.Features.SafeNavigationUnwrap = "smart"

	ctx := &plugin.Context{
		FileSet:     token.NewFileSet(),
		DingoConfig: cfg,
	}

	// Create SafeNavigationExpr: user?.name
	safeNavExpr := &dingoast.SafeNavigationExpr{
		X:     ast.NewIdent("user"),
		OpPos: token.NoPos,
		Sel:   ast.NewIdent("Name"),
	}

	// Transform
	result, err := p.Transform(ctx, safeNavExpr)
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

	// Check if condition is `user != nil`
	binExpr, ok := ifStmt.Cond.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("Expected binary expression in if condition, got %T", ifStmt.Cond)
	}

	if binExpr.Op != token.NEQ {
		t.Errorf("Expected != operator, got %v", binExpr.Op)
	}

	xIdent, ok := binExpr.X.(*ast.Ident)
	if !ok || xIdent.Name != "user" {
		t.Error("Expected condition to check 'user'")
	}

	yIdent, ok := binExpr.Y.(*ast.Ident)
	if !ok || yIdent.Name != "nil" {
		t.Error("Expected condition to check against 'nil'")
	}

	// If body should return user.Name
	if len(ifStmt.Body.List) != 1 {
		t.Fatalf("Expected 1 statement in if body, got %d", len(ifStmt.Body.List))
	}

	returnStmt, ok := ifStmt.Body.List[0].(*ast.ReturnStmt)
	if !ok {
		t.Fatalf("Expected return statement in if body, got %T", ifStmt.Body.List[0])
	}

	if len(returnStmt.Results) != 1 {
		t.Fatalf("Expected 1 return value, got %d", len(returnStmt.Results))
	}

	selExpr, ok := returnStmt.Results[0].(*ast.SelectorExpr)
	if !ok {
		t.Fatalf("Expected selector expression in return, got %T", returnStmt.Results[0])
	}

	if selExpr.Sel.Name != "Name" {
		t.Error("Expected return to access 'Name' field")
	}

	// Second statement should be return nil (zero value)
	returnZero, ok := funcLit.Body.List[1].(*ast.ReturnStmt)
	if !ok {
		t.Fatalf("Expected second statement to be *ast.ReturnStmt, got %T", funcLit.Body.List[1])
	}

	if len(returnZero.Results) != 1 {
		t.Fatalf("Expected 1 return value, got %d", len(returnZero.Results))
	}

	if nilIdent, ok := returnZero.Results[0].(*ast.Ident); !ok || nilIdent.Name != "nil" {
		t.Error("Expected zero value return to be 'nil'")
	}
}

func TestSafeNavTransformAlwaysOptionMode(t *testing.T) {
	p := NewSafeNavigationPlugin()

	cfg := config.DefaultConfig()
	cfg.Features.SafeNavigationUnwrap = "always_option"

	ctx := &plugin.Context{
		FileSet:     token.NewFileSet(),
		DingoConfig: cfg,
	}

	// Create SafeNavigationExpr: user?.name
	safeNavExpr := &dingoast.SafeNavigationExpr{
		X:     ast.NewIdent("user"),
		OpPos: token.NoPos,
		Sel:   ast.NewIdent("Name"),
	}

	// Transform
	result, err := p.Transform(ctx, safeNavExpr)
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

	// Function should have return type Option_T
	if funcLit.Type.Results == nil || len(funcLit.Type.Results.List) != 1 {
		t.Fatal("Expected function to have 1 return type")
	}

	retType, ok := funcLit.Type.Results.List[0].Type.(*ast.Ident)
	if !ok || retType.Name != "Option_T" {
		t.Error("Expected return type to be Option_T")
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

	// If body should return Option_Some(user.Name)
	if len(ifStmt.Body.List) != 1 {
		t.Fatalf("Expected 1 statement in if body, got %d", len(ifStmt.Body.List))
	}

	returnSome, ok := ifStmt.Body.List[0].(*ast.ReturnStmt)
	if !ok {
		t.Fatalf("Expected return statement in if body, got %T", ifStmt.Body.List[0])
	}

	if len(returnSome.Results) != 1 {
		t.Fatalf("Expected 1 return value, got %d", len(returnSome.Results))
	}

	someCall, ok := returnSome.Results[0].(*ast.CallExpr)
	if !ok {
		t.Fatalf("Expected call expression in return, got %T", returnSome.Results[0])
	}

	someFun, ok := someCall.Fun.(*ast.Ident)
	if !ok || someFun.Name != "Option_Some" {
		t.Error("Expected call to Option_Some")
	}

	// Second statement should return Option_None
	returnNone, ok := funcLit.Body.List[1].(*ast.ReturnStmt)
	if !ok {
		t.Fatalf("Expected second statement to be *ast.ReturnStmt, got %T", funcLit.Body.List[1])
	}

	noneCall, ok := returnNone.Results[0].(*ast.CallExpr)
	if !ok {
		t.Fatalf("Expected call expression in return, got %T", returnNone.Results[0])
	}

	noneFun, ok := noneCall.Fun.(*ast.Ident)
	if !ok || noneFun.Name != "Option_None" {
		t.Error("Expected call to Option_None")
	}
}

func TestSafeNavInvalidConfig(t *testing.T) {
	p := NewSafeNavigationPlugin()

	cfg := config.DefaultConfig()
	cfg.Features.SafeNavigationUnwrap = "invalid_mode"

	ctx := &plugin.Context{
		FileSet:     token.NewFileSet(),
		DingoConfig: cfg,
	}

	safeNavExpr := &dingoast.SafeNavigationExpr{
		X:     ast.NewIdent("user"),
		OpPos: token.NoPos,
		Sel:   ast.NewIdent("Name"),
	}

	// Transform should return error
	_, err := p.Transform(ctx, safeNavExpr)
	if err == nil {
		t.Fatal("Expected error for invalid configuration, got nil")
	}

	expectedMsg := "invalid safe_navigation_unwrap mode"
	if !contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain %q, got %q", expectedMsg, err.Error())
	}
}

func TestSafeNavNilConfig(t *testing.T) {
	p := NewSafeNavigationPlugin()

	// Context with nil config should default to smart mode
	ctx := &plugin.Context{
		FileSet:     token.NewFileSet(),
		DingoConfig: nil,
	}

	safeNavExpr := &dingoast.SafeNavigationExpr{
		X:     ast.NewIdent("user"),
		OpPos: token.NoPos,
		Sel:   ast.NewIdent("Name"),
	}

	// Transform should succeed with default behavior
	result, err := p.Transform(ctx, safeNavExpr)
	if err != nil {
		t.Fatalf("Transform() error = %v", err)
	}

	// Should generate IIFE (smart mode default)
	if _, ok := result.(*ast.CallExpr); !ok {
		t.Fatalf("Expected *ast.CallExpr (IIFE), got %T", result)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

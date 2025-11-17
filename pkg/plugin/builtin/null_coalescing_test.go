package builtin

import (
	"go/ast"
	"go/token"
	"go/types"
	"testing"

	dingoast "github.com/MadAppGang/dingo/pkg/ast"
	"github.com/MadAppGang/dingo/pkg/config"
	"github.com/MadAppGang/dingo/pkg/plugin"
)

func TestNewNullCoalescingPlugin(t *testing.T) {
	p := NewNullCoalescingPlugin()

	if p == nil {
		t.Fatal("Expected plugin to be non-nil")
	}

	if p.Name() != "null_coalescing" {
		t.Errorf("Expected name 'null_coalescing', got %q", p.Name())
	}
}

func TestNullCoalesceTransformNonNullCoalesceNode(t *testing.T) {
	p := NewNullCoalescingPlugin()
	ctx := &plugin.Context{
		FileSet: token.NewFileSet(),
	}

	// Create a regular binary expression
	binExpr := &ast.BinaryExpr{
		X:  ast.NewIdent("a"),
		Op: token.ADD,
		Y:  ast.NewIdent("b"),
	}

	// Transform should return the node unchanged
	result, err := p.Transform(ctx, binExpr)
	if err != nil {
		t.Fatalf("Transform() error = %v", err)
	}

	if result != binExpr {
		t.Error("Expected node to be returned unchanged for non-NullCoalescingExpr")
	}
}

func TestNullCoalesceTransformOptionType(t *testing.T) {
	p := NewNullCoalescingPlugin()

	cfg := config.DefaultConfig()
	ctx := &plugin.Context{
		FileSet:     token.NewFileSet(),
		DingoConfig: cfg,
	}

	// Create NullCoalescingExpr: opt ?? default
	ncExpr := &dingoast.NullCoalescingExpr{
		X:     ast.NewIdent("opt"),
		OpPos: token.NoPos,
		Y:     ast.NewIdent("defaultValue"),
	}

	// Transform (without type info, falls back to Option transformation)
	result, err := p.Transform(ctx, ncExpr)
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

	// Function body should have 2 statements: if and return
	if len(funcLit.Body.List) != 2 {
		t.Fatalf("Expected 2 statements in function body, got %d", len(funcLit.Body.List))
	}

	// First statement should be IfStmt
	ifStmt, ok := funcLit.Body.List[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("Expected first statement to be *ast.IfStmt, got %T", funcLit.Body.List[0])
	}

	// Check if condition is opt.IsSome()
	condCall, ok := ifStmt.Cond.(*ast.CallExpr)
	if !ok {
		t.Fatalf("Expected call expression in if condition, got %T", ifStmt.Cond)
	}

	selExpr, ok := condCall.Fun.(*ast.SelectorExpr)
	if !ok {
		t.Fatalf("Expected selector expression in condition call, got %T", condCall.Fun)
	}

	if selExpr.Sel.Name != "IsSome" {
		t.Error("Expected condition to call IsSome()")
	}

	// If body should return opt.Unwrap()
	if len(ifStmt.Body.List) != 1 {
		t.Fatalf("Expected 1 statement in if body, got %d", len(ifStmt.Body.List))
	}

	returnStmt, ok := ifStmt.Body.List[0].(*ast.ReturnStmt)
	if !ok {
		t.Fatalf("Expected return statement in if body, got %T", ifStmt.Body.List[0])
	}

	unwrapCall, ok := returnStmt.Results[0].(*ast.CallExpr)
	if !ok {
		t.Fatalf("Expected call expression in return, got %T", returnStmt.Results[0])
	}

	unwrapSel, ok := unwrapCall.Fun.(*ast.SelectorExpr)
	if !ok {
		t.Fatalf("Expected selector expression in unwrap call, got %T", unwrapCall.Fun)
	}

	if unwrapSel.Sel.Name != "Unwrap" {
		t.Error("Expected return to call Unwrap()")
	}

	// Second statement should return default
	returnDefault, ok := funcLit.Body.List[1].(*ast.ReturnStmt)
	if !ok {
		t.Fatalf("Expected second statement to be *ast.ReturnStmt, got %T", funcLit.Body.List[1])
	}

	if defaultIdent, ok := returnDefault.Results[0].(*ast.Ident); !ok || defaultIdent.Name != "defaultValue" {
		t.Error("Expected return to use default value")
	}
}

func TestNullCoalesceTransformPointerEnabled(t *testing.T) {
	p := NewNullCoalescingPlugin()

	cfg := config.DefaultConfig()
	cfg.Features.NullCoalescingPointers = true

	// Create mock type info with pointer type
	typeInfo := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
	}

	ptrIdent := ast.NewIdent("ptr")
	// Create pointer type: *string
	typeInfo.Types[ptrIdent] = types.TypeAndValue{
		Type: types.NewPointer(types.Typ[types.String]),
	}

	ctx := &plugin.Context{
		FileSet:     token.NewFileSet(),
		DingoConfig: cfg,
		TypeInfo:    typeInfo,
	}

	// Create NullCoalescingExpr: ptr ?? default
	ncExpr := &dingoast.NullCoalescingExpr{
		X:     ptrIdent,
		OpPos: token.NoPos,
		Y:     &ast.BasicLit{Kind: token.STRING, Value: `"default"`},
	}

	// Transform
	result, err := p.Transform(ctx, ncExpr)
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

	// Function body should have 2 statements: if and return
	if len(funcLit.Body.List) != 2 {
		t.Fatalf("Expected 2 statements in function body, got %d", len(funcLit.Body.List))
	}

	// First statement should be IfStmt
	ifStmt, ok := funcLit.Body.List[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("Expected first statement to be *ast.IfStmt, got %T", funcLit.Body.List[0])
	}

	// Check if condition is ptr != nil
	binExpr, ok := ifStmt.Cond.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("Expected binary expression in if condition, got %T", ifStmt.Cond)
	}

	if binExpr.Op != token.NEQ {
		t.Errorf("Expected != operator, got %v", binExpr.Op)
	}

	// If body should return *ptr
	if len(ifStmt.Body.List) != 1 {
		t.Fatalf("Expected 1 statement in if body, got %d", len(ifStmt.Body.List))
	}

	returnStmt, ok := ifStmt.Body.List[0].(*ast.ReturnStmt)
	if !ok {
		t.Fatalf("Expected return statement in if body, got %T", ifStmt.Body.List[0])
	}

	starExpr, ok := returnStmt.Results[0].(*ast.StarExpr)
	if !ok {
		t.Fatalf("Expected star expression (dereference) in return, got %T", returnStmt.Results[0])
	}

	if derefIdent, ok := starExpr.X.(*ast.Ident); !ok || derefIdent.Name != "ptr" {
		t.Error("Expected return to dereference 'ptr'")
	}
}

func TestNullCoalesceTransformPointerDisabled(t *testing.T) {
	p := NewNullCoalescingPlugin()

	cfg := config.DefaultConfig()
	cfg.Features.NullCoalescingPointers = false

	// Create mock type info with pointer type
	typeInfo := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
	}

	ptrIdent := ast.NewIdent("ptr")
	// Create pointer type: *string
	typeInfo.Types[ptrIdent] = types.TypeAndValue{
		Type: types.NewPointer(types.Typ[types.String]),
	}

	ctx := &plugin.Context{
		FileSet:     token.NewFileSet(),
		DingoConfig: cfg,
		TypeInfo:    typeInfo,
	}

	// Create NullCoalescingExpr: ptr ?? default
	ncExpr := &dingoast.NullCoalescingExpr{
		X:     ptrIdent,
		OpPos: token.NoPos,
		Y:     &ast.BasicLit{Kind: token.STRING, Value: `"default"`},
	}

	// Transform should fall back to Option transformation
	result, err := p.Transform(ctx, ncExpr)
	if err != nil {
		t.Fatalf("Transform() error = %v", err)
	}

	// Should generate Option-style transformation (IsSome/Unwrap)
	callExpr, ok := result.(*ast.CallExpr)
	if !ok {
		t.Fatalf("Expected *ast.CallExpr (IIFE), got %T", result)
	}

	funcLit, ok := callExpr.Fun.(*ast.FuncLit)
	if !ok {
		t.Fatalf("Expected FuncLit in CallExpr.Fun, got %T", callExpr.Fun)
	}

	ifStmt, ok := funcLit.Body.List[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("Expected IfStmt in body, got %T", funcLit.Body.List[0])
	}

	// Condition should be IsSome() call, not nil check
	if condCall, ok := ifStmt.Cond.(*ast.CallExpr); ok {
		if sel, ok := condCall.Fun.(*ast.SelectorExpr); ok {
			if sel.Sel.Name == "IsSome" {
				// Correct: fell back to Option transformation
				return
			}
		}
	}

	t.Error("Expected fallback to Option transformation when pointer support disabled")
}

func TestNullCoalesceNoTypeInfo(t *testing.T) {
	p := NewNullCoalescingPlugin()

	cfg := config.DefaultConfig()
	ctx := &plugin.Context{
		FileSet:     token.NewFileSet(),
		DingoConfig: cfg,
		TypeInfo:    nil, // No type info
	}

	ncExpr := &dingoast.NullCoalescingExpr{
		X:     ast.NewIdent("value"),
		OpPos: token.NoPos,
		Y:     ast.NewIdent("default"),
	}

	// Transform should succeed with default Option behavior
	result, err := p.Transform(ctx, ncExpr)
	if err != nil {
		t.Fatalf("Transform() error = %v", err)
	}

	// Should generate IIFE
	if _, ok := result.(*ast.CallExpr); !ok {
		t.Fatalf("Expected *ast.CallExpr (IIFE), got %T", result)
	}
}

func TestNullCoalesceIsOptionType(t *testing.T) {
	p := NewNullCoalescingPlugin()

	tests := []struct {
		name     string
		typeName string
		want     bool
	}{
		{"Option_string", "Option_string", true},
		{"Option_User", "Option_User", true},
		{"Option_int", "Option_int", true},
		{"NotOption", "NotOption", false},
		{"Option", "Option", false}, // Too short
		{"string", "string", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a named type
			obj := types.NewTypeName(token.NoPos, nil, tt.typeName, nil)
			named := types.NewNamed(obj, types.Typ[types.Int], nil)

			got := p.isOptionType(named)
			if got != tt.want {
				t.Errorf("isOptionType(%q) = %v, want %v", tt.typeName, got, tt.want)
			}
		})
	}
}

func TestNullCoalesceIsPointerType(t *testing.T) {
	p := NewNullCoalescingPlugin()

	// Test pointer type
	ptrType := types.NewPointer(types.Typ[types.String])
	if !p.isPointerType(ptrType) {
		t.Error("Expected isPointerType to return true for pointer type")
	}

	// Test non-pointer type
	nonPtrType := types.Typ[types.String]
	if p.isPointerType(nonPtrType) {
		t.Error("Expected isPointerType to return false for non-pointer type")
	}

	// Test nil type
	if p.isPointerType(nil) {
		t.Error("Expected isPointerType to return false for nil type")
	}
}

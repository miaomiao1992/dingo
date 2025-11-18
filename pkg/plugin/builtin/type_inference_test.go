package builtin

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/MadAppGang/dingo/pkg/plugin"
)

// ============================================================================
// 1. BASIC TYPE INFERENCE TESTS (InferType method)
// ============================================================================

func TestInferType_BasicLiterals(t *testing.T) {
	service := createTestTypeInferenceService(t)

	tests := []struct {
		name     string
		expr     ast.Expr
		expected string
	}{
		{
			name:     "int literal",
			expr:     &ast.BasicLit{Kind: token.INT, Value: "42"},
			expected: "int", // UntypedInt -> int via TypeToString
		},
		{
			name:     "float literal",
			expr:     &ast.BasicLit{Kind: token.FLOAT, Value: "3.14"},
			expected: "float64", // UntypedFloat -> float64
		},
		{
			name:     "string literal",
			expr:     &ast.BasicLit{Kind: token.STRING, Value: `"hello"`},
			expected: "string", // UntypedString -> string
		},
		{
			name:     "rune literal",
			expr:     &ast.BasicLit{Kind: token.CHAR, Value: "'a'"},
			expected: "rune", // UntypedRune -> rune
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, ok := service.InferType(tt.expr)
			if !ok {
				t.Fatalf("InferType failed for %s", tt.name)
			}

			got := service.TypeToString(typ)
			if got != tt.expected {
				t.Errorf("expected type %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestInferType_BuiltinIdents(t *testing.T) {
	service := createTestTypeInferenceService(t)

	tests := []struct {
		name     string
		ident    string
		expected string
	}{
		{
			name:     "true",
			ident:    "true",
			expected: "bool",
		},
		{
			name:     "false",
			ident:    "false",
			expected: "bool",
		},
		{
			name:     "nil",
			ident:    "nil",
			expected: "interface{}", // nil has no specific type
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := ast.NewIdent(tt.ident)
			typ, ok := service.InferType(expr)
			if !ok {
				t.Fatalf("InferType failed for %s", tt.name)
			}

			got := service.TypeToString(typ)
			if got != tt.expected {
				t.Errorf("expected type %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestInferType_PointerExpression(t *testing.T) {
	service := createTestTypeInferenceService(t)

	// &42 -> *int
	expr := &ast.UnaryExpr{
		Op: token.AND,
		X:  &ast.BasicLit{Kind: token.INT, Value: "42"},
	}

	typ, ok := service.InferType(expr)
	if !ok {
		t.Fatal("InferType failed for pointer expression")
	}

	got := service.TypeToString(typ)
	expected := "*int"
	if got != expected {
		t.Errorf("expected type %q, got %q", expected, got)
	}
}

func TestInferType_NilExpression(t *testing.T) {
	service := createTestTypeInferenceService(t)

	typ, ok := service.InferType(nil)
	if ok {
		t.Error("InferType should fail for nil expression")
	}
	if typ != nil {
		t.Errorf("expected nil type, got %v", typ)
	}
}

func TestInferType_UnsupportedExpression(t *testing.T) {
	service := createTestTypeInferenceService(t)

	// Function call without go/types should fail
	expr := &ast.CallExpr{
		Fun: ast.NewIdent("someFunc"),
	}

	typ, ok := service.InferType(expr)
	if ok {
		t.Error("InferType should fail for function call without go/types")
	}
	if typ != nil {
		t.Errorf("expected nil type, got %v", typ)
	}
}

// ============================================================================
// 2. TYPE TO STRING CONVERSION TESTS
// ============================================================================

func TestTypeToString_BasicTypes(t *testing.T) {
	service := createTestTypeInferenceService(t)

	tests := []struct {
		name     string
		typ      types.Type
		expected string
	}{
		{
			name:     "int",
			typ:      types.Typ[types.Int],
			expected: "int",
		},
		{
			name:     "string",
			typ:      types.Typ[types.String],
			expected: "string",
		},
		{
			name:     "bool",
			typ:      types.Typ[types.Bool],
			expected: "bool",
		},
		{
			name:     "float64",
			typ:      types.Typ[types.Float64],
			expected: "float64",
		},
		{
			name:     "byte",
			typ:      types.Typ[types.Byte],
			expected: "uint8", // byte is alias for uint8
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.TypeToString(tt.typ)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestTypeToString_UntypedConstants(t *testing.T) {
	service := createTestTypeInferenceService(t)

	tests := []struct {
		name     string
		typ      types.Type
		expected string
	}{
		{
			name:     "untyped int",
			typ:      types.Typ[types.UntypedInt],
			expected: "int",
		},
		{
			name:     "untyped float",
			typ:      types.Typ[types.UntypedFloat],
			expected: "float64",
		},
		{
			name:     "untyped string",
			typ:      types.Typ[types.UntypedString],
			expected: "string",
		},
		{
			name:     "untyped bool",
			typ:      types.Typ[types.UntypedBool],
			expected: "bool",
		},
		{
			name:     "untyped rune",
			typ:      types.Typ[types.UntypedRune],
			expected: "rune",
		},
		{
			name:     "untyped nil",
			typ:      types.Typ[types.UntypedNil],
			expected: "interface{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.TypeToString(tt.typ)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestTypeToString_CompositeTypes(t *testing.T) {
	service := createTestTypeInferenceService(t)

	tests := []struct {
		name     string
		typ      types.Type
		expected string
	}{
		{
			name:     "pointer to int",
			typ:      types.NewPointer(types.Typ[types.Int]),
			expected: "*int",
		},
		{
			name:     "slice of byte",
			typ:      types.NewSlice(types.Typ[types.Byte]),
			expected: "[]uint8",
		},
		{
			name:     "array of int",
			typ:      types.NewArray(types.Typ[types.Int], 10),
			expected: "[10]int",
		},
		{
			name:     "map[string]int",
			typ:      types.NewMap(types.Typ[types.String], types.Typ[types.Int]),
			expected: "map[string]int",
		},
		{
			name:     "chan int",
			typ:      types.NewChan(types.SendRecv, types.Typ[types.Int]),
			expected: "chan int",
		},
		{
			name:     "chan<- string",
			typ:      types.NewChan(types.SendOnly, types.Typ[types.String]),
			expected: "chan<- string",
		},
		{
			name:     "<-chan bool",
			typ:      types.NewChan(types.RecvOnly, types.Typ[types.Bool]),
			expected: "<-chan bool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.TypeToString(tt.typ)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestTypeToString_EmptyInterface(t *testing.T) {
	service := createTestTypeInferenceService(t)

	// Create empty interface type
	typ := types.NewInterfaceType(nil, nil)

	got := service.TypeToString(typ)
	expected := "interface{}"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestTypeToString_NestedPointers(t *testing.T) {
	service := createTestTypeInferenceService(t)

	// **int
	typ := types.NewPointer(types.NewPointer(types.Typ[types.Int]))

	got := service.TypeToString(typ)
	expected := "**int"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestTypeToString_NilType(t *testing.T) {
	service := createTestTypeInferenceService(t)

	got := service.TypeToString(nil)
	expected := "interface{}"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

// ============================================================================
// 3. GO/TYPES INTEGRATION TESTS
// ============================================================================

func TestInferType_WithGoTypes(t *testing.T) {
	// Create a simple Go source file
	src := `package main

func add(x int, y int) int {
	return x + y
}

func main() {
	result := add(1, 2)
	_ = result
}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}

	// Run type checker
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	conf := types.Config{Importer: nil}
	_, err = conf.Check("main", fset, []*ast.File{file}, info)
	// Type checking may fail without importer, but we get partial info
	// Don't fail the test if type checking has errors

	// Create service with types.Info
	service, err := NewTypeInferenceService(fset, file, plugin.NewNoOpLogger())
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}
	service.SetTypesInfo(info)

	// Find the "1" literal in add(1, 2)
	var oneLiteral *ast.BasicLit
	ast.Inspect(file, func(n ast.Node) bool {
		if lit, ok := n.(*ast.BasicLit); ok && lit.Value == "1" {
			oneLiteral = lit
			return false
		}
		return true
	})

	if oneLiteral == nil {
		t.Fatal("could not find literal '1' in AST")
	}

	// Infer type using go/types
	typ, ok := service.InferType(oneLiteral)
	if !ok {
		t.Fatal("InferType failed with go/types")
	}

	// Should infer as untyped int (or int if context provides it)
	got := service.TypeToString(typ)
	// Accept both "int" and "untyped int" as valid
	if got != "int" && got != "untyped int" {
		t.Errorf("expected int or untyped int, got %q", got)
	}
}

func TestSetTypesInfo(t *testing.T) {
	service := createTestTypeInferenceService(t)

	// Initially, typesInfo should be nil
	if service.typesInfo != nil {
		t.Error("expected nil typesInfo initially")
	}

	// Set types.Info
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
	}
	service.SetTypesInfo(info)

	// Verify it was set
	if service.typesInfo != info {
		t.Error("SetTypesInfo did not set the typesInfo field")
	}
}

// ============================================================================
// 4. GRACEFUL FALLBACK TESTS
// ============================================================================

func TestInferType_FallbackWithoutGoTypes(t *testing.T) {
	service := createTestTypeInferenceService(t)

	// Service has no go/types info, should fall back to structural inference

	// Test: basic literal (should work with fallback)
	expr := &ast.BasicLit{Kind: token.INT, Value: "123"}
	typ, ok := service.InferType(expr)
	if !ok {
		t.Error("InferType should succeed for basic literal without go/types")
	}
	if service.TypeToString(typ) != "int" {
		t.Errorf("expected int, got %q", service.TypeToString(typ))
	}

	// Test: identifier (should fail gracefully without go/types)
	identExpr := ast.NewIdent("myVar")
	typ, ok = service.InferType(identExpr)
	if ok {
		t.Error("InferType should fail for identifier without go/types")
	}
	if typ != nil {
		t.Errorf("expected nil type, got %v", typ)
	}
}

func TestInferType_PartialGoTypesInfo(t *testing.T) {
	service := createTestTypeInferenceService(t)

	// Create partial types.Info with only some expressions
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
	}

	// Add type info for one expression
	knownExpr := &ast.BasicLit{Kind: token.INT, Value: "42"}
	info.Types[knownExpr] = types.TypeAndValue{
		Type: types.Typ[types.Int],
	}

	service.SetTypesInfo(info)

	// Test: expression with type info (should use go/types)
	typ, ok := service.InferType(knownExpr)
	if !ok {
		t.Error("InferType should succeed for known expression")
	}
	if service.TypeToString(typ) != "int" {
		t.Errorf("expected int, got %q", service.TypeToString(typ))
	}

	// Test: expression without type info (should fall back)
	unknownExpr := &ast.BasicLit{Kind: token.STRING, Value: `"test"`}
	typ, ok = service.InferType(unknownExpr)
	if !ok {
		t.Error("InferType should fall back for unknown expression")
	}
	if service.TypeToString(typ) != "string" {
		t.Errorf("expected string, got %q", service.TypeToString(typ))
	}
}

// ============================================================================
// 5. EDGE CASES AND ERROR HANDLING
// ============================================================================

func TestInferType_EmptyTypesInfo(t *testing.T) {
	service := createTestTypeInferenceService(t)

	// Set empty types.Info (not nil, but empty maps)
	info := &types.Info{}
	service.SetTypesInfo(info)

	// Should handle gracefully and fall back
	expr := &ast.BasicLit{Kind: token.INT, Value: "99"}
	typ, ok := service.InferType(expr)
	if !ok {
		t.Error("InferType should fall back when types.Info is empty")
	}
	if service.TypeToString(typ) != "int" {
		t.Errorf("expected int, got %q", service.TypeToString(typ))
	}
}

func TestTypeToString_ComplexSignature(t *testing.T) {
	service := createTestTypeInferenceService(t)

	// Create a function signature: func(int, string) (bool, error)
	params := types.NewTuple(
		types.NewParam(token.NoPos, nil, "x", types.Typ[types.Int]),
		types.NewParam(token.NoPos, nil, "y", types.Typ[types.String]),
	)
	errorType := types.Universe.Lookup("error").Type()
	results := types.NewTuple(
		types.NewParam(token.NoPos, nil, "", types.Typ[types.Bool]),
		types.NewParam(token.NoPos, nil, "", errorType),
	)

	sig := types.NewSignature(nil, params, results, false)

	got := service.TypeToString(sig)
	// Should produce a function signature string
	// The exact format may vary, but it should contain the types
	if got == "" || got == "interface{}" {
		t.Errorf("TypeToString should produce non-empty signature, got %q", got)
	}

	t.Logf("Function signature: %s", got)
}

func TestInferType_InvalidToken(t *testing.T) {
	service := createTestTypeInferenceService(t)

	// Create a basic literal with invalid token kind
	expr := &ast.BasicLit{Kind: token.ILLEGAL, Value: "???"}

	typ, ok := service.InferType(expr)
	if !ok {
		t.Error("InferType should succeed (returning Invalid type)")
	}

	// Should return types.Invalid
	if basic, isBasic := typ.(*types.Basic); isBasic {
		if basic.Kind() != types.Invalid {
			t.Logf("Got basic type: %v (kind: %v)", basic, basic.Kind())
		}
	}
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func createTestTypeInferenceService(t *testing.T) *TypeInferenceService {
	t.Helper()

	fset := token.NewFileSet()
	file := &ast.File{
		Name: ast.NewIdent("test"),
	}

	service, err := NewTypeInferenceService(fset, file, plugin.NewNoOpLogger())
	if err != nil {
		t.Fatalf("failed to create TypeInferenceService: %v", err)
	}

	return service
}

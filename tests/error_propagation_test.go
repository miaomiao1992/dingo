package tests

import (
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"strings"
	"testing"

	"github.com/MadAppGang/dingo/pkg/plugin"
	"github.com/MadAppGang/dingo/pkg/plugin/builtin"
)

// TestErrorPropagationSmokeTests performs basic smoke tests for error propagation
func TestErrorPropagationSmokeTests(t *testing.T) {
	tests := []struct {
		name        string
		description string
		input       string
		shouldPass  bool
	}{
		{
			name:        "basic_function",
			description: "Plugin should handle functions without error propagation",
			input: `package main

func simple() int {
	return 42
}`,
			shouldPass: true,
		},
		{
			name:        "function_with_error_return",
			description: "Plugin should handle functions with error returns",
			input: `package main

func fetchUser() (User, error) {
	return User{}, nil
}`,
			shouldPass: true,
		},
		{
			name:        "multiple_functions",
			description: "Plugin should handle multiple functions",
			input: `package main

func first() int { return 1 }
func second() int { return 2 }`,
			shouldPass: true,
		},
		{
			name:        "nested_blocks",
			description: "Plugin should handle nested blocks",
			input: `package main

func process() error {
	if true {
		for i := 0; i < 10; i++ {
			if i > 5 {
				return nil
			}
		}
	}
	return nil
}`,
			shouldPass: true,
		},
		{
			name:        "empty_file",
			description: "Plugin should handle empty package",
			input:       `package main`,
			shouldPass:  true,
		},
		{
			name:        "imports",
			description: "Plugin should handle imports",
			input: `package main

import "fmt"

func hello() {
	fmt.Println("hello")
}`,
			shouldPass: true,
		},
		{
			name:        "structs",
			description: "Plugin should handle struct definitions",
			input: `package main

type User struct {
	ID   int
	Name string
}`,
			shouldPass: true,
		},
		{
			name:        "interfaces",
			description: "Plugin should handle interface definitions",
			input: `package main

type Reader interface {
	Read() ([]byte, error)
}`,
			shouldPass: true,
		},
		{
			name:        "constants",
			description: "Plugin should handle constants",
			input: `package main

const MaxSize = 1024`,
			shouldPass: true,
		},
		{
			name:        "variables",
			description: "Plugin should handle variables",
			input: `package main

var counter int`,
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the input
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.input, 0)
			if err != nil {
				if tt.shouldPass {
					t.Fatalf("Failed to parse input: %v", err)
				}
				return
			}

			// Create plugin
			p := builtin.NewErrorPropagationPlugin()

			// Create context
			ctx := &plugin.Context{
				FileSet: fset,
				Logger:  &testLogger{t: t},
			}

			// Transform
			result, err := p.Transform(ctx, file)
			if err != nil {
				if tt.shouldPass {
					t.Fatalf("Transform failed: %v", err)
				}
				return
			}

			// Verify result is still valid AST
			if result == nil {
				t.Fatal("Transform returned nil")
			}

			// Try to print the result to verify it's valid
			var buf strings.Builder
			if err := printer.Fprint(&buf, fset, result); err != nil {
				t.Fatalf("Failed to print transformed AST: %v", err)
			}

			// The output should be valid (not empty for non-empty inputs)
			output := buf.String()
			if tt.input != "" && output == "" {
				t.Fatal("Transform produced empty output for non-empty input")
			}
		})
	}
}

// TestTypeInference tests the type inference component
func TestTypeInference(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		funcName string
		wantType string // simplified type check
	}{
		{
			name: "int_return",
			code: `package main
func test() (int, error) { return 0, nil }`,
			funcName: "test",
			wantType: "int",
		},
		{
			name: "string_return",
			code: `package main
func test() (string, error) { return "", nil }`,
			funcName: "test",
			wantType: "string",
		},
		{
			name: "bool_return",
			code: `package main
func test() (bool, error) { return false, nil }`,
			funcName: "test",
			wantType: "bool",
		},
		{
			name: "pointer_return",
			code: `package main
type User struct{}
func test() (*User, error) { return nil, nil }`,
			funcName: "test",
			wantType: "pointer",
		},
		{
			name: "slice_return",
			code: `package main
func test() ([]int, error) { return nil, nil }`,
			funcName: "test",
			wantType: "slice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, 0)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			// Create type inference
			ti, err := builtin.NewTypeInference(fset, file)
			if err != nil {
				t.Fatalf("Type inference init failed: %v", err)
			}
			defer ti.Close()

			// Find the function
			var fn *ast.FuncDecl
			ast.Inspect(file, func(n ast.Node) bool {
				if f, ok := n.(*ast.FuncDecl); ok && f.Name.Name == tt.funcName {
					fn = f
					return false
				}
				return true
			})

			if fn == nil {
				t.Fatalf("Function %s not found", tt.funcName)
			}

			// Infer return type
			retType, err := ti.InferFunctionReturnType(fn)
			if err != nil {
				t.Fatalf("Failed to infer return type: %v", err)
			}

			// Generate zero value
			zeroVal := ti.GenerateZeroValue(retType)
			if zeroVal == nil {
				t.Fatal("GenerateZeroValue returned nil")
			}

			// Verify we can print the zero value
			var buf strings.Builder
			if err := printer.Fprint(&buf, fset, zeroVal); err != nil {
				t.Fatalf("Failed to print zero value: %v", err)
			}
		})
	}
}

// TestStatementLifter tests the statement lifting component
func TestStatementLifter(t *testing.T) {
	tests := []struct {
		name  string
		setup func() (*builtin.StatementLifter, ast.Expr, ast.Expr)
	}{
		{
			name: "basic_lift",
			setup: func() (*builtin.StatementLifter, ast.Expr, ast.Expr) {
				sl := builtin.NewStatementLifter()
				expr := &ast.CallExpr{
					Fun: &ast.Ident{Name: "fetch"},
				}
				zeroVal := &ast.Ident{Name: "nil"}
				return sl, expr, zeroVal
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sl, expr, zeroVal := tt.setup()
			result := sl.LiftExpression(expr, zeroVal, nil)

			if result == nil {
				t.Fatal("LiftExpression returned nil")
			}
			if len(result.Statements) == 0 {
				t.Fatal("LiftExpression produced no statements")
			}
			if result.Replacement == nil {
				t.Fatal("LiftExpression produced no replacement")
			}
			if result.TempVarName == "" {
				t.Fatal("LiftExpression produced empty temp var name")
			}
			if result.ErrorVarName == "" {
				t.Fatal("LiftExpression produced empty error var name")
			}
		})
	}
}

// TestErrorWrapper tests the error wrapping component
func TestErrorWrapper(t *testing.T) {
	tests := []struct {
		name    string
		message string
		errVar  string
	}{
		{
			name:    "simple_message",
			message: "failed",
			errVar:  "__err0",
		},
		{
			name:    "message_with_quotes",
			message: `user "admin" not found`,
			errVar:  "__err1",
		},
		{
			name:    "message_with_newline",
			message: "line1\nline2",
			errVar:  "__err2",
		},
		{
			name:    "message_with_tab",
			message: "col1\tcol2",
			errVar:  "__err3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ew := builtin.NewErrorWrapper()
			wrapped := ew.WrapError(tt.errVar, tt.message)

			if wrapped == nil {
				t.Fatal("WrapError returned nil")
			}

			// Verify it's a call expression
			call, ok := wrapped.(*ast.CallExpr)
			if !ok {
				t.Fatal("WrapError didn't return CallExpr")
			}

			// Verify it calls fmt.Errorf
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				t.Fatal("WrapError didn't create selector expression")
			}
			if sel.Sel.Name != "Errorf" {
				t.Fatalf("Expected Errorf, got %s", sel.Sel.Name)
			}

			// Verify it has 2 args
			if len(call.Args) != 2 {
				t.Fatalf("Expected 2 args, got %d", len(call.Args))
			}

			// Print it to verify it's valid
			fset := token.NewFileSet()
			var buf strings.Builder
			if err := printer.Fprint(&buf, fset, wrapped); err != nil {
				t.Fatalf("Failed to print wrapped error: %v", err)
			}

			// Verify output contains %w
			output := buf.String()
			if !strings.Contains(output, "%w") {
				t.Errorf("Output doesn't contain %%w: %s", output)
			}
		})
	}
}

// testLogger implements plugin.Logger for testing
type testLogger struct {
	t *testing.T
}

func (l *testLogger) Info(format string, args ...interface{}) {
	l.t.Logf("[INFO] "+format, args...)
}

func (l *testLogger) Warn(format string, args ...interface{}) {
	l.t.Logf("[WARN] "+format, args...)
}

func (l *testLogger) Error(format string, args ...interface{}) {
	l.t.Logf("[ERROR] "+format, args...)
}

func (l *testLogger) Debug(format string, args ...interface{}) {
	l.t.Logf("[DEBUG] "+format, args...)
}

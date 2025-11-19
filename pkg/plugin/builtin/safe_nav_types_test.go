package builtin

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/MadAppGang/dingo/pkg/plugin"
)

// TestSafeNavTypePlugin_Discovery tests the discovery phase of finding __INFER__ placeholders
func TestSafeNavTypePlugin_Discovery(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		expectedCount  int
		description    string
	}{
		{
			name: "single __INFER__ placeholder",
			source: `package main
func main() {
	x := __INFER__.field
}`,
			expectedCount: 1,
			description:   "Should find one __INFER__ identifier",
		},
		{
			name: "multiple __INFER__ placeholders",
			source: `package main
func main() {
	x := __INFER__.field1
	y := __INFER__.field2
	z := __INFER__.method()
}`,
			expectedCount: 3,
			description:   "Should find three __INFER__ identifiers",
		},
		{
			name: "chained __INFER__ placeholders",
			source: `package main
func main() {
	x := __INFER__.field1.__INFER__.field2
}`,
			expectedCount: 2,
			description:   "Should find two __INFER__ identifiers in chain",
		},
		{
			name: "no __INFER__ placeholders",
			source: `package main
func main() {
	x := user.field
}`,
			expectedCount: 0,
			description:   "Should find zero __INFER__ identifiers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse source
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.source, parser.ParseComments)
			if err != nil {
				t.Fatalf("Failed to parse source: %v", err)
			}

			// Create plugin context
			ctx := &plugin.Context{
				FileSet: fset,
				Logger:  plugin.NewNoOpLogger(),
			}

			// Create plugin
			p := NewSafeNavTypePlugin()
			p.SetContext(ctx)

			// Run discovery phase
			err = p.Process(file)
			if err != nil {
				t.Fatalf("Process failed: %v", err)
			}

			// Verify count
			if len(p.inferNodes) != tt.expectedCount {
				t.Errorf("%s: expected %d __INFER__ nodes, got %d",
					tt.description, tt.expectedCount, len(p.inferNodes))
			}
		})
	}
}

// TestSafeNavTypePlugin_PointerTypeResolution tests type resolution for pointer types
func TestSafeNavTypePlugin_PointerTypeResolution(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		expectedType   string
		expectedPtr    bool
		expectedOption bool
		description    string
	}{
		{
			name: "simple pointer type",
			source: `package main
type User struct {
	name string
}
func main() {
	var user *User
	x := __SAFE_NAV_INFER__(user, "name")
}`,
			expectedType:   "User",
			expectedPtr:    true,
			expectedOption: false,
			description:    "Should resolve *User to pointer type",
		},
		{
			name: "pointer to built-in type",
			source: `package main
func main() {
	var count *int
	x := __SAFE_NAV_INFER__(count, "value")
}`,
			expectedType:   "int",
			expectedPtr:    true,
			expectedOption: false,
			description:    "Should resolve *int to pointer type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse source
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.source, parser.ParseComments)
			if err != nil {
				t.Fatalf("Failed to parse source: %v", err)
			}

			// Type check the file
			conf := types.Config{}
			info := &types.Info{
				Types: make(map[ast.Expr]types.TypeAndValue),
				Defs:  make(map[*ast.Ident]types.Object),
				Uses:  make(map[*ast.Ident]types.Object),
			}
			_, err = conf.Check("test", fset, []*ast.File{file}, info)
			if err != nil {
				// Type errors are expected for __SAFE_NAV_INFER__ function
				// We'll ignore them for testing purposes
				t.Logf("Type check warnings (expected): %v", err)
			}

			// Create plugin context with type info
			ctx := &plugin.Context{
				FileSet:  fset,
				TypeInfo: info,
				Logger:   plugin.NewNoOpLogger(),
			}

			// Create plugin
			p := NewSafeNavTypePlugin()
			p.SetContext(ctx)

			// Run discovery phase
			err = p.Process(file)
			if err != nil {
				t.Fatalf("Process failed: %v", err)
			}

			// Run transform phase
			_, err = p.Transform(file)
			if err != nil {
				t.Fatalf("Transform failed: %v", err)
			}

			// Verify resolution (check inferNodes)
			if len(p.inferNodes) > 0 {
				node := p.inferNodes[0]
				if node.resolvedType != tt.expectedType {
					t.Errorf("%s: expected type %s, got %s",
						tt.description, tt.expectedType, node.resolvedType)
				}
				if node.isPointer != tt.expectedPtr {
					t.Errorf("%s: expected isPointer=%v, got %v",
						tt.description, tt.expectedPtr, node.isPointer)
				}
				if node.isOption != tt.expectedOption {
					t.Errorf("%s: expected isOption=%v, got %v",
						tt.description, tt.expectedOption, node.isOption)
				}
			}
		})
	}
}

// TestSafeNavTypePlugin_OptionTypeResolution tests type resolution for Option types
func TestSafeNavTypePlugin_OptionTypeResolution(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		expectedType   string
		expectedPtr    bool
		expectedOption bool
		description    string
	}{
		{
			name: "Option type detection",
			source: `package main
type OptionTag uint8
const (
	OptionTagSome OptionTag = iota
	OptionTagNone
)
type Option_User struct {
	tag   OptionTag
	value *User
}
func (o Option_User) IsSome() bool {
	return o.tag == OptionTagSome
}
func (o Option_User) IsNone() bool {
	return o.tag == OptionTagNone
}
type User struct {
	name string
}
func main() {
	var user Option_User
	x := __SAFE_NAV_INFER__(user, "name")
}`,
			expectedType:   "Option_User",
			expectedPtr:    false,
			expectedOption: true,
			description:    "Should resolve Option_User to Option type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse source
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.source, parser.ParseComments)
			if err != nil {
				t.Fatalf("Failed to parse source: %v", err)
			}

			// Type check the file
			conf := types.Config{}
			info := &types.Info{
				Types: make(map[ast.Expr]types.TypeAndValue),
				Defs:  make(map[*ast.Ident]types.Object),
				Uses:  make(map[*ast.Ident]types.Object),
			}
			_, err = conf.Check("test", fset, []*ast.File{file}, info)
			if err != nil {
				t.Logf("Type check warnings (expected): %v", err)
			}

			// Create plugin context with type info
			ctx := &plugin.Context{
				FileSet:  fset,
				TypeInfo: info,
				Logger:   plugin.NewNoOpLogger(),
			}

			// Create plugin
			p := NewSafeNavTypePlugin()
			p.SetContext(ctx)

			// Run discovery phase
			err = p.Process(file)
			if err != nil {
				t.Fatalf("Process failed: %v", err)
			}

			// Run transform phase
			_, err = p.Transform(file)
			if err != nil {
				t.Fatalf("Transform failed: %v", err)
			}

			// Verify resolution
			if len(p.inferNodes) > 0 {
				node := p.inferNodes[0]
				if node.resolvedType != tt.expectedType {
					t.Errorf("%s: expected type %s, got %s",
						tt.description, tt.expectedType, node.resolvedType)
				}
				if node.isPointer != tt.expectedPtr {
					t.Errorf("%s: expected isPointer=%v, got %v",
						tt.description, tt.expectedPtr, node.isPointer)
				}
				if node.isOption != tt.expectedOption {
					t.Errorf("%s: expected isOption=%v, got %v",
						tt.description, tt.expectedOption, node.isOption)
				}
			}
		})
	}
}

// TestSafeNavTypePlugin_ErrorReporting tests error reporting for invalid types
func TestSafeNavTypePlugin_ErrorReporting(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		expectError bool
		errorContains string
		description string
	}{
		{
			name: "non-nullable type error",
			source: `package main
type User struct {
	name string
}
func main() {
	var user User
	x := __SAFE_NAV_INFER__(user, "name")
}`,
			expectError:   true,
			errorContains: "nullable type",
			description:   "Should report error for non-nullable type",
		},
		{
			name: "built-in type error",
			source: `package main
func main() {
	var count int
	x := __SAFE_NAV_INFER__(count, "value")
}`,
			expectError:   true,
			errorContains: "nullable type",
			description:   "Should report error for built-in type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse source
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.source, parser.ParseComments)
			if err != nil {
				t.Fatalf("Failed to parse source: %v", err)
			}

			// Type check the file
			conf := types.Config{}
			info := &types.Info{
				Types: make(map[ast.Expr]types.TypeAndValue),
				Defs:  make(map[*ast.Ident]types.Object),
				Uses:  make(map[*ast.Ident]types.Object),
			}
			_, err = conf.Check("test", fset, []*ast.File{file}, info)
			if err != nil {
				t.Logf("Type check warnings (expected): %v", err)
			}

			// Create plugin context with type info
			ctx := &plugin.Context{
				FileSet:  fset,
				TypeInfo: info,
				Logger:   plugin.NewNoOpLogger(),
			}

			// Create plugin
			p := NewSafeNavTypePlugin()
			p.SetContext(ctx)

			// Run discovery phase
			err = p.Process(file)
			if err != nil {
				t.Fatalf("Process failed: %v", err)
			}

			// Run transform phase
			_, err = p.Transform(file)

			// Check for errors
			errors := p.GetErrors()
			hasError := len(errors) > 0 || ctx.HasErrors()

			if tt.expectError && !hasError {
				t.Errorf("%s: expected error but got none", tt.description)
			}

			if tt.expectError && hasError {
				// Verify error message contains expected substring
				foundExpected := false
				// Check plugin errors ([]string)
				for _, e := range errors {
					if containsSubstring(e, tt.errorContains) {
						foundExpected = true
						break
					}
				}
				// Check context errors ([]error)
				if !foundExpected {
					for _, e := range ctx.GetErrors() {
						if e != nil && containsSubstring(e.Error(), tt.errorContains) {
							foundExpected = true
							break
						}
					}
				}
				if !foundExpected {
					t.Errorf("%s: error message should contain '%s', plugin errors: %v, context errors: %v",
						tt.description, tt.errorContains, errors, ctx.GetErrors())
				}
			}
		})
	}
}

// TestSafeNavTypePlugin_Transform tests AST transformation
func TestSafeNavTypePlugin_Transform(t *testing.T) {
	source := `package main
type User struct {
	name string
}
func main() {
	var user *User
	x := __INFER__.name
}`

	// Parse source
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	// Type check the file
	conf := types.Config{}
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}
	_, err = conf.Check("test", fset, []*ast.File{file}, info)
	if err != nil {
		t.Logf("Type check warnings (expected): %v", err)
	}

	// Create plugin context
	ctx := &plugin.Context{
		FileSet:  fset,
		TypeInfo: info,
		Logger:   plugin.NewNoOpLogger(),
	}

	// Create plugin
	p := NewSafeNavTypePlugin()
	p.SetContext(ctx)

	// Run discovery
	err = p.Process(file)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Verify we found __INFER__
	if len(p.inferNodes) == 0 {
		t.Fatal("Expected to find __INFER__ node")
	}

	// Run transform
	transformed, err := p.Transform(file)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	// Verify transformation occurred
	if transformed == nil {
		t.Fatal("Transform returned nil")
	}

	// The __INFER__ should have been replaced
	// Walk the transformed AST to verify no __INFER__ remains
	foundInfer := false
	ast.Inspect(transformed, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok {
			if ident.Name == "__INFER__" {
				foundInfer = true
			}
		}
		return true
	})

	if foundInfer {
		t.Error("Transform should have replaced __INFER__ but it still exists")
	}
}

// Helper function to check if error message contains substring
func containsSubstring(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) &&
		(s == substr || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

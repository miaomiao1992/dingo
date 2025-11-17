package builtin

import (
	"go/parser"
	"go/printer"
	"go/token"
	"strings"
	"testing"

	"github.com/MadAppGang/dingo/pkg/plugin"
)

func TestNewFunctionalUtilitiesPlugin(t *testing.T) {
	p := NewFunctionalUtilitiesPlugin()

	if p.Name() != "functional_utilities" {
		t.Errorf("expected name 'functional_utilities', got %q", p.Name())
	}

	if !p.Enabled() {
		t.Error("expected plugin to be enabled by default")
	}
}

func TestTransformMap(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:  "simple map with multiplication",
			input: "package main; func test() { numbers := []int{1,2,3}; numbers.Map(func(x int) int { return x * 2 }) }",
			expected: "func() []int {\n\tvar __temp0 []int\n\t__temp0 = make([]int, 0, len(numbers))\n\tfor _, x := range numbers {\n\t\t__temp0 = append(__temp0, x*2)\n\t}\n\treturn __temp0\n}()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse input
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.input, 0)
			if err != nil {
				t.Fatalf("failed to parse input: %v", err)
			}

			// Create plugin and context
			p := NewFunctionalUtilitiesPlugin()
			ctx := &plugin.Context{
				FileSet: fset,
				Logger:  &testLogger{},
			}

			// Transform
			result, err := p.Transform(ctx, file)
			if err != nil {
				t.Fatalf("transform failed: %v", err)
			}

			// Print result
			var buf strings.Builder
			if err := printer.Fprint(&buf, fset, result); err != nil {
				t.Fatalf("failed to print result: %v", err)
			}

			// Check if the output contains the expected pattern
			// (exact match is hard due to formatting variations)
			output := buf.String()
			if !strings.Contains(output, "__temp0") {
				t.Errorf("expected output to contain '__temp0', got:\n%s", output)
			}
			if !strings.Contains(output, "make([]int") {
				t.Errorf("expected output to contain 'make([]int', got:\n%s", output)
			}
			if !strings.Contains(output, "for _, x := range") {
				t.Errorf("expected output to contain 'for _, x := range', got:\n%s", output)
			}
		})
	}
}

func TestTransformFilter(t *testing.T) {
	input := "package main; func test() { numbers.filter(func(x int) bool { return x > 0 }) }"

	// Parse input
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", input, 0)
	if err != nil {
		t.Fatalf("failed to parse input: %v", err)
	}

	// Create plugin and context
	p := NewFunctionalUtilitiesPlugin()
	ctx := &plugin.Context{
		FileSet: fset,
		Logger:  &testLogger{},
	}

	// Transform
	result, err := p.Transform(ctx, file)
	if err != nil {
		t.Fatalf("transform failed: %v", err)
	}

	// Print result
	var buf strings.Builder
	if err := printer.Fprint(&buf, fset, result); err != nil {
		t.Fatalf("failed to print result: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "__temp0") {
		t.Errorf("expected output to contain '__temp0', got:\n%s", output)
	}
	if !strings.Contains(output, "if x > 0") {
		t.Errorf("expected output to contain 'if x > 0', got:\n%s", output)
	}
}

func TestTransformReduce(t *testing.T) {
	input := "package main; func test() { numbers.reduce(0, func(acc int, x int) int { return acc + x }) }"

	// Parse input
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", input, 0)
	if err != nil {
		t.Fatalf("failed to parse input: %v", err)
	}

	// Create plugin and context
	p := NewFunctionalUtilitiesPlugin()
	ctx := &plugin.Context{
		FileSet: fset,
		Logger:  &testLogger{},
	}

	// Transform
	result, err := p.Transform(ctx, file)
	if err != nil {
		t.Fatalf("transform failed: %v", err)
	}

	// Print result
	var buf strings.Builder
	if err := printer.Fprint(&buf, fset, result); err != nil {
		t.Fatalf("failed to print result: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "__temp0") {
		t.Errorf("expected output to contain '__temp0', got:\n%s", output)
	}
	if !strings.Contains(output, "for _, x := range") {
		t.Errorf("expected output to contain 'for _, x := range', got:\n%s", output)
	}
}

func TestTransformSum(t *testing.T) {
	input := "package main; func test() { numbers.sum() }"

	// Parse input
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", input, 0)
	if err != nil {
		t.Fatalf("failed to parse input: %v", err)
	}

	// Create plugin and context
	p := NewFunctionalUtilitiesPlugin()
	ctx := &plugin.Context{
		FileSet: fset,
		Logger:  &testLogger{},
	}

	// Transform
	result, err := p.Transform(ctx, file)
	if err != nil {
		t.Fatalf("transform failed: %v", err)
	}

	// Print result
	var buf strings.Builder
	if err := printer.Fprint(&buf, fset, result); err != nil {
		t.Fatalf("failed to print result: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "__temp") {
		t.Errorf("expected output to contain temp variables, got:\n%s", output)
	}
	if !strings.Contains(output, "for _,") {
		t.Errorf("expected output to contain 'for _,', got:\n%s", output)
	}
}

func TestTransformAll(t *testing.T) {
	input := "package main; func test() { numbers.all(func(x int) bool { return x > 0 }) }"

	// Parse input
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", input, 0)
	if err != nil {
		t.Fatalf("failed to parse input: %v", err)
	}

	// Create plugin and context
	p := NewFunctionalUtilitiesPlugin()
	ctx := &plugin.Context{
		FileSet: fset,
		Logger:  &testLogger{},
	}

	// Transform
	result, err := p.Transform(ctx, file)
	if err != nil {
		t.Fatalf("transform failed: %v", err)
	}

	// Print result
	var buf strings.Builder
	if err := printer.Fprint(&buf, fset, result); err != nil {
		t.Fatalf("failed to print result: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "break") {
		t.Errorf("expected output to contain 'break' for early exit, got:\n%s", output)
	}
	if !strings.Contains(output, "true") {
		t.Errorf("expected output to contain 'true', got:\n%s", output)
	}
}

func TestTransformAny(t *testing.T) {
	input := "package main; func test() { numbers.any(func(x int) bool { return x < 0 }) }"

	// Parse input
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", input, 0)
	if err != nil {
		t.Fatalf("failed to parse input: %v", err)
	}

	// Create plugin and context
	p := NewFunctionalUtilitiesPlugin()
	ctx := &plugin.Context{
		FileSet: fset,
		Logger:  &testLogger{},
	}

	// Transform
	result, err := p.Transform(ctx, file)
	if err != nil {
		t.Fatalf("transform failed: %v", err)
	}

	// Print result
	var buf strings.Builder
	if err := printer.Fprint(&buf, fset, result); err != nil {
		t.Fatalf("failed to print result: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "break") {
		t.Errorf("expected output to contain 'break' for early exit, got:\n%s", output)
	}
	if !strings.Contains(output, "false") {
		t.Errorf("expected output to contain 'false', got:\n%s", output)
	}
}

func TestTransformCount(t *testing.T) {
	input := "package main; func test() { numbers.count(func(x int) bool { return x > 5 }) }"

	// Parse input
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", input, 0)
	if err != nil {
		t.Fatalf("failed to parse input: %v", err)
	}

	// Create plugin and context
	p := NewFunctionalUtilitiesPlugin()
	ctx := &plugin.Context{
		FileSet: fset,
		Logger:  &testLogger{},
	}

	// Transform
	result, err := p.Transform(ctx, file)
	if err != nil {
		t.Fatalf("transform failed: %v", err)
	}

	// Print result
	var buf strings.Builder
	if err := printer.Fprint(&buf, fset, result); err != nil {
		t.Fatalf("failed to print result: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "__temp") {
		t.Errorf("expected output to contain temp counter variable, got:\n%s", output)
	}
	if !strings.Contains(output, "for _,") {
		t.Errorf("expected output to contain 'for _,', got:\n%s", output)
	}
	if !strings.Contains(output, "if") {
		t.Errorf("expected output to contain 'if' for conditional counting, got:\n%s", output)
	}
	if !strings.Contains(output, "++") {
		t.Errorf("expected output to contain '++' for counter increment, got:\n%s", output)
	}
}

// testLogger is a simple logger for testing
type testLogger struct{}

func (l *testLogger) Debug(format string, args ...interface{}) {}
func (l *testLogger) Info(format string, args ...interface{})  {}
func (l *testLogger) Warn(format string, args ...interface{})  {}
func (l *testLogger) Error(format string, args ...interface{}) {}

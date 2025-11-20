package preprocessor

import (
	"strings"
	"testing"
)

// Benchmark small files (< 100 lines)
func BenchmarkLambdaProcessor_SmallFile(b *testing.B) {
	processor := NewLambdaProcessor()

	// Small file: 20 lines with 5 lambdas
	source := []byte(`package main

func main() {
	double := (x: int): int => x * 2
	add := (x: int, y: int): int => x + y
	greet := (name: string): string => "Hello, " + name

	numbers := []int{1, 2, 3, 4, 5}
	doubled := map(numbers, (x: int): int => x * 2)
	filtered := filter(numbers, (x: int): bool => x > 2)
}
`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = processor.Process(source)
	}
}

// Benchmark medium files (100-1000 lines)
func BenchmarkLambdaProcessor_MediumFile(b *testing.B) {
	processor := NewLambdaProcessor()

	// Medium file: ~300 lines with 50 lambdas
	var builder strings.Builder
	builder.WriteString("package main\n\nfunc main() {\n")

	// Generate 50 lambda expressions
	for i := 0; i < 50; i++ {
		builder.WriteString("\tdouble := (x: int): int => x * 2\n")
		builder.WriteString("\tadd := (x: int, y: int): int => x + y\n")
		builder.WriteString("\tfiltered := filter(items, (x: int): bool => x > 0)\n")
		builder.WriteString("\tmapped := map(items, (x: int): int => x * 3)\n")
		builder.WriteString("\n")
	}
	builder.WriteString("}\n")

	source := []byte(builder.String())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = processor.Process(source)
	}
}

// Benchmark large files (> 1000 lines)
func BenchmarkLambdaProcessor_LargeFile(b *testing.B) {
	processor := NewLambdaProcessor()

	// Large file: ~2000 lines with 200 lambdas
	var builder strings.Builder
	builder.WriteString("package main\n\nimport \"fmt\"\n\nfunc main() {\n")

	// Generate 200 lambda expressions with varied patterns
	for i := 0; i < 200; i++ {
		builder.WriteString("\t// Lambda group " + string(rune('0'+i%10)) + "\n")
		builder.WriteString("\tdouble := (x: int): int => x * 2\n")
		builder.WriteString("\tadd := (x: int, y: int): int => x + y\n")
		builder.WriteString("\tvalidate := (s: string): bool => len(s) > 0\n")
		builder.WriteString("\ttransform := (x: int): string => fmt.Sprintf(\"%d\", x)\n")
		builder.WriteString("\tcomplex := (a: int, b: int, c: int): int => a * b + c\n")
		builder.WriteString("\n")
	}
	builder.WriteString("}\n")

	source := []byte(builder.String())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = processor.Process(source)
	}
}

// Benchmark nested function calls (complex expressions)
func BenchmarkLambdaProcessor_NestedCalls(b *testing.B) {
	processor := NewLambdaProcessor()

	// File with complex nested function calls in lambda bodies
	source := []byte(`package main

func main() {
	// Nested function calls with multiple arguments
	result1 := map(numbers, (x: int): int => transform(x, 2, 10))
	result2 := map(numbers, (x: int): int => transform(transform(x, 2, 0), 3, 1))
	result3 := filter(users, (u: string): bool => validate(u, context))
	result4 := reduce(numbers, (acc: int, x: int): int => combine(acc, x, seed), 0)

	// Deeply nested expressions
	result5 := map(
		filter(data, (s: string): bool => len(s) > 3),
		(s: string): string => strings.Repeat(s, 2),
	)

	// Chained transformations
	result6 := compose(
		(x: int): int => x * 2,
		(x: int): int => x + 1,
	)
}
`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = processor.Process(source)
	}
}

// Benchmark TypeScript arrow syntax vs Rust pipe syntax
func BenchmarkLambdaProcessor_TypeScriptStyle(b *testing.B) {
	processor := NewLambdaProcessor() // Default: TypeScript

	source := []byte(`package main

func main() {
	double := (x: int): int => x * 2
	add := (x: int, y: int): int => x + y
	filter := (x: int): bool => x > 0
	mapped := map(items, (x: int): int => x * 3)
}
`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = processor.Process(source)
	}
}

func BenchmarkLambdaProcessor_RustStyle(b *testing.B) {
	processor := &LambdaProcessor{style: StyleRust}

	source := []byte(`package main

func main() {
	double := |x: int| -> int { x * 2 }
	add := |x: int, y: int| -> int { x + y }
	filter := |x: int| -> bool { x > 0 }
	mapped := map(items, |x: int| -> int { x * 3 })
}
`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = processor.Process(source)
	}
}

// Benchmark memory allocations
func BenchmarkLambdaProcessor_Allocations(b *testing.B) {
	processor := NewLambdaProcessor()

	source := []byte(`package main

func main() {
	double := (x: int): int => x * 2
	add := (x: int, y: int): int => x + y
	filtered := filter(items, (x: int): bool => x > 0)
	mapped := map(items, (x: int): int => x * 3)
}
`)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = processor.Process(source)
	}
}

// Benchmark with no lambdas (baseline performance)
func BenchmarkLambdaProcessor_NoLambdas(b *testing.B) {
	processor := NewLambdaProcessor()

	source := []byte(`package main

func main() {
	x := 42
	y := "hello"
	z := []int{1, 2, 3}

	for _, v := range z {
		println(v)
	}
}
`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = processor.Process(source)
	}
}

// Benchmark expression vs block lambdas
func BenchmarkLambdaProcessor_ExpressionLambdas(b *testing.B) {
	processor := NewLambdaProcessor()

	source := []byte(`package main

func main() {
	double := (x: int): int => x * 2
	add := (x: int, y: int): int => x + y
	greet := (name: string): string => "Hello, " + name
	isPositive := (n: int): bool => n > 0
}
`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = processor.Process(source)
	}
}

func BenchmarkLambdaProcessor_BlockLambdas(b *testing.B) {
	processor := NewLambdaProcessor()

	source := []byte(`package main

func main() {
	double := (x: int): int => {
		return x * 2
	}
	add := (x: int, y: int): int => {
		return x + y
	}
	greet := (name: string): string => {
		return "Hello, " + name
	}
	isPositive := (n: int): bool => {
		return n > 0
	}
}
`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = processor.Process(source)
	}
}

// Benchmark real-world pattern: functional pipeline
func BenchmarkLambdaProcessor_FunctionalPipeline(b *testing.B) {
	processor := NewLambdaProcessor()

	source := []byte(`package main

func main() {
	// Typical functional programming pipeline
	result := filter(
		map(
			filter(users, (u: User): bool => u.age > 18),
			(u: User): string => u.name,
		),
		(name: string): bool => len(name) > 0,
	)

	// Chained transformations
	processed := users.
		filter((u: User): bool => u.verified).
		map((u: User): string => u.email).
		filter((e: string): bool => strings.Contains(e, "@"))
}
`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = processor.Process(source)
	}
}

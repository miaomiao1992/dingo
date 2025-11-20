package preprocessor

import (
	"testing"
)

// Benchmark data for null coalescing tests

// SIMPLE INLINE BENCHMARKS (identifier ?? literal)

// BenchmarkNullCoalesceInline_NonNil benchmarks value ?? default (non-nil value)
func BenchmarkNullCoalesceInline_NonNil(b *testing.B) {
	value := "hello"
	valuePtr := &value
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Generated code for: valuePtr ?? "default"
		result := func() string {
			if valuePtr != nil {
				return *valuePtr
			}
			return "default"
		}()
		_ = result
	}
}

// BenchmarkHandWrittenCoalesce_NonNil benchmarks hand-written equivalent
func BenchmarkHandWrittenCoalesce_NonNil(b *testing.B) {
	value := "hello"
	valuePtr := &value
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result string
		if valuePtr != nil {
			result = *valuePtr
		} else {
			result = "default"
		}
		_ = result
	}
}

// BenchmarkNullCoalesceInline_Nil benchmarks value ?? default (nil value)
func BenchmarkNullCoalesceInline_Nil(b *testing.B) {
	var valuePtr *string = nil
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Generated code for: valuePtr ?? "default"
		result := func() string {
			if valuePtr != nil {
				return *valuePtr
			}
			return "default"
		}()
		_ = result
	}
}

// BenchmarkHandWrittenCoalesce_Nil benchmarks hand-written with nil
func BenchmarkHandWrittenCoalesce_Nil(b *testing.B) {
	var valuePtr *string = nil
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result string
		if valuePtr != nil {
			result = *valuePtr
		} else {
			result = "default"
		}
		_ = result
	}
}

// INTEGER BENCHMARKS

// BenchmarkNullCoalesceInline_IntNonNil benchmarks integer coalescing
func BenchmarkNullCoalesceInline_IntNonNil(b *testing.B) {
	value := 42
	valuePtr := &value
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Generated code for: valuePtr ?? 0
		result := func() int {
			if valuePtr != nil {
				return *valuePtr
			}
			return 0
		}()
		_ = result
	}
}

// BenchmarkHandWrittenCoalesce_IntNonNil benchmarks hand-written integer
func BenchmarkHandWrittenCoalesce_IntNonNil(b *testing.B) {
	value := 42
	valuePtr := &value
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result int
		if valuePtr != nil {
			result = *valuePtr
		} else {
			result = 0
		}
		_ = result
	}
}

// COMPLEX IIFE BENCHMARKS (function calls, expressions)

// Helper functions for complex benchmarks
func getExpensiveString() *string {
	result := "computed"
	return &result
}

func getDefaultString() string {
	return "default"
}

// BenchmarkNullCoalesceIIFE_FunctionCall benchmarks func() ?? default
func BenchmarkNullCoalesceIIFE_FunctionCall(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Generated code for: getExpensiveString() ?? "default"
		result := func() string {
			__coalesce0 := getExpensiveString()
			if __coalesce0 != nil {
				return *__coalesce0
			}
			return "default"
		}()
		_ = result
	}
}

// BenchmarkHandWrittenCoalesce_FunctionCall benchmarks hand-written function call
func BenchmarkHandWrittenCoalesce_FunctionCall(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		value := getExpensiveString()
		var result string
		if value != nil {
			result = *value
		} else {
			result = "default"
		}
		_ = result
	}
}

// CHAINED COALESCING BENCHMARKS (a ?? b ?? c)

// BenchmarkNullCoalesceChain_FirstNonNil benchmarks a ?? b ?? c (first non-nil)
func BenchmarkNullCoalesceChain_FirstNonNil(b *testing.B) {
	first := "first"
	firstPtr := &first
	second := "second"
	secondPtr := &second
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Generated code for: firstPtr ?? secondPtr ?? "fallback"
		result := func() string {
			__coalesce0 := firstPtr
			if __coalesce0 != nil {
				return *__coalesce0
			}
			__coalesce1 := secondPtr
			if __coalesce1 != nil {
				return *__coalesce1
			}
			return "fallback"
		}()
		_ = result
	}
}

// BenchmarkHandWrittenChain_FirstNonNil benchmarks hand-written chain
func BenchmarkHandWrittenChain_FirstNonNil(b *testing.B) {
	first := "first"
	firstPtr := &first
	second := "second"
	secondPtr := &second
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result string
		if firstPtr != nil {
			result = *firstPtr
		} else if secondPtr != nil {
			result = *secondPtr
		} else {
			result = "fallback"
		}
		_ = result
	}
}

// BenchmarkNullCoalesceChain_SecondNonNil benchmarks a ?? b ?? c (second non-nil)
func BenchmarkNullCoalesceChain_SecondNonNil(b *testing.B) {
	var firstPtr *string = nil
	second := "second"
	secondPtr := &second
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Generated code for: firstPtr ?? secondPtr ?? "fallback"
		result := func() string {
			__coalesce0 := firstPtr
			if __coalesce0 != nil {
				return *__coalesce0
			}
			__coalesce1 := secondPtr
			if __coalesce1 != nil {
				return *__coalesce1
			}
			return "fallback"
		}()
		_ = result
	}
}

// BenchmarkHandWrittenChain_SecondNonNil benchmarks hand-written second hit
func BenchmarkHandWrittenChain_SecondNonNil(b *testing.B) {
	var firstPtr *string = nil
	second := "second"
	secondPtr := &second
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result string
		if firstPtr != nil {
			result = *firstPtr
		} else if secondPtr != nil {
			result = *secondPtr
		} else {
			result = "fallback"
		}
		_ = result
	}
}

// BenchmarkNullCoalesceChain_AllNil benchmarks a ?? b ?? c (all nil, use fallback)
func BenchmarkNullCoalesceChain_AllNil(b *testing.B) {
	var firstPtr *string = nil
	var secondPtr *string = nil
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Generated code for: firstPtr ?? secondPtr ?? "fallback"
		result := func() string {
			__coalesce0 := firstPtr
			if __coalesce0 != nil {
				return *__coalesce0
			}
			__coalesce1 := secondPtr
			if __coalesce1 != nil {
				return *__coalesce1
			}
			return "fallback"
		}()
		_ = result
	}
}

// BenchmarkHandWrittenChain_AllNil benchmarks hand-written all nil
func BenchmarkHandWrittenChain_AllNil(b *testing.B) {
	var firstPtr *string = nil
	var secondPtr *string = nil
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result string
		if firstPtr != nil {
			result = *firstPtr
		} else if secondPtr != nil {
			result = *secondPtr
		} else {
			result = "fallback"
		}
		_ = result
	}
}

// COMBINED WITH SAFE NAV BENCHMARKS

// BenchmarkCoalesceWithSafeNav benchmarks user?.address?.city ?? "Unknown"
func BenchmarkCoalesceWithSafeNav(b *testing.B) {
	user := getBenchUserWithAddress()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Generated code for: user?.address?.city ?? "Unknown"
		// (safe nav produces *string, coalesce unwraps it)
		result := func() string {
			safeNavResult := func() *string {
				if user == nil {
					return nil
				}
				if user.Address == nil {
					return nil
				}
				return &user.Address.City
			}()

			if safeNavResult != nil {
				return *safeNavResult
			}
			return "Unknown"
		}()
		_ = result
	}
}

// BenchmarkHandWrittenCoalesceWithSafeNav benchmarks hand-written combination
func BenchmarkHandWrittenCoalesceWithSafeNav(b *testing.B) {
	user := getBenchUserWithAddress()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result string
		if user != nil && user.Address != nil {
			result = user.Address.City
		} else {
			result = "Unknown"
		}
		_ = result
	}
}

// IIFE OPTIMIZATION VALIDATION FOR NULL COALESCING

// BenchmarkIIFECoalesce_Inline benchmarks inline IIFE overhead
func BenchmarkIIFECoalesce_Inline(b *testing.B) {
	value := 42
	valuePtr := &value
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result := func() int {
			if valuePtr != nil {
				return *valuePtr
			}
			return 0
		}()
		_ = result
	}
}

// BenchmarkTernarySimulation benchmarks simulated ternary
func BenchmarkTernarySimulation(b *testing.B) {
	value := 42
	valuePtr := &value
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result int
		if valuePtr != nil {
			result = *valuePtr
		} else {
			result = 0
		}
		_ = result
	}
}

// COMPLEX EXPRESSION BENCHMARKS

// BenchmarkComplexExpressionCoalesce benchmarks (a + b) ?? default
func BenchmarkComplexExpressionCoalesce(b *testing.B) {
	a := 10
	aPtr := &a
	b_val := 20
	bPtr := &b_val
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Generated code for: (aPtr + bPtr) ?? 0
		// Note: This would be more complex in reality with safe arithmetic
		result := func() int {
			if aPtr != nil && bPtr != nil {
				sum := *aPtr + *bPtr
				return sum
			}
			return 0
		}()
		_ = result
	}
}

// BenchmarkHandWrittenComplexExpression benchmarks hand-written complex
func BenchmarkHandWrittenComplexExpression(b *testing.B) {
	a := 10
	aPtr := &a
	b_val := 20
	bPtr := &b_val
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result int
		if aPtr != nil && bPtr != nil {
			result = *aPtr + *bPtr
		} else {
			result = 0
		}
		_ = result
	}
}

// STRUCT FIELD DEFAULT BENCHMARKS

type BenchConfig struct {
	Timeout  *int
	MaxRetry *int
}

// BenchmarkStructFieldCoalesce benchmarks config.timeout ?? 30
func BenchmarkStructFieldCoalesce(b *testing.B) {
	timeout := 60
	config := BenchConfig{
		Timeout: &timeout,
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Generated code for: config.timeout ?? 30
		result := func() int {
			if config.Timeout != nil {
				return *config.Timeout
			}
			return 30
		}()
		_ = result
	}
}

// BenchmarkHandWrittenStructFieldCoalesce benchmarks hand-written struct field
func BenchmarkHandWrittenStructFieldCoalesce(b *testing.B) {
	timeout := 60
	config := BenchConfig{
		Timeout: &timeout,
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result int
		if config.Timeout != nil {
			result = *config.Timeout
		} else {
			result = 30
		}
		_ = result
	}
}

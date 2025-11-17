package preprocessor

import (
	"strings"
	"testing"
)

func TestErrorPropagationBasic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "simple assignment",
			input: `package main

func readConfig(path: string) ([]byte, error) {
	let data = ReadFile(path)?
	return data, nil
}`,
			expected: `package main

func readConfig(path string) ([]byte, error) {
	__tmp0, __err0 := ReadFile(path)
	// dingo:s:1
	if __err0 != nil {
		return nil, __err0
	}
	// dingo:e:1
	var data = __tmp0
	return data, nil
}`,
		},
		{
			name: "simple return",
			input: `package main

func parseInt(s: string) (int, error) {
	return Atoi(s)?
}`,
			expected: `package main

func parseInt(s string) (int, error) {
	__tmp0, __err0 := Atoi(s)
	// dingo:s:1
	if __err0 != nil {
		return 0, __err0
	}
	// dingo:e:1
	return __tmp0, nil
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New([]byte(tt.input))
			result, _, err := p.Process()
			if err != nil {
				t.Fatalf("preprocessing failed: %v", err)
			}

			actual := strings.TrimSpace(result)
			expected := strings.TrimSpace(tt.expected)

			if actual != expected {
				t.Errorf("output mismatch:\n=== EXPECTED ===\n%s\n\n=== ACTUAL ===\n%s\n", expected, actual)
			}
		})
	}
}

// TestIMPORTANT1_ErrorMessageEscaping tests IMPORTANT-1 fix:
// Error messages with % characters must be escaped to prevent fmt.Errorf panics
func TestIMPORTANT1_ErrorMessageEscaping(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldHave  string
		shouldntHave string
	}{
		{
			name: "percent in error message",
			input: `package main

func readData(path: string) ([]byte, error) {
	let data = ReadFile(path)? "failed: 50% complete"
	return data, nil
}`,
			shouldHave: `fmt.Errorf("failed: 50%% complete: %w"`,
			shouldntHave: `fmt.Errorf("failed: 50% complete: %w"`, // This would panic!
		},
		{
			name: "multiple percents in error message",
			input: `package main

func process() (string, error) {
	return DoWork()? "progress: 25% to 75%"
}`,
			shouldHave: `fmt.Errorf("progress: 25%% to 75%%: %w"`,
			shouldntHave: `fmt.Errorf("progress: 25% to 75%: %w"`, // This would panic!
		},
		{
			name: "percent-w pattern in error message",
			input: `package main

func test() (int, error) {
	return Calc()? "100%w complete"
}`,
			shouldHave: `fmt.Errorf("100%%w complete: %w"`,
			shouldntHave: `fmt.Errorf("100%w complete: %w"`, // Would create %w%w!
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New([]byte(tt.input))
			result, _, err := p.Process()
			if err != nil {
				t.Fatalf("preprocessing failed: %v", err)
			}

			actual := string(result)

			if !strings.Contains(actual, tt.shouldHave) {
				t.Errorf("expected to find:\n%s\n\nActual output:\n%s", tt.shouldHave, actual)
			}

			if strings.Contains(actual, tt.shouldntHave) {
				t.Errorf("should NOT contain (unescaped):\n%s\n\nActual output:\n%s", tt.shouldntHave, actual)
			}
		})
	}
}

// TestIMPORTANT2_TypeAnnotationEnhancement tests IMPORTANT-2 fix:
// Type annotations must handle complex Go types including function types, channels, nested generics
func TestIMPORTANT2_TypeAnnotationEnhancement(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "function type in parameters",
			input: `package main

func process(handler: func(int) error) error {
	return nil
}`,
			expected: `package main

func process(handler func(int) error) error {
	return nil
}`,
		},
		{
			name: "channel with direction",
			input: `package main

func send(ch: <-chan string, out: chan<- int) {
}`,
			expected: `package main

func send(ch <-chan string, out chan<- int) {
}`,
		},
		{
			name: "complex nested generics",
			input: `package main

func lookup(cache: map[string][]interface{}, key: string) {
}`,
			expected: `package main

func lookup(cache map[string][]interface{}, key string) {
}`,
		},
		{
			name: "function returning multiple values",
			input: `package main

func transform(fn: func(a, b int) (string, error)) {
}`,
			expected: `package main

func transform(fn func(a, b int) (string, error)) {
}`,
		},
		{
			name: "nested function types",
			input: `package main

func higher(fn: func() func() error) {
}`,
			expected: `package main

func higher(fn func() func() error) {
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New([]byte(tt.input))
			result, _, err := p.Process()
			if err != nil {
				t.Fatalf("preprocessing failed: %v", err)
			}

			actual := strings.TrimSpace(string(result))
			expected := strings.TrimSpace(tt.expected)

			if actual != expected {
				t.Errorf("output mismatch:\n=== EXPECTED ===\n%s\n\n=== ACTUAL ===\n%s\n", expected, actual)
			}
		})
	}
}

// TestGeminiCodeReviewFixes verifies both IMPORTANT fixes from Gemini code review work together
func TestGeminiCodeReviewFixes(t *testing.T) {
	// This test combines both fixes in a realistic scenario:
	// - IMPORTANT-1: Error message escaping (% → %%)
	// - IMPORTANT-2: Complex type annotations (function types, channels)
	// - Bonus: Ternary detection must ignore : in string literals

	input := `package main

func processData(handler: func([]byte) error, path: string) ([]byte, error) {
	let data = ReadFile(path)? "failed: 50% complete"
	return data, nil
}

func fetchConfig(url: string) ([]byte, error) {
	return HttpGet(url)? "progress: 25% to 75%"
}`

	expected := `package main

import "fmt"

func processData(handler func([]byte) error, path string) ([]byte, error) {
	__tmp0, __err0 := ReadFile(path)
	// dingo:s:1
	if __err0 != nil {
		return nil, fmt.Errorf("failed: 50%% complete: %w", __err0)
	}
	// dingo:e:1
	var data = __tmp0
	return data, nil
}

func fetchConfig(url string) ([]byte, error) {
	__tmp0, __err0 := HttpGet(url)
	// dingo:s:1
	if __err0 != nil {
		return nil, fmt.Errorf("progress: 25%% to 75%%: %w", __err0)
	}
	// dingo:e:1
	return __tmp0, nil
}`

	p := New([]byte(input))
	result, _, err := p.Process()
	if err != nil {
		t.Fatalf("preprocessing failed: %v", err)
	}

	actual := strings.TrimSpace(string(result))
	expectedNorm := strings.TrimSpace(expected)

	if actual != expectedNorm {
		t.Errorf("output mismatch:\n=== EXPECTED ===\n%s\n\n=== ACTUAL ===\n%s\n", expectedNorm, actual)
	}

	// Verify critical aspects of the fixes
	if !strings.Contains(actual, `"failed: 50%% complete: %w"`) {
		t.Error("IMPORTANT-1 failed: % not escaped in first error message")
	}
	if !strings.Contains(actual, `"progress: 25%% to 75%%: %w"`) {
		t.Error("IMPORTANT-1 failed: % not escaped in second error message")
	}
	if !strings.Contains(actual, "handler func([]byte) error") {
		t.Error("IMPORTANT-2 failed: function type not handled correctly")
	}
	if !strings.Contains(actual, "url string") {
		t.Error("Type annotation conversion failed")
	}
}

// TestSourceMapGeneration verifies that source maps are correctly generated
// for error propagation expansions (1 source line → 7 generated lines)
func TestSourceMapGeneration(t *testing.T) {
	input := `package main

func readConfig(path string) ([]byte, error) {
	let data = ReadFile(path)?
	return data, nil
}`

	p := New([]byte(input))
	_, sourceMap, err := p.Process()
	if err != nil {
		t.Fatalf("preprocessing failed: %v", err)
	}

	// The error propagation on line 4 should generate 7 output lines
	// All 7 lines should map back to original line 4

	// Expected mappings (line 4 in input → lines 4-10 in output):
	// Line 4: __tmp0, __err0 := ReadFile(path)
	// Line 5: // dingo:s:1
	// Line 6: if __err0 != nil {
	// Line 7:     return nil, __err0
	// Line 8: }
	// Line 9: // dingo:e:1
	// Line 10: var data = __tmp0

	expectedMappings := []struct {
		originalLine  int
		generatedLine int
	}{
		{4, 4},  // __tmp0, __err0 := ReadFile(path)
		{4, 5},  // // dingo:s:1
		{4, 6},  // if __err0 != nil {
		{4, 7},  // return nil, __err0
		{4, 8},  // }
		{4, 9},  // // dingo:e:1
		{4, 10}, // var data = __tmp0
	}

	if len(sourceMap.Mappings) != len(expectedMappings) {
		t.Errorf("expected %d mappings, got %d", len(expectedMappings), len(sourceMap.Mappings))
		for i, m := range sourceMap.Mappings {
			t.Logf("Mapping %d: orig=%d gen=%d", i, m.OriginalLine, m.GeneratedLine)
		}
		return
	}

	for i, expected := range expectedMappings {
		mapping := sourceMap.Mappings[i]
		if mapping.OriginalLine != expected.originalLine {
			t.Errorf("mapping %d: expected original line %d, got %d",
				i, expected.originalLine, mapping.OriginalLine)
		}
		if mapping.GeneratedLine != expected.generatedLine {
			t.Errorf("mapping %d: expected generated line %d, got %d",
				i, expected.generatedLine, mapping.GeneratedLine)
		}
	}
}

// TestSourceMapMultipleExpansions verifies source maps when multiple
// error propagations occur in the same function
func TestSourceMapMultipleExpansions(t *testing.T) {
	input := `package main

func process(path string) ([]byte, error) {
	let data = ReadFile(path)?
	let result = Process(data)?
	return result, nil
}`

	p := New([]byte(input))
	_, sourceMap, err := p.Process()
	if err != nil {
		t.Fatalf("preprocessing failed: %v", err)
	}

	// Line 4 expands to 7 lines (4-10)
	// Line 5 expands to 7 lines (11-17)
	// Total: 14 mappings

	if len(sourceMap.Mappings) != 14 {
		t.Errorf("expected 14 mappings (7+7), got %d", len(sourceMap.Mappings))
		for i, m := range sourceMap.Mappings {
			t.Logf("Mapping %d: orig=%d gen=%d", i, m.OriginalLine, m.GeneratedLine)
		}
		return
	}

	// First expansion: line 4 → lines 4-10
	for i := 0; i < 7; i++ {
		mapping := sourceMap.Mappings[i]
		if mapping.OriginalLine != 4 {
			t.Errorf("mapping %d: expected original line 4, got %d", i, mapping.OriginalLine)
		}
		expectedGenLine := 4 + i
		if mapping.GeneratedLine != expectedGenLine {
			t.Errorf("mapping %d: expected generated line %d, got %d",
				i, expectedGenLine, mapping.GeneratedLine)
		}
	}

	// Second expansion: line 5 → lines 11-17
	for i := 7; i < 14; i++ {
		mapping := sourceMap.Mappings[i]
		if mapping.OriginalLine != 5 {
			t.Errorf("mapping %d: expected original line 5, got %d", i, mapping.OriginalLine)
		}
		expectedGenLine := 11 + (i - 7)
		if mapping.GeneratedLine != expectedGenLine {
			t.Errorf("mapping %d: expected generated line %d, got %d",
				i, expectedGenLine, mapping.GeneratedLine)
		}
	}
}

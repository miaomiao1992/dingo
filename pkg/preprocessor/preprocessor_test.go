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
	return __tmp0
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

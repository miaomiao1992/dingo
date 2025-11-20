package preprocessor

import (
	"strings"
	"testing"
)

func TestNestedMatchBlocks(t *testing.T) {
	source := `package main

func processNestedMatch(result Result) int {
	return match result {
		Ok(inner) => {
			match inner {
				Some(val) => val,
				None => 0
			}
		},
		Err(e) => -1
	}
}
`

	processor := NewRustMatchProcessor()
	output, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	outputStr := string(output)
	t.Logf("Output:\n%s", outputStr)

	// The inner match should be transformed, not left as "match inner"
	if strings.Contains(outputStr, "match inner") {
		t.Errorf("Nested match block was not transformed! Output still contains 'match inner'")
	}

	// Should have two switch statements (outer and inner)
	switchCount := strings.Count(outputStr, "switch ")
	if switchCount < 2 {
		t.Errorf("Expected at least 2 switch statements (outer + inner), got %d", switchCount)
	}
}

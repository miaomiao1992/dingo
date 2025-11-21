package preprocessor

import (
	"strings"
	"testing"
)

func TestErrorPropProcessorV2_Metadata(t *testing.T) {
	proc := NewErrorPropProcessor()

	input := `package main

func readConfig(path string) ([]byte, error) {
	let data = os.ReadFile(path)?
	return data, nil
}`

	// Test ProcessV2 with PostAST mode
	result, err := proc.ProcessV2([]byte(input), ModePostAST)
	if err != nil {
		t.Fatalf("ProcessV2 failed: %v", err)
	}

	// Verify metadata was generated
	if len(result.Metadata) != 1 {
		t.Errorf("Expected 1 metadata entry, got %d", len(result.Metadata))
	}

	// Verify metadata content
	if len(result.Metadata) > 0 {
		meta := result.Metadata[0]
		if meta.Type != "error_prop" {
			t.Errorf("Expected type 'error_prop', got '%s'", meta.Type)
		}
		if meta.OriginalText != "?" {
			t.Errorf("Expected original text '?', got '%s'", meta.OriginalText)
		}
		if meta.ASTNodeType != "IfStmt" {
			t.Errorf("Expected AST node type 'IfStmt', got '%s'", meta.ASTNodeType)
		}
		if !strings.HasPrefix(meta.GeneratedMarker, "// dingo:e:") {
			t.Errorf("Expected marker to start with '// dingo:e:', got '%s'", meta.GeneratedMarker)
		}
	}

	// Verify marker was inserted in generated code
	output := string(result.Source)
	if !strings.Contains(output, "// dingo:e:0") {
		t.Errorf("Expected marker '// dingo:e:0' in output, got:\n%s", output)
	}
}

func TestErrorPropProcessorV2_UniqueMarkers(t *testing.T) {
	proc := NewErrorPropProcessor()

	input := `package main

func processFiles(path1, path2 string) error {
	let data1 = os.ReadFile(path1)?
	let data2 = os.ReadFile(path2)?
	return nil
}`

	// Test ProcessV2 with PostAST mode
	result, err := proc.ProcessV2([]byte(input), ModePostAST)
	if err != nil {
		t.Fatalf("ProcessV2 failed: %v", err)
	}

	// Verify two metadata entries
	if len(result.Metadata) != 2 {
		t.Errorf("Expected 2 metadata entries, got %d", len(result.Metadata))
	}

	// Verify unique markers
	output := string(result.Source)
	if !strings.Contains(output, "// dingo:e:0") {
		t.Errorf("Expected marker '// dingo:e:0' in output")
	}
	if !strings.Contains(output, "// dingo:e:1") {
		t.Errorf("Expected marker '// dingo:e:1' in output")
	}
}

func TestErrorPropProcessorV2_DualMode(t *testing.T) {
	proc := NewErrorPropProcessor()

	input := `package main

func readConfig(path string) ([]byte, error) {
	let data = os.ReadFile(path)?
	return data, nil
}`

	// Test ProcessV2 with Dual mode
	result, err := proc.ProcessV2([]byte(input), ModeDual)
	if err != nil {
		t.Fatalf("ProcessV2 failed: %v", err)
	}

	// Verify both mappings and metadata were generated
	if len(result.Mappings) == 0 {
		t.Error("Expected mappings in Dual mode, got none")
	}
	if len(result.Metadata) == 0 {
		t.Error("Expected metadata in Dual mode, got none")
	}
}

func TestErrorPropProcessorV2_LegacyMode(t *testing.T) {
	proc := NewErrorPropProcessor()

	input := `package main

func readConfig(path string) ([]byte, error) {
	let data = os.ReadFile(path)?
	return data, nil
}`

	// Test ProcessV2 with Legacy mode
	result, err := proc.ProcessV2([]byte(input), ModeLegacy)
	if err != nil {
		t.Fatalf("ProcessV2 failed: %v", err)
	}

	// Verify only mappings were generated (no metadata)
	if len(result.Mappings) == 0 {
		t.Error("Expected mappings in Legacy mode, got none")
	}
	if len(result.Metadata) != 0 {
		t.Errorf("Expected no metadata in Legacy mode, got %d entries", len(result.Metadata))
	}
}

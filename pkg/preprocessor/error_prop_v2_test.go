package preprocessor

import (
	"strings"
	"testing"
)

func TestErrorPropProcessor_Metadata(t *testing.T) {
	proc := NewErrorPropProcessor()

	input := `package main

func readConfig(path string) ([]byte, error) {
	let data = os.ReadFile(path)?
	return data, nil
}`

	// Test Process
	result, metadata, err := proc.ProcessInternal(input)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Verify metadata was generated
	if len(metadata) != 1 {
		t.Errorf("Expected 1 metadata entry, got %d", len(metadata))
	}

	// Verify metadata content
	if len(metadata) > 0 {
		meta := metadata[0]
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
	if !strings.Contains(result, "// dingo:e:0") {
		t.Errorf("Expected marker '// dingo:e:0' in output, got:\n%s", result)
	}
}

func TestErrorPropProcessor_UniqueMarkers(t *testing.T) {
	proc := NewErrorPropProcessor()

	input := `package main

func processFiles(path1, path2 string) error {
	let data1 = os.ReadFile(path1)?
	let data2 = os.ReadFile(path2)?
	return nil
}`

	// Test Process
	result, metadata, err := proc.ProcessInternal(input)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Verify two metadata entries
	if len(metadata) != 2 {
		t.Errorf("Expected 2 metadata entries, got %d", len(metadata))
	}

	// Verify unique markers
	if !strings.Contains(result, "// dingo:e:0") {
		t.Errorf("Expected marker '// dingo:e:0' in output")
	}
	if !strings.Contains(result, "// dingo:e:1") {
		t.Errorf("Expected marker '// dingo:e:1' in output")
	}
}

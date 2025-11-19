package preprocessor

import (
	"encoding/json"
	"testing"
)

// TestSourceMapRoundTrip validates that source maps are bidirectional
// This is critical for LSP and debugging support
func TestSourceMapRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		mappings []Mapping
		tests    []roundTripTest
	}{
		{
			name: "Simple error propagation",
			mappings: []Mapping{
				{
					Name:            "expr_mapping",
					OriginalLine:    10,
					OriginalColumn:  5,
					GeneratedLine:   15,
					GeneratedColumn: 10,
					Length:          12,
				},
				{
					Name:            "error_prop",
					OriginalLine:    10,
					OriginalColumn:  17, // position of ?
					GeneratedLine:   16,
					GeneratedColumn: 1,
					Length:          1,
				},
			},
			tests: []roundTripTest{
				{
					desc:            "Expression start",
					origLine:        10,
					origCol:         5,
					expectGenLine:   15,
					expectGenCol:    10,
					expectRoundTrip: true,
				},
				{
					desc:            "Expression middle",
					origLine:        10,
					origCol:         10,
					expectGenLine:   15,
					expectGenCol:    15,
					expectRoundTrip: true,
				},
				{
					desc:            "Error operator",
					origLine:        10,
					origCol:         17,
					expectGenLine:   16,
					expectGenCol:    1,
					expectRoundTrip: true,
				},
			},
		},
		{
			name: "Pattern matching",
			mappings: []Mapping{
				{
					Name:            "pattern_match",
					OriginalLine:    20,
					OriginalColumn:  1,
					GeneratedLine:   50,
					GeneratedColumn: 5,
					Length:          25,
				},
			},
			tests: []roundTripTest{
				{
					desc:            "Pattern start",
					origLine:        20,
					origCol:         1,
					expectGenLine:   50,
					expectGenCol:    5,
					expectRoundTrip: true,
				},
				{
					desc:            "Pattern end",
					origLine:        20,
					origCol:         25,
					expectGenLine:   50,
					expectGenCol:    29,
					expectRoundTrip: true,
				},
			},
		},
		{
			name: "Multi-line expansion",
			mappings: []Mapping{
				{
					Name:            "type_annotation",
					OriginalLine:    5,
					OriginalColumn:  10,
					GeneratedLine:   5,
					GeneratedColumn: 10,
					Length:          20,
				},
				// Multiple generated lines can map back to same original line
				{
					Name:            "error_prop_line1",
					OriginalLine:    5,
					OriginalColumn:  30,
					GeneratedLine:   10,
					GeneratedColumn: 1,
					Length:          1,
				},
				{
					Name:            "error_prop_line2",
					OriginalLine:    5,
					OriginalColumn:  30,
					GeneratedLine:   11,
					GeneratedColumn: 5,
					Length:          1,
				},
			},
			tests: []roundTripTest{
				{
					desc:            "Original annotation",
					origLine:        5,
					origCol:         15,
					expectGenLine:   5,
					expectGenCol:    15,
					expectRoundTrip: true,
				},
				{
					desc:            "Generated line 1 maps back",
					origLine:        5,
					origCol:         30,
					expectGenLine:   10,
					expectGenCol:    1,
					expectRoundTrip: false, // Multiple gen positions for same orig
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewSourceMap()
			for _, m := range tt.mappings {
				sm.AddMapping(m)
			}

			for _, rt := range tt.tests {
				t.Run(rt.desc, func(t *testing.T) {
					// Forward: Original → Generated
					genLine, genCol := sm.MapToGenerated(rt.origLine, rt.origCol)

					if genLine != rt.expectGenLine || genCol != rt.expectGenCol {
						t.Errorf("MapToGenerated(%d, %d) = (%d, %d), want (%d, %d)",
							rt.origLine, rt.origCol, genLine, genCol,
							rt.expectGenLine, rt.expectGenCol)
					}

					// Reverse: Generated → Original (if round-trip expected)
					if rt.expectRoundTrip {
						origLine, origCol := sm.MapToOriginal(genLine, genCol)

						if origLine != rt.origLine || origCol != rt.origCol {
							t.Errorf("Round-trip failed: MapToOriginal(%d, %d) = (%d, %d), want (%d, %d)",
								genLine, genCol, origLine, origCol, rt.origLine, rt.origCol)
						}
					}
				})
			}
		})
	}
}

type roundTripTest struct {
	desc            string
	origLine        int
	origCol         int
	expectGenLine   int
	expectGenCol    int
	expectRoundTrip bool
}

// TestSourceMapJSONSerialization validates JSON round-trip
func TestSourceMapJSONSerialization(t *testing.T) {
	original := NewSourceMap()
	original.DingoFile = "/path/to/test.dingo"
	original.GoFile = "/path/to/test.go"

	original.AddMapping(Mapping{
		Name:            "test_mapping",
		OriginalLine:    10,
		OriginalColumn:  5,
		GeneratedLine:   20,
		GeneratedColumn: 15,
		Length:          25,
	})

	// Serialize
	jsonData, err := original.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() failed: %v", err)
	}

	// Validate JSON structure
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	// Check required fields
	if parsed["version"] != float64(1) {
		t.Errorf("version = %v, want 1", parsed["version"])
	}

	if parsed["dingo_file"] != original.DingoFile {
		t.Errorf("dingo_file = %v, want %v", parsed["dingo_file"], original.DingoFile)
	}

	// Deserialize
	restored, err := FromJSON(jsonData)
	if err != nil {
		t.Fatalf("FromJSON() failed: %v", err)
	}

	// Validate restoration
	if restored.Version != original.Version {
		t.Errorf("Version mismatch: %d != %d", restored.Version, original.Version)
	}

	if restored.DingoFile != original.DingoFile {
		t.Errorf("DingoFile mismatch: %s != %s", restored.DingoFile, original.DingoFile)
	}

	if len(restored.Mappings) != len(original.Mappings) {
		t.Errorf("Mappings count mismatch: %d != %d", len(restored.Mappings), len(original.Mappings))
	}

	// Validate mapping fidelity
	for i := range original.Mappings {
		orig := original.Mappings[i]
		rest := restored.Mappings[i]

		if orig.Name != rest.Name ||
		   orig.OriginalLine != rest.OriginalLine ||
		   orig.OriginalColumn != rest.OriginalColumn ||
		   orig.GeneratedLine != rest.GeneratedLine ||
		   orig.GeneratedColumn != rest.GeneratedColumn ||
		   orig.Length != rest.Length {
			t.Errorf("Mapping %d mismatch:\noriginal: %+v\nrestored: %+v", i, orig, rest)
		}
	}
}

// TestSourceMapVersionCompatibility ensures version validation works
func TestSourceMapVersionCompatibility(t *testing.T) {
	// Current version should work
	validJSON := `{
		"version": 1,
		"dingo_file": "/path/to/test.dingo",
		"go_file": "/path/to/test.go",
		"mappings": []
	}`

	_, err := FromJSON([]byte(validJSON))
	if err != nil {
		t.Errorf("Version 1 should be valid: %v", err)
	}

	// Future versions should parse (forward compatibility)
	futureJSON := `{
		"version": 2,
		"dingo_file": "/path/to/test.dingo",
		"go_file": "/path/to/test.go",
		"mappings": [],
		"future_field": "ignored"
	}`

	sm, err := FromJSON([]byte(futureJSON))
	if err != nil {
		t.Errorf("Future version should parse: %v", err)
	}

	// Should still have basic fields
	if sm.Version != 2 {
		t.Errorf("Version = %d, want 2", sm.Version)
	}

	// Invalid JSON should fail
	invalidJSON := `{"version": "not a number"}`
	_, err = FromJSON([]byte(invalidJSON))
	if err == nil {
		t.Error("Invalid JSON should fail to parse")
	}
}

// TestSourceMapMerge validates merging multiple source maps
func TestSourceMapMerge(t *testing.T) {
	sm1 := NewSourceMap()
	sm1.AddMapping(Mapping{
		Name:            "preprocessor1",
		OriginalLine:    10,
		OriginalColumn:  5,
		GeneratedLine:   15,
		GeneratedColumn: 10,
		Length:          10,
	})

	sm2 := NewSourceMap()
	sm2.AddMapping(Mapping{
		Name:            "preprocessor2",
		OriginalLine:    20,
		OriginalColumn:  10,
		GeneratedLine:   25,
		GeneratedColumn: 15,
		Length:          15,
	})

	// Merge sm2 into sm1
	sm1.Merge(sm2)

	// Should have both mappings
	if len(sm1.Mappings) != 2 {
		t.Errorf("Merged map should have 2 mappings, got %d", len(sm1.Mappings))
	}

	// Both mappings should work
	line, col := sm1.MapToOriginal(15, 12)
	if line != 10 || col != 7 {
		t.Errorf("First mapping failed after merge: (%d, %d)", line, col)
	}

	line, col = sm1.MapToOriginal(25, 20)
	if line != 20 || col != 15 {
		t.Errorf("Second mapping failed after merge: (%d, %d)", line, col)
	}
}

// TestSourceMapAccuracy validates mapping accuracy for common patterns
func TestSourceMapAccuracy(t *testing.T) {
	tests := []struct {
		name     string
		mappings []Mapping
		queries  []accuracyTest
	}{
		{
			name: "Error propagation accuracy",
			mappings: []Mapping{
				{
					Name:            "expr_mapping",
					OriginalLine:    1,
					OriginalColumn:  10,
					GeneratedLine:   5,
					GeneratedColumn: 15,
					Length:          20, // "someFunction(args)"
				},
				{
					Name:            "error_prop",
					OriginalLine:    1,
					OriginalColumn:  30, // position of ?
					GeneratedLine:   6,
					GeneratedColumn: 5,
					Length:          1,
				},
			},
			queries: []accuracyTest{
				{
					desc:       "Error at function name maps to expression",
					genLine:    5,
					genCol:     18, // Inside "someFunction"
					wantLine:   1,
					wantCol:    13, // Corresponding position in original
					maxError:   2,  // Allow 2 column deviation
				},
				{
					desc:       "Error at error check maps to ? operator",
					genLine:    6,
					genCol:     5,
					wantLine:   1,
					wantCol:    30,
					maxError:   0, // Exact match for error_prop
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewSourceMap()
			for _, m := range tt.mappings {
				sm.AddMapping(m)
			}

			for _, query := range tt.queries {
				t.Run(query.desc, func(t *testing.T) {
					gotLine, gotCol := sm.MapToOriginal(query.genLine, query.genCol)

					lineError := abs(gotLine - query.wantLine)
					colError := abs(gotCol - query.wantCol)

					if lineError > 0 {
						t.Errorf("Line mapping error: got %d, want %d (diff: %d)",
							gotLine, query.wantLine, lineError)
					}

					if colError > query.maxError {
						t.Errorf("Column mapping error: got %d, want %d (diff: %d, max allowed: %d)",
							gotCol, query.wantCol, colError, query.maxError)
					}
				})
			}
		})
	}
}

type accuracyTest struct {
	desc     string
	genLine  int
	genCol   int
	wantLine int
	wantCol  int
	maxError int // Maximum allowed column deviation
}

// TestSourceMapEmptyMapping validates behavior with no mappings
func TestSourceMapEmptyMapping(t *testing.T) {
	sm := NewSourceMap()

	// Identity mapping for unmapped positions
	tests := []struct {
		line, col int
	}{
		{1, 1},
		{10, 25},
		{100, 200},
	}

	for _, tt := range tests {
		line, col := sm.MapToOriginal(tt.line, tt.col)
		if line != tt.line || col != tt.col {
			t.Errorf("Empty map should return identity: got (%d, %d), want (%d, %d)",
				line, col, tt.line, tt.col)
		}

		line, col = sm.MapToGenerated(tt.line, tt.col)
		if line != tt.line || col != tt.col {
			t.Errorf("Empty map forward should return identity: got (%d, %d), want (%d, %d)",
				line, col, tt.line, tt.col)
		}
	}
}

// TestSourceMapOverlappingMappings validates handling of overlapping ranges
func TestSourceMapOverlappingMappings(t *testing.T) {
	sm := NewSourceMap()

	// Overlapping mappings on same line (should prefer exact match)
	sm.AddMapping(Mapping{
		Name:            "mapping1",
		OriginalLine:    10,
		OriginalColumn:  10,
		GeneratedLine:   20,
		GeneratedColumn: 10,
		Length:          20,
	})

	sm.AddMapping(Mapping{
		Name:            "mapping2",
		OriginalLine:    10,
		OriginalColumn:  15,
		GeneratedLine:   20,
		GeneratedColumn: 15,
		Length:          10,
	})

	// Position 22 is in both ranges, should prefer exact match from mapping2
	line, col := sm.MapToOriginal(20, 22)

	// mapping2: gen_col=15, length=10 → covers 15-24
	// Position 22 with offset 7 → orig_col = 15+7 = 22
	if line != 10 || col != 22 {
		t.Errorf("Overlapping mappings: got (%d, %d), want (10, 22)", line, col)
	}
}

// Benchmark source map operations
func BenchmarkMapToOriginal(b *testing.B) {
	sm := NewSourceMap()

	// Simulate realistic source map with 100 mappings
	for i := 0; i < 100; i++ {
		sm.AddMapping(Mapping{
			Name:            "mapping",
			OriginalLine:    i + 1,
			OriginalColumn:  10,
			GeneratedLine:   i*2 + 1,
			GeneratedColumn: 15,
			Length:          20,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.MapToOriginal(50, 20)
	}
}

func BenchmarkMapToGenerated(b *testing.B) {
	sm := NewSourceMap()

	for i := 0; i < 100; i++ {
		sm.AddMapping(Mapping{
			Name:            "mapping",
			OriginalLine:    i + 1,
			OriginalColumn:  10,
			GeneratedLine:   i*2 + 1,
			GeneratedColumn: 15,
			Length:          20,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.MapToGenerated(25, 15)
	}
}

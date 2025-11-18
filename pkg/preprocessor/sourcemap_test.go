package preprocessor

import (
	"testing"
)

func TestSourceMapExpressionPriority(t *testing.T) {
	// Test case: When both expr_mapping and error_prop exist on same line,
	// and position is closer to expr_mapping, it should be chosen over error_prop

	sm := NewSourceMap()

	// Add mappings that simulate error propagation expansion
	// Line 10: ReadFile(path)? would generate:
	// - expr_mapping for "ReadFile(path)" at column 15, length 12
	// - error_prop for "?" at column 27, length 1
	sm.AddMapping(Mapping{
		Name:            "expr_mapping",
		GeneratedLine:   100,
		GeneratedColumn: 15,
		OriginalLine:    10,
		OriginalColumn:  15,
		Length:          12, // "ReadFile(path)"
	})

	sm.AddMapping(Mapping{
		Name:            "error_prop",
		GeneratedLine:   100,
		GeneratedColumn: 1,
		OriginalLine:    10,
		OriginalColumn:  27, // position of ?
		Length:          1,
	})

	tests := []struct {
		name           string
		line, col      int
		expLine, expCol int
	}{
		{
			name:     "Error position within expression should map to expression",
			line:     100,
			col:      18, // Inside "ReadFile(path)"
			expLine:  10,
			expCol:   18, // Should map to same position in expression
		},
		{
			name:     "Error at start of expression should map to expression",
			line:     100,
			col:      15, // Start of "ReadFile(path)"
			expLine:  10,
			expCol:   15, // Should map to start of expression
		},
		{
			name:     "Error near expression but within reasonable range",
			line:     100,
			col:      25, // Just before ? operator
			expLine:  10,
			expCol:   25, // Should still map to expression if reasonable
		},
		{
			name:     "Error very close to ? should use error_prop",
			line:     100,
			col:      27, // At ? position
			expLine:  10,
			expCol:   27, // Should map to ?
		},
		{
			name:     "Error very far from expression should use fallback",
			line:     100,
			col:      50, // Far from both mappings
			expLine:  10,
			expCol:   15, // Should fallback to expression start
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLine, gotCol := sm.MapToOriginal(tt.line, tt.col)

			if gotLine != tt.expLine || gotCol != tt.expCol {
				t.Errorf("MapToOriginal(%d, %d) = (%d, %d), want (%d, %d)",
					tt.line, tt.col, gotLine, gotCol, tt.expLine, tt.expCol)
			}

			// Verify with debug mode produces same result
			debugLine, debugCol := sm.MapToOriginalWithDebug(tt.line, tt.col, true)
			if debugLine != gotLine || debugCol != gotCol {
				t.Errorf("Debug mode produced different result: (%d, %d) vs (%d, %d)",
					debugLine, debugCol, gotLine, gotCol)
			}
		})
	}
}

func TestSourceMapEdgeCases(t *testing.T) {
	sm := NewSourceMap()

	// Test empty source map
	line, col := sm.MapToOriginal(1, 1)
	if line != 1 || col != 1 {
		t.Errorf("Empty source map should return identity mapping, got (%d, %d)", line, col)
	}

	// Test single mapping
	sm.AddMapping(Mapping{
		Name:            "test_mapping",
		GeneratedLine:   5,
		GeneratedColumn:  10,
		OriginalLine:    2,
		OriginalColumn:  5,
		Length:          15,
	})

	// Exact match test
	line, col = sm.MapToOriginal(5, 12)
	if line != 2 || col != 7 { // 5 + (12-10)
		t.Errorf("Exact match failed, got (%d, %d)", line, col)
	}

	// Outside mapping range with reasonable distance
	line, col = sm.MapToOriginal(5, 25)
	if line != 2 || col != 5 {
		t.Errorf("Outside range should return mapping start, got (%d, %d)", line, col)
	}
}

func TestSourceMapMultiLinePriority(t *testing.T) {
	// Test that mappings on different lines don't interfere
	sm := NewSourceMap()

	// Multiple lines with various mappings
	sm.AddMapping(Mapping{
		Name:            "expr_mapping",
		GeneratedLine:   10,
		GeneratedColumn:  15,
		OriginalLine:    10,
		OriginalColumn:  15,
		Length:          10,
	})

	sm.AddMapping(Mapping{
		Name:            "error_prop",
		GeneratedLine:   11,
		GeneratedColumn:  1,
		OriginalLine:    10,
		OriginalColumn:  26,
		Length:          1,
	})

	// Error on line 10 should use expr_mapping
	line, col := sm.MapToOriginal(10, 18)
	if line != 10 || col != 18 {
		t.Errorf("Line 10 should use expr_mapping, got (%d, %d)", line, col)
	}

	// Error on line 11 should use error_prop
	line, col = sm.MapToOriginal(11, 1)
	if line != 10 || col != 26 {
		t.Errorf("Line 11 should use error_prop, got (%d, %d)", line, col)
	}
}

func TestSourceMapOffsetBounds(t *testing.T) {
	// Test that offset calculations respect reasonable bounds
	sm := NewSourceMap()

	sm.AddMapping(Mapping{
		Name:            "test_mapping",
		GeneratedLine:   10,
		GeneratedColumn:  10,
		OriginalLine:    10,
		OriginalColumn:  10,
		Length:          5,
	})

	tests := []struct {
		name     string
		col      int
		expLine   int
		expCol    int
	}{
		{
			name:   "Within reasonable offset",
			col:    12, // 2 chars into mapping
			expLine: 10,
			expCol:  12, // 10 + 2
		},
		{
			name:   "Just beyond mapping length",
			col:    14, // Just beyond 5-char mapping
			expLine: 10,
			expCol:  14, // Should still apply offset if reasonable
		},
		{
			name:   "Far beyond mapping range",
			col:    50, // Too far from mapping
			expLine: 10,
			expCol:  10, // Should fallback to mapping start
		},
		{
			name:   "Before mapping start",
			col:    5, // Before mapping start
			expLine: 10,
			expCol:  10, // Should fallback to mapping start
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLine, gotCol := sm.MapToOriginal(10, tt.col)

			if gotLine != tt.expLine || gotCol != tt.expCol {
				t.Errorf("Offset bounds test failed for %s: MapToOriginal(10, %d) = (%d, %d), want (%d, %d)",
					tt.name, tt.col, gotLine, gotCol, tt.expLine, tt.expCol)
			}
		})
	}
}
// Package preprocessor implements the Dingo source code preprocessor
// that transforms Dingo syntax to valid Go syntax with semantic placeholders
package preprocessor

import (
	"encoding/json"
	"fmt"
)

// SourceMap tracks position mappings between original Dingo source
// and preprocessed Go source for error reporting and LSP integration
type SourceMap struct {
	Mappings []Mapping `json:"mappings"`
}

// Mapping represents a single position mapping
type Mapping struct {
	// Preprocessed (generated) position
	GeneratedLine   int `json:"generated_line"`
	GeneratedColumn int `json:"generated_column"`

	// Original (Dingo) position
	OriginalLine    int `json:"original_line"`
	OriginalColumn  int `json:"original_column"`

	// Length of the mapped segment
	Length int `json:"length"`

	// Optional name/description for debugging
	Name string `json:"name,omitempty"`
}

// NewSourceMap creates a new empty source map
func NewSourceMap() *SourceMap {
	return &SourceMap{
		Mappings: make([]Mapping, 0),
	}
}

// AddMapping adds a new position mapping
func (sm *SourceMap) AddMapping(m Mapping) {
	sm.Mappings = append(sm.Mappings, m)
}

// MapToOriginal maps a preprocessed position to the original Dingo position
// Returns the mapped position or the input position if no mapping found
func (sm *SourceMap) MapToOriginal(line, col int) (int, int) {
	// Find the mapping that contains this position
	for _, m := range sm.Mappings {
		if m.GeneratedLine == line &&
		   col >= m.GeneratedColumn &&
		   col < m.GeneratedColumn+m.Length {
			// Calculate offset within the mapping
			offset := col - m.GeneratedColumn
			return m.OriginalLine, m.OriginalColumn + offset
		}
	}

	// No mapping found, return as-is
	return line, col
}

// MapToGenerated maps an original Dingo position to the preprocessed position
// Returns the mapped position or the input position if no mapping found
func (sm *SourceMap) MapToGenerated(line, col int) (int, int) {
	// Find the mapping that contains this position
	for _, m := range sm.Mappings {
		if m.OriginalLine == line &&
		   col >= m.OriginalColumn &&
		   col < m.OriginalColumn+m.Length {
			// Calculate offset within the mapping
			offset := col - m.OriginalColumn
			return m.GeneratedLine, m.GeneratedColumn + offset
		}
	}

	// No mapping found, return as-is
	return line, col
}

// ToJSON serializes the source map to JSON
func (sm *SourceMap) ToJSON() ([]byte, error) {
	return json.MarshalIndent(sm, "", "  ")
}

// FromJSON deserializes a source map from JSON
func FromJSON(data []byte) (*SourceMap, error) {
	var sm SourceMap
	if err := json.Unmarshal(data, &sm); err != nil {
		return nil, fmt.Errorf("failed to parse source map: %w", err)
	}
	return &sm, nil
}

// Merge combines multiple source maps into one
// Useful when multiple preprocessors run in sequence
func (sm *SourceMap) Merge(other *SourceMap) {
	sm.Mappings = append(sm.Mappings, other.Mappings...)
}

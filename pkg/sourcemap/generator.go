// Package sourcemap provides source map generation for Dingo → Go transpilation
package sourcemap

import (
	"encoding/base64"
	"fmt"
	"go/token"

	"github.com/go-sourcemap/sourcemap"
)

// Generator collects position mappings during transpilation and generates source maps
type Generator struct {
	sourceFile string
	genFile    string
	mappings   []Mapping
}

// Mapping represents a single position mapping from Dingo source to Go output
type Mapping struct {
	// Source position (original .dingo file)
	SourceLine   int
	SourceColumn int

	// Generated position (output .go file)
	GenLine   int
	GenColumn int

	// Optional: Name at this position (for identifier mapping)
	Name string
}

// NewGenerator creates a new source map generator
func NewGenerator(sourceFile, genFile string) *Generator {
	return &Generator{
		sourceFile: sourceFile,
		genFile:    genFile,
		mappings:   make([]Mapping, 0),
	}
}

// AddMapping records a position mapping from source to generated code
func (g *Generator) AddMapping(src, gen token.Position) {
	g.mappings = append(g.mappings, Mapping{
		SourceLine:   src.Line,
		SourceColumn: src.Column,
		GenLine:      gen.Line,
		GenColumn:    gen.Column,
	})
}

// AddMappingWithName records a position mapping with an identifier name
func (g *Generator) AddMappingWithName(src, gen token.Position, name string) {
	g.mappings = append(g.mappings, Mapping{
		SourceLine:   src.Line,
		SourceColumn: src.Column,
		GenLine:      gen.Line,
		GenColumn:    gen.Column,
		Name:         name,
	})
}

// Generate creates a source map in JSON format
// TODO(Future): Implement full VLQ encoding for production use
// For now, we return a valid but basic source map structure
func (g *Generator) Generate() ([]byte, error) {
	// Sort mappings by generated position
	sortedMappings := make([]Mapping, len(g.mappings))
	copy(sortedMappings, g.mappings)

	// Simple bubble sort for now
	for i := 0; i < len(sortedMappings); i++ {
		for j := i + 1; j < len(sortedMappings); j++ {
			if sortedMappings[i].GenLine > sortedMappings[j].GenLine ||
				(sortedMappings[i].GenLine == sortedMappings[j].GenLine &&
					sortedMappings[i].GenColumn > sortedMappings[j].GenColumn) {
				sortedMappings[i], sortedMappings[j] = sortedMappings[j], sortedMappings[i]
			}
		}
	}

	// For now, return a skeleton source map
	// VLQ encoding will be added in future enhancement
	sourceMap := fmt.Sprintf(`{
  "version": 3,
  "file": "%s",
  "sourceRoot": "",
  "sources": ["%s"],
  "names": %s,
  "mappings": ""
}`, g.genFile, g.sourceFile, g.formatNames())

	return []byte(sourceMap), nil
}

// formatNames formats the names array as JSON
func (g *Generator) formatNames() string {
	names := g.collectNames()
	if len(names) == 0 {
		return "[]"
	}
	result := "["
	for i, name := range names {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf(`"%s"`, name)
	}
	result += "]"
	return result
}

// GenerateInline creates a base64-encoded inline source map comment
func (g *Generator) GenerateInline() (string, error) {
	data, err := g.Generate()
	if err != nil {
		return "", err
	}

	// Base64 encode
	encoded := base64.StdEncoding.EncodeToString(data)

	// Return as inline comment
	return fmt.Sprintf("//# sourceMappingURL=data:application/json;base64,%s", encoded), nil
}

// collectNames extracts unique identifier names from mappings
func (g *Generator) collectNames() []string {
	nameSet := make(map[string]bool)
	names := make([]string, 0)

	for _, m := range g.mappings {
		if m.Name != "" && !nameSet[m.Name] {
			nameSet[m.Name] = true
			names = append(names, m.Name)
		}
	}

	return names
}

// Consumer provides source map lookup functionality
type Consumer struct {
	sm *sourcemap.Consumer
}

// NewConsumer creates a source map consumer from raw source map data
func NewConsumer(data []byte) (*Consumer, error) {
	sm, err := sourcemap.Parse("", data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source map: %w", err)
	}

	return &Consumer{sm: sm}, nil
}

// Source looks up the original source position for a generated position
func (c *Consumer) Source(line, column int) (*token.Position, error) {
	// Note: go-sourcemap uses 0-based indexing, but we use 1-based
	file, _, line, col, ok := c.sm.Source(line-1, column-1)
	if !ok {
		return nil, fmt.Errorf("no mapping found for position %d:%d", line, column)
	}

	return &token.Position{
		Filename: file,
		Line:     line + 1,
		Column:   col + 1,
	}, nil
}

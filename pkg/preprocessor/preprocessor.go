// Package preprocessor transforms Dingo syntax to valid Go syntax
package preprocessor

import (
	"bytes"
	"fmt"
)

// Preprocessor orchestrates multiple feature processors to transform
// Dingo source code into valid Go code with semantic placeholders
type Preprocessor struct {
	source     []byte
	processors []FeatureProcessor
}

// FeatureProcessor defines the interface for individual feature preprocessors
type FeatureProcessor interface {
	// Name returns the feature name for logging/debugging
	Name() string

	// Process transforms the source code and returns:
	// - transformed source
	// - source mappings
	// - error if transformation failed
	Process(source []byte) ([]byte, []Mapping, error)
}

// New creates a new preprocessor with all registered features
func New(source []byte) *Preprocessor {
	return &Preprocessor{
		source: source,
		processors: []FeatureProcessor{
			// Order matters! Process in this sequence:
			// 0. Type annotations (: → space) - must be first
			NewTypeAnnotProcessor(),
			// 1. Error propagation (expr?)
			NewErrorPropProcessor(),
			// 2. Keywords (let → var) - after error prop so it doesn't interfere
			NewKeywordProcessor(),
			// 3. Lambdas (|x| expr)
			// NewLambdaProcessor(),
			// 4. Sum types (enum)
			// NewSumTypeProcessor(),
			// 5. Pattern matching (match)
			// NewPatternMatchProcessor(),
			// 6. Operators (ternary, ??, ?.)
			// NewOperatorProcessor(),
		},
	}
}

// Process runs all feature processors in sequence and combines source maps
func (p *Preprocessor) Process() (string, *SourceMap, error) {
	result := p.source
	sourceMap := NewSourceMap()

	// Run each processor in sequence
	for _, proc := range p.processors {
		processed, mappings, err := proc.Process(result)
		if err != nil {
			return "", nil, fmt.Errorf("%s preprocessing failed: %w", proc.Name(), err)
		}

		// Update result
		result = processed

		// Merge mappings
		for _, m := range mappings {
			sourceMap.AddMapping(m)
		}
	}

	return string(result), sourceMap, nil
}

// ProcessBytes is like Process but returns bytes
func (p *Preprocessor) ProcessBytes() ([]byte, *SourceMap, error) {
	str, sm, err := p.Process()
	if err != nil {
		return nil, nil, err
	}
	return []byte(str), sm, nil
}

// ScanContext provides utilities for scanning and transforming source code
type ScanContext struct {
	Source []byte
	Pos    int
	Line   int
	Column int
}

// NewScanContext creates a new scan context
func NewScanContext(source []byte) *ScanContext {
	return &ScanContext{
		Source: source,
		Pos:    0,
		Line:   1,
		Column: 1,
	}
}

// Peek returns the next byte without advancing
func (sc *ScanContext) Peek() byte {
	if sc.Pos >= len(sc.Source) {
		return 0
	}
	return sc.Source[sc.Pos]
}

// PeekN returns the next n bytes without advancing
func (sc *ScanContext) PeekN(n int) []byte {
	end := sc.Pos + n
	if end > len(sc.Source) {
		end = len(sc.Source)
	}
	return sc.Source[sc.Pos:end]
}

// Advance moves to the next byte
func (sc *ScanContext) Advance() byte {
	if sc.Pos >= len(sc.Source) {
		return 0
	}

	ch := sc.Source[sc.Pos]
	sc.Pos++

	if ch == '\n' {
		sc.Line++
		sc.Column = 1
	} else {
		sc.Column++
	}

	return ch
}

// SkipWhitespace skips whitespace and returns true if any was skipped
func (sc *ScanContext) SkipWhitespace() bool {
	skipped := false
	for sc.Pos < len(sc.Source) {
		ch := sc.Peek()
		if ch == ' ' || ch == '\t' || ch == '\r' {
			sc.Advance()
			skipped = true
		} else {
			break
		}
	}
	return skipped
}

// AtEnd returns true if at end of source
func (sc *ScanContext) AtEnd() bool {
	return sc.Pos >= len(sc.Source)
}

// Buffer is a helper for building transformed output
type Buffer struct {
	buf    bytes.Buffer
	line   int
	column int
}

// NewBuffer creates a new output buffer
func NewBuffer() *Buffer {
	return &Buffer{
		line:   1,
		column: 1,
	}
}

// Write writes bytes to the buffer
func (b *Buffer) Write(p []byte) {
	for _, ch := range p {
		b.buf.WriteByte(ch)
		if ch == '\n' {
			b.line++
			b.column = 1
		} else {
			b.column++
		}
	}
}

// WriteString writes a string to the buffer
func (b *Buffer) WriteString(s string) {
	b.Write([]byte(s))
}

// WriteByte writes a single byte
func (b *Buffer) WriteByte(ch byte) {
	b.buf.WriteByte(ch)
	if ch == '\n' {
		b.line++
		b.column = 1
	} else {
		b.column++
	}
}

// Bytes returns the buffer contents
func (b *Buffer) Bytes() []byte {
	return b.buf.Bytes()
}

// String returns the buffer as a string
func (b *Buffer) String() string {
	return b.buf.String()
}

// Position returns current line and column
func (b *Buffer) Position() (int, int) {
	return b.line, b.column
}

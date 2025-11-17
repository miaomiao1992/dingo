// Package preprocessor transforms Dingo syntax to valid Go syntax
package preprocessor

import (
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

package preprocessor

import (
	"regexp"
)

// Package-level compiled regex (Issue 2: Regex Performance)
var (
	letPattern = regexp.MustCompile(`\blet\s+`)
)

// KeywordProcessor converts Dingo keywords to Go keywords
type KeywordProcessor struct{}

// NewKeywordProcessor creates a new keyword processor
func NewKeywordProcessor() *KeywordProcessor {
	return &KeywordProcessor{}
}

// Name returns the processor name
func (k *KeywordProcessor) Name() string {
	return "keywords"
}

// Process transforms Dingo keywords to Go keywords
// Converts: let → var
func (k *KeywordProcessor) Process(source []byte) ([]byte, []Mapping, error) {
	// Replace `let ` with `var ` (with space to avoid matching `letter`)
	result := letPattern.ReplaceAll(source, []byte("var "))

	return result, nil, nil
}

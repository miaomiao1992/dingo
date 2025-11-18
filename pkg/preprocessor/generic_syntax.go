package preprocessor

import (
	"bytes"
	"regexp"
)

// GenericSyntaxProcessor converts Rust-style generic syntax (<>) to Go syntax ([])
// Transforms: Result<T, E> → Result[T, E]
// Transforms: Option<T> → Option[T]
type GenericSyntaxProcessor struct{}

// Pattern matches generic type declarations with angle brackets
// Matches: TypeName<...> where TypeName is a word (Result, Option, etc.)
// Strategy: Find word followed by <, match up to corresponding >, replace with []
var genericPattern = regexp.MustCompile(`\b([A-Z]\w*)<([^>]+)>`)

// NewGenericSyntaxProcessor creates a new generic syntax processor
func NewGenericSyntaxProcessor() *GenericSyntaxProcessor {
	return &GenericSyntaxProcessor{}
}

// Name returns the processor name
func (g *GenericSyntaxProcessor) Name() string {
	return "generic_syntax"
}

// Process transforms generic syntax from <> to []
func (g *GenericSyntaxProcessor) Process(source []byte) ([]byte, []Mapping, error) {
	// Replace all occurrences of TypeName<...> with TypeName[...]
	result := genericPattern.ReplaceAllFunc(source, func(match []byte) []byte {
		// Extract type name and generic parameters
		submatch := genericPattern.FindSubmatch(match)
		if len(submatch) != 3 {
			return match // Should not happen, but be safe
		}

		typeName := submatch[1]
		genericParams := submatch[2]

		// Build: TypeName[genericParams]
		var buf bytes.Buffer
		buf.Write(typeName)
		buf.WriteByte('[')
		buf.Write(genericParams)
		buf.WriteByte(']')

		return buf.Bytes()
	})

	// No source mappings needed since this is just bracket replacement
	return result, nil, nil
}

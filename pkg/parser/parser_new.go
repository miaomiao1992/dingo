// Package parser provides the new go/parser-based Dingo parser
package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/MadAppGang/dingo/pkg/preprocessor"
)

// NewGoParser creates a parser that uses go/parser with preprocessing
type NewGoParser struct {
	fset *token.FileSet
}

// NewGoParserInstance creates a new go/parser-based parser
func NewGoParserInstance() *NewGoParser {
	return &NewGoParser{
		fset: token.NewFileSet(),
	}
}

// ParseResult holds the parsing results
type ParseResult struct {
	AST       *ast.File
	SourceMap *preprocessor.SourceMap
	FileSet   *token.FileSet
}

// ParseFile parses a Dingo file using preprocessing + go/parser
func (p *NewGoParser) ParseFile(filename string, source []byte) (*ParseResult, error) {
	// Step 1: Preprocess Dingo source to valid Go
	prep := preprocessor.New(source)
	goSource, sourceMap, err := prep.ProcessBytes()
	if err != nil {
		return nil, fmt.Errorf("preprocessing failed: %w", err)
	}

	// Step 2: Parse with go/parser
	file, err := parser.ParseFile(p.fset, filename, goSource, parser.ParseComments)
	if err != nil {
		// Map error positions back to original file
		mappedErr := p.mapError(err, sourceMap)
		return nil, mappedErr
	}

	return &ParseResult{
		AST:       file,
		SourceMap: sourceMap,
		FileSet:   p.fset,
	}, nil
}

// mapError maps go/parser errors to original Dingo positions
func (p *NewGoParser) mapError(err error, sm *preprocessor.SourceMap) error {
	// TODO: Extract position from go/parser error
	// TODO: Map position using source map
	// TODO: Return error with corrected position

	// For now, return as-is
	return err
}

// FileSet returns the token.FileSet used by this parser
func (p *NewGoParser) FileSet() *token.FileSet {
	return p.fset
}

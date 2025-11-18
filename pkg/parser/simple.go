package parser

import (
	"go/parser"
	"go/token"

	dingoast "github.com/MadAppGang/dingo/pkg/ast"
	"github.com/MadAppGang/dingo/pkg/preprocessor"
)

type simpleParser struct {
	mode Mode
}

func newParticipleParser(mode Mode) Parser {
	return &simpleParser{mode: mode}
}

func (p *simpleParser) ParseFile(fset *token.FileSet, filename string, src []byte) (*dingoast.File, error) {
	// Step 1: Preprocess Dingo syntax to valid Go
	prep := preprocessor.New(src)
	goCode, _, err := prep.Process()
	if err != nil {
		return nil, err
	}

	// Step 2: Use go/parser to parse the preprocessed Go code
	var parserMode parser.Mode
	if p.mode&ParseComments != 0 {
		parserMode |= parser.ParseComments
	}
	if p.mode&AllErrors != 0 {
		parserMode |= parser.AllErrors
	}

	file, err := parser.ParseFile(fset, filename, []byte(goCode), parserMode)
	if err != nil {
		return nil, err
	}

	return &dingoast.File{File: file}, nil
}

func (p *simpleParser) ParseExpr(fset *token.FileSet, expr string) (dingoast.DingoNode, error) {
	// Not implemented for now
	return nil, nil
}

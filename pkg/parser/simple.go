package parser

import (
	"go/parser"
	"go/token"

	dingoast "github.com/MadAppGang/dingo/pkg/ast"
)

type simpleParser struct {
	mode Mode
}

func newParticipleParser(mode Mode) Parser {
	return &simpleParser{mode: mode}
}

func (p *simpleParser) ParseFile(fset *token.FileSet, filename string, src []byte) (*dingoast.File, error) {
	// Use go/parser directly since preprocessor already converted to Go
	var parserMode parser.Mode
	if p.mode&ParseComments != 0 {
		parserMode |= parser.ParseComments
	}
	if p.mode&AllErrors != 0 {
		parserMode |= parser.AllErrors
	}

	file, err := parser.ParseFile(fset, filename, src, parserMode)
	if err != nil {
		return nil, err
	}

	return &dingoast.File{File: file}, nil
}

func (p *simpleParser) ParseExpr(fset *token.FileSet, expr string) (dingoast.DingoNode, error) {
	// Not implemented for now
	return nil, nil
}

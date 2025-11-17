package ast

import "go/ast"

// File wraps go/ast.File for Dingo
type File struct {
	*ast.File
}

// DingoNode is a marker interface for Dingo AST nodes
type DingoNode interface {
	Node()
}

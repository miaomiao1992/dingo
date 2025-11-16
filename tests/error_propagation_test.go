package tests

import (
	"go/token"
	"testing"

	"github.com/MadAppGang/dingo/pkg/parser"
	dingoast "github.com/MadAppGang/dingo/pkg/ast"
)

func TestErrorPropagationQuestion(t *testing.T) {
	// Note: Using simplified syntax that the Phase 1 parser supports
	// Full Dingo syntax will be tested in Phase 1.5
	src := []byte(`package main

func fetchUser(id: int) string {
	return "user"
}

func processUser(id: int) string {
	let user = fetchUser(id)?
	return user
}
`)

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.dingo", src, 0)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if file == nil {
		t.Fatal("File is nil")
	}

	// Check that we have Dingo nodes
	if !file.HasDingoNodes() {
		t.Error("Expected Dingo nodes but found none")
	}

	// Verify the ErrorPropagationExpr was created
	found := false
	for _, node := range file.DingoNodes {
		if errExpr, ok := node.(*dingoast.ErrorPropagationExpr); ok {
			if errExpr.Syntax != dingoast.SyntaxQuestion {
				t.Errorf("Expected SyntaxQuestion, got %v", errExpr.Syntax)
			}
			found = true
			break
		}
	}

	if !found {
		t.Error("Did not find ErrorPropagationExpr in parsed file")
	}
}

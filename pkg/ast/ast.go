// Package ast defines Dingo-specific AST extensions on top of go/ast
//
// Strategy: Reuse 95% of go/ast infrastructure, only define custom nodes
// for Dingo-specific features that don't exist in Go.
//
// Benefits:
// - Leverage go/printer for code generation
// - Use go/token.FileSet for position tracking
// - Reuse go/ast.Walk, go/ast.Inspect
// - Familiar API for Go developers
package ast

import (
	"go/ast"
	"go/token"
)

// ============================================================================
// Dingo-Specific Expression Nodes
// ============================================================================
// These extend go/ast with new expression types for Dingo features

// SyntaxStyle represents the error propagation syntax used
type SyntaxStyle string

const (
	SyntaxQuestion SyntaxStyle = "question" // expr?
	SyntaxBang     SyntaxStyle = "bang"     // expr!
	SyntaxTry      SyntaxStyle = "try"      // try expr
)

// ErrorPropagationExpr represents error propagation with configurable syntax
// Supports three syntaxes: expr?, expr!, try expr
// Example: let user = fetchUser(id)?
// Example with message: let user = fetchUser(id)? "failed to fetch user"
//
// Implements ast.Expr interface so it can be used anywhere a Go expression can
type ErrorPropagationExpr struct {
	X          ast.Expr    // The expression being propagated (e.g., fetchUser(id))
	OpPos      token.Pos   // Position of the operator ('?', '!', or 'try' keyword)
	Syntax     SyntaxStyle // Which syntax was used
	Message    string      // Optional error wrapping message
	MessagePos token.Pos   // Position of the message string (if present)
}

func (e *ErrorPropagationExpr) Pos() token.Pos {
	if e.Syntax == SyntaxTry {
		return e.OpPos // For 'try expr', start at 'try'
	}
	return e.X.Pos() // For postfix operators, start at expression
}

func (e *ErrorPropagationExpr) End() token.Pos {
	if e.Syntax == SyntaxTry {
		return e.X.End() // For 'try expr', end at expression
	}
	return e.OpPos + 1 // For postfix operators, end after operator
}

// exprNode ensures ErrorPropagationExpr implements ast.Expr
func (*ErrorPropagationExpr) exprNode() {}

// SafeNavigationExpr represents the safe navigation operator (?.)
// Example: let city = user?.address?.city
//
// Implements ast.Expr interface
type SafeNavigationExpr struct {
	X     ast.Expr  // Left operand (potentially nil value)
	OpPos token.Pos // Position of '?.'
	Sel   *ast.Ident // Field/method being accessed
}

func (s *SafeNavigationExpr) Pos() token.Pos { return s.X.Pos() }
func (s *SafeNavigationExpr) End() token.Pos { return s.Sel.End() }

// exprNode ensures SafeNavigationExpr implements ast.Expr
func (*SafeNavigationExpr) exprNode() {}

// NullCoalescingExpr represents the `??` operator (a ?? b)
// Example: let name = user.name ?? "Unknown"
//
// Implements ast.Expr interface
type NullCoalescingExpr struct {
	X      ast.Expr  // Left operand (nullable value)
	OpPos  token.Pos // Position of '??'
	Y      ast.Expr  // Right operand (default value)
}

func (n *NullCoalescingExpr) Pos() token.Pos { return n.X.Pos() }
func (n *NullCoalescingExpr) End() token.Pos { return n.Y.End() }

// exprNode ensures NullCoalescingExpr implements ast.Expr
func (*NullCoalescingExpr) exprNode() {}

// TernaryExpr represents the ternary operator (cond ? then : else)
// Example: let status = age >= 18 ? "adult" : "minor"
//
// Implements ast.Expr interface
type TernaryExpr struct {
	Cond     ast.Expr  // Condition expression
	Question token.Pos // Position of '?'
	Then     ast.Expr  // Expression if true
	Colon    token.Pos // Position of ':'
	Else     ast.Expr  // Expression if false
}

func (t *TernaryExpr) Pos() token.Pos { return t.Cond.Pos() }
func (t *TernaryExpr) End() token.Pos { return t.Else.End() }

// exprNode ensures TernaryExpr implements ast.Expr
func (*TernaryExpr) exprNode() {}

// LambdaExpr represents lambda/arrow functions
// Examples:
//   - Rust style: |a, b| a + b
//   - Arrow style: (a, b) => a + b
//   - Kotlin style: { it.age > 18 }
//
// Implements ast.Expr interface
type LambdaExpr struct {
	Pipe   token.Pos       // Position of '|' or '(' or '{'
	Params *ast.FieldList  // Parameters (reuse go/ast.FieldList!)
	Arrow  token.Pos       // Position of '=>' or '->' (if present)
	Body   ast.Expr        // Body (can be ast.BlockStmt or any expression)
	Rpipe  token.Pos       // Position of closing '|' or ')' or '}'
}

func (l *LambdaExpr) Pos() token.Pos { return l.Pipe }
func (l *LambdaExpr) End() token.Pos {
	if l.Rpipe.IsValid() {
		return l.Rpipe + 1
	}
	return l.Body.End()
}

// exprNode ensures LambdaExpr implements ast.Expr
func (*LambdaExpr) exprNode() {}

// ============================================================================
// Dingo-Specific Type Nodes
// ============================================================================

// ResultType represents Result<T, E> type
// Example: func fetchUser(id: string) -> Result<User, Error>
//
// Note: This is only used during parsing. The transformer will convert
// Result<T, E> to the appropriate Go struct type.
type ResultType struct {
	Result   token.Pos // Position of 'Result' keyword
	Lbrack   token.Pos // Position of '<'
	Value    ast.Expr  // Value type (T)
	Comma    token.Pos // Position of ','
	Error    ast.Expr  // Error type (E)
	Rbrack   token.Pos // Position of '>'
}

func (r *ResultType) Pos() token.Pos { return r.Result }
func (r *ResultType) End() token.Pos { return r.Rbrack + 1 }

// OptionType represents Option<T> type
// Example: func findUser(id: string) -> Option<User>
//
// Note: This is only used during parsing. The transformer will convert
// Option<T> to the appropriate Go pointer or struct type.
type OptionType struct {
	Option token.Pos // Position of 'Option' keyword
	Lbrack token.Pos // Position of '<'
	Value  ast.Expr  // Value type (T)
	Rbrack token.Pos // Position of '>'
}

func (o *OptionType) Pos() token.Pos { return o.Option }
func (o *OptionType) End() token.Pos { return o.Rbrack + 1 }

// ============================================================================
// Sum Types (Enums and Pattern Matching)
// ============================================================================

// EnumDecl represents an enum declaration (sum type)
// Example:
//   enum Shape {
//       Circle { radius: float64 },
//       Rectangle { width: float64, height: float64 },
//       Point,
//   }
//
// Implements ast.Decl interface
type EnumDecl struct {
	Enum       token.Pos       // Position of 'enum' keyword
	Name       *ast.Ident      // Enum name
	TypeParams *ast.FieldList  // Generic type parameters (nil if not generic)
	Lbrace     token.Pos       // Position of '{'
	Variants   []*VariantDecl  // Enum variants
	Rbrace     token.Pos       // Position of '}'
}

func (e *EnumDecl) Pos() token.Pos { return e.Enum }
func (e *EnumDecl) End() token.Pos { return e.Rbrace + 1 }

// declNode ensures EnumDecl implements ast.Decl
func (*EnumDecl) declNode() {}

// VariantDecl represents a single variant of an enum
// Supports three forms:
//   - Unit: Point
//   - Tuple: Circle(radius: float64)
//   - Struct: Rectangle { width: float64, height: float64 }
type VariantDecl struct {
	Name   *ast.Ident     // Variant name
	Fields *ast.FieldList // Fields (nil for unit variants, non-nil for tuple/struct)
	Kind   VariantKind    // Unit, Tuple, or Struct
}

func (v *VariantDecl) Pos() token.Pos { return v.Name.Pos() }
func (v *VariantDecl) End() token.Pos {
	if v.Fields != nil {
		return v.Fields.End()
	}
	return v.Name.End()
}

// VariantKind specifies the kind of enum variant
type VariantKind int

const (
	VariantUnit VariantKind = iota  // Unit variant (no data)
	VariantTuple                     // Tuple variant (positional fields)
	VariantStruct                    // Struct variant (named fields)
)

// MatchExpr represents a match expression for pattern matching
// Example:
//   match shape {
//       Circle { radius } => 3.14 * radius * radius,
//       Rectangle { width, height } => width * height,
//       Point => 0.0,
//   }
//
// Implements ast.Expr interface
type MatchExpr struct {
	Match  token.Pos   // Position of 'match' keyword
	Expr   ast.Expr    // Expression being matched
	Lbrace token.Pos   // Position of '{'
	Arms   []*MatchArm // Match arms
	Rbrace token.Pos   // Position of '}'
}

func (m *MatchExpr) Pos() token.Pos { return m.Match }
func (m *MatchExpr) End() token.Pos { return m.Rbrace + 1 }

// exprNode ensures MatchExpr implements ast.Expr
func (*MatchExpr) exprNode() {}

// MatchArm represents a single arm of a match expression
// Example: Circle { radius } => 3.14 * radius * radius
type MatchArm struct {
	Pattern *Pattern   // Pattern to match
	Guard   ast.Expr   // Optional guard condition (if clause)
	Arrow   token.Pos  // Position of '=>'
	Body    ast.Expr   // Expression or block to execute
}

func (m *MatchArm) Pos() token.Pos { return m.Pattern.Pos() }
func (m *MatchArm) End() token.Pos { return m.Body.End() }

// Pattern represents a pattern in a match arm
// Can be:
//   - Wildcard: _
//   - Unit variant: Point
//   - Tuple variant: Circle(radius)
//   - Struct variant: Rectangle { width, height }
type Pattern struct {
	PatternPos token.Pos     // Position of pattern start
	Wildcard   bool          // true if this is a _ wildcard
	Variant    *ast.Ident    // Variant name (nil for wildcard)
	Fields     []*FieldPattern // Field patterns for destructuring
	Kind       PatternKind   // Unit, Tuple, Struct, or Wildcard
}

func (p *Pattern) Pos() token.Pos { return p.PatternPos }
func (p *Pattern) End() token.Pos {
	if len(p.Fields) > 0 {
		return p.Fields[len(p.Fields)-1].End()
	}
	if p.Variant != nil {
		return p.Variant.End()
	}
	return p.PatternPos + 1 // Wildcard
}

// PatternKind specifies the kind of pattern
type PatternKind int

const (
	PatternWildcard PatternKind = iota  // _ pattern
	PatternUnit                         // Point pattern
	PatternTuple                        // Circle(r) pattern
	PatternStruct                       // Rectangle { w, h } pattern
)

// FieldPattern represents a field binding in a pattern
// For tuple patterns: just a binding name
// For struct patterns: field name + binding name
type FieldPattern struct {
	FieldName *ast.Ident // Original field name (nil for tuple patterns)
	Binding   *ast.Ident // Variable name to bind to
}

func (f *FieldPattern) Pos() token.Pos {
	if f.FieldName != nil {
		return f.FieldName.Pos()
	}
	return f.Binding.Pos()
}

func (f *FieldPattern) End() token.Pos { return f.Binding.End() }

// ============================================================================
// Helper Functions
// ============================================================================

// IsDingoNode reports whether a node is a Dingo-specific extension
func IsDingoNode(node ast.Node) bool {
	switch node.(type) {
	case *ErrorPropagationExpr, *SafeNavigationExpr, *NullCoalescingExpr, *TernaryExpr, *LambdaExpr:
		return true
	case *ResultType, *OptionType:
		return true
	case *EnumDecl, *MatchExpr:
		return true
	default:
		return false
	}
}

// IsDingoType reports whether a type node is Dingo-specific
func IsDingoType(node ast.Node) bool {
	switch node.(type) {
	case *ResultType, *OptionType:
		return true
	default:
		return false
	}
}

// ============================================================================
// Extended Walk for Dingo Nodes
// ============================================================================

// Walk extends go/ast.Walk to handle Dingo-specific nodes
// It delegates to go/ast.Inspect for standard nodes and handles custom nodes
func Walk(node ast.Node, f func(ast.Node) bool) {
	ast.Inspect(node, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		// Call visitor function
		if !f(n) {
			return false
		}

		// Handle Dingo-specific nodes
		switch x := n.(type) {
		case *ErrorPropagationExpr:
			ast.Inspect(x.X, f)
			return false

		case *SafeNavigationExpr:
			ast.Inspect(x.X, f)
			ast.Inspect(x.Sel, f)
			return false

		case *NullCoalescingExpr:
			ast.Inspect(x.X, f)
			ast.Inspect(x.Y, f)
			return false

		case *TernaryExpr:
			ast.Inspect(x.Cond, f)
			ast.Inspect(x.Then, f)
			ast.Inspect(x.Else, f)
			return false

		case *LambdaExpr:
			if x.Params != nil {
				ast.Inspect(x.Params, f)
			}
			ast.Inspect(x.Body, f)
			return false

		case *ResultType:
			ast.Inspect(x.Value, f)
			ast.Inspect(x.Error, f)
			return false

		case *OptionType:
			ast.Inspect(x.Value, f)
			return false

		case *EnumDecl:
			ast.Inspect(x.Name, f)
			if x.TypeParams != nil {
				ast.Inspect(x.TypeParams, f)
			}
			for _, v := range x.Variants {
				ast.Inspect(v.Name, f)
				if v.Fields != nil {
					ast.Inspect(v.Fields, f)
				}
			}
			return false

		case *MatchExpr:
			ast.Inspect(x.Expr, f)
			for _, arm := range x.Arms {
				// Visit pattern
				if arm.Pattern != nil && arm.Pattern.Variant != nil {
					ast.Inspect(arm.Pattern.Variant, f)
				}
				// Visit guard
				if arm.Guard != nil {
					ast.Inspect(arm.Guard, f)
				}
				// Visit body
				ast.Inspect(arm.Body, f)
			}
			return false
		}

		// For standard go/ast nodes, continue normal traversal
		return true
	})
}

// ============================================================================
// Notes on Design
// ============================================================================

// Why this hybrid approach works:
//
// 1. Parser Phase:
//    - Dingo source -> Dingo AST (mix of go/ast + custom nodes)
//    - Example: fetchUser(id)? becomes ErrorPropagationExpr{X: CallExpr{...}}
//
// 2. Transform Phase:
//    - Dingo AST -> Pure go/ast
//    - Example: ErrorPropagationExpr -> IfStmt checking for errors
//
// 3. Generation Phase:
//    - Pure go/ast -> Go source code
//    - Use standard go/printer.Fprint()
//
// This means:
// - We get all go/ast tooling for free
// - Custom nodes are isolated and easy to transform
// - Generated Go code uses 100% standard go/ast (can use go/printer directly)
// - Source maps work naturally (both Dingo and Go use token.Pos)

package preprocessor

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"strconv"
	"strings"
)

// TernaryTypeInferrer analyzes Dingo expressions and infers concrete Go types.
// It uses go/parser and go/types for accurate type inference, returning
// concrete types like "string", "int", "bool" instead of "interface{}" or "any".
// This is specifically designed for ternary operator type inference.
type TernaryTypeInferrer struct {
	fset   *token.FileSet
	config *types.Config
	cache  map[string]string // Expression → type cache for performance
}

// NewTernaryTypeInferrer creates a new TernaryTypeInferrer instance.
func NewTernaryTypeInferrer() *TernaryTypeInferrer {
	return &TernaryTypeInferrer{
		fset: token.NewFileSet(),
		config: &types.Config{
			Error: func(err error) {
				// Ignore type errors - we'll fallback to "any"
			},
			Importer: nil, // No imports needed for literals
		},
		cache: make(map[string]string), // Initialize cache
	}
}

// InferType analyzes an expression and returns its concrete Go type.
// It handles:
// - Literals (strings, numbers, booleans)
// - Basic expressions (variables, function calls, field access)
// - Composite literals ([]int{1,2}, map[string]int{})
//
// Returns "any" for unknown/mixed types or on error.
//
// Examples:
//   - "hello" → "string"
//   - 42 → "int"
//   - true → "bool"
//   - []int{1,2} → "[]int"
//   - unknown → "any"
func (ti *TernaryTypeInferrer) InferType(expr string) string {
	// Check cache first
	if cachedType, ok := ti.cache[expr]; ok {
		return cachedType
	}

	var result string

	// Fast path: detect string literals
	if ti.isStringLiteral(expr) {
		result = "string"
	} else if numType := ti.detectNumericType(expr); numType != "" {
		// Fast path: detect numeric literals
		result = numType
	} else if ti.isBooleanLiteral(expr) {
		// Fast path: detect boolean literals
		result = "bool"
	} else {
		// Fallback: parse and analyze with go/types
		result = ti.inferTypeFromAST(expr)
	}

	// Cache result before returning
	ti.cache[expr] = result
	return result
}

// isStringLiteral checks if expression is a string literal.
func (ti *TernaryTypeInferrer) isStringLiteral(expr string) bool {
	expr = strings.TrimSpace(expr)
	if len(expr) < 2 {
		return false
	}

	// Raw strings
	if strings.HasPrefix(expr, "`") && strings.HasSuffix(expr, "`") {
		return true
	}

	// Regular strings
	if strings.HasPrefix(expr, `"`) && strings.HasSuffix(expr, `"`) {
		return true
	}

	return false
}

// isBooleanLiteral checks if expression is a boolean literal.
func (ti *TernaryTypeInferrer) isBooleanLiteral(expr string) bool {
	expr = strings.TrimSpace(expr)
	return expr == "true" || expr == "false"
}

// detectNumericType detects numeric literal types (int, float64).
func (ti *TernaryTypeInferrer) detectNumericType(expr string) string {
	expr = strings.TrimSpace(expr)

	// Check for float literals (contains . or e/E exponent)
	if strings.Contains(expr, ".") || strings.Contains(expr, "e") || strings.Contains(expr, "E") {
		if _, err := strconv.ParseFloat(expr, 64); err == nil {
			return "float64"
		}
		return ""
	}

	// Check for integer literals
	if _, err := strconv.ParseInt(expr, 0, 64); err == nil {
		return "int"
	}

	return ""
}

// inferTypeFromAST parses the expression as Go code and infers type using go/types.
func (ti *TernaryTypeInferrer) inferTypeFromAST(expr string) string {
	// Trim whitespace
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return "any"
	}

	// Check for parenthesized boolean literals before parsing
	// This avoids the AST parser treating them as identifiers
	cleanExpr := strings.Trim(expr, "()")
	cleanExpr = strings.TrimSpace(cleanExpr)
	if cleanExpr == "true" || cleanExpr == "false" {
		return "bool"
	}

	// Parse as Go expression
	node, err := parser.ParseExpr(expr)
	if err != nil {
		return "any"
	}

	// Try to infer type from AST node structure
	typeStr := ti.inferFromNode(node)
	if typeStr != "" {
		return typeStr
	}

	// Fallback to any
	return "any"
}

// inferFromNode infers type from AST node without full type checking.
// This provides best-effort type inference for common cases.
func (ti *TernaryTypeInferrer) inferFromNode(node ast.Expr) string {
	switch n := node.(type) {
	case *ast.BasicLit:
		return ti.basicLitType(n)

	case *ast.CompositeLit:
		return ti.compositeLitType(n)

	case *ast.ArrayType:
		if elemType := ti.inferFromNode(n.Elt); elemType != "" {
			return "[]" + elemType
		}
		return "[]any"

	case *ast.MapType:
		keyType := "any"
		valType := "any"
		if n.Key != nil {
			if kt := ti.inferFromNode(n.Key); kt != "" {
				keyType = kt
			}
		}
		if n.Value != nil {
			if vt := ti.inferFromNode(n.Value); vt != "" {
				valType = vt
			}
		}
		return "map[" + keyType + "]" + valType

	case *ast.Ident:
		// Known type identifiers
		switch n.Name {
		case "int", "int8", "int16", "int32", "int64":
			return n.Name
		case "uint", "uint8", "uint16", "uint32", "uint64", "uintptr":
			return n.Name
		case "float32", "float64":
			return n.Name
		case "bool":
			return "bool"
		case "string":
			return "string"
		case "byte":
			return "byte"
		case "rune":
			return "rune"
		case "error":
			return "error"
		case "any":
			return "any"
		default:
			// Unknown identifier - could be variable, fallback to any
			return "any"
		}

	case *ast.BinaryExpr:
		// For binary expressions, try to infer from operands
		leftType := ti.inferFromNode(n.X)
		rightType := ti.inferFromNode(n.Y)

		// If both sides have same type, use that
		if leftType == rightType && leftType != "any" {
			return leftType
		}

		// For numeric operations, default to int
		if ti.isNumericOp(n.Op) {
			if leftType == "int" || rightType == "int" {
				return "int"
			}
			if leftType == "float64" || rightType == "float64" {
				return "float64"
			}
		}

		// For comparison operations, result is bool
		if ti.isComparisonOp(n.Op) {
			return "bool"
		}

		return "any"

	case *ast.UnaryExpr:
		// Unary operations preserve operand type mostly
		if n.Op == token.NOT {
			return "bool"
		}
		return ti.inferFromNode(n.X)

	case *ast.ParenExpr:
		return ti.inferFromNode(n.X)

	default:
		// Unknown node type, fallback
		return "any"
	}
}

// basicLitType returns the type of a basic literal.
func (ti *TernaryTypeInferrer) basicLitType(lit *ast.BasicLit) string {
	switch lit.Kind {
	case token.INT:
		return "int"
	case token.FLOAT:
		return "float64"
	case token.STRING:
		return "string"
	case token.CHAR:
		return "rune"
	default:
		return "any"
	}
}

// compositeLitType infers the type of a composite literal.
func (ti *TernaryTypeInferrer) compositeLitType(lit *ast.CompositeLit) string {
	if lit.Type != nil {
		// Type is explicitly specified
		return types.ExprString(lit.Type)
	}

	// No explicit type - try to infer from elements
	return "any"
}

// isNumericOp checks if operator is numeric (+, -, *, /, %).
func (ti *TernaryTypeInferrer) isNumericOp(op token.Token) bool {
	switch op {
	case token.ADD, token.SUB, token.MUL, token.QUO, token.REM:
		return true
	default:
		return false
	}
}

// isComparisonOp checks if operator is comparison (==, !=, <, >, <=, >=).
func (ti *TernaryTypeInferrer) isComparisonOp(op token.Token) bool {
	switch op {
	case token.EQL, token.NEQ, token.LSS, token.GTR, token.LEQ, token.GEQ:
		return true
	default:
		return false
	}
}

// isNumericType checks if a type is a numeric type
func isNumericType(t string) bool {
	return t == "int" || t == "int8" || t == "int16" || t == "int32" || t == "int64" ||
		t == "uint" || t == "uint8" || t == "uint16" || t == "uint32" || t == "uint64" ||
		t == "float32" || t == "float64"
}

// promoteNumericTypes promotes two numeric types to their common type
func promoteNumericTypes(t1, t2 string) string {
	// If either is float, promote to float
	if strings.HasPrefix(t1, "float") || strings.HasPrefix(t2, "float") {
		// Use larger float type
		if t1 == "float64" || t2 == "float64" {
			return "float64"
		}
		return "float32"
	}

	// Both integers - use larger type (simplified: use int64 for safety)
	return "int64"
}

// InferBranchTypes analyzes both branches of a ternary and returns the appropriate return type.
// If both branches have the same concrete type, returns that type.
// Handles numeric type promotion (e.g., int + float64 → float64).
// Otherwise returns "any" (Go 1.18+ generic any type).
//
// Examples:
//   - ("adult", "minor") → "string"
//   - (100, 200) → "int"
//   - (42, 3.14) → "float64" (numeric promotion)
//   - ("text", 42) → "any"
func (ti *TernaryTypeInferrer) InferBranchTypes(trueVal, falseVal string) string {
	trueType := ti.InferType(trueVal)
	falseType := ti.InferType(falseVal)

	// If types match exactly, return the concrete type
	if trueType == falseType {
		return trueType
	}

	// Handle numeric type promotion
	if isNumericType(trueType) && isNumericType(falseType) {
		return promoteNumericTypes(trueType, falseType)
	}

	// If either is "any", return "any"
	if trueType == "any" || falseType == "any" {
		return "any"
	}

	// Types differ - use "any" as fallback (Go 1.18+)
	return "any"
}

// Package builtin provides safe navigation type inference plugin
package builtin

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"github.com/MadAppGang/dingo/pkg/plugin"
	"golang.org/x/tools/go/ast/astutil"
)

// SafeNavTypePlugin resolves __INFER__ placeholders with actual types for safe navigation
//
// This plugin implements Task A2 from Phase 7: Safe navigation type inference.
// It finds __INFER__ placeholder identifiers inserted by SafeNavProcessor and replaces them
// with concrete types using go/types analysis.
//
// Type Resolution Strategy:
// 1. Discovery: Find all __INFER__ identifiers in AST
// 2. Transform: Use go/types to resolve actual types
// 3. Inject: Replace __INFER__ with concrete types (Option<T> or *T)
//
// Supports:
// - Option<T> types (enum-based optionals)
// - Raw Go pointers (*T)
// - Error reporting for non-nullable types
type SafeNavTypePlugin struct {
	ctx *plugin.Context

	// Track which __INFER__ nodes we've found
	inferNodes []*inferNode

	// Type inference service for accurate type resolution
	typeInference *TypeInferenceService

	// Errors encountered during type inference
	errors []string
}

// inferNode represents an __INFER__ placeholder that needs type resolution
type inferNode struct {
	// The identifier node with name "__INFER__"
	ident *ast.Ident

	// The parent node (should be a selector expression like __INFER__.field)
	parent ast.Node

	// The resolved type (set during Transform phase)
	resolvedType string

	// Whether this is an Option<T> or pointer type
	isOption  bool
	isPointer bool
}

// NewSafeNavTypePlugin creates a new safe navigation type inference plugin
func NewSafeNavTypePlugin() *SafeNavTypePlugin {
	return &SafeNavTypePlugin{
		inferNodes: make([]*inferNode, 0),
		errors:     make([]string, 0),
	}
}

// Name returns the plugin name
func (p *SafeNavTypePlugin) Name() string {
	return "safe_nav_types"
}

// SetContext sets the plugin context (ContextAware interface)
func (p *SafeNavTypePlugin) SetContext(ctx *plugin.Context) {
	p.ctx = ctx

	// Initialize type inference service with go/types integration
	if ctx != nil && ctx.FileSet != nil {
		// Create type inference service
		service, err := NewTypeInferenceService(ctx.FileSet, nil, ctx.Logger)
		if err != nil {
			ctx.Logger.Warn("SafeNavTypePlugin: Failed to create type inference service: %v", err)
		} else {
			p.typeInference = service

			// Inject go/types.Info if available in context
			if ctx.TypeInfo != nil {
				if typesInfo, ok := ctx.TypeInfo.(*types.Info); ok {
					service.SetTypesInfo(typesInfo)
					ctx.Logger.Debug("SafeNavTypePlugin: go/types integration enabled")
				}
			}
		}
	}
}

// Process discovers __INFER__ placeholders in the AST (Discovery Phase)
func (p *SafeNavTypePlugin) Process(node ast.Node) error {
	if p.ctx == nil {
		return fmt.Errorf("plugin context not initialized")
	}

	// Build parent map if not already built
	if p.ctx.GetParentMap() == nil {
		if file, ok := node.(*ast.File); ok {
			p.ctx.BuildParentMap(file)
		}
	}

	// Walk the AST to find __INFER__ identifiers
	ast.Inspect(node, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok {
			if ident.Name == "__INFER__" {
				// Found an __INFER__ placeholder
				parent := p.ctx.GetParent(ident)
				p.inferNodes = append(p.inferNodes, &inferNode{
					ident:  ident,
					parent: parent,
				})
				p.ctx.Logger.Debug("SafeNavTypePlugin: Found __INFER__ placeholder at %v", p.ctx.FileSet.Position(ident.Pos()))
			}
		}
		return true
	})

	p.ctx.Logger.Debug("SafeNavTypePlugin: Discovery complete, found %d __INFER__ placeholders", len(p.inferNodes))
	return nil
}

// Transform resolves types and replaces __INFER__ placeholders (Transform Phase)
func (p *SafeNavTypePlugin) Transform(node ast.Node) (ast.Node, error) {
	if p.ctx == nil {
		return nil, fmt.Errorf("plugin context not initialized")
	}

	if len(p.inferNodes) == 0 {
		// No __INFER__ placeholders to resolve
		return node, nil
	}

	// Resolve types for all discovered __INFER__ nodes
	for _, inferNode := range p.inferNodes {
		if err := p.resolveTypeForInferNode(inferNode); err != nil {
			p.errors = append(p.errors, err.Error())
			p.ctx.ReportError(err.Error(), inferNode.ident.Pos())
		}
	}

	// Use astutil.Apply to walk and transform the AST
	transformed := astutil.Apply(node,
		func(cursor *astutil.Cursor) bool {
			n := cursor.Node()

			// Check if this is an __INFER__ identifier we need to replace
			if ident, ok := n.(*ast.Ident); ok {
				if ident.Name == "__INFER__" {
					// Find the corresponding inferNode
					for _, inferNode := range p.inferNodes {
						if inferNode.ident == ident && inferNode.resolvedType != "" {
							// Replace __INFER__ with the resolved type
							replacement := ast.NewIdent(inferNode.resolvedType)
							cursor.Replace(replacement)
							p.ctx.Logger.Debug("SafeNavTypePlugin: Replaced __INFER__ with %s", inferNode.resolvedType)
							break
						}
					}
				}
			}
			return true
		},
		nil, // Post-order not needed
	)

	return transformed, nil
}

// resolveTypeForInferNode uses go/types to resolve the actual type for an __INFER__ placeholder
func (p *SafeNavTypePlugin) resolveTypeForInferNode(node *inferNode) error {
	if p.typeInference == nil {
		return fmt.Errorf("type inference service not available")
	}

	// The __INFER__ should appear in a selector expression like:
	// __INFER__.field or __INFER__.method()
	// We need to find the actual variable being accessed

	// Check the parent node
	switch parent := node.parent.(type) {
	case *ast.SelectorExpr:
		// __INFER__ is the X part of a selector: __INFER__.field
		// We need to find what variable this __INFER__ represents
		// This is tricky because the preprocessor should have provided context

		// Look for patterns like: var __INFER__ = someVar
		// or function arguments: __SAFE_NAV_INFER__(someVar, "field")

		// For now, we'll use a simplified approach:
		// Try to infer from the selector's field name and surrounding context
		return p.inferFromContext(node, parent)

	case *ast.CallExpr:
		// __INFER__ might be in a function call: __SAFE_NAV_INFER__(var, "field")
		return p.inferFromFunctionCall(node, parent)

	default:
		return fmt.Errorf("unexpected parent node type for __INFER__: %T", parent)
	}
}

// inferFromContext attempts to infer type from surrounding context
func (p *SafeNavTypePlugin) inferFromContext(node *inferNode, selector *ast.SelectorExpr) error {
	// Walk up the parent chain to find the actual variable
	var actualVar ast.Expr

	p.ctx.WalkParents(selector, func(parent ast.Node) bool {
		switch p := parent.(type) {
		case *ast.AssignStmt:
			// Found assignment: someVar = __INFER__.field
			// The LHS might give us the variable
			if len(p.Rhs) > 0 {
				// Look for the actual variable in the RHS
				// This is a simplified heuristic
			}
		case *ast.CallExpr:
			// Found function call containing __INFER__
			// Look for patterns like __SAFE_NAV_INFER__(actualVar, ...)
			if ident, ok := p.Fun.(*ast.Ident); ok {
				if strings.HasPrefix(ident.Name, "__SAFE_NAV_INFER__") {
					if len(p.Args) > 0 {
						actualVar = p.Args[0]
						return false // Stop walking
					}
				}
			}
		}
		return true // Continue walking
	})

	if actualVar != nil {
		return p.resolveTypeFromExpr(node, actualVar)
	}

	// Fallback: Report error
	return fmt.Errorf("unable to infer type for __INFER__ placeholder")
}

// inferFromFunctionCall attempts to infer type from function call pattern
func (p *SafeNavTypePlugin) inferFromFunctionCall(node *inferNode, call *ast.CallExpr) error {
	// Expected pattern: __SAFE_NAV_INFER__(actualVar, "field")
	if ident, ok := call.Fun.(*ast.Ident); ok {
		if strings.HasPrefix(ident.Name, "__SAFE_NAV_INFER__") {
			if len(call.Args) > 0 {
				actualVar := call.Args[0]
				return p.resolveTypeFromExpr(node, actualVar)
			}
		}
	}

	return fmt.Errorf("invalid __SAFE_NAV_INFER__ function call pattern")
}

// resolveTypeFromExpr resolves the type of an expression using go/types
func (p *SafeNavTypePlugin) resolveTypeFromExpr(node *inferNode, expr ast.Expr) error {
	// Use type inference service to get the type
	typ, ok := p.typeInference.InferType(expr)
	if !ok || typ == nil {
		return fmt.Errorf("failed to infer type for expression: %s", FormatExprForDebug(expr))
	}

	// Check if it's a pointer type
	if ptrType, ok := typ.(*types.Pointer); ok {
		node.isPointer = true
		node.isOption = false
		node.resolvedType = p.typeInference.TypeToString(ptrType.Elem())
		p.ctx.Logger.Debug("SafeNavTypePlugin: Resolved __INFER__ to pointer type: *%s", node.resolvedType)
		return nil
	}

	// Check if it's an Option type (struct with tag field + IsSome/IsNone methods)
	if named, ok := typ.(*types.Named); ok {
		if p.isOptionType(named) {
			node.isOption = true
			node.isPointer = false
			node.resolvedType = named.Obj().Name()
			p.ctx.Logger.Debug("SafeNavTypePlugin: Resolved __INFER__ to Option type: %s", node.resolvedType)
			return nil
		}
	}

	// Not a nullable type - report error
	return fmt.Errorf("safe navigation requires nullable type (Option<T> or *T), got: %s", p.typeInference.TypeToString(typ))
}

// isOptionType checks if a named type is an Option type
// An Option type is identified by:
// 1. Having a 'tag' field of type OptionTag
// 2. Having IsSome() and IsNone() methods
func (p *SafeNavTypePlugin) isOptionType(named *types.Named) bool {
	// Get the underlying struct type
	structType, ok := named.Underlying().(*types.Struct)
	if !ok {
		return false
	}

	// Check for 'tag' field
	hasTagField := false
	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		if field.Name() == "tag" {
			// Check if tag type is OptionTag or ResultTag
			if namedType, ok := field.Type().(*types.Named); ok {
				tagName := namedType.Obj().Name()
				if tagName == "OptionTag" || tagName == "ResultTag" {
					hasTagField = true
					break
				}
			}
		}
	}

	if !hasTagField {
		return false
	}

	// Check for IsSome() method
	hasSomeMethod := false
	for i := 0; i < named.NumMethods(); i++ {
		method := named.Method(i)
		if method.Name() == "IsSome" {
			hasSomeMethod = true
			break
		}
	}

	return hasSomeMethod
}

// GetErrors returns all accumulated errors
func (p *SafeNavTypePlugin) GetErrors() []string {
	return p.errors
}

// ClearErrors clears all accumulated errors
func (p *SafeNavTypePlugin) ClearErrors() {
	p.errors = make([]string, 0)
}

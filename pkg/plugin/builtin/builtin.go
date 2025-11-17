// Package builtin provides built-in Dingo transformation plugins
package builtin

import (
	"fmt"

	"github.com/MadAppGang/dingo/pkg/plugin"
)

// NewDefaultRegistry creates a registry with all built-in plugins registered
// This is the standard set of plugins for Dingo compilation
//
// Plugin Execution Order:
//
// The order of plugin registration matters due to dependencies:
//
// 1. Type Plugins (ResultTypePlugin, OptionTypePlugin)
//    - Define core types used by other plugins
//    - Must run first to ensure types are available
//
// 2. SumTypesPlugin
//    - Transforms enum declarations
//    - MUST run before ErrorPropagationPlugin to avoid crashes
//    - ErrorPropagation does type checking which fails on empty GenDecl placeholders
//    - SumTypes transforms and removes those placeholders
//
// 3. ErrorPropagationPlugin
//    - Uses Result types from ResultTypePlugin
//    - Requires clean AST from SumTypesPlugin
//
// 4. FunctionalUtilitiesPlugin
//    - Currently works with plain slices
//    - Future: Will integrate with Result/Option for mapResult/filterSome
//    - When implementing mapResult/filterSome, ensure Result/Option plugins run first
//
// 5. Other utility plugins (SafeNavigation, NullCoalescing, Ternary, Lambda)
//    - Independent of other plugins
//    - Can run in any order relative to each other
func NewDefaultRegistry() (*plugin.Registry, error) {
	registry := plugin.NewRegistry()

	// Register all built-in plugins in dependency order
	plugins := []plugin.Plugin{
		NewResultTypePlugin(),           // 1. Core type: Result<T, E>
		NewOptionTypePlugin(),           // 1. Core type: Option<T>
		NewSumTypesPlugin(),             // 2. Sum types - MUST run before error propagation!
		NewErrorPropagationPlugin(),     // 3. Error propagation (depends on Result, SumTypes cleanup)
		NewFunctionalUtilitiesPlugin(),  // 4. Functional utilities (future: depends on Result/Option)
		NewSafeNavigationPlugin(),       // 5. Safe navigation operator (?.)
		NewNullCoalescingPlugin(),       // 5. Null coalescing operator (??)
		NewTernaryPlugin(),              // 5. Ternary operator (? :)
		NewLambdaPlugin(),               // 5. Lambda functions (|x| expr and (x) => expr)
		// Add more plugins here as they are implemented
	}

	for _, p := range plugins {
		if err := registry.Register(p); err != nil {
			return nil, fmt.Errorf("failed to register plugin %s: %w", p.Name(), err)
		}
	}

	// Sort plugins by dependencies
	if err := registry.SortByDependencies(); err != nil {
		return nil, fmt.Errorf("failed to sort plugins: %w", err)
	}

	// Enable all plugins by default
	for _, p := range plugins {
		p.SetEnabled(true)
	}

	return registry, nil
}

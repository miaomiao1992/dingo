# User Clarifications

## Scope Constraints
- **DO NOT fix tests** - Another agent is handling test fixes
- **DO NOT modify Dingo engine** - Could conflict with other agent's work
- **Focus on features that already work** - No implementation changes

## Strategic Decisions

### v1.0 Scope
Skip Phase V.1 test fixes (handled by another agent). Focus on infrastructure and tooling around working features.

### Package Management Strategy
**Hybrid approach:**
- Libraries → Publish `.go` files (transpile-on-publish)
- Applications → Use `.dingo` files directly

### Constraints
- No changes to transpiler, parser, or AST transformations
- No changes to test suite or golden files
- Focus on: Documentation, package management workflow, CI enhancements, tooling

## Phase V Tasks to Focus On (Non-Engine)

1. **Package Management Documentation**
   - Document hybrid workflow for libraries vs applications
   - Create guides for publishing Dingo packages
   - Example projects showing both patterns

2. **CI/CD Enhancements** (No test changes)
   - Performance tracking dashboard
   - Golden test diff visualization (just visualization, not fixes)
   - Documentation generation automation

3. **Source Map Validation** (Read-only validation)
   - Validation suite that checks existing source maps
   - No changes to source map generation
   - Just verification and reporting

4. **Developer Experience**
   - VS Code extension enhancements (if any needed)
   - Documentation improvements
   - Getting started guides

5. **Workspace Builds** (If time permits)
   - Multi-package build support
   - Dependency resolution
   - Build caching

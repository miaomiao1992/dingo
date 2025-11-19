# Investigation Request: File Organization Strategy

## Problem Statement

Currently, Dingo generates multiple files per source file:
- `.dingo` - Source file
- `.go` - Transpiled output
- `.go.map` - Source map

This creates file clutter with 3 files duplicating each other in the same location.

## User's Proposed Solution

Create three separate folder structures:

1. **Source folder** - Mix of `.dingo` and `.go` files (original source code)
2. **Shadow folder** - Compiled version with only `.go` files (different package, pre-compiled)
3. **Map folder** - Source map files only (maybe in an insert folder)

The Dingo transpiler would:
- Read from source folder
- Transpile `.dingo` files
- Output transpiled `.go` to shadow folder
- Output `.go.map` files to map folder

## Investigation Goals

1. **Analyze current project structure** - How are files organized now?
2. **Research best practices** - How do similar tools handle this?
   - TypeScript (`.ts` → `.js` + `.js.map`)
   - templ (`.templ` → `_templ.go`)
   - Borgo (`.brg` → `.go`)
3. **Evaluate user's proposal** - Is shadow folder approach optimal?
4. **Consider alternatives** - Other organization strategies?
5. **LSP implications** - How does file organization affect gopls integration?
6. **Tooling changes needed** - What needs to change in compiler/LSP?

## Constraints

- Must maintain Go ecosystem compatibility
- LSP must work seamlessly (gopls proxy)
- Source maps needed for IDE features
- Should feel natural to Go developers
- Minimal configuration burden

## Desired Output

Architectural proposal with:
- Recommended file organization strategy
- Rationale comparing alternatives
- Implementation plan for compiler changes
- LSP server changes needed
- Migration strategy for existing code

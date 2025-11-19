# Architectural Consultation: Dingo File Organization Strategy

## Current Problem
Dingo generates output files (.go, .go.map) in the same directory as input files (.dingo), causing:
1. File clutter: 3 files per source (source + output + map)
2. Test directory chaos: 242 files for 62 tests
3. Git complexity: Hard to ignore generated files without ignoring hand-written Go
4. Name collision risk: Can't have foo.dingo and foo.go in same directory

## Goal
Find the BEST file organization strategy for Dingo that balances:
- Developer Experience (natural for Go developers)
- Tool Integration (go build, go test, gopls)
- Scalability (large projects)
- Maintainability (.gitignore, CI/CD)

## Options Considered
A. In-Place Generation (current): .dingo + .go + .go.map in same directory
B. Shadow Folder: src/ with .dingo, gen/ with transpiled .go, maps/ with .go.map
C. Target Directory: *.dingo and *.go mixed, target/ for generated files
D. Suffix Pattern: foo.dingo → foo_gen.go in same directory
E. Hybrid: Your creative solution

## Questions to Answer
1. Should Dingo default to in-place (like templ) or separate directory (like TypeScript)?
2. How to handle mixed .dingo + hand-written .go projects?
3. What's the Go-idiomatic way to organize generated sources?
4. How to maintain gopls integration with separated directories?
5. What configuration should dingo.toml [build] section contain?

## Constraints
- Must preserve gopls integration
- Generated .go must be valid for Go tools
- Source maps needed for LSP position translation
- Go module structure must work correctly
- Keep simple: dingo build file.dingo should work without config

## Provide
1. Recommended strategy with concrete directory structure example
2. Configuration schema for dingo.toml
3. Trade-off analysis vs alternatives
4. Implementation considerations (CLI, LSP changes)
5. Migration path for existing projects

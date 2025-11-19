
# Architectural Consultation: Dingo File Organization Strategy

## Your Role

You are a software architecture consultant evaluating file organization strategies for Dingo, a meta-language that transpiles to Go.

## Context

Read the attached context document which contains:
1. Current implementation analysis (how files are organized now)
2. Similar tools research (TypeScript, Rust, Borgo, templ patterns)
3. Pain points with current approach
4. User's proposed "shadow folder" solution

## Your Task

Propose the **BEST** file organization strategy for Dingo. Consider:

### Evaluation Criteria

1. **Developer Experience**
   - How natural does it feel to Go developers?
   - Configuration complexity?
   - Learning curve?

2. **Tool Integration**
   - Go tooling compatibility (go build, go test, etc.)
   - LSP server complexity (gopls proxy)
   - IDE experience (file navigation, debugging)

3. **Scalability**
   - Large projects (1000+ files)
   - Monorepos with multiple packages
   - Build performance

4. **Maintainability**
   - .gitignore patterns
   - CI/CD integration
   - Cleanup and regeneration

### Your Deliverable

Provide a a **concrete architectural recommendation** with:

1. **Recommended Strategy** - Which approach? (in-place, shadow folder, target dir, hybrid?)
2. **File Layout Example** - Show concrete directory structure
3. **Rationale** - Why is this optimal? What trade-offs?
4. **Comparison** - How does it compare to user's proposal and alternatives?
5. **Implementation Notes** - What changes needed in compiler/LSP?
6. **Migration Path** - How to transition from current approach?

### Alternative Approaches to Consider

A. **In-Place Generation** (current)
   - `.dingo` + `.go` + `.go.map` in same directory
   - Simple but cluttered

B. **Shadow Folder** (user's proposal)
   - Source: `src/` with `.dingo` + `.go` files
   - Output: `gen/` with transpiled `.go` files
   - Maps: `maps/` or `.dingo/maps/`

C. **Target Directory** (TypeScript/Rust model)
   - Source: `*.dingo` and `*.go` mixed
   - Output: `target/` or `dist/` for all generated files

D. **Suffix Pattern** (templ model)
   - `foo.dingo` → `foo_gen.go` in same directory
   - Maps: `.dingo/` hidden folder

E. **Hybrid** - Your creative solution

### Be Specific

Don't just say "shadow folder is good" - specify:
- Exact directory names (`gen/`, `target/`, `.dingo/build/`?)
- Where do source maps go?
- How does gopls find files?
- What goes in .gitignore?
- How does `go build` work?

### Real-World Scenarios

Consider these use cases:
1. Simple project: `main.dingo` + 3 packages
2. Large project: 50 packages, 500 `.dingo` files
3. Mixed project: 30% Dingo, 70% plain Go
4. Library: Dingo code exposing Go API

Thank you for your architectural expertise!
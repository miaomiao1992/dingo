# Dingo Changelog

All notable changes to the Dingo compiler will be documented in this file.

## [Unreleased] - 2025-11-16

### Iteration 2 - Plugin System

**Added:**
- ✨ **Plugin System Architecture** - Complete modular plugin framework
  - `Plugin` interface for extensible features
  - `PluginRegistry` for plugin management and discovery
  - `Pipeline` for AST transformation with dependency resolution
  - Topological sort for correct plugin execution order
  - Circular dependency detection
  - Enable/disable plugin functionality
  - Logging infrastructure (Debug/Info/Warn/Error)
  - `BasePlugin` for easy plugin implementation

- 📦 **New Package: `pkg/plugin/`** - ~681 lines of production code
  - `plugin.go` - Core interfaces and registry (228 lines)
  - `pipeline.go` - Transformation pipeline (106 lines)
  - `logger.go` - Logging infrastructure (83 lines)
  - `base.go` - Base plugin implementation (47 lines)
  - `plugin_test.go` - Comprehensive tests (217 lines, 100% pass rate)

- 📄 **Documentation:**
  - `PLUGIN_SYSTEM_DESIGN.md` - Complete architecture documentation

**Changed:**
- 🔄 **Generator Integration** - Updated `pkg/generator/generator.go`
  - Added plugin pipeline support
  - New `NewWithPlugins()` constructor for custom plugins
  - Transform step in generation pipeline: Parse → Transform → Generate → Format
  - Backward compatible (default generator has no plugins)
  - Logger integration for debugging

- 🐕 **Emoji Update** - Changed mascot from dinosaur 🦕 to dog 🐕
  - Updated CLI header output
  - Updated version command output

**Technical Details:**
- Dependency resolution uses Kahn's algorithm (O(V + E) time complexity)
- Deterministic plugin ordering for consistent builds
- Zero overhead for disabled plugins
- Comprehensive test coverage (8 tests, all passing)

---

### Iteration 1 - Foundation

**Added:**
- ✨ **Basic Transpiler** - Complete Dingo → Go compilation pipeline
- ✨ **`dingo build`** - Transpile .dingo files to .go
- ✨ **`dingo run`** - Compile and execute in one step (like `go run`)
  - Supports passing arguments: `dingo run file.dingo -- arg1 arg2`
  - Passes through stdin/stdout/stderr
  - Preserves program exit codes
- ✨ **Beautiful CLI Output** - lipgloss-powered terminal UI
- ✨ **`dingo version`** - Version information

**Changed:**
- 🔥 **Removed arrow syntax for return types** (breaking, but no releases yet)
  - **Before:** `func max(a: int, b: int) -> int`
  - **After:** `func max(a: int, b: int) int`
  - **Rationale:** Cleaner, closer to Go, arrow adds no value

**Improved:**
- 📝 Better error messages for parse failures
- 🎨 Consistent beautiful output across all commands

## Design Philosophy

**Principle:** Keep syntax changes minimal. Only diverge from Go when there's clear value.

### What We Keep Different
- ✅ **Parameter types with `:`** - `func max(a: int, b: int)` is clearer than `func max(a int, b int)`
- ✅ **`let` keyword** - Explicit immutability by default

### What We Keep Same
- ✅ **Return types** - Just `int`, no arrow (same as Go)
- ✅ **Braces, semicolons, etc.** - Follow Go conventions

---

## [0.1.0-alpha] - 2025-11-16

### Initial Release

#### Core Features
- 🦕 **Dingo Compiler** - Full transpilation pipeline (Dingo → Go)
- 📦 **CLI Tool** with beautiful output (lipgloss-powered)
- ⚡ **Parser** - participle-based with full expression support
- 🎨 **Generator** - go/printer + go/format for clean output
- 🏗️ **Hybrid AST** - Reuses go/ast with custom Dingo nodes

#### Commands
- `dingo build` - Transpile .dingo files to .go
- `dingo run` - Compile and execute immediately
- `dingo version` - Show version information
- `dingo --help` - Full documentation

#### Syntax Support
- ✅ Package declarations
- ✅ Import statements
- ✅ Function declarations with `:` parameter syntax
- ✅ Variable declarations (`let`/`var`)
- ✅ Type annotations
- ✅ Expressions (binary, unary, calls)
- ✅ Operator precedence
- ✅ Comments

#### Developer Experience
- 🌈 Full color terminal output
- 📊 Performance metrics for each build step
- 🎯 Clear, actionable error messages
- ✨ Professional polish matching modern tools

#### Documentation
- 📚 Complete README with examples
- 🎨 CLI showcase with screenshots
- 📝 Syntax design rationale
- 🛠️ Implementation guides

#### Statistics
- **1,486 lines** of production code
- **5 packages** (ast, parser, generator, ui, main)
- **3 example programs** included
- **100% test pass rate**

---

## Future Roadmap

### Phase 2 (Week 2) - Plugin System
- [ ] Plugin architecture
- [ ] Error propagation (`?` operator)
- [ ] Source maps for debugging

### Phase 3 - Core Features
- [ ] `Result<T, E>` type
- [ ] `Option<T>` type
- [ ] Pattern matching
- [ ] Null coalescing (`??`)
- [ ] Ternary operator (`? :`)

### Phase 4 - Advanced Features
- [ ] Lambda functions (multiple syntax styles)
- [ ] Sum types (enums)
- [ ] Functional utilities (map, filter, reduce)
- [ ] Tree-sitter migration
- [ ] Language server (gopls proxy)

---

## Notes

**Breaking Changes:** Since we haven't released v1.0 yet, we're free to make breaking changes to improve the design. The arrow syntax removal is a perfect example - better to fix it now than carry technical debt forever.

**Versioning:** Following semantic versioning once we hit v1.0. Until then, expect API changes.

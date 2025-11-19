# Architecture Pivot: Participle → go/parser + Preprocessor

**Date**: 2025-11-17
**Decision**: Abandon Participle parser, switch to go/parser + Preprocessor
**Rationale**: Leverage Go's native tooling instead of reimplementing Go syntax

## Why This Pivot is Necessary

### Current State (Participle)
- ❌ ~20% Go syntax coverage
- ❌ Need 10,000+ lines to complete
- ❌ Missing: selectors, assignments, ~80+ Go constructs
- ❌ Must track Go evolution manually
- ❌ Blocks Result/Option types (can't parse method calls)

### New State (go/parser + Preprocessor)
- ✅ 100% Go syntax coverage (via go/parser)
- ✅ Only ~2,300 lines custom code
- ✅ Auto-tracks Go evolution
- ✅ All Go features work automatically
- ✅ Unblocks Result/Option and all future features

## Strategic Reasoning

### 1. Align with Go Ecosystem (Don't Reinvent)
**Principle**: Use the same tools Go uses

Go's official tools:
- `go/parser` - Parse Go source → AST
- `go/ast` - AST representation
- `go/printer` - AST → Go source
- `go/types` - Type checker
- `golang.org/x/tools/go/ast/astutil` - AST manipulation

**Our new architecture uses ALL of these.**

### 2. Follow Industry Precedent
Every successful meta-language uses preprocessing:
- **TypeScript**: TS syntax → JS syntax → JS parser
- **Rust macros**: Macro expansion → Rust parser
- **C preprocessor**: #define → C parser
- **templ**: Templates → Go → gopls
- **PREGO, gpp**: Go preprocessors

**This is the proven pattern.**

### 3. Focus Effort Where It Matters
**Waste**: Reimplementing Go's parser (10,000+ lines)
**Value**: Dingo-specific features (~2,300 lines)

With preprocessor approach:
- 100% effort on Dingo features
- 0% effort on Go parsing (go/parser handles it)

### 4. Enable Rapid Feature Development
Adding new Dingo features becomes trivial:

**Participle approach** (per feature):
1. Add parser grammar (100-200 lines)
2. Add AST conversion (100-200 lines)
3. Handle edge cases (50-100 lines)
4. Test parser (50-100 lines)
5. Add transformation (100-200 lines)
**Total: 400-800 lines per feature**

**Preprocessor approach** (per feature):
1. Add preprocessing rule (50-100 lines)
2. Add AST transformation (100-200 lines)
**Total: 150-300 lines per feature**

**2-3x faster feature development!**

## User Confirmation

User explicitly requested:
> "DO NOT reimplement Go features. Use the same tools Go native tools use. We need some tool which solved the majority of tasks for us and all we need is to extend it with dingo features"

**This architecture directly fulfills that requirement.**

## Next Steps

1. Design detailed architecture (golang-architect)
2. Create migration plan (preserve existing working features)
3. Implement in phases (incremental, not big-bang)
4. Validate with existing golden tests

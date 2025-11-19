# Multi-Model Consultation: Package-Wide Scanning for Unqualified Import Inference

## Context

Dingo transpiler needs to implement unqualified import inference:
- Transform `ReadFile(path)` → `os.ReadFile(path)` + add import "os"
- Must detect local user-defined functions to avoid false transforms
- Currently considering single-file scope vs package-wide scope

## Current Approach (Single File Scope)

**Pros:**
- Fast: ~10-50ms per file
- Simple: Fits file-oriented transpilation
- Rare false positives caught by Go compiler

**Cons:**
- Misses functions defined in other files of same package
- Cross-file false transforms possible (though rare)

## The Challenge: Package-Wide Scanning

User wants package-wide scanning from the start to avoid ANY false transforms.

**Key Questions:**
1. How to efficiently scan all files in a package during transpilation?
2. How to cache/share scan results across file transpilations?
3. What's the performance impact (acceptable threshold: <500ms per package)?
4. How to handle incremental builds (one file changed)?
5. Should we use go/packages, manual file discovery, or something else?

## Architecture Constraints

- Dingo transpiles file-by-file (currently stateless between files)
- Uses preprocessor stage (text transformations before go/parser)
- Must remain fast for watch mode / incremental builds
- Should integrate cleanly with existing ImportTracker

## Your Task

**Design the best approach for package-wide local function detection that:**
1. Scans all `.dingo` files in the package
2. Builds a shared exclusion list of local functions
3. Remains fast enough for watch mode (<500ms total per package)
4. Handles incremental builds efficiently
5. Integrates with preprocessor pipeline

**Provide:**
- Specific architecture (components, data flow)
- Caching strategy for performance
- Incremental build handling
- Code structure (which packages/files)
- Performance estimation with rationale
- Trade-offs and edge cases

**Format your response with:**
- ## Proposed Architecture
- ## Caching Strategy
- ## Incremental Build Handling
- ## Performance Analysis
- ## Implementation Plan
- ## Trade-offs & Edge Cases

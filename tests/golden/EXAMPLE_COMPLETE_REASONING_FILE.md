# Example: Complete Reasoning File

This document shows what a fully-developed reasoning file looks like for reference.

## File: error_prop_01_simple.reasoning.md

This is one of the most comprehensive reasoning files, serving as a template for complex features.

### Structure

```markdown
---
# YAML Frontmatter (structured metadata)
title: "Simple error propagation with ? operator in statement"
category: "Error Handling"
subcategory: "Error Propagation"
test_id: "error_prop_01_simple"
order: 1
complexity: "basic"
feature: "error-propagation"
phase: "Phase 2.4"
status: "implemented"
description: "Demonstrates basic error propagation..."
summary: "Basic ? operator in let statement"
code_reduction: 58
lines_dingo: 7
lines_go: 12
go_proposal: "71203"
go_proposal_link: "https://github.com/golang/go/issues/71203"
feature_file: "error-propagation.md"
related_tests:
  - "error_prop_02_multiple"
  - "error_prop_03_expression"
tags:
  - "error-handling"
  - "operator"
keywords:
  - "? operator"
  - "error propagation"
---

# Test Reasoning: error_prop_01_simple

## Test File
- **Source**: test file location
- **Feature**: feature name
- **Phase**: development phase
- **Status**: implementation status

## What This Test Validates
Brief description of what the test proves

### Dingo Code (X lines)
```dingo
// Source code
```

### Generated Go Code (Y lines)
```go
// Generated code
```

## Community Context
### Go Proposal Reference
Links and discussion context

### Why This Matters
Explanation of real-world impact

## Design Decisions
### 1. Design Decision Name
**Rationale**:
- Bullet points

**Alternative Considered**:
- Options that were rejected and why

### 2. Another Design Decision
...

## Feature File Reference
Link to feature specification

## Comparison with Other Languages
### Rust
Code example and comparison

### Swift
Code example and comparison

## Testing Strategy
### What This Test Proves
Validation points

### Edge Cases Covered
List of edge cases

### Edge Cases NOT Covered
What other tests handle

## Success Metrics
**Code Reduction**: X% reduction
**Type Safety**: What safety is gained
**Developer Experience**: UX improvements

## Known Limitations & Future Work
### Current Limitations
List of limitations

### Future Enhancements
Planned improvements

## External References
### Go Proposals
Links

### Community Tools
Links

### Language Documentation
Links

---
**Last Updated**: Date
**Test Status**: Status
**Dingo Version**: Version
```

## Key Sections Explained

### 1. YAML Frontmatter
- **Purpose**: Machine-readable metadata for Astro Content Collections
- **Usage**: Powers navigation, filtering, search
- **Required**: Yes, for all files

### 2. Test File Info
- **Purpose**: Quick reference for file location and context
- **Usage**: Helps developers find the actual test
- **Required**: Yes

### 3. Code Examples
- **Purpose**: Show input and output side-by-side
- **Usage**: Demonstrate transpilation
- **Required**: Yes, both Dingo and Go

### 4. Community Context
- **Purpose**: Link to Go community discussions
- **Usage**: Provides rationale for feature
- **Required**: When Go proposal exists

### 5. Design Decisions
- **Purpose**: Explain implementation choices
- **Usage**: Documents "why" not just "what"
- **Required**: For non-trivial features

### 6. Language Comparisons
- **Purpose**: Show how other languages solve this
- **Usage**: Validates design choices
- **Required**: For major features (Rust always included)

### 7. Success Metrics
- **Purpose**: Quantify benefits
- **Usage**: Justifies feature existence
- **Required**: Code reduction percentage minimum

### 8. External References
- **Purpose**: Link to additional context
- **Usage**: Deep dives and research
- **Required**: Go proposals when applicable

## Concise vs Comprehensive

**Comprehensive** (like error_prop_01_simple):
- Major features (error propagation, sum types, Result/Option)
- High community interest
- Complex design decisions
- Multiple language comparisons

**Concise** (like ternary_01_basic):
- Well-understood patterns
- Simple, straightforward features
- Minimal design decisions
- Focus on frontmatter + brief explanation

## Template Selection Guide

**Use Comprehensive Template When**:
- Feature addresses Go proposal
- Complex design tradeoffs exist
- Multiple implementation approaches considered
- Strong community discussion exists
- Feature is novel to Go ecosystem

**Use Concise Template When**:
- Feature is well-understood (ternary, null coalesce)
- Straightforward implementation
- Minimal design decisions
- Standard pattern from other languages

## Validation Checklist

Before committing a reasoning file:

- [ ] Valid YAML frontmatter (test with YAML parser)
- [ ] All required frontmatter fields present
- [ ] Title is descriptive, not generic
- [ ] Category/subcategory correct
- [ ] Code examples present (Dingo + Go)
- [ ] Code reduction percentage calculated
- [ ] Related tests cross-referenced
- [ ] Go proposal linked (if applicable)
- [ ] File compiles (for longer files)
- [ ] Spell-checked
- [ ] Markdown syntax valid

---

**Document Purpose**: Reference for creating new reasoning files
**Last Updated**: 2025-11-17

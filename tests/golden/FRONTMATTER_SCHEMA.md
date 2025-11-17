# Reasoning File Frontmatter Schema

This document defines the YAML frontmatter schema for `.reasoning.md` files used to enrich the golden tests documentation.

## Purpose

The frontmatter provides structured metadata for:
- **Navigation**: Building hierarchical menus with categories and subcategories
- **Presentation**: Descriptive titles and summaries for better UX
- **Context**: Linking to proposals, features, and external references
- **Metrics**: Code reduction stats and complexity indicators

## Schema Definition

```yaml
---
# === IDENTIFICATION ===
title: "Simple error propagation with ? operator"
category: "Error Handling"
subcategory: "Error Propagation"
test_id: "error_prop_01_simple"
order: 1

# === CLASSIFICATION ===
complexity: "basic"        # basic | intermediate | advanced
feature: "error-propagation"
phase: "Phase 2.4"
status: "implemented"      # implemented | planned | experimental

# === METADATA ===
description: "Demonstrates basic error propagation using the ? operator for simple return statements"
summary: "Basic ? operator usage in single statement"
code_reduction: 65         # Percentage (optional)
lines_dingo: 12           # Lines of Dingo code (optional)
lines_go: 34              # Lines of generated Go code (optional)

# === REFERENCES ===
go_proposal: "71203"       # Go proposal number (without #)
go_proposal_link: "https://github.com/golang/go/issues/71203"
feature_file: "error-propagation.md"
related_tests:
  - "error_prop_02_multiple"
  - "error_prop_03_expression"

# === EXTERNAL CONTEXT ===
tags:
  - "error-handling"
  - "operator"
  - "syntax-sugar"
keywords:
  - "? operator"
  - "error propagation"
  - "early return"
---
```

## Field Descriptions

### Identification Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `title` | string | ✅ Yes | **Human-readable title** displayed in menus. Should be descriptive and specific (e.g., "Simple error propagation with ? operator" not "Basic") |
| `category` | string | ✅ Yes | **Top-level category** for menu grouping (e.g., "Error Handling", "Type System", "Functional Programming") |
| `subcategory` | string | ✅ Yes | **Second-level category** for menu grouping (e.g., "Error Propagation", "Result Type", "Option Type") |
| `test_id` | string | ✅ Yes | **Test file identifier** (matches filename without extension, e.g., "error_prop_01_simple") |
| `order` | number | ✅ Yes | **Sort order** within subcategory (1-based, determines menu order) |

### Classification Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `complexity` | enum | ✅ Yes | **Complexity level**: `basic`, `intermediate`, `advanced` |
| `feature` | string | ✅ Yes | **Feature identifier** (slug format, e.g., "error-propagation", "sum-types") |
| `phase` | string | ⚠️ Recommended | **Development phase** (e.g., "Phase 2.4", "Phase 2.5") |
| `status` | enum | ⚠️ Recommended | **Implementation status**: `implemented`, `planned`, `experimental` |

### Metadata Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `description` | string | ✅ Yes | **Full description** (1-2 sentences, used for tooltips and preview cards) |
| `summary` | string | ⚠️ Recommended | **Short summary** (5-8 words, used in compact views) |
| `code_reduction` | number | ⬜ Optional | **Code reduction percentage** (e.g., 65 means 65% less code than Go) |
| `lines_dingo` | number | ⬜ Optional | **Lines of Dingo code** in the test |
| `lines_go` | number | ⬜ Optional | **Lines of generated Go code** |

### Reference Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `go_proposal` | string | ⬜ Optional | **Go proposal number** (without #, e.g., "71203") |
| `go_proposal_link` | string | ⬜ Optional | **Full URL to Go proposal** |
| `feature_file` | string | ⬜ Optional | **Feature spec filename** (e.g., "error-propagation.md" in features/) |
| `related_tests` | string[] | ⬜ Optional | **Array of related test IDs** (for cross-linking) |

### External Context Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `tags` | string[] | ⬜ Optional | **Semantic tags** for filtering/search (e.g., ["error-handling", "operator"]) |
| `keywords` | string[] | ⬜ Optional | **Search keywords** (e.g., ["? operator", "early return"]) |

## Category Hierarchy

The proposed category structure for all 45 tests:

### 1. Error Handling
- **Error Propagation** (8 tests)
  - Simple, multiple, expression, wrapping, complex types, mixed context, special chars, chained calls
- **Result Type** (5 tests)
  - Basic, propagation, pattern match, chaining, Go interop

### 2. Type System
- **Option Type** (4 tests)
  - Basic, pattern match, chaining, Go interop
- **Sum Types** (5 tests)
  - Simple enum, struct variant, generic, multiple, nested

### 3. Functional Programming
- **Lambda Functions** (4 tests)
  - Basic, multiline, closure, higher-order
- **Functional Utilities** (4 tests)
  - Map, filter, reduce, chaining

### 4. Control Flow
- **Pattern Matching** (4 tests)
  - Basic, guards, nested, exhaustive
- **Ternary Operator** (3 tests)
  - Basic, nested, complex

### 5. Operators
- **Null Coalescing** (3 tests)
  - Basic, chained, with option
- **Safe Navigation** (3 tests)
  - Basic, chained, with methods

### 6. Data Structures
- **Tuples** (3 tests)
  - Basic, destructure, nested

## Complexity Levels

| Level | Definition | Examples |
|-------|-----------|----------|
| **basic** | Single feature, minimal code, no edge cases | `error_prop_01_simple`, `lambda_01_basic` |
| **intermediate** | Multiple related concepts, realistic usage | `error_prop_04_wrapping`, `lambda_03_closure` |
| **advanced** | Complex patterns, edge cases, interop | `pattern_match_04_exhaustive`, `sum_types_05_nested` |

## Example: Complete Frontmatter

```yaml
---
# Simple error propagation test
title: "Simple error propagation with ? operator"
category: "Error Handling"
subcategory: "Error Propagation"
test_id: "error_prop_01_simple"
order: 1

# Classification
complexity: "basic"
feature: "error-propagation"
phase: "Phase 2.4"
status: "implemented"

# Metadata
description: "Demonstrates basic error propagation using the ? operator for simple return statements, eliminating verbose if-err-return boilerplate"
summary: "Basic ? operator usage"
code_reduction: 65
lines_dingo: 12
lines_go: 34

# References
go_proposal: "71203"
go_proposal_link: "https://github.com/golang/go/issues/71203"
feature_file: "error-propagation.md"
related_tests:
  - "error_prop_02_multiple"
  - "result_02_propagation"

# Context
tags:
  - "error-handling"
  - "operator"
  - "syntax-sugar"
keywords:
  - "? operator"
  - "error propagation"
  - "early return"
  - "if err return"
---
```

## Usage in Astro Content Collections

The frontmatter will be parsed by Astro's Content Collections API:

```typescript
// src/content.config.ts
import { defineCollection, z } from 'astro:content';

const reasoningCollection = defineCollection({
  type: 'content',
  schema: z.object({
    // Identification
    title: z.string(),
    category: z.string(),
    subcategory: z.string(),
    test_id: z.string(),
    order: z.number(),

    // Classification
    complexity: z.enum(['basic', 'intermediate', 'advanced']),
    feature: z.string(),
    phase: z.string().optional(),
    status: z.enum(['implemented', 'planned', 'experimental']).optional(),

    // Metadata
    description: z.string(),
    summary: z.string().optional(),
    code_reduction: z.number().optional(),
    lines_dingo: z.number().optional(),
    lines_go: z.number().optional(),

    // References
    go_proposal: z.string().optional(),
    go_proposal_link: z.string().url().optional(),
    feature_file: z.string().optional(),
    related_tests: z.array(z.string()).optional(),

    // Context
    tags: z.array(z.string()).optional(),
    keywords: z.array(z.string()).optional(),
  }),
});

export const collections = {
  'reasoning': reasoningCollection,
};
```

## Migration Strategy

1. **Create frontmatter** for all 45 reasoning files
2. **Preserve existing content** (move to body section)
3. **Validate schema** using Astro's type checking
4. **Build hierarchical menu** from category/subcategory fields
5. **Sort by order** within each subcategory

## Benefits

✅ **Better Navigation**: Clear hierarchy with categories and subcategories
✅ **Descriptive Titles**: "Simple error propagation" instead of "Basic"
✅ **Type Safety**: Zod schema ensures data consistency
✅ **Searchable**: Tags and keywords enable filtering
✅ **Metrics Visible**: Code reduction displayed in UI
✅ **Cross-linking**: Related tests easily discoverable
✅ **Community Context**: Go proposals linked directly

---

**Version**: 1.0
**Last Updated**: 2025-11-17
**Status**: ✅ Ready for implementation

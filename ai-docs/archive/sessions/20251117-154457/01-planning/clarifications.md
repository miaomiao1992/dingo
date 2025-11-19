# User Clarifications

## Question 1: Generic Type Parameters
**Answer**: Full generic enum system (complex, 15-20 hours)

User selected the most comprehensive approach - implementing complete generic type parameters for all enums. This will provide the most powerful and flexible system.

## Question 2: None Type Inference
**Answer**: Make both options available via dingo.toml configuration

User wants to support BOTH:
1. Explicit type annotation required: `let x: Option = None`
2. Context-based inference: Infer type from assignment/return context

This should be configurable via the existing dingo options system in dingo.toml. The configuration follows this pattern:
```toml
[features]
# Example from existing config:
# safe_navigation_unwrap = "smart"  # or "always_option"

# New option to add:
# none_type_inference = "explicit" | "context"
```

## Question 3: Helper Methods
**Answer**: Comprehensive set (Unwrap, Map, AndThen, Filter, etc.)

User wants all 8-10 helper methods per type for the best developer experience.

## Implications
- Implementation timeline: 15-20 hours (full generics + comprehensive helpers)
- Need to research dingo.toml configuration system to understand how to add options
- Higher complexity but better long-term foundation

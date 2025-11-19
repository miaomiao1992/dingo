# User Clarifications

## Question 1: Should we support guards on ALL levels of nested patterns?
**Answer**: Yes, support guards at all nesting levels

**Implication**: Need to implement guard support not just for innermost patterns, but also for outer patterns. Example:
```dingo
Result_Ok(Option_Some(val)) where opt.IsSome() && val > 0 => ...
```

## Question 2: Should we enforce a maximum nesting depth for patterns with guards?
**Answer**: Yes, limit to 2 levels (reasonable)

**Implication**: Add validation to error if nesting depth exceeds 2 levels with guards. This keeps code readable and prevents excessive indentation.

## Question 3: Should we add a dedicated golden test for nested patterns with guards?
**Answer**: Yes, create pattern_match_13_nested_guards_simple.dingo

**Implication**: Create a new, simpler test file that isolates the nested-pattern-with-guard feature for regression testing.

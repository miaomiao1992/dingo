# User Clarifications

## Question 1: Priority Order
**Decision:** Work on both in parallel
- Fix position info bug AND implement pattern destructuring simultaneously
- This maximizes efficiency and gets both done faster

## Question 2: Match Expression IIFE Wrapping
**Decision:** Implement now (Phase 2.5)
- Don't defer to Phase 3
- Add IIFE wrapping for expression contexts in this session
- Complete implementation even though it adds 4-6 hours

## Question 3: Nil Safety for Pattern Destructuring
**Decision:** Make it a configurable feature
- Add runtime nil checks as an **optional feature** controlled by settings
- Not binary (always on vs always off)
- **Three options switchable in settings:**
  1. **Off** - No nil checks (trust constructors, maximum performance)
  2. **On** - Always check for nil (safe, some overhead)
  3. **Debug only** - Checks in debug builds, optimized out in release

This gives users flexibility to choose their safety vs performance tradeoff.

## Implementation Notes

### Nil Safety Settings Design
- Add to `dingo.toml` config file:
  ```toml
  [transpiler]
  nil_safety_checks = "on"  # Options: "off", "on", "debug"
  ```
- Add to plugin configuration system
- Generator reads setting and conditionally emits nil check code
- Default: "on" (prioritize safety for Phase 2)

### Parallel Work Strategy
- Position info bug: Simpler, affects all generated declarations
- Pattern destructuring: More complex, specific to match expressions
- Can work independently with minimal conflicts
- Test both features together at end

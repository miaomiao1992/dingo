# User Request: Investigate Current State and Suggest Next Steps

## Context

The project recently completed Phase 2.11 - Build System Fixes with the following changes:
- Removed duplicate error propagation implementation from transformer
- Added automatic import detection in preprocessor
- Implemented import injection pipeline
- Updated source mappings for injected imports
- Architecture clarification: preprocessor vs transformer responsibilities

## Current Situation

An external code review (GPT-5.1 Codex via claudish) has identified critical issues:

### CRITICAL Issues (2)
1. **Source-map offset bug** - Shifts ALL mappings when imports injected, even for lines before the import block
2. **Multi-value return handling** - Error propagation `return expr?` drops extra non-error values for multi-value returns

### IMPORTANT Issues (2)
1. **Import detection false positives** - User-defined functions with stdlib names trigger unwanted imports
2. **Missing negative tests** - No tests for user-defined function shadowing or mapping offset bugs

## Request

Investigate the current project state and suggest prioritized next steps that:

1. **Address Critical Issues** - Fix the bugs identified in code review
2. **Consider Architecture** - Align with preprocessor/transformer separation
3. **Maintain Quality** - Follow code review best practices
4. **Plan Next Phase** - Suggest what features/improvements to tackle next

## Success Criteria

- All CRITICAL issues have clear fix plans
- IMPORTANT issues are either fixed or deferred with rationale
- Next development phase is well-defined with concrete tasks
- Testing strategy ensures no regressions

# User Clarifications

## Question 1: Already Fixed Issues?
**Answer**: Skip implementation, only add comprehensive tests

**Implication**: Issues #2 (multi-value returns) and #3 (import collision) are already fixed in the codebase. We will verify with tests and mark as 'already fixed' in the review response.

## Question 2: Multi-Value Return Support
**Answer**: Make it configurable via compiler flag

**Implication**: We need to add a compiler flag to control whether `return expr?` supports multi-value propagation. This is a NEW feature requirement, not just a bug fix. Default behavior should match current implementation (full multi-value support).

## Question 3: Future Preprocessor Policy
**Answer**: All processors MUST run before import injection (recommended)

**Implication**: This simplifies the architecture. Import injection is always the final step. The Issue #1 fix (source map offset correction) will work for all future processors. Document this policy in the preprocessor README.

## Summary

**Adjusted Scope**:
1. **Issue #1**: Fix source map offset bug (1-line change)
2. **Issue #2**: Already fixed - add tests only
3. **Issue #3**: Already fixed - add tests only
4. **Issue #4**: Add comprehensive negative tests
5. **NEW**: Add compiler flag for multi-value return mode
6. **NEW**: Document preprocessor ordering policy

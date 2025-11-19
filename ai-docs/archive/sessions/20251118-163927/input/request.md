Command:
Use the Task tool with the following parameters:
- subagent_type: code-reviewer
- description: Review Phase 4.1 implementation
- prompt: [detailed task description below]

Task Description for code-reviewer agent:

You are operating in proxy mode via claudish. Review the following Phase 4.1 implementation for the Dingo project.

Project Context:
- Working directory: /Users/jack/mag/dingo
- Project: Dingo transpiler (meta-language for Go)
- Phase: 4.1 MVP - Pattern Matching + None Inference
- Implementation: Completed (57/57 unit tests passing, 98% integration test pass rate)
- Guidelines: See CLAUDE.md and ai-docs/ directory
- Key technologies: Go, go/parser, go/ast, go/types, source maps

Review Task: Comprehensive code review of Phase 4.1 implementation

Implementation Summary:
Phase 4.1 delivered basic pattern matching with exhaustiveness checking and None type inference. The implementation spans 7 major tasks across 5 parallel batches:

BATCH 1 (Parallel Foundation):
- Task A: Configuration system (dingo.toml support)
  Files: pkg/config/config.go, pkg/config/loader.go, pkg/config/config_test.go
  Tests: 11/11 passing
  
- Task B: AST Parent Tracking
  Files: pkg/plugin/context.go (BuildParentMap, GetParent, WalkParents)
  Tests: 14/14 passing

BATCH 2 (Preprocessor):
- Task C: Rust Pattern Match Preprocessor
  Files: pkg/preprocessor/rust_match.go, rust_match_test.go
  Tests: 12/12 passing
  Golden: pattern_match_01_simple.dingo

BATCH 3 (Parallel Plugins):
- Task D: Pattern Match Plugin (Discovery + Exhaustiveness)
  Files: pkg/plugin/builtin/pattern_match.go, pattern_match_test.go
  Tests: 10/10 passing
  Golden: pattern_match_02_exhaustive.dingo
  
- Task E: None Context Inference Plugin
  Files: pkg/plugin/builtin/none_context.go, none_context_test.go
  Tests: 8/8 passing
  Golden: option_06_none_inference.dingo

BATCH 4 (Transformation):
- Task F: Pattern Match Transformation
  Extended: pkg/plugin/builtin/pattern_match.go (Transform phase)
  Tests: 12 total passing
  Golden: pattern_match_03_result_option.dingo

BATCH 5 (Integration):
- Task G: Generator Integration & E2E Tests
  Modified: pkg/generator/generator.go (24 lines)
  Created: tests/integration_phase4_test.go (408 lines, 4 tests)
  Pass rate: 98%

Review Focus Areas:
1. **Correctness & Bugs**
   - Are the algorithms correct? (exhaustiveness checking, type inference)
   - Are there edge cases not handled?
   - Are there race conditions or concurrency issues?
   - Is error handling comprehensive?

2. **Go Best Practices**
   - Idiomatic Go code?
   - Proper error handling (not panic-driven)?
   - Clear naming and documentation?
   - Effective use of standard library?

3. **Performance**
   - Is parent map construction efficient (<10ms target)?
   - Is exhaustiveness checking efficient (<1ms target)?
   - Are there unnecessary allocations or copies?
   - Benchmark results: Parent map <10ms ✅, Exhaustiveness <1ms ✅

4. **Architecture & Maintainability**
   - Does design align with two-stage pipeline?
   - Are responsibilities clear and separated?
   - Is the code testable and well-tested?
   - Will this code age well?

5. **Test Quality**
   - Are tests comprehensive? (57 unit tests, 4 integration tests)
   - Do they cover edge cases?
   - Are golden tests realistic?
   - Is the 98% pass rate acceptable for release?

Specific Files to Review:

Configuration System (Task A):
- pkg/config/config.go
- pkg/config/loader.go
- pkg/config/config_test.go

AST Parent Tracking (Task B):
- pkg/plugin/context.go (BuildParentMap, GetParent, WalkParents)
- pkg/plugin/context_test.go

Rust Match Preprocessor (Task C):
- pkg/preprocessor/rust_match.go
- pkg/preprocessor/rust_match_test.go

Pattern Match Plugin (Tasks D & F):
- pkg/plugin/builtin/pattern_match.go
- pkg/plugin/builtin/pattern_match_test.go

None Context Inference (Task E):
- pkg/plugin/builtin/none_context.go
- pkg/plugin/builtin/none_context_test.go

Integration (Task G):
- pkg/generator/generator.go (modified 24 lines)
- tests/integration_phase4_test.go (408 lines)

Golden Tests:
- tests/golden/pattern_match_01_simple.dingo
- tests/golden/pattern_match_02_exhaustive.dingo
- tests/golden/pattern_match_03_result_option.dingo
- tests/golden/option_06_none_inference.dingo

Expected Output Format:

### ✅ Strengths
- List what the code does well
- Acknowledge good practices

### ⚠️ Concerns
For each issue, specify:
- **Category** (Correctness/Performance/Maintainability/Testing/Go Best Practices)
- **Severity** (CRITICAL/IMPORTANT/MINOR)
- **Issue**: Specific problem description
- **Location**: File and line number if applicable
- **Impact**: Why this matters
- **Recommendation**: Concrete fix with code example when helpful

### 🔍 Questions
- Clarifying questions about intent or requirements
- Areas where more context would improve the review

### 📊 Summary
- Overall assessment (APPROVED / CHANGES_NEEDED / MAJOR_ISSUES)
- Priority ranking of recommendations
- Testability score (High/Medium/Low) with justification
- Final counts: CRITICAL: N | IMPORTANT: N | MINOR: N

CRITICAL Requirements:
1. Categorize ALL issues by severity (CRITICAL/IMPORTANT/MINOR)
2. Provide specific file/line locations for issues
3. Include concrete recommendations with examples
4. Focus on real issues, not style preferences
5. Consider Go idioms and standard library alternatives

CRITICAL: Do not just acknowledge this request - actually invoke the code-reviewer agent using the Task tool and have it perform the review.

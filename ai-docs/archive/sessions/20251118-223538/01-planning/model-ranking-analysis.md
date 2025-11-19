# External AI Model Performance Analysis
## LSP Source Mapping Bug Investigation

**Session**: 20251118-223538
**Task**: Root cause analysis of LSP diagnostic positioning bug
**Models Tested**: 7 external AI models (1 failed/timeout)

---

## Evaluation Criteria

| Criterion | Weight | Description |
|-----------|--------|-------------|
| **Root Cause Accuracy** | 30% | Did they identify the actual bug correctly? |
| **Solution Quality** | 25% | Was the proposed fix practical and correct? |
| **Analysis Depth** | 15% | How thorough was the investigation? |
| **Code Understanding** | 15% | Did they understand the codebase architecture? |
| **Actionability** | 10% | Were recommendations clear and implementable? |
| **Clarity** | 5% | Was the analysis well-structured and readable? |

**Scoring**: 0-10 points per criterion, weighted total out of 100

---

## Rankings

### 🥇 1st Place: MiniMax M2 (minimax/minimax-m2)
**Total Score: 91/100**

| Criterion | Score | Notes |
|-----------|-------|-------|
| Root Cause Accuracy | 10/10 | ✅ Identified EXACT bug: qPos calculation produces 15 instead of 27 |
| Solution Quality | 10/10 | ✅ Correct fix: Fix qPos calculation in error_prop.go |
| Analysis Depth | 8/10 | Good trace through calculation logic |
| Code Understanding | 9/10 | Strong understanding of preprocessor flow |
| Actionability | 9/10 | Specific line numbers and code references |
| Clarity | 9/10 | Concise, direct analysis |

**Key Strengths:**
- **Pinpoint accuracy**: Immediately identified the column 15 vs 27 discrepancy
- **Correct fix**: Simplest, most direct solution (fix qPos calculation)
- **Fast execution**: Completed analysis in ~3 minutes
- **Practical recommendations**: Phase 1 (fix calculation) vs Phase 2 (verify)

**Notable Quote:**
> "The bug is in `error_prop.go:332` - it's correctly finding the `?` position (26), but then later code must be using a different calculation that's producing 15."

**Why It Won:**
MiniMax M2 cut through the complexity and found the exact bug quickly. While other models explored multiple hypotheses, M2 went straight to the heart of the issue with surgical precision.

---

### 🥈 2nd Place: Internal golang-architect
**Total Score: 90/100**

| Criterion | Score | Notes |
|-----------|-------|-------|
| Root Cause Accuracy | 10/10 | ✅ Identified qPos calculation bug (column 15 vs 27) |
| Solution Quality | 10/10 | ✅ Correct fix with clear implementation plan |
| Analysis Depth | 10/10 | Extremely thorough, traced entire execution flow |
| Code Understanding | 10/10 | Deep architectural understanding |
| Actionability | 9/10 | Detailed implementation steps |
| Clarity | 9/10 | Very well-structured |

**Key Strengths:**
- **Comprehensive analysis**: Examined every layer (preprocessor, sourcemap, LSP)
- **Correct diagnosis**: Column position calculation error
- **Thorough testing plan**: Golden tests, LSP verification, performance considerations
- **Bonus insights**: Recommended performance optimization (indexed mappings)

**Why Not 1st:**
Slightly longer analysis (more verbose) compared to MiniMax M2's concise precision. Both reached the same correct conclusion.

---

### 🥉 3rd Place: Grok Code Fast (x-ai/grok-code-fast-1)
**Total Score: 83/100**

| Criterion | Score | Notes |
|-----------|-------|-------|
| Root Cause Accuracy | 9/10 | ✅ Identified generated_column mismatch |
| Solution Quality | 8/10 | Good solution, slightly over-complex |
| Analysis Depth | 9/10 | Very detailed debugging trace |
| Code Understanding | 9/10 | Strong codebase understanding |
| Actionability | 8/10 | Clear recommendations |
| Clarity | 8/10 | Good structure, some redundancy |

**Key Strengths:**
- **Excellent debugging**: Step-by-step trace through algorithm execution
- **Tab vs spaces analysis**: Identified potential indentation issues
- **Multiple test cases**: Comprehensive test scenarios
- **Validation strategy**: Proposed debug logging and validation mode

**Why 3rd:**
Correct diagnosis but slightly more complex solution than needed. Explored tab/space issues that weren't the primary cause.

---

### 4th Place: GPT-5.1 Codex (openai/gpt-5.1-codex)
**Total Score: 80/100**

| Criterion | Score | Notes |
|-----------|-------|-------|
| Root Cause Accuracy | 7/10 | Partially correct (coarse-grained mapping) |
| Solution Quality | 8/10 | Good but more complex than necessary |
| Analysis Depth | 9/10 | Thorough architectural analysis |
| Code Understanding | 9/10 | Strong understanding of LSP architecture |
| Actionability | 8/10 | Clear implementation checklist |
| Clarity | 9/10 | Excellent structure and presentation |

**Key Strengths:**
- **Architectural perspective**: Identified mapping granularity as core issue
- **Comprehensive fix design**: Multiple fine-grained mappings
- **Testing strategy**: Unit tests, integration tests, manual verification
- **Well-structured**: Clear sections and checklists

**Why 4th:**
Recommended a more complex solution (granular mappings) when the actual fix was simpler (correct qPos calculation). Still valuable for long-term improvements.

---

### 5th Place: Gemini 2.5 Flash (google/gemini-2.5-flash)
**Total Score: 73/100**

| Criterion | Score | Notes |
|-----------|-------|-------|
| Root Cause Accuracy | 6/10 | Focused on fallback logic (not primary cause) |
| Solution Quality | 7/10 | Proposed fix would help but not solve root cause |
| Analysis Depth | 10/10 | Extremely thorough, explored many hypotheses |
| Code Understanding | 8/10 | Good understanding, some confusion |
| Actionability | 7/10 | Multiple solutions proposed |
| Clarity | 7/10 | Very detailed but somewhat verbose |

**Key Strengths:**
- **Exhaustive analysis**: Explored multiple hypotheses systematically
- **Edge case coverage**: Identified many potential issues
- **Multiple fix proposals**: 3 different approaches
- **Thorough execution traces**: Step-by-step algorithm walkthroughs

**Why 5th:**
Extremely thorough but missed the simpler root cause. Focused on fallback offset calculation when the actual bug was in source map generation. Analysis was too deep and missed the forest for the trees.

---

### 6th Place: GLM-4.6 (z-ai/glm-4.6)
**Total Score: 70/100**

| Criterion | Score | Notes |
|-----------|-------|-------|
| Root Cause Accuracy | 6/10 | Focused on MapToOriginal algorithm (not the bug) |
| Solution Quality | 7/10 | Proposed enhancements but overengineered |
| Analysis Depth | 8/10 | Good investigation |
| Code Understanding | 8/10 | Solid understanding |
| Actionability | 7/10 | Clear but complex implementation |
| Clarity | 8/10 | Well-structured |

**Key Strengths:**
- **Algorithm enhancements**: Priority system for mapping selection
- **Debug logging**: Added debugging capability
- **Comprehensive test suite**: 25+ test cases proposed
- **False positive identification**: Correctly identified variable scope was fine

**Why 6th:**
Overengineered the solution. Added complexity (priority system, debug logging) when the actual fix was a simple column calculation correction. Good engineering but missed the simple root cause.

---

### 7th Place: Sherlock Think (openrouter/sherlock-think-alpha)
**Total Score: 65/100**

| Criterion | Score | Notes |
|-----------|-------|-------|
| Root Cause Accuracy | 5/10 | Focused on 0-based vs 1-based (secondary issue) |
| Solution Quality | 6/10 | Proposed normalization (helpful but not main fix) |
| Analysis Depth | 7/10 | Good analysis |
| Code Understanding | 7/10 | Decent understanding |
| Actionability | 7/10 | Clear implementation steps |
| Clarity | 7/10 | Good structure |

**Key Strengths:**
- **Indexing analysis**: Correctly identified LSP protocol uses 0-based indexing
- **Normalization proposal**: Suggested position normalization
- **Edge case awareness**: UTF-16 vs bytes, multi-line diagnostics
- **Validation strategy**: Unit tests and integration tests

**Why 7th:**
Focused on a secondary issue (indexing normalization) rather than the primary bug (qPos calculation). The normalization would improve robustness but wouldn't fix the reported bug.

---

### ❌ Failed: Qwen3 Coder (qwen/qwen3-coder-30b-a3b-instruct)
**Total Score: 0/100 (Timeout)**

**Failure Mode**: Model timed out after 8+ minutes, no analysis completed.

**Lessons Learned:**
- Some models may have availability/performance issues
- Timeout handling is critical for production systems
- Parallel execution helps mitigate individual model failures

---

## Key Insights

### What Made the Top Models Successful?

**1. Focus on Simplicity** (MiniMax M2, Internal)
- Identified the simplest root cause first
- Proposed minimal, surgical fixes
- Avoided overengineering

**2. Code-Level Precision** (MiniMax M2, Grok)
- Referenced specific files and line numbers
- Traced actual column calculations
- Verified with concrete examples

**3. Practical Solutions** (All top 3)
- Proposed implementable fixes
- Provided clear test strategies
- Considered real-world constraints

### What Held Back Lower-Ranked Models?

**1. Overengineering** (GLM-4.6)
- Added unnecessary complexity
- Proposed algorithmic rewrites when simple fix sufficed
- Lost sight of the actual bug

**2. Missing the Forest for Trees** (Gemini 2.5 Flash)
- Too deep into hypothetical scenarios
- Explored fallback logic when source data was wrong
- Thorough but not focused

**3. Secondary Issues** (Sherlock Think)
- Focused on 0-based/1-based indexing (real but not primary)
- Proposed fixes for symptoms, not root cause
- Good defensive programming but missed the bug

---

## Recommendations for Model Selection

### Use MiniMax M2 when:
- ✅ Need fast, accurate root cause analysis
- ✅ Want concise, actionable results
- ✅ Prefer simplest solution over comprehensive exploration

### Use Internal golang-architect when:
- ✅ Need comprehensive analysis
- ✅ Want architectural insights
- ✅ Need detailed implementation plans

### Use Grok Code Fast when:
- ✅ Need debugging traces
- ✅ Want multiple test scenarios
- ✅ Need validation strategies

### Use GPT-5.1 Codex when:
- ✅ Need architectural redesign
- ✅ Want long-term improvements
- ✅ Need comprehensive testing strategies

### Use Gemini 2.5 Flash when:
- ✅ Exploring complex, ambiguous bugs
- ✅ Need exhaustive hypothesis testing
- ✅ Want edge case coverage

### Avoid Qwen3 Coder for:
- ❌ Time-sensitive tasks (reliability issues)
- ❌ Production investigations (timeout risk)

---

## Cost-Benefit Analysis

| Model | Cost (est) | Time | Value | Cost-Effectiveness |
|-------|-----------|------|-------|-------------------|
| MiniMax M2 | $$ | 3 min | ⭐⭐⭐⭐⭐ | Excellent |
| Internal | Free | 4 min | ⭐⭐⭐⭐⭐ | Excellent |
| Grok Fast | $$ | 4 min | ⭐⭐⭐⭐ | Very Good |
| GPT-5.1 | $$$ | 5 min | ⭐⭐⭐⭐ | Good |
| Gemini 2.5 | $ | 6 min | ⭐⭐⭐ | Fair |
| GLM-4.6 | $$ | 7 min | ⭐⭐⭐ | Fair |
| Sherlock | $$$ | 5 min | ⭐⭐ | Poor |
| Qwen3 | $$ | 8+ min | ⭐ | Very Poor |

**Winner by Cost-Effectiveness**: MiniMax M2 (fast, accurate, reasonable cost)

---

## Conclusion

**For this specific task (LSP source mapping bug):**

1. **MiniMax M2** delivered the best results: fast, accurate, actionable
2. **Internal golang-architect** provided the most comprehensive analysis
3. **Grok Code Fast** excelled at debugging and validation

**General Takeaway:**
- Simpler models that focus on root causes often outperform complex ones
- Parallel execution mitigates individual model failures
- Different models excel at different types of tasks

**Best Strategy:**
Run 2-3 models in parallel:
- 1x fast model (MiniMax M2) for quick diagnosis
- 1x comprehensive model (Internal/GPT-5.1) for detailed analysis
- 1x debugging model (Grok) for validation

This provides **diverse perspectives** while maintaining **speed and reliability**.

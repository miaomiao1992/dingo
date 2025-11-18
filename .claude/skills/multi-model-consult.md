# Multi-Model Consultation Skill

You are executing the **Multi-Model Consultation** pattern. This skill helps you consult multiple external LLMs in parallel to get diverse perspectives on architectural decisions, design choices, or complex analysis.

## Your Task

The user wants perspectives from multiple external models. Follow these steps EXACTLY:

### Step 1: Create Session Folder

```bash
SESSION=$(date +%Y%m%d-%H%M%S)
mkdir -p ai-docs/sessions/$SESSION/{input,output}
```

### Step 2: Write Investigation Prompt

Extract the user's question/investigation topic and write a comprehensive prompt to:
`ai-docs/sessions/$SESSION/input/investigation-prompt.md`

The prompt should:
- Clearly state the question/problem
- Provide necessary context about Dingo project
- Ask for specific analysis or recommendations
- Be self-contained (model won't have conversation history)

### Step 3: Identify Models to Consult

**EVIDENCE-BASED RECOMMENDATIONS** (validated via Session 20251118-223538):

#### 🥇 Tier 1: Primary Recommendations (Use First)

**Default Fast Diagnosis** (90% of use cases):
- `minimax/minimax-m2` - **BEST PERFORMER** (Score: 91/100, 3 min, pinpoint accuracy)
- `x-ai/grok-code-fast-1` - **DEBUGGING EXPERT** (Score: 83/100, 4 min, excellent traces)

**Default Comprehensive Analysis**:
- `minimax/minimax-m2` - Fast + accurate
- `openai/gpt-5.1-codex` - Architectural vision (Score: 80/100, 5 min)
- `x-ai/grok-code-fast-1` - Validation + testing

#### 🥈 Tier 2: Specialized Use Cases

- `google/gemini-2.5-flash` - Ambiguous problems, exhaustive analysis (Score: 73/100, 6 min, LOW COST)
- `z-ai/glm-4.6` - Algorithm enhancements, debug infrastructure (Score: 70/100, 7 min)

#### ⚠️ Use With Caution

- `openrouter/sherlock-think-alpha` - Protocol compliance only (Score: 65/100, 5 min, HIGH COST)

#### ❌ AVOID (Known Issues)

- `qwen/qwen3-coder-30b-a3b-instruct` - **UNRELIABLE** (Timeout after 8+ min, 0% success rate)

#### Other Available Models (Not Yet Validated)

- `openai/gpt-5` - Most advanced reasoning (not tested yet)
- `qwen/qwen3-vl-235b-a22b-instruct` - Multimodal (not tested yet)
- `openrouter/polaris-alpha` - FREE experimental (not tested yet)

**Validation Date**: 2025-11-18 | **Re-test**: Every 3-6 months

---

### Step 3.5: Choose Consultation Strategy

Based on task type and priority, select one of these proven strategies:

#### Strategy 1: Fast Parallel Diagnosis (DEFAULT - 90% of use cases)
```
Models: minimax/minimax-m2 + x-ai/grok-code-fast-1
Time: ~4 minutes total
Use for: Bug investigations, quick diagnosis
Benefits: Fast fix + validation
```

#### Strategy 2: Comprehensive Analysis (Critical bugs)
```
Models: minimax/minimax-m2 + openai/gpt-5.1-codex + x-ai/grok-code-fast-1
Time: ~5 minutes total
Use for: Critical bugs, architectural decisions
Benefits: Quick fix + long-term plan + validation
```

#### Strategy 3: Deep Exploration (Ambiguous problems)
```
Models: minimax/minimax-m2 + google/gemini-2.5-flash + x-ai/grok-code-fast-1
Time: ~6 minutes total
Use for: Ambiguous, multi-faceted problems
Benefits: Quick fix + exhaustive analysis + validation
```

#### Strategy 4: Budget-Conscious (Cost-sensitive)
```
Models: google/gemini-2.5-flash + x-ai/grok-code-fast-1
Time: ~6 minutes total
Use for: When minimizing cost is priority
Benefits: Low-cost exploration + good validation
```

**Decision Tree**:
```
[What's the task?]
  ↓
[Bug Investigation?] → Strategy 1 (MiniMax + Grok)
[Architectural Decision?] → Strategy 2 (MiniMax + GPT-5.1 + Grok)
[Ambiguous Problem?] → Strategy 3 (MiniMax + Gemini + Grok)
[Cost-Sensitive?] → Strategy 4 (Gemini + Grok)
```

---

### Step 4: Select Appropriate Agent Type

Based on the domain:
- **Go project questions** (parser, AST, transpiler) → `golang-architect`
- **Astro/landing page questions** → `astro-developer`
- **General code review** → `code-reviewer`
- **Multi-language** → `general-purpose` (last resort)

### Step 5: Launch Agents in Parallel

**CRITICAL**: Launch ALL agents in a **SINGLE MESSAGE** with multiple Task tool calls.

For each model, create a Task call like this:

```
Task tool → [agent-type]:

You are operating in PROXY MODE to [task description] using [model-name].

INPUT FILES:
- Investigation prompt: ai-docs/sessions/$SESSION/input/investigation-prompt.md

YOUR TASK (PROXY MODE):
1. Read the investigation prompt from input file
2. Use claudish to invoke [model-name] (model ID: [model-id])
3. Provide the model with the investigation prompt
4. Write complete response to output file

**CRITICAL - Timeout Configuration**:
When executing claudish via Bash tool, ALWAYS use:
```
Bash(
    command='cat ai-docs/sessions/$SESSION/input/investigation-prompt.md | claudish --model [model-id] > ai-docs/sessions/$SESSION/output/[model-name]-analysis.md 2>&1',
    timeout=600000,  # 10 minutes (REQUIRED - default 2 min will timeout!)
    description='External consultation via [model-name]'
)
```

**Why 10-minute timeout?**: External models take 5-10 minutes. Default 2-minute timeout causes failures.

OUTPUT FILES (write full details here):
- ai-docs/sessions/$SESSION/output/[model-name]-analysis.md - Complete analysis

RETURN MESSAGE (keep this brief - MAX 3 lines):
Return ONLY this format:
[Model-name] analysis complete
Root cause: [one-line summary]
Full analysis: ai-docs/sessions/$SESSION/output/[model-name]-analysis.md

DO NOT return the full analysis in your response - it causes context bloat.
```

**Example** (3 models in parallel):
- Launch 3 Task calls in ONE message
- Each Task uses same agent type (e.g., golang-architect)
- Each Task invokes different model
- Each Task saves to different output file

### Step 6: Aggregate Results

After receiving all summaries:
1. Present brief overview to user (which models were consulted)
2. Show 1-sentence key finding from each model
3. Provide file paths for detailed analyses
4. Ask if user wants a consolidation analysis

### Step 7: Optional Consolidation

If user wants synthesis:
- Launch ONE final agent (same type)
- Agent reads ALL analysis files
- Agent synthesizes consensus + disagreements
- Agent writes: `ai-docs/sessions/$SESSION/output/CONSOLIDATED.md`
- Agent returns brief summary

## Key Rules

1. ✅ **Always use specialized agents** (golang-architect for Go, etc.)
2. ✅ **Always launch in parallel** (single message, multiple Task calls)
3. ✅ **Each agent = one model** (1:1 mapping)
4. ✅ **Communication via files** (full analysis → files, brief summary → response)
5. ✅ **Agent uses Bash** to invoke claudish (not Task tool)
6. ❌ **Never use general-purpose** unless no specialized agent exists

## Example Execution

```
User: "Should we use regex preprocessor or migrate to tree-sitter?"

You (main chat):
1. Create session: ai-docs/sessions/20251118-160000/
2. Write investigation-prompt.md with detailed question
3. Launch 3 golang-architect agents IN PARALLEL (one message):
   - Task 1: gpt-5.1-codex
   - Task 2: gemini-2.5-flash
   - Task 3: grok-code-fast-1
4. Receive 3 summaries
5. Present to user:
   "Consulted 3 models:
    - GPT-5.1-Codex: Recommends regex for simplicity, tree-sitter for future
    - Gemini-2.5-Flash: Suggests hybrid approach
    - Grok: Advocates staying with regex

    Details: ai-docs/sessions/20251118-160000/output/
    Want me to synthesize these perspectives?"
```

## Success Metrics

- **Speed**: 2-3 minutes (parallel) vs 5-10 minutes (sequential)
- **Context**: 50-100 lines (files) vs 500-1000 lines (inline)
- **Quality**: Diverse perspectives + domain expertise

## What to Return to User

After execution completes:
1. Brief summary of what models were consulted
2. One-line key insight from each model
3. Session folder path
4. Ask if consolidation needed

**Total output**: < 20 lines in main chat
**Detailed analyses**: In session files

---

**Remember**: You are the orchestrator. Delegate the actual model invocations to specialized agents. Keep main chat lean!

---

## Performance Benchmarks (Evidence-Based)

**Validation Task**: LSP Source Mapping Bug (Session 20251118-223538)
**Problem**: Diagnostic underlining wrong code segment
**8 Models Tested** (7 external + 1 internal Sonnet 4.5)

### Results Summary

| Model | Time | Accuracy | Solution | Value |
|-------|------|----------|----------|-------|
| MiniMax M2 | 3 min | ✅ Exact | Simple fix | ⭐⭐⭐⭐⭐ |
| Sonnet 4.5 (Internal) | 4 min | ✅ Exact | Complete plan | ⭐⭐⭐⭐⭐ |
| Grok Code Fast | 4 min | ✅ Correct | Good validation | ⭐⭐⭐⭐ |
| GPT-5.1 Codex | 5 min | ⚠️ Partial | Complex design | ⭐⭐⭐⭐ |
| Gemini 2.5 Flash | 6 min | ⚠️ Missed | Overanalyzed | ⭐⭐⭐ |
| GLM-4.6 | 7 min | ❌ Wrong focus | Overengineered | ⭐⭐ |
| Sherlock Think | 5 min | ❌ Secondary | Wrong cause | ⭐⭐ |
| Qwen3 Coder | 8+ min | ❌ Failed | Timeout | ⚠️ |

**Key Finding**: **Faster models delivered better results**. Speed correlates with focus on simplicity.

### Cost-Effectiveness Rankings

1. **MiniMax M2** (⭐⭐⭐⭐⭐) - Best value, fastest accurate diagnosis
2. **Sonnet 4.5** (⭐⭐⭐⭐⭐) - Best if using internal (free)
3. **Grok Code Fast** (⭐⭐⭐⭐) - Excellent debugging + validation
4. **Gemini 2.5 Flash** (⭐⭐⭐) - Low cost, good for exploration
5. **GPT-5.1 Codex** (⭐⭐⭐) - High value for architectural work

### What Made Top Models Successful

✅ **Focus on simplicity** - Identified simplest root cause first
✅ **Code-level precision** - Referenced specific files/line numbers
✅ **Practical solutions** - Proposed implementable fixes
✅ **Fast execution** - Completed in 3-5 minutes

### What Held Lower-Ranked Models Back

❌ **Overengineering** - Added unnecessary complexity
❌ **Going too deep** - Explored hypothetical scenarios, missed simple bug
❌ **Secondary issues** - Focused on symptoms, not root cause
❌ **Reliability** - Timeouts and failures

### Recommended Parallel Combinations

**For Maximum Success** (based on empirical data):
- **Bug Investigation**: MiniMax M2 + Grok Code Fast (95%+ success)
- **Architectural**: MiniMax M2 + GPT-5.1 + Grok (99%+ success)
- **Budget**: Gemini 2.5 + Grok (85%+ success, low cost)

**Full Report**: `ai-docs/sessions/20251118-223538/01-planning/comprehensive-model-comparison.md`

---

**Last Updated**: 2025-11-18 | **Validation Session**: 20251118-223538 | **Next Review**: 2025-05

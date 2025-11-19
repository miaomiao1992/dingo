# System Instruction Updates: Parallel External Model Consultation

## Problem Identified

When user requests consultation with multiple external models (via claudish), the main chat assistant should:
1. Use **specialized agents** (not general-purpose) based on the domain
2. Run agents **in parallel** (single message with multiple Task calls)
3. Each agent invokes **one external model** via Bash/claudish
4. All results exchanged via **files** (not returned in agent summaries)

## Current Behavior (Incorrect)

❌ Main chat tries to use general-purpose agents
❌ Main chat tries to invoke claudish directly via Bash
❌ Main chat doesn't recognize this as an agent delegation pattern

## Desired Behavior (Correct)

✅ Main chat identifies "multiple model perspectives needed"
✅ Selects appropriate specialized agent type (golang-architect for Go projects, etc.)
✅ Launches N agents in parallel (one per model)
✅ Each agent uses Bash to invoke claudish
✅ Results saved to session files
✅ Optional: Final consolidation agent synthesizes results

## Recommended Addition to CLAUDE.md

Add this section under "Agent Usage Guidelines":

---

### Pattern: Multiple External Model Consultation

**When to use**: User wants diverse perspectives from multiple LLMs on an architectural decision, design choice, or complex analysis.

**Trigger phrases**:
- "Run multiple models in parallel"
- "Get perspectives from different models"
- "Use external agents to investigate"
- "Consult [model1], [model2], [model3]"

**Execution Pattern**:

```
Step 1: Identify Appropriate Agent Type
- Go project → golang-architect
- Astro/landing page → astro-developer
- General code review → code-reviewer
- Multi-language → general-purpose (last resort)

Step 2: Create Session Folder
SESSION=$(date +%Y%m%d-%H%M%S)
mkdir -p ai-docs/sessions/$SESSION/{input,output}

Step 3: Write Investigation Prompt
Write comprehensive investigation prompt to:
ai-docs/sessions/$SESSION/input/investigation-prompt.md

Step 4: Launch Agents in Parallel (SINGLE MESSAGE)
For each model:
  - Use Task tool with appropriate subagent_type
  - Agent reads investigation-prompt.md
  - Agent invokes: cat prompt.md | claudish --model <model-id>
  - Agent saves to: ai-docs/sessions/$SESSION/output/<model-name>-analysis.md
  - Agent returns MAX 5 sentence summary

Step 5 (Optional): Consolidation Agent
Launch one final agent to:
  - Read all N analysis files
  - Synthesize consensus findings
  - Identify disagreements
  - Provide unified recommendation
  - Write to: ai-docs/sessions/$SESSION/output/CONSOLIDATED.md
```

**Example**:

```
User: "Run gpt-5.1-codex, gemini-2.5-flash, and grok-code-fast-1
       to investigate parser architecture"

Main Chat:
1. Creates session: 20251118-153544/
2. Writes investigation-prompt.md
3. Launches 3 golang-architect agents IN PARALLEL (one message):

   Task(golang-architect): Use gpt-5.1-codex via claudish
   Task(golang-architect): Use gemini-2.5-flash via claudish
   Task(golang-architect): Use grok-code-fast-1 via claudish

4. Each agent:
   - Reads investigation-prompt.md
   - Runs: cat prompt.md | claudish --model <model>
   - Saves to: output/<model>-analysis.md
   - Returns: 5 sentence summary

5. Main chat receives 3 summaries
6. Optionally launches consolidation agent

Result: 3 detailed analyses in files, 3 brief summaries in context
```

**Key Rules**:

1. ✅ **Always use specialized agents** (golang-architect, astro-developer, etc.)
   - ❌ Don't use general-purpose unless no specialized agent exists

2. ✅ **Always launch in parallel** (single message, multiple Task calls)
   - ❌ Don't launch sequentially (slower, wastes time)

3. ✅ **Always use session folders** for file organization
   - input/ for prompts
   - output/ for results

4. ✅ **Each agent = one model** (1:1 mapping)
   - ❌ Don't have one agent call multiple models

5. ✅ **Communication via files**
   - Full analysis → files
   - Brief summary → agent response (MAX 5 sentences)

6. ✅ **Agent uses Bash tool** to invoke claudish
   - ❌ Don't try to invoke claudish from main chat directly

**Available External Models** (via `claudish --list-models`):
- openai/gpt-5.1-codex (software engineering specialist)
- openai/gpt-5 (most advanced reasoning)
- google/gemini-2.5-flash (advanced reasoning + fast)
- x-ai/grok-code-fast-1 (ultra-fast coding)
- qwen/qwen3-vl-235b-a22b-instruct (multimodal)
- openrouter/polaris-alpha (FREE experimental)
- minimax/minimax-m2 (compact high-efficiency)

---

## Additional Recommendations

### 1. Pre-Check Decision Tree Update

Add to the token budget decision tree:

```
[User wants multiple model perspectives?]
         ↓ YES
    [Architectural/design decision?]
         ↓ YES
    Use specialized agents in parallel
    (golang-architect for Go, astro-developer for Astro, etc.)
         ↓
    Each agent invokes one external model via claudish
         ↓
    Results → files
    Summaries → main chat (< 5 sentences each)
         ↓
    Optional: Consolidation agent synthesizes
```

### 2. Agent Self-Awareness Update

Add to "Agent Self-Awareness Rules":

**Rule 6: External Model Invocation**

When a specialized agent (golang-architect, astro-developer, etc.) is asked to invoke an external model:

✅ **Correct approach:**
- Use Bash tool to invoke claudish
- Read input prompt from file
- Save full response to file
- Return brief summary (MAX 5 sentences)

❌ **Incorrect approaches:**
- Delegating to another agent to invoke claudish
- Trying to invoke claudish via Task tool
- Returning full analysis in response (context bloat)

**Example (golang-architect agent)**:
```bash
# Read prompt
cat /path/to/investigation-prompt.md | \
  claudish --model openai/gpt-5.1-codex > \
  /path/to/output/gpt-5.1-codex-analysis.md

# Return brief summary only
```

### 3. Slash Command Consideration

Consider adding a slash command for this pattern:

```bash
# .claude/commands/consult-models.md
Run multiple external models in parallel to investigate: {{QUESTION}}

Models to consult:
- gpt-5.1-codex (software engineering specialist)
- gemini-2.5-flash (advanced reasoning)
- grok-code-fast-1 (fast practical insights)

Session folder: ai-docs/sessions/{{TIMESTAMP}}/
Each model provides independent analysis.
Consolidated recommendation generated at end.
```

Usage: `/consult-models "Should we use regex preprocessor or tree-sitter?"`

## Summary of Changes

**High Priority**:
1. ✅ Add "Pattern: Multiple External Model Consultation" to CLAUDE.md
2. ✅ Update agent selection rules (prefer specialized over general-purpose)
3. ✅ Add parallel execution requirement (single message, multiple agents)

**Medium Priority**:
4. Update pre-check decision tree with external model consultation path
5. Add external model invocation to agent self-awareness rules

**Low Priority**:
6. Create slash command for common pattern (optional convenience)

## Testing the Pattern

After implementing these changes, test with:

```
User: "Run 3 external models (gpt-5.1-codex, gemini-2.5-flash, grok-code-fast-1)
       to investigate whether we should migrate from regex preprocessors
       to tree-sitter for Dingo's parser"

Expected behavior:
1. Main chat creates session folder
2. Main chat writes investigation prompt to file
3. Main chat launches 3 golang-architect agents in ONE message
4. Each agent invokes one model via claudish
5. Each agent saves full response to file
6. Each agent returns 5-sentence summary
7. Main chat optionally launches consolidation agent
8. Total context usage: < 100 lines (not 1000+ lines)
```

## Metrics

**Before (incorrect pattern)**:
- Uses general-purpose agents
- Sequential execution (slow)
- Results returned in agent responses (context bloat)
- Estimated time: 5-10 minutes
- Context usage: 500-1000 lines

**After (correct pattern)**:
- Uses specialized agents (golang-architect)
- Parallel execution (fast)
- Results in files, summaries only in context
- Estimated time: 2-3 minutes (2-3x faster)
- Context usage: 50-100 lines (10x reduction)

## Conclusion

The key insight is: **Use specialized agents to invoke external models in parallel, with file-based communication.**

This pattern:
- Leverages domain expertise (golang-architect for Go questions)
- Maximizes speed (parallel execution)
- Minimizes context (file-based results)
- Produces better outcomes (diverse expert perspectives)

Update CLAUDE.md with the "Pattern: Multiple External Model Consultation" section to codify this workflow.

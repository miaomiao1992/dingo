# Agent Anti-Recursion Rules

## Problem Statement

**Issue**: Specialized agents (astro-developer, golang-developer, etc.) were attempting to delegate tasks to themselves, causing failures.

**Root Cause**: Agent prompts contain instructions like "Use the Task tool with subagent_type=astro-developer" which are correct for external callers but confusing when the agent reads its own prompt.

**Solution**: Add explicit anti-recursion rules to all agent prompts.

---

## Standard Anti-Recursion Section

Every agent prompt MUST include this section immediately after the "Core Responsibilities" or "Operating Modes" section:

```markdown
## ⚠️ CRITICAL: Anti-Recursion Rule

**YOU ARE THE {AGENT_NAME} AGENT**

DO NOT delegate to another {AGENT_NAME} agent. You ARE the specialized agent that does this work directly.

### Self-Awareness Check

Before using the Task tool, ask yourself:
1. **Am I trying to delegate to my own agent type?** → ❌ STOP. Do it yourself.
2. **Do I need a DIFFERENT specialized agent?** → ✅ OK. Use Task tool with different subagent_type.
3. **Am I following proxy mode instructions?** → ⚠️ Those are for EXTERNAL models, not you.

### When You CAN Delegate

✅ **To a DIFFERENT agent type only:**
- You are `golang-developer` → Can delegate to `golang-tester` or `code-reviewer`
- You are `astro-developer` → Can delegate to `astro-reviewer` (different agent)
- You need investigation → Can delegate to `Explore` agent

❌ **NEVER delegate to:**
- Another `{AGENT_NAME}` agent (that's you!)
- The same agent type you are (recursion)

### Proxy Mode Clarification

**What "Proxy Mode" actually means:**
1. Main chat invokes YOU via Task tool
2. You CAN use `claudish` CLI to consult external models (Grok, Gemini, Codex)
3. Those external models provide suggestions, which you implement
4. Those external models may ALSO invoke you via Task tool
5. **But YOU do not invoke yourself - you ARE the endpoint**

**Correct proxy workflow:**
```
Main Chat
  ↓ [Task tool: golang-developer]
You (golang-developer agent)
  ↓ [claudish: Ask Gemini for algorithm suggestion]
External Model (Gemini)
  → Returns: "Use X algorithm because Y"
You (golang-developer agent)
  → Implements the algorithm directly
```

**WRONG proxy workflow (recursion):**
```
Main Chat
  ↓ [Task tool: golang-developer]
You (golang-developer agent)
  ↓ [Task tool: golang-developer]  ← ❌ WRONG!
Another golang-developer agent
  → ERROR: Recursion detected
```

### Instructions in This Prompt

When you see instructions like:
- "Use the Task tool with subagent_type={AGENT_NAME}"
- "Invoke the {AGENT_NAME} agent"
- "Delegate to {AGENT_NAME} for implementation"

**These are instructions FOR OTHERS to use when calling you, NOT for you to call yourself.**

You are the destination, not the caller.

### Quick Decision Tree

```
Need to use Task tool?
│
├─ Am I {AGENT_NAME}?
│  └─ YES
│     └─ Is the task for {AGENT_NAME}?
│        ├─ YES → ❌ DO NOT delegate. Implement directly.
│        └─ NO → ✅ Can delegate to different agent
│
└─ Unsure which agent I am?
   └─ You are: {AGENT_NAME}
      └─ Never delegate to {AGENT_NAME}
```

### If You Catch Yourself About to Delegate to Yourself

**STOP. Ask:**
1. Why do I think I need to delegate?
2. Am I trying to save context? (Don't - just do the work)
3. Am I following instructions meant for callers? (Yes - ignore those)
4. Can I actually just implement this myself? (Yes - you're the expert)

**Then:** Implement directly. You are the specialized agent for this work.
```

---

## Agent-Specific Replacements

### For golang-developer.md

Replace `{AGENT_NAME}` with:
- `golang-developer`

Add this section **after line 58** ("## Core Competencies" section ends).

### For astro-developer.md

Replace `{AGENT_NAME}` with:
- `astro-developer`

Add this section **after line 59** ("# Core Responsibilities" section ends).

### For golang-architect.md

Replace `{AGENT_NAME}` with:
- `golang-architect`

Add this section at the appropriate location after core responsibilities.

### For tester.md (golang-tester)

Replace `{AGENT_NAME}` with:
- `golang-tester` or `tester`

Add this section at the appropriate location.

### For code-reviewer.md

Replace `{AGENT_NAME}` with:
- `code-reviewer`

Add this section at the appropriate location.

### For astro-reviewer.md

Replace `{AGENT_NAME}` with:
- `astro-reviewer`

Add this section at the appropriate location.

---

## CLAUDE.md Updates

Add the following section to both:
- `/Users/jack/mag/dingo/CLAUDE.md`
- `/Users/jack/mag/dingo/langingpage/CLAUDE.md`

### New Section: "Agent Self-Awareness Rules"

Insert this section in the "🎯 Delegation Strategy & Context Economy" chapter, after the "Main Chat → Agent (Delegation)" subsection:

```markdown
### Agent Self-Awareness Rules (Anti-Recursion)

**CRITICAL FOR ALL AGENTS:**

#### Rule 1: Know Thyself

Every agent MUST be aware of its own type:
- If you are `golang-developer`, you cannot delegate to `golang-developer`
- If you are `astro-developer`, you cannot delegate to `astro-developer`
- If you are `code-reviewer`, you cannot delegate to `code-reviewer`

**Why:** You ARE the specialized agent. Delegating to yourself causes recursion and failures.

#### Rule 2: Delegation Decision Tree

```
Before using Task tool, ask:
│
├─ What is my agent type?
│  └─ I am: [agent-name]
│
├─ What agent type does this task need?
│  ├─ Same as me → ❌ DO NOT delegate. Implement directly.
│  └─ Different → ✅ CAN delegate to that different agent
│
└─ Why do I want to delegate?
   ├─ "To save context" → ❌ WRONG REASON. Just do the work.
   ├─ "Instructions say to" → ❌ Those are for CALLERS, not you.
   └─ "Need different expertise" → ✅ OK if it's a different agent type.
```

#### Rule 3: Proxy Mode Is Not Self-Delegation

**Proxy Mode Means:**
- Using `claudish` to consult external models (Grok, Gemini, Codex)
- Getting suggestions/implementations from those models
- Implementing their suggestions yourself

**Proxy Mode Does NOT Mean:**
- Using Task tool to invoke yourself
- Creating another instance of your own agent
- Delegating work you should do directly

#### Rule 4: Instructions Are Context-Dependent

When you read your own prompt and see:
- "Use the Task tool to invoke the golang-developer agent"
- "Delegate to astro-developer for implementation"

**Understand:** These instructions are for MAIN CHAT and EXTERNAL MODELS to use when calling you.

**NOT** for you to use to call yourself.

You are the **destination** of those calls, not the **source**.

#### Rule 5: When In Doubt, Implement

If you're uncertain whether to delegate:
1. Check if you're delegating to your own agent type → If yes, DON'T
2. Check if you have the expertise to implement → If yes, DO IT
3. Check if you're trying to save context → NOT A VALID REASON

**Default action: Implement directly.**

#### Examples

**✅ CORRECT Delegation:**
- `golang-developer` delegates to `golang-tester` (different agent)
- `astro-developer` delegates to `astro-reviewer` (different agent)
- `golang-developer` delegates to `Explore` (different agent)
- `golang-architect` delegates to `golang-developer` (different agent)

**❌ WRONG Delegation (Recursion):**
- `golang-developer` delegates to `golang-developer` (same agent ❌)
- `astro-developer` delegates to `astro-developer` (same agent ❌)
- `code-reviewer` delegates to `code-reviewer` (same agent ❌)
```

---

## Implementation Checklist

- [ ] Update `/Users/jack/mag/dingo/.claude/agents/golang-developer.md`
- [ ] Update `/Users/jack/mag/dingo/.claude/agents/golang-architect.md`
- [ ] Update `/Users/jack/mag/dingo/.claude/agents/tester.md`
- [ ] Update `/Users/jack/mag/dingo/.claude/agents/code-reviewer.md`
- [ ] Update `/Users/jack/mag/dingo/langingpage/.claude/agents/astro-developer.md`
- [ ] Update `/Users/jack/mag/dingo/langingpage/.claude/agents/astro-reviewer.md`
- [ ] Update `/Users/jack/mag/dingo/CLAUDE.md` (add Agent Self-Awareness section)
- [ ] Update `/Users/jack/mag/dingo/langingpage/CLAUDE.md` (add Agent Self-Awareness section)
- [ ] Review slash commands for similar delegation issues

---

**Created**: 2025-11-18
**Purpose**: Prevent agents from recursively delegating to themselves
**Status**: Ready for implementation

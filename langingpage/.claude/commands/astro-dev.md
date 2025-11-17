# Astro Development Orchestration Workflow

You are now executing the **Astro Development Orchestration** workflow - a comprehensive, multi-phase development process that ensures high-quality Astro implementations through planning, development, review, and iterative refinement.

## Workflow Overview

This orchestration coordinates multiple agents through the following phases:

1. **Planning Phase** - Understand requirements and plan implementation
2. **Development Phase** - Execute the implementation
3. **Review Selection Phase** - Choose review strategy (internal vs multi-LLM)
4. **Review Phase** - Run reviews in parallel
5. **Fix Loop Phase** - Iteratively fix issues until all reviewers approve

---

## Phase 1: Planning & Requirements Gathering

### Step 1.1: Launch Planning Agent

Use the Task tool to launch the `astro-developer` agent in **planning mode**:

**Agent Task**:
```
You are in PLANNING MODE for the Astro development orchestration workflow.

Task: Understand the user's requirements and plan the implementation.

Required Actions:
1. **Understand Requirements**:
   - Ask clarifying questions about what needs to be built/modified
   - Identify scope: new feature, bug fix, refactor, optimization, etc.
   - Determine affected components/pages/layouts

2. **Analyze Current State**:
   - Read relevant files in src/ directory
   - Review existing components and patterns
   - Check current styling approach (Tailwind, scoped styles, etc.)
   - Identify dependencies and integrations

3. **Consult Knowledge Base** (MANDATORY):
   - Read ai-docs/INDEX.md for navigation
   - Read ai-docs/01-why-astro.md for core principles
   - Read ai-docs/best-practices-checklist.md for decision trees
   - Read relevant modules based on task:
     * Components: ai-docs/07-astro-components.md, ai-docs/06-framework-components.md
     * Layouts: ai-docs/08-layouts.md
     * Content: ai-docs/05-content-collections.md
     * Islands: ai-docs/02-islands-architecture.md
     * Recipes: ai-docs/recipes/* (as needed)

4. **Create Implementation Plan**:
   - Break down into specific, actionable steps
   - Reference ai-docs modules for each step
   - Identify decision points (e.g., .astro vs framework component)
   - Estimate complexity and potential issues
   - List files to create/modify
   - Note best practices to follow

5. **Present Plan**:
   - Use clear markdown formatting
   - Include ai-docs references
   - Highlight critical decisions
   - Note any assumptions or ambiguities
   - Provide architecture diagram if complex

Output Format:
# Implementation Plan

## Requirements Summary
[Clear statement of what needs to be built]

## Current State Analysis
[What exists, what patterns are in use, what will be affected]

## Knowledge Base Consultation
**ai-docs Modules Consulted**:
- [List modules with key takeaways]

## Implementation Steps
1. [Step with ai-docs reference]
2. [Step with ai-docs reference]
...

## Key Decisions
- [Decision point with rationale and ai-docs reference]

## Files to Create/Modify
- [File path: description]

## Potential Issues
- [Issue with mitigation strategy]

## Estimated Complexity
[Simple/Medium/Complex with reasoning]
```

**Wait for the planning agent to complete and return the plan.**

### Step 1.2: Review Plan with User

Once you receive the plan from the planning agent:

1. **Present the plan** to the user clearly
2. **Ask for feedback** using this pattern:
   - "Here's the implementation plan from the astro-developer agent."
   - "Please review the plan above. Are you ready to proceed, or would you like any changes?"

3. **Collect feedback**:
   - If user requests changes: Go back to Step 1.1 with updated requirements
   - If user approves: Proceed to Phase 2

---

## Phase 2: Development Execution

### Step 2.1: Launch Development Agent

Once the plan is approved, use the Task tool to launch the `astro-developer` agent in **development mode**:

**Agent Task**:
```
You are in DEVELOPMENT MODE for the Astro development orchestration workflow.

You have an approved implementation plan. Now execute it.

Implementation Plan:
[Insert the approved plan from Phase 1]

Required Actions:
1. **Follow the Plan Exactly**:
   - Execute each step from the approved plan
   - Reference ai-docs modules as specified
   - Make decisions as outlined in the plan

2. **Write Code to Files**:
   - Create/modify all files as specified
   - Follow Astro best practices from ai-docs
   - Use proper component patterns (.astro vs framework)
   - Apply correct client:* directives for Islands
   - Optimize images, use Content Collections, etc.

3. **Self-Validate**:
   - Check against ai-docs/best-practices-checklist.md
   - Ensure no CRITICAL issues
   - Verify proper architecture (server-first, zero JS default)

4. **Document Changes**:
   - List all files created/modified
   - Note any deviations from plan with reasoning
   - Highlight any discovered issues or edge cases

Output Format:
# Implementation Complete

## Files Created/Modified
- [File path: what was done]

## Changes Summary
[Concise summary of implementation]

## ai-docs Validation
[Which checklist items passed, any concerns]

## Notes
[Any deviations, issues discovered, or recommendations]
```

**Wait for the development agent to complete the implementation.**

### Step 2.2: Confirm Development Complete

Once development is done:
- Acknowledge completion to the user
- Summarize what was implemented
- Proceed to Phase 3

---

## Phase 3: Review Strategy Selection

### Step 3.1: Fetch Available LLM Models

Run the following command to get available models from claudish:

```bash
claudish --list-models
```

Parse the output to extract model identifiers (e.g., `x-ai/grok-code-fast-1`, `openai/gpt-4-turbo`, `anthropic/claude-opus-4`, etc.)

### Step 3.2: Present Review Options to User

Use the **AskUserQuestion** tool to present review strategy options:

**Question Structure**:
- **Header**: "Review strategy"
- **Question**: "How would you like to review the implementation? (Select one or more review methods)"
- **Multi-select**: `true`
- **Options**:
  1. **Internal Review Only**
     - Label: "Internal only (astro-reviewer)"
     - Description: "Fast review using the built-in astro-reviewer agent. Best for quick iterations."

  2. **+ Grok Fast** (if available in claudish models)
     - Label: "+ Grok Fast (x-ai/grok-code-fast-1)"
     - Description: "Add fast external review via Grok. Good for catching additional issues."

  3. **+ Grok Quality** (if available)
     - Label: "+ Grok Quality (x-ai/grok-code-quality-1)"
     - Description: "Add thorough external review via Grok. Best for production code."

  4. **+ GPT-4 Turbo** (if available)
     - Label: "+ GPT-4 Turbo (openai/gpt-4-turbo)"
     - Description: "Add OpenAI GPT-4 review. Excellent for best practices and edge cases."

  5. **+ Claude Opus** (if available)
     - Label: "+ Claude Opus (anthropic/claude-opus-4)"
     - Description: "Add Claude Opus review. Best for comprehensive analysis and architecture."

  6. **+ Gemini** (if available)
     - Label: "+ Gemini (google/gemini-2.0-pro)"
     - Description: "Add Google Gemini review. Strong at performance and optimization."

  7. **Custom model**
     - Label: "Custom model (manual entry)"
     - Description: "Specify a custom model identifier from claudish --list-models"

**Note**: Dynamically build this list based on actual models from `claudish --list-models`. Always include "Internal only" as first option and "Custom model" as last option.

### Step 3.3: Process User Selection

After user selects review methods:

1. **Parse selections**:
   - Extract selected model identifiers
   - If "Custom model" selected, prompt for model identifier
   - Validate model identifiers

2. **Build reviewer list**:
   - Always include internal `astro-reviewer` agent
   - Add external reviewers for each selected model

3. **Store configuration** for Phase 4 and Phase 5 (fix loop)

Example configuration:
```javascript
{
  "internal": true,
  "external": [
    "x-ai/grok-code-fast-1",
    "openai/gpt-4-turbo"
  ]
}
```

---

## Phase 4: Parallel Review Execution

### Step 4.1: Prepare Review Context

Create a review context file that all reviewers will use:

**File**: `review-reports/review-context-[timestamp].md`

**Content**:
```markdown
# Review Context - [Timestamp]

## Implementation Summary
[Summary from Phase 2]

## Files Changed
[List of files from Phase 2]

## Review Instructions
Please review this Astro implementation against:
1. ai-docs/ best practices (mandatory reading)
2. Project CLAUDE.md guidelines
3. Performance targets (Core Web Vitals)
4. Code quality and maintainability
5. Visual accuracy (if applicable)

## Knowledge Base
- Location: langingpage/ai-docs/
- Start with: INDEX.md
- Key modules: best-practices-checklist.md, 01-why-astro.md

## Review Report Format
Use the standard Astro Code Review Report format with:
- Issues categorized by severity (CRITICAL/MEDIUM/MINOR)
- ai-docs module references for violations
- Specific fixes with file:line locations
- Performance impact estimates
```

### Step 4.2: Launch Reviews in Parallel

**IMPORTANT**: Launch ALL reviewers in parallel using a single message with multiple Task tool calls.

#### Internal Review

**Task tool call 1** - Launch `astro-reviewer` agent:

```
You are conducting an internal code review as part of the Astro development orchestration workflow.

Review Context File: review-reports/review-context-[timestamp].md

Task: Perform a comprehensive review of the implementation.

Required Actions:
1. Read the review context file
2. Read ai-docs/ knowledge base (INDEX.md, best-practices-checklist.md, relevant modules)
3. Review all modified/created files
4. Validate against best practices checklist
5. Run the dev server and perform visual validation (if applicable)
6. Check performance metrics

Output:
- Write detailed review report to: review-reports/internal-review-[timestamp].md
- Return a concise summary (3-5 bullet points) highlighting CRITICAL and MEDIUM issues only

Format:
# Internal Review Summary
**Status**: PASS / NEEDS_FIXES / MAJOR_ISSUES
**Critical Issues**: [count]
**Medium Issues**: [count]

## Top Issues
1. [Most important issue]
2. [Second most important]
...

Full report: review-reports/internal-review-[timestamp].md
```

#### External Reviews (if any)

For each external model selected, **Task tool call N** - Launch `astro-reviewer` in proxy mode:

```
You are conducting an external code review via claudish proxy as part of the Astro development orchestration workflow.

Model: [model-identifier]
Review Context File: review-reports/review-context-[timestamp].md

Task: Delegate this review to the specified model via claudish.

Claudish Command Pattern:
cat review-reports/review-context-[timestamp].md | claudish --stdin --model [model-identifier] "$(cat <<'PROMPT'
You are using the astro-reviewer agent in proxy mode.

Review Context: [stdin contains the context]

Task: Perform a comprehensive code review of this Astro implementation.

Instructions:
1. Read the review context provided via stdin
2. Consult the knowledge base at langingpage/ai-docs/ (start with INDEX.md)
3. Review all files mentioned in the context
4. Validate against Astro best practices from ai-docs/best-practices-checklist.md
5. Check for:
   - CRITICAL issues (violates core principles)
   - MEDIUM issues (suboptimal patterns)
   - MINOR issues (nice-to-haves)

Output Requirements:
1. Write detailed review to: review-reports/[model-name]-review-[timestamp].md
2. Use standard Astro Code Review Report format
3. Reference ai-docs modules for every issue
4. Provide specific fixes with file:line locations
5. Return concise summary (3-5 bullet points)

Format:
# [Model Name] Review Summary
**Status**: PASS / NEEDS_FIXES / MAJOR_ISSUES
**Critical Issues**: [count]
**Medium Issues**: [count]

## Top Issues
1. [Most important issue]
2. [Second most important]
...

Full report: review-reports/[model-name]-review-[timestamp].md
PROMPT
)"

Required Actions:
1. Execute the claudish command above
2. Capture the output
3. Parse the summary
4. Verify the review report file was created

Return: The concise summary from the external reviewer
```

**Launch all review tasks in a SINGLE message with multiple Task tool calls.**

### Step 4.3: Collect Review Results

Wait for all parallel reviews to complete. You will receive:
- Internal review summary
- External review summaries (one per model)

**Aggregate Results**:
1. Count total CRITICAL issues across all reviews
2. Count total MEDIUM issues across all reviews
3. Identify common issues mentioned by multiple reviewers
4. Note unique issues found by specific reviewers

**Present Summary to User**:
```markdown
# Review Results - [Timestamp]

## Reviews Completed
- ✅ Internal (astro-reviewer)
- ✅ External: [model1]
- ✅ External: [model2]
...

## Aggregate Issues
- **CRITICAL**: [total count]
- **MEDIUM**: [total count]
- **MINOR**: [total count]

## Common Issues (mentioned by 2+ reviewers)
1. [Issue]
2. [Issue]

## Detailed Reports
- Internal: review-reports/internal-review-[timestamp].md
- [Model1]: review-reports/[model1]-review-[timestamp].md
- [Model2]: review-reports/[model2]-review-[timestamp].md

## Status
[PASS / NEEDS_FIXES / MAJOR_ISSUES]
```

---

## Phase 5: Fix Loop (Iterative Refinement)

### Step 5.1: Determine if Fixes Needed

Check the aggregate review status:

- **If Status = PASS**:
  - Congratulate the user
  - Workflow complete ✅
  - Exit

- **If Status = NEEDS_FIXES or MAJOR_ISSUES**:
  - Proceed to Step 5.2

### Step 5.2: Consolidate Review Feedback

Use the Task tool to launch `astro-developer` in **review consolidation mode**:

```
You are in REVIEW CONSOLIDATION MODE for the Astro development orchestration workflow.

Task: Analyze all review reports and create a prioritized fix list.

Review Reports:
- review-reports/internal-review-[timestamp].md
- review-reports/[model1]-review-[timestamp].md
- review-reports/[model2]-review-[timestamp].md
...

Required Actions:
1. **Read All Review Reports**:
   - Parse each review report
   - Extract all CRITICAL and MEDIUM issues
   - Note file:line locations for each issue

2. **Deduplicate Issues**:
   - Identify issues mentioned by multiple reviewers
   - Consolidate into single list with all references

3. **Prioritize Fixes**:
   - CRITICAL issues first (by impact)
   - MEDIUM issues second (by frequency across reviewers)
   - Group by file for efficient fixing

4. **Create Fix Plan**:
   - For each issue, specify:
     * File and line location
     * Current problematic code
     * Proposed fix (with ai-docs reference)
     * Expected impact
   - Estimate total effort

Output Format:
# Consolidated Review Feedback

## Summary
- **Total Issues**: [count]
- **Critical**: [count]
- **Medium**: [count]
- **Files Affected**: [count]

## Prioritized Fix List

### CRITICAL Issues
1. **[Issue Title]**
   - **Mentioned by**: [reviewer1, reviewer2]
   - **Location**: file.astro:123
   - **Current Code**: ```[snippet]```
   - **Fix**: ```[proposed code]```
   - **ai-docs Reference**: [module]
   - **Impact**: [description]

2. [Next critical issue]
...

### MEDIUM Issues
[Same format]

## Fix Strategy
[High-level approach to addressing all issues efficiently]

## Estimated Effort
[Simple/Medium/Complex]
```

Wait for consolidation agent to complete.

### Step 5.3: Execute Fixes

Use the Task tool to launch `astro-developer` in **fix mode**:

```
You are in FIX MODE for the Astro development orchestration workflow.

Task: Apply all fixes from the consolidated review feedback.

Consolidated Feedback: [output from Step 5.2]

Required Actions:
1. **Apply Fixes Systematically**:
   - Address all CRITICAL issues first
   - Then address all MEDIUM issues
   - Follow the proposed fixes from consolidation
   - Reference ai-docs modules as specified

2. **Validate Each Fix**:
   - Ensure fix actually resolves the issue
   - Check against best-practices-checklist.md
   - Verify no new issues introduced

3. **Write to Files**:
   - Modify all affected files
   - Use proper patterns from ai-docs
   - Maintain code quality

4. **Document Changes**:
   - List all fixes applied
   - Note file:line changes
   - Highlight any issues that couldn't be fixed (with reasoning)

Output Format:
# Fixes Applied

## Summary
- **Fixes Applied**: [count]
- **Critical Fixed**: [count]
- **Medium Fixed**: [count]
- **Files Modified**: [list]

## Detailed Changes
1. **[Issue Title]** - FIXED ✅
   - File: file.astro:123
   - Change: [description]
   - ai-docs Reference: [module]

2. [Next fix]
...

## Unable to Fix (if any)
- **[Issue]**: [Reason why couldn't fix, needs manual intervention]

## Validation
[Self-validation against checklist]
```

Wait for fixes to be applied.

### Step 5.4: Re-run Reviews (Same Reviewers)

**IMPORTANT**: Re-run the EXACT same set of reviewers that were used in Phase 4.

Use the stored configuration from Phase 3.3 to determine which reviewers to launch.

**Launch all reviewers in parallel** (same pattern as Phase 4.2):
- Internal: `astro-reviewer` agent
- External: `astro-reviewer` in proxy mode for each selected model

**Use the same review instructions** but update the review context to indicate this is iteration N:

```markdown
# Review Context - Iteration [N] - [Timestamp]

## Previous Review
- Iteration [N-1] found [X] CRITICAL and [Y] MEDIUM issues
- All issues have been addressed by astro-developer

## Fixes Applied
[Summary from Step 5.3]

## Review Instructions
Re-validate the implementation:
1. Verify all previously identified issues are resolved
2. Check for new issues introduced by fixes
3. Validate against ai-docs/ best practices
4. Confirm Core Web Vitals targets met
...
```

**Wait for all reviews to complete.**

### Step 5.5: Check Loop Termination

After re-reviews complete:

1. **Aggregate new results** (same as Phase 4.3)

2. **Check status**:
   - **If Status = PASS**:
     - Workflow complete ✅
     - Present final summary to user
     - Exit

   - **If Status = NEEDS_FIXES or MAJOR_ISSUES**:
     - **If iteration count < 5**:
       - Go back to Step 5.2 (consolidate new feedback)
       - Continue fix loop
     - **If iteration count >= 5**:
       - Alert user: "Maximum iterations reached. Manual intervention may be needed."
       - Present current status
       - Ask user: Continue loop? Or manual review?

### Step 5.6: Final Summary (When Loop Completes)

When all reviewers approve (Status = PASS):

```markdown
# 🎉 Astro Development Workflow Complete!

## Summary
- **Planning**: [brief summary of what was planned]
- **Development**: [what was implemented]
- **Review Iterations**: [N]
- **Total Issues Fixed**: [count]
  - Critical: [count]
  - Medium: [count]

## Final Review Status
All reviewers approved! ✅

### Review Results
- Internal (astro-reviewer): ✅ PASS
- External ([model1]): ✅ PASS
- External ([model2]): ✅ PASS

## Files Modified
[Final list of all files created/modified]

## All Review Reports
[Links to all review reports from all iterations]

## Performance Validation
- Bundle size: [size]
- JavaScript shipped: [size]
- Core Web Vitals: [status]

## Next Steps
- Test the implementation manually
- Deploy to staging environment
- Monitor performance in production
```

---

## Error Handling

### If Planning Agent Fails
- Notify user
- Offer to retry with simplified requirements
- Suggest manual planning

### If Development Agent Fails
- Notify user with error details
- Check if issue is resolvable
- Offer to retry or fall back to manual development

### If Review Agent Fails
- Log the failure
- Continue with remaining reviewers
- Notify user which review failed
- Offer to retry failed review

### If Claudish Unavailable
- Notify user
- Fall back to internal review only
- Suggest installing claudish

### If Fix Loop Stalls
- After 5 iterations without PASS, alert user
- Provide summary of persistent issues
- Suggest manual intervention
- Offer to export consolidated issue list

---

## Usage Notes

- **Run this command**: `/astro-dev` from the landing page directory
- **Preparation**: Ensure `pnpm dev` can run successfully
- **Time**: Full workflow may take 10-30 minutes depending on complexity
- **Cost**: External reviews via claudish may incur API costs
- **Interactivity**: User will be asked questions at key decision points

---

## Success Criteria

The workflow is successful when:
- ✅ All CRITICAL issues resolved
- ✅ All MEDIUM issues resolved (or explicitly accepted)
- ✅ All selected reviewers approve (Status = PASS)
- ✅ Code follows ai-docs/ best practices
- ✅ Performance targets met
- ✅ Visual implementation matches requirements

---

**Now execute this workflow step by step. Start with Phase 1: Planning & Requirements Gathering.**

# Slash Commands - Dingo Landing Page

This directory contains custom slash commands for the Dingo landing page project.

## Available Commands

### `/astro-dev` - Astro Development Orchestration

**Purpose**: A comprehensive, multi-phase development workflow that orchestrates planning, development, review, and iterative refinement for Astro implementations.

**Usage**:
```
/astro-dev
```

**Workflow**:
1. **Planning Phase**: Launches `astro-developer` agent to analyze requirements, consult ai-docs/, and create implementation plan
2. **Development Phase**: Executes the approved plan using `astro-developer` agent
3. **Review Selection**: Asks user to choose review strategy (internal only or multi-LLM external reviews)
4. **Parallel Reviews**: Runs all selected reviewers in parallel (internal + external via claudish)
5. **Fix Loop**: Iteratively consolidates feedback, applies fixes, and re-reviews until all reviewers approve

**Key Features**:
- 📋 **AI-docs driven planning** - Every decision references best practices from ai-docs/
- 🔄 **Iterative refinement** - Automatically fixes issues until reviewers approve
- 🚀 **Parallel reviews** - Run multiple LLM reviews simultaneously for comprehensive feedback
- 📊 **Multi-LLM support** - Choose from Grok, GPT-4, Claude Opus, Gemini, or custom models
- 🎯 **Zero manual review** - Fully automated from planning to approved implementation

**Review Options**:
- **Internal only**: Fast iteration using built-in `astro-reviewer` agent
- **Multi-LLM**: Add external reviews via claudish:
  - x-ai/grok-code-fast-1 (fast)
  - x-ai/grok-code-quality-1 (thorough)
  - openai/gpt-4-turbo (best practices)
  - anthropic/claude-opus-4 (comprehensive)
  - google/gemini-2.0-pro (performance)
  - Custom models (manual entry)

**Prerequisites**:
- Working Astro development environment (`pnpm dev` functional)
- `claudish` CLI installed (for external reviews): https://github.com/anthropics/claudish
- Agents configured: `astro-developer`, `astro-reviewer`

**Time Estimate**:
- Simple tasks: 5-10 minutes
- Medium tasks: 10-20 minutes
- Complex tasks: 20-40 minutes
- External reviews add: 2-5 minutes per model per iteration

**Cost Considerations**:
- Internal review only: No additional cost (uses Claude Code)
- External reviews: API costs via claudish (varies by model and usage)
- Typical external review: $0.01 - $0.10 per review depending on code size and model

**Example Use Cases**:
- "Create a new Hero section component with responsive design"
- "Add a contact form with client-side validation"
- "Refactor the Features section to use Islands Architecture"
- "Optimize images and improve Core Web Vitals"
- "Implement a blog with Content Collections"

**Output**:
- All code changes written to files in src/
- Review reports saved to review-reports/ directory
- Final summary with links to all reports and metrics

**Safety**:
- Maximum 5 fix loop iterations (prevents infinite loops)
- All changes written to files (version control friendly)
- Each iteration documented in review reports
- User can interrupt workflow at decision points

**Tips**:
- Start with internal review only for rapid iteration
- Use multi-LLM reviews for production-critical code
- Check review-reports/ directory for detailed feedback
- Review reports persist across sessions (timestamped)

---

## How Slash Commands Work

Slash commands are markdown files that expand into detailed prompts when you type them in Claude Code.

**Creating a new command**:
1. Create a `.md` file in `.claude/commands/`
2. The filename becomes the command (e.g., `foo.md` → `/foo`)
3. Write the prompt instructions in the markdown file
4. Use it by typing `/command-name` in Claude Code

**Best Practices**:
- Use clear, step-by-step instructions
- Reference specific tools (Task, AskUserQuestion, etc.)
- Include error handling
- Provide examples and usage notes
- Document prerequisites and outputs

---

Last updated: 2025-11-17

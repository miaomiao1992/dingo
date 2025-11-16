# AI Documentation Directory

**Purpose**: This directory contains AI-generated research, design documents, working notes, and session artifacts for AI agents working on the Dingo project.

**Audience**: AI agents (Claude, GPT, Gemini, etc.) and human developers reviewing AI-generated content

**NOT for**: End-user documentation (see `/docs` when created)

---

## Directory Structure

```
ai-docs/
├── README.md              ← You are here
├── research/              ← Multi-source research documents
│   ├── compiler/          ← Compiler/transpiler architecture research
│   └── golang_missing/    ← Research on Go's missing features
├── language/              ← Language design decisions and syntax specs
├── wip/                   ← Work-in-progress documents and planning
└── sessions/              ← /dev orchestrator session artifacts
    └── YYYYMMDD-HHMMSS/   ← Timestamped development sessions
```

---

## 📁 Directory Purposes

### `research/`
**Multi-source AI research on specific topics**

Research gathered from multiple AI models to get diverse perspectives and comprehensive coverage.

#### `research/compiler/`
Research on compiler/transpiler architecture:
- `gemini_research.md` - Gemini's technical blueprint
- `claude-research.md` - Claude's implementation guide
- `chatgpt-research.md` - ChatGPT's analysis

**Purpose**: Architectural guidance, tooling recommendations, implementation strategies

#### `research/golang_missing/`
Research on Go language limitations that Dingo aims to address:
- Multiple AI perspectives (gemini, claude, chatgpt, kimi, grok)
- Identifies pain points in Go
- Proposes solutions for Dingo

**Purpose**: Feature justification, language design rationale

### `language/`
**Language design decisions and specifications**

Documents defining Dingo's syntax, semantics, and design philosophy.

**Contents**:
- `SYNTAX_DESIGN.md` - Syntax decisions with rationale
- `UI_IMPLEMENTATION.md` - UX considerations for the language

**Purpose**:
- Reference for consistent language design
- Justification for syntax choices
- Guide for parser/transpiler implementation

### `wip/`
**Work-in-progress documents and planning**

Active development documents, progress tracking, and evolving designs.

**Contents**:
- `PROGRESS.md` - Weekly/milestone progress reports
- `IMPLEMENTATION_PLAN.md` - Current implementation roadmap
- `BORGO_ANALYSIS.md` - Analysis of Borgo transpiler (reference project)
- `GO_IMPLEMENTATION.md` - Go-specific implementation notes
- `PLUGIN_SYSTEM_DESIGN.md` - Plugin architecture design
- `CLI_SHOWCASE.md` - CLI features and examples

**Purpose**:
- Track development progress
- Document implementation decisions
- Analyze reference projects

**Note**: Documents in `wip/` may become outdated as the project evolves. Check dates and validate against current codebase.

### `sessions/`
**/dev orchestrator session artifacts**

Created by the `/dev` slash command orchestrator. Each session is a complete development workflow with planning, implementation, reviews, and testing.

**Structure**:
```
sessions/YYYYMMDD-HHMMSS/
├── session-state.json         ← Workflow state tracker
├── 01-planning/
│   ├── user-request.md        ← Original user request
│   ├── initial-plan.md        ← Architect's initial plan
│   ├── gaps.json              ← Clarification questions
│   ├── clarifications.md      ← User's answers
│   ├── final-plan.md          ← Finalized implementation plan
│   └── summary.txt            ← Brief plan summary
├── 02-implementation/
│   ├── changes-made.md        ← Files created/modified
│   ├── implementation-notes.md ← Decisions and deviations
│   └── status.txt             ← Implementation status
├── 03-reviews/
│   └── iteration-01/
│       ├── internal-review.md    ← Code-reviewer agent feedback
│       ├── gemini-review.md      ← Gemini's code review
│       ├── gpt4-review.md        ← GPT-4's code review
│       ├── consolidated.md       ← Combined feedback
│       ├── action-items.md       ← Prioritized fixes
│       └── consolidated-summary.txt ← Review summary
├── 04-testing/
│   ├── test-plan.md           ← Test scenarios
│   ├── test-results.md        ← Test execution results
│   └── test-summary.txt       ← Pass/fail status
└── session-logs/
    └── errors.log             ← Error log
```

**Purpose**:
- Complete audit trail for development workflows
- Multi-agent collaboration artifacts
- Reference for similar future implementations

**Lifecycle**: Sessions are preserved indefinitely as historical record.

---

## Organization Rules

### From `/CLAUDE.md`

#### ✅ DO

1. **Save all AI research here**
   - Multi-model perspectives
   - Architecture analysis
   - Design explorations

2. **Document design decisions**
   - Language syntax choices
   - Implementation strategies
   - Architecture rationale

3. **Track work-in-progress**
   - Progress reports
   - Implementation plans
   - Evolving designs

4. **Preserve session artifacts**
   - Development workflows
   - Code review iterations
   - Test results

5. **Use descriptive filenames**
   - `{topic}-{source}.md` (e.g., `compiler-gemini.md`)
   - `{FEATURE}_DESIGN.md` (e.g., `PLUGIN_SYSTEM_DESIGN.md`)
   - All caps for major design docs in `wip/`

6. **Update `/CLAUDE.md` for important context**
   - When adding critical research
   - When making architectural decisions
   - When changing project direction

#### ❌ DON'T

1. **Don't put user-facing docs here**
   - User docs → `/docs` (when created)
   - API docs → `/docs/api`
   - Tutorials → `/docs/tutorials`

2. **Don't create docs in root**
   - Root should only have `README.md` and `CLAUDE.md`
   - Keep root minimal

3. **Don't delete session directories**
   - Sessions are permanent audit trail
   - Can reference for future similar tasks

4. **Don't mix concerns**
   - Research → `research/`
   - Language design → `language/`
   - Work-in-progress → `wip/`
   - Session artifacts → `sessions/`

---

## File Naming Conventions

### Research Documents
- **Multi-source**: `{model-name}-research.md` or `{model-name}.md`
- **Topic-specific**: `{topic}-analysis.md`, `{topic}-proposal.md`

### Language Design
- **Uppercase for visibility**: `SYNTAX_DESIGN.md`, `TYPE_SYSTEM.md`
- **Descriptive**: Include the aspect being designed

### Work-in-Progress
- **Uppercase for active docs**: `PROGRESS.md`, `IMPLEMENTATION_PLAN.md`
- **Include context**: `BORGO_ANALYSIS.md` (what + what it is)

### Sessions
- **Timestamped directories**: `YYYYMMDD-HHMMSS/` (auto-generated)
- **Structured filenames**: Follow the session template structure

---

## Usage Guide for AI Agents

### When Starting Work on Dingo

1. **Read key context files** (in order):
   - `/CLAUDE.md` - Project overview and current status
   - `research/compiler/gemini_research.md` - Technical blueprint
   - `research/compiler/claude-research.md` - Implementation guide
   - `language/SYNTAX_DESIGN.md` - Current syntax decisions
   - `wip/PROGRESS.md` - Latest progress

2. **Understand the phase**:
   - Check `Current Phase` in `/CLAUDE.md`
   - Read relevant `wip/` documents for current work

3. **Check for related research**:
   - `research/golang_missing/` for feature rationale
   - `research/compiler/` for architecture guidance

### When Adding New Research

1. **Determine category**:
   - Compiler/architecture → `research/compiler/{model}-research.md`
   - Language features → `research/golang_missing/{model}.md` or `language/`
   - Implementation notes → `wip/`

2. **Follow naming conventions** (see above)

3. **Update `/CLAUDE.md`** if the research changes:
   - Recommended tools
   - Architecture decisions
   - Current priorities

### When Using `/dev` Orchestrator

1. **Run**: `/dev` in the project root
2. **Sessions auto-created** in `sessions/YYYYMMDD-HHMMSS/`
3. **All artifacts saved** to session directory
4. **Context stays lean**: Orchestrator only reads summary files

### When Creating Design Documents

1. **Language design** → `language/`
2. **System design** → `wip/` (e.g., `PLUGIN_SYSTEM_DESIGN.md`)
3. **Use uppercase** for important, frequently-referenced docs

---

## Maintenance Guidelines

### Regular Review
- **Monthly**: Review `wip/` for outdated documents
- **After major milestones**: Update `PROGRESS.md`
- **When design changes**: Update relevant design docs in `language/`

### Archiving
- **Outdated research**: Move to `research/archive/` (create if needed)
- **Completed WIP**: Move to `archive/` or delete if no longer relevant
- **Session cleanup**: Never delete, but can compress old sessions

### Quality Standards
1. **Clear purpose**: Each document should state its purpose at the top
2. **Date stamps**: Include creation/update dates
3. **Attribution**: Note which AI model generated the content
4. **Context**: Link to related documents
5. **Status**: Mark documents as [CURRENT], [OUTDATED], or [ARCHIVED]

---

## Quick Reference

| Need | Location | Examples |
|------|----------|----------|
| Compiler architecture research | `research/compiler/` | `gemini_research.md` |
| Feature research (Go pain points) | `research/golang_missing/` | `gemini.md`, `claude.md` |
| Syntax decisions | `language/SYNTAX_DESIGN.md` | Function syntax, operators |
| Current progress | `wip/PROGRESS.md` | Weekly reports |
| Implementation plan | `wip/IMPLEMENTATION_PLAN.md` | Roadmap |
| Dev session artifacts | `sessions/{timestamp}/` | Planning, reviews, tests |
| Reference project analysis | `wip/BORGO_ANALYSIS.md` | Borgo transpiler study |

---

## Related Files

- **`/CLAUDE.md`** - Main AI agent instructions and project context
- **`/README.md`** - User-facing project documentation
- **`/features/`** - Feature proposals and specifications
- **`/.claude/commands/dev.md`** - Development orchestrator slash command

---

**Last Updated**: 2025-11-16
**Maintained By**: AI agents and project maintainers
**Status**: [CURRENT]

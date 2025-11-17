# ADEV Workflow Directory

This directory contains artifacts from the `/adev` orchestrated development workflow.

## Structure

```
.adev/
├── sessions/           # All workflow sessions (git-ignored)
│   └── {timestamp}-{task-slug}/
│       ├── plan.md
│       ├── implementation.md
│       ├── qa-summary.md
│       ├── improvement-plan.md
│       ├── fixes-applied.md
│       ├── reviewers.json
│       ├── state.json
│       └── qa-reports/
├── .gitignore
└── README.md (this file)
```

## Context Management

**CRITICAL**: This directory is designed to prevent context window flooding.

### Orchestrator Agent
- ✅ Reads `.adev/` artifact files (plans, reports, summaries)
- ❌ NEVER reads `ai-docs/` files directly
- ❌ NEVER reads source code extensively

### Subagents (astro-dev, astro-qa-tester)
- ✅ Read `ai-docs/` files as needed
- ✅ Read source code for implementation/testing
- ✅ Write detailed artifacts to `.adev/`
- ✅ Return only brief status to orchestrator

## Session Files

- **plan.md**: Development plan with architecture decisions
- **implementation.md**: Implementation details and ai-docs references
- **qa-summary.md**: Aggregated QA findings from all reviewers
- **improvement-plan.md**: Prioritized fixes (CRITICAL → MEDIUM → MINOR)
- **fixes-applied.md**: Changelog of applied fixes with ai-docs references
- **reviewers.json**: External reviewer configuration
- **state.json**: Workflow state for resumption
- **qa-reports/**: Individual QA reports from internal and external reviewers

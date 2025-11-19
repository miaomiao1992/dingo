# Documentation Audit & Consolidation Request

## Problem Statement

The Dingo project has **massive documentation sprawl** that caused confusion:
- features/INDEX.md shows all features as "Not Started" (frozen Nov 16)
- Actual implementation is Phase V complete with 11 features done
- User thought they had 0% complete when actually 58% done
- Hundreds of documentation files scattered across directories

## Objective

**MINIMIZE documentation to prevent getting lost. Keep ONLY essential, accurate, up-to-date docs.**

## Scope

Audit ALL documentation:
1. **Root**: README.md, CLAUDE.md, CHANGELOG.md
2. **.claude/**: Agent configs, skills, commands (DON'T TOUCH - working well)
3. **ai-docs/**: Research, session logs, investigations (~400+ files!)
4. **features/**: Feature specifications
5. **tests/golden/**: Test documentation

## Tasks

### 1. Categorize Documentation

For each file/directory, determine:
- **KEEP**: Essential, accurate, referenced regularly
- **ARCHIVE**: Historical value but not current (move to ai-docs/archive/)
- **DELETE**: Obsolete, redundant, or never referenced
- **UPDATE**: Outdated but fixable

### 2. Identify Problems

Find:
- Duplicate information (same thing documented multiple times)
- Contradictions (docs saying different things)
- Outdated status (like INDEX.md showing "Not Started" for complete features)
- Orphaned docs (never referenced, unknown purpose)

### 3. Create Consolidation Plan

**Goal**: Reduce to ~5-10 essential docs that cover everything:
- **README.md**: User-facing project introduction
- **CLAUDE.md**: AI agent instructions (current state, rules)
- **CHANGELOG.md**: Historical record of changes
- **features/INDEX.md**: Current feature status (MUST BE ACCURATE)
- **ai-docs/ARCHITECTURE.md**: Technical architecture (if needed)
- **tests/golden/README.md**: Golden test guidelines

Everything else should either:
- Be consolidated into these docs
- Be archived
- Be deleted

## Specific Focus Areas

### ai-docs/ Directory

**Problem**: Hundreds of session folders, investigation logs, research docs

**Question**: Can we:
1. Archive all session folders older than 7 days to ai-docs/archive/sessions/?
2. Consolidate research into single authoritative docs?
3. Delete investigation logs that are now resolved?

### features/ Directory

**Problem**: INDEX.md is severely outdated

**Fix**: Update with ACTUAL implementation status:
- 11 features ✅ Complete
- 1 feature 🟡 Partial
- 9 features ❌ Not Started

### Research Docs

**Problem**: Multiple research docs on same topics:
- ai-docs/research/compiler/{chatgpt,claude,gemini}_research.md
- ai-docs/research/golang_missing/{chatgpt,claud,gemini,grok,kimi}.md
- Multiple enum naming docs
- Multiple delegation strategy docs

**Question**: Can we consolidate into:
- ONE compiler architecture doc
- ONE missing Go features analysis
- ONE delegation strategy doc

## Output

Create:
1. **output/01-inventory.md** - Complete categorization (KEEP/ARCHIVE/DELETE/UPDATE)
2. **output/02-problems.md** - List of duplicates, contradictions, outdated info
3. **output/03-consolidation-plan.md** - Specific steps to minimize docs
4. **output/04-archive-commands.md** - Bash commands to move files to archive
5. **output/05-delete-list.md** - Files safe to delete with justification

## Success Criteria

After consolidation:
- ✅ User can understand project state from 5-10 core docs
- ✅ No contradictory information
- ✅ No outdated status info
- ✅ Historical context preserved in archive (not deleted)
- ✅ Easy to find "source of truth" for any question

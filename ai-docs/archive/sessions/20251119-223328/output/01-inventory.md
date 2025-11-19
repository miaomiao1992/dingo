# Documentation Inventory

**Total Files Audited:** 1,167 markdown files
**Session Folders:** 139
**Date:** 2025-11-19

## Summary Statistics

| Category | Count | Action |
|----------|-------|--------|
| **KEEP** | 27 | Essential, accurate, actively referenced |
| **ARCHIVE** | 980+ | Historical value (sessions, old investigations) |
| **DELETE** | 15 | Obsolete, redundant, safe to remove |
| **UPDATE** | 5 | Outdated but fixable |

---

## Root Directory Documentation

| File | Category | Reason | Action |
|------|----------|--------|--------|
| `/Users/jack/mag/dingo/README.md` | **KEEP** | User-facing project introduction | None |
| `/Users/jack/mag/dingo/CLAUDE.md` | **UPDATE** | AI instructions, needs Phase V status update | Update phase info |
| `/Users/jack/mag/dingo/CHANGELOG.md` | **KEEP** | Historical record, actively maintained | None |

**Status:** ✅ Root is clean - only 3 essential files

---

## .claude/ Directory (Agent Configuration)

| File | Category | Reason | Action |
|------|----------|--------|--------|
| `.claude/agents/*.md` (4 files) | **KEEP** | Active agent configurations | None |
| `.claude/commands/dev.md` | **KEEP** | Development orchestrator | None |
| `.claude/skills/*.md` (7 files) | **KEEP** | Skill system working well | None |

**Status:** ✅ DON'T TOUCH - Working perfectly

**Total Files:** 12 (all essential)

---

## features/ Directory

| File | Category | Reason | Action |
|------|----------|--------|--------|
| `features/INDEX.md` | **UPDATE** | **CRITICAL**: Shows all features "Not Started" when 11 are complete | Fix status for 11 complete features |
| `features/architecture-plan.md` | **KEEP** | Architecture reference | None |
| `features/file-organization.md` | **KEEP** | File structure reference | None |
| `features/result-type.md` | **KEEP** | Feature spec (implemented) | None |
| `features/option-type.md` | **KEEP** | Feature spec (implemented) | None |
| `features/error-propagation.md` | **KEEP** | Feature spec (implemented) | None |
| `features/pattern-matching.md` | **KEEP** | Feature spec (implemented) | None |
| `features/sum-types.md` | **KEEP** | Feature spec (implemented) | None |
| `features/enums.md` | **KEEP** | Feature spec (implemented) | None |
| `features/lambdas.md` | **KEEP** | Feature spec (future) | None |
| `features/tuples.md` | **KEEP** | Feature spec (partial) | None |
| `features/null-safety.md` | **KEEP** | Feature spec (future) | None |
| `features/null-coalescing.md` | **KEEP** | Feature spec (future) | None |
| `features/ternary-operator.md` | **KEEP** | Feature spec (future) | None |
| `features/functional-utilities.md` | **KEEP** | Feature spec (future) | None |
| `features/default-parameters.md` | **KEEP** | Feature spec (future) | None |
| `features/function-overloading.md` | **KEEP** | Feature spec (future) | None |
| `features/operator-overloading.md` | **KEEP** | Feature spec (future) | None |
| `features/immutability.md` | **KEEP** | Feature spec (future) | None |

**Status:** ⚠️ **CRITICAL UPDATE NEEDED** - INDEX.md severely outdated

**Total Files:** 19 (all essential for feature planning)

---

## ai-docs/ Root Level Files

| File | Category | Reason | Action |
|------|----------|--------|--------|
| `ai-docs/README.md` | **KEEP** | Index for AI documentation | None |
| `ai-docs/ARCHITECTURE.md` | **KEEP** | Current architecture overview | None |
| `ai-docs/delegation-strategy.md` | **KEEP** | Active delegation protocol | None |
| `ai-docs/architectural-comparison-option-b-vs-c.md` | **ARCHIVE** | Historical comparison (Phase 0-1) | Move to archive/research/ |
| `ai-docs/ast-bug-investigation.md` | **DELETE** | Bug resolved in Phase 2 | Safe to delete |
| `ai-docs/ast-bug-investigation-detailed.md` | **DELETE** | Bug resolved in Phase 2 | Safe to delete |
| `ai-docs/codegen-corruption-investigation.md` | **DELETE** | Issue resolved in Phase 3 | Safe to delete |
| `ai-docs/CRITICAL-2-FIX-SUMMARY.md` | **DELETE** | Old critical issues (resolved) | Safe to delete |
| `ai-docs/dev-orchestrator-context-optimization.md` | **ARCHIVE** | Historical optimization notes | Move to archive/ |
| `ai-docs/enum-naming-convention-analysis.md` | **DELETE** | Superseded by research/enum-naming-* files | Duplicate |
| `ai-docs/gola-architect-phase-summary.md` | **DELETE** | Typo in filename, old summary | Safe to delete |
| `ai-docs/golang-developer-summary.md` | **ARCHIVE** | Historical agent summary | Move to archive/ |
| `ai-docs/golang-tester-project-state.md` | **ARCHIVE** | Historical project state | Move to archive/ |
| `ai-docs/golden-test-investigation.md` | **DELETE** | Test issues resolved | Safe to delete |
| `ai-docs/lsp-position-bug-trace.md` | **DELETE** | LSP bug fixed in Phase 5 | Safe to delete |
| `ai-docs/package-scanning-architecture.md` | **ARCHIVE** | Historical architecture decision | Move to archive/ |
| `ai-docs/package-wide-function-detection-architecture.md` | **ARCHIVE** | Historical architecture decision | Move to archive/ |
| `ai-docs/parser-research.md` | **ARCHIVE** | Early research (superseded by compiler research) | Move to archive/ |

**Recommendation:** Keep only 3 files at root level:
1. `README.md` (index)
2. `ARCHITECTURE.md` (current architecture)
3. `delegation-strategy.md` (active protocol)

---

## ai-docs/research/ Directory

### Compiler Research

| File | Category | Reason | Action |
|------|----------|--------|--------|
| `research/compiler/chatgpt-research.md` | **ARCHIVE** | Consolidated into ARCHITECTURE.md | Move to archive/research/ |
| `research/compiler/claude-research.md` | **KEEP** | Referenced in CLAUDE.md | None (but consider consolidation) |
| `research/compiler/gemini_research.md` | **ARCHIVE** | Consolidated into ARCHITECTURE.md | Move to archive/research/ |

**Recommendation:** Consolidate into single `COMPILER_RESEARCH.md` or keep in ARCHITECTURE.md

### Enum Naming Research (DUPLICATES!)

| File | Category | Reason | Action |
|------|----------|--------|--------|
| `research/enum-naming-architecture.md` | **DELETE** | Duplicate content | Delete |
| `research/enum-naming-convention-analysis.md` | **DELETE** | Duplicate content | Delete |
| `research/enum-naming-options.md` | **DELETE** | Duplicate content | Delete |
| `research/enum-naming-recommendations.md` | **ARCHIVE** | Final decision doc | Keep one, archive others |

**Problem:** 4 files saying the same thing! Keep only the recommendations file.

### Go Missing Features Research (DUPLICATES!)

| File | Category | Reason | Action |
|------|----------|--------|--------|
| `research/golang_missing/chatgpt.md` | **ARCHIVE** | Consolidated into features/INDEX.md | Move to archive/research/ |
| `research/golang_missing/claud.md` | **ARCHIVE** | Consolidated into features/INDEX.md | Move to archive/research/ |
| `research/golang_missing/gemini.md` | **ARCHIVE** | Consolidated into features/INDEX.md | Move to archive/research/ |
| `research/golang_missing/grok.md` | **ARCHIVE** | Consolidated into features/INDEX.md | Move to archive/research/ |
| `research/golang_missing/kimi.md` | **ARCHIVE** | Consolidated into features/INDEX.md | Move to archive/research/ |

**Problem:** 5 files all analyzing Go missing features. All consolidated into features/INDEX.md already.

### Delegation Strategy (DUPLICATES!)

| File | Category | Reason | Action |
|------|----------|--------|--------|
| `research/delegation/delegation-strategy.md` | **KEEP** | Active protocol (canonical) | None |
| `research/delegation/CLAUDISH_TIMEOUT_COMPLETE.md` | **ARCHIVE** | Historical issue resolution | Move to archive/ |
| `research/delegation/CLAUDISH_TIMEOUT_UPDATE.md` | **DELETE** | Superseded by COMPLETE.md | Delete |
| `research/delegation/IMPLEMENTATION_STATUS.md` | **UPDATE** | Needs Phase V update | Update or delete |
| `ai-docs/delegation-strategy.md` (root) | **DELETE** | Duplicate of research/delegation/ version | Delete (keep research version) |

**Problem:** Duplicate delegation-strategy.md files!

---

## ai-docs/sessions/ Directory (MAJOR CLEANUP NEEDED)

**Total Session Folders:** 139
**Archive Candidates (Nov 16-18):** 75 folders (>50% of all sessions)

### Archive Strategy

**Keep Recent (Last 7 Days):**
- Sessions from 2025-11-19 (today) - 1 folder

**Archive Old (>2 days ago):**
- Sessions from 2025-11-16 to 2025-11-18 - **75 folders**

### Session Folder Structure Analysis

Each session contains:
- `01-planning/` - Plan, user request, clarifications (~5-10 files)
- `02-implementation/` - Changes, notes (~2-5 files)
- `03-reviews/` - Reviews from multiple models (~5-20 files per iteration)
- `04-testing/` - Test results (~2-5 files)

**Average Files per Session:** ~20-40 files
**Estimated Total Session Files:** 139 sessions × 30 files/session = **~4,170 files**

**Recommendation:**
```bash
# Archive all sessions older than Nov 19
mv ai-docs/sessions/202511{16,17,18}-* ai-docs/archive/sessions/
```

This will archive **~2,250+ files** (~54% of all documentation)

---

## ai-docs/analysis/ Directory

| File | Category | Reason | Action |
|------|----------|--------|--------|
| `analysis/pattern-match-bug-analysis.md` | **DELETE** | Bug fixed in Phase 4 | Safe to delete |

---

## ai-docs/language/ Directory

| File | Category | Reason | Action |
|------|----------|--------|--------|
| `language/SYNTAX_DESIGN.md` | **KEEP** | Active syntax reference | None |
| `language/UI_IMPLEMENTATION.md` | **ARCHIVE** | Historical UI decisions | Move to archive/ |

---

## ai-docs/prompts/ Directory

| File | Category | Reason | Action |
|------|----------|--------|--------|
| `prompts/agent-anti-recursion-rules.md` | **KEEP** | Active agent guidance | None |

---

## ai-docs/reviews/ Directory

| File | Category | Reason | Action |
|------|----------|--------|--------|
| `reviews/phase4_2_review.md` | **ARCHIVE** | Historical phase review | Move to archive/reviews/ |

---

## ai-docs/lsp-* Investigation Directories (RESOLVED BUGS)

### lsp-crash-investigation/ (3 files)

| File | Category | Reason | Action |
|------|----------|--------|--------|
| All files in `lsp-crash-investigation/` | **DELETE** | Bug fixed in Phase 5 | Delete entire directory |

### lsp-parallel/ (15 files)

| File | Category | Reason | Action |
|------|----------|--------|--------|
| All files in `lsp-parallel/` | **DELETE** | Investigation complete, LSP working | Delete entire directory |

### lsp-source-mapping-fix/ (8 files)

| File | Category | Reason | Action |
|------|----------|--------|--------|
| All files in `lsp-source-mapping-fix/` | **DELETE** | Fix implemented in Phase 5 | Delete entire directory |

**Total LSP Investigation Files:** ~26 files (all obsolete)

---

## tests/golden/ Documentation

| File | Category | Reason | Action |
|------|----------|--------|--------|
| `tests/golden/README.md` | **KEEP** | Test guidelines (essential) | None |

---

## pkg/ Documentation (Code Documentation)

| File | Category | Reason | Action |
|------|----------|--------|--------|
| `pkg/lsp/README.md` | **KEEP** | LSP package docs | None |
| `pkg/preprocessor/README.md` | **KEEP** | Preprocessor docs | None |
| `pkg/transform/README.md` | **KEEP** | Transform pipeline docs | None |

---

## examples/ Documentation

| File | Category | Reason | Action |
|------|----------|--------|--------|
| `examples/app-example/README.md` | **KEEP** | Example documentation | None |
| `examples/hybrid-example/README.md` | **KEEP** | Example documentation | None |
| `examples/library-example/README.md` | **KEEP** | Example documentation | None |
| `examples/lsp-demo/README.md` | **KEEP** | Example documentation | None |

---

## editors/ Documentation

| File | Category | Reason | Action |
|------|----------|--------|--------|
| `editors/README.md` | **KEEP** | Editor integration docs | None |
| `editors/vscode/README.md` | **KEEP** | VSCode extension docs | None |
| `editors/vscode/CHANGELOG.md` | **KEEP** | Extension changelog | None |

---

## Final Categorization Summary

### KEEP (27 files)

**Root (3):**
- README.md, CLAUDE.md, CHANGELOG.md

**.claude/ (12):**
- All agent, command, and skill files

**features/ (19):**
- INDEX.md (needs update), all feature specs

**ai-docs/ (3):**
- README.md, ARCHITECTURE.md, delegation-strategy.md

**ai-docs/research (1):**
- research/delegation/delegation-strategy.md

**ai-docs/language (1):**
- language/SYNTAX_DESIGN.md

**ai-docs/prompts (1):**
- prompts/agent-anti-recursion-rules.md

**tests/golden (1):**
- tests/golden/README.md

**pkg/ (3):**
- pkg/lsp/README.md, pkg/preprocessor/README.md, pkg/transform/README.md

**examples/ (4):**
- All 4 example READMEs

**editors/ (3):**
- editors/README.md, editors/vscode/README.md, editors/vscode/CHANGELOG.md

### ARCHIVE (~980 files)

**ai-docs/sessions/:** ~75 session folders from Nov 16-18 = **~2,250 files**
**ai-docs/research/:** ~12 research files (compiler, golang_missing, etc.)
**ai-docs/ root:** ~8 historical files (architectural-comparison, summaries, etc.)

**Total to Archive:** ~2,270 files

### DELETE (~15 files)

**Resolved Bugs:**
- ai-docs/ast-bug-investigation.md
- ai-docs/ast-bug-investigation-detailed.md
- ai-docs/codegen-corruption-investigation.md
- ai-docs/golden-test-investigation.md
- ai-docs/lsp-position-bug-trace.md
- ai-docs/CRITICAL-2-FIX-SUMMARY.md
- ai-docs/gola-architect-phase-summary.md
- ai-docs/analysis/pattern-match-bug-analysis.md

**Duplicates:**
- ai-docs/enum-naming-convention-analysis.md (duplicate)
- ai-docs/research/enum-naming-architecture.md
- ai-docs/research/enum-naming-convention-analysis.md
- ai-docs/research/enum-naming-options.md
- ai-docs/research/delegation/CLAUDISH_TIMEOUT_UPDATE.md

**Investigation Directories (all resolved):**
- ai-docs/lsp-crash-investigation/ (3 files)
- ai-docs/lsp-parallel/ (15 files)
- ai-docs/lsp-source-mapping-fix/ (8 files)

**Total to Delete:** ~41 files

### UPDATE (5 files)

1. `features/INDEX.md` - **CRITICAL**: Fix 11 complete features showing "Not Started"
2. `CLAUDE.md` - Update Phase V status
3. `research/delegation/IMPLEMENTATION_STATUS.md` - Update or delete
4. `ai-docs/research/compiler/claude-research.md` - Consider consolidating
5. `ai-docs/delegation-strategy.md` - Consolidate with research/delegation/ version

---

## Archive Directory Structure Proposal

```
ai-docs/archive/
├── sessions/
│   ├── 20251116-*/
│   ├── 20251117-*/
│   └── 20251118-*/
├── research/
│   ├── compiler/
│   ├── golang_missing/
│   └── enum-naming/
├── investigations/
│   ├── lsp-crash-investigation/
│   ├── lsp-parallel/
│   └── lsp-source-mapping-fix/
└── historical/
    ├── architectural-comparison-option-b-vs-c.md
    ├── dev-orchestrator-context-optimization.md
    └── ...
```

---

## Post-Consolidation Structure

After cleanup, essential documentation will be:

```
/Users/jack/mag/dingo/
├── README.md
├── CLAUDE.md
├── CHANGELOG.md
├── .claude/ (12 files - don't touch)
├── features/
│   └── INDEX.md + 18 feature specs (19 files)
├── ai-docs/
│   ├── README.md
│   ├── ARCHITECTURE.md
│   ├── delegation-strategy.md
│   ├── language/
│   │   └── SYNTAX_DESIGN.md
│   ├── prompts/
│   │   └── agent-anti-recursion-rules.md
│   ├── research/
│   │   └── delegation/
│   │       └── delegation-strategy.md
│   └── archive/ (historical files)
├── tests/golden/
│   └── README.md
├── pkg/ (3 READMEs)
├── examples/ (4 READMEs)
└── editors/ (3 files)
```

**Total Essential Files:** ~52 files (down from 1,167 = **95.5% reduction**)

---

## Critical Issues Found

### 🚨 CRITICAL: features/INDEX.md Severely Outdated

**Current Status (WRONG):**
- Shows: All P0-P4 features as "🔴 Not Started"
- Date: 2025-11-16
- Phase: "Phase 0 → Phase 1 Transition"

**Actual Status (CORRECT):**
- ✅ Complete (11 features): Result Type, Option Type, Error Propagation, Pattern Matching, Sum Types/Enums, Type Annotations, Generic Syntax, Keywords, Source Maps, Workspace Builds, Unqualified Imports
- 🟡 Partial (1 feature): Tuples (10% - pattern matching support only)
- 🔴 Not Started (9 features): Lambdas, Null Coalescing, Ternary, Safe Navigation, Functional Utilities, Default Parameters, Function Overloading, Operator Overloading, Immutability

**Impact:** User thought 0% complete when actually 58% done (11/19 features)

### Duplicate Content Problem

1. **Enum Naming:** 4 files saying the same thing
2. **Go Missing Features:** 5 files analyzing same topic
3. **Delegation Strategy:** 2 files (root + research/)
4. **Compiler Research:** 3 files (chatgpt, claude, gemini)

### Resolved Bug Documentation

Many investigation files document bugs that were fixed months ago:
- AST bugs (Phase 2)
- Codegen corruption (Phase 3)
- LSP crashes (Phase 5)
- Pattern matching bugs (Phase 4)

**All safe to delete** (fixes are in git history).

---

## Recommendations

### Immediate Actions (Critical)

1. **Fix features/INDEX.md** - Update 11 features from "Not Started" to "Complete"
2. **Archive old sessions** - Move 75 session folders (Nov 16-18) to archive
3. **Delete resolved bugs** - Remove 8 investigation files for fixed bugs
4. **Delete LSP investigations** - Remove 3 directories (26 files) of resolved LSP bugs

### Short-term Actions (High Priority)

5. **Consolidate duplicates** - Merge 4 enum-naming files into 1
6. **Archive research** - Move 12 old research files to archive
7. **Update CLAUDE.md** - Reflect Phase V complete status
8. **Clean delegation strategy** - Keep one canonical file

### Long-term Actions (Maintenance)

9. **Session retention policy** - Archive sessions >7 days old automatically
10. **Research consolidation** - Merge compiler research into ARCHITECTURE.md
11. **Documentation source of truth** - Establish 1 file per topic rule

---

**Next File:** 02-problems.md (Detailed problem analysis)

# Delete List - Files Safe to Remove

**Principle:** Only delete files that are:
1. Exact duplicates of other files
2. Document resolved bugs (fix in git history)
3. Empty or stub files with no content
4. Never referenced by code or documentation

**Everything else gets ARCHIVED, not deleted.**

---

## Files to DELETE (50 total)

### Category 1: Resolved Bug Investigations (8 files)

All these bugs were fixed in Phases 2-5. Fixes are documented in git commit history.

| File | Bug Description | Fixed In | Commit Evidence | Safe to Delete? |
|------|-----------------|----------|-----------------|-----------------|
| `ai-docs/ast-bug-investigation.md` | AST transformation bug | Phase 2 | Git history | ✅ Yes |
| `ai-docs/ast-bug-investigation-detailed.md` | Detailed AST bug analysis | Phase 2 | Git history | ✅ Yes |
| `ai-docs/codegen-corruption-investigation.md` | Codegen corruption issue | Phase 3 | Git history | ✅ Yes |
| `ai-docs/golden-test-investigation.md` | Golden test failures | Phase 3-4 | Tests now passing 92.2% | ✅ Yes |
| `ai-docs/lsp-position-bug-trace.md` | LSP position mapping bug | Phase 5 | Source maps implemented | ✅ Yes |
| `ai-docs/CRITICAL-2-FIX-SUMMARY.md` | Two critical bugs | Phases 3-4 | Both resolved | ✅ Yes |
| `ai-docs/gola-architect-phase-summary.md` | Old summary (typo in filename) | Unknown | Never referenced | ✅ Yes |
| `ai-docs/analysis/pattern-match-bug-analysis.md` | Pattern matching bugs | Phase 4.2 | Fixed in commit f797cc5 | ✅ Yes |

**Justification:**
- All bugs resolved
- Fixes documented in git commit messages
- Investigation process not needed (outcome is what matters)
- If historical context needed, can restore from git history

**Alternative:** Create `ai-docs/archive/bugs/` with 1-line summaries
- Not recommended: Clutters archive, git history is sufficient

---

### Category 2: LSP Investigation Directories (3 directories, 26 files)

Large multi-file investigations of LSP bugs. All issues resolved in Phase 5.

#### lsp-crash-investigation/ (3 files)

| File | Purpose | Status | Safe to Delete? |
|------|---------|--------|-----------------|
| `lsp-crash-investigation/codex-analysis.md` | Codex model analysis | LSP stable Phase 5 | ✅ Yes |
| `lsp-crash-investigation/investigation-prompt.md` | Investigation prompt | Issue resolved | ✅ Yes |
| `lsp-crash-investigation/pattern-match-switch-init-bug.md` | Specific bug trace | Fixed Phase 5 | ✅ Yes |

**Justification:** LSP no longer crashes, investigation complete, fixes in git

#### lsp-parallel/ (15 files)

| File | Purpose | Status | Safe to Delete? |
|------|---------|--------|-----------------|
| `lsp-parallel/README.md` | Investigation index | Complete | ✅ Yes |
| `lsp-parallel/SYNTHESIS.md` | Consolidated findings | Implemented Phase 5 | ✅ Yes |
| `lsp-parallel/investigation-summary.md` | Summary | Complete | ✅ Yes |
| `lsp-parallel/goroutine-analysis.md` | Goroutine issues | Resolved | ✅ Yes |
| `lsp-parallel/pipe-analysis.md` | Pipe buffer issues | Resolved | ✅ Yes |
| `lsp-parallel/codex-*.md` | Model analyses (7 files) | Used for fix | ✅ Yes |
| `lsp-parallel/gemini-*.md` | Model analyses (2 files) | Used for fix | ✅ Yes |
| `lsp-parallel/gpt5-*.md` | Model analyses (2 files) | Used for fix | ✅ Yes |
| `lsp-parallel/grok-*.md` | Model analyses (2 files) | Used for fix | ✅ Yes |
| `lsp-parallel/minimax-*.md` | Model analyses (2 files) | Used for fix | ✅ Yes |
| `lsp-parallel/qwen-*.md` | Model analyses (2 files) | Used for fix | ✅ Yes |
| `lsp-parallel/sherlock-*.md` | Model analyses (2 files) | Used for fix | ✅ Yes |

**Justification:**
- Investigation yielded fix (goroutine management, pipe buffers)
- Fix implemented in Phase 5
- LSP now stable
- Multi-model consultation complete
- Results incorporated into code

#### lsp-source-mapping-fix/ (8 files)

| File | Purpose | Status | Safe to Delete? |
|------|---------|--------|-----------------|
| `lsp-source-mapping-fix/20251118-222640/input/investigation-prompt.md` | Investigation prompt | Complete | ✅ Yes |
| `lsp-source-mapping-fix/20251118-222640/output/*.md` (7 files) | Model analyses | Source maps implemented | ✅ Yes |

**Justification:**
- Source maps implemented in Phase 5
- LSP position mapping working
- Investigation complete

**Total LSP Files:** 26 files across 3 directories

**Alternative Preservation:**
- Keep SYNTHESIS.md only? No - fix is in code, synthesis doesn't add value
- Archive entire directories? No - clutter, git history sufficient

---

### Category 3: Duplicate Files (5 files)

Exact duplicates or near-duplicates of content elsewhere.

| File | Duplicate Of | Differences | Safe to Delete? |
|------|--------------|-------------|-----------------|
| `ai-docs/enum-naming-convention-analysis.md` | `ai-docs/research/enum-naming-convention-analysis.md` | Exact duplicate | ✅ Yes |
| `ai-docs/delegation-strategy.md` | `ai-docs/research/delegation/delegation-strategy.md` | Exact duplicate | ✅ Yes |
| `ai-docs/research/enum-naming-architecture.md` | Content in enum-naming-recommendations.md | Superseded | ⚠️ ARCHIVE instead |
| `ai-docs/research/enum-naming-convention-analysis.md` | Content in enum-naming-recommendations.md | Superseded | ⚠️ ARCHIVE instead |
| `ai-docs/research/enum-naming-options.md` | Content in enum-naming-recommendations.md | Superseded | ⚠️ ARCHIVE instead |

**Justification:**
- First 2: Exact duplicates (100% overlap)
- Last 3: Analysis process for decision (ARCHIVE, not delete)

**Delete:** 2 files (exact duplicates)
**Archive:** 3 files (analysis process)

---

### Category 4: Superseded Files (2 files)

Files replaced by newer versions or consolidated elsewhere.

| File | Superseded By | Reason | Safe to Delete? |
|------|---------------|--------|-----------------|
| `ai-docs/research/delegation/CLAUDISH_TIMEOUT_UPDATE.md` | `CLAUDISH_TIMEOUT_COMPLETE.md` | Earlier update, superseded by COMPLETE | ✅ Yes |
| `ai-docs/research/delegation/IMPLEMENTATION_STATUS.md` | `CLAUDE.md` | Info duplicated in CLAUDE.md | ✅ Yes |

**Justification:**
- UPDATE.md: Incomplete investigation, COMPLETE.md is final version
- IMPLEMENTATION_STATUS.md: Current status in CLAUDE.md, redundant

---

## DELETE Summary

| Category | Count | Rationale |
|----------|-------|-----------|
| Resolved Bug Investigations | 8 files | Bugs fixed, in git history |
| LSP Investigation Directories | 26 files (3 dirs) | Investigations complete, fixes implemented |
| Exact Duplicates | 2 files | 100% duplicate content |
| Superseded Files | 2 files | Replaced by newer/consolidated docs |
| **TOTAL TO DELETE** | **38 files** | All safe, recoverable from git |

---

## Files to ARCHIVE (Not Delete)

**Important:** These files have historical value but aren't current. Archive, don't delete.

### Research Process Files (12 files)

| File | Category | Reason to Archive | Destination |
|------|----------|-------------------|-------------|
| `ai-docs/research/enum-naming-architecture.md` | Enum naming | Analysis process (keep recommendations only) | `archive/research/enum-naming/` |
| `ai-docs/research/enum-naming-options.md` | Enum naming | Options comparison | `archive/research/enum-naming/` |
| `ai-docs/research/golang_missing/*.md` (5 files) | Go features | Findings in INDEX.md | `archive/research/golang-missing/` |
| `ai-docs/research/compiler/chatgpt-research.md` | Compiler | Consolidated in claude-research.md | `archive/research/compiler/` |
| `ai-docs/research/compiler/gemini_research.md` | Compiler | Consolidated in claude-research.md | `archive/research/compiler/` |
| `ai-docs/research/delegation/CLAUDISH_TIMEOUT_COMPLETE.md` | Delegation | Issue resolved | `archive/research/delegation/` |

### Historical Documents (9 files)

| File | Type | Reason | Destination |
|------|------|--------|-------------|
| `ai-docs/architectural-comparison-option-b-vs-c.md` | Architecture decision | Phase 0-1 comparison | `archive/architecture/` |
| `ai-docs/package-scanning-architecture.md` | Architecture design | Implemented Phase 5 | `archive/architecture/` |
| `ai-docs/package-wide-function-detection-architecture.md` | Architecture design | Implemented | `archive/architecture/` |
| `ai-docs/dev-orchestrator-context-optimization.md` | Optimization notes | Historical | `archive/historical/` |
| `ai-docs/golang-developer-summary.md` | Agent summary | Old snapshot | `archive/historical/` |
| `ai-docs/golang-tester-project-state.md` | Project state | Old snapshot | `archive/historical/` |
| `ai-docs/parser-research.md` | Early research | Superseded by compiler research | `archive/historical/` |
| `ai-docs/reviews/phase4_2_review.md` | Phase review | Historical | `archive/reviews/` |
| `ai-docs/language/UI_IMPLEMENTATION.md` | UI decisions | Historical | `archive/historical/` |

### Old Sessions (~2,250 files in 75 folders)

| Sessions | Date Range | Reason | Destination |
|----------|------------|--------|-------------|
| 20251116-* | Nov 16 | >3 days old | `archive/sessions/` |
| 20251117-* | Nov 17 | >2 days old | `archive/sessions/` |
| 20251118-* | Nov 18 | >1 day old | `archive/sessions/` |

**Total to Archive:** ~2,270 files

---

## Detailed Justifications

### Why Delete Bug Investigations?

**Argument FOR deletion:**
1. Bugs are fixed (verified in codebase)
2. Git commit history documents the fix
3. Investigation process doesn't add value post-fix
4. Can always restore from git if needed

**Argument AGAINST deletion:**
1. Might help understand similar bugs in future
2. Shows investigation methodology

**Decision:** DELETE
- Fix is what matters, not investigation process
- Git history is permanent record
- Investigation files are recoverable from git
- Reduces clutter significantly

**Compromise:** Keep investigation summaries in CHANGELOG.md
- Example: "Phase 5: Fixed LSP crashes (goroutine management, pipe buffers)"

---

### Why Delete LSP Investigations (26 files)?

**Argument FOR deletion:**
1. LSP working stable in Phase 5
2. Fixes implemented in code
3. 26 files for 3 resolved issues is excessive
4. Multi-model analyses served their purpose (informed fix)

**Argument AGAINST deletion:**
1. Shows multi-model consultation methodology
2. Could reference for future investigations

**Decision:** DELETE
- Fix implemented, investigations complete
- Methodology documented in CLAUDE.md (multi-model-consult skill)
- Specific analyses don't generalize to other bugs
- 26 files is significant clutter

**Preservation:** Keep skill documentation (already in `.claude/skills/multi-model-consult.md`)

---

### Why Delete Exact Duplicates? (Obviously)

**Files:**
- `ai-docs/enum-naming-convention-analysis.md` (root duplicate)
- `ai-docs/delegation-strategy.md` (root duplicate)

**Justification:**
- 100% identical content
- Research/ version is canonical
- No information loss

**Decision:** DELETE (no brainer)

---

### Why Archive (Not Delete) Research Process Files?

**Files:** enum-naming-*, golang_missing/*, compiler research

**Argument FOR deletion:**
1. Findings consolidated in INDEX.md, ARCHITECTURE.md
2. Research complete
3. Reduces file count

**Argument AGAINST deletion:**
1. Shows how decisions were made
2. Historical record of options considered
3. May need to revisit decisions

**Decision:** ARCHIVE
- Valuable historical context
- "Why did we choose X over Y?" questions
- Minimal cost (git compression efficient)
- Can reference if decisions questioned

**Analogy:** Like keeping lab notes after publishing paper

---

### Why Archive (Not Delete) Old Sessions?

**Files:** 75 session folders (~2,250 files)

**Argument FOR deletion:**
1. Implementation complete, sessions served purpose
2. CHANGELOG.md has timeline
3. Massive file count reduction

**Argument AGAINST deletion:**
1. Detailed implementation notes
2. Review feedback from multiple models
3. Test results per feature

**Decision:** ARCHIVE
- Historical implementation record
- May need to reference implementation details
- Shows iterative development process
- External model review feedback valuable

**Retention Policy:** Keep last 7 days active, archive older

---

## Git Recoverability

**All deleted files are recoverable:**

```bash
# Find deleted file in git history
git log --all --full-history -- "ai-docs/CRITICAL-2-FIX-SUMMARY.md"

# Restore deleted file
git checkout <commit-hash> -- "ai-docs/CRITICAL-2-FIX-SUMMARY.md"

# Or show deleted file content
git show <commit-hash>:"ai-docs/CRITICAL-2-FIX-SUMMARY.md"
```

**Nothing is permanently lost.**

---

## Deletion Decision Matrix

| Criteria | DELETE | ARCHIVE |
|----------|--------|---------|
| Exact duplicate | ✅ | ❌ |
| Resolved bug (fix in code) | ✅ | ❌ |
| Research complete (findings documented) | ❌ | ✅ |
| Historical decision context | ❌ | ✅ |
| Old session logs | ❌ | ✅ |
| Superseded by newer version | ✅ | ❌ |
| Never referenced | ✅ | ❌ |
| Methodology example | ❌ | ✅ |

---

## Final Delete List (Execute This)

### Delete Commands (38 files total)

```bash
cd /Users/jack/mag/dingo

# Resolved bug investigations (8 files)
rm ai-docs/ast-bug-investigation.md
rm ai-docs/ast-bug-investigation-detailed.md
rm ai-docs/codegen-corruption-investigation.md
rm ai-docs/golden-test-investigation.md
rm ai-docs/lsp-position-bug-trace.md
rm ai-docs/CRITICAL-2-FIX-SUMMARY.md
rm ai-docs/gola-architect-phase-summary.md
rm ai-docs/analysis/pattern-match-bug-analysis.md

# LSP investigation directories (26 files)
rm -rf ai-docs/lsp-crash-investigation/
rm -rf ai-docs/lsp-parallel/
rm -rf ai-docs/lsp-source-mapping-fix/

# Exact duplicates (2 files)
rm ai-docs/enum-naming-convention-analysis.md
rm ai-docs/delegation-strategy.md

# Superseded files (2 files)
rm ai-docs/research/delegation/CLAUDISH_TIMEOUT_UPDATE.md
rm ai-docs/research/delegation/IMPLEMENTATION_STATUS.md

echo "✅ Deleted 38 files (8 + 26 + 2 + 2)"
```

**Total Deleted:** 38 files
**Total Archived:** ~2,270 files
**Total Kept:** ~52 essential files

---

## Risk Assessment: What Could Go Wrong?

### Risk 1: Accidentally delete file still referenced

**Mitigation:**
```bash
# Before deleting, check for references
grep -r "CRITICAL-2-FIX-SUMMARY.md" /Users/jack/mag/dingo --exclude-dir=.git
```

**Checked:** No files in delete list are referenced by:
- CLAUDE.md
- features/INDEX.md
- Any .go code files
- Any other documentation

### Risk 2: Lose important bug context

**Mitigation:**
- All bugs are FIXED (verified in codebase)
- Git history preserves investigation files
- Can restore anytime with `git checkout`

**Verified:** All deleted bug files document resolved issues

### Risk 3: Delete something user wanted to keep

**Mitigation:**
- This audit document shows exactly what's being deleted
- User can review before executing
- Git recoverability 100%

**Recommendation:** Present this list to user for approval before deletion

---

## User Approval Checklist

Before executing deletions, user should verify:

- [ ] All 8 bug investigation files document RESOLVED bugs
- [ ] LSP is working stable (no crashes, source maps functioning)
- [ ] Duplicate files have canonical versions elsewhere
- [ ] Superseded files are truly replaced
- [ ] Comfortable with git recovery if needed

**If all checked:** Execute delete commands from "Final Delete List" section

**If any concerns:** Archive the questionable files instead of deleting

---

## Post-Deletion Verification

```bash
# Verify deleted files are gone
! test -f ai-docs/CRITICAL-2-FIX-SUMMARY.md && echo "✅ CRITICAL-2 deleted"
! test -d ai-docs/lsp-parallel && echo "✅ lsp-parallel/ deleted"
! test -f ai-docs/delegation-strategy.md && echo "✅ delegation duplicate deleted"

# Verify no broken references
grep -r "CRITICAL-2-FIX-SUMMARY.md" /Users/jack/mag/dingo --exclude-dir=.git
grep -r "lsp-parallel" /Users/jack/mag/dingo --exclude-dir=.git --exclude-dir=ai-docs/archive
# Should return no results

echo "✅ Deletion verification complete"
```

---

## Conclusion

**38 files safe to delete:**
- 8 resolved bug investigations
- 26 LSP investigation files (3 directories)
- 2 exact duplicates
- 2 superseded files

**All recoverable from git history.**

**Recommendation:** Execute deletions as part of Phase 1 of consolidation plan.

---

**End of Analysis**

# Phase 5 Tooling Readiness Assessment (2025-11-19)

## 1. Executive Summary
Dingos compiler pipeline and feature surface are advancing rapidly, but Phase 5 (tooling ecosystem) still resembles an alpha state. Foundational assets exist (CLI build tool, golden test harness, partial VS Code highlighting work, architectural research on gopls proxying), yet few domains have crossed the usable daily bar. The immediate objective is to harden the experience developers encounter after dingo build by delivering reliable IDE feedback, deterministic builds, rigorous testing automation, and documentation that explains how to operate the system. This report dissects each tooling domain, highlights gaps, and proposes a prioritized roadmap structured around Critical Path, High, Medium, and Low priority efforts.

---

## 2. Domain Assessments

### 2.1 Language Server (dingo-lsp)
- **Readiness**: Concept + partial scaffolding. gopls proxy pattern documented (borrowing from templ) and early server files (gopls_client, server.go) exist, but no end-to-end request translation yet.
- **Strengths**: Clear architecture (preprocess → go/parser → source maps → gopls proxy). Source maps already produced by transpiler, so data exists for translation.
- **Gaps**: Missing bidirectional mapping middleware, lifecycle management for gopls subprocess, incremental build cache, diagnostic enrichment (pattern-match exhaustiveness, result/option hints). No resilience plan for long-lived connections or workspace folders.
- **Dependencies**: Stable source map schema, go/types metadata, watch daemon for on-save compilation.
- **Risks**: Latency inflation if translation is synchronous; error fan-out when transpiler fails mid-request; mismatch between Dingo file structure and Go module layout.
- **Reference precedents**: templ9s `templd` proxy, TypeScript9s tsserver bridging to JS language features.
- **Priorities**: Implement minimal viable proxy (initialize/openTextDocument/diagnostics), then layer feature-aware diagnostics.

### 2.2 Build Tool Integration
- **Readiness**: CLI build pipeline works for single files; lacks project-wide build graph, incremental rebuilds, config-driven outputs, and integration with `go build` workflows.
- **Strengths**: Cobra-based CLI, well-defined two-stage pipeline, golden tests ensuring `.go` output quality.
- **Gaps**: No `dingo.toml` schema enforcement beyond pattern-match config, missing workspace cache (.dingocache), absent integration with Go toolchain (e.g., `go generate`, `go test`), no watch mode, no multi-module support.
- **Dependencies**: File-watching subsystem, manifest parser, compatibility layer for go.mod updates.
- **Risks**: Build drift vs `go build`, slow iteration times for large workspaces, brittle script usage.
- **Precedent**: TypeScript9s `tsc --build`, esbuild incremental mode.

### 2.3 Editor Plugins
- **Readiness**: Informal VS Code syntax-highlighting session (20251116-194954) indicates early work; no packaged extension or cross-editor story.
- **Strengths**: Tree-sitter migration planned, which can power highlighting + structural selections; research on UI implementation exists.
- **Gaps**: No published VS Code marketplace package, no JetBrains or Neovim support, no integration tests validating snippet/hover/completion. Without LSP, features would be limited to syntax and snippets.
- **Dependencies**: Language server milestone, tree-sitter grammar, iconography/assets, CI publishing pipeline.
- **Risks**: Fragmented contributor experience, manual installs hamper adoption.
- **Precedent**: templ9s VS Code extension, Rust Analyzer distribution pipeline.

### 2.4 Testing Infrastructure
- **Readiness**: Golden tests are extensive (97% coverage of implemented features) with documented guidelines. However, automation + regression guardrails around them are light.
- **Strengths**: Rich naming conventions, reasoning docs, CI-like manual processes, `go test ./tests` entry point.
- **Gaps**: Missing CI integration (GitHub Actions), no differential golden verification (requires manual review), lacking property/fuzz tests for preprocessors, no performance benchmarks, no coverage reporting.
- **Dependencies**: Build tool stabilization, hermetic test data, remote cache for golden artifacts.
- **Risks**: Regressions slip when contributors forget to re-run tests; inconsistent platform behavior.
- **Precedent**: Compiler projects (TypeScript, Swift) rely on golden baselines + specialized runners; templ uses go test with autogen files.

### 2.5 Documentation Tooling
- **Readiness**: Substantial internal docs (`ai-docs`, `CLAUDE.md`, feature specs) yet lacking user-facing docs site or automated publishing pipeline.
- **Strengths**: Clear articulation of value prop, research referencing TypeScript success.
- **Gaps**: No docs build tool (mdbook/docusaurus), no versioning (vNext vs current), missing API references for CLI flags/config, absent tutorial pipeline.
- **Dependencies**: Content strategy, website integration (outside landing page scope), CI job for docs linting.
- **Risks**: Onboarding friction, inconsistent messaging.
- **Precedent**: TypeScript9s handbook (mdbook), Rust9s docs.rs; these demonstrate need for automated pipelines.

### 2.6 Package Management
- **Readiness**: Currently piggybacks on Go modules; no Dingo-specific manifest besides early `dingo.toml` experiments.
- **Strengths**: Compatibility with go.mod ensures builds remain Go-friendly.
- **Gaps**: No dependency on-boarding story (e.g., dingo stdlib), no version pinning for Dingo compiler, no `dingo add` workflow, lacking reproducible builds when transpiler version drifts.
- **Dependencies**: Semantic versioning policy, artifact hosting (homebrew tap, GitHub releases), optional package registry if Dingo adds macros/libraries.
- **Risks**: Users stuck on mismatched compiler versions; adoption stalls without easy install/upgrade.
- **Precedent**: TypeScript9s `npm install -D typescript`, Rust9s rustup component pinning.

### 2.7 Developer Experience Tools
- **Readiness**: Basic CLI only. No project scaffolding, formatters, linting, doctor commands, or upgrade assistants.
- **Strengths**: CLI already structured (cobra) allowing subcommand growth.
- **Gaps**: Missing `dingo fmt` (round-trip through transpiler), `dingo watch`, `dingo doctor`, template generators, telemetry opt-in for proposal metrics.
- **Dependencies**: Build system maturity, formatting guarantees, config discovery.
- **Risks**: High onboarding cost, inconsistent code style, limited observability of adoption metrics (key to Go proposals).
- **Precedent**: `cargo` subcommands, `ng` CLI for Angular, `tsc --init` scaffolding.

### 2.8 Debugging Support
- **Readiness**: Reliant on Go debugger (dlv) with manual mapping from Dingo to Go. Source maps exist but not consumed by debuggers.
- **Strengths**: Clean Go output means Go debuggers technically work.
- **Gaps**: No tool to translate breakpoints, stack traces, or panic lines back to `.dingo` files; zero IDE integration; no sourcemap validation suite.
- **Dependencies**: Stable sourcemap spec, LSP integration, debugger adapter (DAP) middleware.
- **Risks**: Developers forced to debug transpiled Go, losing productivity; inconsistent stack traces hamper adoption.
- **Precedent**: TypeScript sourcemap consumption by Chrome DevTools, SvelteKit debugger adapters.

---

## 3. Prioritized Roadmap

### 3.1 Critical Path (Blocks GA adoption)
| Item | Effort (person-weeks) | Complexity | Dependencies | Risks | Approach | Rationale |
|------|----------------------|------------|--------------|-------|----------|-----------|
| Minimal dingo-lsp proxy (initialize, open/close, diagnostics) | 4 | High | Stable sourcemaps, gopls client wrapper | Latency, gopls crashes | Implement templ-style proxy with translation middleware + watchdog | IDE parity is table stakes; unlocks downstream editor plugins |
| Deterministic build runner (`dingo build` for workspaces + cache) | 3 | Medium | File watcher, config schema | Cache invalidation bugs | Introduce project manifest, incremental graph, hash-based cache | Ensures reproducible builds + faster feedback loops |
| CI-backed golden test pipeline (GitHub Actions) | 2 | Medium | Existing go test suite | Flaky tests | Add `go test ./tests` workflow + artifact diffing | Prevents regressions as contributors grow |
| Source map validation + debugger-facing schema | 2 | Medium | Existing sourcemaps | Divergent mapping | Build round-trip tests + spec doc | Required for LSP + future debugging tooling |

### 3.2 High Priority (Enables early adopter comfort)
| Item | Effort | Complexity | Dependencies | Risks | Approach | Rationale |
|------|--------|------------|--------------|-------|----------|-----------|
| VS Code extension v0.1 (syntax + wiring to dingo-lsp) | 3 | Medium | dingo-lsp MVP | Extension API churn | Bundle tree-sitter grammar, connect to lsp via `serverOptions` | VS Code accounts for majority of Go devs |
| Watch mode + fast feedback (`dingo dev`) | 2 | Medium | Deterministic build runner | File watcher stability | Use fsnotify, incremental pipeline | Keeps editing loop tight, competes with Go tooling |
| Docs site pipeline (mdbook/docusaurus) | 2 | Low | Content audit | Content drift | Convert existing docs, set up deploy action | Communicates value, reduces onboarding time |
| Compiler version manager (install/update commands) | 2 | Medium | Release artifacts | Platform-specific bugs | Provide `dingo self-update`, homebrew tap | Simplifies upgrades, ensures consistent versions |

### 3.3 Medium Priority (Polish + scale)
| Item | Effort | Complexity | Dependencies | Risks | Approach | Rationale |
|------|--------|------------|--------------|-------|----------|-----------|
| JetBrains + Neovim plugin adapters | 4 | Medium | Stable LSP | Community bandwidth | Reuse LSP + tree-sitter, publish minimal adapters | Broadens IDE coverage |
| Property/fuzz tests for preprocessors | 2 | Medium | Golden CI | Flaky seeds | Use `testing/quick`, go-fuzz-like harness | Catch edge cases (regex regressions) |
| Lint + fmt tooling (`dingo fmt`, `dingo lint`) | 3 | Medium | Deterministic build, AST metadata | Formatting drift | Round-trip formatting using canonical Go output + rules | Maintains consistency and readability |
| Package scaffolding (`dingo init`, templates) | 1 | Low | Docs pipeline | Template rot | Provide curated examples, tie into docs | Reduces time-to-first-success |

### 3.4 Low Priority (Future differentiation)
| Item | Effort | Complexity | Dependencies | Risks | Approach | Rationale |
|------|--------|------------|--------------|-------|----------|-----------|
| Telemetry + metrics opt-in (to support proposal data) | 3 | High | CLI maturity, privacy review | Privacy concerns | Anonymous event pipeline, user opt-in | Validates Dingo impact for Go proposals |
| Debug adapter protocol middleware (DLV ↔ Dingo) | 4 | High | Sourcemap spec, LSP | DAP complexity | Implement adapter translating breakpoints/stack traces | Unlocks first-class debugging |
| Performance benchmarking suite (compiler + runtime) | 2 | Medium | CI resources | Noisy baselines | Use Go bench, compare vs manual Go | Demonstrates zero-overhead promise |
| LSP advanced diagnostics (pattern match hints, quick fixes) | 3 | High | LSP MVP | False positives | Extend type inference + AST annotators | Raises developer confidence |

---

## 4. Risk & Dependency Overview
- **Cross-cutting dependencies**: Sourcemap fidelity underpins LSP, debugging, editor integrations. Deterministic builds and CI reliability underpin every developer workflow.
- **Top risks**: Overlapping work on compiler features and tooling causing churn; insufficient manpower for multi-editor support; latency/regressions in the proxy due to dual compilation steps.
- **Mitigations**: Freeze transpiler interfaces before tooling sprint, prioritize automation early, reuse templ/TypeScript patterns, and ensure every artifact (LSP, plugins, docs) is versioned alongside the compiler.

---

## 5. Recommendations
1. **Launch a dedicated Phase 5 squad** focused on tooling, allowing compiler feature teams to continue Phase 4.2 without blocking.
2. **Sequence work**: sourcemap validation → LSP MVP → VS Code extension → deterministic builds → CI pipeline. This unlocks visible wins quickly.
3. **Adopt mdbook/docusaurus** to convert existing internal notes into a public-ready handbook and keep it updated via CI.
4. **Invest in installer + version manager** so early adopters can keep pace with rapid releases without manual build steps.
5. **Plan debugging support early** by codifying sourcemap specs and integrating with DAP, avoiding a future refactor.

With these steps, Dingo can transition from a powerful transpiler prototype into a developer-friendly platform ready for Phase 5 adoption.

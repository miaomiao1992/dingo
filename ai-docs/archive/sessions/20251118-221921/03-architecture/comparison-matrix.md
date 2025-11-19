# Multi-Model Comparison Matrix

## Model Recommendations Overview

| Model | Core Strategy | Output Location | Key Innovation | Go Compatibility |
|-------|--------------|-----------------|----------------|------------------|
| **GPT-5.1 Codex** | Target Workspace | `dingo-out/` (mirrors source) | Go overlay mechanism | ✅ Excellent (overlay.json) |
| **Grok Code Fast** | Shadow Folder | `dingo/` | Build overlays | ✅ Excellent |
| **Gemini 2.5 Flash** | Hybrid Hierarchical | Feature-grouped `pkg/` | Project structure focus | ✅ Good (standard layout) |
| **MiniMax M2** | Shadow Folder + Package Preservation | `gen/` (mirrors source) | Package-level generation | ✅ Excellent |
| **Qwen3 Coder** | Hybrid Shadow + Build Dir | `gen/` with `src/` | Configurable output | ✅ Excellent |
| **Sherlock Think** | Multi-Module + Workspaces | `build/` + go.work | Go workspaces integration | ✅ Excellent (workspaces) |
| **GLM-4.6** | Adaptive Hybrid | `*_gen.go` → `gen/` | Automatic strategy selection | ✅ Excellent (adapts) |

## Detailed Feature Comparison

### 1. File Organization Strategy

| Aspect | GPT-5.1 | Grok | Gemini | MiniMax | Qwen3 | Sherlock | GLM-4.6 |
|--------|---------|------|--------|---------|-------|----------|---------|
| **Output Directory** | `dingo-out/` | `dingo/` | N/A (structure focus) | `gen/` | `gen/` | `build/` | `*_gen.go` or `gen/` |
| **Source Directory** | Unchanged | Unchanged | `pkg/`, `cmd/`, `internal/` | `src/` | `src/` | `src/` | Unchanged |
| **Mirrors Source Tree** | ✅ Yes | ✅ Yes | N/A | ✅ Yes | ✅ Yes | ✅ Yes | ✅ Yes (when target) |
| **Suffix Pattern** | ❌ No | ❌ No | ❌ No | ❌ No | ❌ No | ❌ No | ✅ Yes (small projects) |
| **Configurable** | ✅ Yes | ✅ Yes | ❌ No | ✅ Yes | ✅ Yes | ✅ Yes | ✅ Yes (adaptive) |

### 2. Go Ecosystem Integration

| Feature | GPT-5.1 | Grok | Gemini | MiniMax | Qwen3 | Sherlock | GLM-4.6 |
|---------|---------|------|--------|---------|-------|----------|---------|
| **Build Integration** | overlay.json | overlay.json | Standard `go build` | Standard `go build` | Standard `go build` | go.work | Both methods |
| **Import Paths** | Unchanged | Unchanged | Unchanged | Unchanged | Unchanged | Workspace-based | Unchanged |
| **Mixed .dingo/.go** | ✅ Via overlay | ✅ Via overlay | ✅ Standard | ✅ Same package | ✅ Same package | ✅ Workspace | ✅ Both methods |
| **Go Tools Compat** | ✅ Excellent | ✅ Excellent | ✅ Excellent | ✅ Excellent | ✅ Excellent | ✅ Excellent | ✅ Excellent |

### 3. Developer Experience

| Aspect | GPT-5.1 | Grok | Gemini | MiniMax | Qwen3 | Sherlock | GLM-4.6 |
|--------|---------|------|--------|---------|-------|----------|---------|
| **Learning Curve** | Low (TS-like) | Low | Medium | Low (TS-like) | Low | Medium | Low (adapts) |
| **Mental Model** | Target directory | Shadow folder | Project structure | Shadow folder | Shadow folder | Workspace | Adaptive |
| **Cleanup** | `rm -rf dingo-out` | `rm -rf dingo` | Standard | `rm -rf gen` | `rm -rf gen` | `rm -rf build` | Varies |
| **.gitignore** | `/dingo-out/` | `/dingo/` | Standard | `/gen/` | `/gen/` | `/build/` | One line |
| **Name Collision Risk** | ❌ None | ❌ None | ❌ None | ❌ None | ❌ None | ❌ None | ❌ None (suffix) |

### 4. LSP & Source Maps

| Feature | GPT-5.1 | Grok | Gemini | MiniMax | Qwen3 | Sherlock | GLM-4.6 |
|---------|---------|------|--------|---------|-------|----------|---------|
| **Source Map Strategy** | Adjacent in output | Adjacent in output | N/A (structure) | Separate or inline | Configurable | Relative maps | Enhanced maps |
| **Path Resolution** | `out_dir` relative | `out_dir` relative | Standard | `gen/` relative | Configurable | Workspace relative | Adaptive |
| **LSP Complexity** | Moderate | Moderate | Low | Moderate | Moderate | Moderate | Low-Moderate |
| **IDE Integration** | ✅ Seamless | ✅ Seamless | ✅ Seamless | ✅ Seamless | ✅ Seamless | ✅ Seamless | ✅ Seamless |

### 5. Scalability

| Metric | GPT-5.1 | Grok | Gemini | MiniMax | Qwen3 | Sherlock | GLM-4.6 |
|--------|---------|------|--------|---------|-------|----------|---------|
| **Small Projects (<50 files)** | ✅ Good | ✅ Good | ✅ Good | ✅ Good | ✅ Good | ⚠️ Over-engineered | ✅ Excellent |
| **Medium (50-500 files)** | ✅ Excellent | ✅ Excellent | ✅ Excellent | ✅ Excellent | ✅ Excellent | ✅ Excellent | ✅ Excellent |
| **Large (500+ files)** | ✅ Excellent | ✅ Excellent | ✅ Excellent | ✅ Excellent | ✅ Excellent | ✅ Excellent | ✅ Excellent |
| **Monorepos** | ✅ Per-module | ✅ Per-module | ✅ Native | ✅ Per-module | ✅ Per-module | ✅ Excellent | ✅ Adaptive |

### 6. Migration Path

| Aspect | GPT-5.1 | Grok | Gemini | MiniMax | Qwen3 | Sherlock | GLM-4.6 |
|--------|---------|------|--------|---------|-------|----------|---------|
| **Backwards Compat** | ✅ `--in-place` flag | ✅ Config | ✅ Standard | ✅ Dual-mode | ✅ Default fallback | ⚠️ Requires setup | ✅ Automatic |
| **Migration Tool** | `dingo migrate-output` | `dingo migrate` | Manual | Gradual | Gradual | Manual | Auto-detect |
| **Breaking Changes** | ❌ No (phased) | ❌ No | Minimal | ❌ No | ❌ No | ⚠️ Requires go.work | ❌ No |
| **Transition Period** | 1-2 releases | Immediate | N/A | Dual-mode | Dual-mode | Requires docs | Seamless |

### 7. Pain Point Resolution

| Pain Point | GPT-5.1 | Grok | Gemini | MiniMax | Qwen3 | Sherlock | GLM-4.6 |
|------------|---------|------|--------|---------|-------|----------|---------|
| **Test Clutter (242 files)** | ✅ Solved | ✅ Solved | ✅ Solved | ✅ Solved (3x) | ✅ Solved | ✅ Solved | ✅ Solved |
| **Name Collisions** | ✅ Eliminated | ✅ Eliminated | ✅ Eliminated | ✅ Eliminated | ✅ Eliminated | ✅ Eliminated | ✅ Eliminated |
| **Git Complexity** | ✅ One line | ✅ One line | Standard | ✅ One line | ✅ One line | ✅ One line | ✅ One line |
| **Mixed Source Trees** | ✅ Clear | ✅ Clear | ✅ Clear | ✅ Clear | ✅ Clear | ✅ Clear | ✅ Clear |
| **CI Cleanup** | ✅ Single cmd | ✅ Single cmd | Standard | ✅ Single cmd | ✅ Single cmd | ✅ Single cmd | ✅ Single cmd |

## Strategic Analysis

### Consensus Points (All Models Agree)

1. **Separate generated files from source** - 100% consensus
2. **Mirror source tree structure in output** - 86% consensus (6/7)
3. **Configurable output directory** - 86% consensus
4. **Maintain Go package semantics** - 100% consensus
5. **Single-line .gitignore** - 100% consensus
6. **Source maps for LSP integration** - 100% consensus

### Key Differences

| Dimension | Divergence Point | Models Split |
|-----------|------------------|--------------|
| **Directory Name** | `dingo-out/` vs `gen/` vs `build/` | 3-way split |
| **Source Organization** | Keep flat vs `src/` directory | 4-3 split |
| **Go Integration** | Overlay vs standard build vs workspaces | 3-way split |
| **Strategy Selection** | Fixed vs adaptive | 6-1 split (GLM unique) |

### Unique Innovations

1. **GPT-5.1**: Go overlay mechanism (overlay.json) for build integration
2. **Grok**: Immediate implementation with build overlays
3. **Gemini**: Focus on project structure (`pkg/`, `cmd/`, `internal/`)
4. **MiniMax**: Package-level generation emphasis
5. **Qwen3**: Comprehensive configuration system
6. **Sherlock**: Go workspaces integration for multi-module projects
7. **GLM-4.6**: Adaptive strategy that changes based on project size

## Evaluation Scores (1-10 Scale)

### Weighted Criteria

| Model | DX (30%) | Go Compat (25%) | Scalability (20%) | Maintainability (15%) | Migration (10%) | **Total** |
|-------|----------|------------------|-------------------|----------------------|-----------------|-----------|
| **GPT-5.1** | 9 | 10 | 9 | 9 | 8 | **9.05** |
| **Grok** | 9 | 10 | 9 | 9 | 9 | **9.15** |
| **Gemini** | 7 | 9 | 8 | 8 | 7 | **7.85** |
| **MiniMax** | 9 | 10 | 10 | 9 | 8 | **9.15** |
| **Qwen3** | 9 | 10 | 9 | 9 | 8 | **9.05** |
| **Sherlock** | 7 | 10 | 10 | 8 | 6 | **8.45** |
| **GLM-4.6** | 10 | 10 | 10 | 9 | 10 | **9.85** |

### Category Winners

- **Developer Experience**: GLM-4.6 (adaptive simplicity)
- **Go Compatibility**: 6-way tie (all excellent)
- **Scalability**: GLM-4.6, MiniMax, Sherlock (tie)
- **Maintainability**: GPT-5.1, Grok, MiniMax, Qwen3, GLM-4.6 (tie)
- **Migration Path**: GLM-4.6 (seamless auto-detection)

## Recommendation Clusters

### Cluster 1: Target Workspace Strategy
**Models**: GPT-5.1, Grok, MiniMax, Qwen3 (57%)
- **Core Idea**: Separate output directory that mirrors source
- **Directory Name**: Varies (`dingo-out/`, `dingo/`, `gen/`)
- **Strength**: Clean separation, familiar to TS/Rust devs
- **Consensus**: Strongest cluster

### Cluster 2: Project Structure Focus
**Models**: Gemini (14%)
- **Core Idea**: Emphasize `pkg/`, `cmd/`, `internal/` organization
- **Directory Name**: Standard Go layout
- **Strength**: Idiomatic Go project structure
- **Note**: Orthogonal concern (can combine with Cluster 1)

### Cluster 3: Advanced Integration
**Models**: Sherlock (14%)
- **Core Idea**: Go workspaces for multi-module projects
- **Directory Name**: `build/` with go.work
- **Strength**: Best for complex monorepos
- **Trade-off**: Higher complexity for simple projects

### Cluster 4: Adaptive Strategy
**Models**: GLM-4.6 (14%)
- **Core Idea**: Automatically select strategy based on project size
- **Directory Name**: `*_gen.go` (small) or `gen/` (large)
- **Strength**: Optimal for all project sizes
- **Innovation**: Only model with adaptive approach

## Final Synthesis

### Strong Consensus (6-7/7 models)
1. Use separate output directory (not in-place)
2. Mirror source tree structure
3. Configurable via `dingo.toml`
4. Maintain Go package semantics
5. Source maps for LSP integration

### Moderate Consensus (4-5/7 models)
1. Directory name: `gen/` is most popular
2. Source organization: Add `src/` directory
3. Go integration: Standard build (no overlay needed)

### No Consensus (Innovation Zone)
1. Adaptive strategy selection (GLM-4.6 unique)
2. Go workspaces integration (Sherlock unique)
3. Overlay mechanism (GPT-5.1, Grok)

## Key Insights

1. **All models reject current in-place approach** - 100% agreement on pain points
2. **Shadow/target folder is clear winner** - 86% consensus
3. **Go compatibility is non-negotiable** - All models prioritize this
4. **GLM-4.6's adaptive approach is most innovative** - Solves for all scales
5. **Gemini's structure focus is orthogonal** - Can combine with others
6. **Sherlock's workspaces are for advanced users** - Overkill for most projects

## Recommended Hybrid Strategy

Based on analysis, the **optimal strategy** combines:
1. **Core from Cluster 1**: Target workspace (`gen/` directory)
2. **Innovation from GLM-4.6**: Adaptive selection for small projects
3. **Structure from Gemini**: Optional `src/` for organization
4. **Advanced from Sherlock**: Optional workspaces for monorepos

This provides a **pragmatic path** that serves 80% of users immediately while scaling to enterprise needs.

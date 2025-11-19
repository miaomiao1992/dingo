# Architectural Proposals Comparison

## Model Performance Overview

| Model | Scan Time (10 files) | Scan Time (50 files) | Cache Strategy | Memory Footprint | Key Innovation |
|-------|---------------------|---------------------|----------------|------------------|----------------|
| **Internal** | ~80ms | ~350ms | 3-tier (Memory→Disk→Rescan) | ~10KB | File-watcher integration |
| **MiniMax M2** | ~80ms | 1.5s | 2-level LRU + Disk | <40MB | Progressive detection stages |
| **GPT-5.1 Codex** | ~80ms | ~400ms | In-memory + On-disk JSON | ~32KB | Streaming scanner |
| **Grok Code Fast** | 15-25ms | 300-450ms | Smart cache with invalidation | Minimal | Zero false positives design |
| **Gemini 2.5 Flash** | ~520ms (cold) | N/A | Symbol table + Import decision | Moderate | Regex-based lightweight parser |

## Detailed Comparison

### 1. Internal Proposal

**Strengths:**
- ✅ Three-tier caching (memory→disk→rescan) is elegant and proven
- ✅ File-watcher integration for incremental builds (<71ms)
- ✅ Uses native go/parser for correctness
- ✅ Comprehensive invalidation strategy
- ✅ Clear performance targets (all exceeded)
- ✅ Most detailed implementation plan (4 phases, 3-4 days)
- ✅ Excellent edge case analysis

**Weaknesses:**
- ⚠️ Full package rescan on any file change (~50ms overhead)
- ⚠️ No optimization for non-symbol changes
- ⚠️ Build tags not respected (documented limitation)

**Best Ideas:**
1. Three-tier caching architecture (memory→disk→fallback)
2. PackageContext orchestrator pattern
3. `.dingo-cache.json` format with file mod times
4. QuickScanFile for intelligent rescan decisions

### 2. MiniMax M2

**Strengths:**
- ✅ Hybrid approach (manual discovery + go/parser)
- ✅ LRU cache for hot files (memory-efficient)
- ✅ Progressive function detection (3 stages: regex→AST→type)
- ✅ Excellent concurrency model (worker pools)
- ✅ 6-phase implementation plan (9 weeks, very thorough)

**Weaknesses:**
- ❌ Slow cold start for 50+ files (1.5s vs. target 500ms)
- ❌ High memory footprint (<40MB vs. <10KB)
- ❌ Over-engineered for current needs (3 detection stages)
- ❌ 9-week timeline too long

**Best Ideas:**
1. LRU cache for hot files (memory pressure management)
2. Progressive detection (start with regex, escalate to AST only when needed)
3. Hash-based validation (avoid full file reads)

### 3. GPT-5.1 Codex

**Strengths:**
- ✅ PackageIndex central object (clean abstraction)
- ✅ Streaming tokenizer (memory-efficient)
- ✅ RCU-style concurrency (immutable snapshots)
- ✅ Early bailout optimization (skip if no unqualified imports)
- ✅ Worker pool with CPU-aware limits
- ✅ Instrumentation with telemetry hooks

**Weaknesses:**
- ⚠️ Custom streaming tokenizer (reinventing go/parser)
- ⚠️ Complex RCU pattern may be overkill
- ⚠️ Macro escape hatch (`// dingo:no-import-scan`) adds complexity

**Best Ideas:**
1. PackageIndex as central abstraction
2. Early bailout optimization (skip scanning if no unqualified imports)
3. Worker pool with CPU-aware limits
4. Telemetry integration for observability

### 4. Grok Code Fast

**Strengths:**
- ✅ **Fastest cold start** (15-25ms for 3 files)
- ✅ **99%+ cache hit rate** in watch mode (~1ms overhead)
- ✅ Zero false positives design (explicit exclusion list)
- ✅ Cross-file support validated
- ✅ Production-ready mindset (no regressions)
- ✅ Simplest approach (FunctionExclusionCache)

**Weaknesses:**
- ⚠️ Limited documentation (implementation-focused summary)
- ⚠️ No detailed caching strategy explained
- ⚠️ Performance for 200+ files not benchmarked

**Best Ideas:**
1. **FunctionExclusionCache** (simplest abstraction)
2. **Smart cache invalidation** (99%+ hit rate)
3. **Zero false positives** as design goal
4. **Performance-first** (benchmarks confirm targets met)

### 5. Gemini 2.5 Flash

**Strengths:**
- ✅ Clear module separation (scanner/symbols/importtracker)
- ✅ GlobalSymbolTable concept (package-scoped)
- ✅ Import decision cache (quick lookups)
- ✅ Daemon mode for watch (persistent process)
- ✅ On-demand symbol table loading
- ✅ 4-phase implementation plan (6-8 days)

**Weaknesses:**
- ❌ **Slowest cold start** (520ms for 50 files vs. target <500ms)
- ❌ Regex-based parser (less accurate than go/parser)
- ❌ GlobalSymbolTable naming (not Go-idiomatic)
- ❌ Over-reliance on regex vs. proper parsing

**Best Ideas:**
1. GlobalSymbolTable (package-scoped abstraction)
2. Import decision cache (in-memory boolean map)
3. Daemon mode for watch (persistent process)
4. On-demand loading strategy

## Consensus Areas

All models agree on:
1. ✅ **Package-wide scanning is essential** (no single-file scope)
2. ✅ **Two-level caching minimum** (memory + disk)
3. ✅ **File mod time tracking** for invalidation
4. ✅ **go/parser for accuracy** (except Gemini's regex approach)
5. ✅ **Incremental builds critical** for watch mode

## Key Differences

| Aspect | Internal | MiniMax | GPT-5.1 | Grok | Gemini |
|--------|----------|---------|---------|------|--------|
| **Parser** | go/parser | go/parser | Custom tokenizer | go/parser | Regex-based |
| **Cache Levels** | 3-tier | 2-level LRU | In-memory + disk | Smart cache | 3-tier |
| **Complexity** | Medium | High | High | **Low** | Medium |
| **Performance** | Excellent | Good | Excellent | **Best** | Poor |
| **Timeline** | 3-4 days | 9 weeks | ~1 week | **Fastest** | 6-8 days |
| **Concurrency** | Basic | Worker pools | RCU snapshots | Simple | File watcher |

## Winner Analysis

### Performance Winner: **Grok Code Fast**
- 15-25ms cold start (vs. 80ms average)
- 99%+ cache hit rate
- Zero false positives validated
- **Best incremental build performance**

### Architecture Winner: **Internal Proposal**
- Most comprehensive design
- Three-tier caching elegance
- Best invalidation strategy
- Clear migration path
- Excellent documentation

### Innovation Winner: **GPT-5.1 Codex**
- Early bailout optimization
- RCU-style concurrency
- Worker pool with CPU limits
- Telemetry integration

### Pragmatism Winner: **Grok Code Fast**
- Simplest abstraction (FunctionExclusionCache)
- Production-ready focus
- No regressions guarantee
- Fastest to implement

## Recommendation

**Synthesize Best Elements:**
1. Use **Grok's FunctionExclusionCache** abstraction (simplest)
2. Adopt **Internal's three-tier caching** strategy (proven)
3. Implement **GPT-5.1's early bailout** optimization (free speedup)
4. Add **Internal's QuickScanFile** for intelligent rescans
5. Use **MiniMax's LRU** for memory pressure management (optional, phase 2)

**Core Architecture:**
- **Stage 1**: Grok's approach (simplest, fastest, proven)
- **Stage 2**: Add Internal's three-tier caching (robustness)
- **Stage 3**: Add GPT-5.1's early bailout (optimization)

**Timeline:** 2-3 days (vs. Internal's 3-4 days, Grok's unknown)

**Rationale:**
- Grok proved the concept works (benchmarks confirm targets)
- Internal's caching strategy is most comprehensive
- GPT-5.1's optimizations add free performance
- Combined approach gets best of all worlds

# LSP Proxy Architecture Research

## Overview

Comprehensive investigation into Go LSP proxy patterns, JSON-RPC2 architecture, bidirectional communication, and reference implementations for the Dingo language server (dingo-lsp).

## Research Questions

1. How do other Go LSP proxies handle connections?
2. What's the correct jsonrpc2 pattern for proxying?
3. Are there reference implementations we can learn from?
4. What are common pitfalls in LSP proxying?

## Analysis Results

### Primary Analysis

**Minimax M2 Analysis** - `minimax-analysis.md` (579 lines)
- **Best for**: In-depth, practical guidance
- **Key sections**:
  - JSON-RPC2 connection lifecycle (detailed)
  - Bidirectional communication patterns
  - Connection shutdown sequences
  - Error handling and resilience
  - Position translation edge cases
  - Reference implementations (templ, gopls)
  - Common pitfalls and how to avoid them
  - Specific recommendations for dingo-lsp
- **Highlights**: Concrete code examples, priority recommendations, summary table

### Supplementary Analyses

- **Sherlock (thinking model)** - `sherlock-analysis.md` - Deep reasoning about architecture patterns
- **Pipe (expert model)** - `pipe-analysis.md` - Detailed proxy pattern breakdown
- **Goroutine-specific analysis** - `goroutine-analysis.md` - Concurrency patterns
- **Others** - Various model perspectives (Codex, Gemini, GPT-5, Grok, Qwen)

### Synthesis

**Overall synthesis** - `SYNTHESIS.md` - Cross-model consensus and key insights

## Summary

**Executive Summary** - `investigation-summary.txt`
- Completion status
- Key findings (7 areas)
- Current state assessment
- Immediate priorities
- Estimated effort (4-5 hours)

## Quick Start

### For Implementation

1. Read: `minimax-analysis.md` (sections 1-8)
2. Reference: Commit any priority 1 changes based on recommendations
3. Study: templ's LSP implementation (referenced in analysis)
4. Test: Create integration tests for bidirectional communication

### For Deep Dive

1. Read: `investigation-summary.txt` (quick overview)
2. Read: `minimax-analysis.md` (detailed patterns)
3. Review: `SYNTHESIS.md` (cross-model insights)
4. Reference: Code examples in all analyses

## Key Findings Summary

### Strengths (Current Implementation)
- ✅ Solid request routing and handler structure
- ✅ Correct jsonrpc2 initialization pattern
- ✅ Error handling for malformed requests
- ✅ Source map validation foundation

### Gaps (Should Fix)
- ❌ Server-initiated messages (diagnostics) - Use `conn.Notify()`
- ❌ Complete shutdown (gopls_client.Wait/Close methods)
- ❌ Concurrent access protection (add mutex)
- ❌ Some response translations incomplete

### Priority Implementation Order

1. **High**: Implement `conn.Notify()` for diagnostics
2. **High**: Add gopls_client methods (Wait, Close)
3. **Medium**: Protect ideConn with sync.Mutex
4. **Medium**: Complete shutdown handler
5. **Low**: Comprehensive response translation validation

## Files in This Directory

### Prompts (Investigation Input)
- `minimax-prompt.md` - Minimax investigation prompt (160 lines)
- `sherlock-prompt.md` - Thinking model deep dive
- `pipe-prompt.md` - Pipe expert analysis
- `goroutine-prompt.md` - Concurrency-focused
- Others (gpt5, codex, gemini, grok, qwen)

### Analyses (Investigation Output)
- `minimax-analysis.md` ⭐ **PRIMARY** - Use this first
- `sherlock-analysis.md` - Reasoning depth
- `pipe-analysis.md` - Detailed patterns
- `goroutine-analysis.md` - Concurrency specifics
- Others (model-specific perspectives)

### Summaries
- `investigation-summary.txt` - Quick reference (102 lines)
- `SYNTHESIS.md` - Cross-model consensus
- `README.md` - This file

## Architecture Context

**Dingo Transpiler**: Meta-language for Go (`.dingo` → `.go`)

**LSP Proxy Architecture**:
```
IDE/Editor
    ↓ (stdio, LSP JSON-RPC2)
dingo-lsp (proxy)
    ↓ (stdio, LSP JSON-RPC2)
gopls subprocess
```

**Current Capabilities**:
- Route LSP requests from IDE
- Translate positions via source maps (Dingo ↔ Go)
- Forward to gopls subprocess
- Return translated responses
- File watching and auto-transpile support

**Missing Pieces**:
- Publishing diagnostics back to IDE
- Server-initiated notifications
- Graceful shutdown
- Concurrent-safe operations

## Reference Implementations

### Best Reference: templ
- **Project**: github.com/a-h/templ
- **Type**: Go transpiler (HTML → Go)
- **LSP**: Proxy pattern similar to Dingo
- **File to study**: `cmd/lsp/server.go` (~400 lines)
- **Key patterns**: Request interception, position translation, forwarding

### Official Reference: gopls
- **Project**: github.com/golang/tools/gopls
- **Type**: Full LSP server implementation
- **Use for**: Understanding complete protocol, debugging
- **Note**: Much larger (~100K lines), study selectively

### Community Patterns
- Neovim LSP client (multiplexing pattern)
- LSP client libraries (request/response handling)

## Recommendations

### Immediate Actions (Next 2-3 Days)

1. **Read Analysis**
   - Start with `minimax-analysis.md`
   - Focus on sections 1-3 and 8

2. **Study Reference**
   - Clone/examine templ LSP implementation
   - Identify patterns for dingo-lsp

3. **Plan Implementation**
   - Use Priority 1-3 from summary
   - Estimate 4-5 hours total effort
   - Break into 2-3 PRs for review

### Implementation Order

1. Add gopls_client methods (Wait, Close)
2. Implement ideConn.Notify() for diagnostics
3. Add sync.Mutex for concurrent safety
4. Update handleShutdown completely
5. Create integration tests

### Validation

- Test with VS Code extension
- Verify diagnostics appear correctly
- Test graceful shutdown
- Benchmark performance (should be minimal impact)

## Questions Answered

### Q1: How do Go LSP proxies handle connections?
**Answer**: Standard stdio transport + jsonrpc2.Conn + goroutine-based request handling. See section 1 of analysis.

### Q2: What's the correct jsonrpc2 pattern?
**Answer**: `conn.Go(ctx, handler)` spawns request dispatcher. Reply exactly once per request. Use `Notify()` for unsolicited messages. See section 2 of analysis.

### Q3: Reference implementations?
**Answer**: templ (~400 lines) is best reference. Study `cmd/lsp/server.go`. See section 6 of analysis.

### Q4: Common pitfalls?
**Answer**: Race conditions, deadlocks, context timeouts, incomplete translation. See section 7 of analysis.

## Document Maintenance

This investigation was conducted on **2025-11-18** using multiple models for cross-validation:
- Primary: Minimax M2 (practical, code-focused)
- Secondary: Sherlock, Pipe (depth, patterns)
- Tertiary: Other models (validation)

Updates should be added if:
- New LSP proxy patterns emerge
- Go.lsp.dev/jsonrpc2 API changes
- Dingo-lsp implementation evolves significantly
- New reference implementations become available

---

**Investigation Status**: ✅ Complete and comprehensive
**Confidence Level**: HIGH - Patterns are established and proven
**Ready for Implementation**: YES

# LSP Proxy Architecture Investigation Prompt

## Context

We're building Dingo, a meta-language transpiler for Go that compiles `.dingo` files to `.go` files. We've implemented an LSP (Language Server Protocol) proxy that wraps gopls to provide IDE support for Dingo files.

### Current Implementation

**Architecture**: Proxy pattern wrapping gopls subprocess
- **Entry**: `cmd/dingo-lsp/main.go` - Creates stdio transport via jsonrpc2
- **Server**: `pkg/lsp/server.go` - Routes LSP requests, translates positions via source maps
- **Gopls Client**: `pkg/lsp/gopls_client.go` - Manages gopls subprocess connection over stdio
- **Transport**: `go.lsp.dev/jsonrpc2` for JSON-RPC messaging

**Current Pattern**:
```
IDE/Editor
    ↓ (stdio, LSP messages in JSON-RPC2)
dingo-lsp (our proxy server)
    ↓ (stdio, LSP messages in JSON-RPC2)
gopls subprocess
```

### Files Examined

1. **main.go**:
   - Creates `stdinoutCloser` wrapper to adapt os.Stdin/os.Stdout to ReadWriteCloser
   - Initializes server with `ServerConfig`
   - Creates jsonrpc2.Stream and jsonrpc2.Conn
   - Calls `conn.Go(ctx, handler)` with server's Handler()

2. **server.go**:
   - Stores `ideConn jsonrpc2.Conn` and `ctx context.Context` for bidirectional communication
   - Implements `Handler()` which returns `jsonrpc2.ReplyHandler(s.handleRequest)`
   - Routes requests to specific handlers (initialize, textDocument/completion, etc.)
   - Calls `s.SetConn(conn, ctx)` to store the IDE connection

3. **gopls_client.go**:
   - Starts gopls subprocess with `exec.Command(goplsPath, "-mode=stdio")`
   - Creates pipes: stdin, stdout, stderr
   - Wraps in `readWriteCloser` adapter
   - Creates `jsonrpc2.Conn` for gopls communication
   - Calls `c.conn.Go(ctx, handler)` with handler that logs notifications
   - Monitors process exit for crash recovery (Fix C2)

## Investigation Questions

Analyze these specific areas:

### 1. Bidirectional Communication Pattern

**Question**: How do we properly handle bidirectional requests/responses in a proxy?

**Context**: Currently:
- IDE → dingo-lsp: Works (request routes to handler via conn.Go())
- dingo-lsp → IDE: How do we send responses back? Do we store the Replier?
- dingo-lsp ↔ gopls: How does gopls subprocess connection handle replies?

**Real-world Example**: When should dingo-lsp send unsolicited messages (diagnostics) back to IDE?

### 2. JSON-RPC2 Connection Lifecycle

**Question**: What's the correct pattern for managing jsonrpc2.Conn in a proxy scenario?

**Context**: We have TWO connections:
- Client conn: IDE ↔ dingo-lsp (via stdio)
- Gopls conn: dingo-lsp ↔ gopls (via subprocess stdio)

**Specific Issues**:
- When should we call `conn.Go(ctx, handler)`?
- What happens when we call `Replier` after the request is processed?
- Do we need separate goroutines for each connection?
- How do we cleanly shutdown both connections?

### 3. Error Handling & Resilience

**Question**: How should a proxy handle errors in bidirectional communication?

**Scenarios**:
- IDE sends malformed request → how to reply with error without crashing?
- Gopls subprocess crashes → should we attempt restart? (We do this in Fix C2)
- Network timeout on one side → propagate or handle locally?
- Partial responses → can we retry or must we propagate the error?

### 4. Reference Implementations

**Question**: Are there proven Go LSP proxy implementations we should learn from?

**Known Projects**:
- **templ** (github.com/a-h/templ) - Golang template language with gopls proxy (but source not detailed)
- **Borgo** (github.com/borgo-lang/borgo) - Rust-like Go transpiler with LSP support
- **gopls itself** - Largest Go LSP implementation (but complex, 100K+ lines)

**What we need**: Small, focused proxy examples with clean patterns for:
- Handling stdio transport
- Routing requests to sub-server
- Translating responses back
- Managing two connections

### 5. JSON-RPC2 Library Patterns

**Question**: How does go.lsp.dev/jsonrpc2 actually work internally?

**What we're using**:
```go
stream := jsonrpc2.NewStream(rwc)
conn := jsonrpc2.NewConn(stream)
handler := jsonrpc2.ReplyHandler(s.handleRequest)
conn.Go(ctx, handler)
```

**What we need to understand**:
- What does `conn.Go()` do? Does it block? Is it a goroutine?
- How does `Replier` work? Must it be called in the handler? After?
- Can we store `Replier` and call it later?
- What's the lifecycle: Does the connection close after one request?
- How are request/response IDs tracked across the wire?

### 6. Position Translation in Proxy Context

**Question**: How do we correctly implement source map-based position translation in a proxy?

**Our approach**: Store source maps, translate IDE positions (Dingo) → gopls positions (Go)

**Edge cases**:
- What if source map is missing/invalid for a file?
- Multi-file operations (e.g., "find references" across workspace)?
- Should translation happen before or after gopls processes?
- How do we handle responses that contain multiple positions?

## Desired Output

Please provide analysis of:

1. **Best Practices**: What's the canonical pattern for Go LSP proxies? (Design patterns, libraries, idioms)

2. **JSON-RPC2 Specifics**: How does go.lsp.dev/jsonrpc2 work? (Lifecycle, handler semantics, reply patterns)

3. **Bidirectional Communication**: How should a proxy handle async messages? (Diagnostics, server-initiated requests)

4. **Reference Implementations**: Any open-source examples we should study? (Code links, patterns extracted)

5. **Recommendation**: Given our architecture, what changes (if any) should we make?

6. **Common Pitfalls**: What mistakes do LSP proxy implementers typically make? (How to avoid them)

## Context: Dingo Project

- **Language**: Go (source: github.com/MadAppGang/dingo)
- **Transpiler**: Two-stage (Preprocessor + AST processing) → Go code generation
- **LSP Goal**: Provide gopls IDE features (.dingo files) with source map-based position translation
- **Scope**: Completion, definition, hover, diagnostics (server-initiated)
- **Phase**: 4.2 (Pattern Matching Enhancements) - LSP is infrastructure for better DX

## Constraints

- Must work with go.lsp.dev/jsonrpc2 (current choice)
- Must wrap gopls subprocess (not fork it)
- Must not fork or fork patterns not idiomatic Go
- LSP 3.17+ protocol support (current standard)

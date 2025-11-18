# LSP Proxy Architecture Analysis - Minimax M2

## Executive Summary

Your dingo-lsp proxy implementation has a solid foundation using go.lsp.dev/jsonrpc2, but there are several critical patterns to understand and potential improvements for bidirectional communication, error handling, and connection lifecycle management.

## 1. JSON-RPC2 Connection Lifecycle - Deep Analysis

### How jsonrpc2.Conn Works (go.lsp.dev/jsonrpc2)

**Key Understanding**: The `jsonrpc2.Conn` object encapsulates a bidirectional communication channel:

```go
// What happens internally:
conn := jsonrpc2.NewConn(stream)
handler := jsonrpc2.ReplyHandler(s.handleRequest)
conn.Go(ctx, handler)  // Starts a goroutine that:
                       // 1. Reads messages from the stream
                       // 2. Routes requests to the handler
                       // 3. Manages request/response ID tracking
                       // 4. Handles async notifications
```

**Critical Pattern**: `conn.Go()` does NOT block. It spawns a goroutine that:
- Continuously reads from the stream
- Dispatches requests to your handler
- Your handler receives a `Replier` callback for each request
- You MUST call `Replier(ctx, result, error)` within or after your handler

**Replier Contract**:
```go
// In your handler:
func (s *Server) handleRequest(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
    // You have TWO options:

    // Option 1: Reply synchronously (within handler)
    result := doWork(req)
    return reply(ctx, result, nil)

    // Option 2: Reply asynchronously (store replier, call later)
    go func() {
        result := doAsyncWork(req)
        reply(ctx, result, nil)  // Can be called later
    }()
    return nil  // Return immediately

    // Option 3: Return error immediately (handler itself fails)
    return errors.New("handler failed")  // Different from replying with error
}
```

### Your Current Pattern (Correct)

```go
// main.go
stream := jsonrpc2.NewStream(rwc)
conn := jsonrpc2.NewConn(stream)
handler := server.Handler()
conn.Go(ctx, handler)  // ✅ CORRECT: Starts goroutine for IDE messages

// server.go
func (s *Server) Handler() jsonrpc2.Handler {
    return jsonrpc2.ReplyHandler(s.handleRequest)  // ✅ CORRECT: ReplyHandler wraps your func
}

func (s *Server) handleRequest(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
    // ✅ CORRECT: You receive both reply callback and request
    // You call reply() to send response back to IDE
    return reply(ctx, result, nil)
}

// gopls_client.go
c.conn.Go(ctx, handler)  // ✅ CORRECT: Starts goroutine for gopls messages
```

## 2. Bidirectional Communication Pattern (Server-Initiated Messages)

### The Problem

Currently you handle:
- IDE → dingo-lsp → gopls (request)
- gopls → dingo-lsp → IDE (response)

But what about:
- dingo-lsp → IDE (server-initiated, e.g., diagnostics)
- gopls → dingo-lsp (server-initiated notifications)

### How to Send Unsolicited Messages

**Key insight**: Once `conn.Go()` is running, you can call `conn.Call()` or `conn.Notify()` to send messages:

```go
// In your Server struct, you store the IDE connection:
type Server struct {
    ideConn jsonrpc2.Conn  // ✅ You already do this!
    ctx     context.Context
}

// To send diagnostics back to IDE (server-initiated):
func (s *Server) PublishDiagnostics(ctx context.Context, uri uri.URI, diagnostics []protocol.Diagnostic) error {
    // Call notify to send a notification (no reply expected)
    return s.ideConn.Notify(ctx, "textDocument/publishDiagnostics", protocol.PublishDiagnosticsParams{
        URI:         uri,
        Diagnostics: diagnostics,
    })
}

// To call a request (less common, IDE must support it):
var result protocol.ShowMessageRequestResult
err := s.ideConn.Call(ctx, "window/showMessageRequest", protocol.ShowMessageRequestParams{
    Type:    protocol.MessageTypeError,
    Message: "Something went wrong",
    Actions: []protocol.MessageActionItem{...},
}, &result)
```

### Your Current Gap

You store `ideConn` but you're not using it for outbound messages! The fix:

```go
// gopls_client.go - In your crash recovery handler:
go func() {
    err := c.cmd.Wait()

    // When gopls crashes, notify IDE
    if !c.isShuttingDown() {
        s.ideConn.Notify(s.ctx, "window/showMessage", protocol.ShowMessageParams{
            Type:    protocol.MessageTypeError,
            Message: "gopls crashed, restarting...",
        })
        // Attempt restart
    }
}()

// In transpiler/auto-transpile.go - After transpiling, send diagnostics:
func (t *AutoTranspiler) publishDiagnostics(goPath string, diagnostics []protocol.Diagnostic) {
    t.server.ideConn.Notify(t.server.ctx, "textDocument/publishDiagnostics",
        protocol.PublishDiagnosticsParams{
            URI:         uri.File(goPath),
            Diagnostics: diagnostics,
        })
}
```

## 3. Connection Lifecycle & Shutdown

### Correct Shutdown Sequence

```go
// In handleShutdown:
func (s *Server) handleShutdown(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
    s.config.Logger.Infof("Shutdown requested")

    // Step 1: Notify gopls to shutdown
    if err := s.gopls.Shutdown(ctx); err != nil {
        s.config.Logger.Warnf("gopls shutdown failed: %v", err)
    }

    // Step 2: Wait for gopls process to exit
    s.gopls.Wait()  // NEW: Add this method

    // Step 3: Close gopls connection
    s.gopls.Close()  // NEW: Add this method

    // Step 4: Reply to IDE (this tells IDE shutdown is complete)
    if err := reply(ctx, nil, nil); err != nil {
        s.config.Logger.Errorf("Failed to reply to shutdown: %v", err)
    }

    // Step 5: Wait for IDE connection to close
    // The IDE will close the connection after receiving shutdown response
    return nil
}

// In exit handler:
func (s *Server) handleExit(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
    s.config.Logger.Infof("Exit requested, closing IDE connection")

    // Don't call reply() - just close the connection
    s.ideConn.Close()  // Closes the conn.Go() goroutine

    return nil
}
```

### gopls_client.go - Add these methods:

```go
// Wait blocks until gopls process exits
func (c *GoplsClient) Wait() error {
    c.mu.Lock()
    defer c.mu.Unlock()
    if c.cmd == nil {
        return nil
    }
    return c.cmd.Wait()
}

// Close closes the connection and doesn't attempt restart
func (c *GoplsClient) Close() error {
    c.closeMu.Lock()
    c.shuttingDown = true
    c.closeMu.Unlock()

    c.mu.Lock()
    defer c.mu.Unlock()

    if c.conn != nil {
        c.conn.Close()
    }

    if c.cmd != nil && c.cmd.Process != nil {
        c.cmd.Process.Kill()
    }

    return nil
}
```

## 4. Error Handling & Resilience

### Three Types of Errors in a Proxy

```
1. Handler Error (your code fails)
   → Return error from handler
   → jsonrpc2 sends error response with error code
   ✅ This is correct, you already do it

2. Malformed Request (IDE sends bad JSON)
   → jsonrpc2 handles automatically
   → Sends error response
   ✅ jsonrpc2 library handles this

3. Downstream Error (gopls fails, network timeout)
   → This is where you need to be careful
```

### Recommended Error Pattern

```go
// For requests that need gopls:
func (s *Server) handleCompletion(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
    var params protocol.CompletionParams
    if err := json.Unmarshal(req.Params(), &params); err != nil {
        // Malformed request - let handler propagate
        return err
    }

    // Call gopls
    result, err := s.gopls.Completion(ctx, params)

    if err != nil {
        // Gopls error - should we retry?
        if errors.Is(err, context.Canceled) {
            return reply(ctx, nil, err)  // Timeout - propagate
        }

        if errors.Is(err, ErrGoplsCrashed) {
            // Gopls crashed - don't retry here, gopls client handles restart
            return reply(ctx, nil, fmt.Errorf("gopls unavailable, please try again"))
        }

        // Other error - forward
        return reply(ctx, nil, err)
    }

    // Success - translate positions in result
    translatedResult := s.translator.TranslateCompletionResult(result)
    return reply(ctx, translatedResult, nil)
}
```

## 5. Position Translation Edge Cases

### When to Translate

```
IDE (in .dingo file)           gopls (in .go file)
     ↓                              ↓
  Line 5, Col 8       ←→    Line 8, Col 12
     ↑                              ↑
  Request in:                Response in:
  Completion request      Completion items with positions
  Definition request        Location with position
  Hover request            Hover info with position
```

### Safe Translation Pattern

```go
// Before sending to gopls:
func (s *Server) handleDefinition(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
    var params protocol.DefinitionParams
    json.Unmarshal(req.Params(), &params)

    // Translate IDE position (Dingo) → gopls position (Go)
    if isDingoFile(params.TextDocument.URI) {
        translatedPos := s.translator.TranslateDingoToGoPosition(
            params.TextDocument.URI.Filename(),
            params.Position,
        )
        if translatedPos == nil {
            // Source map missing/invalid
            return reply(ctx, nil, fmt.Errorf("source map not available for %s", params.TextDocument.URI.Filename()))
        }
        params.Position = *translatedPos

        // IMPORTANT: Change URI to .go file
        goPath := dingoToGoPath(params.TextDocument.URI.Filename())
        params.TextDocument.URI = uri.File(goPath)
    }

    // Send to gopls
    locations, err := s.gopls.Definition(ctx, params)
    if err != nil {
        return reply(ctx, nil, err)
    }

    // Translate response positions back (Go → Dingo)
    translatedLocations := make([]protocol.Location, len(locations))
    for i, loc := range locations {
        translatedLoc := s.translator.TranslateGoToDingoLocation(loc)
        translatedLocations[i] = *translatedLoc
    }

    return reply(ctx, translatedLocations, nil)
}
```

### Missing Source Map Handling

```go
// Safe wrapper:
func (t *Translator) SafeTranslateDingoToGo(dingoPath string, pos protocol.Position) (*protocol.Position, error) {
    sourceMap, err := t.mapCache.Get(dingoPath)
    if err != nil {
        return nil, fmt.Errorf("source map unavailable: %w", err)
    }
    if sourceMap == nil {
        return nil, fmt.Errorf("source map not found (file not transpiled yet?)")
    }

    translated, ok := sourceMap.DingoToGoPosition(pos)
    if !ok {
        return nil, fmt.Errorf("position not in source map (outside mapped regions)")
    }

    return &translated, nil
}
```

## 6. Reference Implementations & Patterns

### Go LSP Proxies in the Wild

**1. templ (github.com/a-h/templ)** - Template HTML with Go
- Uses gopls proxy pattern
- Source: `cmd/lsp/server.go` (simple implementation ~400 lines)
- Pattern: Intercepts requests, translates positions, forwards to gopls
- Good for: Understanding request/response translation

**2. gopls itself** - The reference
- Source: `github.com/golang/tools/gopls`
- Key file: `internal/server.go` (massive, but structured)
- Pattern: Implements full LSP server from scratch
- Good for: Understanding complete LSP protocol

**3. Neovim LSP client** - Client-side proxy
- Shows how clients wrap multiple servers
- Pattern: Multiplexing, buffer position tracking
- Relevant for: Understanding client expectations

### Canonical Proxy Pattern (From Community)

```
1. Create stdio transport ✅ You do this

2. Create connection and start handler ✅ You do this

3. For each request:
   a. Unmarshal params ✅ You do this
   b. Translate position (if needed) ✅ You do this
   c. Forward to subprocess ✅ You do this
   d. Translate response back ✅ You do this partially
   e. Call Replier ✅ You do this

4. For unsolicited messages:
   a. Use conn.Notify() ❌ You don't do this yet

5. Handle shutdown gracefully ✅ You have structure but incomplete
```

## 7. Common Pitfalls (How to Avoid Them)

### Pitfall 1: Race Conditions on Replier

**Problem**:
```go
// WRONG - Replier called twice
func (s *Server) handleRequest(..., reply jsonrpc2.Replier, ...) error {
    result := doWork()
    reply(ctx, result, nil)  // First reply

    // Later...
    go func() {
        moreWork()
        reply(ctx, extraData, nil)  // ERROR: Can't reply twice!
    }()
}
```

**Solution**: Reply exactly once, or use Notify for follow-ups:
```go
// CORRECT - Reply once, notify for updates
func (s *Server) handleRequest(..., reply jsonrpc2.Replier, ...) error {
    result := doWork()
    reply(ctx, result, nil)  // Reply to the request

    // Publish diagnostics separately
    go func() {
        diagnostics := compute()
        s.ideConn.Notify(ctx, "textDocument/publishDiagnostics", ...)
    }()
}
```

### Pitfall 2: Deadlock on Connection Close

**Problem**:
```go
// WRONG - Waiting for response after closing connection
func (s *Server) shutdown() {
    s.ideConn.Close()  // Close connection
    result := s.gopls.Call(ctx, "shutdown", ...)  // Wait for response - DEADLOCK!
}
```

**Solution**: Close in correct order:
```go
// CORRECT - Shutdown gracefully
func (s *Server) shutdown() {
    s.gopls.Shutdown(ctx)      // Tell gopls to shutdown
    s.gopls.Wait()             // Wait for process exit
    s.gopls.Close()            // Close connection
    // IDE will close its connection when it sees shutdown response
}
```

### Pitfall 3: Context Timeout Issues

**Problem**:
```go
// WRONG - Context deadline exceeded
conn.Go(context.WithTimeout(ctx, 5*time.Second), handler)
// After 5 seconds, conn closes!
```

**Solution**: Use Background or long-lived context:
```go
// CORRECT - conn.Go needs long-lived context
conn.Go(context.Background(), handler)
// Or store context in Server, keep it alive
```

### Pitfall 4: Lost Responses in Translation

**Problem**:
```go
// WRONG - Only translating first item
func translateCompletionResponse(items []protocol.CompletionItem) []protocol.CompletionItem {
    return []protocol.CompletionItem{items[0]}  // Lost all others!
}
```

**Solution**: Translate all items:
```go
// CORRECT - Translate everything
func (t *Translator) TranslateCompletionResponse(items []protocol.CompletionItem) []protocol.CompletionItem {
    result := make([]protocol.CompletionItem, len(items))
    for i, item := range items {
        result[i] = t.TranslateCompletionItem(item)
    }
    return result
}
```

## 8. Specific Recommendations for Dingo-LSP

### Priority 1: Implement Server-Initiated Messages

```go
// In your existing Server struct
type Server struct {
    ideConn jsonrpc2.Conn
    ctx context.Context
    // NEW:
    mu sync.Mutex  // Protect ideConn for concurrent use
}

// NEW METHOD:
func (s *Server) PublishDiagnostics(uri uri.URI, diagnostics []protocol.Diagnostic) {
    s.mu.Lock()
    defer s.mu.Unlock()

    if err := s.ideConn.Notify(s.ctx, "textDocument/publishDiagnostics",
        protocol.PublishDiagnosticsParams{
            URI:         uri,
            Diagnostics: diagnostics,
        }); err != nil {
        s.config.Logger.Warnf("Failed to publish diagnostics: %v", err)
    }
}

// When transpiler detects errors, call this
```

### Priority 2: Implement Complete Shutdown

Add to gopls_client.go:
```go
func (c *GoplsClient) Wait() error { ... }
func (c *GoplsClient) Close() error { ... }
```

Update handleShutdown to call these.

### Priority 3: Add Context Protection

```go
// In Server - protect ideConn access
type Server struct {
    ideConn jsonrpc2.Conn
    ctx context.Context
    mu sync.Mutex  // NEW: Protect concurrent access
}

// Anywhere you use ideConn, lock first:
s.mu.Lock()
defer s.mu.Unlock()
s.ideConn.Notify(...)
```

### Priority 4: Test Bidirectional Communication

Create integration test:
```go
func TestServerInitializedNotification(t *testing.T) {
    // Verify that after initialized, diagnostics are sent
    // Verify that gopls crashes are communicated to IDE
    // Verify that diagnostics are published
}
```

## Summary Table: Current State vs. Recommended

| Feature | Current | Recommended | Impact |
|---------|---------|-------------|--------|
| Request handling | ✅ Good | - | High |
| Response translation | ✅ Partial | Complete all response types | High |
| Server-initiated messages | ❌ Missing | Implement Notify() | High |
| Connection shutdown | ⚠️ Partial | Complete Wait() + Close() | Medium |
| Concurrent access | ⚠️ No locks | Add mu sync.Mutex | Medium |
| Error handling | ✅ Good | Add retry logic for transients | Low |
| Source map validation | ✅ Good | Already good | - |

## Conclusion

Your dingo-lsp proxy is well-architected. The jsonrpc2 library handles most of the complexity. Your main gaps are:

1. **Server-initiated messages** (diagnostics, notifications) - Use `ideConn.Notify()`
2. **Complete shutdown** - Implement gopls_client.Wait() and Close()
3. **Concurrent protection** - Add mutex for ideConn access
4. **Response translation** - Ensure all response types are fully translated

These are all straightforward fixes that follow the canonical LSP proxy pattern. Study templ's implementation for a clean reference, and you'll have a solid foundation for Dingo's IDE support.

The jsonrpc2 library does what it should - you just need to use all of its capabilities (Notify, Call, Close) which you're not currently doing.

# GPT-5.1-Codex: jsonrpc2 Connection Lifecycle Investigation

## Problem Statement

The dingo-lsp server crashes immediately with error: "Cannot call write after a stream was destroyed". The process exits and no LSP server remains running.

## Code Context

### main.go (cmd/dingo-lsp/main.go)
```go
// Create stdio transport using ReadWriteCloser wrapper
rwc := &stdinoutCloser{stdin: os.Stdin, stdout: os.Stdout}
stream := jsonrpc2.NewStream(rwc)
conn := jsonrpc2.NewConn(stream)

// Start serving
ctx := context.Background()

// Create handler and start connection
handler := server.Handler()
conn.Go(ctx, handler)

// Store connection in server for use by handlers
server.SetConn(conn, ctx)

// Wait for connection to close
<-conn.Done()
logger.Infof("Server stopped")
```

**Key Components:**
- `stdinoutCloser`: Wraps os.Stdin/os.Stdout, Close() is no-op
- `jsonrpc2.NewStream(rwc)`: Creates stream from ReadWriteCloser
- `jsonrpc2.NewConn(stream)`: Creates connection from stream
- `conn.Go(ctx, handler)`: Spawns goroutine to read/handle messages
- `<-conn.Done()`: Waits for connection to close

### server.go (pkg/lsp/server.go)

**SetConn method (lines 63-67):**
```go
func (s *Server) SetConn(conn jsonrpc2.Conn, ctx context.Context) {
	s.ideConn = conn
	s.ctx = ctx
}
```

**Server struct (lines 21-32):**
```go
type Server struct {
	config        ServerConfig
	gopls         *GoplsClient
	mapCache      *SourceMapCache
	translator    *Translator
	transpiler    *AutoTranspiler
	watcher       *FileWatcher
	workspacePath string
	initialized   bool
	ideConn       jsonrpc2.Conn  // CRITICAL FIX C1: Store IDE connection
	ctx           context.Context // IMPORTANT FIX I3: Store server context
}
```

## Critical Questions to Investigate

1. **Connection Initialization Race**:
   - Is `conn.Go(ctx, handler)` creating a goroutine that immediately tries to read from stdin?
   - Could the stream be closed before the handler is registered?
   - What happens if the goroutine exits before `<-conn.Done()` is reached?

2. **Stream Destruction Timing**:
   - When/how does the stream get destroyed?
   - Is `stdinoutCloser.Close()` being called unexpectedly?
   - Could there be a double-close or premature close in jsonrpc2?

3. **Handler Pattern Correctness**:
   - Is `jsonrpc2.ReplyHandler(s.handleRequest)` the correct wrapper?
   - Should there be a specific error recovery pattern in the handler?
   - What happens when a handler returns an error?

4. **Context Lifecycle**:
   - Is `context.Background()` the right choice for the main connection context?
   - Should handlers use `s.ctx` instead of the passed `ctx`?
   - Could context cancellation be causing premature shutdown?

5. **gopls Integration**:
   - Could gopls subprocess startup failure be breaking the connection?
   - Are there any goroutine leaks in GoplsClient initialization?
   - Is the stdio stream being consumed by gopls startup code?

## Expected Analysis Output

For each question, provide:
- Root cause analysis
- Code flow diagram showing message lifecycle
- Recommended fixes with code examples
- References to jsonrpc2 documentation patterns
- Common pitfalls in jsonrpc2 server implementations

## Files to Reference

- `go.lsp.dev/jsonrpc2` package documentation
- Standard LSP server implementations (gopls, pylance, etc.)
- Known jsonrpc2 usage patterns in Go ecosystem

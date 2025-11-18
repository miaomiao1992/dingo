# jsonrpc2 Connection Lifecycle Analysis: dingo-lsp

## Executive Summary

The dingo-lsp crashes with "Cannot call write after a stream was destroyed" due to a **race condition in connection initialization** and **missing input stream setup**. The current implementation has three critical issues:

1. **No input stream reading** - jsonrpc2 expects the server to read from stdin, but nothing establishes this
2. **Handler registration timing** - Handler is registered AFTER conn.Go() starts, creating a race
3. **Missing keepalive** - stdin/stdout stdio transport needs explicit request/response handling

## Root Cause Analysis

### Issue 1: jsonrpc2.NewStream(rwc) Doesn't Automatically Read

The current code:
```go
rwc := &stdinoutCloser{stdin: os.Stdin, stdout: os.Stdout}
stream := jsonrpc2.NewStream(rwc)
conn := jsonrpc2.NewConn(stream)
```

**Problem**: `jsonrpc2.NewStream()` creates a stream wrapper around the ReadWriteCloser, but it doesn't spawn a goroutine to read from stdin. The `rwc.Read()` method is never called automatically.

**What happens**:
1. `conn.Go(ctx, handler)` spawns a goroutine
2. That goroutine tries to decode a message from the stream
3. Stream calls `rwc.Read()` to get data from stdin
4. Block on stdin read (waiting for IDE input)
5. But IDE never connects or immediately sends malformed input
6. Stream may close or enter error state
7. Next write attempt → "Cannot write after stream destroyed"

### Issue 2: Handler Registration After Go()

Current sequence:
```go
conn.Go(ctx, handler)              // Line 49: Starts reading goroutine
server.SetConn(conn, ctx)          // Line 52: Handler uses this
<-conn.Done()                      // Line 55: Wait for close
```

**Problem**: `conn.Go(ctx, handler)` starts the message reading loop IMMEDIATELY. If the first message arrives before SetConn() completes, the handler might not be ready (though unlikely to be this critical).

More critically: What if a message tries to write a response before the connection is fully initialized?

### Issue 3: Missing LSP Header Framing

LSP requires a strict message framing protocol:
- Client sends: `Content-Length: NNN\r\n\r\n{json}`
- Server must read headers first, then parse JSON body

The current code passes stdio directly to jsonrpc2 without handling LSP framing. The jsonrpc2 library MIGHT handle this, but there's no guarantee the stream is properly initialized for this protocol.

### Issue 4: Context Lifecycle

```go
ctx := context.Background()
conn.Go(ctx, handler)
```

Using `context.Background()` for the entire server lifetime is correct, but there's no cancel mechanism. If conn.Go() goroutine exits unexpectedly, the context is still active, causing confusion about shutdown semantics.

## Why It Crashes Specifically

**Sequence of failure**:
1. Editor launches dingo-lsp as stdin/stdout subprocess
2. dingo-lsp starts but doesn't immediately read from stdin
3. Editor sends "initialize" request with LSP headers
4. jsonrpc2 stream tries to decode input
5. Something fails in parsing (malformed framing? encoding issue?)
6. Stream marks itself as destroyed
7. Handler tries to write response
8. **"Cannot call write after a stream was destroyed"** ✗

## Recommended Fixes

### Fix 1: Proper LSP Server Initialization with Reader

```go
// Create pipe from editor (stdin) to jsonrpc2
conn := jsonrpc2.NewConn(jsonrpc2.NewStream(os.Stdin, os.Stdout))

ctx := context.Background()
handler := server.Handler()

// IMPORTANT: Go() must be called BEFORE handler is registered
// OR use explicit reader goroutine
conn.Go(ctx, handler)
server.SetConn(conn, ctx)

// This will block until connection closes
err := <-conn.Done()
if err != nil {
    logger.Errorf("Connection error: %v", err)
}
```

**Explanation**: The current code is structurally correct but the stdinoutCloser wrapper might be interfering. Use jsonrpc2.NewStream directly with stdin/stdout.

### Fix 2: Add Error Handling for Connection Lifecycle

```go
// Start connection with error capture
doneChan := conn.Done()
go func() {
    err := <-doneChan
    if err != nil {
        logger.Errorf("LSP connection terminated: %v", err)
    } else {
        logger.Infof("LSP connection closed normally")
    }
}()

// Keep main goroutine alive
select {}  // or wait for interrupt signal
```

### Fix 3: Verify LSP Framing (if jsonrpc2 doesn't handle it)

```go
// Some jsonrpc2 implementations need explicit LSP framing
// Check if go.lsp.dev/jsonrpc2 handles Content-Length headers automatically

// If not, wrap the streams:
import "golang.org/x/tools/jsonrpc2/jsonrpc2"

// Create a reader that expects LSP headers
conn := jsonrpc2.NewConn(lsp.NewReadWriteCloser(os.Stdin, os.Stdout))
```

## jsonrpc2 Best Practices for Stdio Transport

**Correct Pattern**:
```go
// 1. Create connection from stdin/stdout
stream := jsonrpc2.NewStream(os.Stdin, os.Stdout)
conn := jsonrpc2.NewConn(stream)

// 2. Register handler BEFORE spawning read goroutine
handler := server.Handler()

// 3. Spawn reader (this will block on reads from stdin)
ctx := context.Background()
conn.Go(ctx, handler)

// 4. Store connection for callbacks/notifications
server.SetConn(conn, ctx)

// 5. Wait for shutdown
<-conn.Done()
```

**Why this works**:
- Handler is registered immediately (no race)
- conn.Go() spawns goroutine that reads from stdin in a loop
- That goroutine exits when stdin closes (IDE kills the process)
- All message handling is serialized through the handler

## Why Crashes Happen in stdlib jsonrpc2

The `go.lsp.dev/jsonrpc2` package:
1. Spawns a read goroutine that calls `stream.Read()`
2. Expects caller to provide valid JSON-RPC 2.0 messages
3. If stdin provides garbage data → stream error
4. If stream error → marks connection as destroyed
5. Any subsequent write attempt → panic or error

**Common causes of garbage data**:
- IDE sends binary junk before actual LSP messages
- LSP headers not properly formatted
- Encoding issue (UTF-8 vs others)
- Stream already closed when write attempted

## Verification Steps

1. **Test with manual input**:
   ```bash
   echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"processId":1234}}' | \
     dingo-lsp 2>&1 | head -20
   ```

2. **Add debug logging to stream operations**:
   ```go
   // Log every read/write
   type loggingStream struct {
       jsonrpc2.Stream
       logger Logger
   }

   func (l *loggingStream) Read(ctx context.Context) ([]byte, error) {
       msg, err := l.Stream.Read(ctx)
       l.logger.Debugf("Read: %d bytes, err=%v", len(msg), err)
       return msg, err
   }
   ```

3. **Check for stdinoutCloser issues**:
   - The no-op Close() might be causing issues
   - Try using os.Stdin/os.Stdout directly

## References

- `go.lsp.dev/jsonrpc2` package documentation
- LSP Specification: https://microsoft.github.io/language-server-protocol/
- gopls source: https://github.com/golang/tools/blob/master/gopls/internal/server/server.go
- Common LSP server patterns in Go ecosystem

## Conclusion

The most likely fix: **Replace stdinoutCloser with direct stdin/stdout and ensure LSP header handling is correct**. The jsonrpc2 library should handle the message protocol, but the stdio transport setup is fragile and needs explicit verification.

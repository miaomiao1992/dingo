# VS Code LSP Client Stream Analysis - Direct Investigation

## Root Cause: Stream Destruction on Server Errors

The "Cannot call write after a stream was destroyed" error occurs because:

1. **Silent Server Failures**: dingo-lsp crashes or exits unexpectedly (invalid protocol response, panic, etc.)
2. **Insufficient Error Handler**: Current error handler returns `action: 2` (Continue), but this doesn't stop the client from trying to write to a destroyed stream
3. **No Stream State Validation**: Client doesn't check if the underlying stdio stream is still valid before sending messages
4. **Race Condition on Init**: Client might attempt operations before server completes LSP initialize handshake

## Current Problem Areas

### 1. Error Handler Configuration (CRITICAL)
```typescript
errorHandler: {
    error: () => ({ action: 2 }), // Continue - WRONG!
    closed: () => ({ action: 1 })  // Don't restart - WRONG!
}
```

**Issues:**
- `action: 2` (Continue) tells client to ignore the error and keep sending messages
- `action: 1` (Don't restart) means a crashed server stays down
- These actions are backwards for LSP reliability

**Should be:**
```typescript
errorHandler: {
    error: () => ({ action: 1 }), // Restart on error
    closed: () => ({ action: 1 })  // Restart when closed
}
```

### 2. Missing Initialization Handler
The client doesn't handle initialization failures. If dingo-lsp fails during LSP initialization handshake, the stream might be left in an inconsistent state.

**Missing:**
```typescript
initializationFailedHandler: (error: Error) => {
    // Log properly, don't try to continue
    return true; // Stop client
}
```

### 3. No Stream Health Check
Before operations, the client should verify the server is responding.

## LSP Protocol Requirements (Why This Matters)

1. **Initialize Must Complete**: Server must respond to `initialize` request with capabilities
2. **Bidirectional Communication**: Messages flow client→server AND server→client
3. **Stdio Transport Issues**:
   - If server writes invalid JSON to stdout, stream breaks
   - If server crashes, stdin becomes invalid
   - No recovery without restart

## Timing Issue: Race Condition

**Sequence that causes crashes:**
1. Client calls `client.start()`
2. Subprocess spawned, dingo-lsp starts
3. dingo-lsp initializes gopls subprocess
4. If gopls startup fails, dingo-lsp might crash before sending initialize response
5. Client has no stream to write to anymore
6. Subsequent operations: "Cannot call write after a stream was destroyed"

## LSP Server-Side Problems to Check

In dingo-lsp Go code, verify:
1. **Initialize Handler**: Does it respond properly to `initialize` request?
2. **Error Recovery**: Does server handle gopls failures gracefully?
3. **Stdout Validity**: Does server output valid JSON-RPC only?
4. **Stdio Buffering**: Is stdout properly flushed after each message?

## Fixes Required (Priority Order)

### High Priority
1. **Fix error handler actions** - Change `2` (Continue) to `1` (Restart)
2. **Add initialization handler** - Catch init failures early
3. **Add stream state logging** - Log when stream closes to identify crashes
4. **Add connection timeout** - Set maximum initialization wait time

### Medium Priority
5. **Add heartbeat mechanism** - Detect dead servers
6. **Improve error messages** - Show actual crash reason to user
7. **Add server output capture** - Log stderr from dingo-lsp for debugging

### Low Priority
8. **Remove DINGO_AUTO_TRANSPILE** - LSP mode doesn't use this
9. **Validate fileSystemWatcher** - Ensure it doesn't fire during init
10. **Add retry logic** - Gracefully retry failed operations

## Code Pattern for Robust LSP Clients

```typescript
const clientOptions: LanguageClientOptions = {
    documentSelector: [{ scheme: 'file', language: 'dingo' }],
    synchronize: {
        fileEvents: vscode.workspace.createFileSystemWatcher('**/*.{dingo,go.map}')
    },
    outputChannelName: 'Dingo Language Server',

    // Recovery on ANY error
    errorHandler: {
        error: (error, message, count) => {
            // Log the error for debugging
            console.error(`LSP Error [${count}]:`, error, message);
            return { action: 1 }; // Always restart
        },
        closed: () => {
            console.warn('LSP connection closed, restarting');
            return { action: 1 }; // Always restart
        }
    },

    // Fail fast on init problems
    initializationFailedHandler: (error) => {
        console.error('LSP initialization failed:', error);
        return true; // Stop and don't retry
    }
};
```

## Next Steps for Investigation

1. **Check dingo-lsp for**:
   - Initialize response format (must match LSP spec)
   - Proper error propagation
   - Stdout stream integrity

2. **Monitor in VS Code**:
   - Enable debug logging: Set `DINGO_LSP_LOG=debug`
   - Watch the "Dingo Language Server" output channel
   - Look for incomplete JSON-RPC messages

3. **Add telemetry**:
   - Log every client→server message
   - Log every server→client message
   - Log stream state changes

## Why This Matters for Dingo

For a language server to work reliably in production:
- Users need to trust it won't crash
- Crashes should be recoverable (auto-restart)
- Errors should be visible (proper logging)
- The stream lifecycle must be managed defensively

The current configuration treats errors as "continue anyway," which is fundamentally wrong for stdio-based LSP.

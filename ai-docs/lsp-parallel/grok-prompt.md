# VS Code LSP Client Stream Handling Analysis

## Error Context
The VS Code extension reports: "Cannot call write after a stream was destroyed"

This occurs during LSP client initialization/operation with the dingo-lsp server.

## Current Implementation

### Client Configuration (`lspClient.ts`, lines 35-49)
```typescript
const clientOptions: LanguageClientOptions = {
    documentSelector: [
        { scheme: 'file', language: 'dingo' }
    ],
    synchronize: {
        fileEvents: vscode.workspace.createFileSystemWatcher('**/*.{dingo,go.map}')
    },
    outputChannelName: 'Dingo Language Server',
    errorHandler: {
        error: () => ({ action: 2 }), // Continue
        closed: () => ({ action: 1 })  // Don't restart automatically on close
    }
};
```

### Server Invocation (lines 21-32)
```typescript
const serverOptions: ServerOptions = {
    command: lspPath,
    args: [],
    transport: TransportKind.stdio,
    options: {
        env: {
            ...process.env,
            DINGO_LSP_LOG: logLevel,
            DINGO_AUTO_TRANSPILE: transpileOnSave.toString(),
        }
    }
};
```

### Client Lifecycle (lines 59-93)
- `client.start()` called in try/catch
- Error handler configured to continue on errors
- Closed handler configured NOT to auto-restart
- No explicit stream/pipe management

## Investigation Questions

1. **Stream Lifecycle Management**
   - When does the stream get destroyed? (client.stop(), server crash, protocol error?)
   - Is there a race condition between client.start() and immediate file operations?
   - Should we add explicit stream state checking before write operations?

2. **Error Handler Behavior**
   - What does `action: 2` (Continue) actually do? Does it keep the stream alive?
   - Should `error` handler return `action: 1` (Restart) instead for fatal errors?
   - Is the distinction between `error` and `closed` handlers correct?

3. **Initialization Handshake**
   - Does dingo-lsp send proper LSP initialize response?
   - Are there missing capabilities/methods the client expects?
   - Should we add `initializationFailedHandler` to catch startup issues?

4. **File System Watcher**
   - Could the FileSystemWatcher be causing issues? (line 41)
   - Should we dispose it properly on errors?
   - Is there a timing issue where watcher fires before client is ready?

5. **Server Process Management**
   - Should we add heartbeat/keepalive mechanism?
   - Is DINGO_AUTO_TRANSPILE environment variable valid for LSP mode?
   - Should we add explicit stdio pipe error handlers?

## Expected Output

1. **Root Cause Analysis**: What actually causes the "stream destroyed" error?
2. **Error Handler Fix**: Should we change error/closed action codes?
3. **Initialization Fix**: What's missing from client init that causes stream destruction?
4. **Best Practices**: How should VS Code extensions handle LSP server crashes?
5. **Concrete Code Changes**: Specific lspClient.ts modifications needed.

## Context
- Dingo is a Go transpiler (TypeScript-like for Go)
- dingo-lsp is a gopls-wrapping language server
- VS Code extension uses vscode-languageclient 8.1.0
- Server transport: stdio (not TCP)

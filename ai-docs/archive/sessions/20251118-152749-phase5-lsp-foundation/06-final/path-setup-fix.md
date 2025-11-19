# PATH Configuration Fix for gopls

**Date:** 2025-11-18
**Issue:** gopls installed but not found in PATH
**Status:** ✅ RESOLVED

## Problem

After running `go install golang.org/x/tools/gopls@latest`, gopls was installed to `$GOPATH/bin` (`/Users/jack/go/bin/gopls`), but when running `dingo-lsp`, it failed with:

```
[FATAL] gopls not found in $PATH
```

**Root Cause:** `$GOPATH/bin` was not in the user's zsh PATH environment variable.

## Solution

Added the following line to `~/.zshrc`:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

This ensures:
1. gopls is accessible from any terminal session
2. Other Go-installed tools (like `dingo-lsp` if installed via `go install`) are also accessible
3. PATH is automatically configured on shell startup

## Verification

After applying the fix:

```bash
$ which gopls
/Users/jack/go/bin/gopls

$ gopls version
golang.org/x/tools/gopls v0.20.0

$ which dingo-lsp
/usr/local/bin/dingo-lsp

$ ./dingo-lsp
[INFO] Starting dingo-lsp server
[INFO] Found gopls at: /Users/jack/go/bin/gopls
[INFO] gopls started (PID: 71315)
```

✅ All tools now accessible and working correctly.

## Updated Scripts

Updated `scripts/lsp-quicktest.sh` to automatically:
1. Check if gopls is in PATH
2. If not, install gopls
3. Ensure `$GOPATH/bin` is added to PATH
4. Add to `~/.zshrc` if not already present

This prevents future users from encountering the same issue.

## Files Modified

- `~/.zshrc` - Added GOPATH/bin to PATH
- `scripts/lsp-quicktest.sh` - Added automatic PATH configuration

## Next Steps

✅ Ready for manual LSP testing in VSCode!

Follow the guides:
- Quick start: `HOW-TO-TEST-LSP.md`
- Comprehensive: `docs/MANUAL-LSP-TESTING.md`

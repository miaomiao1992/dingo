# Library Example - Publishing a Dingo Library

This example demonstrates how to create and publish a Dingo library that can be consumed by both Go and Dingo projects.

## Overview

**Library Name**: `mathutils`
**Purpose**: Demonstrate safe mathematical operations using Result types
**Target Users**: Go and Dingo developers

## Strategy

This library follows the **transpile-on-publish** strategy:
1. Develop in `.dingo` files
2. Transpile to `.go` files before publishing
3. Commit both `.dingo` (source) and `.go` (transpiled) files
4. Consumers get idiomatic Go code via `go get`

## Files

```
library-example/
├── README.md           # This file
├── go.mod              # Module definition
├── mathutils.dingo     # Source code (Dingo)
├── mathutils.go        # Transpiled Go (committed for consumers)
└── mathutils_test.go   # Tests (Go, works with both)
```

## Building the Library

### Step 1: Develop in Dingo

Edit `mathutils.dingo` to add features.

### Step 2: Transpile

```bash
# From this directory
dingo build mathutils.dingo

# This generates mathutils.go
```

### Step 3: Test

```bash
# Run tests using standard Go tooling
go test -v

# Or use go test with coverage
go test -v -cover
```

### Step 4: Verify Go Compilation

```bash
# Ensure the transpiled code compiles
go build ./...
```

## Publishing Workflow

### Initial Setup

```bash
# Initialize as Go module (already done)
go mod init github.com/yourname/mathutils

# Create repository
git init
git add .
git commit -m "feat: initial commit"
git remote add origin https://github.com/yourname/mathutils.git
git push -u origin main
```

### Publishing a Release

```bash
# 1. Make changes to mathutils.dingo
vim mathutils.dingo

# 2. Transpile
dingo build mathutils.dingo

# 3. Test
go test ./...

# 4. Commit both .dingo and .go files
git add mathutils.dingo mathutils.go
git commit -m "feat: add SafeSqrt function"

# 5. Tag version (semantic versioning)
git tag v1.1.0

# 6. Push
git push origin main --tags
```

### Verification

Your library is now available via:

```bash
go get github.com/yourname/mathutils@v1.1.0
```

## Consuming the Library

### From Pure Go

```go
package main

import (
    "fmt"
    "github.com/yourname/mathutils"
)

func main() {
    // Use the library without Dingo installed!
    result := mathutils.SafeDivide(10, 2)

    if result.IsOk() {
        fmt.Println("Result:", result.Unwrap())
    } else {
        fmt.Println("Error:", result.UnwrapErr())
    }
}
```

### From Dingo

```dingo
package main

import "github.com/yourname/mathutils"

func main() {
    // Use Dingo syntax for error propagation
    value := mathutils.SafeDivide(10, 2)?
    fmt.Println("Result:", value)
}
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Test and Publish

on:
  push:
    tags:
      - 'v*'

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Install Dingo
        run: |
          go install github.com/dingo-lang/dingo/cmd/dingo@latest

      - name: Transpile
        run: dingo build ./...

      - name: Test
        run: go test -v ./...

      - name: Check for uncommitted changes
        run: |
          git diff --exit-code
          # Fail if .go files weren't committed
```

## Best Practices

### ✅ DO

- Commit both `.dingo` and `.go` files
- Run `go test` before publishing
- Use semantic versioning
- Include usage examples in README
- Transpile before every release

### ❌ DON'T

- Gitignore `.go` files in libraries (consumers need them)
- Manually edit transpiled `.go` files (edit `.dingo` instead)
- Publish without running tests
- Forget to tag versions

## Key Advantages

1. **Zero Barrier**: Go developers can use your library without Dingo
2. **Full Tooling**: Works with gopls, go get, pkg.go.dev
3. **Type Safety**: Result/Option types are normal Go structs
4. **Performance**: No runtime overhead, it's just Go

## Next Steps

- See `app-example/` for application development pattern
- See `hybrid-example/` for consuming this library from an app
- Read `docs/package-management.md` for complete guide

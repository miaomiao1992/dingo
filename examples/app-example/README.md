# Application Example - Building with Dingo

This example demonstrates how to build an application using Dingo, keeping `.dingo` source files and transpiling during development.

## Overview

**Application**: Simple TODO CLI
**Strategy**: Keep `.dingo` files, transpile locally
**Target**: Internal use (not published as library)

## Strategy

This application follows the **direct .dingo usage** pattern:
1. Keep `.dingo` files in version control
2. Gitignore transpiled `.go` files
3. Transpile during development and build
4. Deploy compiled binary (no Dingo required in production)

## Files

```
app-example/
├── README.md           # This file
├── go.mod              # Module definition
├── .gitignore          # Ignores *.go files
├── main.dingo          # Entry point
├── tasks.dingo         # Business logic
└── Makefile            # Build automation
```

## Development Workflow

### Step 1: Edit Dingo Files

```bash
# Edit the source
vim main.dingo
vim tasks.dingo
```

### Step 2: Transpile and Run

```bash
# Transpile all .dingo files
dingo build ./...

# Run the application
go run main.go
```

Or use the Makefile:

```bash
make run
```

### Step 3: Build Production Binary

```bash
# Transpile and build
make build

# Run the binary
./todo-cli add "Buy groceries"
./todo-cli list
```

## Makefile Usage

The included Makefile provides convenient commands:

```bash
# Development
make run              # Transpile and run
make watch            # Auto-rebuild on changes (future)

# Testing
make test             # Run tests

# Production
make build            # Build production binary
make clean            # Remove generated files

# Help
make help             # Show all commands
```

## .gitignore Strategy

The `.gitignore` file excludes transpiled Go code:

```gitignore
# Transpiled Go files (application mode)
*.go
!vendor/**/*.go

# Binaries
todo-cli
*.exe

# Source maps (optional)
*.go.golden.map
```

**Why**: The `.dingo` files are the source of truth. Transpiled Go is generated, not authored.

## Production Deployment

### Building for Deployment

```bash
# Cross-compile for different platforms
GOOS=linux GOARCH=amd64 make build
```

### Dockerfile Example

```dockerfile
# Build stage
FROM golang:1.21 AS builder

# Install Dingo
RUN go install github.com/dingo-lang/dingo/cmd/dingo@latest

WORKDIR /app
COPY . .

# Transpile Dingo to Go
RUN dingo build ./...

# Build binary
RUN go build -o todo-cli main.go

# Runtime stage (Dingo not needed!)
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/todo-cli .

ENTRYPOINT ["/app/todo-cli"]
```

**Note**: The final image doesn't need Dingo - only the compiled binary!

## CI/CD Example

### GitHub Actions

```yaml
name: Build and Test

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  build:
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

      - name: Build
        run: go build -o todo-cli main.go

      - name: Upload artifact
        uses: actions/upload-artifact@v3
        with:
          name: todo-cli
          path: todo-cli
```

## Running the TODO CLI

### Add a Task

```bash
./todo-cli add "Buy groceries"
./todo-cli add "Write documentation"
./todo-cli add "Review pull request"
```

### List Tasks

```bash
./todo-cli list
```

Output:
```
TODO List:
1. [ ] Buy groceries
2. [ ] Write documentation
3. [ ] Review pull request
```

### Complete a Task

```bash
./todo-cli complete 1
./todo-cli list
```

Output:
```
TODO List:
1. [x] Buy groceries
2. [ ] Write documentation
3. [ ] Review pull request
```

## Best Practices

### ✅ DO

- Gitignore `.go` files for applications
- Use Makefile or build scripts for automation
- Transpile as part of CI/CD
- Deploy binaries, not source code
- Keep `.dingo` as source of truth

### ❌ DON'T

- Manually edit transpiled `.go` files
- Commit `.go` files for applications
- Require Dingo in production (only build time)
- Mix `.dingo` and `.go` for same functionality

## Key Advantages

1. **Best DX**: Full Dingo syntax during development
2. **Fast Iteration**: No publish step, instant feedback
3. **Better Debugging**: Source maps point to `.dingo` files
4. **LSP Integration**: IDE features work on `.dingo` directly
5. **Production Ready**: Deployed binary is pure Go

## File Structure for Larger Apps

```
myapp/
├── go.mod
├── .gitignore
├── Makefile
├── cmd/
│   └── myapp/
│       └── main.dingo
├── internal/
│   ├── config/
│   │   ├── config.dingo
│   │   └── config_test.go
│   ├── handlers/
│   │   ├── http.dingo
│   │   └── http_test.go
│   └── models/
│       ├── user.dingo
│       └── user_test.go
├── pkg/
│   └── utils/
│       ├── strings.dingo
│       └── strings_test.go
└── deployments/
    ├── Dockerfile
    └── docker-compose.yml
```

## Next Steps

- See `library-example/` for library publishing pattern
- See `hybrid-example/` for consuming published libraries
- Read `docs/package-management.md` for complete guide
- Try building your own Dingo application!

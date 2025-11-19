# Hybrid Example - Consuming a Dingo Library

This example demonstrates how a Dingo application can consume a published Dingo library.

## Overview

**Scenario**: A calculator application consuming the `mathutils` library
**Pattern**: Hybrid - app in Dingo, library published as Go
**Demonstrates**: Interoperability between Dingo code

## Architecture

```
┌─────────────────────────────────┐
│  Calculator App (.dingo)        │
│  - Direct .dingo development    │
│  - Gitignore .go files          │
└────────────┬────────────────────┘
             │ imports
             ↓
┌─────────────────────────────────┐
│  mathutils Library (.go)        │
│  - Published as Go package      │
│  - go get compatible            │
│  - Originally written in Dingo  │
└─────────────────────────────────┘
```

## Files

```
hybrid-example/
├── README.md           # This file
├── go.mod              # Imports mathutils library
├── .gitignore          # Ignores *.go (app mode)
├── calculator.dingo    # Main application
└── Makefile            # Build automation
```

## Setup

### Step 1: Install the Library

This example uses a **local replace** directive for demonstration:

```go.mod
module github.com/example/calculator

go 1.21

// In production, you'd use:
// require github.com/example/mathutils v1.0.0

// For this demo, use local library
replace github.com/example/mathutils => ../library-example
```

In production, you would just:
```bash
go get github.com/example/mathutils@v1.0.0
```

### Step 2: Import and Use

The application imports the library like any Go package:

```dingo
import "github.com/example/mathutils"

func calculate() {
    // Use library's Result type and functions
    result := mathutils.SafeDivide(10, 2)?
    fmt.Println(result)
}
```

## Running the Calculator

### Build and Run

```bash
# Using Makefile
make run

# Or manually
dingo build calculator.dingo
go run calculator.go
```

### Example Session

```bash
$ make run
Transpiling Dingo files...
Running application...

=== Dingo Calculator ===

Testing SafeDivide(10, 2):
Result: 5

Testing SafeDivide(10, 0):
Error: division by zero

Testing SafeSqrt(16):
Result: 4

Testing SafeSqrt(-4):
Error: cannot calculate square root of negative number

Testing SafeModulo(17, 5):
Result: 2

Testing SafeModulo(10, 0):
Error: modulo by zero

=== Demonstrating Error Propagation ===

Chain calculation: sqrt(100) / 2
Result: 5

Chain calculation with error: sqrt(-25) / 2
Error: cannot calculate square root of negative number
```

## Key Concepts

### 1. Seamless Integration

The application uses the library as if it were written in Dingo:

```dingo
// Both app and library speak the same language
result := mathutils.SafeDivide(a, b)?

// The ? operator works because both use Result<T,E>
```

### 2. Type Compatibility

Result types are identical across Dingo code:

```dingo
// Library defines Result[T, E]
// App can use it directly with ? operator
// Full type inference works
```

### 3. Zero Overhead

At runtime, it's all compiled Go code:
- No Dingo runtime
- No type conversions
- Same performance as pure Go

## Comparison: Dingo vs Pure Go

### Using the Library in Dingo (This Example)

```dingo
func chainedCalculation(x: float64) -> Result[float64, error] {
    sqrt := mathutils.SafeSqrt(x)?
    result := mathutils.SafeDivide(sqrt, 2)?
    return Ok[float64, error](result)
}
```

**Lines**: 4
**Error handling**: Implicit with `?`

### Using the Same Library in Pure Go

```go
func chainedCalculation(x float64) (float64, error) {
    sqrtResult := mathutils.SafeSqrt(x)
    if sqrtResult.IsErr() {
        return 0, sqrtResult.UnwrapErr()
    }
    sqrt := sqrtResult.Unwrap()

    divResult := mathutils.SafeDivide(sqrt, 2)
    if divResult.IsErr() {
        return 0, divResult.UnwrapErr()
    }

    return divResult.Unwrap(), nil
}
```

**Lines**: 13
**Error handling**: Explicit checks

**Code reduction**: 67% with Dingo!

## Best Practices

### ✅ DO

- Use `?` operator for error propagation
- Import libraries as normal Go packages
- Let type inference work (don't over-specify types)
- Keep application `.dingo` files (don't gitignore)
- Transpile during build/deployment

### ❌ DON'T

- Manually edit transpiled `.go` files
- Commit `.go` files for applications
- Try to use Dingo syntax on pure Go libraries (won't work)
- Mix Result types from different sources

## Development Workflow

### Typical Day

1. **Edit**: Modify `calculator.dingo`
2. **Transpile**: Run `make run` or `dingo build`
3. **Test**: Application uses library seamlessly
4. **Deploy**: Build binary with `make build`

### Updating the Library

```bash
# If library is published on GitHub
go get github.com/example/mathutils@v1.2.0
go mod tidy

# Rebuild
make build
```

### Local Development with Library Changes

```bash
# Make changes to library
cd ../library-example
vim mathutils.dingo
dingo build mathutils.dingo

# Return to app
cd ../hybrid-example
make run  # Uses updated library via replace directive
```

## Production Deployment

### Dockerfile

```dockerfile
FROM golang:1.21 AS builder

# Install Dingo
RUN go install github.com/dingo-lang/dingo/cmd/dingo@latest

WORKDIR /app

# Copy application
COPY go.mod go.sum ./
RUN go mod download

COPY *.dingo ./

# Transpile
RUN dingo build ./...

# Build binary
RUN go build -o calculator calculator.go

# Runtime image (no Dingo needed)
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/calculator .

ENTRYPOINT ["/app/calculator"]
```

## Key Takeaways

1. **Seamless interop**: Dingo apps consume Dingo libraries naturally
2. **Zero setup**: Libraries work via `go get`, no special tooling
3. **Best of both worlds**: Library compatibility + app developer experience
4. **Type safety**: Full type inference across boundaries
5. **Production ready**: Compiles to pure Go, no runtime dependencies

## Next Steps

- Try modifying the calculator to add more operations
- Create your own Dingo library following `library-example/`
- Publish your library and use it in applications
- Read `docs/package-management.md` for complete guide

.PHONY: help build run test clean

# Default target
help:
	@echo "TODO CLI - Makefile commands:"
	@echo ""
	@echo "  make build    - Transpile and build production binary"
	@echo "  make run      - Transpile and run application"
	@echo "  make test     - Run tests"
	@echo "  make clean    - Remove generated files"
	@echo ""

# Transpile and build production binary
build:
	@echo "Transpiling Dingo files..."
	dingo build ./...
	@echo "Building binary..."
	go build -o todo-cli main.go
	@echo "✓ Built: ./todo-cli"

# Transpile and run for development
run:
	@echo "Transpiling Dingo files..."
	dingo build ./...
	@echo "Running application..."
	go run main.go

# Run tests
test:
	@echo "Transpiling Dingo files..."
	dingo build ./...
	@echo "Running tests..."
	go test -v ./...

# Clean generated files
clean:
	@echo "Cleaning generated files..."
	rm -f *.go
	rm -f *.go.golden.map
	rm -f todo-cli
	@echo "✓ Cleaned"

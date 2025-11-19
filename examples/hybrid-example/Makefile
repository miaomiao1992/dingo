.PHONY: help build run clean

# Default target
help:
	@echo "Calculator - Makefile commands:"
	@echo ""
	@echo "  make build    - Transpile and build production binary"
	@echo "  make run      - Transpile and run calculator"
	@echo "  make clean    - Remove generated files"
	@echo ""

# Transpile and build production binary
build:
	@echo "Transpiling Dingo files..."
	dingo build calculator.dingo
	@echo "Building binary..."
	go build -o calculator calculator.go
	@echo "✓ Built: ./calculator"

# Transpile and run for development
run:
	@echo "Transpiling Dingo files..."
	dingo build calculator.dingo
	@echo "Running calculator..."
	go run calculator.go

# Clean generated files
clean:
	@echo "Cleaning generated files..."
	rm -f *.go
	rm -f *.go.golden.map
	rm -f calculator
	@echo "✓ Cleaned"

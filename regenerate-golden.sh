#!/bin/bash
# regenerate-golden.sh - Regenerate golden test files
# Usage: ./regenerate-golden.sh [pattern]
# Example: ./regenerate-golden.sh null_coalesce  # Only regenerate null_coalesce tests

set -e

# Color output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Build dingo if not present or out of date
if [ ! -f ./dingo ] || [ ./cmd/dingo/main.go -nt ./dingo ]; then
    echo "Building dingo transpiler..."
    go build -o dingo ./cmd/dingo
fi

# Pattern filter (optional)
PATTERN="${1:-*.dingo}"
if [[ "$PATTERN" != *.dingo ]]; then
    PATTERN="${PATTERN}*.dingo"
fi

# Counters
SUCCESS=0
FAILED=0
SKIPPED=0

# Create temp directory for output
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

echo "Regenerating golden test files matching: $PATTERN"
echo "========================================"

for dingo_file in tests/golden/$PATTERN; do
    if [ ! -f "$dingo_file" ]; then
        continue
    fi

    base=$(basename "$dingo_file" .dingo)
    golden_file="tests/golden/${base}.go.golden"

    echo -n "Regenerating ${base}... "

    # Build to temp directory
    if ./dingo build "$dingo_file" -o "$TEMP_DIR/${base}.go" 2>&1 | grep -q "error"; then
        echo -e "${RED}✗ FAILED${NC}"
        FAILED=$((FAILED + 1))
        continue
    fi

    # Check if output was generated
    if [ ! -f "$TEMP_DIR/${base}.go" ]; then
        echo -e "${YELLOW}⊘ SKIPPED (no output)${NC}"
        SKIPPED=$((SKIPPED + 1))
        continue
    fi

    # Move to golden location
    mv "$TEMP_DIR/${base}.go" "$golden_file"

    # Move source map if present
    if [ -f "$TEMP_DIR/${base}.go.map" ]; then
        mv "$TEMP_DIR/${base}.go.map" "${golden_file}.map"
    fi

    # Verify no machine directives remain
    if grep -q "__INFER__\|__UNWRAP__\|__IS_SOME__\|__SAFE_NAV" "$golden_file"; then
        echo -e "${YELLOW}⚠ WARNING: Machine directives still present${NC}"
    fi

    echo -e "${GREEN}✓ SUCCESS${NC}"
    SUCCESS=$((SUCCESS + 1))
done

echo "========================================"
echo -e "${GREEN}Success: $SUCCESS${NC}"
echo -e "${RED}Failed:  $FAILED${NC}"
echo -e "${YELLOW}Skipped: $SKIPPED${NC}"
echo "========================================"

if [ $SUCCESS -gt 0 ]; then
    echo ""
    echo "Verifying machine directives removed:"
    REMAINING=$(grep -l "__INFER__\|__UNWRAP__\|__IS_SOME__\|__SAFE_NAV" tests/golden/*.go.golden 2>/dev/null | wc -l | tr -d ' ')
    if [ "$REMAINING" -eq "0" ]; then
        echo -e "${GREEN}✓ All machine directives removed!${NC}"
    else
        echo -e "${YELLOW}⚠ $REMAINING files still have directives${NC}"
        grep -l "__INFER__\|__UNWRAP__\|__IS_SOME__\|__SAFE_NAV" tests/golden/*.go.golden 2>/dev/null
    fi
fi

#!/bin/bash

# Regenerate all golden test files
# Track success and failures

SUCCESS=0
FAILED=0
FAILED_FILES=""

echo "Regenerating golden test files..."
echo ""

for dingo_file in tests/golden/*.dingo; do
    base_name="${dingo_file%.dingo}"
    golden_file="${base_name}.go.golden"

    echo -n "Processing $(basename "$dingo_file")... "

    if ./dingo build "$dingo_file" -o "$golden_file" 2>&1 | grep -q "Success"; then
        echo "✓"
        SUCCESS=$((SUCCESS + 1))
    else
        echo "✗"
        FAILED=$((FAILED + 1))
        FAILED_FILES="$FAILED_FILES\n  - $(basename "$dingo_file")"
    fi
done

echo ""
echo "================================"
echo "Summary:"
echo "  Succeeded: $SUCCESS"
echo "  Failed: $FAILED"

if [ $FAILED -gt 0 ]; then
    echo ""
    echo "Failed files:"
    echo -e "$FAILED_FILES"
fi

echo ""
echo "Total: $((SUCCESS + FAILED)) files processed"

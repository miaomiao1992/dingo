# Task A: Fix CRITICAL #1 - Source-Map Offset Bug

## Status: Already Fixed

The bug was already resolved in the current codebase.

## Changes Found

**File:** `pkg/preprocessor/preprocessor.go:183-192`

The `adjustMappingsForImports` function already contains the fix:
- Only shifts mappings where `GeneratedLine >= importInsertLine`
- Mappings before import block are preserved

## Test Results

All tests passing:
- TestSourceMappingWithImports - Validates the fix
- 8/8 preprocessor tests passing

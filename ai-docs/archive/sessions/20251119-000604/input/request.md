# Task
Design package-wide scanning approach for unqualified import inference in Dingo.

## Requirements
1. Scan all .dingo files in package to collect local function identifiers.
2. Share exclusion list across file transpilation to prevent auto-qualifying local functions.
3. Stay under 500ms per package for watch mode.
4. Support incremental builds when only subset of files change.
5. Integrate with existing preprocessor / ImportTracker architecture.

## Deliverables
- Architecture description
- Caching strategy
- Incremental build handling
- Performance analysis
- Implementation plan
- Trade-offs / edge cases

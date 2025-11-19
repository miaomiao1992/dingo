# User Request

Focus on the `tests/golden/` folder. Check that all files are good test examples and verify they are not failing.

If tests fail:
1. Run parallel external reviewers using golang-architect agents
2. Use one agent per recommended language model (from our skills)
3. Investigate why tests are failing
4. Consolidate outcomes into an implementation plan

If 80%+ confidence on what needs to be fixed:
- Execute fixes without user permission
- Run reviews again to verify
- Iterate (developer → review) until satisfied

If something tricky requiring user attention:
- Stop and provide full description of the issue

## Test Results

**Status**: Most golden tests are failing

**Common Issues**:
1. Import formatting mismatch (multi-line vs single-line)
2. Missing `// dingo:s:N` comments in transpiler output
3. Golden files have duplicate `// dingo:s:N` comments
4. Golden files use unqualified function calls (won't compile)

**Critical Question**: Are the golden files wrong, or is the transpiler wrong?

**Initial Assessment**: Transpiler output appears MORE correct (compilable Go with proper imports).

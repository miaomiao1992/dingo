package preprocessor

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

// RustMatchProcessor handles Rust-like pattern matching syntax
// Transforms: match expr { Pattern => expression, ... } → Go switch statement with markers
type RustMatchProcessor struct {
	matchCounter int
	mappings     []Mapping
}

// Pattern-matching regex for Rust-like match expressions
var (
	// Match the entire match expression: match expr { ... }
	matchExprPattern = regexp.MustCompile(`(?s)match\s+([^{]+)\s*\{(.+)\}`)
)

// NewRustMatchProcessor creates a new Rust-like match preprocessor
func NewRustMatchProcessor() *RustMatchProcessor {
	return &RustMatchProcessor{
		matchCounter: 0,
		mappings:     []Mapping{},
	}
}

// Name returns the processor name
func (r *RustMatchProcessor) Name() string {
	return "rust_match"
}

// Process transforms Rust-like match expressions
func (r *RustMatchProcessor) Process(source []byte) ([]byte, []Mapping, error) {
	r.mappings = []Mapping{}
	r.matchCounter = 0

	input := string(source)
	lines := strings.Split(input, "\n")

	var output bytes.Buffer
	inputLineNum := 0
	outputLineNum := 1

	for inputLineNum < len(lines) {
		line := lines[inputLineNum]

		// Check if this line starts a match expression (not in comments)
		// Must be "match " followed by identifier (not in middle of word, not in comment)
		trimmed := strings.TrimSpace(line)
		isMatchExpr := false
		if !strings.HasPrefix(trimmed, "//") && !strings.HasPrefix(trimmed, "/*") {
			// Look for standalone "match " keyword (not part of another word)
			if strings.Contains(line, "match ") {
				// Simple heuristic: check it's not in the middle of a word
				// by verifying character before "match" is whitespace or punctuation
				idx := strings.Index(line, "match ")
				if idx == 0 || !isAlphanumeric(rune(line[idx-1])) {
					isMatchExpr = true
				}
			}
		}

		if isMatchExpr {
			// Collect the entire match expression (may span multiple lines)
			matchExpr, linesConsumed := r.collectMatchExpression(lines, inputLineNum)
			if matchExpr != "" {
				// Transform the match expression
				transformed, newMappings, err := r.transformMatch(matchExpr, inputLineNum+1, outputLineNum)
				if err != nil {
					return nil, nil, fmt.Errorf("line %d: %w", inputLineNum+1, err)
				}

				output.WriteString(transformed)
				r.mappings = append(r.mappings, newMappings...)

				// Update line counters
				inputLineNum += linesConsumed
				outputLineNum += strings.Count(transformed, "\n")

				// Add newline if not at end
				if inputLineNum < len(lines) {
					output.WriteByte('\n')
					outputLineNum++
				}
				continue
			}
		}

		// Not a match expression, pass through as-is
		output.WriteString(line)
		if inputLineNum < len(lines)-1 {
			output.WriteByte('\n')
		}
		inputLineNum++
		outputLineNum++
	}

	return output.Bytes(), r.mappings, nil
}

// isAlphanumeric checks if a rune is alphanumeric or underscore
func isAlphanumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}

// collectMatchExpression collects a complete match expression across multiple lines
// Returns: (matchExpression, linesConsumed)
func (r *RustMatchProcessor) collectMatchExpression(lines []string, startLine int) (string, int) {
	var buf bytes.Buffer
	braceDepth := 0
	linesConsumed := 0
	foundMatch := false

	for i := startLine; i < len(lines); i++ {
		line := lines[i]
		buf.WriteString(line)
		linesConsumed++

		// Track brace depth
		for _, ch := range line {
			if ch == '{' {
				braceDepth++
				foundMatch = true
			} else if ch == '}' {
				braceDepth--
				if braceDepth == 0 && foundMatch {
					// Complete match expression
					return buf.String(), linesConsumed
				}
			}
		}

		// Add newline if more lines to come (C7 FIX: Preserve newlines for proper formatting)
		if i < len(lines)-1 {
			buf.WriteByte('\n')
		}
	}

	// Incomplete match expression (missing closing brace)
	return "", 0
}

// transformMatch transforms a Rust-like match expression to Go switch
func (r *RustMatchProcessor) transformMatch(matchExpr string, originalLine int, outputLine int) (string, []Mapping, error) {
	// Extract scrutinee and arms
	matches := matchExprPattern.FindStringSubmatch(matchExpr)
	if len(matches) < 3 {
		return "", nil, fmt.Errorf("invalid match expression syntax")
	}

	scrutinee := strings.TrimSpace(matches[1])
	armsText := matches[2]

	// Check if match expression is in assignment context and extract variable name
	// If the entire match expression starts with "let x = match" or "var x = match",
	// we need to handle it specially using Variable Hoisting pattern
	isInAssignment, assignmentVar := r.extractAssignmentVar(matchExpr)

	// Check if scrutinee is a tuple expression
	isTuple, tupleElements, err := r.detectTuple(scrutinee)
	if err != nil {
		return "", nil, err
	}

	if isTuple {
		// Parse tuple pattern arms
		tupleArms, err := r.parseTupleArms(armsText)
		if err != nil {
			return "", nil, fmt.Errorf("parsing tuple pattern arms: %w", err)
		}

		// Generate tuple match (elements extraction + pattern info)
		result, mappings := r.generateTupleMatch(tupleElements, tupleArms, originalLine, outputLine)
		return result, mappings, nil
	}

	// Parse pattern arms (non-tuple)
	arms, err := r.parseArms(armsText)
	if err != nil {
		return "", nil, fmt.Errorf("parsing pattern arms: %w", err)
	}

	// Generate Go switch statement
	result, mappings := r.generateSwitch(scrutinee, arms, originalLine, outputLine, isInAssignment, assignmentVar)
	return result, mappings, nil
}

// extractAssignmentVar extracts the variable name if match is in assignment context
// Returns: (isInAssignment, variableName)
// Examples:
//   "let x = match ..." -> (true, "x")
//   "var result = match ..." -> (true, "result")
//   "match ..." -> (false, "")
func (r *RustMatchProcessor) extractAssignmentVar(matchExpr string) (bool, string) {
	// Get the text before "match" keyword
	matchIdx := strings.Index(matchExpr, "match")
	if matchIdx == -1 {
		return false, ""
	}

	beforeMatch := strings.TrimSpace(matchExpr[:matchIdx])

	// Check if there's an assignment operator before match
	if !strings.Contains(beforeMatch, "=") {
		return false, ""
	}

	// Extract variable name from patterns like:
	//   "let x ="
	//   "var result ="
	//   "x :="

	// Remove any "let" or "var" keywords
	beforeMatch = strings.TrimPrefix(beforeMatch, "let")
	beforeMatch = strings.TrimPrefix(beforeMatch, "var")
	beforeMatch = strings.TrimSpace(beforeMatch)

	// Now we should have something like "x =" or "result ="
	// Remove the "=" and any ":="
	beforeMatch = strings.TrimSuffix(beforeMatch, "=")
	beforeMatch = strings.TrimSuffix(beforeMatch, ":")
	beforeMatch = strings.TrimSpace(beforeMatch)

	// What remains should be the variable name
	varName := beforeMatch
	if varName == "" {
		return false, ""
	}

	return true, varName
}

// patternArm represents a single pattern arm
type patternArm struct {
	pattern    string // Ok(x), Err(e), Some(v), None, _
	binding    string // x, e, v, etc. (empty for None and _)
	guard      string // Guard condition: "x > 0" (optional, empty if no guard)
	expression string // the expression to execute
}

// parseArms parses pattern arms from the match body
// Handles both simple and block expressions:
//   Ok(x) => x * 2,
//   Err(e) => { log(e); return 0 }
func (r *RustMatchProcessor) parseArms(armsText string) ([]patternArm, error) {
	arms := []patternArm{}
	text := strings.TrimSpace(armsText)

	// Parse arms manually to handle nested braces
	i := 0
	for i < len(text) {
		// Skip whitespace
		for i < len(text) && (text[i] == ' ' || text[i] == '\t' || text[i] == '\n' || text[i] == '\r') {
			i++
		}
		if i >= len(text) {
			break
		}

		// Extract pattern + optional guard (everything before =>)
		arrowPos := strings.Index(text[i:], "=>")
		if arrowPos == -1 {
			break // No more arms
		}

		patternAndGuard := strings.TrimSpace(text[i : i+arrowPos])
		i += arrowPos + 2 // Skip =>

		// Split pattern from guard if present
		// Guard syntax: Pattern if condition => expr
		//           or: Pattern where condition => expr
		pattern, guard := r.splitPatternAndGuard(patternAndGuard)

		// Skip whitespace after =>
		for i < len(text) && (text[i] == ' ' || text[i] == '\t' || text[i] == '\n' || text[i] == '\r') {
			i++
		}
		if i >= len(text) {
			return nil, fmt.Errorf("unexpected end after =>")
		}

		// Extract expression (until comma or end)
		var expr string
		if text[i] == '{' {
			// Block expression - find matching }
			braceCount := 1
			start := i
			i++
			for i < len(text) && braceCount > 0 {
				if text[i] == '{' {
					braceCount++
				} else if text[i] == '}' {
					braceCount--
				}
				i++
			}
			expr = strings.TrimSpace(text[start:i])
		} else {
			// Simple expression - find comma or end
			start := i
			for i < len(text) && text[i] != ',' {
				i++
			}
			expr = strings.TrimSpace(text[start:i])
		}

		// Skip comma if present
		if i < len(text) && text[i] == ',' {
			i++
		}

		// Extract binding from pattern (if present)
		binding := ""
		patternName := pattern
		if strings.Contains(pattern, "(") {
			start := strings.Index(pattern, "(")
			end := strings.Index(pattern, ")")
			if end > start {
				binding = strings.TrimSpace(pattern[start+1 : end])
				patternName = pattern[:start]
			}
		}

		arms = append(arms, patternArm{
			pattern:    patternName,
			binding:    binding,
			guard:      guard,
			expression: expr,
		})
	}

	if len(arms) == 0 {
		return nil, fmt.Errorf("no pattern arms found")
	}

	return arms, nil
}

// splitPatternAndGuard splits a pattern arm into pattern and optional guard
// Supports only 'if' guard keyword (Swift 'where' removed in Phase 4.2)
// Examples:
//   "Ok(x) if x > 0" -> ("Ok(x)", "x > 0")
//   "Ok(x)" -> ("Ok(x)", "")
func (r *RustMatchProcessor) splitPatternAndGuard(patternAndGuard string) (pattern string, guard string) {
	// Strategy: Look for guard keyword (" if ") that comes after a complete pattern
	// Pattern formats:
	//   - Ok(binding)   - ends with )
	//   - None          - bare identifier
	//   - _             - wildcard

	// We need to find " if " (with surrounding spaces) that appears after the pattern
	// To avoid false matches like "diff" containing "if", we require the keyword
	// to be surrounded by spaces

	var guardPos int = -1
	var guardKeywordLen int = 0

	// Find " if "
	idx := strings.Index(patternAndGuard, " if ")
	if idx != -1 {
		// Found " if " - this could be the guard
		// Validate it's after a complete pattern by checking what comes before
		before := patternAndGuard[:idx]
		if r.isCompletePattern(before) {
			guardPos = idx
			guardKeywordLen = 4 // len(" if ")
		}
	}

	if guardPos == -1 {
		// No guard found
		return strings.TrimSpace(patternAndGuard), ""
	}

	// Split at the guard keyword
	pattern = strings.TrimSpace(patternAndGuard[:guardPos])
	guard = strings.TrimSpace(patternAndGuard[guardPos+guardKeywordLen:])

	return pattern, guard
}

// isCompletePattern checks if a string looks like a complete pattern
// (ends with ) for binding patterns, or is a bare identifier/wildcard)
func (r *RustMatchProcessor) isCompletePattern(s string) bool {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return false
	}

	// Pattern with binding: Ok(x), Some(value), etc - must end with )
	if strings.Contains(trimmed, "(") {
		return strings.HasSuffix(trimmed, ")")
	}

	// Bare pattern: None, _, Active, etc - just an identifier
	// Check it's not part of a larger expression (no operators like <, >, &&, etc)
	return true
}

// generateSwitch generates Go switch statement with DINGO_MATCH markers
func (r *RustMatchProcessor) generateSwitch(scrutinee string, arms []patternArm, originalLine int, outputLine int, isInAssignment bool, assignmentVar string) (string, []Mapping) {
	var buf bytes.Buffer
	mappings := []Mapping{}

	matchID := r.matchCounter
	r.matchCounter++

	// Create temporary variable for scrutinee
	scrutineeVar := fmt.Sprintf("__match_%d", matchID)

	// Variable Hoisting Pattern: If in assignment context, declare result variable with proper type
	if isInAssignment && assignmentVar != "" {
		resultType := r.inferMatchResultType(arms)

		// Line 1: Declare result variable with proper type
		buf.WriteString(fmt.Sprintf("var %s %s\n", assignmentVar, resultType))
		mappings = append(mappings, Mapping{
			OriginalLine:    originalLine,
			OriginalColumn:  1,
			GeneratedLine:   outputLine,
			GeneratedColumn: 1,
			Length:          5,
			Name:            "rust_match",
		})
		outputLine++
	}

	// Line 2: DINGO_MATCH_START marker (MUST BE BEFORE temp var)
	buf.WriteString(fmt.Sprintf("// DINGO_MATCH_START: %s\n", scrutinee))
	mappings = append(mappings, Mapping{
		OriginalLine:    originalLine,
		OriginalColumn:  1,
		GeneratedLine:   outputLine,
		GeneratedColumn: 1,
		Length:          5,
		Name:            "rust_match",
	})
	outputLine++

	// Line 3: Store scrutinee in temporary variable
	buf.WriteString(fmt.Sprintf("%s := %s\n", scrutineeVar, scrutinee))
	mappings = append(mappings, Mapping{
		OriginalLine:    originalLine,
		OriginalColumn:  1,
		GeneratedLine:   outputLine,
		GeneratedColumn: 1,
		Length:          5, // "match"
		Name:            "rust_match",
	})
	outputLine++

	// Line 4: switch statement opening (tag-based switch - CORRECT pattern)
	buf.WriteString(fmt.Sprintf("switch %s.tag {\n", scrutineeVar))
	mappings = append(mappings, Mapping{
		OriginalLine:    originalLine,
		OriginalColumn:  1,
		GeneratedLine:   outputLine,
		GeneratedColumn: 1,
		Length:          5,
		Name:            "rust_match",
	})
	outputLine++

	// Generate case statements for each arm
	for _, arm := range arms {
		caseLines, caseMappings := r.generateCase(scrutineeVar, arm, originalLine, outputLine, isInAssignment, assignmentVar)
		buf.WriteString(caseLines)
		mappings = append(mappings, caseMappings...)
		outputLine += strings.Count(caseLines, "\n")
	}

	// Closing brace for switch
	buf.WriteString("}\n")
	mappings = append(mappings, Mapping{
		OriginalLine:    originalLine,
		OriginalColumn:  1,
		GeneratedLine:   outputLine,
		GeneratedColumn: 1,
		Length:          1,
		Name:            "rust_match",
	})
	outputLine++

	// DINGO_MATCH_END marker
	buf.WriteString("// DINGO_MATCH_END\n")
	mappings = append(mappings, Mapping{
		OriginalLine:    originalLine,
		OriginalColumn:  1,
		GeneratedLine:   outputLine,
		GeneratedColumn: 1,
		Length:          1,
		Name:            "rust_match",
	})

	return buf.String(), mappings
}

// inferMatchResultType infers the result type from match arms
// For now, uses simple heuristics based on first arm's pattern
func (r *RustMatchProcessor) inferMatchResultType(arms []patternArm) string {
	if len(arms) == 0 {
		return "interface{}" // Fallback
	}

	// Look at first arm's pattern to infer type
	firstPattern := arms[0].pattern
	switch firstPattern {
	case "Ok", "Err":
		// Result type - need to infer T and E types
		// For now, return generic Result type placeholder
		// TODO: Parse arm expressions to infer exact types
		return "Result_int_error" // Simplified for now
	case "Some", "None":
		// Option type - need to infer T type
		return "Option_int" // Simplified for now
	default:
		// Custom enum or unknown
		return "interface{}"
	}
}

// generateCase generates a single case statement
func (r *RustMatchProcessor) generateCase(scrutineeVar string, arm patternArm, originalLine int, outputLine int, isInAssignment bool, assignmentVar string) (string, []Mapping) {
	var buf bytes.Buffer
	mappings := []Mapping{}

	// Handle wildcard pattern
	if arm.pattern == "_" {
		buf.WriteString("default:\n")
		buf.WriteString(fmt.Sprintf("\t// DINGO_PATTERN: _\n"))

		// Variable Hoisting: Assign to result variable if in assignment context
		if isInAssignment && assignmentVar != "" {
			buf.WriteString(fmt.Sprintf("\t%s = %s\n", assignmentVar, arm.expression))
		} else {
			buf.WriteString(fmt.Sprintf("\t%s\n", arm.expression))
		}

		mappings = append(mappings, Mapping{
			OriginalLine:    originalLine,
			OriginalColumn:  1,
			GeneratedLine:   outputLine,
			GeneratedColumn: 1,
			Length:          1,
			Name:            "rust_match_arm",
		})
		return buf.String(), mappings
	}

	// Generate case tag (tag-based case - CORRECT pattern)
	tagName := r.getTagName(arm.pattern)
	buf.WriteString(fmt.Sprintf("case %s:\n", tagName))
	mappings = append(mappings, Mapping{
		OriginalLine:    originalLine,
		OriginalColumn:  1,
		GeneratedLine:   outputLine,
		GeneratedColumn: 1,
		Length:          len(arm.pattern),
		Name:            "rust_match_arm",
	})
	outputLine++

	// DINGO_PATTERN marker
	patternStr := arm.pattern
	if arm.binding != "" {
		patternStr = fmt.Sprintf("%s(%s)", arm.pattern, arm.binding)
	}
	buf.WriteString(fmt.Sprintf("\t// DINGO_PATTERN: %s", patternStr))

	// Add DINGO_GUARD marker if guard present
	if arm.guard != "" {
		buf.WriteString(fmt.Sprintf(" | DINGO_GUARD: %s", arm.guard))
	}
	buf.WriteString("\n")

	mappings = append(mappings, Mapping{
		OriginalLine:    originalLine,
		OriginalColumn:  1,
		GeneratedLine:   outputLine,
		GeneratedColumn: 1,
		Length:          len(patternStr),
		Name:            "rust_match_arm",
	})
	outputLine++

	// Extract binding if present
	if arm.binding != "" {
		// Generate binding extraction
		// For Ok(x): x := *scrutinee.ok_0
		// For Err(e): e := scrutinee.err_0
		// For Some(v): v := *scrutinee.some_0
		bindingCode := r.generateBinding(scrutineeVar, arm.pattern, arm.binding)
		buf.WriteString(fmt.Sprintf("\t%s\n", bindingCode))
		mappings = append(mappings, Mapping{
			OriginalLine:    originalLine,
			OriginalColumn:  1,
			GeneratedLine:   outputLine,
			GeneratedColumn: 1,
			Length:          len(arm.binding),
			Name:            "rust_match_binding",
		})
		outputLine++
	}

	// Pattern arm expression (C7 FIX: Handle block expressions properly)
	// Variable Hoisting: If in assignment context, assign to result variable
	exprStr := arm.expression
	if strings.HasPrefix(exprStr, "{") && strings.HasSuffix(exprStr, "}") {
		// Block expression: remove outer braces and preserve formatting
		innerBlock := strings.TrimSpace(exprStr[1 : len(exprStr)-1])
		formatted := r.formatBlockStatements(innerBlock)
		for _, line := range strings.Split(formatted, "\n") {
			if trimmed := strings.TrimSpace(line); trimmed != "" {
				buf.WriteString(fmt.Sprintf("\t%s\n", trimmed))
			}
		}
	} else {
		// Simple expression
		if isInAssignment && assignmentVar != "" {
			// Variable Hoisting: Assign to result variable instead of returning
			buf.WriteString(fmt.Sprintf("\t%s = %s\n", assignmentVar, exprStr))
		} else {
			// Not in assignment context: keep expression as-is
			buf.WriteString(fmt.Sprintf("\t%s\n", exprStr))
		}
	}

	mappings = append(mappings, Mapping{
		OriginalLine:    originalLine,
		OriginalColumn:  1,
		GeneratedLine:   outputLine,
		GeneratedColumn: 1,
		Length:          len(arm.expression),
		Name:            "rust_match_expr",
	})

	return buf.String(), mappings
}

// getTagName converts pattern name to Go tag constant name
// Ok → ResultTagOk, Err → ResultTagErr, Some → OptionTagSome, None → OptionTagNone
// Status_Pending → StatusTag_Pending (for custom enums)
func (r *RustMatchProcessor) getTagName(pattern string) string {
	switch pattern {
	case "Ok":
		return "ResultTagOk"
	case "Err":
		return "ResultTagErr"
	case "Some":
		return "OptionTagSome"
	case "None":
		return "OptionTagNone"
	default:
		// Custom enum variant: EnumName_Variant → EnumNameTag_Variant
		// Example: Status_Pending → StatusTag_Pending
		if idx := strings.Index(pattern, "_"); idx > 0 {
			enumName := pattern[:idx]
			variantName := pattern[idx:] // includes the underscore
			return enumName + "Tag" + variantName
		}
		// Bare variant name (shouldn't happen in well-formed Dingo code)
		return pattern + "Tag"
	}
}

// generateBinding generates binding extraction code
func (r *RustMatchProcessor) generateBinding(scrutinee string, pattern string, binding string) string {
	switch pattern {
	case "Ok":
		// For Result<T,E>, Ok value is stored in ok_0 field (pointer to T)
		return fmt.Sprintf("%s := *%s.ok_0", binding, scrutinee)
	case "Err":
		// For Result<T,E>, Err value is stored in err_0 field (E)
		return fmt.Sprintf("%s := %s.err_0", binding, scrutinee)
	case "Some":
		// For Option<T>, Some value is stored in some_0 field (pointer to T)
		return fmt.Sprintf("%s := *%s.some_0", binding, scrutinee)
	default:
		// Custom enum variant: assume field name is lowercased pattern name + _0
		fieldName := strings.ToLower(pattern) + "_0"
		return fmt.Sprintf("%s := %s.%s", binding, scrutinee, fieldName)
	}
}

// formatBlockStatements formats block statements preserving newlines
func (r *RustMatchProcessor) formatBlockStatements(block string) string {
	// Newlines are now preserved, just return the block as-is
	return block
}

// GetNeededImports implements the ImportProvider interface
func (r *RustMatchProcessor) GetNeededImports() []string {
	// Rust match syntax doesn't require additional imports
	return []string{}
}

// detectTuple checks if scrutinee is a tuple expression: (expr1, expr2, ...)
// Returns: (isTuple, elements, error)
func (r *RustMatchProcessor) detectTuple(scrutinee string) (bool, []string, error) {
	trimmed := strings.TrimSpace(scrutinee)

	// Must start/end with parens
	if !strings.HasPrefix(trimmed, "(") || !strings.HasSuffix(trimmed, ")") {
		return false, nil, nil // Not a tuple
	}

	// Parse elements
	inner := trimmed[1 : len(trimmed)-1]
	elements := r.splitTupleElements(inner)

	// Enforce 6-element limit (USER DECISION)
	if len(elements) > 6 {
		return false, nil, fmt.Errorf(
			"tuple patterns limited to 6 elements (found %d)",
			len(elements),
		)
	}

	// Must have at least 2 elements to be a tuple
	if len(elements) < 2 {
		return false, nil, nil
	}

	return true, elements, nil
}

// splitTupleElements splits tuple elements on commas (respects nested parens/brackets)
func (r *RustMatchProcessor) splitTupleElements(s string) []string {
	var elements []string
	var current strings.Builder
	depth := 0

	for _, ch := range s {
		switch ch {
		case '(', '[', '{':
			depth++
			current.WriteRune(ch)
		case ')', ']', '}':
			depth--
			current.WriteRune(ch)
		case ',':
			if depth == 0 {
				elements = append(elements, strings.TrimSpace(current.String()))
				current.Reset()
			} else {
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		elements = append(elements, strings.TrimSpace(current.String()))
	}

	return elements
}

// tuplePatternArm represents a tuple pattern arm
type tuplePatternArm struct {
	patterns   []tupleElementPattern // One per tuple element
	guard      string                // Guard condition (optional)
	expression string                // Expression to execute
}

// tupleElementPattern represents one element in a tuple pattern
type tupleElementPattern struct {
	variant string // Ok, Err, Some, None, _ (wildcard)
	binding string // x, e, v (optional - empty for None/_)
}

// parseTupleArms parses tuple pattern arms from match body
// Example: (Ok(x), Err(e)) => expr1, (Ok(a), Ok(b)) if guard => expr2
func (r *RustMatchProcessor) parseTupleArms(armsText string) ([]tuplePatternArm, error) {
	arms := []tuplePatternArm{}
	text := strings.TrimSpace(armsText)

	i := 0
	for i < len(text) {
		// Skip whitespace
		for i < len(text) && (text[i] == ' ' || text[i] == '\t' || text[i] == '\n' || text[i] == '\r') {
			i++
		}
		if i >= len(text) {
			break
		}

		// Expect tuple pattern: (Pattern1, Pattern2, ...)
		if text[i] != '(' {
			return nil, fmt.Errorf("expected tuple pattern at position %d", i)
		}

		// Find matching close paren
		parenDepth := 1
		tupleStart := i
		i++
		for i < len(text) && parenDepth > 0 {
			if text[i] == '(' {
				parenDepth++
			} else if text[i] == ')' {
				parenDepth--
			}
			i++
		}
		tuplePatternStr := text[tupleStart:i]

		// Parse tuple elements
		tupleElements, err := r.parseTuplePattern(tuplePatternStr)
		if err != nil {
			return nil, fmt.Errorf("parsing tuple pattern: %w", err)
		}

		// Skip whitespace
		for i < len(text) && (text[i] == ' ' || text[i] == '\t' || text[i] == '\n' || text[i] == '\r') {
			i++
		}

		// Check for guard (if)
		guard := ""
		if i < len(text) && strings.HasPrefix(text[i:], "if ") {
			i += 3 // skip "if "

			// Extract guard condition (until =>)
			arrowPos := strings.Index(text[i:], "=>")
			if arrowPos == -1 {
				return nil, fmt.Errorf("expected => after guard")
			}
			guard = strings.TrimSpace(text[i : i+arrowPos])
			i += arrowPos
		}

		// Expect =>
		if !strings.HasPrefix(text[i:], "=>") {
			return nil, fmt.Errorf("expected => at position %d", i)
		}
		i += 2

		// Skip whitespace after =>
		for i < len(text) && (text[i] == ' ' || text[i] == '\t' || text[i] == '\n' || text[i] == '\r') {
			i++
		}

		// Extract expression (until comma or end)
		var expr string
		if text[i] == '{' {
			// Block expression
			braceCount := 1
			start := i
			i++
			for i < len(text) && braceCount > 0 {
				if text[i] == '{' {
					braceCount++
				} else if text[i] == '}' {
					braceCount--
				}
				i++
			}
			expr = strings.TrimSpace(text[start:i])
		} else {
			// Simple expression
			start := i
			for i < len(text) && text[i] != ',' {
				i++
			}
			expr = strings.TrimSpace(text[start:i])
		}

		// Skip comma
		if i < len(text) && text[i] == ',' {
			i++
		}

		arms = append(arms, tuplePatternArm{
			patterns:   tupleElements,
			guard:      guard,
			expression: expr,
		})
	}

	if len(arms) == 0 {
		return nil, fmt.Errorf("no tuple pattern arms found")
	}

	return arms, nil
}

// parseTuplePattern parses a single tuple pattern: (Ok(x), Err(e), _)
func (r *RustMatchProcessor) parseTuplePattern(tupleStr string) ([]tupleElementPattern, error) {
	// Remove outer parens
	tupleStr = strings.TrimSpace(tupleStr)
	if !strings.HasPrefix(tupleStr, "(") || !strings.HasSuffix(tupleStr, ")") {
		return nil, fmt.Errorf("invalid tuple pattern: %s", tupleStr)
	}
	inner := tupleStr[1 : len(tupleStr)-1]

	// Split on commas (respecting nested parens)
	elementStrs := r.splitTupleElements(inner)

	elements := make([]tupleElementPattern, len(elementStrs))
	for i, elemStr := range elementStrs {
		elemStr = strings.TrimSpace(elemStr)

		// Wildcard
		if elemStr == "_" {
			elements[i] = tupleElementPattern{
				variant: "_",
				binding: "",
			}
			continue
		}

		// Pattern with binding: Ok(x), Err(e), Some(v)
		if strings.Contains(elemStr, "(") {
			start := strings.Index(elemStr, "(")
			end := strings.Index(elemStr, ")")
			if end <= start {
				return nil, fmt.Errorf("invalid pattern: %s", elemStr)
			}
			variant := strings.TrimSpace(elemStr[:start])
			binding := strings.TrimSpace(elemStr[start+1 : end])
			elements[i] = tupleElementPattern{
				variant: variant,
				binding: binding,
			}
		} else {
			// Pattern without binding: None
			elements[i] = tupleElementPattern{
				variant: elemStr,
				binding: "",
			}
		}
	}

	return elements, nil
}

// generateTupleMatch generates Go code for tuple pattern matching
func (r *RustMatchProcessor) generateTupleMatch(tupleElements []string, arms []tuplePatternArm, originalLine int, outputLine int) (string, []Mapping) {
	var buf bytes.Buffer
	mappings := []Mapping{}

	matchID := r.matchCounter
	r.matchCounter++

	arity := len(tupleElements)

	// Line 1: DINGO_MATCH_START marker
	scrutineeRepr := "(" + strings.Join(tupleElements, ", ") + ")"
	buf.WriteString(fmt.Sprintf("// DINGO_MATCH_START: %s\n", scrutineeRepr))
	mappings = append(mappings, Mapping{
		OriginalLine:    originalLine,
		OriginalColumn:  1,
		GeneratedLine:   outputLine,
		GeneratedColumn: 1,
		Length:          5,
		Name:            "rust_match",
	})
	outputLine++

	// Line 2: Extract tuple elements into temp vars
	var elemVars []string
	for i := 0; i < arity; i++ {
		elemVars = append(elemVars, fmt.Sprintf("__match_%d_elem%d", matchID, i))
	}
	buf.WriteString(fmt.Sprintf("%s := %s\n",
		strings.Join(elemVars, ", "),
		strings.Join(tupleElements, ", "),
	))
	mappings = append(mappings, Mapping{
		OriginalLine:    originalLine,
		OriginalColumn:  1,
		GeneratedLine:   outputLine,
		GeneratedColumn: 1,
		Length:          5,
		Name:            "rust_match",
	})
	outputLine++

	// Generate DINGO_TUPLE_PATTERN marker with pattern summary
	patternSummary := r.generateTuplePatternSummary(arms)
	buf.WriteString(fmt.Sprintf("// DINGO_TUPLE_PATTERN: %s | ARITY: %d\n", patternSummary, arity))
	mappings = append(mappings, Mapping{
		OriginalLine:    originalLine,
		OriginalColumn:  1,
		GeneratedLine:   outputLine,
		GeneratedColumn: 1,
		Length:          5,
		Name:            "rust_match",
	})
	outputLine++

	// Line 3: switch on first element
	buf.WriteString(fmt.Sprintf("switch %s.tag {\n", elemVars[0]))
	mappings = append(mappings, Mapping{
		OriginalLine:    originalLine,
		OriginalColumn:  1,
		GeneratedLine:   outputLine,
		GeneratedColumn: 1,
		Length:          5,
		Name:            "rust_match",
	})
	outputLine++

	// Generate cases (plugin will transform into nested switches)
	// For now, we just generate flat cases with markers
	// Plugin will detect DINGO_TUPLE_PATTERN and rewrite
	for _, arm := range arms {
		caseLines, caseMappings := r.generateTupleCase(elemVars, arm, originalLine, outputLine)
		buf.WriteString(caseLines)
		mappings = append(mappings, caseMappings...)
		outputLine += strings.Count(caseLines, "\n")
	}

	// Closing brace
	buf.WriteString("}\n")
	mappings = append(mappings, Mapping{
		OriginalLine:    originalLine,
		OriginalColumn:  1,
		GeneratedLine:   outputLine,
		GeneratedColumn: 1,
		Length:          1,
		Name:            "rust_match",
	})
	outputLine++

	// DINGO_MATCH_END marker
	buf.WriteString("// DINGO_MATCH_END\n")
	mappings = append(mappings, Mapping{
		OriginalLine:    originalLine,
		OriginalColumn:  1,
		GeneratedLine:   outputLine,
		GeneratedColumn: 1,
		Length:          1,
		Name:            "rust_match",
	})

	return buf.String(), mappings
}

// generateTuplePatternSummary creates a summary string for DINGO_TUPLE_PATTERN marker
// Example: (Ok, Ok) | (Ok, Err) | (Err, _)
func (r *RustMatchProcessor) generateTuplePatternSummary(arms []tuplePatternArm) string {
	var patterns []string
	for _, arm := range arms {
		var variants []string
		for _, elem := range arm.patterns {
			variants = append(variants, elem.variant)
		}
		patterns = append(patterns, "("+strings.Join(variants, ", ")+")")
	}
	return strings.Join(patterns, " | ")
}

// generateTupleCase generates code for one tuple pattern arm
// This is a simplified placeholder - plugin will do the actual nested switch generation
func (r *RustMatchProcessor) generateTupleCase(elemVars []string, arm tuplePatternArm, originalLine int, outputLine int) (string, []Mapping) {
	var buf bytes.Buffer
	mappings := []Mapping{}

	// Generate case for first element only (plugin will expand to nested switches)
	firstElem := arm.patterns[0]

	if firstElem.variant == "_" {
		// Wildcard - default case
		buf.WriteString("default:\n")
	} else {
		// Specific variant
		tagName := r.getTagName(firstElem.variant)
		buf.WriteString(fmt.Sprintf("case %s:\n", tagName))
	}

	mappings = append(mappings, Mapping{
		OriginalLine:    originalLine,
		OriginalColumn:  1,
		GeneratedLine:   outputLine,
		GeneratedColumn: 1,
		Length:          4,
		Name:            "rust_match_arm",
	})
	outputLine++

	// Add DINGO_TUPLE_ARM marker with full pattern info
	var patternStrs []string
	for _, elem := range arm.patterns {
		if elem.binding != "" {
			patternStrs = append(patternStrs, fmt.Sprintf("%s(%s)", elem.variant, elem.binding))
		} else {
			patternStrs = append(patternStrs, elem.variant)
		}
	}
	patternRepr := "(" + strings.Join(patternStrs, ", ") + ")"

	buf.WriteString(fmt.Sprintf("\t// DINGO_TUPLE_ARM: %s", patternRepr))
	if arm.guard != "" {
		buf.WriteString(fmt.Sprintf(" | DINGO_GUARD: %s", arm.guard))
	}
	buf.WriteString("\n")

	mappings = append(mappings, Mapping{
		OriginalLine:    originalLine,
		OriginalColumn:  1,
		GeneratedLine:   outputLine,
		GeneratedColumn: 1,
		Length:          len(patternRepr),
		Name:            "rust_match_arm",
	})
	outputLine++

	// Plugin will generate:
	// 1. Bindings for all elements
	// 2. Nested switches for remaining elements
	// 3. Guard checks
	// For now, just add expression
	buf.WriteString(fmt.Sprintf("\t%s\n", arm.expression))

	mappings = append(mappings, Mapping{
		OriginalLine:    originalLine,
		OriginalColumn:  1,
		GeneratedLine:   outputLine,
		GeneratedColumn: 1,
		Length:          len(arm.expression),
		Name:            "rust_match_expr",
	})

	return buf.String(), mappings
}

package preprocessor

import (
	"strings"
	"testing"
)

// ====================
// 1. Simple Ternaries (10 tests)
// ====================

func TestTernaryProcessor_SimpleVariable(t *testing.T) {
	processor := NewTernaryProcessor()

	source := `let x = cond ? y : z`
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should generate IIFE
	if !strings.Contains(output, "func()") {
		t.Errorf("Expected IIFE, got: %s", output)
	}

	// Should contain if statement
	if !strings.Contains(output, "if cond {") {
		t.Errorf("Expected if statement, got: %s", output)
	}

	// Should return both branches
	if !strings.Contains(output, "return y") {
		t.Errorf("Expected 'return y', got: %s", output)
	}
	if !strings.Contains(output, "return z") {
		t.Errorf("Expected 'return z', got: %s", output)
	}
}

func TestTernaryProcessor_StringLiterals(t *testing.T) {
	processor := NewTernaryProcessor()

	source := `let status = age >= 18 ? "adult" : "minor"`
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should infer string type
	if !strings.Contains(output, "func() string {") {
		t.Errorf("Expected string type, got: %s", output)
	}

	// Should contain condition
	if !strings.Contains(output, "age >= 18") {
		t.Errorf("Expected condition 'age >= 18', got: %s", output)
	}

	// Should return string literals
	if !strings.Contains(output, `return "adult"`) {
		t.Errorf("Expected 'return \"adult\"', got: %s", output)
	}
	if !strings.Contains(output, `return "minor"`) {
		t.Errorf("Expected 'return \"minor\"', got: %s", output)
	}
}

func TestTernaryProcessor_IntegerLiterals(t *testing.T) {
	processor := NewTernaryProcessor()

	source := `let score = isPrime ? 100 : 0`
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should infer int type
	if !strings.Contains(output, "func() int {") {
		t.Errorf("Expected int type, got: %s", output)
	}

	// Should return integer literals
	if !strings.Contains(output, "return 100") {
		t.Errorf("Expected 'return 100', got: %s", output)
	}
	if !strings.Contains(output, "return 0") {
		t.Errorf("Expected 'return 0', got: %s", output)
	}
}

func TestTernaryProcessor_BooleanLiterals(t *testing.T) {
	processor := NewTernaryProcessor()

	source := `let valid = x > 0 ? true : false`
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should infer bool type
	if !strings.Contains(output, "func() bool {") {
		t.Errorf("Expected bool type, got: %s", output)
	}

	// Should return boolean literals
	if !strings.Contains(output, "return true") {
		t.Errorf("Expected 'return true', got: %s", output)
	}
	if !strings.Contains(output, "return false") {
		t.Errorf("Expected 'return false', got: %s", output)
	}
}

func TestTernaryProcessor_FunctionCalls(t *testing.T) {
	processor := NewTernaryProcessor()

	source := `let result = isValid() ? getSuccess() : getFailure()`
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should contain function calls
	if !strings.Contains(output, "isValid()") {
		t.Errorf("Expected 'isValid()' call, got: %s", output)
	}
	if !strings.Contains(output, "return getSuccess()") {
		t.Errorf("Expected 'return getSuccess()' call, got: %s", output)
	}
	if !strings.Contains(output, "return getFailure()") {
		t.Errorf("Expected 'return getFailure()' call, got: %s", output)
	}
}

func TestTernaryProcessor_ComparisonCondition(t *testing.T) {
	processor := NewTernaryProcessor()

	source := `let msg = count > 10 ? "many" : "few"`
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should preserve comparison
	if !strings.Contains(output, "if count > 10 {") {
		t.Errorf("Expected 'if count > 10', got: %s", output)
	}
}

func TestTernaryProcessor_ComplexExpression(t *testing.T) {
	processor := NewTernaryProcessor()

	source := `let val = x + y > z ? a * b : c / d`
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should preserve complex condition
	if !strings.Contains(output, "x + y > z") {
		t.Errorf("Expected complex condition, got: %s", output)
	}

	// Should preserve complex branches
	if !strings.Contains(output, "a * b") {
		t.Errorf("Expected 'a * b' expression, got: %s", output)
	}
	if !strings.Contains(output, "c / d") {
		t.Errorf("Expected 'c / d' expression, got: %s", output)
	}
}

func TestTernaryProcessor_FieldAccess(t *testing.T) {
	processor := NewTernaryProcessor()

	source := `let name = user.IsAdmin ? user.FullName : user.Username`
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should preserve field access
	if !strings.Contains(output, "user.IsAdmin") {
		t.Errorf("Expected 'user.IsAdmin', got: %s", output)
	}
	if !strings.Contains(output, "user.FullName") {
		t.Errorf("Expected 'user.FullName', got: %s", output)
	}
	if !strings.Contains(output, "user.Username") {
		t.Errorf("Expected 'user.Username', got: %s", output)
	}
}

func TestTernaryProcessor_SingleLiteral(t *testing.T) {
	processor := NewTernaryProcessor()

	source := `let x = true ? 1 : 0`
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should generate IIFE even for simple literal condition
	if !strings.Contains(output, "func()") {
		t.Errorf("Expected IIFE, got: %s", output)
	}
	if !strings.Contains(output, "if true {") {
		t.Errorf("Expected 'if true', got: %s", output)
	}
}

func TestTernaryProcessor_ParenthesizedCondition(t *testing.T) {
	processor := NewTernaryProcessor()

	source := `let x = (a && b) ? 1 : 2`
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should preserve parenthesized condition
	if !strings.Contains(output, "(a && b)") {
		t.Errorf("Expected parenthesized condition, got: %s", output)
	}
}

// ====================
// 2. Nested Ternaries (8 tests)
// ====================

func TestTernaryProcessor_NestedTwoLevels(t *testing.T) {
	processor := NewTernaryProcessor()

	source := `let x = a ? (b ? c : d) : e`
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should generate nested IIFEs
	count := strings.Count(output, "func()")
	if count < 2 {
		t.Errorf("Expected at least 2 IIFEs (nested), found %d in: %s", count, output)
	}
}

func TestTernaryProcessor_NestedThreeLevels(t *testing.T) {
	processor := NewTernaryProcessor()

	source := `let x = a ? (b ? (c ? d : e) : f) : g`
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should generate 3 nested IIFEs
	count := strings.Count(output, "func()")
	if count < 3 {
		t.Errorf("Expected at least 3 IIFEs (nested), found %d in: %s", count, output)
	}
}

func TestTernaryProcessor_NestedFourLevels_Error(t *testing.T) {
	processor := NewTernaryProcessor()

	source := `let x = a ? (b ? (c ? (d ? e : f) : g) : h) : i`
	_, _, err := processor.Process([]byte(source))

	// Should fail with nesting depth error
	if err == nil {
		t.Fatalf("Expected error for 4-level nesting, got none")
	}

	if !strings.Contains(err.Error(), "nesting too deep") {
		t.Errorf("Expected 'nesting too deep' error, got: %v", err)
	}

	if !strings.Contains(err.Error(), "level 4") {
		t.Errorf("Expected 'level 4' in error, got: %v", err)
	}
}

func TestTernaryProcessor_NestedInTrueBranch(t *testing.T) {
	processor := NewTernaryProcessor()

	source := `let x = a ? (b ? c : d) : e`
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should have nested structure in true branch
	if !strings.Contains(output, "func()") {
		t.Errorf("Expected IIFE, got: %s", output)
	}
}

func TestTernaryProcessor_NestedInFalseBranch(t *testing.T) {
	processor := NewTernaryProcessor()

	source := `let x = a ? b : (c ? d : e)`
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should have nested structure in false branch
	if !strings.Contains(output, "func()") {
		t.Errorf("Expected IIFE, got: %s", output)
	}
}

func TestTernaryProcessor_NestedBothBranches(t *testing.T) {
	processor := NewTernaryProcessor()

	source := `let x = a ? (b ? c : d) : (e ? f : g)`
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should have multiple nested IIFEs
	count := strings.Count(output, "func()")
	if count < 3 {
		t.Errorf("Expected at least 3 IIFEs, found %d in: %s", count, output)
	}
}

func TestTernaryProcessor_NestedMaxDepthBoundary(t *testing.T) {
	processor := NewTernaryProcessor()

	// Exactly 3 levels - should succeed
	source := `let x = a ? (b ? (c ? d : e) : f) : g`
	_, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Expected success for 3-level nesting, got error: %v", err)
	}
}

func TestTernaryProcessor_NestedComplexExpressions(t *testing.T) {
	processor := NewTernaryProcessor()

	source := `let x = score > 90 ? (bonus > 10 ? "excellent" : "good") : "average"`
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should preserve all expressions
	if !strings.Contains(output, "score > 90") {
		t.Errorf("Expected 'score > 90', got: %s", output)
	}
	if !strings.Contains(output, "bonus > 10") {
		t.Errorf("Expected 'bonus > 10', got: %s", output)
	}
}

// ====================
// 3. Chained Ternaries (5 tests)
// ====================

func TestTernaryProcessor_ChainedSimple(t *testing.T) {
	processor := NewTernaryProcessor()

	source := `let x = a ? b : c ? d : e`
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should generate nested IIFEs (right-associative)
	if !strings.Contains(output, "func()") {
		t.Errorf("Expected IIFE, got: %s", output)
	}
}

func TestTernaryProcessor_ChainedMultiple(t *testing.T) {
	processor := NewTernaryProcessor()

	source := `let x = a ? b : c ? d : e ? f : g`
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should generate multiple IIFEs
	if !strings.Contains(output, "func()") {
		t.Errorf("Expected IIFE, got: %s", output)
	}
}

func TestTernaryProcessor_ChainedWithComparison(t *testing.T) {
	processor := NewTernaryProcessor()

	source := `let grade = score >= 90 ? "A" : score >= 80 ? "B" : score >= 70 ? "C" : "F"`
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should preserve all comparisons
	if !strings.Contains(output, "score >= 90") {
		t.Errorf("Expected 'score >= 90', got: %s", output)
	}
	if !strings.Contains(output, "score >= 80") {
		t.Errorf("Expected 'score >= 80', got: %s", output)
	}
	if !strings.Contains(output, "score >= 70") {
		t.Errorf("Expected 'score >= 70', got: %s", output)
	}
}

func TestTernaryProcessor_ChainedAtMaxDepth(t *testing.T) {
	processor := NewTernaryProcessor()

	// 3 levels via chaining - should succeed
	source := `let x = a ? b : c ? d : e ? f : g`
	_, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Expected success for 3-level chained ternary, got error: %v", err)
	}
}

func TestTernaryProcessor_ChainedExceedsMaxDepth(t *testing.T) {
	processor := NewTernaryProcessor()

	// 4 levels via chaining - should fail
	source := `let x = a ? b : c ? d : e ? f : g ? h : i`
	_, _, err := processor.Process([]byte(source))

	// Should fail with nesting depth error
	if err == nil {
		t.Fatalf("Expected error for 4-level chained ternary, got none")
	}

	if !strings.Contains(err.Error(), "nesting too deep") {
		t.Errorf("Expected 'nesting too deep' error, got: %v", err)
	}
}

// ====================
// 4. Edge Cases (10 tests)
// ====================

func TestTernaryProcessor_StringWithColon(t *testing.T) {
	processor := NewTernaryProcessor()

	source := `let url = isSecure ? "https://example.com" : "http://example.com"`
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should preserve URLs with colons
	if !strings.Contains(output, `"https://example.com"`) {
		t.Errorf("Expected HTTPS URL preserved, got: %s", output)
	}
	if !strings.Contains(output, `"http://example.com"`) {
		t.Errorf("Expected HTTP URL preserved, got: %s", output)
	}
}

func TestTernaryProcessor_StringWithQuestionMark(t *testing.T) {
	processor := NewTernaryProcessor()

	source := `let msg = hasError ? "Error: What happened?" : "Success!"`
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should preserve question marks in strings
	if !strings.Contains(output, `"Error: What happened?"`) {
		t.Errorf("Expected question mark in string preserved, got: %s", output)
	}
}

func TestTernaryProcessor_NoTernary(t *testing.T) {
	processor := NewTernaryProcessor()

	source := `let x = value`
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should be unchanged
	if output != source {
		t.Errorf("Expected unchanged output, got: %s", output)
	}
}

func TestTernaryProcessor_ErrorPropagationOperator(t *testing.T) {
	processor := NewTernaryProcessor()

	// Single ? without : should NOT be treated as ternary
	source := `let x = getValue()?`
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should be unchanged (error prop handled by different processor)
	if output != source {
		t.Errorf("Expected unchanged output for error prop, got: %s", output)
	}
}

func TestTernaryProcessor_NullCoalesceOperator(t *testing.T) {
	processor := NewTernaryProcessor()

	// ?? should NOT be treated as ternary
	source := `let x = value ?? defaultValue`
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should be unchanged (null coalesce handled by different processor)
	if output != source {
		t.Errorf("Expected unchanged output for null coalesce, got: %s", output)
	}
}

func TestTernaryProcessor_MixedWithNullCoalesce(t *testing.T) {
	processor := NewTernaryProcessor()

	source := `let x = (value ?? defaultValue) ? "yes" : "no"`
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should preserve ?? in condition
	if !strings.Contains(output, "??") {
		t.Errorf("Expected ?? preserved in condition, got: %s", output)
	}

	// Should generate ternary IIFE
	if !strings.Contains(output, "func()") {
		t.Errorf("Expected IIFE for ternary, got: %s", output)
	}
}

func TestTernaryProcessor_EmptyLines(t *testing.T) {
	processor := NewTernaryProcessor()

	source := "\n\nlet x = a ? b : c\n\n"
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should preserve empty lines
	if !strings.HasPrefix(output, "\n\n") {
		t.Errorf("Expected leading newlines preserved, got: %s", output)
	}
	if !strings.HasSuffix(output, "\n\n") {
		t.Errorf("Expected trailing newlines preserved, got: %s", output)
	}
}

func TestTernaryProcessor_MultipleLines(t *testing.T) {
	processor := NewTernaryProcessor()

	source := `let x = a ? b : c
let y = d ? e : f`
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should process both ternaries
	count := strings.Count(output, "func()")
	if count < 2 {
		t.Errorf("Expected 2 IIFEs (one per line), found %d in: %s", count, output)
	}
}

func TestTernaryProcessor_RawStringLiterals(t *testing.T) {
	processor := NewTernaryProcessor()

	source := "let msg = isValid ? `Success!` : `Failed!`"
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should preserve raw string literals
	if !strings.Contains(output, "`Success!`") {
		t.Errorf("Expected raw string preserved, got: %s", output)
	}
	if !strings.Contains(output, "`Failed!`") {
		t.Errorf("Expected raw string preserved, got: %s", output)
	}
}

func TestTernaryProcessor_WhitespaceVariations(t *testing.T) {
	processor := NewTernaryProcessor()

	source := `let x = a?b:c` // No whitespace
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should still generate IIFE
	if !strings.Contains(output, "func()") {
		t.Errorf("Expected IIFE even without whitespace, got: %s", output)
	}
}

// ====================
// 5. Type Inference (5 tests)
// ====================

func TestTernaryProcessor_TypeInference_SameStringType(t *testing.T) {
	processor := NewTernaryProcessor()

	source := `let msg = cond ? "yes" : "no"`
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should infer concrete string type
	if !strings.Contains(output, "func() string {") {
		t.Errorf("Expected string type, got: %s", output)
	}
}

func TestTernaryProcessor_TypeInference_SameIntType(t *testing.T) {
	processor := NewTernaryProcessor()

	source := `let num = cond ? 42 : 100`
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should infer concrete int type
	if !strings.Contains(output, "func() int {") {
		t.Errorf("Expected int type, got: %s", output)
	}
}

func TestTernaryProcessor_TypeInference_DifferentTypes(t *testing.T) {
	processor := NewTernaryProcessor()

	source := `let value = cond ? "text" : 42`
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should fallback to 'any' for mixed types
	if !strings.Contains(output, "func() any {") {
		t.Errorf("Expected any type for mixed types, got: %s", output)
	}
}

func TestTernaryProcessor_TypeInference_ComplexTypes(t *testing.T) {
	processor := NewTernaryProcessor()

	source := `let arr = cond ? []int{1, 2} : []int{3, 4}`
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should infer []int type
	if !strings.Contains(output, "func() []int {") {
		t.Errorf("Expected []int type, got: %s", output)
	}
}

func TestTernaryProcessor_TypeInference_BoolType(t *testing.T) {
	processor := NewTernaryProcessor()

	source := `let flag = cond ? true : false`
	result, _, err := processor.Process([]byte(source))
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Should infer bool type
	if !strings.Contains(output, "func() bool {") {
		t.Errorf("Expected bool type, got: %s", output)
	}
}

// ====================
// 6. CRITICAL Fixes (Code Review Issues)
// ====================

// Issue #1: Expression boundary detection bug
func TestTernaryProcessor_ComparisonInCondition(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"with spaces", `let x = (a == b) ? c : d`},
		{"no spaces", `let x = a==b?c:d`},
		{"complex", `let x = (age >= 18 && score <= 100) ? "pass" : "fail"`},
		{"not equals", `let x = a != b ? 1 : 0`},
		{"walrus", `let x := getValue(); x > 0 ? "pos" : "neg"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proc := NewTernaryProcessor()
			result, _, err := proc.Process([]byte(tt.input))
			if err != nil {
				t.Fatalf("Process failed: %v", err)
			}

			output := string(result)

			// Verify IIFE generated correctly
			if !strings.Contains(output, "func()") {
				t.Errorf("Expected IIFE, got: %s", output)
			}

			// Verify comparison operators preserved
			if strings.Contains(tt.input, "==") && !strings.Contains(output, "==") {
				t.Errorf("Expected == preserved, got: %s", output)
			}
			if strings.Contains(tt.input, ">=") && !strings.Contains(output, ">=") {
				t.Errorf("Expected >= preserved, got: %s", output)
			}
			if strings.Contains(tt.input, "<=") && !strings.Contains(output, "<=") {
				t.Errorf("Expected <= preserved, got: %s", output)
			}
			if strings.Contains(tt.input, "!=") && !strings.Contains(output, "!=") {
				t.Errorf("Expected != preserved, got: %s", output)
			}
		})
	}
}

// Issue #2: Raw string literal handling
func TestTernaryProcessor_RawStringLiterals_InBranches(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"raw in false branch", "let url = local ? \"localhost\" : `example.com:8080`"},
		{"raw in true branch", "let url = prod ? `https://api.com:443` : \"localhost\""},
		{"raw in both", "let url = flag ? `server1:8080` : `server2:9090`"},
		{"raw with colon", "let msg = err ? `Error: failed` : `Success: ok`"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proc := NewTernaryProcessor()
			result, _, err := proc.Process([]byte(tt.input))
			if err != nil {
				t.Fatalf("Process failed: %v", err)
			}

			output := string(result)

			// Verify IIFE generated
			if !strings.Contains(output, "func()") {
				t.Errorf("Expected IIFE, got: %s", output)
			}

			// Verify raw strings preserved
			if strings.Contains(tt.input, "`") && !strings.Contains(output, "`") {
				t.Errorf("Expected raw string preserved, got: %s", output)
			}

			// Count backticks - should match input
			inputBackticks := strings.Count(tt.input, "`")
			outputBackticks := strings.Count(output, "`")
			if inputBackticks != outputBackticks {
				t.Errorf("Expected %d backticks, got %d. Output: %s", inputBackticks, outputBackticks, output)
			}
		})
	}
}

// Issue #3: Multi-ternary per line error
func TestTernaryProcessor_MultiTernaryPerLine_Error(t *testing.T) {
	input := `let x = a ? 1 : 2, y = b ? 3 : 4`
	proc := NewTernaryProcessor()
	_, _, err := proc.Process([]byte(input))

	if err == nil {
		t.Fatalf("Expected error for multiple ternaries, got none")
	}

	if !strings.Contains(err.Error(), "multiple ternary operators") {
		t.Errorf("Expected 'multiple ternary operators' in error, got: %v", err)
	}

	if !strings.Contains(err.Error(), "found 2") {
		t.Errorf("Expected 'found 2' in error, got: %v", err)
	}
}

// Issue #4: Source mapping implementation
func TestTernaryProcessor_SourceMappings(t *testing.T) {
	input := `let x = age > 18 ? "adult" : "minor"`
	proc := NewTernaryProcessor()
	result, mappings, err := proc.Process([]byte(input))

	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	output := string(result)

	// Verify mappings were generated
	if len(mappings) == 0 {
		t.Errorf("Expected source mappings, got none")
	}

	// Should have 3 mappings: condition, trueVal, falseVal
	if len(mappings) != 3 {
		t.Errorf("Expected 3 mappings, got %d", len(mappings))
	}

	// Verify IIFE generated
	if !strings.Contains(output, "func() string {") {
		t.Errorf("Expected IIFE with string type, got: %s", output)
	}

	// Verify condition in generated code
	if !strings.Contains(output, "age > 18") {
		t.Errorf("Expected condition 'age > 18', got: %s", output)
	}
}

// ====================
// 7. Performance Benchmarks (Issue #6)
// ====================

func BenchmarkTernaryProcessor_LargeFile(b *testing.B) {
	// Generate 1000-line file with 100 ternaries
	lines := make([]string, 1000)
	for i := 0; i < 100; i++ {
		lines[i*10] = "let x" + strings.Repeat("a", i%10) + " = cond ? " + strings.Repeat("1", i%5+1) + " : " + strings.Repeat("2", i%5+1)
	}
	source := []byte(strings.Join(lines, "\n"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		proc := NewTernaryProcessor()
		_, _, err := proc.Process(source)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTernaryProcessor_NestedTernaries(b *testing.B) {
	source := []byte(`let x = a ? (b ? 1 : 2) : (c ? 3 : 4)`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		proc := NewTernaryProcessor()
		_, _, err := proc.Process(source)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTernaryProcessor_SimpleTernary(b *testing.B) {
	source := []byte(`let x = a ? b : c`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		proc := NewTernaryProcessor()
		_, _, err := proc.Process(source)
		if err != nil {
			b.Fatal(err)
		}
	}
}

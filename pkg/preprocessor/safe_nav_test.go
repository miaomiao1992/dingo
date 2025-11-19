package preprocessor

import (
	"strings"
	"testing"
)

func TestSafeNavProcessor_PropertyAccess_Option(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string // Strings that must appear in output
		notContains []string // Strings that must NOT appear
	}{
		{
			name: "simple property access",
			input: `let user: UserOption = getUser()
let name = user?.name`,
			contains: []string{
				"func() __INFER__",
				"if user.IsNone()",
				"return __INFER___None()",
				"user.Unwrap()",
				"__user0 :=",
				"return __user0.name",
			},
		},
		{
			name: "chained property access",
			input: `let user: UserOption = getUser()
let city = user?.address?.city`,
			contains: []string{
				"func() __INFER__",
				"if user.IsNone()",
				"__user0 := user.Unwrap()",
				"if __user0.address.IsNone()",
				"__user1 := __user0.address.Unwrap()",
				"return __user1.city",
			},
		},
		{
			name: "three-level chain",
			input: `let user: UserOption = getUser()
let value = user?.profile?.settings?.theme`,
			contains: []string{
				"func() __INFER__",
				"if user.IsNone()",
				"user.Unwrap()",
				"__user0.profile.IsNone()",
				"__user1.settings.IsNone()",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewSafeNavProcessor()
			output, _, err := processor.Process([]byte(tt.input))
			if err != nil {
				t.Fatalf("Process() error = %v", err)
			}

			result := string(output)

			// Check required strings
			for _, str := range tt.contains {
				if !strings.Contains(result, str) {
					t.Errorf("Output missing expected string: %q\nGot:\n%s", str, result)
				}
			}

			// Check forbidden strings
			for _, str := range tt.notContains {
				if strings.Contains(result, str) {
					t.Errorf("Output contains forbidden string: %q\nGot:\n%s", str, result)
				}
			}
		})
	}
}

func TestSafeNavProcessor_PropertyAccess_Pointer(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name: "pointer simple access",
			input: `let user: *User = getUser()
let name = user?.name`,
			contains: []string{
				"func() __INFER__",
				"if user == nil",
				"return nil",
				"return user.name",
			},
		},
		{
			name: "pointer chained access",
			input: `let user: *User = getUser()
let city = user?.address?.city`,
			contains: []string{
				"func() __INFER__",
				"if user == nil",
				"return nil",
				"if user.address == nil",
				"return user.address.city",
			},
		},
		{
			name: "pointer triple chain",
			input: `let config: *Config = getConfig()
let value = config?.db?.connection?.timeout`,
			contains: []string{
				"if config == nil",
				"if config.db == nil",
				"if config.db.connection == nil",
				"return config.db.connection.timeout",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewSafeNavProcessor()
			output, _, err := processor.Process([]byte(tt.input))
			if err != nil {
				t.Fatalf("Process() error = %v", err)
			}

			result := string(output)

			for _, str := range tt.contains {
				if !strings.Contains(result, str) {
					t.Errorf("Output missing expected string: %q\nGot:\n%s", str, result)
				}
			}
		})
	}
}

func TestSafeNavProcessor_UnknownType_Placeholder(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name: "unknown type generates placeholder",
			input: `let user = getUser()
let name = user?.name`,
			contains: []string{
				`__SAFE_NAV_INFER__(user, "name")`,
			},
		},
		{
			name: "unknown type chain generates placeholder",
			input: `let config = getConfig()
let value = config?.settings?.timeout`,
			contains: []string{
				`__SAFE_NAV_INFER__(config, "settings", "timeout")`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewSafeNavProcessor()
			output, _, err := processor.Process([]byte(tt.input))
			if err != nil {
				t.Fatalf("Process() error = %v", err)
			}

			result := string(output)

			for _, str := range tt.contains {
				if !strings.Contains(result, str) {
					t.Errorf("Output missing expected string: %q\nGot:\n%s", str, result)
				}
			}
		})
	}
}

func TestSafeNavProcessor_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "trailing safe nav operator",
			input:       "let x = user?.",
			expectError: true,
			errorMsg:    "trailing ?. operator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewSafeNavProcessor()
			_, _, err := processor.Process([]byte(tt.input))

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing %q, got nil", tt.errorMsg)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestSafeNavProcessor_TypeDetection(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		varName  string
		expected TypeKind
	}{
		{
			name:     "detect pointer type",
			source:   "let user: *User = nil",
			varName:  "user",
			expected: TypePointer,
		},
		{
			name:     "detect option type by suffix",
			source:   "let user: UserOption = None()",
			varName:  "user",
			expected: TypeOption,
		},
		{
			name:     "detect option type by enum",
			source:   "enum UserOption { Some(User), None }\nlet user: UserOption = None()",
			varName:  "user",
			expected: TypeOption,
		},
		{
			name:     "detect unknown type",
			source:   "let user = getUser()",
			varName:  "user",
			expected: TypeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewTypeDetector()
			detector.ParseSource([]byte(tt.source))

			result := detector.DetectType(tt.varName)
			if result != tt.expected {
				t.Errorf("DetectType(%q) = %v, want %v", tt.varName, result, tt.expected)
			}
		})
	}
}

func TestSafeNavProcessor_SourceMappings(t *testing.T) {
	input := `let user: UserOption = getUser()
let name = user?.name`

	processor := NewSafeNavProcessor()
	_, mappings, err := processor.Process([]byte(input))
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	// Should have at least one mapping for the safe nav operation
	if len(mappings) == 0 {
		t.Error("Expected source mappings, got none")
	}

	// Verify mapping has required fields
	for _, m := range mappings {
		if m.OriginalLine == 0 {
			t.Error("Mapping missing OriginalLine")
		}
		if m.GeneratedLine == 0 {
			t.Error("Mapping missing GeneratedLine")
		}
		if m.Name == "" {
			t.Error("Mapping missing Name")
		}
	}
}

func TestSafeNavProcessor_NoChange(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "no safe nav operator",
			input: "let user = getUser()\nlet name = user.name",
		},
		{
			name:  "ternary operator",
			input: "let x = condition ? a : b",
		},
		{
			name:  "question mark in string",
			input: `let msg = "Is this safe?"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewSafeNavProcessor()
			output, _, err := processor.Process([]byte(tt.input))
			if err != nil {
				t.Fatalf("Process() error = %v", err)
			}

			if string(output) != tt.input {
				t.Errorf("Expected unchanged output, got:\n%s", string(output))
			}
		})
	}
}

func TestParseSafeNavChain(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []ChainElement
	}{
		{
			name:  "single property",
			input: "?.name",
			expected: []ChainElement{
				{Name: "name", IsMethod: false},
			},
		},
		{
			name:  "two properties",
			input: "?.address?.city",
			expected: []ChainElement{
				{Name: "address", IsMethod: false},
				{Name: "city", IsMethod: false},
			},
		},
		{
			name:  "simple method call",
			input: "?.getName()",
			expected: []ChainElement{
				{Name: "getName", IsMethod: true, Args: []string{}, RawArgs: ""},
			},
		},
		{
			name:  "method with single arg",
			input: "?.process(arg1)",
			expected: []ChainElement{
				{Name: "process", IsMethod: true, Args: []string{"arg1"}, RawArgs: "arg1"},
			},
		},
		{
			name:  "method with multiple args",
			input: "?.transform(arg1, arg2, arg3)",
			expected: []ChainElement{
				{Name: "transform", IsMethod: true, Args: []string{"arg1", "arg2", "arg3"}, RawArgs: "arg1, arg2, arg3"},
			},
		},
		{
			name:  "chained methods",
			input: "?.getName()?.toUpper()",
			expected: []ChainElement{
				{Name: "getName", IsMethod: true, Args: []string{}, RawArgs: ""},
				{Name: "toUpper", IsMethod: true, Args: []string{}, RawArgs: ""},
			},
		},
		{
			name:  "mixed property and method",
			input: "?.address?.format()",
			expected: []ChainElement{
				{Name: "address", IsMethod: false},
				{Name: "format", IsMethod: true, Args: []string{}, RawArgs: ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseSafeNavChain(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d elements, got %d: %v", len(tt.expected), len(result), result)
				return
			}

			for i, elem := range result {
				if elem.Name != tt.expected[i].Name {
					t.Errorf("Element %d: expected name %q, got %q", i, tt.expected[i].Name, elem.Name)
				}
				if elem.IsMethod != tt.expected[i].IsMethod {
					t.Errorf("Element %d: expected IsMethod %v, got %v", i, tt.expected[i].IsMethod, elem.IsMethod)
				}
				if elem.IsMethod && elem.RawArgs != tt.expected[i].RawArgs {
					t.Errorf("Element %d: expected RawArgs %q, got %q", i, tt.expected[i].RawArgs, elem.RawArgs)
				}
			}
		})
	}
}

func TestClassifyType(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		expected TypeKind
	}{
		{
			name:     "pointer type",
			typeName: "*User",
			expected: TypePointer,
		},
		{
			name:     "option by suffix",
			typeName: "UserOption",
			expected: TypeOption,
		},
		{
			name:     "option generic syntax",
			typeName: "Option<User>",
			expected: TypeOption,
		},
		{
			name:     "regular type",
			typeName: "User",
			expected: TypeRegular,
		},
		{
			name:     "regular type int",
			typeName: "int",
			expected: TypeRegular,
		},
		{
			name:     "pointer to primitive",
			typeName: "*string",
			expected: TypePointer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifyType(tt.typeName)
			if result != tt.expected {
				t.Errorf("classifyType(%q) = %v, want %v", tt.typeName, result, tt.expected)
			}
		})
	}
}

func TestSafeNavProcessor_MultipleChains(t *testing.T) {
	input := `let user: UserOption = getUser()
let config: *Config = getConfig()
let name = user?.name
let timeout = config?.timeout`

	processor := NewSafeNavProcessor()
	output, _, err := processor.Process([]byte(input))
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	result := string(output)

	// Should contain both Option and pointer checks
	expectedStrings := []string{
		"if user.IsNone()", // Option check
		"if config == nil", // Pointer check
	}

	for _, str := range expectedStrings {
		if !strings.Contains(result, str) {
			t.Errorf("Output missing expected string: %q\nGot:\n%s", str, result)
		}
	}
}

func TestParseMethodArgs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "no arguments",
			input:    "",
			expected: []string{},
		},
		{
			name:     "single argument",
			input:    "arg1",
			expected: []string{"arg1"},
		},
		{
			name:     "multiple arguments",
			input:    "arg1, arg2, arg3",
			expected: []string{"arg1", "arg2", "arg3"},
		},
		{
			name:     "nested parentheses",
			input:    "getData(), arg2",
			expected: []string{"getData()", "arg2"},
		},
		{
			name:     "nested function call with args",
			input:    "transform(a, b), c",
			expected: []string{"transform(a, b)", "c"},
		},
		{
			name:     "string literal with comma",
			input:    `"hello, world", arg2`,
			expected: []string{`"hello, world"`, "arg2"},
		},
		{
			name:     "multiple string literals",
			input:    `"first, value", "second, value"`,
			expected: []string{`"first, value"`, `"second, value"`},
		},
		{
			name:     "complex nested",
			input:    "process(getData(x, y), format(z)), final",
			expected: []string{"process(getData(x, y), format(z))", "final"},
		},
		{
			name:     "single quotes",
			input:    "'hello, world', arg2",
			expected: []string{"'hello, world'", "arg2"},
		},
		{
			name:     "backticks",
			input:    "`template, string`, arg2",
			expected: []string{"`template, string`", "arg2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseMethodArgs(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d args, got %d: %v", len(tt.expected), len(result), result)
				return
			}

			for i, arg := range result {
				if arg != tt.expected[i] {
					t.Errorf("Arg %d: expected %q, got %q", i, tt.expected[i], arg)
				}
			}
		})
	}
}

func TestSafeNavProcessor_MethodCalls_Option(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name: "simple method call no args",
			input: `let user: UserOption = getUser()
let name = user?.getName()`,
			contains: []string{
				"func() __INFER__",
				"if user.IsNone()",
				"return __INFER___None()",
				"__user0 := user.Unwrap()",
				"return __user0.getName()",
			},
		},
		{
			name: "method call with single arg",
			input: `let user: UserOption = getUser()
let result = user?.process(arg1)`,
			contains: []string{
				"func() __INFER__",
				"if user.IsNone()",
				"__user0 := user.Unwrap()",
				"return __user0.process(arg1)",
			},
		},
		{
			name: "method call with multiple args",
			input: `let user: UserOption = getUser()
let result = user?.transform(arg1, arg2, arg3)`,
			contains: []string{
				"func() __INFER__",
				"return __user0.transform(arg1, arg2, arg3)",
			},
		},
		{
			name: "chained method calls",
			input: `let user: UserOption = getUser()
let result = user?.getName()?.toUpper()`,
			contains: []string{
				"func() __INFER__",
				"if user.IsNone()",
				"__user0 := user.Unwrap()",
				"__user0.getName()",
				"if __user0.getName().IsNone()",
				"__user1 := __user0.getName().Unwrap()",
				"return __user1.toUpper()",
			},
		},
		{
			name: "mixed property and method",
			input: `let user: UserOption = getUser()
let result = user?.address?.format()`,
			contains: []string{
				"func() __INFER__",
				"if user.IsNone()",
				"__user0 := user.Unwrap()",
				"if __user0.address.IsNone()",
				"__user1 := __user0.address.Unwrap()",
				"return __user1.format()",
			},
		},
		{
			name: "method then property",
			input: `let user: UserOption = getUser()
let result = user?.getProfile()?.name`,
			contains: []string{
				"func() __INFER__",
				"__user0.getProfile()",
				"if __user0.getProfile().IsNone()",
				"__user1 := __user0.getProfile().Unwrap()",
				"return __user1.name",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewSafeNavProcessor()
			output, _, err := processor.Process([]byte(tt.input))
			if err != nil {
				t.Fatalf("Process() error = %v", err)
			}

			result := string(output)

			for _, str := range tt.contains {
				if !strings.Contains(result, str) {
					t.Errorf("Output missing expected string: %q\nGot:\n%s", str, result)
				}
			}
		})
	}
}

func TestSafeNavProcessor_MethodCalls_Pointer(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name: "simple method call",
			input: `let user: *User = getUser()
let name = user?.getName()`,
			contains: []string{
				"func() __INFER__",
				"if user == nil",
				"return nil",
				"return user.getName()",
			},
		},
		{
			name: "method with args",
			input: `let user: *User = getUser()
let result = user?.process(arg1, arg2)`,
			contains: []string{
				"func() __INFER__",
				"if user == nil",
				"return user.process(arg1, arg2)",
			},
		},
		{
			name: "chained methods",
			input: `let user: *User = getUser()
let result = user?.getAddress()?.format()`,
			contains: []string{
				"func() __INFER__",
				"if user == nil",
				"if user.getAddress() == nil",
				"return user.getAddress().format()",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewSafeNavProcessor()
			output, _, err := processor.Process([]byte(tt.input))
			if err != nil {
				t.Fatalf("Process() error = %v", err)
			}

			result := string(output)

			for _, str := range tt.contains {
				if !strings.Contains(result, str) {
					t.Errorf("Output missing expected string: %q\nGot:\n%s", str, result)
				}
			}
		})
	}
}

func TestSafeNavProcessor_MethodCalls_ComplexArgs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name: "nested function call in args",
			input: `let user: UserOption = getUser()
let result = user?.process(getData())`,
			contains: []string{
				"return __user0.process(getData())",
			},
		},
		{
			name: "string literal with comma",
			input: `let user: UserOption = getUser()
let result = user?.format("hello, world")`,
			contains: []string{
				`return __user0.format("hello, world")`,
			},
		},
		{
			name: "multiple complex args",
			input: `let user: UserOption = getUser()
let result = user?.transform(getData(x, y), "test, string", 42)`,
			contains: []string{
				`return __user0.transform(getData(x, y), "test, string", 42)`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewSafeNavProcessor()
			output, _, err := processor.Process([]byte(tt.input))
			if err != nil {
				t.Fatalf("Process() error = %v", err)
			}

			result := string(output)

			for _, str := range tt.contains {
				if !strings.Contains(result, str) {
					t.Errorf("Output missing expected string: %q\nGot:\n%s", str, result)
				}
			}
		})
	}
}

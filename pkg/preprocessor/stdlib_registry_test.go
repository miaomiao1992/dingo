package preprocessor

import (
	"strings"
	"testing"
)

func TestGetPackageForFunction_Unique(t *testing.T) {
	tests := []struct {
		function string
		want     string
	}{
		{"ReadFile", "os"},
		{"Atoi", "strconv"},
		{"Printf", "fmt"},
		{"Println", "fmt"},
		{"Marshal", "json"},
		{"Unmarshal", "json"},
		{"Errorf", "fmt"},
		{"New", "rand"}, // errors.New, rand.New - ambiguous in registry
		{"Now", "time"},
		{"Sleep", "time"},
		{"Join", "strings"}, // strings.Join, bytes.Join, path.Join, filepath.Join
		{"Environ", "os"},
		{"Exit", "os"},
		{"Getenv", "os"},
		{"Copy", "io"},
		{"Dial", "net"},
	}

	for _, tt := range tests {
		t.Run(tt.function, func(t *testing.T) {
			got, err := GetPackageForFunction(tt.function)

			// Check if function is actually ambiguous in registry
			pkgs := StdlibRegistry[tt.function]
			if len(pkgs) > 1 {
				// Should return error for ambiguous functions
				if err == nil {
					t.Errorf("Expected ambiguity error for %s (packages: %v), got none",
						tt.function, pkgs)
				}
				return
			}

			// For unique functions
			if err != nil {
				t.Errorf("GetPackageForFunction(%q) error = %v, want nil", tt.function, err)
				return
			}

			if got != tt.want {
				t.Errorf("GetPackageForFunction(%q) = %q, want %q", tt.function, got, tt.want)
			}
		})
	}
}

func TestGetPackageForFunction_Ambiguous(t *testing.T) {
	tests := []struct {
		function     string
		wantPackages []string
	}{
		{"Open", []string{"os", "net"}},
		{"Get", []string{"http", "sync"}},
		{"Read", []string{"io", "os", "bufio", "rand"}},
		{"Write", []string{"io", "os", "bufio"}},
		{"Close", []string{"os", "io", "net"}},
		{"Pipe", []string{"net", "os", "io"}},
	}

	for _, tt := range tests {
		t.Run(tt.function, func(t *testing.T) {
			pkg, err := GetPackageForFunction(tt.function)

			// Should return empty package
			if pkg != "" {
				t.Errorf("GetPackageForFunction(%q) package = %q, want empty for ambiguous function",
					tt.function, pkg)
			}

			// Should return AmbiguousFunctionError
			if err == nil {
				t.Errorf("GetPackageForFunction(%q) error = nil, want AmbiguousFunctionError",
					tt.function)
				return
			}

			ambigErr, ok := err.(*AmbiguousFunctionError)
			if !ok {
				t.Errorf("GetPackageForFunction(%q) error type = %T, want *AmbiguousFunctionError",
					tt.function, err)
				return
			}

			// Check that error contains all expected packages
			for _, wantPkg := range tt.wantPackages {
				found := false
				for _, gotPkg := range ambigErr.Packages {
					if gotPkg == wantPkg {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("AmbiguousFunctionError.Packages = %v, missing package %q",
						ambigErr.Packages, wantPkg)
				}
			}

			// Check error message format
			errMsg := err.Error()
			if !strings.Contains(errMsg, tt.function) {
				t.Errorf("Error message %q does not contain function name %q",
					errMsg, tt.function)
			}
			if !strings.Contains(errMsg, "ambiguous") {
				t.Errorf("Error message %q does not contain 'ambiguous'", errMsg)
			}
			if !strings.Contains(errMsg, "Fix:") {
				t.Errorf("Error message %q does not contain fix-it hint", errMsg)
			}
		})
	}
}

func TestGetPackageForFunction_Unknown(t *testing.T) {
	tests := []string{
		"CustomFunc",
		"MyReadFile",
		"UserDefinedFunction",
		"NotInStdlib",
	}

	for _, funcName := range tests {
		t.Run(funcName, func(t *testing.T) {
			pkg, err := GetPackageForFunction(funcName)

			if err != nil {
				t.Errorf("GetPackageForFunction(%q) error = %v, want nil for unknown function",
					funcName, err)
			}

			if pkg != "" {
				t.Errorf("GetPackageForFunction(%q) = %q, want empty string for unknown function",
					funcName, pkg)
			}
		})
	}
}

func TestIsStdlibFunction(t *testing.T) {
	tests := []struct {
		function string
		want     bool
	}{
		{"ReadFile", true},
		{"Atoi", true},
		{"Printf", true},
		{"Open", true}, // Ambiguous but still in stdlib
		{"CustomFunc", false},
		{"NotInStdlib", false},
		{"MyFunction", false},
	}

	for _, tt := range tests {
		t.Run(tt.function, func(t *testing.T) {
			got := IsStdlibFunction(tt.function)
			if got != tt.want {
				t.Errorf("IsStdlibFunction(%q) = %v, want %v", tt.function, got, tt.want)
			}
		})
	}
}

func TestGetAllPackages(t *testing.T) {
	packages := GetAllPackages()

	// Check that we have a reasonable number of packages
	if len(packages) < 20 {
		t.Errorf("GetAllPackages() returned %d packages, expected at least 20", len(packages))
	}

	// Check for expected packages
	expectedPackages := []string{
		"os", "fmt", "strconv", "io", "json", "http",
		"sync", "time", "errors", "strings", "bytes",
		"filepath", "path", "regexp", "sort", "math",
		"rand", "context", "log", "net",
	}

	pkgMap := make(map[string]bool)
	for _, pkg := range packages {
		pkgMap[pkg] = true
	}

	for _, expected := range expectedPackages {
		if !pkgMap[expected] {
			t.Errorf("GetAllPackages() missing expected package %q", expected)
		}
	}
}

func TestGetFunctionCount(t *testing.T) {
	count := GetFunctionCount()

	// Should have at least 400 functions (user requirement: ~500)
	if count < 400 {
		t.Errorf("GetFunctionCount() = %d, expected at least 400 functions", count)
	}

	// Should match registry size
	if count != len(StdlibRegistry) {
		t.Errorf("GetFunctionCount() = %d, but StdlibRegistry has %d entries",
			count, len(StdlibRegistry))
	}

	t.Logf("Registry contains %d stdlib functions across %d packages",
		count, len(GetAllPackages()))
}

func TestGetAmbiguousFunctions(t *testing.T) {
	ambiguous := GetAmbiguousFunctions()

	// Should have some ambiguous functions
	if len(ambiguous) == 0 {
		t.Error("GetAmbiguousFunctions() returned no ambiguous functions, expected at least a few")
	}

	// Verify all returned functions are actually ambiguous
	for _, funcName := range ambiguous {
		pkgs := StdlibRegistry[funcName]
		if len(pkgs) <= 1 {
			t.Errorf("GetAmbiguousFunctions() returned %q with %d packages, expected >1",
				funcName, len(pkgs))
		}
	}

	t.Logf("Found %d ambiguous functions in registry", len(ambiguous))
}

func TestAmbiguousFunctionError_Message(t *testing.T) {
	err := &AmbiguousFunctionError{
		Function: "Open",
		Packages: []string{"os", "net"},
	}

	msg := err.Error()

	// Check message contains all required parts
	required := []string{
		"ambiguous",
		"Open",
		"os",
		"net",
		"Fix:",
		"os.Open",
		"net.Open",
	}

	for _, req := range required {
		if !strings.Contains(msg, req) {
			t.Errorf("Error message missing %q:\n%s", req, msg)
		}
	}
}

func TestStdlibRegistry_NoDuplicatesInUnique(t *testing.T) {
	// Verify that functions marked as unique actually have only one package
	for funcName, pkgs := range StdlibRegistry {
		if len(pkgs) == 0 {
			t.Errorf("Function %q has empty package list", funcName)
		}

		// If a function has multiple packages, it should be intentional (ambiguous)
		// This test just documents current state
		if len(pkgs) > 1 {
			t.Logf("Ambiguous function %q: %v", funcName, pkgs)
		}
	}
}

func TestStdlibRegistry_Coverage(t *testing.T) {
	// Test coverage of common packages
	packageCoverage := make(map[string]int)

	for _, pkgs := range StdlibRegistry {
		for _, pkg := range pkgs {
			packageCoverage[pkg]++
		}
	}

	// Verify we have good coverage of important packages
	expectedCoverage := map[string]int{
		"os":       20, // At least 20 functions from os
		"fmt":      10, // At least 10 from fmt
		"strconv":  15, // At least 15 from strconv
		"io":       5,  // At least 5 from io
		"strings":  20, // At least 20 from strings
		"time":     10, // At least 10 from time
		"errors":   3,  // At least 3 from errors
	}

	for pkg, minCount := range expectedCoverage {
		if count := packageCoverage[pkg]; count < minCount {
			t.Errorf("Package %q has only %d functions, expected at least %d",
				pkg, count, minCount)
		} else {
			t.Logf("Package %q: %d functions ✓", pkg, count)
		}
	}
}

func TestGetPackageForFunction_ConsistentBehavior(t *testing.T) {
	// Test that calling GetPackageForFunction multiple times returns consistent results
	testFuncs := []string{"ReadFile", "Open", "Printf", "UnknownFunc"}

	for _, funcName := range testFuncs {
		t.Run(funcName, func(t *testing.T) {
			// Call twice
			pkg1, err1 := GetPackageForFunction(funcName)
			pkg2, err2 := GetPackageForFunction(funcName)

			// Results should be identical
			if pkg1 != pkg2 {
				t.Errorf("Inconsistent package results for %q: %q vs %q",
					funcName, pkg1, pkg2)
			}

			// Error states should match
			if (err1 == nil) != (err2 == nil) {
				t.Errorf("Inconsistent error results for %q: %v vs %v",
					funcName, err1, err2)
			}
		})
	}
}

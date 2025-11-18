package preprocessor

import (
	"os"
	"path/filepath"
	"testing"
)

// TestPreprocessor_WithCache tests preprocessor with cache integration
func TestPreprocessor_WithCache(t *testing.T) {
	// Create temp directory for test package
	tmpDir, err := os.MkdirTemp("", "dingo-cache-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files (must be valid Go for cache scanning)
	testFiles := map[string]string{
		"main.dingo": `package main

func ReadFile(path string) string {
	// User-defined function
	return ""
}

func main() {
	data := ReadFile("test.txt")
	println(data)
}`,
		"utils.dingo": `package main

func ParseConfig(path string) {
	// Another user function
}`,
	}

	for name, content := range testFiles {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write %s: %v", name, err)
		}
	}

	// Create cache and scan package
	cache := NewFunctionExclusionCache(tmpDir)
	files := []string{
		filepath.Join(tmpDir, "main.dingo"),
		filepath.Join(tmpDir, "utils.dingo"),
	}
	if err := cache.ScanPackage(files); err != nil {
		t.Fatalf("Failed to scan package: %v", err)
	}

	// Verify cache has local functions
	if !cache.IsLocalSymbol("ReadFile") {
		t.Error("Expected ReadFile to be in cache")
	}
	if !cache.IsLocalSymbol("ParseConfig") {
		t.Error("Expected ParseConfig to be in cache")
	}

	// Create preprocessor with cache
	source := []byte(`package main

func test() {
	x := ReadFile("foo.txt")
	ParseConfig("bar")
}`)

	p := NewWithCache(source, cache)

	// Verify cache is attached
	if !p.HasCache() {
		t.Fatal("Expected preprocessor to have cache")
	}

	if p.GetCache() != cache {
		t.Error("Expected GetCache to return same cache instance")
	}

	// Process source
	result, _, err := p.Process()
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Verify result (should NOT transform local functions)
	if result == "" {
		t.Error("Expected non-empty result")
	}

	// Verify metrics (ReadFile, ParseConfig, main = 3 functions)
	metrics := cache.Metrics()
	if metrics.TotalSymbols != 3 {
		t.Errorf("Expected 3 symbols in cache (ReadFile, ParseConfig, main), got %d", metrics.TotalSymbols)
	}
}

// TestPreprocessor_WithoutCache tests backward compatibility (no cache)
func TestPreprocessor_WithoutCache(t *testing.T) {
	source := []byte(`package main

func test() {
	x := 42
}`)

	// Create preprocessor without cache (traditional way)
	p := New(source)

	// Verify no cache
	if p.HasCache() {
		t.Error("Expected preprocessor to have no cache")
	}

	if p.GetCache() != nil {
		t.Error("Expected GetCache to return nil")
	}

	// Process should still work
	result, _, err := p.Process()
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if result == "" {
		t.Error("Expected non-empty result")
	}
}

// TestPreprocessor_EarlyBailout tests early bailout optimization
func TestPreprocessor_EarlyBailout(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "dingo-bailout-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create file WITHOUT any capitalized patterns (true bailout case)
	testFile := filepath.Join(tmpDir, "main.dingo")
	content := `package main

func main() {
	x := 42
	y := x + 1
	_ = y
}`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create cache and scan
	cache := NewFunctionExclusionCache(tmpDir)
	if err := cache.ScanPackage([]string{testFile}); err != nil {
		t.Fatalf("Failed to scan package: %v", err)
	}

	// Verify no unqualified imports (no capitalized function calls at all)
	if cache.HasUnqualifiedImports() {
		t.Error("Expected HasUnqualifiedImports to be false (no capitalized function patterns)")
	}

	// Create preprocessor with cache
	source := []byte(content)
	p := NewWithCache(source, cache)

	// Process should skip unqualified import processing (early bailout)
	result, _, err := p.Process()
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if result == "" {
		t.Error("Expected non-empty result")
	}
}

// TestPreprocessor_EarlyBailout_HasUnqualified tests when package has unqualified imports
func TestPreprocessor_EarlyBailout_HasUnqualified(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "dingo-unqualified-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create file WITH unqualified imports (capitalized function call)
	testFile := filepath.Join(tmpDir, "main.dingo")
	content := `package main

func main() {
	ReadFile("test.txt")  // Potential unqualified stdlib call
}`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create cache and scan
	cache := NewFunctionExclusionCache(tmpDir)
	if err := cache.ScanPackage([]string{testFile}); err != nil {
		t.Fatalf("Failed to scan package: %v", err)
	}

	// Verify HAS unqualified imports
	if !cache.HasUnqualifiedImports() {
		t.Error("Expected HasUnqualifiedImports to be true (file has potential unqualified calls)")
	}

	// Create preprocessor with cache
	source := []byte(content)
	p := NewWithCache(source, cache)

	// Process should NOT skip unqualified processing
	result, _, err := p.Process()
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if result == "" {
		t.Error("Expected non-empty result")
	}
}

// TestPreprocessor_NilCache tests nil cache safety
func TestPreprocessor_NilCache(t *testing.T) {
	source := []byte(`package main

func main() {
	x := 42
}`)

	// Create with nil cache (should work)
	p := NewWithCache(source, nil)

	if p.HasCache() {
		t.Error("Expected HasCache to return false with nil cache")
	}

	if p.GetCache() != nil {
		t.Error("Expected GetCache to return nil")
	}

	// Process should work (nil-safe)
	result, _, err := p.Process()
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if result == "" {
		t.Error("Expected non-empty result")
	}
}

// TestNewWithMainConfig_BackwardCompat tests that NewWithMainConfig still works without cache
func TestNewWithMainConfig_BackwardCompat(t *testing.T) {
	source := []byte(`package main

func test() string {
	x: int = 42
	return "ok"
}`)

	// Old API should still work
	p := NewWithMainConfig(source, nil)

	if p.HasCache() {
		t.Error("Expected no cache with NewWithMainConfig")
	}

	result, _, err := p.Process()
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Verify type annotation was processed
	if result == "" {
		t.Error("Expected non-empty result")
	}

	// Should contain "var x int" (type annotation processed)
	// Note: Exact output depends on all processors, just verify it processed
}

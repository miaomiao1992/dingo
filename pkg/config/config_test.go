package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Features.ErrorPropagationSyntax != SyntaxQuestion {
		t.Errorf("Expected default syntax to be 'question', got %q", cfg.Features.ErrorPropagationSyntax)
	}

	if !cfg.SourceMap.Enabled {
		t.Error("Expected source maps to be enabled by default")
	}

	if cfg.SourceMap.Format != FormatInline {
		t.Errorf("Expected default format to be 'inline', got %q", cfg.SourceMap.Format)
	}
}

func TestSyntaxStyleValidation(t *testing.T) {
	tests := []struct {
		style SyntaxStyle
		valid bool
	}{
		{SyntaxQuestion, true},
		{SyntaxBang, true},
		{SyntaxTry, true},
		{SyntaxStyle("invalid"), false},
		{SyntaxStyle(""), false},
		{SyntaxStyle("QUESTION"), false}, // Case sensitive
	}

	for _, tt := range tests {
		t.Run(string(tt.style), func(t *testing.T) {
			got := tt.style.IsValid()
			if got != tt.valid {
				t.Errorf("IsValid() = %v, want %v for %q", got, tt.valid, tt.style)
			}
		})
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid default config",
			config:    DefaultConfig(),
			wantError: false,
		},
		{
			name: "valid bang syntax",
			config: &Config{
				Features: FeatureConfig{
					ErrorPropagationSyntax: SyntaxBang,
				},
				SourceMap: SourceMapConfig{
					Enabled: true,
					Format:  FormatSeparate,
				},
			},
			wantError: false,
		},
		{
			name: "invalid syntax",
			config: &Config{
				Features: FeatureConfig{
					ErrorPropagationSyntax: SyntaxStyle("invalid"),
				},
				SourceMap: SourceMapConfig{
					Enabled: true,
					Format:  FormatInline,
				},
			},
			wantError: true,
			errorMsg:  "invalid error_propagation_syntax",
		},
		{
			name: "invalid source map format",
			config: &Config{
				Features: FeatureConfig{
					ErrorPropagationSyntax: SyntaxQuestion,
				},
				SourceMap: SourceMapConfig{
					Enabled: true,
					Format:  SourceMapFormat("bad_format"),
				},
			},
			wantError: true,
			errorMsg:  "invalid sourcemap format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error containing %q, got nil", tt.errorMsg)
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

func TestLoadConfigNoFiles(t *testing.T) {
	// Create temp directory without config files
	tmpDir, err := os.MkdirTemp("", "dingo-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Override HOME to avoid loading user config
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Load config - should use defaults
	cfg, err := Load(nil)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Features.ErrorPropagationSyntax != SyntaxQuestion {
		t.Errorf("Expected default syntax 'question', got %q", cfg.Features.ErrorPropagationSyntax)
	}
}

func TestLoadConfigProjectFile(t *testing.T) {
	// Create temp directory with project config
	tmpDir, err := os.MkdirTemp("", "dingo-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Write project config
	projectConfig := `[features]
error_propagation_syntax = "bang"

[sourcemaps]
enabled = true
format = "separate"
`
	configPath := filepath.Join(tmpDir, "dingo.toml")
	if err := os.WriteFile(configPath, []byte(projectConfig), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Override HOME to avoid loading user config
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Load config
	cfg, err := Load(nil)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Features.ErrorPropagationSyntax != SyntaxBang {
		t.Errorf("Expected syntax 'bang' from project config, got %q", cfg.Features.ErrorPropagationSyntax)
	}

	if cfg.SourceMap.Format != FormatSeparate {
		t.Errorf("Expected format 'separate' from project config, got %q", cfg.SourceMap.Format)
	}
}

func TestLoadConfigCLIOverride(t *testing.T) {
	// Create temp directory with project config
	tmpDir, err := os.MkdirTemp("", "dingo-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Write project config with "bang"
	projectConfig := `[features]
error_propagation_syntax = "bang"
`
	configPath := filepath.Join(tmpDir, "dingo.toml")
	if err := os.WriteFile(configPath, []byte(projectConfig), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Override HOME
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Load config with CLI override to "try"
	overrides := &Config{
		Features: FeatureConfig{
			ErrorPropagationSyntax: SyntaxTry,
		},
	}

	cfg, err := Load(overrides)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// CLI override should win
	if cfg.Features.ErrorPropagationSyntax != SyntaxTry {
		t.Errorf("Expected syntax 'try' from CLI override, got %q", cfg.Features.ErrorPropagationSyntax)
	}
}

func TestLoadConfigInvalidTOML(t *testing.T) {
	// Create temp directory with invalid TOML
	tmpDir, err := os.MkdirTemp("", "dingo-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Write invalid TOML
	invalidConfig := `[features
error_propagation_syntax = "bang"  # Missing closing bracket
`
	configPath := filepath.Join(tmpDir, "dingo.toml")
	if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Override HOME
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Load should fail
	_, err = Load(nil)
	if err == nil {
		t.Error("Expected error for invalid TOML, got nil")
	}
}

func TestLoadConfigInvalidValue(t *testing.T) {
	// Create temp directory with invalid syntax value
	tmpDir, err := os.MkdirTemp("", "dingo-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Write config with invalid syntax
	invalidConfig := `[features]
error_propagation_syntax = "invalid_syntax"
`
	configPath := filepath.Join(tmpDir, "dingo.toml")
	if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Override HOME
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Load should fail validation
	_, err = Load(nil)
	if err == nil {
		t.Error("Expected validation error, got nil")
	}
	if !contains(err.Error(), "invalid configuration") {
		t.Errorf("Expected 'invalid configuration' error, got %v", err)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

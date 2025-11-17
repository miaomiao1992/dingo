package preprocessor

import "fmt"

// Config holds preprocessor configuration options
type Config struct {
	// MultiValueReturnMode controls error propagation behavior for multi-value returns
	// - "full": Support (A, B, C, error) propagation (default)
	// - "single": Restrict to (T, error) only
	MultiValueReturnMode string
}

// DefaultConfig returns the default preprocessor configuration
func DefaultConfig() *Config {
	return &Config{
		MultiValueReturnMode: "full",
	}
}

// ValidateMultiValueReturnMode checks if the mode is valid
func (c *Config) ValidateMultiValueReturnMode() error {
	switch c.MultiValueReturnMode {
	case "full", "single":
		return nil
	default:
		return fmt.Errorf("invalid multi-value return mode: %q (must be 'full' or 'single')", c.MultiValueReturnMode)
	}
}

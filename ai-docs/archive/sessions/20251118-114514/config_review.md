The `pkg/config` package defines the configuration structure for the Dingo compiler, handling various feature flags, syntax styles, and source map options. It uses the `toml` format for configuration files and supports a cascading load order: built-in defaults, user config (`~/.dingo/config.toml`), project config (`dingo.toml`), and CLI overrides. The design includes robust validation for all configurable fields, ensuring that invalid values are caught early with descriptive error messages. Error handling is consistent, returning wrapped errors for easier debugging.

### Configuration Structure
The main `Config` struct aggregates `FeatureConfig`, `MatchConfig`, and `SourceMapConfig`.
- `FeatureConfig`: Encompasses settings for error propagation syntax (`?`, `!`, `try`), `Result` and `Option` types, nil safety checks, lambda syntax, safe navigation, null coalescing, and operator precedence.
- `MatchConfig`: Defines the syntax style for pattern matching (`rust`, `swift`).
- `SourceMapConfig`: Controls source map generation and format (`inline`, `separate`, `both`, `none`).

### Defaults and Loading Order
`DefaultConfig()` provides a baseline configuration with sensible defaults for all features. The `Load` function implements a clear precedence for loading configurations:
1. CLI overrides (highest priority)
2. Project `dingo.toml`
3. User config `~/.dingo/config.toml`
4. Built-in defaults (lowest priority)

This layered approach allows for granular control over configuration, from global user settings to project-specific overrides and command-line adjustments.

### Validation
A comprehensive `Validate()` method within the `Config` struct ensures that all string-based configuration values adhere to a predefined set of valid options. This is implemented using `switch` statements, providing clear error messages for invalid inputs. This proactive validation prevents the compiler from operating with malformed or unsupported settings.

### Error Handling
The `loadConfigFile` helper gracefully handles the absence of configuration files by returning `nil` if `os.IsNotExist(err)` is true. For actual parsing errors, it returns a wrapped error using `fmt.Errorf("failed to parse %s: %w")`, preserving the original error context. The `Load` function itself wraps validation errors with `fmt.Errorf("invalid configuration: %w")`, maintaining a clear error chain.

### Go Best Practices
- **Clear Structs**: Configuration options are logically grouped into well-named structs.
- **Named Constants**: `SyntaxStyle` and `SourceMapFormat` use named constants, improving readability and maintainability.
- **Methods on Types**: `SyntaxStyle.IsValid()` and `Config.Validate()` demonstrate methods attached to their respective types, promoting encapsulation and clear responsibility.
- **Error Wrapping**: Consistent use of `fmt.Errorf("%w")` for error wrapping.
- **Test Coverage**: `config_test.go` provides good test coverage for default configurations, validation logic, and the cascading load mechanism, including tests for invalid configurations and file loading scenarios.

### Summary
The `pkg/config` package is well-designed, robust, and correctly implements configuration parsing, validation, and loading. It provides a flexible yet controlled mechanism for managing Dingo compiler settings, adhering to Go best practices for error handling and code structure.
# Task E: Implementation Changes

## Files Created

### 1. `/Users/jack/mag/dingo/pkg/preprocessor/config.go` (NEW FILE)
Complete file created with:
- Lines 1-27: Package declaration, Config struct, DefaultConfig(), ValidateMultiValueReturnMode()
- MultiValueReturnMode field: controls "full" (default) vs "single" mode
- Validation ensures only "full" or "single" values are accepted

## Files Modified

### 2. `/Users/jack/mag/dingo/pkg/preprocessor/preprocessor.go`
**Lines 18-21**: Added config field to Preprocessor struct
```go
type Preprocessor struct {
	source     []byte
	processors []FeatureProcessor
	config     *Config // Configuration for preprocessor behavior
}
```

**Lines 42-74**: Replaced New() and added NewWithConfig()
- Old New() function now wraps NewWithConfig() with nil config
- NewWithConfig() accepts optional config, defaults to DefaultConfig() if nil
- Passes config to NewErrorPropProcessorWithConfig() at line 61

### 3. `/Users/jack/mag/dingo/pkg/preprocessor/error_prop.go`
**Line 150**: Added config field to ErrorPropProcessor struct
```go
config *Config // Configuration for preprocessor behavior
```

**Lines 159-173**: Updated constructor functions
- NewErrorPropProcessor() now wraps NewErrorPropProcessorWithConfig(nil)
- NewErrorPropProcessorWithConfig(config) is the new constructor that accepts config
- Config defaults to DefaultConfig() if nil

**Lines 255, 315, 429**: Updated function signatures to return errors
- processLine: (string, []Mapping) → (string, []Mapping, error)
- expandAssignment: (string, []Mapping) → (string, []Mapping, error)
- expandReturn: (string, []Mapping) → (string, []Mapping, error)

**Lines 204-207**: Added error handling in Process() method
```go
transformed, newMappings, err := e.processLine(line, inputLineNum+1, outputLineNum)
if err != nil {
	return nil, nil, fmt.Errorf("line %d: %w", inputLineNum+1, err)
}
```

**Lines 258-293**: Updated processLine to propagate errors from expandAssignment/expandReturn

**Lines 271-275, 284-288**: Error propagation in assignment and return patterns

**Lines 424, 568**: Updated expandAssignment and expandReturn to return nil error on success

**Lines 442-450**: **CORE FEATURE** - Multi-value return mode enforcement in expandReturn()
```go
// Check config mode: enforce single-value restriction if configured
if e.config != nil && e.config.MultiValueReturnMode == "single" && numNonErrorReturns > 1 {
	// Return error - will be caught and reported by Process()
	return "", nil, fmt.Errorf(
		"multi-value error propagation not allowed in 'single' mode (use --multi-value-return=full): function returns %d values plus error",
		numNonErrorReturns,
	)
}
```

### 4. `/Users/jack/mag/dingo/cmd/dingo/main.go`
**Lines 66-70**: Added multiValueReturnMode flag variable to buildCmd

**Lines 86**: Added example usage in buildCmd help text
```
dingo build --multi-value-return=single file.dingo  # Restrict to (T, error) only
```

**Line 95**: Added flag registration
```go
cmd.Flags().StringVar(&multiValueReturnMode, "multi-value-return", "full", "Multi-value return propagation mode: 'full' (default, supports (A,B,error)) or 'single' (restricts to (T,error))")
```

**Line 89**: Updated runBuild call to pass multiValueReturnMode parameter

**Lines 101**: Added multiValueReturnMode flag variable to runCmd

**Line 119**: Added example usage in runCmd help text

**Line 130**: Updated runDingoFile call to pass multiValueReturnMode parameter

**Line 134**: Added flag registration to runCmd

**Lines 142-151**: Added config creation and validation in runBuild()
```go
// Create config from flags
config := &preprocessor.Config{
	MultiValueReturnMode: multiValueReturnMode,
}

// Validate config
if err := config.ValidateMultiValueReturnMode(); err != nil {
	return fmt.Errorf("configuration error: %w", err)
}
```

**Line 167**: Updated buildFile call to pass config parameter

**Line 190**: Updated buildFile signature to accept config parameter

**Line 211**: Updated preprocessor.New() to preprocessor.NewWithConfig(src, config)

**Line 331**: Updated runDingoFile signature to accept multiValueReturnMode parameter

**Lines 350-359**: Added config creation and validation in runDingoFile()

**Line 369**: Updated preprocessor.New() to preprocessor.NewWithConfig(src, config) in runDingoFile

## Summary
- **1 new file created**: config.go
- **4 files modified**: preprocessor.go, error_prop.go, main.go
- **Total lines changed**: ~120 lines across all files
- **Build status**: ✅ Successful (verified with `go build ./cmd/dingo`)

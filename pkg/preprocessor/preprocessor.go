// Package preprocessor transforms Dingo syntax to valid Go syntax
package preprocessor

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"sort"
	"strings"

	"github.com/MadAppGang/dingo/pkg/config"
	"golang.org/x/tools/go/ast/astutil"
)

// Preprocessor orchestrates multiple feature processors to transform
// Dingo source code into valid Go code with semantic placeholders
type Preprocessor struct {
	source     []byte
	processors []FeatureProcessor
	oldConfig  *Config        // Deprecated: Legacy preprocessor-specific config
	config     *config.Config // Main Dingo configuration

	// Package-wide cache (optional, for unqualified import inference)
	// When present, enables early bailout optimization and local function exclusion
	cache *FunctionExclusionCache

	// Source map generation mode (Phase 2: Post-AST support)
	mode PreprocessorMode
}

// PreprocessorMode controls which source map generation mode is used
type PreprocessorMode int

const (
	// ModeLegacy uses the old source map generation (predictions during preprocessing)
	ModeLegacy PreprocessorMode = iota
	// ModePostAST uses the new Post-AST source map generation (Phase 2)
	ModePostAST
	// ModeDual generates both legacy mappings and Post-AST metadata (for testing/migration)
	ModeDual
)

// TransformMetadata holds metadata about a transformation (NOT final mappings)
// This is emitted by preprocessors and used by Post-AST generator to match AST nodes
type TransformMetadata struct {
	Type            string // "error_prop", "type_annot", "enum", etc.
	OriginalLine    int    // Line in .dingo file
	OriginalColumn  int    // Column in .dingo file
	OriginalLength  int    // Length in .dingo file
	OriginalText    string // Original Dingo syntax (e.g., "?")
	GeneratedMarker string // Unique marker in Go code (e.g., "// dingo:e:0")
	ASTNodeType     string // "CallExpr", "FuncDecl", "IfStmt", etc.
}

// ProcessResult holds the result of preprocessing
// Supports both legacy mappings and new Post-AST metadata
type ProcessResult struct {
	Source   []byte              // Transformed Go source code
	Mappings []Mapping           // LEGACY: For backward compatibility
	Metadata []TransformMetadata // NEW: For Post-AST generation
}

// FeatureProcessor defines the interface for individual feature preprocessors
type FeatureProcessor interface {
	// Name returns the feature name for logging/debugging
	Name() string

	// Process transforms the source code and returns:
	// - transformed source
	// - source mappings for error reporting
	// - error if transformation failed
	Process(source []byte) ([]byte, []Mapping, error)
}

// FeatureProcessorV2 is the new interface that supports Post-AST metadata emission
// Processors can implement this interface to support the new Post-AST source map generation
type FeatureProcessorV2 interface {
	FeatureProcessor // Embed the old interface for backward compatibility

	// ProcessV2 transforms the source code and returns a ProcessResult
	// This method supports both legacy mappings and Post-AST metadata
	ProcessV2(source []byte, mode PreprocessorMode) (ProcessResult, error)
}

// ImportProvider is an optional interface for processors that need to add imports
type ImportProvider interface {
	// GetNeededImports returns list of import paths that should be added
	GetNeededImports() []string
}

// New creates a new preprocessor with all registered features and default config
func New(source []byte) *Preprocessor {
	return NewWithMainConfig(source, nil)
}

// NewWithConfig creates a new preprocessor with legacy config (deprecated)
// Use NewWithMainConfig instead
func NewWithConfig(source []byte, legacyConfig *Config) *Preprocessor {
	// Convert legacy config to main config
	cfg := config.DefaultConfig()
	if legacyConfig != nil && legacyConfig.MultiValueReturnMode == "single" {
		// Map legacy mode to main config (feature not in main config yet)
	}
	return NewWithMainConfig(source, cfg)
}

// NewWithMainConfig creates a new preprocessor with main Dingo configuration
func NewWithMainConfig(source []byte, cfg *config.Config) *Preprocessor {
	return newWithConfigAndCache(source, cfg, nil)
}

// newWithConfigAndCache is the internal constructor that accepts an optional cache
func newWithConfigAndCache(source []byte, cfg *config.Config, cache *FunctionExclusionCache) *Preprocessor {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	processors := []FeatureProcessor{
		// Order matters! Process in this sequence:
		// 0. Generic syntax (<> → []) - must be FIRST before type annotations
		NewGenericSyntaxProcessor(),
		// 1. Pattern matching (match) - MUST run BEFORE lambdas (both use =>)
		//    Match arms: Pattern => Expression (structural context)
		//    Lambdas: params => expression (expression context)
		NewRustMatchProcessor(),
		// 2. Lambdas (x => expr, |x| expr) - AFTER pattern matching
		NewLambdaProcessorWithConfig(cfg),
		// 3. Type annotations (: → space) - after lambdas, after generic syntax
		NewTypeAnnotProcessor(),
		// 4. Tuples ((a, b) = (1, 2)) - BEFORE safe navigation (uses . in field access)
		NewTupleProcessor(),
		// 5. Safe navigation (?.) - BEFORE null coalescing (SafeNav handles ?. before NullCoalesce sees ??)
		NewSafeNavProcessor(),
		// 6. Null coalescing (??) - AFTER safe navigation, BEFORE ternary
		//    CRITICAL: Must run BEFORE TernaryProcessor and ErrorPropProcessor
		NewNullCoalesceProcessor(),
		// 7. Ternary operator (? :) - AFTER null coalescing, BEFORE error propagation
		//    Process ternary BEFORE error prop to cleanly separate ? : from single ?
		NewTernaryProcessor(),
		// 8. Error propagation (expr?) - AFTER ternary (handles remaining ?)
		NewErrorPropProcessor(),
	}

	// 9. Enums (enum Name { ... }) - after error prop, before keywords
	processors = append(processors, NewEnumProcessor())

	// 10. Keywords (let → var) - after pattern match, error prop, and enum
	processors = append(processors, NewKeywordProcessor())

	// 11. Unqualified imports (ReadFile → os.ReadFile) - requires cache
	if cache != nil {
		processors = append(processors, NewUnqualifiedImportProcessor(cache))
	}

	return &Preprocessor{
		source:     source,
		config:     cfg,
		oldConfig:  nil, // No longer used
		processors: processors,
		cache:      cache,
		mode:       ModeLegacy, // Default to legacy mode for backward compatibility
	}
}

// SetMode sets the source map generation mode
// This allows switching between legacy and Post-AST modes
func (p *Preprocessor) SetMode(mode PreprocessorMode) {
	p.mode = mode
}

// GetMode returns the current source map generation mode
func (p *Preprocessor) GetMode() PreprocessorMode {
	return p.mode
}

// ProcessWithMetadata runs all feature processors and returns both legacy mappings and Post-AST metadata
// This is the new unified method that supports all modes (Legacy, PostAST, Dual)
func (p *Preprocessor) ProcessWithMetadata() (string, *SourceMap, []TransformMetadata, error) {
	// Early bailout optimization (GPT-5.1): If cache indicates no unqualified imports
	// in this package, skip expensive symbol resolution for unqualified import processors
	skipUnqualifiedProcessing := false
	if p.cache != nil && !p.cache.HasUnqualifiedImports() {
		// This package has no unqualified stdlib calls, skip that processing
		skipUnqualifiedProcessing = true
	}
	_ = skipUnqualifiedProcessing // TODO: Use when UnqualifiedImportProcessor is integrated

	result := p.source
	sourceMap := NewSourceMap()
	allMetadata := []TransformMetadata{}
	neededImports := []string{}

	// Run each processor in sequence
	for _, proc := range p.processors {
		// Check if processor implements V2 interface
		if procV2, ok := proc.(FeatureProcessorV2); ok {
			// Use new ProcessV2 method
			procResult, err := procV2.ProcessV2(result, p.mode)
			if err != nil {
				return "", nil, nil, fmt.Errorf("%s preprocessing failed: %w", proc.Name(), err)
			}

			// Update result
			result = procResult.Source

			// Merge mappings (if in Legacy or Dual mode)
			if p.mode == ModeLegacy || p.mode == ModeDual {
				for _, m := range procResult.Mappings {
					sourceMap.AddMapping(m)
				}
			}

			// Collect metadata (if in PostAST or Dual mode)
			if p.mode == ModePostAST || p.mode == ModeDual {
				allMetadata = append(allMetadata, procResult.Metadata...)
			}
		} else {
			// Fall back to legacy Process method
			processed, mappings, err := proc.Process(result)
			if err != nil {
				return "", nil, nil, fmt.Errorf("%s preprocessing failed: %w", proc.Name(), err)
			}

			// Update result
			result = processed

			// Merge mappings (legacy mode always uses mappings)
			for _, m := range mappings {
				sourceMap.AddMapping(m)
			}
		}

		// Collect needed imports if processor implements ImportProvider
		if importProvider, ok := proc.(ImportProvider); ok {
			imports := importProvider.GetNeededImports()
			neededImports = append(neededImports, imports...)
		}
	}

	// Inject all needed imports at the end (after all transformations complete)
	if len(neededImports) > 0 {
		var importInsertLine, importBlockEndLine int
		var err error
		// CRITICAL FIX: Get both import start and end lines for accurate shifting
		result, importInsertLine, importBlockEndLine, err = injectImportsWithPosition(result, neededImports)
		if err != nil {
			return "", nil, nil, fmt.Errorf("failed to inject imports: %w", err)
		}

		// Calculate how many lines the import block occupies
		// importInsertLine is where imports are inserted (after package declaration)
		// importBlockEndLine is where imports end (last line of import block)
		// CRITICAL FIX: Only apply adjustment if imports were actually added
		if importInsertLine > 0 && importBlockEndLine > 0 {
			importBlockSize := importBlockEndLine - importInsertLine + 1

			// Adjust all source mappings to account for added import lines
			if p.mode == ModeLegacy || p.mode == ModeDual {
				adjustMappingsForImports(sourceMap, importBlockSize, importInsertLine)
			}

			// TODO: Adjust metadata line numbers for Post-AST mode
			// This will be needed when we integrate with Phase 3
		}
	}

	return string(result), sourceMap, allMetadata, nil
}

// Process runs all feature processors in sequence and combines source maps
// This is the legacy method that returns only source maps (for backward compatibility)
func (p *Preprocessor) Process() (string, *SourceMap, error) {
	// Delegate to ProcessWithMetadata and discard metadata
	result, sourceMap, _, err := p.ProcessWithMetadata()
	return result, sourceMap, err
}

// DEPRECATED: Old Process implementation kept for reference during migration
// Will be removed after all callers migrate to ProcessWithMetadata
func (p *Preprocessor) processLegacy() (string, *SourceMap, error) {
	// Early bailout optimization (GPT-5.1): If cache indicates no unqualified imports
	// in this package, skip expensive symbol resolution for unqualified import processors
	skipUnqualifiedProcessing := false
	if p.cache != nil && !p.cache.HasUnqualifiedImports() {
		// This package has no unqualified stdlib calls, skip that processing
		skipUnqualifiedProcessing = true
	}
	_ = skipUnqualifiedProcessing // TODO: Use when UnqualifiedImportProcessor is integrated

	result := p.source
	sourceMap := NewSourceMap()
	neededImports := []string{}

	// Run each processor in sequence
	for _, proc := range p.processors {
		processed, mappings, err := proc.Process(result)
		if err != nil {
			return "", nil, fmt.Errorf("%s preprocessing failed: %w", proc.Name(), err)
		}

		// Update result
		result = processed

		// Merge mappings
		for _, m := range mappings {
			sourceMap.AddMapping(m)
		}

		// Collect needed imports if processor implements ImportProvider
		if importProvider, ok := proc.(ImportProvider); ok {
			imports := importProvider.GetNeededImports()
			neededImports = append(neededImports, imports...)
		}
	}

	// Inject all needed imports at the end (after all transformations complete)
	if len(neededImports) > 0 {
		var importInsertLine, importBlockEndLine int
		var err error
		// CRITICAL FIX: Get both import start and end lines for accurate shifting
		result, importInsertLine, importBlockEndLine, err = injectImportsWithPosition(result, neededImports)
		if err != nil {
			return "", nil, fmt.Errorf("failed to inject imports: %w", err)
		}

		// Calculate how many lines the import block occupies
		// importInsertLine is where imports are inserted (after package declaration)
		// importBlockEndLine is where imports end (last line of import block)
		// CRITICAL FIX: Only apply adjustment if imports were actually added
		if importInsertLine > 0 && importBlockEndLine > 0 {
			// Calculate the number of lines added by the import block
			//
			// Example - multi-line import:
			// BEFORE import injection (preprocessed code):
			//   Line 1: package main
			//   Line 2: [blank]
			//   Line 3: func readConfig(...) {
			//   Line 4:     tmp, err := os.ReadFile(path)  ← mapping says gen_line=4
			//
			// AFTER import injection:
			//   Line 1: package main
			//   Line 2: [blank]
			//   Line 3: import (             ← importInsertLine is BEFORE this (line 2)
			//   Line 4:     "os"
			//   Line 5: )                    ← importBlockEndLine = 5
			//   Line 6: [blank line added by go/printer]
			//   Line 7: func readConfig(...) {
			//   Line 8:     tmp, err := os.ReadFile(path)  ← should be gen_line=8
			//
			// Calculation:
			//   importInsertLine = 2 (line after package, before imports start)
			//   importBlockEndLine = 5 (last line of import block)
			//   Shift needed = 8 - 4 = 4 lines
			//   Formula: importBlockEndLine - importInsertLine + 1 = 5 - 2 + 1 = 4 ✓
			//
			// The +1 accounts for the blank line that go/printer adds after the import block
			importBlockSize := importBlockEndLine - importInsertLine + 1

			// Adjust all source mappings to account for added import lines
			adjustMappingsForImports(sourceMap, importBlockSize, importInsertLine)
		}
	}

	return string(result), sourceMap, nil
}

// ProcessBytes is like Process but returns bytes
func (p *Preprocessor) ProcessBytes() ([]byte, *SourceMap, error) {
	str, sm, err := p.Process()
	if err != nil {
		return nil, nil, err
	}
	return []byte(str), sm, nil
}

// GetCache returns the function exclusion cache (if present)
// Returns nil if preprocessor was created without a cache
func (p *Preprocessor) GetCache() *FunctionExclusionCache {
	return p.cache
}

// HasCache returns true if this preprocessor has a package-wide cache
func (p *Preprocessor) HasCache() bool {
	return p.cache != nil
}

// injectImportsWithPosition adds needed imports to the source code and returns the insertion line and end line
// CRITICAL FIX: Now returns both start and end lines of import block for accurate source map adjustment
// Returns: modified source, import block start line (1-based), import block end line (1-based), and error
func injectImportsWithPosition(source []byte, needed []string) ([]byte, int, int, error) {
	// Parse the source
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", source, parser.ParseComments)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to parse source for import injection: %w", err)
	}

	// Deduplicate and sort needed imports
	importMap := make(map[string]bool)
	for _, pkg := range needed {
		importMap[pkg] = true
	}

	// Remove packages that are already imported
	for _, imp := range node.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		delete(importMap, path)
	}

	// If no new imports needed, return original
	if len(importMap) == 0 {
		return source, 0, 0, nil
	}

	// Determine import insertion line (after package declaration, before first decl)
	importInsertLine := 1
	if node.Name != nil {
		// Line after package declaration (typically line 1 or 2)
		importInsertLine = fset.Position(node.Name.End()).Line + 1
	}

	// Convert map to sorted slice
	finalImports := make([]string, 0, len(importMap))
	for pkg := range importMap {
		finalImports = append(finalImports, pkg)
	}
	sort.Strings(finalImports)

	// Add each import using astutil
	for _, pkg := range finalImports {
		astutil.AddImport(fset, node, pkg)
	}

	// CRITICAL FIX: Print imports in multi-line format to match generator output
	// The generator always uses "import (\n\t...\n)" format, so we must do the same
	// to ensure source map line numbers are correct
	var buf bytes.Buffer
	cfg := printer.Config{
		Mode:     printer.TabIndent | printer.UseSpaces,
		Tabwidth: 8,
	}

	// Print package statement
	fmt.Fprintf(&buf, "package %s\n\n", node.Name.Name)

	// Print imports in multi-line format (matching generator behavior)
	if len(node.Imports) > 0 {
		buf.WriteString("import (\n")
		for _, imp := range node.Imports {
			if err := cfg.Fprint(&buf, fset, imp); err != nil {
				return nil, 0, 0, fmt.Errorf("failed to print import: %w", err)
			}
			buf.WriteString("\n")
		}
		buf.WriteString(")\n\n")
	}

	// Print the rest of declarations (skip imports - already printed)
	for _, decl := range node.Decls {
		// Skip import declarations (already printed above)
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			continue
		}

		if err := cfg.Fprint(&buf, fset, decl); err != nil {
			return nil, 0, 0, fmt.Errorf("failed to print declaration: %w", err)
		}
		buf.WriteString("\n")
	}

	formatted := buf.Bytes()

	// CRITICAL FIX: Determine import block end line by scanning the FORMATTED output
	// This ensures line numbers match the final generated code
	resultStr := string(formatted)
	lines := strings.Split(resultStr, "\n")

	// Find the last import line by looking for import declarations
	importBlockEndLine := importInsertLine
	inImportBlock := false
	for i, line := range lines {
		lineNum := i + 1 // 1-based line number
		trimmed := strings.TrimSpace(line)

		// Check if we're entering the import block
		if strings.HasPrefix(trimmed, "import") {
			inImportBlock = true
			importBlockEndLine = lineNum
			continue
		}

		// If we're in import block, check if we've reached the end
		if inImportBlock {
			// Check if we reached a declaration (end of import section)
			if strings.HasPrefix(trimmed, "func") || strings.HasPrefix(trimmed, "type") ||
				strings.HasPrefix(trimmed, "const") || strings.HasPrefix(trimmed, "var") {
				// Reached first declaration. Import block has ended.
				break
			}

			// Update end line for any non-blank line in import block
			if trimmed != "" {
				importBlockEndLine = lineNum
			}
		}
	}

	return formatted, importInsertLine, importBlockEndLine, nil
}

// adjustMappingsForImports shifts mapping line numbers to account for added imports
// CRITICAL FIX: Shifts mappings for lines AFTER the import insertion point
func adjustMappingsForImports(sourceMap *SourceMap, numImportLines int, importInsertionLine int) {
	for i := range sourceMap.Mappings {
		// CRITICAL FIX: Only shift mappings for lines AFTER import insertion
		//
		// importInsertionLine is the line number (1-based) where imports are inserted
		// (typically line 2 or 3, right after the package declaration).
		//
		// We use > (not >=) to exclude the insertion line itself. Mappings AT the
		// insertion line are for package-level declarations BEFORE the imports, and
		// should NOT be shifted.
		//
		// Example:
		//   Line 1: package main
		//   Line 2: [IMPORTS INSERTED HERE] ← importInsertionLine = 2
		//   Line 3: func foo() { ... } (shifts to line 7 if 4-line import block added)
		//
		// Mappings with GeneratedLine=1 or 2 stay as-is.
		// Mappings with GeneratedLine=3+ are shifted by numImportLines.
		if sourceMap.Mappings[i].GeneratedLine > importInsertionLine {
			sourceMap.Mappings[i].GeneratedLine += numImportLines
		}
	}
}

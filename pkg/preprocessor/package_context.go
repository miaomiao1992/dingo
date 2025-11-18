// Package preprocessor provides package-level transpilation orchestration
package preprocessor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PackageContext orchestrates package-level transpilation with caching support
type PackageContext struct {
	packagePath string
	dingoFiles  []string
	cache       *FunctionExclusionCache

	// Build options
	incremental bool // Watch mode (use cache if valid)
	force       bool // Skip cache, always rescan
	verbose     bool // Show cache statistics
}

// BuildOptions controls package build behavior
type BuildOptions struct {
	Incremental bool // Enable incremental mode (watch mode)
	Force       bool // Force full rebuild, skip cache
	Verbose     bool // Show cache statistics
}

// DefaultBuildOptions returns default build configuration
func DefaultBuildOptions() BuildOptions {
	return BuildOptions{
		Incremental: false,
		Force:       false,
		Verbose:     false,
	}
}

// NewPackageContext creates a new package context with automatic .dingo file discovery
func NewPackageContext(packageDir string, opts BuildOptions) (*PackageContext, error) {
	// Resolve absolute path
	absPath, err := filepath.Abs(packageDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve package path: %w", err)
	}

	// Discover all .dingo files in package
	files, err := discoverDingoFiles(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to discover .dingo files: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no .dingo files found in %s", packageDir)
	}

	// Initialize cache
	cache := NewFunctionExclusionCache(absPath)

	// Try loading from disk (if incremental mode and not forced)
	cacheLoaded := false
	if opts.Incremental && !opts.Force {
		if err := cache.LoadFromDisk(); err == nil {
			// Cache loaded successfully, check if still valid
			if !cache.NeedsRescan(files) {
				// Cache is valid! Use it.
				cacheLoaded = true
				if opts.Verbose {
					metrics := cache.Metrics()
					fmt.Printf("Cache loaded (%d symbols)\n", metrics.TotalSymbols)
				}
			} else {
				if opts.Verbose {
					fmt.Println("Cache invalid (files changed), rescanning...")
				}
			}
		} else {
			if opts.Verbose {
				fmt.Printf("Cache not found or invalid: %v\n", err)
			}
		}
	}

	// Cache miss or invalid → Full rescan
	if !cacheLoaded {
		if err := cache.ScanPackage(files); err != nil {
			return nil, fmt.Errorf("failed to scan package: %w", err)
		}

		// Save cache for next build
		if err := cache.SaveToDisk(); err != nil {
			// Non-fatal: warn but continue
			if opts.Verbose {
				fmt.Printf("Warning: failed to save cache: %v\n", err)
			}
		} else if opts.Verbose {
			metrics := cache.Metrics()
			fmt.Printf("Cache saved (%d symbols, scan time: %v)\n", metrics.TotalSymbols, metrics.ScanDuration)
		}
	}

	return &PackageContext{
		packagePath: absPath,
		dingoFiles:  files,
		cache:       cache,
		incremental: opts.Incremental,
		force:       opts.Force,
		verbose:     opts.Verbose,
	}, nil
}

// discoverDingoFiles finds all .dingo files in the given directory
// Does NOT recurse into subdirectories (matches Go package model)
func discoverDingoFiles(packageDir string) ([]string, error) {
	entries, err := os.ReadDir(packageDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var dingoFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip subdirectories
		}

		name := entry.Name()
		if strings.HasSuffix(name, ".dingo") {
			absPath := filepath.Join(packageDir, name)
			dingoFiles = append(dingoFiles, absPath)
		}
	}

	return dingoFiles, nil
}

// TranspileAll transpiles all .dingo files in the package
func (ctx *PackageContext) TranspileAll() error {
	for _, file := range ctx.dingoFiles {
		if err := ctx.TranspileFile(file); err != nil {
			return fmt.Errorf("failed to transpile %s: %w", file, err)
		}
	}
	return nil
}

// TranspileFile transpiles a single .dingo file
func (ctx *PackageContext) TranspileFile(dingoFile string) error {
	// Read source
	source, err := os.ReadFile(dingoFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Create preprocessor with cache
	preprocessor := NewWithCache(source, ctx.cache)

	// Process
	goSource, sourceMap, err := preprocessor.Process()
	if err != nil {
		return fmt.Errorf("preprocessing failed: %w", err)
	}

	// Write .go file
	goFile := strings.TrimSuffix(dingoFile, ".dingo") + ".go"
	if err := os.WriteFile(goFile, []byte(goSource), 0644); err != nil {
		return fmt.Errorf("failed to write .go file: %w", err)
	}

	// Write .sourcemap file
	if sourceMap != nil {
		mapFile := goFile + ".map"
		mapJSON, err := sourceMap.ToJSON()
		if err != nil {
			return fmt.Errorf("failed to serialize source map: %w", err)
		}
		if err := os.WriteFile(mapFile, []byte(mapJSON), 0644); err != nil {
			return fmt.Errorf("failed to write source map: %w", err)
		}
	}

	if ctx.verbose {
		fmt.Printf("Transpiled: %s → %s\n", filepath.Base(dingoFile), filepath.Base(goFile))
	}

	return nil
}

// GetCache returns the function exclusion cache for this package
func (ctx *PackageContext) GetCache() *FunctionExclusionCache {
	return ctx.cache
}

// GetFiles returns the list of .dingo files in this package
func (ctx *PackageContext) GetFiles() []string {
	return ctx.dingoFiles
}

// PackagePath returns the absolute path to the package directory
func (ctx *PackageContext) PackagePath() string {
	return ctx.packagePath
}

// NewWithCache creates a preprocessor with a package-level cache
// This enables early bailout optimization and local function exclusion
func NewWithCache(source []byte, cache *FunctionExclusionCache) *Preprocessor {
	return newWithConfigAndCache(source, nil, cache)
}

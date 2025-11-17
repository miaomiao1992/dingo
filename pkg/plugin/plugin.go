// Package plugin provides the plugin system for code generation
package plugin

import (
	"go/ast"
	"go/token"
)

// Registry manages plugins
type Registry struct{}

// NewRegistry creates a new plugin registry
func NewRegistry() *Registry {
	return &Registry{}
}

// Pipeline executes plugins in sequence
type Pipeline struct {
	Ctx *Context
}

// NewPipeline creates a new plugin pipeline
func NewPipeline(registry *Registry, ctx *Context) (*Pipeline, error) {
	return &Pipeline{Ctx: ctx}, nil
}

// Transform transforms an AST (no-op for now)
func (p *Pipeline) Transform(file *ast.File) (*ast.File, error) {
	return file, nil
}

// GetStats returns pipeline stats
func (p *Pipeline) GetStats() Stats {
	return Stats{}
}

// SetTypeInferenceFactory sets the type inference factory (no-op)
func (p *Pipeline) SetTypeInferenceFactory(f interface{}) {}

// Stats for pipeline execution
type Stats struct {
	EnabledPlugins int
	TotalPlugins   int
}

// Context holds pipeline context
type Context struct {
	FileSet     *token.FileSet
	TypeInfo    interface{}
	Config      *Config
	Registry    *Registry
	Logger      Logger
	CurrentFile interface{}
}

// Config for code generation
type Config struct {
	EmitGeneratedMarkers bool
}

// Logger interface for plugin logging
type Logger interface {
	Info(msg string)
	Error(msg string)
	Debug(format string, args ...interface{})
	Warn(format string, args ...interface{})
}

// NoOpLogger does nothing
type NoOpLogger struct{}

// NewNoOpLogger creates a no-op logger
func NewNoOpLogger() Logger {
	return &NoOpLogger{}
}

func (n *NoOpLogger) Info(msg string)                            {}
func (n *NoOpLogger) Error(msg string)                           {}
func (n *NoOpLogger) Debug(format string, args ...interface{})   {}
func (n *NoOpLogger) Warn(format string, args ...interface{})    {}

// Plugin interface
type Plugin interface {
	Name() string
	Process(node ast.Node) error
}

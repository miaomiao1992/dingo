package plugin

import (
	"go/ast"
	"go/token"
	"testing"
)

// Mock plugin for testing
type MockPlugin struct {
	*BasePlugin
	transformCalled bool
}

func NewMockPlugin(name string, deps []string) *MockPlugin {
	return &MockPlugin{
		BasePlugin: NewBasePlugin(name, "Mock plugin for testing", deps),
	}
}

func (p *MockPlugin) Transform(ctx *Context, node ast.Node) (ast.Node, error) {
	p.transformCalled = true
	return node, nil
}

func TestRegistryRegister(t *testing.T) {
	registry := NewRegistry()
	plugin := NewMockPlugin("test-plugin", nil)

	err := registry.Register(plugin)
	if err != nil {
		t.Fatalf("Failed to register plugin: %v", err)
	}

	// Try to get the plugin
	retrieved, ok := registry.Get("test-plugin")
	if !ok {
		t.Fatal("Plugin not found after registration")
	}

	if retrieved.Name() != "test-plugin" {
		t.Errorf("Expected plugin name 'test-plugin', got %q", retrieved.Name())
	}
}

func TestRegistryDuplicateRegistration(t *testing.T) {
	registry := NewRegistry()
	plugin1 := NewMockPlugin("test-plugin", nil)
	plugin2 := NewMockPlugin("test-plugin", nil)

	if err := registry.Register(plugin1); err != nil {
		t.Fatalf("First registration failed: %v", err)
	}

	err := registry.Register(plugin2)
	if err == nil {
		t.Fatal("Expected error when registering duplicate plugin")
	}
}

func TestRegistrySortByDependencies(t *testing.T) {
	registry := NewRegistry()

	// Create plugins with dependencies:
	// A -> no deps
	// B -> depends on A
	// C -> depends on B
	pluginA := NewMockPlugin("plugin-a", nil)
	pluginB := NewMockPlugin("plugin-b", []string{"plugin-a"})
	pluginC := NewMockPlugin("plugin-c", []string{"plugin-b"})

	registry.Register(pluginC) // Register in random order
	registry.Register(pluginA)
	registry.Register(pluginB)

	if err := registry.SortByDependencies(); err != nil {
		t.Fatalf("Failed to sort dependencies: %v", err)
	}

	// Check order: should be A, B, C
	enabled := registry.Enabled()
	if len(enabled) != 3 {
		t.Fatalf("Expected 3 enabled plugins, got %d", len(enabled))
	}

	if enabled[0].Name() != "plugin-a" {
		t.Errorf("Expected first plugin to be 'plugin-a', got %q", enabled[0].Name())
	}
	if enabled[1].Name() != "plugin-b" {
		t.Errorf("Expected second plugin to be 'plugin-b', got %q", enabled[1].Name())
	}
	if enabled[2].Name() != "plugin-c" {
		t.Errorf("Expected third plugin to be 'plugin-c', got %q", enabled[2].Name())
	}
}

func TestRegistryCircularDependency(t *testing.T) {
	registry := NewRegistry()

	// Create circular dependency:
	// A -> depends on B
	// B -> depends on A
	pluginA := NewMockPlugin("plugin-a", []string{"plugin-b"})
	pluginB := NewMockPlugin("plugin-b", []string{"plugin-a"})

	registry.Register(pluginA)
	registry.Register(pluginB)

	err := registry.SortByDependencies()
	if err == nil {
		t.Fatal("Expected error for circular dependency")
	}
}

func TestRegistryEnableDisable(t *testing.T) {
	registry := NewRegistry()
	plugin := NewMockPlugin("test-plugin", nil)

	registry.Register(plugin)

	// Plugin should be enabled by default
	if !plugin.Enabled() {
		t.Error("Plugin should be enabled by default")
	}

	// Disable plugin
	if err := registry.DisablePlugin("test-plugin"); err != nil {
		t.Fatalf("Failed to disable plugin: %v", err)
	}

	if plugin.Enabled() {
		t.Error("Plugin should be disabled")
	}

	// Enable plugin
	if err := registry.EnablePlugin("test-plugin"); err != nil {
		t.Fatalf("Failed to enable plugin: %v", err)
	}

	if !plugin.Enabled() {
		t.Error("Plugin should be enabled")
	}
}

func TestRegistryEnableDependencies(t *testing.T) {
	registry := NewRegistry()

	// Create plugins:
	// A -> no deps
	// B -> depends on A
	pluginA := NewMockPlugin("plugin-a", nil)
	pluginB := NewMockPlugin("plugin-b", []string{"plugin-a"})

	pluginA.SetEnabled(false)
	pluginB.SetEnabled(false)

	registry.Register(pluginA)
	registry.Register(pluginB)

	// Enable B should also enable A
	if err := registry.EnablePlugin("plugin-b"); err != nil {
		t.Fatalf("Failed to enable plugin-b: %v", err)
	}

	if !pluginA.Enabled() {
		t.Error("plugin-a should be enabled as dependency")
	}
	if !pluginB.Enabled() {
		t.Error("plugin-b should be enabled")
	}
}

func TestPipelineTransform(t *testing.T) {
	registry := NewRegistry()
	plugin1 := NewMockPlugin("plugin-1", nil)
	plugin2 := NewMockPlugin("plugin-2", nil)

	registry.Register(plugin1)
	registry.Register(plugin2)

	ctx := &Context{
		FileSet: token.NewFileSet(),
		Config:  &Config{},
		Registry: registry,
		Logger:   NewNoOpLogger(),
	}

	pipeline, err := NewPipeline(registry, ctx)
	if err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}

	// Create a simple AST file
	file := &ast.File{
		Name: &ast.Ident{Name: "main"},
	}

	// Transform the file
	_, err = pipeline.Transform(file)
	if err != nil {
		t.Fatalf("Pipeline transform failed: %v", err)
	}

	// Verify both plugins were called
	if !plugin1.transformCalled {
		t.Error("plugin-1 Transform was not called")
	}
	if !plugin2.transformCalled {
		t.Error("plugin-2 Transform was not called")
	}
}

func TestPipelineStats(t *testing.T) {
	registry := NewRegistry()
	plugin1 := NewMockPlugin("plugin-1", nil)
	plugin2 := NewMockPlugin("plugin-2", nil)
	plugin2.SetEnabled(false)

	registry.Register(plugin1)
	registry.Register(plugin2)

	ctx := &Context{
		FileSet:  token.NewFileSet(),
		Config:   &Config{},
		Registry: registry,
		Logger:   NewNoOpLogger(),
	}

	pipeline, err := NewPipeline(registry, ctx)
	if err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}

	stats := pipeline.GetStats()

	if stats.TotalPlugins != 2 {
		t.Errorf("Expected 2 total plugins, got %d", stats.TotalPlugins)
	}

	if stats.EnabledPlugins != 1 {
		t.Errorf("Expected 1 enabled plugin, got %d", stats.EnabledPlugins)
	}

	if len(stats.PluginNames) != 2 {
		t.Errorf("Expected 2 plugin names, got %d", len(stats.PluginNames))
	}

	if len(stats.ExecutionOrder) != 1 {
		t.Errorf("Expected 1 plugin in execution order, got %d", len(stats.ExecutionOrder))
	}

	if stats.ExecutionOrder[0] != "plugin-1" {
		t.Errorf("Expected 'plugin-1' in execution order, got %q", stats.ExecutionOrder[0])
	}
}

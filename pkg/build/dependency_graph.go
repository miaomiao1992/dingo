package build

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// DependencyGraph represents package dependencies
type DependencyGraph struct {
	Nodes map[string]*GraphNode // Package path -> node
}

// GraphNode represents a package in the dependency graph
type GraphNode struct {
	Path         string   // Package path
	Dependencies []string // Packages this one depends on
	Dependents   []string // Packages that depend on this one
}

// buildDependencyGraph creates a dependency graph from packages
func buildDependencyGraph(packages []Package, workspaceRoot string) (*DependencyGraph, error) {
	graph := &DependencyGraph{
		Nodes: make(map[string]*GraphNode),
	}

	// Create nodes for all packages
	for _, pkg := range packages {
		graph.Nodes[pkg.Path] = &GraphNode{
			Path:         pkg.Path,
			Dependencies: make([]string, 0),
			Dependents:   make([]string, 0),
		}
	}

	// Extract dependencies from import statements
	for _, pkg := range packages {
		deps, err := extractDependencies(pkg, workspaceRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to extract dependencies for %s: %w", pkg.Path, err)
		}

		node := graph.Nodes[pkg.Path]
		for _, dep := range deps {
			// Only track internal workspace dependencies
			if depNode, exists := graph.Nodes[dep]; exists {
				node.Dependencies = append(node.Dependencies, dep)
				depNode.Dependents = append(depNode.Dependents, pkg.Path)
			}
		}
	}

	return graph, nil
}

// extractDependencies extracts import statements from .dingo files
func extractDependencies(pkg Package, workspaceRoot string) ([]string, error) {
	deps := make(map[string]bool) // Use map to avoid duplicates

	// Regex to match import statements
	importRegex := regexp.MustCompile(`import\s+(?:"([^"]+)"|([^\s]+))`)

	for _, dingoFile := range pkg.DingoFiles {
		fullPath := filepath.Join(workspaceRoot, dingoFile)
		file, err := os.Open(fullPath)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			matches := importRegex.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				importPath := match[1]
				if importPath == "" {
					importPath = match[2]
				}

				// Convert import path to workspace-relative package path
				pkgPath := importPathToPackagePath(importPath, workspaceRoot)
				if pkgPath != "" {
					deps[pkgPath] = true
				}
			}
		}

		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}

	// Convert map to slice
	result := make([]string, 0, len(deps))
	for dep := range deps {
		result = append(result, dep)
	}

	return result, nil
}

// importPathToPackagePath converts an import path to a workspace-relative package path
func importPathToPackagePath(importPath, workspaceRoot string) string {
	// Handle relative imports (./foo, ../bar)
	if strings.HasPrefix(importPath, ".") {
		return strings.TrimPrefix(importPath, "./")
	}

	// Read module path from go.mod
	modPath, err := getModulePath(workspaceRoot)
	if err != nil {
		return "" // External dependency (not in workspace)
	}

	// Check if import is within workspace module
	if strings.HasPrefix(importPath, modPath) {
		relPath := strings.TrimPrefix(importPath, modPath+"/")
		return relPath
	}

	// External dependency, not tracked
	return ""
}

// getModulePath extracts the module path from go.mod
func getModulePath(root string) (string, error) {
	goMod := filepath.Join(root, "go.mod")
	data, err := os.ReadFile(goMod)
	if err != nil {
		return "", err
	}

	// Parse "module github.com/user/repo" line
	moduleRegex := regexp.MustCompile(`module\s+([^\s]+)`)
	match := moduleRegex.FindSubmatch(data)
	if match == nil {
		return "", fmt.Errorf("no module declaration found in go.mod")
	}

	return string(match[1]), nil
}

// detectCircularDependencies finds circular dependency chains.
// Returns slice of cycles, where each cycle is the full import chain.
func detectCircularDependencies(graph *DependencyGraph) [][]string {
	cycles := make([][]string, 0)

	// Track visited nodes and current path
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := make([]string, 0)

	// DFS to detect cycles
	var detectCycle func(node string) bool
	detectCycle = func(node string) bool {
		visited[node] = true
		recStack[node] = true
		path = append(path, node)

		if graphNode, exists := graph.Nodes[node]; exists {
			for _, dep := range graphNode.Dependencies {
				if !visited[dep] {
					if detectCycle(dep) {
						return true
					}
				} else if recStack[dep] {
					// Found cycle - build full cycle path
					cycleStart := 0
					for i, p := range path {
						if p == dep {
							cycleStart = i
							break
						}
					}
					// Include the closing edge back to start
					cycle := make([]string, len(path)-cycleStart+1)
					copy(cycle, path[cycleStart:])
					cycle[len(cycle)-1] = dep // Complete the cycle
					cycles = append(cycles, cycle)
					return true
				}
			}
		}

		path = path[:len(path)-1]
		recStack[node] = false
		return false
	}

	// Check all nodes
	for node := range graph.Nodes {
		if !visited[node] {
			detectCycle(node)
		}
	}

	return cycles
}

// topologicalSort returns packages in build order (dependencies first)
func topologicalSort(graph *DependencyGraph) []string {
	// Kahn's algorithm
	inDegree := make(map[string]int)
	for _, node := range graph.Nodes {
		inDegree[node.Path] = 0
	}
	for _, node := range graph.Nodes {
		for _, dep := range node.Dependencies {
			inDegree[dep]++
		}
	}

	// Queue of nodes with no dependencies
	queue := make([]string, 0)
	for node, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, node)
		}
	}

	result := make([]string, 0, len(graph.Nodes))
	for len(queue) > 0 {
		// Dequeue
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// Reduce in-degree of dependents
		if node, exists := graph.Nodes[current]; exists {
			for _, dep := range node.Dependents {
				inDegree[dep]--
				if inDegree[dep] == 0 {
					queue = append(queue, dep)
				}
			}
		}
	}

	// If result doesn't contain all nodes, there's a cycle
	if len(result) != len(graph.Nodes) {
		// Return partial order (best effort)
		for node := range graph.Nodes {
			found := false
			for _, r := range result {
				if r == node {
					found = true
					break
				}
			}
			if !found {
				result = append(result, node)
			}
		}
	}

	return result
}

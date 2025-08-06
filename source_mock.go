package echotemplates

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

// MockSource implements TemplateSource for testing purposes using an in-memory map
type MockSource struct {
	templates map[string]string
}

// NewMockSource creates a new mock template source with the given templates
func NewMockSource(templates map[string]string) *MockSource {
	// Create a copy to avoid external modifications
	templatesCopy := make(map[string]string)
	for k, v := range templates {
		templatesCopy[k] = v
	}

	return &MockSource{
		templates: templatesCopy,
	}
}

// Open returns a reader for the template content
func (m *MockSource) Open(path string) (io.ReadCloser, error) {
	content, exists := m.templates[path]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", path)
	}

	return io.NopCloser(bytes.NewReader([]byte(content))), nil
}

// Stat returns information about a template
func (m *MockSource) Stat(path string) (TemplateInfo, error) {
	content, exists := m.templates[path]
	if !exists {
		return TemplateInfo{}, fmt.Errorf("template not found: %s", path)
	}

	return TemplateInfo{
		Path:    path,
		ModTime: time.Now(),
		Size:    int64(len(content)),
		IsDir:   false,
	}, nil
}

// List returns all available template paths
func (m *MockSource) List() ([]string, error) {
	var paths []string
	for path := range m.templates {
		// Only include .md files to match FileSystemSource behavior
		if strings.HasSuffix(path, ".md") {
			paths = append(paths, path)
		}
	}

	sort.Strings(paths)
	return paths, nil
}

// Watch returns nil channel - watching not supported for mock
func (m *MockSource) Watch() (<-chan string, error) {
	return nil, nil
}

// StopWatch is a no-op for mock
func (m *MockSource) StopWatch() error {
	return nil
}

// ResolveImport returns empty string - no custom import resolution
func (m *MockSource) ResolveImport(importPath, currentPath string) string {
	return ""
}

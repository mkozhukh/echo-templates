package echotemplates

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/mkozhukh/echo"
)

// stringSource implements TemplateSource for in-memory string templates
// It treats the path parameter as the actual template content
type stringSource struct{}

// Open returns a reader for the template content (path is the content)
func (s *stringSource) Open(path string) (io.ReadCloser, error) {
	// The path IS the template content for string source
	return io.NopCloser(strings.NewReader(path)), nil
}

// Stat returns information about the template
func (s *stringSource) Stat(path string) (TemplateInfo, error) {
	return TemplateInfo{
		Path:    path,
		ModTime: time.Now(),
		Size:    int64(len(path)),
		IsDir:   false,
	}, nil
}

// List returns an empty list as string source has no files
func (s *stringSource) List() ([]string, error) {
	return []string{}, nil
}

// Watch returns nil as string templates don't change
func (s *stringSource) Watch() (<-chan string, error) {
	return nil, nil
}

// StopWatch is a no-op for string templates
func (s *stringSource) StopWatch() error {
	return nil
}

// ResolveImport returns empty string as imports are not supported
func (s *stringSource) ResolveImport(importPath, currentPath string) string {
	return ""
}

// Singleton string engine for package-level Generate functions
var (
	stringEngine     TemplateEngine
	stringEngineOnce sync.Once
	stringEngineErr  error
)

// getStringEngine returns the singleton string engine
func getStringEngine() (TemplateEngine, error) {
	stringEngineOnce.Do(func() {
		stringEngine, stringEngineErr = New(Config{
			Source:  &stringSource{},
			DevMode: true, // No caching for string templates
		})
	})
	return stringEngine, stringEngineErr
}

// Generate creates messages from a string template
func Generate(content string, vars map[string]any, opts ...GenerateOptions) ([]echo.Message, error) {
	engine, err := getStringEngine()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize string engine: %w", err)
	}

	// For string source, we pass the content as the "template name"
	// The engine will call source.Open(content) which returns the content
	return engine.Generate(content, vars, opts...)
}

// GenerateWithMetadata creates messages from a string template and returns metadata
func GenerateWithMetadata(content string, vars map[string]any, opts ...GenerateOptions) ([]echo.Message, map[string]any, error) {
	engine, err := getStringEngine()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize string engine: %w", err)
	}

	// For string source, we pass the content as the "template name"
	// The engine will call source.Open(content) which returns the content
	return engine.GenerateWithMetadata(content, vars, opts...)
}

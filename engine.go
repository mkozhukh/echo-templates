package echotemplates

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mkozhukh/echo"
)

// templateEngine is the main implementation of TemplateEngine
type templateEngine struct {
	config Config
	cache  *templateCache
}

// New creates a new template engine
func New(config Config) (TemplateEngine, error) {
	// Validate config
	if config.RootDir == "" {
		return nil, fmt.Errorf("RootDir is required")
	}

	// Check if root directory exists
	info, err := os.Stat(config.RootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to access RootDir: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("RootDir is not a directory: %s", config.RootDir)
	}

	// Set defaults
	if config.CacheSize == 0 {
		config.CacheSize = 100
	}
	if !config.EnableCache {
		config.EnableCache = true
	}

	engine := &templateEngine{
		config: config,
	}

	if config.EnableCache {
		engine.cache = newTemplateCache(config.CacheSize)
	}

	return engine, nil
}

// Generate creates messages from a template
func (e *templateEngine) Generate(name string, vars map[string]string, opts ...GenerateOptions) ([]echo.Message, error) {
	options := e.config.DefaultOptions
	if len(opts) > 0 {
		options = opts[0]
	}
	messages, _, err := e.generateInternal(name, vars, options)
	return messages, err
}

// GenerateWithMetadata creates messages and returns template metadata
func (e *templateEngine) GenerateWithMetadata(name string, vars map[string]string, opts ...GenerateOptions) ([]echo.Message, map[string]any, error) {
	options := e.config.DefaultOptions
	if len(opts) > 0 {
		options = opts[0]
	}
	return e.generateInternal(name, vars, options)
}

// ClearCache removes cached templates
func (e *templateEngine) ClearCache() {
	if e.cache != nil {
		e.cache.clear()
	}
}

// generateInternal is the core generation logic
func (e *templateEngine) generateInternal(name string, vars map[string]string, opts GenerateOptions) ([]echo.Message, map[string]any, error) {
	// Ensure .md extension
	if !strings.HasSuffix(name, ".md") {
		name = name + ".md"
	}

	// Resolve the full path
	fullPath := filepath.Join(e.config.RootDir, name)

	// Load and parse the template
	template, err := e.loadTemplate(fullPath, opts)
	if err != nil {
		return nil, nil, err
	}

	// Process imports recursively
	content, err := e.processImports(template.content, vars, opts, name)
	if err != nil {
		return nil, nil, err
	}

	// Merge defaults with provided vars
	mergedVars := make(map[string]string)
	defaults := make(map[string]any)
	if d, ok := template.metadata["defaults"]; ok {
		if defaultsMap, ok := d.(map[string]any); ok {
			defaults = defaultsMap
			for k, v := range defaults {
				mergedVars[k] = toString(v)
			}
		}
	}
	for k, v := range vars {
		mergedVars[k] = v
	}

	// Substitute variables
	content, err = substituteVariables(content, mergedVars, defaults, opts)
	if err != nil {
		return nil, nil, err
	}

	// Parse into messages
	messages := echo.TemplateMessage(content)

	return messages, template.metadata, nil
}

// loadTemplate loads and parses a template file
func (e *templateEngine) loadTemplate(path string, opts GenerateOptions) (*parsedTemplate, error) {
	// Get file info for cache checking
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &TemplateNotFoundError{
				Name: filepath.Base(path),
				Path: path,
			}
		}
		return nil, fmt.Errorf("failed to stat template file: %w", err)
	}

	// Check cache if enabled
	if e.cache != nil && !opts.DisableCache {
		if cached, ok := e.cache.get(path, info.ModTime()); ok {
			return cached, nil
		}
	}

	// Read the file
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open template file: %w", err)
	}
	defer file.Close()

	// Parse front-matter and content
	metadata, content, err := parseFrontMatter(file)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	// Extract imports
	imports := extractImports(content)

	template := &parsedTemplate{
		metadata: metadata,
		content:  content,
		imports:  imports,
	}

	// Cache the parsed template
	if e.cache != nil && !opts.DisableCache {
		e.cache.put(path, template, info.ModTime())
	}

	return template, nil
}

// processImports recursively processes import placeholders
func (e *templateEngine) processImports(content string, vars map[string]string, opts GenerateOptions, currentTemplate string) (string, error) {
	// Keep track of processed imports to avoid infinite recursion
	processed := make(map[string]bool)

	return e.processImportsRecursive(content, vars, opts, currentTemplate, processed)
}

// processImportsRecursive handles the actual recursive import processing
func (e *templateEngine) processImportsRecursive(content string, vars map[string]string, opts GenerateOptions, currentTemplate string, processed map[string]bool) (string, error) {
	// Process imports using the extractImports function which handles nested placeholders
	imports := extractImports(content)

	for _, importPath := range imports {
		fullMatch := "{{@" + importPath + "}}"

		// Handle dynamic imports (e.g., {{@{{template_type}}/header}})
		importPath = placeholderRegex.ReplaceAllStringFunc(importPath, func(innerMatch string) string {
			varName := strings.TrimSpace(innerMatch[2 : len(innerMatch)-2])
			if value, ok := vars[varName]; ok {
				return value
			}
			return innerMatch
		})

		// Ensure .md extension
		if !strings.HasSuffix(importPath, ".md") {
			importPath = importPath + ".md"
		}

		// Resolve the full path
		fullImportPath := filepath.Join(e.config.RootDir, importPath)

		// Check for circular imports
		if processed[fullImportPath] {
			if opts.StrictMode {
				return "", &ImportError{
					ImportPath: importPath,
					Template:   currentTemplate,
					Cause:      fmt.Errorf("circular import detected"),
				}
			}
			// In non-strict mode, just skip the import
			content = strings.ReplaceAll(content, fullMatch, "")
			continue
		}

		// Mark as processed
		processed[fullImportPath] = true

		// Load the imported template
		importedTemplate, err := e.loadTemplate(fullImportPath, opts)
		if err != nil {
			if opts.StrictMode {
				return "", &ImportError{
					ImportPath: importPath,
					Template:   currentTemplate,
					Cause:      err,
				}
			}
			// In non-strict mode, keep the placeholder
			continue
		}

		// Process imports in the imported content recursively
		importedContent, err := e.processImportsRecursive(importedTemplate.content, vars, opts, importPath, processed)
		if err != nil {
			return "", err
		}

		// Replace the import placeholder with the imported content
		content = strings.ReplaceAll(content, fullMatch, importedContent)
	}

	return content, nil
}

// toString converts any value to string representation
func toString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return strconv.Itoa(val)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	default:
		return ""
	}
}

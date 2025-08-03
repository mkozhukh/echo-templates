package echotemplates

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/mkozhukh/echo"
)

// templateEngine is the main implementation of TemplateEngine
type templateEngine struct {
	config    Config
	source    TemplateSource
	cache     *templateCache
	watchChan <-chan string
	devMode   bool
}

// New creates a new template engine
func New(config Config) (TemplateEngine, error) {
	// Validate config
	if config.Source == nil {
		return nil, fmt.Errorf("config.Source is required")
	}

	// Set defaults
	if config.CacheSize == 0 {
		config.CacheSize = 100
	}

	engine := &templateEngine{
		config:  config,
		source:  config.Source,
		devMode: config.DevMode,
	}

	// Initialize cache in production mode
	if !config.DevMode {
		engine.cache = newTemplateCache(config.CacheSize)
	}

	// Start file watching in dev mode
	if config.DevMode {
		watchChan, err := config.Source.Watch()
		if err == nil && watchChan != nil {
			engine.watchChan = watchChan
			go engine.handleFileChanges()
		}
	}

	return engine, nil
}

// handleFileChanges monitors file changes in dev mode
func (e *templateEngine) handleFileChanges() {
	for range e.watchChan {
		// Clear entire cache in dev mode when any file changes
		// This ensures imports are also refreshed
		e.ClearCache()
	}
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
	// Ensure .md extension (except for stringSource where name is the content)
	if _, isStringSource := e.source.(*stringSource); !isStringSource && !strings.HasSuffix(name, ".md") {
		name = name + ".md"
	}

	// Load and parse the template
	template, err := e.loadTemplate(name, opts)
	if err != nil {
		return nil, nil, err
	}

	// Check if we're using stringSource and have imports
	if _, isStringSource := e.source.(*stringSource); isStringSource && len(template.imports) > 0 {
		return nil, nil, fmt.Errorf("imports are not supported in string templates")
	}

	// Process imports recursively
	content, err := e.processImports(template.content, vars, opts, name)
	if err != nil {
		return nil, nil, err
	}

	// Merge defaults with provided vars
	mergedVars := make(map[string]string)
	if d, ok := template.metadata["defaults"]; ok {
		if defaultsMap, ok := d.(map[string]string); ok {
			for k, v := range defaultsMap {
				mergedVars[k] = v
			}
		}
	}
	for k, v := range vars {
		mergedVars[k] = v
	}

	// Substitute variables
	content, err = substituteVariables(content, mergedVars, nil, opts)
	if err != nil {
		return nil, nil, err
	}

	// Parse into messages
	messages := echo.TemplateMessage(content)

	// If no messages were parsed (no role markers), create a single user message
	// This is useful for simple string templates
	if len(messages) == 0 && content != "" {
		messages = []echo.Message{
			{Role: "user", Content: content},
		}
	}

	return messages, template.metadata, nil
}

// loadTemplate loads and parses a template file
func (e *templateEngine) loadTemplate(path string, opts GenerateOptions) (*parsedTemplate, error) {
	// Get file info for cache checking
	info, err := e.source.Stat(path)
	if err != nil {
		return nil, &TemplateNotFoundError{
			Name: strings.TrimSuffix(path, ".md"),
			Path: path,
		}
	}

	// Check cache if enabled (skip in dev mode or if DisableCache is set)
	if e.cache != nil && !e.devMode && !opts.DisableCache {
		if cached, ok := e.cache.get(path, info.ModTime); ok {
			return cached, nil
		}
	}

	// Read the file
	file, err := e.source.Open(path)
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

	// Cache the parsed template (skip in dev mode)
	if e.cache != nil && !e.devMode && !opts.DisableCache {
		e.cache.put(path, template, info.ModTime)
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

		// Allow source to customize import resolution
		if customPath := e.source.ResolveImport(importPath, currentTemplate); customPath != "" {
			importPath = customPath
		}

		// Check for circular imports
		if processed[importPath] {
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
		processed[importPath] = true

		// Load the imported template
		importedTemplate, err := e.loadTemplate(importPath, opts)
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

// ValidateTemplate checks if a template is valid without generating messages
func (e *templateEngine) ValidateTemplate(name string) error {
	// Ensure .md extension
	if !strings.HasSuffix(name, ".md") {
		name = name + ".md"
	}

	// Try to load and parse the template
	_, err := e.loadTemplate(name, e.config.DefaultOptions)
	if err != nil {
		return err
	}

	// Check for circular imports by processing imports with empty vars
	template, _ := e.loadTemplate(name, e.config.DefaultOptions)
	_, err = e.processImports(template.content, make(map[string]string), e.config.DefaultOptions, name)
	return err
}

// GetTemplateVariables returns all variable names used in a template
func (e *templateEngine) GetTemplateVariables(name string) ([]string, error) {
	// Ensure .md extension
	if !strings.HasSuffix(name, ".md") {
		name = name + ".md"
	}

	// Load the template
	template, err := e.loadTemplate(name, e.config.DefaultOptions)
	if err != nil {
		return nil, err
	}

	// Process imports to get full content
	content, err := e.processImports(template.content, make(map[string]string), e.config.DefaultOptions, name)
	if err != nil {
		return nil, err
	}

	// Extract all variables
	variableMap := make(map[string]bool)

	// First, remove triple brace placeholders to avoid double matching
	contentWithoutRaw := rawPlaceholderRegex.ReplaceAllString(content, "")

	// Find variables in double braces (excluding imports)
	matches := placeholderRegex.FindAllStringSubmatch(contentWithoutRaw, -1)
	for _, match := range matches {
		if len(match) > 1 && !strings.HasPrefix(match[0], "{{@") {
			inner := strings.TrimSpace(match[1])
			// Handle default value syntax
			parts := strings.SplitN(inner, "|", 2)
			varName := strings.TrimSpace(parts[0])
			variableMap[varName] = true
		}
	}

	// Find variables in triple braces from original content
	rawMatches := rawPlaceholderRegex.FindAllStringSubmatch(content, -1)
	for _, match := range rawMatches {
		if len(match) > 1 {
			varName := strings.TrimSpace(match[1])
			variableMap[varName] = true
		}
	}

	// Convert map to sorted slice
	var variables []string
	for v := range variableMap {
		variables = append(variables, v)
	}
	sort.Strings(variables)

	return variables, nil
}

// TemplateExists checks if a template file exists
func (e *templateEngine) TemplateExists(name string) bool {
	// Ensure .md extension
	if !strings.HasSuffix(name, ".md") {
		name = name + ".md"
	}

	// Check if file exists
	info, err := e.source.Stat(name)
	return err == nil && !info.IsDir
}

// ListTemplates returns all available template paths relative to source root
func (e *templateEngine) ListTemplates() ([]string, error) {
	templates, err := e.source.List()
	if err != nil {
		return nil, err
	}

	// Remove .md extension for consistency with other methods
	for i, template := range templates {
		templates[i] = strings.TrimSuffix(template, ".md")
	}

	return templates, nil
}

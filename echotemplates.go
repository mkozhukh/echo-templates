package echotemplates

import (
	"github.com/mkozhukh/echo"
)

// TemplateEngine manages template loading and processing
type TemplateEngine interface {
	// Generate creates messages from a template
	// If name doesn't contain .md suffix, it will be added automatically
	Generate(name string, vars map[string]string, opts ...GenerateOptions) ([]echo.Message, error)

	// GenerateWithMetadata creates messages and returns template metadata
	GenerateWithMetadata(name string, vars map[string]string, opts ...GenerateOptions) ([]echo.Message, map[string]any, error)

	// ClearCache removes cached templates (useful for development)
	ClearCache()
}

// GenerateOptions configures template generation behavior
type GenerateOptions struct {
	// AllowMissingVars determines if missing placeholders cause errors
	AllowMissingVars bool

	// StrictMode enables strict parsing (no undefined imports, etc)
	StrictMode bool

	// DisableCache bypasses cache for this generation
	DisableCache bool
}

// Config configures the template engine
type Config struct {
	// RootDir is the base directory for template files
	RootDir string

	// DefaultOptions applies to all Generate calls unless overridden
	DefaultOptions GenerateOptions

	// EnableCache enables in-memory template caching (default: true)
	EnableCache bool

	// CacheSize maximum number of templates to cache (default: 100)
	CacheSize int
}

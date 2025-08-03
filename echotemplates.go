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

	// ValidateTemplate checks if a template is valid without generating messages
	ValidateTemplate(name string) error

	// GetTemplateVariables returns all variable names used in a template
	GetTemplateVariables(name string) ([]string, error)

	// TemplateExists checks if a template file exists
	TemplateExists(name string) bool

	// ListTemplates returns all available template paths relative to RootDir
	ListTemplates() ([]string, error)
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
	// Source is the template source (required)
	Source TemplateSource

	// DevMode disables caching for development (default: false)
	DevMode bool

	// DefaultOptions applies to all Generate calls unless overridden
	DefaultOptions GenerateOptions

	// CacheSize maximum number of templates to cache in production mode (default: 100)
	CacheSize int
}

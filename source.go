package echotemplates

import (
	"io"
	"time"
)

// TemplateSource abstracts the source of templates (filesystem, embedded, etc.)
type TemplateSource interface {
	// Open returns a reader for the template content
	Open(path string) (io.ReadCloser, error)

	// Stat returns information about a template
	Stat(path string) (TemplateInfo, error)

	// List returns all available template paths
	List() ([]string, error)

	// Watch starts watching for changes if supported
	// The returned channel will receive paths of changed templates
	// Returns nil if watching is not supported
	Watch() (<-chan string, error)

	// StopWatch stops watching for changes
	StopWatch() error

	// ResolveImport allows customizing import resolution
	// Given an import path and the current template path, returns the resolved path
	// Return empty string to use default resolution
	ResolveImport(importPath, currentPath string) string
}

// TemplateInfo contains information about a template
type TemplateInfo struct {
	// Path is the template path
	Path string

	// ModTime is the modification time
	ModTime time.Time

	// Size is the template size in bytes
	Size int64

	// IsDir indicates if this is a directory
	IsDir bool
}

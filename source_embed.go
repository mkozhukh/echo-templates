package echotemplates

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

// EmbedSource implements TemplateSource for embedded templates
type EmbedSource struct {
	fs      embed.FS
	rootDir string
}

// NewEmbedSource creates a new embedded template source
func NewEmbedSource(embedFS embed.FS, rootDir string) *EmbedSource {
	// Normalize root directory
	rootDir = strings.TrimPrefix(rootDir, "/")
	rootDir = strings.TrimSuffix(rootDir, "/")

	return &EmbedSource{
		fs:      embedFS,
		rootDir: rootDir,
	}
}

// Open returns a reader for the template content
func (s *EmbedSource) Open(path string) (io.ReadCloser, error) {
	fullPath := path
	if s.rootDir != "" {
		fullPath = filepath.Join(s.rootDir, path)
	}

	data, err := s.fs.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}

	return io.NopCloser(bytes.NewReader(data)), nil
}

// Stat returns information about a template
func (s *EmbedSource) Stat(path string) (TemplateInfo, error) {
	fullPath := path
	if s.rootDir != "" {
		fullPath = filepath.Join(s.rootDir, path)
	}

	// Try to open the file to check if it exists
	file, err := s.fs.Open(fullPath)
	if err != nil {
		return TemplateInfo{}, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return TemplateInfo{}, err
	}

	return TemplateInfo{
		Path:    path,
		ModTime: info.ModTime(),
		Size:    info.Size(),
		IsDir:   info.IsDir(),
	}, nil
}

// List returns all available template paths
func (s *EmbedSource) List() ([]string, error) {
	var templates []string

	rootToWalk := "."
	if s.rootDir != "" {
		rootToWalk = s.rootDir
	}

	err := fs.WalkDir(s.fs, rootToWalk, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Only include .md files
		if strings.HasSuffix(path, ".md") {
			// Get relative path from root
			relPath := path
			if s.rootDir != "" {
				relPath = strings.TrimPrefix(path, s.rootDir+"/")
			}
			templates = append(templates, relPath)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk embedded filesystem: %w", err)
	}

	sort.Strings(templates)
	return templates, nil
}

// Watch returns nil as embedded templates don't change
func (s *EmbedSource) Watch() (<-chan string, error) {
	// Embedded templates don't change at runtime
	return nil, nil
}

// StopWatch is a no-op for embedded templates
func (s *EmbedSource) StopWatch() error {
	return nil
}

// ResolveImport allows customizing import resolution
func (s *EmbedSource) ResolveImport(importPath, currentPath string) string {
	// Default resolution - no custom behavior
	return ""
}

// embedReadCloser wraps embed.FS file to implement io.ReadCloser
type embedReadCloser struct {
	*bytes.Reader
}

func (e *embedReadCloser) Close() error {
	return nil
}

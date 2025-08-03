package echotemplates

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// FileSystemSource implements TemplateSource for filesystem-based templates
type FileSystemSource struct {
	rootDir    string
	watchChan  chan string
	stopWatch  chan struct{}
	watchErr   error
	watchMutex sync.Mutex
	watching   bool
}

// NewFileSystemSource creates a new filesystem template source
func NewFileSystemSource(rootDir string) (*FileSystemSource, error) {
	// Validate root directory
	info, err := os.Stat(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to access root directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("root path is not a directory: %s", rootDir)
	}

	absPath, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	return &FileSystemSource{
		rootDir: absPath,
	}, nil
}

// Open returns a reader for the template content
func (s *FileSystemSource) Open(path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.rootDir, path)
	return os.Open(fullPath)
}

// Stat returns information about a template
func (s *FileSystemSource) Stat(path string) (TemplateInfo, error) {
	fullPath := filepath.Join(s.rootDir, path)
	info, err := os.Stat(fullPath)
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
func (s *FileSystemSource) List() ([]string, error) {
	var templates []string

	err := filepath.WalkDir(s.rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Only include .md files
		if strings.HasSuffix(path, ".md") {
			// Get relative path
			relPath, err := filepath.Rel(s.rootDir, path)
			if err != nil {
				return err
			}
			templates = append(templates, relPath)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.Strings(templates)
	return templates, nil
}

// Watch starts watching for changes
func (s *FileSystemSource) Watch() (<-chan string, error) {
	s.watchMutex.Lock()
	defer s.watchMutex.Unlock()

	if s.watching {
		return s.watchChan, nil
	}

	// For now, we'll use a polling approach
	// In production, you might want to use fsnotify or similar
	s.watchChan = make(chan string, 100)
	s.stopWatch = make(chan struct{})
	s.watching = true

	// Start polling goroutine
	go s.pollChanges()

	return s.watchChan, nil
}

// StopWatch stops watching for changes
func (s *FileSystemSource) StopWatch() error {
	s.watchMutex.Lock()
	defer s.watchMutex.Unlock()

	if !s.watching {
		return nil
	}

	close(s.stopWatch)
	s.watching = false
	close(s.watchChan)

	return nil
}

// ResolveImport allows customizing import resolution
func (s *FileSystemSource) ResolveImport(importPath, currentPath string) string {
	// Default resolution - no custom behavior
	return ""
}

// pollChanges polls for file changes (simple implementation)
func (s *FileSystemSource) pollChanges() {
	// Keep track of file modification times
	modTimes := make(map[string]time.Time)

	// Initial scan
	templates, _ := s.List()
	for _, path := range templates {
		if info, err := s.Stat(path); err == nil {
			modTimes[path] = info.ModTime
		}
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopWatch:
			return
		case <-ticker.C:
			// Check for changes
			templates, _ := s.List()
			for _, path := range templates {
				if info, err := s.Stat(path); err == nil {
					if lastMod, exists := modTimes[path]; !exists || info.ModTime.After(lastMod) {
						// File was added or modified
						select {
						case s.watchChan <- path:
						default:
							// Channel full, skip
						}
						modTimes[path] = info.ModTime
					}
				}
			}

			// Check for deletions
			for path := range modTimes {
				found := false
				for _, t := range templates {
					if t == path {
						found = true
						break
					}
				}
				if !found {
					// File was deleted
					delete(modTimes, path)
					select {
					case s.watchChan <- path:
					default:
						// Channel full, skip
					}
				}
			}
		}
	}
}

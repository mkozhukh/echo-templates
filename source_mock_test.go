package echotemplates

import (
	"io"
	"testing"
)

func TestMockSource(t *testing.T) {
	// Test creating a MockSource with initial templates
	templates := map[string]string{
		"template1.md": "# Template 1\nContent of template 1",
		"template2.md": "# Template 2\nContent of template 2",
		"other.txt":    "Not a markdown file",
	}

	mock := NewMockSource(templates)

	// Test that it implements TemplateSource interface
	var _ TemplateSource = mock

	// Test Open
	t.Run("Open", func(t *testing.T) {
		reader, err := mock.Open("template1.md")
		if err != nil {
			t.Fatalf("Failed to open template: %v", err)
		}
		defer reader.Close()

		content, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("Failed to read content: %v", err)
		}

		expected := "# Template 1\nContent of template 1"
		if string(content) != expected {
			t.Errorf("Expected content %q, got %q", expected, string(content))
		}

		// Test non-existent template
		_, err = mock.Open("nonexistent.md")
		if err == nil {
			t.Error("Expected error for non-existent template")
		}
	})

	// Test Stat
	t.Run("Stat", func(t *testing.T) {
		info, err := mock.Stat("template1.md")
		if err != nil {
			t.Fatalf("Failed to stat template: %v", err)
		}

		if info.Path != "template1.md" {
			t.Errorf("Expected path %q, got %q", "template1.md", info.Path)
		}

		expectedSize := int64(len("# Template 1\nContent of template 1"))
		if info.Size != expectedSize {
			t.Errorf("Expected size %d, got %d", expectedSize, info.Size)
		}

		if info.IsDir {
			t.Error("Expected IsDir to be false")
		}

		// Test non-existent template
		_, err = mock.Stat("nonexistent.md")
		if err == nil {
			t.Error("Expected error for non-existent template")
		}
	})

	// Test List
	t.Run("List", func(t *testing.T) {
		paths, err := mock.List()
		if err != nil {
			t.Fatalf("Failed to list templates: %v", err)
		}

		// Should only return .md files
		expectedPaths := []string{"template1.md", "template2.md"}
		if len(paths) != len(expectedPaths) {
			t.Errorf("Expected %d paths, got %d", len(expectedPaths), len(paths))
		}

		for i, path := range paths {
			if path != expectedPaths[i] {
				t.Errorf("Expected path %q at index %d, got %q", expectedPaths[i], i, path)
			}
		}
	})

	// Test Watch
	t.Run("Watch", func(t *testing.T) {
		watchChan, err := mock.Watch()
		if err != nil {
			t.Fatalf("Failed to call Watch: %v", err)
		}

		if watchChan != nil {
			t.Error("Expected nil watch channel for mock")
		}

		// Stop watching should be no-op
		err = mock.StopWatch()
		if err != nil {
			t.Fatalf("StopWatch failed: %v", err)
		}
	})

	// Test ResolveImport
	t.Run("ResolveImport", func(t *testing.T) {
		result := mock.ResolveImport("some/import", "current/path")
		if result != "" {
			t.Errorf("Expected empty string from ResolveImport, got %q", result)
		}
	})
}

func TestMockSourceEmpty(t *testing.T) {
	// Test creating MockSource with nil map
	mock := NewMockSource(nil)

	paths, err := mock.List()
	if err != nil {
		t.Fatalf("Failed to list templates: %v", err)
	}

	if len(paths) != 0 {
		t.Errorf("Expected 0 templates, got %d", len(paths))
	}
}

func TestMockSourceIsolation(t *testing.T) {
	// Test that external modifications don't affect the MockSource
	original := map[string]string{
		"test.md": "Original content",
	}

	mock := NewMockSource(original)

	// Modify the original map
	original["test.md"] = "Modified content"
	original["new.md"] = "New content"

	// Check that MockSource is not affected
	reader, err := mock.Open("test.md")
	if err != nil {
		t.Fatal("Failed to open test.md")
	}
	defer reader.Close()

	content, _ := io.ReadAll(reader)
	if string(content) != "Original content" {
		t.Errorf("Expected %q, got %q", "Original content", string(content))
	}

	_, err = mock.Open("new.md")
	if err == nil {
		t.Error("New template should not exist in MockSource")
	}
}

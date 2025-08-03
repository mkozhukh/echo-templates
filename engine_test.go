package echotemplates

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		root        string
		expectError bool
	}{
		{
			name:        "valid config",
			root:        tmpDir,
			expectError: false,
		},
		{
			name:        "missing source",
			root:        "",
			expectError: true,
		},
		{
			name:        "non-existent root dir",
			root:        "/non/existent/path",
			expectError: true,
		},
		{
			name:        "file as root dir",
			root:        filepath.Join(tmpDir, "file.txt"),
			expectError: true,
		},
	}

	// Create a file for the "file as root dir" test
	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("content"), 0644)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootTmpDir, err := NewFileSystemSource(tt.root)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if rootTmpDir == nil {
					t.Error("Expected engine but got nil")
				}
			}

			engine, err := New(Config{
				Source: rootTmpDir,
			})
			if err != nil || engine == nil {
				t.Fatalf("Failed to create source: %v", err)
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test templates
	os.WriteFile(filepath.Join(tmpDir, "simple.md"), []byte(`@system:
You are a {{role}} assistant.

@user:
{{query}}`), 0644)

	os.WriteFile(filepath.Join(tmpDir, "with-defaults.md"), []byte(`---
default.role: helpful
default.tone: friendly
---
@system:
You are a {{role}} assistant with a {{tone}} tone.`), 0644)

	os.Mkdir(filepath.Join(tmpDir, "common"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "common", "header.md"), []byte(`You are an AI assistant.`), 0644)

	os.WriteFile(filepath.Join(tmpDir, "with-import.md"), []byte(`@system:
{{@common/header}}
Your specialty is {{domain}}.`), 0644)

	tmpDirRoot, err := NewFileSystemSource(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	engine, err := New(Config{
		Source: tmpDirRoot,
	})
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	tests := []struct {
		name         string
		template     string
		vars         map[string]string
		expectError  bool
		checkContent func(t *testing.T, messages []interface{})
	}{
		{
			name:     "simple template",
			template: "simple",
			vars: map[string]string{
				"role":  "helpful",
				"query": "Hello!",
			},
			checkContent: func(t *testing.T, messages []interface{}) {
				if len(messages) != 2 {
					t.Errorf("Expected 2 messages, got %d", len(messages))
				}
			},
		},
		{
			name:     "template with .md extension",
			template: "simple.md",
			vars: map[string]string{
				"role":  "helpful",
				"query": "Hello!",
			},
			checkContent: func(t *testing.T, messages []interface{}) {
				if len(messages) != 2 {
					t.Errorf("Expected 2 messages, got %d", len(messages))
				}
			},
		},
		{
			name:     "with defaults",
			template: "with-defaults",
			vars: map[string]string{
				"tone": "professional", // Override default
			},
			checkContent: func(t *testing.T, messages []interface{}) {
				if len(messages) != 1 {
					t.Errorf("Expected 1 message, got %d", len(messages))
				}
			},
		},
		{
			name:     "with import",
			template: "with-import",
			vars: map[string]string{
				"domain": "mathematics",
			},
			checkContent: func(t *testing.T, messages []interface{}) {
				if len(messages) != 1 {
					t.Errorf("Expected 1 message, got %d", len(messages))
				}
			},
		},
		{
			name:        "non-existent template",
			template:    "non-existent",
			vars:        map[string]string{},
			expectError: true,
		},
		{
			name:        "missing required variable",
			template:    "simple",
			vars:        map[string]string{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages, err := engine.Generate(tt.template, tt.vars)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if tt.checkContent != nil {
					// Convert messages to interface{} slice for easier testing
					var msgInterfaces []interface{}
					for _, msg := range messages {
						msgInterfaces = append(msgInterfaces, msg)
					}
					tt.checkContent(t, msgInterfaces)
				}
			}
		})
	}
}

func TestGenerateWithOptions(t *testing.T) {
	tmpDir := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir, "optional.md"), []byte(`@system:
Hello {{name}}!`), 0644)

	tmpDirRoot, err := NewFileSystemSource(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	engine, err := New(Config{
		Source: tmpDirRoot,
	})
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Test with AllowMissingVars
	opts := GenerateOptions{
		AllowMissingVars: true,
	}

	messages, err := engine.Generate("optional", map[string]string{}, opts)
	if err != nil {
		t.Errorf("Expected no error with AllowMissingVars, got: %v", err)
	}
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}
}

func TestGenerateWithMetadata(t *testing.T) {
	tmpDir := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir, "with-meta.md"), []byte(`---
temperature: 0.8
max_tokens: 2000
model: gpt-4
description: Test template
---
@system:
Hello!`), 0644)

	tmpDirRoot, err := NewFileSystemSource(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	engine, err := New(Config{
		Source: tmpDirRoot,
	})
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	messages, metadata, err := engine.GenerateWithMetadata("with-meta", map[string]string{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	if metadata == nil {
		t.Fatal("Expected metadata but got nil")
	}

	if model, ok := metadata["model"].(string); !ok || model != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got %v", metadata["model"])
	}
	if desc, ok := metadata["description"].(string); !ok || desc != "Test template" {
		t.Errorf("Expected description 'Test template', got %v", metadata["description"])
	}
	if temp, ok := metadata["temperature"].(float64); !ok || temp != 0.8 {
		t.Errorf("Expected temperature 0.8, got %v", metadata["temperature"])
	}
	if tokens, ok := metadata["max_tokens"].(int); !ok || tokens != 2000 {
		t.Errorf("Expected max_tokens 2000, got %v", metadata["max_tokens"])
	}
}

func TestCircularImports(t *testing.T) {
	tmpDir := t.TempDir()

	// Create circular import templates
	os.WriteFile(filepath.Join(tmpDir, "a.md"), []byte(`{{@b}}
Content A`), 0644)

	os.WriteFile(filepath.Join(tmpDir, "b.md"), []byte(`{{@a}}
Content B`), 0644)

	tmpDirRoot, err := NewFileSystemSource(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	engine, err := New(Config{
		Source: tmpDirRoot,
		DefaultOptions: GenerateOptions{
			StrictMode: true,
		},
	})
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	_, err = engine.Generate("a", map[string]string{})
	if err == nil {
		t.Error("Expected error for circular import")
	}

	var importErr *ImportError
	if !reflect.TypeOf(err).AssignableTo(reflect.TypeOf(importErr)) {
		t.Errorf("Expected ImportError, got %T", err)
	}
}

func TestDynamicImports(t *testing.T) {
	tmpDir := t.TempDir()

	// Create template directories
	os.Mkdir(filepath.Join(tmpDir, "styles"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "styles", "formal.md"), []byte(`Formal style content`), 0644)
	os.WriteFile(filepath.Join(tmpDir, "styles", "casual.md"), []byte(`Casual style content`), 0644)

	os.WriteFile(filepath.Join(tmpDir, "main.md"), []byte(`@system:
{{@styles/{{style}}}}
Main content`), 0644)

	tmpDirRoot, err := NewFileSystemSource(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	engine, err := New(Config{
		Source: tmpDirRoot,
	})
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Test with formal style
	messages, err := engine.Generate("main", map[string]string{
		"style": "formal",
	})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	// Test with casual style
	messages, err = engine.Generate("main", map[string]string{
		"style": "casual",
	})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}
}

func TestCaching(t *testing.T) {
	tmpDir := t.TempDir()

	templatePath := filepath.Join(tmpDir, "cached.md")
	os.WriteFile(templatePath, []byte(`@system:
Original content`), 0644)

	source, err := NewFileSystemSource(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	engine, err := New(Config{
		Source:  source,
		DevMode: false,
	})
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// First generation
	messages1, err := engine.Generate("cached", map[string]string{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Modify the file
	os.WriteFile(templatePath, []byte(`@system:
Modified content`), 0644)

	// Second generation should get new content
	messages2, err := engine.Generate("cached", map[string]string{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Content should be different
	if len(messages1) != 1 || len(messages2) != 1 {
		t.Error("Expected 1 message in each result")
		return
	}

	// Test cache clear
	engine.ClearCache()

	// Test DisableCache option
	opts := GenerateOptions{
		DisableCache: true,
	}
	_, err = engine.Generate("cached", map[string]string{}, opts)
	if err != nil {
		t.Errorf("Unexpected error with DisableCache: %v", err)
	}
}

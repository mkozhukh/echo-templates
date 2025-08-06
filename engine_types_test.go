package echotemplates

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateWithDifferentTypes(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test template
	os.WriteFile(filepath.Join(tmpDir, "types.md"), []byte(`@user:
Name: {{name}}
Age: {{age}}
Score: {{score}}
Tags: {{tags}}`), 0644)

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
		vars         map[string]any
		wantContains []string
	}{
		{
			name: "string values",
			vars: map[string]any{
				"name":  "Alice",
				"age":   "25",
				"score": "98.5",
				"tags":  "go,testing",
			},
			wantContains: []string{"Alice", "25", "98.5", "go,testing"},
		},
		{
			name: "int value",
			vars: map[string]any{
				"name":  "Bob",
				"age":   30,
				"score": "95.0",
				"tags":  "python,ml",
			},
			wantContains: []string{"Bob", "30", "95.0", "python,ml"},
		},
		{
			name: "float64 value",
			vars: map[string]any{
				"name":  "Charlie",
				"age":   "28",
				"score": 92.75,
				"tags":  "rust,wasm",
			},
			wantContains: []string{"Charlie", "28", "92.75", "rust,wasm"},
		},
		{
			name: "[]string value",
			vars: map[string]any{
				"name":  "Diana",
				"age":   35,
				"score": 99.9,
				"tags":  []string{"java", "spring", "microservices"},
			},
			wantContains: []string{"Diana", "35", "99.9", "java,spring,microservices"},
		},
		{
			name: "mixed types",
			vars: map[string]any{
				"name":  "Eve",
				"age":   42,
				"score": 88.5,
				"tags":  []string{"javascript", "react", "node"},
			},
			wantContains: []string{"Eve", "42", "88.5", "javascript,react,node"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages, err := engine.Generate("types", tt.vars)
			if err != nil {
				t.Errorf("Generate() error = %v", err)
				return
			}

			if len(messages) != 1 {
				t.Errorf("Expected 1 message, got %d", len(messages))
				return
			}

			content := messages[0].Content
			for _, want := range tt.wantContains {
				if !contains(content, want) {
					t.Errorf("Expected content to contain %q, got: %s", want, content)
				}
			}
		})
	}
}

func TestGenerateWithUnsupportedTypes(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test template
	os.WriteFile(filepath.Join(tmpDir, "test.md"), []byte(`Value: {{value}}`), 0644)

	tmpDirRoot, err := NewFileSystemSource(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	engine, err := New(Config{
		Source: tmpDirRoot,
		DefaultOptions: GenerateOptions{
			AllowMissingVars: true,
		},
	})
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	tests := []struct {
		name     string
		vars     map[string]any
		expected string
	}{
		{
			name: "unsupported type - bool",
			vars: map[string]any{
				"value": true,
			},
			expected: "Value: ", // Unsupported types convert to empty string
		},
		{
			name: "unsupported type - map",
			vars: map[string]any{
				"value": map[string]string{"key": "val"},
			},
			expected: "Value: ", // Unsupported types convert to empty string
		},
		{
			name: "unsupported type - struct",
			vars: map[string]any{
				"value": struct{ Name string }{"test"},
			},
			expected: "Value: ", // Unsupported types convert to empty string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages, err := engine.Generate("test", tt.vars)
			if err != nil {
				t.Errorf("Generate() error = %v", err)
				return
			}

			if len(messages) != 1 {
				t.Errorf("Expected 1 message, got %d", len(messages))
				return
			}

			if messages[0].Content != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, messages[0].Content)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

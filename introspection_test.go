package echotemplates

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestTemplateIntrospection(t *testing.T) {
	// Create a temporary directory for test templates
	tmpDir, err := os.MkdirTemp("", "echotemplates-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test templates
	testTemplates := map[string]string{
		"simple.md": `@system:
You are a {{role}} assistant.

@user:
{{query}}`,
		"with-vars.md": `# default.style: friendly
# temperature: 0.7

@system:
You are a {{role|helpful}} assistant with {{style}} style.
Please help with {{{raw_content}}}.`,
		"nested/template.md": `@user:
Testing {{var1}} and {{var2|default}}`,
		"with-import.md": `{{@simple}}

@assistant:
I'll help you with {{topic}}.`,
	}

	for path, content := range testTemplates {
		fullPath := filepath.Join(tmpDir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	tmpDirRoot, err := NewFileSystemSource(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	engine, err := New(Config{
		Source:  tmpDirRoot,
		DevMode: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("TemplateExists", func(t *testing.T) {
		tests := []struct {
			name   string
			exists bool
		}{
			{"simple", true},
			{"simple.md", true},
			{"with-vars", true},
			{"nested/template", true},
			{"nonexistent", false},
			{"nested/nonexistent", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				exists := engine.TemplateExists(tt.name)
				if exists != tt.exists {
					t.Errorf("TemplateExists(%q) = %v, want %v", tt.name, exists, tt.exists)
				}
			})
		}
	})

	t.Run("ValidateTemplate", func(t *testing.T) {
		tests := []struct {
			name      string
			wantError bool
		}{
			{"simple", false},
			{"with-vars", false},
			{"nested/template", false},
			{"with-import", false},
			{"nonexistent", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := engine.ValidateTemplate(tt.name)
				if (err != nil) != tt.wantError {
					t.Errorf("ValidateTemplate(%q) error = %v, wantError %v", tt.name, err, tt.wantError)
				}
			})
		}
	})

	t.Run("GetTemplateVariables", func(t *testing.T) {
		tests := []struct {
			name     string
			expected []string
		}{
			{"simple", []string{"query", "role"}},
			{"with-vars", []string{"raw_content", "role", "style"}},
			{"nested/template", []string{"var1", "var2"}},
			{"with-import", []string{"query", "role", "topic"}}, // includes vars from imported template
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				vars, err := engine.GetTemplateVariables(tt.name)
				if err != nil {
					t.Fatalf("GetTemplateVariables(%q) error = %v", tt.name, err)
				}
				sort.Strings(vars)
				sort.Strings(tt.expected)
				if !reflect.DeepEqual(vars, tt.expected) {
					t.Errorf("GetTemplateVariables(%q) = %v, want %v", tt.name, vars, tt.expected)
				}
			})
		}
	})

	t.Run("ListTemplates", func(t *testing.T) {
		templates, err := engine.ListTemplates()
		if err != nil {
			t.Fatalf("ListTemplates() error = %v", err)
		}

		expected := []string{
			"nested/template",
			"simple",
			"with-import",
			"with-vars",
		}

		if !reflect.DeepEqual(templates, expected) {
			t.Errorf("ListTemplates() = %v, want %v", templates, expected)
		}
	})
}

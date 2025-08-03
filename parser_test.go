package echotemplates

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseFrontMatter(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedMeta    map[string]any
		expectedContent string
	}{
		{
			name: "full front-matter",
			input: `---
temperature: 0.7
max_tokens: 1000
model: gpt-4
description: A helpful assistant template
default.role: helpful
default.style: professional
---
@system:
You are a {{role}} assistant.`,
			expectedMeta: map[string]any{
				"temperature": 0.7,
				"max_tokens":  1000,
				"model":       "gpt-4",
				"description": "A helpful assistant template",
				"defaults": map[string]string{
					"role":  "helpful",
					"style": "professional",
				},
			},
			expectedContent: `@system:
You are a {{role}} assistant.`,
		},
		{
			name: "no front-matter",
			input: `@system:
You are an assistant.`,
			expectedMeta: map[string]any{
				"defaults": map[string]string{},
			},
			expectedContent: `@system:
You are an assistant.`,
		},
		{
			name: "partial front-matter",
			input: `---
model: claude-3
default.tone: friendly
---
Content here`,
			expectedMeta: map[string]any{
				"model": "claude-3",
				"defaults": map[string]string{
					"tone": "friendly",
				},
			},
			expectedContent: `Content here`,
		},
		{
			name: "invalid values ignored",
			input: `---
temperature: invalid
max_tokens: not-a-number
model: gpt-4
---
Content`,
			expectedMeta: map[string]any{
				"temperature": "invalid",
				"max_tokens":  "not-a-number",
				"model":       "gpt-4",
				"defaults":    map[string]string{},
			},
			expectedContent: `Content`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			meta, content, err := parseFrontMatter(reader)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check metadata using DeepEqual since it's now a map
			if !reflect.DeepEqual(meta, tt.expectedMeta) {
				t.Errorf("Metadata mismatch:\nexpected: %v\ngot: %v", tt.expectedMeta, meta)
			}

			// Check content
			if content != tt.expectedContent {
				t.Errorf("Content: expected %q, got %q", tt.expectedContent, content)
			}
		})
	}
}

func TestSubstituteVariables(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		vars        map[string]string
		defaults    map[string]string
		opts        GenerateOptions
		expected    string
		expectError bool
	}{
		{
			name:    "simple substitution",
			content: "Hello {{name}}, welcome to {{place}}!",
			vars: map[string]string{
				"name":  "Alice",
				"place": "Wonderland",
			},
			expected: "Hello Alice, welcome to Wonderland!",
		},
		{
			name:    "with defaults",
			content: "Hello {{name|World}}, you are {{role|guest}}!",
			vars: map[string]string{
				"name": "Bob",
			},
			expected: "Hello Bob, you are guest!",
		},
		{
			name:    "raw placeholders",
			content: "Code: {{{code}}} and {{formatted}}",
			vars: map[string]string{
				"code":      "<script>alert('hi')</script>",
				"formatted": "<b>bold</b>",
			},
			expected: "Code: <script>alert('hi')</script> and <b>bold</b>",
		},
		{
			name:    "missing variable error",
			content: "Hello {{name}}!",
			vars:    map[string]string{},
			opts: GenerateOptions{
				AllowMissingVars: false,
			},
			expectError: true,
		},
		{
			name:    "missing variable allowed",
			content: "Hello {{name}}!",
			vars:    map[string]string{},
			opts: GenerateOptions{
				AllowMissingVars: true,
			},
			expected: "Hello {{name}}!",
		},
		{
			name:    "preserve import placeholders",
			content: "{{@common/header}} Hello {{name}}!",
			vars: map[string]string{
				"name": "Charlie",
			},
			expected: "{{@common/header}} Hello Charlie!",
		},
		{
			name:    "use defaults from metadata",
			content: "Style: {{style}}, Tone: {{tone}}",
			vars: map[string]string{
				"style": "modern",
			},
			defaults: map[string]string{
				"style": "classic",
				"tone":  "formal",
			},
			expected: "Style: modern, Tone: formal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Merge defaults into vars to match the new behavior
			mergedVars := make(map[string]string)
			for k, v := range tt.defaults {
				mergedVars[k] = v
			}
			for k, v := range tt.vars {
				mergedVars[k] = v
			}

			result, err := substituteVariables(tt.content, mergedVars, nil, tt.opts)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExtractImports(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "single import",
			content:  "{{@common/header}} Some content",
			expected: []string{"common/header"},
		},
		{
			name:     "multiple imports",
			content:  "{{@header}} Middle {{@footer}} End",
			expected: []string{"header", "footer"},
		},
		{
			name:     "dynamic import",
			content:  "{{@templates/{{type}}/main}}",
			expected: []string{"templates/{{type}}/main"},
		},
		{
			name:     "no imports",
			content:  "Just {{variable}} placeholders",
			expected: []string{},
		},
		{
			name: "mixed content",
			content: `{{@common/personality}}
@system:
You are a {{role}} assistant.
{{@prompts/{{domain}}/system-prompt}}`,
			expected: []string{"common/personality", "prompts/{{domain}}/system-prompt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imports := extractImports(tt.content)

			if !reflect.DeepEqual(imports, tt.expected) {
				t.Errorf("Expected imports %v, got %v", tt.expected, imports)
			}
		})
	}
}

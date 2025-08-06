package echotemplates

import (
	"testing"
)

func TestStringGeneration(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		vars     map[string]any
		wantErr  bool
		wantMsgs int
	}{
		{
			name:     "simple template",
			content:  "Hello {{name}}!",
			vars:     map[string]any{"name": "World"},
			wantMsgs: 1,
		},
		{
			name: "template with roles",
			content: `@system:
You are {{role}}.

@user:
{{question}}`,
			vars:     map[string]any{"role": "helpful", "question": "What is Go?"},
			wantMsgs: 2,
		},
		{
			name: "template with defaults",
			content: `---
default.greeting: Hello
default.name: Friend
---
{{greeting}} {{name}}!`,
			vars:     map[string]any{"name": "Alice"},
			wantMsgs: 1,
		},
		{
			name: "template with raw content",
			content: `Code:
{{{code}}}`,
			vars:     map[string]any{"code": "func main() {\n\tfmt.Println(\"Hello\")\n}"},
			wantMsgs: 1,
		},
		{
			name:    "imports not supported",
			content: "{{@common/header}}",
			vars:    map[string]any{},
			wantErr: true,
		},
		{
			name:    "missing variable in strict mode",
			content: "Hello {{name}}!",
			vars:    map[string]any{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := GenerateOptions{}
			if tt.name == "missing variable in strict mode" {
				opts.AllowMissingVars = false
			}

			messages, err := Generate(tt.content, tt.vars, opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(messages) != tt.wantMsgs {
				t.Errorf("Generate() returned %d messages, want %d", len(messages), tt.wantMsgs)
			}
		})
	}
}

func TestStringGenerationWithMetadata(t *testing.T) {
	content := `---
temperature: 0.7
max_tokens: 1000
model: gpt-4
default.role: assistant
---
@system:
You are a {{role}}.`

	messages, metadata, err := GenerateWithMetadata(content, map[string]any{})
	if err != nil {
		t.Errorf("GenerateWithMetadata() error = %v", err)
		return
	}

	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	// Check metadata
	if temp, ok := metadata["temperature"].(float64); !ok || temp != 0.7 {
		t.Errorf("Expected temperature 0.7, got %v", metadata["temperature"])
	}

	if model, ok := metadata["model"].(string); !ok || model != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got %v", metadata["model"])
	}

	if maxTokens, ok := metadata["max_tokens"].(int); !ok || maxTokens != 1000 {
		t.Errorf("Expected max_tokens 1000, got %v", metadata["max_tokens"])
	}

	// Check default substitution
	if messages[0].Content != "You are a assistant." {
		t.Errorf("Expected 'You are a assistant.', got '%s'", messages[0].Content)
	}
}

func TestStringSource(t *testing.T) {
	source := &stringSource{}

	// Test Open - path is the content
	content := "test content"
	reader, err := source.Open(content)
	if err != nil {
		t.Errorf("Open() error = %v", err)
	}
	reader.Close()

	// Test Stat
	info, err := source.Stat(content)
	if err != nil {
		t.Errorf("Stat() error = %v", err)
	}
	if info.Size != int64(len(content)) {
		t.Errorf("Expected size %d, got %d", len(content), info.Size)
	}

	// Test List (should return empty)
	list, err := source.List()
	if err != nil || len(list) != 0 {
		t.Errorf("List() should return empty list")
	}

	// Test Watch (should return nil)
	ch, err := source.Watch()
	if err != nil || ch != nil {
		t.Errorf("Watch() should return nil channel")
	}

	// Test ResolveImport (should return empty)
	result := source.ResolveImport("any", "path")
	if result != "" {
		t.Errorf("ResolveImport() should return empty string")
	}
}

package echotemplates

import (
	"errors"
	"testing"
)

func TestTemplateNotFoundError(t *testing.T) {
	err := &TemplateNotFoundError{
		Name: "test.md",
		Path: "/path/to/test.md",
	}

	expected := "template not found: test.md (path: /path/to/test.md)"
	if err.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, err.Error())
	}
}

func TestVariableError(t *testing.T) {
	err := &VariableError{
		Variable: "username",
		Template: "greeting.md",
	}

	expected := `variable "username" not found in template "greeting.md"`
	if err.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, err.Error())
	}
}

func TestImportError(t *testing.T) {
	innerErr := errors.New("file not found")
	err := &ImportError{
		ImportPath: "common/header.md",
		Template:   "main.md",
		Cause:      innerErr,
	}

	expected := `failed to import "common/header.md" in template "main.md": file not found`
	if err.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, err.Error())
	}
}

func TestParseError(t *testing.T) {
	tests := []struct {
		name     string
		err      ParseError
		expected string
	}{
		{
			name: "with line number",
			err: ParseError{
				Template: "test.md",
				Line:     42,
				Message:  "unexpected token",
			},
			expected: `parse error in template "test.md" at line 42: unexpected token`,
		},
		{
			name: "without line number",
			err: ParseError{
				Template: "test.md",
				Message:  "invalid syntax",
			},
			expected: `parse error in template "test.md": invalid syntax`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("Expected error message %q, got %q", tt.expected, tt.err.Error())
			}
		})
	}
}

package echotemplates

import "fmt"

// TemplateNotFoundError indicates that a template file was not found
type TemplateNotFoundError struct {
	Name string
	Path string
}

func (e *TemplateNotFoundError) Error() string {
	return fmt.Sprintf("template not found: %s (path: %s)", e.Name, e.Path)
}

// VariableError indicates a missing or invalid variable
type VariableError struct {
	Variable string
	Template string
}

func (e *VariableError) Error() string {
	return fmt.Sprintf("variable %q not found in template %q", e.Variable, e.Template)
}

// ImportError indicates a failure during template import
type ImportError struct {
	ImportPath string
	Template   string
	Cause      error
}

func (e *ImportError) Error() string {
	return fmt.Sprintf("failed to import %q in template %q: %v", e.ImportPath, e.Template, e.Cause)
}

// ParseError indicates a template parsing error
type ParseError struct {
	Template string
	Line     int
	Message  string
}

func (e *ParseError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("parse error in template %q at line %d: %s", e.Template, e.Line, e.Message)
	}
	return fmt.Sprintf("parse error in template %q: %s", e.Template, e.Message)
}

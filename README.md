# Echo Templates

Echo Templates extends the [echo](https://github.com/mkozhukh/echo) library with templating capabilities for LLM prompt construction.

## Installation

```bash
go get github.com/mkozhukh/echo-templates
```

## Features

- 🚀 **Simple string templates** - Generate messages directly from strings
- 📁 **File-based templates** - Organize templates in filesystem
- 📦 **Embedded templates** - Bundle templates in your binary
- 🔄 **Hot reload** - Automatic template reloading in development
- 🎯 **Smart imports** - Compose templates with dynamic imports
- ⚙️ **Extensible** - Custom template sources and import resolution

## Quick Start

### Simple String Template

For quick and simple use cases, generate messages directly from strings:

```go
package main

import (
    "fmt"
    "github.com/mkozhukh/echo-templates"
)

func main() {
    // Generate messages from a string template
    messages, err := echotemplates.Generate(
        "Hello {{name}}! How can I help you with {{topic}} today?",
        map[string]any{
            "name": "Alice",
            "topic": "Go programming",
        },
    )
    if err != nil {
        panic(err)
    }
    
    fmt.Println(messages[0].Content)
    // Output: Hello Alice! How can I help you with Go programming today?
}
```

### File-based Templates

```go
package main

import (
    "context"
    "fmt"
    "github.com/mkozhukh/echo"
    "github.com/mkozhukh/echo-templates"
)

func main() {
    // Create a filesystem source
    source, err := echotemplates.NewFileSystemSource("./prompts")
    if err != nil {
        panic(err)
    }
    
    // Create a template engine
    engine, err := echotemplates.New(echotemplates.Config{
        Source: source,
        DevMode: false, // Enable for development (file watching, no caching)
    })
    if err != nil {
        panic(err)
    }
    
    // Generate messages from a template
    messages, err := engine.Generate("chat/assistant", map[string]any{
        "role": "helpful",
        "domain": "mathematics",
        "user_query": "What is calculus?",
    })
    if err != nil {
        panic(err)
    }
    
    // Use with echo client
    client, _ := echo.NewClient("openai/gpt-4", "api-key")
    resp, _ := client.Call(context.Background(), messages)
    fmt.Println(resp.Text)
}
```

## Template Format

Templates use the same format as echo's TemplateMessage with additional features:

### Basic Structure

```markdown
@system:
You are a {{role}} assistant specializing in {{domain}}.

@user:
{{user_query}}

@assistant:
I'll help you with {{domain}}. Let me analyze your request.
```

### Front-matter (Optional)

Templates can include metadata using YAML-like syntax at the beginning of the file:

```markdown
---
temperature: 0.7
max_tokens: 1000
model: gpt-4
description: A helpful assistant template
default.role: helpful
default.style: professional
---
@system:
You are a {{role}} assistant with a {{style}} communication style.
```

Front-matter must be delimited by `---` lines and appear at the very beginning of the file. It supports any key-value pairs:
- Keys can be any string
- Values can be strings or numbers (integers or floats)
- Keys starting with `default.` define default values for variables
- Common fields include `temperature`, `max_tokens`, `model`, `description`

## Template Syntax

### Placeholders

1. **Simple substitution**: `{{variable_name}}`
   ```markdown
   Hello {{name}}, welcome to {{place}}!
   ```

2. **With default values**: `{{variable_name|default_value}}`
   ```markdown
   You are a {{role|helpful}} assistant.
   ```

3. **Raw content (no escaping)**: `{{{raw_content}}}`
   ```markdown
   Code: {{{code_snippet}}}
   ```

### Imports

Include content from other templates:

1. **Simple import**: `{{@file_path}}`
   ```markdown
   {{@common/header}}
   
   @user:
   {{user_query}}
   ```

2. **Dynamic import**: `{{@folder/{{variable}}/file}}`
   ```markdown
   {{@styles/{{style_type}}/intro}}
   {{@personas/{{persona}}}}
   ```

### Processing Order

1. **Import Resolution** - All `{{@...}}` imports are processed recursively
2. **Variable Substitution** - All `{{variable}}` placeholders are replaced
3. **Message Parsing** - Content is split into messages using `@role:` markers

## Configuration

### Template Sources

Echo Templates supports multiple template sources through the `TemplateSource` interface:

#### Filesystem Source
```go
// Create a filesystem source
source, err := echotemplates.NewFileSystemSource("./prompts")

engine, err := echotemplates.New(echotemplates.Config{
    Source: source,
    DevMode: true,  // Enable file watching and disable caching
})
```

#### Embedded Templates
```go
//go:embed prompts/*
var embeddedTemplates embed.FS

// Create an embedded source
source := echotemplates.NewEmbedSource(embeddedTemplates, "prompts")

engine, err := echotemplates.New(echotemplates.Config{
    Source: source,
    DevMode: false, // Production mode with caching
})
```

### Engine Configuration

```go
engine, err := echotemplates.New(echotemplates.Config{
    // Template source (required)
    Source: source,
    
    // Development mode (default: false)
    // - true: disables caching, enables file watching
    // - false: enables caching for production
    DevMode: false,
    
    // Maximum number of templates to cache (default: 100)
    CacheSize: 100,
    
    // Default options for all Generate calls
    DefaultOptions: echotemplates.GenerateOptions{
        StrictMode: true,
        AllowMissingVars: false,
    },
})
```

### Generation Options

Options are optional and can be passed as the last parameter:

```go
// Without options (uses default options)
messages, err := engine.Generate("template", vars)

// With custom options
messages, err := engine.Generate("template", vars, 
    echotemplates.GenerateOptions{
        // Allow missing variables (default: false)
        AllowMissingVars: true,
        
        // Enable strict parsing (default: false)
        StrictMode: true,
        
        // Bypass cache for this generation (default: false)
        DisableCache: true,
    },
)
```

## API Reference

### Package-level Functions

These functions provide a simple way to generate messages from string templates without creating an engine:

#### Generate

```go
func Generate(content string, vars map[string]any, opts ...GenerateOptions) ([]echo.Message, error)
```

Generates messages from a string template.

**Parameters:**
- `content` - The template string with placeholders
- `vars` - Variables to substitute in the template (supports string, int, float64, []string)
- `opts` - Optional generation options

**Examples:**
```go
// Simple template (creates user message)
messages, err := echotemplates.Generate(
    "Hello {{name}}, welcome to {{place}}!",
    map[string]any{"name": "Alice", "place": "Wonderland"},
)
// messages[0].Role == "user"
// messages[0].Content == "Hello Alice, welcome to Wonderland!"

// Template with different value types
messages, err = echotemplates.Generate(
    "User {{name}} (age {{age}}) scored {{score}} with tags: {{tags}}",
    map[string]any{
        "name": "Bob",
        "age": 25,                // int
        "score": 98.5,             // float64
        "tags": []string{"go", "testing"},  // []string -> "go,testing"
    },
)

// Template with role markers
messages, err = echotemplates.Generate(
    "@system:\nYou are a {{role}} assistant.\n\n@user:\n{{query}}",
    map[string]any{"role": "helpful", "query": "What is Go?"},
)
// messages[0].Role == "system"
// messages[1].Role == "user"
```

#### GenerateWithMetadata

```go
func GenerateWithMetadata(content string, vars map[string]any, opts ...GenerateOptions) ([]echo.Message, map[string]any, error)
```

Generates messages from a string template and returns any metadata defined in front-matter.

**Example:**
```go
template := `---
temperature: 0.7
model: gpt-4
---
Hello {{name}}!`

messages, metadata, err := echotemplates.GenerateWithMetadata(
    template,
    map[string]any{"name": "Bob"},
)
// metadata["temperature"] == 0.7
// metadata["model"] == "gpt-4"
```

**Notes:** 
- String templates do not support imports (`{{@...}}`). Use file-based or embedded sources for templates with imports.
- If no role markers (`@role:`) are present, the content becomes a single user message.

#### CallOptions

```go
func CallOptions(metadata map[string]any) []echo.CallOption
```

Creates echo.CallOption slice from template metadata for configuring LLM API calls.

**Example:**
```go
messages, metadata, _ := engine.GenerateWithMetadata("template", vars)
opts := echotemplates.CallOptions(metadata)
client.Call(ctx, messages, opts...)
```

The `CallOptions` function automatically extracts and converts:

- `model` (string) → `echo.WithModel(model)`
- `temperature` (float64) → `echo.WithTemperature(temp)`
- `max_tokens` (int) → `echo.WithMaxTokens(maxTokens)`


### Engine Functions

For file-based templates, create an engine with a template source:

```go
// Create source
source, err := echotemplates.NewFileSystemSource("./templates")

// Create engine
engine, err := echotemplates.New(echotemplates.Config{
    Source: source,
    DevMode: false,
})

// Generate from file
messages, err := engine.Generate("prompt", vars)
```

## Advanced Usage

### Using Template Metadata

```go
// Get template with metadata
messages, metadata, err := engine.GenerateWithMetadata("ai/creative", vars)

// Access metadata fields
if temp, ok := metadata["temperature"].(float64); ok {
    // Use temperature value
}

if maxTokens, ok := metadata["max_tokens"].(int); ok {
    // Use max_tokens value
}

// Access defaults
if defaults, ok := metadata["defaults"].(map[string]any); ok {
    // Use default values
}
```

### Custom Template Sources

Implement the `TemplateSource` interface to create custom sources:

```go
type TemplateSource interface {
    Open(path string) (io.ReadCloser, error)
    Stat(path string) (TemplateInfo, error)
    List() ([]string, error)
    Watch() (<-chan string, error)
    StopWatch() error
    ResolveImport(importPath, currentPath string) string
}
```

Example: Database-backed templates, remote templates, etc.

### Custom Import Resolution

Override import resolution for relative imports or custom logic:

```go
type customSource struct {
    *echotemplates.FileSystemSource
}

func (s *customSource) ResolveImport(importPath, currentPath string) string {
    // Custom logic: resolve imports relative to current template
    if !filepath.IsAbs(importPath) {
        dir := filepath.Dir(currentPath)
        return filepath.Join(dir, importPath)
    }
    return "" // Use default resolution
}
```

### File Watching in Development

In dev mode, the filesystem source automatically watches for template changes:

```go
engine, err := echotemplates.New(echotemplates.Config{
    Source: source,
    DevMode: true, // Enables file watching
})

// Templates are automatically reloaded when files change
// No need to restart the application during development
```

### Dynamic Imports

Create flexible templates with variable-based imports:

```markdown
{{@personas/{{persona_type}}}}
{{@styles/{{writing_style}}}}

@system:
You are configured with the above persona and style.

@user:
{{user_input}}
```

### Template Introspection

```go
// Check if a template exists
exists := engine.TemplateExists("chat/assistant")

// List all available templates
templates, err := engine.ListTemplates()

// Get all variables used in a template
vars, err := engine.GetTemplateVariables("chat/assistant")

// Validate a template without generating
err := engine.ValidateTemplate("chat/assistant")
```

### Clearing Cache

During development or when templates change:

```go
engine.ClearCache()
```

## Error Handling

The library provides specific error types for better error handling:

```go
messages, err := engine.Generate("template", vars)
if err != nil {
    switch e := err.(type) {
    case *echotemplates.TemplateNotFoundError:
        // Handle missing template
        fmt.Printf("Template not found: %s\n", e.Name)
    case *echotemplates.VariableError:
        // Handle missing variable
        fmt.Printf("Missing variable: %s\n", e.Variable)
    case *echotemplates.ImportError:
        // Handle import failure
        fmt.Printf("Import failed: %s\n", e.ImportPath)
    case *echotemplates.ParseError:
        // Handle parse error
        fmt.Printf("Parse error at line %d: %s\n", e.Line, e.Message)
    }
}
```

## Caching

The template engine implements an LRU cache with automatic invalidation:

- Templates are cached after parsing (before variable substitution)
- Cache is automatically disabled in dev mode
- In production mode with filesystem source, cache is invalidated when template files are modified
- Cache size is configurable
- Can be disabled globally or per-request

## Thread Safety

The template engine is thread-safe and can be used concurrently from multiple goroutines.

## Why Echo Templates?

Echo Templates is about doing one thing well: managing LLM prompts.

### Designed for LLM Prompts

Unlike Go's `text/template` or other template engines, Echo Templates understands the structure of LLM conversations:

- **Role-based messages** - Native support for `@system:`, `@user:`, `@assistant:` markers
- **Message arrays** - Outputs `[]echo.Message` ready for LLM APIs, not just strings
- **Front-matter metadata** - Embed temperature, model preferences, and other LLM parameters directly in templates
- **Minimal syntax** - Just `{{variables}}` and `{{@imports}}`, no loops or conditionals
- **File watching** - Templates reload automatically during prompt iteration

### What It's NOT

Echo Templates intentionally doesn't include:

- ❌ Logic operations (if/else, loops, comparisons)
- ❌ Function calls or pipelines
- ❌ Complex data structures
- ❌ HTML/JS escaping
- ❌ General-purpose templating features

If you need these features, use Go's `text/template`. Echo Templates focuses on making LLM prompt management simple, maintainable, and purpose-built for AI applications.

### Example: Why It Matters

```go
// With text/template (general purpose)
tmpl := template.Must(template.New("prompt").Parse(`
{{if .IsCreative}}You are a creative assistant.{{else}}You are a helpful assistant.{{end}}

User: {{.Query}}`))

var buf bytes.Buffer
tmpl.Execute(&buf, data)
// Now you need to parse this into messages somehow...

// With Echo Templates (purpose-built)
messages, _ := echotemplates.Generate(`
	@system: You are a {{role}} assistant.
	@user:
	{{query}}`,
    map[string]any{"role": roleType, "query": userQuery},
)
// Ready to send to OpenAI/Claude/etc!
```


## License

MIT License

Copyright (c) 2025 Maksim Kozhukh

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

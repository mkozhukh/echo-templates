# Echo Templates

Echo Templates extends the [echo](https://github.com/mkozhukh/echo) library with templating capabilities for LLM prompt construction.

## Installation

```bash
go get github.com/mkozhukh/echo-templates
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "github.com/mkozhukh/echo"
    "github.com/mkozhukh/echo-templates"
)

func main() {
    // Create a template engine
    engine, err := echotemplates.New(echotemplates.Config{
        RootDir: "./prompts",
    })
    if err != nil {
        panic(err)
    }
    
    // Generate messages from a template
    messages, err := engine.Generate("chat/assistant", map[string]string{
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

Templates can include metadata using comment syntax at the beginning of the file:

```markdown
# temperature: 0.7
# max_tokens: 1000
# model: gpt-4
# description: A helpful assistant template
# default.role: helpful
# default.style: professional

@system:
You are a {{role}} assistant with a {{style}} communication style.
```

Front-matter supports any key-value pairs:
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
   {{@personas/{{persona}}}
   ```

### Processing Order

1. **Import Resolution** - All `{{@...}}` imports are processed recursively
2. **Variable Substitution** - All `{{variable}}` placeholders are replaced
3. **Message Parsing** - Content is split into messages using `@role:` markers

## Configuration

### Engine Configuration

```go
engine, err := echotemplates.New(echotemplates.Config{
    // Base directory for template files
    RootDir: "./prompts",
    
    // Enable in-memory caching (default: true)
    EnableCache: true,
    
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

### Clearing Cache

During development, you may want to clear the template cache:

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
- Cache is invalidated when template files are modified
- Cache size is configurable
- Can be disabled globally or per-request

## Thread Safety

The template engine is thread-safe and can be used concurrently from multiple goroutines.

## License

MIT License

Copyright (c) 2025

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
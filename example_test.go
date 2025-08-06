package echotemplates_test

import (
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

	echotemplates "github.com/mkozhukh/echo-templates"
)

func ExampleFileSystemSource() {
	// Create a filesystem source
	source, err := echotemplates.NewFileSystemSource("./templates")
	if err != nil {
		log.Fatal(err)
	}

	// Create engine with dev mode enabled (no caching, file watching)
	engine, err := echotemplates.New(echotemplates.Config{
		Source:  source,
		DevMode: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Generate messages
	messages, err := engine.Generate("hello", map[string]any{
		"name": "World",
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, msg := range messages {
		fmt.Printf("Role: %s, Content: %s\n", msg.Role, msg.Content)
	}
}

func ExampleEmbedSource() {
	// Example with embedded templates
	// Assume you have:
	// //go:embed prompts/*
	// var embeddedTemplates embed.FS

	// For this example, we'll create a dummy embed.FS
	var embeddedTemplates embed.FS

	// Create an embedded source
	source := echotemplates.NewEmbedSource(embeddedTemplates, "prompts")

	// Create engine in production mode (with caching)
	engine, err := echotemplates.New(echotemplates.Config{
		Source:    source,
		DevMode:   false,
		CacheSize: 100,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Generate messages
	messages, err := engine.Generate("example", map[string]any{
		"topic": "AI",
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, msg := range messages {
		fmt.Printf("Role: %s, Content: %s\n", msg.Role, msg.Content)
	}
}

func Example_customImportResolver() {
	// Create a custom source with import resolution
	type customSource struct {
		echotemplates.TemplateSource
		baseDir string
	}

	// Implement custom import resolution
	type customFileSource struct {
		*echotemplates.FileSystemSource
	}

	baseSource, _ := echotemplates.NewFileSystemSource("./templates")

	// Wrap the source to override import resolution
	source := &struct {
		echotemplates.TemplateSource
	}{
		TemplateSource: baseSource,
	}

	// Note: In a real implementation, you would create a proper type
	// that embeds FileSystemSource and overrides the ResolveImport method

	engine, err := echotemplates.New(echotemplates.Config{
		Source: source,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Use the engine...
	_ = engine
}

func Example_stringGeneration() {
	// Generate messages directly from a string template
	messages, _ := echotemplates.Generate("Hello {{name}}!", map[string]any{"name": "World"})
	fmt.Printf("Content: %s\n", messages[0].Content)
	// Output: Content: Hello World!
}

func Example_fileWatching() {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "templates")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a filesystem source
	source, err := echotemplates.NewFileSystemSource(tmpDir)
	if err != nil {
		log.Fatal(err)
	}

	// Create engine with dev mode (enables file watching)
	engine, err := echotemplates.New(echotemplates.Config{
		Source:  source,
		DevMode: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create a template file
	templatePath := filepath.Join(tmpDir, "test.md")
	err = os.WriteFile(templatePath, []byte("Hello {{name}}!"), 0644)
	if err != nil {
		log.Fatal(err)
	}

	// Generate with initial content
	messages, _ := engine.Generate("test", map[string]any{"name": "World"})
	fmt.Printf("Initial: %s\n", messages[0].Content)

	// Modify the template file
	err = os.WriteFile(templatePath, []byte("Hi {{name}}!"), 0644)
	if err != nil {
		log.Fatal(err)
	}

	// Wait a moment for file watcher to detect change
	// (In real usage, this happens automatically)

	// Generate again - will use updated content
	messages, _ = engine.Generate("test", map[string]any{"name": "World"})
	fmt.Printf("Updated: %s\n", messages[0].Content)
}

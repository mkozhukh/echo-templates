package echotemplates

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// parseFrontMatter extracts front-matter from the beginning of a template
func parseFrontMatter(reader io.Reader) (map[string]any, string, error) {
	metadata := make(map[string]any)
	defaults := make(map[string]string)
	metadata["defaults"] = defaults

	scanner := bufio.NewScanner(reader)
	var contentBuilder strings.Builder
	inFrontMatter := true

	for scanner.Scan() {
		line := scanner.Text()

		if inFrontMatter && strings.HasPrefix(line, "# ") {
			// Parse front-matter line
			line = strings.TrimPrefix(line, "# ")
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// Check for default.variable format
				if strings.HasPrefix(key, "default.") {
					varName := strings.TrimPrefix(key, "default.")
					defaults[varName] = value
				} else {
					// Try to parse as number for regular metadata
					if num, err := strconv.ParseFloat(value, 64); err == nil {
						if num == float64(int(num)) {
							metadata[key] = int(num)
						} else {
							metadata[key] = num
						}
					} else {
						metadata[key] = value
					}
				}
			}
		} else {
			// End of front-matter, rest is content
			inFrontMatter = false
			if contentBuilder.Len() > 0 {
				contentBuilder.WriteString("\n")
			}
			contentBuilder.WriteString(line)
		}
	}

	// Read remaining content
	for scanner.Scan() {
		if contentBuilder.Len() > 0 {
			contentBuilder.WriteString("\n")
		}
		contentBuilder.WriteString(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, "", err
	}

	content := strings.TrimRight(contentBuilder.String(), "\n")
	return metadata, content, nil
}

var (
	// Regular expressions for parsing
	placeholderRegex    = regexp.MustCompile(`\{\{([^}]+)\}\}`)
	importRegex         = regexp.MustCompile(`\{\{@(.+?)\}\}`)
	rawPlaceholderRegex = regexp.MustCompile(`\{\{\{([^}]+)\}\}\}`)
)

// parsedTemplate represents a template after initial parsing
type parsedTemplate struct {
	metadata map[string]any
	content  string
	imports  []string
}

// substituteVariables replaces placeholders with actual values
func substituteVariables(content string, vars map[string]string, defaults map[string]string, opts GenerateOptions) (string, error) {
	// First handle triple-brace raw placeholders
	content = rawPlaceholderRegex.ReplaceAllStringFunc(content, func(match string) string {
		varName := strings.TrimSpace(match[3 : len(match)-3])
		if value, ok := vars[varName]; ok {
			return value
		}
		return match // Keep original if not found
	})

	// Handle regular placeholders
	var missingVars []string

	content = placeholderRegex.ReplaceAllStringFunc(content, func(match string) string {
		// Skip import placeholders
		if strings.HasPrefix(match, "{{@") {
			return match
		}

		inner := strings.TrimSpace(match[2 : len(match)-2])

		// Check for default value syntax
		parts := strings.SplitN(inner, "|", 2)
		varName := strings.TrimSpace(parts[0])
		defaultValue := ""
		if len(parts) > 1 {
			defaultValue = strings.TrimSpace(parts[1])
		}

		// Try to get value from vars, then defaults, then use default value
		if value, ok := vars[varName]; ok {
			return value
		}
		if defaultValue != "" {
			return defaultValue
		}

		// Variable not found
		if !opts.AllowMissingVars {
			missingVars = append(missingVars, varName)
		}
		return match // Keep original placeholder
	})

	if len(missingVars) > 0 && !opts.AllowMissingVars {
		return "", &VariableError{
			Variable: strings.Join(missingVars, ", "),
			Template: "current",
		}
	}

	return content, nil
}

// extractImports finds all import placeholders in content
func extractImports(content string) []string {
	// Use a more permissive approach to handle nested placeholders
	imports := []string{}
	start := 0
	for {
		idx := strings.Index(content[start:], "{{@")
		if idx == -1 {
			break
		}
		idx += start

		// Find the closing }}
		end := idx + 3
		braceCount := 1
		for end < len(content) && braceCount > 0 {
			if end+1 < len(content) && content[end:end+2] == "{{" {
				braceCount++
				end += 2
			} else if end+1 < len(content) && content[end:end+2] == "}}" {
				braceCount--
				end += 2
			} else {
				end++
			}
		}

		if braceCount == 0 {
			// Extract the import path (without {{@ and }})
			importPath := content[idx+3 : end-2]
			imports = append(imports, strings.TrimSpace(importPath))
		}

		start = end
	}
	return imports
}

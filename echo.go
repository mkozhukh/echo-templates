package echotemplates

import (
	"github.com/mkozhukh/echo"
)

// CallOptions creates echo.CallOption slice from template metadata
func CallOptions(metadata map[string]any) []echo.CallOption {
	if metadata == nil {
		return nil
	}

	var opts []echo.CallOption

	// Add WithModel if model is defined
	if model, ok := metadata["model"].(string); ok && model != "" {
		opts = append(opts, echo.WithModel(model))
	}

	// Add WithTemperature if temperature is defined
	if temp, ok := metadata["temperature"].(float64); ok {
		opts = append(opts, echo.WithTemperature(float64(temp)))
	}

	// Add WithMaxTokens if max_tokens is defined
	if maxTokens, ok := metadata["max_tokens"].(int); ok {
		opts = append(opts, echo.WithMaxTokens(maxTokens))
	}

	return opts
}

func Extend(metadata map[string]any, content string) map[string]any {
	copy := make(map[string]any)
	for k, v := range metadata {
		copy[k] = v
	}
	copy["user_query"] = content

	return copy
}

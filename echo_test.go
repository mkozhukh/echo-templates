package echotemplates

import (
	"testing"

	"github.com/mkozhukh/echo"
)

func TestCallOptions(t *testing.T) {
	tests := []struct {
		name     string
		metadata map[string]any
		wantLen  int
		validate func(t *testing.T, opts []echo.CallOption)
	}{
		{
			name:     "nil metadata",
			metadata: nil,
			wantLen:  0,
		},
		{
			name:     "empty metadata",
			metadata: map[string]any{},
			wantLen:  0,
		},
		{
			name: "with model only",
			metadata: map[string]any{
				"model": "gpt-4",
			},
			wantLen: 1,
			validate: func(t *testing.T, opts []echo.CallOption) {
				// We can't directly inspect the options, but we can verify count
				if len(opts) != 1 {
					t.Errorf("Expected 1 option, got %d", len(opts))
				}
			},
		},
		{
			name: "with temperature only",
			metadata: map[string]any{
				"temperature": 0.7,
			},
			wantLen: 1,
		},
		{
			name: "with max_tokens only",
			metadata: map[string]any{
				"max_tokens": 1000,
			},
			wantLen: 1,
		},
		{
			name: "with all options",
			metadata: map[string]any{
				"model":       "gpt-4",
				"temperature": 0.8,
				"max_tokens":  2000,
			},
			wantLen: 3,
		},
		{
			name: "with extra fields",
			metadata: map[string]any{
				"model":       "claude-3",
				"temperature": 0.5,
				"max_tokens":  1500,
				"extra_field": "ignored",
				"another":     123,
			},
			wantLen: 3,
		},
		{
			name: "with wrong types",
			metadata: map[string]any{
				"model":       123,        // wrong type, should be ignored
				"temperature": "0.5",      // wrong type, should be ignored
				"max_tokens":  "thousand", // wrong type, should be ignored
			},
			wantLen: 0,
		},
		{
			name: "with empty model string",
			metadata: map[string]any{
				"model":       "",
				"temperature": 0.9,
			},
			wantLen: 1, // only temperature should be added
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := CallOptions(tt.metadata)

			if len(opts) != tt.wantLen {
				t.Errorf("CallOptions() returned %d options, want %d", len(opts), tt.wantLen)
			}

			if tt.validate != nil {
				tt.validate(t, opts)
			}
		})
	}
}

func TestCallOptionsIntegration(t *testing.T) {
	// Test that CallOptions works with actual template metadata
	metadata := map[string]any{
		"model":       "gpt-4-turbo",
		"temperature": 0.7,
		"max_tokens":  1000,
		"defaults": map[string]string{
			"role": "assistant",
		},
	}

	opts := CallOptions(metadata)

	// Should have 3 options (model, temperature, max_tokens)
	// defaults should be ignored
	if len(opts) != 3 {
		t.Errorf("Expected 3 options, got %d", len(opts))
	}
}

func TestCallOptionsTypeConversion(t *testing.T) {
	// Test various numeric type conversions that might come from JSON parsing
	tests := []struct {
		name     string
		metadata map[string]any
		wantLen  int
	}{
		{
			name: "float64 temperature",
			metadata: map[string]any{
				"temperature": float64(0.7),
			},
			wantLen: 1,
		},
		{
			name: "int max_tokens",
			metadata: map[string]any{
				"max_tokens": int(1000),
			},
			wantLen: 1,
		},
		{
			name: "int64 max_tokens should be ignored",
			metadata: map[string]any{
				"max_tokens": int64(1000),
			},
			wantLen: 0, // int64 is not handled, only int
		},
		{
			name: "float32 temperature should be ignored",
			metadata: map[string]any{
				"temperature": float32(0.7),
			},
			wantLen: 0, // float32 is not handled, only float64
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := CallOptions(tt.metadata)
			if len(opts) != tt.wantLen {
				t.Errorf("CallOptions() returned %d options, want %d", len(opts), tt.wantLen)
			}
		})
	}
}

func TestCallOptionsNilReturn(t *testing.T) {
	// Verify that nil metadata returns nil, not empty slice
	opts := CallOptions(nil)
	if opts != nil {
		t.Errorf("CallOptions(nil) = %v, want nil", opts)
	}

	// But empty metadata should return nil or empty slice
	opts = CallOptions(map[string]any{})
	if opts == nil {
		// This is fine
	} else if len(opts) != 0 {
		t.Errorf("CallOptions(empty) = %v, want nil or empty", opts)
	}
}

func TestCallOptionsRealWorld(t *testing.T) {
	// Test with a realistic metadata structure from a parsed template
	metadata := map[string]any{
		"model":       "claude-3-sonnet",
		"temperature": 0.3,
		"max_tokens":  4096,
		"top_p":       0.95,         // not supported yet, should be ignored
		"stream":      true,         // not supported yet, should be ignored
		"system":      "Be helpful", // not a call option, should be ignored
	}

	opts := CallOptions(metadata)

	// Should only process model, temperature, and max_tokens
	if len(opts) != 3 {
		t.Errorf("Expected 3 options from real-world metadata, got %d", len(opts))
	}
}

// Helper function to check if two slices of options are equal
// Note: This is a simplified check since we can't directly compare CallOption values
func optionsEqual(a, b []echo.CallOption) bool {
	return len(a) == len(b)
}

// Benchmark to ensure performance is reasonable
func BenchmarkCallOptions(b *testing.B) {
	metadata := map[string]any{
		"model":       "gpt-4",
		"temperature": 0.7,
		"max_tokens":  2000,
		"extra1":      "ignored",
		"extra2":      123,
		"extra3":      true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CallOptions(metadata)
	}
}

func BenchmarkCallOptionsEmpty(b *testing.B) {
	metadata := map[string]any{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CallOptions(metadata)
	}
}

func BenchmarkCallOptionsNil(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CallOptions(nil)
	}
}

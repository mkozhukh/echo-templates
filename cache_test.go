package echotemplates

import (
	"testing"
	"time"
)

func TestTemplateCache(t *testing.T) {
	cache := newTemplateCache(3)

	// Create test templates
	template1 := &parsedTemplate{
		metadata: map[string]any{"model": "gpt-4"},
		content:  "Template 1",
	}
	template2 := &parsedTemplate{
		metadata: map[string]any{"model": "claude"},
		content:  "Template 2",
	}
	template3 := &parsedTemplate{
		metadata: map[string]any{"model": "gemini"},
		content:  "Template 3",
	}
	template4 := &parsedTemplate{
		metadata: map[string]any{"model": "llama"},
		content:  "Template 4",
	}

	now := time.Now()

	// Test basic put and get
	cache.put("key1", template1, now)

	got, ok := cache.get("key1", now)
	if !ok {
		t.Error("Expected to find key1 in cache")
	}
	if got.content != template1.content {
		t.Errorf("Expected content %q, got %q", template1.content, got.content)
	}

	// Test cache miss
	_, ok = cache.get("nonexistent", now)
	if ok {
		t.Error("Expected cache miss for nonexistent key")
	}

	// Test file modification invalidation
	laterTime := now.Add(1 * time.Second)
	_, ok = cache.get("key1", laterTime)
	if ok {
		t.Error("Expected cache miss due to file modification")
	}

	// Test LRU eviction
	cache.put("key1", template1, now)
	cache.put("key2", template2, now)
	cache.put("key3", template3, now)

	// Access key1 to make it most recently used
	cache.get("key1", now)

	// Add key4, which should evict key2 (least recently used)
	cache.put("key4", template4, now)

	_, ok = cache.get("key2", now)
	if ok {
		t.Error("Expected key2 to be evicted")
	}

	// key1 should still be there
	_, ok = cache.get("key1", now)
	if !ok {
		t.Error("Expected key1 to still be in cache")
	}

	// Test clear
	cache.clear()
	_, ok = cache.get("key1", now)
	if ok {
		t.Error("Expected cache to be empty after clear")
	}
}

func TestCacheUpdate(t *testing.T) {
	cache := newTemplateCache(10)

	template1 := &parsedTemplate{
		metadata: map[string]any{"model": "gpt-4"},
		content:  "Original",
	}
	template2 := &parsedTemplate{
		metadata: map[string]any{"model": "gpt-4"},
		content:  "Updated",
	}

	now := time.Now()

	// Put original
	cache.put("key1", template1, now)

	// Update with new content
	cache.put("key1", template2, now)

	got, ok := cache.get("key1", now)
	if !ok {
		t.Error("Expected to find key1 in cache")
	}
	if got.content != template2.content {
		t.Errorf("Expected updated content %q, got %q", template2.content, got.content)
	}
}

func TestCacheConcurrency(t *testing.T) {
	cache := newTemplateCache(100)
	now := time.Now()

	// Run concurrent operations
	done := make(chan bool)

	// Writers
	for i := 0; i < 10; i++ {
		go func(id int) {
			template := &parsedTemplate{
				content: string(rune('A' + id)),
			}
			for j := 0; j < 100; j++ {
				cache.put(string(rune('A'+id)), template, now)
			}
			done <- true
		}(i)
	}

	// Readers
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				cache.get(string(rune('A'+id)), now)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Should not panic or deadlock
}

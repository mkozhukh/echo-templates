package echotemplates

import (
	"container/list"
	"sync"
	"time"
)

// cacheEntry represents a cached template
type cacheEntry struct {
	template    *parsedTemplate
	modTime     time.Time
	lastChecked time.Time
}

// templateCache implements an LRU cache for templates
type templateCache struct {
	mu        sync.RWMutex
	entries   map[string]*list.Element
	lru       *list.List
	maxSize   int
	checkFreq time.Duration
}

// cacheItem is what we store in the LRU list
type cacheItem struct {
	key   string
	entry *cacheEntry
}

// newTemplateCache creates a new template cache
func newTemplateCache(maxSize int) *templateCache {
	if maxSize <= 0 {
		maxSize = 100
	}
	return &templateCache{
		entries:   make(map[string]*list.Element),
		lru:       list.New(),
		maxSize:   maxSize,
		checkFreq: 5 * time.Second, // Check file modification every 5 seconds
	}
}

// get retrieves a template from cache if it exists and is still valid
func (c *templateCache) get(key string, fileModTime time.Time) (*parsedTemplate, bool) {
	c.mu.RLock()
	elem, exists := c.entries[key]
	c.mu.RUnlock()

	if !exists {
		return nil, false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	item := elem.Value.(*cacheItem)
	entry := item.entry

	// Check if file has been modified
	if fileModTime.After(entry.modTime) {
		// File has been modified, remove from cache
		c.lru.Remove(elem)
		delete(c.entries, key)
		return nil, false
	}

	// Move to front (most recently used)
	c.lru.MoveToFront(elem)
	entry.lastChecked = time.Now()

	return entry.template, true
}

// put adds or updates a template in the cache
func (c *templateCache) put(key string, template *parsedTemplate, modTime time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if already exists
	if elem, exists := c.entries[key]; exists {
		// Update existing entry
		item := elem.Value.(*cacheItem)
		item.entry.template = template
		item.entry.modTime = modTime
		item.entry.lastChecked = time.Now()
		c.lru.MoveToFront(elem)
		return
	}

	// Add new entry
	entry := &cacheEntry{
		template:    template,
		modTime:     modTime,
		lastChecked: time.Now(),
	}

	item := &cacheItem{
		key:   key,
		entry: entry,
	}

	elem := c.lru.PushFront(item)
	c.entries[key] = elem

	// Evict oldest if over capacity
	if c.lru.Len() > c.maxSize {
		oldest := c.lru.Back()
		if oldest != nil {
			oldItem := oldest.Value.(*cacheItem)
			c.lru.Remove(oldest)
			delete(c.entries, oldItem.key)
		}
	}
}

// clear removes all entries from the cache
func (c *templateCache) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*list.Element)
	c.lru = list.New()
}

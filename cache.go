package cache

import (
	"sort"
	"time"
)

// A Cache is a thread-safe store for fast item storage retrieval
type Cache struct {
	ops chan func(map[string]interface{})
}

// Returns an empty cache
func New() *Cache {
	c := &Cache{
		ops: make(chan func(map[string]interface{})),
	}

	go c.loop()
	return c
}

func (c *Cache) loop() {
	items := map[string]interface{}{}
	for op := range c.ops {
		op(items)
	}
}

// Adds an entry into the cache at the specified key.
// If an entry already exists at the specified key, it will be overwritten
func (c *Cache) Add(key string, val interface{}) {
	c.ops <- func(items map[string]interface{}) {
		items[key] = val
	}
}

// Removes all entries from the Cache
func (c *Cache) Clear() {
	c.ops <- func(items map[string]interface{}) {
		for key := range items {
			delete(items, key)
		}
	}
}

// Delete an entry from the cache at the specified key.
// If no entry exists at the specified key, no action is taken
func (c *Cache) Delete(key string) {
	c.ops <- func(items map[string]interface{}) {
		if _, ok := items[key]; ok {
			delete(items, key)
		}
	}
}

// todo: this doesn't clear entries after N time, clears eveyrtintg after N time
// should be clear on this
func (c *Cache) Expire(d time.Duration) *time.Ticker {
	ticker := time.NewTicker(d)
	go func() {
		for range ticker.C {
			c.Clear()
		}
	}()

	return ticker
}

// Retrieves an entry at the specified key
func (c *Cache) Get(key string) interface{} {
	result := make(chan interface{}, 1)
	c.ops <- func(items map[string]interface{}) {
		result <- items[key]
	}

	return <-result
}

// Retrieves an entry at the specified key.
// Returns bool specifying if the entry exists
func (c *Cache) Getf(key string) (interface{}, bool) {
	result := make(chan interface{}, 1)
	exists := make(chan bool, 1)
	c.ops <- func(items map[string]interface{}) {
		v, ok := items[key]
		result <- v
		exists <- ok
	}

	return <-result, <-exists
}

// Retrieves all entry in the cache
func (c *Cache) Items() map[string]interface{} {
	result := make(chan map[string]interface{}, 1)
	c.ops <- func(items map[string]interface{}) {
		result <- items
	}

	return <-result
}

// Retrieves a sorted list of all keys in the cache
func (c *Cache) Keys() []string {
	result := make(chan []string, 1)
	c.ops <- func(items map[string]interface{}) {
		keys := make([]string, 0, len(items))
		for k := range items {
			keys = append(keys, k)
		}

		sort.Strings(keys)
		result <- keys
	}

	return <-result
}
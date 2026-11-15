package livecoding

import (
	"testing"
	"time"
)

// §4.6 — expiring cache: per-entry TTL, lazy expiry on read,
// with an INJECTED clock so tests never sleep.

type cacheEntry struct {
	value     string
	expiresAt time.Time
}

// Cache is a per-entry TTL cache. `now` is injectable for tests;
// production passes time.Now.
type Cache struct {
	data map[string]cacheEntry
	now  func() time.Time
}

func NewCache(now func() time.Time) *Cache {
	return &Cache{data: map[string]cacheEntry{}, now: now}
}

func (c *Cache) Set(key, value string, ttl time.Duration) {
	c.data[key] = cacheEntry{value: value, expiresAt: c.now().Add(ttl)}
}

// Get is lazy: expired entries are dropped on access. This avoids a
// janitor goroutine (which would need its own stop channel and owner).
func (c *Cache) Get(key string) (string, bool) {
	e, ok := c.data[key]
	if !ok {
		return "", false
	}
	if c.now().After(e.expiresAt) {
		delete(c.data, key)
		return "", false
	}
	return e.value, true
}

func TestCacheHitAndExpiry(t *testing.T) {
	now := time.Unix(1000, 0)
	cache := NewCache(func() time.Time { return now }) // frozen clock, advanced manually

	cache.Set("city", "tampere", 10*time.Second)

	if v, ok := cache.Get("city"); !ok || v != "tampere" {
		t.Fatalf("Get before expiry = (%q, %v)", v, ok)
	}

	now = now.Add(11 * time.Second) // time travel past the TTL

	if _, ok := cache.Get("city"); ok {
		t.Error("entry should be expired")
	}
	// expired entry must have been evicted, not just hidden:
	if _, stillThere := cache.data["city"]; stillThere {
		t.Error("expired entry was not evicted")
	}
}

func TestCachePerEntryTTLIsolation(t *testing.T) {
	now := time.Unix(1000, 0)
	cache := NewCache(func() time.Time { return now })

	cache.Set("short", "a", 1*time.Second)
	cache.Set("long", "b", 1*time.Hour)

	now = now.Add(2 * time.Second)

	if _, ok := cache.Get("short"); ok {
		t.Error("short TTL should have expired")
	}
	if v, ok := cache.Get("long"); !ok || v != "b" {
		t.Errorf("long TTL should survive: (%q, %v)", v, ok)
	}
}

func TestCacheZeroTTLIsImmediatelyExpired(t *testing.T) {
	now := time.Unix(1000, 0)
	cache := NewCache(func() time.Time { return now })

	cache.Set("x", "v", 0) // expires at "now" exactly

	now = now.Add(time.Nanosecond)
	if _, ok := cache.Get("x"); ok {
		t.Error("zero TTL should expire immediately")
	}
}

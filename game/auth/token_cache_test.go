package auth

import (
	"testing"
	"time"
)

func (c *tokenCache) size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.tokens)
}

func TestTokenCacheTTL(t *testing.T) {
	base := time.Now()
	tests := []struct {
		name string
		ttl  time.Duration
		age  time.Duration
		want bool
	}{
		{"fresh entry", 50 * time.Millisecond, 0, true},
		{"entry just under ttl", 50 * time.Millisecond, 49 * time.Millisecond, true},
		{"entry exactly at ttl", 50 * time.Millisecond, 50 * time.Millisecond, false},
		{"entry past ttl", 50 * time.Millisecond, 60 * time.Millisecond, false},
		{"tiny ttl already stale", time.Nanosecond, time.Millisecond, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cache := newTokenCache(tc.ttl)
			defer cache.stop()

			cache.now = func() time.Time { return base }
			cache.put("player-1", "token-1")

			cache.now = func() time.Time { return base.Add(tc.age) }
			token, ok := cache.get("player-1")
			if ok != tc.want {
				t.Fatalf("get returned ok=%v, want %v", ok, tc.want)
			}
			if tc.want && token != "token-1" {
				t.Fatalf("expected cached token, got %q", token)
			}
		})
	}
}

func TestTokenCacheExpiresInRealTime(t *testing.T) {
	ttl := 30 * time.Millisecond
	cache := newTokenCache(ttl)
	defer cache.stop()

	cache.put("player-1", "token-1")
	if token, ok := cache.get("player-1"); !ok || token != "token-1" {
		t.Fatalf("expected fresh token, got %q (%v)", token, ok)
	}

	time.Sleep(2 * ttl)
	if _, ok := cache.get("player-1"); ok {
		t.Fatal("expected token to expire after the ttl")
	}
}

func TestTokenCacheEvictsStaleEntriesOnTick(t *testing.T) {
	ttl := 20 * time.Millisecond
	cache := newTokenCache(ttl)
	defer cache.stop()

	cache.put("player-1", "token-1")
	cache.put("player-2", "token-2")

	deadline := time.Now().Add(time.Second)
	for cache.size() > 0 {
		if time.Now().After(deadline) {
			t.Fatalf("expected background eviction, %d entries left", cache.size())
		}
		time.Sleep(ttl / 2)
	}
}

func TestTokenCachePutReplacesActiveToken(t *testing.T) {
	cache := newTokenCache(time.Minute)
	defer cache.stop()

	cache.put("player-1", "token-1")
	cache.put("player-1", "token-2")

	token, ok := cache.get("player-1")
	if !ok || token != "token-2" {
		t.Fatalf("expected latest token, got %q (%v)", token, ok)
	}
	if cache.size() != 1 {
		t.Fatalf("expected a single entry per player, got %d", cache.size())
	}
}

func TestTokenCacheEvictStaleKeepsFreshEntries(t *testing.T) {
	base := time.Now()
	cache := newTokenCache(50 * time.Millisecond)
	defer cache.stop()

	cache.now = func() time.Time { return base }
	cache.put("stale-player", "token-1")

	cache.now = func() time.Time { return base.Add(40 * time.Millisecond) }
	cache.put("fresh-player", "token-2")

	cache.now = func() time.Time { return base.Add(60 * time.Millisecond) }
	cache.evictStale()

	if _, ok := cache.get("stale-player"); ok {
		t.Fatal("expected stale entry to be evicted")
	}
	if token, ok := cache.get("fresh-player"); !ok || token != "token-2" {
		t.Fatalf("expected fresh entry to survive, got %q (%v)", token, ok)
	}
	if cache.size() != 1 {
		t.Fatalf("expected one remaining entry, got %d", cache.size())
	}
}

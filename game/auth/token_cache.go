package auth

import (
	"sync"
	"time"
)

var (
	TokenCacheTTL = 5 * time.Minute

	liveTokens = newTokenCache(TokenCacheTTL)
)

type cacheEntry struct {
	token    string
	storedAt time.Time
}

type tokenCache struct {
	mu     sync.RWMutex
	ttl    time.Duration
	timer  *time.Ticker
	tokens map[string]cacheEntry
	now    func() time.Time // used for syncing timings over time zones and tests
}

func newTokenCache(ttl time.Duration) *tokenCache {
	c := &tokenCache{
		ttl:    ttl,
		timer:  time.NewTicker(ttl),
		tokens: map[string]cacheEntry{},
		now:    time.Now,
	}
	go c.evictOnTick()
	return c
}

func (c *tokenCache) stop() {
	c.timer.Stop()
}

func (c *tokenCache) evictOnTick() {
	for range c.timer.C {
		c.evictStale()
	}
}

func (c *tokenCache) evictStale() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for playerID, entry := range c.tokens {
		if c.isStale(entry) {
			delete(c.tokens, playerID)
		}
	}
}

func (c *tokenCache) isStale(entry cacheEntry) bool {
	return c.now().Sub(entry.storedAt) >= c.ttl
}

func (c *tokenCache) put(playerID string, token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tokens[playerID] = cacheEntry{token: token, storedAt: c.now()}
}

func (c *tokenCache) get(playerID string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.tokens[playerID]
	if !ok || c.isStale(entry) {
		return "", false
	}
	return entry.token, true
}

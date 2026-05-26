package osv

import (
	"context"
	"sync"
	"time"

	"github.com/yemresalcan/mr-ducksafe/internal/matcher"
)

const defaultCacheTTL = 24 * time.Hour

// CachedClient bir Client'ı in-memory TTL cache ile sarmalayarak
// aynı (identifier, version) sorgusunun tekrar atılmasını önler.
type CachedClient struct {
	inner Client
	ttl   time.Duration

	mu      sync.RWMutex
	entries map[string]cacheEntry
	now     func() time.Time
}

type cacheEntry struct {
	vulns     []Vulnerability
	expiresAt time.Time
}

// NewCachedClient varsayılan 24 saat TTL ile cache wrapper döner.
func NewCachedClient(inner Client) *CachedClient {
	return &CachedClient{
		inner:   inner,
		ttl:     defaultCacheTTL,
		entries: make(map[string]cacheEntry),
		now:     time.Now,
	}
}

// Query önce cache'e bakar, miss/expired ise inner'a delege eder.
func (c *CachedClient) Query(ctx context.Context, id matcher.Identifier, version string) ([]Vulnerability, error) {
	key := id.Value + "@" + version

	c.mu.RLock()
	e, ok := c.entries[key]
	c.mu.RUnlock()
	if ok && c.now().Before(e.expiresAt) {
		return e.vulns, nil
	}

	vulns, err := c.inner.Query(ctx, id, version)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.entries[key] = cacheEntry{vulns: vulns, expiresAt: c.now().Add(c.ttl)}
	c.mu.Unlock()

	return vulns, nil
}

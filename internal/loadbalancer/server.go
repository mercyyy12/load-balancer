package loadbalancer

import (
	"net/http/httputil"
	"net/url"
	"sync"
)

// Backend represents a backend server with its URL, proxy, and health status
type Backend struct {
	URL   *url.URL
	Alive bool
	Proxy *httputil.ReverseProxy
	Mu    sync.RWMutex
}

// It locks the mutex so background health checks don't collide with active traffic.
func (b *Backend) SetAlive(alive bool) {
	b.Mu.Lock()
	b.Alive = alive
	b.Mu.Unlock()
}

// Uses a read-lock to allow multiple users to check status simultaneously.
func (b *Backend) IsAlive() bool {
	b.Mu.RLock()
	alive := b.Alive
	b.Mu.RUnlock()
	return alive
}

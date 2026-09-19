package loadbalancer

import (
	"log/slog"
	"net/http"
	"sync/atomic"
)

// Pool holds the list of all available backend servers
type Pool struct {
	Backends []*Backend
	Counter  uint64
}

// It routes every incoming user request to the healthy backend server
func (p *Pool) HomePage(w http.ResponseWriter, r *http.Request) {
	// Safely increment the counter using atomic operations
	// This prevents race conditions when thousands of users connect at the same time
	newCounter := atomic.AddUint64(&p.Counter, 1)

	// picking the next server in the slice (Round-Robin)
	proxy := p.Backends[int(newCounter)%len(p.Backends)]
	status := proxy.IsAlive()

	// If the chosen server is dead keep looping until we find a healthy one
	if status == false {
		counter := 0
		for !status {
			slog.Warn("Server is dead, skipping", "url", proxy.URL.String())

			// Try the next server in line
			newCounter += 1
			proxy = p.Backends[int(newCounter)%len(p.Backends)]
			status = proxy.IsAlive()

			//if we have checked every single server and they are all dead, break the loop
			counter++
			if counter > len(p.Backends) {
				break
			}
		}
	}

	// If the loop broke because the entire server is down return a error
	if status == false {
		http.Error(w, "All backend servers are currently offline!", http.StatusServiceUnavailable)
		return
	}

	// Forward the HTTP request to the healthy backend server
	proxy.Proxy.ServeHTTP(w, r)
}

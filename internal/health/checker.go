package health

import (
	"log/slog"
	"net"
	"time"

	"github.com/mercyyy12/load-balancer/internal/loadbalancer"
)

// CheckHealth runs in the background and pings all backend servers
// to ensure they are still online and accepting connections.
func CheckHealth(p *loadbalancer.Pool) {
	for _, v := range p.Backends {
		// Attempt to open a raw TCP connection to the server's port
		// If it takes longer than 2 seconds consider it dead
		conn, err := net.DialTimeout("tcp", v.URL.Host, 2*time.Second)
		if err != nil {
			slog.Error("Health check failed", "url", v.URL.String(), "err", err)
			v.SetAlive(false)
			continue
		}

		// The connection succeeded! Close it immediately and mark the server as healthy.
		conn.Close()
		v.SetAlive(true)
	}
}

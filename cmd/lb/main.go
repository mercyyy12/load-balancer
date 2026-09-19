package main

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/mercyyy12/load-balancer/internal/health"
	"github.com/mercyyy12/load-balancer/internal/loadbalancer"
)

func main() {
	var poolProxy loadbalancer.Pool

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Read backend URLs from environment variables
	backendUrlString := os.Getenv("BACKENDS")
	if backendUrlString == "" {
		backendUrlString = "http://localhost:8081,http://localhost:8082,http://localhost:8083"
	}
	rawUrls := strings.Split(backendUrlString, ",")

	// Initialize the Server Pool nd create a ReverseProxy for it
	poolProxy.Counter = 0
	for _, v := range rawUrls {
		target, err := url.Parse(v)
		if err != nil {
			slog.Error("Url parsing went wrong", "err", err)
			os.Exit(1)
		}

		backendPool := &loadbalancer.Backend{
			Alive: true, // Assume servers are healthy
			Proxy: httputil.NewSingleHostReverseProxy(target),
			URL:   target,
		}
		poolProxy.Backends = append(poolProxy.Backends, backendPool)
	}

	// Create the HTTP router for the Load Balancer
	mux := http.NewServeMux()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Route ALL incoming traffic ("/") to our Round-Robin load balancing logic
	mux.HandleFunc("/", poolProxy.HomePage)

	// a background Goroutine that acts as our Active Health Checker
	// wakes up every 10 seconds, pings the servers, and updates it's status
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			<-ticker.C
			health.CheckHealth(&poolProxy)
		}
	}()

	// Start the main Load Balancer web server in the background
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to start server", "err", err)
			os.Exit(1)
		}
	}()

	// Graceful Shutdown logic:
	// We freeze the main function here and wait until the OS sends an interrupt
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit // Blocks until interrupted

	// Once interrupted, give the server 5 seconds to finish any active user requests before dying
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}
	slog.Info("Server shutdown successfully")
}

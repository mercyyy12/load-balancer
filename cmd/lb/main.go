package main

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"time"
)

type Targets struct {
	Urls []string
}

type Pool struct {
	proxies []*httputil.ReverseProxy
	Counter int
}

func (p *Pool) homePage(w http.ResponseWriter, r *http.Request) {
	proxy := p.proxies[p.Counter%len(p.proxies)]
	p.Counter++
	proxy.ServeHTTP(w, r)
}

func main() {
	var poolProxy Pool

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	urls := &Targets{
		Urls: []string{"http://localhost:8081", "http://localhost:8082", "http://localhost:8083"},
	}
	poolProxy.Counter = 0
	for _, v := range urls.Urls {
		target, err := url.Parse(v)
		if err != nil {
			slog.Error("Url parsing went wrong", "err", err)
			os.Exit(1)
		}
		proxy := httputil.NewSingleHostReverseProxy(target)
		poolProxy.proxies = append(poolProxy.proxies, proxy)

	}
	mux := http.NewServeMux()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	mux.HandleFunc("/", poolProxy.homePage)

	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			slog.Error("Failed tp start a server", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}
	slog.Info("Server shutdown successfully")
}

package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// holds port configuration
type Backend struct {
	Port string
}

// data sent to client
type JSONResponse struct {
	Message string `json:"message"`
}

func (b *Backend) handleRoot(w http.ResponseWriter, r *http.Request) {
	resp := JSONResponse{
		Message: "Server started on port " + b.Port,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	backend := Backend{
		Port: port,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", backend.handleRoot)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	slog.Info("Backend server starting...", "port", port)

	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to start server", "err", err.Error())
			log.Fatal(err)
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

	slog.Info("Server shut down successfully")
}

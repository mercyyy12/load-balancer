package main

import (
	"context"
	"encoding/json"
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

	backend := Backend{
		Port: os.Getenv("PORT"),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", backend.handleRoot)

	srv := &http.Server{
		Addr:    ":" + backend.Port,
		Handler: mux,
	}

	slog.Info("Backend server starting...", "port", backend.Port)

	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to start server", "err", err.Error())
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

	slog.Info("Server shut down successfully")
}

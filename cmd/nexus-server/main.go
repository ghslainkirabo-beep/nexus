package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nexus/core/internal/handlers"
)

// HealthResponse defines the telemetry status payload
type HealthResponse struct {
	Status    string    `json:"status"`
	Service   string    `json:"service"`
	Timestamp time.Time `json:"timestamp"`
	Uptime    string    `json:"uptime"`
}

var startTime = time.Now()

func main() {
	mux := http.NewServeMux()

	// Health Telemetry Endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		response := HealthResponse{
			Status:    "ONLINE",
			Service:   "Nexus Core Engine (Go/Spotify Pillar)",
			Timestamp: time.Now().UTC(),
			Uptime:    time.Since(startTime).String(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("[ERROR] Failed to encode health response: %v", err)
		}
	})

	// Bimodal Authentication Gateway
	mux.HandleFunc("/api/v1/auth/register", handlers.HandleRegister)

	// Server Configuration with strict timeouts
	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to listen for graceful shutdown signals
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("[NEXUS] Core Engine active and listening on port 8080...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[FATAL] Server failed to start: %v", err)
		}
	}()

	// Block until shutdown signal is received
	sig := <-shutdownChan
	log.Printf("[NEXUS] Shutdown signal received (%v). Gracefully draining connections...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("[FATAL] Server forced to shutdown: %v", err)
	}

	log.Println("[NEXUS] Core Engine successfully shut down. Sovereign state preserved.")
}
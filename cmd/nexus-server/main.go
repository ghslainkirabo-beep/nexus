package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nexus/core/internal/handlers"
	"github.com/nexus/core/internal/streaming"
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
	// Initialize the Spotify Pillar Real-Time Streaming Hub
	hub := streaming.NewHub()
	go hub.Run()
	go hub.SimulatePulseGenerator()

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

	// Google Pillar Semantic Matching Endpoint
	mux.HandleFunc("/api/v1/intelligence/match", handlers.HandleSemanticMatch)

	// Real-Time Capability Pulse Stream (Server-Sent Events / Spotify Pillar Stream)
	mux.HandleFunc("/api/v1/stream/pulses", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported by client connection", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		clientChan := make(chan streaming.CapabilityPulse, 10)
		hub.Register <- clientChan
		defer func() {
			hub.Unregister <- clientChan
		}()

		notify := r.Context().Done()

		for {
			select {
			case <-notify:
				return
			case pulse := <-clientChan:
				data, err := json.Marshal(pulse)
				if err != nil {
					continue
				}
				fmt.Fprintf(w, "data: %s\n\n", string(data))
				flusher.Flush()
			}
		}
	})

	// Server Configuration with strict timeouts
	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 0, // Infinite for SSE streaming
		IdleTimeout:  120 * time.Second,
	}

	// Channel to listen for graceful shutdown signals
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("[NEXUS] Core Engine & Streaming Hub active and listening on port 8080...")
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
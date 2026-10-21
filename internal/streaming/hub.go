package streaming

import (
	"log"
	"sync"
	"time"
)

// CapabilityPulse represents a real-time professional capability match event
type CapabilityPulse struct {
	PulseID            string    `json:"pulse_id"`
	NodeID             string    `json:"node_id"`
	MatchedOpportunity string    `json:"matched_opportunity"`
	LatencyNs          int64     `json:"latency_ns"`
	Status             string    `json:"status"`
	Timestamp          time.Time `json:"timestamp"`
}

// Hub maintains the set of active client channels and broadcasts real-time pulses
type Hub struct {
	Clients    map[chan CapabilityPulse]bool
	Broadcast  chan CapabilityPulse
	Register   chan chan CapabilityPulse
	Unregister chan chan CapabilityPulse
	mu         sync.Mutex
}

// NewHub initializes our high-concurrency event streaming hub
func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan CapabilityPulse, 256),
		Register:   make(chan chan CapabilityPulse),
		Unregister: make(chan chan CapabilityPulse),
		Clients:    make(map[chan CapabilityPulse]bool),
	}
}

// Run starts the core event loop managing concurrent stream listeners
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client] = true
			h.mu.Unlock()
			log.Printf("[STREAM HUB] New client node connected to capability stream.")

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client)
			}
			h.mu.Unlock()
			log.Printf("[STREAM HUB] Client node disconnected from capability stream.")

		case pulse := <-h.Broadcast:
			h.mu.Lock()
			for client := range h.Clients {
				select {
				case client <- pulse:
				default:
					close(client)
					delete(h.Clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

// SimulatePulseGenerator periodically injects high-speed matching pulses into the sink (Simulating Google Pillar vector matches)
func (h *Hub) SimulatePulseGenerator() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		startTime := time.Now()
		pulse := CapabilityPulse{
			PulseID:            time.Now().Format("20060102-150405") + "-PULSE",
			NodeID:             "GLOBAL-VECTOR-ENGINE",
			MatchedOpportunity: "Autonomous Quantum System Architect - Kigali / Remote",
			LatencyNs:          time.Since(startTime).Nanoseconds(),
			Status:             "OPTIMAL_MATCH",
			Timestamp:          time.Now().UTC(),
		}
		h.Broadcast <- pulse
	}
}

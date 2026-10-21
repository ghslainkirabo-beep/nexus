package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/nexus/core/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// HandleRegister processes secure bimodal user registration
func HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate Email and Password
	if req.Email == "" || len(req.Password) < 8 {
		http.Error(w, "Email is required and password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	// Validate Role Bimodality
	var userRole models.Role
	switch req.Role {
	case "TALENT":
		userRole = models.RoleTalent
	case "ENTITY":
		userRole = models.RoleEntity
	default:
		http.Error(w, "Invalid role specified. Must be 'TALENT' or 'ENTITY'", http.StatusBadRequest)
		return
	}

	// Cryptographic Password Hashing (Cost factor 12)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		log.Printf("[ERROR] Failed to hash password: %v", err)
		http.Error(w, "Internal cryptographic error", http.StatusInternalServerError)
		return
	}

	// Construct User Node (In-memory storage for now, upgrading to CockroachDB in Phase 4)
	newUser := models.User{
		ID:           generateSecureUUID(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         userRole,
		CreatedAt:    time.Now().UTC(),
	}

	log.Printf("[NEXUS SEC] Registered new [%s] node: %s", newUser.Role, newUser.Email)

	// Respond with success and sanitized user profile
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		statusKey:   "SUCCESS",
		messageKey:  "Node registered successfully",
		"user":      newUser,
	})
}

const (
	statusKey  = "status"
	messageKey = "message"
)

// Simple pseudo-UUID generator for the initial core engine
func generateSecureUUID() string {
	return time.Now().Format("20060102150405") + "-NEXUS-NODE"
}
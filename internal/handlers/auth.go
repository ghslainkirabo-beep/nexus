package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nexus/core/internal/middleware"
	"github.com/nexus/core/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// HandleRegister processes secure bimodal user registration and issues a JWT token
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

	if req.Email == "" || len(req.Password) < 8 {
		http.Error(w, "Email is required and password must be at least 8 characters", http.StatusBadRequest)
		return
	}

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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		log.Printf("[ERROR] Failed to hash password: %v", err)
		http.Error(w, "Internal cryptographic error", http.StatusInternalServerError)
		return
	}

	newUser := models.User{
		ID:           generateSecureUUID(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         userRole,
		CreatedAt:    time.Now().UTC(),
	}

	// Mint JWT Token valid for 24 hours
	claims := middleware.JWTClaims{
		Email: newUser.Email,
		Role:  string(newUser.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(middleware.JwtSecret)
	if err != nil {
		log.Printf("[ERROR] Failed to sign token: %v", err)
		http.Error(w, "Failed to generate session token", http.StatusInternalServerError)
		return
	}

	log.Printf("[NEXUS SEC] Registered and minted token for [%s] node: %s", newUser.Role, newUser.Email)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "SUCCESS",
		"message": "Node registered successfully with active session",
		"token":   tokenString,
		"user":    newUser,
	})
}

func generateSecureUUID() string {
	return time.Now().Format("20060102150405") + "-NEXUS-NODE"
}
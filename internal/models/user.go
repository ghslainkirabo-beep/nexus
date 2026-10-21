package models

import (
	"time"
)

// Role defines whether the user is a Talent node or an Entity command
type Role string

const (
	RoleTalent Role = "TALENT"
	RoleEntity Role = "ENTITY"
)

// User represents a registered participant in the Nexus hyper-platform
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Never exposed in JSON responses
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

// RegisterRequest defines the incoming payload for user registration
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"` // "TALENT" or "ENTITY"
}
// Package admin provides platform-level user management functionality.
package admin

import (
	"time"

	"apprun/ent"
)

// ============================================================================
// API Response Types (Handler Layer)
// ============================================================================

// UserResponse represents user data in API responses
type UserResponse struct {
	ID        int64   `json:"id"`
	Email     string  `json:"email"`
	Nickname  *string `json:"nickname,omitempty"`
	Role      string  `json:"role"`
	IsActive  bool    `json:"is_active"`
	IsSystem  bool    `json:"is_system,omitempty"`
	CreatedAt string  `json:"created_at,omitempty"`
	UpdatedAt string  `json:"updated_at"`
}

// CreateUserResponse represents the response after creating a user
type CreateUserResponse struct {
	User              UserResponse `json:"user"`
	GeneratedPassword *string      `json:"generated_password,omitempty"` // Only returned when auto-generated
}

// ============================================================================
// Conversion Functions
// ============================================================================

// ToUserResponse converts Ent User entity to API response format
func ToUserResponse(user *ent.User, includeSystem, includeCreatedAt bool) UserResponse {
	var nickname *string
	if user.Nickname != "" {
		nickname = &user.Nickname
	}

	resp := UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Nickname:  nickname,
		Role:      user.Role,
		IsActive:  user.Status == 1,
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}

	if includeSystem {
		resp.IsSystem = user.IsSystem
	}
	if includeCreatedAt {
		resp.CreatedAt = user.CreatedAt.Format(time.RFC3339)
	}

	return resp
}

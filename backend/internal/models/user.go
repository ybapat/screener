package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleSeller UserRole = "seller"
	RoleBuyer  UserRole = "buyer"
	RoleAdmin  UserRole = "admin"
)

type User struct {
	ID                   uuid.UUID `json:"id"`
	Email                string    `json:"email"`
	PasswordHash         string    `json:"-"`
	DisplayName          string    `json:"display_name"`
	Role                 UserRole  `json:"role"`
	AgeRange             *string   `json:"age_range,omitempty"`
	Country              *string   `json:"country,omitempty"`
	Timezone             *string   `json:"timezone,omitempty"`
	CreditBalance        int64     `json:"credit_balance"`
	GlobalEpsilonBudget  float64   `json:"global_epsilon_budget"`
	EpsilonSpent         float64   `json:"epsilon_spent"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

func (u *User) EpsilonRemaining() float64 {
	return u.GlobalEpsilonBudget - u.EpsilonSpent
}

type RefreshToken struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	TokenHash string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	Revoked   bool      `json:"revoked"`
	CreatedAt time.Time `json:"created_at"`
}

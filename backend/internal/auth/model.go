package auth

import "time"

// LoginRequest adalah body request untuk endpoint login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// TokenResponse adalah response setelah login berhasil.
type TokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      UserInfo  `json:"user"`
}

// UserInfo adalah data user yang dikembalikan setelah login.
type UserInfo struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	Department string `json:"department,omitempty"`
	Position   string `json:"position,omitempty"`
	PhotoURL   string `json:"photo_url,omitempty"`
}

// Claims adalah JWT claims yang disimpan dalam token.
type Claims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CompanyID string `json:"company_id"`
}

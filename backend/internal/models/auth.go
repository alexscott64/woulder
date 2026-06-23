package models

import "time"

const (
	RoleAdmin     = "admin"
	RoleDeveloper = "developer"
	RoleViewer    = "viewer"
)

type User struct {
	ID           string     `json:"id"`
	Email        string     `json:"email"`
	DisplayName  string     `json:"display_name"`
	PasswordHash string     `json:"-"`
	Role         string     `json:"role"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
}

type AuthRefreshToken struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type CurrentUser struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

type AuthLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthRefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type AuthLogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type AuthResponse struct {
	User         CurrentUser `json:"user"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresAt    int64       `json:"expires_at"`
}

type AccessClaims struct {
	UserID    string
	Email     string
	Role      string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

func (u User) Current() CurrentUser {
	return CurrentUser{ID: u.ID, Email: u.Email, DisplayName: u.DisplayName, Role: u.Role}
}

func CanWriteMoney(role string) bool {
	return role == RoleAdmin || role == RoleDeveloper
}

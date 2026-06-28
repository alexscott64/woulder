package auth

import (
	"context"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

type Repository interface {
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	CreateUser(ctx context.Context, user models.User) (*models.User, error)
	UpdateUserCredentials(ctx context.Context, user models.User) (*models.User, error)
	UpdateLastLogin(ctx context.Context, userID string) error
	CreateRefreshToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) (*models.AuthRefreshToken, error)
	GetRefreshToken(ctx context.Context, tokenHash string) (*models.AuthRefreshToken, *models.User, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
}

package auth

import (
	"context"
	"time"

	"github.com/alexscott64/woulder/backend/internal/database/dberrors"
	"github.com/alexscott64/woulder/backend/internal/models"
)

type PostgresRepository struct {
	db DBConn
}

func NewPostgresRepository(db DBConn) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return r.scanUser(r.db.QueryRowContext(ctx, queryGetUserByEmail, email))
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	return r.scanUser(r.db.QueryRowContext(ctx, queryGetUserByID, id))
}

func (r *PostgresRepository) CreateUser(ctx context.Context, user models.User) (*models.User, error) {
	return r.scanUser(r.db.QueryRowContext(ctx, queryCreateUser, user.Email, user.DisplayName, user.PasswordHash, user.Role))
}

func (r *PostgresRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, queryUpdateLastLogin, userID)
	return err
}

func (r *PostgresRepository) CreateRefreshToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) (*models.AuthRefreshToken, error) {
	var rt models.AuthRefreshToken
	err := r.db.QueryRowContext(ctx, queryCreateRefreshToken, userID, tokenHash, expiresAt).Scan(
		&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.RevokedAt, &rt.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *PostgresRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*models.AuthRefreshToken, *models.User, error) {
	var rt models.AuthRefreshToken
	var u models.User
	err := r.db.QueryRowContext(ctx, queryGetRefreshToken, tokenHash).Scan(
		&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.RevokedAt, &rt.CreatedAt,
		&u.ID, &u.Email, &u.DisplayName, &u.PasswordHash, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt, &u.LastLoginAt,
	)
	if err != nil {
		return nil, nil, dberrors.WrapNotFound(err)
	}
	return &rt, &u, nil
}

func (r *PostgresRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := r.db.ExecContext(ctx, queryRevokeRefreshToken, tokenHash)
	return err
}

type userScanner interface {
	Scan(dest ...interface{}) error
}

func (r *PostgresRepository) scanUser(row userScanner) (*models.User, error) {
	var u models.User
	err := row.Scan(&u.ID, &u.Email, &u.DisplayName, &u.PasswordHash, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt, &u.LastLoginAt)
	if err != nil {
		return nil, dberrors.WrapNotFound(err)
	}
	return &u, nil
}

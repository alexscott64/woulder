package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/alexscott64/woulder/backend/internal/config"
	"github.com/alexscott64/woulder/backend/internal/database/auth"
	"github.com/alexscott64/woulder/backend/internal/database/dberrors"
	"github.com/alexscott64/woulder/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

var ErrAuthInvalidCredentials = errors.New("invalid credentials")
var ErrAuthInactiveUser = errors.New("inactive user")
var ErrAuthInvalidToken = errors.New("invalid token")

type AuthService struct {
	repo auth.Repository
	cfg  config.AuthConfig
}

func NewAuthService(repo auth.Repository, cfg config.AuthConfig) *AuthService {
	return &AuthService{repo: repo, cfg: cfg}
}

func (s *AuthService) BootstrapAdmin(ctx context.Context) error {
	if strings.TrimSpace(s.cfg.AdminEmail) == "" || s.cfg.AdminPassword == "" {
		return nil
	}
	_, err := s.repo.GetUserByEmail(ctx, s.cfg.AdminEmail)
	if err == nil {
		return nil
	}
	if !dberrors.IsNotFound(err) {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(s.cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.repo.CreateUser(ctx, models.User{Email: strings.ToLower(strings.TrimSpace(s.cfg.AdminEmail)), DisplayName: s.cfg.AdminDisplayName, PasswordHash: string(hash), Role: models.RoleAdmin})
	return err
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*models.AuthResponse, error) {
	u, err := s.repo.GetUserByEmail(ctx, strings.TrimSpace(email))
	if err != nil {
		return nil, ErrAuthInvalidCredentials
	}
	if !u.IsActive {
		return nil, ErrAuthInactiveUser
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return nil, ErrAuthInvalidCredentials
	}
	_ = s.repo.UpdateLastLogin(ctx, u.ID)
	return s.issueTokenPair(ctx, *u)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*models.AuthResponse, error) {
	hash := hashToken(refreshToken)
	rt, u, err := s.repo.GetRefreshToken(ctx, hash)
	if err != nil || rt.RevokedAt != nil || time.Now().After(rt.ExpiresAt) || !u.IsActive {
		return nil, ErrAuthInvalidToken
	}
	if err := s.repo.RevokeRefreshToken(ctx, hash); err != nil {
		return nil, err
	}
	return s.issueTokenPair(ctx, *u)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	if strings.TrimSpace(refreshToken) == "" {
		return nil
	}
	return s.repo.RevokeRefreshToken(ctx, hashToken(refreshToken))
}

func (s *AuthService) CurrentUser(ctx context.Context, id string) (*models.CurrentUser, error) {
	u, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !u.IsActive {
		return nil, ErrAuthInactiveUser
	}
	cur := u.Current()
	return &cur, nil
}

func (s *AuthService) ValidateAccessToken(token string) (*models.AccessClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrAuthInvalidToken
	}
	signed := parts[0] + "." + parts[1]
	expected := signHMAC([]byte(s.cfg.JWTSecret), signed)
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return nil, ErrAuthInvalidToken
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrAuthInvalidToken
	}
	var payload struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Role  string `json:"role"`
		Iat   int64  `json:"iat"`
		Exp   int64  `json:"exp"`
	}
	if json.Unmarshal(payloadBytes, &payload) != nil || payload.Sub == "" || payload.Exp <= time.Now().Unix() {
		return nil, ErrAuthInvalidToken
	}
	return &models.AccessClaims{UserID: payload.Sub, Email: payload.Email, Role: payload.Role, IssuedAt: time.Unix(payload.Iat, 0), ExpiresAt: time.Unix(payload.Exp, 0)}, nil
}

func (s *AuthService) issueTokenPair(ctx context.Context, u models.User) (*models.AuthResponse, error) {
	access, exp, err := s.makeAccessToken(u)
	if err != nil {
		return nil, err
	}
	refresh, err := randomToken(32)
	if err != nil {
		return nil, err
	}
	_, err = s.repo.CreateRefreshToken(ctx, u.ID, hashToken(refresh), time.Now().Add(time.Duration(s.cfg.RefreshTokenDays)*24*time.Hour))
	if err != nil {
		return nil, err
	}
	return &models.AuthResponse{User: u.Current(), AccessToken: access, RefreshToken: refresh, ExpiresAt: exp.Unix()}, nil
}

func (s *AuthService) makeAccessToken(u models.User) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(time.Duration(s.cfg.AccessTokenMinutes) * time.Minute)
	header, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	payload, _ := json.Marshal(map[string]interface{}{"sub": u.ID, "email": u.Email, "role": u.Role, "iat": now.Unix(), "exp": exp.Unix()})
	head := base64.RawURLEncoding.EncodeToString(header)
	body := base64.RawURLEncoding.EncodeToString(payload)
	signed := head + "." + body
	return signed + "." + signHMAC([]byte(s.cfg.JWTSecret), signed), exp, nil
}

func signHMAC(secret []byte, message string) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(message))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func randomToken(bytesLen int) (string, error) {
	b := make([]byte, bytesLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (s *AuthService) HashPasswordForTest(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}

func (s *AuthService) TokenDebugString() string {
	return fmt.Sprintf("access=%dm refresh=%dd", s.cfg.AccessTokenMinutes, s.cfg.RefreshTokenDays)
}

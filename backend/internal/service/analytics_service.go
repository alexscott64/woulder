package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alexscott64/woulder/backend/internal/database/analytics"
	"github.com/alexscott64/woulder/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// AnalyticsService handles analytics business logic.
type AnalyticsService struct {
	repo      analytics.Repository
	jwtSecret []byte
}

// NewAnalyticsService creates a new analytics service.
func NewAnalyticsService(repo analytics.Repository) *AnalyticsService {
	secret := os.Getenv("ANALYTICS_JWT_SECRET")
	if secret == "" {
		secret = "woulder-analytics-default-secret-change-me"
		log.Println("WARNING: ANALYTICS_JWT_SECRET not set, using default (insecure)")
	}

	svc := &AnalyticsService{
		repo:      repo,
		jwtSecret: []byte(secret),
	}

	// Ensure admin user exists on startup
	svc.ensureAdminUser()

	return svc
}

// ensureAdminUser creates/updates the admin user from environment variables.
func (s *AnalyticsService) ensureAdminUser() {
	username := os.Getenv("ANALYTICS_ADMIN_USERNAME")
	password := os.Getenv("ANALYTICS_ADMIN_PASSWORD")

	if username == "" || password == "" {
		log.Println("ANALYTICS_ADMIN_USERNAME or ANALYTICS_ADMIN_PASSWORD not set, skipping admin setup")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash admin password: %v", err)
		return
	}

	ctx := context.Background()
	if err := s.repo.UpsertAdmin(ctx, username, string(hash)); err != nil {
		log.Printf("Failed to upsert admin user: %v", err)
		return
	}

	log.Printf("Analytics admin user '%s' configured", username)
}

// --- Session operations ---

// CreateSession creates a new visitor session.
func (s *AnalyticsService) CreateSession(ctx context.Context, req *models.CreateSessionRequest, ipAddress string) error {
	session := &models.AnalyticsSession{
		SessionID:    req.SessionID,
		VisitorID:    req.VisitorID,
		IPAddress:    &ipAddress,
		UserAgent:    req.UserAgent,
		Referrer:     req.Referrer,
		DeviceType:   req.DeviceType,
		Browser:      req.Browser,
		OS:           req.OS,
		ScreenWidth:  req.ScreenWidth,
		ScreenHeight: req.ScreenHeight,
	}
	if err := s.repo.CreateSession(ctx, session); err != nil {
		return err
	}

	// Async IP geolocation lookup
	go s.lookupGeo(req.SessionID, ipAddress)

	return nil
}

// lookupGeo resolves IP address to country/region/city using ip-api.com (free, no key needed).
func (s *AnalyticsService) lookupGeo(sessionID, ipAddress string) {
	if ipAddress == "" || ipAddress == "127.0.0.1" || ipAddress == "::1" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,country,regionName,city", ipAddress)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("[analytics] geo lookup failed for %s: %v", ipAddress, err)
		return
	}
	defer resp.Body.Close()

	var result struct {
		Status     string `json:"status"`
		Country    string `json:"country"`
		RegionName string `json:"regionName"`
		City       string `json:"city"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return
	}
	if result.Status != "success" || result.Country == "" {
		return
	}

	if err := s.repo.UpdateSessionGeo(context.Background(), sessionID, result.Country, result.RegionName, result.City); err != nil {
		log.Printf("[analytics] failed to update geo for session %s: %v", sessionID, err)
	}
}

// Heartbeat updates session activity.
func (s *AnalyticsService) Heartbeat(ctx context.Context, sessionID string) error {
	return s.repo.UpdateSessionActivity(ctx, sessionID)
}

// --- Event operations ---

// TrackEvent records a single analytics event.
func (s *AnalyticsService) TrackEvent(ctx context.Context, req *models.TrackEventRequest) error {
	event := &models.AnalyticsEvent{
		SessionID: req.SessionID,
		EventType: req.EventType,
		EventName: req.EventName,
		PagePath:  req.PagePath,
		Metadata:  req.Metadata,
	}
	if err := s.repo.InsertEvent(ctx, event); err != nil {
		return err
	}
	// Update session activity after each event
	return s.repo.UpdateSessionActivity(ctx, req.SessionID)
}

// TrackBatchEvents records multiple events.
func (s *AnalyticsService) TrackBatchEvents(ctx context.Context, req *models.BatchEventsRequest) error {
	events := make([]models.AnalyticsEvent, len(req.Events))
	for i, e := range req.Events {
		events[i] = models.AnalyticsEvent{
			SessionID: req.SessionID,
			EventType: e.EventType,
			EventName: e.EventName,
			PagePath:  e.PagePath,
			Metadata:  e.Metadata,
		}
	}
	if err := s.repo.InsertEvents(ctx, events); err != nil {
		return err
	}
	// Update session activity after batch
	return s.repo.UpdateSessionActivity(ctx, req.SessionID)
}

// --- Auth operations ---

// Login validates credentials and returns a JWT token.
func (s *AnalyticsService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	user, err := s.repo.GetAdminByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("invalid credentials")
		}
		return nil, fmt.Errorf("auth error: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Update last login
	_ = s.repo.UpdateLastLogin(ctx, req.Username)

	// Generate JWT
	expiresAt := time.Now().Add(24 * time.Hour)
	token, err := s.generateJWT(req.Username, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &models.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt.Unix(),
	}, nil
}

// ValidateToken validates a JWT token and returns the username.
func (s *AnalyticsService) ValidateToken(tokenString string) (string, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid token format")
	}

	// Verify signature
	signingInput := parts[0] + "." + parts[1]
	expectedSig := s.signHMAC([]byte(signingInput))
	actualSig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return "", fmt.Errorf("invalid signature encoding")
	}

	if !hmac.Equal(expectedSig, actualSig) {
		return "", fmt.Errorf("invalid signature")
	}

	// Decode payload
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("invalid payload encoding")
	}

	var payload struct {
		Sub string `json:"sub"`
		Exp int64  `json:"exp"`
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return "", fmt.Errorf("invalid payload")
	}

	// Check expiry
	if time.Now().Unix() > payload.Exp {
		return "", fmt.Errorf("token expired")
	}

	return payload.Sub, nil
}

// generateJWT creates a simple HMAC-SHA256 JWT token.
func (s *AnalyticsService) generateJWT(username string, expiresAt time.Time) (string, error) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	payload, err := json.Marshal(map[string]interface{}{
		"sub": username,
		"iat": time.Now().Unix(),
		"exp": expiresAt.Unix(),
	})
	if err != nil {
		return "", err
	}
	payloadEncoded := base64.RawURLEncoding.EncodeToString(payload)

	signingInput := header + "." + payloadEncoded
	signature := base64.RawURLEncoding.EncodeToString(s.signHMAC([]byte(signingInput)))

	return signingInput + "." + signature, nil
}

// signHMAC signs data with HMAC-SHA256.
func (s *AnalyticsService) signHMAC(data []byte) []byte {
	mac := hmac.New(sha256.New, s.jwtSecret)
	mac.Write(data)
	return mac.Sum(nil)
}

// --- Metrics operations ---

// parsePeriod converts a period string to a time.Time "since" value.
func parsePeriod(period string) time.Time {
	switch period {
	case "7d":
		return time.Now().AddDate(0, 0, -7)
	case "30d":
		return time.Now().AddDate(0, 0, -30)
	case "90d":
		return time.Now().AddDate(0, 0, -90)
	case "all":
		return time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	default:
		return time.Now().AddDate(0, 0, -30)
	}
}

// GetOverviewMetrics returns high-level dashboard metrics.
func (s *AnalyticsService) GetOverviewMetrics(ctx context.Context, period string) (*models.OverviewMetrics, error) {
	return s.repo.GetOverviewMetrics(ctx, parsePeriod(period))
}

// GetVisitorsOverTime returns daily visitor data points.
func (s *AnalyticsService) GetVisitorsOverTime(ctx context.Context, period string) ([]models.VisitorDataPoint, error) {
	return s.repo.GetVisitorsOverTime(ctx, parsePeriod(period))
}

// GetTopPages returns the most viewed pages.
func (s *AnalyticsService) GetTopPages(ctx context.Context, period string, limit int) ([]models.TopPage, error) {
	return s.repo.GetTopPages(ctx, parsePeriod(period), limit)
}

// GetTopLocations returns the most viewed climbing locations.
func (s *AnalyticsService) GetTopLocations(ctx context.Context, period string, limit int) ([]models.TopLocation, error) {
	return s.repo.GetTopLocations(ctx, parsePeriod(period), limit)
}

// GetTopAreas returns the most viewed climbing areas.
func (s *AnalyticsService) GetTopAreas(ctx context.Context, period string, limit int) ([]models.TopArea, error) {
	return s.repo.GetTopAreas(ctx, parsePeriod(period), limit)
}

// GetTopRoutes returns the most viewed routes/boulders.
func (s *AnalyticsService) GetTopRoutes(ctx context.Context, period string, limit int) ([]models.TopRoute, error) {
	return s.repo.GetTopRoutes(ctx, parsePeriod(period), limit)
}

// GetFeatureUsage returns feature usage breakdown.
func (s *AnalyticsService) GetFeatureUsage(ctx context.Context, period string) ([]models.FeatureUsage, error) {
	return s.repo.GetFeatureUsage(ctx, parsePeriod(period))
}

// GetGeography returns visitor geographic distribution.
func (s *AnalyticsService) GetGeography(ctx context.Context, period string, limit int) ([]models.GeoLocation, error) {
	return s.repo.GetGeography(ctx, parsePeriod(period), limit)
}

// GetDeviceBreakdown returns device/browser/OS stats.
func (s *AnalyticsService) GetDeviceBreakdown(ctx context.Context, period string) (map[string][]models.DeviceBreakdown, error) {
	since := parsePeriod(period)
	devices, err := s.repo.GetDeviceBreakdown(ctx, since)
	if err != nil {
		return nil, err
	}
	browsers, err := s.repo.GetBrowserBreakdown(ctx, since)
	if err != nil {
		return nil, err
	}
	oses, err := s.repo.GetOSBreakdown(ctx, since)
	if err != nil {
		return nil, err
	}
	return map[string][]models.DeviceBreakdown{
		"devices":  devices,
		"browsers": browsers,
		"os":       oses,
	}, nil
}

// GetReferrers returns top referrer sources.
func (s *AnalyticsService) GetReferrers(ctx context.Context, period string, limit int) ([]models.ReferrerInfo, error) {
	return s.repo.GetReferrers(ctx, parsePeriod(period), limit)
}

// GetRecentSessions returns recent sessions with details.
func (s *AnalyticsService) GetRecentSessions(ctx context.Context, limit int) ([]models.SessionDetail, error) {
	return s.repo.GetRecentSessions(ctx, limit)
}

// GetSessionEvents returns events for a specific session.
func (s *AnalyticsService) GetSessionEvents(ctx context.Context, sessionID string) ([]models.AnalyticsEvent, error) {
	return s.repo.GetSessionEvents(ctx, sessionID)
}

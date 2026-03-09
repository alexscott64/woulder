package api

import (
	"log"
	"net/http"
	"strconv"

	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/gin-gonic/gin"
)

// --- Event Collection Endpoints (No Auth) ---

// CreateAnalyticsSession handles POST /api/analytics/session
func (h *Handler) CreateAnalyticsSession(c *gin.Context) {
	var req models.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.SessionID == "" || req.VisitorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id and visitor_id are required"})
		return
	}

	ipAddress := c.ClientIP()
	if err := h.analyticsService.CreateSession(c.Request.Context(), &req, ipAddress); err != nil {
		// Session might already exist (ON CONFLICT DO NOTHING), that's fine
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "created"})
}

// TrackAnalyticsEvents handles POST /api/analytics/events
func (h *Handler) TrackAnalyticsEvents(c *gin.Context) {
	var req models.BatchEventsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.SessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	if len(req.Events) == 0 {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "count": 0})
		return
	}

	if err := h.analyticsService.TrackBatchEvents(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "count": len(req.Events)})
}

// AnalyticsHeartbeat handles POST /api/analytics/heartbeat
func (h *Handler) AnalyticsHeartbeat(c *gin.Context) {
	var req models.HeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.SessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	if err := h.analyticsService.Heartbeat(c.Request.Context(), req.SessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update heartbeat"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// --- Auth Endpoints ---

// AnalyticsLogin handles POST /api/analytics/auth/login
func (h *Handler) AnalyticsLogin(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password are required"})
		return
	}

	resp, err := h.analyticsService.Login(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// --- Metrics Endpoints (Auth Required) ---

func getPeriod(c *gin.Context) string {
	period := c.Query("period")
	if period == "" {
		period = "30d"
	}
	return period
}

func getLimit(c *gin.Context, defaultLimit int) int {
	limitStr := c.Query("limit")
	if limitStr == "" {
		return defaultLimit
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		return defaultLimit
	}
	if limit > 100 {
		return 100
	}
	return limit
}

// GetAnalyticsOverview handles GET /api/analytics/metrics/overview
func (h *Handler) GetAnalyticsOverview(c *gin.Context) {
	metrics, err := h.analyticsService.GetOverviewMetrics(c.Request.Context(), getPeriod(c))
	if err != nil {
		log.Printf("[analytics] overview metrics error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get overview metrics"})
		return
	}
	c.JSON(http.StatusOK, metrics)
}

// GetAnalyticsVisitors handles GET /api/analytics/metrics/visitors
func (h *Handler) GetAnalyticsVisitors(c *gin.Context) {
	data, err := h.analyticsService.GetVisitorsOverTime(c.Request.Context(), getPeriod(c))
	if err != nil {
		log.Printf("[analytics] visitors error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get visitor data"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// GetAnalyticsPages handles GET /api/analytics/metrics/pages
func (h *Handler) GetAnalyticsPages(c *gin.Context) {
	data, err := h.analyticsService.GetTopPages(c.Request.Context(), getPeriod(c), getLimit(c, 20))
	if err != nil {
		log.Printf("[analytics] pages error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get page data"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// GetAnalyticsLocations handles GET /api/analytics/metrics/locations
func (h *Handler) GetAnalyticsLocations(c *gin.Context) {
	data, err := h.analyticsService.GetTopLocations(c.Request.Context(), getPeriod(c), getLimit(c, 20))
	if err != nil {
		log.Printf("[analytics] locations error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get location data"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// GetAnalyticsAreas handles GET /api/analytics/metrics/areas
func (h *Handler) GetAnalyticsAreas(c *gin.Context) {
	data, err := h.analyticsService.GetTopAreas(c.Request.Context(), getPeriod(c), getLimit(c, 20))
	if err != nil {
		log.Printf("[analytics] areas error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get area data"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// GetAnalyticsRoutes handles GET /api/analytics/metrics/routes
func (h *Handler) GetAnalyticsRoutes(c *gin.Context) {
	data, err := h.analyticsService.GetTopRoutes(c.Request.Context(), getPeriod(c), getLimit(c, 20))
	if err != nil {
		log.Printf("[analytics] routes error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get route data"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// GetAnalyticsFeatures handles GET /api/analytics/metrics/features
func (h *Handler) GetAnalyticsFeatures(c *gin.Context) {
	data, err := h.analyticsService.GetFeatureUsage(c.Request.Context(), getPeriod(c))
	if err != nil {
		log.Printf("[analytics] features error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get feature data"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// GetAnalyticsGeography handles GET /api/analytics/metrics/geography
func (h *Handler) GetAnalyticsGeography(c *gin.Context) {
	data, err := h.analyticsService.GetGeography(c.Request.Context(), getPeriod(c), getLimit(c, 50))
	if err != nil {
		log.Printf("[analytics] geography error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get geography data"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// GetAnalyticsDevices handles GET /api/analytics/metrics/devices
func (h *Handler) GetAnalyticsDevices(c *gin.Context) {
	data, err := h.analyticsService.GetDeviceBreakdown(c.Request.Context(), getPeriod(c))
	if err != nil {
		log.Printf("[analytics] devices error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get device data"})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetAnalyticsReferrers handles GET /api/analytics/metrics/referrers
func (h *Handler) GetAnalyticsReferrers(c *gin.Context) {
	data, err := h.analyticsService.GetReferrers(c.Request.Context(), getPeriod(c), getLimit(c, 20))
	if err != nil {
		log.Printf("[analytics] referrers error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get referrer data"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// GetAnalyticsSessions handles GET /api/analytics/sessions
func (h *Handler) GetAnalyticsSessions(c *gin.Context) {
	sessions, err := h.analyticsService.GetRecentSessions(c.Request.Context(), getLimit(c, 50))
	if err != nil {
		log.Printf("[analytics] sessions error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get sessions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": sessions})
}

// GetAnalyticsSessionEvents handles GET /api/analytics/sessions/:session_id/events
func (h *Handler) GetAnalyticsSessionEvents(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	events, err := h.analyticsService.GetSessionEvents(c.Request.Context(), sessionID)
	if err != nil {
		log.Printf("[analytics] session events error (session=%s): %v", sessionID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get session events"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": events})
}

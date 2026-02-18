package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// KayaLocationResponse represents a Kaya climbing location for API responses
type KayaLocationResponse struct {
	ID          int      `json:"id"`
	KayaID      string   `json:"kaya_id"`
	Slug        string   `json:"slug"`
	Name        string   `json:"name"`
	Latitude    *float64 `json:"latitude,omitempty"`
	Longitude   *float64 `json:"longitude,omitempty"`
	PhotoURL    *string  `json:"photo_url,omitempty"`
	Description *string  `json:"description,omitempty"`
	ClimbCount  int      `json:"climb_count"`
	AscentCount int      `json:"ascent_count"`
	LastSynced  *string  `json:"last_synced,omitempty"`
}

// KayaClimbResponse represents a Kaya climb for API responses
type KayaClimbResponse struct {
	ID          int      `json:"id"`
	KayaID      string   `json:"kaya_id"`
	Slug        string   `json:"slug"`
	Name        string   `json:"name"`
	Grade       *string  `json:"grade,omitempty"`
	ClimbType   *string  `json:"climb_type,omitempty"`
	Rating      *float64 `json:"rating,omitempty"`
	AscentCount int      `json:"ascent_count"`
	Location    *string  `json:"location,omitempty"`
	IsModerated bool     `json:"is_moderated"`
}

// KayaMatchResponse represents a Kaya <-> MP route match
type KayaMatchResponse struct {
	KayaClimbID    string   `json:"kaya_climb_id"`
	KayaClimbName  string   `json:"kaya_climb_name"`
	MPRouteID      int64    `json:"mp_route_id"`
	MPRouteName    string   `json:"mp_route_name"`
	Confidence     float64  `json:"confidence"`
	MatchType      string   `json:"match_type"`
	NameSimilarity *float64 `json:"name_similarity,omitempty"`
	DistanceKM     *float64 `json:"distance_km,omitempty"`
	IsVerified     bool     `json:"is_verified"`
}

// GetKayaLocations returns a list of Kaya locations
// GET /api/kaya/locations?limit=50&offset=0
func (h *Handler) GetKayaLocations(c *gin.Context) {
	// Parse query parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// TODO: Query database for locations using Kaya repository
	// query := `SELECT ... FROM woulder.kaya_locations ...`

	// For now, return placeholder data
	locations := []KayaLocationResponse{
		{
			ID:          1,
			KayaID:      "2958",
			Slug:        "Leavenworth-344933",
			Name:        "Leavenworth",
			ClimbCount:  1553,
			AscentCount: 510,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"locations": locations,
		"count":     len(locations),
		"limit":     limit,
		"offset":    offset,
	})
}

// GetKayaLocation returns a single Kaya location by slug
// GET /api/kaya/locations/:slug
func (h *Handler) GetKayaLocation(c *gin.Context) {
	slug := c.Param("slug")

	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug parameter required"})
		return
	}

	// TODO: Query database for location
	// query := `SELECT ... FROM woulder.kaya_locations WHERE slug = $1`

	// For now, return placeholder
	location := KayaLocationResponse{
		ID:          1,
		KayaID:      "2958",
		Slug:        slug,
		Name:        "Leavenworth",
		ClimbCount:  1553,
		AscentCount: 510,
	}

	c.JSON(http.StatusOK, location)
}

// GetKayaClimbs returns climbs for a location
// GET /api/kaya/locations/:slug/climbs?grade=V4&limit=50
func (h *Handler) GetKayaClimbs(c *gin.Context) {
	slug := c.Param("slug")

	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug parameter required"})
		return
	}

	// Parse filters
	grade := c.Query("grade")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	_ = grade // Will be used in filtering
	_ = limit

	// TODO: Query database for climbs
	// query := `SELECT ... FROM woulder.kaya_climbs WHERE ...`

	climbs := []KayaClimbResponse{
		{
			ID:          1,
			KayaID:      "12345",
			Slug:        "the-prism-v9",
			Name:        "The Prism",
			Grade:       strPtr("V9"),
			ClimbType:   strPtr("Boulder"),
			AscentCount: 45,
			IsModerated: true,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"climbs": climbs,
		"count":  len(climbs),
	})
}

// GetKayaMPMatches returns route matches between Kaya and MP
// GET /api/kaya/matches?location=Leavenworth&min_confidence=0.85
func (h *Handler) GetKayaMPMatches(c *gin.Context) {
	location := c.Query("location")
	minConfidence, _ := strconv.ParseFloat(c.DefaultQuery("min_confidence", "0.85"), 64)

	_ = location
	_ = minConfidence

	// TODO: Query database for matches
	// query := `SELECT ... FROM woulder.kaya_mp_route_matches WHERE ...`

	matches := []KayaMatchResponse{
		{
			KayaClimbID:   "12345",
			KayaClimbName: "The Prism",
			MPRouteID:     108123456,
			MPRouteName:   "The Prism",
			Confidence:    0.98,
			MatchType:     "exact_name",
			IsVerified:    false,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"matches": matches,
		"count":   len(matches),
	})
}

// GetKayaSyncStatus returns sync progress for all locations
// GET /api/kaya/sync/status
func (h *Handler) GetKayaSyncStatus(c *gin.Context) {
	// TODO: Query sync progress table
	// query := `SELECT ... FROM woulder.kaya_sync_progress ...`

	// Return placeholder
	status := gin.H{
		"total_locations_synced": 1,
		"total_climbs":           1553,
		"total_ascents":          510,
		"last_sync":              "2026-02-17T23:35:00Z",
		"sync_jobs_running":      0,
	}

	c.JSON(http.StatusOK, status)
}

func strPtr(s string) *string {
	return &s
}

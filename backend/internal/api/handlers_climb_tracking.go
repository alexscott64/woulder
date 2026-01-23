package api

import (
	"net/http"
	"strconv"

	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/gin-gonic/gin"
)

// RefreshClimbData triggers a manual refresh of climb data from Mountain Project
// POST /api/climbs/refresh
func (h *Handler) RefreshClimbData(c *gin.Context) {
	// Check if already syncing
	isSyncing, lastSync := h.climbTrackingService.GetSyncStatus()
	if isSyncing {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Sync already in progress",
			"message": "A sync operation is currently running. Please wait for it to complete.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Manual refresh endpoint not yet implemented. Use the sync_climbs script instead.",
		"last_sync": lastSync,
		"note":      "Run: cd backend && go run cmd/sync_climbs/main.go",
	})
}

// GetLastClimbedForLocation retrieves the most recent climb info for a specific location
// GET /api/climbs/location/:id
func (h *Handler) GetLastClimbedForLocation(c *gin.Context) {
	// Parse location ID from URL
	locationIDStr := c.Param("id")
	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid location ID"})
		return
	}

	// Fetch last climbed info
	lastClimbed, err := h.climbTrackingService.GetLastClimbedForLocation(c.Request.Context(), locationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve climb data"})
		return
	}

	// Return null if no data found
	if lastClimbed == nil {
		c.JSON(http.StatusOK, gin.H{"last_climbed_info": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"last_climbed_info": lastClimbed})
}

// GetAreasOrderedByActivity retrieves areas ordered by most recent climb activity
// GET /api/climbs/location/:id/areas
func (h *Handler) GetAreasOrderedByActivity(c *gin.Context) {
	// Parse location ID from URL
	locationIDStr := c.Param("id")
	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid location ID"})
		return
	}

	// Fetch areas ordered by activity
	areas, err := h.climbTrackingService.GetAreasOrderedByActivity(c.Request.Context(), locationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve area activity data"})
		return
	}

	// Return empty array if no data found
	if areas == nil {
		areas = []models.AreaActivitySummary{}
	}

	c.JSON(http.StatusOK, areas)
}

// GetSubareasOrderedByActivity retrieves subareas of a parent area ordered by recent climb activity
// GET /api/climbs/location/:id/areas/:area_id/subareas
func (h *Handler) GetSubareasOrderedByActivity(c *gin.Context) {
	// Parse location ID and area ID from URL
	locationIDStr := c.Param("id")
	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid location ID"})
		return
	}

	areaID := c.Param("area_id")
	if areaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Area ID is required"})
		return
	}

	// Fetch subareas ordered by activity
	subareas, err := h.climbTrackingService.GetSubareasOrderedByActivity(c.Request.Context(), areaID, locationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve subarea activity data"})
		return
	}

	// Return empty array if no data found
	if subareas == nil {
		subareas = []models.AreaActivitySummary{}
	}

	c.JSON(http.StatusOK, subareas)
}

// GetRoutesOrderedByActivity retrieves routes in an area ordered by recent climb activity
// GET /api/climbs/location/:id/areas/:area_id/routes?limit=50
func (h *Handler) GetRoutesOrderedByActivity(c *gin.Context) {
	// Parse location ID and area ID from URL
	locationIDStr := c.Param("id")
	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid location ID"})
		return
	}

	areaID := c.Param("area_id")
	if areaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Area ID is required"})
		return
	}

	// Parse optional limit query parameter (default 50, max 200)
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
			return
		}
		if parsedLimit < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Limit must be at least 1"})
			return
		}
		if parsedLimit > 200 {
			parsedLimit = 200
		}
		limit = parsedLimit
	}

	// Fetch routes ordered by activity
	routes, err := h.climbTrackingService.GetRoutesOrderedByActivity(c.Request.Context(), areaID, locationID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve route activity data"})
		return
	}

	// Return empty array if no data found
	if routes == nil {
		routes = []models.RouteActivitySummary{}
	}

	c.JSON(http.StatusOK, routes)
}

// GetRecentTicksForRoute retrieves recent ticks for a specific route
// GET /api/climbs/routes/:route_id/ticks?limit=5
func (h *Handler) GetRecentTicksForRoute(c *gin.Context) {
	// Parse route ID from URL
	routeID := c.Param("route_id")
	if routeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Route ID is required"})
		return
	}

	// Parse optional limit query parameter (default 5, max 20)
	limit := 5
	if limitStr := c.Query("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
			return
		}
		if parsedLimit < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Limit must be at least 1"})
			return
		}
		if parsedLimit > 20 {
			parsedLimit = 20
		}
		limit = parsedLimit
	}

	// Fetch recent ticks for route
	ticks, err := h.climbTrackingService.GetRecentTicksForRoute(c.Request.Context(), routeID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tick data"})
		return
	}

	// Return empty array if no data found
	if ticks == nil {
		ticks = []models.ClimbHistoryEntry{}
	}

	c.JSON(http.StatusOK, ticks)
}

// SearchInLocation searches all areas and routes in a location by name
// GET /api/climbs/location/:id/search-all?q=query&limit=50
func (h *Handler) SearchInLocation(c *gin.Context) {
	// Parse location ID from URL
	locationIDStr := c.Param("id")
	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid location ID"})
		return
	}

	// Get search query from query parameter
	searchQuery := c.Query("q")
	if searchQuery == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query 'q' is required"})
		return
	}

	// Parse optional limit query parameter (default 50, max 200)
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
			return
		}
		if parsedLimit < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Limit must be at least 1"})
			return
		}
		if parsedLimit > 200 {
			parsedLimit = 200
		}
		limit = parsedLimit
	}

	// Search both areas and routes
	results, err := h.climbTrackingService.SearchInLocation(c.Request.Context(), locationID, searchQuery, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search location"})
		return
	}

	// Return empty array if no data found
	if results == nil {
		results = []models.SearchResult{}
	}

	c.JSON(http.StatusOK, results)
}

// SearchRoutesInLocation searches all routes in a location by name, grade, or area
// GET /api/climbs/location/:id/search?q=query&limit=50
func (h *Handler) SearchRoutesInLocation(c *gin.Context) {
	// Parse location ID from URL
	locationIDStr := c.Param("id")
	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid location ID"})
		return
	}

	// Get search query from query parameter
	searchQuery := c.Query("q")
	if searchQuery == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query 'q' is required"})
		return
	}

	// Parse optional limit query parameter (default 50, max 200)
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
			return
		}
		if parsedLimit < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Limit must be at least 1"})
			return
		}
		if parsedLimit > 200 {
			parsedLimit = 200
		}
		limit = parsedLimit
	}

	// Search routes
	routes, err := h.climbTrackingService.SearchRoutesInLocation(c.Request.Context(), locationID, searchQuery, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search routes"})
		return
	}

	// Return empty array if no data found
	if routes == nil {
		routes = []models.RouteActivitySummary{}
	}

	c.JSON(http.StatusOK, routes)
}


// GetBoulderDryingStatus calculates boulder-specific drying status
// GET /api/climbs/routes/:route_id/drying-status
func (h *Handler) GetBoulderDryingStatus(c *gin.Context) {
	// Parse route ID from URL
	routeID := c.Param("route_id")
	if routeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Route ID is required"})
		return
	}

	// Calculate boulder drying status
	status, err := h.boulderDryingService.GetBoulderDryingStatus(c.Request.Context(), routeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate boulder drying status", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetAreaDryingStats calculates aggregated drying statistics for an area
// GET /api/climbs/location/:id/areas/:area_id/drying-stats
func (h *Handler) GetAreaDryingStats(c *gin.Context) {
	// Parse location ID from URL
	locationIDStr := c.Param("id")
	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid location ID"})
		return
	}

	// Parse area ID from URL
	areaID := c.Param("area_id")
	if areaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Area ID is required"})
		return
	}

	// Calculate area drying stats
	stats, err := h.boulderDryingService.GetAreaDryingStats(c.Request.Context(), areaID, locationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate area drying stats", "details": err.Error()})
		return
	}

	// Return null if no routes with GPS data
	if stats == nil {
		c.JSON(http.StatusOK, gin.H{"message": "No routes with GPS data found for this area"})
		return
	}

	c.JSON(http.StatusOK, stats)
}


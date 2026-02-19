package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

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

	// Fetch areas ordered by activity (based on MP data)
	areas, err := h.climbTrackingService.GetAreasOrderedByActivity(c.Request.Context(), locationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve area activity data"})
		return
	}

	// Return empty array if no data found
	if areas == nil {
		areas = []models.AreaActivitySummary{}
	}

	// Update area activity with Kaya data if newer
	// For each area, check if any matched Kaya climbs have more recent activity
	for i := range areas {
		area := &areas[i]

		// Fetch matched Kaya climbs for this area
		kayaClimbs, err := h.kayaRepo.Climbs().GetMatchedClimbsForArea(c.Request.Context(), area.MPAreaID, 1)
		if err != nil || len(kayaClimbs) == 0 {
			continue // Skip if no Kaya data or error
		}

		// If Kaya has more recent activity, update the area's last activity
		if kayaClimbs[0].LastClimbAt.After(area.LastClimbAt) {
			area.LastClimbAt = kayaClimbs[0].LastClimbAt
			area.DaysSinceClimb = kayaClimbs[0].DaysSinceClimb
		}
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

	areaIDStr := c.Param("area_id")
	if areaIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Area ID is required"})
		return
	}

	areaID, err := strconv.ParseInt(areaIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid area ID"})
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

	areaIDStr := c.Param("area_id")
	if areaIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Area ID is required"})
		return
	}

	areaID, err := strconv.ParseInt(areaIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid area ID"})
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

// GetUnifiedRoutesOrderedByActivity retrieves both MP routes and Kaya climbs ordered by activity
// GET /api/climbs/location/:id/areas/:area_id/unified-routes?limit=200
func (h *Handler) GetUnifiedRoutesOrderedByActivity(c *gin.Context) {
	// Parse location ID and area ID from URL
	locationIDStr := c.Param("id")
	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid location ID"})
		return
	}

	areaIDStr := c.Param("area_id")
	if areaIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Area ID is required"})
		return
	}

	areaID, err := strconv.ParseInt(areaIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid area ID"})
		return
	}

	// Parse optional limit query parameter (default 200, max 200)
	limit := 200
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

	// Fetch MP routes for this specific area
	mpRoutes, err := h.climbTrackingService.GetRoutesOrderedByActivity(c.Request.Context(), areaID, locationID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve MP route activity data"})
		return
	}

	// Convert MP routes to unified format
	unifiedRoutes := make([]models.UnifiedRouteActivitySummary, 0, len(mpRoutes)+50)
	for _, route := range mpRoutes {
		mpRouteID := route.MPRouteID
		mpAreaID := route.MPAreaID
		unified := models.UnifiedRouteActivitySummary{
			ID:             fmt.Sprintf("mp-%d", route.MPRouteID),
			Name:           route.Name,
			Rating:         route.Rating,
			AreaName:       "", // Will be populated from most recent tick if available
			LastClimbAt:    route.LastClimbAt,
			DaysSinceClimb: route.DaysSinceClimb,
			Source:         "mp",
			MPRouteID:      &mpRouteID,
			MPAreaID:       &mpAreaID,
			MostRecentTick: route.MostRecentTick,
		}
		if route.MostRecentTick != nil {
			unified.AreaName = route.MostRecentTick.AreaName
		}
		unifiedRoutes = append(unifiedRoutes, unified)
	}

	// Create a map of MP route ID -> index in unifiedRoutes for quick lookup
	mpRouteMap := make(map[int64]int)
	for i, route := range unifiedRoutes {
		if route.MPRouteID != nil {
			mpRouteMap[*route.MPRouteID] = i
		}
	}

	// Fetch Kaya climbs that have been matched to MP routes in this area
	kayaClimbs, err := h.kayaRepo.Climbs().GetMatchedClimbsForArea(c.Request.Context(), areaID, limit)
	if err != nil {
		// Log error but continue - Kaya data is optional
		fmt.Printf("Warning: Failed to fetch matched Kaya climbs: %v\n", err)
	} else {
		for _, kayaClimb := range kayaClimbs {
			if kayaClimb.MPRouteID != nil {
				// If this Kaya climb matches an existing MP route, update the MP route
				// with the more recent activity if Kaya has newer data
				if idx, exists := mpRouteMap[*kayaClimb.MPRouteID]; exists {
					mpRoute := &unifiedRoutes[idx]
					// Update if Kaya has more recent activity
					if kayaClimb.LastClimbAt.After(mpRoute.LastClimbAt) {
						mpRoute.LastClimbAt = kayaClimb.LastClimbAt
						mpRoute.DaysSinceClimb = kayaClimb.DaysSinceClimb
						// Keep the MP route but note that most recent activity is from Kaya
						if kayaClimb.MostRecentAscent != nil {
							// We could add a field to indicate latest source, but for now just update the date
						}
					}
					// Don't add as separate entry - we updated the existing MP route
					continue
				}
			}
			// Add Kaya climbs that don't match any MP route
			unifiedRoutes = append(unifiedRoutes, kayaClimb)
		}
	}

	// Sort all results by most recent activity
	// Sort in place by LastClimbAt descending
	for i := 0; i < len(unifiedRoutes); i++ {
		for j := i + 1; j < len(unifiedRoutes); j++ {
			if unifiedRoutes[j].LastClimbAt.After(unifiedRoutes[i].LastClimbAt) {
				unifiedRoutes[i], unifiedRoutes[j] = unifiedRoutes[j], unifiedRoutes[i]
			}
		}
	}

	// Trim to limit
	if len(unifiedRoutes) > limit {
		unifiedRoutes = unifiedRoutes[:limit]
	}

	c.JSON(http.StatusOK, unifiedRoutes)
}

// GetRecentTicksForRoute retrieves recent ticks for a specific route
// GET /api/climbs/routes/:route_id/ticks?limit=5
func (h *Handler) GetRecentTicksForRoute(c *gin.Context) {
	// Parse route ID from URL
	routeIDStr := c.Param("route_id")
	if routeIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Route ID is required"})
		return
	}

	routeID, err := strconv.ParseInt(routeIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid route ID"})
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

	// Fetch recent MP ticks for route
	mpTicks, err := h.climbTrackingService.GetRecentTicksForRoute(c.Request.Context(), routeID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tick data"})
		return
	}

	// Initialize combined ticks with MP data and set source
	allTicks := []models.ClimbHistoryEntry{}
	if mpTicks != nil {
		// Set source field for all MP ticks
		for i := range mpTicks {
			mpTicks[i].Source = "mp"
		}
		allTicks = mpTicks
	}

	// Try to fetch Kaya ascents for matched route
	kayaAscents, err := h.kayaRepo.Ascents().GetAscentsForMatchedRoute(c.Request.Context(), routeID, limit)
	if err != nil {
		// Log but don't fail - Kaya data is optional
		fmt.Printf("Warning: Failed to fetch Kaya ascents for route %d: %v\n", routeID, err)
	} else {
		// Convert Kaya ascents to ClimbHistoryEntry format and add them
		// Populate ALL fields so frontend doesn't need to handle source differences
		for _, ascent := range kayaAscents {
			daysSince := int(time.Since(ascent.Date).Hours() / 24)

			entry := models.ClimbHistoryEntry{
				MPRouteID:      routeID, // Use the MP route ID since they're matched
				RouteName:      ascent.ClimbName,
				RouteRating:    *ascent.ClimbGrade,
				MPAreaID:       0, // We don't have MP area ID for Kaya
				AreaName:       ascent.AreaName,
				ClimbedAt:      ascent.Date,
				ClimbedBy:      ascent.Username,
				Style:          "", // Kaya doesn't have style (Flash/Redpoint/etc)
				Comment:        ascent.Comment,
				DaysSinceClimb: daysSince,
				Source:         "kaya",
			}
			allTicks = append(allTicks, entry)
		}
	}

	// Sort combined ticks by date descending
	for i := 0; i < len(allTicks); i++ {
		for j := i + 1; j < len(allTicks); j++ {
			if allTicks[j].ClimbedAt.After(allTicks[i].ClimbedAt) {
				allTicks[i], allTicks[j] = allTicks[j], allTicks[i]
			}
		}
	}

	// Trim to limit after combining
	if len(allTicks) > limit {
		allTicks = allTicks[:limit]
	}

	c.JSON(http.StatusOK, allTicks)
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

// GetBatchBoulderDryingStatus calculates boulder-specific drying status for multiple routes
// GET /api/climbs/routes/batch-drying-status?route_ids=id1,id2,id3
func (h *Handler) GetBatchBoulderDryingStatus(c *gin.Context) {
	// Parse route IDs from query parameter
	routeIDsStr := c.Query("route_ids")
	if routeIDsStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "route_ids query parameter is required"})
		return
	}

	// Split comma-separated route IDs and convert to int64
	routeIDs := []int64{}
	for _, id := range strings.Split(routeIDsStr, ",") {
		trimmed := strings.TrimSpace(id)
		if trimmed != "" {
			routeID, err := strconv.ParseInt(trimmed, 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid route ID: %s", trimmed)})
				return
			}
			routeIDs = append(routeIDs, routeID)
		}
	}

	if len(routeIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one route ID is required"})
		return
	}

	// Limit batch size to prevent abuse
	if len(routeIDs) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Maximum 200 route IDs allowed per batch request"})
		return
	}

	// Calculate boulder drying statuses in batch
	statuses, err := h.boulderDryingService.GetBatchBoulderDryingStatus(c.Request.Context(), routeIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate boulder drying statuses", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, statuses)
}

// GetBoulderDryingStatus calculates boulder-specific drying status
// GET /api/climbs/routes/:route_id/drying-status
func (h *Handler) GetBoulderDryingStatus(c *gin.Context) {
	// Parse route ID from URL
	routeIDStr := c.Param("route_id")
	if routeIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Route ID is required"})
		return
	}

	routeID, err := strconv.ParseInt(routeIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid route ID"})
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
	areaIDStr := c.Param("area_id")
	if areaIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Area ID is required"})
		return
	}

	areaID, err := strconv.ParseInt(areaIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid area ID"})
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

// GetBatchAreaDryingStats calculates drying statistics for multiple areas in a single call
// GET /api/climbs/location/:id/batch-area-drying-stats?area_ids=id1,id2,id3
func (h *Handler) GetBatchAreaDryingStats(c *gin.Context) {
	// Parse location ID from URL
	locationIDStr := c.Param("id")
	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid location ID"})
		return
	}

	// Parse area IDs from query parameter
	areaIDsStr := c.Query("area_ids")
	if areaIDsStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "area_ids query parameter is required"})
		return
	}

	// Split comma-separated area IDs and convert to int64
	areaIDs := []int64{}
	for _, id := range strings.Split(areaIDsStr, ",") {
		trimmed := strings.TrimSpace(id)
		if trimmed != "" {
			areaID, err := strconv.ParseInt(trimmed, 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid area ID: %s", trimmed)})
				return
			}
			areaIDs = append(areaIDs, areaID)
		}
	}

	if len(areaIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one area ID is required"})
		return
	}

	// Limit batch size to prevent abuse
	if len(areaIDs) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Maximum 100 area IDs allowed per batch request"})
		return
	}

	// Calculate area drying stats in batch
	stats, err := h.boulderDryingService.GetBatchAreaDryingStats(c.Request.Context(), areaIDs, locationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate batch area drying stats", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

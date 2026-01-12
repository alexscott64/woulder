package api

import (
	"net/http"
	"strconv"

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

package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alexscott64/woulder/backend/internal/database"
	"github.com/gin-gonic/gin"
)

// GetHeatMapActivity returns aggregated climbing activity for the heat map
// GET /api/heat-map/activity?start_date=2024-01-01&end_date=2024-12-31&min_lat=...&max_lat=...&min_lon=...&max_lon=...&min_activity=5&limit=500
func (h *Handler) GetHeatMapActivity(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse date range (required)
	startDateStr := c.Query("start_date")
	if startDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "start_date is required (format: YYYY-MM-DD)",
		})
		return
	}

	endDateStr := c.Query("end_date")
	if endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "end_date is required (format: YYYY-MM-DD)",
		})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid start_date format (use YYYY-MM-DD)",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid end_date format (use YYYY-MM-DD)",
		})
		return
	}

	// Parse optional geographic bounds
	var bounds *database.GeoBounds
	if c.Query("min_lat") != "" {
		minLat, err1 := strconv.ParseFloat(c.Query("min_lat"), 64)
		maxLat, err2 := strconv.ParseFloat(c.Query("max_lat"), 64)
		minLon, err3 := strconv.ParseFloat(c.Query("min_lon"), 64)
		maxLon, err4 := strconv.ParseFloat(c.Query("max_lon"), 64)

		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid bounds parameters. All 4 bounds required: min_lat, max_lat, min_lon, max_lon",
			})
			return
		}

		bounds = &database.GeoBounds{
			MinLat: minLat, MaxLat: maxLat,
			MinLon: minLon, MaxLon: maxLon,
		}
	}

	// Parse optional filters
	minActivity := 1
	if val := c.Query("min_activity"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			minActivity = parsed
		}
	}

	limit := 500
	if val := c.Query("limit"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			limit = parsed
			if limit > 1000 {
				limit = 1000 // Cap at 1000
			}
		}
	}

	// Fetch heat map data
	points, err := h.heatMapService.GetHeatMapData(ctx, startDate, endDate, bounds, minActivity, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch heat map data",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"points": points,
		"count":  len(points),
		"filters": gin.H{
			"start_date":   startDate.Format("2006-01-02"),
			"end_date":     endDate.Format("2006-01-02"),
			"min_activity": minActivity,
			"limit":        limit,
		},
	})
}

// GetHeatMapAreaDetail returns detailed activity for a specific area
// GET /api/heat-map/area/:area_id/detail?start_date=...&end_date=...
func (h *Handler) GetHeatMapAreaDetail(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse area ID
	areaIDStr := c.Param("area_id")
	areaID, err := strconv.ParseInt(areaIDStr, 10, 64)
	if err != nil || areaID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid area ID",
		})
		return
	}

	// Parse date range
	startDateStr := c.Query("start_date")
	if startDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "start_date is required (format: YYYY-MM-DD)",
		})
		return
	}

	endDateStr := c.Query("end_date")
	if endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "end_date is required (format: YYYY-MM-DD)",
		})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid start_date format (use YYYY-MM-DD)",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid end_date format (use YYYY-MM-DD)",
		})
		return
	}

	// Fetch detailed activity
	detail, err := h.heatMapService.GetAreaActivityDetail(ctx, areaID, startDate, endDate)
	if err != nil {
		if strings.Contains(err.Error(), "area not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Area not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch area detail",
		})
		return
	}

	c.JSON(http.StatusOK, detail)
}

// GetHeatMapRoutes returns routes within bounds with activity
// GET /api/heat-map/routes?min_lat=...&max_lat=...&min_lon=...&max_lon=...&start_date=...&end_date=...&limit=100
func (h *Handler) GetHeatMapRoutes(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse bounds (required)
	minLatStr := c.Query("min_lat")
	maxLatStr := c.Query("max_lat")
	minLonStr := c.Query("min_lon")
	maxLonStr := c.Query("max_lon")

	if minLatStr == "" || maxLatStr == "" || minLonStr == "" || maxLonStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "All bounds parameters required: min_lat, max_lat, min_lon, max_lon",
		})
		return
	}

	minLat, err1 := strconv.ParseFloat(minLatStr, 64)
	maxLat, err2 := strconv.ParseFloat(maxLatStr, 64)
	minLon, err3 := strconv.ParseFloat(minLonStr, 64)
	maxLon, err4 := strconv.ParseFloat(maxLonStr, 64)

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid bounds parameters (must be valid floats)",
		})
		return
	}

	bounds := database.GeoBounds{
		MinLat: minLat, MaxLat: maxLat,
		MinLon: minLon, MaxLon: maxLon,
	}

	// Parse date range
	startDate, _ := time.Parse("2006-01-02", c.Query("start_date"))
	endDate, _ := time.Parse("2006-01-02", c.Query("end_date"))

	// If dates not provided, use last 30 days
	if startDate.IsZero() || endDate.IsZero() {
		endDate = time.Now()
		startDate = endDate.AddDate(0, 0, -30)
	}

	// Parse limit
	limit := 100
	if val := c.Query("limit"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			limit = parsed
			if limit > 500 {
				limit = 500
			}
		}
	}

	// Fetch routes
	routes, err := h.heatMapService.GetRoutesByBounds(ctx, bounds, startDate, endDate, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch routes",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"routes": routes,
		"count":  len(routes),
	})
}

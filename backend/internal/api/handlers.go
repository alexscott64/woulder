package api

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/alexscott64/woulder/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	locationService      *service.LocationService
	weatherService       *service.WeatherService
	riverService         *service.RiverService
	climbTrackingService *service.ClimbTrackingService
	boulderDryingService *service.BoulderDryingService
}

func NewHandler(
	locationService *service.LocationService,
	weatherService *service.WeatherService,
	riverService *service.RiverService,
	climbTrackingService *service.ClimbTrackingService,
	boulderDryingService *service.BoulderDryingService,
) *Handler {
	return &Handler{
		locationService:      locationService,
		weatherService:       weatherService,
		riverService:         riverService,
		climbTrackingService: climbTrackingService,
		boulderDryingService: boulderDryingService,
	}
}

// StartBackgroundRefresh starts a goroutine that refreshes weather data periodically
func (h *Handler) StartBackgroundRefresh(interval time.Duration) {
	// Start periodic refresh using weather service
	h.weatherService.StartBackgroundRefresh(interval)
	log.Printf("Background weather refresh scheduled every %v", interval)
}

// StartBackgroundTickSync starts a goroutine that syncs new Mountain Project ticks periodically
func (h *Handler) StartBackgroundTickSync(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		log.Printf("Starting background tick sync scheduler (every %v)", interval)

		// Run immediately on startup
		log.Println("Running initial tick sync...")
		ctx := context.Background()
		if err := h.climbTrackingService.SyncNewTicksForAllLocations(ctx); err != nil {
			log.Printf("Error in initial tick sync: %v", err)
		}

		// Then run on schedule
		for range ticker.C {
			log.Println("Starting scheduled tick sync...")
			ctx := context.Background()
			if err := h.climbTrackingService.SyncNewTicksForAllLocations(ctx); err != nil {
				log.Printf("Error in scheduled tick sync: %v", err)
			} else {
				log.Println("Scheduled tick sync completed successfully")
			}
		}
	}()
}

// StartBackgroundRouteSync starts a goroutine that checks for and syncs new Mountain Project routes periodically
func (h *Handler) StartBackgroundRouteSync(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		log.Printf("Starting background route sync scheduler (every %v)", interval)

		// Run immediately on startup
		log.Println("Running initial route sync...")
		ctx := context.Background()
		if err := h.climbTrackingService.SyncNewRoutesForAllStates(ctx); err != nil {
			log.Printf("Error in initial route sync: %v", err)
		}

		// Then run on schedule
		for range ticker.C {
			log.Println("Starting scheduled route sync...")
			ctx := context.Background()
			if err := h.climbTrackingService.SyncNewRoutesForAllStates(ctx); err != nil {
				log.Printf("Error in scheduled route sync: %v", err)
			} else {
				log.Println("Scheduled route sync completed successfully")
			}
		}
	}()
}

// StartPriorityRecalculation starts a background job that recalculates route sync priorities daily
// Only applies to non-location routes (location routes always sync daily)
func (h *Handler) StartPriorityRecalculation(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		log.Printf("Starting priority recalculation scheduler (every %v)", interval)

		// Run immediately on startup (before first sync)
		ctx := context.Background()
		if err := h.climbTrackingService.RecalculateAllPriorities(ctx); err != nil {
			log.Printf("Error in initial priority calculation: %v", err)
		}

		// Then run on schedule
		for range ticker.C {
			log.Println("Running scheduled priority recalculation...")
			ctx := context.Background()
			if err := h.climbTrackingService.RecalculateAllPriorities(ctx); err != nil {
				log.Printf("Error in priority recalculation: %v", err)
			}
		}
	}()
}

// StartLocationRouteSync starts a background job that syncs ticks and comments for location routes daily
// Location routes (with location_id) ALWAYS sync daily regardless of activity
func (h *Handler) StartLocationRouteSync(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		log.Printf("Starting location route sync scheduler (every %v)", interval)

		// Run immediately on startup (most critical - these are woulder locations)
		h.runLocationRouteSync()

		// Then run on schedule
		for range ticker.C {
			h.runLocationRouteSync()
		}
	}()
}

func (h *Handler) runLocationRouteSync() {
	log.Println("Starting location route sync (ticks + comments)...")
	ctx := context.Background()

	// Sync ticks for location routes
	if err := h.climbTrackingService.SyncLocationRouteTicks(ctx); err != nil {
		log.Printf("Error in location route tick sync: %v", err)
	}

	// Sync comments for location routes
	if err := h.climbTrackingService.SyncLocationRouteComments(ctx); err != nil {
		log.Printf("Error in location route comment sync: %v", err)
	}

	log.Println("Location route sync complete")
}

// StartHighPrioritySync starts a background job that syncs high-priority non-location routes daily
func (h *Handler) StartHighPrioritySync(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		log.Printf("Starting high-priority sync scheduler (every %v)", interval)

		// Run immediately on startup
		h.runHighPrioritySync()

		// Then run on schedule
		for range ticker.C {
			h.runHighPrioritySync()
		}
	}()
}

func (h *Handler) runHighPrioritySync() {
	log.Println("Starting high-priority sync (non-location routes, ticks + comments)...")
	ctx := context.Background()

	// Sync ticks for high-priority NON-LOCATION routes
	if err := h.climbTrackingService.SyncTicksByPriority(ctx, "high"); err != nil {
		log.Printf("Error in high-priority tick sync: %v", err)
	}

	// Sync comments for high-priority NON-LOCATION routes
	if err := h.climbTrackingService.SyncCommentsByPriority(ctx, "high"); err != nil {
		log.Printf("Error in high-priority comment sync: %v", err)
	}

	log.Println("High-priority sync complete")
}

// StartMediumPrioritySync starts a background job that syncs medium-priority non-location routes weekly
func (h *Handler) StartMediumPrioritySync(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		log.Printf("Starting medium-priority sync scheduler (every %v)", interval)

		// Run immediately on startup
		h.runMediumPrioritySync()

		// Then run on schedule
		for range ticker.C {
			h.runMediumPrioritySync()
		}
	}()
}

func (h *Handler) runMediumPrioritySync() {
	log.Println("Starting medium-priority sync (non-location routes, ticks + comments)...")
	ctx := context.Background()

	// Sync ticks for medium-priority NON-LOCATION routes
	if err := h.climbTrackingService.SyncTicksByPriority(ctx, "medium"); err != nil {
		log.Printf("Error in medium-priority tick sync: %v", err)
	}

	// Sync comments for medium-priority NON-LOCATION routes
	if err := h.climbTrackingService.SyncCommentsByPriority(ctx, "medium"); err != nil {
		log.Printf("Error in medium-priority comment sync: %v", err)
	}

	log.Println("Medium-priority sync complete")
}

// StartLowPrioritySync starts a background job that syncs low-priority non-location routes monthly
func (h *Handler) StartLowPrioritySync(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		log.Printf("Starting low-priority sync scheduler (every %v)", interval)

		// Run immediately on startup
		h.runLowPrioritySync()

		// Then run on schedule
		for range ticker.C {
			h.runLowPrioritySync()
		}
	}()
}

func (h *Handler) runLowPrioritySync() {
	log.Println("Starting low-priority sync (non-location routes, ticks + comments)...")
	ctx := context.Background()

	// Sync ticks for low-priority NON-LOCATION routes
	if err := h.climbTrackingService.SyncTicksByPriority(ctx, "low"); err != nil {
		log.Printf("Error in low-priority tick sync: %v", err)
	}

	// Sync comments for low-priority NON-LOCATION routes
	if err := h.climbTrackingService.SyncCommentsByPriority(ctx, "low"); err != nil {
		log.Printf("Error in low-priority comment sync: %v", err)
	}

	log.Println("Low-priority sync complete")
}

// HealthCheck returns service health status
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "woulder-api",
		"time":    time.Now().Format(time.RFC3339),
	})
}

// GetAllLocations returns all saved locations
func (h *Handler) GetAllLocations(c *gin.Context) {
	ctx := c.Request.Context()

	locations, err := h.locationService.GetAllLocations(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch locations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"locations": locations,
		"count":     len(locations),
	})
}

// GetWeatherForLocation returns complete weather forecast for a location
func (h *Handler) GetWeatherForLocation(c *gin.Context) {
	ctx := c.Request.Context()

	locationID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid location ID"})
		return
	}

	forecast, err := h.weatherService.GetLocationWeather(ctx, locationID)
	if err != nil {
		log.Printf("Error fetching weather for location %d: %v", locationID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch weather data"})
		return
	}

	c.JSON(http.StatusOK, forecast)
}

// GetWeatherByCoordinates returns weather for arbitrary coordinates
func (h *Handler) GetWeatherByCoordinates(c *gin.Context) {
	ctx := c.Request.Context()

	latStr := c.Query("lat")
	lonStr := c.Query("lon")

	if latStr == "" || lonStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing lat or lon query parameters"})
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid latitude"})
		return
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid longitude"})
		return
	}

	forecast, err := h.weatherService.GetWeatherByCoordinates(ctx, lat, lon)
	if err != nil {
		log.Printf("Error fetching weather for coordinates (%.2f, %.2f): %v", lat, lon, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch weather data"})
		return
	}

	c.JSON(http.StatusOK, forecast)
}

// GetAllWeather returns weather for all locations or filtered by area
func (h *Handler) GetAllWeather(c *gin.Context) {
	ctx := c.Request.Context()

	var areaID *int
	if areaIDStr := c.Query("area_id"); areaIDStr != "" {
		parsedID, err := strconv.Atoi(areaIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid area_id"})
			return
		}
		areaID = &parsedID
	}

	forecasts, err := h.weatherService.GetAllWeather(ctx, areaID)
	if err != nil {
		log.Printf("Error fetching all weather: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch weather data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"forecasts":  forecasts,
		"count":      len(forecasts),
		"updated_at": time.Now().Format(time.RFC3339),
	})
}

// RefreshWeather manually triggers a weather data refresh
func (h *Handler) RefreshWeather(c *gin.Context) {
	ctx := c.Request.Context()

	log.Println("Manual weather refresh triggered")

	err := h.weatherService.RefreshAllWeather(ctx)
	if err != nil {
		log.Printf("Error during manual refresh: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Weather data refresh completed",
		"updated_at": time.Now().Format(time.RFC3339),
	})
}

// RefreshRoutes manually triggers a new route sync check
func (h *Handler) RefreshRoutes(c *gin.Context) {
	ctx := c.Request.Context()

	log.Println("Manual route sync triggered")

	err := h.climbTrackingService.SyncNewRoutesForAllStates(ctx)
	if err != nil {
		log.Printf("Error during manual route sync: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Route sync completed",
		"updated_at": time.Now().Format(time.RFC3339),
	})
}

// GetRiverDataForLocation returns current river data for all rivers at a location
func (h *Handler) GetRiverDataForLocation(c *gin.Context) {
	ctx := c.Request.Context()

	locationID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid location ID"})
		return
	}

	riverData, err := h.riverService.GetRiverDataForLocation(ctx, locationID)
	if err != nil {
		log.Printf("Error fetching rivers for location %d: %v", locationID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch river data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rivers":     riverData,
		"updated_at": time.Now().Format(time.RFC3339),
	})
}

// GetRiverDataByID returns current data for a specific river crossing
func (h *Handler) GetRiverDataByID(c *gin.Context) {
	ctx := c.Request.Context()

	riverID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid river ID"})
		return
	}

	log.Printf("Fetching river ID: %d", riverID)

	riverData, err := h.riverService.GetRiverDataByID(ctx, riverID)
	if err != nil {
		log.Printf("Error fetching river %d: %v", riverID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch river data"})
		return
	}

	c.JSON(http.StatusOK, riverData)
}

// GetAllAreas returns all climbing areas with location counts
func (h *Handler) GetAllAreas(c *gin.Context) {
	ctx := c.Request.Context()

	areas, err := h.locationService.GetAreasWithLocationCounts(ctx)
	if err != nil {
		log.Printf("Error fetching areas: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch areas"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"areas": areas,
		"count": len(areas),
	})
}

// GetLocationsByArea returns all locations in a specific area
func (h *Handler) GetLocationsByArea(c *gin.Context) {
	ctx := c.Request.Context()

	areaID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid area ID"})
		return
	}

	locations, err := h.locationService.GetLocationsByArea(ctx, areaID)
	if err != nil {
		log.Printf("Error fetching locations for area %d: %v", areaID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch locations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"locations": locations,
		"count":     len(locations),
	})
}

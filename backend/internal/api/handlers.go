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
}

func NewHandler(
	locationService *service.LocationService,
	weatherService *service.WeatherService,
	riverService *service.RiverService,
	climbTrackingService *service.ClimbTrackingService,
) *Handler {
	return &Handler{
		locationService:      locationService,
		weatherService:       weatherService,
		riverService:         riverService,
		climbTrackingService: climbTrackingService,
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

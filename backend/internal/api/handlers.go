package api

import (
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/alexscott64/woulder/backend/internal/database"
	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/alexscott64/woulder/backend/internal/weather"
	"github.com/alexscott64/woulder/backend/internal/rivers"
)

type Handler struct {
	db             *database.Database
	weatherService *weather.WeatherService
	riverClient    *rivers.USGSClient
	refreshMutex   sync.Mutex
	lastRefresh    time.Time
	isRefreshing   bool
}

func NewHandler(db *database.Database, weatherService *weather.WeatherService) *Handler {
	h := &Handler{
		db:             db,
		weatherService: weatherService,
		riverClient:    rivers.NewUSGSClient(),
	}
	return h
}

// StartBackgroundRefresh starts a goroutine that refreshes weather data periodically
func (h *Handler) StartBackgroundRefresh(interval time.Duration) {
	// Do initial refresh on startup
	go func() {
		log.Println("Starting initial weather data refresh...")
		h.refreshAllWeatherData()
	}()

	// Start periodic refresh
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			log.Println("Starting scheduled weather data refresh...")
			h.refreshAllWeatherData()
		}
	}()

	log.Printf("Background weather refresh scheduled every %v", interval)
}

// refreshAllWeatherData fetches fresh weather data for all locations
func (h *Handler) refreshAllWeatherData() {
	h.refreshMutex.Lock()
	if h.isRefreshing {
		h.refreshMutex.Unlock()
		log.Println("Refresh already in progress, skipping")
		return
	}
	h.isRefreshing = true
	h.refreshMutex.Unlock()

	defer func() {
		h.refreshMutex.Lock()
		h.isRefreshing = false
		h.lastRefresh = time.Now()
		h.refreshMutex.Unlock()
	}()

	locations, err := h.db.GetAllLocations()
	if err != nil {
		log.Printf("Error fetching locations for refresh: %v", err)
		return
	}

	// Fetch sequentially with delay to avoid rate limiting
	for _, location := range locations {
		log.Printf("Refreshing weather data for location %d (%s)", location.ID, location.Name)

		current, forecast, err := h.weatherService.GetCurrentAndForecast(location.Latitude, location.Longitude)
		if err != nil {
			log.Printf("Error refreshing weather for location %d: %v", location.ID, err)
			continue
		}

		current.LocationID = location.ID
		if err := h.db.SaveWeatherData(current); err != nil {
			log.Printf("Error saving current weather: %v", err)
		}

		for _, f := range forecast {
			f.LocationID = location.ID
			if err := h.db.SaveWeatherData(&f); err != nil {
				log.Printf("Error saving forecast data: %v", err)
			}
		}

		// Small delay between locations to be nice to the API
		time.Sleep(500 * time.Millisecond)
	}

	// Clean old data
	if err := h.db.CleanOldWeatherData(14); err != nil {
		log.Printf("Error cleaning old weather data: %v", err)
	}

	log.Println("Weather data refresh complete")
}

// HealthCheck returns API health status
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Woulder API is running",
		"time":    time.Now().Format(time.RFC3339),
	})
}

// GetAllLocations returns all saved locations
func (h *Handler) GetAllLocations(c *gin.Context) {
	locations, err := h.db.GetAllLocations()
	if err != nil {
		log.Printf("Error fetching locations: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch locations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"locations": locations,
	})
}

// GetWeatherForLocation returns current weather and forecast for a location
func (h *Handler) GetWeatherForLocation(c *gin.Context) {
	locationID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid location ID"})
		return
	}

	// Get location details
	location, err := h.db.GetLocation(locationID)
	if err != nil {
		log.Printf("Error fetching location %d: %v", locationID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Location not found"})
		return
	}

	// Fetch current weather and forecast in a single API call
	current, forecast, err := h.weatherService.GetCurrentAndForecast(location.Latitude, location.Longitude)
	if err != nil {
		log.Printf("Error fetching weather for location %d: %v", locationID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch weather"})
		return
	}
	current.LocationID = locationID

	// Save current weather to database
	if err := h.db.SaveWeatherData(current); err != nil {
		log.Printf("Error saving current weather: %v", err)
	}

	// Save forecast to database
	for _, f := range forecast {
		f.LocationID = locationID
		if err := h.db.SaveWeatherData(&f); err != nil {
			log.Printf("Error saving forecast data: %v", err)
		}
	}

	// Get historical data (last 7 days)
	historical, err := h.db.GetHistoricalWeather(locationID, 7)
	if err != nil {
		log.Printf("Error fetching historical weather: %v", err)
		historical = []models.WeatherData{}
	}

	response := models.WeatherForecast{
		LocationID: locationID,
		Location:   *location,
		Current:    *current,
		Hourly:     forecast,
		Historical: historical,
	}

	c.JSON(http.StatusOK, response)
}

// GetWeatherByCoordinates returns weather for specific coordinates (for custom locations)
func (h *Handler) GetWeatherByCoordinates(c *gin.Context) {
	latStr := c.Query("lat")
	lonStr := c.Query("lon")

	if latStr == "" || lonStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lat and lon query parameters are required"})
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

	// Fetch current weather
	current, err := h.weatherService.GetCurrentWeather(lat, lon)
	if err != nil {
		log.Printf("Error fetching weather for coordinates (%.6f, %.6f): %v", lat, lon, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch weather"})
		return
	}

	// Fetch forecast
	forecast, err := h.weatherService.GetForecast(lat, lon)
	if err != nil {
		log.Printf("Error fetching forecast for coordinates (%.6f, %.6f): %v", lat, lon, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch forecast"})
		return
	}

	response := gin.H{
		"current":  current,
		"forecast": forecast,
	}

	c.JSON(http.StatusOK, response)
}

// GetAllWeather returns weather for all locations from cache (fast response)
func (h *Handler) GetAllWeather(c *gin.Context) {
	locations, err := h.db.GetAllLocations()
	if err != nil {
		log.Printf("Error fetching locations: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch locations"})
		return
	}

	// Serve data from cache/database immediately
	allWeather := make([]models.WeatherForecast, 0, len(locations))

	for _, location := range locations {
		// Get current weather from database
		current, err := h.db.GetCurrentWeather(location.ID)
		if err != nil {
			log.Printf("No cached weather for location %d: %v", location.ID, err)
			continue
		}

		// Get forecast from database
		forecast, err := h.db.GetForecastWeather(location.ID)
		if err != nil {
			log.Printf("Error fetching forecast for location %d: %v", location.ID, err)
			forecast = []models.WeatherData{}
		}

		// Get historical data
		historical, err := h.db.GetHistoricalWeather(location.ID, 7)
		if err != nil {
			log.Printf("Error fetching historical weather: %v", err)
			historical = []models.WeatherData{}
		}

		allWeather = append(allWeather, models.WeatherForecast{
			LocationID: location.ID,
			Location:   location,
			Current:    *current,
			Hourly:     forecast,
			Historical: historical,
		})
	}

	// Get last refresh time
	h.refreshMutex.Lock()
	lastRefresh := h.lastRefresh
	isRefreshing := h.isRefreshing
	h.refreshMutex.Unlock()

	updatedAt := lastRefresh.Format(time.RFC3339)
	if lastRefresh.IsZero() {
		updatedAt = "refreshing..."
	}

	c.JSON(http.StatusOK, gin.H{
		"weather":      allWeather,
		"updated_at":   updatedAt,
		"is_refreshing": isRefreshing,
	})
}

// RefreshWeather triggers a manual refresh of weather data (runs in background)
func (h *Handler) RefreshWeather(c *gin.Context) {
	h.refreshMutex.Lock()
	isRefreshing := h.isRefreshing
	h.refreshMutex.Unlock()

	if isRefreshing {
		c.JSON(http.StatusAccepted, gin.H{
			"message": "Refresh already in progress",
			"status":  "in_progress",
		})
		return
	}

	// Trigger refresh in background
	go h.refreshAllWeatherData()

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Refresh started",
		"status":  "started",
	})
}

// GetRiverDataForLocation returns current river crossing data for a location
func (h *Handler) GetRiverDataForLocation(c *gin.Context) {
	locationID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid location ID"})
		return
	}

	// Get all rivers for this location
	log.Printf("Fetching rivers for location ID: %d", locationID)
	locationRivers, err := h.db.GetRiversByLocation(locationID)
	if err != nil {
		log.Printf("Error fetching rivers for location %d: %v", locationID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch river data"})
		return
	}

	log.Printf("Found %d rivers for location %d", len(locationRivers), locationID)
	if len(locationRivers) == 0 {
		log.Printf("No rivers found for location %d, returning empty array", locationID)
		c.JSON(http.StatusOK, gin.H{"rivers": []models.RiverData{}})
		return
	}

	// Fetch current data for each river
	var riverData []models.RiverData
	for _, river := range locationRivers {
		gaugeFlowCFS, gaugeHeightFt, timestamp, err := h.riverClient.GetRiverData(river.GaugeID)
		if err != nil {
			log.Printf("Error fetching USGS data for gauge %s: %v", river.GaugeID, err)
			continue
		}

		// Apply flow estimation if needed
		actualFlowCFS := gaugeFlowCFS
		if river.IsEstimated {
			if river.FlowDivisor != nil && *river.FlowDivisor > 0 {
				// Simple divisor method (e.g., gauge / 2 for North Fork at Index)
				actualFlowCFS = gaugeFlowCFS / *river.FlowDivisor
				log.Printf("Estimated flow for %s: gauge %.0f CFS / %.1f = river %.0f CFS",
					river.RiverName, gaugeFlowCFS, *river.FlowDivisor, actualFlowCFS)
			} else if river.DrainageAreaSqMi != nil && river.GaugeDrainageAreaSqMi != nil {
				// Drainage area ratio method
				actualFlowCFS = rivers.EstimateFlowFromDrainageRatio(
					gaugeFlowCFS,
					*river.DrainageAreaSqMi,
					*river.GaugeDrainageAreaSqMi,
				)
				log.Printf("Estimated flow for %s: gauge %.0f CFS -> river %.0f CFS (drainage ratio %.3f)",
					river.RiverName, gaugeFlowCFS, actualFlowCFS, *river.DrainageAreaSqMi / *river.GaugeDrainageAreaSqMi)
			}
		}

		status, message, isSafe, percentOfSafe := rivers.CalculateCrossingStatus(river, actualFlowCFS)

		riverData = append(riverData, models.RiverData{
			River:         river,
			FlowCFS:       actualFlowCFS,
			GaugeHeightFt: gaugeHeightFt,
			IsSafe:        isSafe,
			Status:        status,
			StatusMessage: message,
			Timestamp:     timestamp,
			PercentOfSafe: percentOfSafe,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"rivers":     riverData,
		"updated_at": time.Now().Format(time.RFC3339),
	})
}

// GetRiverDataByID returns current data for a specific river crossing
func (h *Handler) GetRiverDataByID(c *gin.Context) {
	riverID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid river ID"})
		return
	}

	// Get river info from database
	river, err := h.db.GetRiverByID(riverID)
	if err != nil {
		log.Printf("Error fetching river %d: %v", riverID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "River not found"})
		return
	}

	// Fetch current USGS data
	gaugeFlowCFS, gaugeHeightFt, timestamp, err := h.riverClient.GetRiverData(river.GaugeID)
	if err != nil {
		log.Printf("Error fetching USGS data for gauge %s: %v", river.GaugeID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch river data"})
		return
	}

	// Apply flow estimation if needed
	actualFlowCFS := gaugeFlowCFS
	if river.IsEstimated {
		if river.FlowDivisor != nil && *river.FlowDivisor > 0 {
			// Simple divisor method (e.g., gauge / 2 for North Fork at Index)
			actualFlowCFS = gaugeFlowCFS / *river.FlowDivisor
			log.Printf("Estimated flow for %s: gauge %.0f CFS / %.1f = river %.0f CFS",
				river.RiverName, gaugeFlowCFS, *river.FlowDivisor, actualFlowCFS)
		} else if river.DrainageAreaSqMi != nil && river.GaugeDrainageAreaSqMi != nil {
			// Drainage area ratio method
			actualFlowCFS = rivers.EstimateFlowFromDrainageRatio(
				gaugeFlowCFS,
				*river.DrainageAreaSqMi,
				*river.GaugeDrainageAreaSqMi,
			)
			log.Printf("Estimated flow for %s: gauge %.0f CFS -> river %.0f CFS (drainage ratio %.3f)",
				river.RiverName, gaugeFlowCFS, actualFlowCFS, *river.DrainageAreaSqMi / *river.GaugeDrainageAreaSqMi)
		}
	}

	status, message, isSafe, percentOfSafe := rivers.CalculateCrossingStatus(*river, actualFlowCFS)

	riverData := models.RiverData{
		River:         *river,
		FlowCFS:       actualFlowCFS,
		GaugeHeightFt: gaugeHeightFt,
		IsSafe:        isSafe,
		Status:        status,
		StatusMessage: message,
		Timestamp:     timestamp,
		PercentOfSafe: percentOfSafe,
	}

	c.JSON(http.StatusOK, riverData)
}

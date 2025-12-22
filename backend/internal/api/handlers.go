package api

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/alexscott64/woulder/backend/internal/database"
	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/alexscott64/woulder/backend/internal/weather"
)

type Handler struct {
	db             *database.Database
	weatherService *weather.WeatherService
}

func NewHandler(db *database.Database, weatherService *weather.WeatherService) *Handler {
	return &Handler{
		db:             db,
		weatherService: weatherService,
	}
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

// GetAllWeather returns weather for all locations (convenient for dashboard)
func (h *Handler) GetAllWeather(c *gin.Context) {
	locations, err := h.db.GetAllLocations()
	if err != nil {
		log.Printf("Error fetching locations: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch locations"})
		return
	}

	// Fetch weather for all locations in parallel with concurrency limit
	type result struct {
		forecast models.WeatherForecast
		err      error
		index    int
	}

	results := make(chan result, len(locations))
	// Limit to 3 concurrent requests to avoid rate limiting
	semaphore := make(chan struct{}, 3)

	for i, location := range locations {
		go func(loc models.Location, idx int) {
			semaphore <- struct{}{} // Acquire
			defer func() { <-semaphore }() // Release
			log.Printf("Fetching fresh data from API for location %d", loc.ID)

			// Fetch current weather and forecast in a single API call
			current, forecast, err := h.weatherService.GetCurrentAndForecast(loc.Latitude, loc.Longitude)
			if err != nil {
				log.Printf("Error fetching weather for location %d: %v", loc.ID, err)
				results <- result{err: err, index: idx}
				return
			}
			current.LocationID = loc.ID

			// Save current weather to database
			if err := h.db.SaveWeatherData(current); err != nil {
				log.Printf("Error saving weather data: %v", err)
			}

			// Save forecast
			for _, f := range forecast {
				f.LocationID = loc.ID
				if err := h.db.SaveWeatherData(&f); err != nil {
					log.Printf("Error saving forecast data: %v", err)
				}
			}

			// Get historical
			historical, err := h.db.GetHistoricalWeather(loc.ID, 7)
			if err != nil {
				log.Printf("Error fetching historical weather: %v", err)
				historical = []models.WeatherData{}
			}

			results <- result{
				forecast: models.WeatherForecast{
					LocationID: loc.ID,
					Location:   loc,
					Current:    *current,
					Hourly:     forecast,
					Historical: historical,
				},
				index: idx,
			}
		}(location, i)
	}

	// Collect results maintaining order
	allWeather := make([]models.WeatherForecast, 0, len(locations))
	resultMap := make(map[int]models.WeatherForecast)

	for i := 0; i < len(locations); i++ {
		res := <-results
		if res.err == nil {
			resultMap[res.index] = res.forecast
		}
	}

	// Maintain original order
	for i := 0; i < len(locations); i++ {
		if forecast, ok := resultMap[i]; ok {
			allWeather = append(allWeather, forecast)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"weather":    allWeather,
		"updated_at": time.Now().Format(time.RFC3339),
	})
}

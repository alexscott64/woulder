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
	db            *database.Database
	weatherClient *weather.Client
}

func NewHandler(db *database.Database, weatherClient *weather.Client) *Handler {
	return &Handler{
		db:            db,
		weatherClient: weatherClient,
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

	// Fetch current weather
	current, err := h.weatherClient.GetCurrentWeather(location.Latitude, location.Longitude)
	if err != nil {
		log.Printf("Error fetching current weather for location %d: %v", locationID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch current weather"})
		return
	}
	current.LocationID = locationID

	// Save current weather to database
	if err := h.db.SaveWeatherData(current); err != nil {
		log.Printf("Error saving current weather: %v", err)
	}

	// Fetch forecast
	forecast, err := h.weatherClient.GetForecast(location.Latitude, location.Longitude)
	if err != nil {
		log.Printf("Error fetching forecast for location %d: %v", locationID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch forecast"})
		return
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
	current, err := h.weatherClient.GetCurrentWeather(lat, lon)
	if err != nil {
		log.Printf("Error fetching weather for coordinates (%.6f, %.6f): %v", lat, lon, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch weather"})
		return
	}

	// Fetch forecast
	forecast, err := h.weatherClient.GetForecast(lat, lon)
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

	var allWeather []models.WeatherForecast

	for _, location := range locations {
		log.Printf("Fetching fresh data from API for location %d", location.ID)

		// Fetch current weather
		current, err := h.weatherClient.GetCurrentWeather(location.Latitude, location.Longitude)
		if err != nil {
			log.Printf("Error fetching weather for location %d: %v", location.ID, err)
			continue
		}
		current.LocationID = location.ID

		// Save to database
		if err := h.db.SaveWeatherData(current); err != nil {
			log.Printf("Error saving weather data: %v", err)
		}

		// Fetch forecast
		forecast, err := h.weatherClient.GetForecast(location.Latitude, location.Longitude)
		if err != nil {
			log.Printf("Error fetching forecast for location %d: %v", location.ID, err)
			continue
		}

		// Save forecast
		for _, f := range forecast {
			f.LocationID = location.ID
			if err := h.db.SaveWeatherData(&f); err != nil {
				log.Printf("Error saving forecast data: %v", err)
			}
		}

		// Get historical
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

	c.JSON(http.StatusOK, gin.H{
		"weather":    allWeather,
		"updated_at": time.Now().Format(time.RFC3339),
	})
}

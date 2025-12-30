package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/alexscott64/woulder/backend/internal/api"
	"github.com/alexscott64/woulder/backend/internal/database"
	"github.com/alexscott64/woulder/backend/internal/weather"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Initialize database
	db, err := database.New()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize weather service (Open-Meteo with OpenWeatherMap fallback)
	weatherService := weather.NewWeatherService()

	// Initialize API handler
	handler := api.NewHandler(db, weatherService)

	// Start background weather refresh (every 2 hours)
	handler.StartBackgroundRefresh(2 * time.Hour)

	// Set Gin mode
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = "release"
	}
	gin.SetMode(ginMode)

	// Create Gin router
	router := gin.Default()

	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // In production, replace with specific origins
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// API routes
	apiGroup := router.Group("/api")
	{
		apiGroup.GET("/health", handler.HealthCheck)
		apiGroup.GET("/locations", handler.GetAllLocations)
		apiGroup.GET("/areas", handler.GetAllAreas)
		apiGroup.GET("/areas/:id/locations", handler.GetLocationsByArea)
		apiGroup.GET("/weather/all", handler.GetAllWeather)
		apiGroup.GET("/weather/:id", handler.GetWeatherForLocation)
		apiGroup.GET("/weather/coordinates", handler.GetWeatherByCoordinates)
		apiGroup.POST("/weather/refresh", handler.RefreshWeather)
		apiGroup.GET("/rivers/location/:id", handler.GetRiverDataForLocation)
		apiGroup.GET("/rivers/:id", handler.GetRiverDataByID)
	}

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("Starting Woulder API server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

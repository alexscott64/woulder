package main

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/alexscott64/woulder/backend/internal/api"
	"github.com/alexscott64/woulder/backend/internal/config"
	"github.com/alexscott64/woulder/backend/internal/database"
	"github.com/alexscott64/woulder/backend/internal/mountainproject"
	"github.com/alexscott64/woulder/backend/internal/rivers"
	"github.com/alexscott64/woulder/backend/internal/service"
	"github.com/alexscott64/woulder/backend/internal/weather"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := database.New()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize external API clients
	weatherClient := weather.NewWeatherService(cfg.Weather.OpenWeatherMapAPIKey)
	riverClient := rivers.NewUSGSClient()
	mpClient := mountainproject.NewClient()

	// Initialize services with dependency injection
	locationService := service.NewLocationService(db)
	climbTrackingService := service.NewClimbTrackingService(db, mpClient)
	weatherServiceLayer := service.NewWeatherService(db, weatherClient, climbTrackingService)
	riverServiceLayer := service.NewRiverService(db, riverClient)

	// Initialize API handler with services
	handler := api.NewHandler(locationService, weatherServiceLayer, riverServiceLayer, climbTrackingService)

	// Start background weather refresh (every 2 hours)
	handler.StartBackgroundRefresh(2 * time.Hour)

	// Start background tick sync (every 24 hours)
	handler.StartBackgroundTickSync(24 * time.Hour)

	// Set Gin mode
	gin.SetMode(cfg.Server.GinMode)

	// Create Gin router
	router := gin.Default()

	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.Server.CORS.AllowOrigins,
		AllowMethods:     cfg.Server.CORS.AllowMethods,
		AllowHeaders:     cfg.Server.CORS.AllowHeaders,
		ExposeHeaders:    cfg.Server.CORS.ExposeHeaders,
		AllowCredentials: cfg.Server.CORS.AllowCredentials,
		MaxAge:           cfg.Server.CORS.MaxAge,
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
		apiGroup.POST("/climbs/refresh", handler.RefreshClimbData)
		apiGroup.GET("/climbs/location/:id", handler.GetLastClimbedForLocation)
		apiGroup.GET("/climbs/location/:id/areas", handler.GetAreasOrderedByActivity)
		apiGroup.GET("/climbs/location/:id/areas/:area_id/subareas", handler.GetSubareasOrderedByActivity)
		apiGroup.GET("/climbs/location/:id/areas/:area_id/routes", handler.GetRoutesOrderedByActivity)
		apiGroup.GET("/climbs/routes/:route_id/ticks", handler.GetRecentTicksForRoute)
		apiGroup.GET("/climbs/location/:id/search-all", handler.SearchInLocation)
		apiGroup.GET("/climbs/location/:id/search", handler.SearchRoutesInLocation)
	}

	// Start server
	log.Printf("Starting Woulder API server on port %s", cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

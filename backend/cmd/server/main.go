package main

// Trigger Air recompilation to test checkpoint recovery

import (
	"context"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/alexscott64/woulder/backend/internal/api"
	"github.com/alexscott64/woulder/backend/internal/config"
	"github.com/alexscott64/woulder/backend/internal/database"
	"github.com/alexscott64/woulder/backend/internal/monitoring"
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

	// Initialize job monitor
	jobMonitor := monitoring.NewJobMonitor(db.Conn())

	// Initialize services with dependency injection
	locationService := service.NewLocationService(db.Locations(), db.Areas())
	climbTrackingService := service.NewClimbTrackingService(db.MountainProject(), db.Climbing(), mpClient, jobMonitor)

	// Recover any interrupted jobs from previous run (before starting new background jobs)
	log.Println("Checking for interrupted jobs from previous run...")
	recoverInterruptedJobs(jobMonitor, climbTrackingService)

	weatherServiceLayer := service.NewWeatherService(db.Weather(), db.Locations(), db.Rocks(), weatherClient, climbTrackingService)
	riverServiceLayer := service.NewRiverService(db.Rivers(), riverClient)
	boulderDryingService := service.NewBoulderDryingService(db.Boulders(), db.Weather(), db.Locations(), db.Rocks(), db.MountainProject(), weatherClient)
	heatMapService := service.NewHeatMapService(db.HeatMap())

	// Initialize API handler with services
	handler := api.NewHandler(locationService, weatherServiceLayer, riverServiceLayer, climbTrackingService, boulderDryingService, heatMapService, db.Kaya(), jobMonitor)

	// Start background syncs only if not disabled (e.g., in development)
	if cfg.Server.DisableBackgroundSyncs {
		log.Println("⚠️  Background syncs DISABLED via config (DISABLE_BACKGROUND_SYNCS=true)")
	} else {
		log.Println("✓ Starting background syncs...")

		// Start background weather refresh (every 1 hour)
		// The refresh automatically checks if data is fresh and skips API calls if updated within the last hour
		handler.StartBackgroundRefresh(1 * time.Hour)

		// Start dual-track sync system for Mountain Project ticks/comments
		// Priority recalculation runs FIRST (populates priorities for non-location routes)
		handler.StartPriorityRecalculation(24 * time.Hour)
		// Location route sync runs SECOND (ensures woulder locations are fresh - most critical)
		handler.StartLocationRouteSync(24 * time.Hour)
		// High-priority sync runs THIRD (ensures popular non-location routes are fresh)
		handler.StartHighPrioritySync(24 * time.Hour)
		// Medium-priority sync runs weekly
		handler.StartMediumPrioritySync(7 * 24 * time.Hour)
		// Low-priority sync runs monthly
		handler.StartLowPrioritySync(30 * 24 * time.Hour)

		// Start background route sync (every 24 hours)
		handler.StartBackgroundRouteSync(24 * time.Hour)
	}

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
		apiGroup.POST("/routes/refresh", handler.RefreshRoutes)
		apiGroup.GET("/rivers/location/:id", handler.GetRiverDataForLocation)
		apiGroup.GET("/rivers/:id", handler.GetRiverDataByID)
		apiGroup.POST("/climbs/refresh", handler.RefreshClimbData)
		apiGroup.GET("/climbs/location/:id", handler.GetLastClimbedForLocation)
		apiGroup.GET("/climbs/location/:id/areas", handler.GetAreasOrderedByActivity)
		apiGroup.GET("/climbs/location/:id/areas/:area_id/subareas", handler.GetSubareasOrderedByActivity)
		apiGroup.GET("/climbs/location/:id/areas/:area_id/routes", handler.GetRoutesOrderedByActivity)
		apiGroup.GET("/climbs/location/:id/areas/:area_id/unified-routes", handler.GetUnifiedRoutesOrderedByActivity)
		apiGroup.GET("/climbs/location/:id/areas/:area_id/drying-stats", handler.GetAreaDryingStats)
		apiGroup.GET("/climbs/location/:id/batch-area-drying-stats", handler.GetBatchAreaDryingStats)
		apiGroup.GET("/climbs/routes/:route_id/ticks", handler.GetRecentTicksForRoute)
		apiGroup.GET("/climbs/routes/:route_id/drying-status", handler.GetBoulderDryingStatus)
		apiGroup.GET("/climbs/routes/batch-drying-status", handler.GetBatchBoulderDryingStatus)
		apiGroup.GET("/climbs/location/:id/search-all", handler.SearchInLocation)
		apiGroup.GET("/climbs/location/:id/search", handler.SearchRoutesInLocation)

		// Heat map routes
		apiGroup.GET("/heat-map/activity", handler.GetHeatMapActivity)
		apiGroup.GET("/heat-map/area/:area_id/detail", handler.GetHeatMapAreaDetail)
		apiGroup.GET("/heat-map/routes", handler.GetHeatMapRoutes)
		apiGroup.GET("/heat-map/route/:route_id/ticks", handler.GetRouteTicksInDateRange)
		apiGroup.POST("/heat-map/cluster/search-routes", handler.SearchClusterRoutes)

		// Kaya routes
		apiGroup.GET("/kaya/location/:id/ascents", handler.GetKayaAscentsForLocation)

		// Job monitoring routes
		apiGroup.GET("/monitoring/jobs/active", handler.GetActiveJobs)
		apiGroup.GET("/monitoring/jobs/history", handler.GetJobHistory)
		apiGroup.GET("/monitoring/jobs/summary", handler.GetJobsSummary)
		apiGroup.GET("/monitoring/jobs/:job_id", handler.GetJobStatus)
	}

	// Serve job monitoring dashboard at /jtrack
	router.StaticFile("/jtrack", "./static/jtrack.html")
	log.Printf("Job monitoring dashboard available at /jtrack")

	// Start server
	log.Printf("Starting Woulder API server on port %s", cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// recoverInterruptedJobs checks for jobs that were running when server stopped
// and automatically resumes them
func recoverInterruptedJobs(
	jobMonitor *monitoring.JobMonitor,
	climbService *service.ClimbTrackingService,
) {
	ctx := context.Background()

	// Find jobs interrupted in the last 24 hours
	interrupted, err := jobMonitor.RecoverInterruptedJobs(ctx, 24*time.Hour)
	if err != nil {
		log.Printf("Error checking for interrupted jobs: %v", err)
		return
	}

	if len(interrupted) == 0 {
		log.Println("No interrupted jobs found - starting clean")
		return
	}

	log.Printf("Found %d interrupted job(s), attempting to resume...", len(interrupted))

	for _, job := range interrupted {
		log.Printf("  - %s (ID: %d, started: %v, status: %s)",
			job.JobName, job.ID, job.StartedAt, job.Status)

		// Resume based on job type
		switch job.JobName {
		case "route_sync_all_states":
			// Don't start in goroutine - resumes automatically when background scheduler runs
			log.Printf("Route sync job will be resumed by StartBackgroundRouteSync")

		case "location_tick_sync":
			// Don't start in goroutine - resumes automatically when background scheduler runs
			log.Printf("Location tick sync job will be resumed by StartLocationRouteSync")

		case "location_comment_sync":
			// Don't start in goroutine - resumes automatically when background scheduler runs
			log.Printf("Location comment sync job will be resumed by StartLocationRouteSync")

		default:
			// Priority-based sync jobs (high_priority_tick_sync, medium_priority_tick_sync, etc.)
			// will also resume automatically when their schedulers run
			log.Printf("Job %s will be resumed by its scheduled background sync", job.JobName)
		}
	}
}

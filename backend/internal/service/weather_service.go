package service

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/alexscott64/woulder/backend/internal/database/locations"
	"github.com/alexscott64/woulder/backend/internal/database/rocks"
	"github.com/alexscott64/woulder/backend/internal/database/weather"
	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/alexscott64/woulder/backend/internal/pests"
	weatherPkg "github.com/alexscott64/woulder/backend/internal/weather"
	"github.com/alexscott64/woulder/backend/internal/weather/calculator"
	"github.com/alexscott64/woulder/backend/internal/weather/client"
	"github.com/alexscott64/woulder/backend/internal/weather/rock_drying"
	sunpkg "github.com/alexscott64/woulder/backend/internal/weather/sun"
)

type WeatherService struct {
	weatherRepo          weather.Repository
	locationsRepo        locations.Repository
	rocksRepo            rocks.Repository
	weatherClient        *weatherPkg.WeatherService
	rockCalculator       *rock_drying.Calculator
	pestAnalyzer         *pests.PestAnalyzer
	climbTrackingService *ClimbTrackingService

	// Background refresh management
	refreshMutex sync.Mutex
	lastRefresh  time.Time
	isRefreshing bool
}

func NewWeatherService(
	weatherRepo weather.Repository,
	locationsRepo locations.Repository,
	rocksRepo rocks.Repository,
	client *weatherPkg.WeatherService,
	climbService *ClimbTrackingService,
) *WeatherService {
	return &WeatherService{
		weatherRepo:          weatherRepo,
		locationsRepo:        locationsRepo,
		rocksRepo:            rocksRepo,
		weatherClient:        client,
		rockCalculator:       &rock_drying.Calculator{},
		pestAnalyzer:         &pests.PestAnalyzer{},
		climbTrackingService: climbService,
	}
}

// GetLocationWeather retrieves complete weather forecast for a location
// Uses cached data from database if available and fresh (< 1 hour old)
// includeClimbHistory controls whether to fetch climb history (expensive query)
func (s *WeatherService) GetLocationWeather(ctx context.Context, locationID int) (*models.WeatherForecast, error) {
	return s.getLocationWeatherWithOptions(ctx, locationID, true)
}

// getLocationWeatherWithOptions is the internal implementation with configurable options
func (s *WeatherService) getLocationWeatherWithOptions(ctx context.Context, locationID int, includeClimbHistory bool) (*models.WeatherForecast, error) {
	// 1. Get location
	location, err := s.locationsRepo.GetByID(ctx, locationID)
	if err != nil {
		return nil, fmt.Errorf("location not found: %w", err)
	}

	// 2. Try to get current weather from database cache first
	var current *models.WeatherData
	var hourlyForecast []models.WeatherData
	var sunTimes *client.SunTimes

	cachedCurrent, err := s.weatherRepo.GetCurrent(ctx, locationID)
	if err == nil && cachedCurrent != nil {
		// Check if cached data is fresh (created/updated less than 1 hour ago)
		age := time.Since(cachedCurrent.CreatedAt)
		if age < 1*time.Hour {
			log.Printf("Using cached weather data for location %d (age: %v)", locationID, age.Round(time.Minute))
			current = cachedCurrent

			// Also get cached forecast data (next 7 days)
			hourlyForecast, err = s.weatherRepo.GetForecast(ctx, locationID, 168) // 7 days
			if err != nil {
				log.Printf("Warning: failed to get forecast from cache: %v", err)
				hourlyForecast = []models.WeatherData{}
			}

			// We don't have cached sun times, but that's okay - we can skip for cached responses
			sunTimes = nil
		}
	}

	// 3. If no cached data or data is stale, fetch from API and persist to DB
	if current == nil {
		log.Printf("Cache miss or stale data, fetching fresh weather for location %d", locationID)
		var fetchErr error
		current, hourlyForecast, sunTimes, fetchErr = s.weatherClient.GetCurrentAndForecast(
			location.Latitude, location.Longitude,
		)
		if fetchErr != nil {
			return nil, fmt.Errorf("failed to fetch weather: %w", fetchErr)
		}

		// Persist fresh data to DB so subsequent requests use the cache.
		// Delete stale future forecasts first, then save all hourly data.
		current.LocationID = locationID
		if err := s.weatherRepo.DeleteFutureForLocation(ctx, locationID); err != nil {
			log.Printf("Warning: failed to purge stale forecasts for location %d: %v", locationID, err)
		}
		if err := s.weatherRepo.Save(ctx, current); err != nil {
			log.Printf("Warning: failed to save current weather for location %d: %v", locationID, err)
		}
		for i := range hourlyForecast {
			hourlyForecast[i].LocationID = locationID
			if err := s.weatherRepo.Save(ctx, &hourlyForecast[i]); err != nil {
				log.Printf("Warning: failed to save forecast hour for location %d: %v", locationID, err)
			}
		}
		log.Printf("Persisted fresh weather data for location %d (%d hours)", locationID, len(hourlyForecast))
	}

	// Extract sunrise/sunset
	var sunrise, sunset string
	if sunTimes != nil {
		sunrise = sunTimes.Sunrise
		sunset = sunTimes.Sunset
	}

	// Set location IDs for the weather data
	current.LocationID = locationID
	for i := range hourlyForecast {
		hourlyForecast[i].LocationID = locationID
	}

	// 3. Get recent hourly historical data for near-term/UI logic.
	historical, err := s.weatherRepo.GetHistorical(ctx, locationID, 7) // 7 days hourly
	if err != nil {
		log.Printf("Warning: failed to get historical weather: %v", err)
		historical = []models.WeatherData{}
	}

	// 4. Build analytics history: hourly for last 30 days + expanded daily aggregates before that.
	analyticsHistorical, err := s.getHistoricalForAnalytics(ctx, locationID, 90)
	if err != nil {
		log.Printf("Warning: failed to build analytics historical weather: %v", err)
		analyticsHistorical = historical
	}

	// 5. Calculate snow depth
	snowDepth := s.calculateSnowDepth(location, current, hourlyForecast, analyticsHistorical)
	dailySnowDepth := s.calculateDailySnowDepth(location, current, hourlyForecast, analyticsHistorical)

	// 6. Calculate rock drying status (use high-fidelity recent hourly history)
	rockStatus, err := s.calculateRockDryingStatus(ctx, location, current, historical, snowDepth)
	if err != nil {
		log.Printf("Warning: failed to calculate rock drying: %v", err)
	}

	// 7. Calculate climbing conditions
	conditionCalc := &weatherPkg.ConditionCalculator{}
	todayCondition := conditionCalc.CalculateTodayCondition(current, hourlyForecast, historical)
	rainLast48h := conditionCalc.CalculateRainLast48h(historical, hourlyForecast)
	rainNext48h := s.calculateRainNext48h(hourlyForecast)

	// 8. Calculate pest conditions (use analytics history)
	pestConditions := s.calculatePestConditions(current, analyticsHistorical)

	// 8. Fetch climb history from Mountain Project data (conditionally, as it's expensive)
	var lastClimbedInfo *models.LastClimbedInfo
	var climbHistory []models.ClimbHistoryEntry
	if includeClimbHistory && s.climbTrackingService != nil {
		// Fetch last 5 climbs for timeline display
		history, err := s.climbTrackingService.GetClimbHistoryForLocation(ctx, locationID, 5)
		if err != nil {
			log.Printf("Warning: failed to get climb history for location %d: %v", locationID, err)
		} else {
			climbHistory = history
			// Also populate the deprecated lastClimbedInfo field for backwards compatibility
			if len(history) > 0 {
				first := history[0]
				lastClimbedInfo = &models.LastClimbedInfo{
					RouteName:      first.RouteName,
					RouteRating:    first.RouteRating,
					ClimbedAt:      first.ClimbedAt,
					ClimbedBy:      first.ClimbedBy,
					Style:          first.Style,
					Comment:        first.Comment,
					DaysSinceClimb: first.DaysSinceClimb,
				}
			}
		}
	}

	// 9. Build response with sun times
	var dailySunTimes []models.DailySunTimes
	if sunTimes != nil && len(sunTimes.Daily) > 0 {
		for _, st := range sunTimes.Daily {
			dailySunTimes = append(dailySunTimes, models.DailySunTimes{
				Date:    st.Date,
				Sunrise: st.Sunrise,
				Sunset:  st.Sunset,
			})
		}
	} else {
		// Cache path fallback: compute sunrise/sunset locally so frontend still gets daily sun times
		dailySunTimes = buildDailySunTimesFallback(location.Latitude, location.Longitude, 7)
		if sunrise == "" && len(dailySunTimes) > 0 {
			sunrise = dailySunTimes[0].Sunrise
			sunset = dailySunTimes[0].Sunset
		}
	}

	forecast := &models.WeatherForecast{
		LocationID:       locationID,
		Location:         *location,
		Current:          *current,
		Hourly:           hourlyForecast,
		Historical:       historical,
		Sunrise:          sunrise,
		Sunset:           sunset,
		DailySunTimes:    dailySunTimes,
		RockDryingStatus: rockStatus,
		SnowDepthInches:  snowDepth,
		DailySnowDepth:   dailySnowDepth,
		TodayCondition:   &todayCondition,
		RainLast48h:      &rainLast48h,
		RainNext48h:      &rainNext48h,
		PestConditions:   pestConditions,
		LastClimbedInfo:  lastClimbedInfo,
		ClimbHistory:     climbHistory,
	}

	return forecast, nil
}

func buildDailySunTimesFallback(lat, lon float64, days int) []models.DailySunTimes {
	if days <= 0 {
		return nil
	}

	pacificTZ, err := time.LoadLocation("America/Los_Angeles")
	if err != nil || pacificTZ == nil {
		pacificTZ = time.UTC
	}

	nowLocal := time.Now().In(pacificTZ)
	daily := make([]models.DailySunTimes, 0, days)
	for i := 0; i < days; i++ {
		day := nowLocal.AddDate(0, 0, i)
		sunrise, sunset := sunpkg.GetSunriseAndSunset(lat, lon, day)
		daily = append(daily, models.DailySunTimes{
			Date:    day.Format("2006-01-02"),
			Sunrise: sunrise.Format(time.RFC3339),
			Sunset:  sunset.Format(time.RFC3339),
		})
	}

	return daily
}

// calculateSnowDepth calculates current snow depth on ground
func (s *WeatherService) calculateSnowDepth(
	location *models.Location,
	current *models.WeatherData,
	hourly []models.WeatherData,
	historical []models.WeatherData,
) *float64 {
	// Combine current and hourly into future data
	futureData := append([]models.WeatherData{*current}, hourly...)

	// Calculate snow depth
	snowDepth := calculator.GetCurrentSnowDepth(historical, futureData, float64(location.ElevationFt))

	// Return snow depth (even if zero, for visibility)
	return &snowDepth
}

// calculateDailySnowDepth calculates daily snow depth forecast for 6 days
func (s *WeatherService) calculateDailySnowDepth(
	location *models.Location,
	current *models.WeatherData,
	hourly []models.WeatherData,
	historical []models.WeatherData,
) map[string]float64 {
	// Combine current and hourly into future data
	futureData := append([]models.WeatherData{*current}, hourly...)

	// Calculate daily snow depth forecast
	dailySnowDepth := calculator.CalculateSnowAccumulation(historical, futureData, float64(location.ElevationFt))

	// Ensure today's date is in the map by using current snow depth
	// This fixes the issue where today's date might be missing or 0 if forecast starts tomorrow
	pacificTZ, _ := time.LoadLocation("America/Los_Angeles")
	if pacificTZ == nil {
		pacificTZ = time.UTC
	}
	todayKey := current.Timestamp.In(pacificTZ).Format("2006-01-02")

	// If today is missing or zero, calculate current snow depth and add it
	if depth, exists := dailySnowDepth[todayKey]; !exists || depth == 0 {
		currentSnowDepth := calculator.GetCurrentSnowDepth(historical, futureData, float64(location.ElevationFt))
		if currentSnowDepth > 0 {
			dailySnowDepth[todayKey] = currentSnowDepth
			if location.ID == 1 {
				log.Printf("  Added missing today (%s) with current snow depth: %.2f\"", todayKey, currentSnowDepth)
			}
		}
	}

	return dailySnowDepth
}

// calculateRainNext48h calculates forecast rain in next 48 hours
func (s *WeatherService) calculateRainNext48h(hourly []models.WeatherData) float64 {
	now := time.Now()
	fortyEightHoursFromNow := now.Add(48 * time.Hour)
	total := 0.0
	for _, h := range hourly {
		if h.Timestamp.After(now) && (h.Timestamp.Before(fortyEightHoursFromNow) || h.Timestamp.Equal(fortyEightHoursFromNow)) {
			total += h.Precipitation
		}
	}
	return total
}

// calculateRockDryingStatus computes rock drying status
func (s *WeatherService) calculateRockDryingStatus(
	ctx context.Context,
	location *models.Location,
	current *models.WeatherData,
	historical []models.WeatherData,
	snowDepth *float64,
) (*models.RockDryingStatus, error) {
	// Get rock types
	rockTypes, err := s.rocksRepo.GetRockTypesByLocation(ctx, location.ID)
	if err != nil || len(rockTypes) == 0 {
		return nil, fmt.Errorf("no rock types for location")
	}

	// Get sun exposure
	sunExposure, _ := s.rocksRepo.GetSunExposureByLocation(ctx, location.ID)

	// Calculate with full rock type data
	status := s.rockCalculator.CalculateDryingStatus(
		rockTypes,
		current,
		historical,
		sunExposure,
		location.HasSeepageRisk,
		snowDepth,
	)

	return &status, nil
}

// IsWeatherDataFresh checks if weather data is less than the specified duration old
func (s *WeatherService) IsWeatherDataFresh(ctx context.Context, maxAge time.Duration) (bool, error) {
	locations, err := s.locationsRepo.GetAll(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get locations: %w", err)
	}

	// Check if all locations have recent weather data
	for _, loc := range locations {
		weather, err := s.weatherRepo.GetCurrent(ctx, loc.ID)
		if err != nil || weather == nil {
			// No weather data exists for this location
			return false, nil
		}

		// Check if weather data was fetched/created too long ago
		age := time.Since(weather.CreatedAt)
		if age > maxAge {
			return false, nil
		}
	}

	return true, nil
}

// RefreshAllWeather refreshes weather for all locations (background job)
// Set forceRefresh=true to bypass freshness check
func (s *WeatherService) RefreshAllWeather(ctx context.Context) error {
	return s.RefreshAllWeatherWithOptions(ctx, false)
}

// RefreshAllWeatherWithOptions refreshes weather with control over freshness check
func (s *WeatherService) RefreshAllWeatherWithOptions(ctx context.Context, forceRefresh bool) error {
	s.refreshMutex.Lock()
	if s.isRefreshing {
		s.refreshMutex.Unlock()
		return fmt.Errorf("refresh already in progress")
	}

	// Check if data is fresh (less than 1 hour old) unless force refresh
	if !forceRefresh {
		s.refreshMutex.Unlock() // Unlock before checking database
		isFresh, err := s.IsWeatherDataFresh(ctx, 1*time.Hour)
		if err != nil {
			log.Printf("Warning: Failed to check weather data freshness: %v", err)
		} else if isFresh {
			log.Println("Weather data is fresh (less than 1 hour old), skipping refresh")
			return nil
		}
		s.refreshMutex.Lock()
	}

	s.isRefreshing = true
	s.refreshMutex.Unlock()

	defer func() {
		s.refreshMutex.Lock()
		s.isRefreshing = false
		s.lastRefresh = time.Now()
		s.refreshMutex.Unlock()
	}()

	locations, err := s.locationsRepo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to get locations: %w", err)
	}

	log.Printf("Refreshing weather for %d locations...", len(locations))

	pacificTZ, tzErr := time.LoadLocation("America/Los_Angeles")
	if tzErr != nil || pacificTZ == nil {
		pacificTZ = time.UTC
	}
	aggregateEndDate := time.Now().In(pacificTZ).Format("2006-01-02")
	aggregateStartDate := time.Now().In(pacificTZ).AddDate(0, 0, -35).Format("2006-01-02")

	for _, loc := range locations {
		// Fetch and save historical weather data (last 7 days) to database
		// This ensures rain_last_48h calculations use fresh data
		historical, err := s.weatherClient.GetHistoricalWeather(loc.Latitude, loc.Longitude, 7)
		if err != nil {
			log.Printf("Failed to fetch historical weather for location %d: %v", loc.ID, err)
		} else {
			// Delete old hourly weather data (older than 30 days)
			log.Printf("Deleting old weather data for location %d (keeping last 30 days)", loc.ID)
			if err := s.weatherRepo.DeleteOldForLocation(ctx, loc.ID, 30); err != nil {
				log.Printf("ERROR: failed to delete old weather data for location %d: %v", loc.ID, err)
			} else {
				log.Printf("Successfully deleted old weather data for location %d", loc.ID)
			}

			// Save historical data to database
			for i := range historical {
				historical[i].LocationID = loc.ID
				if err := s.weatherRepo.Save(ctx, &historical[i]); err != nil {
					log.Printf("Failed to save historical weather for location %d: %v", loc.ID, err)
				}
			}
			log.Printf("Updated historical weather for location %d (%d hours)", loc.ID, len(historical))
		}

		// Fetch and save forecast data (next 16 days) to database
		// This is CRITICAL for boulder drying 6-day forecasts to work
		forecast, err := s.weatherClient.GetForecast(loc.Latitude, loc.Longitude)
		if err != nil {
			log.Printf("Failed to fetch forecast weather for location %d: %v", loc.ID, err)
		} else {
			// CRITICAL: Delete ALL future forecast data before saving fresh data.
			// This prevents stale forecasts from persisting when timestamps don't
			// exactly match (e.g., due to previous timezone bugs or model changes).
			if err := s.weatherRepo.DeleteFutureForLocation(ctx, loc.ID); err != nil {
				log.Printf("ERROR: failed to delete future weather data for location %d: %v", loc.ID, err)
			} else {
				log.Printf("Purged stale future forecasts for location %d", loc.ID)
			}

			// Save fresh forecast data to database
			for i := range forecast {
				forecast[i].LocationID = loc.ID
				if err := s.weatherRepo.Save(ctx, &forecast[i]); err != nil {
					log.Printf("Failed to save forecast weather for location %d: %v", loc.ID, err)
				}
			}
			log.Printf("Updated forecast weather for location %d (%d hours)", loc.ID, len(forecast))
		}

		if err := s.weatherRepo.UpsertDailyAggregates(ctx, loc.ID, aggregateStartDate, aggregateEndDate); err != nil {
			log.Printf("Failed to upsert daily weather aggregates for location %d: %v", loc.ID, err)
		}

		// Fetch current/forecast weather (this also triggers calculations)
		if _, err := s.GetLocationWeather(ctx, loc.ID); err != nil {
			log.Printf("Failed to refresh location %d: %v", loc.ID, err)
		}
	}

	return nil
}

// StartBackgroundRefresh starts automatic weather refresh
func (s *WeatherService) StartBackgroundRefresh(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			if err := s.RefreshAllWeather(ctx); err != nil {
				log.Printf("Background refresh failed: %v", err)
			}
			cancel()
		}
	}()
}

// GetWeatherByCoordinates fetches weather for arbitrary coordinates
func (s *WeatherService) GetWeatherByCoordinates(ctx context.Context, lat, lon float64) (*models.WeatherForecast, error) {
	// Fetch weather from API
	current, hourlyForecast, sunTimes, err := s.weatherClient.GetCurrentAndForecast(lat, lon)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather: %w", err)
	}

	// Extract sunrise/sunset
	var sunrise, sunset string
	if sunTimes != nil {
		sunrise = sunTimes.Sunrise
		sunset = sunTimes.Sunset
	}

	// Build response (no location, no rock drying)
	forecast := &models.WeatherForecast{
		Current:    *current,
		Hourly:     hourlyForecast,
		Historical: []models.WeatherData{},
		Sunrise:    sunrise,
		Sunset:     sunset,
	}

	return forecast, nil
}

// GetAllWeather retrieves weather for all locations or filtered by area
// Fetches weather data concurrently for better performance
func (s *WeatherService) GetAllWeather(ctx context.Context, areaID *int) ([]models.WeatherForecast, error) {
	start := time.Now()
	var locations []models.Location
	var err error

	if areaID != nil {
		locations, err = s.locationsRepo.GetByArea(ctx, *areaID)
	} else {
		locations, err = s.locationsRepo.GetAll(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get locations: %w", err)
	}

	// Return empty if no locations
	if len(locations) == 0 {
		return []models.WeatherForecast{}, nil
	}
	log.Printf("GetAllWeather: Got %d locations in %v", len(locations), time.Since(start))

	// PERFORMANCE OPTIMIZATION: Batch fetch climb history for all locations in a single query
	// This replaces N separate queries with 1 batch query, significantly reducing database round trips
	// Original: 9 locations × 1 query each = 9 database queries (slow!)
	// Optimized: 1 batch query for all locations = 1 database query (fast!)
	climbStart := time.Now()
	climbHistoryMap := make(map[int][]models.ClimbHistoryEntry)
	if s.climbTrackingService != nil {
		locationIDs := make([]int, len(locations))
		for i, loc := range locations {
			locationIDs[i] = loc.ID
		}
		climbHistoryMap, err = s.climbTrackingService.GetClimbHistoryForLocations(ctx, locationIDs, 5)
		if err != nil {
			log.Printf("Warning: failed to batch fetch climb history: %v", err)
			// Don't fail the whole request, just proceed without climb history
			climbHistoryMap = make(map[int][]models.ClimbHistoryEntry)
		}
	}
	log.Printf("GetAllWeather: Fetched climb history in %v", time.Since(climbStart))

	// Build the forecasts concurrently with a limited number of workers
	// to balance parallelism and database connection pressure
	forecastStart := time.Now()
	type result struct {
		forecast *models.WeatherForecast
		err      error
	}

	results := make(chan result, len(locations))
	var wg sync.WaitGroup
	// Use a semaphore to limit concurrent workers
	// Since weather data is usually cached, we can be more aggressive with parallelism
	sem := make(chan struct{}, 10) // Max 10 concurrent workers (covers all PNW locations)

	for _, loc := range locations {
		wg.Add(1)
		go func(locationID int, climbHistory []models.ClimbHistoryEntry) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			forecast, err := s.getLocationWeatherWithClimbHistory(ctx, locationID, climbHistory)
			results <- result{forecast: forecast, err: err}
		}(loc.ID, climbHistoryMap[loc.ID])
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var forecasts []models.WeatherForecast
	for res := range results {
		if res.err != nil {
			log.Printf("Warning: failed to get weather for location: %v", res.err)
			continue
		}
		if res.forecast != nil {
			forecasts = append(forecasts, *res.forecast)
		}
	}
	log.Printf("GetAllWeather: Built %d forecasts in %v", len(forecasts), time.Since(forecastStart))
	log.Printf("GetAllWeather: Total time %v", time.Since(start))

	return forecasts, nil
}

// getLocationWeatherWithClimbHistory is a helper that fetches weather with pre-fetched climb history
// This is used by GetAllWeather to avoid N+1 queries when fetching multiple locations
func (s *WeatherService) getLocationWeatherWithClimbHistory(ctx context.Context, locationID int, climbHistory []models.ClimbHistoryEntry) (*models.WeatherForecast, error) {
	// Get all the weather data using the standard method without climb history
	forecast, err := s.getLocationWeatherWithOptions(ctx, locationID, false)
	if err != nil {
		return nil, err
	}

	// Inject the pre-fetched climb history
	forecast.ClimbHistory = climbHistory

	// Populate the deprecated lastClimbedInfo field for backwards compatibility
	if len(climbHistory) > 0 {
		first := climbHistory[0]
		forecast.LastClimbedInfo = &models.LastClimbedInfo{
			RouteName:      first.RouteName,
			RouteRating:    first.RouteRating,
			ClimbedAt:      first.ClimbedAt,
			ClimbedBy:      first.ClimbedBy,
			Style:          first.Style,
			Comment:        first.Comment,
			DaysSinceClimb: first.DaysSinceClimb,
		}
	}

	return forecast, nil
}

// calculatePestConditions calculates pest activity levels based on current and historical weather
func (s *WeatherService) calculatePestConditions(
	current *models.WeatherData,
	historical []models.WeatherData,
) *models.PestConditions {
	// Assess pest conditions using the pest analyzer
	result := s.pestAnalyzer.AssessConditions(current, historical)

	// Convert to models.PestConditions
	return &models.PestConditions{
		MosquitoLevel:    string(result.MosquitoLevel),
		MosquitoScore:    result.MosquitoScore,
		OutdoorPestLevel: string(result.OutdoorPestLevel),
		OutdoorPestScore: result.OutdoorPestScore,
		Factors:          result.Factors,
	}
}

// getHistoricalForAnalytics builds a hybrid history timeline:
// - hourly rows for the most recent 30 days
// - expanded daily aggregates for older days in the requested window
func (s *WeatherService) getHistoricalForAnalytics(ctx context.Context, locationID int, totalDays int) ([]models.WeatherData, error) {
	if totalDays <= 0 {
		return []models.WeatherData{}, nil
	}

	hourlyDays := totalDays
	if hourlyDays > 30 {
		hourlyDays = 30
	}

	recentHourly, err := s.weatherRepo.GetHistorical(ctx, locationID, hourlyDays)
	if err != nil {
		return nil, err
	}

	if totalDays <= 30 {
		return recentHourly, nil
	}

	pacificTZ, tzErr := time.LoadLocation("America/Los_Angeles")
	if tzErr != nil || pacificTZ == nil {
		pacificTZ = time.UTC
	}

	nowLocal := time.Now().In(pacificTZ)
	aggStartDate := nowLocal.AddDate(0, 0, -totalDays).Format("2006-01-02")
	aggEndDate := nowLocal.AddDate(0, 0, -31).Format("2006-01-02")

	dailyAgg, err := s.weatherRepo.GetDailyAggregates(ctx, locationID, aggStartDate, aggEndDate)
	if err != nil {
		return nil, err
	}

	hybrid := make([]models.WeatherData, 0, len(recentHourly)+(len(dailyAgg)*24))
	hybrid = append(hybrid, recentHourly...)

	for _, agg := range dailyAgg {
		hybrid = append(hybrid, expandDailyAggregateToHourly(agg, pacificTZ)...)
	}

	sort.Slice(hybrid, func(i, j int) bool {
		return hybrid[i].Timestamp.Before(hybrid[j].Timestamp)
	})

	return hybrid, nil
}

func expandDailyAggregateToHourly(agg models.WeatherDailyAggregate, loc *time.Location) []models.WeatherData {
	base, err := time.ParseInLocation("2006-01-02", agg.LocalDate, loc)
	if err != nil {
		return []models.WeatherData{}
	}

	hours := 24
	if agg.SourceHourCount > 0 && agg.SourceHourCount < 24 {
		hours = agg.SourceHourCount
	}

	precipPerHour := 0.0
	if hours > 0 {
		precipPerHour = agg.TotalPrecipitation / float64(hours)
	}

	humidity := int(math.Round(agg.AvgHumidity))
	if humidity < 0 {
		humidity = 0
	}
	if humidity > 100 {
		humidity = 100
	}

	result := make([]models.WeatherData, 0, hours)
	for h := 0; h < hours; h++ {
		ts := time.Date(base.Year(), base.Month(), base.Day(), h, 0, 0, 0, loc).UTC()
		result = append(result, models.WeatherData{
			LocationID:    agg.LocationID,
			Timestamp:     ts,
			Temperature:   agg.AvgTemperature,
			FeelsLike:     agg.AvgTemperature,
			Precipitation: precipPerHour,
			Humidity:      humidity,
			WindSpeed:     agg.AvgWindSpeed,
			WindDirection: 0,
			CloudCover:    0,
			Pressure:      0,
			Description:   "daily aggregate expanded",
			Icon:          "",
		})
	}

	return result
}

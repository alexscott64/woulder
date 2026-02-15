package service

import (
	"context"
	"fmt"
	"log"
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
func (s *WeatherService) GetLocationWeather(ctx context.Context, locationID int) (*models.WeatherForecast, error) {
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
		// Check if cached data is fresh (less than 1 hour old)
		age := time.Since(cachedCurrent.Timestamp)
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

	// 3. If no cached data or data is stale, fetch from API
	if current == nil {
		log.Printf("Cache miss or stale data, fetching fresh weather for location %d", locationID)
		var fetchErr error
		current, hourlyForecast, sunTimes, fetchErr = s.weatherClient.GetCurrentAndForecast(
			location.Latitude, location.Longitude,
		)
		if fetchErr != nil {
			return nil, fmt.Errorf("failed to fetch weather: %w", fetchErr)
		}
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

	// 3. Get historical data from database for rock drying and snow calculation
	// Get 7 days (168 hours) for better snow accumulation tracking
	historical, err := s.weatherRepo.GetHistorical(ctx, locationID, 7) // 7 days
	if err != nil {
		log.Printf("Warning: failed to get historical weather: %v", err)
		historical = []models.WeatherData{}
	}

	// 4. Calculate snow depth
	snowDepth := s.calculateSnowDepth(location, current, hourlyForecast, historical)
	dailySnowDepth := s.calculateDailySnowDepth(location, current, hourlyForecast, historical)

	// 5. Calculate rock drying status
	rockStatus, err := s.calculateRockDryingStatus(ctx, location, current, historical, snowDepth)
	if err != nil {
		log.Printf("Warning: failed to calculate rock drying: %v", err)
	}

	// 6. Calculate climbing conditions
	conditionCalc := &weatherPkg.ConditionCalculator{}
	todayCondition := conditionCalc.CalculateTodayCondition(current, hourlyForecast, historical)
	rainLast48h := conditionCalc.CalculateRainLast48h(historical, hourlyForecast)
	rainNext48h := s.calculateRainNext48h(hourlyForecast)

	// 7. Calculate pest conditions
	pestConditions := s.calculatePestConditions(current, historical)

	// 8. Fetch climb history from Mountain Project data
	var lastClimbedInfo *models.LastClimbedInfo
	var climbHistory []models.ClimbHistoryEntry
	if s.climbTrackingService != nil {
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

	// 9. Build response with fresh API data
	var dailySunTimes []models.DailySunTimes
	if sunTimes != nil && len(sunTimes.Daily) > 0 {
		for _, st := range sunTimes.Daily {
			dailySunTimes = append(dailySunTimes, models.DailySunTimes{
				Date:    st.Date,
				Sunrise: st.Sunrise,
				Sunset:  st.Sunset,
			})
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

		// Check if weather data is too old
		age := time.Since(weather.Timestamp)
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

	for _, loc := range locations {
		// Fetch and save historical weather data (last 7 days) to database
		// This ensures rain_last_48h calculations use fresh data
		historical, err := s.weatherClient.GetHistoricalWeather(loc.Latitude, loc.Longitude, 7)
		if err != nil {
			log.Printf("Failed to fetch historical weather for location %d: %v", loc.ID, err)
		} else {
			// Delete old historical data (older than 7 days) to prevent stale precipitation data
			// This ensures only fresh API data is used for rain calculations
			log.Printf("Deleting old weather data for location %d (keeping last 7 days)", loc.ID)
			if err := s.weatherRepo.DeleteOldForLocation(ctx, loc.ID, 7); err != nil {
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

		// Fetch and save forecast data (next 7 days) to database
		// This is CRITICAL for boulder drying 6-day forecasts to work
		forecast, err := s.weatherClient.GetForecast(loc.Latitude, loc.Longitude)
		if err != nil {
			log.Printf("Failed to fetch forecast weather for location %d: %v", loc.ID, err)
		} else {
			// Save forecast data to database
			for i := range forecast {
				forecast[i].LocationID = loc.ID
				if err := s.weatherRepo.Save(ctx, &forecast[i]); err != nil {
					log.Printf("Failed to save forecast weather for location %d: %v", loc.ID, err)
				}
			}
			log.Printf("Updated forecast weather for location %d (%d hours)", loc.ID, len(forecast))
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

	// Fetch weather concurrently for all locations
	type result struct {
		forecast *models.WeatherForecast
		err      error
	}

	results := make(chan result, len(locations))
	var wg sync.WaitGroup

	for _, loc := range locations {
		wg.Add(1)
		go func(locationID int) {
			defer wg.Done()
			forecast, err := s.GetLocationWeather(ctx, locationID)
			results <- result{forecast: forecast, err: err}
		}(loc.ID)
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

	return forecasts, nil
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

package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/alexscott64/woulder/backend/internal/database"
	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/alexscott64/woulder/backend/internal/weather"
	"github.com/alexscott64/woulder/backend/internal/weather/calculator"
)

type WeatherService struct {
	repo           database.Repository
	weatherClient  *weather.WeatherService
	rockCalculator *weather.RockDryingCalculator

	// Background refresh management
	refreshMutex sync.Mutex
	lastRefresh  time.Time
	isRefreshing bool
}

func NewWeatherService(repo database.Repository, client *weather.WeatherService) *WeatherService {
	return &WeatherService{
		repo:           repo,
		weatherClient:  client,
		rockCalculator: &weather.RockDryingCalculator{},
	}
}

// GetLocationWeather retrieves complete weather forecast for a location
func (s *WeatherService) GetLocationWeather(ctx context.Context, locationID int) (*models.WeatherForecast, error) {
	// 1. Get location
	location, err := s.repo.GetLocation(ctx, locationID)
	if err != nil {
		return nil, fmt.Errorf("location not found: %w", err)
	}

	// 2. Fetch weather from API (fresh data, not from DB)
	current, hourlyForecast, sunTimes, err := s.weatherClient.GetCurrentAndForecast(
		location.Latitude, location.Longitude,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather: %w", err)
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
	historical, err := s.repo.GetHistoricalWeather(ctx, locationID, 168)
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

	// 6. Build response with fresh API data
	forecast := &models.WeatherForecast{
		LocationID:       locationID,
		Location:         *location,
		Current:          *current,
		Hourly:           hourlyForecast,
		Historical:       historical,
		Sunrise:          sunrise,
		Sunset:           sunset,
		RockDryingStatus: rockStatus,
		SnowDepthInches:  snowDepth,
		DailySnowDepth:   dailySnowDepth,
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

	// DEBUG: Log data being used (only for Money Creek - location ID 1)
	if location.ID == 1 {
		log.Printf("[SNOW DEBUG] calculateSnowDepth for location %d (%s)", location.ID, location.Name)
		log.Printf("  Historical count: %d", len(historical))
		if len(historical) > 0 {
			log.Printf("  Historical range: %s to %s", historical[0].Timestamp, historical[len(historical)-1].Timestamp)
		}
		log.Printf("  FutureData count: %d", len(futureData))
		if len(futureData) > 0 {
			log.Printf("  FutureData range: %s to %s", futureData[0].Timestamp, futureData[len(futureData)-1].Timestamp)
		}
	}

	// Calculate snow depth
	snowDepth := calculator.GetCurrentSnowDepth(historical, futureData, float64(location.ElevationFt))

	if location.ID == 1 {
		log.Printf("  Result: %.2f\"", snowDepth)
	}

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

	// DEBUG: Log data being used (only for Money Creek - location ID 1)
	if location.ID == 1 {
		log.Printf("[SNOW DEBUG] calculateDailySnowDepth for location %d (%s)", location.ID, location.Name)
		log.Printf("  Historical count: %d", len(historical))
		if len(historical) > 0 {
			log.Printf("  Historical range: %s to %s", historical[0].Timestamp, historical[len(historical)-1].Timestamp)
		}
		log.Printf("  FutureData count: %d", len(futureData))
		if len(futureData) > 0 {
			log.Printf("  FutureData range: %s to %s", futureData[0].Timestamp, futureData[len(futureData)-1].Timestamp)
		}
	}

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

	// DEBUG: Log results (only for Money Creek - location ID 1)
	if location.ID == 1 {
		log.Printf("  Daily snow depth map:")
		for date, depth := range dailySnowDepth {
			log.Printf("    %s: %.2f\"", date, depth)
		}
	}

	return dailySnowDepth
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
	rockTypes, err := s.repo.GetRockTypesByLocation(ctx, location.ID)
	if err != nil || len(rockTypes) == 0 {
		return nil, fmt.Errorf("no rock types for location")
	}

	// Get sun exposure
	sunExposure, _ := s.repo.GetSunExposureByLocation(ctx, location.ID)

	// Calculate
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

// RefreshAllWeather refreshes weather for all locations (background job)
func (s *WeatherService) RefreshAllWeather(ctx context.Context) error {
	s.refreshMutex.Lock()
	if s.isRefreshing {
		s.refreshMutex.Unlock()
		return fmt.Errorf("refresh already in progress")
	}
	s.isRefreshing = true
	s.refreshMutex.Unlock()

	defer func() {
		s.refreshMutex.Lock()
		s.isRefreshing = false
		s.lastRefresh = time.Now()
		s.refreshMutex.Unlock()
	}()

	locations, err := s.repo.GetAllLocations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get locations: %w", err)
	}

	log.Printf("Refreshing weather for %d locations...", len(locations))

	for _, loc := range locations {
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
		locations, err = s.repo.GetLocationsByArea(ctx, *areaID)
	} else {
		locations, err = s.repo.GetAllLocations(ctx)
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

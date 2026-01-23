package service

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/alexscott64/woulder/backend/internal/database"
	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/alexscott64/woulder/backend/internal/weather"
	"github.com/alexscott64/woulder/backend/internal/weather/boulder_drying"
	"github.com/alexscott64/woulder/backend/internal/weather/rock_drying"
)

// BoulderDryingService handles boulder-specific drying calculations
type BoulderDryingService struct {
	repo          database.Repository
	calculator    *boulder_drying.Calculator
	weatherClient *weather.WeatherService
}

// NewBoulderDryingService creates a new boulder drying service
func NewBoulderDryingService(repo database.Repository, weatherClient *weather.WeatherService) *BoulderDryingService {
	// Get API key from environment
	apiKey := os.Getenv("IPGEOLOCATION_API_KEY")
	if apiKey == "" {
		log.Printf("Warning: IPGEOLOCATION_API_KEY not set - sun exposure calculations will use fallback estimates")
	}

	return &BoulderDryingService{
		repo:          repo,
		calculator:    boulder_drying.NewCalculator(apiKey),
		weatherClient: weatherClient,
	}
}

// GetBoulderDryingStatus calculates the drying status for a specific boulder
func (s *BoulderDryingService) GetBoulderDryingStatus(
	ctx context.Context,
	mpRouteID string,
) (*boulder_drying.BoulderDryingStatus, error) {
	// Get the route
	route, err := s.repo.GetMPRouteByID(ctx, mpRouteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get route: %w", err)
	}
	if route == nil {
		return nil, fmt.Errorf("route not found: %s", mpRouteID)
	}

	// Get the location ID
	if route.LocationID == nil {
		return nil, fmt.Errorf("route has no associated location")
	}

	// Get location-level rock drying status
	locationDrying, err := s.getLocationRockDryingStatus(ctx, *route.LocationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get location drying status: %w", err)
	}

	// Get boulder drying profile (if exists)
	profile, err := s.repo.GetBoulderDryingProfile(ctx, mpRouteID)
	if err != nil {
		log.Printf("Warning: Failed to get boulder drying profile for %s: %v", mpRouteID, err)
		profile = nil // Continue without profile
	}

	// Get location sun exposure (for tree coverage)
	sunExposure, err := s.repo.GetSunExposureByLocation(ctx, *route.LocationID)
	if err != nil {
		log.Printf("Warning: Failed to get sun exposure for location %d: %v", *route.LocationID, err)
		sunExposure = nil
	}

	// Extract location tree coverage
	locationTreeCoverage := 0.0
	if sunExposure != nil {
		locationTreeCoverage = sunExposure.TreeCoveragePercent
	}

	// Get hourly forecast for 6-day forecast
	hourlyForecast, err := s.repo.GetForecastWeather(ctx, *route.LocationID, 144) // 6 days = 144 hours
	if err != nil {
		log.Printf("Warning: Failed to get hourly forecast for location %d: %v", *route.LocationID, err)
		hourlyForecast = nil // Continue without forecast
	}

	// Calculate boulder-specific drying status
	status, err := s.calculator.CalculateBoulderDryingStatus(ctx, route, locationDrying, profile, locationTreeCoverage, hourlyForecast)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate boulder drying status: %w", err)
	}

	// Save/update cache if we have GPS and calculated sun exposure
	if status.Latitude != 0 && status.Longitude != 0 && status.SunExposureHours > 0 {
		cacheData, err := s.calculator.GetSunExposureCacheData(
			status.Aspect,
			status.TreeCoveragePercent,
			status.SunExposureHours,
		)
		if err == nil {
			// Convert []byte to string for JSONB storage
			cacheStr := string(cacheData)

			// Save or update boulder drying profile with cache
			profileToSave := &models.BoulderDryingProfile{
				MPRouteID:             mpRouteID,
				TreeCoveragePercent:   &status.TreeCoveragePercent,
				SunExposureHoursCache: &cacheStr,
			}
			if err := s.repo.SaveBoulderDryingProfile(ctx, profileToSave); err != nil {
				log.Printf("Warning: Failed to save boulder drying profile cache: %v", err)
			} else {
				log.Printf("Saved sun exposure cache for route %s", mpRouteID)
			}
		}
	}

	return status, nil
}

// getLocationRockDryingStatus calculates location-level rock drying status
func (s *BoulderDryingService) getLocationRockDryingStatus(
	ctx context.Context,
	locationID int,
) (*models.RockDryingStatus, error) {
	// Get location to fetch coordinates
	location, err := s.repo.GetLocation(ctx, locationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get location: %w", err)
	}

	// Get FRESH current weather from API (not stale database cache)
	currentWeather, _, _, err := s.weatherClient.GetCurrentAndForecast(location.Latitude, location.Longitude)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch fresh current weather: %w", err)
	}
	currentWeather.LocationID = locationID

	// Get historical weather (last 7 days) from database
	historicalWeather, err := s.repo.GetHistoricalWeather(ctx, locationID, 168) // 7 days
	if err != nil {
		return nil, fmt.Errorf("failed to get historical weather: %w", err)
	}

	// Get rock types
	rockTypes, err := s.repo.GetRockTypesByLocation(ctx, locationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rock types: %w", err)
	}

	// Get sun exposure
	sunExposure, err := s.repo.GetSunExposureByLocation(ctx, locationID)
	if err != nil {
		log.Printf("Warning: Failed to get sun exposure for location %d: %v", locationID, err)
		sunExposure = nil
	}

	// Calculate rock drying status using existing calculator
	calc := &rock_drying.Calculator{}
	dryingStatus := calc.CalculateDryingStatus(
		rockTypes,
		currentWeather,
		historicalWeather,
		sunExposure,
		false, // hasSeepageRisk - TODO: Get from location profile if needed
		nil,   // snowDepthInches - not tracked in WeatherData
	)

	return &dryingStatus, nil
}

// GetAreaDryingStats calculates aggregated drying statistics for an area
func (s *BoulderDryingService) GetAreaDryingStats(
	ctx context.Context,
	mpAreaID string,
	locationID int,
) (*models.AreaDryingStats, error) {
	// Get all routes with GPS in this area (including subareas)
	routes, err := s.repo.GetRoutesWithGPSByArea(ctx, mpAreaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get routes for area: %w", err)
	}

	if len(routes) == 0 {
		// No routes with GPS data - return nil
		return nil, nil
	}

	// Get location-level rock drying status (shared for all routes in location)
	locationDrying, err := s.getLocationRockDryingStatus(ctx, locationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get location drying status: %w", err)
	}

	// Get location sun exposure (for tree coverage fallback)
	sunExposure, err := s.repo.GetSunExposureByLocation(ctx, locationID)
	if err != nil {
		log.Printf("Warning: Failed to get sun exposure for location %d: %v", locationID, err)
		sunExposure = nil
	}

	locationTreeCoverage := 0.0
	if sunExposure != nil {
		locationTreeCoverage = sunExposure.TreeCoveragePercent
	}

	// Get hourly forecast for 6-day forecast (shared across all routes)
	hourlyForecast, err := s.repo.GetForecastWeather(ctx, locationID, 144) // 6 days
	if err != nil {
		log.Printf("Warning: Failed to get hourly forecast for location %d: %v", locationID, err)
		hourlyForecast = nil
	}

	// Calculate drying status for each route and aggregate
	stats := &models.AreaDryingStats{
		TotalRoutes: len(routes),
	}

	var totalHoursUntilDry float64
	var wetRouteCount int
	var totalTreeCoverage float64
	var treeCoverageCount int
	var totalConfidence int

	for _, route := range routes {
		// Get boulder drying profile if exists
		profile, err := s.repo.GetBoulderDryingProfile(ctx, route.MPRouteID)
		if err != nil {
			log.Printf("Warning: Failed to get boulder drying profile for %s: %v", route.MPRouteID, err)
			profile = nil
		}

		// Calculate boulder-specific drying status
		status, err := s.calculator.CalculateBoulderDryingStatus(
			ctx,
			route,
			locationDrying,
			profile,
			locationTreeCoverage,
			hourlyForecast,
		)
		if err != nil {
			log.Printf("Warning: Failed to calculate drying status for %s: %v", route.MPRouteID, err)
			continue
		}

		// Aggregate statistics
		if !status.IsWet {
			stats.DryCount++
		} else if status.HoursUntilDry > 0 && status.HoursUntilDry <= 24 {
			stats.DryingCount++
			totalHoursUntilDry += status.HoursUntilDry
			wetRouteCount++
		} else {
			stats.WetCount++
			totalHoursUntilDry += status.HoursUntilDry
			wetRouteCount++
		}

		// Tree coverage
		if status.TreeCoveragePercent > 0 {
			totalTreeCoverage += status.TreeCoveragePercent
			treeCoverageCount++
		}

		// Confidence
		totalConfidence += status.ConfidenceScore
	}

	// Calculate percentages and averages
	if stats.TotalRoutes > 0 {
		stats.PercentDry = (float64(stats.DryCount) / float64(stats.TotalRoutes)) * 100
	}

	if wetRouteCount > 0 {
		stats.AvgHoursUntilDry = totalHoursUntilDry / float64(wetRouteCount)
	}

	if treeCoverageCount > 0 {
		stats.AvgTreeCoverage = totalTreeCoverage / float64(treeCoverageCount)
	}

	if stats.TotalRoutes > 0 {
		stats.ConfidenceScore = totalConfidence / stats.TotalRoutes
	}

	return stats, nil
}

package service

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/alexscott64/woulder/backend/internal/database"
	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/alexscott64/woulder/backend/internal/weather/boulder_drying"
	"github.com/alexscott64/woulder/backend/internal/weather/rock_drying"
)

// BoulderDryingService handles boulder-specific drying calculations
type BoulderDryingService struct {
	repo       database.Repository
	calculator *boulder_drying.Calculator
}

// NewBoulderDryingService creates a new boulder drying service
func NewBoulderDryingService(repo database.Repository) *BoulderDryingService {
	// Get API key from environment
	apiKey := os.Getenv("IPGEOLOCATION_API_KEY")
	if apiKey == "" {
		log.Printf("Warning: IPGEOLOCATION_API_KEY not set - sun exposure calculations will use fallback estimates")
	}

	return &BoulderDryingService{
		repo:       repo,
		calculator: boulder_drying.NewCalculator(apiKey),
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

	// Calculate boulder-specific drying status
	status, err := s.calculator.CalculateBoulderDryingStatus(ctx, route, locationDrying, profile)
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
	// Get current weather
	currentWeather, err := s.repo.GetCurrentWeather(ctx, locationID)
	if err != nil || currentWeather == nil {
		return nil, fmt.Errorf("failed to get current weather: %w", err)
	}

	// Get historical weather (last 7 days)
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

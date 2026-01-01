package service

import (
	"context"
	"fmt"

	"github.com/alexscott64/woulder/backend/internal/database"
	"github.com/alexscott64/woulder/backend/internal/models"
)

// LocationService handles location-related business logic
type LocationService struct {
	repo database.Repository
}

// NewLocationService creates a new LocationService
func NewLocationService(repo database.Repository) *LocationService {
	return &LocationService{
		repo: repo,
	}
}

// GetAllLocations retrieves all locations
func (s *LocationService) GetAllLocations(ctx context.Context) ([]models.Location, error) {
	locations, err := s.repo.GetAllLocations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all locations: %w", err)
	}
	return locations, nil
}

// GetLocation retrieves a single location by ID
func (s *LocationService) GetLocation(ctx context.Context, id int) (*models.Location, error) {
	location, err := s.repo.GetLocation(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get location %d: %w", id, err)
	}
	return location, nil
}

// GetLocationsByArea retrieves all locations in a specific area
func (s *LocationService) GetLocationsByArea(ctx context.Context, areaID int) ([]models.Location, error) {
	locations, err := s.repo.GetLocationsByArea(ctx, areaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get locations for area %d: %w", areaID, err)
	}
	return locations, nil
}

// GetAllAreas retrieves all areas
func (s *LocationService) GetAllAreas(ctx context.Context) ([]models.Area, error) {
	areas, err := s.repo.GetAllAreas(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all areas: %w", err)
	}
	return areas, nil
}

// GetAreasWithLocationCounts retrieves areas with their location counts
func (s *LocationService) GetAreasWithLocationCounts(ctx context.Context) ([]models.AreaWithLocationCount, error) {
	areas, err := s.repo.GetAreasWithLocationCounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get areas with counts: %w", err)
	}
	return areas, nil
}

// GetAreaByID retrieves a specific area by ID
func (s *LocationService) GetAreaByID(ctx context.Context, id int) (*models.Area, error) {
	area, err := s.repo.GetAreaByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get area %d: %w", id, err)
	}
	return area, nil
}

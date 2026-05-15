package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/alexscott64/woulder/backend/internal/database/areas"
	"github.com/alexscott64/woulder/backend/internal/database/locations"
	"github.com/alexscott64/woulder/backend/internal/geo"
	"github.com/alexscott64/woulder/backend/internal/models"
)

// ErrInvalidTimezone is returned by CreateLocation when the supplied
// Timezone is not a valid IANA timezone name (i.e. time.LoadLocation
// rejects it). Handlers should map this to HTTP 400.
var ErrInvalidTimezone = errors.New("invalid IANA timezone")

// LocationService handles location-related business logic
type LocationService struct {
	locationsRepo locations.Repository
	areasRepo     areas.Repository
}

// NewLocationService creates a new LocationService
func NewLocationService(locationsRepo locations.Repository, areasRepo areas.Repository) *LocationService {
	return &LocationService{
		locationsRepo: locationsRepo,
		areasRepo:     areasRepo,
	}
}

// GetAllLocations retrieves all locations
func (s *LocationService) GetAllLocations(ctx context.Context) ([]models.Location, error) {
	locations, err := s.locationsRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all locations: %w", err)
	}
	return locations, nil
}

// GetLocation retrieves a single location by ID
func (s *LocationService) GetLocation(ctx context.Context, id int) (*models.Location, error) {
	location, err := s.locationsRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get location %d: %w", id, err)
	}
	return location, nil
}

// GetLocationsByArea retrieves all locations in a specific area
func (s *LocationService) GetLocationsByArea(ctx context.Context, areaID int) ([]models.Location, error) {
	locations, err := s.locationsRepo.GetByArea(ctx, areaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get locations for area %d: %w", areaID, err)
	}
	return locations, nil
}

// GetAllAreas retrieves all areas
func (s *LocationService) GetAllAreas(ctx context.Context) ([]models.Area, error) {
	areas, err := s.areasRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all areas: %w", err)
	}
	return areas, nil
}

// GetAreasWithLocationCounts retrieves areas with their location counts
func (s *LocationService) GetAreasWithLocationCounts(ctx context.Context) ([]models.AreaWithLocationCount, error) {
	areas, err := s.areasRepo.GetAllWithLocationCounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get areas with counts: %w", err)
	}
	return areas, nil
}

// GetAreaByID retrieves a specific area by ID
func (s *LocationService) GetAreaByID(ctx context.Context, id int) (*models.Area, error) {
	area, err := s.areasRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get area %d: %w", id, err)
	}
	return area, nil
}

// CreateLocation inserts a new location, deriving its IANA timezone from
// (latitude, longitude) when not supplied.
//
// Behaviour:
//   - If loc.Timezone is empty, it is populated via geo.LookupTimezone (which
//     itself falls back to "America/Los_Angeles" on lookup failure).
//   - If loc.Timezone is non-empty, it is validated via time.LoadLocation;
//     unknown names return ErrInvalidTimezone (handlers should map to HTTP 400).
//
// Returns the newly inserted location's ID.
func (s *LocationService) CreateLocation(ctx context.Context, loc models.Location) (int, error) {
	if loc.Timezone == "" {
		loc.Timezone = geo.LookupTimezone(loc.Latitude, loc.Longitude)
		if loc.Timezone == "" {
			// geo.LookupTimezone is documented to never return "" today,
			// but defend against future behaviour changes.
			log.Printf("WARN: tz lookup failed for (%f,%f); defaulting to America/Los_Angeles", loc.Latitude, loc.Longitude)
			loc.Timezone = "America/Los_Angeles"
		}
	}

	if _, err := time.LoadLocation(loc.Timezone); err != nil {
		return 0, fmt.Errorf("%w: %q", ErrInvalidTimezone, loc.Timezone)
	}

	id, err := s.locationsRepo.Create(ctx, loc)
	if err != nil {
		return 0, fmt.Errorf("failed to create location: %w", err)
	}
	return id, nil
}

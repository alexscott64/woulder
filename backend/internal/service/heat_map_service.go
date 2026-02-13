package service

import (
	"context"
	"fmt"
	"time"

	"github.com/alexscott64/woulder/backend/internal/database"
	"github.com/alexscott64/woulder/backend/internal/models"
)

type HeatMapService struct {
	repo database.Repository
}

func NewHeatMapService(repo database.Repository) *HeatMapService {
	return &HeatMapService{repo: repo}
}

// GetHeatMapData retrieves aggregated activity data for the map
func (s *HeatMapService) GetHeatMapData(
	ctx context.Context,
	startDate, endDate time.Time,
	bounds *database.GeoBounds,
	minActivity, limit int,
) ([]models.HeatMapPoint, error) {
	// Validate inputs
	if startDate.After(endDate) {
		return nil, fmt.Errorf("start_date must be before end_date")
	}

	if minActivity < 1 {
		minActivity = 1
	}

	if limit < 1 || limit > 1000 {
		limit = 500 // Default to 500
	}

	if bounds != nil {
		if err := bounds.Validate(); err != nil {
			return nil, fmt.Errorf("invalid bounds: %w", err)
		}
	}

	// Fetch raw data
	points, err := s.repo.GetHeatMapData(ctx, startDate, endDate, bounds, minActivity, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch heat map data: %w", err)
	}

	// Calculate activity scores (recency weighting)
	for i := range points {
		points[i].ActivityScore = s.calculateActivityScore(
			points[i].TotalTicks,
			points[i].LastActivity,
			endDate,
		)
	}

	return points, nil
}

// calculateActivityScore weights recent activity higher
func (s *HeatMapService) calculateActivityScore(tickCount int, lastActivity, endDate time.Time) int {
	daysSince := endDate.Sub(lastActivity).Hours() / 24

	// Recency multiplier: 2x for last week, 1.5x for last month, 1x otherwise
	var multiplier float64 = 1.0
	if daysSince <= 7 {
		multiplier = 2.0
	} else if daysSince <= 30 {
		multiplier = 1.5
	}

	score := int(float64(tickCount) * multiplier)
	if score < 1 {
		score = 1
	}

	return score
}

// GetAreaActivityDetail retrieves detailed activity information for a specific area
func (s *HeatMapService) GetAreaActivityDetail(
	ctx context.Context,
	areaID int64,
	startDate, endDate time.Time,
) (*models.AreaActivityDetail, error) {
	if areaID <= 0 {
		return nil, fmt.Errorf("invalid area ID: %d", areaID)
	}

	if startDate.After(endDate) {
		return nil, fmt.Errorf("start_date must be before end_date")
	}

	detail, err := s.repo.GetAreaActivityDetail(ctx, areaID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch area detail: %w", err)
	}

	return detail, nil
}

// GetRoutesByBounds retrieves routes within geographic bounds with activity
func (s *HeatMapService) GetRoutesByBounds(
	ctx context.Context,
	bounds database.GeoBounds,
	startDate, endDate time.Time,
	limit int,
) ([]models.RouteActivity, error) {
	if err := bounds.Validate(); err != nil {
		return nil, fmt.Errorf("invalid bounds: %w", err)
	}

	if startDate.After(endDate) {
		return nil, fmt.Errorf("start_date must be before end_date")
	}

	if limit < 1 || limit > 500 {
		limit = 100 // Default to 100 for routes
	}

	routes, err := s.repo.GetRoutesByBounds(ctx, bounds, startDate, endDate, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch routes: %w", err)
	}

	return routes, nil
}

package service

import (
	"context"
	"fmt"
	"time"

	"github.com/alexscott64/woulder/backend/internal/database"
	"github.com/alexscott64/woulder/backend/internal/database/heatmap"
	"github.com/alexscott64/woulder/backend/internal/models"
)

type HeatMapService struct {
	db *database.Database
}

func NewHeatMapService(db *database.Database) *HeatMapService {
	return &HeatMapService{db: db}
}

// GetHeatMapData retrieves aggregated activity data for the map
// Supports route type filtering and lightweight mode for performance
func (s *HeatMapService) GetHeatMapData(
	ctx context.Context,
	startDate, endDate time.Time,
	bounds *database.GeoBounds,
	minActivity, limit int,
	routeTypes []string,
	lightweight bool,
) ([]models.HeatMapPoint, error) {
	// Validate inputs
	if startDate.After(endDate) {
		return nil, fmt.Errorf("start_date must be before end_date")
	}

	if minActivity < 1 {
		minActivity = 1
	}

	if limit < 1 || limit > 10000 {
		limit = 10000 // Increased default to support showing all points
	}

	if bounds != nil {
		if err := bounds.Validate(); err != nil {
			return nil, fmt.Errorf("invalid bounds: %w", err)
		}
	}

	// Fetch raw data with route type filtering and lightweight option
	var heatmapBounds *heatmap.GeoBounds
	if bounds != nil {
		heatmapBounds = &heatmap.GeoBounds{
			MinLat: bounds.MinLat,
			MaxLat: bounds.MaxLat,
			MinLon: bounds.MinLon,
			MaxLon: bounds.MaxLon,
		}
	}
	points, err := s.db.HeatMap().GetHeatMapData(ctx, startDate, endDate, heatmapBounds, minActivity, limit, routeTypes, lightweight)
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

	detail, err := s.db.HeatMap().GetAreaActivityDetail(ctx, areaID, startDate, endDate)
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

	// Convert database.GeoBounds to heatmap.GeoBounds
	heatmapBounds := heatmap.GeoBounds{
		MinLat: bounds.MinLat,
		MaxLat: bounds.MaxLat,
		MinLon: bounds.MinLon,
		MaxLon: bounds.MaxLon,
	}

	routes, err := s.db.HeatMap().GetRoutesByBounds(ctx, heatmapBounds, startDate, endDate, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch routes: %w", err)
	}

	return routes, nil
}

// GetRouteTicksInDateRange retrieves all ticks for a specific route within a date range
func (s *HeatMapService) GetRouteTicksInDateRange(
	ctx context.Context,
	routeID int64,
	startDate, endDate time.Time,
	limit int,
) ([]models.TickDetail, error) {
	if routeID <= 0 {
		return nil, fmt.Errorf("invalid route ID: %d", routeID)
	}

	if startDate.After(endDate) {
		return nil, fmt.Errorf("start_date must be before end_date")
	}

	if limit < 1 || limit > 500 {
		limit = 100 // Default to 100 for route ticks
	}

	ticks, err := s.db.HeatMap().GetRouteTicksInDateRange(ctx, routeID, startDate, endDate, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch route ticks: %w", err)
	}

	return ticks, nil
}

// SearchRoutesInAreas searches for routes within specified areas by name
func (s *HeatMapService) SearchRoutesInAreas(
	ctx context.Context,
	areaIDs []int64,
	searchQuery string,
	startDate, endDate time.Time,
	limit int,
) ([]models.RouteActivity, error) {
	if len(areaIDs) == 0 {
		return []models.RouteActivity{}, nil
	}

	if searchQuery == "" {
		return []models.RouteActivity{}, nil
	}

	if startDate.After(endDate) {
		return nil, fmt.Errorf("start_date must be before end_date")
	}

	if limit < 1 || limit > 500 {
		limit = 100 // Default to 100 for route search
	}

	routes, err := s.db.HeatMap().SearchRoutesInAreas(ctx, areaIDs, searchQuery, startDate, endDate, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search routes: %w", err)
	}

	return routes, nil
}

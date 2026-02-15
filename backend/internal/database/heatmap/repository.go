// Package heatmap provides repository operations for heat map and activity visualization.
// Heat map data aggregates climbing activity across geographic regions for
// performance-critical clustering and visualization in the web UI.
package heatmap

import (
	"context"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// GeoBounds represents a geographic bounding box for filtering.
type GeoBounds struct {
	MinLat float64
	MaxLat float64
	MinLon float64
	MaxLon float64
}

// Repository defines operations for heat map and activity data.
// All methods are safe for concurrent use.
type Repository interface {
	// GetHeatMapData returns aggregated climbing activity for geographic areas.
	// Supports lightweight mode for clustering performance with minimal data.
	// Route type filtering allows boulder/sport/trad specific visualizations.
	// Results are ordered by activity (tick count) descending.
	GetHeatMapData(
		ctx context.Context,
		startDate, endDate time.Time,
		bounds *GeoBounds,
		minActivity, limit int,
		routeTypes []string,
		lightweight bool,
	) ([]models.HeatMapPoint, error)

	// GetAreaActivityDetail returns comprehensive activity data for a specific area.
	// Includes recent ticks, comments, activity timeline, and top routes.
	// Used for detailed area drill-down views in the UI.
	GetAreaActivityDetail(
		ctx context.Context,
		areaID int64,
		startDate, endDate time.Time,
	) (*models.AreaActivityDetail, error)

	// GetRoutesByBounds returns routes within geographic bounds with activity.
	// Used for precision route-level clustering at high zoom levels.
	// Results are ordered by tick count descending.
	GetRoutesByBounds(
		ctx context.Context,
		bounds GeoBounds,
		startDate, endDate time.Time,
		limit int,
	) ([]models.RouteActivity, error)

	// GetRouteTicksInDateRange returns all ticks for a specific route within a date range.
	// Used for route detail views showing recent climbing activity.
	// Results are ordered by climbed_at descending (most recent first).
	GetRouteTicksInDateRange(
		ctx context.Context,
		routeID int64,
		startDate, endDate time.Time,
		limit int,
	) ([]models.TickDetail, error)

	// SearchRoutesInAreas searches for routes within specified areas by name.
	// Case-insensitive partial match search.
	// Only returns routes with activity in the date range (tick_count > 0).
	// Results ordered by tick count descending, then name ascending.
	SearchRoutesInAreas(
		ctx context.Context,
		areaIDs []int64,
		searchQuery string,
		startDate, endDate time.Time,
		limit int,
	) ([]models.RouteActivity, error)
}

package heatmap

import (
	"context"
	"fmt"
	"time"

	"github.com/alexscott64/woulder/backend/internal/database/dberrors"
	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/lib/pq"
)

// PostgresRepository implements Repository using PostgreSQL.
type PostgresRepository struct {
	db DBConn
}

// NewPostgresRepository creates a new PostgreSQL heat map repository.
func NewPostgresRepository(db DBConn) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Validate ensures bounds are logical.
func (b *GeoBounds) Validate() error {
	if b.MinLat > b.MaxLat {
		return fmt.Errorf("min_lat cannot be greater than max_lat")
	}
	if b.MinLon > b.MaxLon {
		return fmt.Errorf("min_lon cannot be greater than max_lon")
	}
	if b.MinLat < -90 || b.MaxLat > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}
	if b.MinLon < -180 || b.MaxLon > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}
	return nil
}

// GetHeatMapData returns aggregated climbing activity for geographic areas.
func (r *PostgresRepository) GetHeatMapData(
	ctx context.Context,
	startDate, endDate time.Time,
	bounds *GeoBounds,
	minActivity, limit int,
	routeTypes []string,
	lightweight bool,
) ([]models.HeatMapPoint, error) {
	// Validate bounds if provided
	if bounds != nil {
		if err := bounds.Validate(); err != nil {
			return nil, fmt.Errorf("invalid bounds: %w", err)
		}
	}

	// Select query based on lightweight mode
	query := queryHeatMapDataFull
	if lightweight {
		query = queryHeatMapDataLightweight
	}

	// Prepare bound parameters (nil if no bounds specified)
	var minLat, maxLat, minLon, maxLon interface{}
	if bounds != nil {
		minLat, maxLat = bounds.MinLat, bounds.MaxLat
		minLon, maxLon = bounds.MinLon, bounds.MaxLon
	}

	// Convert route types to PostgreSQL array format
	// IMPORTANT: Pass nil for empty slice, not pq.Array([]), to avoid SQL syntax errors
	var routeTypesParam interface{}
	if len(routeTypes) > 0 {
		routeTypesParam = pq.Array(routeTypes)
	}

	rows, err := r.db.QueryContext(ctx, query,
		startDate, endDate,
		minLat, maxLat, minLon, maxLon,
		routeTypesParam,
		minActivity, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query heat map data: %w", err)
	}
	defer rows.Close()

	var points []models.HeatMapPoint
	for rows.Next() {
		var p models.HeatMapPoint
		if err := rows.Scan(
			&p.MPAreaID, &p.Name, &p.Latitude, &p.Longitude,
			&p.ActiveRoutes, &p.TotalTicks, &p.LastActivity,
			&p.UniqueClimbers, &p.HasSubareas,
		); err != nil {
			return nil, fmt.Errorf("failed to scan heat map point: %w", err)
		}
		points = append(points, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating heat map rows: %w", err)
	}

	return points, nil
}

// GetAreaActivityDetail returns comprehensive activity data for a specific area.
func (r *PostgresRepository) GetAreaActivityDetail(
	ctx context.Context,
	areaID int64,
	startDate, endDate time.Time,
) (*models.AreaActivityDetail, error) {
	var detail models.AreaActivityDetail

	// Step 1: Get base area info
	err := r.db.QueryRowContext(ctx, queryAreaInfo, areaID).Scan(
		&detail.MPAreaID,
		&detail.Name,
		&detail.ParentMPAreaID,
		&detail.Latitude,
		&detail.Longitude,
	)
	if err != nil {
		return nil, fmt.Errorf("area not found: %w", dberrors.WrapNotFound(err))
	}

	// Step 2: Get activity statistics
	err = r.db.QueryRowContext(ctx, queryAreaActivityStats, areaID, startDate, endDate).Scan(
		&detail.TotalTicks,
		&detail.ActiveRoutes,
		&detail.UniqueClimbers,
		&detail.LastActivity,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch activity stats: %w", err)
	}

	// Step 3: Get recent ticks
	detail.RecentTicks, err = r.fetchRecentTicks(ctx, areaID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Step 4: Get recent comments
	detail.RecentComments, err = r.fetchRecentComments(ctx, areaID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Step 5: Get activity timeline
	detail.ActivityTimeline, err = r.fetchActivityTimeline(ctx, areaID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Step 6: Get top routes
	detail.TopRoutes, err = r.fetchTopRoutes(ctx, areaID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	return &detail, nil
}

// fetchRecentTicks retrieves recent ticks for an area (helper method).
func (r *PostgresRepository) fetchRecentTicks(ctx context.Context, areaID int64, startDate, endDate time.Time) ([]models.TickDetail, error) {
	rows, err := r.db.QueryContext(ctx, queryRecentTicks, areaID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recent ticks: %w", err)
	}
	defer rows.Close()

	var ticks []models.TickDetail
	for rows.Next() {
		var tick models.TickDetail
		if err := rows.Scan(
			&tick.MPRouteID,
			&tick.RouteName,
			&tick.Rating,
			&tick.UserName,
			&tick.ClimbedAt,
			&tick.Style,
			&tick.Comment,
		); err != nil {
			return nil, fmt.Errorf("failed to scan tick: %w", err)
		}
		ticks = append(ticks, tick)
	}

	return ticks, nil
}

// fetchRecentComments retrieves recent comments for an area (helper method).
func (r *PostgresRepository) fetchRecentComments(ctx context.Context, areaID int64, startDate, endDate time.Time) ([]models.CommentSummary, error) {
	rows, err := r.db.QueryContext(ctx, queryRecentComments, areaID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch comments: %w", err)
	}
	defer rows.Close()

	var comments []models.CommentSummary
	for rows.Next() {
		var comment models.CommentSummary
		if err := rows.Scan(
			&comment.ID,
			&comment.UserName,
			&comment.CommentText,
			&comment.CommentedAt,
			&comment.MPRouteID,
			&comment.RouteName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

// fetchActivityTimeline retrieves daily activity aggregation for an area (helper method).
func (r *PostgresRepository) fetchActivityTimeline(ctx context.Context, areaID int64, startDate, endDate time.Time) ([]models.DailyActivity, error) {
	rows, err := r.db.QueryContext(ctx, queryActivityTimeline, areaID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch timeline: %w", err)
	}
	defer rows.Close()

	var timeline []models.DailyActivity
	for rows.Next() {
		var activity models.DailyActivity
		if err := rows.Scan(
			&activity.Date,
			&activity.TickCount,
			&activity.RouteCount,
		); err != nil {
			return nil, fmt.Errorf("failed to scan timeline: %w", err)
		}
		timeline = append(timeline, activity)
	}

	return timeline, nil
}

// fetchTopRoutes retrieves the most active routes in an area (helper method).
func (r *PostgresRepository) fetchTopRoutes(ctx context.Context, areaID int64, startDate, endDate time.Time) ([]models.TopRouteSummary, error) {
	rows, err := r.db.QueryContext(ctx, queryTopRoutes, areaID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch top routes: %w", err)
	}
	defer rows.Close()

	var routes []models.TopRouteSummary
	for rows.Next() {
		var route models.TopRouteSummary
		if err := rows.Scan(
			&route.MPRouteID,
			&route.Name,
			&route.Rating,
			&route.TickCount,
			&route.LastActivity,
		); err != nil {
			return nil, fmt.Errorf("failed to scan route: %w", err)
		}
		routes = append(routes, route)
	}

	return routes, nil
}

// GetRoutesByBounds returns routes within geographic bounds with activity.
func (r *PostgresRepository) GetRoutesByBounds(
	ctx context.Context,
	bounds GeoBounds,
	startDate, endDate time.Time,
	limit int,
) ([]models.RouteActivity, error) {
	if err := bounds.Validate(); err != nil {
		return nil, fmt.Errorf("invalid bounds: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, queryRoutesByBounds,
		bounds.MinLat, bounds.MaxLat,
		bounds.MinLon, bounds.MaxLon,
		startDate, endDate,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query routes by bounds: %w", err)
	}
	defer rows.Close()

	var routes []models.RouteActivity
	for rows.Next() {
		var route models.RouteActivity
		if err := rows.Scan(
			&route.MPRouteID, &route.Name, &route.Rating,
			&route.Latitude, &route.Longitude,
			&route.TickCount, &route.LastActivity,
			&route.MPAreaID, &route.AreaName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan route: %w", err)
		}
		routes = append(routes, route)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating route rows: %w", err)
	}

	return routes, nil
}

// GetRouteTicksInDateRange returns all ticks for a specific route within a date range.
func (r *PostgresRepository) GetRouteTicksInDateRange(
	ctx context.Context,
	routeID int64,
	startDate, endDate time.Time,
	limit int,
) ([]models.TickDetail, error) {
	rows, err := r.db.QueryContext(ctx, queryRouteTicksInDateRange, routeID, startDate, endDate, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query route ticks: %w", err)
	}
	defer rows.Close()

	var ticks []models.TickDetail
	for rows.Next() {
		var tick models.TickDetail
		if err := rows.Scan(
			&tick.MPRouteID,
			&tick.RouteName,
			&tick.Rating,
			&tick.UserName,
			&tick.ClimbedAt,
			&tick.Style,
			&tick.Comment,
		); err != nil {
			return nil, fmt.Errorf("failed to scan tick: %w", err)
		}
		ticks = append(ticks, tick)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tick rows: %w", err)
	}

	return ticks, nil
}

// SearchRoutesInAreas searches for routes within specified areas by name.
func (r *PostgresRepository) SearchRoutesInAreas(
	ctx context.Context,
	areaIDs []int64,
	searchQuery string,
	startDate, endDate time.Time,
	limit int,
) ([]models.RouteActivity, error) {
	searchPattern := "%" + searchQuery + "%"

	rows, err := r.db.QueryContext(ctx, querySearchRoutesInAreas,
		pq.Array(areaIDs),
		startDate, endDate,
		searchPattern,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search routes in areas: %w", err)
	}
	defer rows.Close()

	var routes []models.RouteActivity
	for rows.Next() {
		var route models.RouteActivity
		if err := rows.Scan(
			&route.MPRouteID, &route.Name, &route.Rating,
			&route.Latitude, &route.Longitude,
			&route.TickCount, &route.LastActivity,
			&route.MPAreaID, &route.AreaName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan route: %w", err)
		}
		routes = append(routes, route)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating route rows: %w", err)
	}

	return routes, nil
}

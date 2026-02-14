package database

import (
	"context"
	"fmt"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/lib/pq"
)

// Validate ensures bounds are logical
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

// GetHeatMapData returns aggregated climbing activity for geographic areas
// Supports route type filtering and lightweight mode for performance
func (db *Database) GetHeatMapData(
	ctx context.Context,
	startDate, endDate time.Time,
	bounds *GeoBounds,
	minActivity int,
	limit int,
	routeTypes []string,
	lightweight bool,
) ([]models.HeatMapPoint, error) {
	// Build the query based on lightweight mode
	var query string

	if lightweight {
		// Lightweight query: minimal data for clustering performance
		query = `
			SELECT
				a.mp_area_id,
				a.name,
				a.latitude,
				a.longitude,
				0 as active_routes,  -- Not calculated in lightweight mode
				COUNT(t.id) as total_ticks,
				MAX(t.climbed_at) as last_activity,
				COUNT(DISTINCT CASE WHEN t.user_name IS NOT NULL THEN t.user_name END) as unique_climbers,
				false as has_subareas  -- Not calculated in lightweight mode
			FROM woulder.mp_areas a
			JOIN woulder.mp_routes r ON r.mp_area_id = a.mp_area_id
			JOIN woulder.mp_ticks t ON t.mp_route_id = r.mp_route_id
			WHERE t.climbed_at >= $1
				AND t.climbed_at <= $2
				AND a.latitude IS NOT NULL
				AND a.longitude IS NOT NULL
				AND ($3::float IS NULL OR (
					a.latitude BETWEEN $3 AND $4
					AND a.longitude BETWEEN $5 AND $6
				))
				AND ($7::text[] IS NULL OR r.route_type = ANY($7))
			GROUP BY a.mp_area_id, a.name, a.latitude, a.longitude
			HAVING COUNT(t.id) >= $8
			ORDER BY COUNT(t.id) DESC
			LIMIT $9;
		`
	} else {
		// Full query: complete data with all aggregations
		query = `
			SELECT
				a.mp_area_id,
				a.name,
				a.latitude,
				a.longitude,
				COUNT(DISTINCT t.mp_route_id) as active_routes,
				COUNT(t.id) as total_ticks,
				MAX(t.climbed_at) as last_activity,
				COUNT(DISTINCT t.user_name) as unique_climbers,
				EXISTS(
					SELECT 1 FROM woulder.mp_areas sub
					WHERE sub.parent_mp_area_id = a.mp_area_id
					LIMIT 1
				) as has_subareas
			FROM woulder.mp_areas a
			JOIN woulder.mp_routes r ON r.mp_area_id = a.mp_area_id
			JOIN woulder.mp_ticks t ON t.mp_route_id = r.mp_route_id
			WHERE t.climbed_at >= $1
				AND t.climbed_at <= $2
				AND a.latitude IS NOT NULL
				AND a.longitude IS NOT NULL
				AND ($3::float IS NULL OR (
					a.latitude BETWEEN $3 AND $4
					AND a.longitude BETWEEN $5 AND $6
				))
				AND ($7::text[] IS NULL OR r.route_type = ANY($7))
			GROUP BY a.mp_area_id, a.name, a.latitude, a.longitude
			HAVING COUNT(t.id) >= $8
			ORDER BY COUNT(t.id) DESC
			LIMIT $9;
		`
	}

	var minLat, maxLat, minLon, maxLon interface{}
	if bounds != nil {
		minLat, maxLat = bounds.MinLat, bounds.MaxLat
		minLon, maxLon = bounds.MinLon, bounds.MaxLon
	}

	// Convert route types to PostgreSQL array format using pq.Array
	// IMPORTANT: Pass nil for empty slice, not pq.Array([]), to avoid SQL syntax errors
	var routeTypesParam interface{}
	if len(routeTypes) > 0 {
		routeTypesParam = pq.Array(routeTypes)
	} else {
		routeTypesParam = nil
	}

	rows, err := db.conn.QueryContext(ctx, query,
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

// GetAreaActivityDetail returns detailed activity breakdown for a specific area
func (db *Database) GetAreaActivityDetail(
	ctx context.Context,
	areaID int64,
	startDate, endDate time.Time,
) (*models.AreaActivityDetail, error) {
	// Base area info
	areaQuery := `
		SELECT
			a.mp_area_id,
			a.name,
			a.parent_mp_area_id,
			a.latitude,
			a.longitude
		FROM woulder.mp_areas a
		WHERE a.mp_area_id = $1
	`

	var detail models.AreaActivityDetail
	err := db.conn.QueryRowContext(ctx, areaQuery, areaID).Scan(
		&detail.MPAreaID,
		&detail.Name,
		&detail.ParentMPAreaID,
		&detail.Latitude,
		&detail.Longitude,
	)
	if err != nil {
		return nil, fmt.Errorf("area not found: %w", err)
	}

	// Activity stats
	statsQuery := `
		SELECT
			COUNT(t.id) as total_ticks,
			COUNT(DISTINCT t.mp_route_id) as active_routes,
			COUNT(DISTINCT CASE WHEN t.user_name IS NOT NULL THEN t.user_name END) as unique_climbers,
			MAX(t.climbed_at) as last_activity
		FROM woulder.mp_ticks t
		JOIN woulder.mp_routes r ON t.mp_route_id = r.mp_route_id
		WHERE r.mp_area_id = $1
			AND t.climbed_at >= $2
			AND t.climbed_at <= $3
	`

	err = db.conn.QueryRowContext(ctx, statsQuery, areaID, startDate, endDate).Scan(
		&detail.TotalTicks,
		&detail.ActiveRoutes,
		&detail.UniqueClimbers,
		&detail.LastActivity,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch activity stats: %w", err)
	}

	// Recent ticks (last 20)
	ticksQuery := `
		SELECT
			t.mp_route_id,
			r.name as route_name,
			COALESCE(r.rating, '') as rating,
			COALESCE(t.user_name, '') as user_name,
			t.climbed_at,
			COALESCE(t.style, '') as style,
			COALESCE(t.comment, '') as comment
		FROM woulder.mp_ticks t
		JOIN woulder.mp_routes r ON t.mp_route_id = r.mp_route_id
		WHERE r.mp_area_id = $1
			AND t.climbed_at >= $2
			AND t.climbed_at <= $3
		ORDER BY t.climbed_at DESC
		LIMIT 20
	`

	rows, err := db.conn.QueryContext(ctx, ticksQuery, areaID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recent ticks: %w", err)
	}
	defer rows.Close()

	detail.RecentTicks = []models.TickDetail{}
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
		detail.RecentTicks = append(detail.RecentTicks, tick)
	}

	// Recent comments (last 10)
	// Join through routes to get comments for routes in this area
	// Strip HTML tags from comment text using regex
	commentsQuery := `
		SELECT
			c.id,
			c.user_name,
			REGEXP_REPLACE(
				REGEXP_REPLACE(c.comment_text, '<[^>]+>', '', 'g'),
				'\s+', ' ', 'g'
			) as comment_text,
			c.commented_at,
			c.mp_route_id,
			r.name as route_name
		FROM woulder.mp_comments c
		JOIN woulder.mp_routes r ON c.mp_route_id = r.mp_route_id
		WHERE r.mp_area_id = $1
			AND c.commented_at >= $2
			AND c.commented_at <= $3
		ORDER BY c.commented_at DESC
		LIMIT 10
	`

	rows, err = db.conn.QueryContext(ctx, commentsQuery, areaID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch comments: %w", err)
	}
	defer rows.Close()

	detail.RecentComments = []models.CommentSummary{}
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
		detail.RecentComments = append(detail.RecentComments, comment)
	}

	// Activity timeline (daily aggregation)
	timelineQuery := `
		SELECT
			DATE(t.climbed_at) as date,
			COUNT(t.id) as tick_count,
			COUNT(DISTINCT t.mp_route_id) as route_count
		FROM woulder.mp_ticks t
		JOIN woulder.mp_routes r ON t.mp_route_id = r.mp_route_id
		WHERE r.mp_area_id = $1
			AND t.climbed_at >= $2
			AND t.climbed_at <= $3
		GROUP BY DATE(t.climbed_at)
		ORDER BY DATE(t.climbed_at) ASC
	`

	rows, err = db.conn.QueryContext(ctx, timelineQuery, areaID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch timeline: %w", err)
	}
	defer rows.Close()

	detail.ActivityTimeline = []models.DailyActivity{}
	for rows.Next() {
		var activity models.DailyActivity
		if err := rows.Scan(
			&activity.Date,
			&activity.TickCount,
			&activity.RouteCount,
		); err != nil {
			return nil, fmt.Errorf("failed to scan timeline: %w", err)
		}
		detail.ActivityTimeline = append(detail.ActivityTimeline, activity)
	}

	// Top routes by activity
	topRoutesQuery := `
		SELECT
			r.mp_route_id,
			r.name,
			r.rating,
			COUNT(t.id) as tick_count,
			MAX(t.climbed_at) as last_activity
		FROM woulder.mp_routes r
		JOIN woulder.mp_ticks t ON r.mp_route_id = t.mp_route_id
		WHERE r.mp_area_id = $1
			AND t.climbed_at >= $2
			AND t.climbed_at <= $3
		GROUP BY r.mp_route_id, r.name, r.rating
		ORDER BY COUNT(t.id) DESC
		LIMIT 10
	`

	rows, err = db.conn.QueryContext(ctx, topRoutesQuery, areaID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch top routes: %w", err)
	}
	defer rows.Close()

	detail.TopRoutes = []models.TopRouteSummary{}
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
		detail.TopRoutes = append(detail.TopRoutes, route)
	}

	return &detail, nil
}

// GetRoutesByBounds returns routes within geographic bounds with activity
func (db *Database) GetRoutesByBounds(
	ctx context.Context,
	bounds GeoBounds,
	startDate, endDate time.Time,
	limit int,
) ([]models.RouteActivity, error) {
	query := `
		SELECT
			r.mp_route_id,
			r.name,
			r.rating,
			r.latitude,
			r.longitude,
			COUNT(t.id) as tick_count,
			MAX(t.climbed_at) as last_activity,
			r.mp_area_id,
			a.name as area_name
		FROM woulder.mp_routes r
		JOIN woulder.mp_ticks t ON r.mp_route_id = t.mp_route_id
		JOIN woulder.mp_areas a ON r.mp_area_id = a.mp_area_id
		WHERE r.latitude IS NOT NULL
			AND r.longitude IS NOT NULL
			AND r.latitude BETWEEN $1 AND $2
			AND r.longitude BETWEEN $3 AND $4
			AND t.climbed_at >= $5
			AND t.climbed_at <= $6
		GROUP BY r.mp_route_id, r.name, r.rating, r.latitude, r.longitude, r.mp_area_id, a.name
		ORDER BY COUNT(t.id) DESC
		LIMIT $7
	`

	rows, err := db.conn.QueryContext(ctx, query,
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
		var r models.RouteActivity
		if err := rows.Scan(
			&r.MPRouteID, &r.Name, &r.Rating,
			&r.Latitude, &r.Longitude,
			&r.TickCount, &r.LastActivity,
			&r.MPAreaID, &r.AreaName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan route: %w", err)
		}
		routes = append(routes, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating route rows: %w", err)
	}

	return routes, nil
}

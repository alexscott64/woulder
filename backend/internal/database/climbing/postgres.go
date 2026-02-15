package climbing

import (
	"context"
	"database/sql"

	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/lib/pq"
)

// PostgresRepository implements Repository using PostgreSQL.
// Implements the composite pattern with sub-repositories for organization.
type PostgresRepository struct {
	db DBConn
}

// NewPostgresRepository creates a new PostgreSQL climbing repository.
func NewPostgresRepository(db DBConn) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Implement Repository interface by returning self for sub-repositories
func (r *PostgresRepository) History() HistoryRepository   { return r }
func (r *PostgresRepository) Activity() ActivityRepository { return r }
func (r *PostgresRepository) Search() SearchRepository     { return r }

// ====================
// History Repository
// ====================

// GetLastClimbedForLocation retrieves the most recent climb for a location.
func (r *PostgresRepository) GetLastClimbedForLocation(ctx context.Context, locationID int) (*models.LastClimbedInfo, error) {
	var info models.LastClimbedInfo
	err := r.db.QueryRowContext(ctx, queryGetLastClimbedForLocation, locationID).Scan(
		&info.RouteName,
		&info.RouteRating,
		&info.ClimbedAt,
		&info.ClimbedBy,
		&info.Style,
		&info.Comment,
		&info.DaysSinceClimb,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No climb data for this location
	}

	if err != nil {
		return nil, err
	}

	return &info, nil
}

// GetClimbHistoryForLocation retrieves recent climb history for a location.
func (r *PostgresRepository) GetClimbHistoryForLocation(ctx context.Context, locationID int, limit int) ([]models.ClimbHistoryEntry, error) {
	rows, err := r.db.QueryContext(ctx, queryGetClimbHistoryForLocation, locationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []models.ClimbHistoryEntry
	for rows.Next() {
		var entry models.ClimbHistoryEntry
		var climbedBy, style sql.NullString
		var comment sql.NullString

		err := rows.Scan(
			&entry.MPRouteID,
			&entry.RouteName,
			&entry.RouteRating,
			&entry.MPAreaID,
			&entry.AreaName,
			&entry.ClimbedAt,
			&climbedBy,
			&style,
			&comment,
			&entry.DaysSinceClimb,
		)
		if err != nil {
			return nil, err
		}

		// Handle nullable fields
		entry.ClimbedBy = climbedBy.String // Will be empty string if NULL
		entry.Style = style.String         // Will be empty string if NULL
		if comment.Valid {
			entry.Comment = &comment.String
		}

		history = append(history, entry)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return history, nil
}

// GetClimbHistoryForLocations retrieves recent climb history for multiple locations in a single query.
// Returns a map of locationID -> []ClimbHistoryEntry for efficient batch fetching.
func (r *PostgresRepository) GetClimbHistoryForLocations(ctx context.Context, locationIDs []int, limit int) (map[int][]models.ClimbHistoryEntry, error) {
	if len(locationIDs) == 0 {
		return make(map[int][]models.ClimbHistoryEntry), nil
	}

	// Convert to int64 array for PostgreSQL compatibility
	int64IDs := make([]int64, len(locationIDs))
	for i, id := range locationIDs {
		int64IDs[i] = int64(id)
	}

	// Use pq.Array to properly pass array parameter to PostgreSQL
	rows, err := r.db.QueryContext(ctx, queryGetClimbHistoryForLocations, pq.Array(int64IDs), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Pre-allocate the result map
	result := make(map[int][]models.ClimbHistoryEntry, len(locationIDs))

	for rows.Next() {
		var locationID int
		var entry models.ClimbHistoryEntry
		var climbedBy, style sql.NullString
		var comment sql.NullString

		err := rows.Scan(
			&locationID,
			&entry.MPRouteID,
			&entry.RouteName,
			&entry.RouteRating,
			&entry.MPAreaID,
			&entry.AreaName,
			&entry.ClimbedAt,
			&climbedBy,
			&style,
			&comment,
			&entry.DaysSinceClimb,
		)
		if err != nil {
			return nil, err
		}

		// Handle nullable fields
		entry.ClimbedBy = climbedBy.String // Will be empty string if NULL
		entry.Style = style.String         // Will be empty string if NULL
		if comment.Valid {
			entry.Comment = &comment.String
		}

		// Append to the slice for this location
		result[locationID] = append(result[locationID], entry)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// ====================
// Activity Repository
// ====================

// GetAreasOrderedByActivity retrieves top-level areas ordered by activity.
func (r *PostgresRepository) GetAreasOrderedByActivity(ctx context.Context, locationID int) ([]models.AreaActivitySummary, error) {
	rows, err := r.db.QueryContext(ctx, queryGetAreasOrderedByActivity, locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var areas []models.AreaActivitySummary
	for rows.Next() {
		var area models.AreaActivitySummary
		err := rows.Scan(
			&area.MPAreaID,
			&area.Name,
			&area.ParentMPAreaID,
			&area.LastClimbAt,
			&area.UniqueRoutes,
			&area.TotalTicks,
			&area.DaysSinceClimb,
			&area.HasSubareas,
			&area.SubareaCount,
		)
		if err != nil {
			return nil, err
		}
		areas = append(areas, area)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return areas, nil
}

// GetSubareasOrderedByActivity retrieves subareas ordered by activity.
func (r *PostgresRepository) GetSubareasOrderedByActivity(ctx context.Context, parentAreaID int64, locationID int) ([]models.AreaActivitySummary, error) {
	rows, err := r.db.QueryContext(ctx, queryGetSubareasOrderedByActivity, parentAreaID, locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var areas []models.AreaActivitySummary
	for rows.Next() {
		var area models.AreaActivitySummary
		err := rows.Scan(
			&area.MPAreaID,
			&area.Name,
			&area.ParentMPAreaID,
			&area.LastClimbAt,
			&area.UniqueRoutes,
			&area.TotalTicks,
			&area.DaysSinceClimb,
			&area.HasSubareas,
			&area.SubareaCount,
		)
		if err != nil {
			return nil, err
		}
		areas = append(areas, area)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return areas, nil
}

// GetRoutesOrderedByActivity retrieves routes in an area ordered by activity.
func (r *PostgresRepository) GetRoutesOrderedByActivity(ctx context.Context, areaID int64, locationID int, limit int) ([]models.RouteActivitySummary, error) {
	rows, err := r.db.QueryContext(ctx, queryGetRoutesOrderedByActivity, areaID, locationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routes []models.RouteActivitySummary
	for rows.Next() {
		var route models.RouteActivitySummary
		var climbedBy, style, comment, areaName sql.NullString
		var climbedAt sql.NullTime
		var noTicks int

		err := rows.Scan(
			&route.MPRouteID,
			&route.Name,
			&route.Rating,
			&route.MPAreaID,
			&route.LastClimbAt,
			&route.DaysSinceClimb,
			&climbedBy,
			&climbedAt,
			&style,
			&comment,
			&areaName,
			&noTicks,
		)
		if err != nil {
			return nil, err
		}

		// Only populate most recent tick if the route has been climbed
		if climbedAt.Valid {
			mostRecentTick := &models.ClimbHistoryEntry{
				MPRouteID:      route.MPRouteID,
				RouteName:      route.Name,
				RouteRating:    route.Rating,
				MPAreaID:       route.MPAreaID,
				ClimbedBy:      climbedBy.String,
				ClimbedAt:      climbedAt.Time,
				Style:          style.String,
				AreaName:       areaName.String,
				DaysSinceClimb: route.DaysSinceClimb,
			}
			if comment.Valid {
				mostRecentTick.Comment = &comment.String
			}
			route.MostRecentTick = mostRecentTick
		}

		routes = append(routes, route)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return routes, nil
}

// GetRecentTicksForRoute retrieves the most recent ticks for a specific route.
func (r *PostgresRepository) GetRecentTicksForRoute(ctx context.Context, routeID int64, limit int) ([]models.ClimbHistoryEntry, error) {
	rows, err := r.db.QueryContext(ctx, queryGetRecentTicksForRoute, routeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ticks []models.ClimbHistoryEntry
	for rows.Next() {
		var tick models.ClimbHistoryEntry
		var climbedBy, style sql.NullString
		var comment sql.NullString

		err := rows.Scan(
			&tick.MPRouteID,
			&tick.RouteName,
			&tick.RouteRating,
			&tick.MPAreaID,
			&tick.AreaName,
			&tick.ClimbedAt,
			&climbedBy,
			&style,
			&comment,
			&tick.DaysSinceClimb,
		)
		if err != nil {
			return nil, err
		}

		// Handle nullable fields
		tick.ClimbedBy = climbedBy.String // Will be empty string if NULL
		tick.Style = style.String         // Will be empty string if NULL
		if comment.Valid {
			tick.Comment = &comment.String
		}

		ticks = append(ticks, tick)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ticks, nil
}

// ====================
// Search Repository
// ====================

// SearchInLocation searches all areas and routes in a location by name.
func (r *PostgresRepository) SearchInLocation(ctx context.Context, locationID int, searchQuery string, limit int) ([]models.SearchResult, error) {
	searchPattern := "%" + searchQuery + "%"
	rows, err := r.db.QueryContext(ctx, querySearchInLocation, locationID, searchPattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.SearchResult
	for rows.Next() {
		var result models.SearchResult
		var rating, areaName, climbedBy, style, comment sql.NullString
		var totalTicks, uniqueRoutes sql.NullInt64
		var climbedAt sql.NullTime
		var noTicks int

		err := rows.Scan(
			&result.ResultType,
			&result.ID,
			&result.Name,
			&rating,
			&result.MPAreaID,
			&areaName,
			&result.LastClimbAt,
			&result.DaysSinceClimb,
			&totalTicks,
			&uniqueRoutes,
			&climbedBy,
			&climbedAt,
			&style,
			&comment,
			&noTicks,
		)
		if err != nil {
			return nil, err
		}

		// Populate optional fields based on result type
		if rating.Valid {
			result.Rating = &rating.String
		}
		if areaName.Valid {
			result.AreaName = &areaName.String
		}
		if totalTicks.Valid {
			ticks := int(totalTicks.Int64)
			result.TotalTicks = &ticks
		}
		if uniqueRoutes.Valid {
			routes := int(uniqueRoutes.Int64)
			result.UniqueRoutes = &routes
		}

		// Only populate most recent tick for routes with activity
		if result.ResultType == "route" && climbedAt.Valid {
			mostRecentTick := &models.ClimbHistoryEntry{
				MPRouteID:      result.ID,
				RouteName:      result.Name,
				RouteRating:    *result.Rating,
				MPAreaID:       result.MPAreaID,
				ClimbedBy:      climbedBy.String,
				ClimbedAt:      climbedAt.Time,
				Style:          style.String,
				AreaName:       *result.AreaName,
				DaysSinceClimb: result.DaysSinceClimb,
			}
			if comment.Valid {
				mostRecentTick.Comment = &comment.String
			}
			result.MostRecentTick = mostRecentTick
		}

		results = append(results, result)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// SearchRoutesInLocation searches routes in a location by name, rating, or area name.
func (r *PostgresRepository) SearchRoutesInLocation(ctx context.Context, locationID int, searchQuery string, limit int) ([]models.RouteActivitySummary, error) {
	searchPattern := "%" + searchQuery + "%"
	rows, err := r.db.QueryContext(ctx, querySearchRoutesInLocation, locationID, searchPattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routes []models.RouteActivitySummary
	for rows.Next() {
		var route models.RouteActivitySummary
		var climbedBy, style, comment, areaName sql.NullString
		var climbedAt sql.NullTime
		var noTicks int

		err := rows.Scan(
			&route.MPRouteID,
			&route.Name,
			&route.Rating,
			&route.MPAreaID,
			&route.LastClimbAt,
			&route.DaysSinceClimb,
			&climbedBy,
			&climbedAt,
			&style,
			&comment,
			&areaName,
			&noTicks,
		)
		if err != nil {
			return nil, err
		}

		// Only populate most recent tick if the route has been climbed
		if climbedAt.Valid {
			mostRecentTick := &models.ClimbHistoryEntry{
				MPRouteID:      route.MPRouteID,
				RouteName:      route.Name,
				RouteRating:    route.Rating,
				MPAreaID:       route.MPAreaID,
				ClimbedBy:      climbedBy.String,
				ClimbedAt:      climbedAt.Time,
				Style:          style.String,
				AreaName:       areaName.String,
				DaysSinceClimb: route.DaysSinceClimb,
			}
			if comment.Valid {
				mostRecentTick.Comment = &comment.String
			}
			route.MostRecentTick = mostRecentTick
		}

		routes = append(routes, route)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return routes, nil
}

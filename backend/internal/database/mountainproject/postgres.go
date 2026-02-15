package mountainproject

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// PostgresRepository implements all Mountain Project sub-repositories.
type PostgresRepository struct {
	db DBConn
}

// NewPostgresRepository creates a new PostgreSQL-backed Mountain Project repository.
func NewPostgresRepository(db DBConn) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Implement Repository interface - return self as all sub-repositories
func (r *PostgresRepository) Areas() AreasRepository       { return r }
func (r *PostgresRepository) Routes() RoutesRepository     { return r }
func (r *PostgresRepository) Ticks() TicksRepository       { return r }
func (r *PostgresRepository) Comments() CommentsRepository { return r }
func (r *PostgresRepository) Sync() SyncRepository         { return r }

// AreasRepository implementation

func (r *PostgresRepository) SaveArea(ctx context.Context, area *models.MPArea) error {
	_, err := r.db.ExecContext(ctx, querySaveArea,
		area.MPAreaID,
		area.Name,
		area.ParentMPAreaID,
		area.AreaType,
		area.LocationID,
		area.Latitude,
		area.Longitude,
		time.Now(),
	)
	return err
}

func (r *PostgresRepository) GetByID(ctx context.Context, mpAreaID int64) (*models.MPArea, error) {
	var area models.MPArea
	err := r.db.QueryRowContext(ctx, queryGetAreaByID, mpAreaID).Scan(
		&area.ID,
		&area.MPAreaID,
		&area.Name,
		&area.ParentMPAreaID,
		&area.AreaType,
		&area.LocationID,
		&area.Latitude,
		&area.Longitude,
		&area.LastSyncedAt,
		&area.CreatedAt,
		&area.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &area, nil
}

func (r *PostgresRepository) UpdateRouteCount(ctx context.Context, mpAreaID string, total int) error {
	_, err := r.db.ExecContext(ctx, queryUpdateAreaRouteCount, total, mpAreaID)
	return err
}

func (r *PostgresRepository) GetRouteCount(ctx context.Context, mpAreaID string) (int, error) {
	var count sql.NullInt64
	err := r.db.QueryRowContext(ctx, queryGetAreaRouteCount, mpAreaID).Scan(&count)
	if err == sql.ErrNoRows {
		return -1, nil
	}
	if err != nil {
		return -1, err
	}
	if !count.Valid {
		return -1, nil // Count not yet set
	}
	return int(count.Int64), nil
}

func (r *PostgresRepository) GetChildAreas(ctx context.Context, parentMPAreaID string) ([]ChildArea, error) {
	rows, err := r.db.QueryContext(ctx, queryGetChildAreas, parentMPAreaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var areas []ChildArea
	for rows.Next() {
		var area ChildArea
		if err := rows.Scan(&area.MPAreaID, &area.Name); err != nil {
			return nil, err
		}
		areas = append(areas, area)
	}
	return areas, rows.Err()
}

func (r *PostgresRepository) GetAllStateConfigs(ctx context.Context) ([]StateConfig, error) {
	rows, err := r.db.QueryContext(ctx, queryGetAllStateConfigs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []StateConfig
	for rows.Next() {
		var config StateConfig
		if err := rows.Scan(&config.StateName, &config.MPAreaID, &config.IsActive); err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}
	return configs, rows.Err()
}

// RoutesRepository implementation

func (r *PostgresRepository) SaveRoute(ctx context.Context, route *models.MPRoute) error {
	_, err := r.db.ExecContext(ctx, querySaveRoute,
		route.MPRouteID,
		route.MPAreaID,
		route.Name,
		route.RouteType,
		route.Rating,
		route.LocationID,
		route.Latitude,
		route.Longitude,
		route.Aspect,
	)
	return err
}

func (r *PostgresRepository) GetAllIDsForLocation(ctx context.Context, locationID int) ([]int64, error) {
	rows, err := r.db.QueryContext(ctx, queryGetAllRouteIDsForLocation, locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routeIDs []int64
	for rows.Next() {
		var routeID int64
		if err := rows.Scan(&routeID); err != nil {
			return nil, err
		}
		routeIDs = append(routeIDs, routeID)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return routeIDs, nil
}

func (r *PostgresRepository) UpdateGPS(ctx context.Context, routeID int64, latitude, longitude float64, aspect string) error {
	_, err := r.db.ExecContext(ctx, queryUpdateRouteGPS, latitude, longitude, aspect, routeID)
	return err
}

func (r *PostgresRepository) GetIDsForArea(ctx context.Context, mpAreaID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, queryGetRouteIDsForArea, mpAreaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routeIDs []string
	for rows.Next() {
		var routeID string
		if err := rows.Scan(&routeID); err != nil {
			return nil, err
		}
		routeIDs = append(routeIDs, routeID)
	}
	return routeIDs, rows.Err()
}

func (r *PostgresRepository) UpsertRoute(ctx context.Context, mpRouteID, mpAreaID int64, locationID *int, name, routeType, rating string, lat, lon *float64, aspect *string) error {
	_, err := r.db.ExecContext(ctx, queryUpsertRoute, mpRouteID, mpAreaID, locationID, name, routeType, rating, lat, lon, aspect)
	return err
}

// TicksRepository implementation

func (r *PostgresRepository) SaveTick(ctx context.Context, tick *models.MPTick) error {
	// Insert the tick
	_, err := r.db.ExecContext(ctx, querySaveTick,
		tick.MPRouteID,
		tick.UserName,
		tick.ClimbedAt,
		tick.Style,
		tick.Comment,
	)
	if err != nil {
		return err
	}

	// Update last_tick_sync_at timestamp for this route
	_, err = r.db.ExecContext(ctx, queryUpdateRouteTickSyncTimestamp, tick.MPRouteID)
	return err
}

func (r *PostgresRepository) GetLastTimestampForRoute(ctx context.Context, routeID int64) (*time.Time, error) {
	var lastTick *time.Time
	err := r.db.QueryRowContext(ctx, queryGetLastTickTimestamp, routeID).Scan(&lastTick)

	if err == sql.ErrNoRows {
		return nil, nil // No ticks for this route yet
	}

	if err != nil {
		return nil, err
	}

	return lastTick, nil
}

func (r *PostgresRepository) UpsertTick(ctx context.Context, mpRouteID int64, userName string, climbedAt time.Time, style string, comment *string) error {
	_, err := r.db.ExecContext(ctx, querySaveTick, mpRouteID, userName, climbedAt, style, comment)
	return err
}

// CommentsRepository implementation

func (r *PostgresRepository) SaveAreaComment(ctx context.Context, mpCommentID, mpAreaID int64, userName, commentText string, commentedAt time.Time) error {
	_, err := r.db.ExecContext(ctx, querySaveAreaComment,
		mpCommentID,
		mpAreaID,
		userName,
		commentText,
		commentedAt,
	)
	return err
}

func (r *PostgresRepository) SaveRouteComment(ctx context.Context, mpCommentID, mpRouteID int64, userName, commentText string, commentedAt time.Time) error {
	// Insert/update the comment
	_, err := r.db.ExecContext(ctx, querySaveRouteComment,
		mpCommentID,
		mpRouteID,
		userName,
		commentText,
		commentedAt,
	)
	if err != nil {
		return err
	}

	// Update last_comment_sync_at timestamp for this route
	_, err = r.db.ExecContext(ctx, queryUpdateRouteCommentSyncTimestamp, mpRouteID)
	return err
}

func (r *PostgresRepository) UpsertAreaComment(ctx context.Context, mpCommentID, mpAreaID int64, userName string, userID *string, commentText string, commentedAt time.Time) error {
	_, err := r.db.ExecContext(ctx, queryUpsertAreaComment,
		mpCommentID,
		mpAreaID,
		userName,
		userID,
		commentText,
		commentedAt,
	)
	return err
}

func (r *PostgresRepository) UpsertRouteComment(ctx context.Context, mpCommentID, mpRouteID int64, userName string, userID *string, commentText string, commentedAt time.Time) error {
	_, err := r.db.ExecContext(ctx, queryUpsertRouteComment,
		mpCommentID,
		mpRouteID,
		userName,
		userID,
		commentText,
		commentedAt,
	)
	return err
}

// SyncRepository implementation

func (r *PostgresRepository) UpdateRoutePriorities(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, queryUpdateRouteSyncPriorities)
	return err
}

func (r *PostgresRepository) GetLocationRoutesDueForSync(ctx context.Context, syncType string) ([]int64, error) {
	var query string

	if syncType == "ticks" {
		query = queryGetLocationRoutesDueForTickSync
	} else if syncType == "comments" {
		query = queryGetLocationRoutesDueForCommentSync
	} else {
		return nil, fmt.Errorf("invalid syncType: %s (must be 'ticks' or 'comments')", syncType)
	}

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routeIDs []int64
	for rows.Next() {
		var routeID int64
		if err := rows.Scan(&routeID); err != nil {
			return nil, err
		}
		routeIDs = append(routeIDs, routeID)
	}

	return routeIDs, rows.Err()
}

func (r *PostgresRepository) GetRoutesDueForTickSync(ctx context.Context, priority string) ([]int64, error) {
	var interval string
	switch priority {
	case "high":
		interval = "24 hours"
	case "medium":
		interval = "7 days"
	case "low":
		interval = "30 days"
	default:
		return nil, fmt.Errorf("invalid priority: %s (must be 'high', 'medium', or 'low')", priority)
	}

	query := fmt.Sprintf(queryGetRoutesDueForTickSyncTemplate, interval)

	rows, err := r.db.QueryContext(ctx, query, priority)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routeIDs []int64
	for rows.Next() {
		var routeID int64
		if err := rows.Scan(&routeID); err != nil {
			return nil, err
		}
		routeIDs = append(routeIDs, routeID)
	}

	return routeIDs, rows.Err()
}

func (r *PostgresRepository) GetRoutesDueForCommentSync(ctx context.Context, priority string) ([]int64, error) {
	var interval string
	switch priority {
	case "high":
		interval = "24 hours"
	case "medium":
		interval = "7 days"
	case "low":
		interval = "30 days"
	default:
		return nil, fmt.Errorf("invalid priority: %s (must be 'high', 'medium', or 'low')", priority)
	}

	query := fmt.Sprintf(queryGetRoutesDueForCommentSyncTemplate, interval)

	rows, err := r.db.QueryContext(ctx, query, priority)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routeIDs []int64
	for rows.Next() {
		var routeID int64
		if err := rows.Scan(&routeID); err != nil {
			return nil, err
		}
		routeIDs = append(routeIDs, routeID)
	}

	return routeIDs, rows.Err()
}

func (r *PostgresRepository) GetPriorityDistribution(ctx context.Context) (map[string]int, error) {
	rows, err := r.db.QueryContext(ctx, queryGetPriorityDistribution)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	distribution := make(map[string]int)
	for rows.Next() {
		var priority string
		var count int
		if err := rows.Scan(&priority, &count); err != nil {
			return nil, err
		}
		distribution[priority] = count
	}

	return distribution, rows.Err()
}

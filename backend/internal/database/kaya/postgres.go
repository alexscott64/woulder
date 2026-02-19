package kaya

import (
	"context"
	"database/sql"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// PostgresRepository implements all Kaya sub-repositories.
type PostgresRepository struct {
	db DBConn
}

// NewPostgresRepository creates a new PostgreSQL-backed Kaya repository.
func NewPostgresRepository(db DBConn) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Implement Repository interface - return self as all sub-repositories
func (r *PostgresRepository) Users() UsersRepository         { return r }
func (r *PostgresRepository) Locations() LocationsRepository { return r }
func (r *PostgresRepository) Climbs() ClimbsRepository       { return r }
func (r *PostgresRepository) Ascents() AscentsRepository     { return r }
func (r *PostgresRepository) Posts() PostsRepository         { return r }
func (r *PostgresRepository) Sync() SyncRepository           { return r }

// UsersRepository implementation

func (r *PostgresRepository) SaveUser(ctx context.Context, user *models.KayaUser) error {
	_, err := r.db.ExecContext(ctx, querySaveUser,
		user.KayaUserID,
		user.Username,
		user.Fname,
		user.Lname,
		user.PhotoURL,
		user.Bio,
		user.Height,
		user.ApeIndex,
		user.LimitGradeBoulderingID,
		user.LimitGradeBoulderingName,
		user.LimitGradeRoutesID,
		user.LimitGradeRoutesName,
		user.IsPrivate,
		user.IsPremium,
	)
	return err
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, kayaUserID string) (*models.KayaUser, error) {
	var user models.KayaUser
	err := r.db.QueryRowContext(ctx, queryGetUserByID, kayaUserID).Scan(
		&user.ID,
		&user.KayaUserID,
		&user.Username,
		&user.Fname,
		&user.Lname,
		&user.PhotoURL,
		&user.Bio,
		&user.Height,
		&user.ApeIndex,
		&user.LimitGradeBoulderingID,
		&user.LimitGradeBoulderingName,
		&user.LimitGradeRoutesID,
		&user.LimitGradeRoutesName,
		&user.IsPrivate,
		&user.IsPremium,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *PostgresRepository) GetUserByUsername(ctx context.Context, username string) (*models.KayaUser, error) {
	var user models.KayaUser
	err := r.db.QueryRowContext(ctx, queryGetUserByUsername, username).Scan(
		&user.ID,
		&user.KayaUserID,
		&user.Username,
		&user.Fname,
		&user.Lname,
		&user.PhotoURL,
		&user.Bio,
		&user.Height,
		&user.ApeIndex,
		&user.LimitGradeBoulderingID,
		&user.LimitGradeBoulderingName,
		&user.LimitGradeRoutesID,
		&user.LimitGradeRoutesName,
		&user.IsPrivate,
		&user.IsPremium,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// LocationsRepository implementation

func (r *PostgresRepository) SaveLocation(ctx context.Context, location *models.KayaLocation) error {
	_, err := r.db.ExecContext(ctx, querySaveLocation,
		location.KayaLocationID,
		location.Slug,
		location.Name,
		location.Latitude,
		location.Longitude,
		location.PhotoURL,
		location.Description,
		location.LocationTypeID,
		location.LocationTypeName,
		location.ParentLocationID,
		location.ParentLocationSlug,
		location.ParentLocationName,
		location.ClimbCount,
		location.BoulderCount,
		location.RouteCount,
		location.AscentCount,
		location.IsGBModeratedBouldering,
		location.IsGBModeratedRoutes,
		location.IsAccessSensitive,
		location.IsClosed,
		location.HasMapsDisabled,
		location.ClosedDate,
		location.DescriptionBouldering,
		location.DescriptionRoutes,
		location.DescriptionShortBouldering,
		location.DescriptionShortRoutes,
		location.AccessDescriptionBouldering,
		location.AccessDescriptionRoutes,
		location.AccessIssuesDescriptionBouldering,
		location.AccessIssuesDescriptionRoutes,
		location.ClimbTypeID,
		location.WoulderLocationID,
		time.Now(),
	)
	return err
}

func (r *PostgresRepository) GetLocationByID(ctx context.Context, kayaLocationID string) (*models.KayaLocation, error) {
	var loc models.KayaLocation
	err := r.db.QueryRowContext(ctx, queryGetLocationByID, kayaLocationID).Scan(
		&loc.ID,
		&loc.KayaLocationID,
		&loc.Slug,
		&loc.Name,
		&loc.Latitude,
		&loc.Longitude,
		&loc.PhotoURL,
		&loc.Description,
		&loc.LocationTypeID,
		&loc.LocationTypeName,
		&loc.ParentLocationID,
		&loc.ParentLocationSlug,
		&loc.ParentLocationName,
		&loc.ClimbCount,
		&loc.BoulderCount,
		&loc.RouteCount,
		&loc.AscentCount,
		&loc.IsGBModeratedBouldering,
		&loc.IsGBModeratedRoutes,
		&loc.IsAccessSensitive,
		&loc.IsClosed,
		&loc.HasMapsDisabled,
		&loc.ClosedDate,
		&loc.DescriptionBouldering,
		&loc.DescriptionRoutes,
		&loc.DescriptionShortBouldering,
		&loc.DescriptionShortRoutes,
		&loc.AccessDescriptionBouldering,
		&loc.AccessDescriptionRoutes,
		&loc.AccessIssuesDescriptionBouldering,
		&loc.AccessIssuesDescriptionRoutes,
		&loc.ClimbTypeID,
		&loc.WoulderLocationID,
		&loc.LastSyncedAt,
		&loc.CreatedAt,
		&loc.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &loc, nil
}

func (r *PostgresRepository) GetLocationBySlug(ctx context.Context, slug string) (*models.KayaLocation, error) {
	var loc models.KayaLocation
	err := r.db.QueryRowContext(ctx, queryGetLocationBySlug, slug).Scan(
		&loc.ID,
		&loc.KayaLocationID,
		&loc.Slug,
		&loc.Name,
		&loc.Latitude,
		&loc.Longitude,
		&loc.PhotoURL,
		&loc.Description,
		&loc.LocationTypeID,
		&loc.LocationTypeName,
		&loc.ParentLocationID,
		&loc.ParentLocationSlug,
		&loc.ParentLocationName,
		&loc.ClimbCount,
		&loc.BoulderCount,
		&loc.RouteCount,
		&loc.AscentCount,
		&loc.IsGBModeratedBouldering,
		&loc.IsGBModeratedRoutes,
		&loc.IsAccessSensitive,
		&loc.IsClosed,
		&loc.HasMapsDisabled,
		&loc.ClosedDate,
		&loc.DescriptionBouldering,
		&loc.DescriptionRoutes,
		&loc.DescriptionShortBouldering,
		&loc.DescriptionShortRoutes,
		&loc.AccessDescriptionBouldering,
		&loc.AccessDescriptionRoutes,
		&loc.AccessIssuesDescriptionBouldering,
		&loc.AccessIssuesDescriptionRoutes,
		&loc.ClimbTypeID,
		&loc.WoulderLocationID,
		&loc.LastSyncedAt,
		&loc.CreatedAt,
		&loc.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &loc, nil
}

func (r *PostgresRepository) GetSubLocations(ctx context.Context, parentKayaLocationID string) ([]*models.KayaLocation, error) {
	rows, err := r.db.QueryContext(ctx, queryGetSubLocations, parentKayaLocationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []*models.KayaLocation
	for rows.Next() {
		var loc models.KayaLocation
		if err := rows.Scan(
			&loc.ID,
			&loc.KayaLocationID,
			&loc.Slug,
			&loc.Name,
			&loc.Latitude,
			&loc.Longitude,
			&loc.PhotoURL,
			&loc.Description,
			&loc.LocationTypeID,
			&loc.LocationTypeName,
			&loc.ParentLocationID,
			&loc.ParentLocationSlug,
			&loc.ParentLocationName,
			&loc.ClimbCount,
			&loc.BoulderCount,
			&loc.RouteCount,
			&loc.AscentCount,
			&loc.IsGBModeratedBouldering,
			&loc.IsGBModeratedRoutes,
			&loc.IsAccessSensitive,
			&loc.IsClosed,
			&loc.HasMapsDisabled,
			&loc.ClosedDate,
			&loc.DescriptionBouldering,
			&loc.DescriptionRoutes,
			&loc.DescriptionShortBouldering,
			&loc.DescriptionShortRoutes,
			&loc.AccessDescriptionBouldering,
			&loc.AccessDescriptionRoutes,
			&loc.AccessIssuesDescriptionBouldering,
			&loc.AccessIssuesDescriptionRoutes,
			&loc.ClimbTypeID,
			&loc.WoulderLocationID,
			&loc.LastSyncedAt,
			&loc.CreatedAt,
			&loc.UpdatedAt,
		); err != nil {
			return nil, err
		}
		locations = append(locations, &loc)
	}

	return locations, rows.Err()
}

func (r *PostgresRepository) GetAllLocations(ctx context.Context) ([]*models.KayaLocation, error) {
	rows, err := r.db.QueryContext(ctx, queryGetAllLocations)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []*models.KayaLocation
	for rows.Next() {
		var loc models.KayaLocation
		if err := rows.Scan(
			&loc.ID,
			&loc.KayaLocationID,
			&loc.Slug,
			&loc.Name,
			&loc.Latitude,
			&loc.Longitude,
			&loc.PhotoURL,
			&loc.Description,
			&loc.LocationTypeID,
			&loc.LocationTypeName,
			&loc.ParentLocationID,
			&loc.ParentLocationSlug,
			&loc.ParentLocationName,
			&loc.ClimbCount,
			&loc.BoulderCount,
			&loc.RouteCount,
			&loc.AscentCount,
			&loc.IsGBModeratedBouldering,
			&loc.IsGBModeratedRoutes,
			&loc.IsAccessSensitive,
			&loc.IsClosed,
			&loc.HasMapsDisabled,
			&loc.ClosedDate,
			&loc.DescriptionBouldering,
			&loc.DescriptionRoutes,
			&loc.DescriptionShortBouldering,
			&loc.DescriptionShortRoutes,
			&loc.AccessDescriptionBouldering,
			&loc.AccessDescriptionRoutes,
			&loc.AccessIssuesDescriptionBouldering,
			&loc.AccessIssuesDescriptionRoutes,
			&loc.ClimbTypeID,
			&loc.WoulderLocationID,
			&loc.LastSyncedAt,
			&loc.CreatedAt,
			&loc.UpdatedAt,
		); err != nil {
			return nil, err
		}
		locations = append(locations, &loc)
	}

	return locations, rows.Err()
}

// ClimbsRepository implementation

func (r *PostgresRepository) SaveClimb(ctx context.Context, climb *models.KayaClimb) error {
	_, err := r.db.ExecContext(ctx, querySaveClimb,
		climb.KayaClimbID,
		climb.Slug,
		climb.Name,
		climb.GradeID,
		climb.GradeName,
		climb.GradeOrdering,
		climb.GradeClimbTypeID,
		climb.ClimbTypeID,
		climb.ClimbTypeName,
		climb.Rating,
		climb.AscentCount,
		climb.KayaDestinationID,
		climb.KayaDestinationName,
		climb.KayaAreaID,
		climb.KayaAreaName,
		climb.ColorName,
		climb.GymName,
		climb.BoardName,
		climb.IsGBModerated,
		climb.IsAccessSensitive,
		climb.IsClosed,
		climb.IsOffensive,
		climb.WoulderLocationID,
		time.Now(),
	)
	return err
}

func (r *PostgresRepository) GetClimbBySlug(ctx context.Context, slug string) (*models.KayaClimb, error) {
	var climb models.KayaClimb
	err := r.db.QueryRowContext(ctx, queryGetClimbBySlug, slug).Scan(
		&climb.ID,
		&climb.KayaClimbID,
		&climb.Slug,
		&climb.Name,
		&climb.GradeID,
		&climb.GradeName,
		&climb.GradeOrdering,
		&climb.GradeClimbTypeID,
		&climb.ClimbTypeID,
		&climb.ClimbTypeName,
		&climb.Rating,
		&climb.AscentCount,
		&climb.KayaDestinationID,
		&climb.KayaDestinationName,
		&climb.KayaAreaID,
		&climb.KayaAreaName,
		&climb.ColorName,
		&climb.GymName,
		&climb.BoardName,
		&climb.IsGBModerated,
		&climb.IsAccessSensitive,
		&climb.IsClosed,
		&climb.IsOffensive,
		&climb.WoulderLocationID,
		&climb.LastSyncedAt,
		&climb.CreatedAt,
		&climb.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &climb, nil
}

func (r *PostgresRepository) GetClimbsByLocation(ctx context.Context, kayaLocationID string) ([]*models.KayaClimb, error) {
	rows, err := r.db.QueryContext(ctx, queryGetClimbsByLocation, kayaLocationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanClimbs(rows)
}

func (r *PostgresRepository) GetClimbsByDestination(ctx context.Context, kayaDestinationID string) ([]*models.KayaClimb, error) {
	rows, err := r.db.QueryContext(ctx, queryGetClimbsByDestination, kayaDestinationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanClimbs(rows)
}

// GetClimbsOrderedByActivityForWoulderLocation retrieves Kaya climbs with recent activity for a Woulder location
func (r *PostgresRepository) GetClimbsOrderedByActivityForWoulderLocation(ctx context.Context, woulderLocationID int, limit int) ([]models.UnifiedRouteActivitySummary, error) {
	rows, err := r.db.QueryContext(ctx, queryGetClimbsOrderedByActivityForWoulderLocation, woulderLocationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.UnifiedRouteActivitySummary
	for rows.Next() {
		var (
			slug        string
			name        string
			rating      string
			areaName    string
			lastClimbAt time.Time
			daysSince   int
			ascentID    string
			climbedBy   string
			comment     sql.NullString
			userGrade   sql.NullString
		)

		if err := rows.Scan(
			&slug,
			&name,
			&rating,
			&areaName,
			&lastClimbAt,
			&daysSince,
			&ascentID,
			&climbedBy,
			&comment,
			&userGrade,
		); err != nil {
			return nil, err
		}

		// Build most recent ascent summary
		mostRecentAscent := &models.KayaAscentSummary{
			KayaAscentID: ascentID,
			ClimbedAt:    lastClimbAt,
			ClimbedBy:    climbedBy,
		}
		if comment.Valid {
			mostRecentAscent.Comment = &comment.String
		}
		if userGrade.Valid {
			mostRecentAscent.GradeName = &userGrade.String
		}

		// Build unified route summary
		result := models.UnifiedRouteActivitySummary{
			ID:               "kaya-" + slug,
			Name:             name,
			Rating:           rating,
			AreaName:         areaName,
			LastClimbAt:      lastClimbAt,
			DaysSinceClimb:   daysSince,
			Source:           "kaya",
			KayaClimbSlug:    &slug,
			MostRecentAscent: mostRecentAscent,
		}

		results = append(results, result)
	}

	return results, rows.Err()
}

// GetMatchedClimbsForArea retrieves Kaya climbs matched to MP routes in a specific area
func (r *PostgresRepository) GetMatchedClimbsForArea(ctx context.Context, mpAreaID int64, limit int) ([]models.UnifiedRouteActivitySummary, error) {
	rows, err := r.db.QueryContext(ctx, queryGetMatchedClimbsForArea, mpAreaID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.UnifiedRouteActivitySummary
	for rows.Next() {
		var (
			slug            string
			name            string
			rating          string
			areaName        string
			lastClimbAt     time.Time
			daysSince       int
			ascentID        string
			climbedBy       string
			comment         sql.NullString
			userGrade       sql.NullString
			mpRouteID       int64
			matchConfidence float64
		)

		if err := rows.Scan(
			&slug,
			&name,
			&rating,
			&areaName,
			&lastClimbAt,
			&daysSince,
			&ascentID,
			&climbedBy,
			&comment,
			&userGrade,
			&mpRouteID,
			&matchConfidence,
		); err != nil {
			return nil, err
		}

		// Build most recent ascent summary
		mostRecentAscent := &models.KayaAscentSummary{
			KayaAscentID: ascentID,
			ClimbedAt:    lastClimbAt,
			ClimbedBy:    climbedBy,
		}
		if comment.Valid {
			mostRecentAscent.Comment = &comment.String
		}
		if userGrade.Valid {
			mostRecentAscent.GradeName = &userGrade.String
		}

		// Build unified route summary with MP route association
		result := models.UnifiedRouteActivitySummary{
			ID:               "kaya-" + slug,
			Name:             name,
			Rating:           rating,
			AreaName:         areaName,
			LastClimbAt:      lastClimbAt,
			DaysSinceClimb:   daysSince,
			Source:           "kaya",
			KayaClimbSlug:    &slug,
			MPRouteID:        &mpRouteID, // Include MP route ID from match
			MostRecentAscent: mostRecentAscent,
		}

		results = append(results, result)
	}

	return results, rows.Err()
}

func (r *PostgresRepository) scanClimbs(rows *sql.Rows) ([]*models.KayaClimb, error) {
	var climbs []*models.KayaClimb
	for rows.Next() {
		var climb models.KayaClimb
		if err := rows.Scan(
			&climb.ID,
			&climb.KayaClimbID,
			&climb.Slug,
			&climb.Name,
			&climb.GradeID,
			&climb.GradeName,
			&climb.GradeOrdering,
			&climb.GradeClimbTypeID,
			&climb.ClimbTypeID,
			&climb.ClimbTypeName,
			&climb.Rating,
			&climb.AscentCount,
			&climb.KayaDestinationID,
			&climb.KayaDestinationName,
			&climb.KayaAreaID,
			&climb.KayaAreaName,
			&climb.ColorName,
			&climb.GymName,
			&climb.BoardName,
			&climb.IsGBModerated,
			&climb.IsAccessSensitive,
			&climb.IsClosed,
			&climb.IsOffensive,
			&climb.WoulderLocationID,
			&climb.LastSyncedAt,
			&climb.CreatedAt,
			&climb.UpdatedAt,
		); err != nil {
			return nil, err
		}
		climbs = append(climbs, &climb)
	}

	return climbs, rows.Err()
}

// AscentsRepository implementation

func (r *PostgresRepository) SaveAscent(ctx context.Context, ascent *models.KayaAscent) error {
	_, err := r.db.ExecContext(ctx, querySaveAscent,
		ascent.KayaAscentID,
		ascent.KayaClimbSlug,
		ascent.KayaUserID,
		ascent.Date,
		ascent.Comment,
		ascent.Rating,
		ascent.Stiffness,
		ascent.GradeID,
		ascent.GradeName,
		ascent.PhotoURL,
		ascent.PhotoThumbURL,
		ascent.VideoURL,
		ascent.VideoThumbURL,
	)
	return err
}

func (r *PostgresRepository) GetAscentByID(ctx context.Context, kayaAscentID string) (*models.KayaAscent, error) {
	var ascent models.KayaAscent
	err := r.db.QueryRowContext(ctx, queryGetAscentByID, kayaAscentID).Scan(
		&ascent.ID,
		&ascent.KayaAscentID,
		&ascent.KayaClimbSlug,
		&ascent.KayaUserID,
		&ascent.Date,
		&ascent.Comment,
		&ascent.Rating,
		&ascent.Stiffness,
		&ascent.GradeID,
		&ascent.GradeName,
		&ascent.PhotoURL,
		&ascent.PhotoThumbURL,
		&ascent.VideoURL,
		&ascent.VideoThumbURL,
		&ascent.CreatedAt,
		&ascent.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &ascent, nil
}

func (r *PostgresRepository) GetAscentsByClimb(ctx context.Context, kayaClimbSlug string, limit int) ([]*models.KayaAscent, error) {
	rows, err := r.db.QueryContext(ctx, queryGetAscentsByClimb, kayaClimbSlug, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanAscents(rows)
}

func (r *PostgresRepository) GetAscentsByUser(ctx context.Context, kayaUserID string, limit int) ([]*models.KayaAscent, error) {
	rows, err := r.db.QueryContext(ctx, queryGetAscentsByUser, kayaUserID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanAscents(rows)
}

func (r *PostgresRepository) GetRecentAscents(ctx context.Context, limit int) ([]*models.KayaAscent, error) {
	rows, err := r.db.QueryContext(ctx, queryGetRecentAscents, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanAscents(rows)
}

func (r *PostgresRepository) GetAscentsByWoulderLocation(ctx context.Context, woulderLocationID int, limit int) ([]*models.KayaAscent, error) {
	rows, err := r.db.QueryContext(ctx, queryGetAscentsByWoulderLocation, woulderLocationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanAscents(rows)
}

// GetAscentsWithDetailsForWoulderLocation retrieves ascents with climb and user details in a single optimized query
func (r *PostgresRepository) GetAscentsWithDetailsForWoulderLocation(ctx context.Context, woulderLocationID int, limit int) ([]KayaAscentWithDetails, error) {
	rows, err := r.db.QueryContext(ctx, queryGetAscentsWithDetailsForWoulderLocation, woulderLocationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []KayaAscentWithDetails
	for rows.Next() {
		var result KayaAscentWithDetails
		if err := rows.Scan(
			&result.KayaAscentID,
			&result.KayaClimbSlug,
			&result.Date,
			&result.Comment,
			&result.ClimbName,
			&result.ClimbGrade,
			&result.AreaName,
			&result.Username,
		); err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, rows.Err()
}

// GetAscentsForMatchedRoute retrieves Kaya ascents for climbs matched to a specific MP route
func (r *PostgresRepository) GetAscentsForMatchedRoute(ctx context.Context, mpRouteID int64, limit int) ([]KayaAscentWithDetails, error) {
	rows, err := r.db.QueryContext(ctx, queryGetAscentsForMatchedRoute, mpRouteID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []KayaAscentWithDetails
	for rows.Next() {
		var result KayaAscentWithDetails
		if err := rows.Scan(
			&result.KayaAscentID,
			&result.KayaClimbSlug,
			&result.Date,
			&result.Comment,
			&result.ClimbName,
			&result.ClimbGrade,
			&result.AreaName,
			&result.Username,
		); err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, rows.Err()
}

func (r *PostgresRepository) scanAscents(rows *sql.Rows) ([]*models.KayaAscent, error) {
	var ascents []*models.KayaAscent
	for rows.Next() {
		var ascent models.KayaAscent
		if err := rows.Scan(
			&ascent.ID,
			&ascent.KayaAscentID,
			&ascent.KayaClimbSlug,
			&ascent.KayaUserID,
			&ascent.Date,
			&ascent.Comment,
			&ascent.Rating,
			&ascent.Stiffness,
			&ascent.GradeID,
			&ascent.GradeName,
			&ascent.PhotoURL,
			&ascent.PhotoThumbURL,
			&ascent.VideoURL,
			&ascent.VideoThumbURL,
			&ascent.CreatedAt,
			&ascent.UpdatedAt,
		); err != nil {
			return nil, err
		}
		ascents = append(ascents, &ascent)
	}

	return ascents, rows.Err()
}

// PostsRepository implementation

func (r *PostgresRepository) SavePost(ctx context.Context, post *models.KayaPost) error {
	_, err := r.db.ExecContext(ctx, querySavePost,
		post.KayaPostID,
		post.KayaUserID,
		post.DateCreated,
	)
	return err
}

func (r *PostgresRepository) SavePostItem(ctx context.Context, item *models.KayaPostItem) error {
	_, err := r.db.ExecContext(ctx, querySavePostItem,
		item.KayaPostItemID,
		item.KayaPostID,
		item.KayaClimbSlug,
		item.KayaAscentID,
		item.PhotoURL,
		item.VideoURL,
		item.VideoThumbnailURL,
		item.Caption,
	)
	return err
}

func (r *PostgresRepository) GetPostByID(ctx context.Context, kayaPostID string) (*models.KayaPost, error) {
	var post models.KayaPost
	err := r.db.QueryRowContext(ctx, queryGetPostByID, kayaPostID).Scan(
		&post.ID,
		&post.KayaPostID,
		&post.KayaUserID,
		&post.DateCreated,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &post, nil
}

func (r *PostgresRepository) GetPostItemsByPost(ctx context.Context, kayaPostID string) ([]*models.KayaPostItem, error) {
	rows, err := r.db.QueryContext(ctx, queryGetPostItemsByPost, kayaPostID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*models.KayaPostItem
	for rows.Next() {
		var item models.KayaPostItem
		if err := rows.Scan(
			&item.ID,
			&item.KayaPostItemID,
			&item.KayaPostID,
			&item.KayaClimbSlug,
			&item.KayaAscentID,
			&item.PhotoURL,
			&item.VideoURL,
			&item.VideoThumbnailURL,
			&item.Caption,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}

	return items, rows.Err()
}

func (r *PostgresRepository) GetRecentPosts(ctx context.Context, limit int) ([]*models.KayaPost, error) {
	rows, err := r.db.QueryContext(ctx, queryGetRecentPosts, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.KayaPost
	for rows.Next() {
		var post models.KayaPost
		if err := rows.Scan(
			&post.ID,
			&post.KayaPostID,
			&post.KayaUserID,
			&post.DateCreated,
			&post.CreatedAt,
			&post.UpdatedAt,
		); err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}

	return posts, rows.Err()
}

// SyncRepository implementation

func (r *PostgresRepository) SaveSyncProgress(ctx context.Context, progress *models.KayaSyncProgress) error {
	_, err := r.db.ExecContext(ctx, querySaveSyncProgress,
		progress.KayaLocationID,
		progress.LocationName,
		progress.Status,
		progress.LastSyncAt,
		progress.NextSyncAt,
		progress.SyncError,
		progress.ClimbsSynced,
		progress.AscentsSynced,
		progress.SubLocationsSynced,
	)
	return err
}

func (r *PostgresRepository) GetSyncProgress(ctx context.Context, kayaLocationID string) (*models.KayaSyncProgress, error) {
	var progress models.KayaSyncProgress
	err := r.db.QueryRowContext(ctx, queryGetSyncProgress, kayaLocationID).Scan(
		&progress.ID,
		&progress.KayaLocationID,
		&progress.LocationName,
		&progress.Status,
		&progress.LastSyncAt,
		&progress.NextSyncAt,
		&progress.SyncError,
		&progress.ClimbsSynced,
		&progress.AscentsSynced,
		&progress.SubLocationsSynced,
		&progress.CreatedAt,
		&progress.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &progress, nil
}

func (r *PostgresRepository) GetLocationsDueForSync(ctx context.Context, limit int) ([]*models.KayaSyncProgress, error) {
	rows, err := r.db.QueryContext(ctx, queryGetLocationsDueForSync, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var progressList []*models.KayaSyncProgress
	for rows.Next() {
		var progress models.KayaSyncProgress
		if err := rows.Scan(
			&progress.ID,
			&progress.KayaLocationID,
			&progress.LocationName,
			&progress.Status,
			&progress.LastSyncAt,
			&progress.NextSyncAt,
			&progress.SyncError,
			&progress.ClimbsSynced,
			&progress.AscentsSynced,
			&progress.SubLocationsSynced,
			&progress.CreatedAt,
			&progress.UpdatedAt,
		); err != nil {
			return nil, err
		}
		progressList = append(progressList, &progress)
	}

	return progressList, rows.Err()
}

func (r *PostgresRepository) UpdateSyncStatus(ctx context.Context, kayaLocationID string, status string, syncError *string) error {
	_, err := r.db.ExecContext(ctx, queryUpdateSyncStatus, kayaLocationID, status, syncError)
	return err
}

func (r *PostgresRepository) IncrementSyncCounters(ctx context.Context, kayaLocationID string, climbs, ascents, subLocations int) error {
	_, err := r.db.ExecContext(ctx, queryIncrementSyncCounters, kayaLocationID, climbs, ascents, subLocations)
	return err
}

package database

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/alexscott64/woulder/backend/internal/database/areas"
	"github.com/alexscott64/woulder/backend/internal/database/boulders"
	"github.com/alexscott64/woulder/backend/internal/database/climbing"
	"github.com/alexscott64/woulder/backend/internal/database/heatmap"
	"github.com/alexscott64/woulder/backend/internal/database/kaya"
	"github.com/alexscott64/woulder/backend/internal/database/locations"
	"github.com/alexscott64/woulder/backend/internal/database/mountainproject"
	"github.com/alexscott64/woulder/backend/internal/database/rivers"
	"github.com/alexscott64/woulder/backend/internal/database/rocks"
	"github.com/alexscott64/woulder/backend/internal/database/weather"
	"github.com/alexscott64/woulder/backend/internal/models"
	_ "github.com/lib/pq"
)

//go:embed setup_postgres.sql
var setupSQL string

type Database struct {
	conn *sql.DB
}

func New() (*Database, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	if port == "" {
		port = "5432"
	}
	if sslmode == "" {
		sslmode = "require"
	}

	if host == "" || user == "" || password == "" || dbname == "" {
		return nil, fmt.Errorf("missing required database configuration")
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	database := &Database{conn: db}

	needsInit, err := database.needsInitialization()
	if err != nil {
		return nil, err
	}

	if needsInit {
		log.Println("Database schema not found, running setup...")
		if err := database.runSetup(); err != nil {
			return nil, err
		}
	}

	return database, nil
}

func (db *Database) needsInitialization() (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM information_schema.schemata WHERE schema_name = 'woulder')`
	err := db.conn.QueryRow(query).Scan(&exists)
	return !exists, err
}

func (db *Database) runSetup() error {
	_, err := db.conn.Exec(setupSQL)
	return err
}

func (db *Database) Close() error {
	return db.conn.Close()
}

func (db *Database) Ping(ctx context.Context) error {
	return db.conn.PingContext(ctx)
}

// Conn returns the underlying database connection for direct SQL access.
// Used by monitoring and other utilities that need raw database access.
func (db *Database) Conn() *sql.DB {
	return db.conn
}

// Domain Repository Accessors
// These methods provide unified access to domain-specific repositories,
// breaking import cycles by having the database package import domains
// instead of domains importing database.

// Rivers returns the rivers repository for river data operations.
func (db *Database) Rivers() rivers.Repository {
	return rivers.NewPostgresRepository(db.conn)
}

// Weather returns the weather repository for weather data operations.
func (db *Database) Weather() weather.Repository {
	return weather.NewPostgresRepository(db.conn)
}

// Areas returns the areas repository for geographic area operations.
func (db *Database) Areas() areas.Repository {
	return areas.NewPostgresRepository(db.conn)
}

// Locations returns the locations repository for climbing location operations.
func (db *Database) Locations() locations.Repository {
	return locations.NewPostgresRepository(db.conn)
}

// Rocks returns the rocks repository for rock type and sun exposure operations.
func (db *Database) Rocks() rocks.Repository {
	return rocks.NewPostgresRepository(db.conn)
}

// Boulders returns the boulders repository for boulder drying profile operations.
func (db *Database) Boulders() boulders.Repository {
	return boulders.NewPostgresRepository(db.conn)
}

// HeatMap returns the heatmap repository for activity visualization operations.
func (db *Database) HeatMap() heatmap.Repository {
	return heatmap.NewPostgresRepository(db.conn)
}

// Climbing returns the climbing repository for activity and history operations.
func (db *Database) Climbing() climbing.Repository {
	return climbing.NewPostgresRepository(db.conn)
}

// MountainProject returns the Mountain Project repository for MP data operations.
func (db *Database) MountainProject() mountainproject.Repository {
	return mountainproject.NewPostgresRepository(db.conn)
}

// Kaya returns the Kaya repository for Kaya data operations.
func (db *Database) Kaya() kaya.Repository {
	return kaya.NewPostgresRepository(db.conn)
}

// Legacy wrapper methods for backward compatibility
// These delegate to the new domain repositories

// GetMPAreaByID delegates to MountainProject().Areas().GetAreaByID()
func (db *Database) GetMPAreaByID(ctx context.Context, mpAreaID int64) (*models.MPArea, error) {
	return db.MountainProject().Areas().GetAreaByID(ctx, mpAreaID)
}

// GetMPRouteByID delegates to MountainProject().Routes().GetByID()
func (db *Database) GetMPRouteByID(ctx context.Context, mpRouteID int64) (*models.MPRoute, error) {
	return db.MountainProject().Routes().GetByID(ctx, mpRouteID)
}

// GetMPRoutesByIDs delegates to MountainProject().Routes().GetByIDs()
func (db *Database) GetMPRoutesByIDs(ctx context.Context, mpRouteIDs []int64) (map[int64]*models.MPRoute, error) {
	return db.MountainProject().Routes().GetByIDs(ctx, mpRouteIDs)
}

// GetRoutesWithGPSByArea delegates to MountainProject().Routes().GetWithGPSByArea()
func (db *Database) GetRoutesWithGPSByArea(ctx context.Context, mpAreaID int64) ([]*models.MPRoute, error) {
	return db.MountainProject().Routes().GetWithGPSByArea(ctx, mpAreaID)
}

// Weather wrapper methods

// SaveWeatherData delegates to Weather().Save()
func (db *Database) SaveWeatherData(ctx context.Context, data *models.WeatherData) error {
	return db.Weather().Save(ctx, data)
}

// GetHistoricalWeather delegates to Weather().GetHistorical()
func (db *Database) GetHistoricalWeather(ctx context.Context, locationID int, hours int) ([]models.WeatherData, error) {
	return db.Weather().GetHistorical(ctx, locationID, hours)
}

// GetForecastWeather delegates to Weather().GetForecast()
func (db *Database) GetForecastWeather(ctx context.Context, locationID int, hours int) ([]models.WeatherData, error) {
	return db.Weather().GetForecast(ctx, locationID, hours)
}

// GetCurrentWeather delegates to Weather().GetCurrent()
func (db *Database) GetCurrentWeather(ctx context.Context, locationID int) (*models.WeatherData, error) {
	return db.Weather().GetCurrent(ctx, locationID)
}

// CleanOldWeatherData delegates to Weather().CleanOld()
func (db *Database) CleanOldWeatherData(ctx context.Context, daysToKeep int) error {
	return db.Weather().CleanOld(ctx, daysToKeep)
}

// DeleteOldWeatherData delegates to Weather().DeleteOldForLocation()
func (db *Database) DeleteOldWeatherData(ctx context.Context, locationID int, daysToKeep int) error {
	return db.Weather().DeleteOldForLocation(ctx, locationID, daysToKeep)
}

// Location wrapper methods

// GetAllLocations delegates to Locations().GetAll()
func (db *Database) GetAllLocations(ctx context.Context) ([]models.Location, error) {
	return db.Locations().GetAll(ctx)
}

// GetLocation delegates to Locations().GetByID()
func (db *Database) GetLocation(ctx context.Context, id int) (*models.Location, error) {
	return db.Locations().GetByID(ctx, id)
}

// GetLocationsByArea delegates to Locations().GetByArea()
func (db *Database) GetLocationsByArea(ctx context.Context, areaID int) ([]models.Location, error) {
	return db.Locations().GetByArea(ctx, areaID)
}

// GetLocationByID delegates to Locations().GetByID()
func (db *Database) GetLocationByID(ctx context.Context, locationID int) (*models.Location, error) {
	return db.Locations().GetByID(ctx, locationID)
}

// River wrapper methods

// GetRiversByLocation delegates to Rivers().GetByLocation()
func (db *Database) GetRiversByLocation(ctx context.Context, locationID int) ([]models.River, error) {
	return db.Rivers().GetByLocation(ctx, locationID)
}

// GetRiverByID delegates to Rivers().GetByID()
func (db *Database) GetRiverByID(ctx context.Context, id int) (*models.River, error) {
	return db.Rivers().GetByID(ctx, id)
}

// Area wrapper methods

// GetAllAreas delegates to Areas().GetAll()
func (db *Database) GetAllAreas(ctx context.Context) ([]models.Area, error) {
	return db.Areas().GetAll(ctx)
}

// GetAreasWithLocationCounts delegates to Areas().GetAllWithLocationCounts()
func (db *Database) GetAreasWithLocationCounts(ctx context.Context) ([]models.AreaWithLocationCount, error) {
	return db.Areas().GetAllWithLocationCounts(ctx)
}

// GetAreaByID delegates to Areas().GetByID()
func (db *Database) GetAreaByID(ctx context.Context, id int) (*models.Area, error) {
	return db.Areas().GetByID(ctx, id)
}

// Rock type wrapper methods

// GetRockTypesByLocation delegates to Rocks().GetRockTypesByLocation()
func (db *Database) GetRockTypesByLocation(ctx context.Context, locationID int) ([]models.RockType, error) {
	return db.Rocks().GetRockTypesByLocation(ctx, locationID)
}

// GetPrimaryRockType delegates to Rocks().GetPrimaryRockType()
func (db *Database) GetPrimaryRockType(ctx context.Context, locationID int) (*models.RockType, error) {
	return db.Rocks().GetPrimaryRockType(ctx, locationID)
}

// GetSunExposureByLocation delegates to Rocks().GetSunExposureByLocation()
func (db *Database) GetSunExposureByLocation(ctx context.Context, locationID int) (*models.LocationSunExposure, error) {
	return db.Rocks().GetSunExposureByLocation(ctx, locationID)
}

// Mountain Project wrapper methods (additional)

// SaveMPArea delegates to MountainProject().Areas().SaveArea()
func (db *Database) SaveMPArea(ctx context.Context, area *models.MPArea) error {
	return db.MountainProject().Areas().SaveArea(ctx, area)
}

// SaveMPRoute delegates to MountainProject().Routes().SaveRoute()
func (db *Database) SaveMPRoute(ctx context.Context, route *models.MPRoute) error {
	return db.MountainProject().Routes().SaveRoute(ctx, route)
}

// SaveMPTick delegates to MountainProject().Ticks().SaveTick()
func (db *Database) SaveMPTick(ctx context.Context, tick *models.MPTick) error {
	return db.MountainProject().Ticks().SaveTick(ctx, tick)
}

// SaveAreaComment delegates to MountainProject().Comments().SaveAreaComment()
func (db *Database) SaveAreaComment(ctx context.Context, mpCommentID, mpAreaID int64, userName, commentText string, commentedAt time.Time) error {
	return db.MountainProject().Comments().SaveAreaComment(ctx, mpCommentID, mpAreaID, userName, commentText, commentedAt)
}

// SaveRouteComment delegates to MountainProject().Comments().SaveRouteComment()
func (db *Database) SaveRouteComment(ctx context.Context, mpCommentID, mpRouteID int64, userName, commentText string, commentedAt time.Time) error {
	return db.MountainProject().Comments().SaveRouteComment(ctx, mpCommentID, mpRouteID, userName, commentText, commentedAt)
}

// UpdateRouteGPS delegates to MountainProject().Routes().UpdateGPS()
func (db *Database) UpdateRouteGPS(ctx context.Context, routeID int64, latitude, longitude float64, aspect string) error {
	return db.MountainProject().Routes().UpdateGPS(ctx, routeID, latitude, longitude, aspect)
}

// GetLastTickTimestampForRoute delegates to MountainProject().Ticks().GetLastTimestampForRoute()
func (db *Database) GetLastTickTimestampForRoute(ctx context.Context, routeID int64) (*time.Time, error) {
	return db.MountainProject().Ticks().GetLastTimestampForRoute(ctx, routeID)
}

// GetAllRouteIDsForLocation delegates to MountainProject().Routes().GetAllIDsForLocation()
func (db *Database) GetAllRouteIDsForLocation(ctx context.Context, locationID int) ([]int64, error) {
	return db.MountainProject().Routes().GetAllIDsForLocation(ctx, locationID)
}

// UpdateAreaRouteCount delegates to MountainProject().Areas().UpdateRouteCount()
func (db *Database) UpdateAreaRouteCount(ctx context.Context, mpAreaID string, total int) error {
	return db.MountainProject().Areas().UpdateRouteCount(ctx, mpAreaID, total)
}

// GetAreaRouteCount delegates to MountainProject().Areas().GetRouteCount()
func (db *Database) GetAreaRouteCount(ctx context.Context, mpAreaID string) (int, error) {
	return db.MountainProject().Areas().GetRouteCount(ctx, mpAreaID)
}

// GetChildAreas delegates to MountainProject().Areas().GetChildAreas()
func (db *Database) GetChildAreas(ctx context.Context, parentMPAreaID string) ([]struct {
	MPAreaID string
	Name     string
}, error) {
	childAreas, err := db.MountainProject().Areas().GetChildAreas(ctx, parentMPAreaID)
	if err != nil {
		return nil, err
	}
	result := make([]struct {
		MPAreaID string
		Name     string
	}, len(childAreas))
	for i, ca := range childAreas {
		result[i] = struct {
			MPAreaID string
			Name     string
		}{
			MPAreaID: ca.MPAreaID,
			Name:     ca.Name,
		}
	}
	return result, nil
}

// GetRouteIDsForArea delegates to MountainProject().Routes().GetIDsForArea()
func (db *Database) GetRouteIDsForArea(ctx context.Context, mpAreaID string) ([]string, error) {
	return db.MountainProject().Routes().GetIDsForArea(ctx, mpAreaID)
}

// GetAllStateConfigs delegates to MountainProject().Areas().GetAllStateConfigs()
func (db *Database) GetAllStateConfigs(ctx context.Context) ([]struct {
	StateName string
	MPAreaID  string
	IsActive  bool
}, error) {
	stateConfigs, err := db.MountainProject().Areas().GetAllStateConfigs(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]struct {
		StateName string
		MPAreaID  string
		IsActive  bool
	}, len(stateConfigs))
	for i, sc := range stateConfigs {
		result[i] = struct {
			StateName string
			MPAreaID  string
			IsActive  bool
		}{
			StateName: sc.StateName,
			MPAreaID:  sc.MPAreaID,
			IsActive:  sc.IsActive,
		}
	}
	return result, nil
}

// UpsertRoute delegates to MountainProject().Routes().UpsertRoute()
func (db *Database) UpsertRoute(ctx context.Context, mpRouteID, mpAreaID int64, locationID *int, name, routeType, rating string, lat, lon *float64, aspect *string) error {
	return db.MountainProject().Routes().UpsertRoute(ctx, mpRouteID, mpAreaID, locationID, name, routeType, rating, lat, lon, aspect)
}

// UpsertTick delegates to MountainProject().Ticks().UpsertTick()
func (db *Database) UpsertTick(ctx context.Context, mpRouteID int64, userName string, climbedAt time.Time, style string, comment *string) error {
	return db.MountainProject().Ticks().UpsertTick(ctx, mpRouteID, userName, climbedAt, style, comment)
}

// UpsertAreaComment delegates to MountainProject().Comments().UpsertAreaComment()
func (db *Database) UpsertAreaComment(ctx context.Context, mpCommentID, mpAreaID int64, userName string, userID *string, commentText string, commentedAt time.Time) error {
	return db.MountainProject().Comments().UpsertAreaComment(ctx, mpCommentID, mpAreaID, userName, userID, commentText, commentedAt)
}

// UpsertRouteComment delegates to MountainProject().Comments().UpsertRouteComment()
func (db *Database) UpsertRouteComment(ctx context.Context, mpCommentID, mpRouteID int64, userName string, userID *string, commentText string, commentedAt time.Time) error {
	return db.MountainProject().Comments().UpsertRouteComment(ctx, mpCommentID, mpRouteID, userName, userID, commentText, commentedAt)
}

// UpdateRouteSyncPriorities delegates to MountainProject().Sync().UpdateRoutePriorities()
func (db *Database) UpdateRouteSyncPriorities(ctx context.Context) error {
	return db.MountainProject().Sync().UpdateRoutePriorities(ctx)
}

// GetLocationRoutesDueForSync delegates to MountainProject().Sync().GetLocationRoutesDueForSync()
func (db *Database) GetLocationRoutesDueForSync(ctx context.Context, syncType string) ([]int64, error) {
	return db.MountainProject().Sync().GetLocationRoutesDueForSync(ctx, syncType)
}

// GetRoutesDueForTickSync delegates to MountainProject().Sync().GetRoutesDueForTickSync()
func (db *Database) GetRoutesDueForTickSync(ctx context.Context, priority string) ([]int64, error) {
	return db.MountainProject().Sync().GetRoutesDueForTickSync(ctx, priority)
}

// GetRoutesDueForCommentSync delegates to MountainProject().Sync().GetRoutesDueForCommentSync()
func (db *Database) GetRoutesDueForCommentSync(ctx context.Context, priority string) ([]int64, error) {
	return db.MountainProject().Sync().GetRoutesDueForCommentSync(ctx, priority)
}

// GetPriorityDistribution delegates to MountainProject().Sync().GetPriorityDistribution()
func (db *Database) GetPriorityDistribution(ctx context.Context) (map[string]int, error) {
	return db.MountainProject().Sync().GetPriorityDistribution(ctx)
}

// Boulder drying wrapper methods

// GetBoulderDryingProfile delegates to Boulders().GetProfile()
func (db *Database) GetBoulderDryingProfile(ctx context.Context, mpRouteID int64) (*models.BoulderDryingProfile, error) {
	return db.Boulders().GetProfile(ctx, mpRouteID)
}

// GetBoulderDryingProfilesByRouteIDs delegates to Boulders().GetProfilesByIDs()
func (db *Database) GetBoulderDryingProfilesByRouteIDs(ctx context.Context, mpRouteIDs []int64) (map[int64]*models.BoulderDryingProfile, error) {
	return db.Boulders().GetProfilesByIDs(ctx, mpRouteIDs)
}

// SaveBoulderDryingProfile delegates to Boulders().SaveProfile()
func (db *Database) SaveBoulderDryingProfile(ctx context.Context, profile *models.BoulderDryingProfile) error {
	return db.Boulders().SaveProfile(ctx, profile)
}

// Climbing/Activity wrapper methods

// GetLastClimbedForLocation delegates to Climbing().History().GetLastClimbedForLocation()
func (db *Database) GetLastClimbedForLocation(ctx context.Context, locationID int) (*models.LastClimbedInfo, error) {
	return db.Climbing().History().GetLastClimbedForLocation(ctx, locationID)
}

// GetClimbHistoryForLocation delegates to Climbing().History().GetClimbHistoryForLocation()
func (db *Database) GetClimbHistoryForLocation(ctx context.Context, locationID int, limit int) ([]models.ClimbHistoryEntry, error) {
	return db.Climbing().History().GetClimbHistoryForLocation(ctx, locationID, limit)
}

// GetAreasOrderedByActivity delegates to Climbing().Activity().GetAreasOrderedByActivity()
func (db *Database) GetAreasOrderedByActivity(ctx context.Context, locationID int) ([]models.AreaActivitySummary, error) {
	return db.Climbing().Activity().GetAreasOrderedByActivity(ctx, locationID)
}

// GetSubareasOrderedByActivity delegates to Climbing().Activity().GetSubareasOrderedByActivity()
func (db *Database) GetSubareasOrderedByActivity(ctx context.Context, parentAreaID int64, locationID int) ([]models.AreaActivitySummary, error) {
	return db.Climbing().Activity().GetSubareasOrderedByActivity(ctx, parentAreaID, locationID)
}

// GetRoutesOrderedByActivity delegates to Climbing().Activity().GetRoutesOrderedByActivity()
func (db *Database) GetRoutesOrderedByActivity(ctx context.Context, areaID int64, locationID int, limit int) ([]models.RouteActivitySummary, error) {
	return db.Climbing().Activity().GetRoutesOrderedByActivity(ctx, areaID, locationID, limit)
}

// GetRecentTicksForRoute delegates to Climbing().Activity().GetRecentTicksForRoute()
func (db *Database) GetRecentTicksForRoute(ctx context.Context, routeID int64, limit int) ([]models.ClimbHistoryEntry, error) {
	return db.Climbing().Activity().GetRecentTicksForRoute(ctx, routeID, limit)
}

// SearchInLocation delegates to Climbing().Search().SearchInLocation()
func (db *Database) SearchInLocation(ctx context.Context, locationID int, searchQuery string, limit int) ([]models.SearchResult, error) {
	return db.Climbing().Search().SearchInLocation(ctx, locationID, searchQuery, limit)
}

// SearchRoutesInLocation delegates to Climbing().Search().SearchRoutesInLocation()
func (db *Database) SearchRoutesInLocation(ctx context.Context, locationID int, searchQuery string, limit int) ([]models.RouteActivitySummary, error) {
	return db.Climbing().Search().SearchRoutesInLocation(ctx, locationID, searchQuery, limit)
}

// Heat map wrapper methods

// GetHeatMapData delegates to HeatMap().GetData()
func (db *Database) GetHeatMapData(ctx context.Context, startDate, endDate time.Time, bounds *GeoBounds, minActivity, limit int, routeTypes []string, lightweight bool) ([]models.HeatMapPoint, error) {
	var heatmapBounds *heatmap.GeoBounds
	if bounds != nil {
		heatmapBounds = &heatmap.GeoBounds{
			MinLat: bounds.MinLat,
			MaxLat: bounds.MaxLat,
			MinLon: bounds.MinLon,
			MaxLon: bounds.MaxLon,
		}
	}
	return db.HeatMap().GetHeatMapData(ctx, startDate, endDate, heatmapBounds, minActivity, limit, routeTypes, lightweight)
}

// GetAreaActivityDetail delegates to HeatMap().GetAreaActivityDetail()
func (db *Database) GetAreaActivityDetail(ctx context.Context, areaID int64, startDate, endDate time.Time) (*models.AreaActivityDetail, error) {
	return db.HeatMap().GetAreaActivityDetail(ctx, areaID, startDate, endDate)
}

// GetRoutesByBounds delegates to HeatMap().GetRoutesByBounds()
func (db *Database) GetRoutesByBounds(ctx context.Context, bounds GeoBounds, startDate, endDate time.Time, limit int) ([]models.RouteActivity, error) {
	heatmapBounds := heatmap.GeoBounds{
		MinLat: bounds.MinLat,
		MaxLat: bounds.MaxLat,
		MinLon: bounds.MinLon,
		MaxLon: bounds.MaxLon,
	}
	return db.HeatMap().GetRoutesByBounds(ctx, heatmapBounds, startDate, endDate, limit)
}

// GetRouteTicksInDateRange delegates to HeatMap().GetRouteTicksInDateRange()
func (db *Database) GetRouteTicksInDateRange(ctx context.Context, routeID int64, startDate, endDate time.Time, limit int) ([]models.TickDetail, error) {
	return db.HeatMap().GetRouteTicksInDateRange(ctx, routeID, startDate, endDate, limit)
}

// SearchRoutesInAreas delegates to HeatMap().SearchRoutesInAreas()
func (db *Database) SearchRoutesInAreas(ctx context.Context, areaIDs []int64, searchQuery string, startDate, endDate time.Time, limit int) ([]models.RouteActivity, error) {
	return db.HeatMap().SearchRoutesInAreas(ctx, areaIDs, searchQuery, startDate, endDate, limit)
}

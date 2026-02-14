package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alexscott64/woulder/backend/internal/database"
	"github.com/alexscott64/woulder/backend/internal/models"
)

// MockHeatMapRepository for testing
type MockHeatMapRepository struct {
	GetHeatMapDataFunc        func(ctx context.Context, startDate, endDate time.Time, bounds *database.GeoBounds, minActivity, limit int, routeTypes []string, lightweight bool) ([]models.HeatMapPoint, error)
	GetAreaActivityDetailFunc func(ctx context.Context, areaID int64, startDate, endDate time.Time) (*models.AreaActivityDetail, error)
	GetRoutesByBoundsFunc     func(ctx context.Context, bounds database.GeoBounds, startDate, endDate time.Time, limit int) ([]models.RouteActivity, error)
}

func (m *MockHeatMapRepository) GetHeatMapData(ctx context.Context, startDate, endDate time.Time, bounds *database.GeoBounds, minActivity, limit int, routeTypes []string, lightweight bool) ([]models.HeatMapPoint, error) {
	if m.GetHeatMapDataFunc != nil {
		return m.GetHeatMapDataFunc(ctx, startDate, endDate, bounds, minActivity, limit, routeTypes, lightweight)
	}
	return nil, errors.New("not implemented")
}

func (m *MockHeatMapRepository) GetAreaActivityDetail(ctx context.Context, areaID int64, startDate, endDate time.Time) (*models.AreaActivityDetail, error) {
	if m.GetAreaActivityDetailFunc != nil {
		return m.GetAreaActivityDetailFunc(ctx, areaID, startDate, endDate)
	}
	return nil, errors.New("not implemented")
}

func (m *MockHeatMapRepository) GetRoutesByBounds(ctx context.Context, bounds database.GeoBounds, startDate, endDate time.Time, limit int) ([]models.RouteActivity, error) {
	if m.GetRoutesByBoundsFunc != nil {
		return m.GetRoutesByBoundsFunc(ctx, bounds, startDate, endDate, limit)
	}
	return nil, errors.New("not implemented")
}

// Implement other Repository methods as no-ops
func (m *MockHeatMapRepository) GetAllLocations(ctx context.Context) ([]models.Location, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetLocation(ctx context.Context, id int) (*models.Location, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetLocationsByArea(ctx context.Context, areaID int) ([]models.Location, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) SaveWeatherData(ctx context.Context, data *models.WeatherData) error {
	return errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetHistoricalWeather(ctx context.Context, locationID int, hours int) ([]models.WeatherData, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetForecastWeather(ctx context.Context, locationID int, hours int) ([]models.WeatherData, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetCurrentWeather(ctx context.Context, locationID int) (*models.WeatherData, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) CleanOldWeatherData(ctx context.Context, daysToKeep int) error {
	return errors.New("not implemented")
}
func (m *MockHeatMapRepository) DeleteOldWeatherData(ctx context.Context, locationID int, daysToKeep int) error {
	return errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetRiversByLocation(ctx context.Context, locationID int) ([]models.River, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetRiverByID(ctx context.Context, id int) (*models.River, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetAllAreas(ctx context.Context) ([]models.Area, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetAreasWithLocationCounts(ctx context.Context) ([]models.AreaWithLocationCount, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetAreaByID(ctx context.Context, id int) (*models.Area, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetRockTypesByLocation(ctx context.Context, locationID int) ([]models.RockType, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetPrimaryRockType(ctx context.Context, locationID int) (*models.RockType, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetSunExposureByLocation(ctx context.Context, locationID int) (*models.LocationSunExposure, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) SaveMPArea(ctx context.Context, area *models.MPArea) error {
	return errors.New("not implemented")
}
func (m *MockHeatMapRepository) SaveMPRoute(ctx context.Context, route *models.MPRoute) error {
	return errors.New("not implemented")
}
func (m *MockHeatMapRepository) SaveMPTick(ctx context.Context, tick *models.MPTick) error {
	return errors.New("not implemented")
}
func (m *MockHeatMapRepository) SaveAreaComment(ctx context.Context, mpCommentID, mpAreaID int64, userName, commentText string, commentedAt time.Time) error {
	return errors.New("not implemented")
}
func (m *MockHeatMapRepository) SaveRouteComment(ctx context.Context, mpCommentID, mpRouteID int64, userName, commentText string, commentedAt time.Time) error {
	return errors.New("not implemented")
}
func (m *MockHeatMapRepository) UpdateRouteGPS(ctx context.Context, routeID int64, latitude, longitude float64, aspect string) error {
	return errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetLastClimbedForLocation(ctx context.Context, locationID int) (*models.LastClimbedInfo, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetClimbHistoryForLocation(ctx context.Context, locationID int, limit int) ([]models.ClimbHistoryEntry, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetMPAreaByID(ctx context.Context, mpAreaID int64) (*models.MPArea, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetMPRouteByID(ctx context.Context, mpRouteID int64) (*models.MPRoute, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetLastTickTimestampForRoute(ctx context.Context, routeID int64) (*time.Time, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetAllRouteIDsForLocation(ctx context.Context, locationID int) ([]int64, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetAreasOrderedByActivity(ctx context.Context, locationID int) ([]models.AreaActivitySummary, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetSubareasOrderedByActivity(ctx context.Context, parentAreaID int64, locationID int) ([]models.AreaActivitySummary, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetRoutesOrderedByActivity(ctx context.Context, areaID int64, locationID int, limit int) ([]models.RouteActivitySummary, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetRecentTicksForRoute(ctx context.Context, routeID int64, limit int) ([]models.ClimbHistoryEntry, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) SearchInLocation(ctx context.Context, locationID int, searchQuery string, limit int) ([]models.SearchResult, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) SearchRoutesInLocation(ctx context.Context, locationID int, searchQuery string, limit int) ([]models.RouteActivitySummary, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) UpdateAreaRouteCount(ctx context.Context, mpAreaID string, total int) error {
	return errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetAreaRouteCount(ctx context.Context, mpAreaID string) (int, error) {
	return 0, errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetChildAreas(ctx context.Context, parentMPAreaID string) ([]struct {
	MPAreaID string
	Name     string
}, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetRouteIDsForArea(ctx context.Context, mpAreaID string) ([]string, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetAllStateConfigs(ctx context.Context) ([]struct {
	StateName string
	MPAreaID  string
	IsActive  bool
}, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) UpsertRoute(ctx context.Context, mpRouteID, mpAreaID int64, locationID *int, name, routeType, rating string, lat, lon *float64, aspect *string) error {
	return errors.New("not implemented")
}
func (m *MockHeatMapRepository) UpsertTick(ctx context.Context, mpRouteID int64, userName string, climbedAt time.Time, style string, comment *string) error {
	return errors.New("not implemented")
}
func (m *MockHeatMapRepository) UpsertAreaComment(ctx context.Context, mpCommentID, mpAreaID int64, userName string, userID *string, commentText string, commentedAt time.Time) error {
	return errors.New("not implemented")
}
func (m *MockHeatMapRepository) UpsertRouteComment(ctx context.Context, mpCommentID, mpRouteID int64, userName string, userID *string, commentText string, commentedAt time.Time) error {
	return errors.New("not implemented")
}
func (m *MockHeatMapRepository) UpdateRouteSyncPriorities(ctx context.Context) error {
	return errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetLocationRoutesDueForSync(ctx context.Context, syncType string) ([]int64, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetRoutesDueForTickSync(ctx context.Context, priority string) ([]int64, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetRoutesDueForCommentSync(ctx context.Context, priority string) ([]int64, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetPriorityDistribution(ctx context.Context) (map[string]int, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetBoulderDryingProfile(ctx context.Context, mpRouteID int64) (*models.BoulderDryingProfile, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetBoulderDryingProfilesByRouteIDs(ctx context.Context, mpRouteIDs []int64) (map[int64]*models.BoulderDryingProfile, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) SaveBoulderDryingProfile(ctx context.Context, profile *models.BoulderDryingProfile) error {
	return errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetLocationByID(ctx context.Context, locationID int) (*models.Location, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetRoutesWithGPSByArea(ctx context.Context, mpAreaID int64) ([]*models.MPRoute, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetMPRoutesByIDs(ctx context.Context, mpRouteIDs []int64) (map[int64]*models.MPRoute, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) Ping(ctx context.Context) error {
	return errors.New("not implemented")
}
func (m *MockHeatMapRepository) Close() error {
	return errors.New("not implemented")
}
func (m *MockHeatMapRepository) GetRouteTicksInDateRange(ctx context.Context, routeID int64, startDate, endDate time.Time, limit int) ([]models.TickDetail, error) {
	return nil, errors.New("not implemented")
}
func (m *MockHeatMapRepository) SearchRoutesInAreas(ctx context.Context, areaIDs []int64, searchQuery string, startDate, endDate time.Time, limit int) ([]models.RouteActivity, error) {
	return nil, errors.New("not implemented")
}

// Tests
func TestHeatMapService_GetHeatMapData(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	thirtyDaysAgo := now.AddDate(0, 0, -30)

	t.Run("successfully retrieves and calculates activity scores", func(t *testing.T) {
		mockRepo := &MockHeatMapRepository{
			GetHeatMapDataFunc: func(ctx context.Context, startDate, endDate time.Time, bounds *database.GeoBounds, minActivity, limit int, routeTypes []string, lightweight bool) ([]models.HeatMapPoint, error) {
				return []models.HeatMapPoint{
					{
						MPAreaID:       1,
						Name:           "Test Area 1",
						Latitude:       47.5,
						Longitude:      -121.5,
						TotalTicks:     100,
						LastActivity:   now.AddDate(0, 0, -5), // 5 days ago
						ActiveRoutes:   50,
						UniqueClimbers: 25,
					},
					{
						MPAreaID:       2,
						Name:           "Test Area 2",
						Latitude:       47.6,
						Longitude:      -121.6,
						TotalTicks:     50,
						LastActivity:   now.AddDate(0, 0, -20), // 20 days ago
						ActiveRoutes:   25,
						UniqueClimbers: 15,
					},
				}, nil
			},
		}

		service := NewHeatMapService(mockRepo)
		points, err := service.GetHeatMapData(ctx, thirtyDaysAgo, now, nil, 1, 500, nil, false)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(points) != 2 {
			t.Fatalf("Expected 2 points, got %d", len(points))
		}

		// First point should have higher score (more recent)
		if points[0].ActivityScore <= points[1].ActivityScore {
			t.Errorf("Expected first point to have higher activity score")
		}

		// Verify recency multiplier applied
		if points[0].ActivityScore != 200 { // 100 * 2.0 (last week)
			t.Errorf("Expected activity score 200, got %d", points[0].ActivityScore)
		}

		if points[1].ActivityScore != 75 { // 50 * 1.5 (last month)
			t.Errorf("Expected activity score 75, got %d", points[1].ActivityScore)
		}
	})

	t.Run("validates date range", func(t *testing.T) {
		mockRepo := &MockHeatMapRepository{}
		service := NewHeatMapService(mockRepo)

		// Invalid: start after end
		_, err := service.GetHeatMapData(ctx, now, thirtyDaysAgo, nil, 1, 500, nil, false)

		if err == nil {
			t.Error("Expected error for invalid date range")
		}
	})

	t.Run("validates bounds", func(t *testing.T) {
		mockRepo := &MockHeatMapRepository{}
		service := NewHeatMapService(mockRepo)

		invalidBounds := &database.GeoBounds{
			MinLat: 50.0,
			MaxLat: 40.0, // Invalid: min > max
			MinLon: -125.0,
			MaxLon: -120.0,
		}

		_, err := service.GetHeatMapData(ctx, thirtyDaysAgo, now, invalidBounds, 1, 500, nil, false)

		if err == nil {
			t.Error("Expected error for invalid bounds")
		}
	})
}

func TestHeatMapService_calculateActivityScore(t *testing.T) {
	service := &HeatMapService{}
	now := time.Now()

	tests := []struct {
		name         string
		tickCount    int
		lastActivity time.Time
		endDate      time.Time
		wantScore    int
	}{
		{
			name:         "last week - 2x multiplier",
			tickCount:    100,
			lastActivity: now.AddDate(0, 0, -3),
			endDate:      now,
			wantScore:    200,
		},
		{
			name:         "last month - 1.5x multiplier",
			tickCount:    100,
			lastActivity: now.AddDate(0, 0, -20),
			endDate:      now,
			wantScore:    150,
		},
		{
			name:         "older - 1x multiplier",
			tickCount:    100,
			lastActivity: now.AddDate(0, 0, -60),
			endDate:      now,
			wantScore:    100,
		},
		{
			name:         "minimum score of 1",
			tickCount:    0,
			lastActivity: now,
			endDate:      now,
			wantScore:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := service.calculateActivityScore(tt.tickCount, tt.lastActivity, tt.endDate)
			if score != tt.wantScore {
				t.Errorf("Expected score %d, got %d", tt.wantScore, score)
			}
		})
	}
}

func TestGeoBounds_Validate(t *testing.T) {
	tests := []struct {
		name    string
		bounds  database.GeoBounds
		wantErr bool
	}{
		{
			name: "valid bounds",
			bounds: database.GeoBounds{
				MinLat: 40.0, MaxLat: 45.0,
				MinLon: -125.0, MaxLon: -120.0,
			},
			wantErr: false,
		},
		{
			name: "invalid - minLat > maxLat",
			bounds: database.GeoBounds{
				MinLat: 45.0, MaxLat: 40.0,
				MinLon: -125.0, MaxLon: -120.0,
			},
			wantErr: true,
		},
		{
			name: "invalid - latitude out of range",
			bounds: database.GeoBounds{
				MinLat: -95.0, MaxLat: 45.0,
				MinLon: -125.0, MaxLon: -120.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.bounds.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

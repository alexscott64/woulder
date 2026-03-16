package service

import (
	"context"
	"testing"
	"time"

	"github.com/alexscott64/woulder/backend/internal/database/heatmap"
	"github.com/alexscott64/woulder/backend/internal/models"
)

// Tests
func TestHeatMapService_GetHeatMapData(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	thirtyDaysAgo := now.AddDate(0, 0, -30)

	t.Run("successfully retrieves and calculates activity scores", func(t *testing.T) {
		mockRepo := &MockHeatMapRepository{
			GetHeatMapDataFn: func(ctx context.Context, startDate, endDate time.Time, bounds *heatmap.GeoBounds, minActivity, limit int, routeTypes []string, lightweight bool, gradeOrders []int) ([]models.HeatMapPoint, error) {
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
		points, err := service.GetHeatMapData(ctx, thirtyDaysAgo, now, nil, 1, 500, nil, false, nil)

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
		_, err := service.GetHeatMapData(ctx, now, thirtyDaysAgo, nil, 1, 500, nil, false, nil)

		if err == nil {
			t.Error("Expected error for invalid date range")
		}
	})

	t.Run("passes grade orders filter to repository", func(t *testing.T) {
		var receivedGradeOrders []int
		mockRepo := &MockHeatMapRepository{
			GetHeatMapDataFn: func(ctx context.Context, startDate, endDate time.Time, bounds *heatmap.GeoBounds, minActivity, limit int, routeTypes []string, lightweight bool, gradeOrders []int) ([]models.HeatMapPoint, error) {
				receivedGradeOrders = gradeOrders
				return []models.HeatMapPoint{
					{
						MPAreaID:     1,
						Name:         "Boulder Area",
						TotalTicks:   50,
						LastActivity: now.AddDate(0, 0, -5),
					},
				}, nil
			},
		}

		service := NewHeatMapService(mockRepo)

		// Test boulder V9-V17 filtering (the reported bug scenario)
		boulderGradeOrders := []int{9, 10, 11, 12, 13, 14, 15, 16, 17}
		points, err := service.GetHeatMapData(ctx, thirtyDaysAgo, now, nil, 1, 500, []string{"Boulder"}, false, boulderGradeOrders)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(points) != 1 {
			t.Fatalf("Expected 1 point, got %d", len(points))
		}

		if len(receivedGradeOrders) != 9 {
			t.Fatalf("Expected 9 grade orders passed to repo, got %d", len(receivedGradeOrders))
		}

		if receivedGradeOrders[0] != 9 || receivedGradeOrders[8] != 17 {
			t.Errorf("Expected grade orders [9..17], got %v", receivedGradeOrders)
		}
	})

	t.Run("passes multi-type grade filters to repository", func(t *testing.T) {
		var receivedGradeOrders []int
		var receivedRouteTypes []string
		mockRepo := &MockHeatMapRepository{
			GetHeatMapDataFn: func(ctx context.Context, startDate, endDate time.Time, bounds *heatmap.GeoBounds, minActivity, limit int, routeTypes []string, lightweight bool, gradeOrders []int) ([]models.HeatMapPoint, error) {
				receivedGradeOrders = gradeOrders
				receivedRouteTypes = routeTypes
				return []models.HeatMapPoint{}, nil
			},
		}

		service := NewHeatMapService(mockRepo)

		// Multi-type: Boulder V0-V2 + Ice WI1-WI3
		multiGradeOrders := []int{0, 1, 2, 200, 201, 202}
		multiRouteTypes := []string{"Boulder", "Ice"}
		_, err := service.GetHeatMapData(ctx, thirtyDaysAgo, now, nil, 1, 500, multiRouteTypes, false, multiGradeOrders)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(receivedGradeOrders) != 6 {
			t.Fatalf("Expected 6 grade orders, got %d", len(receivedGradeOrders))
		}

		if len(receivedRouteTypes) != 2 {
			t.Fatalf("Expected 2 route types, got %d", len(receivedRouteTypes))
		}

		// Verify boulder orders (0-2) and WI orders (200-202) are both present
		hasV0 := false
		hasWI1 := false
		for _, o := range receivedGradeOrders {
			if o == 0 {
				hasV0 = true
			}
			if o == 200 {
				hasWI1 = true
			}
		}
		if !hasV0 {
			t.Error("Expected V0 (order 0) in grade orders")
		}
		if !hasWI1 {
			t.Error("Expected WI1 (order 200) in grade orders")
		}
	})

	t.Run("nil grade orders passes through (no filtering)", func(t *testing.T) {
		var receivedGradeOrders []int
		mockRepo := &MockHeatMapRepository{
			GetHeatMapDataFn: func(ctx context.Context, startDate, endDate time.Time, bounds *heatmap.GeoBounds, minActivity, limit int, routeTypes []string, lightweight bool, gradeOrders []int) ([]models.HeatMapPoint, error) {
				receivedGradeOrders = gradeOrders
				return []models.HeatMapPoint{}, nil
			},
		}

		service := NewHeatMapService(mockRepo)
		_, err := service.GetHeatMapData(ctx, thirtyDaysAgo, now, nil, 1, 500, nil, false, nil)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if receivedGradeOrders != nil {
			t.Errorf("Expected nil grade orders (no filter), got %v", receivedGradeOrders)
		}
	})

	t.Run("validates bounds", func(t *testing.T) {
		mockRepo := &MockHeatMapRepository{}
		service := NewHeatMapService(mockRepo)

		invalidBounds := &heatmap.GeoBounds{
			MinLat: 50.0,
			MaxLat: 40.0, // Invalid: min > max
			MinLon: -125.0,
			MaxLon: -120.0,
		}

		_, err := service.GetHeatMapData(ctx, thirtyDaysAgo, now, invalidBounds, 1, 500, nil, false, nil)

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
		bounds  heatmap.GeoBounds
		wantErr bool
	}{
		{
			name: "valid bounds",
			bounds: heatmap.GeoBounds{
				MinLat: 40.0, MaxLat: 45.0,
				MinLon: -125.0, MaxLon: -120.0,
			},
			wantErr: false,
		},
		{
			name: "invalid - minLat > maxLat",
			bounds: heatmap.GeoBounds{
				MinLat: 45.0, MaxLat: 40.0,
				MinLon: -125.0, MaxLon: -120.0,
			},
			wantErr: true,
		},
		{
			name: "invalid - latitude out of range",
			bounds: heatmap.GeoBounds{
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

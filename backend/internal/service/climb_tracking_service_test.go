package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/alexscott64/woulder/backend/internal/database"
	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/alexscott64/woulder/backend/internal/mountainproject"
	"github.com/stretchr/testify/assert"
)

// Helper function to create a Tick with user data
func createTickWithUser(date, userName, style string) mountainproject.Tick {
	userJSON, _ := json.Marshal(map[string]interface{}{
		"id":   123,
		"name": userName,
	})
	textJSON, _ := json.Marshal("")

	return mountainproject.Tick{
		Date:  date,
		Style: style,
		User:  userJSON,
		Text:  textJSON,
	}
}

// Mock Mountain Project Client
type MockMPClient struct {
	GetRouteTicksFn func(routeID string) ([]mountainproject.Tick, error)
	GetAreaFn       func(areaID string) (*mountainproject.AreaResponse, error)
}

func (m *MockMPClient) GetRouteTicks(routeID string) ([]mountainproject.Tick, error) {
	if m.GetRouteTicksFn != nil {
		return m.GetRouteTicksFn(routeID)
	}
	return nil, nil
}

func (m *MockMPClient) GetArea(areaID string) (*mountainproject.AreaResponse, error) {
	if m.GetAreaFn != nil {
		return m.GetAreaFn(areaID)
	}
	return nil, nil
}

func TestClimbTrackingService_GetClimbHistoryForLocation(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		mockFn  func(ctx context.Context, locationID int, limit int) ([]models.ClimbHistoryEntry, error)
		locID   int
		limit   int
		want    int
		wantErr bool
	}{
		{
			name: "success with results",
			mockFn: func(ctx context.Context, locationID int, limit int) ([]models.ClimbHistoryEntry, error) {
				return []models.ClimbHistoryEntry{
					{
						MPRouteID:      "123",
						RouteName:      "Test Route",
						RouteRating:    "V5",
						MPAreaID:       "456",
						AreaName:       "Test Area",
						ClimbedAt:      now,
						ClimbedBy:      "Test User",
						Style:          "Flash",
						DaysSinceClimb: 1,
					},
				}, nil
			},
			locID:   1,
			limit:   5,
			want:    1,
			wantErr: false,
		},
		{
			name: "empty results",
			mockFn: func(ctx context.Context, locationID int, limit int) ([]models.ClimbHistoryEntry, error) {
				return []models.ClimbHistoryEntry{}, nil
			},
			locID:   999,
			limit:   5,
			want:    0,
			wantErr: false,
		},
		{
			name: "database error",
			mockFn: func(ctx context.Context, locationID int, limit int) ([]models.ClimbHistoryEntry, error) {
				return nil, errors.New("database error")
			},
			locID:   1,
			limit:   5,
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &database.MockRepository{
				GetClimbHistoryForLocationFn: tt.mockFn,
			}

			service := NewClimbTrackingService(mockRepo, &MockMPClient{})
			history, err := service.GetClimbHistoryForLocation(context.Background(), tt.locID, tt.limit)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, history, tt.want)
			}
		})
	}
}

func TestClimbTrackingService_SyncNewTicksForLocation(t *testing.T) {
	now := time.Now()
	oldTick := now.Add(-48 * time.Hour)
	newTick := now.Add(-1 * time.Hour)

	tests := []struct {
		name           string
		locationID     int
		mockRouteIDs   []string
		mockLastTick   *time.Time
		mockMPTicks    []mountainproject.Tick
		mockGetRoutes  func(ctx context.Context, locationID int) ([]string, error)
		mockGetLastFn  func(ctx context.Context, routeID string) (*time.Time, error)
		mockSaveTickFn func(ctx context.Context, tick *models.MPTick) error
		wantErr        bool
		expectedSaved  int
	}{
		{
			name:         "success - new ticks only",
			locationID:   1,
			mockRouteIDs: []string{"123"},
			mockLastTick: &oldTick,
			mockMPTicks: []mountainproject.Tick{
				createTickWithUser(newTick.Format("Jan 2, 2006, 3:04 pm"), "TestUser", "Flash"),
			},
			mockGetRoutes: func(ctx context.Context, locationID int) ([]string, error) {
				return []string{"123"}, nil
			},
			mockGetLastFn: func(ctx context.Context, routeID string) (*time.Time, error) {
				return &oldTick, nil
			},
			mockSaveTickFn: func(ctx context.Context, tick *models.MPTick) error {
				return nil
			},
			wantErr:       false,
			expectedSaved: 1,
		},
		{
			name:         "skip old ticks",
			locationID:   1,
			mockRouteIDs: []string{"123"},
			mockLastTick: &newTick,
			mockMPTicks: []mountainproject.Tick{
				createTickWithUser(oldTick.Format("Jan 2, 2006, 3:04 pm"), "TestUser", "Send"),
			},
			mockGetRoutes: func(ctx context.Context, locationID int) ([]string, error) {
				return []string{"123"}, nil
			},
			mockGetLastFn: func(ctx context.Context, routeID string) (*time.Time, error) {
				return &newTick, nil
			},
			mockSaveTickFn: func(ctx context.Context, tick *models.MPTick) error {
				t.Error("SaveMPTick should not be called for old ticks")
				return nil
			},
			wantErr:       false,
			expectedSaved: 0,
		},
		{
			name:         "no routes for location",
			locationID:   999,
			mockRouteIDs: []string{},
			mockGetRoutes: func(ctx context.Context, locationID int) ([]string, error) {
				return []string{}, nil
			},
			wantErr:       false,
			expectedSaved: 0,
		},
		{
			name:       "error getting routes",
			locationID: 1,
			mockGetRoutes: func(ctx context.Context, locationID int) ([]string, error) {
				return nil, errors.New("database error")
			},
			wantErr:       true,
			expectedSaved: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			savedCount := 0

			mockRepo := &database.MockRepository{
				GetAllRouteIDsForLocationFn: tt.mockGetRoutes,
				GetLastTickTimestampForRouteFn: tt.mockGetLastFn,
				SaveMPTickFn: func(ctx context.Context, tick *models.MPTick) error {
					savedCount++
					if tt.mockSaveTickFn != nil {
						return tt.mockSaveTickFn(ctx, tick)
					}
					return nil
				},
			}

			mockMPClient := &MockMPClient{
				GetRouteTicksFn: func(routeID string) ([]mountainproject.Tick, error) {
					return tt.mockMPTicks, nil
				},
			}

			service := NewClimbTrackingService(mockRepo, mockMPClient)
			err := service.SyncNewTicksForLocation(context.Background(), tt.locationID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSaved, savedCount, "Expected %d ticks to be saved, got %d", tt.expectedSaved, savedCount)
			}
		})
	}
}

func TestClimbTrackingService_GetSyncStatus(t *testing.T) {
	mockRepo := &database.MockRepository{}
	mockMPClient := &MockMPClient{}

	service := NewClimbTrackingService(mockRepo, mockMPClient)

	// Initially not syncing
	isSyncing, lastSync := service.GetSyncStatus()
	assert.False(t, isSyncing)
	assert.True(t, lastSync.IsZero())

	// Start a sync
	service.syncMutex.Lock()
	service.isSyncing = true
	service.syncMutex.Unlock()

	isSyncing, _ = service.GetSyncStatus()
	assert.True(t, isSyncing)

	// Complete sync
	service.syncMutex.Lock()
	service.isSyncing = false
	service.lastSyncTime = time.Now()
	service.syncMutex.Unlock()

	isSyncing, lastSync = service.GetSyncStatus()
	assert.False(t, isSyncing)
	assert.False(t, lastSync.IsZero())
}

func TestClimbTrackingService_ConcurrentSyncPrevention(t *testing.T) {
	mockRepo := &database.MockRepository{
		GetAllRouteIDsForLocationFn: func(ctx context.Context, locationID int) ([]string, error) {
			// Simulate slow operation
			time.Sleep(100 * time.Millisecond)
			return []string{}, nil
		},
	}
	mockMPClient := &MockMPClient{}

	service := NewClimbTrackingService(mockRepo, mockMPClient)

	// Start first sync
	go func() {
		_ = service.SyncNewTicksForLocation(context.Background(), 1)
	}()

	// Wait a bit to ensure first sync has started
	time.Sleep(10 * time.Millisecond)

	// Try to start second sync - should fail
	err := service.SyncNewTicksForLocation(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sync already in progress")
}

func TestClimbTrackingService_DateParsing(t *testing.T) {
	tests := []struct {
		name       string
		dateFormat string
		shouldSave bool
	}{
		{
			name:       "format 1 - Jan 2, 2006, 3:04 pm",
			dateFormat: "Jan 2, 2006, 3:04 pm",
			shouldSave: true,
		},
		{
			name:       "format 2 - 2006-01-02 15:04:05",
			dateFormat: "2006-01-02 15:04:05",
			shouldSave: true,
		},
		{
			name:       "format 3 - 2006-01-02",
			dateFormat: "2006-01-02",
			shouldSave: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			savedCount := 0
			testTime := time.Now().Add(-1 * time.Hour)

			mockRepo := &database.MockRepository{
				GetAllRouteIDsForLocationFn: func(ctx context.Context, locationID int) ([]string, error) {
					return []string{"123"}, nil
				},
				GetLastTickTimestampForRouteFn: func(ctx context.Context, routeID string) (*time.Time, error) {
					oldTime := testTime.Add(-48 * time.Hour)
					return &oldTime, nil
				},
				SaveMPTickFn: func(ctx context.Context, tick *models.MPTick) error {
					savedCount++
					return nil
				},
			}

			mockMPClient := &MockMPClient{
				GetRouteTicksFn: func(routeID string) ([]mountainproject.Tick, error) {
					return []mountainproject.Tick{
						createTickWithUser(testTime.Format(tt.dateFormat), "TestUser", "Send"),
					}, nil
				},
			}

			service := NewClimbTrackingService(mockRepo, mockMPClient)
			err := service.SyncNewTicksForLocation(context.Background(), 1)

			assert.NoError(t, err)
			if tt.shouldSave {
				assert.Equal(t, 1, savedCount, "Expected tick to be saved for format: %s", tt.dateFormat)
			}
		})
	}
}

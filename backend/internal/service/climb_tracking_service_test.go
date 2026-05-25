package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/alexscott64/woulder/backend/internal/monitoring"
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
	GetRouteTicksFn    func(routeID string) ([]mountainproject.Tick, error)
	GetRouteFn         func(routeID string) (*mountainproject.RouteResponse, error)
	GetAreaFn          func(areaID string) (*mountainproject.AreaResponse, error)
	GetAreaCommentsFn  func(areaID string) ([]mountainproject.Comment, error)
	GetRouteCommentsFn func(routeID string) ([]mountainproject.Comment, error)
}

func (m *MockMPClient) GetRouteTicks(routeID string) ([]mountainproject.Tick, error) {
	if m.GetRouteTicksFn != nil {
		return m.GetRouteTicksFn(routeID)
	}
	return nil, nil
}

func (m *MockMPClient) GetRoute(routeID string) (*mountainproject.RouteResponse, error) {
	if m.GetRouteFn != nil {
		return m.GetRouteFn(routeID)
	}
	return nil, nil
}

func (m *MockMPClient) GetArea(areaID string) (*mountainproject.AreaResponse, error) {
	if m.GetAreaFn != nil {
		return m.GetAreaFn(areaID)
	}
	return nil, nil
}

func (m *MockMPClient) GetAreaComments(areaID string) ([]mountainproject.Comment, error) {
	if m.GetAreaCommentsFn != nil {
		return m.GetAreaCommentsFn(areaID)
	}
	return nil, nil
}

func (m *MockMPClient) GetRouteComments(routeID string) ([]mountainproject.Comment, error) {
	if m.GetRouteCommentsFn != nil {
		return m.GetRouteCommentsFn(routeID)
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
						MPRouteID:      123,
						RouteName:      "Test Route",
						RouteRating:    "V5",
						MPAreaID:       456,
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
			mockMPRepo := NewMockMountainProjectRepository()
			mockClimbingRepo := NewMockClimbingRepository()
			mockClimbingRepo.history.GetClimbHistoryForLocationFn = tt.mockFn

			service := NewClimbTrackingService(mockMPRepo, mockClimbingRepo, &MockMPClient{}, nil)
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
		mockRouteIDs   []int64
		mockLastTick   *time.Time
		mockMPTicks    []mountainproject.Tick
		mockGetRoutes  func(ctx context.Context, locationID int) ([]int64, error)
		mockGetLastFn  func(ctx context.Context, routeID int64) (*time.Time, error)
		mockSaveTickFn func(ctx context.Context, tick *models.MPTick) error
		wantErr        bool
		expectedSaved  int
	}{
		{
			name:         "success - new ticks only",
			locationID:   1,
			mockRouteIDs: []int64{123},
			mockLastTick: &oldTick,
			mockMPTicks: []mountainproject.Tick{
				createTickWithUser(newTick.Format("Jan 2, 2006, 3:04 pm"), "TestUser", "Flash"),
			},
			mockGetRoutes: func(ctx context.Context, locationID int) ([]int64, error) {
				return []int64{123}, nil
			},
			mockGetLastFn: func(ctx context.Context, routeID int64) (*time.Time, error) {
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
			mockRouteIDs: []int64{123},
			mockLastTick: &newTick,
			mockMPTicks: []mountainproject.Tick{
				createTickWithUser(oldTick.Format("Jan 2, 2006, 3:04 pm"), "TestUser", "Send"),
			},
			mockGetRoutes: func(ctx context.Context, locationID int) ([]int64, error) {
				return []int64{123}, nil
			},
			mockGetLastFn: func(ctx context.Context, routeID int64) (*time.Time, error) {
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
			mockRouteIDs: []int64{},
			mockGetRoutes: func(ctx context.Context, locationID int) ([]int64, error) {
				return []int64{}, nil
			},
			wantErr:       false,
			expectedSaved: 0,
		},
		{
			name:       "error getting routes",
			locationID: 1,
			mockGetRoutes: func(ctx context.Context, locationID int) ([]int64, error) {
				return nil, errors.New("database error")
			},
			wantErr:       true,
			expectedSaved: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			savedCount := 0

			mockMPRepo := NewMockMountainProjectRepository()
			mockMPRepo.routes.GetAllIDsForLocationFn = tt.mockGetRoutes
			mockMPRepo.ticks.GetLastTimestampForRouteFn = tt.mockGetLastFn
			mockMPRepo.ticks.SaveTickFn = func(ctx context.Context, tick *models.MPTick) error {
				savedCount++
				if tt.mockSaveTickFn != nil {
					return tt.mockSaveTickFn(ctx, tick)
				}
				return nil
			}

			mockClimbingRepo := NewMockClimbingRepository()

			mockMPClient := &MockMPClient{
				GetRouteTicksFn: func(routeID string) ([]mountainproject.Tick, error) {
					return tt.mockMPTicks, nil
				},
			}

			service := NewClimbTrackingService(mockMPRepo, mockClimbingRepo, mockMPClient, nil)
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
	mockMPRepo := NewMockMountainProjectRepository()
	mockClimbingRepo := NewMockClimbingRepository()
	mockMPClient := &MockMPClient{}

	service := NewClimbTrackingService(mockMPRepo, mockClimbingRepo, mockMPClient, nil)

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
	mockMPRepo := NewMockMountainProjectRepository()
	mockMPRepo.routes.GetAllIDsForLocationFn = func(ctx context.Context, locationID int) ([]int64, error) {
		// Simulate slow operation
		time.Sleep(100 * time.Millisecond)
		return []int64{}, nil
	}
	mockClimbingRepo := NewMockClimbingRepository()
	mockMPClient := &MockMPClient{}

	service := NewClimbTrackingService(mockMPRepo, mockClimbingRepo, mockMPClient, nil)

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

			mockMPRepo := NewMockMountainProjectRepository()
			mockMPRepo.routes.GetAllIDsForLocationFn = func(ctx context.Context, locationID int) ([]int64, error) {
				return []int64{123}, nil
			}
			mockMPRepo.ticks.GetLastTimestampForRouteFn = func(ctx context.Context, routeID int64) (*time.Time, error) {
				oldTime := testTime.Add(-48 * time.Hour)
				return &oldTime, nil
			}
			mockMPRepo.ticks.SaveTickFn = func(ctx context.Context, tick *models.MPTick) error {
				savedCount++
				return nil
			}

			mockClimbingRepo := NewMockClimbingRepository()

			mockMPClient := &MockMPClient{
				GetRouteTicksFn: func(routeID string) ([]mountainproject.Tick, error) {
					return []mountainproject.Tick{
						createTickWithUser(testTime.Format(tt.dateFormat), "TestUser", "Send"),
					}, nil
				},
			}

			service := NewClimbTrackingService(mockMPRepo, mockClimbingRepo, mockMPClient, nil)
			err := service.SyncNewTicksForLocation(context.Background(), 1)

			assert.NoError(t, err)
			if tt.shouldSave {
				assert.Equal(t, 1, savedCount, "Expected tick to be saved for format: %s", tt.dateFormat)
			}
		})
	}
}

// TestDateParsingWithTimezone verifies that Mountain Project dates are correctly parsed in Pacific timezone
func TestDateParsingWithTimezone(t *testing.T) {
	tests := []struct {
		name         string
		inputDate    string
		expectedHour int // Hour in UTC after conversion from Pacific
		expectedDay  int // Day in UTC (may differ due to timezone)
	}{
		{
			name:         "Afternoon Pacific time stays same day in UTC",
			inputDate:    "Jan 14, 2026, 3:04 pm",
			expectedHour: 23, // 3 PM PST = 11 PM UTC (PST is UTC-8)
			expectedDay:  14,
		},
		{
			name:         "Morning Pacific time stays same day in UTC",
			inputDate:    "Jan 14, 2026, 9:30 am",
			expectedHour: 17, // 9:30 AM PST = 5:30 PM UTC
			expectedDay:  14,
		},
		{
			name:         "Late evening Pacific crosses date boundary to next day UTC",
			inputDate:    "Jan 14, 2026, 11:00 pm",
			expectedHour: 7,  // 11 PM PST = 7 AM UTC next day
			expectedDay:  15, // Crosses to next day in UTC
		},
		{
			name:         "Early morning Pacific is previous day in UTC",
			inputDate:    "Jan 14, 2026, 12:30 am",
			expectedHour: 8, // 12:30 AM PST = 8:30 AM UTC
			expectedDay:  14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			savedTicks := []*models.MPTick{}

			mockMPRepo := NewMockMountainProjectRepository()
			mockMPRepo.routes.GetAllIDsForLocationFn = func(ctx context.Context, locationID int) ([]int64, error) {
				return []int64{123456789}, nil
			}
			mockMPRepo.ticks.GetLastTimestampForRouteFn = func(ctx context.Context, routeID int64) (*time.Time, error) {
				return nil, nil // No previous ticks
			}
			mockMPRepo.ticks.SaveTickFn = func(ctx context.Context, tick *models.MPTick) error {
				savedTicks = append(savedTicks, tick)
				return nil
			}

			mockClimbingRepo := NewMockClimbingRepository()

			mockMPClient := &MockMPClient{
				GetRouteTicksFn: func(routeID string) ([]mountainproject.Tick, error) {
					return []mountainproject.Tick{
						createTickWithUser(tt.inputDate, "TestUser", "Send"),
					}, nil
				},
			}

			service := NewClimbTrackingService(mockMPRepo, mockClimbingRepo, mockMPClient, nil)
			err := service.SyncNewTicksForLocation(context.Background(), 1)

			assert.NoError(t, err)
			assert.Equal(t, 1, len(savedTicks), "Expected 1 tick to be saved")

			if len(savedTicks) > 0 {
				savedTick := savedTicks[0]

				// Convert to UTC for comparison
				utcTime := savedTick.ClimbedAt.UTC()

				// Verify the hour matches expected UTC hour
				assert.Equal(t, tt.expectedHour, utcTime.Hour(),
					"Hour mismatch for input %s: expected %d UTC, got %d UTC (full time: %s)",
					tt.inputDate, tt.expectedHour, utcTime.Hour(), utcTime.Format(time.RFC3339))

				// Verify the day matches expected UTC day
				assert.Equal(t, tt.expectedDay, utcTime.Day(),
					"Day mismatch for input %s: expected day %d UTC, got day %d UTC (full time: %s)",
					tt.inputDate, tt.expectedDay, utcTime.Day(), utcTime.Format(time.RFC3339))

				// Verify the timezone is preserved (should be Pacific)
				zone, _ := savedTick.ClimbedAt.Zone()
				assert.Contains(t, []string{"PST", "PDT"}, zone,
					"Expected Pacific timezone (PST/PDT), got %s", zone)
			}
		})
	}
}

// TestTimezoneConsistencyBetweenSyncs verifies dates are consistent when parsed and compared
func TestTimezoneConsistencyBetweenSyncs(t *testing.T) {
	pacificTZ, err := time.LoadLocation("America/Los_Angeles")
	assert.NoError(t, err, "Failed to load Pacific timezone")

	// Jamie's climb on Jan 14, 2026 at 2:30 PM Pacific
	referenceTime := time.Date(2026, 1, 14, 14, 30, 0, 0, pacificTZ)

	tests := []struct {
		name          string
		tickDate      string
		shouldBeSaved bool
		description   string
	}{
		{
			name:          "Exact same time - should be skipped",
			tickDate:      "Jan 14, 2026, 2:30 pm",
			shouldBeSaved: false,
			description:   "Tick at exact reference time should not be saved (already have it)",
		},
		{
			name:          "Earlier time - should be skipped",
			tickDate:      "Jan 13, 2026, 10:00 am",
			shouldBeSaved: false,
			description:   "Tick before reference time should not be saved",
		},
		{
			name:          "Later time - should be saved",
			tickDate:      "Jan 15, 2026, 9:00 am",
			shouldBeSaved: true,
			description:   "Tick after reference time should be saved",
		},
		{
			name:          "One minute later - should be saved",
			tickDate:      "Jan 14, 2026, 2:31 pm",
			shouldBeSaved: true,
			description:   "Tick 1 minute after reference time should be saved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			savedTicks := []*models.MPTick{}

			mockMPRepo := NewMockMountainProjectRepository()
			mockMPRepo.routes.GetAllIDsForLocationFn = func(ctx context.Context, locationID int) ([]int64, error) {
				return []int64{100001}, nil
			}
			mockMPRepo.ticks.GetLastTimestampForRouteFn = func(ctx context.Context, routeID int64) (*time.Time, error) {
				return &referenceTime, nil
			}
			mockMPRepo.ticks.SaveTickFn = func(ctx context.Context, tick *models.MPTick) error {
				savedTicks = append(savedTicks, tick)
				return nil
			}

			mockClimbingRepo := NewMockClimbingRepository()

			mockMPClient := &MockMPClient{
				GetRouteTicksFn: func(routeID string) ([]mountainproject.Tick, error) {
					return []mountainproject.Tick{
						createTickWithUser(tt.tickDate, "TestUser", "Send"),
					}, nil
				},
			}

			service := NewClimbTrackingService(mockMPRepo, mockClimbingRepo, mockMPClient, nil)
			err := service.SyncNewTicksForLocation(context.Background(), 1)

			assert.NoError(t, err)

			if tt.shouldBeSaved {
				assert.Equal(t, 1, len(savedTicks), "%s: Expected tick to be saved", tt.description)
				if len(savedTicks) > 0 {
					assert.True(t, savedTicks[0].ClimbedAt.After(referenceTime),
						"%s: Saved tick should be after reference time", tt.description)
				}
			} else {
				assert.Equal(t, 0, len(savedTicks), "%s: Expected tick to be skipped", tt.description)
			}
		})
	}
}

// TestFutureDateFiltering verifies that ticks with future dates are rejected
func TestFutureDateFiltering(t *testing.T) {
	now := time.Now()
	futureDate := now.Add(7 * 24 * time.Hour) // 7 days in the future

	tests := []struct {
		name          string
		tickDate      string
		shouldBeSaved bool
		description   string
	}{
		{
			name:          "Past date - should be saved",
			tickDate:      now.Add(-24 * time.Hour).Format("Jan 2, 2006, 3:04 pm"),
			shouldBeSaved: true,
			description:   "Tick from yesterday should be saved",
		},
		{
			name:          "Current time - should be saved",
			tickDate:      now.Format("Jan 2, 2006, 3:04 pm"),
			shouldBeSaved: true,
			description:   "Tick from now should be saved",
		},
		{
			name:          "Within 24h buffer - should be saved",
			tickDate:      now.Add(12 * time.Hour).Format("Jan 2, 2006, 3:04 pm"),
			shouldBeSaved: true,
			description:   "Tick within 24h future buffer should be saved (clock skew)",
		},
		{
			name:          "7 days in future - should be rejected",
			tickDate:      futureDate.Format("Jan 2, 2006, 3:04 pm"),
			shouldBeSaved: false,
			description:   "Tick 7 days in future should be rejected",
		},
		{
			name:          "2 days in future - should be rejected",
			tickDate:      now.Add(48 * time.Hour).Format("Jan 2, 2006, 3:04 pm"),
			shouldBeSaved: false,
			description:   "Tick 2 days in future should be rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			savedTicks := []*models.MPTick{}

			mockMPRepo := NewMockMountainProjectRepository()
			mockMPRepo.routes.GetAllIDsForLocationFn = func(ctx context.Context, locationID int) ([]int64, error) {
				return []int64{123456789}, nil
			}
			mockMPRepo.ticks.GetLastTimestampForRouteFn = func(ctx context.Context, routeID int64) (*time.Time, error) {
				return nil, nil // No previous ticks
			}
			mockMPRepo.ticks.SaveTickFn = func(ctx context.Context, tick *models.MPTick) error {
				savedTicks = append(savedTicks, tick)
				return nil
			}

			mockClimbingRepo := NewMockClimbingRepository()

			mockMPClient := &MockMPClient{
				GetRouteTicksFn: func(routeID string) ([]mountainproject.Tick, error) {
					return []mountainproject.Tick{
						createTickWithUser(tt.tickDate, "TestUser", "Send"),
					}, nil
				},
			}

			service := NewClimbTrackingService(mockMPRepo, mockClimbingRepo, mockMPClient, nil)
			err := service.SyncNewTicksForLocation(context.Background(), 1)

			assert.NoError(t, err)

			if tt.shouldBeSaved {
				assert.Equal(t, 1, len(savedTicks), "%s: Expected tick to be saved", tt.description)
			} else {
				assert.Equal(t, 0, len(savedTicks), "%s: Expected tick to be rejected", tt.description)
			}
		})
	}
}

// TestCommentCleaning verifies that HTML entities are decoded and cleaned from comments
func TestCommentCleaning(t *testing.T) {
	tests := []struct {
		name            string
		inputText       string
		expectedComment *string
		description     string
	}{
		{
			name:            "HTML entity middot at start",
			inputText:       "&middot; Great climb!",
			expectedComment: strPtr("Great climb!"),
			description:     "Should decode and remove leading &middot;",
		},
		{
			name:            "Actual middot character at start",
			inputText:       "· Fun problem",
			expectedComment: strPtr("Fun problem"),
			description:     "Should remove leading middot character",
		},
		{
			name:            "Multiple HTML entities",
			inputText:       "&middot; Rock &amp; roll!",
			expectedComment: strPtr("Rock & roll!"),
			description:     "Should decode all HTML entities",
		},
		{
			name:            "Just middot - should be empty",
			inputText:       "&middot;",
			expectedComment: nil,
			description:     "Only middot should result in nil comment",
		},
		{
			name:            "Empty string",
			inputText:       "",
			expectedComment: nil,
			description:     "Empty string should result in nil comment",
		},
		{
			name:            "Normal comment without entities",
			inputText:       "Awesome send!",
			expectedComment: strPtr("Awesome send!"),
			description:     "Normal comment should be preserved",
		},
		{
			name:            "Middot with whitespace",
			inputText:       "&middot;  \t  Great route",
			expectedComment: strPtr("Great route"),
			description:     "Should remove middot and trim whitespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			savedTicks := []*models.MPTick{}

			mockMPRepo := NewMockMountainProjectRepository()
			mockMPRepo.routes.GetAllIDsForLocationFn = func(ctx context.Context, locationID int) ([]int64, error) {
				return []int64{123456789}, nil
			}
			mockMPRepo.ticks.GetLastTimestampForRouteFn = func(ctx context.Context, routeID int64) (*time.Time, error) {
				return nil, nil
			}
			mockMPRepo.ticks.SaveTickFn = func(ctx context.Context, tick *models.MPTick) error {
				savedTicks = append(savedTicks, tick)
				return nil
			}

			mockClimbingRepo := NewMockClimbingRepository()

			// Create tick with comment text
			testTime := time.Now().Add(-1 * time.Hour)
			userJSON, _ := json.Marshal(map[string]interface{}{
				"id":   123,
				"name": "TestUser",
			})
			textJSON, _ := json.Marshal(tt.inputText)

			mockMPClient := &MockMPClient{
				GetRouteTicksFn: func(routeID string) ([]mountainproject.Tick, error) {
					return []mountainproject.Tick{
						{
							Date:  testTime.Format("Jan 2, 2006, 3:04 pm"),
							Style: "Send",
							User:  userJSON,
							Text:  textJSON,
						},
					}, nil
				},
			}

			service := NewClimbTrackingService(mockMPRepo, mockClimbingRepo, mockMPClient, nil)
			err := service.SyncNewTicksForLocation(context.Background(), 1)

			assert.NoError(t, err)
			assert.Equal(t, 1, len(savedTicks), "%s: Expected tick to be saved", tt.description)

			if len(savedTicks) > 0 {
				tick := savedTicks[0]
				if tt.expectedComment == nil {
					assert.Nil(t, tick.Comment, "%s: Expected nil comment", tt.description)
				} else {
					assert.NotNil(t, tick.Comment, "%s: Expected non-nil comment", tt.description)
					if tick.Comment != nil {
						assert.Equal(t, *tt.expectedComment, *tick.Comment,
							"%s: Comment mismatch", tt.description)
					}
				}
			}
		})
	}
}

// Helper function to create string pointer
func strPtr(s string) *string {
	return &s
}

// TestJamieZeldaRailsScenario tests the exact scenario from the bug report
func TestJamieZeldaRailsScenario(t *testing.T) {
	// Jamie sent Zelda Rails on Jan 14, 2026
	// Mountain Project shows Jan 14th
	// But it was appearing as Jan 13th in the database

	savedTicks := []*models.MPTick{}

	mockMPRepo := NewMockMountainProjectRepository()
	mockMPRepo.routes.GetAllIDsForLocationFn = func(ctx context.Context, locationID int) ([]int64, error) {
		return []int64{999888777}, nil
	}
	mockMPRepo.ticks.GetLastTimestampForRouteFn = func(ctx context.Context, routeID int64) (*time.Time, error) {
		return nil, nil // No previous ticks
	}
	mockMPRepo.ticks.SaveTickFn = func(ctx context.Context, tick *models.MPTick) error {
		savedTicks = append(savedTicks, tick)
		return nil
	}

	mockClimbingRepo := NewMockClimbingRepository()

	mockMPClient := &MockMPClient{
		GetRouteTicksFn: func(routeID string) ([]mountainproject.Tick, error) {
			return []mountainproject.Tick{
				createTickWithUser("Jan 14, 2026, 2:30 pm", "Jamie", "Send"),
			}, nil
		},
	}

	service := NewClimbTrackingService(mockMPRepo, mockClimbingRepo, mockMPClient, nil)
	err := service.SyncNewTicksForLocation(context.Background(), 2) // Location 2 is Index, WA

	assert.NoError(t, err)
	assert.Equal(t, 1, len(savedTicks), "Expected Jamie's tick to be saved")

	if len(savedTicks) > 0 {
		tick := savedTicks[0]
		assert.Equal(t, "Jamie", tick.UserName)
		assert.Equal(t, "Send", tick.Style)

		// The critical assertion: when viewed in Pacific timezone, it should be Jan 14
		pacificTZ, err := time.LoadLocation("America/Los_Angeles")
		assert.NoError(t, err)

		tickInPacific := tick.ClimbedAt.In(pacificTZ)
		assert.Equal(t, 14, tickInPacific.Day(),
			"Jamie's tick should be on Jan 14 when viewed in Pacific timezone, got day %d (full time: %s)",
			tickInPacific.Day(), tickInPacific.Format(time.RFC3339))
		assert.Equal(t, time.January, tickInPacific.Month())
		assert.Equal(t, 2026, tickInPacific.Year())

		// Also verify that when the database returns this timestamp,
		// and we calculate days_since_climb, it should be based on the Pacific date
		t.Logf("Saved tick time: %s", tick.ClimbedAt.Format(time.RFC3339))
		t.Logf("Tick time in Pacific: %s", tickInPacific.Format(time.RFC3339))
	}
}

// ============================================================================
// SyncLocationAreaDiscovery tests
// ============================================================================

// mockAreaDiscoveryJobMonitor implements AreaDiscoveryJobMonitor and records
// the calls made against it. We use a local mock (rather than extending
// mocks_test.go) because *monitoring.JobMonitor is a concrete struct used by
// many other service methods directly; this narrow interface is only relevant
// to SyncLocationAreaDiscovery.
type mockAreaDiscoveryJobMonitor struct {
	StartJobCalls          []startJobCall
	CompleteJobCalls       []int64
	FailJobCalls           []failJobCall
	MarkJobPausedCalls     []int64
	UpdateProgressCalls    []updateProgressCall
	SaveCheckpointCalls    []saveCheckpointCall
	GetInterruptedJobCalls []string

	nextJobID           int64
	StartJobErr         error
	GetInterruptedJobFn func(ctx context.Context, jobName string) (*monitoring.JobExecution, error)
}

type startJobCall struct {
	JobName    string
	JobType    string
	TotalItems int
	Metadata   map[string]interface{}
}

type failJobCall struct {
	JobID    int64
	ErrorMsg string
}

type updateProgressCall struct {
	JobID          int64
	ItemsProcessed int
	Succeeded      int
	Failed         int
}

type saveCheckpointCall struct {
	JobID      int64
	Checkpoint map[string]interface{}
}

func (m *mockAreaDiscoveryJobMonitor) StartJob(ctx context.Context, jobName, jobType string, totalItems int, metadata map[string]interface{}) (*monitoring.JobExecution, error) {
	m.StartJobCalls = append(m.StartJobCalls, startJobCall{
		JobName:    jobName,
		JobType:    jobType,
		TotalItems: totalItems,
		Metadata:   metadata,
	})
	if m.StartJobErr != nil {
		return nil, m.StartJobErr
	}
	m.nextJobID++
	return &monitoring.JobExecution{
		ID:       m.nextJobID,
		JobName:  jobName,
		JobType:  jobType,
		Status:   monitoring.StatusRunning,
		Metadata: metadata,
	}, nil
}

func (m *mockAreaDiscoveryJobMonitor) CompleteJob(ctx context.Context, jobID int64) error {
	m.CompleteJobCalls = append(m.CompleteJobCalls, jobID)
	return nil
}

func (m *mockAreaDiscoveryJobMonitor) FailJob(ctx context.Context, jobID int64, errorMsg string) error {
	m.FailJobCalls = append(m.FailJobCalls, failJobCall{JobID: jobID, ErrorMsg: errorMsg})
	return nil
}

func (m *mockAreaDiscoveryJobMonitor) MarkJobPaused(ctx context.Context, jobID int64) error {
	m.MarkJobPausedCalls = append(m.MarkJobPausedCalls, jobID)
	return nil
}

func (m *mockAreaDiscoveryJobMonitor) UpdateProgress(ctx context.Context, jobID int64, itemsProcessed, succeeded, failed int) error {
	m.UpdateProgressCalls = append(m.UpdateProgressCalls, updateProgressCall{
		JobID: jobID, ItemsProcessed: itemsProcessed, Succeeded: succeeded, Failed: failed,
	})
	return nil
}

func (m *mockAreaDiscoveryJobMonitor) SaveCheckpoint(ctx context.Context, jobID int64, checkpoint map[string]interface{}) error {
	m.SaveCheckpointCalls = append(m.SaveCheckpointCalls, saveCheckpointCall{
		JobID: jobID, Checkpoint: checkpoint,
	})
	return nil
}

func (m *mockAreaDiscoveryJobMonitor) GetInterruptedJob(ctx context.Context, jobName string) (*monitoring.JobExecution, error) {
	m.GetInterruptedJobCalls = append(m.GetInterruptedJobCalls, jobName)
	if m.GetInterruptedJobFn != nil {
		return m.GetInterruptedJobFn(ctx, jobName)
	}
	return nil, nil
}

// TestSyncLocationAreaDiscovery_FantasiaBoulders is the regression test for
// the exact scenario that motivated this job's existence: a new MP sub-area
// (Fantasia Boulders, mp_area_id 202944793) is added underneath an existing
// Squamish parent (Apron Boulders, mp_area_id 106025685) and must be picked
// up automatically without manual seeding.
func TestSyncLocationAreaDiscovery_FantasiaBoulders(t *testing.T) {
	const (
		apronParentID    = int64(106025685)
		fantasiaChildID  = int64(202944793)
		syntheticRoute1  = 900000001
		syntheticRoute2  = 900000002
		squamishLocation = 6
	)

	// Configure LocationRoots to expose exactly one root for this test.
	restoreRoots := SetLocationRootsForTest([]LocationRootConfig{
		{
			LocationName: "Squamish",
			LocationID:   squamishLocation,
			MPAreaIDs:    []int64{apronParentID},
		},
	})
	defer restoreRoots()

	mockMPRepo := NewMockMountainProjectRepository()
	mockClimbingRepo := NewMockClimbingRepository()

	// Capture SaveArea calls so we can assert on them.
	var savedAreas []*models.MPArea
	mockMPRepo.areas.SaveAreaFn = func(ctx context.Context, area *models.MPArea) error {
		// Copy to avoid retaining a shared pointer.
		copy := *area
		savedAreas = append(savedAreas, &copy)
		return nil
	}

	// Capture SaveRoute calls (SyncAreaRecursive uses SaveRoute, not UpsertRoute).
	var savedRoutes []*models.MPRoute
	mockMPRepo.routes.SaveRouteFn = func(ctx context.Context, route *models.MPRoute) error {
		copy := *route
		savedRoutes = append(savedRoutes, &copy)
		return nil
	}

	// Mock MP API: parent area returns Fantasia as its only child; Fantasia
	// returns two synthetic route children with no further sub-areas.
	mockMPClient := &MockMPClient{
		GetAreaFn: func(areaID string) (*mountainproject.AreaResponse, error) {
			switch areaID {
			case "106025685":
				return &mountainproject.AreaResponse{
					ID:    int(apronParentID),
					Title: "Apron Boulders",
					Type:  "Area",
					Children: []mountainproject.ChildElement{
						{
							ID:    int(fantasiaChildID),
							Title: "Fantasia Boulders",
							Type:  "Area",
						},
					},
				}, nil
			case "202944793":
				return &mountainproject.AreaResponse{
					ID:    int(fantasiaChildID),
					Title: "Fantasia Boulders",
					Type:  "Area",
					Children: []mountainproject.ChildElement{
						{
							ID:         syntheticRoute1,
							Title:      "Fantasia Direct",
							Type:       "Route",
							RouteTypes: []string{"Boulder"},
						},
						{
							ID:         syntheticRoute2,
							Title:      "Fantasia Traverse",
							Type:       "Route",
							RouteTypes: []string{"Boulder"},
						},
					},
				}, nil
			}
			return nil, errors.New("unexpected area: " + areaID)
		},
		// GetRouteTicks / GetRouteComments return empty - keep the test focused.
		GetRouteTicksFn:    func(_ string) ([]mountainproject.Tick, error) { return nil, nil },
		GetRouteCommentsFn: func(_ string) ([]mountainproject.Comment, error) { return nil, nil },
		GetAreaCommentsFn:  func(_ string) ([]mountainproject.Comment, error) { return nil, nil },
	}

	service := NewClimbTrackingService(mockMPRepo, mockClimbingRepo, mockMPClient, nil)
	mockMonitor := &mockAreaDiscoveryJobMonitor{}
	restoreMonitor := service.SetAreaDiscoveryJobMonitorForTest(mockMonitor)
	defer restoreMonitor()

	err := service.SyncLocationAreaDiscovery(context.Background())
	assert.NoError(t, err, "discovery should succeed when no roots fail")

	// --- Assertion 1: Fantasia Boulders area was saved with correct parent + location.
	var fantasia *models.MPArea
	for _, a := range savedAreas {
		if a.MPAreaID == fantasiaChildID {
			fantasia = a
			break
		}
	}
	if assert.NotNil(t, fantasia, "Fantasia Boulders (mp_area_id=%d) must be discovered and saved", fantasiaChildID) {
		assert.Equal(t, "Fantasia Boulders", fantasia.Name)
		if assert.NotNil(t, fantasia.ParentMPAreaID, "parent_mp_area_id must be set") {
			assert.Equal(t, apronParentID, *fantasia.ParentMPAreaID,
				"Fantasia's parent must be Apron Boulders (mp_area_id=%d)", apronParentID)
		}
		if assert.NotNil(t, fantasia.LocationID, "location_id must be set") {
			assert.Equal(t, squamishLocation, *fantasia.LocationID,
				"Fantasia must be tagged with Squamish location_id=%d", squamishLocation)
		}
	}

	// --- Assertion 2: synthetic routes were saved with the same location_id.
	assert.GreaterOrEqual(t, len(savedRoutes), 2, "expected at least the two synthetic Fantasia routes to be saved")
	routeIDs := map[int64]*models.MPRoute{}
	for _, r := range savedRoutes {
		routeIDs[r.MPRouteID] = r
	}
	for _, id := range []int64{syntheticRoute1, syntheticRoute2} {
		r, ok := routeIDs[id]
		if assert.True(t, ok, "route %d should have been saved", id) {
			if assert.NotNil(t, r.LocationID, "route %d location_id must be set", id) {
				assert.Equal(t, squamishLocation, *r.LocationID,
					"route %d must be tagged with Squamish location_id", id)
			}
		}
	}

	// --- Assertion 3: job_monitor was driven with the right job_name + total.
	if assert.Len(t, mockMonitor.StartJobCalls, 1, "StartJob should be called exactly once") {
		call := mockMonitor.StartJobCalls[0]
		assert.Equal(t, "location_area_discovery", call.JobName, "job_name must match for jtrack pickup")
		assert.Equal(t, "location_area_discovery", call.JobType, "job_type must match")
		assert.Equal(t, 1, call.TotalItems, "total = 1 root configured in this test")
	}

	// --- Assertion 4: CompleteJob fired with success (no FailJob).
	assert.Len(t, mockMonitor.CompleteJobCalls, 1, "CompleteJob should fire once on success")
	assert.Empty(t, mockMonitor.FailJobCalls, "FailJob should not fire when all roots succeed")
}

// TestSyncLocationAreaDiscovery_ErrorIsolation verifies that a failure on
// one root does not prevent subsequent roots from being processed, and that
// the job is marked as failed (not silently completed) when any root errors.
func TestSyncLocationAreaDiscovery_ErrorIsolation(t *testing.T) {
	const (
		failingRootID = int64(111111111)
		workingRootID = int64(222222222)
	)

	restoreRoots := SetLocationRootsForTest([]LocationRootConfig{
		{LocationName: "Failing", LocationID: 100, MPAreaIDs: []int64{failingRootID}},
		{LocationName: "Working", LocationID: 101, MPAreaIDs: []int64{workingRootID}},
	})
	defer restoreRoots()

	mockMPRepo := NewMockMountainProjectRepository()
	mockClimbingRepo := NewMockClimbingRepository()

	var savedAreas []int64
	mockMPRepo.areas.SaveAreaFn = func(ctx context.Context, area *models.MPArea) error {
		savedAreas = append(savedAreas, area.MPAreaID)
		return nil
	}

	getAreaCalls := map[string]int{}
	mockMPClient := &MockMPClient{
		GetAreaFn: func(areaID string) (*mountainproject.AreaResponse, error) {
			getAreaCalls[areaID]++
			switch areaID {
			case "111111111":
				return nil, errors.New("simulated MP API failure")
			case "222222222":
				return &mountainproject.AreaResponse{
					ID:       int(workingRootID),
					Title:    "Working Root",
					Type:     "Area",
					Children: nil,
				}, nil
			}
			return nil, errors.New("unexpected area: " + areaID)
		},
	}

	service := NewClimbTrackingService(mockMPRepo, mockClimbingRepo, mockMPClient, nil)
	mockMonitor := &mockAreaDiscoveryJobMonitor{}
	restoreMonitor := service.SetAreaDiscoveryJobMonitorForTest(mockMonitor)
	defer restoreMonitor()

	err := service.SyncLocationAreaDiscovery(context.Background())
	assert.Error(t, err, "an error should be returned when any root fails")

	// Both roots were attempted (error isolation - second still runs).
	assert.GreaterOrEqual(t, getAreaCalls["111111111"], 1, "failing root must have been attempted")
	assert.GreaterOrEqual(t, getAreaCalls["222222222"], 1, "second root must still be attempted after the first fails")

	// Working root's area was saved (proves second root executed end-to-end).
	assert.Contains(t, savedAreas, workingRootID,
		"working root's area must be saved even though the first root failed")

	// Job ended in a non-success state.
	assert.Empty(t, mockMonitor.CompleteJobCalls,
		"CompleteJob should NOT fire when any root failed")
	assert.Len(t, mockMonitor.FailJobCalls, 1,
		"FailJob should fire exactly once to surface the partial-failure outcome")

	// Total items reported as 2 (one per root).
	if assert.Len(t, mockMonitor.StartJobCalls, 1) {
		assert.Equal(t, 2, mockMonitor.StartJobCalls[0].TotalItems)
	}
}

// TestSyncLocationAreaDiscovery_ContextCancellation verifies the loop exits
// cleanly between roots when the context is cancelled, without panicking,
// and marks the job as paused (mirroring SyncLocationRouteTicks' behavior).
func TestSyncLocationAreaDiscovery_ContextCancellation(t *testing.T) {
	const (
		firstRootID  = int64(333333333)
		secondRootID = int64(444444444)
	)

	restoreRoots := SetLocationRootsForTest([]LocationRootConfig{
		{LocationName: "First", LocationID: 200, MPAreaIDs: []int64{firstRootID}},
		{LocationName: "Second", LocationID: 201, MPAreaIDs: []int64{secondRootID}},
	})
	defer restoreRoots()

	mockMPRepo := NewMockMountainProjectRepository()
	mockClimbingRepo := NewMockClimbingRepository()

	ctx, cancel := context.WithCancel(context.Background())

	getAreaCalls := map[string]int{}
	mockMPClient := &MockMPClient{
		GetAreaFn: func(areaID string) (*mountainproject.AreaResponse, error) {
			getAreaCalls[areaID]++
			// After the first root finishes, cancel the context so the
			// loop exits before processing the second root.
			if areaID == "333333333" {
				cancel()
				return &mountainproject.AreaResponse{
					ID: int(firstRootID), Title: "First", Type: "Area",
				}, nil
			}
			return &mountainproject.AreaResponse{
				ID: int(secondRootID), Title: "Second", Type: "Area",
			}, nil
		},
	}

	service := NewClimbTrackingService(mockMPRepo, mockClimbingRepo, mockMPClient, nil)
	mockMonitor := &mockAreaDiscoveryJobMonitor{}
	restoreMonitor := service.SetAreaDiscoveryJobMonitorForTest(mockMonitor)
	defer restoreMonitor()

	// Should not panic.
	err := service.SyncLocationAreaDiscovery(ctx)
	assert.ErrorIs(t, err, context.Canceled, "cancellation should propagate as context.Canceled")

	// Second root should NOT have been processed.
	assert.Equal(t, 1, getAreaCalls["333333333"], "first root processed")
	assert.Equal(t, 0, getAreaCalls["444444444"], "second root must be skipped after cancellation")

	// Job marked paused (not completed/failed) so it can resume on next boot.
	assert.Empty(t, mockMonitor.CompleteJobCalls, "CompleteJob must not fire on cancellation")
	assert.Empty(t, mockMonitor.FailJobCalls, "FailJob must not fire on cancellation")
	assert.Len(t, mockMonitor.MarkJobPausedCalls, 1,
		"MarkJobPaused should fire exactly once so the next boot resumes the run")
}

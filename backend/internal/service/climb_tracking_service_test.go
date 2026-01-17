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
			expectedHour: 7, // 11 PM PST = 7 AM UTC next day
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

			mockRepo := &database.MockRepository{
				GetAllRouteIDsForLocationFn: func(ctx context.Context, locationID int) ([]string, error) {
					return []string{"test-route-123"}, nil
				},
				GetLastTickTimestampForRouteFn: func(ctx context.Context, routeID string) (*time.Time, error) {
					return nil, nil // No previous ticks
				},
				SaveMPTickFn: func(ctx context.Context, tick *models.MPTick) error {
					savedTicks = append(savedTicks, tick)
					return nil
				},
			}

			mockMPClient := &MockMPClient{
				GetRouteTicksFn: func(routeID string) ([]mountainproject.Tick, error) {
					return []mountainproject.Tick{
						createTickWithUser(tt.inputDate, "TestUser", "Send"),
					}, nil
				},
			}

			service := NewClimbTrackingService(mockRepo, mockMPClient)
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

			mockRepo := &database.MockRepository{
				GetAllRouteIDsForLocationFn: func(ctx context.Context, locationID int) ([]string, error) {
					return []string{"route-1"}, nil
				},
				GetLastTickTimestampForRouteFn: func(ctx context.Context, routeID string) (*time.Time, error) {
					return &referenceTime, nil
				},
				SaveMPTickFn: func(ctx context.Context, tick *models.MPTick) error {
					savedTicks = append(savedTicks, tick)
					return nil
				},
			}

			mockMPClient := &MockMPClient{
				GetRouteTicksFn: func(routeID string) ([]mountainproject.Tick, error) {
					return []mountainproject.Tick{
						createTickWithUser(tt.tickDate, "TestUser", "Send"),
					}, nil
				},
			}

			service := NewClimbTrackingService(mockRepo, mockMPClient)
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

// TestJamieZeldaRailsScenario tests the exact scenario from the bug report
func TestJamieZeldaRailsScenario(t *testing.T) {
	// Jamie sent Zelda Rails on Jan 14, 2026
	// Mountain Project shows Jan 14th
	// But it was appearing as Jan 13th in the database

	savedTicks := []*models.MPTick{}

	mockRepo := &database.MockRepository{
		GetAllRouteIDsForLocationFn: func(ctx context.Context, locationID int) ([]string, error) {
			return []string{"zelda-rails-route"}, nil
		},
		GetLastTickTimestampForRouteFn: func(ctx context.Context, routeID string) (*time.Time, error) {
			return nil, nil // No previous ticks
		},
		SaveMPTickFn: func(ctx context.Context, tick *models.MPTick) error {
			savedTicks = append(savedTicks, tick)
			return nil
		},
	}

	mockMPClient := &MockMPClient{
		GetRouteTicksFn: func(routeID string) ([]mountainproject.Tick, error) {
			return []mountainproject.Tick{
				createTickWithUser("Jan 14, 2026, 2:30 pm", "Jamie", "Send"),
			}, nil
		},
	}

	service := NewClimbTrackingService(mockRepo, mockMPClient)
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

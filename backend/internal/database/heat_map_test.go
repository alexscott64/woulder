package database

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetAreaActivityDetail_UniqueClimbersWithNullUsernames tests that unique_climbers
// counts correctly even when some ticks have NULL user_name values
func TestGetAreaActivityDetail_UniqueClimbersWithNullUsernames(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}

	db, err := New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()

	// Test with a known area that has activity
	// Using a recent date range
	endDate := time.Now()
	startDate := endDate.AddDate(0, -3, 0) // 3 months ago

	// Query for any area with activity
	query := `
		SELECT r.mp_area_id
		FROM woulder.mp_routes r
		JOIN woulder.mp_ticks t ON r.mp_route_id = t.mp_route_id
		WHERE t.climbed_at >= $1 AND t.climbed_at <= $2
		GROUP BY r.mp_area_id
		HAVING COUNT(t.id) >= 5
		LIMIT 1
	`

	var areaID int64
	err = db.conn.QueryRowContext(ctx, query, startDate, endDate).Scan(&areaID)
	if err != nil {
		t.Skip("No areas with sufficient activity found for test")
	}

	// Get area activity detail
	detail, err := db.GetAreaActivityDetail(ctx, areaID, startDate, endDate)
	require.NoError(t, err)
	require.NotNil(t, detail)

	// Verify that unique_climbers is calculated correctly
	// It should not be 0 if there are ticks (unless all user_names are NULL)
	if detail.TotalTicks > 0 {
		// Query to check if there are any non-NULL user_names
		var nonNullUserCount int
		err = db.conn.QueryRowContext(ctx, `
			SELECT COUNT(DISTINCT t.user_name)
			FROM woulder.mp_ticks t
			JOIN woulder.mp_routes r ON t.mp_route_id = r.mp_route_id
			WHERE r.mp_area_id = $1
				AND t.climbed_at >= $2
				AND t.climbed_at <= $3
				AND t.user_name IS NOT NULL
		`, areaID, startDate, endDate).Scan(&nonNullUserCount)
		require.NoError(t, err)

		// The unique_climbers should match the count of non-NULL distinct user_names
		assert.Equal(t, nonNullUserCount, detail.UniqueClimbers,
			"unique_climbers should count only non-NULL user_names")
	}

	// Basic sanity checks
	assert.Greater(t, detail.TotalTicks, 0, "should have ticks")
	assert.Greater(t, detail.ActiveRoutes, 0, "should have active routes")
	assert.NotEmpty(t, detail.Name, "area should have a name")
}

// TestGetAreaActivityDetail_NullFieldHandling tests that NULL fields in ticks
// are properly converted to empty strings
func TestGetAreaActivityDetail_NullFieldHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}

	db, err := New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()

	// Find an area with recent activity
	endDate := time.Now()
	startDate := endDate.AddDate(0, -1, 0) // 1 month ago

	query := `
		SELECT r.mp_area_id
		FROM woulder.mp_routes r
		JOIN woulder.mp_ticks t ON r.mp_route_id = t.mp_route_id
		WHERE t.climbed_at >= $1 AND t.climbed_at <= $2
		GROUP BY r.mp_area_id
		HAVING COUNT(t.id) >= 1
		LIMIT 1
	`

	var areaID int64
	err = db.conn.QueryRowContext(ctx, query, startDate, endDate).Scan(&areaID)
	if err != nil {
		t.Skip("No areas with sufficient activity found for test")
	}

	// Get area activity detail
	detail, err := db.GetAreaActivityDetail(ctx, areaID, startDate, endDate)
	require.NoError(t, err)
	require.NotNil(t, detail)

	// Verify that recent ticks have no nil string fields
	// (NULL values should be converted to empty strings)
	for _, tick := range detail.RecentTicks {
		// These assertions ensure no panic from nil pointer dereference
		assert.NotNil(t, tick.UserName, "user_name should not be nil")
		assert.NotNil(t, tick.Rating, "rating should not be nil")
		assert.NotNil(t, tick.Style, "style should not be nil")
		assert.NotNil(t, tick.Comment, "comment should not be nil")
	}
}

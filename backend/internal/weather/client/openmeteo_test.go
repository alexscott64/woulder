package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseTimestampUTC(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantYear  int
		wantMonth int
		wantDay   int
		wantHour  int
		wantErr   bool
	}{
		{
			name:      "bare timestamp treated as UTC",
			input:     "2026-03-14T15:00",
			wantYear:  2026,
			wantMonth: 3,
			wantDay:   14,
			wantHour:  15,
			wantErr:   false,
		},
		{
			name:      "midnight UTC stays midnight UTC",
			input:     "2026-03-11T00:00",
			wantYear:  2026,
			wantMonth: 3,
			wantDay:   11,
			wantHour:  0,
			wantErr:   false,
		},
		{
			name:      "RFC3339 timestamp with Z suffix",
			input:     "2026-03-14T15:00:00Z",
			wantYear:  2026,
			wantMonth: 3,
			wantDay:   14,
			wantHour:  15,
			wantErr:   false,
		},
		{
			name:      "RFC3339 timestamp with offset",
			input:     "2026-03-14T08:00:00-07:00",
			wantYear:  2026,
			wantMonth: 3,
			wantDay:   14,
			wantHour:  15, // 8am Pacific = 3pm UTC
			wantErr:   false,
		},
		{
			name:    "invalid format",
			input:   "not-a-timestamp",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTimestampUTC(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseTimestampUTC(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseTimestampUTC(%q) unexpected error: %v", tt.input, err)
			}

			if got.Year() != tt.wantYear {
				t.Errorf("Year = %d, want %d", got.Year(), tt.wantYear)
			}
			if int(got.Month()) != tt.wantMonth {
				t.Errorf("Month = %d, want %d", got.Month(), tt.wantMonth)
			}
			if got.Day() != tt.wantDay {
				t.Errorf("Day = %d, want %d", got.Day(), tt.wantDay)
			}
			if got.Hour() != tt.wantHour {
				t.Errorf("Hour = %d, want %d", got.Hour(), tt.wantHour)
			}
			if got.Location().String() != "UTC" {
				t.Errorf("Location = %q, want UTC", got.Location().String())
			}
		})
	}
}

func TestParseSunTimestamp(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantHour int // Expected UTC hour
		wantErr  bool
	}{
		{
			name:     "bare sunrise timestamp as UTC",
			input:    "2026-01-15T07:30",
			wantHour: 7, // Already UTC from API (timezone=UTC)
			wantErr:  false,
		},
		{
			name:     "bare sunset timestamp as UTC",
			input:    "2026-06-15T20:45",
			wantHour: 20, // Already UTC from API (timezone=UTC)
			wantErr:  false,
		},
		{
			name:    "invalid format",
			input:   "not-a-timestamp",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSunTimestamp(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseSunTimestamp(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseSunTimestamp(%q) unexpected error: %v", tt.input, err)
			}

			if got.Hour() != tt.wantHour {
				t.Errorf("Hour = %d, want %d (input: %s, parsed UTC: %s)", got.Hour(), tt.wantHour, tt.input, got.String())
			}
			if got.Location().String() != "UTC" {
				t.Errorf("Location = %q, want UTC", got.Location().String())
			}
		})
	}
}

func TestIsNightTime(t *testing.T) {
	tests := []struct {
		name    string
		time    string
		sunrise string
		sunset  string
		want    bool
	}{
		{
			name:    "midday is not night",
			time:    "2026-03-14T12:00",
			sunrise: "2026-03-14T06:30",
			sunset:  "2026-03-14T18:30",
			want:    false,
		},
		{
			name:    "before sunrise is night",
			time:    "2026-03-14T05:00",
			sunrise: "2026-03-14T06:30",
			sunset:  "2026-03-14T18:30",
			want:    true,
		},
		{
			name:    "after sunset is night",
			time:    "2026-03-14T20:00",
			sunrise: "2026-03-14T06:30",
			sunset:  "2026-03-14T18:30",
			want:    true,
		},
		{
			name:    "at sunset is night",
			time:    "2026-03-14T18:30",
			sunrise: "2026-03-14T06:30",
			sunset:  "2026-03-14T18:30",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isNightTime(tt.time, tt.sunrise, tt.sunset)
			if got != tt.want {
				t.Errorf("isNightTime(%q, %q, %q) = %v, want %v", tt.time, tt.sunrise, tt.sunset, got, tt.want)
			}
		})
	}
}

// TestGetCurrentAndForecast_RejectsTruncatedResponse asserts that the
// Open-Meteo client returns an error containing the truncated-response
// sentinel prefix when the upstream returns HTTP 200 with a hourly array
// shorter than expectedMinForecastHours. This is the in-client guard that
// complements the service-layer length check (Fix 1).
//
// Bug context: in production, Open-Meteo intermittently returned 69-359
// hourly entries instead of the expected ~396, and the service silently
// accepted these stubs and overwrote the cache. The client now refuses
// short responses so callers can preserve the existing cache.
func TestGetCurrentAndForecast_RejectsTruncatedResponse(t *testing.T) {
	// Build a fake Open-Meteo response with only 100 hourly rows — well
	// below the 336-hour threshold but above the existing len==0 check.
	const truncatedHours = 100

	hourlyTimes := make([]string, truncatedHours)
	hourlyTemp := make([]float64, truncatedHours)
	hourlyHum := make([]int, truncatedHours)
	hourlyPrecip := make([]float64, truncatedHours)
	hourlyRain := make([]float64, truncatedHours)
	hourlySnow := make([]float64, truncatedHours)
	hourlyCloud := make([]int, truncatedHours)
	hourlyWind := make([]float64, truncatedHours)
	hourlyWindDir := make([]int, truncatedHours)
	hourlyCode := make([]int, truncatedHours)
	hourlyApp := make([]float64, truncatedHours)
	hourlyPress := make([]float64, truncatedHours)
	hourlySW := make([]float64, truncatedHours)
	hourlyDir := make([]float64, truncatedHours)
	hourlyDiff := make([]float64, truncatedHours)
	hourlyDew := make([]float64, truncatedHours)

	for i := 0; i < truncatedHours; i++ {
		hourlyTimes[i] = fmt.Sprintf("2026-03-14T%02d:00", i%24)
		hourlyTemp[i] = 50.0
		hourlyHum[i] = 60
		hourlyPrecip[i] = 0.0
		hourlyRain[i] = 0.0
		hourlySnow[i] = 0.0
		hourlyCloud[i] = 50
		hourlyWind[i] = 5.0
		hourlyWindDir[i] = 180
		hourlyCode[i] = 1
		hourlyApp[i] = 48.0
		hourlyPress[i] = 1013.0
	}

	resp := map[string]interface{}{
		"current": map[string]interface{}{
			"time":                 "2026-03-14T12:00",
			"temperature_2m":       50.0,
			"relative_humidity_2m": 60,
			"precipitation":        0.0,
			"rain":                 0.0,
			"snowfall":             0.0,
			"cloud_cover":          50,
			"wind_speed_10m":       5.0,
			"wind_direction_10m":   180,
			"weather_code":         1,
			"apparent_temperature": 48.0,
			"surface_pressure":     1013.0,
			"shortwave_radiation":  0.0,
			"direct_radiation":     0.0,
			"diffuse_radiation":    0.0,
			"dew_point_2m":         40.0,
		},
		"hourly": map[string]interface{}{
			"time":                 hourlyTimes,
			"temperature_2m":       hourlyTemp,
			"relative_humidity_2m": hourlyHum,
			"precipitation":        hourlyPrecip,
			"rain":                 hourlyRain,
			"snowfall":             hourlySnow,
			"cloud_cover":          hourlyCloud,
			"wind_speed_10m":       hourlyWind,
			"wind_direction_10m":   hourlyWindDir,
			"weather_code":         hourlyCode,
			"apparent_temperature": hourlyApp,
			"surface_pressure":     hourlyPress,
			"shortwave_radiation":  hourlySW,
			"direct_radiation":     hourlyDir,
			"diffuse_radiation":    hourlyDiff,
			"dew_point_2m":         hourlyDew,
		},
		"daily": map[string]interface{}{
			"time":    []string{"2026-03-14"},
			"sunrise": []string{"2026-03-14T07:00"},
			"sunset":  []string{"2026-03-14T19:00"},
		},
	}

	// Track number of HTTP calls so we can also assert the one-shot retry
	// behavior added by GetCurrentAndForecast for truncation errors.
	var callCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Redirect the client at the test server. Restore on exit so other tests
	// in this package are unaffected.
	originalURL := openMeteoForecastURL
	openMeteoForecastURL = server.URL
	defer func() { openMeteoForecastURL = originalURL }()

	client := NewOpenMeteoClient()
	current, forecast, _, err := client.GetCurrentAndForecast(47.0, -121.0)

	if err == nil {
		t.Fatalf("expected truncation error, got nil (current=%v, forecast_len=%d)", current, len(forecast))
	}
	if !strings.HasPrefix(err.Error(), truncatedResponseErrPrefix) {
		t.Errorf("expected error to start with %q, got: %v", truncatedResponseErrPrefix, err)
	}
	if current != nil || forecast != nil {
		t.Errorf("expected nil current/forecast on truncation, got current=%v forecast_len=%d", current, len(forecast))
	}
	// Retry classifier should have produced exactly one extra attempt
	// (initial + 1 retry = 2 total calls) before giving up.
	if callCount != 2 {
		t.Errorf("expected exactly 2 HTTP calls (1 + 1 retry) on persistent truncation, got %d", callCount)
	}
}

// TestIsRetryableTruncationErr verifies the truncation-error sentinel
// classifier used by the retry path. Direct white-box test on the helper
// to lock in the contract used by GetCurrentAndForecast.
func TestIsRetryableTruncationErr(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"unrelated error", fmt.Errorf("something else went wrong"), false},
		{"truncation error", fmt.Errorf("%s: got 100 hours, expected at least 336", truncatedResponseErrPrefix), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRetryableTruncationErr(tt.err); got != tt.want {
				t.Errorf("isRetryableTruncationErr(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

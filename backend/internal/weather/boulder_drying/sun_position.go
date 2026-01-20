package boulder_drying

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SunPositionClient fetches sun position data from IP Geolocation Astronomy API
type SunPositionClient struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// SunPositionData represents hourly sun position (azimuth, elevation)
type SunPositionData struct {
	Timestamp time.Time `json:"timestamp"` // Hour timestamp
	Azimuth   float64   `json:"azimuth"`   // Sun direction (0° = North, 90° = East, 180° = South, 270° = West)
	Elevation float64   `json:"elevation"` // Sun angle above horizon (0° = horizon, 90° = zenith)
}

// AstronomyAPIResponse represents the response from IP Geolocation Astronomy API
type AstronomyAPIResponse struct {
	Location struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"location"`
	Date          string `json:"date"`
	CurrentTime   string `json:"current_time"`
	Sunrise       string `json:"sunrise"`
	Sunset        string `json:"sunset"`
	SolarNoon     string `json:"solar_noon"`
	DayLength     string `json:"day_length"`
	SunAltitude   float64 `json:"sun_altitude"`
	SunDistance   float64 `json:"sun_distance"`
	SunAzimuth    float64 `json:"sun_azimuth"`
	Moonrise      string `json:"moonrise"`
	Moonset       string `json:"moonset"`
	MoonAltitude  float64 `json:"moon_altitude"`
	MoonDistance  float64 `json:"moon_distance"`
	MoonAzimuth   float64 `json:"moon_azimuth"`
	MoonParallax  float64 `json:"moon_parallax"`
}

// NewSunPositionClient creates a new sun position API client
func NewSunPositionClient(apiKey string) *SunPositionClient {
	return &SunPositionClient{
		apiKey:  apiKey,
		baseURL: "https://api.ipgeolocation.io/astronomy",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetSunPositionForecast fetches hourly sun position data for the next N days
// Returns array of hourly sun positions (azimuth, elevation)
func (c *SunPositionClient) GetSunPositionForecast(
	ctx context.Context,
	lat, lon float64,
	days int,
) ([]SunPositionData, error) {
	var allData []SunPositionData

	// Fetch sun data for each day
	for i := 0; i < days; i++ {
		date := time.Now().AddDate(0, 0, i)
		dayData, err := c.getSunDataForDay(ctx, lat, lon, date)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch sun data for day %d: %w", i, err)
		}
		allData = append(allData, dayData...)
	}

	return allData, nil
}

// getSunDataForDay fetches sun position data for a specific day
// Estimates hourly sun positions based on sunrise, sunset, and solar noon
func (c *SunPositionClient) getSunDataForDay(
	ctx context.Context,
	lat, lon float64,
	date time.Time,
) ([]SunPositionData, error) {
	// Build API request URL
	url := fmt.Sprintf("%s?apiKey=%s&lat=%.6f&long=%.6f&date=%s",
		c.baseURL,
		c.apiKey,
		lat,
		lon,
		date.Format("2006-01-02"),
	)

	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sun data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResp AstronomyAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Generate hourly sun position data
	return c.generateHourlySunPositions(date, apiResp)
}

// generateHourlySunPositions estimates hourly sun positions based on daily data
// Uses sunrise, sunset, solar noon to calculate approximate hourly positions
func (c *SunPositionClient) generateHourlySunPositions(
	date time.Time,
	apiResp AstronomyAPIResponse,
) ([]SunPositionData, error) {
	// Parse sunrise and sunset times
	sunrise, err := parseTimeOfDay(date, apiResp.Sunrise)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sunrise: %w", err)
	}

	sunset, err := parseTimeOfDay(date, apiResp.Sunset)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sunset: %w", err)
	}

	solarNoon, err := parseTimeOfDay(date, apiResp.SolarNoon)
	if err != nil {
		return nil, fmt.Errorf("failed to parse solar noon: %w", err)
	}

	// Generate hourly data points
	var hourlyData []SunPositionData

	// Before sunrise: sun below horizon
	for hour := 0; hour < sunrise.Hour(); hour++ {
		hourlyData = append(hourlyData, SunPositionData{
			Timestamp: time.Date(date.Year(), date.Month(), date.Day(), hour, 0, 0, 0, date.Location()),
			Azimuth:   90.0,  // East (arbitrary, sun not visible)
			Elevation: -10.0, // Below horizon
		})
	}

	// Daylight hours: sun above horizon
	daylightHours := int(sunset.Sub(sunrise).Hours()) + 1
	for i := 0; i < daylightHours; i++ {
		hour := sunrise.Hour() + i
		if hour >= 24 {
			break
		}

		timestamp := time.Date(date.Year(), date.Month(), date.Day(), hour, 0, 0, 0, date.Location())

		// Calculate azimuth (approximate)
		// Sunrise: ~90° (East), Solar Noon: ~180° (South), Sunset: ~270° (West)
		var azimuth float64
		if timestamp.Before(solarNoon) {
			// Morning: interpolate from 90° (East) to 180° (South)
			progress := timestamp.Sub(sunrise).Hours() / solarNoon.Sub(sunrise).Hours()
			azimuth = 90.0 + (90.0 * progress)
		} else {
			// Afternoon: interpolate from 180° (South) to 270° (West)
			progress := timestamp.Sub(solarNoon).Hours() / sunset.Sub(solarNoon).Hours()
			azimuth = 180.0 + (90.0 * progress)
		}

		// Calculate elevation (approximate parabolic arc)
		// Peak at solar noon, 0° at sunrise/sunset
		hoursSinceSunrise := timestamp.Sub(sunrise).Hours()
		totalDaylight := sunset.Sub(sunrise).Hours()
		normalizedTime := (hoursSinceSunrise / totalDaylight) * 2.0 - 1.0 // Range [-1, 1]
		elevation := 90.0 * (1.0 - normalizedTime*normalizedTime) * 0.6 // Max ~54° elevation (typical for mid-latitudes)

		hourlyData = append(hourlyData, SunPositionData{
			Timestamp: timestamp,
			Azimuth:   azimuth,
			Elevation: elevation,
		})
	}

	// After sunset: sun below horizon
	for hour := sunset.Hour() + 1; hour < 24; hour++ {
		hourlyData = append(hourlyData, SunPositionData{
			Timestamp: time.Date(date.Year(), date.Month(), date.Day(), hour, 0, 0, 0, date.Location()),
			Azimuth:   270.0, // West (arbitrary, sun not visible)
			Elevation: -10.0, // Below horizon
		})
	}

	return hourlyData, nil
}

// parseTimeOfDay converts "HH:MM" string to time.Time on the given date
func parseTimeOfDay(date time.Time, timeStr string) (time.Time, error) {
	// Parse "HH:MM" format
	var hour, minute int
	_, err := fmt.Sscanf(timeStr, "%d:%d", &hour, &minute)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time format '%s': %w", timeStr, err)
	}

	return time.Date(date.Year(), date.Month(), date.Day(), hour, minute, 0, 0, date.Location()), nil
}

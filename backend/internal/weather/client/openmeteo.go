package client

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// openMeteoForecastURL / openMeteoHistoricalURL are vars (not consts) so tests
// can point them at httptest servers without exporting an injection seam. They
// are not modified at runtime in production code paths.
var (
	openMeteoForecastURL   = "https://api.open-meteo.com/v1/forecast"
	openMeteoHistoricalURL = "https://api.open-meteo.com/v1/archive"
)

const (
	maxRetries        = 3
	initialRetryDelay = 1 * time.Second

	// expectedMinForecastHours is the lower bound for `hourly.time` length on
	// the GetCurrentAndForecast endpoint (which requests
	// forecast_days=16&past_hours=12 ≈ 396 hours). We tolerate ~60 hours of
	// upstream slack and only reject responses with fewer than 14 days × 24h
	// of data — this matches the service-layer threshold and addresses the
	// observed bug where Open-Meteo intermittently returned 69-359 hours.
	expectedMinForecastHours = 14 * 24 // 336 hours
)

// errOpenMeteoTruncated is returned by the client (and recognized by the retry
// loop) when Open-Meteo responds with HTTP 200 but a hourly array shorter than
// expectedMinForecastHours. Using a sentinel-style prefix lets retryableGet
// classify the error without coupling to fmt.Errorf-wrapped formatting.
const truncatedResponseErrPrefix = "open-meteo response truncated"

// OpenMeteoClient handles API calls to Open-Meteo
type OpenMeteoClient struct {
	httpClient *http.Client
}

// Open-Meteo API response structures
// Uses default model only (no multi-model) for consistent, accurate forecasts.
type openMeteoResponse struct {
	Current *struct {
		Time                string  `json:"time"`
		Temperature2m       float64 `json:"temperature_2m"`
		RelativeHumidity2m  int     `json:"relative_humidity_2m"`
		Precipitation       float64 `json:"precipitation"`
		Rain                float64 `json:"rain"`
		Snowfall            float64 `json:"snowfall"`
		CloudCover          int     `json:"cloud_cover"`
		WindSpeed10m        float64 `json:"wind_speed_10m"`
		WindDirection10m    int     `json:"wind_direction_10m"`
		WeatherCode         int     `json:"weather_code"`
		ApparentTemperature float64 `json:"apparent_temperature"`
		Pressure            float64 `json:"surface_pressure"`
		ShortwaveRadiation  float64 `json:"shortwave_radiation"`
		DirectRadiation     float64 `json:"direct_radiation"`
		DiffuseRadiation    float64 `json:"diffuse_radiation"`
		Dewpoint2m          float64 `json:"dew_point_2m"`
	} `json:"current"`
	Hourly struct {
		Time                []string  `json:"time"`
		Temperature2m       []float64 `json:"temperature_2m"`
		RelativeHumidity2m  []int     `json:"relative_humidity_2m"`
		Precipitation       []float64 `json:"precipitation"`
		Rain                []float64 `json:"rain"`
		Snowfall            []float64 `json:"snowfall"`
		CloudCover          []int     `json:"cloud_cover"`
		WindSpeed10m        []float64 `json:"wind_speed_10m"`
		WindDirection10m    []int     `json:"wind_direction_10m"`
		WeatherCode         []int     `json:"weather_code"`
		ApparentTemperature []float64 `json:"apparent_temperature"`
		Pressure            []float64 `json:"surface_pressure"`
		ShortwaveRadiation  []float64 `json:"shortwave_radiation"`
		DirectRadiation     []float64 `json:"direct_radiation"`
		DiffuseRadiation    []float64 `json:"diffuse_radiation"`
		Dewpoint2m          []float64 `json:"dew_point_2m"`
	} `json:"hourly"`
	Daily *struct {
		Time    []string `json:"time"`
		Sunrise []string `json:"sunrise"`
		Sunset  []string `json:"sunset"`
	} `json:"daily"`
}

// DailySunTime represents sunrise/sunset for a single day
type DailySunTime struct {
	Date    string
	Sunrise string
	Sunset  string
}

// SunTimes contains sunrise and sunset information
type SunTimes struct {
	Sunrise string         // Today's sunrise
	Sunset  string         // Today's sunset
	Daily   []DailySunTime // All days' sunrise/sunset
}

// NewOpenMeteoClient creates a new Open-Meteo API client
func NewOpenMeteoClient() *OpenMeteoClient {
	return &OpenMeteoClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SetForecastBaseURLForTest overrides the base Open-Meteo forecast URL used by
// all OpenMeteoClient instances created in this process. It returns a restore
// function the test should defer to put the original URL back. This is the
// single supported test seam for redirecting clients at httptest servers; it
// is not safe for concurrent use across parallel tests.
func SetForecastBaseURLForTest(url string) (restore func()) {
	original := openMeteoForecastURL
	openMeteoForecastURL = url
	return func() { openMeteoForecastURL = original }
}

// retryableGet performs an HTTP GET with retry logic for rate limiting and transient errors
func (c *OpenMeteoClient) retryableGet(url string) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Calculate exponential backoff: 1s, 2s, 4s
			delay := initialRetryDelay * time.Duration(1<<uint(attempt-1))
			log.Printf("Retry attempt %d/%d for Open-Meteo after %v", attempt, maxRetries, delay)
			time.Sleep(delay)
		}

		resp, err := c.httpClient.Get(url)
		if err != nil {
			lastErr = err
			// Network errors are retryable
			log.Printf("Open-Meteo request failed (attempt %d/%d): %v", attempt+1, maxRetries+1, err)
			continue
		}

		// Check for rate limiting or server errors
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("Open-Meteo API error (status %d): %s", resp.StatusCode, string(body))
			log.Printf("Open-Meteo returned %d (attempt %d/%d): %s", resp.StatusCode, attempt+1, maxRetries+1, string(body))

			// For 429, check if Retry-After header is present
			if resp.StatusCode == http.StatusTooManyRequests {
				if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
					log.Printf("Open-Meteo rate limited. Retry-After: %s", retryAfter)
				}
			}
			continue
		}

		// Success or non-retryable error
		return resp, nil
	}

	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// isRetryableTruncationErr reports whether an error from a higher-level
// fetch call (e.g. GetCurrentAndForecast) represents a truncated upstream
// response that is worth retrying once. This is checked at the public API
// method level rather than inside retryableGet because truncation is a
// JSON-payload condition, not an HTTP-status one.
func isRetryableTruncationErr(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return len(s) >= len(truncatedResponseErrPrefix) && s[:len(truncatedResponseErrPrefix)] == truncatedResponseErrPrefix
}

// parseTimestampUTC parses a timestamp string and returns it as UTC.
// All API calls use timezone=UTC so bare timestamps (without timezone info)
// are interpreted as UTC directly.
func parseTimestampUTC(timeStr string) (time.Time, error) {
	// Try RFC3339 first (includes timezone)
	timestamp, err := time.Parse(time.RFC3339, timeStr)
	if err == nil {
		return timestamp.UTC(), nil
	}

	// Parse bare timestamp (e.g., "2025-12-30T16:00") as UTC.
	// All Open-Meteo API calls in this client use timezone=UTC,
	// so bare timestamps are always in UTC.
	timestamp, err = time.Parse("2006-01-02T15:04", timeStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse timestamp '%s': %w", timeStr, err)
	}

	// Interpret directly as UTC - no timezone conversion needed
	return timestamp.UTC(), nil
}

// parseSunTimestamp parses sunrise/sunset timestamps from the daily endpoint.
// These are returned in UTC (since we request timezone=UTC) and need to be
// converted to RFC3339 format so the frontend can properly interpret the timezone.
func parseSunTimestamp(timeStr string) (time.Time, error) {
	// Try RFC3339 first (already has timezone info)
	timestamp, err := time.Parse(time.RFC3339, timeStr)
	if err == nil {
		return timestamp.UTC(), nil
	}

	// Parse bare timestamp as UTC (Open-Meteo returns UTC when timezone=UTC is set)
	timestamp, err = time.Parse("2006-01-02T15:04", timeStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse sun timestamp '%s': %w", timeStr, err)
	}

	// Interpret as UTC since the API was called with timezone=UTC
	return time.Date(
		timestamp.Year(), timestamp.Month(), timestamp.Day(),
		timestamp.Hour(), timestamp.Minute(), timestamp.Second(), 0,
		time.UTC,
	), nil
}

// formatSunTimestampUTC takes a bare timestamp from Open-Meteo (e.g., "2026-04-10T13:25")
// returned in UTC and converts it to RFC3339 format (e.g., "2026-04-10T13:25:00Z")
// so the frontend can correctly interpret the timezone and display in local time.
func formatSunTimestampUTC(timeStr string) string {
	parsed, err := parseSunTimestamp(timeStr)
	if err != nil {
		return timeStr // Return original on parse failure
	}
	return parsed.Format(time.RFC3339)
}

// GetCurrentWeather fetches current weather with both current conditions and hourly forecast
// Uses default Open-Meteo model for accurate, consistent data.
func (c *OpenMeteoClient) GetCurrentWeather(lat, lon float64) (*models.WeatherData, error) {
	url := fmt.Sprintf("%s?latitude=%.8f&longitude=%.8f&current=temperature_2m,relative_humidity_2m,precipitation,rain,snowfall,cloud_cover,wind_speed_10m,wind_direction_10m,weather_code,apparent_temperature,surface_pressure,shortwave_radiation,direct_radiation,diffuse_radiation,dew_point_2m&hourly=precipitation,rain,snowfall&temperature_unit=fahrenheit&wind_speed_unit=mph&precipitation_unit=inch&timezone=UTC&forecast_days=1",
		openMeteoForecastURL, lat, lon)

	resp, err := c.retryableGet(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch current weather from Open-Meteo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Open-Meteo API error (status %d): %s", resp.StatusCode, string(body))
	}

	var data openMeteoResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode Open-Meteo response: %w", err)
	}

	if data.Current == nil {
		return nil, fmt.Errorf("no current weather data returned from Open-Meteo")
	}

	if len(data.Hourly.Time) == 0 {
		return nil, fmt.Errorf("no hourly data returned from Open-Meteo")
	}

	precipitation := data.Hourly.Precipitation
	if len(precipitation) == 0 {
		return nil, fmt.Errorf("no precipitation data returned from Open-Meteo")
	}

	// Parse current weather timestamp as UTC
	timestamp, err := parseTimestampUTC(data.Current.Time)
	if err != nil {
		return nil, err
	}

	weather := &models.WeatherData{
		Timestamp:          timestamp,
		Temperature:        data.Current.Temperature2m,
		FeelsLike:          data.Current.ApparentTemperature,
		Precipitation:      precipitation[0],
		Humidity:           data.Current.RelativeHumidity2m,
		WindSpeed:          data.Current.WindSpeed10m,
		WindDirection:      data.Current.WindDirection10m,
		CloudCover:         data.Current.CloudCover,
		Pressure:           int(data.Current.Pressure),
		Description:        getWeatherDescription(data.Current.WeatherCode),
		Icon:               getWeatherIcon(data.Current.WeatherCode),
		ShortwaveRadiation: data.Current.ShortwaveRadiation,
		DirectRadiation:    data.Current.DirectRadiation,
		DiffuseRadiation:   data.Current.DiffuseRadiation,
		DewpointF:          data.Current.Dewpoint2m,
	}

	return weather, nil
}

// GetCurrentAndForecast fetches both current weather and forecast in a single API call.
// Uses default Open-Meteo model with timezone=UTC for consistent timestamp handling.
// Daily sunrise/sunset is also returned in UTC and converted to RFC3339 for proper frontend display.
//
// On a truncated upstream response (see expectedMinForecastHours), the call is
// retried once with a short backoff. After one failed retry the truncation
// error is returned to the caller, which in the service layer triggers the
// length-validation guard and preserves the existing cache.
func (c *OpenMeteoClient) GetCurrentAndForecast(lat, lon float64) (*models.WeatherData, []models.WeatherData, *SunTimes, error) {
	current, forecast, sunTimes, err := c.getCurrentAndForecastOnce(lat, lon)
	if err != nil && isRetryableTruncationErr(err) {
		log.Printf("Open-Meteo returned truncated forecast for (%.5f,%.5f); retrying once: %v", lat, lon, err)
		time.Sleep(initialRetryDelay)
		current, forecast, sunTimes, err = c.getCurrentAndForecastOnce(lat, lon)
	}
	return current, forecast, sunTimes, err
}

// getCurrentAndForecastOnce performs a single Open-Meteo fetch and parse.
// It is the workhorse called by GetCurrentAndForecast (which adds one-shot
// retry on truncated responses).
func (c *OpenMeteoClient) getCurrentAndForecastOnce(lat, lon float64) (*models.WeatherData, []models.WeatherData, *SunTimes, error) {
	// All data (hourly, current, daily) uses timezone=UTC for consistent timestamp handling.
	// Sunrise/sunset timestamps are converted to RFC3339 with Z suffix so the frontend
	// can correctly interpret them as UTC and display in the user's local timezone.
	url := fmt.Sprintf("%s?latitude=%.8f&longitude=%.8f&current=temperature_2m,relative_humidity_2m,cloud_cover,wind_speed_10m,wind_direction_10m,weather_code,apparent_temperature,surface_pressure,shortwave_radiation,direct_radiation,diffuse_radiation,dew_point_2m&hourly=temperature_2m,relative_humidity_2m,precipitation,rain,snowfall,cloud_cover,wind_speed_10m,wind_direction_10m,weather_code,apparent_temperature,surface_pressure,shortwave_radiation,direct_radiation,diffuse_radiation,dew_point_2m&daily=sunrise,sunset&temperature_unit=fahrenheit&wind_speed_unit=mph&precipitation_unit=inch&timezone=UTC&forecast_days=16&past_hours=12",
		openMeteoForecastURL, lat, lon)

	resp, err := c.retryableGet(url)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to fetch weather from Open-Meteo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, nil, nil, fmt.Errorf("Open-Meteo API error (status %d): %s", resp.StatusCode, string(body))
	}

	var data openMeteoResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode Open-Meteo response: %w", err)
	}

	if data.Current == nil {
		return nil, nil, nil, fmt.Errorf("no current weather data returned from Open-Meteo")
	}

	if len(data.Hourly.Time) == 0 {
		return nil, nil, nil, fmt.Errorf("no hourly data returned from Open-Meteo")
	}

	// FIX: Open-Meteo intermittently returns HTTP 200 with a truncated hourly
	// array (observed: 69-359 hours instead of the expected ~396). Reject
	// short responses here so the caller can preserve its existing cache
	// rather than overwriting it with a stub. The error prefix is used by
	// isRetryableTruncationErr() / GetCurrentAndForecastWithRetry() to drive
	// at most one extra attempt.
	if len(data.Hourly.Time) < expectedMinForecastHours {
		return nil, nil, nil, fmt.Errorf("%s: got %d hours, expected at least %d",
			truncatedResponseErrPrefix, len(data.Hourly.Time), expectedMinForecastHours)
	}

	precipitation := data.Hourly.Precipitation
	if len(precipitation) == 0 {
		return nil, nil, nil, fmt.Errorf("no precipitation data returned from Open-Meteo")
	}

	// Extract sunrise/sunset for all days
	var sunTimes *SunTimes
	if data.Daily != nil && len(data.Daily.Sunrise) > 0 && len(data.Daily.Sunset) > 0 {
		sunTimes = &SunTimes{
			Sunrise: formatSunTimestampUTC(data.Daily.Sunrise[0]),
			Sunset:  formatSunTimestampUTC(data.Daily.Sunset[0]),
		}
		for i := 0; i < len(data.Daily.Time) && i < len(data.Daily.Sunrise) && i < len(data.Daily.Sunset); i++ {
			sunTimes.Daily = append(sunTimes.Daily, DailySunTime{
				Date:    data.Daily.Time[i],
				Sunrise: formatSunTimestampUTC(data.Daily.Sunrise[i]),
				Sunset:  formatSunTimestampUTC(data.Daily.Sunset[i]),
			})
		}
	}

	// Parse current weather timestamp as UTC
	timestamp, err := parseTimestampUTC(data.Current.Time)
	if err != nil {
		return nil, nil, nil, err
	}

	// Determine if it's currently night time for the icon
	isNight := false
	if sunTimes != nil {
		isNight = isNightTime(data.Current.Time, sunTimes.Sunrise, sunTimes.Sunset)
	}

	// Create current weather data
	current := &models.WeatherData{
		Timestamp:          timestamp,
		Temperature:        data.Current.Temperature2m,
		FeelsLike:          data.Current.ApparentTemperature,
		Precipitation:      precipitation[0],
		Humidity:           data.Current.RelativeHumidity2m,
		WindSpeed:          data.Current.WindSpeed10m,
		WindDirection:      data.Current.WindDirection10m,
		CloudCover:         data.Current.CloudCover,
		Pressure:           int(data.Current.Pressure),
		Description:        getWeatherDescription(data.Current.WeatherCode),
		Icon:               getWeatherIconWithTime(data.Current.WeatherCode, isNight),
		ShortwaveRadiation: data.Current.ShortwaveRadiation,
		DirectRadiation:    data.Current.DirectRadiation,
		DiffuseRadiation:   data.Current.DiffuseRadiation,
		DewpointF:          data.Current.Dewpoint2m,
	}

	// Parse forecast data (all hourly data)
	var forecast []models.WeatherData
	for i := 0; i < len(data.Hourly.Time); i++ {
		if i >= len(data.Hourly.Temperature2m) || i >= len(data.Hourly.ApparentTemperature) ||
			i >= len(precipitation) || i >= len(data.Hourly.RelativeHumidity2m) ||
			i >= len(data.Hourly.WindSpeed10m) || i >= len(data.Hourly.WindDirection10m) ||
			i >= len(data.Hourly.CloudCover) || i >= len(data.Hourly.Pressure) ||
			i >= len(data.Hourly.WeatherCode) ||
			i >= len(data.Hourly.ShortwaveRadiation) || i >= len(data.Hourly.DirectRadiation) ||
			i >= len(data.Hourly.DiffuseRadiation) || i >= len(data.Hourly.Dewpoint2m) {
			log.Printf("Skipping hour %d due to incomplete data arrays", i)
			continue
		}

		ts, err := parseTimestampUTC(data.Hourly.Time[i])
		if err != nil {
			log.Printf("Failed to parse hourly timestamp '%s': %v", data.Hourly.Time[i], err)
			continue
		}

		// Determine day/night for this forecast hour
		hourIsNight := false
		if sunTimes != nil {
			hourIsNight = isNightTimeForForecast(data.Hourly.Time[i], data.Daily)
		}

		weather := models.WeatherData{
			Timestamp:          ts,
			Temperature:        data.Hourly.Temperature2m[i],
			FeelsLike:          data.Hourly.ApparentTemperature[i],
			Precipitation:      precipitation[i],
			Humidity:           data.Hourly.RelativeHumidity2m[i],
			WindSpeed:          data.Hourly.WindSpeed10m[i],
			WindDirection:      data.Hourly.WindDirection10m[i],
			CloudCover:         data.Hourly.CloudCover[i],
			Pressure:           int(data.Hourly.Pressure[i]),
			Description:        getWeatherDescription(data.Hourly.WeatherCode[i]),
			Icon:               getWeatherIconWithTime(data.Hourly.WeatherCode[i], hourIsNight),
			ShortwaveRadiation: data.Hourly.ShortwaveRadiation[i],
			DirectRadiation:    data.Hourly.DirectRadiation[i],
			DiffuseRadiation:   data.Hourly.DiffuseRadiation[i],
			DewpointF:          data.Hourly.Dewpoint2m[i],
		}

		forecast = append(forecast, weather)
	}

	return current, forecast, sunTimes, nil
}

// GetForecast fetches hourly forecast data.
// Uses default Open-Meteo model with timezone=UTC for consistent timestamp storage.
func (c *OpenMeteoClient) GetForecast(lat, lon float64) ([]models.WeatherData, error) {
	url := fmt.Sprintf("%s?latitude=%.8f&longitude=%.8f&hourly=temperature_2m,relative_humidity_2m,precipitation,rain,snowfall,cloud_cover,wind_speed_10m,wind_direction_10m,weather_code,apparent_temperature,surface_pressure,shortwave_radiation,direct_radiation,diffuse_radiation,dew_point_2m&temperature_unit=fahrenheit&wind_speed_unit=mph&precipitation_unit=inch&timezone=UTC&forecast_days=16",
		openMeteoForecastURL, lat, lon)

	resp, err := c.retryableGet(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch forecast from Open-Meteo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Open-Meteo API error (status %d): %s", resp.StatusCode, string(body))
	}

	var data openMeteoResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode Open-Meteo response: %w", err)
	}

	precipitation := data.Hourly.Precipitation
	if len(precipitation) == 0 {
		return nil, fmt.Errorf("no precipitation data returned from Open-Meteo")
	}

	var forecast []models.WeatherData
	for i := 0; i < len(data.Hourly.Time); i++ {
		if i >= len(data.Hourly.Temperature2m) || i >= len(data.Hourly.ApparentTemperature) ||
			i >= len(precipitation) || i >= len(data.Hourly.RelativeHumidity2m) ||
			i >= len(data.Hourly.WindSpeed10m) || i >= len(data.Hourly.WindDirection10m) ||
			i >= len(data.Hourly.CloudCover) || i >= len(data.Hourly.Pressure) ||
			i >= len(data.Hourly.WeatherCode) ||
			i >= len(data.Hourly.ShortwaveRadiation) || i >= len(data.Hourly.DirectRadiation) ||
			i >= len(data.Hourly.DiffuseRadiation) || i >= len(data.Hourly.Dewpoint2m) {
			log.Printf("Skipping hour %d due to incomplete data arrays", i)
			continue
		}

		timestamp, err := parseTimestampUTC(data.Hourly.Time[i])
		if err != nil {
			log.Printf("Failed to parse hourly timestamp '%s': %v", data.Hourly.Time[i], err)
			continue
		}

		weather := models.WeatherData{
			Timestamp:          timestamp,
			Temperature:        data.Hourly.Temperature2m[i],
			FeelsLike:          data.Hourly.ApparentTemperature[i],
			Precipitation:      precipitation[i],
			Humidity:           data.Hourly.RelativeHumidity2m[i],
			WindSpeed:          data.Hourly.WindSpeed10m[i],
			WindDirection:      data.Hourly.WindDirection10m[i],
			CloudCover:         data.Hourly.CloudCover[i],
			Pressure:           int(data.Hourly.Pressure[i]),
			Description:        getWeatherDescription(data.Hourly.WeatherCode[i]),
			Icon:               getWeatherIcon(data.Hourly.WeatherCode[i]),
			ShortwaveRadiation: data.Hourly.ShortwaveRadiation[i],
			DirectRadiation:    data.Hourly.DirectRadiation[i],
			DiffuseRadiation:   data.Hourly.DiffuseRadiation[i],
			DewpointF:          data.Hourly.Dewpoint2m[i],
		}

		forecast = append(forecast, weather)
	}

	return forecast, nil
}

// GetHistoricalWeather fetches recent historical weather data using forecast API with past_days.
// Uses default model which provides reanalysis/observed data for accurate historical precipitation.
func (c *OpenMeteoClient) GetHistoricalWeather(lat, lon float64, days int) ([]models.WeatherData, error) {
	url := fmt.Sprintf("%s?latitude=%.8f&longitude=%.8f&past_days=%d&forecast_days=1&hourly=temperature_2m,relative_humidity_2m,precipitation,rain,snowfall,cloud_cover,wind_speed_10m,wind_direction_10m,weather_code,apparent_temperature,surface_pressure,shortwave_radiation,direct_radiation,diffuse_radiation,dew_point_2m&temperature_unit=fahrenheit&wind_speed_unit=mph&precipitation_unit=inch&timezone=UTC",
		openMeteoForecastURL, lat, lon, days)

	resp, err := c.retryableGet(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch historical weather from Open-Meteo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Open-Meteo API error (status %d): %s", resp.StatusCode, string(body))
	}

	var data openMeteoResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode Open-Meteo response: %w", err)
	}

	precipitation := data.Hourly.Precipitation

	now := time.Now()
	var historical []models.WeatherData
	for i := range data.Hourly.Time {
		timestamp, err := parseTimestampUTC(data.Hourly.Time[i])
		if err != nil {
			log.Printf("Failed to parse historical timestamp '%s': %v", data.Hourly.Time[i], err)
			continue
		}

		// Only include past data (skip future timestamps)
		if timestamp.After(now) {
			continue
		}

		// Bounds check for all arrays
		if i >= len(data.Hourly.Temperature2m) || i >= len(data.Hourly.ApparentTemperature) ||
			i >= len(precipitation) || i >= len(data.Hourly.RelativeHumidity2m) ||
			i >= len(data.Hourly.WindSpeed10m) || i >= len(data.Hourly.WindDirection10m) ||
			i >= len(data.Hourly.CloudCover) || i >= len(data.Hourly.Pressure) ||
			i >= len(data.Hourly.WeatherCode) ||
			i >= len(data.Hourly.ShortwaveRadiation) || i >= len(data.Hourly.DirectRadiation) ||
			i >= len(data.Hourly.DiffuseRadiation) || i >= len(data.Hourly.Dewpoint2m) {
			log.Printf("Skipping historical hour %d due to incomplete data arrays", i)
			continue
		}

		weather := models.WeatherData{
			Timestamp:          timestamp,
			Temperature:        data.Hourly.Temperature2m[i],
			FeelsLike:          data.Hourly.ApparentTemperature[i],
			Precipitation:      precipitation[i],
			Humidity:           data.Hourly.RelativeHumidity2m[i],
			WindSpeed:          data.Hourly.WindSpeed10m[i],
			WindDirection:      data.Hourly.WindDirection10m[i],
			CloudCover:         data.Hourly.CloudCover[i],
			Pressure:           int(data.Hourly.Pressure[i]),
			Description:        getWeatherDescription(data.Hourly.WeatherCode[i]),
			Icon:               getWeatherIcon(data.Hourly.WeatherCode[i]),
			ShortwaveRadiation: data.Hourly.ShortwaveRadiation[i],
			DirectRadiation:    data.Hourly.DirectRadiation[i],
			DiffuseRadiation:   data.Hourly.DiffuseRadiation[i],
			DewpointF:          data.Hourly.Dewpoint2m[i],
		}

		historical = append(historical, weather)
	}

	log.Printf("Got %d historical data points for (%.4f, %.4f)", len(historical), lat, lon)
	return historical, nil
}

// Map WMO weather codes to descriptions
func getWeatherDescription(code int) string {
	descriptions := map[int]string{
		0:  "Clear sky",
		1:  "Mainly clear",
		2:  "Partly cloudy",
		3:  "Overcast",
		45: "Foggy",
		48: "Depositing rime fog",
		51: "Light drizzle",
		53: "Moderate drizzle",
		55: "Dense drizzle",
		61: "Slight rain",
		63: "Moderate rain",
		65: "Heavy rain",
		71: "Slight snow",
		73: "Moderate snow",
		75: "Heavy snow",
		77: "Snow grains",
		80: "Slight rain showers",
		81: "Moderate rain showers",
		82: "Violent rain showers",
		85: "Slight snow showers",
		86: "Heavy snow showers",
		95: "Thunderstorm",
		96: "Thunderstorm with slight hail",
		99: "Thunderstorm with heavy hail",
	}

	if desc, ok := descriptions[code]; ok {
		return desc
	}
	return "Unknown"
}

// isNightTime checks if the given time is before sunrise or after sunset
func isNightTime(timeStr, sunrise, sunset string) bool {
	// Parse time strings (format: "2025-12-27T16:15" or "2025-12-27T07:54")
	// Extract just the time portion for comparison
	getTimeMinutes := func(s string) int {
		// Find the T separator
		for i, c := range s {
			if c == 'T' {
				timepart := s[i+1:]
				// Parse HH:MM
				var hour, min int
				fmt.Sscanf(timepart, "%d:%d", &hour, &min)
				return hour*60 + min
			}
		}
		return 12 * 60 // Default to noon if parsing fails
	}

	currentMinutes := getTimeMinutes(timeStr)
	sunriseMinutes := getTimeMinutes(sunrise)
	sunsetMinutes := getTimeMinutes(sunset)

	// Night is before sunrise or after sunset
	return currentMinutes < sunriseMinutes || currentMinutes >= sunsetMinutes
}

// isNightTimeForForecast checks if a forecast hour is night time using the daily sunrise/sunset data
func isNightTimeForForecast(timeStr string, daily *struct {
	Time    []string `json:"time"`
	Sunrise []string `json:"sunrise"`
	Sunset  []string `json:"sunset"`
}) bool {
	if daily == nil || len(daily.Time) == 0 {
		return false
	}

	if len(daily.Sunrise) == 0 || len(daily.Sunset) == 0 {
		return false
	}

	// Extract date from timeStr (format: "2025-12-27T15:00")
	dateStr := ""
	for i, c := range timeStr {
		if c == 'T' {
			dateStr = timeStr[:i]
			break
		}
	}

	// Find matching day in daily data
	for i, day := range daily.Time {
		if day == dateStr && i < len(daily.Sunrise) && i < len(daily.Sunset) {
			return isNightTime(timeStr, daily.Sunrise[i], daily.Sunset[i])
		}
	}

	// If no match found, use first day's sunrise/sunset as approximation
	return isNightTime(timeStr, daily.Sunrise[0], daily.Sunset[0])
}

// Map WMO weather codes to OpenWeatherMap-like icon codes for consistency
func getWeatherIcon(code int) string {
	return getWeatherIconWithTime(code, false)
}

// getWeatherIconWithTime returns weather icon with day/night suffix
func getWeatherIconWithTime(code int, isNight bool) string {
	suffix := "d"
	if isNight {
		suffix = "n"
	}

	// Map to OpenWeatherMap icon codes to maintain frontend compatibility
	// Icons that change with day/night: 01 (clear), 02 (few clouds), 10 (rain)
	iconMap := map[int]string{
		0:  "01", // Clear sky
		1:  "02", // Mainly clear
		2:  "03", // Partly cloudy
		3:  "04", // Overcast
		45: "50", // Fog
		48: "50", // Rime fog
		51: "09", // Light drizzle
		53: "09", // Moderate drizzle
		55: "09", // Dense drizzle
		61: "10", // Slight rain
		63: "10", // Moderate rain
		65: "10", // Heavy rain
		71: "13", // Slight snow
		73: "13", // Moderate snow
		75: "13", // Heavy snow
		77: "13", // Snow grains
		80: "09", // Rain showers
		81: "09", // Moderate rain showers
		82: "09", // Violent rain showers
		85: "13", // Snow showers
		86: "13", // Heavy snow showers
		95: "11", // Thunderstorm
		96: "11", // Thunderstorm with hail
		99: "11", // Thunderstorm with heavy hail
	}

	if icon, ok := iconMap[code]; ok {
		return icon + suffix
	}
	return "01" + suffix // Default to clear sky
}

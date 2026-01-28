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

const (
	openMeteoForecastURL   = "https://api.open-meteo.com/v1/forecast"
	openMeteoHistoricalURL = "https://api.open-meteo.com/v1/archive"
)

// OpenMeteoClient handles API calls to Open-Meteo
type OpenMeteoClient struct {
	httpClient *http.Client
}

// Open-Meteo API response structures
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
		// Multi-model fields for precipitation data
		PrecipitationNAM []*float64 `json:"precipitation_ncep_nam_conus"`
		RainNAM          []*float64 `json:"rain_ncep_nam_conus"`
		SnowfallNAM      []*float64 `json:"snowfall_ncep_nam_conus"`
		PrecipitationBM  []float64  `json:"precipitation_best_match"`
		RainBM           []float64  `json:"rain_best_match"`
		SnowfallBM       []float64  `json:"snowfall_best_match"`
		// Multi-model fields for other weather data (used when multi-model request)
		Temperature2mBM       []float64 `json:"temperature_2m_best_match"`
		RelativeHumidity2mBM  []int     `json:"relative_humidity_2m_best_match"`
		CloudCoverBM          []int     `json:"cloud_cover_best_match"`
		WindSpeed10mBM        []float64 `json:"wind_speed_10m_best_match"`
		WindDirection10mBM    []int     `json:"wind_direction_10m_best_match"`
		WeatherCodeBM         []int     `json:"weather_code_best_match"`
		ApparentTemperatureBM []float64 `json:"apparent_temperature_best_match"`
		PressureBM            []float64 `json:"surface_pressure_best_match"`
	} `json:"hourly"`
	Daily *struct {
		Time              []string `json:"time"`
		Sunrise           []string `json:"sunrise"`
		Sunset            []string `json:"sunset"`
		SunriseNAM        []string `json:"sunrise_ncep_nam_conus"`
		SunsetNAM         []string `json:"sunset_ncep_nam_conus"`
		SunriseBestMatch  []string `json:"sunrise_best_match"`
		SunsetBestMatch   []string `json:"sunset_best_match"`
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

// mergePrecipitationData merges precipitation data from NCEP NAM CONUS (0-60h) and best_match (fallback)
// NAM CONUS provides more accurate precipitation forecasts but only for 60 hours
func mergePrecipitationData(namData []*float64, bmData []float64) []float64 {
	if len(bmData) == 0 {
		return []float64{}
	}

	result := make([]float64, len(bmData))
	for i := range bmData {
		// Use NAM data if available (not nil), otherwise fallback to best_match
		if i < len(namData) && namData[i] != nil {
			result[i] = *namData[i]
		} else {
			result[i] = bmData[i]
		}
	}
	return result
}

// getHourlyData returns the appropriate data arrays, preferring multi-model best_match fields
// when available, falling back to single-model fields
func (r *openMeteoResponse) getHourlyData() (temp, feelsLike []float64, humidity, cloudCover, windDir, weatherCode []int, windSpeed, pressure []float64) {
	// Prefer multi-model best_match fields if available
	if len(r.Hourly.Temperature2mBM) > 0 {
		return r.Hourly.Temperature2mBM, r.Hourly.ApparentTemperatureBM,
			r.Hourly.RelativeHumidity2mBM, r.Hourly.CloudCoverBM,
			r.Hourly.WindDirection10mBM, r.Hourly.WeatherCodeBM,
			r.Hourly.WindSpeed10mBM, r.Hourly.PressureBM
	}
	// Fallback to single-model fields
	return r.Hourly.Temperature2m, r.Hourly.ApparentTemperature,
		r.Hourly.RelativeHumidity2m, r.Hourly.CloudCover,
		r.Hourly.WindDirection10m, r.Hourly.WeatherCode,
		r.Hourly.WindSpeed10m, r.Hourly.Pressure
}

// parseTimestampUTC parses a timestamp and ensures it's in UTC
func parseTimestampUTC(timeStr string) (time.Time, error) {
	// Try RFC3339 first (includes timezone)
	timestamp, err := time.Parse(time.RFC3339, timeStr)
	if err == nil {
		return timestamp.UTC(), nil
	}

	// Try parsing without timezone (e.g., "2025-12-30T16:00")
	// Open-Meteo returns times in Pacific timezone when we request timezone=America/Los_Angeles
	timestamp, err = time.Parse("2006-01-02T15:04", timeStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse timestamp '%s': %w", timeStr, err)
	}

	// Load Pacific timezone
	pacificTZ, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to load Pacific timezone: %w", err)
	}

	// Interpret the timestamp as Pacific time, then convert to UTC
	timestampInPacific := time.Date(
		timestamp.Year(), timestamp.Month(), timestamp.Day(),
		timestamp.Hour(), timestamp.Minute(), timestamp.Second(), 0,
		pacificTZ,
	)

	return timestampInPacific.UTC(), nil
}

// GetCurrentWeather fetches current weather with both current conditions and hourly forecast
// Uses NCEP NAM CONUS model for precipitation data with best_match fallback
func (c *OpenMeteoClient) GetCurrentWeather(lat, lon float64) (*models.WeatherData, error) {
	// Fetch both current and hourly data to get accurate precipitation
	url := fmt.Sprintf("%s?latitude=%.8f&longitude=%.8f&current=temperature_2m,relative_humidity_2m,cloud_cover,wind_speed_10m,wind_direction_10m,weather_code,apparent_temperature,surface_pressure&hourly=precipitation,rain,snowfall&temperature_unit=fahrenheit&wind_speed_unit=mph&precipitation_unit=inch&timezone=UTC&forecast_days=1&models=ncep_nam_conus,best_match",
		openMeteoForecastURL, lat, lon)

	resp, err := c.httpClient.Get(url)
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

	// Merge precipitation data: use NAM CONUS if available, fallback to best_match
	precipitation := mergePrecipitationData(data.Hourly.PrecipitationNAM, data.Hourly.PrecipitationBM)

	// Fallback to single-model fields if multi-model not available
	if len(precipitation) == 0 {
		precipitation = data.Hourly.Precipitation
	}

	// Verify we have precipitation data
	if len(precipitation) == 0 {
		return nil, fmt.Errorf("no precipitation data returned from Open-Meteo")
	}

	// Parse current weather timestamp - ensure it's in UTC
	timestamp, err := parseTimestampUTC(data.Current.Time)
	if err != nil {
		return nil, err
	}

	// Use current conditions but get precipitation from the matching hourly forecast (merged data)
	weather := &models.WeatherData{
		Timestamp:     timestamp,
		Temperature:   data.Current.Temperature2m,
		FeelsLike:     data.Current.ApparentTemperature,
		Precipitation: precipitation[0],
		Humidity:      data.Current.RelativeHumidity2m,
		WindSpeed:     data.Current.WindSpeed10m,
		WindDirection: data.Current.WindDirection10m,
		CloudCover:    data.Current.CloudCover,
		Pressure:      int(data.Current.Pressure),
		Description:   getWeatherDescription(data.Current.WeatherCode),
		Icon:          getWeatherIcon(data.Current.WeatherCode),
	}

	return weather, nil
}

// GetCurrentAndForecast fetches both current weather and forecast in a single API call
// Uses NCEP NAM CONUS model for precipitation data (0-60h) with best_match fallback
func (c *OpenMeteoClient) GetCurrentAndForecast(lat, lon float64) (*models.WeatherData, []models.WeatherData, *SunTimes, error) {
	url := fmt.Sprintf("%s?latitude=%.8f&longitude=%.8f&current=temperature_2m,relative_humidity_2m,cloud_cover,wind_speed_10m,wind_direction_10m,weather_code,apparent_temperature,surface_pressure&hourly=temperature_2m,relative_humidity_2m,precipitation,rain,snowfall,cloud_cover,wind_speed_10m,wind_direction_10m,weather_code,apparent_temperature,surface_pressure&daily=sunrise,sunset&temperature_unit=fahrenheit&wind_speed_unit=mph&precipitation_unit=inch&timezone=America/Los_Angeles&forecast_days=16&models=ncep_nam_conus,best_match",
		openMeteoForecastURL, lat, lon)

	resp, err := c.httpClient.Get(url)
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

	// Merge precipitation data: use NAM CONUS for 0-60h, fallback to best_match
	precipitation := mergePrecipitationData(data.Hourly.PrecipitationNAM, data.Hourly.PrecipitationBM)

	// Fallback to single-model fields if multi-model not available
	if len(precipitation) == 0 {
		precipitation = data.Hourly.Precipitation
	}

	// Verify we have precipitation data
	if len(precipitation) == 0 {
		return nil, nil, nil, fmt.Errorf("no precipitation data returned from Open-Meteo")
	}

	// Get hourly data arrays (handles both multi-model and single-model responses)
	temp, feelsLike, humidity, cloudCover, windDir, weatherCode, windSpeed, pressure := data.getHourlyData()

	// Extract sunrise/sunset for all days
	// Use best_match model for sun times (more reliable than NAM CONUS)
	var sunTimes *SunTimes
	if data.Daily != nil {
		var sunriseData, sunsetData []string

		// Select which model's sun data to use (prefer best_match, fallback to NAM, then regular fields)
		if len(data.Daily.SunriseBestMatch) > 0 && len(data.Daily.SunsetBestMatch) > 0 {
			sunriseData = data.Daily.SunriseBestMatch
			sunsetData = data.Daily.SunsetBestMatch
		} else if len(data.Daily.SunriseNAM) > 0 && len(data.Daily.SunsetNAM) > 0 {
			sunriseData = data.Daily.SunriseNAM
			sunsetData = data.Daily.SunsetNAM
		} else if len(data.Daily.Sunrise) > 0 && len(data.Daily.Sunset) > 0 {
			sunriseData = data.Daily.Sunrise
			sunsetData = data.Daily.Sunset
		}

		if len(sunriseData) > 0 && len(sunsetData) > 0 {
			sunTimes = &SunTimes{
				Sunrise: sunriseData[0],
				Sunset:  sunsetData[0],
			}
			// Build daily sun times array
			for i := 0; i < len(data.Daily.Time) && i < len(sunriseData) && i < len(sunsetData); i++ {
				sunTimes.Daily = append(sunTimes.Daily, DailySunTime{
					Date:    data.Daily.Time[i],
					Sunrise: sunriseData[i],
					Sunset:  sunsetData[i],
				})
			}
		}
	}

	// Parse current weather timestamp - ensure it's in UTC
	timestamp, err := parseTimestampUTC(data.Current.Time)
	if err != nil {
		return nil, nil, nil, err
	}

	// Determine if it's currently night time for the icon
	isNight := false
	if sunTimes != nil {
		isNight = isNightTime(data.Current.Time, sunTimes.Sunrise, sunTimes.Sunset)
	}

	// Create current weather data using merged precipitation
	current := &models.WeatherData{
		Timestamp:     timestamp,
		Temperature:   data.Current.Temperature2m,
		FeelsLike:     data.Current.ApparentTemperature,
		Precipitation: precipitation[0],
		Humidity:      data.Current.RelativeHumidity2m,
		WindSpeed:     data.Current.WindSpeed10m,
		WindDirection: data.Current.WindDirection10m,
		CloudCover:    data.Current.CloudCover,
		Pressure:      int(data.Current.Pressure),
		Description:   getWeatherDescription(data.Current.WeatherCode),
		Icon:          getWeatherIconWithTime(data.Current.WeatherCode, isNight),
	}

	// Parse forecast data (all hourly data) using merged precipitation
	var forecast []models.WeatherData
	for i := 0; i < len(data.Hourly.Time); i++ {
		// Bounds check: ensure all arrays have data for this index
		if i >= len(temp) || i >= len(feelsLike) ||
			i >= len(precipitation) || i >= len(humidity) ||
			i >= len(windSpeed) || i >= len(windDir) ||
			i >= len(cloudCover) || i >= len(pressure) ||
			i >= len(weatherCode) {
			log.Printf("Skipping hour %d due to incomplete data arrays", i)
			continue
		}

		timestamp, err := parseTimestampUTC(data.Hourly.Time[i])
		if err != nil {
			log.Printf("Failed to parse hourly timestamp '%s': %v", data.Hourly.Time[i], err)
			continue
		}

		// Determine day/night for this forecast hour
		hourIsNight := false
		if sunTimes != nil {
			// Find the appropriate sunrise/sunset for this day
			hourIsNight = isNightTimeForForecast(data.Hourly.Time[i], data.Daily)
		}

		weather := models.WeatherData{
			Timestamp:     timestamp,
			Temperature:   temp[i],
			FeelsLike:     feelsLike[i],
			Precipitation: precipitation[i],
			Humidity:      humidity[i],
			WindSpeed:     windSpeed[i],
			WindDirection: windDir[i],
			CloudCover:    cloudCover[i],
			Pressure:      int(pressure[i]),
			Description:   getWeatherDescription(weatherCode[i]),
			Icon:          getWeatherIconWithTime(weatherCode[i], hourIsNight),
		}

		forecast = append(forecast, weather)
	}

	return current, forecast, sunTimes, nil
}

// GetForecast fetches hourly forecast data
// Uses NCEP NAM CONUS model for precipitation data (0-60h) with best_match fallback
func (c *OpenMeteoClient) GetForecast(lat, lon float64) ([]models.WeatherData, error) {
	url := fmt.Sprintf("%s?latitude=%.8f&longitude=%.8f&hourly=temperature_2m,relative_humidity_2m,precipitation,rain,snowfall,cloud_cover,wind_speed_10m,wind_direction_10m,weather_code,apparent_temperature,surface_pressure&temperature_unit=fahrenheit&wind_speed_unit=mph&precipitation_unit=inch&timezone=UTC&forecast_days=16&models=ncep_nam_conus,best_match",
		openMeteoForecastURL, lat, lon)

	resp, err := c.httpClient.Get(url)
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

	// Merge precipitation data: use NAM CONUS for 0-60h, fallback to best_match
	precipitation := mergePrecipitationData(data.Hourly.PrecipitationNAM, data.Hourly.PrecipitationBM)

	// Fallback to single-model fields if multi-model not available
	if len(precipitation) == 0 {
		precipitation = data.Hourly.Precipitation
	}

	// Verify we have precipitation data
	if len(precipitation) == 0 {
		return nil, fmt.Errorf("no precipitation data returned from Open-Meteo")
	}

	// Get hourly data arrays (handles both multi-model and single-model responses)
	temp, feelsLike, humidity, cloudCover, windDir, weatherCode, windSpeed, pressure := data.getHourlyData()

	var forecast []models.WeatherData
	// Return all hourly data using merged precipitation
	for i := 0; i < len(data.Hourly.Time); i++ {
		// Bounds check: ensure all arrays have data for this index
		if i >= len(temp) || i >= len(feelsLike) ||
			i >= len(precipitation) || i >= len(humidity) ||
			i >= len(windSpeed) || i >= len(windDir) ||
			i >= len(cloudCover) || i >= len(pressure) ||
			i >= len(weatherCode) {
			log.Printf("Skipping hour %d due to incomplete data arrays", i)
			continue
		}

		timestamp, err := parseTimestampUTC(data.Hourly.Time[i])
		if err != nil {
			log.Printf("Failed to parse hourly timestamp '%s': %v", data.Hourly.Time[i], err)
			continue
		}

		weather := models.WeatherData{
			Timestamp:     timestamp,
			Temperature:   temp[i],
			FeelsLike:     feelsLike[i],
			Precipitation: precipitation[i],
			Humidity:      humidity[i],
			WindSpeed:     windSpeed[i],
			WindDirection: windDir[i],
			CloudCover:    cloudCover[i],
			Pressure:      int(pressure[i]),
			Description:   getWeatherDescription(weatherCode[i]),
			Icon:          getWeatherIcon(weatherCode[i]),
		}

		forecast = append(forecast, weather)
	}

	return forecast, nil
}

// GetHistoricalWeather fetches recent historical weather data using forecast API with past_days
// Uses default model (reanalysis/observations) for accurate historical precipitation
// NAM CONUS is a forecast model and returns stale forecast data, not actual observations
func (c *OpenMeteoClient) GetHistoricalWeather(lat, lon float64, days int) ([]models.WeatherData, error) {
	// Use forecast API with past_days - gives us recent historical data for rain calculations
	// IMPORTANT: Do NOT use NAM CONUS for historical data - it returns old forecast data
	// Use default model which provides reanalysis/observed data for past hours
	url := fmt.Sprintf("%s?latitude=%.8f&longitude=%.8f&past_days=%d&forecast_days=1&hourly=temperature_2m,relative_humidity_2m,precipitation,rain,snowfall,cloud_cover,wind_speed_10m,wind_direction_10m,weather_code,apparent_temperature,surface_pressure&temperature_unit=fahrenheit&wind_speed_unit=mph&precipitation_unit=inch&timezone=UTC",
		openMeteoForecastURL, lat, lon, days)

	resp, err := c.httpClient.Get(url)
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

	// For historical data, use default precipitation (reanalysis/observations)
	// Do NOT merge NAM CONUS data here - NAM is a forecast model and returns stale data
	precipitation := data.Hourly.Precipitation

	// Get hourly data arrays (handles both multi-model and single-model responses)
	temp, feelsLike, humidity, cloudCover, windDir, weatherCode, windSpeed, pressure := data.getHourlyData()

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
		if i >= len(temp) || i >= len(feelsLike) || i >= len(precipitation) ||
			i >= len(humidity) || i >= len(windSpeed) || i >= len(windDir) ||
			i >= len(cloudCover) || i >= len(pressure) || i >= len(weatherCode) {
			log.Printf("Skipping historical hour %d due to incomplete data arrays", i)
			continue
		}

		weather := models.WeatherData{
			Timestamp:     timestamp,
			Temperature:   temp[i],
			FeelsLike:     feelsLike[i],
			Precipitation: precipitation[i],
			Humidity:      humidity[i],
			WindSpeed:     windSpeed[i],
			WindDirection: windDir[i],
			CloudCover:    cloudCover[i],
			Pressure:      int(pressure[i]),
			Description:   getWeatherDescription(weatherCode[i]),
			Icon:          getWeatherIcon(weatherCode[i]),
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
	Time              []string `json:"time"`
	Sunrise           []string `json:"sunrise"`
	Sunset            []string `json:"sunset"`
	SunriseNAM        []string `json:"sunrise_ncep_nam_conus"`
	SunsetNAM         []string `json:"sunset_ncep_nam_conus"`
	SunriseBestMatch  []string `json:"sunrise_best_match"`
	SunsetBestMatch   []string `json:"sunset_best_match"`
}) bool {
	if daily == nil || len(daily.Time) == 0 {
		return false
	}

	// Select which model's sun data to use (prefer best_match, fallback to NAM, then regular fields)
	var sunriseData, sunsetData []string
	if len(daily.SunriseBestMatch) > 0 && len(daily.SunsetBestMatch) > 0 {
		sunriseData = daily.SunriseBestMatch
		sunsetData = daily.SunsetBestMatch
	} else if len(daily.SunriseNAM) > 0 && len(daily.SunsetNAM) > 0 {
		sunriseData = daily.SunriseNAM
		sunsetData = daily.SunsetNAM
	} else if len(daily.Sunrise) > 0 && len(daily.Sunset) > 0 {
		sunriseData = daily.Sunrise
		sunsetData = daily.Sunset
	}

	if len(sunriseData) == 0 || len(sunsetData) == 0 {
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
		if day == dateStr && i < len(sunriseData) && i < len(sunsetData) {
			return isNightTime(timeStr, sunriseData[i], sunsetData[i])
		}
	}

	// If no match found, use first day's sunrise/sunset as approximation
	if len(sunriseData) > 0 && len(sunsetData) > 0 {
		return isNightTime(timeStr, sunriseData[0], sunsetData[0])
	}

	return false
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

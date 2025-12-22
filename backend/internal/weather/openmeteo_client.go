package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

const (
	openMeteoForecastURL   = "https://api.open-meteo.com/v1/forecast"
	openMeteoHistoricalURL = "https://api.open-meteo.com/v1/archive"
)

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
	} `json:"hourly"`
}

func NewOpenMeteoClient() *OpenMeteoClient {
	return &OpenMeteoClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetCurrentWeather fetches current weather observations
func (c *OpenMeteoClient) GetCurrentWeather(lat, lon float64) (*models.WeatherData, error) {
	url := fmt.Sprintf("%s?latitude=%.8f&longitude=%.8f&current=temperature_2m,relative_humidity_2m,precipitation,rain,snowfall,cloud_cover,wind_speed_10m,wind_direction_10m,weather_code,apparent_temperature,surface_pressure&temperature_unit=fahrenheit&wind_speed_unit=mph&precipitation_unit=inch&timezone=auto",
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

	// Parse current weather timestamp
	timestamp, err := time.Parse(time.RFC3339, data.Current.Time)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	weather := &models.WeatherData{
		Timestamp:     timestamp,
		Temperature:   data.Current.Temperature2m,
		FeelsLike:     data.Current.ApparentTemperature,
		Precipitation: data.Current.Rain + data.Current.Snowfall, // Combine rain and snow
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

// GetForecast fetches hourly forecast data
func (c *OpenMeteoClient) GetForecast(lat, lon float64) ([]models.WeatherData, error) {
	url := fmt.Sprintf("%s?latitude=%.8f&longitude=%.8f&hourly=temperature_2m,relative_humidity_2m,precipitation,rain,snowfall,cloud_cover,wind_speed_10m,wind_direction_10m,weather_code,apparent_temperature,surface_pressure&temperature_unit=fahrenheit&wind_speed_unit=mph&precipitation_unit=inch&timezone=auto&forecast_days=16",
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

	var forecast []models.WeatherData
	for i := range data.Hourly.Time {
		timestamp, err := time.Parse(time.RFC3339, data.Hourly.Time[i])
		if err != nil {
			continue
		}

		weather := models.WeatherData{
			Timestamp:     timestamp,
			Temperature:   data.Hourly.Temperature2m[i],
			FeelsLike:     data.Hourly.ApparentTemperature[i],
			Precipitation: data.Hourly.Rain[i] + data.Hourly.Snowfall[i], // Combine rain and snow
			Humidity:      data.Hourly.RelativeHumidity2m[i],
			WindSpeed:     data.Hourly.WindSpeed10m[i],
			WindDirection: data.Hourly.WindDirection10m[i],
			CloudCover:    data.Hourly.CloudCover[i],
			Pressure:      int(data.Hourly.Pressure[i]),
			Description:   getWeatherDescription(data.Hourly.WeatherCode[i]),
			Icon:          getWeatherIcon(data.Hourly.WeatherCode[i]),
		}

		forecast = append(forecast, weather)
	}

	return forecast, nil
}

// GetHistoricalWeather fetches historical weather data
func (c *OpenMeteoClient) GetHistoricalWeather(lat, lon float64, days int) ([]models.WeatherData, error) {
	endDate := time.Now().Add(-24 * time.Hour).Format("2006-01-02") // Yesterday
	startDate := time.Now().Add(-time.Duration(days) * 24 * time.Hour).Format("2006-01-02")

	url := fmt.Sprintf("%s?latitude=%.8f&longitude=%.8f&start_date=%s&end_date=%s&hourly=temperature_2m,relative_humidity_2m,precipitation,rain,snowfall,cloud_cover,wind_speed_10m,wind_direction_10m,weather_code,apparent_temperature,surface_pressure&temperature_unit=fahrenheit&wind_speed_unit=mph&precipitation_unit=inch&timezone=auto",
		openMeteoHistoricalURL, lat, lon, startDate, endDate)

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

	var historical []models.WeatherData
	for i := range data.Hourly.Time {
		timestamp, err := time.Parse(time.RFC3339, data.Hourly.Time[i])
		if err != nil {
			continue
		}

		weather := models.WeatherData{
			Timestamp:     timestamp,
			Temperature:   data.Hourly.Temperature2m[i],
			FeelsLike:     data.Hourly.ApparentTemperature[i],
			Precipitation: data.Hourly.Rain[i] + data.Hourly.Snowfall[i],
			Humidity:      data.Hourly.RelativeHumidity2m[i],
			WindSpeed:     data.Hourly.WindSpeed10m[i],
			WindDirection: data.Hourly.WindDirection10m[i],
			CloudCover:    data.Hourly.CloudCover[i],
			Pressure:      int(data.Hourly.Pressure[i]),
			Description:   getWeatherDescription(data.Hourly.WeatherCode[i]),
			Icon:          getWeatherIcon(data.Hourly.WeatherCode[i]),
		}

		historical = append(historical, weather)
	}

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

// Map WMO weather codes to OpenWeatherMap-like icon codes for consistency
func getWeatherIcon(code int) string {
	// Map to OpenWeatherMap icon codes to maintain frontend compatibility
	iconMap := map[int]string{
		0:  "01d", // Clear sky
		1:  "02d", // Mainly clear
		2:  "03d", // Partly cloudy
		3:  "04d", // Overcast
		45: "50d", // Fog
		48: "50d", // Rime fog
		51: "09d", // Light drizzle
		53: "09d", // Moderate drizzle
		55: "09d", // Dense drizzle
		61: "10d", // Slight rain
		63: "10d", // Moderate rain
		65: "10d", // Heavy rain
		71: "13d", // Slight snow
		73: "13d", // Moderate snow
		75: "13d", // Heavy snow
		77: "13d", // Snow grains
		80: "09d", // Rain showers
		81: "09d", // Moderate rain showers
		82: "09d", // Violent rain showers
		85: "13d", // Snow showers
		86: "13d", // Heavy snow showers
		95: "11d", // Thunderstorm
		96: "11d", // Thunderstorm with hail
		99: "11d", // Thunderstorm with heavy hail
	}

	if icon, ok := iconMap[code]; ok {
		return icon
	}
	return "01d" // Default to clear sky
}

package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

const (
	baseURL = "https://api.openweathermap.org/data/2.5"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

// OpenWeatherMap API response structures
type owmResponse struct {
	List []struct {
		Dt   int64 `json:"dt"`
		Main struct {
			Temp      float64 `json:"temp"`
			FeelsLike float64 `json:"feels_like"`
			Pressure  int     `json:"pressure"`
			Humidity  int     `json:"humidity"`
		} `json:"main"`
		Weather []struct {
			Description string `json:"description"`
			Icon        string `json:"icon"`
		} `json:"weather"`
		Clouds struct {
			All int `json:"all"`
		} `json:"clouds"`
		Wind struct {
			Speed float64 `json:"speed"`
			Deg   int     `json:"deg"`
		} `json:"wind"`
		Rain struct {
			ThreeH float64 `json:"3h"`
		} `json:"rain"`
		Snow struct {
			ThreeH float64 `json:"3h"`
		} `json:"snow"`
		Pop float64 `json:"pop"` // Probability of precipitation
	} `json:"list"`
}

type owmCurrentResponse struct {
	Dt   int64 `json:"dt"`
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		Pressure  int     `json:"pressure"`
		Humidity  int     `json:"humidity"`
	} `json:"main"`
	Weather []struct {
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`
	Wind struct {
		Speed float64 `json:"speed"`
		Deg   int     `json:"deg"`
	} `json:"wind"`
	Rain struct {
		OneH float64 `json:"1h"`
	} `json:"rain"`
}

func NewClient() *Client {
	return &Client{
		apiKey: os.Getenv("OPENWEATHERMAP_API_KEY"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetCurrentWeather fetches current weather for a location
func (c *Client) GetCurrentWeather(lat, lon float64) (*models.WeatherData, error) {
	url := fmt.Sprintf("%s/weather?lat=%.8f&lon=%.8f&appid=%s&units=imperial", baseURL, lat, lon, c.apiKey)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch current weather: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var data owmCurrentResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	weather := &models.WeatherData{
		Timestamp:     time.Unix(data.Dt, 0),
		Temperature:   data.Main.Temp,
		FeelsLike:     data.Main.FeelsLike,
		Precipitation: data.Rain.OneH / 25.4, // Convert mm to inches
		Humidity:      data.Main.Humidity,
		WindSpeed:     data.Wind.Speed,
		WindDirection: data.Wind.Deg,
		CloudCover:    data.Clouds.All,
		Pressure:      data.Main.Pressure,
	}

	if len(data.Weather) > 0 {
		weather.Description = data.Weather[0].Description
		weather.Icon = data.Weather[0].Icon
	}

	return weather, nil
}

// GetForecast fetches 5-day/3-hour forecast for a location
func (c *Client) GetForecast(lat, lon float64) ([]models.WeatherData, error) {
	url := fmt.Sprintf("%s/forecast?lat=%.8f&lon=%.8f&appid=%s&units=imperial", baseURL, lat, lon, c.apiKey)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch forecast: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var data owmResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var forecast []models.WeatherData
	for _, item := range data.List {
		// Combine rain and snow into total precipitation
		totalPrecip := (item.Rain.ThreeH + item.Snow.ThreeH) / 25.4 // Convert mm to inches

		weather := models.WeatherData{
			Timestamp:     time.Unix(item.Dt, 0),
			Temperature:   item.Main.Temp,
			FeelsLike:     item.Main.FeelsLike,
			Precipitation: totalPrecip,
			Humidity:      item.Main.Humidity,
			WindSpeed:     item.Wind.Speed,
			WindDirection: item.Wind.Deg,
			CloudCover:    item.Clouds.All,
			Pressure:      item.Main.Pressure,
		}

		if len(item.Weather) > 0 {
			weather.Description = item.Weather[0].Description
			weather.Icon = item.Weather[0].Icon
		}

		forecast = append(forecast, weather)
	}

	return forecast, nil
}

package weather

import (
	"log"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// WeatherService provides weather data with fallback support
type WeatherService struct {
	openMeteo       *OpenMeteoClient
	openWeatherMap  *Client
	preferOpenMeteo bool
}

// NewWeatherService creates a new weather service with both providers
func NewWeatherService() *WeatherService {
	return &WeatherService{
		openMeteo:       NewOpenMeteoClient(),
		openWeatherMap:  NewClient(),
		preferOpenMeteo: true, // Prefer Open-Meteo by default
	}
}

// GetCurrentAndForecast fetches both current weather and forecast in a single API call
func (s *WeatherService) GetCurrentAndForecast(lat, lon float64) (*models.WeatherData, []models.WeatherData, error) {
	if s.preferOpenMeteo {
		current, forecast, err := s.openMeteo.GetCurrentAndForecast(lat, lon)
		if err == nil {
			log.Printf("Successfully fetched current + forecast from Open-Meteo for (%.6f, %.6f) - %d hours", lat, lon, len(forecast))
			return current, forecast, nil
		}
		log.Printf("Open-Meteo failed for current+forecast (%.6f, %.6f): %v, falling back to separate calls", lat, lon, err)
	}

	// Fallback to separate calls
	current, err := s.GetCurrentWeather(lat, lon)
	if err != nil {
		return nil, nil, err
	}
	forecast, err := s.GetForecast(lat, lon)
	if err != nil {
		return nil, nil, err
	}
	return current, forecast, nil
}

// GetCurrentWeather fetches current weather with fallback
func (s *WeatherService) GetCurrentWeather(lat, lon float64) (*models.WeatherData, error) {
	if s.preferOpenMeteo {
		data, err := s.openMeteo.GetCurrentWeather(lat, lon)
		if err == nil {
			log.Printf("Successfully fetched current weather from Open-Meteo for (%.6f, %.6f)", lat, lon)
			return data, nil
		}
		log.Printf("Open-Meteo failed for current weather (%.6f, %.6f): %v, falling back to OpenWeatherMap", lat, lon, err)
	}

	data, err := s.openWeatherMap.GetCurrentWeather(lat, lon)
	if err != nil {
		return nil, err
	}
	log.Printf("Successfully fetched current weather from OpenWeatherMap for (%.6f, %.6f)", lat, lon)
	return data, nil
}

// GetForecast fetches forecast with fallback
func (s *WeatherService) GetForecast(lat, lon float64) ([]models.WeatherData, error) {
	if s.preferOpenMeteo {
		data, err := s.openMeteo.GetForecast(lat, lon)
		if err == nil {
			log.Printf("Successfully fetched forecast from Open-Meteo for (%.6f, %.6f) - %d hours", lat, lon, len(data))
			return data, nil
		}
		log.Printf("Open-Meteo failed for forecast (%.6f, %.6f): %v, falling back to OpenWeatherMap", lat, lon, err)
	}

	data, err := s.openWeatherMap.GetForecast(lat, lon)
	if err != nil {
		return nil, err
	}
	log.Printf("Successfully fetched forecast from OpenWeatherMap for (%.6f, %.6f) - %d hours", lat, lon, len(data))
	return data, nil
}

// GetHistoricalWeather fetches historical weather (Open-Meteo only, no fallback needed)
func (s *WeatherService) GetHistoricalWeather(lat, lon float64, days int) ([]models.WeatherData, error) {
	// Open-Meteo has true historical data, so we use it exclusively for this
	data, err := s.openMeteo.GetHistoricalWeather(lat, lon, days)
	if err != nil {
		log.Printf("Open-Meteo failed for historical weather (%.6f, %.6f): %v", lat, lon, err)
		return nil, err
	}
	log.Printf("Successfully fetched historical weather from Open-Meteo for (%.6f, %.6f) - %d hours", lat, lon, len(data))
	return data, nil
}

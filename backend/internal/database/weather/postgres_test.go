package weather

import (
	"strings"
	"testing"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

func modelsWeatherDataForTest() models.WeatherData {
	return models.WeatherData{
		LocationID:         1,
		Timestamp:          time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC),
		Temperature:        70,
		FeelsLike:          70,
		Precipitation:      0,
		Humidity:           50,
		WindSpeed:          3,
		WindDirection:      180,
		CloudCover:         10,
		Pressure:           1013,
		Description:        "clear",
		Icon:               "01d",
		ShortwaveRadiation: 100,
		DirectRadiation:    80,
		DiffuseRadiation:   20,
		DewpointF:          40,
	}
}

func TestBuildBulkInsertQueryUsesWeatherNoOpGuard(t *testing.T) {
	query, args := buildBulkInsertQuery([]models.WeatherData{modelsWeatherDataForTest()})

	if len(args) != weatherDataColumnCount {
		t.Fatalf("args length = %d, want %d", len(args), weatherDataColumnCount)
	}
	for _, want := range []string{
		"ON CONFLICT(location_id, timestamp) DO UPDATE SET",
		"WHERE weather_data.temperature",
		"OR weather_data.dewpoint_f",
	} {
		if !strings.Contains(query, want) {
			t.Fatalf("bulk insert query missing %q", want)
		}
	}
}

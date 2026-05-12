package rock_drying

import (
	"testing"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// TestCalculator_TracePrecipitationDoesNotTriggerDrying is a regression test
// for the "Index" bug (location_id=2): a single sub-mm precipitation reading
// (e.g. 0.016") from Open-Meteo — which is sensor noise / dew, not real rain —
// was being treated as a rain event by findLastRainEvent. That caused the
// rock-drying status to report "drying / 1 hour until dry" indefinitely even
// though no actual rain had fallen.
//
// With the rainEventThresholdInches constant (0.03"), trace values must be
// ignored and the rock should be reported as dry.
func TestCalculator_TracePrecipitationDoesNotTriggerDrying(t *testing.T) {
	calc := &Calculator{}

	rockTypes := []models.RockType{
		{
			Name:            "Granite",
			BaseDryingHours: 6.0,
			PorosityPercent: 1.0,
			IsWetSensitive:  false,
			GroupName:       "Granite",
		},
	}

	now := time.Now()

	// Current weather: dry, mild, breezy — Index-like conditions
	currentWeather := &models.WeatherData{
		Temperature:   63.7,
		Precipitation: 0.0,
		Humidity:      74,
		WindSpeed:     7.4,
		CloudCover:    58,
		Timestamp:     now,
	}

	// Build 168h (7d) of historical data, all zero except a single
	// trace reading at T-117h (the actual Index bug fixture: 0.016").
	historical := make([]models.WeatherData, 168)
	for i := 0; i < 168; i++ {
		hoursAgo := 168 - i // oldest first
		historical[i] = models.WeatherData{
			Temperature:   60.0,
			Precipitation: 0.0,
			Humidity:      70,
			WindSpeed:     5.0,
			CloudCover:    50,
			Timestamp:     now.Add(time.Duration(-hoursAgo) * time.Hour),
		}
	}
	// Inject the trace precip ~117h ago (index 168-117 = 51).
	historical[51].Precipitation = 0.016

	status := calc.CalculateDryingStatus(rockTypes, currentWeather, historical, nil, false, nil)

	if status.IsWet {
		t.Errorf("Expected IsWet=false (trace precip 0.016\" is sensor noise, not rain), got IsWet=true")
	}
	if status.HoursUntilDry != 0 {
		t.Errorf("Expected HoursUntilDry=0 (no real rain event), got %.2f", status.HoursUntilDry)
	}
	if status.Status != "good" {
		t.Errorf("Expected Status=\"good\" (rock is dry), got %q (message: %q)", status.Status, status.Message)
	}
	if !status.IsSafe {
		t.Errorf("Expected IsSafe=true, got false")
	}
}

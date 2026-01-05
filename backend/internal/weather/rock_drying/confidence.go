package rock_drying

import (
	"math"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// calculateConfidence calculates confidence score (0-100) for the prediction
func calculateConfidence(
	isDry bool,
	lastRainTime time.Time,
	historicalWeather []models.WeatherData,
	isWetSensitive bool,
	sunExposure *models.LocationSunExposure,
) int {
	confidence := 75.0 // Start with baseline

	// Data completeness factor
	if len(historicalWeather) < 24 {
		confidence -= 15 // -15 for insufficient historical data
	} else if len(historicalWeather) < 48 {
		confidence -= 8 // -8 for limited historical data
	}

	// Sun exposure profile availability
	if sunExposure == nil {
		confidence -= 10 // -10 for missing sun exposure data
	}

	// Time since rain factor (affects uncertainty)
	if !isDry && !lastRainTime.IsZero() {
		hoursSinceRain := time.Since(lastRainTime).Hours()

		if hoursSinceRain < 6 {
			confidence -= 5 // Recent rain = slightly more uncertain
		} else if hoursSinceRain > 72 {
			confidence += 8 // Long time = more confident in dryness
		}
	}

	// Weather stability (check for variable conditions)
	if len(historicalWeather) >= 12 {
		tempVariance := calculateTemperatureVariance(historicalWeather)
		if tempVariance > 15 {
			confidence -= 8 // High variance = less confident
		} else if tempVariance > 10 {
			confidence -= 4
		}
	}

	// Wet-sensitive rock factor (err on side of caution)
	if isWetSensitive && !isDry {
		confidence -= 5 // Lower confidence when wet-sensitive rock is wet
	}

	// Dry conditions boost
	if isDry && (lastRainTime.IsZero() || time.Since(lastRainTime).Hours() > 96) {
		confidence += 10 // High confidence in long-term dry conditions
	}

	// Clamp to realistic range (20-95)
	if confidence < 20 {
		confidence = 20
	}
	if confidence > 95 {
		confidence = 95
	}

	return int(confidence)
}

// calculateTemperatureVariance calculates temperature variance in recent weather
func calculateTemperatureVariance(weather []models.WeatherData) float64 {
	if len(weather) == 0 {
		return 0
	}

	// Calculate mean
	sum := 0.0
	for _, w := range weather {
		sum += w.Temperature
	}
	mean := sum / float64(len(weather))

	// Calculate variance
	varianceSum := 0.0
	for _, w := range weather {
		diff := w.Temperature - mean
		varianceSum += diff * diff
	}
	variance := varianceSum / float64(len(weather))

	return math.Sqrt(variance) // Return standard deviation
}

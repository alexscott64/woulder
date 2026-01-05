package pests

import (
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// PestLevel represents the severity of pest activity
type PestLevel string

const (
	PestLevelLow      PestLevel = "low"
	PestLevelModerate PestLevel = "moderate"
	PestLevelHigh     PestLevel = "high"
	PestLevelVeryHigh PestLevel = "very_high"
	PestLevelExtreme  PestLevel = "extreme"
)

// PestConditions represents overall pest conditions
type PestConditions struct {
	MosquitoLevel    PestLevel `json:"mosquito_level"`
	MosquitoScore    int       `json:"mosquito_score"` // 0-100
	OutdoorPestLevel PestLevel `json:"outdoor_pest_level"`
	OutdoorPestScore int       `json:"outdoor_pest_score"` // 0-100
	Factors          []string  `json:"factors"`
}

// GetSeasonalMosquitoFactor returns activity multiplier based on month
// Based on typical northern hemisphere patterns (WA state latitude ~47-49°N)
func GetSeasonalMosquitoFactor(month time.Month) float64 {
	factors := map[time.Month]float64{
		time.January:   0.0, // Dormant
		time.February:  0.0, // Dormant
		time.March:     0.1, // Emerging
		time.April:     0.3, // Increasing
		time.May:       0.6, // Active
		time.June:      0.9, // High activity
		time.July:      1.0, // Peak
		time.August:    1.0, // Peak
		time.September: 0.7, // Declining
		time.October:   0.3, // Late season
		time.November:  0.1, // Dying off
		time.December:  0.0, // Dormant
	}
	return factors[month]
}

// GetSeasonalPestFactor returns general pest activity multiplier based on month
func GetSeasonalPestFactor(month time.Month) float64 {
	factors := map[time.Month]float64{
		time.January:   0.1, // Minimal
		time.February:  0.1, // Minimal
		time.March:     0.3, // Emerging
		time.April:     0.5, // Increasing
		time.May:       0.8, // Active
		time.June:      1.0, // High
		time.July:      1.0, // Peak
		time.August:    1.0, // Peak
		time.September: 0.8, // Still high
		time.October:   0.5, // Declining
		time.November:  0.2, // Low
		time.December:  0.1, // Minimal
	}
	return factors[month]
}

// ScoreToLevel converts a 0-100 score to a pest level category
func ScoreToLevel(score int) PestLevel {
	if score >= 80 {
		return PestLevelExtreme
	}
	if score >= 60 {
		return PestLevelVeryHigh
	}
	if score >= 40 {
		return PestLevelHigh
	}
	if score >= 20 {
		return PestLevelModerate
	}
	return PestLevelLow
}

// CalculateRecentRainfall sums precipitation from last N days
func CalculateRecentRainfall(historical []models.WeatherData, daysAgo int) float64 {
	cutoffTime := time.Now().Add(-time.Duration(daysAgo) * 24 * time.Hour)
	total := 0.0

	for _, h := range historical {
		if h.Timestamp.After(cutoffTime) || h.Timestamp.Equal(cutoffTime) {
			total += h.Precipitation
		}
	}

	return total
}

// GetMaxTemperature returns the maximum temperature from weather data
func GetMaxTemperature(data []models.WeatherData) float64 {
	if len(data) == 0 {
		return 0
	}

	max := data[0].Temperature
	for _, d := range data[1:] {
		if d.Temperature > max {
			max = d.Temperature
		}
	}
	return max
}

// GetAverageValue returns average of a weather metric
func GetAverageValue(data []models.WeatherData, metric string) float64 {
	if len(data) == 0 {
		return 0
	}

	sum := 0.0
	for _, d := range data {
		switch metric {
		case "temperature":
			sum += d.Temperature
		case "humidity":
			sum += float64(d.Humidity)
		case "wind_speed":
			sum += d.WindSpeed
		}
	}

	return sum / float64(len(data))
}

// GetWorstLevel returns the worse of two pest levels
func GetWorstLevel(level1, level2 PestLevel) PestLevel {
	levelOrder := []PestLevel{
		PestLevelLow,
		PestLevelModerate,
		PestLevelHigh,
		PestLevelVeryHigh,
		PestLevelExtreme,
	}

	index1 := 0
	index2 := 0

	for i, l := range levelOrder {
		if l == level1 {
			index1 = i
		}
		if l == level2 {
			index2 = i
		}
	}

	if index1 > index2 {
		return level1
	}
	return level2
}

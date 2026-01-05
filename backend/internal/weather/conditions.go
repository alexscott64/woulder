package weather

import (
	"fmt"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/alexscott64/woulder/backend/internal/weather/calculator"
)

// ConditionCalculator calculates climbing conditions from weather data
type ConditionCalculator struct{}

// CalculateTodayCondition calculates today's overall climbing condition
// This matches the frontend WeatherCard todayCondition logic
func (c *ConditionCalculator) CalculateTodayCondition(
	current *models.WeatherData,
	hourlyForecast []models.WeatherData,
	historical []models.WeatherData,
) models.ClimbingCondition {
	if current == nil || len(hourlyForecast) == 0 {
		return c.calculateInstantCondition(current, historical)
	}

	// Get today's date in Pacific timezone
	pacificLoc, _ := time.LoadLocation("America/Los_Angeles")
	if pacificLoc == nil {
		pacificLoc = time.UTC
	}
	now := time.Now().In(pacificLoc)
	todayStr := now.Format("2006-01-02")

	// Combine current + hourly for today
	allData := append([]models.WeatherData{*current}, hourlyForecast...)
	var todayHours []models.WeatherData

	for _, data := range allData {
		dataPacific := data.Timestamp.In(pacificLoc)
		dateKey := dataPacific.Format("2006-01-02")
		if dateKey == todayStr {
			todayHours = append(todayHours, data)
		}
	}

	if len(todayHours) == 0 {
		return c.calculateInstantCondition(current, historical)
	}

	// Calculate condition for each hour
	var badHours []hourCondition
	var marginalHours []hourCondition

	for i, hour := range todayHours {
		// Get recent hours for context (previous 2 hours from today's data)
		var recentHours []models.WeatherData
		if i > 0 {
			start := i - 2
			if start < 0 {
				start = 0
			}
			recentHours = todayHours[start:i]
		}

		cond := c.calculateInstantCondition(&hour, recentHours)
		hc := hourCondition{
			condition:      cond,
			hour:           hour,
			isClimbingTime: isClimbingHour(hour.Timestamp, pacificLoc),
		}

		// Filter: include all climbing hours, or non-climbing hours with rain/wind issues
		include := hc.isClimbingTime
		if !include {
			for _, r := range cond.Reasons {
				lowerReason := toLower(r)
				if contains(lowerReason, "rain") || contains(lowerReason, "precip") || contains(lowerReason, "wind") {
					include = true
					break
				}
			}
		}

		if !include {
			continue
		}

		if cond.Level == "bad" {
			badHours = append(badHours, hc)
		} else if cond.Level == "marginal" {
			marginalHours = append(marginalHours, hc)
		}
	}

	// Determine overall level and consolidate reasons
	level := "good"
	var reasons []string

	if len(badHours) > 0 {
		level = "bad"
		reasons = append(reasons, consolidateReasons(badHours)...)
		if len(badHours) > 1 {
			reasons = append(reasons, fmt.Sprintf("%d hours with poor conditions", len(badHours)))
		}
	} else if len(marginalHours) > 0 {
		level = "marginal"
		reasons = append(reasons, consolidateReasons(marginalHours)...)
		if len(marginalHours) > 1 {
			reasons = append(reasons, fmt.Sprintf("%d hours with fair conditions", len(marginalHours)))
		}
	} else {
		reasons = append(reasons, "Good climbing conditions all day")
	}

	// Add unified 48h rain total
	rainLast48h := c.CalculateRainLast48h(historical, hourlyForecast)
	if rainLast48h > 0.5 {
		if level == "good" {
			level = "marginal"
		}
		reasons = append(reasons, fmt.Sprintf("Recent heavy rain (%.2fin in last 48h)", rainLast48h))
	} else if rainLast48h > 0.2 {
		reasons = append(reasons, fmt.Sprintf("Recent rain (%.2fin in last 48h)", rainLast48h))
	}

	return models.ClimbingCondition{Level: level, Reasons: reasons}
}

// CalculateRainLast48h calculates total rain in last 48 hours from all available data (exported)
func (c *ConditionCalculator) CalculateRainLast48h(historical []models.WeatherData, hourly []models.WeatherData) float64 {
	now := time.Now()
	fortyEightHoursAgo := now.Add(-48 * time.Hour)

	// Deduplicate by timestamp
	dataMap := make(map[string]models.WeatherData)
	for _, h := range historical {
		dataMap[h.Timestamp.Format(time.RFC3339)] = h
	}
	for _, h := range hourly {
		dataMap[h.Timestamp.Format(time.RFC3339)] = h
	}

	// Sum rain in window
	total := 0.0
	for _, data := range dataMap {
		if (data.Timestamp.After(fortyEightHoursAgo) || data.Timestamp.Equal(fortyEightHoursAgo)) &&
			(data.Timestamp.Before(now) || data.Timestamp.Equal(now)) {
			total += data.Precipitation
		}
	}

	return total
}

// calculateInstantCondition calculates condition for a single point in time
func (c *ConditionCalculator) calculateInstantCondition(
	weather *models.WeatherData,
	recentWeather []models.WeatherData,
) models.ClimbingCondition {
	if weather == nil {
		return models.ClimbingCondition{Level: "good", Reasons: []string{}}
	}

	reasons := []string{}
	level := "good"

	// Check precipitation
	if weather.Precipitation >= 0.05 {
		level = "bad"
		description := "Moderate"
		if weather.Precipitation >= 0.3 {
			description = "Heavy"
		}
		reasons = append(reasons, fmt.Sprintf("%s rain (%.2fin/hr)", description, weather.Precipitation))
	} else if weather.Precipitation >= 0.01 {
		level = "marginal"
		// Check for persistent drizzle
		isPersistent := calculator.HasPersistentPrecipitation(recentWeather, 0.01)
		if isPersistent && len(recentWeather) > 0 {
			recentTotal := calculator.GetTotalPrecipitation(recentWeather)
			total := recentTotal + weather.Precipitation
			reasons = append(reasons, fmt.Sprintf("Persistent drizzle (%.2fin over %dh)", total, len(recentWeather)))
		} else {
			reasons = append(reasons, fmt.Sprintf("Light rain (%.2fin/hr)", weather.Precipitation))
		}
	}

	// Check temperature
	if weather.Temperature < 40 {
		if level == "good" {
			level = "bad"
		}
		reasons = append(reasons, fmt.Sprintf("Too cold (%.0f°F)", weather.Temperature))
	} else if weather.Temperature < 45 {
		if level == "good" {
			level = "marginal"
		}
		reasons = append(reasons, fmt.Sprintf("Cold (%.0f°F)", weather.Temperature))
	} else if weather.Temperature > 90 {
		if level == "good" {
			level = "bad"
		}
		reasons = append(reasons, fmt.Sprintf("Too hot (%.0f°F)", weather.Temperature))
	} else if weather.Temperature > 85 {
		if level == "good" {
			level = "marginal"
		}
		reasons = append(reasons, fmt.Sprintf("Warm (%.0f°F)", weather.Temperature))
	}

	// Check wind
	if weather.WindSpeed > 30 {
		if level == "good" {
			level = "bad"
		}
		reasons = append(reasons, fmt.Sprintf("Dangerous winds (%.0fmph)", weather.WindSpeed))
	} else if weather.WindSpeed > 20 {
		if level == "good" {
			level = "marginal"
		}
		reasons = append(reasons, fmt.Sprintf("Strong winds (%.0fmph)", weather.WindSpeed))
	} else if weather.WindSpeed > 12 {
		if level == "good" {
			level = "marginal"
		}
		reasons = append(reasons, fmt.Sprintf("Moderate winds (%.0fmph)", weather.WindSpeed))
	}

	// Check humidity
	if weather.Humidity >= 85 {
		if level == "good" {
			level = "marginal"
		}
		reasons = append(reasons, fmt.Sprintf("High humidity (%d%%)", weather.Humidity))
	}

	return models.ClimbingCondition{Level: level, Reasons: reasons}
}

// Helper types and functions

type hourCondition struct {
	condition      models.ClimbingCondition
	hour           models.WeatherData
	isClimbingTime bool
}

func isClimbingHour(timestamp time.Time, pacificLoc *time.Location) bool {
	hourPacific := timestamp.In(pacificLoc).Hour()
	return hourPacific >= 9 && hourPacific < 20 // 9am-8pm Pacific
}

func consolidateReasons(hours []hourCondition) []string {
	// Track worst value for each factor type
	factorMap := make(map[string]factorValue)

	for _, hc := range hours {
		for _, reason := range hc.condition.Reasons {
			// Skip "Drying slowly" or "recent rain" - unified 48h calc added separately
			lowerReason := toLower(reason)
			if contains(lowerReason, "drying slowly") || contains(lowerReason, "recent rain") {
				continue
			}

			// Extract and track worst values
			updateFactorMap(factorMap, reason)
		}
	}

	// Convert map to slice of reasons
	var consolidated []string
	for _, fv := range factorMap {
		consolidated = append(consolidated, fv.reason)
	}

	return consolidated
}

type factorValue struct {
	reason string
	value  float64
}

func updateFactorMap(factorMap map[string]factorValue, reason string) {
	// Simple implementation - just keep unique reasons for now
	// Frontend did complex extraction, but for MVP we can keep it simple
	key := toLower(reason)
	key = removeNonAlpha(key)
	factorMap[key] = factorValue{reason: reason, value: 0}
}

func toLower(s string) string {
	// Simple ASCII lowercase
	result := ""
	for _, c := range s {
		if c >= 'A' && c <= 'Z' {
			result += string(c + 32)
		} else {
			result += string(c)
		}
	}
	return result
}

func removeNonAlpha(s string) string {
	result := ""
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			result += string(c)
		}
	}
	return result
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

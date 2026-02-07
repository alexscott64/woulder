package rock_drying

import (
	"math"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// Calculator calculates rock drying status based on weather and rock type
type Calculator struct{}

// CalculateDryingStatus determines if rock is dry and safe to climb
func (c *Calculator) CalculateDryingStatus(
	rockTypes []models.RockType,
	currentWeather *models.WeatherData,
	historicalWeather []models.WeatherData,
	sunExposure *models.LocationSunExposure,
	hasSeepageRisk bool,
	snowDepthInches *float64,
) models.RockDryingStatus {
	if len(rockTypes) == 0 {
		return models.RockDryingStatus{
			IsWet:           false,
			IsSafe:          true,
			Status:          "good",
			Message:         "No rock type data available",
			RockTypes:       []string{},
			ConfidenceScore: 30, // Low confidence due to missing data
		}
	}

	// Determine primary rock type (first in list, should be marked as primary in DB)
	primaryRock := rockTypes[0]
	rockTypeNames := make([]string, len(rockTypes))
	hasWetSensitive := false

	for i, rt := range rockTypes {
		rockTypeNames[i] = rt.Name
		if rt.IsWetSensitive {
			hasWetSensitive = true
		}
	}

	// Check for snow on ground - this is CRITICAL only for wet-sensitive rocks
	if snowDepthInches != nil && *snowDepthInches > 0.5 {
		return c.handleSnowOnGround(
			*snowDepthInches,
			currentWeather,
			historicalWeather,
			primaryRock,
			sunExposure,
			hasWetSensitive,
			rockTypeNames,
		)
	}

	// Check for freezing temperatures - water frozen on rock
	if currentWeather.Temperature <= 32 {
		status := c.handleFreezingConditions(
			currentWeather,
			historicalWeather,
			primaryRock,
			sunExposure,
			hasWetSensitive,
			rockTypeNames,
		)
		if status != nil {
			return *status
		}
	}

	// If currently raining
	if currentWeather.Precipitation > 0.01 {
		return c.handleCurrentRain(
			currentWeather,
			historicalWeather,
			primaryRock,
			sunExposure,
			hasSeepageRisk,
			hasWetSensitive,
			rockTypeNames,
		)
	}

	// Find last rain event
	lastRainEvent := findLastRainEvent(historicalWeather, currentWeather)

	// If no recent rain
	if lastRainEvent == nil {
		return models.RockDryingStatus{
			IsWet:            false,
			IsSafe:           true,
			IsWetSensitive:   hasWetSensitive,
			HoursUntilDry:    0,
			Status:           "good",
			Message:          "Rock is dry - no recent rain",
			RockTypes:        rockTypeNames,
			PrimaryRockType:  primaryRock.Name,
			PrimaryGroupName: primaryRock.GroupName,
			ConfidenceScore:  calculateConfidence(true, time.Time{}, historicalWeather, hasWetSensitive, sunExposure),
		}
	}

	// Calculate time-weighted drying progress
	requiredDryingTime := estimateDryingTime(primaryRock, currentWeather, historicalWeather,
		sunExposure, hasSeepageRisk, lastRainEvent.TotalRain)

	// Calculate how much drying has occurred (time-weighted)
	dryingProgress := calculateTimeWeightedDrying(lastRainEvent.EndTime, currentWeather, historicalWeather, sunExposure)

	// Adjust required time based on actual drying progress
	effectiveDryingTime := requiredDryingTime * (1.0 - dryingProgress)

	// Check if rock has dried
	if effectiveDryingTime <= 0 {
		return models.RockDryingStatus{
			IsWet:             false,
			IsSafe:            true,
			IsWetSensitive:    hasWetSensitive,
			HoursUntilDry:     0,
			LastRainTimestamp: lastRainEvent.EndTime.Format(time.RFC3339),
			Status:            "good",
			Message:           "Rock is dry and safe to climb",
			RockTypes:         rockTypeNames,
			PrimaryRockType:   primaryRock.Name,
			PrimaryGroupName:  primaryRock.GroupName,
			ConfidenceScore:   calculateConfidence(true, lastRainEvent.EndTime, historicalWeather, hasWetSensitive, sunExposure),
		}
	}

	// Rock is still drying
	status := "fair"
	message := "Rock is drying"

	if effectiveDryingTime > requiredDryingTime*0.5 {
		// More than 50% of drying time remaining
		status = "poor"
		message = "Rock is still wet"
	}

	// Wet-sensitive rocks (sandstone, arkose, graywacke) are ALWAYS critical when wet
	if hasWetSensitive {
		status = "critical"
		message = "DO NOT CLIMB - " + primaryRock.GroupName + " is wet-sensitive and still wet"
	}

	return models.RockDryingStatus{
		IsWet:             true,
		IsSafe:            status != "critical",
		IsWetSensitive:    hasWetSensitive,
		HoursUntilDry:     math.Ceil(effectiveDryingTime),
		LastRainTimestamp: lastRainEvent.EndTime.Format(time.RFC3339),
		Status:            status,
		Message:           message,
		RockTypes:         rockTypeNames,
		PrimaryRockType:   primaryRock.Name,
		PrimaryGroupName:  primaryRock.GroupName,
		ConfidenceScore:   calculateConfidence(false, lastRainEvent.EndTime, historicalWeather, hasWetSensitive, sunExposure),
	}
}

// handleSnowOnGround handles the case when there's snow on the ground
func (c *Calculator) handleSnowOnGround(
	snowDepthInches float64,
	currentWeather *models.WeatherData,
	historicalWeather []models.WeatherData,
	primaryRock models.RockType,
	sunExposure *models.LocationSunExposure,
	hasWetSensitive bool,
	rockTypeNames []string,
) models.RockDryingStatus {
	// Only mark as "critical" for wet-sensitive rocks (sandstone, arkose, graywacke)
	status := "poor"
	message := "Snow on ground - rock may be wet"

	if hasWetSensitive {
		status = "critical"
		message = "DO NOT CLIMB - " + primaryRock.GroupName + " is wet-sensitive and there is snow on ground"
	} else if snowDepthInches > 2.0 {
		message = "Significant snow accumulation on ground"
	}

	// Estimate snow melt time based on conditions
	estimatedDryTime := estimateSnowMeltTime(
		snowDepthInches,
		currentWeather,
		historicalWeather,
		primaryRock,
		sunExposure,
	)

	return models.RockDryingStatus{
		IsWet:             true,
		IsSafe:            false,
		IsWetSensitive:    hasWetSensitive,
		HoursUntilDry:     estimatedDryTime,
		LastRainTimestamp: time.Now().Format(time.RFC3339),
		Status:            status,
		Message:           message,
		RockTypes:         rockTypeNames,
		PrimaryRockType:   primaryRock.Name,
		PrimaryGroupName:  primaryRock.GroupName,
		ConfidenceScore:   95, // High confidence - snow is visible
	}
}

// handleFreezingConditions handles frozen precipitation/ice on rock
func (c *Calculator) handleFreezingConditions(
	currentWeather *models.WeatherData,
	historicalWeather []models.WeatherData,
	primaryRock models.RockType,
	sunExposure *models.LocationSunExposure,
	hasWetSensitive bool,
	rockTypeNames []string,
) *models.RockDryingStatus {
	// Check if there was recent precipitation that's now frozen
	recentPrecip := 0.0
	cutoffTime := time.Now().Add(-48 * time.Hour)

	for _, h := range historicalWeather {
		if h.Timestamp.After(cutoffTime) && h.Precipitation > 0.01 {
			recentPrecip += h.Precipitation
		}
	}

	if recentPrecip > 0.1 {
		// Only mark as "critical" for wet-sensitive rocks (sandstone, arkose, graywacke)
		status := "poor"
		message := "Freezing temps with recent precipitation - ice on rock"

		if hasWetSensitive {
			status = "critical"
			message = "DO NOT CLIMB - " + primaryRock.GroupName + " is wet-sensitive and may have ice"
		}

		// Estimate ice melt time based on conditions
		estimatedDryTime := estimateIceMeltTime(
			recentPrecip,
			currentWeather,
			historicalWeather,
			primaryRock,
			sunExposure,
		)

		return &models.RockDryingStatus{
			IsWet:             true,
			IsSafe:            false,
			IsWetSensitive:    hasWetSensitive,
			HoursUntilDry:     estimatedDryTime,
			LastRainTimestamp: time.Now().Format(time.RFC3339),
			Status:            status,
			Message:           message,
			RockTypes:         rockTypeNames,
			PrimaryRockType:   primaryRock.Name,
			PrimaryGroupName:  primaryRock.GroupName,
			ConfidenceScore:   90,
		}
	}

	return nil
}

// handleCurrentRain handles currently raining conditions
func (c *Calculator) handleCurrentRain(
	currentWeather *models.WeatherData,
	historicalWeather []models.WeatherData,
	primaryRock models.RockType,
	sunExposure *models.LocationSunExposure,
	hasSeepageRisk bool,
	hasWetSensitive bool,
	rockTypeNames []string,
) models.RockDryingStatus {
	status := "poor"
	message := "Currently raining - rock is wet"

	if hasWetSensitive {
		status = "critical"
		message = "DO NOT CLIMB - " + primaryRock.GroupName + " is wet-sensitive and currently raining"
	}

	// Estimate drying time from current rain
	dryingTime := estimateDryingTime(primaryRock, currentWeather, historicalWeather,
		sunExposure, hasSeepageRisk, currentWeather.Precipitation)

	return models.RockDryingStatus{
		IsWet:             true,
		IsSafe:            false,
		IsWetSensitive:    hasWetSensitive,
		HoursUntilDry:     dryingTime,
		LastRainTimestamp: time.Now().Format(time.RFC3339),
		Status:            status,
		Message:           message,
		RockTypes:         rockTypeNames,
		PrimaryRockType:   primaryRock.Name,
		PrimaryGroupName:  primaryRock.GroupName,
		ConfidenceScore:   calculateConfidence(false, time.Now(), historicalWeather, hasWetSensitive, sunExposure),
	}
}

// findLastRainEvent finds the most recent rain event and calculates its characteristics
func findLastRainEvent(historical []models.WeatherData, current *models.WeatherData) *models.RainEvent {
	if len(historical) == 0 && current == nil {
		return nil
	}

	var event *models.RainEvent
	var totalRain float64
	var maxRate float64
	var rainHours int
	var startTime, endTime time.Time

	// Check current weather first (if it's raining NOW)
	if current != nil && current.Precipitation > 0.01 {
		event = &models.RainEvent{
			StartTime: current.Timestamp,
			EndTime:   current.Timestamp,
		}
		endTime = current.Timestamp
		startTime = current.Timestamp
		totalRain = current.Precipitation
		rainHours = 1
		maxRate = current.Precipitation
	}

	// Check historical data (reversed to go from most recent to oldest)
	for i := len(historical) - 1; i >= 0; i-- {
		h := historical[i]
		if h.Precipitation > 0.01 {
			// Part of rain event
			if event == nil {
				// Start of new rain event
				event = &models.RainEvent{
					StartTime: h.Timestamp,
					EndTime:   h.Timestamp,
				}
				endTime = h.Timestamp
			}
			startTime = h.Timestamp
			totalRain += h.Precipitation
			rainHours++

			// Track max hourly rate
			if h.Precipitation > maxRate {
				maxRate = h.Precipitation
			}
		} else if event != nil {
			// End of rain event (found dry period)
			break
		}
	}

	if event == nil {
		return nil
	}

	// Finalize rain event
	event.StartTime = startTime
	event.EndTime = endTime
	event.TotalRain = totalRain
	event.Duration = endTime.Sub(startTime).Hours()
	event.MaxHourlyRate = maxRate

	if rainHours > 0 {
		event.AvgHourlyRate = totalRain / float64(rainHours)
	}

	return event
}

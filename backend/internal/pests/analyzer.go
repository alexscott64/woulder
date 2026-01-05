package pests

import (
	"fmt"
	"math"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// PestAnalyzer calculates pest activity conditions
type PestAnalyzer struct{}

// MosquitoScoreResult contains mosquito score and contributing factors
type MosquitoScoreResult struct {
	Score   int
	Factors []string
}

// CalculateMosquitoScore calculates mosquito activity level based on weather conditions
//
// Key factors from entomological research:
// - Optimal temp: 70-85°F (active 50-95°F, dormant below 50°F)
// - Humidity > 50% preferred (they dehydrate easily)
// - Standing water from rain 7-14 days ago = breeding sites
// - Low wind (< 10 mph) allows flight
// - Season matters: peak activity late spring through early fall
func (a *PestAnalyzer) CalculateMosquitoScore(
	currentTemp float64,
	currentHumidity int,
	currentWind float64,
	recentRainfall float64,
	month time.Month,
) MosquitoScoreResult {
	score := 0
	factors := []string{}

	// Temperature is a GATING factor - mosquitoes are dormant below 50°F
	if currentTemp < 50 {
		factors = append(factors, "Too cold for mosquitoes (below 50°F)")
		return MosquitoScoreResult{
			Score:   int(math.Min(5, math.Round(currentTemp/10))),
			Factors: factors,
		}
	}

	// Temperature factor (0-30 points)
	if currentTemp >= 70 && currentTemp <= 85 {
		score += 30
		factors = append(factors, "Optimal mosquito temperature")
	} else if currentTemp >= 60 && currentTemp < 70 {
		score += 20
		factors = append(factors, "Warm enough for mosquito activity")
	} else if currentTemp >= 85 && currentTemp <= 95 {
		score += 20
		factors = append(factors, "Hot but mosquitoes still active")
	} else if currentTemp >= 50 && currentTemp < 60 {
		score += 10
		factors = append(factors, "Cool - reduced mosquito activity")
	} else if currentTemp > 95 {
		score += 5
		factors = append(factors, "Too hot - mosquitoes seek shade")
	}

	// Humidity factor (0-25 points)
	if currentHumidity >= 70 {
		score += 25
		factors = append(factors, "High humidity favors mosquitoes")
	} else if currentHumidity >= 50 {
		score += 15
		factors = append(factors, "Moderate humidity")
	} else if currentHumidity >= 30 {
		score += 5
		factors = append(factors, "Low humidity limits mosquitoes")
	} else {
		score += 0
		factors = append(factors, "Very dry - mosquitoes dehydrate")
	}

	// Recent rainfall factor (0-25 points)
	if recentRainfall >= 2 {
		score += 25
		factors = append(factors, "Recent heavy rain created breeding sites")
	} else if recentRainfall >= 1 {
		score += 20
		factors = append(factors, "Recent rain provides breeding habitat")
	} else if recentRainfall >= 0.5 {
		score += 12
		factors = append(factors, "Some recent moisture")
	} else if recentRainfall >= 0.1 {
		score += 5
		factors = append(factors, "Minimal recent rainfall")
	} else {
		score += 0
		factors = append(factors, "Dry conditions limit breeding")
	}

	// Wind factor (0-10 points)
	if currentWind <= 5 {
		score += 10
		factors = append(factors, "Calm conditions favor mosquitoes")
	} else if currentWind <= 10 {
		score += 6
		factors = append(factors, "Light wind")
	} else if currentWind <= 15 {
		score += 3
		factors = append(factors, "Moderate wind limits flight")
	} else {
		score += 0
		factors = append(factors, "Strong wind grounds mosquitoes")
	}

	// Seasonal factor (0-10 points)
	seasonalMultiplier := GetSeasonalMosquitoFactor(month)
	score += int(seasonalMultiplier * 10)

	if seasonalMultiplier >= 0.8 {
		factors = append(factors, "Peak mosquito season")
	} else if seasonalMultiplier >= 0.5 {
		factors = append(factors, "Active mosquito season")
	} else if seasonalMultiplier >= 0.2 {
		factors = append(factors, "Early/late season")
	} else {
		factors = append(factors, "Off-season for mosquitoes")
	}

	finalScore := int(math.Min(100, math.Max(0, float64(score))))
	return MosquitoScoreResult{Score: finalScore, Factors: factors}
}

// CalculateOutdoorPestScore calculates general outdoor pest activity
// Determines threat from flies, gnats, wasps, bees, ants, spiders, etc.
func (a *PestAnalyzer) CalculateOutdoorPestScore(
	currentTemp float64,
	currentHumidity int,
	recentRainfall float64,
	month time.Month,
) MosquitoScoreResult {
	score := 0
	factors := []string{}

	// Temperature is a GATING factor
	if currentTemp < 50 {
		factors = append(factors, "Too cold for most insects (below 50°F)")
		return MosquitoScoreResult{
			Score:   int(math.Min(8, math.Round(currentTemp/6))),
			Factors: factors,
		}
	}

	// Temperature factor (0-40 points)
	if currentTemp >= 75 && currentTemp <= 90 {
		score += 40
		factors = append(factors, "Optimal temperature for insect activity")
	} else if currentTemp >= 65 && currentTemp < 75 {
		score += 30
		factors = append(factors, "Warm - good insect activity")
	} else if currentTemp >= 90 && currentTemp <= 100 {
		score += 30
		factors = append(factors, "Hot - high insect activity")
	} else if currentTemp >= 55 && currentTemp < 65 {
		score += 15
		factors = append(factors, "Cool - reduced insect activity")
	} else if currentTemp >= 50 && currentTemp < 55 {
		score += 8
		factors = append(factors, "Near insect activity threshold")
	} else if currentTemp > 100 {
		score += 15
		factors = append(factors, "Extreme heat - insects seek shade")
	}

	// Humidity factor (0-20 points)
	if currentHumidity >= 60 && currentHumidity <= 80 {
		score += 20
		factors = append(factors, "Ideal humidity for insects")
	} else if currentHumidity > 80 {
		score += 15
		factors = append(factors, "High humidity")
	} else if currentHumidity >= 40 && currentHumidity < 60 {
		score += 12
		factors = append(factors, "Moderate humidity")
	} else {
		score += 5
		factors = append(factors, "Dry conditions")
	}

	// Recent rainfall factor (0-20 points)
	if recentRainfall >= 1 && recentRainfall < 3 {
		score += 20
		factors = append(factors, "Recent rain increased pest breeding")
	} else if recentRainfall >= 0.5 && recentRainfall < 1 {
		score += 15
		factors = append(factors, "Some moisture aids pest activity")
	} else if recentRainfall >= 3 {
		score += 12
		factors = append(factors, "Heavy rain - mixed pest effects")
	} else {
		score += 8
		factors = append(factors, "Dry period")
	}

	// Seasonal factor (0-20 points)
	seasonalMultiplier := GetSeasonalPestFactor(month)
	score += int(seasonalMultiplier * 20)

	if seasonalMultiplier >= 0.8 {
		factors = append(factors, "Peak pest season")
	} else if seasonalMultiplier >= 0.5 {
		factors = append(factors, "Active pest season")
	} else if seasonalMultiplier >= 0.2 {
		factors = append(factors, "Early/late pest season")
	} else {
		factors = append(factors, "Low pest season")
	}

	finalScore := int(math.Min(100, math.Max(0, float64(score))))
	return MosquitoScoreResult{Score: finalScore, Factors: factors}
}

// AssessConditions calculates comprehensive pest conditions
func (a *PestAnalyzer) AssessConditions(
	current *models.WeatherData,
	historical []models.WeatherData,
) PestConditions {
	if current == nil {
		return PestConditions{
			MosquitoLevel:    PestLevelLow,
			OutdoorPestLevel: PestLevelLow,
		}
	}

	now := time.Now()
	month := now.Month()

	// Calculate recent rainfall (last 7 days for mosquito breeding)
	recentRainfall := CalculateRecentRainfall(historical, 7)

	// Calculate mosquito conditions
	mosquitoResult := a.CalculateMosquitoScore(
		current.Temperature,
		current.Humidity,
		current.WindSpeed,
		recentRainfall,
		month,
	)

	// Calculate outdoor pest conditions
	pestResult := a.CalculateOutdoorPestScore(
		current.Temperature,
		current.Humidity,
		recentRainfall,
		month,
	)

	// Combine unique factors (limit to top 4)
	factorMap := make(map[string]bool)
	uniqueFactors := []string{}

	for _, f := range append(mosquitoResult.Factors, pestResult.Factors...) {
		if !factorMap[f] && len(uniqueFactors) < 4 {
			factorMap[f] = true
			uniqueFactors = append(uniqueFactors, f)
		}
	}

	return PestConditions{
		MosquitoLevel:    ScoreToLevel(mosquitoResult.Score),
		MosquitoScore:    mosquitoResult.Score,
		OutdoorPestLevel: ScoreToLevel(pestResult.Score),
		OutdoorPestScore: pestResult.Score,
		Factors:          uniqueFactors,
	}
}

// String returns a human-readable description of the pest level
func (l PestLevel) String() string {
	switch l {
	case PestLevelLow:
		return "Low"
	case PestLevelModerate:
		return "Moderate"
	case PestLevelHigh:
		return "High"
	case PestLevelVeryHigh:
		return "Very High"
	case PestLevelExtreme:
		return "Extreme"
	default:
		return "Unknown"
	}
}

// MarshalJSON implements custom JSON marshaling for PestLevel
func (l PestLevel) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, l)), nil
}

package rock_temp

import (
	"fmt"
	"math"

	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/alexscott64/woulder/backend/internal/weather/sun"
)

// calculator.go is the orchestrator that wires every helper in this
// package together to produce a complete RockTemperatureStatus from a
// single Inputs bundle. See plans/rock-temperature-UPDATE.md sections 5
// and 6 for the full data flow.

// Calculator is the public entry point. It is stateless; constructing
// one is free.
type Calculator struct{}

// Calculate produces a complete RockTemperatureStatus from an Inputs
// bundle. Steps performed:
//
//  1. Resolve thermal params from RockTypeGroup (granite fallback).
//  2. Resolve dominant facet (aspect, dip) from sun-exposure profile.
//  3. Tree fraction from sun exposure (0 if missing).
//  4. Concatenate past + forecast hours; degrade gracefully on empty.
//  5. Build a per-hour T_eq closure that captures location, params,
//     facet, and tree fraction; apply elevation corrections to
//     irradiance and sky temperature.
//  6. Run thermal-lag integration with 12h spin-up.
//  7. Locate the "now" index in the concatenated slice.
//  8. Build the future hourly forecast slice (excluding heavy
//     condensation hours when feeding DetectSendWindows).
//  9. Compute current condition + condensation severity + friction.
//  10. Build the CondensationInfo struct (always populated).
//  11. Detect send windows and the next condition transition.
//  12. Compute confidence + factor strings (with cloud/wind variability).
//  13. Build a short user-facing message.
//  14. Return the assembled status.
func (c *Calculator) Calculate(in Inputs) models.RockTemperatureStatus {
	// 1. Thermal params.
	params, paramsKnown := ParamsForGroup(in.RockTypeGroup)
	if in.RockTypeGroup == "" {
		paramsKnown = false
	}

	// 2. Dominant facet.
	facet := ResolveDominantFacet(in.SunExposure)
	aspectDeg := facet.AspectDeg
	dipDeg := facet.DipDeg

	// 3. Tree fraction.
	tree := TreeFraction(in.SunExposure)

	// 4. Concatenate past + forecast.
	all := make([]models.WeatherData, 0, len(in.PastHourly)+len(in.Forecast))
	all = append(all, in.PastHourly...)
	all = append(all, in.Forecast...)
	if len(all) == 0 {
		return models.RockTemperatureStatus{
			Message:           "insufficient weather data for rock temp calculation",
			ConfidenceScore:   20,
			ConfidenceFactors: []string{"no weather data available"},
			RockType:          params.GroupName,
		}
	}

	// 5. T_eq closure.
	elevKm := 0.0
	if in.Location != nil {
		elevKm = float64(in.Location.ElevationFt) * 0.0003048
	}
	irradianceMult := 1.0 + 0.05*elevKm

	var lat, lon float64
	if in.Location != nil {
		lat = in.Location.Latitude
		lon = in.Location.Longitude
	}

	computeEq := func(_ int, w models.WeatherData) float64 {
		sunPos := sun.Calculate(lat, lon, w.Timestamp)

		directH, diffuseH := w.DirectRadiation, w.DiffuseRadiation
		if directH == 0 && diffuseH == 0 && w.ShortwaveRadiation > 0 {
			directH, diffuseH = SplitShortwave(w.ShortwaveRadiation, float64(w.CloudCover))
		}
		// Elevation correction on irradiance.
		directH *= irradianceMult
		diffuseH *= irradianceMult

		dni := DNIFromHorizontal(directH, sunPos.Elevation)
		geom := GeometricFactor(sunPos.Azimuth, sunPos.Elevation, aspectDeg, dipDeg)
		skyView := SkyViewFactor(dipDeg)
		faceIrr := FaceIrradiance(dni, diffuseH, geom, skyView, tree)

		// Sky temperature with extra clear-night cooling at altitude.
		cloudFrac := float64(w.CloudCover) / 100.0
		baseSkyF := SkyTemperatureF(w.Temperature, float64(w.CloudCover))
		extraSkyCoolingF := 3.0 * elevKm * (1 - cloudFrac)
		skyF := baseSkyF - extraSkyCoolingF

		hConv := ConvectiveCoeff(w.WindSpeed)
		hRad := RadiativeCoeff()
		return EquilibriumTempF(w.Temperature, skyF, params.Absorptivity, faceIrr, hConv, hRad)
	}

	// 6. Run integration.
	rockTemps := SpinUpAndIntegrate(in.PastHourly, in.Forecast, computeEq, params.TauMinutes)

	// 7. Locate "now".
	nowIdx := len(in.PastHourly)
	if in.Now != nil {
		found := -1
		for i, w := range all {
			if !w.Timestamp.Before(in.Now.Timestamp) {
				found = i
				break
			}
		}
		if found >= 0 {
			nowIdx = found
		}
	}
	if nowIdx >= len(all) {
		nowIdx = len(all) - 1
	}
	if nowIdx < 0 {
		nowIdx = 0
	}

	// 8. Build hourly forecast (future portion: nowIdx..end).
	future := all[nowIdx:]
	hourlyForecast := make([]models.RockTempHour, 0, len(future))
	windowEligibleHours := make([]models.RockTempHour, 0, len(future))
	for i, w := range future {
		surfF := rockTemps[nowIdx+i]
		condSev := ClassifyCondensation(surfF, w.DewpointF)
		cond := ClassifyTempCondition(surfF, params.Thresholds)
		// `Condensing` flag is true only for light condensation; heavy
		// hours are filtered out of the window-eligibility slice and
		// "none" hours leave Condensing=false (so DryThroughout stays
		// true). This matches the plan's send-window rules.
		condensing := condSev == "light"
		hr := models.RockTempHour{
			Time:       w.Timestamp,
			SurfaceF:   round1(surfF),
			AirF:       round1(w.Temperature),
			DewpointF:  round1(w.DewpointF),
			Condensing: condensing,
			Condition:  cond,
		}
		hourlyForecast = append(hourlyForecast, hr)
		if condSev != "heavy" {
			windowEligibleHours = append(windowEligibleHours, hr)
		}
	}

	// 9. Current condition.
	currentSurfF := rockTemps[nowIdx]
	currentAirF := all[nowIdx].Temperature
	currentDewF := all[nowIdx].DewpointF
	currentWind := all[nowIdx].WindSpeed
	currentHumidity := float64(all[nowIdx].Humidity)
	tempCondition := ClassifyTempCondition(currentSurfF, params.Thresholds)
	condensationSeverity := ClassifyCondensation(currentSurfF, currentDewF)
	friction := ComputeFrictionQuality(tempCondition, condensationSeverity)

	// 10. Condensation info struct (always populated).
	condInfo := models.CondensationInfo{
		Active:            condensationSeverity == "heavy",
		DewpointF:         round1(currentDewF),
		SurfaceVsDewpoint: round1(currentSurfF - currentDewF),
		Severity:          condensationSeverity,
		Reason:            CondensationReason(currentSurfF, currentDewF, currentAirF, currentWind, currentHumidity),
	}
	if condensationSeverity != "none" {
		// Build HourPoint slice from the future portion to find clears-at.
		points := make([]HourPoint, 0, len(future))
		for i, w := range future {
			points = append(points, HourPoint{
				Time:      w.Timestamp,
				SurfaceF:  rockTemps[nowIdx+i],
				DewpointF: w.DewpointF,
			})
		}
		condInfo.ClearsAt = FindClearsAt(points, -1)
	}

	// 11. Send windows + next transition.
	sendWindows := DetectSendWindows(windowEligibleHours, SendWindowOptions{})
	next := NextTransition(hourlyForecast, tempCondition)

	// Aggregate hourly forecast into per-day summaries for the ForecastView.
	// Use the location's local timezone if available (lat-based fallback to UTC for
	// now; weather_service can later pass a real IANA tz string).
	tzName := "" // TODO: wire from in.Location once timezone field is available
	dailyForecast := AggregateDaily(hourlyForecast, sendWindows, tzName)

	// 12. Confidence.
	aspectKnown := in.SunExposure != nil && !facet.Mixed
	dipKnown := in.SunExposure != nil && !facet.Mixed
	confIn := ConfidenceInputs{
		AspectKnown:       aspectKnown,
		DipKnown:          dipKnown,
		RockTypeKnown:     paramsKnown,
		MixedFacets:       facet.Mixed,
		ForecastHorizonH:  len(in.Forecast),
		CloudVariableHigh: cloudVariability(all) > 25.0,
		WindVariableHigh:  windVariability(all) > 5.0,
		SpinUpComplete:    len(in.PastHourly) >= 6,
	}
	conf := ComputeConfidence(confIn)
	// Append facet reason if not already implied.
	if facet.Reason != "" && facet.Mixed {
		// MixedFacets factor is already "mixed facets; result averaged across multiple aspects" — append a finer reason.
		conf.Factors = append(conf.Factors, facet.Reason)
	}

	// 13. Message.
	message := buildMessage(tempCondition, condensationSeverity, currentSurfF, currentAirF)

	// 14. Return.
	return models.RockTemperatureStatus{
		EstimatedSurfaceTempF: round1(currentSurfF),
		AirTempF:              round1(currentAirF),
		TempDifferentialF:     round1(currentSurfF - currentAirF),
		Condition:             tempCondition,
		FrictionQuality:       friction,
		NextTransition:        next,
		Message:               message,
		SendWindows:           sendWindows,
		HourlyForecast:        hourlyForecast,
		DailyForecast:         dailyForecast,
		Condensation:          &condInfo,
		ConfidenceScore:       conf.Score,
		ConfidenceFactors:     conf.Factors,
		RockType:              params.GroupName,
	}
}

// round1 rounds to one decimal place.
func round1(x float64) float64 {
	return math.Round(x*10) / 10
}

// cloudVariability returns the standard deviation of CloudCover (0..100)
// across the supplied slice. Empty/single-element slices return 0.
func cloudVariability(all []models.WeatherData) float64 {
	if len(all) < 2 {
		return 0
	}
	var sum float64
	for _, w := range all {
		sum += float64(w.CloudCover)
	}
	mean := sum / float64(len(all))
	var sq float64
	for _, w := range all {
		d := float64(w.CloudCover) - mean
		sq += d * d
	}
	return math.Sqrt(sq / float64(len(all)))
}

// windVariability returns the standard deviation of WindSpeed (mph)
// across the supplied slice. Empty/single-element slices return 0.
func windVariability(all []models.WeatherData) float64 {
	if len(all) < 2 {
		return 0
	}
	var sum float64
	for _, w := range all {
		sum += w.WindSpeed
	}
	mean := sum / float64(len(all))
	var sq float64
	for _, w := range all {
		d := w.WindSpeed - mean
		sq += d * d
	}
	return math.Sqrt(sq / float64(len(all)))
}

// buildMessage produces a short human-readable summary string keyed off
// (tempCondition, condensationSeverity). Heavy condensation always wins
// the headline because wet rock is unclimbable regardless of temp.
func buildMessage(tempCondition, condensationSeverity string, surfF, airF float64) string {
	if condensationSeverity == "heavy" {
		return "Wet rock surface (condensation) — friction is poor"
	}
	diff := surfF - airF
	switch tempCondition {
	case "very_poor":
		return fmt.Sprintf("Rock at %.0f°F (%.0f above air) — friction is poor", surfF, math.Abs(diff))
	case "poor":
		if diff >= 0 {
			return fmt.Sprintf("Hot rock at %.0f°F (%.0f above air) — friction is poor", surfF, diff)
		}
		return fmt.Sprintf("Cold rock at %.0f°F — friction is poor", surfF)
	case "marginal":
		if condensationSeverity == "light" {
			return fmt.Sprintf("Marginal rock temp at %.0f°F with damp surface", surfF)
		}
		return fmt.Sprintf("Marginal rock temp at %.0f°F", surfF)
	case "good":
		if condensationSeverity == "light" {
			return fmt.Sprintf("Good rock temp at %.0f°F (damp surface)", surfF)
		}
		return fmt.Sprintf("Good rock temp at %.0f°F", surfF)
	case "prime":
		if condensationSeverity == "light" {
			return fmt.Sprintf("Prime rock temp at %.0f°F (surface damp)", surfF)
		}
		return fmt.Sprintf("Prime rock temp at %.0f°F", surfF)
	case "too_cold":
		return fmt.Sprintf("Rock too cold at %.0f°F — friction is poor", surfF)
	}
	return fmt.Sprintf("Rock at %.0f°F", surfF)
}

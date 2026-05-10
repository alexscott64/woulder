package rock_temp

import "time"

// HourPoint is a small forecast hour record used by the condensation
// scanner. It is exported so the calculator orchestrator (subtask 3)
// can build slices of it from []models.RockTempHour without importing a
// private struct.
type HourPoint struct {
	Time      time.Time
	SurfaceF  float64
	DewpointF float64
}

// ClassifyCondensation returns the condensation severity for the given
// surface and dewpoint temperatures (°F):
//
//	"none"  if surfaceF - dewpointF >  2
//	"light" if 0  <  surfaceF - dewpointF <= 2
//	"heavy" if         surfaceF - dewpointF <= 0
//
// Boundary semantics: exactly +2°F → "light"; exactly 0°F → "heavy".
func ClassifyCondensation(surfaceF, dewpointF float64) string {
	diff := surfaceF - dewpointF
	switch {
	case diff > 2:
		return "none"
	case diff > 0:
		return "light"
	default:
		return "heavy"
	}
}

// FindClearsAt scans forward from startIdx+1 for the first hour where
// surfaceF >= dewpointF + 2.0 (a 2°F hysteresis margin to avoid
// flapping). Returns a pointer to the time of that hour, or nil if the
// condition never clears within the supplied slice.
func FindClearsAt(hours []HourPoint, startIdx int) *time.Time {
	if startIdx < -1 {
		startIdx = -1
	}
	for i := startIdx + 1; i < len(hours); i++ {
		if hours[i].SurfaceF >= hours[i].DewpointF+2.0 {
			t := hours[i].Time
			return &t
		}
	}
	return nil
}

// CondensationReason returns a short human-readable string explaining
// why condensation is (or could be) occurring. Selection order:
//
//   - "Air at saturation (fog conditions)" if humidityPct >= 99 OR
//     air-dewpoint spread <= 1.5°F.
//   - "Rock cooled by wind below dewpoint" if wind >= 8 mph AND surface
//     is below dewpoint AND air-dewpoint spread > 3°F.
//   - "Cold rock surface + humid air" otherwise (the typical predawn
//     case).
func CondensationReason(surfaceF, dewpointF, airF, windSpeedMph, humidityPct float64) string {
	if humidityPct >= 99 || airF-dewpointF <= 1.5 {
		return "Air at saturation (fog conditions)"
	}
	if windSpeedMph >= 8 && surfaceF < dewpointF && airF-dewpointF > 3 {
		return "Rock cooled by wind below dewpoint"
	}
	return "Cold rock surface + humid air"
}

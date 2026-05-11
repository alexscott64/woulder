package rock_temp

import (
	"fmt"
	"math"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// Inputs is the bundle of data the rock_temp Calculator (subtask 3)
// consumes. Subtask 2 only defines the type and the helpers below; the
// actual orchestration lives in calculator.go (subtask 3).
type Inputs struct {
	RockTypeGroup string                      // canonical group name from rock_type_groups
	SunExposure   *models.LocationSunExposure // may be nil
	Location      *models.Location            // for elevation, lat, lon
	PastHourly    []models.WeatherData        // ~12h pre-now for spin-up
	Forecast      []models.WeatherData        // future hours
	Now           *models.WeatherData         // current observation
	// TimezoneName is the IANA timezone name (e.g. "America/Los_Angeles")
	// used by send-window midnight-splitting and per-day aggregation.
	// Empty string is acceptable and downstream callers (DetectSendWindows,
	// AggregateDaily) gracefully fall back to UTC.
	TimezoneName string
}

// WeightedAspect computes a single representative aspect angle (degrees
// from north, clockwise) from the four face-percentage breakdowns. The
// result is the azimuth of the unit-vector sum:
//
//	x = e·sin(90°) + s·sin(180°) + w·sin(270°) + n·sin(0°)
//	y = e·cos(90°) + s·cos(180°) + w·cos(270°) + n·cos(0°)
//	az = atan2(x, y), normalized to [0, 360)
//
// Returns ok=false when the resultant vector is essentially zero
// (truly omnidirectional, e.g., 50% E + 50% W); callers should handle
// that case by treating the location as multi-facet.
func WeightedAspect(southPct, westPct, eastPct, northPct float64) (float64, bool) {
	x := eastPct*math.Sin(deg2rad(90)) +
		southPct*math.Sin(deg2rad(180)) +
		westPct*math.Sin(deg2rad(270)) +
		northPct*math.Sin(deg2rad(0))
	y := eastPct*math.Cos(deg2rad(90)) +
		southPct*math.Cos(deg2rad(180)) +
		westPct*math.Cos(deg2rad(270)) +
		northPct*math.Cos(deg2rad(0))
	if math.Abs(x) < 1e-9 && math.Abs(y) < 1e-9 {
		return 0, false
	}
	az := math.Atan2(x, y) * 180 / math.Pi
	if az < 0 {
		az += 360
	}
	return az, true
}

// WeightedDip computes a single representative dip angle (degrees) from
// the slab/overhang percentages. Vertical is implied as the remainder.
// Representative angles: slab=45°, vertical=90°, overhang=110°.
//
//	vertical = clamp(100 - slab - overhang, 0, 100)
//	dip      = (slab*45 + vertical*90 + overhang*110) / 100
func WeightedDip(slabPct, overhangPct float64) float64 {
	if slabPct < 0 {
		slabPct = 0
	}
	if overhangPct < 0 {
		overhangPct = 0
	}
	vertical := 100 - slabPct - overhangPct
	if vertical < 0 {
		vertical = 0
	} else if vertical > 100 {
		vertical = 100
	}
	return (slabPct*45 + vertical*90 + overhangPct*110) / 100
}

// DominantFacet captures the single representative aspect/dip used for
// the energy balance, plus a flag and reason string indicating whether
// the location is genuinely mixed (no facet dominates).
type DominantFacet struct {
	AspectDeg float64
	DipDeg    float64
	Mixed     bool
	Reason    string
}

// ResolveDominantFacet returns a DominantFacet for the given sun-
// exposure profile, applying the plan's "Practical recommendation":
//
//   - If any face percentage exceeds 60%, that face is the dominant
//     aspect (S=180, W=270, E=90, N=0). Otherwise the location is
//     mixed and the weighted aspect is used (or 180° if degenerate).
//   - Same logic for dip: slab>60 → 45°, overhang>60 → 110°, else
//     WeightedDip.
//   - Nil sun exposure defaults to south-facing vertical and is
//     marked Mixed=true.
func ResolveDominantFacet(se *models.LocationSunExposure) DominantFacet {
	if se == nil {
		return DominantFacet{
			AspectDeg: 180,
			DipDeg:    90,
			Mixed:     true,
			Reason:    "no sun exposure data; defaulted to south-facing vertical",
		}
	}

	mixed := false
	var aspectReason, dipReason string

	// Aspect resolution.
	var aspectDeg float64
	switch {
	case se.SouthFacingPercent > 60:
		aspectDeg = 180
		aspectReason = "south-facing dominant"
	case se.WestFacingPercent > 60:
		aspectDeg = 270
		aspectReason = "west-facing dominant"
	case se.EastFacingPercent > 60:
		aspectDeg = 90
		aspectReason = "east-facing dominant"
	case se.NorthFacingPercent > 60:
		aspectDeg = 0
		aspectReason = "north-facing dominant"
	default:
		mixed = true
		az, ok := WeightedAspect(
			se.SouthFacingPercent, se.WestFacingPercent,
			se.EastFacingPercent, se.NorthFacingPercent,
		)
		if !ok {
			aspectDeg = 180
			aspectReason = "mixed facets, weighted aspect degenerate; defaulted to south"
		} else {
			aspectDeg = az
			aspectReason = "mixed facets; using weighted aspect"
		}
	}

	// Dip resolution.
	var dipDeg float64
	switch {
	case se.SlabPercent > 60:
		dipDeg = 45
		dipReason = "slab dominant"
	case se.OverhangPercent > 60:
		dipDeg = 110
		dipReason = "overhang dominant"
	default:
		// Treat anything not dominated by slab/overhang as mixed too if
		// neither slab nor overhang is the clear majority.
		if se.SlabPercent >= 30 || se.OverhangPercent >= 30 {
			mixed = true
		}
		dipDeg = WeightedDip(se.SlabPercent, se.OverhangPercent)
		dipReason = "mixed dip; using weighted average"
	}

	return DominantFacet{
		AspectDeg: aspectDeg,
		DipDeg:    dipDeg,
		Mixed:     mixed,
		Reason:    fmt.Sprintf("%s; %s", aspectReason, dipReason),
	}
}

// TreeFraction returns the location's tree coverage as a 0..1 fraction
// (clamped). Returns 0 when sun exposure data is missing.
func TreeFraction(se *models.LocationSunExposure) float64 {
	if se == nil {
		return 0
	}
	f := se.TreeCoveragePercent / 100
	if f < 0 {
		return 0
	}
	if f > 1 {
		return 1
	}
	return f
}

// ResolveRockTypeGroup picks the rock type group name for the energy
// balance. Resolution order:
//
//  1. If override is non-empty, use it (confident=true).
//  2. Else if rockTypes is non-empty and the first row has a non-empty
//     GroupName, use it (confident=true). The repository orders rows by
//     is_primary DESC so [0] is the primary when one is set.
//  3. Else fall back to "Granite" (confident=false).
func ResolveRockTypeGroup(rockTypes []models.RockType, override string) (string, bool) {
	if override != "" {
		return override, true
	}
	if len(rockTypes) > 0 && rockTypes[0].GroupName != "" {
		return rockTypes[0].GroupName, true
	}
	return "Granite", false
}

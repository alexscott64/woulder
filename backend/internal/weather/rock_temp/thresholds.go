// Package rock_temp implements the rock surface temperature and friction
// condition calculator. This file contains rock-type-specific thermal
// constants and the friction-quality temperature thresholds used to
// classify a surface temperature into a tier (prime/good/marginal/...).
package rock_temp

import "strings"

// ThermalParams holds rock-type-specific physics constants used by the
// energy balance and thermal lag models, plus the friction temperature
// thresholds.
type ThermalParams struct {
	GroupName    string  // canonical name returned by rock_type_groups.group_name
	Absorptivity float64 // alpha, solar absorptivity 0..1
	TauMinutes   float64 // thermal time constant, minutes
	Thresholds   ConditionThresholds
}

// ConditionThresholds defines surface temperature boundaries (°F) for
// classifying friction quality tiers. Boundaries are inclusive on the
// lower side per row:
//
//	Prime    = [PrimeMin, PrimeMax)
//	Good     = [PrimeMax, GoodMax)
//	Marginal = [GoodMax, MarginalMax)
//	Poor     = [MarginalMax, PoorMax)
//	VeryPoor = >= PoorMax
//	TooCold  = < TooColdMax (strict)
type ConditionThresholds struct {
	TooColdMax  float64
	PrimeMin    float64
	PrimeMax    float64
	GoodMax     float64
	MarginalMax float64
	PoorMax     float64
}

// Granite is the default thermal profile and also the fallback when the
// rock type group is unknown.
var Granite = ThermalParams{
	GroupName:    "Granite",
	Absorptivity: 0.70,
	TauMinutes:   105,
	Thresholds: ConditionThresholds{
		TooColdMax: 30, PrimeMin: 35, PrimeMax: 55, GoodMax: 65, MarginalMax: 72, PoorMax: 85,
	},
}

// Sandstone — porous, low thermal mass, dark desert varnish high alpha.
var Sandstone = ThermalParams{
	GroupName:    "Sandstone",
	Absorptivity: 0.75,
	TauMinutes:   50,
	Thresholds: ConditionThresholds{
		TooColdMax: 25, PrimeMin: 30, PrimeMax: 50, GoodMax: 60, MarginalMax: 68, PoorMax: 80,
	},
}

// BasaltGabbro — dark, high absorptivity, granite-like thermal mass.
var BasaltGabbro = ThermalParams{
	GroupName:    "Basalt/Gabbro",
	Absorptivity: 0.90,
	TauMinutes:   105,
	Thresholds: ConditionThresholds{
		TooColdMax: 30, PrimeMin: 35, PrimeMax: 55, GoodMax: 63, MarginalMax: 70, PoorMax: 85,
	},
}

// Limestone — lighter color, lower alpha, mid-range thermal mass.
var Limestone = ThermalParams{
	GroupName:    "Limestone",
	Absorptivity: 0.60,
	TauMinutes:   75,
	Thresholds: ConditionThresholds{
		TooColdMax: 30, PrimeMin: 40, PrimeMax: 60, GoodMax: 68, MarginalMax: 75, PoorMax: 85,
	},
}

// Quartzite — granite-like alpha and tau, slightly different temp tolerances.
var Quartzite = ThermalParams{
	GroupName:    "Quartzite",
	Absorptivity: 0.70,
	TauMinutes:   105,
	Thresholds: ConditionThresholds{
		TooColdMax: 30, PrimeMin: 35, PrimeMax: 58, GoodMax: 67, MarginalMax: 74, PoorMax: 85,
	},
}

// DefaultThermalParams returns the granite profile, used as a fallback
// when the rock type group is unknown.
func DefaultThermalParams() ThermalParams {
	return Granite
}

// ParamsForGroup returns the thermal profile for the given canonical
// rock_type_groups.group_name. Matching is case-insensitive and accepts
// common variations such as "basalt", "gabbro", and "basalt/gabbro" all
// mapping to the BasaltGabbro entry. The boolean is false when the name
// did not match a known family (in which case Granite is returned).
func ParamsForGroup(groupName string) (ThermalParams, bool) {
	n := strings.ToLower(strings.TrimSpace(groupName))
	switch n {
	case "granite":
		return Granite, true
	case "sandstone":
		return Sandstone, true
	case "limestone":
		return Limestone, true
	case "quartzite":
		return Quartzite, true
	case "basalt", "gabbro", "basalt/gabbro", "basalt / gabbro", "basalt-gabbro":
		return BasaltGabbro, true
	}
	// Substring fallback for variants like "Basalt (columnar)" etc.
	if strings.Contains(n, "basalt") || strings.Contains(n, "gabbro") {
		return BasaltGabbro, true
	}
	if strings.Contains(n, "sandstone") {
		return Sandstone, true
	}
	if strings.Contains(n, "limestone") {
		return Limestone, true
	}
	if strings.Contains(n, "quartzite") {
		return Quartzite, true
	}
	if strings.Contains(n, "granite") {
		return Granite, true
	}
	return Granite, false
}

// ClassifyTempCondition returns the friction tier for the given surface
// temperature (°F) under the supplied thresholds. See ConditionThresholds
// for the exact boundary semantics.
func ClassifyTempCondition(surfaceF float64, t ConditionThresholds) string {
	switch {
	case surfaceF < t.TooColdMax:
		return "too_cold"
	case surfaceF >= t.PrimeMin && surfaceF < t.PrimeMax:
		return "prime"
	case surfaceF >= t.PrimeMax && surfaceF < t.GoodMax:
		return "good"
	case surfaceF >= t.GoodMax && surfaceF < t.MarginalMax:
		return "marginal"
	case surfaceF >= t.MarginalMax && surfaceF < t.PoorMax:
		return "poor"
	case surfaceF >= t.PoorMax:
		return "very_poor"
	}
	// Gap between TooColdMax and PrimeMin (e.g., 30..35 for granite):
	// classify as too_cold since the rock isn't yet in the prime band.
	return "too_cold"
}

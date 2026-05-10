package rock_temp

// surface_temp.go implements the linearized energy-balance model used to
// estimate the equilibrium rock surface temperature for a single hour.
//
// All temperatures in this package are tracked in Fahrenheit (°F) to
// match the rest of the codebase. The energy-balance coefficients
// (h_conv, h_rad) are physically defined in W/(m²·K). Since a 1 K
// temperature differential equals a 1 °C differential equals a (1/1.8)
// °F differential, conversions to/from °F differentials use the 1.8
// factor. Absolute temperatures are never converted to Kelvin in this
// file — only differentials matter for the linearized formula.

// mphToMps converts wind speed from miles per hour to meters per second.
const mphToMps = 0.44704

// SkyTemperatureF computes the effective sky temperature in °F based on
// cloud cover. The plan's piecewise model:
//
//	Clear (cloud < 20%):    T_sky ≈ T_air - 20°F
//	Overcast (cloud > 80%): T_sky ≈ T_air
//	In between:             linear interpolation between those endpoints.
//
// Cloud cover values outside [0, 100] are clamped before interpolation.
func SkyTemperatureF(airTempF, cloudCoverPct float64) float64 {
	c := cloudCoverPct
	if c < 0 {
		c = 0
	} else if c > 100 {
		c = 100
	}
	var deltaF float64
	switch {
	case c < 20:
		deltaF = 20.0
	case c > 80:
		deltaF = 0.0
	default:
		// (20, 20) → (80, 0). Slope = -20/60 per percentage point.
		deltaF = 20.0 - (c-20.0)/(80.0-20.0)*20.0
	}
	return airTempF - deltaF
}

// ConvectiveCoeff returns the convective heat-transfer coefficient
// h_conv = 5.7 + 3.8·v (W/(m²·K)), where v is wind speed in m/s.
//
// Input is in mph (codebase convention) and converted internally.
func ConvectiveCoeff(windMph float64) float64 {
	if windMph < 0 {
		windMph = 0
	}
	return 5.7 + 3.8*(windMph*mphToMps)
}

// RadiativeCoeff returns the linearized sky-radiation coefficient h_rad
// in W/(m²·K). Approximately 5.5 W/(m²·K) for an emissivity of ε ≈ 0.9
// and typical earth-surface temperatures. Constant for now; if a future
// subtask wants an emissivity-aware value, it can take an argument.
func RadiativeCoeff() float64 {
	return 5.5
}

// EquilibriumTempF computes the equilibrium rock surface temperature
// (°F) for one hour using the linearized energy balance with sky
// radiative loss:
//
//	T_eq = T_air + ( α·I_face - h_rad·(T_air - T_sky) ) / (h_conv + h_rad)
//
// where the bracketed numerator is in W/m² and the denominator is in
// W/(m²·K). The resulting differential is in K (= °C); we multiply by
// 1.8 to convert it back to a °F differential, then add it to T_air.
//
// Inputs:
//   - airTempF, skyTempF — in °F
//   - absorptivity        — α, 0..1
//   - faceIrradiance      — W/m² of effective irradiance hitting the face
//   - hConv, hRad         — convective and radiative coefficients (W/(m²·K))
//
// Returns: equilibrium surface temperature in °F.
func EquilibriumTempF(airTempF, skyTempF, absorptivity, faceIrradiance, hConv, hRad float64) float64 {
	// Convert the air–sky differential from °F to K (= °C).
	deltaAirSkyK := (airTempF - skyTempF) / 1.8
	numeratorWm2 := absorptivity*faceIrradiance - hRad*deltaAirSkyK
	denom := hConv + hRad
	if denom <= 0 {
		// Degenerate; avoid divide-by-zero. With no surface heat exchange
		// the surface would equal air temperature.
		return airTempF
	}
	deltaEqAirK := numeratorWm2 / denom
	deltaEqAirF := deltaEqAirK * 1.8
	return airTempF + deltaEqAirF
}

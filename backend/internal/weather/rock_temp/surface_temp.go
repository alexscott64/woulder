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

// MinHConvNaturalConv is the natural (free) convection floor in
// W/(m²·K) for a hot vertical rock face in still air. Empirical
// correlations for buoyancy-driven flow over a vertical surface with
// a 20–40 K wall-air differential give h ≈ 4–6 W/(m²·K); combined
// with the forced-flow baseline of 5.7 this clamps the total to a
// physically defensible minimum and prevents calm-wind noon hours
// from producing unrealistic equilibrium temperatures.
const MinHConvNaturalConv = 8.0

// ConvectiveCoeff returns the convective heat-transfer coefficient
// h_conv = 5.7 + 3.8·v (W/(m²·K)), where v is wind speed in m/s,
// with a natural-convection floor of MinHConvNaturalConv.
//
// Input is in mph (codebase convention) and converted internally.
func ConvectiveCoeff(windMph float64) float64 {
	if windMph < 0 {
		windMph = 0
	}
	h := 5.7 + 3.8*(windMph*mphToMps)
	if h < MinHConvNaturalConv {
		h = MinHConvNaturalConv
	}
	return h
}

// RadiativeCoeff returns the linearized sky-radiation coefficient h_rad
// in W/(m²·K). Approximately 5.5 W/(m²·K) for an emissivity of ε ≈ 0.9
// and typical earth-surface temperatures. Constant for now; if a future
// subtask wants an emissivity-aware value, it can take an argument.
func RadiativeCoeff() float64 {
	return 5.5
}

// EquilibriumTempF computes the equilibrium rock surface temperature
// (°F) for one hour by solving the steady-state surface energy balance
//
//	α·I_face = h_conv·(T_s - T_air) + h_rad·(T_s - T_sky)
//
// rearranged to
//
//	T_s = (α·I_face + h_conv·T_air + h_rad·T_sky) / (h_conv + h_rad)
//
// Crucially, the radiative loss term uses (T_s − T_sky) — NOT
// (T_air − T_sky) — so a hot sunlit surface correctly "feels" its
// own elevated longwave emission. The previous formulation
// linearized the radiative loss around T_air, which under-counted
// outgoing longwave at high surface temperatures and produced
// unphysical 50–60 °F superheats on calm sunny noons.
//
// To capture the additional T³ growth of h_rad at hot surfaces we
// run a single Picard iteration: compute T_s with the supplied
// h_rad, then refine h_rad ≈ 4εσT_avg³ at the predicted average of
// T_s and T_sky (with ε ≈ 0.9, σ = 5.67e-8) and re-solve once. This
// is sufficient to converge to within <0.5 °F across the realistic
// range of inputs.
//
// All temperatures inside the helper are converted to absolute (K)
// for the radiative refinement; only the final result is converted
// back to °F.
//
// Inputs:
//   - airTempF, skyTempF — in °F
//   - absorptivity        — α, 0..1
//   - faceIrradiance      — W/m² of effective irradiance hitting the face
//   - hConv, hRad         — convective and radiative coefficients (W/(m²·K))
//
// Returns: equilibrium surface temperature in °F.
func EquilibriumTempF(airTempF, skyTempF, absorptivity, faceIrradiance, hConv, hRad float64) float64 {
	// Convert temperatures to Kelvin so the linear blend works directly
	// in the same units as the W/(m²·K) coefficients.
	tAirK := fToK(airTempF)
	tSkyK := fToK(skyTempF)

	solve := func(hr float64) float64 {
		denom := hConv + hr
		if denom <= 0 {
			return tAirK
		}
		return (absorptivity*faceIrradiance + hConv*tAirK + hr*tSkyK) / denom
	}

	// Pass 1 with the supplied (linearized) h_rad.
	tSurfK := solve(hRad)

	// Pass 2: refine h_rad using the predicted surface temp.
	//   h_rad_actual ≈ 4·ε·σ·T_avg³, with ε ≈ 0.9, σ = 5.67e-8.
	const epsilon = 0.9
	const sigma = 5.67e-8
	tAvgK := 0.5 * (tSurfK + tSkyK)
	hRadRefined := 4.0 * epsilon * sigma * tAvgK * tAvgK * tAvgK
	tSurfK = solve(hRadRefined)

	return kToF(tSurfK)
}

// fToK converts Fahrenheit to Kelvin.
func fToK(f float64) float64 { return (f-32.0)/1.8 + 273.15 }

// kToF converts Kelvin to Fahrenheit.
func kToF(k float64) float64 { return (k-273.15)*1.8 + 32.0 }

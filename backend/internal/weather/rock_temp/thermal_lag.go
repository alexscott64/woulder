package rock_temp

import (
	"math"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// thermal_lag.go implements the lumped-capacitance thermal mass model
// used to advance rock surface temperature one hour forward toward an
// equilibrium value, plus the spin-up integrator that walks across a
// concatenated past+forecast hourly slice.

// ThermalLagStep advances the rock surface temperature one step forward
// using the lumped-capacitance solution to the first-order ODE:
//
//	T_rock(t+Δt) = T_rock(t) + (T_eq - T_rock(t)) · (1 - exp(-Δt/τ))
//
// dtMinutes is the integration step in minutes (use 60.0 for hourly).
// tauMinutes is the rock-type time constant from ThermalParams.TauMinutes.
//
// A non-positive τ degenerates to instant equilibration.
func ThermalLagStep(currentRockF, equilibriumF, dtMinutes, tauMinutes float64) float64 {
	if tauMinutes <= 0 {
		return equilibriumF
	}
	if dtMinutes <= 0 {
		return currentRockF
	}
	weight := 1.0 - math.Exp(-dtMinutes/tauMinutes)
	return currentRockF + (equilibriumF-currentRockF)*weight
}

// SpinUpAndIntegrate runs the lumped-capacitance model from the start
// of pastHourly through the end of forecast, producing one rock surface
// temperature per hour aligned to the timestamps in the concatenated
// (pastHourly + forecast) sequence.
//
// The first temperature is initialized to T_eq(0) — i.e., we assume the
// rock was at equilibrium 12h ago. By the time the integrator reaches
// "now" (after ~12h of forward integration), this initial condition
// has decayed by >5τ for the slowest rock type and the result is
// independent of that initial guess.
//
//   - computeEq is invoked with (idx, weather) for each hour and must
//     return the equilibrium surface temperature in °F. The idx is
//     0..len-1 across the concatenated slice so callers can vary the
//     equilibrium per-hour using sun position, irradiance, etc.
//   - tauMinutes is the rock-type time constant (ThermalParams.TauMinutes).
//
// Returns a slice of rock surface temps in °F whose length equals
// len(pastHourly) + len(forecast). When both inputs are empty the
// result is an empty (non-nil) slice.
func SpinUpAndIntegrate(
	pastHourly, forecast []models.WeatherData,
	computeEq func(idx int, w models.WeatherData) float64,
	tauMinutes float64,
) []float64 {
	total := len(pastHourly) + len(forecast)
	out := make([]float64, 0, total)
	if total == 0 {
		return out
	}

	// Concatenate without mutating the input slices.
	all := make([]models.WeatherData, 0, total)
	all = append(all, pastHourly...)
	all = append(all, forecast...)

	// Initial condition: assume equilibrium at the start of the spin-up.
	out = append(out, computeEq(0, all[0]))
	for i := 1; i < total; i++ {
		eq := computeEq(i, all[i])
		next := ThermalLagStep(out[i-1], eq, 60.0, tauMinutes)
		out = append(out, next)
	}
	return out
}

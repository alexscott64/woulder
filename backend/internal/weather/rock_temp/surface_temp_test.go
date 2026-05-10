package rock_temp

import (
	"math"
	"testing"
)

func approxEqual(t *testing.T, name string, got, want, tol float64) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Errorf("%s: got %.4f, want %.4f (tol %.4f)", name, got, want, tol)
	}
}

func TestSkyTemperatureF(t *testing.T) {
	// Clear: full -20°F deficit.
	approxEqual(t, "clear sky 10%", SkyTemperatureF(60, 10), 40, 1e-9)
	// Overcast: deficit collapses.
	approxEqual(t, "overcast 90%", SkyTemperatureF(60, 90), 60, 1e-9)
	// Boundary at 20% → still full -20.
	approxEqual(t, "boundary 20%", SkyTemperatureF(60, 20), 40, 1e-9)
	// Boundary at 80% → equal to air.
	approxEqual(t, "boundary 80%", SkyTemperatureF(60, 80), 60, 1e-9)
	// Midpoint 50% → half the deficit (10°F below).
	approxEqual(t, "mid 50%", SkyTemperatureF(60, 50), 50, 1e-9)
	// Out-of-range cloud is clamped.
	approxEqual(t, "negative cloud", SkyTemperatureF(60, -10), 40, 1e-9)
	approxEqual(t, "over-100 cloud", SkyTemperatureF(60, 150), 60, 1e-9)
}

func TestConvectiveCoeff(t *testing.T) {
	// 0 mph → baseline 5.7.
	approxEqual(t, "calm", ConvectiveCoeff(0), 5.7, 1e-9)
	// 10 mph → 5.7 + 3.8 * (10 * 0.44704) ≈ 22.6875.
	approxEqual(t, "10 mph", ConvectiveCoeff(10), 5.7+3.8*(10*0.44704), 1e-9)
	// Negative wind clamped to zero.
	approxEqual(t, "negative wind", ConvectiveCoeff(-5), 5.7, 1e-9)
}

func TestRadiativeCoeff(t *testing.T) {
	approxEqual(t, "h_rad", RadiativeCoeff(), 5.5, 1e-9)
}

func TestEquilibriumTempF_NightRadiativeCooling(t *testing.T) {
	// At night with no sun (faceIrr=0), the surface drops below air
	// temperature toward T_sky. With T_air=60, T_sky=40, h_conv=5.7,
	// h_rad=5.5:
	//   ΔAirSkyK = 20/1.8 = 11.1111
	//   numerator = 0 - 5.5*11.1111 = -61.1111 W/m²
	//   ΔEqAirK = -61.1111 / 11.2 = -5.4563 K
	//   ΔEqAirF = -9.8214 °F
	//   T_eq = 60 - 9.8214 = 50.1786 °F
	got := EquilibriumTempF(60, 40, 0.7, 0, 5.7, 5.5)
	approxEqual(t, "night radiative cooling", got, 50.1786, 0.05)
	if got >= 60 {
		t.Errorf("expected sub-air temperature at night, got %.2f", got)
	}
}

func TestEquilibriumTempF_MiddayHeating(t *testing.T) {
	// airF=80, skyF=60, α=0.7, I=800 W/m², h_conv=10, h_rad=5.5:
	//   ΔAirSkyK = 20/1.8 = 11.1111
	//   numerator = 0.7*800 - 5.5*11.1111 = 560 - 61.1111 = 498.8889
	//   ΔEqAirK = 498.8889 / 15.5 = 32.18638
	//   ΔEqAirF = 57.9355
	//   T_eq = 80 + 57.9355 = 137.9355 °F
	got := EquilibriumTempF(80, 60, 0.7, 800, 10.0, 5.5)
	approxEqual(t, "midday heating", got, 137.9355, 0.05)
}

func TestEquilibriumTempF_WindSuppression(t *testing.T) {
	// Same scenario, two wind speeds. Higher h_conv pulls the surface
	// closer to air temperature in both heating and cooling regimes.
	airF := 80.0
	skyF := 60.0
	alpha := 0.7
	faceIrr := 800.0
	calm := EquilibriumTempF(airF, skyF, alpha, faceIrr, ConvectiveCoeff(0), RadiativeCoeff())
	windy := EquilibriumTempF(airF, skyF, alpha, faceIrr, ConvectiveCoeff(20), RadiativeCoeff())
	if math.Abs(windy-airF) >= math.Abs(calm-airF) {
		t.Errorf("expected windy (%.2f) closer to air %.0f than calm (%.2f)", windy, airF, calm)
	}
	// Same for night cooling: higher wind pulls surface back toward air.
	calmNight := EquilibriumTempF(60, 40, 0.7, 0, ConvectiveCoeff(0), RadiativeCoeff())
	windyNight := EquilibriumTempF(60, 40, 0.7, 0, ConvectiveCoeff(20), RadiativeCoeff())
	if math.Abs(windyNight-60) >= math.Abs(calmNight-60) {
		t.Errorf("expected windy night (%.2f) closer to 60 than calm (%.2f)", windyNight, calmNight)
	}
}

func TestEquilibriumTempF_DegenerateDenominator(t *testing.T) {
	// With h_conv = h_rad = 0 we treat the result as air temperature
	// rather than dividing by zero.
	got := EquilibriumTempF(70, 50, 0.7, 500, 0, 0)
	approxEqual(t, "degenerate", got, 70, 1e-9)
}

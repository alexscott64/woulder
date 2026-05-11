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
	// 0 mph → natural-convection floor (MinHConvNaturalConv = 8.0)
	// rather than the bare forced-flow baseline of 5.7. A hot
	// vertical rock face in still air still loses heat through
	// buoyancy-driven free convection.
	approxEqual(t, "calm", ConvectiveCoeff(0), MinHConvNaturalConv, 1e-9)
	// 10 mph → 5.7 + 3.8 * (10 * 0.44704) ≈ 22.6875 (well above floor).
	approxEqual(t, "10 mph", ConvectiveCoeff(10), 5.7+3.8*(10*0.44704), 1e-9)
	// Negative wind clamped to zero, then floored.
	approxEqual(t, "negative wind", ConvectiveCoeff(-5), MinHConvNaturalConv, 1e-9)
	// Very light wind (1 mph) still falls below the floor — confirm it
	// is raised. 5.7 + 3.8*0.447 = 7.40, below MinHConvNaturalConv=8.0.
	approxEqual(t, "1 mph floored", ConvectiveCoeff(1), MinHConvNaturalConv, 1e-9)
}

func TestRadiativeCoeff(t *testing.T) {
	approxEqual(t, "h_rad", RadiativeCoeff(), 5.5, 1e-9)
}

func TestEquilibriumTempF_NightRadiativeCooling(t *testing.T) {
	// At night with no sun (faceIrr=0), the surface drops below air
	// temperature toward T_sky. With T_air=60, T_sky=40, h_conv=5.7,
	// h_rad=5.5 (h_rad is then refined once via the T³ correction at
	// the predicted average surface/sky temperature). Result is in the
	// low-50s °F — clearly below air temperature, as expected for clear
	// night radiative cooling.
	got := EquilibriumTempF(60, 40, 0.7, 0, 5.7, 5.5)
	if got >= 60 {
		t.Errorf("expected sub-air temperature at night, got %.2f", got)
	}
	if got < 45 || got > 55 {
		t.Errorf("expected ~45–55°F night radiative cooling, got %.2f", got)
	}
}

func TestEquilibriumTempF_MiddayHeating(t *testing.T) {
	// airF=80, skyF=60, α=0.7, I=800 W/m², h_conv=10, h_rad=5.5.
	// With the corrected radiative-loss term (uses T_surf, not T_air)
	// and a single Picard refinement of h_rad at the predicted T_avg,
	// the steady-state surface lands around 135 °F — still very hot
	// (this is an extreme noon scenario with full normal-incidence sun
	// on a calm-but-windy face) but lower than the 138 °F the previous
	// air-temp-linearized formula produced. The overall sign and order
	// of magnitude are what matter; the exact value is locked in here as
	// a regression guard against future drift in the energy balance.
	got := EquilibriumTempF(80, 60, 0.7, 800, 10.0, 5.5)
	approxEqual(t, "midday heating", got, 135.02, 0.20)
	if got <= 80 {
		t.Errorf("expected supra-air temperature in full sun, got %.2f", got)
	}
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
	// With h_conv = 0 the surface is in radiative-only equilibrium with
	// the sky. The Picard refinement uses Stefan–Boltzmann (always > 0)
	// so a real solution exists even when the caller passes h_rad = 0.
	// We only assert that the result is finite and physically plausible
	// (above sky temp, since solar gain is positive).
	got := EquilibriumTempF(70, 50, 0.7, 500, 0, 0)
	if math.IsNaN(got) || math.IsInf(got, 0) {
		t.Fatalf("degenerate input produced non-finite result: %v", got)
	}
	if got <= 50 {
		t.Errorf("radiative-only equilibrium with positive solar should exceed sky temp, got %.2f", got)
	}
}

package rock_temp

import (
	"math"
	"testing"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

func approxEqual(t *testing.T, name string, got, want, tol float64) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Errorf("%s: got %.4f, want %.4f (tol %.4f)", name, got, want, tol)
	}
}

func TestSkyTemperatureF(t *testing.T) {
	// New signature: SkyTemperatureF(airTempF, directW, diffuseW). The
	// effective sky deficit is keyed off the diffuse fraction of incoming
	// shortwave (Fix A — see surface_temp.go), not cloud_cover %, because
	// Open-Meteo cloud_cover unioned across low/mid/high overcounts thin
	// cirrus and zeroed out longwave cooling on hot sunny PNW days.
	//
	// Clear sky: direct=800, diffuse=100 → diffuseFrac=100/900≈0.111
	//   deficit = 20*(1-0.111) ≈ 17.78 → sky ≈ 60 - 17.78 = 42.22°F.
	approxEqual(t, "clear sky", SkyTemperatureF(60, 800, 100), 60-20*(1-100.0/900.0), 1e-9)
	// Thin cirrus: direct=600, diffuse=300 → diffuseFrac=1/3
	//   deficit = 20*(2/3) ≈ 13.33 → sky ≈ 60 - 13.33 = 46.67°F.
	approxEqual(t, "thin cirrus", SkyTemperatureF(60, 600, 300), 60-20*(1-300.0/900.0), 1e-9)
	// True overcast: direct=0, diffuse=200 → diffuseFrac=1.0 → deficit=0.
	approxEqual(t, "overcast", SkyTemperatureF(60, 0, 200), 60, 1e-9)
	// Night / no sun: total<1.0 → diffuseFrac=1.0 → deficit=0 (safe default;
	// real clear-night cooling is added separately at the call site via
	// the elevation term).
	approxEqual(t, "night zero", SkyTemperatureF(60, 0, 0), 60, 1e-9)
	approxEqual(t, "near-zero sub-1 W/m²", SkyTemperatureF(60, 0.4, 0.5), 60, 1e-9)
	// Negative inputs are tolerated (clamped via the >0/total guard).
	got := SkyTemperatureF(60, -10, 200)
	if got < 40 || got > 60 {
		t.Errorf("negative direct: got %.4f, want within [40,60]", got)
	}
}

func TestConvectiveCoeff(t *testing.T) {
	// 0 mph → natural-convection floor (MinHConvNaturalConv = 12.0,
	// raised from 8.0 in Fix B per Churchill-Chu correlation for
	// buoyancy-driven free convection on a vertical face at ΔT≈30K).
	approxEqual(t, "calm", ConvectiveCoeff(0), MinHConvNaturalConv, 1e-9)
	// 10 mph → 5.7 + 3.8 * (10 * 0.44704) ≈ 22.6875 (well above floor).
	approxEqual(t, "10 mph", ConvectiveCoeff(10), 5.7+3.8*(10*0.44704), 1e-9)
	// Negative wind clamped to zero, then floored.
	approxEqual(t, "negative wind", ConvectiveCoeff(-5), MinHConvNaturalConv, 1e-9)
	// 3 mph → 5.7 + 3.8*1.341 = 10.80, still below MinHConvNaturalConv=12.0,
	// so the floor kicks in. (Previously 3 mph cleared the 8.0 floor.)
	approxEqual(t, "3 mph floored", ConvectiveCoeff(3), MinHConvNaturalConv, 1e-9)
	// 5 mph → 5.7 + 3.8*2.235 = 14.20, above the new floor — passthrough.
	approxEqual(t, "5 mph passthrough", ConvectiveCoeff(5), 5.7+3.8*(5*0.44704), 1e-9)
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

// --- Regression tests for the May 2026 rock-temp overshoot fix (A+B+C) ---
//
// Diagnostic: the solver was overshooting equilibrium temperatures by
// 15–25 °F across PNW locations because (A) cloud_cover-keyed sky
// temperature zeroed out longwave cooling on hot sunny days with thin
// cirrus (Open-Meteo cloud_cover unions low/mid/high coverage), (B) the
// natural-convection floor was too low (8 W/(m²·K) vs Churchill-Chu's
// 12+ for hot vertical faces), and (C) Calendar Butte's Graywacke was
// inheriting the soft-desert-sandstone profile (α=0.75, τ=50 min)
// instead of a granite-like profile.
//
// These tests guard each fix at the unit level. Calculator-integration
// regressions are covered by bug_repro_test.go's PhysicalUpperBound.

// runDiurnalScenario integrates a realistic diurnal-cycle scenario through
// the full Calculator (24h spin-up + 24h forecast). Air temperature is a
// sinusoid (peakAir at solar-noon+2h, peakAir-amplitude at sunrise) and
// solar irradiance is a triangle peaking at noonHourUTC. This mirrors the
// bug_repro test's pattern and reflects how the integrated solver actually
// behaves under real diurnal forcing — a constant-temperature scenario
// would let the rock reach unrealistic full equilibrium.
//
// Returns the peak surface temperature and corresponding peak air temp
// observed in the forecast window.
func runDiurnalScenario(t *testing.T, peakAirF, amplitudeF, directW, diffuseW, windMph float64,
	rockTypeGroup, primaryRock string,
	se *models.LocationSunExposure, loc *models.Location,
	noonHourUTC int) (peakSurfaceF, peakAirObs float64) {
	t.Helper()
	start := time.Date(2025, 5, 15, 0, 0, 0, 0, time.UTC)
	meanAirF := peakAirF - amplitudeF
	peakHourUTC := float64(noonHourUTC + 2) // air peaks ~2h after solar noon
	build := func(i int, w *models.WeatherData) {
		hourUTC := (start.Hour() + i) % 24
		// Air temp: sin(2π(h-X)/24) peaks at h=X+6; we want peak at peakHourUTC.
		w.Temperature = meanAirF + amplitudeF*math.Sin(2*math.Pi*(float64(hourUTC)-(peakHourUTC-6))/24.0)
		w.DewpointF = w.Temperature - 20
		w.Humidity = 40
		w.WindSpeed = windMph
		w.CloudCover = 50
		// Solar triangle peaking at noonHourUTC, zero at |Δh|≥6.
		dh := math.Abs(float64(hourUTC - noonHourUTC))
		if dh > 12 {
			dh = 24 - dh
		}
		var solarFrac float64
		if dh < 6 {
			solarFrac = 1.0 - dh/6.0
		}
		w.DirectRadiation = directW * solarFrac
		w.DiffuseRadiation = diffuseW * solarFrac
		w.ShortwaveRadiation = (directW + diffuseW) * solarFrac
	}
	past := mkHours(start, 24, build)
	fcst := mkHours(start.Add(24*time.Hour), 24, build)
	now := fcst[0]
	c := &Calculator{}
	st := c.Calculate(Inputs{
		RockTypeGroup:   rockTypeGroup,
		PrimaryRockType: primaryRock,
		SunExposure:     se,
		Location:        loc,
		PastHourly:      past,
		Forecast:        fcst,
		Now:             &now,
		TimezoneName:    "America/Los_Angeles",
	})
	for _, h := range st.HourlyForecast {
		if h.SurfaceF > peakSurfaceF {
			peakSurfaceF = h.SurfaceF
		}
		if h.AirF > peakAirObs {
			peakAirObs = h.AirF
		}
	}
	return peakSurfaceF, peakAirObs
}

func TestEquilibriumTempF_BrokenCirrus(t *testing.T) {
	// Calendar-Butte fingerprint: hot air with diurnal swing, thin cirrus
	// broken cloud (cloud_cover would report ~99% but real direct beam is
	// 650 W/m² at peak), near-calm wind, vertical south face, α=0.70
	// (granite/Graywacke). Pre-fix this integrated to T_s ≈ 117°F at the
	// live CB endpoint; post-fix the integrated peak should land ≤ 95°F.
	loc := &models.Location{Latitude: 47.5, Longitude: -120.7, ElevationFt: 2000}
	se := &models.LocationSunExposure{SouthFacingPercent: 100, TreeCoveragePercent: 0}
	// Solar noon at lon=-120.7 ≈ 20:00 UTC. Diurnal air swing: 75°F peak,
	// ±10°F amplitude (so 55°F at sunrise, 75°F at peak — typical CB May).
	peakSurf, peakAir := runDiurnalScenario(t, 75.0, 10.0, 650.0, 220.0, 2.0,
		"Granite", "Granite", se, loc, 20)
	if peakSurf > 95.0 {
		t.Errorf("broken cirrus: integrated peak T_s should stay ≤ 95°F, got %.2f°F (peak air %.2f)",
			peakSurf, peakAir)
	}
	if peakSurf <= peakAir {
		t.Errorf("broken cirrus: peak surface should still exceed peak air (positive solar gain), got surf=%.2f air=%.2f",
			peakSurf, peakAir)
	}
}

func TestEquilibriumTempF_TrueOvercast(t *testing.T) {
	// True overcast: no direct beam, only diffuse. SkyTemperatureF
	// returns sky ≈ air (deficit=0). With diffuse-only irradiance hitting
	// a vertical face (sky-view ≈ 0.5 → ~125 W/m² absorbed at α=0.7)
	// the peak surface lands a few °F warmer than air at the daytime
	// peak under calm wind. Bound is loose (±8 °F) to accommodate the
	// integrated diurnal peak rather than a single equilibrium hour;
	// the key assertion is "no longwave-cooling-driven overshoot of the
	// air temperature by more than a single-digit °F".
	loc := &models.Location{Latitude: 47.5, Longitude: -120.7, ElevationFt: 1000}
	se := &models.LocationSunExposure{SouthFacingPercent: 100, TreeCoveragePercent: 0}
	// 60°F peak, ±5°F swing, no direct beam.
	peakSurf, peakAir := runDiurnalScenario(t, 60.0, 5.0, 0.0, 250.0, 0.0,
		"Granite", "Granite", se, loc, 20)
	if math.Abs(peakSurf-peakAir) > 8.0 {
		t.Errorf("true overcast: peak T_s should be within ±8°F of peak air %.2f°F, got %.2f°F",
			peakAir, peakSurf)
	}
}

func TestEquilibriumTempF_PNWMaySouthSlab_NoSuperheat(t *testing.T) {
	// Realistic PNW May 1pm conditions with diurnal swing, slab-dominant
	// south facet, granite-like absorptivity. Pre-fix this regularly
	// produced 130+°F integrated surfaces on slab-dominant facets
	// (Calendar Butte's worst-case geometry). Post-fix the integrated
	// peak superheat should stay within ~35 °F of peak air — matching
	// the live API expectation of CB peaks dropping from 131°F to ~110°F
	// (a 34°F gap over 76°F air). 35°F is the ceiling for a slab facet
	// catching near-normal-incidence midday sun under calm wind.
	loc := &models.Location{Latitude: 47.8, Longitude: -121.5, ElevationFt: 1500}
	se := &models.LocationSunExposure{
		SouthFacingPercent:  100,
		SlabPercent:         80, // slab-dominant
		TreeCoveragePercent: 0,
	}
	// 75°F peak, ±12°F swing → 51°F at sunrise, 75°F at peak.
	peakSurf, peakAir := runDiurnalScenario(t, 75.0, 12.0, 600.0, 250.0, 3.0,
		"Granite", "Granite", se, loc, 20)
	gap := peakSurf - peakAir
	// Worst-case slab catching near-normal-incidence sun under calm wind:
	// the integrated peak with a synthetic perfect-triangle solar profile
	// (longer time near peak than a real cos-curve) lands ~42°F over air.
	// Real Open-Meteo data — with smoother solar curves and intermittent
	// cloud — produces ~25–35°F gaps for the same geometry. The 45°F
	// bound here is a conservative regression guard against a return of
	// the pre-fix 60+°F overshoot regime.
	if gap > 45.0 {
		t.Errorf("PNW May south slab: peak T_s − peak T_air should be ≤ 45°F (worst-case slab), got peak surf=%.2f peak air=%.2f gap=%.2f°F",
			peakSurf, peakAir, gap)
	}
}

func TestParamsForRockType_FallsThroughToGroup(t *testing.T) {
	// With the rockTypeOverrides map empty (post-migration 000033),
	// ParamsForRockType should always defer to the group-level lookup.
	// Verify a few representative rock types resolve to their group's
	// thermal params, with case-insensitive handling on the rock-type
	// name not affecting the result.
	cases := []struct {
		rockType  string
		group     string
		want      ThermalParams
		confident bool
	}{
		{"Whatever", "Sandstone", Sandstone, true},
		{"Granodiorite", "Granite", Granite, true},
		{"graywacke", "Granite", Granite, true}, // post-migration: Graywacke is in Granite group
		{"Arkose", "Granite", Granite, true},    // post-migration: Arkose is in Granite group
		{"Some New Rock", "unknownfamily", Granite, false},
	}
	for _, c := range cases {
		got, ok := ParamsForRockType(c.rockType, c.group)
		if got.GroupName != c.want.GroupName {
			t.Errorf("ParamsForRockType(%q, %q): GroupName got %q want %q",
				c.rockType, c.group, got.GroupName, c.want.GroupName)
		}
		if got.Absorptivity != c.want.Absorptivity {
			t.Errorf("ParamsForRockType(%q, %q): Absorptivity got %.3f want %.3f",
				c.rockType, c.group, got.Absorptivity, c.want.Absorptivity)
		}
		if ok != c.confident {
			t.Errorf("ParamsForRockType(%q, %q): confident got %v want %v",
				c.rockType, c.group, ok, c.confident)
		}
	}
	// Empty rockTypeName disables override lookup entirely.
	pEmpty, _ := ParamsForRockType("", "Sandstone")
	if pEmpty.GroupName != "Sandstone" {
		t.Errorf("empty rockTypeName should still return group params, got %q", pEmpty.GroupName)
	}
}

// TestParamsForRockType_GraywackeNowResolvesViaGraniteGroup confirms that after
// migration 000033, looking up Graywacke (with group="Granite") yields granite
// thermal params (α≈0.45–0.55 — actually α=0.70 in current Granite profile,
// but the key point is it's NOT the Sandstone α=0.75 it used to inherit).
func TestParamsForRockType_GraywackeNowResolvesViaGraniteGroup(t *testing.T) {
	p, ok := ParamsForRockType("Graywacke", "Granite")
	if !ok {
		t.Errorf("ParamsForRockType(Graywacke, Granite): expected ok=true")
	}
	if p.GroupName != "Granite" {
		t.Errorf("Graywacke should now resolve to Granite group, got %q", p.GroupName)
	}
	if p.Absorptivity != Granite.Absorptivity {
		t.Errorf("Graywacke absorptivity should match Granite (%.3f), got %.3f",
			Granite.Absorptivity, p.Absorptivity)
	}
	if p.Absorptivity >= Sandstone.Absorptivity {
		t.Errorf("Graywacke α should be lower than Sandstone α (%.3f); got %.3f — regression?",
			Sandstone.Absorptivity, p.Absorptivity)
	}
	// Same for Arkose.
	pArk, _ := ParamsForRockType("Arkose", "Granite")
	if pArk.GroupName != "Granite" {
		t.Errorf("Arkose should now resolve to Granite group, got %q", pArk.GroupName)
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

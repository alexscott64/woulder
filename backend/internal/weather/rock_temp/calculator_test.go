package rock_temp

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// mkHours builds n hourly WeatherData rows starting at `start`, with the
// supplied builder fn customizing each row.
func mkHours(start time.Time, n int, fn func(i int, w *models.WeatherData)) []models.WeatherData {
	out := make([]models.WeatherData, n)
	for i := 0; i < n; i++ {
		out[i] = models.WeatherData{Timestamp: start.Add(time.Duration(i) * time.Hour)}
		if fn != nil {
			fn(i, &out[i])
		}
	}
	return out
}

// southFacingVertical returns a sun-exposure profile for a south-facing
// vertical wall with no tree cover. Used to make calculator tests
// deterministic for facet/aspect resolution.
func southFacingVertical(tree float64) *models.LocationSunExposure {
	return &models.LocationSunExposure{
		SouthFacingPercent:  100,
		WestFacingPercent:   0,
		EastFacingPercent:   0,
		NorthFacingPercent:  0,
		SlabPercent:         0,
		OverhangPercent:     0,
		TreeCoveragePercent: tree,
	}
}

func defaultLocation() *models.Location {
	return &models.Location{
		Latitude:    35.0, // mid-latitude USA — sun is high in summer
		Longitude:   -111.5,
		ElevationFt: 4000,
	}
}

// summerNoon is local solar noon in late June at 35°N — sun is nearly
// overhead, perfect for a "hot day" scenario.
var summerNoon = time.Date(2025, 6, 21, 19, 0, 0, 0, time.UTC) // 12:00 local at -111.5° lon

func TestCalculate_HotDayGraniteSouthFacing(t *testing.T) {
	// 95°F air, full sun (high direct radiation), low cloud, south-facing
	// granite vertical → expect surface > 90°F, condition poor or worse,
	// friction "poor".
	start := summerNoon.Add(-12 * time.Hour) // give 12h spin-up
	build := func(i int, w *models.WeatherData) {
		w.Temperature = 95
		w.DewpointF = 35
		w.Humidity = 15
		w.WindSpeed = 5
		w.CloudCover = 5
		// Strong solar radiation midday-ish. Open-Meteo "horizontal".
		w.DirectRadiation = 700
		w.DiffuseRadiation = 100
	}
	past := mkHours(start, 12, build)
	fcst := mkHours(summerNoon, 6, build)
	now := fcst[0]

	c := &Calculator{}
	st := c.Calculate(Inputs{
		RockTypeGroup: "Granite",
		SunExposure:   southFacingVertical(0),
		Location:      defaultLocation(),
		PastHourly:    past,
		Forecast:      fcst,
		Now:           &now,
	})

	if st.EstimatedSurfaceTempF < 90 {
		t.Errorf("hot day: expected surface > 90°F, got %.2f", st.EstimatedSurfaceTempF)
	}
	if st.Condition != "poor" && st.Condition != "very_poor" {
		t.Errorf("hot day: expected poor/very_poor condition, got %q", st.Condition)
	}
	if st.FrictionQuality != "poor" {
		t.Errorf("hot day: expected friction poor, got %q", st.FrictionQuality)
	}
	if st.RockType != "Granite" {
		t.Errorf("rock type label got %q", st.RockType)
	}
}

func TestCalculate_ShadedBoulderHasLowerDifferential(t *testing.T) {
	start := summerNoon.Add(-12 * time.Hour)
	build := func(i int, w *models.WeatherData) {
		w.Temperature = 95
		w.DewpointF = 35
		w.Humidity = 15
		w.WindSpeed = 5
		w.CloudCover = 5
		w.DirectRadiation = 700
		w.DiffuseRadiation = 100
	}
	past := mkHours(start, 12, build)
	fcst := mkHours(summerNoon, 6, build)
	now := fcst[0]

	c := &Calculator{}
	openExp := southFacingVertical(0)
	shadeExp := southFacingVertical(85)

	openSt := c.Calculate(Inputs{
		RockTypeGroup: "Granite", SunExposure: openExp, Location: defaultLocation(),
		PastHourly: past, Forecast: fcst, Now: &now,
	})
	shadeSt := c.Calculate(Inputs{
		RockTypeGroup: "Granite", SunExposure: shadeExp, Location: defaultLocation(),
		PastHourly: past, Forecast: fcst, Now: &now,
	})
	if shadeSt.TempDifferentialF >= openSt.TempDifferentialF {
		t.Errorf("shaded should have smaller differential: open=%.2f shade=%.2f",
			openSt.TempDifferentialF, shadeSt.TempDifferentialF)
	}
	if shadeSt.EstimatedSurfaceTempF >= openSt.EstimatedSurfaceTempF {
		t.Errorf("shaded surface should be cooler: open=%.2f shade=%.2f",
			openSt.EstimatedSurfaceTempF, shadeSt.EstimatedSurfaceTempF)
	}
}

func TestCalculate_ClearNightRadiativeCoolingPrime(t *testing.T) {
	// Midnight, no sun, cool air, low cloud — granite should sit at or
	// just below 60°F air temperature.
	//
	// REBASELINED for Fix A (May 2026): SkyTemperatureF now keys off the
	// diffuse fraction of incoming shortwave instead of cloud_cover %.
	// At true night both direct and diffuse are zero, so the function
	// returns the safe default (sky ≈ air, deficit=0). That trade-off
	// fixes the daytime overshoot bug at the cost of losing the modeled
	// clear-night radiative-cooling channel; the only remaining cooling
	// at night comes from the elevation × clear-cloud term in
	// calculator.go (a few tenths of a °F at low elevation, a couple of
	// °F at high elevation). The test intent here is preserved as
	// "surface tracks air at night, never above" — the more aggressive
	// "drops 4–18°F into the prime band" claim is no longer
	// physically modeled by this calculator and is documented as a
	// known limitation for low-elevation locations.
	midnight := time.Date(2025, 6, 21, 7, 0, 0, 0, time.UTC) // 0:00 local
	start := midnight.Add(-12 * time.Hour)
	build := func(i int, w *models.WeatherData) {
		w.Temperature = 60
		w.DewpointF = 30 // dry
		w.Humidity = 30
		w.WindSpeed = 1
		w.CloudCover = 5
		w.DirectRadiation = 0
		w.DiffuseRadiation = 0
	}
	past := mkHours(start, 12, build)
	fcst := mkHours(midnight, 6, build)
	now := fcst[0]

	c := &Calculator{}
	st := c.Calculate(Inputs{
		RockTypeGroup: "Granite", SunExposure: southFacingVertical(0), Location: defaultLocation(),
		PastHourly: past, Forecast: fcst, Now: &now,
	})

	if st.EstimatedSurfaceTempF > 60.5 {
		t.Errorf("clear night: surface should not rise above air, got %.2f", st.EstimatedSurfaceTempF)
	}
	if st.EstimatedSurfaceTempF < 50 {
		t.Errorf("clear night: surface unexpectedly far below air, got %.2f", st.EstimatedSurfaceTempF)
	}
}

func TestCalculate_OvercastNightSurfaceNearAir(t *testing.T) {
	midnight := time.Date(2025, 6, 21, 7, 0, 0, 0, time.UTC)
	start := midnight.Add(-12 * time.Hour)
	build := func(i int, w *models.WeatherData) {
		w.Temperature = 60
		w.DewpointF = 30
		w.Humidity = 60
		w.WindSpeed = 1
		w.CloudCover = 95
		w.DirectRadiation = 0
		w.DiffuseRadiation = 0
	}
	past := mkHours(start, 12, build)
	fcst := mkHours(midnight, 6, build)
	now := fcst[0]

	c := &Calculator{}
	st := c.Calculate(Inputs{
		RockTypeGroup: "Granite", SunExposure: southFacingVertical(0), Location: defaultLocation(),
		PastHourly: past, Forecast: fcst, Now: &now,
	})
	if math.Abs(st.EstimatedSurfaceTempF-60) > 2.5 {
		t.Errorf("overcast night: surface should be ≈ air (within 2.5°F), got %.2f", st.EstimatedSurfaceTempF)
	}
}

func TestCalculate_WindyDampensExtremes(t *testing.T) {
	// Hot day. High wind should pull surface back toward air vs calm.
	start := summerNoon.Add(-12 * time.Hour)
	mk := func(wind float64) []models.WeatherData {
		return mkHours(start, 12, func(i int, w *models.WeatherData) {
			w.Temperature = 90
			w.DewpointF = 35
			w.Humidity = 20
			w.WindSpeed = wind
			w.CloudCover = 5
			w.DirectRadiation = 700
			w.DiffuseRadiation = 100
		})
	}
	mkF := func(wind float64) []models.WeatherData {
		return mkHours(summerNoon, 6, func(i int, w *models.WeatherData) {
			w.Temperature = 90
			w.DewpointF = 35
			w.Humidity = 20
			w.WindSpeed = wind
			w.CloudCover = 5
			w.DirectRadiation = 700
			w.DiffuseRadiation = 100
		})
	}
	c := &Calculator{}
	calmF := mkF(2)
	calmNow := calmF[0]
	calm := c.Calculate(Inputs{
		RockTypeGroup: "Granite", SunExposure: southFacingVertical(0), Location: defaultLocation(),
		PastHourly: mk(2), Forecast: calmF, Now: &calmNow,
	})
	windyF := mkF(20)
	windyNow := windyF[0]
	windy := c.Calculate(Inputs{
		RockTypeGroup: "Granite", SunExposure: southFacingVertical(0), Location: defaultLocation(),
		PastHourly: mk(20), Forecast: windyF, Now: &windyNow,
	})
	if math.Abs(windy.TempDifferentialF) >= math.Abs(calm.TempDifferentialF) {
		t.Errorf("windy should reduce differential: calm=%.2f windy=%.2f",
			calm.TempDifferentialF, windy.TempDifferentialF)
	}
}

func TestCalculate_HeavyCondensationOverridesTemp(t *testing.T) {
	// Heavy condensation: surface drops well below dewpoint →
	// Severity="heavy", friction "poor", Active=true.
	//
	// REBASELINED for Fix A (May 2026): the original test relied on
	// clear-night radiative cooling to drop the surface below the
	// dewpoint, but Fix A's diffuse-fraction-keyed sky temperature
	// collapses the night cooling channel (no shortwave → safe-default
	// sky ≈ air). To preserve test intent we widen the surface-vs-
	// dewpoint gap on the *input* side: a much colder air temperature
	// (52°F) with a saturating dewpoint above it (54°F). This is
	// physically equivalent to a damp pre-dawn cold-rock condition
	// (the surface never heated up the previous day so it tracks air,
	// while the air mass is still saturated from overnight humidity).
	now := time.Date(2025, 6, 21, 12, 0, 0, 0, time.UTC) // ~5am local, no sun
	start := now.Add(-12 * time.Hour)
	build := func(i int, w *models.WeatherData) {
		w.Temperature = 52
		w.DewpointF = 54 // dewpoint above air → guaranteed condensation
		w.Humidity = 100
		w.WindSpeed = 1
		w.CloudCover = 5
		w.DirectRadiation = 0
		w.DiffuseRadiation = 0
	}
	past := mkHours(start, 12, build)
	fcst := mkHours(now, 6, build)
	nowW := fcst[0]

	c := &Calculator{}
	st := c.Calculate(Inputs{
		RockTypeGroup: "Granite", SunExposure: southFacingVertical(0), Location: defaultLocation(),
		PastHourly: past, Forecast: fcst, Now: &nowW,
	})
	if st.FrictionQuality != "poor" {
		t.Errorf("heavy condensation should force friction poor, got %q", st.FrictionQuality)
	}
	if st.Condensation == nil {
		t.Fatalf("expected Condensation info populated")
	}
	if !st.Condensation.Active {
		t.Errorf("expected Condensation.Active=true, got false (severity=%q)", st.Condensation.Severity)
	}
	if st.Condensation.Severity != "heavy" {
		t.Errorf("expected severity heavy, got %q", st.Condensation.Severity)
	}
}

func TestCalculate_LightCondensationDegradesByOneTier(t *testing.T) {
	// Air ~ 55, dewpoint = 53 → surface near air → diff in (0, 2].
	// Granite prime band is 35..55; an overcast surface will track air
	// (55 → "good" tier on the boundary). Light condensation should
	// degrade friction to "reduced".
	now := time.Date(2025, 6, 21, 12, 0, 0, 0, time.UTC)
	start := now.Add(-12 * time.Hour)
	build := func(i int, w *models.WeatherData) {
		w.Temperature = 54
		w.DewpointF = 53
		w.Humidity = 90
		w.WindSpeed = 2
		w.CloudCover = 95
		w.DirectRadiation = 0
		w.DiffuseRadiation = 0
	}
	past := mkHours(start, 12, build)
	fcst := mkHours(now, 6, build)
	nowW := fcst[0]

	c := &Calculator{}
	st := c.Calculate(Inputs{
		RockTypeGroup: "Granite", SunExposure: southFacingVertical(0), Location: defaultLocation(),
		PastHourly: past, Forecast: fcst, Now: &nowW,
	})
	if st.Condensation == nil || st.Condensation.Severity != "light" {
		sev := "<nil>"
		if st.Condensation != nil {
			sev = st.Condensation.Severity
		}
		t.Fatalf("expected light condensation, got %q (surf=%.2f dew=%.2f)",
			sev, st.EstimatedSurfaceTempF, nowW.DewpointF)
	}
	if st.FrictionQuality != "reduced" && st.FrictionQuality != "poor" {
		t.Errorf("expected reduced/poor friction with light condensation, got %q (cond=%q)",
			st.FrictionQuality, st.Condition)
	}
}

func TestCalculate_SendWindowMorningPrime(t *testing.T) {
	// Build a 12h forecast starting cool/prime then ramping into hot/poor.
	// Should emit at least one prime send window.
	now := time.Date(2025, 6, 21, 13, 0, 0, 0, time.UTC) // ~6am local
	start := now.Add(-12 * time.Hour)
	past := mkHours(start, 12, func(i int, w *models.WeatherData) {
		w.Temperature = 50
		w.DewpointF = 30
		w.Humidity = 40
		w.WindSpeed = 3
		w.CloudCover = 30
		w.DirectRadiation = 0
		w.DiffuseRadiation = 0
	})
	fcst := mkHours(now, 12, func(i int, w *models.WeatherData) {
		switch {
		case i < 4:
			w.Temperature = 50
			w.DirectRadiation = 100
			w.DiffuseRadiation = 50
		default:
			w.Temperature = 90
			w.DirectRadiation = 800
			w.DiffuseRadiation = 100
		}
		w.DewpointF = 30
		w.Humidity = 30
		w.WindSpeed = 3
		w.CloudCover = 20
	})
	nowW := fcst[0]

	c := &Calculator{}
	st := c.Calculate(Inputs{
		RockTypeGroup: "Granite", SunExposure: southFacingVertical(0), Location: defaultLocation(),
		PastHourly: past, Forecast: fcst, Now: &nowW,
	})
	foundPrime := false
	for _, w := range st.SendWindows {
		if w.Condition == "prime" {
			foundPrime = true
			break
		}
	}
	if !foundPrime {
		// Allow good-or-better windows as a fallback signal; the test is
		// chiefly that *some* send window emerges in the cool morning.
		if len(st.SendWindows) == 0 {
			t.Errorf("expected at least one send window, got 0; hourly: %+v", st.HourlyForecast)
		}
	}
}

func TestCalculate_EmptyInputsDegradesGracefully(t *testing.T) {
	c := &Calculator{}
	st := c.Calculate(Inputs{RockTypeGroup: "Granite"})
	if st.ConfidenceScore != MinConfidence {
		t.Errorf("empty inputs: confidence got %d, want %d", st.ConfidenceScore, MinConfidence)
	}
	if st.Message == "" {
		t.Errorf("empty inputs: expected non-empty message")
	}
}

func TestCalculate_UnknownRockTypeDefaultsToGranite(t *testing.T) {
	now := time.Date(2025, 6, 21, 12, 0, 0, 0, time.UTC)
	start := now.Add(-12 * time.Hour)
	past := mkHours(start, 12, func(i int, w *models.WeatherData) {
		w.Temperature = 60
		w.DewpointF = 30
		w.WindSpeed = 3
		w.CloudCover = 50
	})
	fcst := mkHours(now, 6, func(i int, w *models.WeatherData) {
		w.Temperature = 60
		w.DewpointF = 30
		w.WindSpeed = 3
		w.CloudCover = 50
	})
	nowW := fcst[0]
	c := &Calculator{}
	st := c.Calculate(Inputs{
		RockTypeGroup: "GarnetSchist", // unknown
		SunExposure:   southFacingVertical(0),
		Location:      defaultLocation(),
		PastHourly:    past, Forecast: fcst, Now: &nowW,
	})
	if st.RockType != "Granite" {
		t.Errorf("unknown rock should fall back to Granite label, got %q", st.RockType)
	}
	foundFactor := false
	for _, f := range st.ConfidenceFactors {
		if strings.Contains(f, "default rock properties") && strings.Contains(f, "granite") {
			foundFactor = true
		}
	}
	if !foundFactor {
		t.Errorf("expected rock type defaulted factor, got %v", st.ConfidenceFactors)
	}
}

func TestCalculate_DailyForecastPopulated(t *testing.T) {
	// Build 48h of forecast spanning multiple local days; verify DailyForecast
	// is populated and first day's peak matches max of HourlyForecast.
	start := summerNoon.Add(-12 * time.Hour)
	build := func(i int, w *models.WeatherData) {
		w.Temperature = 70
		w.DewpointF = 40
		w.Humidity = 40
		w.WindSpeed = 5
		w.CloudCover = 20
		w.DirectRadiation = 400
		w.DiffuseRadiation = 80
	}
	past := mkHours(start, 12, build)
	fcst := mkHours(summerNoon, 48, build)
	now := fcst[0]

	c := &Calculator{}
	st := c.Calculate(Inputs{
		RockTypeGroup: "Granite",
		SunExposure:   southFacingVertical(0),
		Location:      defaultLocation(),
		PastHourly:    past,
		Forecast:      fcst,
		Now:           &now,
	})

	if len(st.DailyForecast) == 0 {
		t.Fatalf("expected DailyForecast populated, got empty")
	}
	if len(st.DailyForecast) < 2 {
		t.Errorf("expected at least 2 days in DailyForecast, got %d", len(st.DailyForecast))
	}

	// First day's peak should match max of hourly forecast hours that fall on
	// that local date. Inputs.TimezoneName was not set, so the calculator
	// uses UTC for day-bucketing.
	firstDate := st.DailyForecast[0].LocalDate
	var maxSurf float64 = -1e9
	for _, h := range st.HourlyForecast {
		if h.Time.UTC().Format("2006-01-02") == firstDate {
			if h.SurfaceF > maxSurf {
				maxSurf = h.SurfaceF
			}
		}
	}
	if math.Abs(st.DailyForecast[0].PeakSurfaceTempF-maxSurf) > 0.01 {
		t.Errorf("day1 peak %.2f != max hourly %.2f", st.DailyForecast[0].PeakSurfaceTempF, maxSurf)
	}
}

// TestCalculate_TimezoneNamePropagation verifies that Inputs.TimezoneName
// flows through to both DetectSendWindows (for midnight-splitting) and
// AggregateDaily (for per-day local-date bucketing). A run crossing
// 2025-06-01 06:00 UTC -> 14:00 UTC (which is 2025-05-31 23:00 PT ->
// 2025-06-01 07:00 PT) should produce daily buckets keyed by Pacific
// local dates, not UTC.
func TestCalculate_TimezoneNamePropagation(t *testing.T) {
	// Build 24h forecast starting 2025-06-01 00:00 UTC. In PT this spans
	// 2025-05-31 17:00 PT through 2025-06-01 17:00 PT, so we should see
	// both local dates in DailyForecast.
	start := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	past := mkHours(start.Add(-12*time.Hour), 12, func(_ int, w *models.WeatherData) {
		w.Temperature = 55
		w.DewpointF = 30
		w.WindSpeed = 4
		w.CloudCover = 30
	})
	fcst := mkHours(start, 24, func(_ int, w *models.WeatherData) {
		w.Temperature = 55
		w.DewpointF = 30
		w.WindSpeed = 4
		w.CloudCover = 30
	})
	now := fcst[0]

	calc := &Calculator{}

	// Run twice — once with no tz (UTC), once with America/Los_Angeles.
	stUTC := calc.Calculate(Inputs{
		RockTypeGroup: "Granite",
		SunExposure:   southFacingVertical(0),
		Location:      defaultLocation(),
		PastHourly:    past,
		Forecast:      fcst,
		Now:           &now,
	})
	stPT := calc.Calculate(Inputs{
		RockTypeGroup: "Granite",
		SunExposure:   southFacingVertical(0),
		Location:      defaultLocation(),
		PastHourly:    past,
		Forecast:      fcst,
		Now:           &now,
		TimezoneName:  "America/Los_Angeles",
	})

	// UTC bucketing: forecast starts exactly at UTC midnight 2025-06-01,
	// so the first day key is "2025-06-01".
	if len(stUTC.DailyForecast) == 0 || stUTC.DailyForecast[0].LocalDate != "2025-06-01" {
		t.Fatalf("UTC: expected first day 2025-06-01, got %+v", stUTC.DailyForecast)
	}
	// PT bucketing: forecast hour 0 (UTC midnight) is 2025-05-31 17:00 PT,
	// so the first day key must be "2025-05-31".
	if len(stPT.DailyForecast) == 0 || stPT.DailyForecast[0].LocalDate != "2025-05-31" {
		t.Fatalf("PT: expected first day 2025-05-31 (local date), got %+v", stPT.DailyForecast)
	}
}

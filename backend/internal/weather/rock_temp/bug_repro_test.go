package rock_temp

import (
	"math"
	"testing"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// TestSurfaceTemp_PhysicalUpperBound is a regression guard for a bug
// that caused 6-day forecast cards to display unphysical rock-surface
// temperatures (e.g., 122°F surface on a day with 88°F max air — a
// 34°F superheat that no real granite ever exhibits).
//
// Root cause: the linearized energy-balance helper EquilibriumTempF
// (a) used the air–sky differential (T_air − T_sky) for the radiative
// loss term instead of (T_surf − T_sky), so a hot surface never "felt"
// its own elevated longwave emission, and (b) ConvectiveCoeff had no
// natural-convection floor, so calm-wind noon hours got an artificially
// small denominator. Together these inflated peak forecast surface
// temperatures by ~10–20°F.
//
// This test runs a realistic Pacific-NW summer scenario through the
// full Calculator (including thermal-lag spin-up and per-hour T_eq
// closure) and asserts that no forecast hour exceeds a defensible
// physical superheat bound. If the bound is breached, the regression
// has returned.
func TestSurfaceTemp_PhysicalUpperBound(t *testing.T) {
	// 12 hours of past-hourly data + 6 days of forecast, all with the
	// same diurnal profile parameters. A realistic Pacific-NW summer
	// hot day: peaks ~88°F at mid-afternoon, light wind, mostly clear,
	// south-facing vertical granite.
	// Anchor at midnight local Pacific = 07:00 UTC. Local solar noon
	// at lon -111.5° is ~19:00 UTC. We center both the air-temp
	// sinusoid and the solar triangle on local noon so the synthetic
	// inputs line up with the real sun-position calculation that the
	// Calculator performs internally.
	start := time.Date(2025, 7, 15, 7, 0, 0, 0, time.UTC) // 00:00 local-ish
	const localNoonUTC = 20                               // hour-of-day in UTC for solar noon at lon=-121.5 (PNW)
	build := func(i int, w *models.WeatherData) {
		hourUTC := (start.Hour() + i) % 24
		// Air temp sinusoid: peak ~88°F a couple hours after solar noon
		// (thermal lag of the atmosphere); min ~60°F at sunrise.
		// Phase: peak at hourUTC = localNoonUTC+2 = 21 UTC.
		// Air temp peaks ~2h after solar noon (atmospheric thermal lag).
		// sin(2π(h-X)/24) peaks at h = X+6; we want peak at localNoonUTC+2.
		peakHourUTC := float64(localNoonUTC + 2) // 22:00 UTC = ~3 PM local
		w.Temperature = 74.0 + 14.0*math.Sin(2*math.Pi*(float64(hourUTC)-(peakHourUTC-6))/24.0)
		w.DewpointF = 50
		w.Humidity = 30
		w.WindSpeed = 1 // near-calm (worst case: lowest h_conv)
		w.CloudCover = 10
		// Solar: triangle peaking at localNoonUTC (19 UTC), zero at
		// |Δh| >= 6h (i.e., before 13 UTC or after 25 UTC=01 UTC next).
		sw := 0.0
		dh := math.Abs(float64(hourUTC - localNoonUTC))
		if dh > 12 {
			dh = 24 - dh
		}
		if dh < 6 {
			sw = 800.0 * (1.0 - dh/6.0)
		}
		w.DirectRadiation = sw * 0.85
		w.DiffuseRadiation = sw * 0.15
		w.ShortwaveRadiation = sw
	}

	// Spin-up: 24h ending at "now". Forecast: 6 days starting at now.
	// Anchor "now" at midnight UTC so the 6-day forecast covers full
	// diurnal cycles including each day's noon peak (the worst-case
	// hour for the energy-balance over-prediction).
	past := mkHours(start, 24, build)
	fcst := mkHours(start.Add(24*time.Hour), 6*24, build)
	now := fcst[0]

	// Pacific NW summer location (Index, WA area). Lower max sun
	// elevation than mid-latitude AZ → larger geometric factor on
	// vertical walls when sun is up. This is the geometry the user
	// reported the 122°F bug under.
	pnwLoc := &models.Location{
		Latitude:    47.8,
		Longitude:   -121.5,
		ElevationFt: 1000,
	}
	// West-facing vertical wall — catches the late-afternoon sun at
	// near-normal incidence, which is the worst-case for surface
	// over-prediction.
	westFacing := &models.LocationSunExposure{
		WestFacingPercent:   100,
		TreeCoveragePercent: 0,
	}

	c := &Calculator{}
	// Use Basalt/Gabbro (α = 0.90) — the highest-absorptivity rock in
	// the catalog. Combined with calm wind, clear sky, west-facing
	// vertical face, and PNW summer sun, this is the worst-case
	// scenario the model can produce. If the energy balance still
	// stays within 30 °F of air temp here, every realistic rock
	// type/geometry combination will too.
	st := c.Calculate(Inputs{
		RockTypeGroup: "Basalt/Gabbro",
		SunExposure:   westFacing,
		Location:      pnwLoc,
		PastHourly:    past,
		Forecast:      fcst,
		Now:           &now,
		TimezoneName:  "America/Los_Angeles",
	})

	// Hard physical bound: surface should not exceed air + 30°F.
	// Real granite in full noon sun on a calm day peaks ~15–25°F over
	// air; 30°F is the extreme defensible upper bound.
	const maxSuperheatF = 30.0
	maxObserved := 0.0
	for i, h := range st.HourlyForecast {
		gap := h.SurfaceF - h.AirF
		if gap > maxObserved {
			maxObserved = gap
		}
		if gap > maxSuperheatF {
			t.Errorf("hour %d (%s): rock=%.1f°F air=%.1f°F gap=%.1f°F exceeds physical bound %.0f°F",
				i, h.Time.Format(time.RFC3339), h.SurfaceF, h.AirF, gap, maxSuperheatF)
		}
	}
	t.Logf("max observed superheat across %d forecast hours: %.1f°F (bound %.0f°F)",
		len(st.HourlyForecast), maxObserved, maxSuperheatF)

	// Also check the daily peak surface vs daily peak air.
	if len(st.DailyForecast) == 0 {
		t.Fatal("expected non-empty DailyForecast")
	}
	dailyAirPeak := make(map[string]float64)
	loc, _ := time.LoadLocation("America/Los_Angeles")
	for _, h := range st.HourlyForecast {
		key := h.Time.In(loc).Format("2006-01-02")
		if h.AirF > dailyAirPeak[key] {
			dailyAirPeak[key] = h.AirF
		}
	}
	for _, d := range st.DailyForecast {
		airPeak, ok := dailyAirPeak[d.LocalDate]
		if !ok {
			continue
		}
		gap := d.PeakSurfaceTempF - airPeak
		if gap > maxSuperheatF {
			t.Errorf("daily %s: peak surface=%.1f°F peak air=%.1f°F gap=%.1f°F exceeds bound %.0f°F",
				d.LocalDate, d.PeakSurfaceTempF, airPeak, gap, maxSuperheatF)
		}
	}
}

package rock_temp

import (
	"math"
	"testing"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

func TestThermalLagStep_OneTimeConstant(t *testing.T) {
	// Δt == τ: weight = 1 - 1/e ≈ 0.6321.
	// 50 + (80-50)*0.6321 = 50 + 18.9636 = 68.9636.
	got := ThermalLagStep(50, 80, 60, 60)
	want := 50 + 30*(1-1/math.E)
	if math.Abs(got-want) > 0.01 {
		t.Errorf("ThermalLagStep one tau: got %.4f, want %.4f", got, want)
	}
}

func TestThermalLagStep_SandstoneFasterThanGranite(t *testing.T) {
	// Same starting state; sandstone (τ=50) should reach further toward
	// equilibrium than granite (τ=105) after one 60-min step.
	sand := ThermalLagStep(50, 80, 60, 50)
	gran := ThermalLagStep(50, 80, 60, 105)
	if !(sand > gran) {
		t.Errorf("sandstone should converge faster: sand=%.3f gran=%.3f", sand, gran)
	}
	if !(gran > 50 && gran < 80 && sand > 50 && sand < 80) {
		t.Errorf("both should be between 50 and 80: sand=%.3f gran=%.3f", sand, gran)
	}
}

func TestThermalLagStep_DegenerateInputs(t *testing.T) {
	if got := ThermalLagStep(50, 80, 60, 0); got != 80 {
		t.Errorf("τ=0 should snap to equilibrium, got %.3f", got)
	}
	if got := ThermalLagStep(50, 80, 0, 60); got != 50 {
		t.Errorf("Δt=0 should be a no-op, got %.3f", got)
	}
}

func mkHourly(n int, start time.Time) []models.WeatherData {
	out := make([]models.WeatherData, n)
	for i := 0; i < n; i++ {
		out[i] = models.WeatherData{Timestamp: start.Add(time.Duration(i) * time.Hour)}
	}
	return out
}

func TestSpinUpAndIntegrate_Empty(t *testing.T) {
	got := SpinUpAndIntegrate(nil, nil, func(int, models.WeatherData) float64 { return 70 }, 105)
	if len(got) != 0 {
		t.Errorf("expected empty output, got len=%d", len(got))
	}
}

func TestSpinUpAndIntegrate_ConstantEquilibriumStaysAtEquilibrium(t *testing.T) {
	start := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	past := mkHourly(12, start)
	fcst := mkHourly(24, start.Add(12*time.Hour))
	eqFn := func(int, models.WeatherData) float64 { return 70 }
	out := SpinUpAndIntegrate(past, fcst, eqFn, 105)
	if len(out) != 36 {
		t.Fatalf("length mismatch: got %d, want 36", len(out))
	}
	for i, v := range out {
		if math.Abs(v-70) > 1e-9 {
			t.Errorf("idx %d: got %.6f, want 70", i, v)
		}
	}
}

func TestSpinUpAndIntegrate_StepResponse(t *testing.T) {
	// 12h of past at T_eq=50, then 24h of forecast at T_eq=80.
	// At idx 11 (last past hour) rock should still be 50.
	// At idx 12 it begins migrating toward 80; after enough hours it
	// asymptotes near 80.
	start := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	past := mkHourly(12, start)
	fcst := mkHourly(24, start.Add(12*time.Hour))
	eqFn := func(idx int, _ models.WeatherData) float64 {
		if idx < 12 {
			return 50
		}
		return 80
	}
	out := SpinUpAndIntegrate(past, fcst, eqFn, 60) // τ=60min for clarity
	if math.Abs(out[11]-50) > 1e-9 {
		t.Errorf("idx 11 should still be 50, got %.6f", out[11])
	}
	// After one τ (idx 12 = first forecast hour, one 60min step):
	want12 := 50 + 30*(1-1/math.E)
	if math.Abs(out[12]-want12) > 0.01 {
		t.Errorf("idx 12 got %.4f, want %.4f", out[12], want12)
	}
	// After ~5τ of forecast (idx 16), should be within 1°F of 80.
	if math.Abs(out[16]-80) > 1.0 {
		t.Errorf("idx 16 should be near 80, got %.4f", out[16])
	}
	// Strictly monotonic increase from 12 onward.
	for i := 13; i < len(out); i++ {
		if out[i] <= out[i-1] {
			t.Errorf("idx %d not strictly increasing: %.4f -> %.4f", i, out[i-1], out[i])
			break
		}
	}
}

func TestSpinUpAndIntegrate_LengthMatchesInputs(t *testing.T) {
	start := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	past := mkHourly(3, start)
	fcst := mkHourly(7, start.Add(3*time.Hour))
	out := SpinUpAndIntegrate(past, fcst, func(int, models.WeatherData) float64 { return 60 }, 105)
	if len(out) != 10 {
		t.Errorf("length got %d, want 10", len(out))
	}
}

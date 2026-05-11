package rock_temp

import (
	"strings"
	"testing"
)

func TestComputeConfidence_AllKnown(t *testing.T) {
	r := ComputeConfidence(ConfidenceInputs{
		AspectKnown: true, DipKnown: true, RockTypeKnown: true,
		SpinUpComplete: true,
	})
	if r.Score != 100 {
		t.Errorf("all known: got %d want 100", r.Score)
	}
	if len(r.Factors) != 0 {
		t.Errorf("no factors expected, got %v", r.Factors)
	}
}

func TestComputeConfidence_AspectDefaulted(t *testing.T) {
	r := ComputeConfidence(ConfidenceInputs{
		AspectKnown: false, DipKnown: true, RockTypeKnown: true, SpinUpComplete: true,
	})
	// 100 - 15 (no facet ≥60% dominance) = 85
	if r.Score != 85 {
		t.Errorf("aspect defaulted alone: expected 85, got %d", r.Score)
	}
	found := false
	for _, f := range r.Factors {
		if strings.Contains(f, "mixed sun exposure") {
			found = true
		}
	}
	if !found {
		t.Errorf("missing aspect factor mentioning mixed sun exposure: %v", r.Factors)
	}
}

func TestComputeConfidence_DipDefaulted(t *testing.T) {
	r := ComputeConfidence(ConfidenceInputs{
		AspectKnown: true, DipKnown: false, RockTypeKnown: true, SpinUpComplete: true,
	})
	// 100 - 10 (slab/overhang not dominant) = 90
	if r.Score != 90 {
		t.Errorf("dip defaulted: expected 90, got %d", r.Score)
	}
	found := false
	for _, f := range r.Factors {
		if strings.Contains(f, "vertical walls") && strings.Contains(f, "slab") && strings.Contains(f, "overhang") {
			found = true
		}
	}
	if !found {
		t.Errorf("missing dip factor mentioning vertical walls / slab / overhang: %v", r.Factors)
	}
}

func TestComputeConfidence_RockTypeDefaulted(t *testing.T) {
	r := ComputeConfidence(ConfidenceInputs{
		AspectKnown: true, DipKnown: true, RockTypeKnown: false, SpinUpComplete: true,
	})
	if r.Score != 95 {
		t.Errorf("rock type defaulted: expected 95, got %d", r.Score)
	}
	found := false
	for _, f := range r.Factors {
		if strings.Contains(f, "default rock properties") && strings.Contains(f, "granite") {
			found = true
		}
	}
	if !found {
		t.Errorf("rock type factor should mention default rock properties (granite): %v", r.Factors)
	}
}

func TestComputeConfidence_FloorAt30(t *testing.T) {
	// All deductions piled on; should clamp at MinConfidence (30).
	r := ComputeConfidence(ConfidenceInputs{
		AspectKnown: false, DipKnown: false, RockTypeKnown: false,
		MixedFacets: true, ForecastHorizonH: 168,
		CloudVariableHigh: true, WindVariableHigh: true, SpinUpComplete: false,
		NoSunExposureRow: true,
		Mode:             ConfidenceModeHour,
	})
	if r.Score != 30 {
		t.Errorf("floor: got %d want 30", r.Score)
	}
}

func TestComputeConfidence_Forecast48h_HourMode(t *testing.T) {
	r := ComputeConfidence(ConfidenceInputs{
		AspectKnown: true, DipKnown: true, RockTypeKnown: true,
		ForecastHorizonH: 48, SpinUpComplete: true,
		Mode: ConfidenceModeHour,
	})
	// 48h is in the 24..72 tier → -5
	if r.Score != 95 {
		t.Errorf("48h forecast (hour mode): got %d want 95", r.Score)
	}
}

func TestComputeConfidence_SpinUpMissing(t *testing.T) {
	r := ComputeConfidence(ConfidenceInputs{
		AspectKnown: true, DipKnown: true, RockTypeKnown: true, SpinUpComplete: false,
	})
	if r.Score != 95 {
		t.Errorf("missing spin-up: got %d want 95", r.Score)
	}
}

func TestComputeConfidence_MixedFacets(t *testing.T) {
	// Aspect known + mixed → -8 mixed facets reason fires.
	r := ComputeConfidence(ConfidenceInputs{
		AspectKnown: true, DipKnown: true, RockTypeKnown: true,
		MixedFacets: true, SpinUpComplete: true,
	})
	if r.Score != 92 {
		t.Errorf("mixed facets: got %d want 92", r.Score)
	}
}

// New tests --------------------------------------------------------------

// TestComputeConfidence_StatusModeOmitsHorizonPenalty: in status mode
// the long-horizon penalty must not fire and must not appear in factors.
func TestComputeConfidence_StatusModeOmitsHorizonPenalty(t *testing.T) {
	r := ComputeConfidence(ConfidenceInputs{
		AspectKnown: true, DipKnown: true, RockTypeKnown: true,
		ForecastHorizonH: 168, SpinUpComplete: true,
		Mode: ConfidenceModeStatus,
	})
	if r.Score != 100 {
		t.Errorf("status mode ignores horizon: got %d want 100", r.Score)
	}
	for _, f := range r.Factors {
		if strings.Contains(f, "Forecast") && strings.Contains(f, "h out") {
			t.Errorf("status mode should not include horizon factor; got %q", f)
		}
	}
}

// TestComputeConfidence_HourModeAppliesHorizonPenaltyByDistance: increasing
// horizons produce non-decreasing penalties matching the tier table.
func TestComputeConfidence_HourModeAppliesHorizonPenaltyByDistance(t *testing.T) {
	cases := []struct {
		hours     int
		wantScore int
	}{
		{24, 100}, // 0..24h → no penalty
		{72, 95},  // 24..72h → -5
		{120, 90}, // 72..120h → -10
		{168, 85}, // 120..168h → -15
	}
	for _, c := range cases {
		r := ComputeConfidence(ConfidenceInputs{
			AspectKnown: true, DipKnown: true, RockTypeKnown: true,
			SpinUpComplete:   true,
			ForecastHorizonH: c.hours,
			Mode:             ConfidenceModeHour,
		})
		if r.Score != c.wantScore {
			t.Errorf("hour=%d: got score %d want %d (factors=%v)", c.hours, r.Score, c.wantScore, r.Factors)
		}
	}
}

// TestComputeConfidence_NoDuplicateMixedFacetsReason: when MixedFacets
// is true alongside known aspect, only ONE mixed/multiple-facets reason
// is emitted (not two from separate code paths).
func TestComputeConfidence_NoDuplicateMixedFacetsReason(t *testing.T) {
	r := ComputeConfidence(ConfidenceInputs{
		AspectKnown: true, DipKnown: true, RockTypeKnown: true,
		MixedFacets: true, SpinUpComplete: true,
		Mode: ConfidenceModeStatus,
	})
	count := 0
	for _, f := range r.Factors {
		lf := strings.ToLower(f)
		if strings.Contains(lf, "varied terrain") || strings.Contains(lf, "mixed sun exposure") {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 mixed/facets reason, got %d (factors=%v)", count, r.Factors)
	}
}

// TestComputeConfidence_TypicalSpreadFacets_NotPinnedAtFloor: the audit
// shows all 34 active locations have a location_sun_exposure row, but
// most have facets spread across multiple compass directions (no single
// face reaches 60%). That realistic case should land comfortably above
// the floor, not be pinned at 30.
func TestComputeConfidence_TypicalSpreadFacets_NotPinnedAtFloor(t *testing.T) {
	r := ComputeConfidence(ConfidenceInputs{
		// Row exists but no facet dominates → AspectKnown/DipKnown false,
		// MixedFacets true. NoSunExposureRow stays false because the row
		// is present.
		AspectKnown: false, DipKnown: false, RockTypeKnown: true,
		MixedFacets:      true,
		NoSunExposureRow: false,
		SpinUpComplete:   true,
		Mode:             ConfidenceModeStatus,
	})
	// 100 - 15 (no aspect dominance) - 10 (no dip dominance) - 8 (mixed) = 67
	if r.Score < 60 || r.Score > 75 {
		t.Errorf("typical spread-facets location: expected score in [60,75], got %d (factors=%v)", r.Score, r.Factors)
	}
	if r.Score == MinConfidence {
		t.Errorf("score should not be pinned at floor (%d) for a typical case", MinConfidence)
	}
	for _, f := range r.Factors {
		if strings.Contains(f, "Sun exposure data not yet set") {
			t.Errorf("row-exists case must not emit the no-row reason; got %q", f)
		}
	}
}

// TestComputeConfidence_NoSunExposureRow: when the row is missing
// entirely, emit a distinct stronger reason that points at the seed
// template, and suppress the per-field aspect/dip reasons.
func TestComputeConfidence_NoSunExposureRow(t *testing.T) {
	r := ComputeConfidence(ConfidenceInputs{
		AspectKnown: false, DipKnown: false, RockTypeKnown: true,
		NoSunExposureRow: true,
		SpinUpComplete:   true,
		Mode:             ConfidenceModeStatus,
	})
	// 100 - 25 (no row) = 75
	if r.Score != 75 {
		t.Errorf("no-row case: expected 75, got %d (factors=%v)", r.Score, r.Factors)
	}
	foundNoRow := false
	for _, f := range r.Factors {
		if strings.Contains(f, "Sun exposure data not yet set") {
			foundNoRow = true
		}
		if strings.Contains(f, "mixed sun exposure") || strings.Contains(f, "vertical walls") {
			t.Errorf("per-field reasons should be suppressed when NoSunExposureRow=true; got %q", f)
		}
	}
	if !foundNoRow {
		t.Errorf("expected no-row reason in plain-English form; factors=%v", r.Factors)
	}
}

// TestComputeConfidence_RowExistsButNoDominantFacet: with all four
// _facing_percent values near 25%, the row exists but no facet
// dominates. The "no facet reaches 60% dominance" reason fires; the
// "no row" reason does NOT.
func TestComputeConfidence_RowExistsButNoDominantFacet(t *testing.T) {
	r := ComputeConfidence(ConfidenceInputs{
		AspectKnown: false, DipKnown: false, RockTypeKnown: true,
		NoSunExposureRow: false, // row present
		SpinUpComplete:   true,
		Mode:             ConfidenceModeStatus,
	})
	// 100 - 15 (aspect) - 10 (dip) = 75
	if r.Score != 75 {
		t.Errorf("row-exists-no-dominance: expected 75, got %d (factors=%v)", r.Score, r.Factors)
	}
	foundAspect := false
	foundDip := false
	for _, f := range r.Factors {
		if strings.Contains(f, "mixed sun exposure") {
			foundAspect = true
		}
		if strings.Contains(f, "vertical walls") && strings.Contains(f, "slab") && strings.Contains(f, "overhang") {
			foundDip = true
		}
		if strings.Contains(f, "Sun exposure data not yet set") {
			t.Errorf("row-exists case must not emit no-row reason; got %q", f)
		}
	}
	if !foundAspect {
		t.Errorf("expected aspect mixed-sun-exposure reason; factors=%v", r.Factors)
	}
	if !foundDip {
		t.Errorf("expected dip reason mentioning vertical walls / slab / overhang; factors=%v", r.Factors)
	}
}

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
	if r.Score > 75 {
		t.Errorf("aspect defaulted: score should be ≤75, got %d", r.Score)
	}
	if r.Score != 75 {
		t.Errorf("aspect defaulted alone: expected 75, got %d", r.Score)
	}
	found := false
	for _, f := range r.Factors {
		if strings.Contains(f, "aspect") {
			found = true
		}
	}
	if !found {
		t.Errorf("missing aspect factor: %v", r.Factors)
	}
}

func TestComputeConfidence_DipDefaulted(t *testing.T) {
	r := ComputeConfidence(ConfidenceInputs{
		AspectKnown: true, DipKnown: false, RockTypeKnown: true, SpinUpComplete: true,
	})
	if r.Score != 85 {
		t.Errorf("dip defaulted: expected 85, got %d", r.Score)
	}
}

func TestComputeConfidence_FloorAt20(t *testing.T) {
	r := ComputeConfidence(ConfidenceInputs{
		AspectKnown: false, DipKnown: false, RockTypeKnown: false,
		MixedFacets: true, ForecastHorizonH: 168, // 7 days * 5 = 35
		CloudVariableHigh: true, WindVariableHigh: true, SpinUpComplete: false,
	})
	if r.Score != 20 {
		t.Errorf("floor: got %d want 20", r.Score)
	}
}

func TestComputeConfidence_Forecast48h(t *testing.T) {
	r := ComputeConfidence(ConfidenceInputs{
		AspectKnown: true, DipKnown: true, RockTypeKnown: true,
		ForecastHorizonH: 48, SpinUpComplete: true,
	})
	// 48h → floor(48/24)*5 = 10
	if r.Score != 90 {
		t.Errorf("48h forecast: got %d want 90", r.Score)
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
	r := ComputeConfidence(ConfidenceInputs{
		AspectKnown: true, DipKnown: true, RockTypeKnown: true,
		MixedFacets: true, SpinUpComplete: true,
	})
	if r.Score != 92 {
		t.Errorf("mixed facets: got %d want 92", r.Score)
	}
}

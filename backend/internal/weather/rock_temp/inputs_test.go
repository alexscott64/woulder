package rock_temp

import (
	"math"
	"testing"

	"github.com/alexscott64/woulder/backend/internal/models"
)

func TestWeightedAspect_AllSouth(t *testing.T) {
	az, ok := WeightedAspect(100, 0, 0, 0)
	if !ok {
		t.Fatalf("expected ok")
	}
	if math.Abs(az-180) > 1e-6 {
		t.Errorf("100%% south: got %.4f want 180", az)
	}
}

func TestWeightedAspect_Degenerate(t *testing.T) {
	// 50% east + 50% west cancels.
	if _, ok := WeightedAspect(0, 50, 50, 0); ok {
		t.Errorf("E+W cancellation should be degenerate")
	}
	// 50% north + 50% south cancels.
	if _, ok := WeightedAspect(50, 0, 0, 50); ok {
		t.Errorf("N+S cancellation should be degenerate")
	}
}

func TestWeightedAspect_75South25West(t *testing.T) {
	az, ok := WeightedAspect(75, 25, 0, 0)
	if !ok {
		t.Fatalf("expected ok")
	}
	// Computed analytically: ~198.43°, between south (180) and west (270),
	// closer to south. Tolerance is generous.
	if az < 180 || az > 220 {
		t.Errorf("75%%S+25%%W: got %.4f, expected ~198", az)
	}
}

func TestWeightedAspect_AllNorth(t *testing.T) {
	az, ok := WeightedAspect(0, 0, 0, 100)
	if !ok {
		t.Fatalf("expected ok")
	}
	if math.Abs(az) > 1e-6 && math.Abs(az-360) > 1e-6 {
		t.Errorf("100%% north: got %.4f want 0", az)
	}
}

func TestWeightedDip(t *testing.T) {
	cases := []struct {
		slab, overhang, want float64
	}{
		{100, 0, 45},   // all slab
		{0, 100, 110},  // all overhang
		{0, 0, 90},     // all vertical
		{50, 0, 67.5},  // 50% slab + 50% vertical
		{50, 50, 77.5}, // 50/50 slab/overhang
	}
	for _, c := range cases {
		got := WeightedDip(c.slab, c.overhang)
		if math.Abs(got-c.want) > 1e-6 {
			t.Errorf("dip slab=%.0f over=%.0f: got %.4f want %.4f", c.slab, c.overhang, got, c.want)
		}
	}
}

func TestResolveDominantFacet_Nil(t *testing.T) {
	d := ResolveDominantFacet(nil)
	if d.AspectDeg != 180 || d.DipDeg != 90 || !d.Mixed {
		t.Errorf("nil sun exposure: got %+v", d)
	}
	if d.Reason == "" {
		t.Errorf("expected reason string")
	}
}

func TestResolveDominantFacet_DominantSouth(t *testing.T) {
	se := &models.LocationSunExposure{
		SouthFacingPercent: 80, WestFacingPercent: 10, EastFacingPercent: 10, NorthFacingPercent: 0,
		SlabPercent: 0, OverhangPercent: 0,
	}
	d := ResolveDominantFacet(se)
	if d.AspectDeg != 180 {
		t.Errorf("south>60 should give aspect=180, got %.2f", d.AspectDeg)
	}
	if d.Mixed {
		t.Errorf("south dominant should not be mixed")
	}
}

func TestResolveDominantFacet_MixedAspectDominantSlab(t *testing.T) {
	se := &models.LocationSunExposure{
		SouthFacingPercent: 40, WestFacingPercent: 30, EastFacingPercent: 30, NorthFacingPercent: 0,
		SlabPercent: 80, OverhangPercent: 0,
	}
	d := ResolveDominantFacet(se)
	if !d.Mixed {
		t.Errorf("expected mixed aspect")
	}
	if d.DipDeg != 45 {
		t.Errorf("slab>60 should give dip=45, got %.2f", d.DipDeg)
	}
}

func TestResolveDominantFacet_DominantOverhang(t *testing.T) {
	se := &models.LocationSunExposure{
		SouthFacingPercent: 80, OverhangPercent: 80,
	}
	d := ResolveDominantFacet(se)
	if d.DipDeg != 110 {
		t.Errorf("overhang>60 should give dip=110, got %.2f", d.DipDeg)
	}
}

func TestTreeFraction(t *testing.T) {
	if got := TreeFraction(nil); got != 0 {
		t.Errorf("nil: got %.4f want 0", got)
	}
	se := &models.LocationSunExposure{TreeCoveragePercent: 60}
	if got := TreeFraction(se); math.Abs(got-0.6) > 1e-9 {
		t.Errorf("60%%: got %.4f want 0.6", got)
	}
	se2 := &models.LocationSunExposure{TreeCoveragePercent: 150}
	if got := TreeFraction(se2); got != 1 {
		t.Errorf("150%% clamp: got %.4f want 1", got)
	}
	se3 := &models.LocationSunExposure{TreeCoveragePercent: -10}
	if got := TreeFraction(se3); got != 0 {
		t.Errorf("-10%% clamp: got %.4f want 0", got)
	}
}

func TestResolveRockTypeGroup(t *testing.T) {
	// Override wins.
	if g, ok := ResolveRockTypeGroup(nil, "Sandstone"); g != "Sandstone" || !ok {
		t.Errorf("override: got (%q,%v)", g, ok)
	}
	// First row primary.
	rt := []models.RockType{{GroupName: "Limestone"}}
	if g, ok := ResolveRockTypeGroup(rt, ""); g != "Limestone" || !ok {
		t.Errorf("first row: got (%q,%v)", g, ok)
	}
	// Empty.
	if g, ok := ResolveRockTypeGroup(nil, ""); g != "Granite" || ok {
		t.Errorf("empty: got (%q,%v) want (Granite,false)", g, ok)
	}
	// First row has empty group name → fallback.
	rtEmpty := []models.RockType{{GroupName: ""}}
	if g, ok := ResolveRockTypeGroup(rtEmpty, ""); g != "Granite" || ok {
		t.Errorf("empty group name: got (%q,%v) want (Granite,false)", g, ok)
	}
}

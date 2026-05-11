package rock_temp

import (
	"testing"
	"time"
)

func TestClassifyCondensation(t *testing.T) {
	cases := []struct {
		surf, dew float64
		want      string
	}{
		{60, 57, "none"},  // diff=3
		{60, 59, "light"}, // diff=1
		{60, 58, "light"}, // diff=2 → light by spec
		{60, 60, "heavy"}, // diff=0 → heavy
		{60, 61, "heavy"}, // diff=-1
		{50, 55, "heavy"}, // diff=-5
	}
	for _, c := range cases {
		got := ClassifyCondensation(c.surf, c.dew)
		if got != c.want {
			t.Errorf("surf=%.1f dew=%.1f: got %q want %q", c.surf, c.dew, got, c.want)
		}
	}
}

func TestFindClearsAt(t *testing.T) {
	base := time.Date(2026, 1, 1, 5, 0, 0, 0, time.UTC)
	hours := []HourPoint{
		{Time: base.Add(0 * time.Hour), SurfaceF: 50, DewpointF: 55}, // condensing
		{Time: base.Add(1 * time.Hour), SurfaceF: 52, DewpointF: 55}, // condensing
		{Time: base.Add(2 * time.Hour), SurfaceF: 56, DewpointF: 55}, // diff=1, NOT cleared (need >=2)
		{Time: base.Add(3 * time.Hour), SurfaceF: 58, DewpointF: 55}, // diff=3, cleared
		{Time: base.Add(4 * time.Hour), SurfaceF: 60, DewpointF: 55},
	}
	got := FindClearsAt(hours, 0)
	if got == nil {
		t.Fatalf("expected ClearsAt to find an hour")
	}
	want := base.Add(3 * time.Hour)
	if !got.Equal(want) {
		t.Errorf("clears_at: got %v want %v", got, want)
	}
}

func TestFindClearsAt_NeverClears(t *testing.T) {
	base := time.Date(2026, 1, 1, 5, 0, 0, 0, time.UTC)
	hours := []HourPoint{
		{Time: base, SurfaceF: 50, DewpointF: 55},
		{Time: base.Add(time.Hour), SurfaceF: 51, DewpointF: 55},
	}
	if got := FindClearsAt(hours, 0); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestCondensationReason(t *testing.T) {
	// Air at saturation: humidity high.
	if r := CondensationReason(50, 55, 56, 0, 99); r != "Air at saturation (fog conditions)" {
		t.Errorf("humidity 99 → %q", r)
	}
	// Air at saturation: tiny air-dewpoint spread.
	if r := CondensationReason(50, 55, 56, 0, 80); r != "Air at saturation (fog conditions)" {
		t.Errorf("spread 1 → %q", r)
	}
	// Wind cooling.
	if r := CondensationReason(50, 55, 65, 10, 60); r != "Rock cooled by wind below dewpoint" {
		t.Errorf("wind 10 + cold rock → %q", r)
	}
	// Default predawn case.
	if r := CondensationReason(54, 55, 60, 2, 70); r != "Cold rock surface + humid air" {
		t.Errorf("default → %q", r)
	}
}

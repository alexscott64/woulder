package rock_temp

import (
	"testing"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

func makeHours(start time.Time, conditions []string, surf []float64, condensing []bool) []models.RockTempHour {
	out := make([]models.RockTempHour, len(conditions))
	for i := range conditions {
		out[i] = models.RockTempHour{
			Time:       start.Add(time.Duration(i) * time.Hour),
			SurfaceF:   surf[i],
			Condition:  conditions[i],
			Condensing: condensing[i],
		}
	}
	return out
}

func TestDetectSendWindows_MorningPrime(t *testing.T) {
	base := time.Date(2026, 6, 1, 6, 0, 0, 0, time.UTC)
	conds := make([]string, 24)
	surf := make([]float64, 24)
	cond := make([]bool, 24)
	for i := 0; i < 24; i++ {
		switch {
		case i < 4: // 6,7,8,9 prime
			conds[i] = "prime"
			surf[i] = 50
		case i < 8: // 10..13 good
			conds[i] = "good"
			surf[i] = 60
		default:
			conds[i] = "poor"
			surf[i] = 90
		}
	}
	hours := makeHours(base, conds, surf, cond)
	wins := DetectSendWindows(hours, SendWindowOptions{})
	// Expect: one prime window (4h) and one good-or-better window (8h covering both).
	var prime, good *models.SendWindow
	for i := range wins {
		w := wins[i]
		switch w.Condition {
		case "prime":
			prime = &wins[i]
		case "good":
			good = &wins[i]
		}
	}
	if prime == nil {
		t.Fatalf("expected prime window")
	}
	if prime.DurationH != 4 {
		t.Errorf("prime duration: got %.2f want 4", prime.DurationH)
	}
	if !prime.DryThroughout {
		t.Errorf("prime should be dry")
	}
	if good == nil {
		t.Fatalf("expected good-or-better window")
	}
	if good.DurationH != 8 {
		t.Errorf("good-or-better duration: got %.2f want 8", good.DurationH)
	}
}

func TestDetectSendWindows_FiltersShortWindows(t *testing.T) {
	base := time.Date(2026, 6, 1, 6, 0, 0, 0, time.UTC)
	// Only 1 hour of prime → below 1.5h default → filtered.
	conds := []string{"prime", "poor", "poor"}
	surf := []float64{50, 90, 90}
	cond := []bool{false, false, false}
	hours := makeHours(base, conds, surf, cond)
	wins := DetectSendWindows(hours, SendWindowOptions{})
	if len(wins) != 0 {
		t.Errorf("expected no windows under 1.5h, got %d", len(wins))
	}
}

func TestDetectSendWindows_AllPoor(t *testing.T) {
	base := time.Date(2026, 6, 1, 6, 0, 0, 0, time.UTC)
	conds := make([]string, 24)
	surf := make([]float64, 24)
	cond := make([]bool, 24)
	for i := range conds {
		conds[i] = "poor"
		surf[i] = 90
	}
	hours := makeHours(base, conds, surf, cond)
	if wins := DetectSendWindows(hours, SendWindowOptions{}); len(wins) != 0 {
		t.Errorf("all-poor: got %d windows", len(wins))
	}
}

func TestDetectSendWindows_AllGoodOneWindow(t *testing.T) {
	base := time.Date(2026, 6, 1, 6, 0, 0, 0, time.UTC)
	conds := make([]string, 12)
	surf := make([]float64, 12)
	cond := make([]bool, 12)
	for i := range conds {
		conds[i] = "good"
		surf[i] = 60
	}
	hours := makeHours(base, conds, surf, cond)
	wins := DetectSendWindows(hours, SendWindowOptions{})
	if len(wins) != 1 {
		t.Fatalf("expected exactly 1 good-or-better window, got %d", len(wins))
	}
	if wins[0].Condition != "good" || wins[0].DurationH != 12 {
		t.Errorf("got %+v", wins[0])
	}
}

func TestDetectSendWindows_PrimeContainedInGoodEmitsBoth(t *testing.T) {
	base := time.Date(2026, 6, 1, 6, 0, 0, 0, time.UTC)
	// good, prime, prime, good → prime span (2h) is shorter than good span (4h)
	conds := []string{"good", "prime", "prime", "good"}
	surf := []float64{60, 50, 50, 60}
	cond := []bool{false, false, false, false}
	hours := makeHours(base, conds, surf, cond)
	wins := DetectSendWindows(hours, SendWindowOptions{})
	hasPrime := false
	hasGood := false
	for _, w := range wins {
		if w.Condition == "prime" && w.DurationH == 2 {
			hasPrime = true
		}
		if w.Condition == "good" && w.DurationH == 4 {
			hasGood = true
		}
	}
	if !hasPrime {
		t.Errorf("expected prime sub-window, got %+v", wins)
	}
	if !hasGood {
		t.Errorf("expected good-or-better wrapping window, got %+v", wins)
	}
}

func TestDetectSendWindows_CondensingMarksDryFalse(t *testing.T) {
	base := time.Date(2026, 6, 1, 6, 0, 0, 0, time.UTC)
	conds := []string{"prime", "prime", "prime"}
	surf := []float64{50, 50, 50}
	cond := []bool{true, false, false} // first hour condensing
	hours := makeHours(base, conds, surf, cond)
	wins := DetectSendWindows(hours, SendWindowOptions{})
	if len(wins) == 0 {
		t.Fatalf("expected window")
	}
	for _, w := range wins {
		if w.DryThroughout {
			t.Errorf("condensing hour should mark dry_throughout=false, got %+v", w)
		}
	}
}

func TestDetectSendWindows_AvgAndPeak(t *testing.T) {
	base := time.Date(2026, 6, 1, 6, 0, 0, 0, time.UTC)
	conds := []string{"prime", "prime"}
	surf := []float64{50, 60}
	cond := []bool{false, false}
	hours := makeHours(base, conds, surf, cond)
	wins := DetectSendWindows(hours, SendWindowOptions{})
	if len(wins) != 1 {
		t.Fatalf("expected 1 window, got %d", len(wins))
	}
	w := wins[0]
	if w.AvgTempF != 55 {
		t.Errorf("avg: got %.2f want 55", w.AvgTempF)
	}
	if w.PeakTempF != 60 {
		t.Errorf("peak: got %.2f want 60", w.PeakTempF)
	}
}

func TestNextTransition(t *testing.T) {
	base := time.Date(2026, 6, 1, 6, 0, 0, 0, time.UTC)
	hours := makeHours(base,
		[]string{"prime", "prime", "good", "poor"},
		[]float64{50, 50, 60, 90},
		[]bool{false, false, false, false},
	)
	tr := NextTransition(hours, "prime")
	if tr == nil {
		t.Fatalf("expected transition")
	}
	if tr.ToCondition != "good" {
		t.Errorf("transition: got %q want good", tr.ToCondition)
	}
	want := base.Add(2 * time.Hour)
	if !tr.Time.Equal(want) {
		t.Errorf("transition time: got %v want %v", tr.Time, want)
	}
}

func TestNextTransition_NoChange(t *testing.T) {
	base := time.Date(2026, 6, 1, 6, 0, 0, 0, time.UTC)
	hours := makeHours(base,
		[]string{"prime", "prime", "prime"},
		[]float64{50, 50, 50},
		[]bool{false, false, false},
	)
	if tr := NextTransition(hours, "prime"); tr != nil {
		t.Errorf("expected nil, got %+v", tr)
	}
}

func TestDetectSendWindows_MinDurationConfigurable(t *testing.T) {
	base := time.Date(2026, 6, 1, 6, 0, 0, 0, time.UTC)
	// 2 hours of prime should be filtered when min=3h.
	conds := []string{"prime", "prime", "poor"}
	surf := []float64{50, 50, 90}
	cond := []bool{false, false, false}
	hours := makeHours(base, conds, surf, cond)
	wins := DetectSendWindows(hours, SendWindowOptions{MinDurationH: 3})
	if len(wins) != 0 {
		t.Errorf("expected 0 windows with min=3h, got %d", len(wins))
	}
	wins2 := DetectSendWindows(hours, SendWindowOptions{MinDurationH: 1})
	if len(wins2) != 1 {
		t.Errorf("expected 1 window with min=1h, got %d", len(wins2))
	}
}

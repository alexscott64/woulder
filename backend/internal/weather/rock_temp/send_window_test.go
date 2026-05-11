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
	wins := DetectSendWindows(hours, "UTC", SendWindowOptions{})
	// Expect: one prime window (4h) and one good-or-better window (8h covering both).
	// Good overlaps prime by 4h/8h = 50% → not subsumed.
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
	wins := DetectSendWindows(hours, "UTC", SendWindowOptions{})
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
	if wins := DetectSendWindows(hours, "UTC", SendWindowOptions{}); len(wins) != 0 {
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
	wins := DetectSendWindows(hours, "UTC", SendWindowOptions{})
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
	// Good overlaps prime by 2h/4h = 50% → not subsumed.
	conds := []string{"good", "prime", "prime", "good"}
	surf := []float64{60, 50, 50, 60}
	cond := []bool{false, false, false, false}
	hours := makeHours(base, conds, surf, cond)
	wins := DetectSendWindows(hours, "UTC", SendWindowOptions{})
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
	wins := DetectSendWindows(hours, "UTC", SendWindowOptions{})
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
	wins := DetectSendWindows(hours, "UTC", SendWindowOptions{})
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
	wins := DetectSendWindows(hours, "UTC", SendWindowOptions{MinDurationH: 3})
	if len(wins) != 0 {
		t.Errorf("expected 0 windows with min=3h, got %d", len(wins))
	}
	wins2 := DetectSendWindows(hours, "UTC", SendWindowOptions{MinDurationH: 1})
	if len(wins2) != 1 {
		t.Errorf("expected 1 window with min=1h, got %d", len(wins2))
	}
}

// --- New tests covering the dedupe / midnight-split / subsumption fixes. ---

// TestDetectSendWindows_NoSubsetEmission verifies that a single 12h
// contiguous prime run produces exactly 1 window (not 12 windows for
// every starting hour within the run).
func TestDetectSendWindows_NoSubsetEmission(t *testing.T) {
	// Start at midnight so the entire 12h run sits within one local day.
	base := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	conds := make([]string, 12)
	surf := make([]float64, 12)
	cond := make([]bool, 12)
	for i := range conds {
		conds[i] = "prime"
		surf[i] = 55
	}
	hours := makeHours(base, conds, surf, cond)
	wins := DetectSendWindows(hours, "UTC", SendWindowOptions{})
	primeCount := 0
	for _, w := range wins {
		if w.Condition == "prime" {
			primeCount++
		}
	}
	if primeCount != 1 {
		t.Fatalf("expected exactly 1 prime window for a 12h contiguous run, got %d (windows=%+v)", primeCount, wins)
	}
}

// TestDetectSendWindows_MidnightSplit verifies a run spanning 9 PM Sat
// → 11 AM Mon (38 hours) is split into 3 windows: Sat partial, full
// Sunday, Mon partial. Each ≤24h, none crossing local midnight.
func TestDetectSendWindows_MidnightSplit(t *testing.T) {
	// Saturday 2026-06-06 21:00 local (UTC for simplicity).
	base := time.Date(2026, 6, 6, 21, 0, 0, 0, time.UTC)
	const n = 38 // 21:00 Sat .. 10:00 Mon (last hour) → covers up to 11:00 Mon
	conds := make([]string, n)
	surf := make([]float64, n)
	cond := make([]bool, n)
	for i := range conds {
		conds[i] = "prime"
		surf[i] = 55
	}
	hours := makeHours(base, conds, surf, cond)
	wins := DetectSendWindows(hours, "UTC", SendWindowOptions{})
	primeCount := 0
	for _, w := range wins {
		if w.Condition == "prime" {
			primeCount++
		}
		if w.DurationH > 24 {
			t.Errorf("window exceeds 24h: %+v", w)
		}
	}
	if primeCount != 3 {
		t.Fatalf("expected 3 prime windows after midnight split, got %d (%+v)", primeCount, wins)
	}
	// Expect Sat partial = 3h (21,22,23), Sun full = 24h (00..23), Mon partial = 11h (00..10).
	gotDur := []float64{}
	for _, w := range wins {
		if w.Condition == "prime" {
			gotDur = append(gotDur, w.DurationH)
		}
	}
	wantDur := []float64{3, 24, 11}
	if len(gotDur) != len(wantDur) {
		t.Fatalf("durations: got %v want %v", gotDur, wantDur)
	}
	for i := range gotDur {
		if gotDur[i] != wantDur[i] {
			t.Errorf("duration[%d]: got %.1f want %.1f", i, gotDur[i], wantDur[i])
		}
	}
}

// TestDetectSendWindows_MinDurationAfterSplit verifies that a midnight
// split which produces a sub-30-minute fragment drops that fragment.
// We construct a 2h prime run that crosses midnight at the 1h mark,
// producing two 1h sub-windows — both below the 1.5h default → both
// dropped.
func TestDetectSendWindows_MinDurationAfterSplit(t *testing.T) {
	// 23:00 .. 00:00 (1h) + 00:00 .. 01:00 (1h) → split into 1h + 1h.
	base := time.Date(2026, 6, 1, 23, 0, 0, 0, time.UTC)
	hours := makeHours(base,
		[]string{"prime", "prime"},
		[]float64{55, 55},
		[]bool{false, false},
	)
	wins := DetectSendWindows(hours, "UTC", SendWindowOptions{})
	if len(wins) != 0 {
		t.Errorf("expected 0 windows (both halves <1.5h), got %d: %+v", len(wins), wins)
	}

	// Same data but with a 1h minimum → expect 2 split windows.
	wins2 := DetectSendWindows(hours, "UTC", SendWindowOptions{MinDurationH: 1})
	if len(wins2) != 2 {
		t.Fatalf("expected 2 split windows with min=1h, got %d: %+v", len(wins2), wins2)
	}
	for _, w := range wins2 {
		if w.DurationH != 1 {
			t.Errorf("expected each split window to be 1h, got %+v", w)
		}
	}
}

// TestDetectSendWindows_SubsumedGoodDropped verifies that a "good"
// window where ≥80% of its duration overlaps a "prime" window on the
// same day is dropped.
func TestDetectSendWindows_SubsumedGoodDropped(t *testing.T) {
	base := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	// Hours: 5h "good" then 4h "prime" then 1h "good" — one contiguous
	// good-or-better run of 10h (00..09); prime sub-run is 4h (05..08).
	// To make prime ≥80% of good, build instead:
	// 1h good, 8h prime, 1h good → good run = 10h, prime run = 8h → 80% overlap.
	conds := make([]string, 10)
	surf := make([]float64, 10)
	cond := make([]bool, 10)
	for i := 0; i < 10; i++ {
		switch {
		case i == 0 || i == 9:
			conds[i] = "good"
			surf[i] = 65
		default:
			conds[i] = "prime"
			surf[i] = 55
		}
	}
	hours := makeHours(base, conds, surf, cond)
	wins := DetectSendWindows(hours, "UTC", SendWindowOptions{})
	for _, w := range wins {
		if w.Condition == "good" {
			t.Errorf("expected good window to be dropped (80%% subsumed by prime), got %+v", w)
		}
	}
	primeCount := 0
	for _, w := range wins {
		if w.Condition == "prime" {
			primeCount++
		}
	}
	if primeCount != 1 {
		t.Errorf("expected 1 prime window, got %d", primeCount)
	}
}

// TestDetectSendWindows_NonSubsumedGoodKept verifies a "good" window
// with only a small overlap (<80%) with a "prime" window keeps both.
func TestDetectSendWindows_NonSubsumedGoodKept(t *testing.T) {
	base := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	// 2h prime + 8h good → good run = 10h, prime run = 2h → 20% overlap → keep both.
	conds := []string{"prime", "prime", "good", "good", "good", "good", "good", "good", "good", "good"}
	surf := []float64{55, 55, 65, 65, 65, 65, 65, 65, 65, 65}
	cond := make([]bool, 10)
	hours := makeHours(base, conds, surf, cond)
	wins := DetectSendWindows(hours, "UTC", SendWindowOptions{})
	hasPrime, hasGood := false, false
	for _, w := range wins {
		if w.Condition == "prime" {
			hasPrime = true
		}
		if w.Condition == "good" {
			hasGood = true
		}
	}
	if !hasPrime {
		t.Errorf("expected prime window kept")
	}
	if !hasGood {
		t.Errorf("expected good window kept (only 20%% overlap)")
	}
}

// TestDetectSendWindows_DeterministicSort verifies that output is
// sorted by StartTime, with prime tier before good when start times tie.
func TestDetectSendWindows_DeterministicSort(t *testing.T) {
	base := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	// Two separate runs: an early good run (00..03) and a later prime
	// run (10..13) inside a wrapping good span (08..15). Expect output
	// ordered by start time.
	conds := []string{
		"good", "good", "good", "good", // 00..03 good run #1
		"poor", "poor", "poor", "poor", // 04..07 poor
		"good", "good", // 08..09 good (start of run #2)
		"prime", "prime", "prime", "prime", // 10..13 prime
		"good", "good", // 14..15 good (tail of run #2)
	}
	surf := make([]float64, len(conds))
	for i, c := range conds {
		switch c {
		case "prime":
			surf[i] = 55
		case "good":
			surf[i] = 65
		default:
			surf[i] = 90
		}
	}
	cond := make([]bool, len(conds))
	hours := makeHours(base, conds, surf, cond)
	wins := DetectSendWindows(hours, "UTC", SendWindowOptions{})

	// Must be non-decreasing by StartTime.
	for i := 1; i < len(wins); i++ {
		if wins[i].StartTime.Before(wins[i-1].StartTime) {
			t.Errorf("windows not sorted by StartTime: %+v", wins)
		}
	}
	// First window should start at 00:00 (good run #1).
	if !wins[0].StartTime.Equal(base) {
		t.Errorf("first window StartTime: got %v want %v", wins[0].StartTime, base)
	}
}

// TestDetectSendWindows_TimezoneSplit verifies that midnight splitting
// uses the supplied tzName, not UTC. A run from 2025-06-01 06:00 UTC
// (= 2025-05-31 23:00 PT) to 09:00 UTC (= 02:00 PT) crosses local
// midnight in PT and should split into 2 windows when given LA tz, but
// remain 1 window in UTC.
func TestDetectSendWindows_TimezoneSplit(t *testing.T) {
	base := time.Date(2025, 6, 1, 6, 0, 0, 0, time.UTC) // 23:00 PT prev day
	conds := make([]string, 4)                          // 06,07,08,09 UTC = 23,00,01,02 PT
	surf := make([]float64, 4)
	cond := make([]bool, 4)
	for i := range conds {
		conds[i] = "prime"
		surf[i] = 55
	}
	hours := makeHours(base, conds, surf, cond)

	utcWins := DetectSendWindows(hours, "UTC", SendWindowOptions{})
	primeUTC := 0
	for _, w := range utcWins {
		if w.Condition == "prime" {
			primeUTC++
		}
	}
	if primeUTC != 1 {
		t.Errorf("UTC: expected 1 prime window (no midnight cross), got %d: %+v", primeUTC, utcWins)
	}

	laWins := DetectSendWindows(hours, "America/Los_Angeles", SendWindowOptions{})
	primeLA := 0
	for _, w := range laWins {
		if w.Condition == "prime" {
			primeLA++
		}
	}
	// 23:00 PT alone = 1h (below 1.5h min) → dropped.
	// 00:00..02:00 PT = 3h → kept.
	if primeLA != 1 {
		t.Errorf("LA: expected 1 prime window after split (the 3h sub-window; 1h half dropped), got %d: %+v", primeLA, laWins)
	}
	if primeLA == 1 && laWins[0].DurationH != 3 {
		t.Errorf("LA: expected surviving sub-window to be 3h, got %.1f", laWins[0].DurationH)
	}
}

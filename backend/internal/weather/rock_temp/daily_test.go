package rock_temp

import (
	"testing"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

func TestAggregateDaily_Empty(t *testing.T) {
	out := AggregateDaily(nil, nil, "UTC")
	if len(out) != 0 {
		t.Fatalf("expected empty output, got %d entries", len(out))
	}
	out = AggregateDaily([]models.RockTempHour{}, []models.SendWindow{}, "")
	if len(out) != 0 {
		t.Fatalf("expected empty output, got %d entries", len(out))
	}
}

func TestAggregateDaily_TwoDays(t *testing.T) {
	// Day 1: prime peak with a prime window. Day 2: all poor.
	day1 := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	day2 := time.Date(2025, 6, 2, 0, 0, 0, 0, time.UTC)

	hourly := []models.RockTempHour{
		{Time: day1.Add(6 * time.Hour), SurfaceF: 50, Condition: "good"},
		{Time: day1.Add(10 * time.Hour), SurfaceF: 65, Condition: "prime"},
		{Time: day1.Add(14 * time.Hour), SurfaceF: 60, Condition: "good"},
		{Time: day2.Add(10 * time.Hour), SurfaceF: 95, Condition: "poor"},
		{Time: day2.Add(14 * time.Hour), SurfaceF: 100, Condition: "poor"},
	}
	windows := []models.SendWindow{
		{
			StartTime: day1.Add(8 * time.Hour),
			EndTime:   day1.Add(12 * time.Hour),
			DurationH: 4,
			Condition: "prime",
		},
	}
	out := AggregateDaily(hourly, windows, "UTC")
	if len(out) != 2 {
		t.Fatalf("expected 2 days, got %d", len(out))
	}
	if out[0].LocalDate != "2025-06-01" {
		t.Errorf("expected first day 2025-06-01, got %s", out[0].LocalDate)
	}
	if out[0].PeakSurfaceTempF != 65 {
		t.Errorf("expected peak 65, got %v", out[0].PeakSurfaceTempF)
	}
	if out[0].MinSurfaceTempF != 50 {
		t.Errorf("expected min 50, got %v", out[0].MinSurfaceTempF)
	}
	if out[0].PeakCondition != "prime" {
		t.Errorf("expected peak cond prime, got %s", out[0].PeakCondition)
	}
	if out[0].OverallCondition != "good" {
		t.Errorf("expected overall good (worst non-prime), got %s", out[0].OverallCondition)
	}
	if out[0].BestSendWindow == nil || out[0].BestSendWindow.Condition != "prime" {
		t.Errorf("expected prime window on day 1")
	}
	if out[0].WindowCount != 1 {
		t.Errorf("expected 1 window on day 1, got %d", out[0].WindowCount)
	}
	if out[1].OverallCondition != "poor" {
		t.Errorf("expected day2 overall poor, got %s", out[1].OverallCondition)
	}
	if out[1].PeakSurfaceTempF != 100 {
		t.Errorf("expected day2 peak 100, got %v", out[1].PeakSurfaceTempF)
	}
	if out[1].BestSendWindow != nil {
		t.Errorf("expected no window on day 2")
	}
	if out[1].WindowCount != 0 {
		t.Errorf("expected 0 windows on day 2, got %d", out[1].WindowCount)
	}
}

func TestAggregateDaily_Condensation(t *testing.T) {
	day1 := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	day2 := time.Date(2025, 6, 2, 0, 0, 0, 0, time.UTC)
	hourly := []models.RockTempHour{
		{Time: day1.Add(6 * time.Hour), SurfaceF: 50, Condition: "good", Condensing: true},
		{Time: day1.Add(12 * time.Hour), SurfaceF: 60, Condition: "prime"},
		{Time: day2.Add(12 * time.Hour), SurfaceF: 62, Condition: "prime"},
	}
	out := AggregateDaily(hourly, nil, "UTC")
	if len(out) != 2 {
		t.Fatalf("expected 2 days, got %d", len(out))
	}
	if !out[0].HasCondensation {
		t.Errorf("expected day1 has condensation")
	}
	if out[1].HasCondensation {
		t.Errorf("expected day2 no condensation")
	}
}

func TestAggregateDaily_WindowSpanningMidnight(t *testing.T) {
	day1 := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	day2 := day1.AddDate(0, 0, 1)
	hourly := []models.RockTempHour{
		{Time: day1.Add(22 * time.Hour), SurfaceF: 60, Condition: "prime"},
		{Time: day2.Add(2 * time.Hour), SurfaceF: 58, Condition: "prime"},
	}
	windows := []models.SendWindow{
		{
			StartTime: day1.Add(22 * time.Hour),
			EndTime:   day2.Add(3 * time.Hour),
			DurationH: 5,
			Condition: "prime",
		},
		{
			// Shorter prime window only on day 2, to verify "best" picks longer
			StartTime: day2.Add(10 * time.Hour),
			EndTime:   day2.Add(11 * time.Hour),
			DurationH: 1,
			Condition: "prime",
		},
	}
	out := AggregateDaily(hourly, windows, "UTC")
	if len(out) != 2 {
		t.Fatalf("expected 2 days, got %d", len(out))
	}
	if out[0].WindowCount != 1 {
		t.Errorf("expected day1 window count 1, got %d", out[0].WindowCount)
	}
	if out[1].WindowCount != 2 {
		t.Errorf("expected day2 window count 2 (spanning + standalone), got %d", out[1].WindowCount)
	}
	if out[1].BestSendWindow == nil || out[1].BestSendWindow.DurationH != 5 {
		t.Errorf("expected day2 best window to be 5h spanning window")
	}
}

func TestAggregateDaily_TimezoneHandling(t *testing.T) {
	// Hour at 2025-06-01 06:00 UTC == 2025-05-31 23:00 PT.
	hourly := []models.RockTempHour{
		{Time: time.Date(2025, 6, 1, 6, 0, 0, 0, time.UTC), SurfaceF: 60, Condition: "good"},
	}
	utcOut := AggregateDaily(hourly, nil, "UTC")
	if len(utcOut) != 1 || utcOut[0].LocalDate != "2025-06-01" {
		t.Errorf("UTC: expected 2025-06-01, got %+v", utcOut)
	}
	laOut := AggregateDaily(hourly, nil, "America/Los_Angeles")
	if len(laOut) != 1 || laOut[0].LocalDate != "2025-05-31" {
		t.Errorf("LA: expected 2025-05-31, got %+v", laOut)
	}
}

func TestWorseCondition(t *testing.T) {
	cases := []struct {
		a, b, want string
	}{
		{"prime", "good", "good"},
		{"good", "marginal", "marginal"},
		{"marginal", "too_cold", "too_cold"},
		{"too_cold", "poor", "poor"},
		{"poor", "very_poor", "very_poor"},
		{"prime", "prime", "prime"},
	}
	for _, c := range cases {
		got := worseCondition(c.a, c.b)
		if got != c.want {
			t.Errorf("worseCondition(%q,%q) = %q, want %q", c.a, c.b, got, c.want)
		}
	}
}

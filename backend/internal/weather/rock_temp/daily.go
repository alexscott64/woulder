package rock_temp

import (
	"sort"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// AggregateDaily groups an hourly rock-temp forecast into per-day summaries
// using the supplied IANA timezone (e.g., "America/Los_Angeles") so day
// boundaries match the user's local clock. If tzName is empty or invalid,
// falls back to UTC.
//
// sendWindows is the full set of windows from DetectSendWindows; the function
// picks the best window per day (prime preferred; ties broken by longer duration).
//
// Output is sorted by LocalDate ascending.
func AggregateDaily(
	hourly []models.RockTempHour,
	sendWindows []models.SendWindow,
	tzName string,
) []models.DailyRockTemp {
	if len(hourly) == 0 {
		return nil
	}

	loc, err := time.LoadLocation(tzName)
	if err != nil || tzName == "" {
		loc = time.UTC
	}

	// Group hours by local date.
	byDay := make(map[string][]models.RockTempHour)
	for _, h := range hourly {
		key := h.Time.In(loc).Format("2006-01-02")
		byDay[key] = append(byDay[key], h)
	}

	// Build daily summaries.
	out := make([]models.DailyRockTemp, 0, len(byDay))
	for date, hours := range byDay {
		if len(hours) == 0 {
			continue
		}
		peak := hours[0].SurfaceF
		peakCond := hours[0].Condition
		minT := hours[0].SurfaceF
		overall := hours[0].Condition
		hasCond := false
		for _, h := range hours {
			if h.SurfaceF > peak {
				peak = h.SurfaceF
				peakCond = h.Condition
			}
			if h.SurfaceF < minT {
				minT = h.SurfaceF
			}
			overall = worseCondition(overall, h.Condition)
			if h.Condensing {
				hasCond = true
			}
		}

		// Find overlapping windows for this day and pick best.
		var best *models.SendWindow
		windowCount := 0
		for i := range sendWindows {
			w := sendWindows[i]
			if windowOverlapsLocalDate(w, date, loc) {
				windowCount++
				if best == nil || isBetterWindow(w, *best) {
					ww := w
					best = &ww
				}
			}
		}

		out = append(out, models.DailyRockTemp{
			LocalDate:        date,
			PeakSurfaceTempF: peak,
			MinSurfaceTempF:  minT,
			PeakCondition:    peakCond,
			OverallCondition: overall,
			HasCondensation:  hasCond,
			BestSendWindow:   best,
			WindowCount:      windowCount,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].LocalDate < out[j].LocalDate
	})
	return out
}

// windowOverlapsLocalDate returns true if any portion of the window falls on
// the given local date in the supplied location.
func windowOverlapsLocalDate(w models.SendWindow, date string, loc *time.Location) bool {
	startDate := w.StartTime.In(loc).Format("2006-01-02")
	endDate := w.EndTime.In(loc).Format("2006-01-02")
	return startDate == date || endDate == date || (startDate < date && date < endDate)
}

// isBetterWindow returns true if `a` is better than `b`. Prime > good; ties
// broken by longer duration.
func isBetterWindow(a, b models.SendWindow) bool {
	if a.Condition == "prime" && b.Condition != "prime" {
		return true
	}
	if a.Condition != "prime" && b.Condition == "prime" {
		return false
	}
	return a.DurationH > b.DurationH
}

// worseCondition returns the more concerning of two condition tier strings.
// Ranking (worst -> best): very_poor > poor > too_cold > marginal > good > prime.
// too_cold ranks as more concerning than marginal because climbing in too_cold
// conditions is actually difficult (skin won't grip well, holds slick).
func worseCondition(a, b string) string {
	rank := func(c string) int {
		switch c {
		case "very_poor":
			return 5
		case "poor":
			return 4
		case "too_cold":
			return 3
		case "marginal":
			return 2
		case "good":
			return 1
		case "prime":
			return 0
		}
		return -1
	}
	if rank(b) > rank(a) {
		return b
	}
	return a
}

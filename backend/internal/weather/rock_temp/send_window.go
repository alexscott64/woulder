package rock_temp

import (
	"sort"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// DefaultMinWindowH is the minimum send-window duration in hours when
// SendWindowOptions.MinDurationH is unset.
const DefaultMinWindowH = 1.5

// SendWindowOptions configures DetectSendWindows.
type SendWindowOptions struct {
	// MinDurationH is the minimum window duration to emit. Values <= 0
	// fall back to DefaultMinWindowH.
	MinDurationH float64
}

// DetectSendWindows scans an hourly rock-temp forecast and emits
// contiguous send windows. It performs two passes:
//
//   - Prime windows: hours where Condition == "prime".
//   - Good-or-better windows: hours where Condition is "prime" or
//     "good".
//
// Algorithm:
//
//  1. For each pass, find every maximal contiguous run of qualifying
//     hours (one window per run — never multiple windows for sub-spans
//     of the same run).
//  2. Split each run at local-midnight boundaries (so a multi-day
//     "always good" stretch becomes one window per local calendar day).
//     This avoids meaningless 70h+ windows that aren't useful for
//     trip planning.
//  3. Drop split sub-windows shorter than minDurationH.
//  4. Drop any "good" window whose duration overlaps a "prime" window
//     by ≥80% — climbers don't need to see both when the prime span
//     covers most of the good span.
//  5. Sort deterministically by StartTime, then prime-before-good.
//
// tzName is an IANA timezone name (e.g., "America/Los_Angeles") used
// to determine local midnight. Empty/invalid values fall back to UTC.
//
// Hours flagged Condensing=true are *included* in the window but cause
// DryThroughout=false on that window. The caller is responsible for
// excluding heavy-condensation hours upstream (e.g., by setting
// Condensing=false on light-only hours and dropping heavy hours from
// the slice).
//
// DurationH assumes hours are evenly spaced 1h apart, so each entry in
// the window represents 1.0 hours of conditions. EndTime is the
// timestamp of the *last* hour in the run (start of that hour); the
// window's covered interval is [StartTime, EndTime + 1h).
func DetectSendWindows(hours []models.RockTempHour, tzName string, opts SendWindowOptions) []models.SendWindow {
	minH := opts.MinDurationH
	if minH <= 0 {
		minH = DefaultMinWindowH
	}

	loc, err := time.LoadLocation(tzName)
	if err != nil || tzName == "" {
		loc = time.UTC
	}

	primeMatch := func(h models.RockTempHour) bool { return h.Condition == "prime" }
	goodMatch := func(h models.RockTempHour) bool {
		return h.Condition == "prime" || h.Condition == "good"
	}

	primeRuns := findContiguousRuns(hours, primeMatch)
	goodRuns := findContiguousRuns(hours, goodMatch)

	primeWindows := buildWindowsFromRuns(hours, primeRuns, "prime", loc, minH)
	goodWindows := buildWindowsFromRuns(hours, goodRuns, "good", loc, minH)

	goodWindows = removeSubsumed(goodWindows, primeWindows, 0.80)

	out := make([]models.SendWindow, 0, len(primeWindows)+len(goodWindows))
	out = append(out, primeWindows...)
	out = append(out, goodWindows...)
	sortWindows(out)
	return out
}

// hourRun is a half-open index range [start, end) into the hours slice
// representing a maximal contiguous run of qualifying hours.
type hourRun struct {
	start int
	end   int // exclusive
}

// findContiguousRuns returns one hourRun per maximal contiguous run of
// hours satisfying match. Each run is emitted exactly once.
func findContiguousRuns(hours []models.RockTempHour, match func(models.RockTempHour) bool) []hourRun {
	var runs []hourRun
	i := 0
	for i < len(hours) {
		if !match(hours[i]) {
			i++
			continue
		}
		start := i
		for i < len(hours) && match(hours[i]) {
			i++
		}
		runs = append(runs, hourRun{start: start, end: i})
	}
	return runs
}

// buildWindowsFromRuns converts each run into one or more SendWindows,
// splitting at local-midnight boundaries and dropping sub-windows
// shorter than minDurationH.
func buildWindowsFromRuns(
	hours []models.RockTempHour,
	runs []hourRun,
	label string,
	loc *time.Location,
	minDurationH float64,
) []models.SendWindow {
	var out []models.SendWindow
	for _, r := range runs {
		segments := splitRunAtLocalMidnight(hours, r, loc)
		for _, seg := range segments {
			w, ok := makeWindow(hours, seg, label, minDurationH)
			if ok {
				out = append(out, w)
			}
		}
	}
	return out
}

// splitRunAtLocalMidnight breaks a single hourRun into one or more
// half-open index sub-ranges so that no sub-range crosses a local-
// midnight boundary in loc. Each emitted segment covers a contiguous
// slice of indices into hours.
func splitRunAtLocalMidnight(hours []models.RockTempHour, r hourRun, loc *time.Location) []hourRun {
	if r.end-r.start <= 0 {
		return nil
	}
	var segs []hourRun
	segStart := r.start
	for i := r.start + 1; i < r.end; i++ {
		prevDay := hours[i-1].Time.In(loc).Format("2006-01-02")
		curDay := hours[i].Time.In(loc).Format("2006-01-02")
		if prevDay != curDay {
			segs = append(segs, hourRun{start: segStart, end: i})
			segStart = i
		}
	}
	segs = append(segs, hourRun{start: segStart, end: r.end})
	return segs
}

// makeWindow turns an index range into a SendWindow, computing
// avg/peak/dry stats. Returns ok=false when the duration is below
// minDurationH or the range is empty.
func makeWindow(
	hours []models.RockTempHour,
	r hourRun,
	label string,
	minDurationH float64,
) (models.SendWindow, bool) {
	count := r.end - r.start
	if count <= 0 {
		return models.SendWindow{}, false
	}
	duration := float64(count) // 1h per slot
	if duration < minDurationH {
		return models.SendWindow{}, false
	}

	var sum float64
	peak := hours[r.start].SurfaceF
	dry := true
	for j := r.start; j < r.end; j++ {
		sum += hours[j].SurfaceF
		if hours[j].SurfaceF > peak {
			peak = hours[j].SurfaceF
		}
		if hours[j].Condensing {
			dry = false
		}
	}
	avg := sum / float64(count)

	return models.SendWindow{
		StartTime:     hours[r.start].Time,
		EndTime:       hours[r.end-1].Time,
		DurationH:     duration,
		Condition:     label,
		AvgTempF:      avg,
		PeakTempF:     peak,
		DryThroughout: dry,
	}, true
}

// removeSubsumed drops any window from goods whose covered interval
// overlaps a prime window by at least overlapFrac (e.g., 0.80 = 80%).
// The covered interval of a window is [StartTime, EndTime + 1h).
func removeSubsumed(goods, primes []models.SendWindow, overlapFrac float64) []models.SendWindow {
	if len(primes) == 0 || len(goods) == 0 {
		return goods
	}
	out := make([]models.SendWindow, 0, len(goods))
	for _, g := range goods {
		gStart, gEnd := windowInterval(g)
		gDur := gEnd.Sub(gStart).Hours()
		if gDur <= 0 {
			out = append(out, g)
			continue
		}
		var maxOverlap float64
		for _, p := range primes {
			pStart, pEnd := windowInterval(p)
			ov := overlapHours(gStart, gEnd, pStart, pEnd)
			if ov > maxOverlap {
				maxOverlap = ov
			}
		}
		if maxOverlap/gDur >= overlapFrac {
			continue // subsumed by a prime window
		}
		out = append(out, g)
	}
	return out
}

// windowInterval returns the covered half-open interval
// [start, end+1h) of a SendWindow.
func windowInterval(w models.SendWindow) (time.Time, time.Time) {
	return w.StartTime, w.EndTime.Add(time.Hour)
}

// overlapHours returns the overlap duration in hours between two
// half-open intervals [aStart, aEnd) and [bStart, bEnd).
func overlapHours(aStart, aEnd, bStart, bEnd time.Time) float64 {
	start := aStart
	if bStart.After(start) {
		start = bStart
	}
	end := aEnd
	if bEnd.Before(end) {
		end = bEnd
	}
	if !end.After(start) {
		return 0
	}
	return end.Sub(start).Hours()
}

// sortWindows sorts in place by StartTime ascending; ties broken by
// tier (prime before good) so output is deterministic.
func sortWindows(ws []models.SendWindow) {
	sort.SliceStable(ws, func(i, j int) bool {
		if !ws[i].StartTime.Equal(ws[j].StartTime) {
			return ws[i].StartTime.Before(ws[j].StartTime)
		}
		return tierRank(ws[i].Condition) < tierRank(ws[j].Condition)
	})
}

func tierRank(c string) int {
	switch c {
	case "prime":
		return 0
	case "good":
		return 1
	}
	return 2
}

// NextTransition returns the first hour in the supplied slice whose
// Condition differs from currentCondition, or nil when no transition
// is found within the slice. The slice is scanned from index 0 forward.
func NextTransition(hours []models.RockTempHour, currentCondition string) *models.Transition {
	for _, h := range hours {
		if h.Condition != currentCondition {
			return &models.Transition{Time: h.Time, ToCondition: h.Condition}
		}
	}
	return nil
}

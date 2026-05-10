package rock_temp

import (
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
// Each pass groups consecutive matching hours, computes summary stats
// (avg / peak surface temp, duration), and filters by
// opts.MinDurationH. Hours flagged Condensing=true are *included* in
// the window but cause DryThroughout=false on that window. The
// caller is responsible for excluding heavy-condensation hours
// upstream (e.g., by setting Condensing=false on light-only hours and
// dropping heavy hours from the slice).
//
// Duplicate detection: when a prime-only window has identical start
// and end times as a good-or-better window, only the prime variant is
// emitted. Otherwise both are returned and the UI can de-dupe further.
//
// DurationH assumes hours are evenly spaced 1h apart, so each entry in
// the window represents 1.0 hours of conditions.
func DetectSendWindows(hours []models.RockTempHour, opts SendWindowOptions) []models.SendWindow {
	minH := opts.MinDurationH
	if minH <= 0 {
		minH = DefaultMinWindowH
	}

	primeMatch := func(h models.RockTempHour) bool { return h.Condition == "prime" }
	goodMatch := func(h models.RockTempHour) bool {
		return h.Condition == "prime" || h.Condition == "good"
	}

	primeWindows := scanWindows(hours, primeMatch, "prime", minH)
	goodWindows := scanWindows(hours, goodMatch, "good", minH)

	// Drop good-or-better windows that have an identical prime window
	// with the same start AND end (prime is the more specific tier and
	// should "win" the dedupe — caller can match by tier otherwise).
	out := make([]models.SendWindow, 0, len(primeWindows)+len(goodWindows))
	out = append(out, primeWindows...)
	for _, gw := range goodWindows {
		dup := false
		for _, pw := range primeWindows {
			if pw.StartTime.Equal(gw.StartTime) && pw.EndTime.Equal(gw.EndTime) {
				dup = true
				break
			}
		}
		if !dup {
			out = append(out, gw)
		}
	}
	return out
}

// scanWindows finds contiguous runs of hours satisfying match and
// converts each run >= minDurationH hours long into a SendWindow with
// Condition=label.
func scanWindows(
	hours []models.RockTempHour,
	match func(models.RockTempHour) bool,
	label string,
	minDurationH float64,
) []models.SendWindow {
	var out []models.SendWindow
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
		end := i // exclusive
		count := end - start
		if count <= 0 {
			continue
		}
		duration := float64(count) // 1h per slot
		if duration < minDurationH {
			continue
		}

		var sum, peak float64
		dry := true
		peak = hours[start].SurfaceF
		for j := start; j < end; j++ {
			sum += hours[j].SurfaceF
			if hours[j].SurfaceF > peak {
				peak = hours[j].SurfaceF
			}
			if hours[j].Condensing {
				dry = false
			}
		}
		avg := sum / float64(count)

		out = append(out, models.SendWindow{
			StartTime:     hours[start].Time,
			EndTime:       hours[end-1].Time,
			DurationH:     duration,
			Condition:     label,
			AvgTempF:      avg,
			PeakTempF:     peak,
			DryThroughout: dry,
		})
	}
	return out
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

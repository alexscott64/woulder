package rock_temp

// ConfidenceMode controls how horizon-related deductions are applied.
//
// ConfidenceModeStatus is for the single "now"/current-hour score that
// gets attached to RockTemperatureStatus.Confidence. The long-horizon
// forecast penalty is suppressed because future-week uncertainty does
// not reflect on the present-hour estimate.
//
// ConfidenceModeHour is for per-hour or per-day scoring downstream of
// "now". The horizon penalty is applied tiered by distance from now.
type ConfidenceMode int

const (
	ConfidenceModeStatus ConfidenceMode = iota
	ConfidenceModeHour
)

// MinConfidence is the floor applied to every score. Raised from the
// historical 20 to 30 so that "I have low data" still conveys that the
// calculator did meaningful work.
const MinConfidence = 30

// ConfidenceInputs collects the observable factors that drive the
// confidence score for a rock temperature prediction.
//
// Aspect/Dip semantics: a value of false means the calculator could not
// pick a single dominant face/dip from the location_sun_exposure row
// (no *_facing_percent or slab/overhang exceeds the 60% threshold) and
// fell back to a weighted average or default. It does NOT mean the row
// is missing — that case is covered by NoSunExposureRow below, which
// takes precedence and emits a stronger reason.
type ConfidenceInputs struct {
	AspectKnown       bool // false → no facet ≥60% dominant; aspect derived/defaulted (−15)
	DipKnown          bool // false → slab/overhang not dominant; dip derived/defaulted (−10)
	RockTypeKnown     bool // false → no row in location_rock_types (−5)
	MixedFacets       bool // true  → location has heterogeneous facets (−8)
	ForecastHorizonH  int  // hours ahead of "now"; only applied in ConfidenceModeHour
	CloudVariableHigh bool // cloud cover stddev > threshold (−5)
	WindVariableHigh  bool // wind forecast highly variable (−10)
	SpinUpComplete    bool // false → past_hours data missing (−5)
	// NoSunExposureRow indicates there is no location_sun_exposure row
	// for this location at all (a much worse signal than "row exists
	// but no facet dominates"). When true, a single −25 deduction and a
	// distinct "no row" reason are emitted, and the per-field
	// AspectKnown/DipKnown reasons are suppressed to avoid noise.
	NoSunExposureRow bool
	Mode             ConfidenceMode
}

// ConfidenceResult is the score (0..100) plus a list of human-readable
// factor strings explaining each deduction.
type ConfidenceResult struct {
	Score   int
	Factors []string
}

// ComputeConfidence applies the deductions to a baseline of 100, floors
// the result at MinConfidence, and assembles a list of factor strings
// explaining each deduction in the same order they were applied.
//
// Reason strings below are intentionally written in plain English for
// end-user climbers (no column names, no SQL, no acronyms). Each one is
// preceded by an "OPERATOR HINT" code comment describing what a
// developer/operator can do to raise the score on that factor; those
// hints are for future maintainers and never reach the API.
func ComputeConfidence(in ConfidenceInputs) ConfidenceResult {
	score := 100
	factors := []string{}

	// Missing-row case is reported as a single, stronger signal and
	// suppresses the per-field aspect/dip reasons since they'd just
	// repeat the same root cause.
	if in.NoSunExposureRow {
		score -= 25
		// OPERATOR HINT: Seed a row in location_sun_exposure for this location.
		// Template: backend/internal/database/migrations/seed_location_sun_exposure_TEMPLATE.sql
		factors = append(factors, "Sun exposure data not yet set for this location")
	} else {
		if !in.AspectKnown {
			score -= 15
			// OPERATOR HINT: To raise the score on this factor, refine the
			// *_facing_percent values in location_sun_exposure so one direction
			// (N/E/S/W) reaches ≥60% dominance.
			// See backend/internal/database/migrations/seed_location_sun_exposure_TEMPLATE.sql
			factors = append(factors, "Crag has mixed sun exposure — using an average across directions")
		}
		if !in.DipKnown {
			score -= 10
			// OPERATOR HINT: To raise the score on this factor, refine slab_percent
			// and overhang_percent in location_sun_exposure so a non-vertical
			// face dominates (>60%).
			factors = append(factors, "Assuming vertical walls — slab and overhang amounts not specified")
		}
	}
	if !in.RockTypeKnown {
		score -= 5
		// OPERATOR HINT: Add a row for this location in woulder.location_rock_types
		// so the calculator can use the correct thermal properties instead of the
		// granite default.
		factors = append(factors, "Using default rock properties (granite) for this location")
	}
	// Only emit the mixed-facets reason when the per-field aspect/dip
	// reasons aren't already covering the same root cause. Two cases:
	//   - Row exists, both AspectKnown && DipKnown false → both
	//     per-field reasons fire; an extra "mixed facets" line is noise.
	//   - Row missing (NoSunExposureRow) → per-field reasons are
	//     suppressed, so mixed-facets, if flagged, is non-duplicative.
	if in.MixedFacets && (in.NoSunExposureRow || in.AspectKnown || in.DipKnown) {
		score -= 8
		// OPERATOR HINT: Inherent uncertainty — no schema action will remove this.
		// Varied terrain genuinely has differing conditions across faces.
		factors = append(factors, "Varied terrain — some areas sunnier than others")
	}
	if in.Mode == ConfidenceModeHour && in.ForecastHorizonH > 0 {
		deduction, reason := horizonPenalty(in.ForecastHorizonH)
		if deduction > 0 {
			score -= deduction
			// OPERATOR HINT: Inherent uncertainty — no schema action will remove this.
			factors = append(factors, reason)
		}
	}
	if in.CloudVariableHigh {
		score -= 5
		// OPERATOR HINT: Inherent uncertainty — no schema action will remove this.
		factors = append(factors, "Variable cloud cover — sunlight changes hour to hour")
	}
	if in.WindVariableHigh {
		score -= 10
		// OPERATOR HINT: Inherent uncertainty — no schema action will remove this.
		factors = append(factors, "Variable wind forecast — conditions may shift")
	}
	if !in.SpinUpComplete {
		score -= 5
		// OPERATOR HINT: Ensure the weather sync is back-filling past_hours
		// data so the thermal-lag model has enough recent history to warm up.
		factors = append(factors, "Limited recent weather history — accuracy improves over time")
	}

	if score < MinConfidence {
		score = MinConfidence
	}
	if score > 100 {
		score = 100
	}
	return ConfidenceResult{Score: score, Factors: factors}
}

// horizonPenalty returns the deduction and reason for a given forecast
// horizon (hours ahead of now). Tiers per spec:
//
//	0-24h    → 0     (no penalty)
//	24-72h   → 5     (small)
//	72-120h  → 10    (moderate)
//	120-168h → 15    (larger)
//	>168h    → 15    (capped)
func horizonPenalty(hours int) (int, string) {
	// Plain-English reason; {N} is the hours-ahead value. Operator note:
	// inherent forecast uncertainty — no schema action will remove this.
	reason := "Forecast " + itoa(hours) + "h out — accuracy decreases further from now"
	switch {
	case hours <= 24:
		return 0, ""
	case hours <= 72:
		return 5, reason
	case hours <= 120:
		return 10, reason
	default:
		return 15, reason
	}
}

// itoa is a tiny dependency-free integer-to-string helper to avoid
// pulling in strconv just for confidence factor strings.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

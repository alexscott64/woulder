package rock_temp

// ConfidenceInputs collects the observable factors that drive the
// confidence score for a rock temperature prediction. Each field
// corresponds to a deduction described in the plan section 5.
type ConfidenceInputs struct {
	AspectKnown       bool // false → aspect was defaulted (−25)
	DipKnown          bool // false → dip was defaulted to vertical (−15)
	RockTypeKnown     bool // false → rock type was defaulted (−5)
	MixedFacets       bool // true  → location is mixed (−8)
	ForecastHorizonH  int  // hours ahead of "now"; subtract floor(h/24) * 5
	CloudVariableHigh bool // cloud cover stddev > 30% across the day (−10)
	WindVariableHigh  bool // wind forecast highly variable (−10)
	SpinUpComplete    bool // false → past_hours data missing (−5)
}

// ConfidenceResult is the score (0..100) plus a list of human-readable
// factor strings explaining each deduction.
type ConfidenceResult struct {
	Score   int
	Factors []string
}

// ComputeConfidence applies the deductions described in plan section 5
// to a baseline of 100, floors the result at 20, and assembles a list
// of factor strings explaining each deduction in the same order they
// were applied.
func ComputeConfidence(in ConfidenceInputs) ConfidenceResult {
	score := 100
	factors := []string{}

	if !in.AspectKnown {
		score -= 25
		factors = append(factors, "face aspect defaulted")
	}
	if !in.DipKnown {
		score -= 15
		factors = append(factors, "face dip defaulted to vertical")
	}
	if !in.RockTypeKnown {
		score -= 5
		factors = append(factors, "rock type defaulted to granite")
	}
	if in.MixedFacets {
		score -= 8
		factors = append(factors, "mixed facets; result averaged across multiple aspects")
	}
	if in.ForecastHorizonH > 0 {
		days := in.ForecastHorizonH / 24
		if days > 0 {
			deduction := days * 5
			score -= deduction
			factors = append(factors, formatHorizonFactor(in.ForecastHorizonH))
		}
	}
	if in.CloudVariableHigh {
		score -= 10
		factors = append(factors, "irradiance variable due to broken clouds")
	}
	if in.WindVariableHigh {
		score -= 10
		factors = append(factors, "wind forecast variable")
	}
	if !in.SpinUpComplete {
		score -= 5
		factors = append(factors, "thermal lag spin-up incomplete")
	}

	if score < 20 {
		score = 20
	}
	if score > 100 {
		score = 100
	}
	return ConfidenceResult{Score: score, Factors: factors}
}

func formatHorizonFactor(hours int) string {
	// Concise integer hours format, e.g., "forecast 48h out".
	return "forecast " + itoa(hours) + "h out"
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

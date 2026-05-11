package rock_temp

// ComputeFrictionQuality combines a temperature-tier condition with a
// condensation severity into a single user-facing friction rating.
//
// Inputs:
//   - tempCondition: one of "prime", "good", "marginal", "poor",
//     "very_poor", "too_cold" (output of ClassifyTempCondition).
//   - condensationSeverity: one of "none", "light", "heavy" (output of
//     ClassifyCondensation).
//
// Output: one of "excellent", "good", "reduced", "poor", or "unknown" if
// either input is not recognized.
//
// Rules (per the plan):
//   - heavy condensation → "poor" regardless of temp.
//   - light condensation → degrade by one tier:
//     prime/good → "reduced"; marginal/poor/very_poor/too_cold → "poor".
//   - dry: prime → "excellent", good → "good", marginal → "reduced",
//     poor/very_poor/too_cold → "poor".
func ComputeFrictionQuality(tempCondition, condensationSeverity string) string {
	switch condensationSeverity {
	case "heavy":
		return "poor"
	case "light":
		switch tempCondition {
		case "prime", "good":
			return "reduced"
		case "marginal", "poor", "very_poor", "too_cold":
			return "poor"
		default:
			return "unknown"
		}
	case "none":
		switch tempCondition {
		case "prime":
			return "excellent"
		case "good":
			return "good"
		case "marginal":
			return "reduced"
		case "poor", "very_poor", "too_cold":
			return "poor"
		default:
			return "unknown"
		}
	}
	return "unknown"
}

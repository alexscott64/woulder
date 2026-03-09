// Package grades provides grade parsing, ordering, and classification utilities
// for climbing route grades across different discipline grading systems.
//
// Supported grade families:
//   - V-scale (bouldering): V0 through V17
//   - YDS (sport/trad): 5.4 through 5.15d
//   - Ice (WI): WI1 through WI7
//   - Alpine Ice (AI): AI1 through AI6
//   - Mixed (M): M1 through M13
package grades

import (
	"strings"
)

// Grade families returned by Family().
const (
	FamilyV     = "v"
	FamilyYDS   = "yds"
	FamilyWI    = "wi"
	FamilyAI    = "ai"
	FamilyMixed = "mixed"
	FamilyAid   = "aid"
)

// Ordered V-scale grades (index = sort order).
var vScaleGrades = []string{
	"V0", "V1", "V2", "V3", "V4", "V5", "V6", "V7", "V8", "V9",
	"V10", "V11", "V12", "V13", "V14", "V15", "V16", "V17",
}

// Ordered YDS grades (index = sort order, offset by 100 to avoid collision).
var ydsGrades = []string{
	"5.4", "5.5", "5.6", "5.7", "5.8", "5.9",
	"5.10a", "5.10b", "5.10c", "5.10d",
	"5.11a", "5.11b", "5.11c", "5.11d",
	"5.12a", "5.12b", "5.12c", "5.12d",
	"5.13a", "5.13b", "5.13c", "5.13d",
	"5.14a", "5.14b", "5.14c", "5.14d",
	"5.15a", "5.15b", "5.15c", "5.15d",
}

// Ordered WI grades.
var wiGrades = []string{
	"WI1", "WI2", "WI3", "WI4", "WI5", "WI6", "WI7",
}

// Ordered AI (Alpine Ice) grades.
var aiGrades = []string{
	"AI1", "AI2", "AI3", "AI4", "AI5", "AI6",
}

// Ordered mixed grades.
var mixedGrades = []string{
	"M1", "M2", "M3", "M4", "M5", "M6", "M7", "M8", "M9",
	"M10", "M11", "M12", "M13",
}

// Grade order offsets per family to create a single numeric namespace.
// V-scale: 0-17, YDS: 100-129, WI: 200-206, Mixed: 300-312, AI: 400-405
const (
	offsetV     = 0
	offsetYDS   = 100
	offsetWI    = 200
	offsetMixed = 300
	offsetAI    = 400
)

// Precomputed lookup maps for fast grade → order conversion.
var gradeToOrder map[string]int

func init() {
	gradeToOrder = make(map[string]int, len(vScaleGrades)+len(ydsGrades)+len(wiGrades)+len(aiGrades)+len(mixedGrades))

	for i, g := range vScaleGrades {
		gradeToOrder[strings.ToUpper(g)] = offsetV + i
	}
	for i, g := range ydsGrades {
		gradeToOrder[strings.ToUpper(g)] = offsetYDS + i
	}
	for i, g := range wiGrades {
		gradeToOrder[strings.ToUpper(g)] = offsetWI + i
	}
	for i, g := range aiGrades {
		gradeToOrder[strings.ToUpper(g)] = offsetAI + i
	}
	for i, g := range mixedGrades {
		gradeToOrder[strings.ToUpper(g)] = offsetMixed + i
	}
}

// Family classifies a grade string into its grade family.
// Returns empty string for unrecognized grades.
func Family(grade string) string {
	g := strings.ToUpper(strings.TrimSpace(grade))
	if g == "" {
		return ""
	}

	switch {
	case strings.HasPrefix(g, "V") && len(g) > 1 && g[1] >= '0' && g[1] <= '9':
		return FamilyV
	case strings.HasPrefix(g, "WI"):
		return FamilyWI
	case strings.HasPrefix(g, "AI") && len(g) > 2 && g[2] >= '0' && g[2] <= '9':
		return FamilyAI
	case strings.HasPrefix(g, "M") && len(g) > 1 && g[1] >= '0' && g[1] <= '9':
		return FamilyMixed
	case strings.HasPrefix(g, "A") || strings.HasPrefix(g, "C"):
		return FamilyAid
	case strings.HasPrefix(g, "5."):
		return FamilyYDS
	case strings.HasPrefix(g, "5") && len(g) > 1 && g[1] >= '0' && g[1] <= '9':
		return FamilyYDS
	default:
		return ""
	}
}

// ToOrder converts a grade string to its numeric sort order.
// Returns -1 if the grade is not recognized.
func ToOrder(grade string) int {
	g := strings.ToUpper(strings.TrimSpace(grade))
	if g == "" {
		return -1
	}

	// Try direct lookup first
	if order, ok := gradeToOrder[g]; ok {
		return order
	}

	// Handle grades with modifiers like "V4+" or "5.10a/b" or "5.10" (no letter)
	normalized := normalizeGrade(g)
	if order, ok := gradeToOrder[normalized]; ok {
		return order
	}

	return -1
}

// normalizeGrade attempts to clean grade strings for lookup.
func normalizeGrade(g string) string {
	// Strip +/- suffix ("V4+" → "V4", "5.11a/b" → "5.11A")
	g = strings.TrimRight(g, "+-")

	// Handle slash grades: "5.10a/b" → "5.10A" (take first)
	if idx := strings.Index(g, "/"); idx > 0 {
		g = g[:idx]
	}

	// Handle "5.10" without letter → "5.10A"
	family := Family(g)
	if family == FamilyYDS && len(g) >= 4 {
		last := g[len(g)-1]
		if last >= '0' && last <= '9' {
			// Check if this is a numeric-only YDS grade like "5.10"
			numPart := g[2:] // Strip "5."
			if len(numPart) >= 2 {
				// "5.10" → "5.10A", "5.11" → "5.11A"
				g = g + "A"
			}
		}
	}

	return g
}

// OrderToGrade converts a numeric order back to its grade string.
// Returns empty string if the order is not valid.
func OrderToGrade(order int) string {
	switch {
	case order >= offsetAI && order < offsetAI+len(aiGrades):
		return aiGrades[order-offsetAI]
	case order >= offsetMixed && order < offsetMixed+len(mixedGrades):
		return mixedGrades[order-offsetMixed]
	case order >= offsetWI && order < offsetWI+len(wiGrades):
		return wiGrades[order-offsetWI]
	case order >= offsetYDS && order < offsetYDS+len(ydsGrades):
		return ydsGrades[order-offsetYDS]
	case order >= offsetV && order < offsetV+len(vScaleGrades):
		return vScaleGrades[order-offsetV]
	default:
		return ""
	}
}

// VScaleGrades returns the ordered list of V-scale grades.
func VScaleGrades() []string {
	result := make([]string, len(vScaleGrades))
	copy(result, vScaleGrades)
	return result
}

// YDSGrades returns the ordered list of YDS grades.
func YDSGrades() []string {
	result := make([]string, len(ydsGrades))
	copy(result, ydsGrades)
	return result
}

// WIGrades returns the ordered list of WI grades.
func WIGrades() []string {
	result := make([]string, len(wiGrades))
	copy(result, wiGrades)
	return result
}

// AIGrades returns the ordered list of Alpine Ice grades.
func AIGrades() []string {
	result := make([]string, len(aiGrades))
	copy(result, aiGrades)
	return result
}

// MixedGrades returns the ordered list of mixed grades.
func MixedGrades() []string {
	result := make([]string, len(mixedGrades))
	copy(result, mixedGrades)
	return result
}

// ParseGradeRange converts min/max grade strings to their numeric orders.
// If a grade string is empty or unrecognized, the corresponding return value is nil.
func ParseGradeRange(minGrade, maxGrade string) (*int, *int) {
	var minOrder, maxOrder *int

	if minGrade != "" {
		if order := ToOrder(minGrade); order >= 0 {
			minOrder = &order
		}
	}

	if maxGrade != "" {
		if order := ToOrder(maxGrade); order >= 0 {
			maxOrder = &order
		}
	}

	return minOrder, maxOrder
}

package grades

import (
	"testing"
)

func TestFamily(t *testing.T) {
	tests := []struct {
		grade string
		want  string
	}{
		// V-scale
		{"V0", FamilyV},
		{"V4", FamilyV},
		{"V10", FamilyV},
		{"V17", FamilyV},
		{"v3", FamilyV},
		{"V4+", FamilyV},

		// YDS
		{"5.4", FamilyYDS},
		{"5.9", FamilyYDS},
		{"5.10a", FamilyYDS},
		{"5.11b", FamilyYDS},
		{"5.12c", FamilyYDS},
		{"5.14d", FamilyYDS},
		{"5.15a", FamilyYDS},

		// WI
		{"WI1", FamilyWI},
		{"WI3", FamilyWI},
		{"WI7", FamilyWI},
		{"wi4", FamilyWI},

		// Mixed
		{"M1", FamilyMixed},
		{"M5", FamilyMixed},
		{"M13", FamilyMixed},
		{"m7", FamilyMixed},

		// Aid
		{"A0", FamilyAid},
		{"A3", FamilyAid},
		{"C2", FamilyAid},

		// Unknown
		{"", ""},
		{"Font 7A", ""},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.grade, func(t *testing.T) {
			got := Family(tt.grade)
			if got != tt.want {
				t.Errorf("Family(%q) = %q, want %q", tt.grade, got, tt.want)
			}
		})
	}
}

func TestToOrder(t *testing.T) {
	tests := []struct {
		grade string
		want  int
	}{
		// V-scale
		{"V0", 0},
		{"V1", 1},
		{"V4", 4},
		{"V10", 10},
		{"V17", 17},

		// V-scale case insensitive
		{"v3", 3},

		// V-scale with modifiers
		{"V4+", 4},

		// YDS
		{"5.4", 100},
		{"5.5", 101},
		{"5.9", 105},
		{"5.10a", 106},
		{"5.10b", 107},
		{"5.10c", 108},
		{"5.10d", 109},
		{"5.11a", 110},
		{"5.12a", 114},
		{"5.13a", 118},
		{"5.14a", 122},
		{"5.15a", 126},
		{"5.15d", 129},

		// YDS without letter
		{"5.10", 106}, // maps to 5.10a

		// WI
		{"WI1", 200},
		{"WI3", 202},
		{"WI7", 206},

		// Mixed
		{"M1", 300},
		{"M5", 304},
		{"M13", 312},

		// Unknown
		{"", -1},
		{"unknown", -1},
		{"Font 7A", -1},
	}

	for _, tt := range tests {
		t.Run(tt.grade, func(t *testing.T) {
			got := ToOrder(tt.grade)
			if got != tt.want {
				t.Errorf("ToOrder(%q) = %d, want %d", tt.grade, got, tt.want)
			}
		})
	}
}

func TestOrderToGrade(t *testing.T) {
	tests := []struct {
		order int
		want  string
	}{
		{0, "V0"},
		{4, "V4"},
		{17, "V17"},
		{100, "5.4"},
		{106, "5.10a"},
		{129, "5.15d"},
		{200, "WI1"},
		{206, "WI7"},
		{300, "M1"},
		{312, "M13"},
		{-1, ""},
		{999, ""},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := OrderToGrade(tt.order)
			if got != tt.want {
				t.Errorf("OrderToGrade(%d) = %q, want %q", tt.order, got, tt.want)
			}
		})
	}
}

func TestToOrderAndBack(t *testing.T) {
	// Roundtrip test: every grade in our scales should survive ToOrder → OrderToGrade
	allGrades := append(append(append(VScaleGrades(), YDSGrades()...), WIGrades()...), MixedGrades()...)
	for _, g := range allGrades {
		order := ToOrder(g)
		if order < 0 {
			t.Errorf("ToOrder(%q) returned -1, expected valid order", g)
			continue
		}
		back := OrderToGrade(order)
		if back != g {
			t.Errorf("Roundtrip failed: %q → %d → %q", g, order, back)
		}
	}
}

func TestParseGradeRange(t *testing.T) {
	tests := []struct {
		name     string
		minGrade string
		maxGrade string
		wantMin  *int
		wantMax  *int
	}{
		{
			name:     "V-scale range",
			minGrade: "V3",
			maxGrade: "V10",
			wantMin:  intPtr(3),
			wantMax:  intPtr(10),
		},
		{
			name:     "YDS range",
			minGrade: "5.9",
			maxGrade: "5.12a",
			wantMin:  intPtr(105),
			wantMax:  intPtr(114),
		},
		{
			name:     "empty min",
			minGrade: "",
			maxGrade: "V10",
			wantMin:  nil,
			wantMax:  intPtr(10),
		},
		{
			name:     "empty max",
			minGrade: "V3",
			maxGrade: "",
			wantMin:  intPtr(3),
			wantMax:  nil,
		},
		{
			name:     "both empty",
			minGrade: "",
			maxGrade: "",
			wantMin:  nil,
			wantMax:  nil,
		},
		{
			name:     "unknown grades",
			minGrade: "junk",
			maxGrade: "trash",
			wantMin:  nil,
			wantMax:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMin, gotMax := ParseGradeRange(tt.minGrade, tt.maxGrade)
			if !intPtrEqual(gotMin, tt.wantMin) {
				t.Errorf("ParseGradeRange(%q, %q) min = %v, want %v", tt.minGrade, tt.maxGrade, derefInt(gotMin), derefInt(tt.wantMin))
			}
			if !intPtrEqual(gotMax, tt.wantMax) {
				t.Errorf("ParseGradeRange(%q, %q) max = %v, want %v", tt.minGrade, tt.maxGrade, derefInt(gotMax), derefInt(tt.wantMax))
			}
		})
	}
}

func TestVScaleOrdering(t *testing.T) {
	// V0 should be less than V1, V1 less than V2, etc.
	grades := VScaleGrades()
	for i := 1; i < len(grades); i++ {
		prev := ToOrder(grades[i-1])
		curr := ToOrder(grades[i])
		if prev >= curr {
			t.Errorf("Expected %s (%d) < %s (%d)", grades[i-1], prev, grades[i], curr)
		}
	}
}

func TestYDSOrdering(t *testing.T) {
	// 5.4 should be less than 5.5, etc.
	grades := YDSGrades()
	for i := 1; i < len(grades); i++ {
		prev := ToOrder(grades[i-1])
		curr := ToOrder(grades[i])
		if prev >= curr {
			t.Errorf("Expected %s (%d) < %s (%d)", grades[i-1], prev, grades[i], curr)
		}
	}
}

func TestSlashGrade(t *testing.T) {
	// "5.10a/b" should map to 5.10a
	order := ToOrder("5.10a/b")
	expected := ToOrder("5.10a")
	if order != expected {
		t.Errorf("ToOrder(\"5.10a/b\") = %d, want %d (same as 5.10a)", order, expected)
	}
}

func intPtr(v int) *int {
	return &v
}

func intPtrEqual(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func derefInt(p *int) string {
	if p == nil {
		return "nil"
	}
	return string(rune('0' + *p))
}

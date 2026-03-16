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

func TestParseGradeOrders(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []int
	}{
		{
			name:  "empty string",
			input: "",
			want:  nil,
		},
		{
			name:  "single boulder grade V9",
			input: "9",
			want:  []int{9},
		},
		{
			name:  "boulder V9-V17 range",
			input: "9,10,11,12,13,14,15,16,17",
			want:  []int{9, 10, 11, 12, 13, 14, 15, 16, 17},
		},
		{
			name:  "YDS grades 5.10a-5.12d",
			input: "106,107,108,109,110,111,112,113,114,115,116,117",
			want:  []int{106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117},
		},
		{
			name:  "multi-type: boulder + ice filtering",
			input: "0,1,2,200,201,202",
			want:  []int{0, 1, 2, 200, 201, 202},
		},
		{
			name:  "multi-type: boulder V9-V17 + WI grades",
			input: "9,10,11,12,13,14,15,16,17,200,201,202,203,204,205,206",
			want:  []int{9, 10, 11, 12, 13, 14, 15, 16, 17, 200, 201, 202, 203, 204, 205, 206},
		},
		{
			name:  "with spaces",
			input: " 9 , 10 , 11 ",
			want:  []int{9, 10, 11},
		},
		{
			name:  "invalid entries skipped",
			input: "9,abc,11,xyz",
			want:  []int{9, 11},
		},
		{
			name:  "negative values skipped",
			input: "9,-1,10",
			want:  []int{9, 10},
		},
		{
			name:  "all invalid returns nil",
			input: "abc,def",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseGradeOrders(tt.input)
			if tt.want == nil {
				if got != nil {
					t.Errorf("ParseGradeOrders(%q) = %v, want nil", tt.input, got)
				}
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("ParseGradeOrders(%q) length = %d, want %d", tt.input, len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ParseGradeOrders(%q)[%d] = %d, want %d", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestGradeOrdersMatchExpectedValues(t *testing.T) {
	// Verify that grade order values used in the API match the expected grades.
	// This test ensures the frontend grade_orders parameter values are correct.

	// Boulder V9-V17 should be orders 9-17
	for i := 9; i <= 17; i++ {
		grade := OrderToGrade(i)
		if grade == "" {
			t.Errorf("OrderToGrade(%d) = empty, expected V%d", i, i)
		}
		expectedGrade := "V" + string(rune('0'+i))
		if i >= 10 {
			// Need proper string conversion for V10+
			switch i {
			case 10:
				expectedGrade = "V10"
			case 11:
				expectedGrade = "V11"
			case 12:
				expectedGrade = "V12"
			case 13:
				expectedGrade = "V13"
			case 14:
				expectedGrade = "V14"
			case 15:
				expectedGrade = "V15"
			case 16:
				expectedGrade = "V16"
			case 17:
				expectedGrade = "V17"
			}
		}
		if grade != expectedGrade {
			t.Errorf("OrderToGrade(%d) = %q, want %q", i, grade, expectedGrade)
		}
	}

	// WI1-WI7 should be orders 200-206
	for i := 0; i < 7; i++ {
		grade := OrderToGrade(200 + i)
		if grade == "" {
			t.Errorf("OrderToGrade(%d) = empty, expected WI%d", 200+i, i+1)
		}
	}

	// Multi-type filtering: all grade families should have distinct order ranges
	// (no collisions between boulder, YDS, WI, Mixed, AI)
	boulderMax := ToOrder("V17") // 17
	ydsMin := ToOrder("5.4")     // 100
	ydsMax := ToOrder("5.15d")   // 129
	wiMin := ToOrder("WI1")      // 200
	mixedMin := ToOrder("M1")    // 300

	if boulderMax >= ydsMin {
		t.Errorf("Boulder max order (%d) overlaps with YDS min (%d)", boulderMax, ydsMin)
	}
	if ydsMax >= wiMin {
		t.Errorf("YDS max order (%d) overlaps with WI min (%d)", ydsMax, wiMin)
	}
	if wiMin >= mixedMin {
		// WI and Mixed should not overlap
		wiMax := ToOrder("WI7") // 206
		if wiMax >= mixedMin {
			t.Errorf("WI max order (%d) overlaps with Mixed min (%d)", wiMax, mixedMin)
		}
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

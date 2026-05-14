package main

import "testing"

func TestIsCompatibleMatch_HardRejectsBoulderToIceRouteType(t *testing.T) {
	ok := isCompatibleMatch("Bouldering", "V4", "Ice", "WI2")
	if ok {
		t.Fatalf("expected incompatibility for Kaya boulder vs MP ice")
	}
}

func TestIsCompatibleMatch_HardRejectsGradeFamilyMismatch(t *testing.T) {
	ok := isCompatibleMatch("Bouldering", "V4", "Boulder", "WI2")
	if ok {
		t.Fatalf("expected incompatibility for V-grade vs WI-grade")
	}
}

func TestIsCompatibleMatch_AllowsBoulderToBoulder(t *testing.T) {
	ok := isCompatibleMatch("Bouldering", "V4", "Boulder", "V4")
	if !ok {
		t.Fatalf("expected compatibility for Kaya boulder vs MP boulder with V-grade")
	}
}

func TestIsCompatibleMatch_HardRejectsBoulderToNonBoulderRouteType(t *testing.T) {
	ok := isCompatibleMatch("Bouldering", "V5", "Sport", "5.12a")
	if ok {
		t.Fatalf("expected incompatibility for Kaya boulder vs MP sport route")
	}
}

// TestIsCompatibleMatch_TableDriven covers the Font-grade and unknown-Kaya-discipline
// scenarios that were previously bypassing the discipline guard.
func TestIsCompatibleMatch_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		kayaType    string
		kayaGrade   string
		mpRouteType string
		mpRating    string
		want        bool
	}{
		{
			name:        "Font-grade boulder vs MP Trad",
			kayaType:    "",
			kayaGrade:   "7A",
			mpRouteType: "Trad",
			mpRating:    "5.10a",
			want:        false,
		},
		{
			name:        "Font-grade boulder vs MP Ice",
			kayaType:    "",
			kayaGrade:   "6C+",
			mpRouteType: "Ice",
			mpRating:    "WI4",
			want:        false,
		},
		{
			name:        "Font-grade boulder vs MP Boulder",
			kayaType:    "",
			kayaGrade:   "7A",
			mpRouteType: "Boulder",
			mpRating:    "V6",
			want:        true,
		},
		{
			name:        "Explicit Bouldering vs mixed Sport,Boulder type",
			kayaType:    "Bouldering",
			kayaGrade:   "V5",
			mpRouteType: "Sport, Boulder",
			mpRating:    "V5",
			want:        true,
		},
		{
			name:        "Unknown Kaya discipline (Font 7A) vs MP Sport",
			kayaType:    "",
			kayaGrade:   "7A",
			mpRouteType: "Sport",
			mpRating:    "5.11a",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isCompatibleMatch(tt.kayaType, tt.kayaGrade, tt.mpRouteType, tt.mpRating)
			if got != tt.want {
				t.Fatalf("isCompatibleMatch(%q,%q,%q,%q) = %v, want %v",
					tt.kayaType, tt.kayaGrade, tt.mpRouteType, tt.mpRating, got, tt.want)
			}
		})
	}
}

func TestGradeFamily(t *testing.T) {
	tests := []struct {
		name  string
		grade string
		want  string
	}{
		{name: "v scale", grade: "V8", want: "v"},
		{name: "wi", grade: "WI3", want: "wi"},
		{name: "yds", grade: "5.11b", want: "yds"},
		{name: "unknown", grade: "Font 7A", want: ""},
		{name: "font 7A", grade: "7A", want: "v"},
		{name: "font 6C+", grade: "6C+", want: "v"},
		{name: "font 8B+", grade: "8B+", want: "v"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := gradeFamily(tt.grade)
			if got != tt.want {
				t.Fatalf("gradeFamily(%q) = %q, want %q", tt.grade, got, tt.want)
			}
		})
	}
}

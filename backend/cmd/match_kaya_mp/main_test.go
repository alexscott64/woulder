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

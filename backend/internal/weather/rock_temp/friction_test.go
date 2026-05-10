package rock_temp

import "testing"

func TestComputeFrictionQuality_Dry(t *testing.T) {
	cases := []struct {
		temp string
		want string
	}{
		{"prime", "excellent"},
		{"good", "good"},
		{"marginal", "reduced"},
		{"poor", "poor"},
		{"very_poor", "poor"},
		{"too_cold", "poor"},
	}
	for _, c := range cases {
		got := ComputeFrictionQuality(c.temp, "none")
		if got != c.want {
			t.Errorf("dry %s: got %q want %q", c.temp, got, c.want)
		}
	}
}

func TestComputeFrictionQuality_Light(t *testing.T) {
	cases := []struct {
		temp string
		want string
	}{
		{"prime", "reduced"},
		{"good", "reduced"},
		{"marginal", "poor"},
		{"poor", "poor"},
		{"very_poor", "poor"},
		{"too_cold", "poor"},
	}
	for _, c := range cases {
		got := ComputeFrictionQuality(c.temp, "light")
		if got != c.want {
			t.Errorf("light %s: got %q want %q", c.temp, got, c.want)
		}
	}
}

func TestComputeFrictionQuality_Heavy(t *testing.T) {
	for _, temp := range []string{"prime", "good", "marginal", "poor", "very_poor", "too_cold"} {
		got := ComputeFrictionQuality(temp, "heavy")
		if got != "poor" {
			t.Errorf("heavy %s: got %q want poor", temp, got)
		}
	}
}

func TestComputeFrictionQuality_Unknown(t *testing.T) {
	if got := ComputeFrictionQuality("nope", "none"); got != "unknown" {
		t.Errorf("unknown temp + none: got %q", got)
	}
	if got := ComputeFrictionQuality("prime", "weird"); got != "unknown" {
		t.Errorf("prime + unknown severity: got %q", got)
	}
}

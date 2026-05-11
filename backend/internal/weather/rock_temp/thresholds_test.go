package rock_temp

import "testing"

func TestClassifyTempCondition_Granite(t *testing.T) {
	th := Granite.Thresholds
	cases := []struct {
		surface float64
		want    string
	}{
		{20, "too_cold"}, // < 30
		{29.99, "too_cold"},
		{30, "too_cold"}, // gap between TooColdMax(30) and PrimeMin(35) is too_cold
		{34.99, "too_cold"},
		{35, "prime"},
		{45, "prime"},
		{54.99, "prime"},
		{55, "good"},
		{64.99, "good"},
		{65, "marginal"},
		{71.99, "marginal"},
		{72, "poor"},
		{84.99, "poor"},
		{85, "very_poor"},
		{120, "very_poor"},
	}
	for _, c := range cases {
		got := ClassifyTempCondition(c.surface, th)
		if got != c.want {
			t.Errorf("granite %.2f°F: got %q want %q", c.surface, got, c.want)
		}
	}
}

func TestClassifyTempCondition_Sandstone(t *testing.T) {
	th := Sandstone.Thresholds
	if ClassifyTempCondition(24, th) != "too_cold" {
		t.Errorf("sandstone 24°F should be too_cold")
	}
	if ClassifyTempCondition(30, th) != "prime" {
		t.Errorf("sandstone 30°F should be prime")
	}
	if ClassifyTempCondition(50, th) != "good" {
		t.Errorf("sandstone 50°F should be good")
	}
	if ClassifyTempCondition(81, th) != "very_poor" {
		t.Errorf("sandstone 81°F should be very_poor")
	}
}

func TestClassifyTempCondition_Limestone(t *testing.T) {
	th := Limestone.Thresholds
	if ClassifyTempCondition(40, th) != "prime" {
		t.Errorf("limestone 40°F should be prime")
	}
	if ClassifyTempCondition(86, th) != "very_poor" {
		t.Errorf("limestone 86°F should be very_poor")
	}
}

func TestClassifyTempCondition_Quartzite(t *testing.T) {
	th := Quartzite.Thresholds
	if ClassifyTempCondition(35, th) != "prime" {
		t.Errorf("quartzite 35°F should be prime")
	}
	if ClassifyTempCondition(58, th) != "good" {
		t.Errorf("quartzite 58°F should be good")
	}
}

func TestClassifyTempCondition_Basalt(t *testing.T) {
	th := BasaltGabbro.Thresholds
	if ClassifyTempCondition(35, th) != "prime" {
		t.Errorf("basalt 35°F should be prime")
	}
	if ClassifyTempCondition(63, th) != "marginal" {
		t.Errorf("basalt 63°F should be marginal")
	}
	if ClassifyTempCondition(86, th) != "very_poor" {
		t.Errorf("basalt 86°F should be very_poor")
	}
}

func TestParamsForGroup_CaseInsensitiveAndVariants(t *testing.T) {
	cases := []struct {
		in        string
		want      string
		confident bool
	}{
		{"Granite", "Granite", true},
		{"granite", "Granite", true},
		{"GRANITE", "Granite", true},
		{"Sandstone", "Sandstone", true},
		{"sandstone", "Sandstone", true},
		{"Basalt", "Basalt/Gabbro", true},
		{"Gabbro", "Basalt/Gabbro", true},
		{"basalt/gabbro", "Basalt/Gabbro", true},
		{"Basalt / Gabbro", "Basalt/Gabbro", true},
		{"Quartzite", "Quartzite", true},
		{"Limestone", "Limestone", true},
		{"Yosemite Granite", "Granite", true}, // substring match
		{"unknownrock", "Granite", false},
		{"", "Granite", false},
	}
	for _, c := range cases {
		got, ok := ParamsForGroup(c.in)
		if got.GroupName != c.want {
			t.Errorf("ParamsForGroup(%q): got %q want %q", c.in, got.GroupName, c.want)
		}
		if ok != c.confident {
			t.Errorf("ParamsForGroup(%q): confidence got %v want %v", c.in, ok, c.confident)
		}
	}
}

func TestDefaultThermalParamsIsGranite(t *testing.T) {
	if DefaultThermalParams().GroupName != "Granite" {
		t.Errorf("default should be Granite")
	}
}

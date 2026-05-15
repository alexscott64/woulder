package geo

import "testing"

func TestLookupTimezone(t *testing.T) {
	cases := []struct {
		name     string
		lat, lon float64
		want     string
	}{
		{"Icicle Creek / Leavenworth, WA", 47.59, -120.78, "America/Los_Angeles"},
		{"Red Rocks, NV", 36.13, -115.45, "America/Los_Angeles"},
		{"Squamish, BC", 49.70, -123.16, "America/Vancouver"},
		{"Boulder, CO", 40.01, -105.27, "America/Denver"},
		{"New York City, NY", 40.71, -74.00, "America/New_York"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := LookupTimezone(tc.lat, tc.lon)
			if got != tc.want {
				t.Errorf("LookupTimezone(%v, %v) = %q, want %q",
					tc.lat, tc.lon, got, tc.want)
			}
		})
	}
}

// TestLookupTimezone_OutOfRangeFallback covers a coordinate in the middle of
// the Atlantic where tzf may return an empty string. The wrapper must always
// return a non-empty IANA name (defaulting to America/Los_Angeles when tzf
// returns nothing).
func TestLookupTimezone_OutOfRangeFallback(t *testing.T) {
	got := LookupTimezone(0.0, 0.0)
	if got == "" {
		t.Fatalf("LookupTimezone(0,0) returned empty string; expected non-empty fallback")
	}
	t.Logf("LookupTimezone(0,0) returned %q (acceptable as long as it is non-empty)", got)
}

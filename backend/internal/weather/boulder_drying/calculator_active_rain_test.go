package boulder_drying

import (
	"context"
	"testing"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

// TestActiveRainShowsWet verifies boulders show wet during active rain
func TestActiveRainShowsWet(t *testing.T) {
	now := time.Now()

	// Scenario: It's actively raining/snowing right now
	locationDrying := &models.RockDryingStatus{
		IsWet:             true,
		HoursUntilDry:     24.0, // Location says 24h until dry
		Message:           "Wet - 24h until dry",
		IsWetSensitive:    false,
		PrimaryRockType:   "Granite",
		LastRainTimestamp: now.Format(time.RFC3339),
	}

	route := &models.MPRoute{
		MPRouteID: "test-route-1",
		Name:      "Test Boulder",
		Latitude:  ptrFloat64(37.7),
		Longitude: ptrFloat64(-119.6),
		Aspect:    ptrString("South"),
	}

	profile := &models.BoulderDryingProfile{
		TreeCoveragePercent:   ptrFloat64(30.0),
		SunExposureHoursCache: ptrString("30.0"), // Good sun exposure
	}

	calc := NewCalculator("")
	status, err := calc.CalculateBoulderDryingStatus(
		context.Background(),
		route,
		locationDrying,
		profile,
		30.0, // locationTreeCoverage
		nil,  // hourlyForecast
	)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// CRITICAL: Boulder MUST show wet during active rain
	if !status.IsWet {
		t.Errorf("CRITICAL BUG: Boulder shows DRY during active rain (location HoursUntilDry=%.1f)",
			locationDrying.HoursUntilDry)
		t.Errorf("Boulder HoursUntilDry=%.1f (should be > 0)", status.HoursUntilDry)
	}

	if status.HoursUntilDry <= 0 {
		t.Errorf("CRITICAL BUG: Boulder HoursUntilDry=%.1f (should be > 0 when location is wet)",
			status.HoursUntilDry)
	}

	if status.Status == "good" {
		t.Errorf("CRITICAL BUG: Boulder status='good' during active rain (should be 'poor' or 'fair')")
	}
}

// TestLocationDryTimeZeroMeansDry verifies that HoursUntilDry=0 means dry
func TestLocationDryTimeZeroMeansDry(t *testing.T) {
	// Scenario: Location is completely dry (no recent rain)
	locationDrying := &models.RockDryingStatus{
		IsWet:             false,
		HoursUntilDry:     0, // Location is dry
		Message:           "Dry and ready to climb",
		IsWetSensitive:    false,
		PrimaryRockType:   "Granite",
		LastRainTimestamp: time.Now().Add(-72 * time.Hour).Format(time.RFC3339),
	}

	route := &models.MPRoute{
		MPRouteID: "test-route-2",
		Name:      "Test Boulder",
		Latitude:  ptrFloat64(37.7),
		Longitude: ptrFloat64(-119.6),
		Aspect:    ptrString("South"),
	}

	calc := NewCalculator("")
	status, err := calc.CalculateBoulderDryingStatus(
		context.Background(),
		route,
		locationDrying,
		nil, // No profile
		30.0, // locationTreeCoverage
		nil,  // hourlyForecast
	)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// When location is dry, boulder should also be dry
	if status.IsWet {
		t.Errorf("Boulder shows WET when location is dry (location HoursUntilDry=%.1f)",
			locationDrying.HoursUntilDry)
	}

	if status.HoursUntilDry != 0 {
		t.Errorf("Boulder HoursUntilDry=%.1f (should be 0 when location is dry)",
			status.HoursUntilDry)
	}

	if status.Status != "good" {
		t.Errorf("Boulder status='%s' (should be 'good' when dry)", status.Status)
	}
}

// TestHeavySnowShowsWet verifies boulders show wet during heavy snow
func TestHeavySnowShowsWet(t *testing.T) {
	now := time.Now()

	// Scenario: Heavy snow accumulation (like Tramway)
	locationDrying := &models.RockDryingStatus{
		IsWet:             true,
		HoursUntilDry:     120.0, // 5 days until dry due to snow
		Message:           "Snow/ice present - 120h until dry",
		IsWetSensitive:    false,
		PrimaryRockType:   "Granite",
		LastRainTimestamp: now.Format(time.RFC3339),
	}

	route := &models.MPRoute{
		MPRouteID: "test-route-3",
		Name:      "Snowy Boulder",
		Latitude:  ptrFloat64(33.8),
		Longitude: ptrFloat64(-116.6),
		Aspect:    ptrString("North"), // Worst aspect for drying
	}

	profile := &models.BoulderDryingProfile{
		TreeCoveragePercent:   ptrFloat64(60.0), // Heavy tree cover
		SunExposureHoursCache: ptrString("12.0"), // Poor sun exposure
	}

	calc := NewCalculator("")
	status, err := calc.CalculateBoulderDryingStatus(
		context.Background(),
		route,
		locationDrying,
		profile,
		60.0, // locationTreeCoverage
		nil,  // hourlyForecast
	)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// CRITICAL: Boulder MUST show wet during heavy snow
	if !status.IsWet {
		t.Errorf("CRITICAL BUG: Boulder shows DRY during heavy snow (location HoursUntilDry=%.1f)",
			locationDrying.HoursUntilDry)
	}

	// Snow with poor sun/heavy trees should take even longer than location
	if status.HoursUntilDry < locationDrying.HoursUntilDry {
		t.Errorf("Boulder dries faster than location (%.1fh vs %.1fh) - should be slower with poor conditions",
			status.HoursUntilDry, locationDrying.HoursUntilDry)
	}
}

// Helper functions
func ptrFloat64(f float64) *float64 {
	return &f
}

func ptrString(s string) *string {
	return &s
}

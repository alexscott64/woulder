package boulder_drying

import (
	"context"
	"log"
	"os"
	"strings"

	earthengine "github.com/alexscott64/go-earthengine"
	"github.com/alexscott64/go-earthengine/helpers"
)

// TreeCoverClient fetches tree canopy coverage data from Google Earth Engine
type TreeCoverClient struct {
	geeClient *earthengine.Client
	enabled   bool
}

// IsEnabled returns whether the Google Earth Engine API is enabled
func (c *TreeCoverClient) IsEnabled() bool {
	return c.enabled
}

// NewTreeCoverClient creates a new tree cover API client using Google Earth Engine
func NewTreeCoverClient() *TreeCoverClient {
	projectID := os.Getenv("GOOGLE_EARTH_ENGINE_PROJECT_ID")
	clientEmail := os.Getenv("GOOGLE_EARTH_ENGINE_CLIENT_EMAIL")
	privateKey := os.Getenv("GOOGLE_EARTH_ENGINE_PRIVATE_KEY")

	if projectID == "" || clientEmail == "" || privateKey == "" {
		log.Printf("Warning: Google Earth Engine credentials not set - using location-based estimates")
		log.Printf("Missing: project_id=%v, client_email=%v, private_key=%v",
			projectID == "", clientEmail == "", privateKey == "")
		return &TreeCoverClient{enabled: false}
	}

	// Fix private key: replace literal \n with actual newlines
	// .env files often store the private key with literal \n characters
	privateKey = strings.ReplaceAll(privateKey, "\\n", "\n")

	// Set the fixed private key back to environment for the library to use
	os.Setenv("GOOGLE_EARTH_ENGINE_PRIVATE_KEY", privateKey)

	// Create Earth Engine client using environment variables
	ctx := context.Background()
	client, err := earthengine.NewClient(ctx,
		earthengine.WithProject(projectID),
		earthengine.WithServiceAccountEnv(),
	)
	if err != nil {
		log.Printf("Warning: Failed to initialize Earth Engine client: %v - using location-based estimates", err)
		return &TreeCoverClient{enabled: false}
	}

	log.Printf("Google Earth Engine client initialized successfully (project: %s)", projectID)
	return &TreeCoverClient{
		geeClient: client,
		enabled:   true,
	}
}

// GetTreeCoverageWithDefault returns tree canopy coverage percentage for a GPS coordinate
// Uses NLCD 2023 dataset (USA) or Hansen Global Forest Change (international)
// Returns percentage 0-100
// locationTreeCoverage: optional location-level tree coverage to use as fallback (pass 0 to use GPS-based estimates)
func (c *TreeCoverClient) GetTreeCoverageWithDefault(ctx context.Context, lat, lon, locationTreeCoverage float64) (float64, error) {
	if !c.enabled {
		// Use location-level tree coverage if provided, otherwise estimate from GPS
		var coverage float64
		if locationTreeCoverage > 0 {
			coverage = locationTreeCoverage
			log.Printf("Tree coverage for (%.6f, %.6f): %.1f%% (from location)", lat, lon, coverage)
		} else {
			coverage = c.estimateTreeCoverageFromLocation(lat, lon)
			log.Printf("Tree coverage estimate for (%.6f, %.6f): %.1f%% (GPS-based)", lat, lon, coverage)
		}
		return coverage, nil
	}

	// Try to fetch from Earth Engine using the go-earthengine library
	// This will use NLCD 2023 for USA locations (most accurate)
	coverage, err := helpers.TreeCoverage(c.geeClient, lat, lon)
	if err != nil {
		log.Printf("Warning: Earth Engine query failed, using fallback: %v", err)
		// Fallback to location tree coverage first, then GPS estimates
		if locationTreeCoverage > 0 {
			coverage = locationTreeCoverage
			log.Printf("Tree coverage for (%.6f, %.6f): %.1f%% (from location fallback)", lat, lon, coverage)
		} else {
			coverage = c.estimateTreeCoverageFromLocation(lat, lon)
			log.Printf("Tree coverage for (%.6f, %.6f): %.1f%% (GPS-based fallback)", lat, lon, coverage)
		}
	} else {
		log.Printf("Tree coverage for (%.6f, %.6f): %.1f%% (from Earth Engine)", lat, lon, coverage)
	}

	return coverage, nil
}

// GetTreeCoverage returns tree canopy coverage percentage for a GPS coordinate (no location default)
// Wrapper for backwards compatibility
func (c *TreeCoverClient) GetTreeCoverage(ctx context.Context, lat, lon float64) (float64, error) {
	return c.GetTreeCoverageWithDefault(ctx, lat, lon, 0)
}

// estimateTreeCoverageFromLocation provides tree coverage estimates
// Based on typical forest coverage in known climbing areas
func (c *TreeCoverClient) estimateTreeCoverageFromLocation(lat, lon float64) float64 {
	// Leavenworth area (Washington Cascades): moderate to high tree coverage
	// Index Town Wall, Icicle Creek, Peshastin Pinnacles
	if lat >= 47.5 && lat <= 48.0 && lon >= -121.7 && lon <= -120.5 {
		// Different zones based on longitude
		if lon < -121.4 { // Index/Town Wall area - dense forest
			return 65.0 // Very dense forest (Index, Zelda)
		} else if lon < -120.8 { // Icicle Creek
			return 60.0 // Dense forest
		} else { // Town walls and Peshastin
			return 25.0 // Mixed, more exposed
		}
	}

	// Bishop area (California): sparse trees
	// Buttermilks, Happy Boulders, Sad Boulders
	if lat >= 37.0 && lat <= 37.5 && lon >= -119.0 && lon <= -118.5 {
		if lat > 37.3 { // Buttermilks area
			return 5.0 // Very sparse, high desert
		}
		return 15.0 // Slightly more vegetation in canyons
	}

	// Squamish area (British Columbia): heavy tree coverage
	if lat >= 49.5 && lat <= 50.0 && lon >= -123.5 && lon <= -123.0 {
		return 70.0 // Coastal rainforest
	}

	// Red Rocks (Nevada): minimal trees
	if lat >= 36.0 && lat <= 36.3 && lon >= -115.6 && lon <= -115.3 {
		return 2.0 // Desert
	}

	// Smith Rock (Oregon): sparse trees
	if lat >= 44.3 && lat <= 44.4 && lon >= -121.2 && lon <= -121.1 {
		return 10.0 // High desert with juniper
	}

	// Joshua Tree (California): minimal trees
	if lat >= 33.8 && lat <= 34.2 && lon >= -116.4 && lon <= -116.0 {
		return 3.0 // Desert with joshua trees
	}

	// Yosemite Valley (California): heavy tree coverage
	if lat >= 37.7 && lat <= 37.8 && lon >= -119.7 && lon <= -119.5 {
		return 55.0 // Mixed conifer forest
	}

	// Default: moderate coverage for unknown areas
	return 30.0
}

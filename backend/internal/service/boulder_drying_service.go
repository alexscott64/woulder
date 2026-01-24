package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/alexscott64/woulder/backend/internal/database"
	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/alexscott64/woulder/backend/internal/weather"
	"github.com/alexscott64/woulder/backend/internal/weather/boulder_drying"
	"github.com/alexscott64/woulder/backend/internal/weather/calculator"
	"github.com/alexscott64/woulder/backend/internal/weather/rock_drying"
)

// BoulderDryingService handles boulder-specific drying calculations
type BoulderDryingService struct {
	repo          database.Repository
	calculator    *boulder_drying.Calculator
	weatherClient *weather.WeatherService
}

// NewBoulderDryingService creates a new boulder drying service
func NewBoulderDryingService(repo database.Repository, weatherClient *weather.WeatherService) *BoulderDryingService {
	return &BoulderDryingService{
		repo:          repo,
		calculator:    boulder_drying.NewCalculator(""), // API key no longer used (offline sun calculations)
		weatherClient: weatherClient,
	}
}

// GetBatchBoulderDryingStatus calculates the drying status for multiple boulders efficiently
func (s *BoulderDryingService) GetBatchBoulderDryingStatus(
	ctx context.Context,
	mpRouteIDs []string,
) (map[string]*boulder_drying.BoulderDryingStatus, error) {
	batchStart := time.Now()
	log.Printf("[PERF] GetBatchBoulderDryingStatus: Starting batch request for %d routes", len(mpRouteIDs))

	if len(mpRouteIDs) == 0 {
		return make(map[string]*boulder_drying.BoulderDryingStatus), nil
	}

	results := make(map[string]*boulder_drying.BoulderDryingStatus)
	failedRoutes := make(map[string]string) // Track routes that failed

	// Fetch ALL routes in a single query (eliminates N+1 problem)
	routeFetchStart := time.Now()
	routesMap, err := s.repo.GetMPRoutesByIDs(ctx, mpRouteIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch routes: %w", err)
	}
	log.Printf("[PERF] Route batch fetching took %v for %d routes", time.Since(routeFetchStart), len(mpRouteIDs))

	// Fetch ALL boulder drying profiles in a single query (eliminates N+1 problem)
	profileFetchStart := time.Now()
	profilesMap, err := s.repo.GetBoulderDryingProfilesByRouteIDs(ctx, mpRouteIDs)
	if err != nil {
		log.Printf("Warning: Failed to fetch boulder drying profiles: %v", err)
		profilesMap = make(map[string]*models.BoulderDryingProfile) // Continue with empty profiles
	}
	log.Printf("[PERF] Profile batch fetching took %v for %d routes (%d profiles found)", time.Since(profileFetchStart), len(mpRouteIDs), len(profilesMap))

	// Group routes by location to share expensive calculations
	routesByLocation := make(map[int][]*models.MPRoute)
	for _, routeID := range mpRouteIDs {
		route, found := routesMap[routeID]
		if !found {
			log.Printf("Warning: Route %s not found", routeID)
			failedRoutes[routeID] = "Route not found"
			continue
		}
		if route.LocationID == nil {
			log.Printf("Warning: Route %s has no location", routeID)
			failedRoutes[routeID] = "Route has no location"
			continue
		}

		routesByLocation[*route.LocationID] = append(routesByLocation[*route.LocationID], route)
	}

	// Process each location group
	for locationID, routes := range routesByLocation {
		locationStart := time.Now()
		log.Printf("[PERF] Processing location %d with %d routes", locationID, len(routes))

		// Get location-level rock drying status (shared for all routes in this location)
		dryingStart := time.Now()
		locationDrying, freshForecast, err := s.getLocationRockDryingStatus(ctx, locationID)
		if err != nil {
			log.Printf("Warning: Failed to get location drying status for %d: %v", locationID, err)
			continue
		}
		log.Printf("[PERF] getLocationRockDryingStatus took %v", time.Since(dryingStart))

		// Get location sun exposure (shared for all routes in this location)
		sunStart := time.Now()
		sunExposure, err := s.repo.GetSunExposureByLocation(ctx, locationID)
		if err != nil {
			log.Printf("Warning: Failed to get sun exposure for location %d: %v", locationID, err)
			sunExposure = nil
		}
		log.Printf("[PERF] GetSunExposureByLocation took %v", time.Since(sunStart))

		locationTreeCoverage := 0.0
		if sunExposure != nil {
			locationTreeCoverage = sunExposure.TreeCoveragePercent
		}

		// Use the fresh forecast from getLocationRockDryingStatus (already fetched from API)
		// This ensures boulder 6-day forecast matches the location drying calculation
		hourlyForecast := freshForecast
		log.Printf("[PERF] Using fresh forecast from API (%d hours)", len(hourlyForecast))

		// Calculate drying status for each route in this location
		for _, route := range routes {
			routeStart := time.Now()

			// Look up profile from pre-fetched map (no database query)
			profile := profilesMap[route.MPRouteID]

			calcStart := time.Now()
			status, err := s.calculator.CalculateBoulderDryingStatus(
				ctx,
				route,
				locationDrying,
				profile,
				locationTreeCoverage,
				hourlyForecast,
			)
			if err != nil {
				log.Printf("Warning: Failed to calculate boulder drying status for %s: %v", route.MPRouteID, err)
				failedRoutes[route.MPRouteID] = fmt.Sprintf("Calculation failed: %v", err)
				continue
			}
			calcTime := time.Since(calcStart)

			// NOTE: Profile caching disabled during batch requests for performance
			// Profiles should be pre-populated by background job (cmd/sync_boulder_profiles)
			// Saving profiles during user requests adds significant latency

			results[route.MPRouteID] = status
			log.Printf("[PERF] Route %s: calc=%v total=%v", route.MPRouteID, calcTime, time.Since(routeStart))
		}

		log.Printf("[PERF] Location %d processing took %v", locationID, time.Since(locationStart))
	}

	// Add placeholder statuses for routes that failed
	// This prevents frontend from falling back to potentially stale cached data
	for routeID, errorMsg := range failedRoutes {
		results[routeID] = &boulder_drying.BoulderDryingStatus{
			MPRouteID:       routeID,
			IsWet:           true, // Assume wet for safety
			IsSafe:          false,
			HoursUntilDry:   999, // Unknown
			Status:          "poor",
			Message:         fmt.Sprintf("Unable to calculate: %s", errorMsg),
			ConfidenceScore: 0,
			RockType:        "Unknown",
			Aspect:          "Unknown",
		}
	}

	log.Printf("[PERF] GetBatchBoulderDryingStatus: TOTAL TIME %v for %d routes (%d successful, %d failed)",
		time.Since(batchStart), len(mpRouteIDs), len(results)-len(failedRoutes), len(failedRoutes))

	return results, nil
}

// GetBoulderDryingStatus calculates the drying status for a specific boulder
func (s *BoulderDryingService) GetBoulderDryingStatus(
	ctx context.Context,
	mpRouteID string,
) (*boulder_drying.BoulderDryingStatus, error) {
	// Get the route
	route, err := s.repo.GetMPRouteByID(ctx, mpRouteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get route: %w", err)
	}
	if route == nil {
		return nil, fmt.Errorf("route not found: %s", mpRouteID)
	}

	// Get the location ID
	if route.LocationID == nil {
		return nil, fmt.Errorf("route has no associated location")
	}

	// Get location-level rock drying status (includes fresh forecast data)
	locationDrying, hourlyForecast, err := s.getLocationRockDryingStatus(ctx, *route.LocationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get location drying status: %w", err)
	}

	// Get boulder drying profile (if exists)
	profile, err := s.repo.GetBoulderDryingProfile(ctx, mpRouteID)
	if err != nil {
		log.Printf("Warning: Failed to get boulder drying profile for %s: %v", mpRouteID, err)
		profile = nil // Continue without profile
	}

	// Get location sun exposure (for tree coverage)
	sunExposure, err := s.repo.GetSunExposureByLocation(ctx, *route.LocationID)
	if err != nil {
		log.Printf("Warning: Failed to get sun exposure for location %d: %v", *route.LocationID, err)
		sunExposure = nil
	}

	// Extract location tree coverage
	locationTreeCoverage := 0.0
	if sunExposure != nil {
		locationTreeCoverage = sunExposure.TreeCoveragePercent
	}

	// hourlyForecast already obtained from getLocationRockDryingStatus (fresh API data)

	// Calculate boulder-specific drying status
	status, err := s.calculator.CalculateBoulderDryingStatus(ctx, route, locationDrying, profile, locationTreeCoverage, hourlyForecast)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate boulder drying status: %w", err)
	}

	// NOTE: Sun exposure is NOT cached because it's time-dependent (next 6 days from NOW)
	// Tree coverage should be pre-populated by background job (cmd/sync_tree_cover)

	return status, nil
}

// getLocationRockDryingStatus calculates location-level rock drying status
// Returns both the drying status and the fresh forecast data used for calculation
func (s *BoulderDryingService) getLocationRockDryingStatus(
	ctx context.Context,
	locationID int,
) (*models.RockDryingStatus, []models.WeatherData, error) {
	// Get location for elevation data (needed for snow calculation)
	locStart := time.Now()
	location, err := s.repo.GetLocation(ctx, locationID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get location: %w", err)
	}
	log.Printf("[PERF]   GetLocation took %v", time.Since(locStart))

	// CRITICAL: Fetch FRESH weather from API, not database cache
	// Database cache is stale and missing accurate snow/precipitation data
	// This is slower but accuracy is more important than speed for rock drying
	weatherStart := time.Now()
	currentWeather, hourlyForecast, _, err := s.weatherClient.GetCurrentAndForecast(
		location.Latitude, location.Longitude,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch current weather from API: %w", err)
	}
	log.Printf("[PERF]   GetCurrentAndForecast (API) took %v", time.Since(weatherStart))

	// Set location ID on weather data
	currentWeather.LocationID = locationID
	for i := range hourlyForecast {
		hourlyForecast[i].LocationID = locationID
	}

	// Get historical weather (last 7 days) from database
	histStart := time.Now()
	historicalWeather, err := s.repo.GetHistoricalWeather(ctx, locationID, 168) // 7 days
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get historical weather: %w", err)
	}
	log.Printf("[PERF]   GetHistoricalWeather took %v (got %d hours)", time.Since(histStart), len(historicalWeather))

	// Get rock types
	rockStart := time.Now()
	rockTypes, err := s.repo.GetRockTypesByLocation(ctx, locationID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get rock types: %w", err)
	}
	log.Printf("[PERF]   GetRockTypesByLocation took %v", time.Since(rockStart))

	// Get sun exposure
	sunStart := time.Now()
	sunExposure, err := s.repo.GetSunExposureByLocation(ctx, locationID)
	if err != nil {
		log.Printf("Warning: Failed to get sun exposure for location %d: %v", locationID, err)
		sunExposure = nil
	}
	log.Printf("[PERF]   GetSunExposureByLocation took %v", time.Since(sunStart))

	// CRITICAL FIX: Calculate current snow depth for rock drying
	// This was missing and causing boulders to show dry during active snow
	snowStart := time.Now()

	// Combine current and forecast into future data (same as weather service does)
	futureData := append([]models.WeatherData{*currentWeather}, hourlyForecast...)

	currentSnowDepth := calculator.GetCurrentSnowDepth(historicalWeather, futureData, float64(location.ElevationFt))
	var snowDepthPtr *float64
	if currentSnowDepth > 0 {
		snowDepthPtr = &currentSnowDepth
		log.Printf("[PERF]   Calculated snow depth: %.2f inches (location %d)", currentSnowDepth, locationID)
	} else {
		log.Printf("[PERF]   No snow depth calculated (location %d)", locationID)
	}
	log.Printf("[PERF]   GetCurrentSnowDepth took %v", time.Since(snowStart))

	// Calculate rock drying status using existing calculator
	calcStart := time.Now()
	calc := &rock_drying.Calculator{}

	dryingStatus := calc.CalculateDryingStatus(
		rockTypes,
		currentWeather,
		historicalWeather,
		sunExposure,
		false,       // hasSeepageRisk - TODO: Get from location profile if needed
		snowDepthPtr, // FIXED: Now passing actual snow depth from API
	)
	log.Printf("[PERF]   CalculateDryingStatus took %v", time.Since(calcStart))

	// Return both the drying status and the fresh forecast data
	// The forecast is reused by callers to avoid duplicate API calls
	return &dryingStatus, hourlyForecast, nil
}

// GetAreaDryingStats calculates aggregated drying statistics for an area
// CRITICAL: This method uses GetBatchBoulderDryingStatus internally to ensure
// area stats and individual route statuses use the EXACT same weather data
func (s *BoulderDryingService) GetAreaDryingStats(
	ctx context.Context,
	mpAreaID string,
	locationID int,
) (*models.AreaDryingStats, error) {
	// Get all routes with GPS in this area (including subareas)
	routes, err := s.repo.GetRoutesWithGPSByArea(ctx, mpAreaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get routes for area: %w", err)
	}

	if len(routes) == 0 {
		// No routes with GPS data - return nil
		return nil, nil
	}

	// Extract route IDs
	routeIDs := make([]string, len(routes))
	for i, route := range routes {
		routeIDs[i] = route.MPRouteID
	}

	// CRITICAL: Use batch endpoint to get all route statuses with SAME weather data
	// This ensures area stats match individual route displays
	routeStatuses, err := s.GetBatchBoulderDryingStatus(ctx, routeIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get batch boulder drying statuses: %w", err)
	}

	// Calculate drying status for each route and aggregate
	stats := &models.AreaDryingStats{
		TotalRoutes: len(routes),
	}

	var totalHoursUntilDry float64
	var wetRouteCount int
	var totalTreeCoverage float64
	var treeCoverageCount int
	var totalConfidence int
	var mostRecentRain *time.Time

	for _, route := range routes {
		// Get status from batch results
		status, ok := routeStatuses[route.MPRouteID]
		if !ok {
			log.Printf("Warning: No drying status found for route %s", route.MPRouteID)
			continue
		}

		// Aggregate statistics
		if !status.IsWet {
			stats.DryCount++
		} else if status.HoursUntilDry > 0 && status.HoursUntilDry <= 24 {
			stats.DryingCount++
			totalHoursUntilDry += status.HoursUntilDry
			wetRouteCount++
		} else {
			stats.WetCount++
			totalHoursUntilDry += status.HoursUntilDry
			wetRouteCount++
		}

		// Tree coverage
		if status.TreeCoveragePercent > 0 {
			totalTreeCoverage += status.TreeCoveragePercent
			treeCoverageCount++
		}

		// Confidence
		totalConfidence += status.ConfidenceScore

		// Track most recent rain timestamp
		if status.LastRainTimestamp != nil {
			if mostRecentRain == nil || status.LastRainTimestamp.After(*mostRecentRain) {
				mostRecentRain = status.LastRainTimestamp
			}
		}
	}

	// Calculate percentages and averages
	if stats.TotalRoutes > 0 {
		stats.PercentDry = (float64(stats.DryCount) / float64(stats.TotalRoutes)) * 100
	}

	if wetRouteCount > 0 {
		stats.AvgHoursUntilDry = totalHoursUntilDry / float64(wetRouteCount)
	}

	if treeCoverageCount > 0 {
		stats.AvgTreeCoverage = totalTreeCoverage / float64(treeCoverageCount)
	}

	if stats.TotalRoutes > 0 {
		stats.ConfidenceScore = totalConfidence / stats.TotalRoutes
	}

	// Set most recent rain timestamp
	stats.LastRainTimestamp = mostRecentRain

	return stats, nil
}

// GetBatchAreaDryingStats calculates drying statistics for multiple areas in a single call
// OPTIMIZED: Calculates location-level weather ONCE and reuses for all areas (massive speedup)
func (s *BoulderDryingService) GetBatchAreaDryingStats(
	ctx context.Context,
	areaIDs []string,
	locationID int,
) (map[string]*models.AreaDryingStats, error) {
	batchStart := time.Now()
	log.Printf("[PERF] GetBatchAreaDryingStats: Starting OPTIMIZED batch request for %d areas in location %d", len(areaIDs), locationID)

	if len(areaIDs) == 0 {
		return make(map[string]*models.AreaDryingStats), nil
	}

	// OPTIMIZATION: Calculate location-level rock drying status ONCE for all areas
	// This eliminates redundant weather queries (was: N * 80ms, now: 1 * 80ms)
	// Also returns fresh forecast data from API for accurate boulder 6-day forecasts
	locationStart := time.Now()
	locationDrying, hourlyForecast, err := s.getLocationRockDryingStatus(ctx, locationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get location drying status: %w", err)
	}
	log.Printf("[PERF] Calculated location %d rock drying ONCE (took %v, will be reused for all %d areas)",
		locationID, time.Since(locationStart), len(areaIDs))

	// Get shared location data (also reused for all areas)
	sunExposure, err := s.repo.GetSunExposureByLocation(ctx, locationID)
	if err != nil {
		log.Printf("Warning: Failed to get sun exposure for location %d: %v", locationID, err)
		sunExposure = nil
	}
	locationTreeCoverage := 0.0
	if sunExposure != nil {
		locationTreeCoverage = sunExposure.TreeCoveragePercent
	}

	// hourlyForecast already obtained from getLocationRockDryingStatus (fresh API data)
	log.Printf("[PERF] Using fresh forecast from API (%d hours)", len(hourlyForecast))

	// Fetch ALL routes for ALL areas in a single batch
	allRouteIDs := []string{}
	for _, areaID := range areaIDs {
		routes, err := s.repo.GetRoutesWithGPSByArea(ctx, areaID)
		if err != nil {
			log.Printf("Warning: Failed to get routes for area %s: %v", areaID, err)
			continue
		}
		for _, route := range routes {
			allRouteIDs = append(allRouteIDs, route.MPRouteID)
		}
	}

	log.Printf("[PERF] Fetched %d total routes across %d areas", len(allRouteIDs), len(areaIDs))

	// Batch fetch all route details and profiles
	routesMap, err := s.repo.GetMPRoutesByIDs(ctx, allRouteIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch routes: %w", err)
	}

	profilesMap, err := s.repo.GetBoulderDryingProfilesByRouteIDs(ctx, allRouteIDs)
	if err != nil {
		log.Printf("Warning: Failed to fetch boulder drying profiles: %v", err)
		profilesMap = make(map[string]*models.BoulderDryingProfile)
	}

	// Calculate boulder drying status for ALL routes using the SHARED location data
	calcStart := time.Now()
	statusMap := make(map[string]*boulder_drying.BoulderDryingStatus)
	for _, routeID := range allRouteIDs {
		route, found := routesMap[routeID]
		if !found {
			continue
		}

		profile := profilesMap[routeID]
		status, err := s.calculator.CalculateBoulderDryingStatus(
			ctx,
			route,
			locationDrying, // REUSED for all routes
			profile,
			locationTreeCoverage, // REUSED for all routes
			hourlyForecast,       // REUSED for all routes
		)
		if err != nil {
			log.Printf("Warning: Failed to calculate status for route %s: %v", routeID, err)
			continue
		}
		statusMap[routeID] = status
	}
	log.Printf("[PERF] Calculated %d boulder statuses using SHARED location data (took %v)",
		len(statusMap), time.Since(calcStart))

	// Group routes by area and compute stats
	results := make(map[string]*models.AreaDryingStats)
	for _, areaID := range areaIDs {
		routes, err := s.repo.GetRoutesWithGPSByArea(ctx, areaID)
		if err != nil || len(routes) == 0 {
			continue
		}

		// Compute area-level stats from boulder statuses
		stats := &models.AreaDryingStats{
			TotalRoutes: len(routes),
		}

		var totalHoursUntilDry, totalTreeCoverage float64
		var wetRouteCount, treeCoverageCount, totalConfidence int
		mostRecentRain := ""

		for _, route := range routes {
			status, found := statusMap[route.MPRouteID]
			if !found {
				continue
			}

			if status.IsWet {
				stats.WetCount++
				if status.HoursUntilDry < 999 {
					wetRouteCount++
					totalHoursUntilDry += status.HoursUntilDry
				}
			} else {
				stats.DryCount++
			}

			if status.HoursUntilDry > 0 && status.HoursUntilDry < 999 {
				stats.DryingCount++
			}

			if status.TreeCoveragePercent > 0 {
				treeCoverageCount++
				totalTreeCoverage += status.TreeCoveragePercent
			}

			totalConfidence += status.ConfidenceScore

			if status.LastRainTimestamp != nil {
				rainStr := status.LastRainTimestamp.Format(time.RFC3339)
				if mostRecentRain == "" || rainStr > mostRecentRain {
					mostRecentRain = rainStr
				}
			}
		}

		if stats.TotalRoutes > 0 {
			stats.PercentDry = (float64(stats.DryCount) / float64(stats.TotalRoutes)) * 100
		}

		if wetRouteCount > 0 {
			stats.AvgHoursUntilDry = totalHoursUntilDry / float64(wetRouteCount)
		}

		if treeCoverageCount > 0 {
			stats.AvgTreeCoverage = totalTreeCoverage / float64(treeCoverageCount)
		}

		if stats.TotalRoutes > 0 {
			stats.ConfidenceScore = totalConfidence / stats.TotalRoutes
		}

		if mostRecentRain != "" {
			if parsed, err := time.Parse(time.RFC3339, mostRecentRain); err == nil {
				stats.LastRainTimestamp = &parsed
			}
		}

		results[areaID] = stats
	}

	log.Printf("[PERF] GetBatchAreaDryingStats: TOTAL TIME %v for %d areas (%d routes total, %d with stats) - OPTIMIZED",
		time.Since(batchStart), len(areaIDs), len(allRouteIDs), len(results))

	return results, nil
}

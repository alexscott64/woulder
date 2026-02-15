package service

import (
	"context"
	"fmt"
	"html"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alexscott64/woulder/backend/internal/database/climbing"
	"github.com/alexscott64/woulder/backend/internal/database/mountainproject"
	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/alexscott64/woulder/backend/internal/monitoring"
	mpClient "github.com/alexscott64/woulder/backend/internal/mountainproject"
	"github.com/alexscott64/woulder/backend/internal/weather/boulder_drying"
)

// MPClientInterface defines the interface for Mountain Project API operations
type MPClientInterface interface {
	GetRouteTicks(routeID string) ([]mpClient.Tick, error)
	GetArea(areaID string) (*mpClient.AreaResponse, error)
	GetAreaComments(areaID string) ([]mpClient.Comment, error)
	GetRouteComments(routeID string) ([]mpClient.Comment, error)
}

// Ensure real Client implements the interface
var _ MPClientInterface = (*mpClient.Client)(nil)

// ClimbTrackingService handles Mountain Project data synchronization and retrieval
type ClimbTrackingService struct {
	mountainProjectRepo mountainproject.Repository
	climbingRepo        climbing.Repository
	mpClient            MPClientInterface
	jobMonitor          *monitoring.JobMonitor
	syncMutex           sync.Mutex
	lastSyncTime        time.Time
	isSyncing           bool
}

// NewClimbTrackingService creates a new climb tracking service
func NewClimbTrackingService(
	mountainProjectRepo mountainproject.Repository,
	climbingRepo climbing.Repository,
	mpClient MPClientInterface,
	jobMonitor *monitoring.JobMonitor,
) *ClimbTrackingService {
	return &ClimbTrackingService{
		mountainProjectRepo: mountainProjectRepo,
		climbingRepo:        climbingRepo,
		mpClient:            mpClient,
		jobMonitor:          jobMonitor,
	}
}

// areaQueueItem represents an area to be processed with its associated location
type areaQueueItem struct {
	mpAreaID   string
	locationID *int
	parentID   *string
}

// cleanCommentText cleans and decodes HTML entities from Mountain Project comment text
// Removes common prefixes like "&middot;" and other HTML entities
func cleanCommentText(text string) string {
	if text == "" {
		return ""
	}

	// Decode HTML entities (e.g., &middot; -> ·, &amp; -> &)
	decoded := html.UnescapeString(text)

	// Remove leading middot character (both HTML entity and actual character)
	decoded = strings.TrimPrefix(decoded, "·")
	decoded = strings.TrimPrefix(decoded, "&middot;")

	// Remove leading/trailing whitespace
	decoded = strings.TrimSpace(decoded)

	return decoded
}

// isTickDateValid checks if a tick date is reasonable (not in the future)
// Mountain Project sometimes has data quality issues with future dates
func isTickDateValid(tickDate time.Time) bool {
	// Allow a small buffer for clock skew (24 hours), but reject anything beyond that
	maxFutureTime := time.Now().Add(24 * time.Hour)
	return tickDate.Before(maxFutureTime) || tickDate.Equal(maxFutureTime)
}

// SyncAreaRecursive performs breadth-first traversal of Mountain Project areas/routes
// and syncs all data to the database with rate limiting
func (s *ClimbTrackingService) SyncAreaRecursive(
	ctx context.Context,
	rootAreaID string,
	locationID *int,
) error {
	s.syncMutex.Lock()
	if s.isSyncing {
		s.syncMutex.Unlock()
		return fmt.Errorf("sync already in progress")
	}
	s.isSyncing = true
	s.syncMutex.Unlock()

	defer func() {
		s.syncMutex.Lock()
		s.isSyncing = false
		s.lastSyncTime = time.Now()
		s.syncMutex.Unlock()
	}()

	// Initialize breadth-first queue
	queue := []areaQueueItem{{
		mpAreaID:   rootAreaID,
		locationID: locationID,
		parentID:   nil,
	}}

	processedAreas := make(map[string]bool)
	routeCount := 0
	areaCount := 0

	for len(queue) > 0 {
		// Pop from queue
		item := queue[0]
		queue = queue[1:]

		// Skip if already processed
		if processedAreas[item.mpAreaID] {
			continue
		}

		// Check context for cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Fetch area data from Mountain Project
		log.Printf("Fetching area: %s", item.mpAreaID)
		areaData, err := s.mpClient.GetArea(item.mpAreaID)
		if err != nil {
			log.Printf("Error fetching area %s: %v", item.mpAreaID, err)
			continue
		}

		// Convert area ID to string for API interactions
		areaIDStr := strconv.Itoa(areaData.ID)

		// Convert parent ID string to int64 if exists
		var parentIDInt64 *int64
		if item.parentID != nil {
			if parentID, err := strconv.ParseInt(*item.parentID, 10, 64); err == nil {
				parentIDInt64 = &parentID
			}
		}

		// Save area to database
		area := &models.MPArea{
			MPAreaID:       int64(areaData.ID),
			Name:           areaData.Title,
			ParentMPAreaID: parentIDInt64,
			AreaType:       areaData.Type,
			LocationID:     item.locationID,
		}

		// Extract GPS coordinates if available
		if len(areaData.Coordinates) == 2 {
			longitude := areaData.Coordinates[0]
			latitude := areaData.Coordinates[1]
			area.Longitude = &longitude
			area.Latitude = &latitude
			log.Printf("Area GPS: %.4f, %.4f", latitude, longitude)
		}

		if err := s.mountainProjectRepo.Areas().SaveArea(ctx, area); err != nil {
			log.Printf("Error saving area %s: %v", item.mpAreaID, err)
			continue
		}

		processedAreas[item.mpAreaID] = true
		areaCount++
		log.Printf("Saved area: %s (%s) - Total areas: %d", areaData.Title, areaIDStr, areaCount)

		// Sync area comments
		if err := s.syncAreaComments(ctx, areaIDStr); err != nil {
			log.Printf("Warning: failed to sync comments for area %s: %v", areaIDStr, err)
			// Continue processing even if comment sync fails
		}

		// Collect boulder routes for GPS distribution
		var boulderRoutes []int64 // Store route IDs for GPS calculation

		// Process children
		for _, child := range areaData.Children {
			childIDStr := strconv.Itoa(child.ID)

			if child.Type == "Area" {
				// Add subarea to queue for processing
				queue = append(queue, areaQueueItem{
					mpAreaID:   childIDStr,
					locationID: item.locationID,
					parentID:   &item.mpAreaID,
				})
			} else {
				// It's a route - check if it's a boulder route (for GPS calculation only)
				isBoulder := false
				for _, rt := range child.RouteTypes {
					if strings.ToLower(rt) == "boulder" {
						isBoulder = true
						break
					}
				}

				// Convert area ID string to int64
				areaIDInt64, err := strconv.ParseInt(item.mpAreaID, 10, 64)
				if err != nil {
					log.Printf("Error parsing area ID %s: %v", item.mpAreaID, err)
					continue
				}

				// Save route (all types: boulder, sport, trad, etc.)
				route := &models.MPRoute{
					MPRouteID:  int64(child.ID),
					MPAreaID:   areaIDInt64,
					Name:       child.Title,
					RouteType:  strings.Join(child.RouteTypes, ", "),
					Rating:     "", // Rating not in children response, could fetch separately if needed
					LocationID: item.locationID,
				}

				if err := s.mountainProjectRepo.Routes().SaveRoute(ctx, route); err != nil {
					log.Printf("Error saving route %s: %v", childIDStr, err)
					continue
				}

				// Add to boulder routes list for GPS calculation (only boulders need GPS distribution)
				if isBoulder {
					boulderRoutes = append(boulderRoutes, int64(child.ID))
				}

				routeCount++
				log.Printf("Saved route: %s (%s) - Total routes: %d", child.Title, childIDStr, routeCount)

				// Fetch and save ticks for this route
				if err := s.syncRouteTicks(ctx, childIDStr); err != nil {
					log.Printf("Error syncing ticks for route %s: %v", childIDStr, err)
					// Continue processing other routes even if tick sync fails
				}

				// Fetch and save comments for this route
				if err := s.syncRouteComments(ctx, childIDStr); err != nil {
					log.Printf("Warning: failed to sync comments for route %s: %v", childIDStr, err)
					// Continue processing other routes even if comment sync fails
				}
			}
		}

		// Calculate and update GPS positions for all boulders in this area
		if len(boulderRoutes) > 0 && area.Latitude != nil && area.Longitude != nil {
			if err := s.calculateBoulderGPS(ctx, boulderRoutes, *area.Latitude, *area.Longitude); err != nil {
				log.Printf("Error calculating boulder GPS for area %s: %v", item.mpAreaID, err)
				// Continue processing even if GPS calculation fails
			}
		}
	}

	log.Printf("Sync complete for area %s: %d areas, %d routes processed", rootAreaID, areaCount, routeCount)
	return nil
}

// syncRouteTicks fetches and saves all ticks for a given route
func (s *ClimbTrackingService) syncRouteTicks(ctx context.Context, routeID string) error {
	// Convert route ID string to int64
	routeIDInt64, err := strconv.ParseInt(routeID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid route ID %s: %w", routeID, err)
	}

	ticks, tickErr := s.mpClient.GetRouteTicks(routeID)
	if tickErr != nil {
		return fmt.Errorf("failed to fetch ticks: %w", tickErr)
	}

	// Load Pacific timezone (America/Los_Angeles) for Mountain Project dates
	// Mountain Project returns dates in local time (typically Pacific for US climbing areas)
	pacificTZ, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		log.Printf("Warning: failed to load Pacific timezone, falling back to UTC: %v", err)
		pacificTZ = time.UTC
	}

	tickCount := 0
	for _, tick := range ticks {
		// Parse the date - Mountain Project API returns multiple formats
		var climbedAt time.Time
		var err error

		// Format 1: "Jan 2, 2006, 3:04 pm" - Most common format
		climbedAt, err = time.ParseInLocation("Jan 2, 2006, 3:04 pm", tick.Date, pacificTZ)
		if err != nil {
			// Format 2: "2006-01-02 15:04:05"
			climbedAt, err = time.ParseInLocation("2006-01-02 15:04:05", tick.Date, pacificTZ)
			if err != nil {
				// Format 3: "2006-01-02" - Date only, use noon Pacific as default time
				climbedAt, err = time.ParseInLocation("2006-01-02", tick.Date, pacificTZ)
				if err != nil {
					log.Printf("Warning: invalid date format for tick on route %s: %s", routeID, tick.Date)
					continue
				}
			}
		}

		// Skip ticks with future dates (data quality issue)
		if !isTickDateValid(climbedAt) {
			log.Printf("Warning: skipping tick with future date on route %s: %s", routeID, tick.Date)
			continue
		}

		tickModel := &models.MPTick{
			MPRouteID: routeIDInt64,
			UserName:  tick.GetUserName(),
			ClimbedAt: climbedAt,
			Style:     tick.Style,
		}

		// Use the 'text' field as the comment, clean HTML entities
		textStr := tick.GetTextString()
		if textStr != "" {
			cleanedText := cleanCommentText(textStr)
			if cleanedText != "" {
				tickModel.Comment = &cleanedText
			}
		}

		if err := s.mountainProjectRepo.Ticks().SaveTick(ctx, tickModel); err != nil {
			log.Printf("Error saving tick for route %s: %v", routeID, err)
			continue
		}

		tickCount++
	}

	if tickCount > 0 {
		log.Printf("Saved %d ticks for route %s", tickCount, routeID)
	}

	return nil
}

// syncAreaComments fetches and saves all comments for a given area
func (s *ClimbTrackingService) syncAreaComments(ctx context.Context, areaID string) error {
	comments, err := s.mpClient.GetAreaComments(areaID)
	if err != nil {
		return fmt.Errorf("failed to fetch comments: %w", err)
	}

	// Convert area ID string to int64
	areaIDInt64, err := strconv.ParseInt(areaID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid area ID %s: %w", areaID, err)
	}

	commentCount := 0
	for _, comment := range comments {
		commentedAt := time.Unix(comment.Created, 0)
		userName := comment.GetUserInfo()

		if err := s.mountainProjectRepo.Comments().SaveAreaComment(ctx, int64(comment.ID), areaIDInt64, userName, comment.Message, commentedAt); err != nil {
			log.Printf("Error saving area comment: %v", err)
			continue
		}
		commentCount++
	}

	if commentCount > 0 {
		log.Printf("Saved %d comments for area %s", commentCount, areaID)
	}

	return nil
}

// syncRouteComments fetches and saves all comments for a given route
func (s *ClimbTrackingService) syncRouteComments(ctx context.Context, routeID string) error {
	comments, err := s.mpClient.GetRouteComments(routeID)
	if err != nil {
		return fmt.Errorf("failed to fetch comments: %w", err)
	}

	// Convert route ID string to int64
	routeIDInt64, err := strconv.ParseInt(routeID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid route ID %s: %w", routeID, err)
	}

	commentCount := 0
	for _, comment := range comments {
		commentedAt := time.Unix(comment.Created, 0)
		userName := comment.GetUserInfo()

		if err := s.mountainProjectRepo.Comments().SaveRouteComment(ctx, int64(comment.ID), routeIDInt64, userName, comment.Message, commentedAt); err != nil {
			log.Printf("Error saving route comment: %v", err)
			continue
		}
		commentCount++
	}

	if commentCount > 0 {
		log.Printf("Saved %d comments for route %s", commentCount, routeID)
	}

	return nil
}

// GetLastClimbedForLocation retrieves the most recent climb info for a location
// DEPRECATED: Use GetClimbHistoryForLocation instead
func (s *ClimbTrackingService) GetLastClimbedForLocation(
	ctx context.Context,
	locationID int,
) (*models.LastClimbedInfo, error) {
	return s.climbingRepo.History().GetLastClimbedForLocation(ctx, locationID)
}

// GetClimbHistoryForLocation retrieves recent climb history for a location
func (s *ClimbTrackingService) GetClimbHistoryForLocation(
	ctx context.Context,
	locationID int,
	limit int,
) ([]models.ClimbHistoryEntry, error) {
	return s.climbingRepo.History().GetClimbHistoryForLocation(ctx, locationID, limit)
}

// GetAreasOrderedByActivity retrieves areas ordered by most recent climb activity
func (s *ClimbTrackingService) GetAreasOrderedByActivity(
	ctx context.Context,
	locationID int,
) ([]models.AreaActivitySummary, error) {
	return s.climbingRepo.Activity().GetAreasOrderedByActivity(ctx, locationID)
}

// GetSubareasOrderedByActivity retrieves subareas of a parent area ordered by recent climb activity
func (s *ClimbTrackingService) GetSubareasOrderedByActivity(
	ctx context.Context,
	parentAreaID int64,
	locationID int,
) ([]models.AreaActivitySummary, error) {
	return s.climbingRepo.Activity().GetSubareasOrderedByActivity(ctx, parentAreaID, locationID)
}

// GetRoutesOrderedByActivity retrieves routes in an area ordered by recent climb activity
func (s *ClimbTrackingService) GetRoutesOrderedByActivity(
	ctx context.Context,
	areaID int64,
	locationID int,
	limit int,
) ([]models.RouteActivitySummary, error) {
	return s.climbingRepo.Activity().GetRoutesOrderedByActivity(ctx, areaID, locationID, limit)
}

// GetRecentTicksForRoute retrieves recent ticks for a specific route
func (s *ClimbTrackingService) GetRecentTicksForRoute(
	ctx context.Context,
	routeID int64,
	limit int,
) ([]models.ClimbHistoryEntry, error) {
	return s.climbingRepo.Activity().GetRecentTicksForRoute(ctx, routeID, limit)
}

// SearchInLocation searches all areas and routes in a location by name
func (s *ClimbTrackingService) SearchInLocation(
	ctx context.Context,
	locationID int,
	searchQuery string,
	limit int,
) ([]models.SearchResult, error) {
	return s.climbingRepo.Search().SearchInLocation(ctx, locationID, searchQuery, limit)
}

// SearchRoutesInLocation searches all routes in a location by name or grade
func (s *ClimbTrackingService) SearchRoutesInLocation(
	ctx context.Context,
	locationID int,
	searchQuery string,
	limit int,
) ([]models.RouteActivitySummary, error) {
	return s.climbingRepo.Search().SearchRoutesInLocation(ctx, locationID, searchQuery, limit)
}

// GetSyncStatus returns the current sync status
func (s *ClimbTrackingService) GetSyncStatus() (isSyncing bool, lastSync time.Time) {
	s.syncMutex.Lock()
	defer s.syncMutex.Unlock()
	return s.isSyncing, s.lastSyncTime
}

// SyncNewTicksForLocation performs incremental sync of only new ticks for a location
// This is much more efficient than full sync as it only fetches ticks newer than the last known tick
func (s *ClimbTrackingService) SyncNewTicksForLocation(ctx context.Context, locationID int) error {
	s.syncMutex.Lock()
	if s.isSyncing {
		s.syncMutex.Unlock()
		return fmt.Errorf("sync already in progress")
	}
	s.isSyncing = true
	s.syncMutex.Unlock()

	defer func() {
		s.syncMutex.Lock()
		s.isSyncing = false
		s.lastSyncTime = time.Now()
		s.syncMutex.Unlock()
	}()

	// Get all route IDs for this location
	routeIDs, err := s.mountainProjectRepo.Routes().GetAllIDsForLocation(ctx, locationID)
	if err != nil {
		return fmt.Errorf("failed to get route IDs: %w", err)
	}

	if len(routeIDs) == 0 {
		log.Printf("No routes found for location %d, skipping tick sync", locationID)
		return nil
	}

	log.Printf("Starting incremental tick sync for location %d (%d routes)", locationID, len(routeIDs))

	// Load Pacific timezone for Mountain Project dates
	pacificTZ, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		log.Printf("Warning: failed to load Pacific timezone, falling back to UTC: %v", err)
		pacificTZ = time.UTC
	}

	totalNewTicks := 0
	routesProcessed := 0

	// For each route, sync only new ticks
	for _, routeID := range routeIDs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Get the timestamp of the last tick we have for this route
		lastTickTime, err := s.mountainProjectRepo.Ticks().GetLastTimestampForRoute(ctx, routeID)
		if err != nil {
			log.Printf("Error getting last tick for route %d: %v", routeID, err)
			continue
		}

		// Fetch ticks from Mountain Project (API requires string)
		routeIDStr := strconv.FormatInt(routeID, 10)
		ticks, err := s.mpClient.GetRouteTicks(routeIDStr)
		if err != nil {
			log.Printf("Error fetching ticks for route %d: %v", routeID, err)
			continue
		}

		newTickCount := 0
		for _, tick := range ticks {
			// Parse the date in Pacific timezone
			var climbedAt time.Time
			climbedAt, err = time.ParseInLocation("Jan 2, 2006, 3:04 pm", tick.Date, pacificTZ)
			if err != nil {
				climbedAt, err = time.ParseInLocation("2006-01-02 15:04:05", tick.Date, pacificTZ)
				if err != nil {
					climbedAt, err = time.ParseInLocation("2006-01-02", tick.Date, pacificTZ)
					if err != nil {
						log.Printf("Warning: invalid date format for tick on route %d: %s", routeID, tick.Date)
						continue
					}
				}
			}

			// Skip ticks with future dates (data quality issue)
			if !isTickDateValid(climbedAt) {
				log.Printf("Warning: skipping tick with future date on route %d: %s", routeID, tick.Date)
				continue
			}

			// Skip if we already have this tick (incremental check)
			if lastTickTime != nil && !climbedAt.After(*lastTickTime) {
				continue // Already have this tick or older
			}

			tickModel := &models.MPTick{
				MPRouteID: routeID,
				UserName:  tick.GetUserName(),
				ClimbedAt: climbedAt,
				Style:     tick.Style,
			}

			textStr := tick.GetTextString()
			if textStr != "" {
				cleanedText := cleanCommentText(textStr)
				if cleanedText != "" {
					tickModel.Comment = &cleanedText
				}
			}

			if err := s.mountainProjectRepo.Ticks().SaveTick(ctx, tickModel); err != nil {
				log.Printf("Error saving tick for route %d: %v", routeID, err)
				continue
			}

			newTickCount++
		}

		if newTickCount > 0 {
			totalNewTicks += newTickCount
			// Verbose per-route logging disabled - see summary log below
			// log.Printf("Route %d: %d new ticks", routeID, newTickCount)
		}

		routesProcessed++
	}

	log.Printf("Incremental sync complete for location %d: %d new ticks across %d routes",
		locationID, totalNewTicks, routesProcessed)

	return nil
}

// SyncNewTicksForAllLocations performs incremental sync for all locations with MP data
func (s *ClimbTrackingService) SyncNewTicksForAllLocations(ctx context.Context) error {
	// Hardcoded location IDs that have Mountain Project data
	// These match the locations in cmd/sync_climbs/main.go
	locationIDs := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}

	log.Printf("Starting incremental tick sync for all locations (%d total)", len(locationIDs))

	successCount := 0
	failCount := 0

	for _, locationID := range locationIDs {
		log.Printf("Syncing location %d...", locationID)
		if err := s.SyncNewTicksForLocation(ctx, locationID); err != nil {
			log.Printf("Error syncing location %d: %v", locationID, err)
			failCount++
		} else {
			successCount++
		}
	}

	log.Printf("All locations sync complete: %d succeeded, %d failed", successCount, failCount)

	if failCount > 0 {
		return fmt.Errorf("sync completed with %d failures", failCount)
	}

	return nil
}

// calculateBoulderGPS calculates GPS positions for boulders using circular distribution
// and updates the database with calculated coordinates and aspects
func (s *ClimbTrackingService) calculateBoulderGPS(
	ctx context.Context,
	routeIDs []int64,
	centerLat, centerLon float64,
) error {
	if len(routeIDs) == 0 {
		return nil
	}

	// Calculate radius based on number of boulders
	radius := boulder_drying.CalculateRadiusForArea(len(routeIDs))

	// Calculate positions for all boulders
	positions := boulder_drying.CalculateBoulderPositions(centerLat, centerLon, len(routeIDs), radius)

	log.Printf("Calculating GPS for %d boulders (radius: %.4f degrees)", len(routeIDs), radius)

	// Update each route with its calculated GPS position and aspect
	updateCount := 0
	for i, routeID := range routeIDs {
		pos := positions[i]

		err := s.mountainProjectRepo.Routes().UpdateGPS(ctx, routeID, pos.Latitude, pos.Longitude, pos.Aspect)
		if err != nil {
			log.Printf("Error updating GPS for route %d: %v", routeID, err)
			continue
		}
		updateCount++
	}

	log.Printf("Successfully updated GPS for %d/%d boulders", updateCount, len(routeIDs))
	return nil
}

// SyncNewRoutesForAllStates checks all root areas for new routes and syncs them
// Uses smart binary-search traversal to minimize API calls
func (s *ClimbTrackingService) SyncNewRoutesForAllStates(ctx context.Context) error {
	log.Println("Starting new route sync for all states...")

	// Get all state configurations
	states, err := s.mountainProjectRepo.Areas().GetAllStateConfigs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get state configs: %w", err)
	}

	successCount := 0
	failCount := 0
	totalNewRoutes := 0

	for _, state := range states {
		log.Printf("Checking state: %s (MP Area ID: %s)", state.StateName, state.MPAreaID)

		newRoutes, err := s.checkAreaForNewRoutes(ctx, state.MPAreaID)
		if err != nil {
			log.Printf("Error checking %s: %v", state.StateName, err)
			failCount++
			continue
		}

		if newRoutes > 0 {
			log.Printf("✓ %s: Found and synced %d new route(s)", state.StateName, newRoutes)
			totalNewRoutes += newRoutes
		}

		successCount++
	}

	log.Printf("New route sync complete: %d states checked, %d new routes found, %d failures", successCount, totalNewRoutes, failCount)

	if failCount > 0 {
		return fmt.Errorf("sync completed with %d failures", failCount)
	}

	return nil
}

// checkAreaForNewRoutes recursively checks an area and its children for new routes
// Returns the number of new routes found and synced
func (s *ClimbTrackingService) checkAreaForNewRoutes(ctx context.Context, areaID string) (int, error) {
	// Check context for cancellation
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	// Fetch current area data from Mountain Project API
	areaResp, err := s.mpClient.GetArea(areaID)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch area %s: %w", areaID, err)
	}

	// Get cached route count from database
	cachedCount, err := s.mountainProjectRepo.Areas().GetRouteCount(ctx, areaID)
	if err != nil {
		return 0, fmt.Errorf("failed to get cached count for area %s: %w", areaID, err)
	}

	// Get current count from API response
	currentCount := 0
	if areaResp.RouteTypeCounts != nil {
		currentCount = areaResp.RouteTypeCounts.Total
	}

	// If first time checking or counts match, update cache and return
	if cachedCount == -1 {
		// First time checking this area - just cache the count
		if err := s.mountainProjectRepo.Areas().UpdateRouteCount(ctx, areaID, currentCount); err != nil {
			log.Printf("Warning: failed to cache count for area %s: %v", areaID, err)
		}
		return 0, nil
	}

	if cachedCount == currentCount {
		// No change detected
		if err := s.mountainProjectRepo.Areas().UpdateRouteCount(ctx, areaID, currentCount); err != nil {
			log.Printf("Warning: failed to update check time for area %s: %v", areaID, err)
		}
		return 0, nil
	}

	// Change detected!
	if currentCount < cachedCount {
		// Route count decreased (route deleted) - just update cache
		log.Printf("Area %s (%s): route count decreased from %d to %d (route deleted)", areaID, areaResp.Title, cachedCount, currentCount)
		if err := s.mountainProjectRepo.Areas().UpdateRouteCount(ctx, areaID, currentCount); err != nil {
			log.Printf("Warning: failed to update count for area %s: %v", areaID, err)
		}
		return 0, nil
	}

	// Route count increased - find and sync new routes
	newRouteCount := currentCount - cachedCount
	log.Printf("Area %s (%s): route count increased from %d to %d (+%d new route(s))", areaID, areaResp.Title, cachedCount, currentCount, newRouteCount)

	// Update cached count
	if err := s.mountainProjectRepo.Areas().UpdateRouteCount(ctx, areaID, currentCount); err != nil {
		log.Printf("Warning: failed to update count for area %s: %v", areaID, err)
	}

	// Check if this area has children (subareas)
	hasSubareas := false
	hasRoutes := false
	for _, child := range areaResp.Children {
		if child.Type == "Area" {
			hasSubareas = true
		} else if child.Type == "Route" {
			hasRoutes = true
		}
	}

	// If this area has subareas, recursively check them to find which one changed
	if hasSubareas {
		totalNewRoutes := 0
		for _, child := range areaResp.Children {
			if child.Type == "Area" {
				childID := strconv.Itoa(child.ID)
				newRoutes, err := s.checkAreaForNewRoutes(ctx, childID)
				if err != nil {
					log.Printf("Error checking child area %s: %v", childID, err)
					continue
				}
				totalNewRoutes += newRoutes
			}
		}
		return totalNewRoutes, nil
	}

	// This is a leaf area with routes - sync the new ones
	if hasRoutes {
		return s.syncNewRoutesInArea(ctx, areaID, areaResp)
	}

	// No subareas and no routes - shouldn't happen but handle gracefully
	return 0, nil
}

// syncNewRoutesInArea syncs new routes in a specific area
func (s *ClimbTrackingService) syncNewRoutesInArea(ctx context.Context, areaID string, areaResp *mpClient.AreaResponse) (int, error) {
	// Get existing route IDs from database
	existingRouteIDs, err := s.mountainProjectRepo.Routes().GetIDsForArea(ctx, areaID)
	if err != nil {
		return 0, fmt.Errorf("failed to get existing routes for area %s: %w", areaID, err)
	}

	// Create a map for quick lookup
	existingRoutes := make(map[string]bool)
	for _, routeID := range existingRouteIDs {
		existingRoutes[routeID] = true
	}

	// Find new routes
	var newRoutes []mpClient.ChildElement
	for _, child := range areaResp.Children {
		if child.Type == "Route" {
			routeID := strconv.Itoa(child.ID)
			if !existingRoutes[routeID] {
				newRoutes = append(newRoutes, child)
			}
		}
	}

	if len(newRoutes) == 0 {
		log.Printf("No new routes found in area %s (expected based on count change)", areaID)
		return 0, nil
	}

	log.Printf("Syncing %d new route(s) in area %s (%s)", len(newRoutes), areaID, areaResp.Title)

	// Convert area ID to int64 for database operations
	areaIDInt64, err := strconv.ParseInt(areaID, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse area ID %s: %w", areaID, err)
	}

	// Get GPS coordinates from area
	var lat, lon *float64
	if len(areaResp.Coordinates) >= 2 {
		latitude := areaResp.Coordinates[1] // [longitude, latitude]
		longitude := areaResp.Coordinates[0]
		lat = &latitude
		lon = &longitude
	}

	// Sync each new route
	syncedCount := 0
	for _, route := range newRoutes {
		routeID := strconv.Itoa(route.ID)
		routeIDInt64 := int64(route.ID)

		// Determine route type
		routeType := ""
		if len(route.RouteTypes) > 0 {
			routeType = route.RouteTypes[0]
		}

		// Insert route
		err := s.mountainProjectRepo.Routes().UpsertRoute(ctx, routeIDInt64, areaIDInt64, nil, route.Title, routeType, "", lat, lon, nil)
		if err != nil {
			log.Printf("Error syncing route %s: %v", routeID, err)
			continue
		}

		// Fetch and sync ticks
		ticks, err := s.mpClient.GetRouteTicks(routeID)
		if err != nil {
			log.Printf("Warning: failed to fetch ticks for route %s: %v", routeID, err)
		} else {
			for _, tick := range ticks {
				// Parse tick date
				tickDate, err := time.Parse("Jan 2, 2006, 3:04 pm", tick.Date)
				if err != nil {
					log.Printf("Warning: failed to parse tick date '%s': %v", tick.Date, err)
					continue
				}

				// Skip invalid dates
				if !isTickDateValid(tickDate) {
					continue
				}

				// Get user name
				userName := tick.GetUserName()
				if userName == "" {
					userName = "Anonymous"
				}

				// Get comment
				var comment *string
				commentText := tick.GetTextString()
				if commentText != "" {
					cleaned := cleanCommentText(commentText)
					comment = &cleaned
				}

				// Insert tick
				if err := s.mountainProjectRepo.Ticks().UpsertTick(ctx, routeIDInt64, userName, tickDate, tick.Style, comment); err != nil {
					log.Printf("Warning: failed to insert tick for route %s: %v", routeID, err)
				}
			}
		}

		// Fetch and sync comments
		comments, err := s.mpClient.GetRouteComments(routeID)
		if err != nil {
			log.Printf("Warning: failed to fetch comments for route %s: %v", routeID, err)
		} else {
			for _, comment := range comments {
				commentTime := time.Unix(comment.Created, 0)
				userName := comment.GetUserInfo()
				cleanedText := cleanCommentText(comment.Message)

				commentID := int64(comment.ID)
				if err := s.mountainProjectRepo.Comments().UpsertRouteComment(ctx, commentID, routeIDInt64, userName, nil, cleanedText, commentTime); err != nil {
					log.Printf("Warning: failed to insert comment for route %s: %v", routeID, err)
				}
			}
		}

		syncedCount++
	}

	log.Printf("Successfully synced %d/%d new routes in area %s", syncedCount, len(newRoutes), areaID)
	return syncedCount, nil
}

// RecalculateAllPriorities recalculates sync priorities for all non-location routes
// Only applies to routes without location_id (location routes always sync daily)
func (s *ClimbTrackingService) RecalculateAllPriorities(ctx context.Context) error {
	log.Println("Recalculating route sync priorities (non-location routes only)...")

	if err := s.mountainProjectRepo.Sync().UpdateRoutePriorities(ctx); err != nil {
		return fmt.Errorf("failed to recalculate priorities: %w", err)
	}

	// Query and log priority distribution
	distribution, err := s.mountainProjectRepo.Sync().GetPriorityDistribution(ctx)
	if err != nil {
		log.Printf("Warning: failed to get priority distribution: %v", err)
	} else {
		log.Printf("Priority recalculation complete: high=%d, medium=%d, low=%d [non-location routes]",
			distribution["high"], distribution["medium"], distribution["low"])
	}

	return nil
}

// SyncLocationRouteTicks syncs ticks for ALL routes with location_id (woulder locations - always daily)
func (s *ClimbTrackingService) SyncLocationRouteTicks(ctx context.Context) error {
	startTime := time.Now()

	// Get ALL location routes due for tick sync
	routeIDs, err := s.mountainProjectRepo.Sync().GetLocationRoutesDueForSync(ctx, "ticks")
	if err != nil {
		return fmt.Errorf("failed to get location routes for tick sync: %w", err)
	}

	if len(routeIDs) == 0 {
		log.Println("No location routes due for tick sync")
		return nil
	}

	log.Printf("Syncing ticks for %d location routes...", len(routeIDs))

	// Load Pacific timezone for Mountain Project dates
	pacificTZ, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		log.Printf("Warning: failed to load Pacific timezone, falling back to UTC: %v", err)
		pacificTZ = time.UTC
	}

	totalNewTicks := 0

	// Sync ticks for each route with rate limiting
	err = s.rateLimitedSync(ctx, routeIDs, func(routeID string) error {
		// Get last tick timestamp for this route
		routeIDInt64, _ := strconv.ParseInt(routeID, 10, 64)
		lastTickTime, err := s.mountainProjectRepo.Ticks().GetLastTimestampForRoute(ctx, routeIDInt64)
		if err != nil {
			return fmt.Errorf("failed to get last tick for route %s: %w", routeID, err)
		}

		// Fetch ticks from MP API
		ticks, err := s.mpClient.GetRouteTicks(routeID)
		if err != nil {
			return fmt.Errorf("failed to fetch ticks: %w", err)
		}

		// Process only new ticks
		newTickCount := 0
		for _, tick := range ticks {
			// Parse tick date
			var climbedAt time.Time
			climbedAt, err = time.ParseInLocation("Jan 2, 2006, 3:04 pm", tick.Date, pacificTZ)
			if err != nil {
				climbedAt, err = time.ParseInLocation("2006-01-02 15:04:05", tick.Date, pacificTZ)
				if err != nil {
					climbedAt, err = time.ParseInLocation("2006-01-02", tick.Date, pacificTZ)
					if err != nil {
						continue
					}
				}
			}

			// Skip invalid dates
			if !isTickDateValid(climbedAt) {
				continue
			}

			// Skip if we already have this tick
			if lastTickTime != nil && !climbedAt.After(*lastTickTime) {
				continue
			}

			// Save new tick
			tickModel := &models.MPTick{
				MPRouteID: routeIDInt64,
				UserName:  tick.GetUserName(),
				ClimbedAt: climbedAt,
				Style:     tick.Style,
			}

			textStr := tick.GetTextString()
			if textStr != "" {
				cleanedText := cleanCommentText(textStr)
				if cleanedText != "" {
					tickModel.Comment = &cleanedText
				}
			}

			if err := s.mountainProjectRepo.Ticks().SaveTick(ctx, tickModel); err != nil {
				log.Printf("Error saving tick for route %s: %v", routeID, err)
				continue
			}

			newTickCount++
		}

		totalNewTicks += newTickCount
		return nil
	})

	duration := time.Since(startTime)
	if err != nil {
		return fmt.Errorf("location route tick sync error: %w", err)
	}

	log.Printf("Location route tick sync complete: %d routes, %d new ticks, %.1f minutes",
		len(routeIDs), totalNewTicks, duration.Minutes())

	return nil
}

// SyncLocationRouteComments syncs comments for ALL routes with location_id (woulder locations - always daily)
func (s *ClimbTrackingService) SyncLocationRouteComments(ctx context.Context) error {
	startTime := time.Now()

	// Get ALL location routes due for comment sync
	routeIDs, err := s.mountainProjectRepo.Sync().GetLocationRoutesDueForSync(ctx, "comments")
	if err != nil {
		return fmt.Errorf("failed to get location routes for comment sync: %w", err)
	}

	if len(routeIDs) == 0 {
		log.Println("No location routes due for comment sync")
		return nil
	}

	log.Printf("Syncing comments for %d location routes...", len(routeIDs))

	totalNewComments := 0

	// Sync comments for each route with rate limiting
	err = s.rateLimitedSync(ctx, routeIDs, func(routeID string) error {
		routeIDInt64, _ := strconv.ParseInt(routeID, 10, 64)

		// Fetch comments from MP API
		comments, err := s.mpClient.GetRouteComments(routeID)
		if err != nil {
			return fmt.Errorf("failed to fetch comments: %w", err)
		}

		// Save comments (upsert handles duplicates)
		newCommentCount := 0
		for _, comment := range comments {
			commentedAt := time.Unix(comment.Created, 0)
			userName := comment.GetUserInfo()
			cleanedText := cleanCommentText(comment.Message)

			if err := s.mountainProjectRepo.Comments().SaveRouteComment(ctx, int64(comment.ID), routeIDInt64, userName, cleanedText, commentedAt); err != nil {
				log.Printf("Error saving comment for route %s: %v", routeID, err)
				continue
			}

			newCommentCount++
		}

		totalNewComments += newCommentCount
		return nil
	})

	duration := time.Since(startTime)
	if err != nil {
		return fmt.Errorf("location route comment sync error: %w", err)
	}

	log.Printf("Location route comment sync complete: %d routes, %d new comments, %.1f minutes",
		len(routeIDs), totalNewComments, duration.Minutes())

	return nil
}

// SyncTicksByPriority syncs ticks for non-location routes at a specific priority tier
func (s *ClimbTrackingService) SyncTicksByPriority(ctx context.Context, priority string) error {
	startTime := time.Now()

	// Get non-location routes due for tick sync at this priority
	routeIDs, err := s.mountainProjectRepo.Sync().GetRoutesDueForTickSync(ctx, priority)
	if err != nil {
		return fmt.Errorf("failed to get routes for priority %s tick sync: %w", priority, err)
	}

	if len(routeIDs) == 0 {
		log.Printf("No %s priority routes due for tick sync", priority)
		return nil
	}

	// START MONITORING: Create job execution record
	jobName := fmt.Sprintf("%s_priority_tick_sync", priority)
	jobExec, err := s.jobMonitor.StartJob(ctx, jobName, "tick_sync", len(routeIDs), map[string]interface{}{
		"priority": priority,
	})
	if err != nil {
		log.Printf("Warning: failed to start job monitoring: %v", err)
		jobExec = nil // Continue without monitoring
	}

	// Create progress reporter (updates DB every 10 items or every 5 seconds)
	var reporter *monitoring.ProgressReporter
	if jobExec != nil {
		reporter = monitoring.NewProgressReporter(s.jobMonitor, jobExec.ID, len(routeIDs), 10)
		defer func() {
			// Final flush to ensure last progress is saved
			reporter.FlushProgress(ctx)
		}()
	}

	log.Printf("Syncing ticks for %d %s priority routes...", len(routeIDs), priority)

	// Load Pacific timezone for Mountain Project dates
	pacificTZ, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		log.Printf("Warning: failed to load Pacific timezone, falling back to UTC: %v", err)
		pacificTZ = time.UTC
	}

	totalNewTicks := 0

	// Sync ticks for each route with rate limiting
	err = s.rateLimitedSync(ctx, routeIDs, func(routeID string) error {
		// Get last tick timestamp for this route
		routeIDInt64, _ := strconv.ParseInt(routeID, 10, 64)

		// Get route name for tracking
		if jobExec != nil {
			route, routeErr := s.mountainProjectRepo.Routes().GetByID(ctx, routeIDInt64)
			if routeErr == nil && route != nil {
				// Update current route being synced
				s.jobMonitor.UpdateCurrentItem(ctx, jobExec.ID, map[string]interface{}{
					"current_route_id":   routeIDInt64,
					"current_route_name": route.Name,
					"current_route_type": route.RouteType,
					"current_rating":     route.Rating,
				})
			}
		}

		lastTickTime, tickErr := s.mountainProjectRepo.Ticks().GetLastTimestampForRoute(ctx, routeIDInt64)
		if tickErr != nil {
			// Report progress
			if reporter != nil {
				reporter.Increment(ctx, false)
			}
			return fmt.Errorf("failed to get last tick for route %s: %w", routeID, tickErr)
		}

		// Fetch ticks from MP API
		ticks, tickErr := s.mpClient.GetRouteTicks(routeID)
		if tickErr != nil {
			// Report progress
			if reporter != nil {
				reporter.Increment(ctx, false)
			}
			return fmt.Errorf("failed to fetch ticks: %w", tickErr)
		}

		// Process only new ticks
		newTickCount := 0
		for _, tick := range ticks {
			// Parse tick date
			var climbedAt time.Time
			climbedAt, tickErr = time.ParseInLocation("Jan 2, 2006, 3:04 pm", tick.Date, pacificTZ)
			if tickErr != nil {
				climbedAt, tickErr = time.ParseInLocation("2006-01-02 15:04:05", tick.Date, pacificTZ)
				if tickErr != nil {
					climbedAt, tickErr = time.ParseInLocation("2006-01-02", tick.Date, pacificTZ)
					if tickErr != nil {
						continue
					}
				}
			}

			// Skip invalid dates
			if !isTickDateValid(climbedAt) {
				continue
			}

			// Skip if we already have this tick
			if lastTickTime != nil && !climbedAt.After(*lastTickTime) {
				continue
			}

			// Save new tick
			tickModel := &models.MPTick{
				MPRouteID: routeIDInt64,
				UserName:  tick.GetUserName(),
				ClimbedAt: climbedAt,
				Style:     tick.Style,
			}

			textStr := tick.GetTextString()
			if textStr != "" {
				cleanedText := cleanCommentText(textStr)
				if cleanedText != "" {
					tickModel.Comment = &cleanedText
				}
			}

			if tickErr := s.mountainProjectRepo.Ticks().SaveTick(ctx, tickModel); tickErr != nil {
				log.Printf("Error saving tick for route %s: %v", routeID, tickErr)
				continue
			}

			newTickCount++
		}

		totalNewTicks += newTickCount

		// Report progress to monitoring system (success)
		if reporter != nil {
			reporter.Increment(ctx, true)
		}

		return nil
	})

	duration := time.Since(startTime)

	// COMPLETE MONITORING: Mark job as completed or failed
	if jobExec != nil {
		if err != nil {
			s.jobMonitor.FailJob(ctx, jobExec.ID, err.Error())
		} else {
			s.jobMonitor.CompleteJob(ctx, jobExec.ID)
		}
	}

	if err != nil {
		return fmt.Errorf("priority %s tick sync error: %w", priority, err)
	}

	log.Printf("Priority %s tick sync complete: %d routes, %d new ticks, %.1f minutes",
		priority, len(routeIDs), totalNewTicks, duration.Minutes())

	return nil
}

// SyncCommentsByPriority syncs comments for non-location routes at a specific priority tier
func (s *ClimbTrackingService) SyncCommentsByPriority(ctx context.Context, priority string) error {
	startTime := time.Now()

	// Get non-location routes due for comment sync at this priority
	routeIDs, err := s.mountainProjectRepo.Sync().GetRoutesDueForCommentSync(ctx, priority)
	if err != nil {
		return fmt.Errorf("failed to get routes for priority %s comment sync: %w", priority, err)
	}

	if len(routeIDs) == 0 {
		log.Printf("No %s priority routes due for comment sync", priority)
		return nil
	}

	// START MONITORING: Create job execution record
	jobName := fmt.Sprintf("%s_priority_comment_sync", priority)
	jobExec, err := s.jobMonitor.StartJob(ctx, jobName, "comment_sync", len(routeIDs), map[string]interface{}{
		"priority": priority,
	})
	if err != nil {
		log.Printf("Warning: failed to start job monitoring: %v", err)
		jobExec = nil // Continue without monitoring
	}

	// Create progress reporter (updates DB every 10 items or every 5 seconds)
	var reporter *monitoring.ProgressReporter
	if jobExec != nil {
		reporter = monitoring.NewProgressReporter(s.jobMonitor, jobExec.ID, len(routeIDs), 10)
		defer func() {
			// Final flush to ensure last progress is saved
			reporter.FlushProgress(ctx)
		}()
	}

	log.Printf("Syncing comments for %d %s priority routes...", len(routeIDs), priority)

	totalNewComments := 0

	// Sync comments for each route with rate limiting
	err = s.rateLimitedSync(ctx, routeIDs, func(routeID string) error {
		routeIDInt64, _ := strconv.ParseInt(routeID, 10, 64)

		// Fetch comments from MP API
		comments, commentErr := s.mpClient.GetRouteComments(routeID)
		if commentErr != nil {
			// Report progress (failure)
			if reporter != nil {
				reporter.Increment(ctx, false)
			}
			return fmt.Errorf("failed to fetch comments: %w", commentErr)
		}

		// Save comments (upsert handles duplicates)
		newCommentCount := 0
		for _, comment := range comments {
			commentedAt := time.Unix(comment.Created, 0)
			userName := comment.GetUserInfo()
			cleanedText := cleanCommentText(comment.Message)

			if commentErr := s.mountainProjectRepo.Comments().SaveRouteComment(ctx, int64(comment.ID), routeIDInt64, userName, cleanedText, commentedAt); commentErr != nil {
				log.Printf("Error saving comment for route %s: %v", routeID, commentErr)
				continue
			}

			newCommentCount++
		}

		totalNewComments += newCommentCount

		// Report progress to monitoring system (success)
		if reporter != nil {
			reporter.Increment(ctx, true)
		}

		return nil
	})

	duration := time.Since(startTime)

	// COMPLETE MONITORING: Mark job as completed or failed
	if jobExec != nil {
		if err != nil {
			s.jobMonitor.FailJob(ctx, jobExec.ID, err.Error())
		} else {
			s.jobMonitor.CompleteJob(ctx, jobExec.ID)
		}
	}

	if err != nil {
		return fmt.Errorf("priority %s comment sync error: %w", priority, err)
	}

	log.Printf("Priority %s comment sync complete: %d routes, %d new comments, %.1f minutes",
		priority, len(routeIDs), totalNewComments, duration.Minutes())

	return nil
}

// rateLimitedSync processes routes with consistent rate limiting
// 50ms between requests, 10 second pause every 500 requests
func (s *ClimbTrackingService) rateLimitedSync(
	ctx context.Context,
	routeIDs []int64,
	syncFunc func(routeID string) error,
) error {
	requestCount := 0

	for _, routeID := range routeIDs {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Apply sync function
		if err := syncFunc(strconv.FormatInt(routeID, 10)); err != nil {
			log.Printf("Error syncing route %d: %v", routeID, err)
			continue
		}

		requestCount++

		// Rate limiting: 50ms between requests
		time.Sleep(50 * time.Millisecond)

		// Every 500 requests, pause for 10 seconds
		if requestCount%500 == 0 {
			log.Printf("Processed %d requests, pausing for 10 seconds...", requestCount)
			time.Sleep(10 * time.Second)
		}
	}

	return nil
}

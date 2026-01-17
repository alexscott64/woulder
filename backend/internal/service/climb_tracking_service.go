package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alexscott64/woulder/backend/internal/database"
	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/alexscott64/woulder/backend/internal/mountainproject"
)

// MPClientInterface defines the interface for Mountain Project API operations
type MPClientInterface interface {
	GetRouteTicks(routeID string) ([]mountainproject.Tick, error)
	GetArea(areaID string) (*mountainproject.AreaResponse, error)
}

// Ensure real Client implements the interface
var _ MPClientInterface = (*mountainproject.Client)(nil)

// ClimbTrackingService handles Mountain Project data synchronization and retrieval
type ClimbTrackingService struct {
	repo         database.Repository
	mpClient     MPClientInterface
	syncMutex    sync.Mutex
	lastSyncTime time.Time
	isSyncing    bool
}

// NewClimbTrackingService creates a new climb tracking service
func NewClimbTrackingService(
	repo database.Repository,
	mpClient MPClientInterface,
) *ClimbTrackingService {
	return &ClimbTrackingService{
		repo:     repo,
		mpClient: mpClient,
	}
}

// areaQueueItem represents an area to be processed with its associated location
type areaQueueItem struct {
	mpAreaID   string
	locationID *int
	parentID   *string
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

		// Convert area ID to string
		areaIDStr := strconv.Itoa(areaData.ID)

		// Save area to database
		area := &models.MPArea{
			MPAreaID:       areaIDStr,
			Name:           areaData.Title,
			ParentMPAreaID: item.parentID,
			AreaType:       areaData.Type,
			LocationID:     item.locationID,
		}

		if err := s.repo.SaveMPArea(ctx, area); err != nil {
			log.Printf("Error saving area %s: %v", item.mpAreaID, err)
			continue
		}

		processedAreas[item.mpAreaID] = true
		areaCount++
		log.Printf("Saved area: %s (%s) - Total areas: %d", areaData.Title, areaIDStr, areaCount)

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
				// It's a route - check if it's a boulder route
				isBoulder := false
				for _, rt := range child.RouteTypes {
					if strings.ToLower(rt) == "boulder" {
						isBoulder = true
						break
					}
				}

				if !isBoulder {
					continue // Skip non-boulder routes
				}

				// Save route
				route := &models.MPRoute{
					MPRouteID:  childIDStr,
					MPAreaID:   item.mpAreaID,
					Name:       child.Title,
					RouteType:  strings.Join(child.RouteTypes, ", "),
					Rating:     "", // Rating not in children response, could fetch separately if needed
					LocationID: item.locationID,
				}

				if err := s.repo.SaveMPRoute(ctx, route); err != nil {
					log.Printf("Error saving route %s: %v", childIDStr, err)
					continue
				}

				routeCount++
				log.Printf("Saved route: %s (%s) - Total routes: %d", child.Title, childIDStr, routeCount)

				// Fetch and save ticks for this route
				if err := s.syncRouteTicks(ctx, childIDStr); err != nil {
					log.Printf("Error syncing ticks for route %s: %v", childIDStr, err)
					// Continue processing other routes even if tick sync fails
				}
			}
		}
	}

	log.Printf("Sync complete for area %s: %d areas, %d routes processed", rootAreaID, areaCount, routeCount)
	return nil
}

// syncRouteTicks fetches and saves all ticks for a given route
func (s *ClimbTrackingService) syncRouteTicks(ctx context.Context, routeID string) error {
	ticks, err := s.mpClient.GetRouteTicks(routeID)
	if err != nil {
		return fmt.Errorf("failed to fetch ticks: %w", err)
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

		tickModel := &models.MPTick{
			MPRouteID: routeID,
			UserName:  tick.GetUserName(),
			ClimbedAt: climbedAt,
			Style:     tick.Style,
		}

		// Use the 'text' field as the comment
		textStr := tick.GetTextString()
		if textStr != "" {
			tickModel.Comment = &textStr
		}

		if err := s.repo.SaveMPTick(ctx, tickModel); err != nil {
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

// GetLastClimbedForLocation retrieves the most recent climb info for a location
// DEPRECATED: Use GetClimbHistoryForLocation instead
func (s *ClimbTrackingService) GetLastClimbedForLocation(
	ctx context.Context,
	locationID int,
) (*models.LastClimbedInfo, error) {
	return s.repo.GetLastClimbedForLocation(ctx, locationID)
}

// GetClimbHistoryForLocation retrieves recent climb history for a location
func (s *ClimbTrackingService) GetClimbHistoryForLocation(
	ctx context.Context,
	locationID int,
	limit int,
) ([]models.ClimbHistoryEntry, error) {
	return s.repo.GetClimbHistoryForLocation(ctx, locationID, limit)
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
	routeIDs, err := s.repo.GetAllRouteIDsForLocation(ctx, locationID)
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
		lastTickTime, err := s.repo.GetLastTickTimestampForRoute(ctx, routeID)
		if err != nil {
			log.Printf("Error getting last tick for route %s: %v", routeID, err)
			continue
		}

		// Fetch ticks from Mountain Project
		ticks, err := s.mpClient.GetRouteTicks(routeID)
		if err != nil {
			log.Printf("Error fetching ticks for route %s: %v", routeID, err)
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
						log.Printf("Warning: invalid date format for tick on route %s: %s", routeID, tick.Date)
						continue
					}
				}
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
				tickModel.Comment = &textStr
			}

			if err := s.repo.SaveMPTick(ctx, tickModel); err != nil {
				log.Printf("Error saving tick for route %s: %v", routeID, err)
				continue
			}

			newTickCount++
		}

		if newTickCount > 0 {
			totalNewTicks += newTickCount
			log.Printf("Route %s: %d new ticks", routeID, newTickCount)
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

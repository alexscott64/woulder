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

// ClimbTrackingService handles Mountain Project data synchronization and retrieval
type ClimbTrackingService struct {
	repo         database.Repository
	mpClient     *mountainproject.Client
	syncMutex    sync.Mutex
	lastSyncTime time.Time
	isSyncing    bool
}

// NewClimbTrackingService creates a new climb tracking service
func NewClimbTrackingService(
	repo database.Repository,
	mpClient *mountainproject.Client,
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

	tickCount := 0
	for _, tick := range ticks {
		// Parse the date - Mountain Project API returns multiple formats
		var climbedAt time.Time
		var err error

		// Format 1: "Jan 2, 2006, 3:04 pm"
		climbedAt, err = time.Parse("Jan 2, 2006, 3:04 pm", tick.Date)
		if err != nil {
			// Format 2: "2006-01-02 15:04:05"
			climbedAt, err = time.Parse("2006-01-02 15:04:05", tick.Date)
			if err != nil {
				// Format 3: "2006-01-02"
				climbedAt, err = time.Parse("2006-01-02", tick.Date)
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

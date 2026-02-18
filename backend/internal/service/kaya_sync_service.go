package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	kayaDB "github.com/alexscott64/woulder/backend/internal/database/kaya"
	kayaClient "github.com/alexscott64/woulder/backend/internal/kaya"
	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/alexscott64/woulder/backend/internal/monitoring"
)

// KayaClientInterface defines the interface for Kaya API operations
type KayaClientInterface interface {
	GetLocation(slug string) (*kayaClient.WebLocation, error)
	GetSubLocations(locationID string, climbTypeID *string, offset, count int) ([]*kayaClient.WebLocation, error)
	GetClimbs(locationID string, climbTypeID *string, offset, count int) ([]*kayaClient.WebClimb, error)
	GetAscents(locationID string, offset, count int) ([]*kayaClient.WebAscent, error)
	GetPosts(locationID string, subLocationIDs []string, offset, count int) ([]*kayaClient.WebPost, error)
}

// Ensure both Client and BrowserClient implement the interface
var _ KayaClientInterface = (*kayaClient.Client)(nil)
var _ KayaClientInterface = (*kayaClient.BrowserClient)(nil)

// KayaSyncService handles Kaya data synchronization and retrieval
type KayaSyncService struct {
	kayaRepo   kayaDB.Repository
	kayaClient KayaClientInterface
	jobMonitor *monitoring.JobMonitor
	syncMutex  sync.Mutex
	isSyncing  bool
}

// NewKayaSyncService creates a new Kaya sync service
func NewKayaSyncService(
	kayaRepo kayaDB.Repository,
	kayaClient KayaClientInterface,
	jobMonitor *monitoring.JobMonitor,
) *KayaSyncService {
	return &KayaSyncService{
		kayaRepo:   kayaRepo,
		kayaClient: kayaClient,
		jobMonitor: jobMonitor,
	}
}

// SyncLocationBySlug syncs a single location and optionally its sub-locations
func (s *KayaSyncService) SyncLocationBySlug(ctx context.Context, slug string, recursive bool) error {
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
		s.syncMutex.Unlock()
	}()

	log.Printf("[Kaya] Starting sync for location slug: %s (recursive: %v)", slug, recursive)

	// Fetch location from Kaya API
	location, err := s.kayaClient.GetLocation(slug)
	if err != nil {
		return fmt.Errorf("failed to fetch location %s: %w", slug, err)
	}

	if location == nil {
		return fmt.Errorf("location not found: %s", slug)
	}

	// Save location to database
	if err := s.saveLocation(ctx, location); err != nil {
		return fmt.Errorf("failed to save location: %w", err)
	}

	// Initialize sync progress tracking
	progress := &models.KayaSyncProgress{
		KayaLocationID: location.ID,
		LocationName:   location.Name,
		Status:         "in_progress",
		LastSyncAt:     timePtr(time.Now()),
	}
	if err := s.kayaRepo.Sync().SaveSyncProgress(ctx, progress); err != nil {
		log.Printf("[Kaya] Warning: failed to save sync progress: %v", err)
	}

	var syncError error

	// Sync climbs for this location (both boulders and routes)
	climbsSynced := 0
	if climbs, err := s.syncClimbsForLocation(ctx, location.ID); err != nil {
		log.Printf("[Kaya] Warning: failed to sync climbs for location %s: %v", location.ID, err)
		syncError = err
	} else {
		climbsSynced = climbs
	}

	// Sync ascents for this location
	ascentsSynced := 0
	if ascents, err := s.syncAscentsForLocation(ctx, location.ID); err != nil {
		log.Printf("[Kaya] Warning: failed to sync ascents for location %s: %v", location.ID, err)
		if syncError == nil {
			syncError = err
		}
	} else {
		ascentsSynced = ascents
	}

	// Sync sub-locations if recursive
	subLocationsSynced := 0
	if recursive {
		if subLocs, err := s.syncSubLocations(ctx, location.ID); err != nil {
			log.Printf("[Kaya] Warning: failed to sync sub-locations for %s: %v", location.ID, err)
			if syncError == nil {
				syncError = err
			}
		} else {
			subLocationsSynced = subLocs
		}
	}

	// Update sync progress
	status := "completed"
	var errMsg *string
	if syncError != nil {
		status = "failed"
		msg := syncError.Error()
		errMsg = &msg
	}

	nextSync := time.Now().Add(24 * time.Hour) // Re-sync daily
	progress = &models.KayaSyncProgress{
		KayaLocationID:     location.ID,
		LocationName:       location.Name,
		Status:             status,
		LastSyncAt:         timePtr(time.Now()),
		NextSyncAt:         &nextSync,
		SyncError:          errMsg,
		ClimbsSynced:       climbsSynced,
		AscentsSynced:      ascentsSynced,
		SubLocationsSynced: subLocationsSynced,
	}
	if err := s.kayaRepo.Sync().SaveSyncProgress(ctx, progress); err != nil {
		log.Printf("[Kaya] Warning: failed to save final sync progress: %v", err)
	}

	log.Printf("[Kaya] Sync completed for %s: climbs=%d, ascents=%d, sub-locations=%d",
		location.Name, climbsSynced, ascentsSynced, subLocationsSynced)

	return syncError
}

// saveLocation converts API location to model and saves it
func (s *KayaSyncService) saveLocation(ctx context.Context, apiLoc *kayaClient.WebLocation) error {
	// Convert string lat/lon to float64
	var lat, lon *float64
	if apiLoc.Latitude != nil {
		if f, err := strconv.ParseFloat(*apiLoc.Latitude, 64); err == nil {
			lat = &f
		}
	}
	if apiLoc.Longitude != nil {
		if f, err := strconv.ParseFloat(*apiLoc.Longitude, 64); err == nil {
			lon = &f
		}
	}

	loc := &models.KayaLocation{
		KayaLocationID:                    apiLoc.ID,
		Slug:                              apiLoc.Slug,
		Name:                              apiLoc.Name,
		Latitude:                          lat,
		Longitude:                         lon,
		PhotoURL:                          apiLoc.PhotoURL,
		Description:                       apiLoc.Description,
		ClimbCount:                        apiLoc.ClimbCount,
		BoulderCount:                      apiLoc.BoulderCount,
		RouteCount:                        apiLoc.RouteCount,
		AscentCount:                       apiLoc.AscentCount,
		IsGBModeratedBouldering:           apiLoc.IsGBModeratedBouldering,
		IsGBModeratedRoutes:               apiLoc.IsGBModeratedRoutes,
		IsAccessSensitive:                 apiLoc.IsAccessSensitive,
		IsClosed:                          apiLoc.IsClosed,
		HasMapsDisabled:                   apiLoc.HasMapsDisabled,
		DescriptionBouldering:             apiLoc.DescriptionBouldering,
		DescriptionRoutes:                 apiLoc.DescriptionRoutes,
		DescriptionShortBouldering:        apiLoc.DescriptionShortBouldering,
		DescriptionShortRoutes:            apiLoc.DescriptionShortRoutes,
		AccessDescriptionBouldering:       apiLoc.AccessDescriptionBouldering,
		AccessDescriptionRoutes:           apiLoc.AccessDescriptionRoutes,
		AccessIssuesDescriptionBouldering: apiLoc.AccessIssuesDescriptionBouldering,
		AccessIssuesDescriptionRoutes:     apiLoc.AccessIssuesDescriptionRoutes,
		ClimbTypeID:                       apiLoc.ClimbTypeID,
	}

	// Handle location type
	if apiLoc.LocationType != nil {
		loc.LocationTypeID = &apiLoc.LocationType.ID
		loc.LocationTypeName = &apiLoc.LocationType.Name
	}

	// Handle parent location
	if apiLoc.ParentLocation != nil {
		loc.ParentLocationID = &apiLoc.ParentLocation.ID
		loc.ParentLocationSlug = &apiLoc.ParentLocation.Slug
		loc.ParentLocationName = &apiLoc.ParentLocation.Name
	}

	// Handle closed date
	if apiLoc.ClosedDate != nil {
		if t, err := time.Parse(time.RFC3339, *apiLoc.ClosedDate); err == nil {
			loc.ClosedDate = &t
		}
	}

	return s.kayaRepo.Locations().SaveLocation(ctx, loc)
}

// syncClimbsForLocation syncs all climbs for a location with pagination
func (s *KayaSyncService) syncClimbsForLocation(ctx context.Context, locationID string) (int, error) {
	const pageSize = 20 // Kaya API limit
	offset := 0
	totalSynced := 0

	// Sync both boulders (type 1) and routes (type 2)
	climbTypes := []string{"1", "2"}

	for _, climbTypeID := range climbTypes {
		offset = 0
		for {
			climbs, err := s.kayaClient.GetClimbs(locationID, &climbTypeID, offset, pageSize)
			if err != nil {
				return totalSynced, fmt.Errorf("failed to fetch climbs (type %s, offset %d): %w", climbTypeID, offset, err)
			}

			if len(climbs) == 0 {
				break // No more climbs
			}

			// Save each climb
			for _, climb := range climbs {
				if err := s.saveClimb(ctx, climb); err != nil {
					log.Printf("[Kaya] Warning: failed to save climb %s: %v", climb.Slug, err)
					continue
				}
				totalSynced++
			}

			log.Printf("[Kaya] Synced %d climbs (type %s) for location %s", len(climbs), climbTypeID, locationID)

			// Check if we've reached the end
			if len(climbs) < pageSize {
				break
			}

			offset += pageSize
		}
	}

	return totalSynced, nil
}

// saveClimb converts API climb to model and saves it
func (s *KayaSyncService) saveClimb(ctx context.Context, apiClimb *kayaClient.WebClimb) error {
	// First, ensure user exists if present in climb data
	// (for future enhancement when climbs have user references)

	climb := &models.KayaClimb{
		Slug:              apiClimb.Slug,
		Name:              apiClimb.Name,
		Rating:            apiClimb.Rating,
		AscentCount:       apiClimb.AscentCount,
		IsGBModerated:     apiClimb.IsGBModerated,
		IsAccessSensitive: apiClimb.IsAccessSensitive,
		IsClosed:          apiClimb.IsClosed,
		IsOffensive:       apiClimb.IsOffensive,
	}

	// Handle grade
	if apiClimb.Grade != nil {
		climb.GradeID = &apiClimb.Grade.ID
		climb.GradeName = &apiClimb.Grade.Name
		climb.GradeOrdering = apiClimb.Grade.Ordering
		climb.GradeClimbTypeID = apiClimb.Grade.ClimbTypeID
	}

	// Handle climb type
	if apiClimb.ClimbType != nil {
		climb.ClimbTypeName = &apiClimb.ClimbType.Name
	}

	// Handle destination (top-level location)
	if apiClimb.Destination != nil {
		climb.KayaDestinationName = &apiClimb.Destination.Name
	}

	// Handle area (sub-location)
	if apiClimb.Area != nil {
		climb.KayaAreaName = &apiClimb.Area.Name
	}

	// Handle indoor climb fields
	if apiClimb.Color != nil {
		climb.ColorName = &apiClimb.Color.Name
	}
	if apiClimb.Gym != nil {
		climb.GymName = &apiClimb.Gym.Name
	}
	if apiClimb.Board != nil {
		climb.BoardName = &apiClimb.Board.Name
	}

	return s.kayaRepo.Climbs().SaveClimb(ctx, climb)
}

// syncAscentsForLocation syncs recent ascents for a location with pagination
func (s *KayaSyncService) syncAscentsForLocation(ctx context.Context, locationID string) (int, error) {
	const pageSize = 15    // Kaya API limit (captured queries use 15)
	const maxAscents = 500 // Limit to recent ascents to avoid overwhelming database
	offset := 0
	totalSynced := 0

	for totalSynced < maxAscents {
		ascents, err := s.kayaClient.GetAscents(locationID, offset, pageSize)
		if err != nil {
			return totalSynced, fmt.Errorf("failed to fetch ascents (offset %d): %w", offset, err)
		}

		if len(ascents) == 0 {
			break // No more ascents
		}

		// Save each ascent
		for _, ascent := range ascents {
			if err := s.saveAscent(ctx, ascent); err != nil {
				log.Printf("[Kaya] Warning: failed to save ascent %s: %v", ascent.ID, err)
				continue
			}
			totalSynced++
		}

		log.Printf("[Kaya] Synced %d ascents for location %s (total: %d)", len(ascents), locationID, totalSynced)

		// Check if we've reached the end or limit
		if len(ascents) < pageSize || totalSynced >= maxAscents {
			break
		}

		offset += pageSize
	}

	return totalSynced, nil
}

// saveAscent converts API ascent to model and saves it (with user)
func (s *KayaSyncService) saveAscent(ctx context.Context, apiAscent *kayaClient.WebAscent) error {
	// First, ensure user exists
	if apiAscent.User != nil {
		if err := s.saveUser(ctx, apiAscent.User); err != nil {
			log.Printf("[Kaya] Warning: failed to save user %s: %v", apiAscent.User.ID, err)
		}
	}

	// Ensure climb exists (save basic info from ascent)
	if apiAscent.Climb != nil {
		if err := s.saveClimb(ctx, apiAscent.Climb); err != nil {
			log.Printf("[Kaya] Warning: failed to save climb %s: %v", apiAscent.Climb.Slug, err)
		}
	}

	// Parse date
	ascentDate, err := time.Parse(time.RFC3339, apiAscent.Date)
	if err != nil {
		// Try alternative formats
		ascentDate, err = time.Parse("2006-01-02", apiAscent.Date)
		if err != nil {
			return fmt.Errorf("failed to parse ascent date %s: %w", apiAscent.Date, err)
		}
	}

	ascent := &models.KayaAscent{
		KayaAscentID: apiAscent.ID,
		Date:         ascentDate,
		Comment:      apiAscent.Comment,
		Rating:       apiAscent.Rating,
		Stiffness:    apiAscent.Stiffness,
	}

	// Handle climb reference
	if apiAscent.Climb != nil {
		ascent.KayaClimbSlug = apiAscent.Climb.Slug
	}

	// Handle user reference
	if apiAscent.User != nil {
		ascent.KayaUserID = apiAscent.User.ID
	}

	// Handle grade
	if apiAscent.Grade != nil {
		ascent.GradeID = &apiAscent.Grade.ID
		ascent.GradeName = &apiAscent.Grade.Name
	}

	// Handle media
	if apiAscent.Photo != nil {
		ascent.PhotoURL = apiAscent.Photo.PhotoURL
		ascent.PhotoThumbURL = apiAscent.Photo.ThumbURL
	}
	if apiAscent.Video != nil {
		ascent.VideoURL = apiAscent.Video.VideoURL
		ascent.VideoThumbURL = apiAscent.Video.ThumbURL
	}

	return s.kayaRepo.Ascents().SaveAscent(ctx, ascent)
}

// saveUser converts API user to model and saves it
func (s *KayaSyncService) saveUser(ctx context.Context, apiUser *kayaClient.WebUser) error {
	// Convert height from float64 to int (API returns float, DB expects int)
	var height *int
	if apiUser.Height != nil {
		h := int(*apiUser.Height)
		height = &h
	}

	user := &models.KayaUser{
		KayaUserID: apiUser.ID,
		Username:   apiUser.Username,
		Fname:      apiUser.Fname,
		Lname:      apiUser.Lname,
		PhotoURL:   apiUser.PhotoURL,
		Bio:        apiUser.Bio,
		Height:     height,
		ApeIndex:   apiUser.ApeIndex,
		IsPrivate:  apiUser.IsPrivate,
		IsPremium:  apiUser.IsPremium,
	}

	// Handle limit grades
	if apiUser.LimitGradeBouldering != nil {
		user.LimitGradeBoulderingID = &apiUser.LimitGradeBouldering.ID
		user.LimitGradeBoulderingName = &apiUser.LimitGradeBouldering.Name
	}
	if apiUser.LimitGradeRoutes != nil {
		user.LimitGradeRoutesID = &apiUser.LimitGradeRoutes.ID
		user.LimitGradeRoutesName = &apiUser.LimitGradeRoutes.Name
	}

	return s.kayaRepo.Users().SaveUser(ctx, user)
}

// syncSubLocations syncs all sub-locations for a location
func (s *KayaSyncService) syncSubLocations(ctx context.Context, locationID string) (int, error) {
	const pageSize = 20 // Kaya API limit
	offset := 0
	totalSynced := 0

	for {
		subLocs, err := s.kayaClient.GetSubLocations(locationID, nil, offset, pageSize)
		if err != nil {
			return totalSynced, fmt.Errorf("failed to fetch sub-locations (offset %d): %w", offset, err)
		}

		if len(subLocs) == 0 {
			break // No more sub-locations
		}

		// Save each sub-location
		for _, subLoc := range subLocs {
			if err := s.saveLocation(ctx, subLoc); err != nil {
				log.Printf("[Kaya] Warning: failed to save sub-location %s: %v", subLoc.Slug, err)
				continue
			}
			totalSynced++

			// Recursively sync this sub-location (climbs and ascents only, not more sub-locations)
			log.Printf("[Kaya] Syncing sub-location: %s", subLoc.Name)
			if climbs, err := s.syncClimbsForLocation(ctx, subLoc.ID); err != nil {
				log.Printf("[Kaya] Warning: failed to sync climbs for sub-location %s: %v", subLoc.ID, err)
			} else {
				log.Printf("[Kaya] Synced %d climbs for sub-location %s", climbs, subLoc.Name)
			}

			if ascents, err := s.syncAscentsForLocation(ctx, subLoc.ID); err != nil {
				log.Printf("[Kaya] Warning: failed to sync ascents for sub-location %s: %v", subLoc.ID, err)
			} else {
				log.Printf("[Kaya] Synced %d ascents for sub-location %s", ascents, subLoc.Name)
			}
		}

		log.Printf("[Kaya] Synced %d sub-locations for location %s", len(subLocs), locationID)

		// Check if we've reached the end
		if len(subLocs) < pageSize {
			break
		}

		offset += pageSize
	}

	return totalSynced, nil
}

// GetLocationBySlug retrieves a location from the database
func (s *KayaSyncService) GetLocationBySlug(ctx context.Context, slug string) (*models.KayaLocation, error) {
	return s.kayaRepo.Locations().GetLocationBySlug(ctx, slug)
}

// GetClimbsByLocation retrieves all climbs for a location
func (s *KayaSyncService) GetClimbsByLocation(ctx context.Context, locationID string) ([]*models.KayaClimb, error) {
	return s.kayaRepo.Climbs().GetClimbsByLocation(ctx, locationID)
}

// GetRecentAscents retrieves recent ascents across all climbs
func (s *KayaSyncService) GetRecentAscents(ctx context.Context, limit int) ([]*models.KayaAscent, error) {
	return s.kayaRepo.Ascents().GetRecentAscents(ctx, limit)
}

// timePtr returns a pointer to a time.Time
func timePtr(t time.Time) *time.Time {
	return &t
}

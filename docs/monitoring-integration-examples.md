package service

// EXAMPLE: Integration of JobMonitor into ClimbTrackingService
// This file shows how to integrate the monitoring system into existing sync jobs

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/alexscott64/woulder/backend/internal/monitoring"
)

// Example 1: Simple integration for SyncTicksByPriority
func (s *ClimbTrackingService) SyncTicksByPriorityWithMonitoring(
	ctx context.Context,
	priority string,
	jobMonitor *monitoring.JobMonitor,
) error {
	// Get routes to process
	routeIDs, err := s.mountainProjectRepo.Sync().GetRoutesDueForTickSync(ctx, priority)
	if err != nil {
		return fmt.Errorf("failed to get routes: %w", err)
	}

	if len(routeIDs) == 0 {
		log.Printf("No %s priority routes due for tick sync", priority)
		return nil
	}

	// START MONITORING: Create job execution record
	jobName := fmt.Sprintf("%s_priority_tick_sync", priority)
	jobExec, err := jobMonitor.StartJob(ctx, jobName, "tick_sync", len(routeIDs), map[string]interface{}{
		"priority": priority,
	})
	if err != nil {
		// Log but don't fail - monitoring is non-critical
		log.Printf("Warning: failed to start job monitoring: %v", err)
		jobExec = nil // Continue without monitoring
	}

	// Create progress reporter (updates DB every 10 items or every 5 seconds)
	var reporter *monitoring.ProgressReporter
	if jobExec != nil {
		reporter = monitoring.NewProgressReporter(jobMonitor, jobExec.ID, len(routeIDs), 10)
		defer func() {
			// Final flush to ensure last progress is saved
			reporter.FlushProgress(ctx)
		}()
	}

	log.Printf("Syncing ticks for %d %s priority routes...", len(routeIDs), priority)

	// Process routes with progress tracking
	successCount := 0
	failCount := 0

	for _, routeID := range routeIDs {
		// Sync ticks for this route (existing logic)
		err := s.syncRouteTicksInternal(ctx, routeID)

		// Track progress
		success := err == nil
		if success {
			successCount++
		} else {
			failCount++
			log.Printf("Error syncing route %s: %v", routeID, err)
		}

		// Report progress to monitoring system
		if reporter != nil {
			reporter.Increment(ctx, success)
		}
	}

	// COMPLETE MONITORING: Mark job as completed or failed
	if jobExec != nil {
		if failCount == 0 {
			jobMonitor.CompleteJob(ctx, jobExec.ID)
		} else if successCount == 0 {
			// All failed
			jobMonitor.FailJob(ctx, jobExec.ID, fmt.Sprintf("All %d routes failed", failCount))
		} else {
			// Partial success - still mark as complete but log failures
			jobMonitor.CompleteJob(ctx, jobExec.ID)
			log.Printf("Job completed with %d failures out of %d routes", failCount, len(routeIDs))
		}
	}

	log.Printf("Priority %s tick sync complete: %d success, %d failed", priority, successCount, failCount)
	return nil
}

// Example 2: Integration with handler's background sync starter
func (h *Handler) StartHighPrioritySyncWithMonitoring(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		log.Printf("Starting high-priority sync scheduler (every %v)", interval)

		// Run immediately on startup
		ctx := context.Background()
		h.runHighPrioritySyncWithMonitoring(ctx)

		// Then run on schedule
		for range ticker.C {
			log.Println("Starting scheduled high-priority sync...")
			ctx := context.Background()
			h.runHighPrioritySyncWithMonitoring(ctx)
		}
	}()
}

func (h *Handler) runHighPrioritySyncWithMonitoring(ctx context.Context) {
	log.Println("Starting high-priority sync (ticks + comments)...")

	// Sync ticks with monitoring
	if err := h.climbTrackingService.SyncTicksByPriorityWithMonitoring(ctx, "high", h.jobMonitor); err != nil {
		log.Printf("Error in high-priority tick sync: %v", err)
	}

	// Sync comments with monitoring
	if err := h.climbTrackingService.SyncCommentsByPriorityWithMonitoring(ctx, "high", h.jobMonitor); err != nil {
		log.Printf("Error in high-priority comment sync: %v", err)
	}

	log.Println("High-priority sync complete")
}

// Example 3: Integration with priority recalculation
func (s *ClimbTrackingService) RecalculateAllPrioritiesWithMonitoring(
	ctx context.Context,
	jobMonitor *monitoring.JobMonitor,
) error {
	// START MONITORING
	jobExec, err := jobMonitor.StartJob(ctx, "priority_recalculation", "priority_calc", 1, map[string]interface{}{
		"description": "Recalculating route sync priorities for all non-location routes",
	})
	if err != nil {
		log.Printf("Warning: failed to start job monitoring: %v", err)
	}

	// Run the recalculation (existing logic)
	err = s.mountainProjectRepo.Sync().UpdateRouteSyncPriorities(ctx)

	// COMPLETE MONITORING
	if jobExec != nil {
		if err != nil {
			jobMonitor.FailJob(ctx, jobExec.ID, err.Error())
		} else {
			// Mark as processed (this is a single-step job)
			jobMonitor.UpdateProgress(ctx, jobExec.ID, 1, 1, 0)
			jobMonitor.CompleteJob(ctx, jobExec.ID)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to recalculate priorities: %w", err)
	}

	// Get priority distribution for logging
	distribution, err := s.mountainProjectRepo.Sync().GetPriorityDistribution(ctx)
	if err == nil {
		log.Printf("Priority distribution: high=%d, medium=%d, low=%d",
			distribution["high"], distribution["medium"], distribution["low"])
	}

	return nil
}

// Example 4: Integration pattern for existing methods
// This shows the minimal changes needed to add monitoring to any sync function

// BEFORE (no monitoring):
func (s *ClimbTrackingService) SyncSomethingOld(ctx context.Context) error {
	items := getItemsToSync()

	for _, item := range items {
		syncItem(item)
	}

	return nil
}

// AFTER (with monitoring):
func (s *ClimbTrackingService) SyncSomethingNew(
	ctx context.Context,
	jobMonitor *monitoring.JobMonitor,
) error {
	items := getItemsToSync()

	// Add 3 lines to start monitoring
	jobExec, err := jobMonitor.StartJob(ctx, "sync_something", "sync_type", len(items), nil)
	if err != nil {
		log.Printf("Warning: failed to start monitoring: %v", err)
	}
	reporter := monitoring.NewProgressReporter(jobMonitor, jobExec.ID, len(items), 10)

	for _, item := range items {
		err := syncItem(item)

		// Add 1 line to report progress
		reporter.Increment(ctx, err == nil)
	}

	// Add 1 line to complete monitoring
	jobMonitor.CompleteJob(ctx, jobExec.ID)

	return nil
}

// Example 5: Error handling best practices
func (s *ClimbTrackingService) SyncWithProperErrorHandling(
	ctx context.Context,
	jobMonitor *monitoring.JobMonitor,
) error {
	jobExec, err := jobMonitor.StartJob(ctx, "my_sync_job", "sync_type", 100, nil)
	if err != nil {
		log.Printf("Warning: monitoring not available: %v", err)
		// Continue with sync even if monitoring fails
		jobExec = nil
	}

	// Ensure cleanup happens even on panic
	defer func() {
		if jobExec != nil && r := recover(); r != nil {
			errMsg := fmt.Sprintf("Panic during sync: %v", r)
			jobMonitor.FailJob(context.Background(), jobExec.ID, errMsg)
			panic(r) // Re-panic after recording
		}
	}()

	// Your sync logic here
	err = doActualSync()

	// Complete monitoring
	if jobExec != nil {
		if err != nil {
			jobMonitor.FailJob(ctx, jobExec.ID, err.Error())
		} else {
			jobMonitor.CompleteJob(ctx, jobExec.ID)
		}
	}

	return err
}

// Example 6: Adding metadata for rich monitoring
func (s *ClimbTrackingService) SyncWithMetadata(
	ctx context.Context,
	jobMonitor *monitoring.JobMonitor,
	priority string,
) error {
	routes, err := s.getRoutesForPriority(ctx, priority)
	if err != nil {
		return err
	}

	// Add metadata that will be visible in monitoring
	metadata := map[string]interface{}{
		"priority":       priority,
		"total_routes":   len(routes),
		"sync_type":      "full",
		"initiated_by":   "scheduler",
		"server_version": "1.0.0",
	}

	jobExec, err := jobMonitor.StartJob(ctx,
		fmt.Sprintf("%s_priority_sync", priority),
		"tick_sync",
		len(routes),
		metadata)

	if err != nil {
		log.Printf("Warning: monitoring unavailable: %v", err)
		jobExec = nil
	}

	reporter := monitoring.NewProgressReporter(jobMonitor, jobExec.ID, len(routes), 10)

	// Track metrics
	startTime := time.Now()
	var totalBytes int64

	for _, route := range routes {
		result := syncRoute(route)
		reporter.Increment(ctx, result.Success)
		totalBytes += result.BytesDownloaded
	}

	// Update metadata with final statistics
	if jobExec != nil {
		duration := time.Since(startTime).Seconds()
		metadata["duration_seconds"] = duration
		metadata["bytes_downloaded"] = totalBytes
		metadata["routes_per_second"] = float64(len(routes)) / duration

		jobMonitor.UpdateMetadata(ctx, jobExec.ID, metadata)
		jobMonitor.CompleteJob(ctx, jobExec.ID)
	}

	return nil
}

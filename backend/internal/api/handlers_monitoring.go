package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/alexscott64/woulder/backend/internal/monitoring"
	"github.com/gin-gonic/gin"
)

// JobExecutionResponse extends JobExecution with calculated fields
type JobExecutionResponse struct {
	*monitoring.JobExecution
	ProgressPercent           float64  `json:"progress_percent"`
	ElapsedSeconds            int64    `json:"elapsed_seconds"`
	EstimatedRemainingSeconds *int64   `json:"estimated_remaining_seconds,omitempty"`
	ItemsPerSecond            *float64 `json:"items_per_second,omitempty"`
}

// JobsSummaryResponse provides overview of all job types
type JobsSummaryResponse struct {
	Summary map[string]*JobSummaryItem `json:"summary"`
}

// JobSummaryItem contains summary info for a job type
type JobSummaryItem struct {
	LastRun         *time.Time `json:"last_run,omitempty"`
	Status          string     `json:"status"`
	Progress        *string    `json:"progress,omitempty"`
	DurationSeconds *int64     `json:"duration_seconds,omitempty"`
	NextScheduled   *time.Time `json:"next_scheduled,omitempty"`
	ErrorMessage    *string    `json:"error_message,omitempty"`
}

// GetActiveJobs returns all currently running jobs
// GET /api/monitoring/jobs/active
func (h *Handler) GetActiveJobs(c *gin.Context) {
	ctx := c.Request.Context()

	jobs, err := h.jobMonitor.GetActiveJobs(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get active jobs: %v", err)})
		return
	}

	// Enhance with calculated fields
	response := make([]*JobExecutionResponse, len(jobs))
	for i, job := range jobs {
		response[i] = enhanceJobExecution(job)
	}

	c.JSON(http.StatusOK, gin.H{"jobs": response})
}

// GetJobHistory returns recent job executions
// GET /api/monitoring/jobs/history?job_name=high_priority_tick_sync&limit=10
func (h *Handler) GetJobHistory(c *gin.Context) {
	ctx := c.Request.Context()

	jobName := c.Query("job_name")
	limitStr := c.DefaultQuery("limit", "20")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter (must be 1-100)"})
		return
	}

	var jobs []*monitoring.JobExecution
	if jobName != "" {
		jobs, err = h.jobMonitor.GetJobHistory(ctx, jobName, limit)
	} else {
		jobs, err = h.jobMonitor.GetAllJobHistory(ctx, limit)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get job history: %v", err)})
		return
	}

	// Enhance with calculated fields
	response := make([]*JobExecutionResponse, len(jobs))
	for i, job := range jobs {
		response[i] = enhanceJobExecution(job)
	}

	c.JSON(http.StatusOK, gin.H{"jobs": response})
}

// GetJobStatus returns specific job status
// GET /api/monitoring/jobs/:job_id
func (h *Handler) GetJobStatus(c *gin.Context) {
	ctx := c.Request.Context()

	jobIDStr := c.Param("job_id")
	jobID, err := strconv.ParseInt(jobIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	job, err := h.jobMonitor.GetJobStatus(ctx, jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Job not found: %v", err)})
		return
	}

	response := enhanceJobExecution(job)
	c.JSON(http.StatusOK, response)
}

// GetJobsSummary returns summary of all job types with latest status
// GET /api/monitoring/jobs/summary
func (h *Handler) GetJobsSummary(c *gin.Context) {
	ctx := c.Request.Context()

	// Define all known job names and their schedule intervals
	jobSchedules := map[string]time.Duration{
		"priority_recalculation":       24 * time.Hour,
		"location_tick_sync":           24 * time.Hour,
		"location_comment_sync":        24 * time.Hour,
		"high_priority_tick_sync":      24 * time.Hour,
		"high_priority_comment_sync":   24 * time.Hour,
		"medium_priority_tick_sync":    7 * 24 * time.Hour,
		"medium_priority_comment_sync": 7 * 24 * time.Hour,
		"low_priority_tick_sync":       30 * 24 * time.Hour,
		"low_priority_comment_sync":    30 * 24 * time.Hour,
		"route_count_backfill":         24 * time.Hour,
	}

	summary := make(map[string]*JobSummaryItem)

	// Get latest execution for each job type
	for jobName, interval := range jobSchedules {
		job, err := h.jobMonitor.GetLatestJobByName(ctx, jobName)
		if err != nil {
			// Job never run yet
			summary[jobName] = &JobSummaryItem{
				Status: "never_run",
			}
			continue
		}

		item := &JobSummaryItem{
			LastRun: &job.StartedAt,
			Status:  job.Status,
		}

		// Calculate duration
		if job.CompletedAt != nil {
			duration := job.CompletedAt.Sub(job.StartedAt)
			durationSec := int64(duration.Seconds())
			item.DurationSeconds = &durationSec
		} else if job.Status == monitoring.StatusRunning {
			// Running - show elapsed time
			duration := time.Since(job.StartedAt)
			durationSec := int64(duration.Seconds())
			item.DurationSeconds = &durationSec

			// Show progress
			if job.TotalItems > 0 {
				progress := fmt.Sprintf("%d/%d (%.1f%%)",
					job.ItemsProcessed,
					job.TotalItems,
					float64(job.ItemsProcessed)/float64(job.TotalItems)*100)
				item.Progress = &progress
			}
		}

		// Calculate next scheduled run
		if job.Status == monitoring.StatusCompleted {
			nextRun := job.StartedAt.Add(interval)
			item.NextScheduled = &nextRun
		}

		// Add error message if failed
		if job.Status == monitoring.StatusFailed && job.ErrorMessage != nil {
			item.ErrorMessage = job.ErrorMessage
		}

		summary[jobName] = item
	}

	c.JSON(http.StatusOK, JobsSummaryResponse{Summary: summary})
}

// enhanceJobExecution adds calculated fields to job execution
func enhanceJobExecution(job *monitoring.JobExecution) *JobExecutionResponse {
	response := &JobExecutionResponse{
		JobExecution: job,
	}

	// Calculate progress percentage
	if job.TotalItems > 0 {
		response.ProgressPercent = float64(job.ItemsProcessed) / float64(job.TotalItems) * 100
	}

	// Calculate elapsed time
	var endTime time.Time
	if job.CompletedAt != nil {
		endTime = *job.CompletedAt
	} else {
		endTime = time.Now()
	}
	elapsed := endTime.Sub(job.StartedAt)
	response.ElapsedSeconds = int64(elapsed.Seconds())

	// Calculate items per second and estimated remaining time (only for running jobs)
	if job.Status == monitoring.StatusRunning && job.ItemsProcessed > 0 && elapsed.Seconds() > 0 {
		itemsPerSec := float64(job.ItemsProcessed) / elapsed.Seconds()
		response.ItemsPerSecond = &itemsPerSec

		// Estimate remaining time
		if job.TotalItems > job.ItemsProcessed && itemsPerSec > 0 {
			remaining := job.TotalItems - job.ItemsProcessed
			estimatedSec := int64(float64(remaining) / itemsPerSec)
			response.EstimatedRemainingSeconds = &estimatedSec
		}
	}

	return response
}

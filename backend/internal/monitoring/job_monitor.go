package monitoring

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// JobExecution represents a running or completed job
type JobExecution struct {
	ID             int64                  `json:"id"`
	JobName        string                 `json:"job_name"`
	JobType        string                 `json:"job_type"`
	Status         string                 `json:"status"`
	TotalItems     int                    `json:"total_items"`
	ItemsProcessed int                    `json:"items_processed"`
	ItemsSucceeded int                    `json:"items_succeeded"`
	ItemsFailed    int                    `json:"items_failed"`
	ErrorMessage   *string                `json:"error_message,omitempty"`
	StartedAt      time.Time              `json:"started_at"`
	CompletedAt    *time.Time             `json:"completed_at,omitempty"`
	UpdatedAt      time.Time              `json:"updated_at"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// JobStatus constants
const (
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusCancelled = "cancelled"
)

// JobMonitor tracks job execution progress
type JobMonitor struct {
	db *sql.DB
}

// NewJobMonitor creates a new job monitor
func NewJobMonitor(db *sql.DB) *JobMonitor {
	return &JobMonitor{db: db}
}

// StartJob creates a new job execution record
func (m *JobMonitor) StartJob(ctx context.Context, jobName, jobType string, totalItems int, metadata map[string]interface{}) (*JobExecution, error) {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO woulder.job_executions (
			job_name, job_type, status, total_items, started_at, metadata
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, started_at, updated_at
	`

	job := &JobExecution{
		JobName:    jobName,
		JobType:    jobType,
		Status:     StatusRunning,
		TotalItems: totalItems,
		Metadata:   metadata,
	}

	err = m.db.QueryRowContext(
		ctx, query,
		jobName, jobType, StatusRunning, totalItems, time.Now(), metadataJSON,
	).Scan(&job.ID, &job.StartedAt, &job.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to start job: %w", err)
	}

	return job, nil
}

// UpdateProgress updates job progress
func (m *JobMonitor) UpdateProgress(ctx context.Context, jobID int64, itemsProcessed, succeeded, failed int) error {
	query := `
		UPDATE woulder.job_executions
		SET items_processed = $1,
		    items_succeeded = $2,
		    items_failed = $3,
		    updated_at = NOW()
		WHERE id = $4 AND status = $5
	`

	result, err := m.db.ExecContext(ctx, query, itemsProcessed, succeeded, failed, jobID, StatusRunning)
	if err != nil {
		return fmt.Errorf("failed to update progress: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("job not found or not running (id=%d)", jobID)
	}

	return nil
}

// UpdateMetadata updates job metadata
func (m *JobMonitor) UpdateMetadata(ctx context.Context, jobID int64, metadata map[string]interface{}) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE woulder.job_executions
		SET metadata = $1,
		    updated_at = NOW()
		WHERE id = $2
	`

	_, err = m.db.ExecContext(ctx, query, metadataJSON, jobID)
	if err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	return nil
}

// UpdateCurrentItem updates the current item being processed (route/area info)
func (m *JobMonitor) UpdateCurrentItem(ctx context.Context, jobID int64, currentItemInfo map[string]interface{}) error {
	// Get existing metadata
	var metadataJSON []byte
	query := `SELECT metadata FROM woulder.job_executions WHERE id = $1`
	err := m.db.QueryRowContext(ctx, query, jobID).Scan(&metadataJSON)
	if err != nil {
		return fmt.Errorf("failed to get existing metadata: %w", err)
	}

	// Parse existing metadata
	var metadata map[string]interface{}
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
			metadata = make(map[string]interface{})
		}
	} else {
		metadata = make(map[string]interface{})
	}

	// Merge current item info
	for k, v := range currentItemInfo {
		metadata[k] = v
	}

	// Save updated metadata
	return m.UpdateMetadata(ctx, jobID, metadata)
}

// CompleteJob marks job as completed
func (m *JobMonitor) CompleteJob(ctx context.Context, jobID int64) error {
	query := `
		UPDATE woulder.job_executions
		SET status = $1,
		    completed_at = $2
		WHERE id = $3
	`

	_, err := m.db.ExecContext(ctx, query, StatusCompleted, time.Now(), jobID)
	if err != nil {
		return fmt.Errorf("failed to complete job: %w", err)
	}

	return nil
}

// FailJob marks job as failed
func (m *JobMonitor) FailJob(ctx context.Context, jobID int64, errorMsg string) error {
	query := `
		UPDATE woulder.job_executions
		SET status = $1,
		    completed_at = $2,
		    error_message = $3
		WHERE id = $4
	`

	_, err := m.db.ExecContext(ctx, query, StatusFailed, time.Now(), errorMsg, jobID)
	if err != nil {
		return fmt.Errorf("failed to mark job as failed: %w", err)
	}

	return nil
}

// GetActiveJobs returns the most recent running job for each job_name
func (m *JobMonitor) GetActiveJobs(ctx context.Context) ([]*JobExecution, error) {
	query := `
		WITH ranked_jobs AS (
			SELECT id, job_name, job_type, status, total_items, items_processed,
			       items_succeeded, items_failed, error_message, started_at,
			       completed_at, updated_at, metadata,
			       ROW_NUMBER() OVER (PARTITION BY job_name ORDER BY started_at DESC) as rn
			FROM woulder.job_executions
			WHERE status = $1
		)
		SELECT id, job_name, job_type, status, total_items, items_processed,
		       items_succeeded, items_failed, error_message, started_at,
		       completed_at, updated_at, metadata
		FROM ranked_jobs
		WHERE rn = 1
		ORDER BY started_at DESC
	`

	rows, err := m.db.QueryContext(ctx, query, StatusRunning)
	if err != nil {
		return nil, fmt.Errorf("failed to query active jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*JobExecution
	for rows.Next() {
		job, err := scanJobExecution(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// GetJobHistory returns recent job executions for a specific job name
func (m *JobMonitor) GetJobHistory(ctx context.Context, jobName string, limit int) ([]*JobExecution, error) {
	query := `
		SELECT id, job_name, job_type, status, total_items, items_processed,
		       items_succeeded, items_failed, error_message, started_at,
		       completed_at, updated_at, metadata
		FROM woulder.job_executions
		WHERE job_name = $1
		ORDER BY started_at DESC
		LIMIT $2
	`

	rows, err := m.db.QueryContext(ctx, query, jobName, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query job history: %w", err)
	}
	defer rows.Close()

	var jobs []*JobExecution
	for rows.Next() {
		job, err := scanJobExecution(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// GetAllJobHistory returns recent job executions for all jobs
func (m *JobMonitor) GetAllJobHistory(ctx context.Context, limit int) ([]*JobExecution, error) {
	query := `
		SELECT id, job_name, job_type, status, total_items, items_processed,
		       items_succeeded, items_failed, error_message, started_at,
		       completed_at, updated_at, metadata
		FROM woulder.job_executions
		ORDER BY started_at DESC
		LIMIT $1
	`

	rows, err := m.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query all job history: %w", err)
	}
	defer rows.Close()

	var jobs []*JobExecution
	for rows.Next() {
		job, err := scanJobExecution(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// GetJobStatus returns current status of a specific job execution
func (m *JobMonitor) GetJobStatus(ctx context.Context, jobID int64) (*JobExecution, error) {
	query := `
		SELECT id, job_name, job_type, status, total_items, items_processed,
		       items_succeeded, items_failed, error_message, started_at,
		       completed_at, updated_at, metadata
		FROM woulder.job_executions
		WHERE id = $1
	`

	row := m.db.QueryRowContext(ctx, query, jobID)
	return scanJobExecution(row)
}

// GetLatestJobByName returns the most recent execution for a job name
func (m *JobMonitor) GetLatestJobByName(ctx context.Context, jobName string) (*JobExecution, error) {
	query := `
		SELECT id, job_name, job_type, status, total_items, items_processed,
		       items_succeeded, items_failed, error_message, started_at,
		       completed_at, updated_at, metadata
		FROM woulder.job_executions
		WHERE job_name = $1
		ORDER BY started_at DESC
		LIMIT 1
	`

	row := m.db.QueryRowContext(ctx, query, jobName)
	return scanJobExecution(row)
}

// scanJobExecution scans a row into a JobExecution struct
func scanJobExecution(scanner interface {
	Scan(dest ...interface{}) error
}) (*JobExecution, error) {
	job := &JobExecution{}
	var metadataJSON []byte

	err := scanner.Scan(
		&job.ID,
		&job.JobName,
		&job.JobType,
		&job.Status,
		&job.TotalItems,
		&job.ItemsProcessed,
		&job.ItemsSucceeded,
		&job.ItemsFailed,
		&job.ErrorMessage,
		&job.StartedAt,
		&job.CompletedAt,
		&job.UpdatedAt,
		&metadataJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("job not found")
		}
		return nil, fmt.Errorf("failed to scan job: %w", err)
	}

	// Parse metadata JSON
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &job.Metadata); err != nil {
			log.Printf("Warning: failed to unmarshal metadata: %v", err)
			job.Metadata = make(map[string]interface{})
		}
	} else {
		job.Metadata = make(map[string]interface{})
	}

	return job, nil
}

// ProgressReporter provides automatic progress tracking with batched updates
type ProgressReporter struct {
	monitor       *JobMonitor
	jobID         int64
	totalItems    int
	processed     int
	succeeded     int
	failed        int
	lastUpdate    time.Time
	updateEvery   int           // Update DB every N items
	updateMinTime time.Duration // Or every X seconds, whichever comes first
}

// NewProgressReporter creates a new progress reporter
func NewProgressReporter(monitor *JobMonitor, jobID int64, totalItems, updateEvery int) *ProgressReporter {
	return &ProgressReporter{
		monitor:       monitor,
		jobID:         jobID,
		totalItems:    totalItems,
		updateEvery:   updateEvery,
		updateMinTime: 5 * time.Second, // Update at least every 5 seconds
		lastUpdate:    time.Now(),
	}
}

// Increment increments progress counter and updates DB if threshold reached
func (r *ProgressReporter) Increment(ctx context.Context, success bool) error {
	r.processed++
	if success {
		r.succeeded++
	} else {
		r.failed++
	}

	// Check if we should update DB
	shouldUpdate := r.processed%r.updateEvery == 0 ||
		time.Since(r.lastUpdate) >= r.updateMinTime ||
		r.processed == r.totalItems

	if shouldUpdate {
		if err := r.FlushProgress(ctx); err != nil {
			// Log but don't fail - monitoring is non-critical
			log.Printf("Warning: failed to update job progress: %v", err)
		}
	}

	return nil
}

// SetInitialProgress sets the initial progress counters (used when resuming jobs)
func (r *ProgressReporter) SetInitialProgress(processed, succeeded, failed int) {
	r.processed = processed
	r.succeeded = succeeded
	r.failed = failed
}

// FlushProgress forces an immediate progress update to the database
func (r *ProgressReporter) FlushProgress(ctx context.Context) error {
	err := r.monitor.UpdateProgress(ctx, r.jobID, r.processed, r.succeeded, r.failed)
	if err == nil {
		r.lastUpdate = time.Now()
	}
	return err
}

// GetProgress returns current progress statistics
func (r *ProgressReporter) GetProgress() (processed, succeeded, failed int) {
	return r.processed, r.succeeded, r.failed
}

// SaveCheckpoint updates job metadata with checkpoint data
func (m *JobMonitor) SaveCheckpoint(ctx context.Context, jobID int64, checkpoint map[string]interface{}) error {
	// Get existing metadata
	var metadataJSON []byte
	query := `SELECT metadata FROM woulder.job_executions WHERE id = $1`
	err := m.db.QueryRowContext(ctx, query, jobID).Scan(&metadataJSON)
	if err != nil {
		return fmt.Errorf("failed to get metadata: %w", err)
	}

	// Parse and merge
	var metadata map[string]interface{}
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
			metadata = make(map[string]interface{})
		}
	} else {
		metadata = make(map[string]interface{})
	}

	// Update checkpoint
	metadata["checkpoint"] = checkpoint
	metadata["last_checkpoint_time"] = time.Now().Format(time.RFC3339)

	// Save back
	updatedJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	updateQuery := `
		UPDATE woulder.job_executions
		SET metadata = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err = m.db.ExecContext(ctx, updateQuery, updatedJSON, jobID)
	if err != nil {
		return fmt.Errorf("failed to update checkpoint: %w", err)
	}

	return nil
}

// GetInterruptedJob finds the most recent interrupted job for a job name
func (m *JobMonitor) GetInterruptedJob(ctx context.Context, jobName string) (*JobExecution, error) {
	query := `
		SELECT id, job_name, job_type, status, total_items, items_processed,
		       items_succeeded, items_failed, error_message, started_at,
		       completed_at, updated_at, metadata
		FROM woulder.job_executions
		WHERE job_name = $1
		  AND status IN ('running', 'paused')
		  AND started_at > NOW() - INTERVAL '24 hours'
		ORDER BY started_at DESC
		LIMIT 1
	`

	row := m.db.QueryRowContext(ctx, query, jobName)
	job, err := scanJobExecution(row)
	if err != nil {
		if err.Error() == "job not found" {
			return nil, nil // No interrupted job, this is OK
		}
		return nil, err
	}
	return job, nil
}

// MarkJobPaused marks a job as paused (for graceful shutdown)
func (m *JobMonitor) MarkJobPaused(ctx context.Context, jobID int64) error {
	query := `
		UPDATE woulder.job_executions
		SET status = 'paused', updated_at = NOW()
		WHERE id = $1 AND status = 'running'
	`
	result, err := m.db.ExecContext(ctx, query, jobID)
	if err != nil {
		return fmt.Errorf("failed to pause job: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("job not found or not running")
	}
	return nil
}

// RecoverInterruptedJobs finds all jobs interrupted in the last maxAge duration
func (m *JobMonitor) RecoverInterruptedJobs(ctx context.Context, maxAge time.Duration) ([]*JobExecution, error) {
	query := `
		SELECT id, job_name, job_type, status, total_items, items_processed,
		       items_succeeded, items_failed, error_message, started_at,
		       completed_at, updated_at, metadata
		FROM woulder.job_executions
		WHERE status IN ('running', 'paused')
		  AND started_at > $1
		ORDER BY started_at DESC
	`

	cutoff := time.Now().Add(-maxAge)
	rows, err := m.db.QueryContext(ctx, query, cutoff)
	if err != nil {
		return nil, fmt.Errorf("failed to query interrupted jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*JobExecution
	for rows.Next() {
		job, err := scanJobExecution(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// MonitorClient handles API communication
type MonitorClient struct {
	baseURL string
	client  *http.Client
}

// JobExecution represents a job execution from the API
type JobExecution struct {
	ID                        int64                  `json:"id"`
	JobName                   string                 `json:"job_name"`
	JobType                   string                 `json:"job_type"`
	Status                    string                 `json:"status"`
	TotalItems                int                    `json:"total_items"`
	ItemsProcessed            int                    `json:"items_processed"`
	ItemsSucceeded            int                    `json:"items_succeeded"`
	ItemsFailed               int                    `json:"items_failed"`
	ErrorMessage              *string                `json:"error_message"`
	StartedAt                 time.Time              `json:"started_at"`
	CompletedAt               *time.Time             `json:"completed_at"`
	ProgressPercent           float64                `json:"progress_percent"`
	ElapsedSeconds            int64                  `json:"elapsed_seconds"`
	EstimatedRemainingSeconds *int64                 `json:"estimated_remaining_seconds"`
	ItemsPerSecond            *float64               `json:"items_per_second"`
	Metadata                  map[string]interface{} `json:"metadata"`
}

// JobsSummary represents summary response
type JobsSummary struct {
	Summary map[string]*JobSummaryItem `json:"summary"`
}

// JobSummaryItem contains summary for a job
type JobSummaryItem struct {
	LastRun         *time.Time `json:"last_run"`
	Status          string     `json:"status"`
	Progress        *string    `json:"progress"`
	DurationSeconds *int64     `json:"duration_seconds"`
	NextScheduled   *time.Time `json:"next_scheduled"`
	ErrorMessage    *string    `json:"error_message"`
}

func main() {
	baseURL := os.Getenv("WOULDER_API_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	client := &MonitorClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}

	rootCmd := &cobra.Command{
		Use:   "job_monitor",
		Short: "Monitor Woulder background jobs",
		Long:  "A CLI tool to monitor Woulder's background sync jobs on remote servers",
	}

	// Active jobs command
	activeCmd := &cobra.Command{
		Use:   "active",
		Short: "Show all active (running) jobs",
		Run: func(cmd *cobra.Command, args []string) {
			client.showActiveJobs()
		},
	}

	// Watch command (real-time updates)
	watchCmd := &cobra.Command{
		Use:   "watch",
		Short: "Watch active jobs in real-time (updates every 2 seconds)",
		Run: func(cmd *cobra.Command, args []string) {
			client.watchJobs()
		},
	}

	// History command
	var historyJobName string
	var historyLimit int
	historyCmd := &cobra.Command{
		Use:   "history",
		Short: "Show job execution history",
		Run: func(cmd *cobra.Command, args []string) {
			client.showHistory(historyJobName, historyLimit)
		},
	}
	historyCmd.Flags().StringVar(&historyJobName, "job", "", "Filter by job name")
	historyCmd.Flags().IntVar(&historyLimit, "limit", 10, "Number of recent jobs to show")

	// Summary command
	summaryCmd := &cobra.Command{
		Use:   "summary",
		Short: "Show summary of all job types",
		Run: func(cmd *cobra.Command, args []string) {
			client.showSummary()
		},
	}

	// Status command for specific job
	var statusJobID int64
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show status of a specific job execution",
		Run: func(cmd *cobra.Command, args []string) {
			client.showStatus(statusJobID)
		},
	}
	statusCmd.Flags().Int64Var(&statusJobID, "id", 0, "Job execution ID")
	statusCmd.MarkFlagRequired("id")

	rootCmd.AddCommand(activeCmd, watchCmd, historyCmd, summaryCmd, statusCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (c *MonitorClient) showActiveJobs() {
	jobs, err := c.getActiveJobs()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if len(jobs) == 0 {
		fmt.Println("No active jobs running")
		return
	}

	c.printJobs(jobs)
}

func (c *MonitorClient) watchJobs() {
	fmt.Println("Watching active jobs (press Ctrl+C to stop)...")
	fmt.Println()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		// Clear screen
		fmt.Print("\033[H\033[2J")

		jobs, err := c.getActiveJobs()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}

		fmt.Printf("Active Jobs (updated %s)\n", time.Now().Format("15:04:05"))
		fmt.Println(strings.Repeat("=", 80))

		if len(jobs) == 0 {
			fmt.Println("No active jobs running")
		} else {
			c.printJobs(jobs)
		}

		<-ticker.C
	}
}

func (c *MonitorClient) showHistory(jobName string, limit int) {
	url := fmt.Sprintf("%s/api/monitoring/jobs/history?limit=%d", c.baseURL, limit)
	if jobName != "" {
		url += fmt.Sprintf("&job_name=%s", jobName)
	}

	resp, err := c.client.Get(url)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Jobs []*JobExecution `json:"jobs"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
	}

	if len(result.Jobs) == 0 {
		fmt.Println("No job history found")
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Job Name", "Status", "Started", "Duration", "Items"})
	table.SetBorder(true)

	for _, job := range result.Jobs {
		duration := formatDuration(job.ElapsedSeconds)
		items := fmt.Sprintf("%d/%d", job.ItemsProcessed, job.TotalItems)

		table.Append([]string{
			fmt.Sprintf("%d", job.ID),
			job.JobName,
			job.Status,
			job.StartedAt.Format("01-02 15:04"),
			duration,
			items,
		})
	}

	table.Render()
}

func (c *MonitorClient) showSummary() {
	url := fmt.Sprintf("%s/api/monitoring/jobs/summary", c.baseURL)

	resp, err := c.client.Get(url)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var summary JobsSummary
	if err := json.Unmarshal(body, &summary); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Job Name", "Status", "Last Run", "Duration", "Next Run"})
	table.SetBorder(true)
	table.SetRowLine(true)

	for jobName, item := range summary.Summary {
		lastRun := "Never"
		if item.LastRun != nil {
			lastRun = item.LastRun.Format("01-02 15:04")
		}

		duration := "-"
		if item.DurationSeconds != nil {
			duration = formatDuration(*item.DurationSeconds)
		}

		nextRun := "-"
		if item.NextScheduled != nil {
			nextRun = item.NextScheduled.Format("01-02 15:04")
		}

		status := item.Status
		if item.Progress != nil {
			status += "\n" + *item.Progress
		}

		table.Append([]string{
			jobName,
			status,
			lastRun,
			duration,
			nextRun,
		})
	}

	table.Render()
}

func (c *MonitorClient) showStatus(jobID int64) {
	url := fmt.Sprintf("%s/api/monitoring/jobs/%d", c.baseURL, jobID)

	resp, err := c.client.Get(url)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Job not found (ID: %d)\n", jobID)
		os.Exit(1)
	}

	body, _ := io.ReadAll(resp.Body)
	var job JobExecution
	if err := json.Unmarshal(body, &job); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Job Execution #%d\n", job.ID)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Job Name:        %s\n", job.JobName)
	fmt.Printf("Job Type:        %s\n", job.JobType)
	fmt.Printf("Status:          %s\n", job.Status)
	fmt.Printf("Progress:        %d/%d (%.1f%%)\n", job.ItemsProcessed, job.TotalItems, job.ProgressPercent)
	fmt.Printf("Success/Failed:  %d / %d\n", job.ItemsSucceeded, job.ItemsFailed)
	fmt.Printf("Started:         %s\n", job.StartedAt.Format("2006-01-02 15:04:05"))

	if job.CompletedAt != nil {
		fmt.Printf("Completed:       %s\n", job.CompletedAt.Format("2006-01-02 15:04:05"))
	}

	fmt.Printf("Elapsed:         %s\n", formatDuration(job.ElapsedSeconds))

	if job.EstimatedRemainingSeconds != nil {
		fmt.Printf("Est. Remaining:  %s\n", formatDuration(*job.EstimatedRemainingSeconds))
	}

	if job.ItemsPerSecond != nil {
		fmt.Printf("Rate:            %.2f items/sec\n", *job.ItemsPerSecond)
	}

	if job.ErrorMessage != nil {
		fmt.Printf("\nError: %s\n", *job.ErrorMessage)
	}
}

func (c *MonitorClient) getActiveJobs() ([]*JobExecution, error) {
	url := fmt.Sprintf("%s/api/monitoring/jobs/active", c.baseURL)

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Jobs []*JobExecution `json:"jobs"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result.Jobs, nil
}

func (c *MonitorClient) printJobs(jobs []*JobExecution) {
	for i, job := range jobs {
		if i > 0 {
			fmt.Println()
		}

		fmt.Printf("╔══════════════════════════════════════════════════════════════════╗\n")
		fmt.Printf("║ Job: %-59s ║\n", job.JobName)
		fmt.Printf("║ Type: %-58s ║\n", job.JobType)
		fmt.Printf("║ Progress: %d/%d (%.1f%%)%*s║\n",
			job.ItemsProcessed, job.TotalItems, job.ProgressPercent,
			40-len(fmt.Sprintf("%d/%d (%.1f%%)", job.ItemsProcessed, job.TotalItems, job.ProgressPercent)), "")

		// Progress bar
		progressBar := makeProgressBar(job.ProgressPercent, 60)
		fmt.Printf("║ %s ║\n", progressBar)

		fmt.Printf("║ Success: %-4d | Failed: %-4d%*s║\n",
			job.ItemsSucceeded, job.ItemsFailed,
			33, "")

		elapsed := formatDuration(job.ElapsedSeconds)
		remaining := "-"
		if job.EstimatedRemainingSeconds != nil {
			remaining = formatDuration(*job.EstimatedRemainingSeconds)
		}
		fmt.Printf("║ Elapsed: %-10s | Remaining: ~%-10s%*s║\n",
			elapsed, remaining, 18, "")

		if job.ItemsPerSecond != nil {
			fmt.Printf("║ Rate: %.2f items/sec%*s║\n",
				*job.ItemsPerSecond,
				45-len(fmt.Sprintf("%.2f", *job.ItemsPerSecond)), "")
		}

		fmt.Printf("╚══════════════════════════════════════════════════════════════════╝\n")
	}
}

func makeProgressBar(percent float64, width int) string {
	filled := int(percent / 100 * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	bar := "[" + strings.Repeat("█", filled) + strings.Repeat("░", empty) + "]"
	return fmt.Sprintf("%-*s", width+2, bar)
}

func formatDuration(seconds int64) string {
	duration := time.Duration(seconds) * time.Second

	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	} else if seconds < 3600 {
		return fmt.Sprintf("%dm %ds", seconds/60, seconds%60)
	} else {
		hours := seconds / 3600
		minutes := (seconds % 3600) / 60
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
}

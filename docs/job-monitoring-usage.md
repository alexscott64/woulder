# Job Monitoring System - Usage Guide

Complete guide for monitoring Woulder's background sync jobs.

## Overview

The job monitoring system provides real-time tracking of background sync jobs including:
- Priority recalculation (routes sync priority assignment)
- Location route syncs (Woulder-specific locations)
- Priority-based syncs (high/medium/low priority routes)
- Route count backfills

## Components

1. **Database Layer** - Tracks job execution state persistently
2. **Go Service Layer** - `JobMonitor` service tracks jobs programmatically
3. **REST API** - Query job status via HTTP endpoints
4. **CLI Tool** - Terminal-based monitoring for remote servers

## Setup

### 1. Run Database Migration

The monitoring system requires migration `000022`:

```bash
cd backend
# If using golang-migrate
migrate -path internal/database/migrations -database "postgresql://..." up

# Or run the migration manually
psql -U your_user -d your_db -f internal/database/migrations/000022_add_job_monitoring.up.sql
```

### 2. Install CLI Dependencies

```bash
cd backend
go get github.com/spf13/cobra@latest
go get github.com/olekukonko/tablewriter@latest
go mod tidy
```

### 3. Build CLI Tool

```bash
cd backend
go build -o job_monitor cmd/job_monitor/main.go

# Or build for specific platform
GOOS=linux GOARCH=amd64 go build -o job_monitor cmd/job_monitor/main.go
```

### 4. Deploy and Run

The monitoring system is automatically active when you start the Woulder server:

```bash
cd backend
go run cmd/server/main.go
```

## API Endpoints

All endpoints are under `/api/monitoring`:

### GET `/api/monitoring/jobs/active`

Returns all currently running jobs.

**Response:**
```json
{
  "jobs": [
    {
      "id": 1234,
      "job_name": "high_priority_tick_sync",
      "job_type": "tick_sync",
      "status": "running",
      "total_items": 2847,
      "items_processed": 1523,
      "items_succeeded": 1520,
      "items_failed": 3,
      "progress_percent": 53.5,
      "started_at": "2026-02-15T06:00:00Z",
      "elapsed_seconds": 215,
      "estimated_remaining_seconds": 186,
      "items_per_second": 7.08,
      "metadata": {
        "priority": "high"
      }
    }
  ]
}
```

### GET `/api/monitoring/jobs/history`

Returns recent job executions.

**Query Parameters:**
- `job_name` (optional) - Filter by specific job name
- `limit` (optional, default: 20) - Number of results (max 100)

**Examples:**
```bash
# Get last 20 jobs (all types)
curl http://localhost:8080/api/monitoring/jobs/history

# Get last 10 high-priority tick syncs
curl http://localhost:8080/api/monitoring/jobs/history?job_name=high_priority_tick_sync&limit=10
```

### GET `/api/monitoring/jobs/summary`

Returns summary of all job types with latest status.

**Response:**
```json
{
  "summary": {
    "high_priority_tick_sync": {
      "last_run": "2026-02-15T06:00:00Z",
      "status": "running",
      "progress": "1523/2847 (53.5%)",
      "duration_seconds": 215
    },
    "medium_priority_tick_sync": {
      "last_run": "2026-02-14T00:00:00Z",
      "status": "completed",
      "duration_seconds": 3241,
      "next_scheduled": "2026-02-21T00:00:00Z"
    }
  }
}
```

### GET `/api/monitoring/jobs/:job_id`

Returns detailed status of a specific job execution.

**Example:**
```bash
curl http://localhost:8080/api/monitoring/jobs/1234
```

## CLI Tool Usage

### Basic Commands

```bash
# Show active jobs
./job_monitor active

# Watch jobs in real-time (refreshes every 2s)
./job_monitor watch

# Show job history
./job_monitor history

# Show summary of all job types
./job_monitor summary

# Show specific job details
./job_monitor status --id 1234
```

### Remote Server Monitoring

#### Option 1: SSH Tunnel

Create a tunnel to your remote server:

```bash
# Create tunnel
ssh -L 8080:localhost:8080 user@your-server.com

# In another terminal, run monitor
export WOULDER_API_URL=http://localhost:8080
./job_monitor watch
```

#### Option 2: Direct Access

If your API is publicly accessible:

```bash
export WOULDER_API_URL=https://api.yourdomain.com
./job_monitor watch
```

### Advanced Usage

```bash
# Filter history by job name
./job_monitor history --job high_priority_tick_sync --limit 5

# Monitor in background and log to file
./job_monitor active > job_status.log 2>&1

# Create a watch script
cat > watch.sh << 'EOF'
#!/bin/bash
export WOULDER_API_URL=https://api.yourdomain.com
while true; do
    clear
    ./job_monitor summary
    sleep 10
done
EOF
chmod +x watch.sh
./watch.sh
```

## Job Types and Names

### Sync Jobs

| Job Name | Description | Frequency | Typical Duration |
|----------|-------------|-----------|------------------|
| `priority_recalculation` | Recalculates route priorities | 24h | ~45s |
| `location_tick_sync` | Syncs location route ticks | 24h | ~5-10m |
| `location_comment_sync` | Syncs location route comments | 24h | ~3-5m |
| `high_priority_tick_sync` | Syncs high-priority ticks | 24h | ~3-5m |
| `high_priority_comment_sync` | Syncs high-priority comments | 24h | ~2-3m |
| `medium_priority_tick_sync` | Syncs medium-priority ticks | 7d | ~50-85m |
| `medium_priority_comment_sync` | Syncs medium-priority comments | 7d | ~30-50m |
| `low_priority_tick_sync` | Syncs low-priority ticks | 30d | ~6-8h |
| `low_priority_comment_sync` | Syncs low-priority comments | 30d | ~4-6h |
| `route_count_backfill` | Updates route counts | 24h | ~1-2m |

## Monitoring Best Practices

### 1. Daily Monitoring

Check the summary view each morning:

```bash
./job_monitor summary
```

Look for:
- ✅ Completed jobs from previous day
- ⚠️ Failed jobs (investigate errors)
- 🔄 Running jobs (verify progress)

### 2. Long-Running Job Monitoring

For monthly low-priority syncs (6-8 hours), use watch mode:

```bash
# Watch in real-time
./job_monitor watch

# Or check periodically
watch -n 60 './job_monitor active'
```

### 3. Troubleshooting Failed Jobs

When a job fails:

```bash
# Get the job ID from history
./job_monitor history --limit 20

# View detailed error
./job_monitor status --id <job_id>
```

Common failures:
- **Mountain Project API rate limits** - Wait and retry
- **Network timeouts** - Check connection to MP API
- **Database connection errors** - Verify DB availability

### 4. Performance Monitoring

Track job performance over time:

```bash
# Compare recent executions
./job_monitor history --job high_priority_tick_sync --limit 5
```

Look for:
- Increasing duration (may need optimization)
- Higher failure rates (investigate cause)
- Growing item counts (expected as routes are added)

## Integration with Monitoring Tools

### Prometheus/Grafana Integration

Export metrics from the API:

```python
# Example Python exporter
import requests
import time
from prometheus_client import start_http_server, Gauge

job_duration = Gauge('woulder_job_duration_seconds', 'Job duration', ['job_name'])
job_items_processed = Gauge('woulder_job_items_processed', 'Items processed', ['job_name'])

def collect_metrics():
    response = requests.get('http://localhost:8080/api/monitoring/jobs/summary')
    summary = response.json()['summary']
    
    for job_name, item in summary.items():
        if item.get('duration_seconds'):
            job_duration.labels(job_name=job_name).set(item['duration_seconds'])

if __name__ == '__main__':
    start_http_server(9090)
    while True:
        collect_metrics()
        time.sleep(60)
```

### Alerting

Set up alerts for failures:

```bash
#!/bin/bash
# alert_on_failure.sh

RESPONSE=$(curl -s http://localhost:8080/api/monitoring/jobs/history?limit=5)
FAILURES=$(echo $RESPONSE | jq '.jobs[] | select(.status == "failed") | .job_name' | wc -l)

if [ $FAILURES -gt 0 ]; then
    echo "ALERT: $FAILURES failed jobs detected"
    # Send email, Slack notification, etc.
fi
```

### Logging

Aggregate job logs:

```bash
# Add to crontab for hourly summary
0 * * * * /path/to/job_monitor summary >> /var/log/woulder/jobs_$(date +\%Y\%m\%d).log
```

## Security Considerations

### Authentication

The monitoring endpoints currently have no authentication. To add auth:

1. **Modify API handlers** to require authentication:

```go
// In handlers_monitoring.go
func (h *Handler) GetActiveJobs(c *gin.Context) {
    // Add auth check
    token := c.GetHeader("Authorization")
    if !isValidToken(token) {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }
    // ... rest of handler
}
```

2. **Update CLI tool** to send auth token:

```go
// In job_monitor/main.go
func (c *MonitorClient) getActiveJobs() ([]*JobExecution, error) {
    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Set("Authorization", "Bearer "+os.Getenv("WOULDER_API_TOKEN"))
    // ...
}
```

### Firewall Rules

Restrict API access:

```bash
# Allow only from specific IPs
iptables -A INPUT -p tcp --dport 8080 -s 192.168.1.0/24 -j ACCEPT
iptables -A INPUT -p tcp --dport 8080 -j DROP
```

## Troubleshooting

### CLI Tool Issues

#### "Connection refused"
```bash
# Check if server is running
curl http://localhost:8080/api/health

# Verify correct URL
echo $WOULDER_API_URL

# Test with full URL
curl $WOULDER_API_URL/api/monitoring/jobs/active
```

#### "No jobs showing"
```bash
# Check if jobs are scheduled
# Look at server logs
tail -f /var/log/woulder/server.log | grep "Starting.*sync"

# Verify database has job_executions table
psql -U your_user -d your_db -c "\dt woulder.job_executions"
```

### API Issues

#### "404 Not Found"
- Verify endpoints are registered in [`cmd/server/main.go`](backend/cmd/server/main.go:117)
- Check Gin router configuration
- Ensure monitoring handlers are compiled

#### "500 Internal Server Error"
- Check server logs for detailed error
- Verify database migration ran successfully
- Test database connection

## Files Reference

**Core Implementation:**
- [`backend/internal/monitoring/job_monitor.go`](backend/internal/monitoring/job_monitor.go) - JobMonitor service
- [`backend/internal/api/handlers_monitoring.go`](backend/internal/api/handlers_monitoring.go) - API endpoints
- [`backend/internal/database/migrations/000022_add_job_monitoring.up.sql`](backend/internal/database/migrations/000022_add_job_monitoring.up.sql) - Database schema
- [`backend/cmd/job_monitor/main.go`](backend/cmd/job_monitor/main.go) - CLI tool
- [`backend/cmd/server/main.go`](backend/cmd/server/main.go:40) - Server initialization

**Documentation:**
- [`plans/job-monitoring-system.md`](plans/job-monitoring-system.md) - Architecture design
- [`backend/cmd/job_monitor/README.md`](backend/cmd/job_monitor/README.md) - CLI tool docs

## Next Steps

After implementing the monitoring system:

1. **Integrate with existing sync jobs** - Wrap each sync function with `JobMonitor.StartJob()`
2. **Test monitoring** - Run a sync job and verify it appears in the monitor
3. **Set up alerts** - Create scripts to notify on failures
4. **Document for team** - Share CLI tool usage with team members
5. **Optional: Build web dashboard** - Create React component for browser-based monitoring

## Support

For issues or questions:
1. Check server logs: `tail -f /var/log/woulder/server.log`
2. Test API directly: `curl http://localhost:8080/api/monitoring/jobs/active`
3. Verify database: `psql -U user -d db -c "SELECT * FROM woulder.job_executions ORDER BY started_at DESC LIMIT 5;"`

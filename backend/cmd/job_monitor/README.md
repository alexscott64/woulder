# Job Monitor CLI Tool

A command-line tool for monitoring Woulder's background sync jobs on remote servers.

## Installation

```bash
cd backend
go get github.com/spf13/cobra@latest
go get github.com/olekukonko/tablewriter@latest
go mod tidy
```

## Build

```bash
cd backend
go build -o job_monitor cmd/job_monitor/main.go
```

## Configuration

Set the API endpoint using an environment variable:

```bash
export WOULDER_API_URL=https://your-server.com
```

If not set, defaults to `http://localhost:8080`

## Usage

### Show Active Jobs

Display all currently running jobs:

```bash
./job_monitor active
```

### Watch Jobs in Real-Time

Auto-refresh every 2 seconds to show live progress:

```bash
./job_monitor watch
```

Press `Ctrl+C` to stop watching.

### Show Job History

View recent job executions:

```bash
# Show last 10 jobs of all types
./job_monitor history

# Show last 20 jobs
./job_monitor history --limit 20

# Filter by specific job name
./job_monitor history --job high_priority_tick_sync --limit 5
```

### Show Summary

View summary of all job types with their latest status:

```bash
./job_monitor summary
```

### Show Specific Job Status

Get detailed information about a specific job execution:

```bash
./job_monitor status --id 1234
```

## Example Output

### Active Jobs
```
╔══════════════════════════════════════════════════════════════════╗
║ Job: high_priority_tick_sync                                     ║
║ Type: tick_sync                                                  ║
║ Progress: 1523/2847 (53.5%)                                      ║
║ [████████████████░░░░░░░░░░░░░░░░░░] 53.5%                     ║
║ Success: 1520 | Failed: 3                                        ║
║ Elapsed: 3m 35s    | Remaining: ~3m 6s                          ║
║ Rate: 7.08 items/sec                                             ║
╚══════════════════════════════════════════════════════════════════╝
```

### Summary
```
+---------------------------+-----------+-------------+----------+-------------+
|         JOB NAME          |  STATUS   |  LAST RUN   | DURATION |  NEXT RUN   |
+---------------------------+-----------+-------------+----------+-------------+
| priority_recalculation    | completed | 02-15 00:00 | 45s      | 02-16 00:00 |
| high_priority_tick_sync   | running   | 02-15 06:00 | 3m 35s   | -           |
|                           | 1523/2847 |             |          |             |
| medium_priority_tick_sync | completed | 02-14 00:00 | 54m 1s   | 02-21 00:00 |
| low_priority_tick_sync    | completed | 02-01 00:00 | 6h 48m   | 03-03 00:00 |
+---------------------------+-----------+-------------+----------+-------------+
```

## Job Names

The following job names are tracked:

- `priority_recalculation` - Recalculates route sync priorities (daily)
- `location_tick_sync` - Syncs ticks for Woulder location routes (daily)
- `location_comment_sync` - Syncs comments for location routes (daily)
- `high_priority_tick_sync` - Syncs ticks for high-priority routes (daily)
- `high_priority_comment_sync` - Syncs comments for high-priority routes (daily)
- `medium_priority_tick_sync` - Syncs ticks for medium-priority routes (weekly)
- `medium_priority_comment_sync` - Syncs comments for medium-priority routes (weekly)
- `low_priority_tick_sync` - Syncs ticks for low-priority routes (monthly)
- `low_priority_comment_sync` - Syncs comments for low-priority routes (monthly)
- `route_count_backfill` - Updates route counts per area (daily)

## Remote Server Monitoring

### SSH Tunnel Method

If your API is not publicly accessible, create an SSH tunnel:

```bash
ssh -L 8080:localhost:8080 user@your-server.com
```

Then run the monitor locally:

```bash
export WOULDER_API_URL=http://localhost:8080
./job_monitor watch
```

### Direct Access Method

If your API is public or accessible from your network:

```bash
export WOULDER_API_URL=https://api.yourdomain.com
./job_monitor watch
```

## Troubleshooting

### Connection Refused
- Verify the server is running
- Check the `WOULDER_API_URL` is correct
- Ensure firewall allows connections

### No Jobs Showing
- Jobs may not be running at the moment
- Check `history` command to see past executions
- Verify sync jobs are scheduled in the server

### Authentication Errors
- If you add authentication to the monitoring endpoints, you'll need to modify the CLI to include auth headers

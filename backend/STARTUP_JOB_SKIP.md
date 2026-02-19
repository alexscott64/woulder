# Background Job Startup Skip Feature

## Overview

This feature prevents background jobs from running immediately on server startup during development, avoiding unnecessary load when frequently restarting the server with Air.

## How It Works

### 1. Job Completion Tracking

- Added [`WasJobCompletedRecently()`](backend/internal/monitoring/job_monitor.go:557) method to [`JobMonitor`](backend/internal/monitoring/job_monitor.go:38)
- Checks the `job_executions` table for recent successful completions
- Configurable threshold (default: 1 hour)

### 2. Startup Check Logic

- Added [`shouldRunImmediately()`](backend/internal/api/handlers.go:50) helper to [`Handler`](backend/internal/api/handlers.go:16)
- Called before each job's immediate execution
- Logs when jobs are skipped

### 3. Modified Background Jobs

All background jobs now check for recent completion before running immediately:

1. **Weather Refresh** ([`StartBackgroundRefresh`](backend/internal/api/handlers.go:68))
   - Job name: `weather_refresh`
   - Threshold: 1 hour
   
2. **Priority Recalculation** ([`StartPriorityRecalculation`](backend/internal/api/handlers.go:129))
   - Job name: `priority_recalculation`
   - Threshold: 1 hour
   
3. **Location Route Sync** ([`StartLocationRouteSync`](backend/internal/api/handlers.go:156))
   - Job name: `location_route_sync`
   - Threshold: 1 hour
   
4. **High Priority Sync** ([`StartHighPrioritySync`](backend/internal/api/handlers.go:191))
   - Job name: `high_priority_sync`
   - Threshold: 1 hour
   
5. **Medium Priority Sync** ([`StartMediumPrioritySync`](backend/internal/api/handlers.go:224))
   - Job name: `medium_priority_sync`
   - Threshold: 1 hour
   
6. **Low Priority Sync** ([`StartLowPrioritySync`](backend/internal/api/handlers.go:257))
   - Job name: `low_priority_sync`
   - Threshold: 1 hour
   
7. **Background Route Sync** ([`StartBackgroundRouteSync`](backend/internal/api/handlers.go:101))
   - Job name: `background_route_sync`
   - Threshold: 1 hour

## Behavior

### First Startup (No Recent Completions)
```
[GIN-debug] Starting priority recalculation scheduler (every 24h0m0s)
[GIN-debug] Running initial priority recalculation...
[GIN-debug] Starting location route sync scheduler (every 24h0m0s)
[GIN-debug] Starting location route sync (ticks + comments)...
```

### Subsequent Restart Within 1 Hour
```
[GIN-debug] Starting priority recalculation scheduler (every 24h0m0s)
[GIN-debug] Skipping immediate run of priority_recalculation (completed recently within 1h0m0s)
[GIN-debug] Starting location route sync scheduler (every 24h0m0s)
[GIN-debug] Skipping immediate run of location_route_sync (completed recently within 1h0m0s)
```

### Scheduled Runs (Ticker)
- **Not affected** - periodic ticker runs execute normally
- Only the immediate startup run is skipped

### Error Handling
- If checking recent completion fails, the job runs anyway (fail-safe)
- Logged as warning: `"Warning: failed to check recent completion for {job_name}: {error} (running anyway)"`

## Testing

### Manual Testing

1. **Start the server:**
   ```bash
   cd backend
   air
   ```

2. **Observe initial run:**
   - Check logs for "Running initial..." messages
   - Verify jobs execute

3. **Trigger Air reload:**
   - Modify a file (e.g., add a comment in `main.go`)
   - Air will rebuild and restart

4. **Verify skip behavior:**
   - Check logs for "Skipping immediate run..." messages
   - Confirm jobs don't re-execute

5. **Wait 1+ hour and test again:**
   - Jobs should run again after threshold expires

### Database Verification

Check recent job completions:
```sql
SELECT 
    job_name, 
    status, 
    completed_at,
    NOW() - completed_at as age
FROM woulder.job_executions
WHERE status = 'completed'
ORDER BY completed_at DESC
LIMIT 10;
```

### Adjusting Threshold

To change the skip threshold, modify the duration passed to [`shouldRunImmediately()`](backend/internal/api/handlers.go:50):

```go
// Skip if completed within last 30 minutes
if h.shouldRunImmediately(ctx, "job_name", 30*time.Minute) {
    // ...
}

// Skip if completed within last 2 hours
if h.shouldRunImmediately(ctx, "job_name", 2*time.Hour) {
    // ...
}
```

## Production Considerations

### Pros
- Reduces unnecessary load during development
- Prevents duplicate work when frequently restarting
- No impact on scheduled periodic runs

### Cons
- Could delay important updates if server restarted after job completion
- Reliant on `job_executions` table accuracy

### Recommendations

1. **Keep 1-hour threshold for development** - Good balance for Air reloads
2. **Consider shorter threshold (15-30 min) for production** - More responsive to restarts
3. **Monitor job execution frequency** - Use `/jtrack` dashboard to verify behavior
4. **Critical jobs** - May want different thresholds per job type:
   - Weather refresh: 15 minutes (more frequent updates)
   - Low priority sync: 2 hours (less critical)

## Files Modified

1. [`backend/internal/monitoring/job_monitor.go`](backend/internal/monitoring/job_monitor.go) - Added `WasJobCompletedRecently()`
2. [`backend/internal/api/handlers.go`](backend/internal/api/handlers.go) - Added helper and updated all Start* methods

## See Also

- [Job Monitoring Dashboard](http://localhost:8080/jtrack)
- [Job Execution API](http://localhost:8080/api/monitoring/jobs/history)

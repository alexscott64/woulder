# Kaya Sync - systemd Setup Guide for Debian

## What You're Setting Up

1. **Service** (`kaya-sync.service`) - Defines how to run the sync job
2. **Timer** (`kaya-sync.timer`) - Schedules when to run it (daily at 2 AM)

The service runs **incrementally** by default (only syncs locations that haven't been synced in the last 24 hours).

## Prerequisites

```bash
# Install Go (if not already installed)
sudo apt update
sudo apt install golang-go

# Verify Go installation
go version

# Install PostgreSQL (if not already)
sudo apt install postgresql postgresql-contrib
```

## Setup Instructions

### 1. Update Service File

Edit `backend/deployment/kaya-sync.service` and replace placeholders:

```bash
# Replace these values:
User=your-username              → your actual username (e.g., alex)
Group=your-username             → your actual username
WorkingDirectory=/path/to/...   → /home/alex/woulder/backend
GOPATH=/home/your-username/go   → /home/alex/go
EnvironmentFile=/path/to/...    → /home/alex/woulder/backend/.env
```

Example after editing:
```ini
User=alex
Group=alex
WorkingDirectory=/home/alex/woulder/backend
GOPATH=/home/alex/go
EnvironmentFile=/home/alex/woulder/backend/.env
```

### 2. Create Log Directory

```bash
# Create directory for logs
sudo mkdir -p /var/log/kaya-sync

# Set ownership to your user
sudo chown your-username:your-username /var/log/kaya-sync

# Example:
sudo chown alex:alex /var/log/kaya-sync
```

### 3. Install systemd Files

```bash
cd /home/your-username/woulder

# Copy service and timer to systemd directory
sudo cp backend/deployment/kaya-sync.service /etc/systemd/system/
sudo cp backend/deployment/kaya-sync.timer /etc/systemd/system/

# Set correct permissions
sudo chmod 644 /etc/systemd/system/kaya-sync.service
sudo chmod 644 /etc/systemd/system/kaya-sync.timer

# Reload systemd to recognize new files
sudo systemctl daemon-reload
```

### 4. Enable and Start Timer

```bash
# Enable timer (will start on boot)
sudo systemctl enable kaya-sync.timer

# Start timer immediately
sudo systemctl start kaya-sync.timer

# Check timer status
sudo systemctl status kaya-sync.timer
```

You should see:
```
● kaya-sync.timer - Daily Kaya Climbing Data Sync Timer
   Loaded: loaded (/etc/systemd/system/kaya-sync.timer; enabled)
   Active: active (waiting) since ...
```

### 5. Verify Setup

```bash
# List all timers (should see kaya-sync.timer)
systemctl list-timers

# Check when next sync will run
systemctl list-timers kaya-sync.timer

# Output shows:
# NEXT                         LEFT        LAST  PASSED  UNIT              ACTIVATES
# Thu 2026-02-19 02:00:00 EST  7h left     n/a   n/a     kaya-sync.timer   kaya-sync.service
```

## Manual Testing

### Test the service manually (before timer)

```bash
# Run the service once manually
sudo systemctl start kaya-sync.service

# Watch logs in real-time
tail -f /var/log/kaya-sync/sync.log

# Check if it succeeded
sudo systemctl status kaya-sync.service

# View recent logs
sudo journalctl -u kaya-sync.service -n 50
```

### Test with specific options

```bash
# Test mode (only 3 destinations)
cd /home/your-username/woulder/backend
go run cmd/sync_kaya_job/main.go --test

# Full sync (ignores incremental check)
go run cmd/sync_kaya_job/main.go --incremental=false
```

## Viewing Logs

```bash
# Real-time sync logs
tail -f /var/log/kaya-sync/sync.log

# Error logs
tail -f /var/log/kaya-sync/error.log

# systemd journal logs
sudo journalctl -u kaya-sync.service -f

# Show logs from last run
sudo journalctl -u kaya-sync.service --since "1 hour ago"
```

## Monitoring

### Check timer schedule

```bash
# When will next sync run?
systemctl list-timers kaya-sync.timer

# Timer configuration
systemctl cat kaya-sync.timer
```

### Check job results in database

```sql
-- Connect to database
psql -U postgres -d woulder

-- Check recent sync jobs
SELECT 
    job_name,
    status,
    total_items,
    items_succeeded,
    items_failed,
    started_at,
    completed_at,
    completed_at - started_at as duration
FROM woulder.job_executions
WHERE job_name LIKE 'kaya%'
ORDER BY started_at DESC
LIMIT 10;

-- Check what's been synced
SELECT 
    location_name,
    total_climbs,
    total_ascents,
    last_synced_at
FROM woulder.kaya_sync_progress
ORDER BY last_synced_at DESC
LIMIT 20;
```

## Troubleshooting

### Timer not starting

```bash
# Check timer is enabled
sudo systemctl is-enabled kaya-sync.timer

# Enable if needed
sudo systemctl enable kaya-sync.timer

# Start it
sudo systemctl start kaya-sync.timer
```

### Service failing

```bash
# Check error logs
sudo journalctl -u kaya-sync.service -n 100

# Common issues:
# 1. Wrong WorkingDirectory path
# 2. Missing .env file
# 3. Database connection issues
# 4. Go not in PATH

# Test manually to see errors
cd /home/your-username/woulder/backend
go run cmd/sync_kaya_job/main.go --test
```

### Check environment variables

```bash
# Verify .env file exists and has correct values
cat /home/your-username/woulder/backend/.env | grep DB_
```

## Maintenance Commands

```bash
# Stop timer
sudo systemctl stop kaya-sync.timer

# Disable timer (won't start on boot)
sudo systemctl disable kaya-sync.timer

# Restart timer (after config changes)
sudo systemctl restart kaya-sync.timer

# Reload systemd after editing service/timer files
sudo systemctl daemon-reload
sudo systemctl restart kaya-sync.timer

# Force run now (doesn't wait for timer)
sudo systemctl start kaya-sync.service
```

## Modifying the Schedule

To change when the sync runs, edit `/etc/systemd/system/kaya-sync.timer`:

```ini
# Change this line
OnCalendar=*-*-* 02:00:00

# Examples:
OnCalendar=*-*-* 03:30:00    # 3:30 AM daily
OnCalendar=Mon-Fri 02:00:00  # 2 AM weekdays only
OnCalendar=Sun 03:00:00       # 3 AM Sundays only
```

Then reload:
```bash
sudo systemctl daemon-reload
sudo systemctl restart kaya-sync.timer
```

## Uninstall

```bash
# Stop and disable
sudo systemctl stop kaya-sync.timer
sudo systemctl disable kaya-sync.timer

# Remove files
sudo rm /etc/systemd/system/kaya-sync.service
sudo rm /etc/systemd/system/kaya-sync.timer

# Reload systemd
sudo systemctl daemon-reload

# Remove logs (optional)
sudo rm -rf /var/log/kaya-sync
```

## Route Matching - Do I Run It Once?

### Initial Run (After Global Sync Completes)

```bash
# Run once to create all matches for existing data
cd /home/your-username/woulder/backend

# Dry run first to see what matches
go run cmd/match_kaya_mp/main.go --location Leavenworth --dry-run --limit 100

# Then run for real (all locations)
go run cmd/match_kaya_mp/main.go --min-confidence 0.85
```

### Ongoing Maintenance

**Option A: Run periodically** (recommended for new routes)
```bash
# Add to cron or create another systemd timer
# Weekly on Sundays at 3 AM
0 3 * * 0 cd /home/alex/woulder/backend && go run cmd/match_kaya_mp/main.go --min-confidence 0.85 >> /var/log/kaya-sync/matching.log 2>&1
```

**Option B: Run manually when needed**
```bash
# Run after syncing new locations
go run cmd/match_kaya_mp/main.go --location "New Location Name"
```

**Why run periodically?**
- New climbs added to Kaya need matching
- New routes added to Mountain Project need matching
- But: Existing matches are stable (ON CONFLICT DO UPDATE)
- Weekly or monthly is sufficient for most use cases

### Recommended Strategy

1. **First time** (after global sync): Run match for all locations
2. **Ongoing**: Weekly matching run via cron OR manual after big sync jobs
3. **Verification**: Check match quality quarterly and adjust confidence threshold

```sql
-- Check match coverage
SELECT 
    COUNT(DISTINCT kaya_climb_id) as matched_climbs,
    COUNT(*) as total_matches,
    AVG(match_confidence) as avg_confidence
FROM woulder.kaya_mp_route_matches;

-- Check what's not matched
SELECT COUNT(*) 
FROM woulder.kaya_climbs 
WHERE kaya_climb_id NOT IN (
    SELECT DISTINCT kaya_climb_id FROM woulder.kaya_mp_route_matches
);
```

## Summary

**systemd setup**: One-time setup, runs daily automatically at 2 AM  
**Route matching**: Run once after initial sync, then weekly/monthly for maintenance  

The sync job is incremental (only syncs what changed), so daily runs are fast after the initial global sync.

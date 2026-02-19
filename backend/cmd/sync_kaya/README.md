# Kaya Sync Tool

Command-line tool for syncing climbing data from the Kaya API into the Woulder database.

## Features

- ✅ Sync individual locations by slug
- ✅ Sync all 105 official Kaya destinations globally
- ✅ Recursive sub-location syncing
- ✅ Rate limiting with configurable delays
- ✅ Progress tracking and error recovery
- ✅ Retry logic for transient failures

## Usage

### Sync Single Location

```bash
# Sync a specific location
go run cmd/sync_kaya/main.go --slug Leavenworth-344933

# Without recursive sub-locations
go run cmd/sync_kaya/main.go --slug Bishop-316882 --recursive=false
```

### Sync All Destinations (Global Crawl)

```bash
# Sync all 105 official destinations (takes 6-8 hours)
go run cmd/sync_kaya/main.go --all

# With custom delay between destinations (recommended: 2-5 seconds)
go run cmd/sync_kaya/main.go --all --delay 3

# Non-recursive (only top-level locations)
go run cmd/sync_kaya/main.go --all --recursive=false --delay 2
```

### Test Mode

```bash
# Quick test with Leavenworth only
go run cmd/sync_kaya/main.go --test
```

## Command-Line Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--slug` | string | "" | Specific location slug to sync (e.g., 'Leavenworth-344933') |
| `--all` | bool | false | Sync all 105 official Kaya destinations from docs/kaya-destinations.txt |
| `--recursive` | bool | true | Sync sub-locations recursively |
| `--test` | bool | false | Test mode: only sync Leavenworth |
| `--delay` | int | 2 | Delay in seconds between syncing destinations (for --all mode) |
| `--token` | string | "" | Kaya API JWT token (optional, or set KAYA_AUTH_TOKEN env var) |

## Destination List

The tool loads all official Kaya destinations from [`docs/kaya-destinations.txt`](../../docs/kaya-destinations.txt), which contains 105 curated outdoor climbing destinations including:

- **USA**: Leavenworth, Bishop, Red Rocks, Joshua Tree, Hueco Tanks, Yosemite, Squamish, etc.
- **Canada**: Squamish, Vancouver Island, Kelowna, etc.
- **International**: Wadi Rum (Jordan), and more

## Examples

### Example 1: Test Sync
```bash
go run cmd/sync_kaya/main.go --test
```
Output:
```
Starting Kaya climb data sync...
TEST MODE: Only syncing Leavenworth
Processing location 1/1: Leavenworth
✓ Successfully synced Leavenworth
Sync completed: 1,553 climbs, 510 ascents
```

### Example 2: Specific Location
```bash
go run cmd/sync_kaya/main.go --slug Red-Rocks-331387
```

### Example 3: Full Global Crawl
```bash
# Run overnight during off-peak hours
nohup go run cmd/sync_kaya/main.go --all --delay 3 > sync.log 2>&1 &

# Monitor progress
tail -f sync.log
```

## How It Works

1. **Load Destinations**: Reads slugs from `docs/kaya-destinations.txt` (in --all mode)
2. **Initialize Client**: Creates standard HTTP GraphQL client with proper headers
3. **Sync Loop**: For each destination:
   - Fetches location data via GraphQL
   - Syncs climbs (pageSize: 20)
   - Syncs ascents (pageSize: 15)
   - Syncs sub-locations recursively (if enabled)
   - Updates sync progress tracking
   - Delays before next destination (rate limiting)
4. **Error Handling**: Retries transient errors, continues on failures in --all mode
5. **Progress Tracking**: Logs X of Y destinations, success/failure counts, elapsed time

## Data Synced

For each location, the tool syncs:

- **Location metadata**: Name, GPS coordinates, description, climb counts
- **Climbs/Routes**: Name, grade, rating, ascent count, photos
- **Ascents/Ticks**: User, date, rating, grade, photos, videos
- **Users**: Username, bio, height, limit grades, premium status
- **Sub-locations**: All child locations recursively

## Database Tables

Data is stored in these tables (migration 000023):

- `kaya_locations` - Climbing areas and destinations
- `kaya_climbs` - Individual boulder problems and routes
- `kaya_ascents` - User tick/ascent records
- `kaya_users` - Climber profiles
- `kaya_posts` - Social media posts
- `kaya_post_items` - Media items in posts
- `kaya_sync_progress` - Tracks last sync time per location

## Performance

### Sync Speed
- **Single location**: 2-5 minutes (e.g., Leavenworth: 5m31s for 1,553 climbs)
- **Full global crawl**: 6-8 hours for 105 destinations
- **Rate limiting**: 2-3 second delays between destinations recommended

### API Limits
- Climbs: 20 per page (API limit: "Count Limit Exceeded" at higher values)
- Ascents: 15 per page
- 1 second delay between paginated requests
- No hard rate limits observed from Kaya API

## Troubleshooting

### Error: "Count Limit Exceeded"
The Kaya API limits page sizes. This is already configured correctly (20 for climbs, 15 for ascents).

### Error: Connection timeouts
Use `--delay` flag to increase delays between destinations:
```bash
go run cmd/sync_kaya/main.go --all --delay 5
```

### Error: Database connection failed
Ensure PostgreSQL is running and `.env` is configured:
```bash
cd backend
cat .env
# Should have DATABASE_URL set
```

### Check sync progress
```sql
SELECT 
    location_name, 
    total_climbs, 
    total_ascents, 
    last_synced_at 
FROM kaya_sync_progress 
ORDER BY last_synced_at DESC;
```

## Environment Variables

```bash
# Optional: Kaya API authentication (not required for public data)
KAYA_AUTH_TOKEN=your_jwt_token_here

# Required: Database connection
DATABASE_URL=postgresql://user:pass@localhost:5432/woulder
```

## Next Steps

After syncing Kaya data:

1. **Run Route Matching**: Link Kaya climbs to Mountain Project routes
   ```bash
   # TODO: Implement matching CLI
   go run cmd/match_kaya_mp/main.go --location Leavenworth
   ```

2. **Set up Scheduled Sync**: Run daily incremental syncs
   ```bash
   # TODO: Implement scheduled job
   go run cmd/sync_kaya_job/main.go
   ```

3. **Expose via API**: Add Kaya endpoints to Woulder API
   - `/api/kaya/locations`
   - `/api/kaya/climbs`
   - `/api/kaya/matches` (Kaya ↔ MP route matches)

## Related Documentation

- [Kaya Implementation Summary](../../docs/KAYA_IMPLEMENTATION_SUMMARY.md) - Complete overview
- [Kaya Global Sync Plan](../../docs/KAYA_GLOBAL_SYNC_PLAN.md) - Detailed plan
- [Kaya Destinations List](../../docs/kaya-destinations.txt) - All 105 destinations
- [Kaya Context Summary](../../docs/KAYA_CONTEXT_SUMMARY.md) - API fields reference

## Permission

This sync tool is used with express permission from the Kaya founders to integrate their public outdoor climbing data into the Woulder platform.

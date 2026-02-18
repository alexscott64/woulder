# Kaya Sync CLI Tool

Command-line tool for syncing climbing data from Kaya API to Woulder database.

## Usage

### Test Mode (Leavenworth only)
```bash
cd backend
go run cmd/sync_kaya/main.go -test
```

### Sync Specific Location
```bash
go run cmd/sync_kaya/main.go -slug "Leavenworth-344933"
```

### Sync All Configured Locations
```bash
go run cmd/sync_kaya/main.go
```

### Sync Without Sub-Locations
```bash
go run cmd/sync_kaya/main.go -slug "Leavenworth-344933" -recursive=false
```

## Flags

- `-slug`: Specific location slug to sync (e.g., "Leavenworth-344933")
- `-recursive`: Sync sub-locations recursively (default: true)
- `-test`: Test mode - only sync Leavenworth

## Examples

### Test with Leavenworth
```bash
# Quick test to validate the integration
go run cmd/sync_kaya/main.go -test
```

### Sync a new location
```bash
# Find the slug from Kaya URL: https://kaya-beta.kayaclimb.com/location/Icicle-Canyon-345011
go run cmd/sync_kaya/main.go -slug "Icicle-Canyon-345011"
```

## Data Synced

For each location, the tool syncs:
1. **Location metadata** - name, GPS, descriptions, counts
2. **Climbs** - all routes and boulders with grades and ratings
3. **Ascents** - recent user tick logs (limited to 500 per location)
4. **Users** - user profiles referenced by ascents
5. **Sub-locations** - child areas (if recursive=true)

## Rate Limiting

The Kaya client implements a 1-second delay between requests to be respectful of their API.

## Error Handling

- Transient errors (timeouts, 5xx) trigger automatic retry after 5 seconds
- Partial failures log warnings but continue processing
- Sync progress is tracked in `kaya_sync_progress` table

## Notes

- First run will be slow as it populates the database
- Subsequent runs update existing records
- Check logs for detailed progress and any warnings

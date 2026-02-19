# Kaya Location Mapping Fix - Using Route Matching Table

## Problem

Two critical issues were discovered with the Kaya integration:

1. **Incorrect Area Mapping**: When drilling down into MP sub-areas (e.g., "Zelda Boulders"), Kaya climbs from different areas (e.g., "Sasquatch Boulders") were appearing incorrectly. This was because Kaya climbs were being fetched for the entire Woulder location regardless of which MP sub-area was selected.

2. **Heat Map 500 Errors**: The heat map was throwing 500 errors due to complex FULL OUTER JOIN queries attempting to merge MP and Kaya data.

## Root Cause

### Issue 1: Area Mapping
The `GetUnifiedRoutesOrderedByActivity` endpoint (`/api/climbs/location/:id/areas/:area_id/unified-routes`) was:
- Correctly filtering MP routes by the specific `area_id` parameter
- BUT incorrectly fetching ALL Kaya climbs for the entire location via `GetClimbsOrderedByActivityForWoulderLocation(locationID)`

This caused Kaya climbs to appear in every MP sub-area because the query ignored the `area_id` parameter completely.

**The solution was obvious once pointed out**: We have a `kaya_mp_route_matches` table that links Kaya climbs to specific MP routes! This table was created specifically for this purpose but wasn't being used in the unified routes endpoint.

### Issue 2: Heat Map Errors
The heat map queries were using complex CTEs with FULL OUTER JOINs that were failing due to incomplete foreign key relationships in the Kaya data.

## Solution

### Issue 1: Use Route Matching Table
**Files Modified**:
- [`backend/internal/database/kaya/queries.go`](backend/internal/database/kaya/queries.go:277) - Added `queryGetMatchedClimbsForArea`
- [`backend/internal/database/kaya/repository.go`](backend/internal/database/kaya/repository.go:89) - Added `GetMatchedClimbsForArea()` method to interface
- [`backend/internal/database/kaya/postgres.go`](backend/internal/database/kaya/postgres.go:534) - Implemented `GetMatchedClimbsForArea()`
- [`backend/internal/api/handlers_climb_tracking.go`](backend/internal/api/handlers_climb_tracking.go:221) - Updated handler to use matched climbs

**New Query** (`queryGetMatchedClimbsForArea`):
```sql
WITH matched_routes AS (
    SELECT DISTINCT
        m.kaya_climb_id,
        m.mp_route_id,
        m.match_confidence
    FROM kaya_mp_route_matches m
    JOIN woulder.mp_routes r ON m.mp_route_id = r.mp_route_id
    WHERE r.mp_area_id = $1
        AND m.match_confidence >= 0.60
)
-- ... rest of query joins with kaya_climbs and kaya_ascents
```

**How It Works**:
1. Queries `kaya_mp_route_matches` table for Kaya climbs matched to MP routes in the specific area
2. Only returns Kaya climbs with match confidence ≥ 60%
3. Joins with ascent data to show recent activity
4. Returns `UnifiedRouteActivitySummary` with both Kaya slug AND MP route ID

**Result**: 
- Kaya climbs now appear ONLY in the MP areas they've been matched to
- No more cross-contamination between different areas
- Properly leverages the route matching intelligence we built

### Issue 2: Simplified Heat Map Queries
**File**: [`backend/internal/database/heatmap/queries.go`](backend/internal/database/heatmap/queries.go:11)

Reverted both `queryHeatMapDataLightweight` and `queryHeatMapDataFull` to simple MP-only queries, removing Kaya CTEs and FULL OUTER JOINs.

**Result**:
- Heat map loads successfully without 500 errors
- Shows only MP activity data (which is the primary data source)
- Can re-add Kaya integration later once all location mappings are validated

### Issue 3: Removed Debug Logging
Removed `[FORECAST DEBUG]` console.log statements from [`RouteListItem.tsx`](frontend/src/components/RouteListItem.tsx:50-57).

## What Now Works

✅ Kaya climbs appear CORRECTLY in "By Area" drill-down (only in matched areas)
✅ Kaya climbs still appear in Recent Activity modal (location-wide view)
✅ Kaya climb performance optimization maintained (from previous N+1 fix)
✅ Kaya sync job and route matching CLI tools working
✅ Heat map displays MP data correctly without errors
✅ Clean console without debug spam

## Prerequisites for Kaya to Appear

For Kaya climbs to show up in the "By Area" view, you must:

1. **Run the global Kaya sync** to populate `kaya_climbs` and `kaya_ascents`:
   ```bash
   cd backend
   go run cmd/sync_kaya/main.go --all
   ```

2. **Run the route matching CLI** to populate `kaya_mp_route_matches`:
   ```bash
   go run cmd/match_kaya_mp/main.go --location "Leavenworth" --confidence 0.60
   ```

Without step 2, no Kaya climbs will appear because the `GetMatchedClimbsForArea()` query requires entries in the matching table.

## Files Modified

1. `backend/internal/database/kaya/queries.go` (added `queryGetMatchedClimbsForArea`)
2. `backend/internal/database/kaya/repository.go` (added `GetMatchedClimbsForArea` to interface)
3. `backend/internal/database/kaya/postgres.go` (implemented `GetMatchedClimbsForArea`)
4. `backend/internal/api/handlers_climb_tracking.go` (updated to use matched climbs)
5. `backend/internal/database/heatmap/queries.go` (simplified to MP-only)
6. `frontend/src/components/RouteListItem.tsx` (removed debug logs)

## How Route Matching Works

The `kaya_mp_route_matches` table links Kaya climbs to MP routes using:

- **Name similarity** (Levenshtein distance, 70% weight)
- **Location name matching** (20% weight)
- **GPS proximity** (Haversine distance, 10% weight)

Matches are assigned:
- `match_confidence`: 0.0 to 1.0 score
- `match_type`: 'exact_name', 'fuzzy_name', 'location_name', 'gps_proximity', 'manual'

The new query filters for `match_confidence >= 0.60` to ensure quality matches.

## Testing

To verify the fixes:

1. **Sync Kaya data**: Run `cmd/sync_kaya/main.go --all`
2. **Match routes**: Run `cmd/match_kaya_mp/main.go --location "Leavenworth" --confidence 0.60`
3. **Navigate to area drill-down**: Select "Zelda Boulders" sub-area
4. **Verify correct climbs**: Only Kaya climbs matched to routes in that specific area should appear
5. **Check Recent Activity**: Kaya climbs still appear with purple "Kaya" badges
6. **Test heat map**: Verify no 500 errors

## Related Documentation

- [`KAYA_PERFORMANCE_FIX.md`](KAYA_PERFORMANCE_FIX.md) - Previous optimization of N+1 queries
- [`backend/STARTUP_JOB_SKIP.md`](backend/STARTUP_JOB_SKIP.md) - Job monitoring integration
- [`docs/KAYA_TOOLS_GUIDE.md`](docs/KAYA_TOOLS_GUIDE.md) - Kaya sync and route matching tools
- [`backend/internal/database/migrations/000024_add_kaya_mp_route_matching.up.sql`](backend/internal/database/migrations/000024_add_kaya_mp_route_matching.up.sql) - Route matching table schema

# Boulder Drying Performance Optimization

This document explains the caching strategy and performance optimizations for the boulder drying system.

## Caching Strategy

### What We Cache (Expensive, Rarely Changes)

1. **Tree Coverage** (`boulder_drying_profiles.tree_coverage_percent`)
   - **Source**: Google Earth Engine API (satellite data)
   - **Cost**: 3-5 seconds per boulder
   - **How Often**: Changes very rarely (years)
   - **Sync Method**: Manual `cmd/sync_tree_cover` tool
   - **Why Cache**: External API is extremely slow, data is static

2. **GPS Coordinates** (`mp_routes.latitude`, `mp_routes.longitude`, `mp_routes.aspect`)
   - **Source**: Mountain Project API + circular distribution algorithm
   - **Cost**: Fast API calls + geometry calculations
   - **How Often**: Changes when new routes added
   - **Sync Method**: Automatic during Mountain Project sync
   - **Why Cache**: Boulder positions don't move

### What We Calculate Fresh (Fast, Time-Dependent)

1. **Sun Exposure** (NOT cached)
   - **Source**: Offline astronomical calculations in Go
   - **Cost**: <10ms per boulder (144 hours calculated)
   - **How Often**: Changes every hour (time-dependent)
   - **Algorithm**: Pure Go math (sun position from GPS + date/time)
   - **Why NOT Cache**: "Next 6 days from NOW" changes constantly, calculation is fast

2. **Weather Data**
   - **Current Weather**: Always fetched fresh from Open-Meteo API
   - **Historical Weather**: Cached in database (refreshed every 10 minutes)
   - **Forecast Weather**: Cached in database (refreshed every 10 minutes)
   - **Why**: Weather changes frequently, API is fast

3. **Drying Calculations**
   - **Source**: Rock drying module calculations
   - **Cost**: <50ms per boulder (includes 6-day forecast)
   - **How Often**: Depends on weather + time
   - **Why NOT Cache**: Weather-dependent, time-dependent

## Performance Optimization History

### Problem: Extremely Slow Performance

**Initial Issues:**
1. Individual API calls per route (N+1 problem)
2. Google Earth Engine API calls during requests (3-5s per boulder)
3. Sun exposure caching (causing stale data)
4. Profile saving during batch requests (database writes)
5. Layout shift during loading (UX issue)

### Solution: Multi-Phase Optimization

#### Phase 1: Batch Endpoints
- Created `GetBatchBoulderDryingStatus` to fetch multiple routes in one call
- Created `GetBatchAreaDryingStats` to fetch multiple areas in one call
- Grouped routes by location to share expensive calculations
- **Result**: Eliminated N+1 API calls

#### Phase 2: Remove External API Calls
- Changed tree coverage to ONLY use cached data from profiles
- Removed Google Earth Engine calls during requests
- Pre-populate via background job (`cmd/sync_tree_cover`)
- **Result**: Eliminated 3-5s per boulder API latency

#### Phase 3: Remove Sun Exposure Caching
- Removed `sun_exposure_hours_cache` field and caching logic
- Always calculate fresh using fast offline algorithm
- **Result**: Eliminated stale data, <10ms per calculation

#### Phase 4: Remove Profile Saves
- Disabled profile saving during batch requests
- Profiles pre-populated by background job
- **Result**: Eliminated database write latency

#### Phase 5: Skeleton Loaders
- Added fixed-dimension skeleton loaders (60px × 24px for routes, 80px height for areas)
- Block route rendering until batch data loads
- Show areas immediately with skeleton for drying stats
- **Result**: Eliminated layout shift, better perceived performance

#### Phase 6: Cache Synchronization
- Synchronized all React Query caches to 2min staleTime
- Added refetchOnMount and refetchOnWindowFocus
- Use batch data in area stats calculation
- **Result**: Eliminated area/route data mismatch

#### Phase 7: Performance Logging
- Added comprehensive `[PERF]` logs throughout batch operations
- Track timing for each database query, API call, and calculation
- Identify actual bottlenecks with real data
- **Result**: Can measure and optimize actual slow operations

## Performance Targets

### Current Performance (After Optimizations)
- **Single Route**: <100ms
- **Batch 10 Routes (same location)**: <200ms
- **Batch 50 Routes (same location)**: <500ms
- **Batch 100 Routes (5 locations)**: <1000ms
- **Area Stats (20 routes)**: <300ms

### Bottlenecks (In Order)
1. **Weather API Calls** - 200-500ms per location (only fetched once per location)
2. **Database Queries** - 10-50ms per query (historical weather, forecasts, profiles)
3. **Drying Calculations** - 5-10ms per route
4. **Sun Exposure** - <10ms per route
5. **Forecast Generation** - <50ms per route

## Background Jobs

### sync_tree_cover Tool

**Purpose**: Pre-populate tree coverage cache from Google Earth Engine

**Usage**:
```bash
cd backend
go run cmd/sync_tree_cover/main.go [--force]
```

**Flags**:
- `--force`: Re-sync routes that already have tree coverage

**Smart Caching**:
- Checks if coverage exists, skips if present (unless --force)
- Upserts data (ON CONFLICT DO UPDATE)
- Rate limiting: Pauses every 50 routes

**When to Run**:
- After Mountain Project sync (new routes added)
- Manually when tree coverage data is missing
- Optionally via cron (monthly is sufficient)

**Example Cron**:
```bash
# Run monthly at 3 AM
0 3 1 * * cd /path/to/woulder/backend && go run cmd/sync_tree_cover/main.go
```

## Monitoring Performance

### Performance Logs

All batch operations log timing with `[PERF]` prefix:

```
[PERF] GetBatchBoulderDryingStatus: Starting batch request for 20 routes
[PERF] Route fetching took 15ms for 20 routes
[PERF] Processing location 1 with 20 routes
[PERF]   GetLocation took 2ms
[PERF]   GetCurrentAndForecast (API) took 450ms   <-- SLOWEST
[PERF]   GetHistoricalWeather took 25ms (got 168 hours)
[PERF]   GetRockTypesByLocation took 3ms
[PERF]   GetSunExposureByLocation took 2ms
[PERF]   CalculateDryingStatus took 5ms
[PERF] getLocationRockDryingStatus took 487ms
[PERF] GetSunExposureByLocation took 2ms
[PERF] GetForecastWeather took 30ms (got 144 hours)
[PERF] Route 123: profile=5ms calc=8ms total=13ms
[PERF] Route 456: profile=4ms calc=7ms total=11ms
...
[PERF] Location 1 processing took 550ms
[PERF] GetBatchBoulderDryingStatus: TOTAL TIME 565ms for 20 routes (20 successful, 0 failed)
```

### Identifying Bottlenecks

1. **Check total time**: `TOTAL TIME` log at end
2. **Find slow operations**: Look for operations >100ms
3. **Weather API calls**: Should be 200-500ms per location (only once)
4. **Database queries**: Should be <50ms each
5. **Calculations**: Should be <10ms per route

### Common Issues

**Slow Weather API** (>1000ms):
- Open-Meteo API might be overloaded
- Network latency
- Solution: Add Redis caching for weather API responses

**Slow Database Queries** (>100ms):
- Missing indexes
- Too much historical data
- Solution: Add indexes, clean old weather data

**Slow Calculations** (>100ms per route):
- Bug in calculation logic
- Solution: Profile Go code with pprof

## Future Optimizations (If Needed)

If performance is still not acceptable, consider:

1. **Redis Caching for Weather API**
   - Cache weather API responses (5-10 min TTL)
   - Reduces 450ms per location to <10ms

2. **Database Indexes**
   - Add indexes on `mp_routes.location_id`, `weather_data.timestamp`
   - Speeds up queries by 10-100x

3. **Connection Pooling**
   - Increase database connection pool size
   - Reduces connection overhead

4. **Goroutines for Parallel Processing**
   - Process routes in parallel with goroutines
   - Could reduce batch time by 50-70%

5. **Denormalization**
   - Store pre-computed area stats in database
   - Update via background job every 10 minutes
   - Trades accuracy for speed

## Testing Performance

### Load Test Script

```bash
# Test single route
time curl http://localhost:8080/api/climbs/routes/123/drying-status

# Test batch 10 routes
time curl "http://localhost:8080/api/climbs/batch-boulder-drying-status?route_ids=1,2,3,4,5,6,7,8,9,10"

# Test area stats
time curl http://localhost:8080/api/climbs/location/1/area/456/drying-stats
```

### Benchmark in Go

```go
func BenchmarkBatchBoulderDryingStatus(b *testing.B) {
    service := NewBoulderDryingService(repo, weatherClient)
    routeIDs := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := service.GetBatchBoulderDryingStatus(context.Background(), routeIDs)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## Summary

**Cache expensive, static data** (tree coverage, GPS) via background jobs.

**Calculate cheap, dynamic data** (sun exposure, drying times) on-demand.

**Batch operations** to eliminate N+1 API calls.

**Monitor with logs** to identify real bottlenecks.

**Target: Sub-second response** for typical batch requests (20-50 routes).

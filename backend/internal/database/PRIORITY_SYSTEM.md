# Hybrid Multi-Signal Priority System for Route Syncing

## Overview

The priority system uses multiple signals to intelligently determine which Mountain Project routes need daily, weekly, or monthly syncing. This optimizes API usage while ensuring we never miss important climbing activity, especially for seasonal routes and remote areas.

## Problem Statement

**Challenge**: With 284,000+ routes, syncing all routes daily would require 568,000+ API calls (ticks + comments) taking 8+ hours.

**Goal**: Reduce API calls by 90%+ while maintaining fresh data for:
- Seasonal routes (ice climbing, alpine) during their active seasons
- Remote areas with lower population but important local activity
- Busy routes with high activity
- New routes that might become classics

## Solution: Hybrid Multi-Signal System

Instead of a single threshold, we use **5 complementary signals** evaluated in priority order:

### Signal 1: Seasonal Route Type (Highest Priority)
```sql
WHEN route_type IN ('Ice', 'Alpine', 'Snow', 'Mixed') AND tick_count_90d >= 1 THEN 'high'
```

**Purpose**: Catch seasonal climbing activity
**Why**: Ice/alpine routes only get ticks during specific seasons (winter/summer). A single tick indicates the season has started.
**Result**: ~100-500 routes marked HIGH during their seasons

**Example**:
- December: Ice climbing routes in Colorado get their first ticks → HIGH
- June: Alpine routes in Cascades become accessible → HIGH

### Signal 2: Activity Surge Detection
```sql
WHEN tick_count_14d >= 1 AND days_since_last_tick > 90 THEN 'high'
```

**Purpose**: Detect when ANY route goes from inactive → active
**Why**: Catches season starts for ALL route types, not just ice/alpine
**Result**: ~200-400 routes marked HIGH at season transitions

**Example**:
- Spring: Desert sport climbing routes get first ticks after winter → HIGH
- Fall: High-elevation routes get last activity before snow → HIGH

### Signal 3: Per-Area Percentile Ranking
```sql
WHEN area_percentile >= 0.90 THEN 'high'
```

**Purpose**: Adjust for population differences between areas
**Why**: Top routes in remote areas deserve same priority as top routes in busy areas
**Result**: ~10% of routes per area marked HIGH (population-adjusted)

**Example**:
- Busy area (Yosemite): Route with 50 ticks/90d → top 10% → HIGH
- Remote area (Idaho): Route with 3 ticks/90d → top 10% → HIGH

### Signal 4: Absolute Threshold (Safety Net)
```sql
WHEN tick_count_90d >= 20 THEN 'high'
```

**Purpose**: Catch extremely busy routes regardless of area percentile
**Why**: Very active routes should always stay fresh
**Result**: ~1,000-1,500 routes marked HIGH

**Example**:
- Classic boulder problem: 50 ticks in 90 days → HIGH (even if not top 10% in very busy area)

### Signal 5: New Route Discovery
```sql
WHEN route_age_days < 90 AND total_tick_count > 0 THEN 'high'
```

**Purpose**: Catch new routes that might become classics
**Why**: New routes with early activity often become popular
**Result**: ~50-100 new routes marked HIGH

## Medium Priority
```sql
WHEN tick_count_90d >= 1 OR area_percentile >= 0.50 THEN 'medium'
```

**Synced**: Weekly (7 days)
**Routes**: Any activity or above-average for area
**API calls**: ~30,000-50,000 routes × 2 calls = 60k-100k calls/week

## Low Priority
```sql
ELSE 'low'
```

**Synced**: Monthly (30 days)
**Routes**: No recent activity and below-average
**API calls**: ~200,000+ routes × 2 calls = 400k+ calls/month

## Performance Metrics

### Expected Distribution (after full recalculation)
- **HIGH**: ~2,000-3,000 routes (1%)
  - Daily API calls: ~4,000-6,000
  - Sync time: ~3-5 minutes

- **MEDIUM**: ~30,000-50,000 routes (15-20%)
  - Weekly API calls: ~60,000-100,000
  - Sync time: ~50-85 minutes

- **LOW**: ~230,000+ routes (80%)
  - Monthly API calls: ~460,000+
  - Sync time: ~6-8 hours

### Total API Call Reduction
- **Before**: 568,000 calls/day (8+ hours)
- **After**: ~5,000 calls/day + ~9,000/week + ~15,000/month
  - Daily average: ~7,000 calls/day
  - **Savings**: 98.7% reduction

## Implementation Details

### Database Columns

```sql
-- Migration 000020 added:
tick_count_14d INTEGER          -- Ticks in last 14 days (surge detection)
area_percentile NUMERIC(5,4)    -- Percentile within area (0.0-1.0)

-- Existing from migration 000019:
tick_count_90d INTEGER          -- Ticks in last 90 days
total_tick_count INTEGER        -- Lifetime ticks
days_since_last_tick INTEGER    -- Days since most recent tick
sync_priority VARCHAR(10)       -- 'high', 'medium', 'low'
```

### Query Structure

The priority calculation uses a 3-CTE structure:

1. **route_metrics**: Calculates tick counts and time metrics per route
2. **area_percentiles**: Calculates percentile ranking within each area
3. **UPDATE**: Joins metrics + percentiles, applies multi-signal logic

### Indexes

```sql
-- For percentile calculation (migration 000020)
CREATE INDEX idx_mp_routes_area_activity ON mp_routes(mp_area_id, tick_count_90d);

-- For sync queries (migration 000019)
CREATE INDEX idx_mp_routes_sync_priority ON mp_routes(sync_priority, last_tick_sync_at);
CREATE INDEX idx_mp_routes_location_priority ON mp_routes(location_id, sync_priority);
```

## Testing the System

### Query to See Distribution
```sql
SELECT
    sync_priority,
    COUNT(*) as route_count,
    ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER (), 2) as percentage
FROM woulder.mp_routes
WHERE location_id IS NULL
GROUP BY sync_priority
ORDER BY
    CASE sync_priority
        WHEN 'high' THEN 1
        WHEN 'medium' THEN 2
        WHEN 'low' THEN 3
    END;
```

### Query to See Seasonal Routes
```sql
SELECT
    route_type,
    sync_priority,
    COUNT(*) as route_count
FROM woulder.mp_routes
WHERE location_id IS NULL
    AND route_type IN ('Ice', 'Alpine', 'Snow', 'Mixed')
GROUP BY route_type, sync_priority
ORDER BY route_type, sync_priority;
```

### Query to See Activity Surges
```sql
SELECT
    mp_route_id,
    route_type,
    tick_count_14d,
    tick_count_90d,
    days_since_last_tick,
    sync_priority
FROM woulder.mp_routes
WHERE location_id IS NULL
    AND tick_count_14d >= 1
    AND days_since_last_tick > 90
ORDER BY tick_count_14d DESC
LIMIT 20;
```

## Edge Cases Handled

### 1. Ice Climbing Season Start (December)
- **Before**: Ice routes at LOW priority (0 ticks in 90 days)
- **After**: First tick → HIGH priority immediately
- **Result**: Catch season start within 24 hours

### 2. Remote Area with Low Activity
- **Before**: Route with 2 ticks → LOW (didn't meet absolute threshold)
- **After**: Route in top 10% of remote area → HIGH
- **Result**: Local "classics" stay fresh

### 3. Very Busy Area (Yosemite, Red Rocks)
- **Before**: All popular routes synced daily (inefficient)
- **After**: Only top 10% + 20+ tick routes synced daily
- **Result**: 90% of routes drop to weekly/monthly

### 4. Spring Thaw (Seasonal Transition)
- **Before**: Missed first 1-2 weeks of activity
- **After**: Activity surge detection catches first ticks → HIGH
- **Result**: Catch season transitions across all route types

## Maintenance

### Daily (Automatic)
- Priority recalculation runs on server startup
- High-priority sync runs immediately after

### Monitoring Queries

**Check if priorities need adjustment**:
```sql
-- Should be roughly: 1-2% high, 15-20% medium, 80% low
SELECT
    sync_priority,
    COUNT(*),
    ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER (), 2) as pct
FROM woulder.mp_routes
WHERE location_id IS NULL
GROUP BY sync_priority;
```

**Check seasonal route distribution**:
```sql
-- Ice routes should spike to HIGH in winter
SELECT
    EXTRACT(MONTH FROM NOW()) as current_month,
    route_type,
    sync_priority,
    COUNT(*)
FROM woulder.mp_routes
WHERE route_type IN ('Ice', 'Alpine', 'Snow', 'Mixed')
    AND location_id IS NULL
GROUP BY route_type, sync_priority
ORDER BY route_type, sync_priority;
```

## Future Enhancements

### Potential Additions
1. **Historical trending**: Routes gaining popularity over time → boost priority
2. **Grade-based adjustment**: Hard grades (V10+, 5.13+) get lower thresholds
3. **Regional seasonal calendars**: Area-specific season definitions
4. **Weather-triggered boosts**: After good weather, boost nearby routes

### Configuration
Currently thresholds are hardcoded. Future: Make configurable via environment variables or database table.

## Related Files

- **Migration**: `000020_add_advanced_priority_metrics.up.sql`
- **Implementation**: `backend/internal/database/mountain_project.go` (UpdateRouteSyncPriorities)
- **Service**: `backend/internal/service/climb_tracking_service.go` (RecalculateAllPriorities)
- **Scheduler**: `backend/cmd/server/main.go` (StartPriorityRecalculation)

## Summary

The hybrid multi-signal system adapts to:
- ✅ Seasonal patterns (ice/alpine)
- ✅ Activity surges (season starts)
- ✅ Population differences (per-area ranking)
- ✅ Absolute activity (busy routes)
- ✅ New route discovery

**Result**: 98.7% reduction in daily API calls while maintaining fresh data for all important routes.

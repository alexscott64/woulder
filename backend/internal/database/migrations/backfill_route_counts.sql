-- Backfill script to populate initial route_count_total values
-- Run this ONCE after migration 000018 to initialize route counts from existing data

-- Step 1: For all areas, count routes recursively including descendants
-- Uses a recursive CTE to traverse the area tree and aggregate route counts

WITH RECURSIVE area_tree AS (
    -- Base case: Start with all areas
    SELECT
        mp_area_id,
        parent_mp_area_id,
        mp_area_id as original_area_id
    FROM woulder.mp_areas

    UNION ALL

    -- Recursive case: Get all descendants
    SELECT
        a.mp_area_id,
        a.parent_mp_area_id,
        at.original_area_id
    FROM woulder.mp_areas a
    INNER JOIN area_tree at ON a.parent_mp_area_id = at.mp_area_id
),
route_counts AS (
    -- Count routes for each area (including all descendants)
    SELECT
        at.original_area_id as mp_area_id,
        COUNT(DISTINCT r.mp_route_id) as total_routes
    FROM area_tree at
    LEFT JOIN woulder.mp_routes r ON r.mp_area_id = at.mp_area_id
    GROUP BY at.original_area_id
)
UPDATE woulder.mp_areas a
SET
    route_count_total = COALESCE(rc.total_routes, 0),
    route_count_last_checked = NOW()
FROM route_counts rc
WHERE a.mp_area_id = rc.mp_area_id;

-- Step 2: Verify the backfill
-- Show summary by root areas (states)
SELECT
    a.mp_area_id,
    a.name,
    a.route_count_total,
    a.route_count_last_checked
FROM woulder.mp_areas a
WHERE a.parent_mp_area_id IS NULL
ORDER BY a.route_count_total DESC;

-- Step 3: Show overall statistics
SELECT
    COUNT(*) as total_areas,
    COUNT(CASE WHEN route_count_total > 0 THEN 1 END) as areas_with_routes,
    SUM(route_count_total) as total_route_count,
    MAX(route_count_total) as max_routes_in_area,
    AVG(route_count_total) as avg_routes_per_area
FROM woulder.mp_areas
WHERE route_count_total IS NOT NULL;

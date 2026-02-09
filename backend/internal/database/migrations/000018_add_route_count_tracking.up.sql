-- Add route count tracking columns to mp_areas table for efficient change detection
ALTER TABLE woulder.mp_areas
ADD COLUMN route_count_total INTEGER,
ADD COLUMN route_count_last_checked TIMESTAMPTZ;

-- Add index for efficient route count queries
CREATE INDEX idx_mp_areas_route_count ON woulder.mp_areas(route_count_total);

-- Add comment for documentation
COMMENT ON COLUMN woulder.mp_areas.route_count_total IS 'Total number of routes under this area (including all descendants), cached from Mountain Project API';
COMMENT ON COLUMN woulder.mp_areas.route_count_last_checked IS 'Timestamp when route_count_total was last checked against MP API';

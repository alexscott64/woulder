-- Add index on route_type for efficient filtering
CREATE INDEX IF NOT EXISTS idx_mp_routes_route_type ON woulder.mp_routes(route_type) WHERE route_type IS NOT NULL;

-- Add comment for documentation
COMMENT ON INDEX woulder.idx_mp_routes_route_type IS 'Speeds up heat map queries filtered by route type (Boulder, Sport, Trad, Ice, etc.)';

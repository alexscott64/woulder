-- Migration to add Kaya <-> Mountain Project route matching table
-- This enables smart linking between Kaya climbs and MP routes

CREATE TABLE kaya_mp_route_matches (
    id SERIAL PRIMARY KEY,
    
    -- Reference IDs (no foreign keys due to type mismatch between VARCHAR and BIGINT)
    kaya_climb_id VARCHAR(50) NOT NULL,
    mp_route_id BIGINT NOT NULL,
    
    -- Match confidence and type
    match_confidence DECIMAL(3,2) NOT NULL CHECK (match_confidence >= 0 AND match_confidence <= 1),
    match_type VARCHAR(50) NOT NULL, -- 'exact_name', 'fuzzy_name', 'location_name', 'gps_proximity', 'manual'
    
    -- Source data used for matching (denormalized for auditability)
    kaya_climb_name VARCHAR(255) NOT NULL,
    kaya_location_name VARCHAR(255),
    kaya_latitude DECIMAL(10, 7),
    kaya_longitude DECIMAL(10, 7),
    
    mp_route_name VARCHAR(255) NOT NULL,
    mp_area_name VARCHAR(500),
    mp_latitude DECIMAL(10, 7),
    mp_longitude DECIMAL(10, 7),
    
    -- Match scoring details
    name_similarity DECIMAL(3,2), -- 0.0 to 1.0, Levenshtein-based
    location_distance_km DECIMAL(10,2), -- GPS distance between locations
    location_name_match BOOLEAN DEFAULT false, -- Does "Leavenworth" match area hierarchy?
    
    -- Verification status
    is_verified BOOLEAN DEFAULT false,
    verified_by VARCHAR(100), -- username or system identifier
    verified_at TIMESTAMP,
    
    -- Match metadata
    match_notes TEXT, -- Additional context or reasoning
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure each pair is unique
    UNIQUE(kaya_climb_id, mp_route_id)
);

-- Indexes for efficient querying
CREATE INDEX idx_kaya_mp_matches_kaya_climb ON kaya_mp_route_matches(kaya_climb_id);
CREATE INDEX idx_kaya_mp_matches_mp_route ON kaya_mp_route_matches(mp_route_id);
CREATE INDEX idx_kaya_mp_matches_confidence ON kaya_mp_route_matches(match_confidence DESC);
CREATE INDEX idx_kaya_mp_matches_verified ON kaya_mp_route_matches(is_verified) WHERE is_verified = true;
CREATE INDEX idx_kaya_mp_matches_type ON kaya_mp_route_matches(match_type);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_kaya_mp_matches_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically update updated_at
CREATE TRIGGER update_kaya_mp_matches_timestamp
    BEFORE UPDATE ON kaya_mp_route_matches
    FOR EACH ROW
    EXECUTE FUNCTION update_kaya_mp_matches_updated_at();

-- Add comment explaining the table's purpose
COMMENT ON TABLE kaya_mp_route_matches IS 'Links Kaya climbs to Mountain Project routes using smart matching algorithms. Enables cross-referencing between the two databases.';
COMMENT ON COLUMN kaya_mp_route_matches.match_confidence IS 'Confidence score 0-1: <0.7=low, 0.7-0.85=medium, >0.85=high';
COMMENT ON COLUMN kaya_mp_route_matches.match_type IS 'Algorithm used: exact_name (1.0), fuzzy_name (Levenshtein), location_name (area match), gps_proximity (coords), manual (human verified)';

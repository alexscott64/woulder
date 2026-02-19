-- Migration 000026: Add detailed route information from Mountain Project API
-- Adds additional fields that come from the route details endpoint

-- Add new columns to mp_routes table
ALTER TABLE woulder.mp_routes
    ADD COLUMN difficulty VARCHAR(50),              -- Difficulty grade (V4, 5.10a, etc.) - separate from rating
    ADD COLUMN pitches INT,                        -- Number of pitches
    ADD COLUMN height_feet INT,                    -- Route height in feet
    ADD COLUMN mp_rating FLOAT,                    -- Mountain Project star rating (0-4)
    ADD COLUMN popularity FLOAT,                   -- Popularity score
    ADD COLUMN description_text TEXT,              -- HTML description from sections
    ADD COLUMN location_text TEXT,                 -- HTML location info from sections
    ADD COLUMN protection_text TEXT,               -- HTML protection info from sections
    ADD COLUMN safety_text TEXT;                   -- HTML safety info from sections

-- Create index on difficulty for filtering
CREATE INDEX idx_mp_routes_difficulty ON woulder.mp_routes(difficulty);

-- Create index on mp_rating for sorting
CREATE INDEX idx_mp_routes_rating ON woulder.mp_routes(mp_rating);

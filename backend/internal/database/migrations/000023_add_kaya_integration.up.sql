-- Migration 000023: Add Kaya climbing app integration tables
-- Stores locations, climbs, ascents, and user data from Kaya API

-- Table 1: Kaya Users
-- Stores user profile information from ascents and posts
CREATE TABLE woulder.kaya_users (
    id SERIAL PRIMARY KEY,
    kaya_user_id VARCHAR(50) UNIQUE NOT NULL,       -- Kaya user ID
    username VARCHAR(255) NOT NULL,
    fname VARCHAR(255),
    lname VARCHAR(255),
    photo_url TEXT,
    bio TEXT,
    height INTEGER,                                  -- Height in cm
    ape_index DECIMAL(5, 2),                        -- Ape index ratio
    limit_grade_bouldering_id VARCHAR(50),
    limit_grade_bouldering_name VARCHAR(50),
    limit_grade_routes_id VARCHAR(50),
    limit_grade_routes_name VARCHAR(50),
    is_private BOOLEAN DEFAULT FALSE,
    is_premium BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for kaya_users
CREATE INDEX idx_kaya_users_kaya_user_id ON woulder.kaya_users(kaya_user_id);
CREATE INDEX idx_kaya_users_username ON woulder.kaya_users(username);

-- Table 2: Kaya Locations
-- Stores destinations and areas with hierarchical relationships
CREATE TABLE woulder.kaya_locations (
    id SERIAL PRIMARY KEY,
    kaya_location_id VARCHAR(50) UNIQUE NOT NULL,   -- Kaya location ID
    slug VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    photo_url TEXT,
    description TEXT,
    
    -- Type information
    location_type_id VARCHAR(50),
    location_type_name VARCHAR(100),
    
    -- Parent relationship (hierarchical)
    parent_location_id VARCHAR(50),
    parent_location_slug VARCHAR(255),
    parent_location_name VARCHAR(255),
    
    -- Counts
    climb_count INTEGER DEFAULT 0,
    boulder_count INTEGER DEFAULT 0,
    route_count INTEGER DEFAULT 0,
    ascent_count INTEGER DEFAULT 0,
    
    -- Flags
    is_gb_moderated_bouldering BOOLEAN DEFAULT FALSE,
    is_gb_moderated_routes BOOLEAN DEFAULT FALSE,
    is_access_sensitive BOOLEAN DEFAULT FALSE,
    is_closed BOOLEAN DEFAULT FALSE,
    has_maps_disabled BOOLEAN DEFAULT FALSE,
    closed_date TIMESTAMPTZ,
    
    -- Descriptions by climb type
    description_bouldering TEXT,
    description_routes TEXT,
    description_short_bouldering TEXT,
    description_short_routes TEXT,
    access_description_bouldering TEXT,
    access_description_routes TEXT,
    access_issues_description_bouldering TEXT,
    access_issues_description_routes TEXT,
    
    -- Climb type filter (1=Boulder, 2=Routes)
    climb_type_id VARCHAR(50),
    
    -- Optional mapping to Woulder locations
    woulder_location_id INTEGER REFERENCES woulder.locations(id) ON DELETE SET NULL,
    
    -- Metadata
    last_synced_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for kaya_locations
CREATE INDEX idx_kaya_locations_kaya_location_id ON woulder.kaya_locations(kaya_location_id);
CREATE INDEX idx_kaya_locations_slug ON woulder.kaya_locations(slug);
CREATE INDEX idx_kaya_locations_parent ON woulder.kaya_locations(parent_location_id);
CREATE INDEX idx_kaya_locations_woulder_location ON woulder.kaya_locations(woulder_location_id);
CREATE INDEX idx_kaya_locations_lat_lng ON woulder.kaya_locations(latitude, longitude);

-- Table 3: Kaya Climbs
-- Stores individual routes and boulder problems (unified table)
CREATE TABLE woulder.kaya_climbs (
    id SERIAL PRIMARY KEY,
    kaya_climb_id VARCHAR(50) UNIQUE,               -- Derived from slug if available
    slug VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    
    -- Grade information
    grade_id VARCHAR(50),
    grade_name VARCHAR(50),                         -- "V4", "5.10a", etc.
    grade_ordering INTEGER,                         -- For sorting by difficulty
    grade_climb_type_id VARCHAR(50),
    
    -- Type (Boulder=1, Sport=2, Trad, etc.)
    climb_type_id VARCHAR(50),
    climb_type_name VARCHAR(50),
    
    -- Ratings
    rating DECIMAL(3, 2),                           -- Average star rating (0-5)
    ascent_count INTEGER DEFAULT 0,
    
    -- Location references
    kaya_destination_id VARCHAR(50),                -- Top-level location
    kaya_destination_name VARCHAR(255),
    kaya_area_id VARCHAR(50),                       -- Sub-location/area
    kaya_area_name VARCHAR(255),
    
    -- Gym/Board info (for indoor climbs)
    color_name VARCHAR(50),
    gym_name VARCHAR(255),
    board_name VARCHAR(255),
    
    -- Flags
    is_gb_moderated BOOLEAN DEFAULT FALSE,
    is_access_sensitive BOOLEAN DEFAULT FALSE,
    is_closed BOOLEAN DEFAULT FALSE,
    is_offensive BOOLEAN DEFAULT FALSE,
    
    -- Optional mapping to Woulder location
    woulder_location_id INTEGER REFERENCES woulder.locations(id) ON DELETE SET NULL,
    
    -- Metadata
    last_synced_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for kaya_climbs
CREATE INDEX idx_kaya_climbs_kaya_climb_id ON woulder.kaya_climbs(kaya_climb_id);
CREATE INDEX idx_kaya_climbs_slug ON woulder.kaya_climbs(slug);
CREATE INDEX idx_kaya_climbs_destination ON woulder.kaya_climbs(kaya_destination_id);
CREATE INDEX idx_kaya_climbs_area ON woulder.kaya_climbs(kaya_area_id);
CREATE INDEX idx_kaya_climbs_climb_type ON woulder.kaya_climbs(climb_type_id);
CREATE INDEX idx_kaya_climbs_grade ON woulder.kaya_climbs(grade_ordering);
CREATE INDEX idx_kaya_climbs_woulder_location ON woulder.kaya_climbs(woulder_location_id);

-- Table 4: Kaya Ascents
-- Stores user tick logs and send reports
CREATE TABLE woulder.kaya_ascents (
    id SERIAL PRIMARY KEY,
    kaya_ascent_id VARCHAR(50) UNIQUE NOT NULL,     -- Kaya ascent ID
    kaya_climb_slug VARCHAR(255) NOT NULL,          -- Reference to climb
    kaya_user_id VARCHAR(50) NOT NULL,              -- Reference to user
    
    -- Ascent details
    date TIMESTAMPTZ NOT NULL,                      -- When the ascent occurred
    comment TEXT,
    rating INTEGER,                                  -- Star rating (1-5)
    stiffness INTEGER,                               -- Grade accuracy feedback
    
    -- Grade at time of ascent (may differ from current climb grade)
    grade_id VARCHAR(50),
    grade_name VARCHAR(50),
    
    -- Media
    photo_url TEXT,
    photo_thumb_url TEXT,
    video_url TEXT,
    video_thumb_url TEXT,
    
    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Foreign keys
    FOREIGN KEY (kaya_climb_slug) REFERENCES woulder.kaya_climbs(slug) ON DELETE CASCADE,
    FOREIGN KEY (kaya_user_id) REFERENCES woulder.kaya_users(kaya_user_id) ON DELETE CASCADE
);

-- Indexes for kaya_ascents
CREATE INDEX idx_kaya_ascents_kaya_ascent_id ON woulder.kaya_ascents(kaya_ascent_id);
CREATE INDEX idx_kaya_ascents_climb ON woulder.kaya_ascents(kaya_climb_slug);
CREATE INDEX idx_kaya_ascents_user ON woulder.kaya_ascents(kaya_user_id);
CREATE INDEX idx_kaya_ascents_date ON woulder.kaya_ascents(date DESC);
-- Unique constraint to prevent duplicate ascents (same climb, user, timestamp)
CREATE UNIQUE INDEX idx_kaya_ascents_unique ON woulder.kaya_ascents(kaya_climb_slug, kaya_user_id, date);

-- Table 5: Kaya Posts
-- Stores user posts with photos/videos (beta videos, trip reports)
CREATE TABLE woulder.kaya_posts (
    id SERIAL PRIMARY KEY,
    kaya_post_id VARCHAR(50) UNIQUE NOT NULL,       -- Kaya post ID
    kaya_user_id VARCHAR(50) NOT NULL,              -- Post author
    date_created TIMESTAMPTZ NOT NULL,
    
    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Foreign key
    FOREIGN KEY (kaya_user_id) REFERENCES woulder.kaya_users(kaya_user_id) ON DELETE CASCADE
);

-- Indexes for kaya_posts
CREATE INDEX idx_kaya_posts_kaya_post_id ON woulder.kaya_posts(kaya_post_id);
CREATE INDEX idx_kaya_posts_user ON woulder.kaya_posts(kaya_user_id);
CREATE INDEX idx_kaya_posts_date ON woulder.kaya_posts(date_created DESC);

-- Table 6: Kaya Post Items
-- Stores individual media items within posts (photos, videos)
CREATE TABLE woulder.kaya_post_items (
    id SERIAL PRIMARY KEY,
    kaya_post_item_id VARCHAR(50) UNIQUE NOT NULL,  -- Kaya post item ID
    kaya_post_id VARCHAR(50) NOT NULL,              -- Parent post
    kaya_climb_slug VARCHAR(255),                   -- Optional climb reference
    kaya_ascent_id VARCHAR(50),                     -- Optional ascent reference
    
    -- Media
    photo_url TEXT,
    video_url TEXT,
    video_thumbnail_url TEXT,
    caption TEXT,
    
    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Foreign keys
    FOREIGN KEY (kaya_post_id) REFERENCES woulder.kaya_posts(kaya_post_id) ON DELETE CASCADE,
    FOREIGN KEY (kaya_climb_slug) REFERENCES woulder.kaya_climbs(slug) ON DELETE SET NULL,
    FOREIGN KEY (kaya_ascent_id) REFERENCES woulder.kaya_ascents(kaya_ascent_id) ON DELETE SET NULL
);

-- Indexes for kaya_post_items
CREATE INDEX idx_kaya_post_items_kaya_post_item_id ON woulder.kaya_post_items(kaya_post_item_id);
CREATE INDEX idx_kaya_post_items_post ON woulder.kaya_post_items(kaya_post_id);
CREATE INDEX idx_kaya_post_items_climb ON woulder.kaya_post_items(kaya_climb_slug);
CREATE INDEX idx_kaya_post_items_ascent ON woulder.kaya_post_items(kaya_ascent_id);

-- Table 7: Kaya Sync Progress
-- Tracks sync status for locations (similar to mp_sync_progress)
CREATE TABLE woulder.kaya_sync_progress (
    id SERIAL PRIMARY KEY,
    kaya_location_id VARCHAR(50) UNIQUE NOT NULL,
    location_name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,                    -- 'pending', 'in_progress', 'completed', 'failed'
    last_sync_at TIMESTAMPTZ,
    next_sync_at TIMESTAMPTZ,
    sync_error TEXT,
    
    -- Sync statistics
    climbs_synced INTEGER DEFAULT 0,
    ascents_synced INTEGER DEFAULT 0,
    sub_locations_synced INTEGER DEFAULT 0,
    
    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for kaya_sync_progress
CREATE INDEX idx_kaya_sync_progress_location ON woulder.kaya_sync_progress(kaya_location_id);
CREATE INDEX idx_kaya_sync_progress_status ON woulder.kaya_sync_progress(status);
CREATE INDEX idx_kaya_sync_progress_next_sync ON woulder.kaya_sync_progress(next_sync_at);

-- Add auto-update triggers for updated_at columns
CREATE TRIGGER update_kaya_users_updated_at BEFORE UPDATE ON woulder.kaya_users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_kaya_locations_updated_at BEFORE UPDATE ON woulder.kaya_locations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_kaya_climbs_updated_at BEFORE UPDATE ON woulder.kaya_climbs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_kaya_ascents_updated_at BEFORE UPDATE ON woulder.kaya_ascents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_kaya_posts_updated_at BEFORE UPDATE ON woulder.kaya_posts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_kaya_post_items_updated_at BEFORE UPDATE ON woulder.kaya_post_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_kaya_sync_progress_updated_at BEFORE UPDATE ON woulder.kaya_sync_progress
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

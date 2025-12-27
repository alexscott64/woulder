-- Woulder SQLite Schema
-- This file is tracked in git and used to initialize the database

-- Locations table
CREATE TABLE IF NOT EXISTS locations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    latitude REAL NOT NULL,
    longitude REAL NOT NULL,
    elevation_ft INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Weather data table
CREATE TABLE IF NOT EXISTS weather_data (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    location_id INTEGER NOT NULL,
    timestamp DATETIME NOT NULL,
    temperature REAL NOT NULL,
    feels_like REAL NOT NULL,
    precipitation REAL DEFAULT 0,
    humidity INTEGER DEFAULT 0,
    wind_speed REAL DEFAULT 0,
    wind_direction INTEGER DEFAULT 0,
    cloud_cover INTEGER DEFAULT 0,
    pressure INTEGER DEFAULT 0,
    description TEXT DEFAULT '',
    icon TEXT DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (location_id) REFERENCES locations(id) ON DELETE CASCADE,
    UNIQUE(location_id, timestamp)
);

-- Rivers table
CREATE TABLE IF NOT EXISTS rivers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    location_id INTEGER NOT NULL,
    gauge_id TEXT NOT NULL,
    river_name TEXT NOT NULL,
    safe_crossing_cfs INTEGER NOT NULL,
    caution_crossing_cfs INTEGER NOT NULL,
    drainage_area_sq_mi REAL,
    gauge_drainage_area_sq_mi REAL,
    flow_divisor REAL,
    is_estimated INTEGER DEFAULT 0,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (location_id) REFERENCES locations(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_weather_data_location_timestamp ON weather_data(location_id, timestamp);
CREATE INDEX IF NOT EXISTS idx_weather_data_timestamp ON weather_data(timestamp);
CREATE INDEX IF NOT EXISTS idx_rivers_location ON rivers(location_id);

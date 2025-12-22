-- Woulder Database Schema

-- Locations table
CREATE TABLE IF NOT EXISTS locations (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_coordinates (latitude, longitude)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Weather data table (stores both current and historical data)
CREATE TABLE IF NOT EXISTS weather_data (
    id INT AUTO_INCREMENT PRIMARY KEY,
    location_id INT NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    temperature DECIMAL(5, 2) NOT NULL COMMENT 'Temperature in Fahrenheit',
    feels_like DECIMAL(5, 2) NOT NULL COMMENT 'Feels like temperature in Fahrenheit',
    precipitation DECIMAL(6, 3) DEFAULT 0 COMMENT 'Precipitation in inches',
    humidity TINYINT NOT NULL COMMENT 'Humidity percentage (0-100)',
    wind_speed DECIMAL(5, 2) NOT NULL COMMENT 'Wind speed in mph',
    wind_direction SMALLINT NOT NULL COMMENT 'Wind direction in degrees (0-360)',
    cloud_cover TINYINT NOT NULL COMMENT 'Cloud cover percentage (0-100)',
    pressure SMALLINT NOT NULL COMMENT 'Atmospheric pressure in hPa',
    description VARCHAR(255) NOT NULL COMMENT 'Weather description',
    icon VARCHAR(10) NOT NULL COMMENT 'OpenWeatherMap icon code',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (location_id) REFERENCES locations(id) ON DELETE CASCADE,
    INDEX idx_location_timestamp (location_id, timestamp),
    INDEX idx_timestamp (timestamp),
    UNIQUE KEY unique_location_time (location_id, timestamp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert default locations
INSERT INTO locations (name, latitude, longitude) VALUES
('Skykomish', 47.70000522, -121.46672102),
('Index', 47.82083333, -121.55611111),
('Gold Bar', 47.85555556, -121.69694444),
('Bellingham', 48.75969444, -122.48847222),
('Icicle Creek (Leavenworth)', 47.59527778, -120.78361111),
('Squamish', 49.70147778, -123.15572222)
ON DUPLICATE KEY UPDATE name=name;

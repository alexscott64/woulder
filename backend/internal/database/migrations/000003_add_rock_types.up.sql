-- Create rock_types table
CREATE TABLE IF NOT EXISTS woulder.rock_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    base_drying_hours DECIMAL(4,1) NOT NULL, -- Base hours to dry after 0.1" rain
    porosity_percent DECIMAL(4,1), -- Average porosity percentage
    is_wet_sensitive BOOLEAN DEFAULT FALSE, -- True for sandstone/soft rocks
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create junction table for locations and rock types (many-to-many)
CREATE TABLE IF NOT EXISTS woulder.location_rock_types (
    location_id INTEGER NOT NULL REFERENCES woulder.locations(id) ON DELETE CASCADE,
    rock_type_id INTEGER NOT NULL REFERENCES woulder.rock_types(id) ON DELETE CASCADE,
    is_primary BOOLEAN DEFAULT FALSE, -- Mark the primary rock type for the location
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (location_id, rock_type_id)
);

-- Insert rock types with scientific data
INSERT INTO woulder.rock_types (name, base_drying_hours, porosity_percent, is_wet_sensitive, description) VALUES
-- Wet-sensitive rocks
('Sandstone', 36.0, 20.0, TRUE, 'Soft sedimentary rock that absorbs water and becomes friable when wet. DO NOT CLIMB WHEN WET.'),
('Arkose', 36.0, 18.0, TRUE, 'Feldspar-rich sandstone. Soft and water-absorbent. DO NOT CLIMB WHEN WET.'),
('Graywacke', 30.0, 15.0, TRUE, 'Hard sandstone with clay matrix. Absorbs water. DO NOT CLIMB WHEN WET.'),

-- Fast-drying rocks
('Granite', 6.0, 1.0, FALSE, 'Hard crystalline igneous rock. Non-porous, dries quickly.'),
('Granodiorite', 6.0, 1.2, FALSE, 'Coarse-grained igneous rock similar to granite. Dries quickly.'),
('Tonalite', 6.5, 1.5, FALSE, 'Plagioclase-rich igneous rock. Similar to granite, dries quickly.'),
('Rhyolite', 8.0, 7.0, FALSE, 'Fine-grained volcanic rock. Glassy texture sheds water well.'),

-- Medium-drying rocks
('Basalt', 10.0, 5.0, FALSE, 'Dense volcanic rock. May have vesicles that trap water.'),
('Andesite', 10.0, 6.0, FALSE, 'Intermediate volcanic rock. Moderate drying time.'),

-- Slow-drying rocks
('Schist', 12.0, 3.5, FALSE, 'Foliated metamorphic rock. Water can seep between layers.'),
('Phyllite', 20.0, 10.0, FALSE, 'Fine-grained metamorphic rock. Holds moisture in foliation.'),
('Argillite', 24.0, 12.0, FALSE, 'Clay-rich sedimentary rock. Absorbs and retains water.'),
('Chert', 14.0, 3.0, FALSE, 'Dense sedimentary rock. Micro-pores can hold water.'),
('Metavolcanic', 14.0, 4.0, FALSE, 'Metamorphosed volcanic rock. Moderate absorption.');

-- Create indexes for performance
CREATE INDEX idx_location_rock_types_location ON woulder.location_rock_types(location_id);
CREATE INDEX idx_location_rock_types_rock_type ON woulder.location_rock_types(rock_type_id);

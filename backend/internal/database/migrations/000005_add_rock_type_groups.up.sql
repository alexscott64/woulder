-- Create rock_type_groups table
CREATE TABLE IF NOT EXISTS woulder.rock_type_groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert rock type groups
INSERT INTO woulder.rock_type_groups (name, description) VALUES
('Wet-Sensitive Rocks', 'Soft rocks that are permanently damaged when climbed wet. DO NOT CLIMB WHEN WET.'),
('Fast-Drying Rocks', 'Hard, non-porous rocks that dry quickly after rain.'),
('Medium-Drying Rocks', 'Rocks with moderate porosity that take longer to dry.'),
('Slow-Drying Rocks', 'Rocks that absorb and retain water, requiring extended drying time.');

-- Add rock_type_group_id column to rock_types table
ALTER TABLE woulder.rock_types ADD COLUMN rock_type_group_id INTEGER REFERENCES woulder.rock_type_groups(id);

-- Update existing rock types with their group assignments
UPDATE woulder.rock_types SET rock_type_group_id = (SELECT id FROM woulder.rock_type_groups WHERE name = 'Wet-Sensitive Rocks')
WHERE name IN ('Sandstone', 'Arkose', 'Graywacke');

UPDATE woulder.rock_types SET rock_type_group_id = (SELECT id FROM woulder.rock_type_groups WHERE name = 'Fast-Drying Rocks')
WHERE name IN ('Granite', 'Granodiorite', 'Tonalite', 'Rhyolite');

UPDATE woulder.rock_types SET rock_type_group_id = (SELECT id FROM woulder.rock_type_groups WHERE name = 'Medium-Drying Rocks')
WHERE name IN ('Basalt', 'Andesite', 'Schist');

UPDATE woulder.rock_types SET rock_type_group_id = (SELECT id FROM woulder.rock_type_groups WHERE name = 'Slow-Drying Rocks')
WHERE name IN ('Phyllite', 'Argillite', 'Chert', 'Metavolcanic');

-- Make rock_type_group_id NOT NULL now that all rows have been updated
ALTER TABLE woulder.rock_types ALTER COLUMN rock_type_group_id SET NOT NULL;

-- Create index for performance
CREATE INDEX idx_rock_types_group ON woulder.rock_types(rock_type_group_id);

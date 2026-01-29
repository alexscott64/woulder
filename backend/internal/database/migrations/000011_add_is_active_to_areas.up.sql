-- Add is_active column to areas table to control which areas are visible in the UI
-- This allows filtering out newly imported Mountain Project areas until explicitly enabled

ALTER TABLE woulder.areas
ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;

-- Add index for efficient filtering of active areas
CREATE INDEX idx_areas_is_active ON woulder.areas(is_active);

-- Comment: Existing areas default to active (true)
-- Future imports should set is_active = false to prevent auto-display in UI

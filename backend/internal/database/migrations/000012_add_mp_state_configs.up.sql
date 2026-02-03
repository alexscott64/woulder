-- Migration 000012: Add Mountain Project state configurations table
-- Create table to configure which Mountain Project states to sync
-- This drives the 50-state sync process and controls future API visibility
-- Keeps 'areas' table strictly for weather frontend

CREATE TABLE woulder.mp_state_configs (
    id SERIAL PRIMARY KEY,
    state_name VARCHAR(100) NOT NULL UNIQUE,
    mp_area_id VARCHAR(50) NOT NULL UNIQUE,    -- Top-level MP area ID for state
    region VARCHAR(50),                         -- e.g., 'West', 'Southwest', 'Midwest'
    is_active BOOLEAN NOT NULL DEFAULT FALSE,   -- Controls visibility/API access
    display_order INTEGER DEFAULT 0,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Index for active state queries
CREATE INDEX idx_mp_state_configs_active ON woulder.mp_state_configs(is_active);
CREATE INDEX idx_mp_state_configs_region ON woulder.mp_state_configs(region);

-- Add auto-update trigger for updated_at
CREATE TRIGGER update_mp_state_configs_updated_at BEFORE UPDATE ON woulder.mp_state_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Populate 50 U.S. states with their Mountain Project area IDs
INSERT INTO woulder.mp_state_configs (state_name, mp_area_id, region, display_order, is_active) VALUES
  -- West (display_order 100-111)
  ('Alaska', '105909311', 'West', 100, FALSE),
  ('Arizona', '105708962', 'Southwest', 101, FALSE),
  ('California', '105708959', 'West', 102, FALSE),
  ('Colorado', '105708956', 'West', 103, FALSE),
  ('Hawaii', '106316122', 'West', 104, FALSE),
  ('Idaho', '105708958', 'West', 105, FALSE),
  ('Montana', '105907492', 'West', 106, FALSE),
  ('Nevada', '105708961', 'West', 107, FALSE),
  ('Oregon', '105708960', 'West', 108, FALSE),
  ('Utah', '105708957', 'West', 109, FALSE),
  ('Washington', '105708966', 'West', 110, FALSE),
  ('Wyoming', '105708960', 'West', 111, FALSE),

  -- Southwest (display_order 120-122)
  ('New Mexico', '105708964', 'Southwest', 120, FALSE),
  ('Texas', '105709046', 'Southwest', 121, FALSE),
  ('Oklahoma', '105911215', 'Southwest', 122, FALSE),

  -- Midwest (display_order 130-141)
  ('Illinois', '105911816', 'Midwest', 130, FALSE),
  ('Indiana', '105908701', 'Midwest', 131, FALSE),
  ('Iowa', '106092653', 'Midwest', 132, FALSE),
  ('Kansas', '105910743', 'Midwest', 133, FALSE),
  ('Michigan', '106113246', 'Midwest', 134, FALSE),
  ('Minnesota', '105910238', 'Midwest', 135, FALSE),
  ('Missouri', '105901239', 'Midwest', 136, FALSE),
  ('Nebraska', '105911954', 'Midwest', 137, FALSE),
  ('North Dakota', '105912378', 'Midwest', 138, FALSE),
  ('Ohio', '105994953', 'Midwest', 139, FALSE),
  ('South Dakota', '105708963', 'Midwest', 140, FALSE),
  ('Wisconsin', '105708968', 'Midwest', 141, FALSE),

  -- Southeast (display_order 150-161)
  ('Alabama', '105905173', 'Southeast', 150, FALSE),
  ('Arkansas', '105901027', 'Southeast', 151, FALSE),
  ('Florida', '111721391', 'Southeast', 152, FALSE),
  ('Georgia', '105897947', 'Southeast', 153, FALSE),
  ('Kentucky', '105868674', 'Southeast', 154, FALSE),
  ('Louisiana', '105910197', 'Southeast', 155, FALSE),
  ('Mississippi', '105907550', 'Southeast', 156, FALSE),
  ('North Carolina', '105905638', 'Southeast', 157, FALSE),
  ('South Carolina', '105906948', 'Southeast', 158, FALSE),
  ('Tennessee', '105905667', 'Southeast', 159, FALSE),
  ('Virginia', '105906523', 'Southeast', 160, FALSE),
  ('West Virginia', '105907015', 'Southeast', 161, FALSE),

  -- Northeast (display_order 170-180)
  ('Connecticut', '105806977', 'Northeast', 170, FALSE),
  ('Delaware', '106861605', 'Northeast', 171, FALSE),
  ('Maine', '105948977', 'Northeast', 172, FALSE),
  ('Maryland', '105908492', 'Northeast', 173, FALSE),
  ('Massachusetts', '105907214', 'Northeast', 174, FALSE),
  ('New Hampshire', '105872225', 'Northeast', 175, FALSE),
  ('New Jersey', '105909408', 'Northeast', 176, FALSE),
  ('New York', '105800424', 'Northeast', 177, FALSE),
  ('Pennsylvania', '105908143', 'Northeast', 178, FALSE),
  ('Rhode Island', '105912449', 'Northeast', 179, FALSE),
  ('Vermont', '105891603', 'Northeast', 180, FALSE);

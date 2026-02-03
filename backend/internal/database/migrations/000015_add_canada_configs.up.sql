-- Migration 000015: Add Canadian provinces and territories to mp_state_configs
-- Adds all 12 Canadian provinces/territories as individual areas to sync

INSERT INTO woulder.mp_state_configs (state_name, mp_area_id, region, display_order, is_active) VALUES
  -- Canada (display_order 200-211)
  ('Alberta', '105946432', 'Canada', 200, FALSE),
  ('British Columbia', '105946429', 'Canada', 201, FALSE),
  ('Manitoba', '106998817', 'Canada', 202, FALSE),
  ('New Brunswick', '106797047', 'Canada', 203, FALSE),
  ('Newfoundland and Labrador', '106715353', 'Canada', 204, FALSE),
  ('Northwest Territories', '106998828', 'Canada', 205, FALSE),
  ('Nova Scotia', '105894393', 'Canada', 206, FALSE),
  ('Nunavut', '107109005', 'Canada', 207, FALSE),
  ('Ontario', '105948616', 'Canada', 208, FALSE),
  ('Quebec', '106142016', 'Canada', 209, FALSE),
  ('Saskatchewan', '113243903', 'Canada', 210, FALSE),
  ('Yukon Territory', '106998806', 'Canada', 211, FALSE);

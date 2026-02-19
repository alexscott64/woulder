-- Migration 000023 down: Remove Kaya climbing app integration tables

-- Drop tables in reverse order (respecting foreign key constraints)
DROP TABLE IF EXISTS woulder.kaya_sync_progress;
DROP TABLE IF EXISTS woulder.kaya_post_items;
DROP TABLE IF EXISTS woulder.kaya_posts;
DROP TABLE IF EXISTS woulder.kaya_ascents;
DROP TABLE IF EXISTS woulder.kaya_climbs;
DROP TABLE IF EXISTS woulder.kaya_locations;
DROP TABLE IF EXISTS woulder.kaya_users;

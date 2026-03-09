-- Migration 000031 Rollback: Remove analytics tables

DROP TABLE IF EXISTS woulder.analytics_events;
DROP TABLE IF EXISTS woulder.analytics_sessions;
DROP TABLE IF EXISTS woulder.analytics_admin_users;

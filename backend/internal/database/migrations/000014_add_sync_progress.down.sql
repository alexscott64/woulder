-- Migration 000014: Rollback sync progress tracking table

DROP TRIGGER IF EXISTS update_sync_progress_updated_at ON woulder.sync_progress;
DROP INDEX IF EXISTS woulder.idx_sync_progress_mp_area;
DROP INDEX IF EXISTS woulder.idx_sync_progress_area;
DROP INDEX IF EXISTS woulder.idx_sync_progress_status;
DROP TABLE IF EXISTS woulder.sync_progress;

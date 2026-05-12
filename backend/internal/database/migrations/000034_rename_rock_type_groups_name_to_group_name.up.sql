-- Schema-drift converger for woulder.rock_type_groups.
--
-- (a) WHAT THIS FIXES:
--     Production's woulder.rock_type_groups has long used the column name
--     `group_name`, while migration 000005 declared it as `name`. All Go
--     queries, tests, and migration 000033 use `group_name`. This migration
--     renames `name` -> `group_name` on any environment where the old name
--     is still present (i.e. fresh DBs bootstrapped strictly from migrations).
--
-- (b) WHY IT IS GUARDED:
--     On production the column is already `group_name`, so the rename must
--     be a safe no-op there. The DO-block checks information_schema before
--     touching anything, making this migration idempotent and safe to run
--     on any environment (prod, dev-fresh, dev-already-converged).
--
-- (c) See migration 000005 for the original (drifted) table declaration.

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'woulder'
          AND table_name = 'rock_type_groups'
          AND column_name = 'name'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'woulder'
          AND table_name = 'rock_type_groups'
          AND column_name = 'group_name'
    ) THEN
        ALTER TABLE woulder.rock_type_groups RENAME COLUMN name TO group_name;
    END IF;
END $$;

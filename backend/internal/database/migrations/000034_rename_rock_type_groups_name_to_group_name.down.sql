-- Reverse of 000034.
--
-- (a) WHAT THIS FIXES (when run): reverts the converger rename of
--     woulder.rock_type_groups.group_name back to `name`.
-- (b) WHY IT IS GUARDED: only renames if `group_name` exists and `name`
--     does not, so this is a safe no-op on environments that never had
--     the rename applied. Intended for emergency rollback only — running
--     this WILL break Go code that selects `group_name`.
-- (c) See migration 000005 for the original (drifted) table declaration.

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'woulder'
          AND table_name = 'rock_type_groups'
          AND column_name = 'group_name'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'woulder'
          AND table_name = 'rock_type_groups'
          AND column_name = 'name'
    ) THEN
        ALTER TABLE woulder.rock_type_groups RENAME COLUMN group_name TO name;
    END IF;
END $$;

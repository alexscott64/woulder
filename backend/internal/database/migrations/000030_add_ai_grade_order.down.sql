-- Down migration: clear AI grade_order values (set back to NULL)
UPDATE woulder.mp_routes
SET grade_order = NULL
WHERE grade_order >= 400 AND grade_order <= 405;

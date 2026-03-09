-- Backfill grade_order for AI (Alpine Ice) grades that were not included in migration 29.
UPDATE woulder.mp_routes
SET grade_order = CASE
    WHEN UPPER(COALESCE(difficulty, rating, '')) LIKE 'AI1%' THEN 400
    WHEN UPPER(COALESCE(difficulty, rating, '')) LIKE 'AI2%' THEN 401
    WHEN UPPER(COALESCE(difficulty, rating, '')) LIKE 'AI3%' THEN 402
    WHEN UPPER(COALESCE(difficulty, rating, '')) LIKE 'AI4%' THEN 403
    WHEN UPPER(COALESCE(difficulty, rating, '')) LIKE 'AI5%' THEN 404
    WHEN UPPER(COALESCE(difficulty, rating, '')) LIKE 'AI6%' THEN 405
END
WHERE grade_order IS NULL
  AND UPPER(COALESCE(difficulty, rating, '')) LIKE 'AI%';

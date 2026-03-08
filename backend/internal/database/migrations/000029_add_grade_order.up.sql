-- Migration 000029: Add grade_order column for efficient grade range filtering
-- Stores a numeric sort order for each route's grade, enabling fast BETWEEN queries
-- on the heatmap without runtime string parsing.
--
-- Grade order namespaces:
--   V-scale (bouldering): 0-17   (V0=0, V1=1, ..., V17=17)
--   YDS (sport/trad):     100-129 (5.4=100, 5.5=101, ..., 5.15d=129)
--   WI (ice):             200-206 (WI1=200, ..., WI7=206)
--   Mixed:                300-312 (M1=300, ..., M13=312)

-- Add grade_order column
ALTER TABLE woulder.mp_routes ADD COLUMN IF NOT EXISTS grade_order INTEGER;

-- Create index for efficient range filtering
CREATE INDEX IF NOT EXISTS idx_mp_routes_grade_order ON woulder.mp_routes(grade_order) WHERE grade_order IS NOT NULL;

-- Backfill grade_order from existing rating/difficulty data.
-- Uses COALESCE(difficulty, rating) to pick the best available grade string,
-- then maps known grade patterns to their numeric sort order.
UPDATE woulder.mp_routes
SET grade_order = CASE
    -- V-scale (bouldering): V0=0, V1=1, ..., V17=17
    WHEN UPPER(COALESCE(difficulty, rating, '')) ~ '^V[0-9]' THEN
        CASE
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'V0%' THEN 0
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'V1 %' OR UPPER(COALESCE(difficulty, rating)) = 'V1' OR UPPER(COALESCE(difficulty, rating)) LIKE 'V1-%' OR UPPER(COALESCE(difficulty, rating)) LIKE 'V1+%' THEN 1
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'V10%' THEN 10
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'V11%' THEN 11
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'V12%' THEN 12
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'V13%' THEN 13
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'V14%' THEN 14
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'V15%' THEN 15
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'V16%' THEN 16
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'V17%' THEN 17
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'V2%' THEN 2
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'V3%' THEN 3
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'V4%' THEN 4
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'V5%' THEN 5
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'V6%' THEN 6
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'V7%' THEN 7
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'V8%' THEN 8
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'V9%' THEN 9
            ELSE NULL
        END

    -- YDS (sport/trad): 5.4=100, 5.5=101, ..., 5.15d=129
    WHEN COALESCE(difficulty, rating, '') ~ '^5\.' THEN
        CASE
            WHEN COALESCE(difficulty, rating) LIKE '5.4%' THEN 100
            WHEN COALESCE(difficulty, rating) LIKE '5.5%' THEN 101
            WHEN COALESCE(difficulty, rating) LIKE '5.6%' THEN 102
            WHEN COALESCE(difficulty, rating) LIKE '5.7%' THEN 103
            WHEN COALESCE(difficulty, rating) LIKE '5.8%' THEN 104
            WHEN COALESCE(difficulty, rating) LIKE '5.9%' THEN 105
            WHEN LOWER(COALESCE(difficulty, rating)) LIKE '5.10d%' THEN 109
            WHEN LOWER(COALESCE(difficulty, rating)) LIKE '5.10c%' THEN 108
            WHEN LOWER(COALESCE(difficulty, rating)) LIKE '5.10b%' THEN 107
            WHEN LOWER(COALESCE(difficulty, rating)) LIKE '5.10a%' OR COALESCE(difficulty, rating) LIKE '5.10 %' OR COALESCE(difficulty, rating) = '5.10' THEN 106
            WHEN LOWER(COALESCE(difficulty, rating)) LIKE '5.10%' THEN 106  -- fallback for 5.10 without letter
            WHEN LOWER(COALESCE(difficulty, rating)) LIKE '5.11d%' THEN 113
            WHEN LOWER(COALESCE(difficulty, rating)) LIKE '5.11c%' THEN 112
            WHEN LOWER(COALESCE(difficulty, rating)) LIKE '5.11b%' THEN 111
            WHEN LOWER(COALESCE(difficulty, rating)) LIKE '5.11a%' OR COALESCE(difficulty, rating) = '5.11' THEN 110
            WHEN LOWER(COALESCE(difficulty, rating)) LIKE '5.11%' THEN 110
            WHEN LOWER(COALESCE(difficulty, rating)) LIKE '5.12d%' THEN 117
            WHEN LOWER(COALESCE(difficulty, rating)) LIKE '5.12c%' THEN 116
            WHEN LOWER(COALESCE(difficulty, rating)) LIKE '5.12b%' THEN 115
            WHEN LOWER(COALESCE(difficulty, rating)) LIKE '5.12a%' OR COALESCE(difficulty, rating) = '5.12' THEN 114
            WHEN LOWER(COALESCE(difficulty, rating)) LIKE '5.12%' THEN 114
            WHEN LOWER(COALESCE(difficulty, rating)) LIKE '5.13d%' THEN 121
            WHEN LOWER(COALESCE(difficulty, rating)) LIKE '5.13c%' THEN 120
            WHEN LOWER(COALESCE(difficulty, rating)) LIKE '5.13b%' THEN 119
            WHEN LOWER(COALESCE(difficulty, rating)) LIKE '5.13a%' OR COALESCE(difficulty, rating) = '5.13' THEN 118
            WHEN LOWER(COALESCE(difficulty, rating)) LIKE '5.13%' THEN 118
            WHEN LOWER(COALESCE(difficulty, rating)) LIKE '5.14d%' THEN 125
            WHEN LOWER(COALESCE(difficulty, rating)) LIKE '5.14c%' THEN 124
            WHEN LOWER(COALESCE(difficulty, rating)) LIKE '5.14b%' THEN 123
            WHEN LOWER(COALESCE(difficulty, rating)) LIKE '5.14a%' OR COALESCE(difficulty, rating) = '5.14' THEN 122
            WHEN LOWER(COALESCE(difficulty, rating)) LIKE '5.14%' THEN 122
            WHEN LOWER(COALESCE(difficulty, rating)) LIKE '5.15d%' THEN 129
            WHEN LOWER(COALESCE(difficulty, rating)) LIKE '5.15c%' THEN 128
            WHEN LOWER(COALESCE(difficulty, rating)) LIKE '5.15b%' THEN 127
            WHEN LOWER(COALESCE(difficulty, rating)) LIKE '5.15a%' OR COALESCE(difficulty, rating) = '5.15' THEN 126
            WHEN LOWER(COALESCE(difficulty, rating)) LIKE '5.15%' THEN 126
            ELSE NULL
        END

    -- WI (ice): WI1=200, ..., WI7=206
    WHEN UPPER(COALESCE(difficulty, rating, '')) LIKE 'WI%' THEN
        CASE
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'WI1%' THEN 200
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'WI2%' THEN 201
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'WI3%' THEN 202
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'WI4%' THEN 203
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'WI5%' THEN 204
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'WI6%' THEN 205
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'WI7%' THEN 206
            ELSE NULL
        END

    -- Mixed: M1=300, ..., M13=312
    WHEN UPPER(COALESCE(difficulty, rating, '')) ~ '^M[0-9]' THEN
        CASE
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'M10%' THEN 309
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'M11%' THEN 310
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'M12%' THEN 311
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'M13%' THEN 312
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'M1%' THEN 300
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'M2%' THEN 301
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'M3%' THEN 302
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'M4%' THEN 303
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'M5%' THEN 304
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'M6%' THEN 305
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'M7%' THEN 306
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'M8%' THEN 307
            WHEN UPPER(COALESCE(difficulty, rating)) LIKE 'M9%' THEN 308
            ELSE NULL
        END

    ELSE NULL
END
WHERE COALESCE(difficulty, rating, '') != '';

-- Add comment for documentation
COMMENT ON COLUMN woulder.mp_routes.grade_order IS 'Numeric sort order for grade range filtering. V-scale: 0-17, YDS: 100-129, WI: 200-206, Mixed: 300-312. NULL for unrecognized grades.';

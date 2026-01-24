-- Clean existing Mountain Project tick comments
-- Removes HTML entities (&middot;, &amp;, etc.) and leading middot characters

-- Step 1: Decode HTML entities and remove leading middot characters
UPDATE mp_ticks
SET comment = NULLIF(
    TRIM(
        REGEXP_REPLACE(
            REGEXP_REPLACE(
                REGEXP_REPLACE(
                    REGEXP_REPLACE(
                        REGEXP_REPLACE(
                            REGEXP_REPLACE(
                                comment,
                                '^&middot;\s*', '', 'g'  -- Remove leading &middot; with optional space
                            ),
                            '^·\s*', '', 'g'  -- Remove leading · character with optional space
                        ),
                        '&middot;', '', 'g'  -- Remove any remaining &middot;
                    ),
                    '&amp;', '&', 'g'  -- Decode &amp; to &
                ),
                '&quot;', '"', 'g'  -- Decode &quot; to "
            ),
            '&lt;|&gt;', '', 'g'  -- Remove &lt; and &gt;
        )
    ),
    ''  -- Set to NULL if result is empty string
)
WHERE comment IS NOT NULL
  AND (comment LIKE '%&middot;%' OR comment LIKE '%·%' OR comment LIKE '%&amp;%' OR comment LIKE '%&quot;%' OR comment LIKE '%&lt;%' OR comment LIKE '%&gt;%');

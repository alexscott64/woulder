-- Rollback seed data migration
-- Removes all seeded data in reverse order

SET search_path TO woulder, public;

-- Delete rivers first (foreign key to locations)
DELETE FROM woulder.rivers WHERE location_id IN (
    SELECT id FROM woulder.locations WHERE name IN (
        'Skykomish - Money Creek',
        'Index',
        'Gold Bar',
        'Skykomish - Paradise'
    )
);

-- Delete locations (foreign key to areas)
DELETE FROM woulder.locations WHERE name IN (
    -- Pacific Northwest
    'Skykomish - Money Creek',
    'Index',
    'Gold Bar',
    'Bellingham',
    'Icicle Creek (Leavenworth)',
    'Squamish',
    'Skykomish - Paradise',
    'Treasury',
    'Calendar Butte',
    -- Southern California
    'Joshua Tree',
    'Black Mountain',
    'Buttermilks',
    'Happy / Sad Boulders',
    'Yosemite',
    'Tramway'
);

-- Delete areas
DELETE FROM woulder.areas WHERE name IN (
    'Pacific Northwest',
    'Southern California'
);

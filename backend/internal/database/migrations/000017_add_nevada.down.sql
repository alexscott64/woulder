-- Migration Rollback: Remove Nevada (Red Rock Canyon) area and locations
-- Created: 2026-02-03

-- Remove sun exposure profiles for Nevada locations
DELETE FROM location_sun_exposure
WHERE location_id IN (
    SELECT id FROM locations WHERE name IN (
        'Oak Creek Canyon Boulders',
        'Black Velvet Canyon Boulders',
        'Windy Canyon Boulders',
        'Second Pullout Boulders',
        'Willow Spring Boulders',
        'First Pullout Boulders',
        'White Rock Spring Boulders',
        'Pine Creek Canyon Boulders',
        'Sandstone Quarry Boulders',
        'Ice Box Canyon Boulders',
        'Juniper Canyon Boulders',
        'Southern Outcrops Boulders',
        'Mustang Canyon Boulders',
        'First Creek Canyon Boulders',
        'Kraft Boulders',
        'Red Spring Boulders',
        'Gateway Canyon',
        'Little Springs Canyon Boulders',
        'Ash Spring Boulders'
    )
);

-- Remove rock types for Nevada locations
DELETE FROM location_rock_types
WHERE location_id IN (
    SELECT id FROM locations WHERE name IN (
        'Oak Creek Canyon Boulders',
        'Black Velvet Canyon Boulders',
        'Windy Canyon Boulders',
        'Second Pullout Boulders',
        'Willow Spring Boulders',
        'First Pullout Boulders',
        'White Rock Spring Boulders',
        'Pine Creek Canyon Boulders',
        'Sandstone Quarry Boulders',
        'Ice Box Canyon Boulders',
        'Juniper Canyon Boulders',
        'Southern Outcrops Boulders',
        'Mustang Canyon Boulders',
        'First Creek Canyon Boulders',
        'Kraft Boulders',
        'Red Spring Boulders',
        'Gateway Canyon',
        'Little Springs Canyon Boulders',
        'Ash Spring Boulders'
    )
);

-- Remove Nevada locations
DELETE FROM locations WHERE name IN (
    'Oak Creek Canyon Boulders',
    'Black Velvet Canyon Boulders',
    'Windy Canyon Boulders',
    'Second Pullout Boulders',
    'Willow Spring Boulders',
    'First Pullout Boulders',
    'White Rock Spring Boulders',
    'Pine Creek Canyon Boulders',
    'Sandstone Quarry Boulders',
    'Ice Box Canyon Boulders',
    'Juniper Canyon Boulders',
    'Southern Outcrops Boulders',
    'Mustang Canyon Boulders',
    'First Creek Canyon Boulders',
    'Kraft Boulders',
    'Red Spring Boulders',
    'Gateway Canyon',
    'Little Springs Canyon Boulders',
    'Ash Spring Boulders'
);

-- Remove Nevada area
DELETE FROM areas WHERE name = 'Nevada';

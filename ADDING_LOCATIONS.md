# Adding New Climbing Locations to Woulder

This guide explains how to add new climbing areas to the Woulder weather dashboard.

## Quick Steps

1. **Get location information**
2. **Add to database seed file**
3. **Add river data (if applicable)**
4. **Update documentation**
5. **Test the changes**

---

## Step 1: Gather Location Information

You'll need the following for each location:

- **Name**: Descriptive name (e.g., "Treasury", "Index")
- **Latitude**: Decimal degrees (e.g., 47.76086166)
- **Longitude**: Decimal degrees (e.g., -121.12877297)
- **Elevation**: Feet above sea level (e.g., 3650)

### How to Find Coordinates

**Option 1: Google Maps**
1. Right-click on the climbing area in Google Maps
2. Click the coordinates to copy them
3. Format: `47.76086166, -121.12877297`

**Option 2: Mountain Project**
1. Find the area on Mountain Project
2. Coordinates are usually listed on the area page

**Option 3: GPS Device**
1. Use a GPS at the approach/base of the climbing area
2. Record coordinates in decimal degrees format

---

## Step 2: Add Location to Database Seed

Edit `backend/internal/database/seed.sql`:

```sql
-- Locations
INSERT OR IGNORE INTO locations (name, latitude, longitude, elevation_ft) VALUES
    ('Skykomish - Money Creek', 47.69727769, -121.47884640, 1000),
    ('Index', 47.82061272, -121.55492795, 500),
    -- ... existing locations ...
    ('Treasury', 47.76086166, -121.12877297, 3650),
    ('Your New Location', 48.12345, -122.67890, 2000);  -- Add your new location here
```

### Field Descriptions

- `name`: Display name (string, max 255 chars)
- `latitude`: Decimal degrees, positive for North
- `longitude`: Decimal degrees, negative for West (in North America)
- `elevation_ft`: Elevation in feet (integer)

---

## Step 3: Add River Crossing Data (Optional)

If the climbing area requires a river crossing, add river data after the locations section.

### Example: Direct Gauge Reading

```sql
-- Your Location - Direct USGS gauge
INSERT OR IGNORE INTO rivers (location_id, gauge_id, river_name, safe_crossing_cfs, caution_crossing_cfs, is_estimated, description)
SELECT id, '12134500', 'Skykomish River', 3000, 4500, 0,
       'Direct gauge reading from USGS Skykomish River near Gold Bar'
FROM locations WHERE name = 'Your New Location';
```

### Example: Estimated from Nearby Gauge

```sql
-- Your Location - Estimated using drainage area ratio
INSERT OR IGNORE INTO rivers (location_id, gauge_id, river_name, safe_crossing_cfs, caution_crossing_cfs, drainage_area_sq_mi, gauge_drainage_area_sq_mi, is_estimated, description)
SELECT id, '12134500', 'Money Creek (estimated)', 60, 90, 18.0, 355.0, 1,
       'Flow estimated from nearby gauge using drainage area ratio'
FROM locations WHERE name = 'Your New Location';
```

### Example: Estimated using Flow Divisor

```sql
-- Your Location - Estimated using divisor
INSERT OR IGNORE INTO rivers (location_id, gauge_id, river_name, safe_crossing_cfs, caution_crossing_cfs, flow_divisor, is_estimated, description)
SELECT id, '12134500', 'North Fork Skykomish River', 800, 900, 2.0, 1,
       'Flow estimated as gauge reading / 2 per local climbing guide'
FROM locations WHERE name = 'Your New Location';
```

### Finding River Information

**USGS Gauge IDs:**
1. Go to https://waterdata.usgs.gov/nwis/rt
2. Find the nearest gauge station to your climbing area
3. Copy the 8-digit gauge ID (e.g., `12134500`)

**Safe Crossing Thresholds:**
- Check local climbing guides or forums
- Ask experienced local climbers
- Use conservative estimates until validated
- Document the source in the `description` field

**Drainage Area:**
- Available on USGS gauge station pages
- Used to scale flow estimates from larger rivers
- Formula: `EstimatedFlow = GaugeFlow * (LocalArea / GaugeArea)`

---

## Step 4: Update Documentation

### README.md

Update the locations table:

```markdown
| Location | Coordinates | Elevation | Region |
|----------|-------------|-----------|--------|
| Treasury | 47.76, -121.13 | 3,650 ft | Washington |
| Your New Location | 48.12, -122.68 | 2,000 ft | Washington |
```

Update the feature count:

```markdown
- **10 Climbing Locations** - ... Treasury, Your New Location
```

### QUICKSTART.md

Update the location count in the "Prerequisites Completed" section:

```markdown
- [x] All 10 default locations added to database
```

Update the "What Should You See?" section:

```markdown
- Grid of weather cards for all 10 locations:
  - ... existing locations ...
  - Treasury
  - Your New Location
```

---

## Step 5: Test the Changes

### Reset the Database

```bash
# Delete the existing database
rm backend/woulder.db

# Restart the backend (will recreate database with new seed data)
cd backend
go run cmd/server/main.go
```

### Verify in the Frontend

```bash
cd frontend
npm run dev
```

**Check for:**
1. New location card appears in the grid
2. Weather data loads correctly
3. 6-day forecast expands properly
4. River crossing icon appears (if applicable)
5. No console errors

### Test the API Directly

```bash
# Get all locations
curl http://localhost:8080/api/locations

# Get weather for new location (replace ID)
curl http://localhost:8080/api/weather/9

# Get river data (if applicable)
curl http://localhost:8080/api/rivers/location/9
```

---

## Common Issues

### Location Not Appearing

**Problem**: New location doesn't show in the UI

**Solutions**:
- Check that the database was actually reset
- Verify the seed.sql syntax is correct
- Check backend logs for SQL errors
- Ensure `INSERT OR IGNORE` is used (prevents duplicates)

### Weather Data Not Loading

**Problem**: Card shows "Loading..." indefinitely

**Solutions**:
- Check backend logs for Open-Meteo API errors
- Verify coordinates are in decimal degrees format
- Ensure latitude/longitude are not swapped
- Check internet connectivity

### River Data Errors

**Problem**: River info shows incorrect flow or status

**Solutions**:
- Verify USGS gauge ID is correct (8 digits)
- Check that gauge is active (not decommissioned)
- Validate safe/caution thresholds with local knowledge
- Test with different flow conditions

---

## Advanced: Adding Rivers to Existing Locations

If you want to add river crossing data to a location that was already added without rivers:

```sql
-- Add river to existing location
INSERT OR IGNORE INTO rivers (location_id, gauge_id, river_name, safe_crossing_cfs, caution_crossing_cfs, is_estimated, description)
SELECT id, '12134500', 'Creek Name', 100, 150, 1,
       'Description of crossing and how flow is estimated'
FROM locations WHERE name = 'Existing Location Name';
```

Then reset the database to apply changes.

---

## Tips for Choosing Safe Crossing Thresholds

1. **Start Conservative**: Use lower thresholds until validated by field experience
2. **Consider River Width**: Wider rivers are more dangerous at the same CFS
3. **Account for Substrate**: Rocky bottoms are more treacherous
4. **Seasonal Variations**: Spring snowmelt makes crossings more dangerous
5. **Local Knowledge**: Always defer to experienced local climbers
6. **Document Sources**: Note where threshold came from in description

### Example Thresholds by River Type

- **Small Creek (10-20 ft wide)**: Safe <50 CFS, Caution 50-80, Unsafe >80
- **Medium Creek (20-40 ft wide)**: Safe <150 CFS, Caution 150-300, Unsafe >300
- **Large River (>40 ft wide)**: Safe <800 CFS, Caution 800-1500, Unsafe >1500

**These are rough guidelines only.** Always validate with local experience.

---

## Need Help?

- Check existing locations in `seed.sql` for examples
- USGS Water Data: https://waterdata.usgs.gov/nwis/rt
- Open-Meteo API: https://open-meteo.com/
- Mountain Project: https://www.mountainproject.com/

---

**Remember**: River crossing safety data should be treated as estimates. Always use your own judgment and experience when assessing river crossings in the field.

---

## Sun Exposure Profile (rock-temp confidence)

Every active location should have a row in `woulder.location_sun_exposure`. Without one, the rock-temperature **confidence score** is reduced by **40 points** (25 for aspect + 15 for dip) via [`confidence.go:51-59`](backend/internal/weather/rock_temp/confidence.go:51), capping it at ~60/100 regardless of how good the weather data is.

### What's in the table

There is **one row per location** in [`woulder.location_sun_exposure`](backend/internal/database/migrations/000006_add_sun_exposure_profiles.up.sql:5). The schema is **percentage-based** — there are no `aspect_degrees` or `face_dip_degrees` columns. The rock-temp calculator derives a single representative aspect/dip on the fly from these fields:

| Column | Range | Meaning |
|---|---|---|
| `south_facing_percent` | 0–100 | % of climbing surface facing south (180°) |
| `west_facing_percent` | 0–100 | % facing west (270°) |
| `east_facing_percent` | 0–100 | % facing east (90°) |
| `north_facing_percent` | 0–100 | % facing north (0°) |
| `slab_percent` | 0–100 | % low-angle (face dip ≈ 45°) |
| `overhang_percent` | 0–100 | % overhanging (face dip ≈ 110°) |
| `tree_coverage_percent` | 0–100 | canopy shading the rock |
| `description` | text | free-form notes |

The four facing-percent fields should sum to roughly 100. `slab_percent + overhang_percent` should be ≤ 100; the remainder is treated as vertical wall (dip = 90°).

### How aspect and dip are derived

From [`ResolveDominantFacet()`](backend/internal/weather/rock_temp/inputs.go:99):

**Aspect resolution**

- If any one facing-percent is **> 60**, that face dominates: S=180°, W=270°, E=90°, N=0°.
- Otherwise the four percentages are vector-summed via [`WeightedAspect()`](backend/internal/weather/rock_temp/inputs.go:38) into a single representative azimuth. If the resultant vector is degenerate (e.g., 50% E + 50% W cancel out), the calculator defaults to south-facing (180°).

**Dip resolution**

- `slab_percent > 60` → dip = 45°
- `overhang_percent > 60` → dip = 110°
- Otherwise the calculator uses [`WeightedDip()`](backend/internal/weather/rock_temp/inputs.go:63), a blend with weights 45° (slab) / 90° (vertical) / 110° (overhang).

**Practical implication for operators:** to get a *clean* (non-mixed) aspect and dip, push one face above 60% and pick either slab or overhang above 60%. If your area is genuinely mixed, leave it mixed — the calculator handles that case correctly, it just emits a "mixed facets" reason string at lower confidence weight.

### Visual references

**Aspect (compass bearing the rock face points toward)**

```text
        N (0°)
          │
W (270°)──┼──E (90°)
          │
        S (180°)
```

**Dip (face angle from horizontal)**

```text
0°   ─────  flat top / horizontal slab
45°  ╱      typical low-angle slab
75°  │      steep wall
90°  │      vertical
110° ╲      overhanging (past vertical)
```

### Editing the data

Two paths:

1. **For a brand-new location being added via `seed.sql`** (Step 2 above): include a parallel `INSERT INTO location_sun_exposure …` block in the same edit, modeled after the existing entries near the bottom of [`backend/internal/database/seed.sql`](backend/internal/database/seed.sql:137).

2. **For one-off updates to an existing live location:** use the seed template, which is set up for idempotent UPSERTs by name.

   ```cmd
   cd backend/internal/database/migrations
   copy seed_location_sun_exposure_TEMPLATE.sql seed_location_sun_exposure.sql
   :: edit seed_location_sun_exposure.sql in your editor
   set PGPASSWORD=<from backend/.env>
   psql "host=<DB_HOST> port=5432 user=woulder dbname=woulder sslmode=require" -f seed_location_sun_exposure.sql
   ```

   The template (`_TEMPLATE` suffix means it is **not** a numbered migration; do not move it into a `000NNN_*.sql` slot) contains one `INSERT … SELECT … ON CONFLICT (location_id) DO UPDATE` block per active location, alphabetized by region → area → name and pre-populated from the live DB so you only have to change the values you want to refine.

### Re-auditing

To check current coverage and find any location that's missing a row or has all-zero facets/dip:

```cmd
set PGPASSWORD=<from backend/.env>
psql "host=<DB_HOST> port=5432 user=woulder dbname=woulder sslmode=require" -f backend/internal/database/migrations/audit_location_sun_exposure.sql
```

The script ([`audit_location_sun_exposure.sql`](backend/internal/database/migrations/audit_location_sun_exposure.sql)) is read-only and prints both a summary block (totals + counts of `rows_missing` / `rows_all_zero_facets` / `rows_all_zero_dip`) and a per-location detail listing with `aspect_status` and `dip_status` flags.

### Known issue: misleading confidence reason strings

The reason strings emitted by [`confidence.go:54`](backend/internal/weather/rock_temp/confidence.go:54) and [`confidence.go:58`](backend/internal/weather/rock_temp/confidence.go:58) tell operators to populate `aspect_degrees` and `face_dip_degrees` in `location_sun_exposure`. **Those columns do not exist** — the schema is percentage-based as documented above. Until those strings are corrected (Code-mode follow-up), interpret any "aspect defaulted" / "dip defaulted" reason as: *"this location does not have a `location_sun_exposure` row at all — insert one with the seven percentage fields."* The deduction is gated on `in.SunExposure != nil` ([`calculator.go:206-207`](backend/internal/weather/rock_temp/calculator.go:206)), not on individual NULL columns, so a row with all-zero values still suppresses the deduction (though it will produce a degenerate aspect and the calculator will fall back to south-facing).

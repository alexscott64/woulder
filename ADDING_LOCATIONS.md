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

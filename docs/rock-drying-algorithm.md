# Rock Drying Algorithm - Scientific Documentation

## Overview

The rock drying algorithm calculates whether climbing rock is dry and safe after precipitation events. It uses a comprehensive model that considers rock properties, weather conditions, sun exposure geometry, and environmental factors to provide accurate drying time estimates with confidence scoring.

## Core Principles

### 1. Time-Weighted Drying
Unlike simple binary rain/no-rain calculations, our algorithm recognizes that drying is a continuous process affected by changing environmental conditions. A warm, windy, sunny hour provides more drying than a cold, humid, calm night. The algorithm tracks "effective drying hours" where favorable conditions count for more than baseline, and poor conditions count for less.

### 2. Rain Event Analysis
Rain is analyzed as discrete events rather than individual readings. The algorithm groups contiguous precipitation into events and calculates:
- **Total rainfall**: Sum of all precipitation in the event
- **Duration**: Time from first to last rain reading
- **Max hourly rate**: Peak intensity during the event
- **Average hourly rate**: Mean intensity across the event

This approach correctly distinguishes between "0.3 inches in 30 minutes" (intense) and "0.3 inches over 12 hours" (drizzle).

### 3. Confidence Scoring
Every prediction includes a confidence score (0-100) representing the trustworthiness of the estimate, NOT the accuracy. Factors affecting confidence:
- Data completeness (historical weather depth)
- Time since last rain (more time = more confident)
- Weather stability (variable conditions = less confident)
- Sun exposure profile availability
- Rock sensitivity (wet-sensitive rocks = lower confidence when wet)

Confidence is clamped to 20-95 range to avoid false certainty.

## Rock Properties

### Base Drying Hours
Time required for rock to dry after 0.1" of rain in ideal conditions (70°F, 30% humidity, 10mph wind, sunny, south-facing slab).

| Rock Type | Base Hours | Notes |
|-----------|-----------|-------|
| Granite | 6.0 | Non-porous, crystalline structure |
| Granodiorite | 6.0 | Similar to granite |
| Tonalite | 6.5 | Slightly more weathered than granite |
| Rhyolite | 8.0 | Fine-grained, glassy texture |
| Basalt | 10.0 | Dense but may have vesicles |
| Andesite | 10.0 | Intermediate volcanic rock |
| Schist | 12.0 | Foliated, water between layers |
| Chert | 14.0 | Dense with micro-pores |
| Metavolcanic | 14.0 | Metamorphosed volcanic |
| Phyllite | 20.0 | Fine foliation holds moisture |
| Argillite | 24.0 | Clay-rich, highly absorbent |
| Graywacke | 30.0 | Hard sandstone with clay matrix |
| Sandstone | 36.0 | Highly porous and absorbent |
| Arkose | 36.0 | Feldspar-rich sandstone |

### Porosity
Percentage of rock volume that is pore space capable of holding water.

**Algorithm Integration**:
```
porosityFactor = 1.0 + ((rockPorosityPercent - 5.0) / 100.0)
dryingTime = baseDryingTime × porosityFactor
```

**Examples**:
- Granite (1% porosity): Factor = 0.96x (4% faster drying)
- Basalt (5% porosity): Factor = 1.0x (baseline)
- Sandstone (20% porosity): Factor = 1.15x (15% slower drying)

Clamped to 0.7-1.5 range to prevent extreme values.

### Wet-Sensitive Rocks
Certain rocks (sandstone, arkose, graywacke) are permanently damaged when climbed wet. These receive:
- **+50% drying time** safety margin
- **"critical" status** when any moisture present
- **"DO NOT CLIMB"** warnings
- **Reduced confidence** scores when wet

## Environmental Factors

### Temperature
Warm temperatures increase evaporation rate, cold temperatures slow it.

| Temperature (°F) | Drying Modifier | Reasoning |
|-----------------|----------------|-----------|
| > 70° | 0.75x (faster) | High evaporation rate |
| 65-70° | 0.85x (faster) | Moderate evaporation |
| 55-65° | 1.0x (baseline) | Standard conditions |
| 50-55° | 1.2x (slower) | Reduced evaporation |
| < 50° | 1.4x (slower) | Low evaporation rate |

### Humidity
Lower humidity creates stronger vapor pressure gradient, accelerating drying.

| Humidity (%) | Drying Modifier | Reasoning |
|-------------|----------------|-----------|
| < 40% | 0.75x (faster) | Very dry air pulls moisture rapidly |
| 40-50% | 0.85x (faster) | Dry air accelerates drying |
| 50-70% | 1.0x (baseline) | Standard conditions |
| 70-80% | 1.2x (slower) | Humid air slows evaporation |
| > 80% | 1.35x (slower) | Saturated air, minimal drying |

### Wind Speed
Wind removes moisture-saturated air from rock surface, enabling faster evaporation.

| Wind Speed (mph) | Drying Modifier | Reasoning |
|-----------------|----------------|-----------|
| 5-15 | 0.8x (faster) | Ideal wind for surface drying |
| 15-25 | 0.9x (faster) | Moderate beneficial wind |
| 3-5 | 1.0x (baseline) | Light wind |
| < 3 | 1.1x (slower) | Stagnant air reduces drying |
| > 25 | 1.0x (baseline) | Strong wind (no additional benefit) |

### Cloud Cover
Direct sunlight significantly accelerates drying through radiant heating.

| Cloud Cover (%) | Drying Modifier | Reasoning |
|----------------|----------------|-----------|
| < 30% | 0.8x (faster) | Full sun, maximum radiant heating |
| 30-50% | 0.9x (faster) | Partly cloudy, good sun exposure |
| 50-80% | 1.0x (baseline) | Mostly cloudy |
| > 80% | 1.0x (baseline) | Overcast, no sun benefit |

### Rain Amount
More rain = longer drying time (proportional scaling from 0.1" baseline).

```
rainFactor = totalRainfall / 0.1"
if rainFactor < 0.5: rainFactor = 0.5  # Minimum (light mist)
if rainFactor > 3.0: rainFactor = 3.0  # Maximum (heavy rain)
dryingTime = baseDryingTime × rainFactor
```

**Examples**:
- 0.05" rain: Factor = 0.5x (minimum)
- 0.1" rain: Factor = 1.0x (baseline)
- 0.3" rain: Factor = 3.0x (capped maximum)
- 0.5" rain: Factor = 3.0x (capped maximum)

## Sun Exposure Geometry

### Aspect (Compass Direction)
Rock face orientation relative to sun path affects total solar radiation received.

| Aspect | Multiplier | Daily Sun Hours | Reasoning |
|--------|-----------|----------------|-----------|
| South | +30% | 8-10 hours | Maximum sun in northern hemisphere |
| West | +15% | 6-8 hours | Strong afternoon sun |
| East | +5% | 4-6 hours | Morning sun only |
| North | -15% | 2-4 hours | Minimal direct sunlight |

**Algorithm**:
```
aspectBonus = (southPercent × 0.30) + (westPercent × 0.15) +
              (eastPercent × 0.05) + (northPercent × -0.15)
```

### Rock Angle
Slope affects both water runoff and sun exposure angle.

| Angle Type | Multiplier | Reasoning |
|-----------|-----------|-----------|
| Slab (< 80°) | +20% faster | Water runs off, perpendicular to sun |
| Vertical (80-100°) | Baseline | Standard wall climbing |
| Overhang (> 100°) | -10% slower | Water stays, angled away from sun |

### Tree Coverage
Forest canopy blocks sunlight and increases humidity through transpiration.

| Tree Coverage | Drying Modifier | Reasoning |
|--------------|----------------|-----------|
| 0-25% | 1.0x (baseline) | Open area, full sun |
| 25-50% | 0.9x (slower) | Partial shade |
| 50-75% | 0.8x (slower) | Heavy shade |
| 75-100% | 0.7x (slower) | Dense forest, minimal sun |

### Combined Sun Exposure Factor
The final sun exposure multiplier combines all three geometry factors:

```
sunFactor = 1.0 + aspectBonus + angleBonus
sunFactor = sunFactor × treeCoverageModifier
sunFactor = clamp(sunFactor, 0.5, 1.5)  # Safety bounds
```

**Example Calculation**:

Location: Money Creek (35% S, 25% W, 25% E, 15% N, 40% slab, 20% overhang, 60% trees)

```
aspectBonus = (0.35 × 0.30) + (0.25 × 0.15) + (0.25 × 0.05) + (0.15 × -0.15)
            = 0.105 + 0.0375 + 0.0125 - 0.0225
            = 0.1325

angleBonus = (0.40 × 0.20) + (0.20 × -0.10)
           = 0.08 - 0.02
           = 0.06

sunFactor = 1.0 + 0.1325 + 0.06 = 1.1925

treeCoverageModifier = 0.8 (50-75% range, 60% coverage)

finalSunFactor = 1.1925 × 0.8 = 0.954
```

Result: 95% of baseline drying speed due to heavy tree coverage offsetting good aspect/angle.

## Seepage Risk

Some locations have persistent moisture from:
- Groundwater seepage
- Snowmelt above the climbing area
- Water table intersecting rock face
- Springs or creek spray

Locations flagged with `has_seepage_risk = true` receive **+40% drying time** modifier.

**Current Flagged Locations**:
- Treasury (high alpine, snowmelt)
- Skykomish - Paradise (high elevation, creek proximity)
- Bellingham (groundwater seepage in sandstone)

## Time-Weighted Drying Calculation

### Concept
Instead of assuming constant drying rate, the algorithm calculates "effective drying hours" by summing hourly drying power for each hour since rain stopped.

### Hourly Drying Power
Each hour gets a power multiplier based on conditions (1.0 = baseline):

```
power = 1.0
power × temperature_modifier
power × wind_modifier
power × humidity_modifier
power × sun_exposure_modifier
power × cloud_cover_modifier
```

### Examples

**Hour 1**: 75°F, 35% humidity, 12mph wind, 20% clouds, good sun
```
power = 1.0 × 1.3 (temp) × 1.25 (wind) × 1.3 (humidity) × 1.2 (sun) × 1.2 (clouds)
      = 1.0 × 1.3 × 1.25 × 1.3 × 1.2 × 1.2
      = 2.92 effective hours per actual hour
```

**Hour 2**: 50°F, 80% humidity, 2mph wind, 90% clouds, poor sun
```
power = 1.0 × 0.6 (temp) × 0.85 (wind) × 0.6 (humidity) × 0.85 (sun) × 0.85 (clouds)
      = 0.22 effective hours per actual hour
```

### Drying Progress
```
totalEffectiveHours = sum of all hourly powers since rain stopped
actualHoursElapsed = time since rain stopped
dryingProgress = totalEffectiveHours / actualHoursElapsed
```

If `dryingProgress >= 1.0`, rock is fully dry.

Otherwise:
```
remainingDryingTime = requiredDryingTime × (1.0 - dryingProgress)
```

## Algorithm Flow

### 1. Check Current Precipitation
```
if currently_raining:
    status = "poor" (or "critical" for wet-sensitive)
    estimate drying time from current precipitation rate
    return wet status
```

### 2. Find Last Rain Event
```
scan historical weather backwards
group contiguous precipitation readings
calculate event totals, duration, rates
```

### 3. No Recent Rain
```
if no rain event found:
    status = "good"
    confidence = high (80-95)
    return dry status
```

### 4. Calculate Required Drying Time
```
base = rock_type.base_drying_hours
adjust for: rain_amount, porosity, temperature, humidity,
            wind, clouds, sun_exposure, seepage, wet_sensitive
requiredTime = base × all_modifiers
```

### 5. Calculate Time-Weighted Progress
```
for each hour since rain stopped:
    calculate hourly_drying_power
    sum effective_drying_hours

progress = effective_hours / actual_hours
remaining = required × (1.0 - progress)
```

### 6. Determine Status
```
if remaining <= 0:
    status = "good" (dry)
elif remaining < 50% of required:
    status = "fair" (drying)
else:
    status = "poor" (wet)

if wet_sensitive and wet:
    status = "critical" (DO NOT CLIMB)
```

### 7. Calculate Confidence
```
confidence = 75 (baseline)
adjust for: data_completeness, sun_exposure_data, time_since_rain,
            weather_stability, rock_sensitivity
clamp to 20-95 range
```

## Data Sources

### Rock Type Properties
- **Porosity**: Literature values from geology handbooks and climbing guides
- **Base Drying Times**: Empirical data from climbing community observations
- **Wet Sensitivity**: Known from climbing ethics and rock preservation science

### Location Profiles
- **Sun Exposure**: Estimated from topographic maps, satellite imagery, and local knowledge
- **Seepage Risk**: Derived from guidebook descriptions and climber reports
- **Rock Types**: Mapped from geological surveys and climbing area descriptions

### Weather Data
- **Current/Forecast**: Open-Meteo API (European Centre for Medium-Range Weather Forecasts models)
- **Historical**: Open-Meteo historical weather database
- **Temporal Resolution**: Hourly readings for accurate rain event analysis

## Validation and Calibration

### Field Testing Needed
The algorithm should be validated against:
1. Actual climbing conditions at known locations
2. Local climber knowledge and experience
3. Direct rock moisture measurements (if available)

### Calibration Parameters
Adjustable multipliers for field calibration:
- Temperature thresholds and modifiers
- Humidity effect curves
- Sun exposure weighting factors
- Tree coverage impact
- Rock type base drying hours

### Known Limitations
1. **Microclimate Variation**: Algorithm uses single weather reading per location
2. **Rock Texture**: Surface texture affects drying but not captured in model
3. **Crack Systems**: Deep cracks dry slower than exposed faces
4. **Seasonal Factors**: Algorithm doesn't account for seasonal sun angle changes
5. **Wind Direction**: Only speed considered, not interaction with aspect

## Future Enhancements

### Short Term
1. Add seasonal sun angle adjustments
2. Implement caching for rain event analysis (performance optimization)
3. Collect user feedback on accuracy
4. Fine-tune confidence scoring based on real-world data

### Long Term
1. Machine learning calibration from user reports
2. Microclimate modeling using elevation and terrain
3. Rock texture factors (smooth vs. featured)
4. Crack depth and orientation modeling
5. Real-time moisture sensor integration (if available)

## References

### Scientific Literature
- Evaporation and vapor pressure principles: "The Physics of Atmosphere" (Wallace & Hobbs, 2006)
- Rock porosity and water absorption: "Engineering Geology" (Waltham, 2009)
- Solar radiation and aspect: "Mountain Weather and Climate" (Barry, 2008)

### Climbing Resources
- Rock drying times: Mountain Project community knowledge base
- Wet-sensitive rock ethics: Access Fund rock damage guidelines
- Location-specific info: Various regional climbing guidebooks

### Data Sources
- Weather: Open-Meteo API (https://open-meteo.com)
- Topography: USGS Digital Elevation Models
- Rock types: USGS Geological Survey maps

---

**Version**: 1.0
**Last Updated**: December 2024
**Maintainer**: Woulder Development Team

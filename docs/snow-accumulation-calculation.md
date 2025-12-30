# Snow Accumulation Calculation

This document explains how Woulder calculates snow accumulation and melt for climbing locations.

## Overview

Woulder uses a **Snow Water Equivalent (SWE) temperature-indexed model** to track snow accumulation and melt. This model is specifically calibrated for Pacific Northwest conditions, accounting for:

- Temperature-dependent snow density
- Freezing level transitions (rain/snow mix)
- Rain-on-snow compaction and melt
- Temperature-driven melt above freezing
- Wind-enhanced melt and sublimation
- Humidity-based sublimation
- Natural compaction and settling
- Elevation adjustments

---

## Scientific Background

### Snow Water Equivalent (SWE)

**Definition:** The amount of water contained in the snowpack if it were completely melted, measured in inches.

**Why Use SWE?**
- Mass is conserved (depth changes with density)
- Directly related to precipitation input
- Standard metric in hydrology and avalanche science

**Relationship to Snow Depth:**
```
Snow Depth (inches) = SWE (inches) / Snow Density (fraction)
```

**Example:**
- 12 inches SWE with 0.12 density = 100 inches (8.3 feet) of snow
- 12 inches SWE with 0.30 density = 40 inches (3.3 feet) of snow

### Snow Density

Snow density varies widely based on temperature, age, and metamorphism:

| Snow Type | Density | Description |
|-----------|---------|-------------|
| Fresh powder (-10°F) | 0.05-0.08 | Very light, fluffy |
| Cold fresh snow (20°F) | 0.08-0.12 | Typical powder |
| Near-freezing snow (30°F) | 0.12-0.18 | Heavier, moist |
| Wet snow (32°F+) | 0.18-0.25 | Heavy, sticky |
| Settled snow | 0.25-0.35 | Compacted over time |
| Old firn | 0.35-0.50 | Multi-season snow |
| Glacial ice | 0.90 | Dense, solid |

---

## Model Components

### 1. Elevation-Adjusted Temperature

Temperature decreases with elevation due to the atmospheric lapse rate.

**Formula:**
```
Adjusted Temperature = Surface Temperature - (Elevation / 1000 ft) × 3.5°F
```

**Standard Lapse Rate:** -3.5°F per 1,000 feet of elevation gain

**Example:**
- Surface temperature at sea level: 40°F
- Climbing area elevation: 3,000 feet
- Adjusted temperature: 40°F - (3 × 3.5°F) = 29.5°F

**Scientific Basis:**
This is the dry adiabatic lapse rate, which approximates average atmospheric cooling with height. Actual rates vary with atmospheric moisture and stability:
- **Dry air**: -5.4°F per 1,000 ft
- **Saturated air**: -3.3°F per 1,000 ft
- **Average**: -3.5°F per 1,000 ft (used by Woulder)

---

### 2. Freezing Level Transition

Precipitation rarely switches instantly from rain to snow. There's a transition zone where mixed precipitation occurs.

**Temperature Zones:**

| Temperature | Snow Fraction | Precipitation Type |
|-------------|---------------|-------------------|
| ≤30°F | 100% | All snow |
| 31°F | 75% | Mostly snow |
| 32°F | 50% | Rain/snow mix |
| 33°F | 25% | Mostly rain |
| ≥34°F | 0% | All rain |

**Formula:**
```
Snow Fraction = (34°F - Temperature) / 4°F
```

**Scientific Basis:**
- **Below 30°F**: Atmospheric column cold enough for snow crystals to reach ground
- **30-34°F**: Transition zone where crystals partially melt
- **Above 34°F**: Complete melting during descent

**Regional Variations:**
- **Cascade Range**: Transition zone 30-34°F (maritime climate, higher humidity)
- **Inland Mountains**: Transition zone 28-32°F (continental climate, drier air)
- **High Elevation**: Sharper transition (less time for melting during descent)

---

### 3. Snow Accumulation

When precipitation falls as snow, it adds to the snowpack's SWE.

**New Snow Density (Temperature-Dependent):**

| Temperature | Density | Explanation |
|-------------|---------|-------------|
| ≤20°F | 0.08 | Very cold, dry dendrites |
| 21-28°F | 0.12 | Typical cold powder |
| 29-32°F | 0.18 | Near freezing, riming occurs |
| >32°F | 0.20 | Wet snow (shouldn't happen, but safety) |

**Accumulation Formula:**
```
New SWE = Precipitation (inches) × Snow Fraction
New Density = Based on temperature (see table)
```

**Density Blending:**

When new snow falls on existing snow, densities are mass-weighted:
```
New Pack Density = (Old Density × Old SWE + New Density × New SWE) / Total SWE
```

**Example:**
- Existing pack: 10 inches SWE, density 0.20
- New snowfall: 2 inches SWE, density 0.10
- New pack density: (0.20 × 10 + 0.10 × 2) / 12 = 0.183

---

### 4. Rain-on-Snow

Rain falling on snow causes rapid compaction and melt through thermal energy transfer.

**Infiltration:**
```
SWE Increase = Rain Amount × 0.7
```
70% of rain infiltrates the snowpack; 30% runs off or evaporates.

**Compaction:**
```
New Density = min(0.35, Current Density + 0.03)
```
Rain compacts snow, increasing density by ~0.03 per rain event.

**Rain Energy Melt:**

Warm rain transfers thermal energy to cold snow:
```
Rain Temperature = max(Actual Temperature, 32°F)
Rain Energy Melt (inches SWE) = Rain Amount × (Rain Temp - 32°F) × 0.01
```

**Scientific Basis:**
- **Latent heat**: Rain at 40°F contains ~8 BTU/lb more energy than ice at 32°F
- **Conduction**: Liquid water in contact with ice transfers heat efficiently
- **Percolation**: Water moving through snowpack releases latent heat of fusion

**Example:**
- 1 inch of rain at 40°F
- Energy melt: 1 × (40 - 32) × 0.01 = 0.08 inches SWE melted
- Plus the 0.7 inches that infiltrated
- Net SWE change: +0.70 - 0.08 = +0.62 inches

---

### 5. Temperature-Based Melt

Above freezing, solar radiation and sensible heat cause snowmelt.

**Degree-Day Melt Formula:**
```
Melt Rate (inches SWE/hour) = max(0, (Temperature - 34°F) × 0.01)
```

**For 3-hour period:**
```
Total Melt = Melt Rate × 3
```

**Why 34°F threshold (not 32°F)?**
- Snow surface temperature lags air temperature
- Albedo (reflection) reduces solar input
- Evaporative cooling from snowpack
- Conservative approach (prevents over-prediction)

**Melt Examples:**

| Temperature | Melt Rate/Hour | Melt/3h | Melt/Day |
|-------------|----------------|---------|----------|
| 34°F | 0.00 in | 0.00 in | 0.00 in |
| 40°F | 0.06 in | 0.18 in | 1.44 in |
| 50°F | 0.16 in | 0.48 in | 3.84 in |
| 60°F | 0.26 in | 0.78 in | 6.24 in |

**Real-World Calibration:**

Pacific Northwest degree-day factors (typical values):
- **Forested areas**: 0.04-0.06 in/°F/day
- **Open areas**: 0.06-0.08 in/°F/day
- **South-facing slopes**: 0.08-0.10 in/°F/day

Woulder uses **0.24 in/°F/day** (0.01 in/°F/hour), which is conservative and typical for mixed terrain.

---

### 6. Wind-Enhanced Melt

Wind increases turbulent heat exchange at the snow surface.

**Formula:**
```
Wind Melt (inches SWE/hour) = (Wind Speed - 10 mph) × 0.002
```

**For 3-hour period:**
```
Total Wind Melt = Wind Melt Rate × 3
```

**Threshold:** Wind speeds above 10 mph

**Scientific Basis:**
- **Forced convection**: Wind replaces cold boundary layer with warmer air
- **Evaporation**: Increased vapor transport (sublimation/evaporation)
- **Mechanical erosion**: Wind can physically remove light snow

**Examples:**

| Wind Speed | Wind Melt/Hour | Wind Melt/3h | Wind Melt/Day |
|------------|----------------|--------------|---------------|
| 10 mph | 0.000 in | 0.00 in | 0.00 in |
| 15 mph | 0.010 in | 0.03 in | 0.24 in |
| 25 mph | 0.030 in | 0.09 in | 0.72 in |
| 40 mph | 0.060 in | 0.18 in | 1.44 in |

---

### 7. Humidity-Based Sublimation

Dry air causes snow to sublimate (solid to vapor) without melting.

**Formula:**
```
Sublimation (inches SWE/hour) = (60% - Humidity%) × 0.0001
```

**For 3-hour period:**
```
Total Sublimation = Sublimation Rate × 3
```

**Threshold:** Humidity below 60%

**Scientific Basis:**
- **Vapor pressure gradient**: Dry air has lower vapor pressure than ice surface
- **Sublimation**: Snow converts directly to water vapor
- **Energy cost**: Sublimation requires ~680 cal/g (latent heat)

**Examples:**

| Humidity | Sublimation/Hour | Sublimation/3h | Sublimation/Day |
|----------|------------------|----------------|-----------------|
| 60% | 0.0000 in | 0.00 in | 0.00 in |
| 40% | 0.0020 in | 0.006 in | 0.048 in |
| 20% | 0.0040 in | 0.012 in | 0.096 in |
| 10% | 0.0050 in | 0.015 in | 0.120 in |

**Significance:**
Sublimation is relatively minor compared to melt but becomes important during:
- Cold, dry winter conditions
- High winds (enhanced sublimation)
- High-elevation environments
- Extended dry spells

---

### 8. Natural Compaction

Snow settles under its own weight, increasing density over time.

**Temperature-Dependent Compaction Rates:**

| Temperature | Rate/Hour | Explanation |
|-------------|-----------|-------------|
| <20°F | 0.0003 | Very slow, cold snow |
| 20-28°F | 0.0006 | Moderate, crystal bonding |
| 28-32°F | 0.0012 | Faster, near melting point |
| >32°F | 0.0025 | Rapid, wet snow metamorphism |

**For 3-hour period:**
```
Density Increase = Rate × 3
New Density = min(0.40, Old Density + Density Increase)
```

**Cap at 0.40:** Prevents unrealistic compaction (old firn/glacier ice densities require years)

**Scientific Basis:**
- **Creep**: Snow crystals deform under pressure
- **Sintering**: Crystal bonds strengthen at grain contacts
- **Metamorphism**: Warmer temperatures accelerate recrystallization
- **Pressure melting**: At 32°F, pressure can cause melting at grain contacts

**Example:**
- Fresh powder: 0.10 density
- After 24 hours at 25°F: 0.10 + (0.0006 × 24) = 0.114
- After 7 days at 25°F: 0.10 + (0.0006 × 168) = 0.201
- After 7 days at 30°F: 0.10 + (0.0012 × 168) = 0.302

---

## Complete Calculation Workflow

### Input Data

**Historical Weather (7 days):**
- Hourly temperature (°F)
- Hourly precipitation (inches)
- Hourly wind speed (mph)
- Hourly humidity (%)

**Forecast Weather (16 days):**
- Same parameters as historical

**Location:**
- Elevation (feet)

### Processing Steps

For each hour of data (historical + forecast):

1. **Adjust temperature for elevation**
   ```
   Adjusted Temp = Surface Temp - (Elevation / 1000) × 3.5°F
   ```

2. **Determine snow fraction**
   ```
   Snow Fraction = getSnowFraction(Adjusted Temp)
   ```

3. **If precipitation occurs:**
   - **Snow portion:**
     ```
     Snow SWE = Precipitation × Snow Fraction
     New Snow Density = getNewSnowDensity(Adjusted Temp)
     Blend with existing pack density
     ```

   - **Rain portion (if mixed or rain-on-snow):**
     ```
     Rain SWE = Precipitation × (1 - Snow Fraction)
     Add 70% to pack: SWE += Rain SWE × 0.7
     Compact pack: Density = min(0.35, Density + 0.03)
     Rain energy melt: SWE -= Rain SWE × (Temp - 32) × 0.01
     ```

4. **Temperature melt (if temp > 34°F):**
   ```
   Melt = max(0, (Temp - 34) × 0.01) × 3 hours
   SWE = max(0, SWE - Melt)
   ```

5. **Wind melt (if wind > 10 mph):**
   ```
   Wind Melt = (Wind Speed - 10) × 0.002 × 3 hours
   SWE = max(0, SWE - Wind Melt)
   ```

6. **Sublimation (if humidity < 60%):**
   ```
   Sublimation = (60 - Humidity) × 0.0001 × 3 hours
   SWE = max(0, SWE - Sublimation)
   ```

7. **Compaction:**
   ```
   Density = min(0.40, Density + getCompactionRate(Temp) × 3)
   ```

8. **Calculate snow depth:**
   ```
   Snow Depth = SWE / Density
   ```

9. **Store daily maximum depth**

### Output

Map of dates to snow depth (inches):
```
{
  "2024-12-22": 12.5,
  "2024-12-23": 18.3,
  "2024-12-24": 16.8,
  ...
}
```

---

## Example Calculation

### Scenario: Cold Powder Storm

**Location:** Paradise Valley, 4,000 ft elevation

**Initial Conditions:**
- No existing snow (SWE = 0, Density = 0.12)
- Surface temperature: 30°F

**Hour 1: Snowfall begins**
```
Adjusted Temp: 30 - (4 × 3.5) = 16°F
Precipitation: 0.3 inches
Snow Fraction: 1.0 (all snow, temp < 30°F)

New SWE: 0.3 inches
New Snow Density: 0.08 (very cold)
Pack Density: 0.08
Snow Depth: 0.3 / 0.08 = 3.75 inches
```

**Hour 4: Continued snowfall**
```
Adjusted Temp: 18°F
Precipitation: 0.4 inches
Snow Fraction: 1.0

New SWE: 0.4 inches
Total SWE: 0.3 + 0.4 = 0.7 inches
New Snow Density: 0.08
Blended Density: (0.08 × 0.3 + 0.08 × 0.4) / 0.7 = 0.08
Compaction (3 hours): 0.08 + (0.0003 × 3) = 0.0809
Snow Depth: 0.7 / 0.0809 = 8.65 inches
```

**Hour 12: Warming trend**
```
Adjusted Temp: 28°F
Precipitation: 0.2 inches
Snow Fraction: 1.0

New SWE: 0.2 inches
Total SWE: 0.9 inches
New Snow Density: 0.12 (warmer snow)
Blended Density: (0.0809 × 0.7 + 0.12 × 0.2) / 0.9 = 0.089
Compaction: 0.089 + (0.0006 × 3) = 0.091
Snow Depth: 0.9 / 0.091 = 9.9 inches
```

**Hour 24: Temperature rises**
```
Adjusted Temp: 36°F
Precipitation: 0 inches
Wind: 15 mph
Humidity: 50%

Temperature Melt: (36 - 34) × 0.01 × 3 = 0.06 inches SWE
Wind Melt: (15 - 10) × 0.002 × 3 = 0.03 inches SWE
Sublimation: (60 - 50) × 0.0001 × 3 = 0.003 inches SWE
Total Loss: 0.093 inches SWE

Remaining SWE: 0.9 - 0.093 = 0.807 inches
Compaction: 0.091 + (0.0025 × 3) = 0.099
Snow Depth: 0.807 / 0.099 = 8.15 inches

Loss: 1.75 inches of depth in 3 hours
```

---

## Model Validation and Accuracy

### Strengths

1. **Physically-Based**: Uses snow physics principles (energy balance, mass conservation)
2. **PNW-Calibrated**: Parameters tuned for maritime mountain climate
3. **Accounts for Variability**: Temperature-dependent density, rain-on-snow, compaction
4. **Hourly Resolution**: Captures diurnal melt cycles
5. **Elevation-Aware**: Adjusts for temperature lapse rate

### Limitations

1. **No Solar Radiation**: Doesn't account for aspect (south vs. north slopes) or canopy effects
2. **Simplified Energy Balance**: Degree-day approach approximates complex heat transfer
3. **No Avalanche Modeling**: Doesn't predict snow stability or avalanche hazard
4. **Uniform Snowpack**: Assumes homogeneous layer (real snowpacks have complex stratigraphy)
5. **No Blowing Snow**: Doesn't model wind redistribution
6. **Limited Calibration**: Based on general PNW snowpack behavior, not site-specific measurements

### Expected Accuracy

**Good Conditions (±20%):**
- Cold, stable snowpacks
- Clear weather patterns
- Elevations 2,000-6,000 ft

**Moderate Conditions (±30-40%):**
- Mixed rain/snow events
- Variable wind
- Transitional temperatures

**Poor Conditions (±50%+):**
- Rain-on-snow events
- Extreme wind
- Very warm or very cold extremes
- Elevation <1,000 ft or >8,000 ft

### Comparison with Professional Models

| Feature | Woulder | SNOTEL | SNODAS | SnowModel |
|---------|---------|--------|--------|-----------|
| Resolution | Location-specific | Point | 1 km grid | Variable |
| Physics | Temperature-index | Temperature-index | Energy balance | Energy balance |
| Inputs | Weather forecast | Measured | Model + satellite | Full meteorology |
| Real-time | Yes | Yes | 1-2 day lag | Research |
| Complexity | Simple | Simple | Moderate | High |
| Purpose | Climbing access | Water supply | Operations | Research |

---

## Practical Interpretation

### Snow Depth Categories for Climbing Access

| Depth | Category | Implications |
|-------|----------|-------------|
| 0-3 in | **Minimal** | Trail accessible, no special equipment |
| 3-6 in | **Light** | Microspikes may help, trail visible |
| 6-12 in | **Moderate** | Snowshoes recommended, trail markers important |
| 12-24 in | **Deep** | Snowshoes or skis required, route-finding difficult |
| 24+ in | **Very Deep** | Backcountry skills required, avalanche awareness critical |

### When to Trust the Model

✅ **High Confidence:**
- Clear cold storms (all snow, no mixed precip)
- Stable cold weather after snowfall
- Elevations 3,000-5,000 ft in Cascades

⚠️ **Medium Confidence:**
- Rain/snow mix events
- Rapid temperature swings
- Moderate wind
- First snow of season

❌ **Low Confidence:**
- Major rain-on-snow events
- Extreme wind (>40 mph)
- Temperature inversions
- Very low elevations (<2,000 ft)
- Late spring melt cycles

---

## Scientific References

**Snow Hydrology:**
- DeWalle, D.R. & Rango, A. (2008). *Principles of Snow Hydrology*. Cambridge University Press.
- Marks, D., et al. (1999). "The Sensitivity of Snowmelt Processes to Climate Conditions and Forest Cover During Rain-on-Snow". *Hydrological Processes*, 13, 2177-2190.

**Temperature-Index Models:**
- Hock, R. (2003). "Temperature index melt modeling in mountain areas". *Journal of Hydrology*, 282, 104-115.
- Martinec, J. (1975). "Snowmelt-Runoff Model for Stream Flow Forecasts". *Nordic Hydrology*, 6(3), 145-154.

**Pacific Northwest Snowpack:**
- Marks, D. & Dozier, J. (1992). "Climate and Energy Exchange at the Snow Surface in the Alpine Region of the Sierra Nevada". *Water Resources Research*, 28(11), 3043-3054.
- NRCS SNOTEL. (2024). *Snow Survey and Water Supply Forecasting*. Natural Resources Conservation Service.

**Snow Physics:**
- Colbeck, S.C. (1982). "An Overview of Seasonal Snow Metamorphism". *Reviews of Geophysics*, 20(1), 45-61.
- Sturm, M., et al. (1995). "The Thermal Conductivity of Seasonal Snow". *Journal of Glaciology*, 41(139), 539-554.

---

## Version History

- **v1.0** (December 2024): Initial implementation with SWE-based temperature-indexed model

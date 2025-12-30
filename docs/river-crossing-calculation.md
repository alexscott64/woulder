# River Crossing Safety Calculation

This document explains how Woulder determines river crossing safety levels for climbing approach trails.

## Overview

Many climbing areas in the Pacific Northwest require crossing rivers or creeks on approach trails. Woulder provides real-time safety assessments based on current streamflow data from the USGS (United States Geological Survey) Water Data system.

### Safety Levels

Woulder categorizes river crossings into three levels:

- **Safe**: Flow is at or below the safe crossing threshold
- **Caution**: Flow exceeds safe threshold but is below the dangerous threshold
- **Unsafe**: Flow exceeds the dangerous threshold - do not attempt

---

## Data Sources

### USGS Stream Gauges

The USGS maintains a network of over 8,500 real-time stream gauges across the United States. These gauges continuously monitor:

- **Discharge (Flow Rate)**: Measured in cubic feet per second (CFS)
- **Gauge Height**: Measured in feet above an arbitrary reference point
- **Timestamp**: When the measurement was taken (typically updated every 15 minutes)

**API Endpoint:** `https://waterservices.usgs.gov/nwis/iv/`

**Parameters Monitored:**
- `00060`: Discharge (CFS)
- `00065`: Gauge height (feet)

### Data Reliability

- Updated every 15 minutes
- Real-time telemetry via satellite or cellular
- Quality-controlled by USGS hydrologists
- Historical data available for validation

---

## Flow Estimation Methods

Not all river crossings have gauges directly at the crossing point. Woulder uses two scientifically-validated methods to estimate flow at ungauged locations.

### Method 1: Drainage Area Ratio Method

This is the preferred method when drainage area data is available for both the gauge and the crossing location.

**Formula:**
```
Q_crossing = Q_gauge × (A_crossing / A_gauge)^0.7
```

Where:
- `Q_crossing` = Estimated flow at crossing (CFS)
- `Q_gauge` = Measured flow at gauge (CFS)
- `A_crossing` = Drainage area at crossing (square miles)
- `A_gauge` = Drainage area at gauge (square miles)
- `0.7` = Scaling exponent for mountainous terrain

**Scientific Basis:**

The drainage area ratio method is based on the principle that streamflow is proportional to the watershed area that drains to that point. The exponent of 0.7 (instead of 1.0) accounts for:

1. **Non-linear scaling**: Larger watersheds don't produce proportionally more flow
2. **Storage effects**: Larger basins have more natural storage (lakes, wetlands)
3. **Travel time**: Water takes longer to reach the outlet in larger basins
4. **Evapotranspiration**: Larger basins lose more water to evaporation

**Exponent Selection:**
- **Mountainous terrain** (Cascade Range): 0.7
- **Flat terrain**: 0.8-0.9
- **Urban areas**: 0.6-0.7 (flashier response)

**Example: Money Creek at Skykomish**

Crossing drainage area: 18.0 sq mi
Gauge drainage area: 355.0 sq mi
Gauge flow: 1,000 CFS

```
Ratio = 18.0 / 355.0 = 0.0507
Scale Factor = 0.0507^0.7 = 0.136
Estimated Flow = 1,000 × 0.136 = 136 CFS
```

**Accuracy:** Typically within 20-30% of actual flow when watersheds have similar characteristics.

### Method 2: Simple Flow Divisor

This empirical method is used when drainage areas are unknown but local knowledge provides a ratio.

**Formula:**
```
Q_crossing = Q_gauge / divisor
```

**Example: North Fork Skykomish at Index**

Some locations use a simple divisor based on field observations and historical correlations:
```
Gauge flow: 2,000 CFS
Divisor: 2.0
Estimated flow = 2,000 / 2 = 1,000 CFS
```

**When to Use:**
- Drainage areas unavailable
- Simple relationship observed over time
- Regular field validation available

**Limitations:**
- Less accurate during extreme events (floods, droughts)
- Assumes constant ratio (not always valid)
- Should be calibrated with field measurements

---

## Safety Thresholds

Each river crossing has two critical thresholds determined by:

1. **Field Assessment**: On-site evaluation of crossing conditions
2. **River Morphology**: Width, depth, gradient, substrate
3. **Historical Data**: Observations from climbers and local knowledge
4. **Safety Margin**: Conservative values to account for variability

### Safe Crossing Threshold

**Definition:** The maximum flow rate at which an experienced hiker can safely cross using standard techniques (wading, using trekking poles, solid footing).

**Determination Factors:**
- Water depth at typical crossing point
- Current velocity
- Streambed composition (bedrock, cobbles, gravel)
- Channel width
- Presence of natural aids (rocks, logs)

**Typical Values:**
- Small creeks (5-15 feet wide): 30-80 CFS
- Medium streams (15-30 feet wide): 80-300 CFS
- Large rivers (30+ feet wide): 300-1,000 CFS

### Caution Crossing Threshold

**Definition:** The flow rate above which crossing becomes hazardous for most hikers. Special techniques (linking arms, ropes, scouting alternate routes) may be required.

**Safety Margin:** Typically 150-200% of safe threshold

**Risk Factors at Caution Level:**
- Deeper water (thigh to waist level)
- Stronger current
- Reduced visibility of streambed
- Cold water temperature (hypothermia risk)
- Difficulty returning if crossing fails

**Typical Values:**
- Safe threshold × 1.5 to 2.0

### Example: Money Creek

| Threshold | Flow (CFS) | Description |
|-----------|------------|-------------|
| Safe | 60 | Ankle to mid-shin depth, easy crossing |
| Caution | 90 | Shin to knee depth, careful footing required |
| Unsafe | >90 | Knee+ depth, strong current, do not attempt |

---

## Safety Assessment Algorithm

### Step 1: Obtain Current Flow

1. Query USGS API for gauge data
2. Parse discharge value (CFS)
3. Check data timestamp (reject if >6 hours old)

### Step 2: Estimate Flow at Crossing

If gauge is not at crossing location:

**Priority 1:** Use drainage area ratio method if available
```go
if river.DrainageAreaSqMi != nil && river.GaugeDrainageAreaSqMi != nil {
    actualFlow = gaugeFlow × (crossing_area / gauge_area)^0.7
}
```

**Priority 2:** Use flow divisor if available
```go
else if river.FlowDivisor != nil {
    actualFlow = gaugeFlow / divisor
}
```

**Priority 3:** Use gauge flow directly (gauge at crossing)
```go
else {
    actualFlow = gaugeFlow
}
```

### Step 3: Determine Safety Level

```go
if actualFlow <= SafeCrossingCFS {
    status = "safe"
    message = "Safe to cross. Flow is X% of safe threshold."
    isSafe = true
}
else if actualFlow <= CautionCrossingCFS {
    status = "caution"
    message = "Use caution. Flow is X% of safe threshold."
    isSafe = false
}
else {
    status = "unsafe"
    message = "Unsafe to cross! Flow is X% of safe threshold."
    isSafe = false
}
```

### Step 4: Display to User

- Show current flow (CFS)
- Show percentage of safe threshold
- Display status with color coding:
  - **Green**: Safe
  - **Yellow**: Caution
  - **Red**: Unsafe
- Include gauge measurement timestamp
- Note if flow is estimated vs. measured

---

## Understanding Streamflow

### What is CFS?

**CFS (Cubic Feet per Second)** is the volume of water passing a point per second.

**Visualization:**
- 100 CFS ≈ Filling a bathtub in less than 1 second
- 500 CFS ≈ Filling a swimming pool in 30 seconds
- 1,000 CFS ≈ Filling an Olympic pool in 3 minutes

**Relationship to Depth & Velocity:**

For a rectangular channel:
```
CFS = Width (ft) × Average Depth (ft) × Velocity (ft/s)
```

**Example:**
- Width: 20 feet
- Depth: 2 feet (knee-deep)
- Velocity: 2.5 ft/s
- Flow: 20 × 2 × 2.5 = **100 CFS**

### Seasonal Patterns

**Spring Snowmelt (May-June):**
- Highest flows of the year
- Peak typically in late May/early June
- Can be 10-50× higher than summer base flow
- Many crossings impossible during peak melt

**Summer Base Flow (July-September):**
- Lowest flows of the year
- Groundwater-fed streams remain stable
- Most crossings safe or low caution
- Rain events cause temporary spikes

**Fall Rains (October-December):**
- Moderate to high flows
- Atmospheric rivers can cause rapid increases
- Variable conditions week-to-week

**Winter Low Flow (January-April):**
- Precipitation stored as snow
- Low flows except during rain-on-snow events
- Coldest water temperatures

---

## Safety Guidelines

### Never Cross If:

1. **Flow exceeds unsafe threshold**: Even by a small amount
2. **Water is opaque/muddy**: Cannot see bottom = depth unknown
3. **Heavy debris in water**: Logs, branches indicate high flow upstream
4. **You're alone**: Always cross rivers with a partner
5. **It's getting dark**: Allow time for alternate routes
6. **You're uncertain**: Trust your instincts

### Safe Crossing Techniques

1. **Scout the crossing**: Look for shallow, wide sections
2. **Face upstream**: Maintain balance against current
3. **Use trekking poles**: Create tripod with legs
4. **Unbuckle pack hip belt**: Quick release if you fall
5. **Link arms with partner**: Mutual support
6. **Shuffle, don't step**: Keep contact with streambed
7. **Cross at angle**: Move slightly downstream
8. **Choose right time**: Cross in morning when flows lowest (snowmelt streams)

---

## Example Calculations

### Money Creek (Skykomish)

**Configuration:**
- Gauge: South Fork Skykomish at Skykomish (USGS 12131500)
- Crossing drainage: 18.0 sq mi
- Gauge drainage: 355.0 sq mi
- Safe threshold: 60 CFS
- Caution threshold: 90 CFS

**Scenario 1: Spring Runoff**
```
Gauge flow: 1,200 CFS
Ratio: 18.0 / 355.0 = 0.0507
Scale factor: 0.0507^0.7 = 0.136
Estimated crossing flow: 1,200 × 0.136 = 163 CFS

Status: UNSAFE (163 CFS > 90 CFS)
Percent of safe: 272%
Message: "Unsafe to cross! Flow is 272% of safe threshold."
```

**Scenario 2: Summer Base Flow**
```
Gauge flow: 400 CFS
Ratio: 0.0507
Scale factor: 0.136
Estimated crossing flow: 400 × 0.136 = 54 CFS

Status: SAFE (54 CFS < 60 CFS)
Percent of safe: 90%
Message: "Safe to cross. Flow is 90% of safe threshold."
```

---

## Limitations and Uncertainties

### Estimation Accuracy

**Drainage Area Method:**
- ±20-30% typical error
- Higher error during extreme events
- Assumes similar watershed characteristics
- Better for nearby gauges (<20 miles)

**Factors Affecting Accuracy:**
1. Elevation differences
2. Aspect (north vs. south facing)
3. Forest cover differences
4. Snowpack distribution
5. Recent precipitation patterns

### When Estimates Fail

Estimates become less reliable when:
- Rainfall is localized (thunderstorms)
- Rapid snowmelt events
- Rain-on-snow conditions
- Ice jams or debris dams
- Spring timing differs between watersheds

### Conservative Approach

Woulder uses conservative thresholds:
- Safe threshold set below actual ford-able flow
- Caution range provides buffer
- Percentage display shows margin
- Always err on the side of caution

---

## Scientific References

**Streamflow Estimation:**
- Ries, K.G. & Friesz, P.J. (2000). "Methods for Estimating Low-Flow Statistics for Massachusetts Streams". USGS Water-Resources Investigations Report 00-4135.
- Farmer, W.H., et al. (2014). "Multiple Regression and Machine Learning for Predicting Streamflow Statistics at Ungauged Sites". USGS Techniques and Methods.

**Drainage Area Scaling:**
- Hirsch, R.M. (1979). "An Evaluation of Some Record Reconstruction Techniques". Water Resources Research, 15(6), 1781-1790.
- Emerson, D.G., et al. (2005). "USGS Streamstats: Streamflow Statistics and Spatial Analysis Tools for Water-Resources Applications". USGS Fact Sheet 2005-3096.

**River Crossing Safety:**
- American Whitewater. (2019). *River Safety Handbook*.
- National Outdoor Leadership School. (2020). *Wilderness Medicine*.

---

## Version History

- **v1.0** (December 2024): Initial implementation with drainage area ratio and simple divisor methods

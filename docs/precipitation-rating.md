# Precipitation-Based Condition Rating

## Overview

Woulder's condition rating system now accounts for **precipitation patterns** and **surface drying conditions**, not just instantaneous rainfall amounts.

## Key Concepts

### 1. Precipitation Rate
**Definition**: The amount of rainfall per hour (inches/hour or mm/hour)

Open-Meteo provides precipitation data in 3-hour periods, so we analyze:
- **Current precipitation**: Amount in the current 3-hour window
- **Recent precipitation history**: Previous 6-9 hours (2-3 data points)

### 2. Surface Drying Conditions
Surfaces dry faster when there's:
- **Warm temperature** (>55°F / 13°C)
- **Low cloud cover** (<50%) - sunlight and UV help evaporation
- **Moderate wind** (5-20 mph / 8-32 km/h) - aids evaporation without creating spray

### 3. Precipitation Patterns

The system distinguishes between two scenarios:

#### Scenario A: Brief Isolated Rain + Good Drying
**Example**: 0.06" rain in 1 hour, then sunny/warm/windy for 3 hours

- **Rating**: May stay "Good" if drying conditions are excellent
- **Reason**: "Brief light rain (0.02in/3h, drying fast)"
- **Rock condition**: Likely dry within 1-2 hours

#### Scenario B: Persistent Drizzle + Poor Drying
**Example**: 0.02" rain every hour for 6+ hours, overcast, cool temps

- **Rating**: Downgraded to "Fair/Marginal"
- **Reason**: "Persistent drizzle (0.12in over 9h)" or "Drying slowly after rain"
- **Rock condition**: Stays wet for many hours

## Implementation Logic

### Precipitation Thresholds

```typescript
if (precipitation > 0.1") {
  // Heavy rain: >0.1" in 3h period
  → BAD condition
}
else if (precipitation > 0.05") {
  // Moderate rain: 0.05-0.1" in 3h period
  → MARGINAL condition
}
else if (precipitation > 0.01") {
  // Light rain/drizzle: 0.01-0.05" in 3h period

  if (recent 2+ periods also had rain) {
    → MARGINAL (persistent drizzle)
  }
  else if (poor drying conditions) {
    → MARGINAL (stays wet)
  }
  else {
    → Note it but don't downgrade (dries fast)
  }
}
else if (recent rain > 0.05" AND poor drying) {
  → MARGINAL (still wet from earlier)
}
```

### Drying Conditions

```typescript
hasDryingConditions =
  temperature > 55°F &&
  cloud_cover < 50% &&
  wind_speed > 5 mph &&
  wind_speed < 20 mph
```

## Terminology

### Precipitation Rate
- **0.01-0.02"/hr**: Light drizzle
- **0.02-0.05"/hr**: Light rain
- **0.05-0.1"/hr**: Moderate rain
- **>0.1"/hr**: Heavy rain

### Cumulative Precipitation
Total rainfall over a time window (e.g., "0.12in over 9 hours")

## User-Facing Messages

The system provides context-aware messages:

- ✅ **"Brief light rain (0.02in/3h, drying fast)"** - User sees it rained but conditions are improving
- ⚠️ **"Persistent drizzle (0.12in over 9h)"** - User knows surfaces stay wet despite low rate
- ⚠️ **"Light rain, poor drying (0.03in/3h)"** - User knows it won't dry quickly
- ⚠️ **"Drying slowly after rain (0.08in recently)"** - No current rain but still wet

## Real-World Examples

### Example 1: Morning Shower → Sunny Afternoon
```
Time:  6am   9am   12pm   3pm
Temp:  45°F  55°F  68°F   72°F
Rain:  0.06" 0.0"  0.0"   0.0"
Cloud: 90%   40%   10%    5%
Wind:  3mph  8mph  12mph  10mph

→ Condition at 12pm: GOOD
→ Reason: "Brief light rain, drying fast"
→ Rock likely dry by noon
```

### Example 2: All-Day Drizzle
```
Time:  6am   9am   12pm   3pm
Temp:  48°F  50°F  52°F   50°F
Rain:  0.02" 0.02" 0.02"  0.02"
Cloud: 95%   90%   95%    100%
Wind:  2mph  4mph  3mph   5mph

→ Condition at 12pm: MARGINAL
→ Reason: "Persistent drizzle (0.08in over 9h)"
→ Rock stays wet all day
```

### Example 3: Recent Rain, Poor Drying
```
Time:  6am   9am   12pm   3pm
Temp:  45°F  48°F  50°F   48°F
Rain:  0.10" 0.04" 0.0"   0.0"
Cloud: 100%  90%   85%    90%
Wind:  8mph  6mph  4mph   3mph

→ Condition at 12pm: MARGINAL
→ Reason: "Drying slowly after rain (0.14in recently)"
→ Rock still damp despite no current rain
```

## Benefits

1. **More Accurate**: Distinguishes between brief showers and persistent drizzle
2. **Context-Aware**: Considers temperature, sun, and wind for drying speed
3. **User-Friendly**: Explains *why* conditions are rated a certain way
4. **Real-World Aligned**: Matches climbers' actual experience with rock wetness

## Future Enhancements

Potential improvements:
- Rock type consideration (granite dries faster than sandstone)
- Aspect/exposure (south-facing dries faster than north-facing)
- Season-specific drying rates (summer vs winter)
- Altitude adjustment (high-alpine drying rates)

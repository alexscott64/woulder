# Pest Activity Calculation

This document explains how woulder calculates pest activity levels for outdoor climbing locations.

## Overview

woulder provides two pest activity metrics:
- **Mosquito Activity**: Focuses specifically on mosquito populations
- **Outdoor Pest Activity**: Covers flies, gnats, wasps, ants, and other nuisance insects

Both metrics are calculated on a 0-100 scale and categorized into five levels: Low, Moderate, High, Very High, and Extreme.

---

## Scientific Background

Insect activity is heavily dependent on environmental conditions. Temperature, humidity, wind, and recent rainfall all play critical roles in determining pest populations and their activity levels.

### Key Environmental Factors

1. **Temperature**: Controls insect metabolism and flight capability
2. **Humidity**: Affects survival rates and breeding success
3. **Wind**: Limits flight activity and dispersal
4. **Recent Rainfall**: Creates breeding sites (standing water for mosquitoes)
5. **Season**: Determines baseline population levels

---

## Mosquito Activity Calculation

### Temperature Gating (Primary Factor)

Mosquitoes are cold-blooded and cannot maintain activity below certain temperatures:

- **Below 50°F (10°C)**: Mosquitoes become dormant or die
  - Score capped at 5 regardless of other factors
  - Explanation: "Too cold for mosquitoes (below 50°F)"

- **50-60°F (10-15°C)**: Minimal activity
  - Base score: 10-20
  - Slow metabolism, limited flight

- **60-75°F (15-24°C)**: Moderate activity range
  - Base score scales linearly with temperature
  - Optimal for some species

- **75-85°F (24-29°C)**: Peak activity
  - Base score: 70-90
  - Maximum metabolism and aggressiveness
  - Explanation: "Optimal mosquito temperature (75-85°F)"

- **Above 90°F (32°C)**: Reduced activity
  - Heat stress reduces feeding behavior
  - Score decreases above 90°F

### Humidity Modifier

Mosquitoes require high humidity to prevent desiccation:

- **Below 30%**: Very dry conditions
  - Multiplier: 0.5 (50% reduction)
  - Explanation: "Low humidity (drying effect)"

- **30-50%**: Low humidity
  - Multiplier: 0.7 (30% reduction)
  - Mosquitoes must seek shelter

- **50-70%**: Moderate humidity
  - Multiplier: 1.0 (no effect)
  - Comfortable conditions

- **Above 70%**: High humidity
  - Multiplier: 1.2-1.3 (20-30% increase)
  - Optimal for mosquito survival
  - Explanation: "High humidity (mosquito-friendly)"

### Wind Speed Modifier

Wind physically impairs mosquito flight:

- **0-5 mph**: Calm conditions
  - Multiplier: 1.2 (20% increase)
  - Explanation: "Calm conditions"
  - Easy flight and host detection

- **5-10 mph**: Light breeze
  - Multiplier: 1.0 (no effect)
  - Minimal impact on flight

- **10-15 mph**: Moderate wind
  - Multiplier: 0.7 (30% reduction)
  - Explanation: "Moderate winds"
  - Difficult to maintain stable flight

- **Above 15 mph**: Strong wind
  - Multiplier: 0.4-0.5 (50-60% reduction)
  - Explanation: "Strong winds grounding mosquitoes"
  - Most mosquitoes seek shelter

### Recent Rainfall Modifier

Rain creates standing water, which mosquitoes require for breeding:

**Breeding Cycle Timeline:**
- Days 1-6: Eggs hatch, larvae develop
- Days 7-14: Peak emergence of adults
- After Day 14: Population declines

**Rainfall Scoring:**
- **0-0.5 inches in last 14 days**: Minimal breeding
  - Multiplier: 0.8 (20% reduction)
  - Explanation: "Low recent rainfall (limited breeding sites)"

- **0.5-1.5 inches**: Moderate breeding
  - Multiplier: 1.0 (no effect)
  - Sufficient breeding sites

- **1.5-3 inches**: High breeding
  - Multiplier: 1.3 (30% increase)
  - Explanation: "Recent rainfall (breeding sites abundant)"
  - Abundant standing water

- **Above 3 inches**: Peak breeding
  - Multiplier: 1.5 (50% increase)
  - Maximum breeding site availability

**Peak Activity Window:**
- 7-10 days after rainfall: Multiplier 1.5x
- 10-14 days after rainfall: Multiplier 1.3x
- Explanation: "Peak mosquito emergence (7-14 days post-rain)"

### Seasonal Factor

Baseline population levels vary by season:

| Month | Factor | Explanation |
|-------|--------|-------------|
| January | 0.0 | Winter dormancy |
| February | 0.0 | Winter dormancy |
| March | 0.1 | Early emergence |
| April | 0.3 | Spring awakening |
| May | 0.6 | Rapid population growth |
| June | 0.9 | Near-peak activity |
| July | 1.0 | Peak season |
| August | 1.0 | Peak season |
| September | 0.7 | Population decline |
| October | 0.3 | Late season |
| November | 0.1 | Dormancy begins |
| December | 0.0 | Winter dormancy |

**Applied as:** `score × seasonalFactor`

---

## Outdoor Pest Activity Calculation

Covers flies, gnats, wasps, ants, and other insects common at outdoor climbing areas.

### Temperature Scoring (Primary Factor)

Unlike mosquitoes, many outdoor pests remain active across a wider temperature range:

- **Below 40°F (4°C)**: Minimal activity
  - Base score: 5
  - Most insects dormant

- **40-55°F (4-13°C)**: Low activity
  - Base score: 10-25
  - Limited to hardy species

- **55-70°F (13-21°C)**: Moderate activity
  - Base score: 30-60
  - Most species active

- **70-90°F (21-32°C)**: High activity
  - Base score: 70-90
  - Peak activity for most pests
  - Explanation: "Warm temperatures (peak insect activity)"

- **Above 95°F (35°C)**: Reduced activity
  - Score caps at 85
  - Heat stress limits some species

### Humidity Modifier

Outdoor pests are more resilient to dry conditions than mosquitoes:

- **Below 30%**: Very dry
  - Multiplier: 0.7 (30% reduction)
  - Some reduction in activity

- **30-60%**: Moderate humidity
  - Multiplier: 1.0 (no effect)
  - Comfortable range

- **Above 60%**: High humidity
  - Multiplier: 1.2 (20% increase)
  - Explanation: "High humidity (favorable for insects)"
  - Optimal for most species

### Recent Rainfall Modifier

Rain provides water sources and increases food availability:

- **0-0.3 inches in last 7 days**: Dry conditions
  - Multiplier: 0.9 (10% reduction)
  - Explanation: "Dry conditions"

- **0.3-1 inch**: Moderate moisture
  - Multiplier: 1.0 (no effect)

- **1-2 inches**: Ample moisture
  - Multiplier: 1.2 (20% increase)
  - Explanation: "Recent rainfall (increased activity)"

- **Above 2 inches**: Very wet
  - Multiplier: 1.3 (30% increase)
  - Maximum food and water availability

### Seasonal Factor

Similar to mosquitoes but with broader activity windows:

| Month | Factor |
|-------|--------|
| January | 0.1 |
| February | 0.1 |
| March | 0.3 |
| April | 0.5 |
| May | 0.7 |
| June | 0.9 |
| July | 1.0 |
| August | 1.0 |
| September | 0.8 |
| October | 0.5 |
| November | 0.2 |
| December | 0.1 |

---

## Activity Level Categories

Final scores (0-100) are converted to descriptive levels:

| Score Range | Level | Description |
|-------------|-------|-------------|
| 0-19 | **Low** | Minimal pest presence, unlikely to be bothersome |
| 20-39 | **Moderate** | Noticeable but manageable pest activity |
| 40-59 | **High** | Significant pest activity, may interfere with climbing |
| 60-79 | **Very High** | Heavy pest activity, consider bringing repellent |
| 80-100 | **Extreme** | Severe pest activity, may make climbing unpleasant |

---

## Contributing Factors Display

When displaying pest activity, woulder shows up to 4 key contributing factors:

**Example factors:**
- "Optimal mosquito temperature (75-85°F)"
- "High humidity (mosquito-friendly)"
- "Calm conditions"
- "Recent rainfall (breeding sites abundant)"
- "Peak mosquito emergence (7-14 days post-rain)"
- "Warm temperatures (peak insect activity)"

---

## Data Sources

- **Temperature**: Current and forecast data from Open-Meteo
- **Humidity**: Relative humidity measurements
- **Wind Speed**: 10-meter wind speed measurements
- **Precipitation**: Hourly rainfall data for last 14 days
- **Season**: Calculated from current date

---

## Limitations and Accuracy

### Known Limitations

1. **Local Variations**: Microclimates near water sources may have higher mosquito activity
2. **Species Differences**: Different mosquito species have different temperature optima
3. **Time of Day**: Most mosquitoes are crepuscular (dawn/dusk), but this model uses daily averages
4. **Elevation Effects**: High elevation locations may have lower pest activity
5. **Vegetation**: Areas with dense vegetation often harbor more insects

### Accuracy Considerations

- Model is calibrated for general outdoor conditions in temperate climates
- Works best for elevations below 8,000 feet
- More accurate during peak season (June-August)
- Local knowledge always supersedes model predictions

---

## Scientific References

**Mosquito Biology:**
- Clements, A.N. (1999). *The Biology of Mosquitoes*. Chapman & Hall.
- Reisen, W.K. (1995). "Effect of Temperature on *Culex tarsalis*". *Journal of Medical Entomology*, 32(5), 594-602.

**Environmental Factors:**
- Paaijmans, K.P., et al. (2010). "Influence of climate on malaria transmission depends on daily temperature variation". *PNAS*, 107(34), 15135-15139.
- Shaman, J., et al. (2002). "Drought-Induced Amplification of Saint Louis Encephalitis Virus". *Emerging Infectious Diseases*, 8(6), 575-580.

**Insect Activity Patterns:**
- Chapman, R.F. (1998). *The Insects: Structure and Function*. Cambridge University Press.
- Gullan, P.J. & Cranston, P.S. (2014). *The Insects: An Outline of Entomology*. Wiley-Blackwell.

---

## Version History

- **v1.0** (December 2024): Initial implementation with temperature, humidity, wind, rainfall, and seasonal factors

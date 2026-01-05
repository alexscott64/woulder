# Rock Drying Calculator Module

This module provides rock drying calculations for climbing conditions, including snow/ice melt estimation.

## Structure

### Core Files

- **calculator.go** - Main calculator and status logic
  - `Calculator` type with `CalculateDryingStatus()` method
  - Handles all rock drying scenarios (rain, snow, ice, freezing)
  - Returns `RockDryingStatus` with detailed information

- **drying_time.go** - Drying time estimation
  - `estimateDryingTime()` - Calculates hours needed for rock to dry after rain
  - `calculateTimeWeightedDrying()` - Tracks drying progress over time
  - `calculateHourlyDryingPower()` - Hourly drying effectiveness
  - `calculateSunExposureFactor()` - Sun exposure impact on drying

- **snow_melt.go** - Snow melt calculations
  - `estimateSnowMeltTime()` - Estimates hours until snow melts off rock
  - Considers temperature trends, sun exposure, rock type, and season
  - Provides realistic estimates even in freezing conditions

- **ice_melt.go** - Ice melt calculations
  - `estimateIceMeltTime()` - Estimates hours until ice melts from rock
  - Handles frozen precipitation scenarios
  - Ice melts faster than snow (denser, better heat conduction)

- **confidence.go** - Confidence scoring
  - `calculateConfidence()` - Confidence score (0-100) for predictions
  - `calculateTemperatureVariance()` - Weather stability analysis

## Key Features

### Smart Snow/Ice Handling
- **No more "Unknown" estimates** - Provides season-based estimates when current temp is freezing
- **Warming trend detection** - Checks last 12 hours to detect warming patterns
- **Seasonal adjustments**:
  - Summer: 2-3 days (48h base + 12h/inch snow)
  - Spring/Fall: 4-7 days (96h base + 24h/inch)
  - Winter: 1-2 weeks (168h base + 36h/inch)

### Comprehensive Factors
- Temperature, humidity, wind speed, cloud cover
- Sun exposure (aspect, tree coverage, rock angle)
- Rock type (porosity, thermal properties)
- Seepage risk
- Wet-sensitive rocks (sandstone, arkose, graywacke)

## Testing

Run tests with:
```bash
go test ./internal/weather/rock_drying/... -v
```

### Test Coverage
- **snow_melt_test.go** - Snow melt scenarios including:
  - Warm weather melting
  - Freezing with warming trends
  - Seasonal variations (summer/winter)
  - Sun exposure effects
  - Rock type differences

## Usage Example

```go
import "github.com/alexscott64/woulder/backend/internal/weather/rock_drying"

calc := &rock_drying.Calculator{}
status := calc.CalculateDryingStatus(
    rockTypes,
    currentWeather,
    historicalWeather,
    sunExposure,
    hasSeepageRisk,
    snowDepthInches,
)

// status contains:
// - IsWet, IsSafe, IsWetSensitive
// - HoursUntilDry (realistic estimate, no more 999)
// - Status ("critical", "poor", "fair", "good")
// - Message (human-readable)
// - ConfidenceScore (0-100)
```
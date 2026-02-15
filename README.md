# woulder

> A comprehensive weather intelligence platform for bouldering.

[![Live Demo](https://img.shields.io/badge/demo-live-brightgreen)](https://woulder.com)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![React](https://img.shields.io/badge/React-18-61DAFB?logo=react)](https://reactjs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5-3178C6?logo=typescript)](https://www.typescriptlang.org/)

Track comprehensive climbing conditions including weather, river crossings, pest activity, and snow accumulation for locations across the Pacific Northwest, Southern California, and eventually more.

[woulder.com](https://woulder.com)

---

## Features

### Core Weather Intelligence
- **13 Climbing Locations** across Washington and British Columbia
- **Geographic Areas** - Pacific Northwest (9 locations) and Southern California (4 locations)
- **Real-time Weather** - Temperature, precipitation, wind, humidity, cloud cover
- **Intelligent Condition Analysis** - Multi-factor climbing suitability assessment
- **6-Day Forecast** - Extended hourly forecasts with daily summaries
- **Sunrise/Sunset Times** - Daily solar data for each location
- **16-Day Historical Data** - Past weather for trend analysis

### Advanced Condition Monitoring

#### Rock Drying Intelligence
- **Multi-Factor Analysis** - Temperature, humidity, wind, sun exposure, rock type
- **Snow/Ice Melt Estimation** - Season-aware calculations (no more "unknown")
- **Wet-Sensitive Detection** - Critical warnings for sandstone, arkose, graywacke
- **Time-Weighted Drying** - Realistic estimates based on actual conditions
- **Confidence Scoring** - 0-100% confidence in predictions
- **Critical Override** - Wet-sensitive rock status automatically sets "DO NOT CLIMB"

#### Boulder-Specific Drying
- **Individual Boulder Analysis** - Precise drying estimates for specific problems
- **Aspect-Based Calculations** - North/South/East/West facing adjustments
- **Sun Exposure Tracking** - Hours of direct sunlight over next 6 days
- **Tree Coverage Impact** - Adjusts drying time based on canopy coverage
- **6-Day Drying Forecast** - Visual timeline showing wet/drying/dry periods
- **Recent Activity Tracking** - See recent ascents and climber feedback
- **Batch API Endpoints** - Efficient fetching for multiple boulders

#### River Crossing Safety
- **Real-time Flow Data** - Live USGS stream gauge readings
- **Safety Assessments** - Safe/Caution/Unsafe indicators
- **Flow Estimation** - Drainage area ratio and empirical methods
- **Threshold Alerts** - Percentage of safe crossing levels
- **Multiple Rivers** - Money Creek, West/East Fork Miller River, North Fork Skykomish

#### Pest Activity Forecasts
- **Mosquito Activity** - Temperature-gated scoring with breeding cycle tracking
- **Outdoor Pests** - Flies, gnats, wasps, ants forecasts
- **5-Level System** - Low, Moderate, High, Very High, Extreme
- **Contributing Factors** - Up to 4 key environmental factors displayed
- **Seasonal Patterns** - Month-based population adjustments

#### Snow Accumulation Tracking
- **SWE-Based Model** - Snow Water Equivalent physics
- **Temperature-Indexed** - Degree-day melt calculations
- **Multi-day Forecasts** - Up to 16 days of snow depth
- **Elevation Adjustment** - Lapse rate corrections (-3.5°F per 1,000 ft)
- **Rain-on-Snow** - Compaction and melt modeling

### Interactive Activity Heat Map
- **Geographic Visualization** - DeckGL-powered 3D map showing climbing activity hotspots
- **Time-Based Filtering** - View activity from last 7 days, 30 days, 90 days, or all time
- **Route Type Filtering** - Filter by Boulder, Sport, Trad, Alpine, or combinations
- **Activity Scoring** - Recent activity weighted more heavily (2x last week, 1.5x last month)
- **Cluster Intelligence** - Automatically groups nearby routes, shows tick counts
- **Area Deep Dive** - Click clusters to see detailed area statistics and top routes
- **Route Details** - View individual routes with recent tick history
- **Search Functionality** - Find specific routes across all areas
- **Lightweight Mode** - Optimized data loading for smooth performance

### User Experience
- **Dark Mode** - Persistent theme switching with localStorage
- **Area Filtering** - Collapsible sidebar to filter locations by region
- **Settings Panel** - Centralized configuration management
- **Responsive Design** - Optimized for mobile, tablet, and desktop
- **Auto-refresh** - Updates every 10 minutes with React Query
- **Offline Detection** - Connection status indicators
- **Smart Caching** - Instant data display with background updates

---

## Quick Start

### Prerequisites

- [Go 1.21+](https://go.dev/dl/)
- [Node.js 18+](https://nodejs.org/)
- PostgreSQL 18 (recommended) / SQLite (included) / MySQL 8.0+
- [Open-Meteo API](https://open-meteo.com/) (free, no key required)
- [USGS Water Services](https://waterservices.usgs.gov/) (free, no key required)

### Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/alexscott64/woulder.git
   cd woulder
   ```

2. **Initialize database**
   ```bash
   cd scripts
   npm install
   node init-db.js
   ```
   This creates tables, seeds 13 locations across 2 areas, and sets up river crossings.

3. **Run backend**
   ```bash
   cd backend
   cp .env.example .env
   # Edit .env with your database credentials
   go mod download
   go run cmd/server/main.go
   ```
   Backend runs on port 8080

4. **Run frontend** (in a new terminal)
   ```bash
   cd frontend
   npm install
   npm run dev
   ```
   Frontend runs on port 5173

5. **Open browser**
   - Frontend: http://localhost:5173
   - Backend API: http://localhost:8080/api/health


---

## Documentation

### User Guides
- **[ADDING_LOCATIONS.md](ADDING_LOCATIONS.md)** - Add a new location

### Scientific Documentation
- **[docs/pest-activity-calculation.md](docs/pest-activity-calculation.md)** - How pest forecasts are calculated
- **[docs/river-crossing-calculation.md](docs/river-crossing-calculation.md)** - River safety assessment methodology
- **[docs/snow-accumulation-calculation.md](docs/snow-accumulation-calculation.md)** - Snow physics and modeling
- **[docs/precipitation-rating.md](docs/precipitation-rating.md)** - Precipitation condition assessment
- **[backend/internal/weather/rock_drying/README.md](backend/internal/weather/rock_drying/README.md)** - Rock drying module documentation


---

## Testing

### Frontend Tests

woulder uses [Vitest](https://vitest.dev/) for comprehensive unit testing:

```bash
cd frontend
npm test                # Run all tests
npm run test:ui         # Run with UI
npm run test:coverage   # Generate coverage report
```

**Test Coverage:**
- 40+ frontend tests across 3 test suites
- Weather analyzers (temperature, wind, precipitation, conditions)
- Weather display formatters (dry time, snow depth)
- UI display components
- Backend tests for rock drying, pest activity, and conditions

**Test Files:**
- `frontend/src/utils/weather/__tests__/` - Weather formatters and utilities
- `frontend/src/components/weather/__tests__/` - Weather display components
- `frontend/src/services/__tests__/` - API client tests
- `backend/internal/weather/rock_drying/*_test.go` - Rock drying tests
- `backend/internal/weather/boulder_drying/*_test.go` - Boulder drying tests
- `backend/internal/service/*_test.go` - Service layer tests
- `backend/internal/pests/analyzer_test.go` - Pest analyzer tests
- `backend/internal/weather/conditions_test.go` - Condition tests

### Backend Tests

woulder uses Go's built-in testing framework with a well-organized test structure:

```bash
cd backend
go test ./...           # Run all tests (82+ passing)
go test -v ./...        # Verbose output
go test -cover ./...    # With coverage
go test ./internal/service/... # Test specific package
```

**Test Organization:**
- **Service Tests** (`internal/service/*_test.go`) - Business logic tests with function-based mocks
- **Single Shared Mocks File** ([`internal/service/mocks_test.go`](backend/internal/service/mocks_test.go)) - Organized mock implementations by domain (weather, locations, boulders, climbing, etc.)
- **Domain Tests** - Rock drying, boulder drying, pest activity, snow accumulation, conditions
- **Repository Tests** - Database layer tests for each repository

**Mock Design:**
- Function-based mocks for flexible test setup
- Organized with clear section headers by domain
- Shared across all service tests (follows Go conventions)
- No code duplication, easy to maintain

### Manual API Testing

```bash
# Health check
curl http://localhost:8080/api/health

# List all areas with location counts
curl http://localhost:8080/api/areas

# Get locations by area
curl http://localhost:8080/api/areas/1/locations

# Weather for all locations
curl http://localhost:8080/api/weather/all

# Weather for specific area
curl http://localhost:8080/api/weather/all?area_id=1

# Weather for specific location
curl http://localhost:8080/api/weather/1

# River data for location
curl http://localhost:8080/api/rivers/location/1

# Boulder drying status for route
curl http://localhost:8080/api/boulder-drying/123456789

# Batch boulder drying status
curl "http://localhost:8080/api/boulder-drying/batch?ids=123456789,987654321"

# Area boulder statistics
curl http://localhost:8080/api/boulder-drying/area/1/stats

# Batch area statistics
curl "http://localhost:8080/api/boulder-drying/batch-area-stats?location_ids=1,2,3"
```

---

## Tech Stack

### Backend
- **[Go 1.21+](https://go.dev/)** - Fast, compiled language with excellent concurrency
- **[Gin](https://gin-gonic.com/)** - High-performance HTTP framework
- **[PostgreSQL 18](https://www.postgresql.org/)** - Primary database (recommended)
- **[SQLite](https://www.sqlite.org/)** / **[MySQL 8.0+](https://www.mysql.com/)** - Alternative databases
- **[Open-Meteo API](https://open-meteo.com/)** - Weather data (free, no key)
- **[USGS Water Services](https://waterservices.usgs.gov/)** - River flow data (free, no key)

### Frontend
- **[React 18](https://react.dev/)** - Modern UI library with hooks
- **[TypeScript 5](https://www.typescriptlang.org/)** - Type safety and better DX
- **[Vite](https://vitejs.dev/)** - Lightning-fast build tool and HMR
- **[Tailwind CSS 3](https://tailwindcss.com/)** - Utility-first CSS with dark mode
- **[React Query](https://tanstack.com/query)** - Data fetching, caching, and synchronization
- **[Axios](https://axios-http.com/)** - Promise-based HTTP client
- **[Lucide React](https://lucide.dev/)** - Beautiful, consistent icons
- **[date-fns](https://date-fns.org/)** - Modern date utility library
- **[Vitest](https://vitest.dev/)** - Fast unit testing framework

---

## API Endpoints

### Core Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/health` | Health check and server status |
| GET | `/api/locations` | List all climbing locations |
| GET | `/api/areas` | List geographic areas with location counts |
| GET | `/api/areas/:id/locations` | Get area details with locations |

### Weather Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/weather/all` | Weather for all locations |
| GET | `/api/weather/all?area_id=X` | Weather filtered by area |
| GET | `/api/weather/:id` | Weather for specific location |
| GET | `/api/weather/coordinates?lat=X&lon=Y` | Weather by coordinates |

### River Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/rivers/location/:id` | River crossing data for location |
| GET | `/api/rivers/:id` | Specific river crossing data |

### Boulder Drying Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/boulder-drying/:mp_route_id` | Boulder drying status for specific route |
| GET | `/api/boulder-drying/batch?ids=X,Y,Z` | Batch boulder drying status (efficient) |
| GET | `/api/boulder-drying/area/:location_id/stats` | Area-wide boulder drying statistics |
| GET | `/api/boulder-drying/batch-area-stats?location_ids=X,Y,Z` | Batch area statistics (efficient) |

### Activity Heat Map Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/heatmap/data` | Get climbing activity heat map points with time/route filters |
| GET | `/api/heatmap/area/:area_id` | Detailed area activity statistics |
| GET | `/api/heatmap/routes` | Routes within geographic bounds with activity data |
| GET | `/api/heatmap/route/:route_id/ticks` | Recent tick history for specific route |
| GET | `/api/heatmap/search` | Search routes by name across areas |

### Climb Tracking Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/climb-tracking/history/:location_id` | Climb history for location |
| POST | `/api/climb-tracking/sync/:location_id` | Sync new ticks from Mountain Project |
| GET | `/api/climb-tracking/sync-status/:location_id` | Check sync status |

---

## Project Structure

```
woulder/
├── backend/                      # Go API server
│   ├── cmd/
│   │   └── server/
│   │       └── main.go           # Application entry point
│   ├── internal/
│   │   ├── api/                  # HTTP handlers and routes
│   │   │   ├── handlers.go       # Core request handlers
│   │   │   ├── handlers_heat_map.go    # Heat map endpoints
│   │   │   └── handlers_climb_tracking.go # Climb tracking endpoints
│   │   ├── database/             # Database layer (modular repositories)
│   │   │   ├── repository.go     # Main repository interface
│   │   │   ├── areas/            # Areas repository
│   │   │   ├── boulders/         # Boulder drying profiles repository
│   │   │   ├── climbing/         # Climbing history & activity repository
│   │   │   ├── heatmap/          # Heat map data repository
│   │   │   ├── locations/        # Locations repository
│   │   │   ├── mountainproject/  # Mountain Project sync repository
│   │   │   ├── rivers/           # River crossings repository
│   │   │   ├── rocks/            # Rock types & sun exposure repository
│   │   │   ├── weather/          # Weather data repository
│   │   │   └── migrations/       # SQL migrations
│   │   ├── models/               # Data structures
│   │   │   ├── location.go       # Location & River models
│   │   │   ├── rock_type.go      # Rock type & sun exposure models
│   │   │   ├── area.go           # Area models
│   │   │   ├── heat_map.go       # Heat map & activity models
│   │   │   └── mountain_project.go # Mountain Project sync models
│   │   ├── service/              # Business logic layer
│   │   │   ├── weather_service.go        # Weather orchestration
│   │   │   ├── boulder_drying_service.go # Boulder drying service
│   │   │   ├── heat_map_service.go       # Heat map service
│   │   │   ├── climb_tracking_service.go # Climb tracking service
│   │   │   ├── location_service.go       # Location service
│   │   │   ├── river_service.go          # River service
│   │   │   └── mocks_test.go             # Shared test mocks (organized by domain)
│   │   ├── weather/              # Weather domain
│   │   │   ├── client/
│   │   │   │   └── openmeteo.go  # Open-Meteo API client
│   │   │   ├── calculator/       # Snow accumulation calculations
│   │   │   │   └── snow_accumulation.go
│   │   │   ├── rock_drying/      # Rock drying module (modular)
│   │   │   │   ├── calculator.go # Main calculator & status logic
│   │   │   │   ├── drying_time.go # Drying time estimation
│   │   │   │   ├── snow_melt.go  # Snow melt calculations
│   │   │   │   ├── ice_melt.go   # Ice melt calculations
│   │   │   │   ├── confidence.go # Confidence scoring
│   │   │   │   ├── snow_melt_test.go # Comprehensive tests
│   │   │   │   └── README.md     # Module documentation
│   │   │   ├── boulder_drying/   # Boulder-specific drying module
│   │   │   │   ├── calculator.go # Boulder drying calculations
│   │   │   │   └── calculator_test.go # Boulder tests
│   │   │   └── conditions.go     # Climbing condition analysis
│   │   ├── service/              # Business logic services
│   │   │   ├── weather_service.go    # Weather orchestration
│   │   │   └── boulder_drying_service.go # Boulder drying service
│   │   ├── pests/                # Pest activity domain
│   │   │   ├── analyzer.go       # Pest condition analyzer
│   │   │   └── analyzer_test.go  # Pest analyzer tests
│   │   └── rivers/               # River data service
│   │       └── usgs_client.go    # USGS API client
│   ├── .env                      # Configuration (not in git)
│   └── go.mod                    # Go dependencies
│
├── frontend/                     # React web application
│   ├── src/
│   │   ├── components/           # React components
│   │   │   ├── WeatherCard.tsx   # Main weather display
│   │   │   ├── ForecastView.tsx  # 6-day forecast view
│   │   │   ├── ConditionsModal.tsx # Comprehensive conditions modal
│   │   │   ├── AreaSelector.tsx  # Area filtering component
│   │   │   ├── SettingsModal.tsx # User settings
│   │   │   ├── RiverInfoModal.tsx    # River crossing details
│   │   │   ├── PestInfoModal.tsx     # Pest activity details
│   │   │   ├── RouteListItem.tsx     # Boulder/route display item
│   │   │   └── DryingForecastTimeline.tsx # 6-day drying forecast visualization
│   │   ├── components/pests/     # Pest UI components
│   │   │   ├── pestDisplay.ts    # Pest UI helpers
│   │   │   └── __tests__/        # Pest display tests
│   │   ├── components/weather/   # Weather UI components
│   │   │   ├── weatherDisplay.ts # Weather UI helpers
│   │   │   └── __tests__/        # Weather display tests
│   │   ├── contexts/             # React contexts
│   │   │   └── SettingsContext.tsx  # Global settings state
│   │   ├── services/             # API client layer
│   │   │   └── api.ts            # HTTP requests
│   │   ├── types/                # TypeScript definitions
│   │   │   ├── weather.ts        # Weather, rock, pest types
│   │   │   ├── river.ts          # River crossing types
│   │   │   └── area.ts           # Area types
│   │   ├── utils/                # Utility functions
│   │   │   ├── geolocation.ts    # Location utilities
│   │   │   └── weather/          # Weather display utilities
│   │   │       ├── analyzers/    # Display helpers
│   │   │       │   ├── ConditionCalculator.ts  # Condition colors/labels
│   │   │       │   ├── TemperatureAnalyzer.ts  # Temperature display
│   │   │       │   ├── WindAnalyzer.ts         # Wind display
│   │   │       │   └── index.ts
│   │   │       ├── calculations/ # (Moved to backend)
│   │   │       │   └── index.ts  # Re-exports for compatibility
│   │   │       ├── formatters.ts # Display formatters (dry time, snow)
│   │   │       ├── index.ts
│   │   │       └── __tests__/
│   │   │           └── dryTimeDisplay.test.ts
│   │   └── App.tsx               # Root component
│   ├── .env                      # Frontend configuration
│   ├── package.json              # npm dependencies
│   └── vitest.config.ts          # Test configuration
│
├── scripts/                      # Utility scripts
│   └── init-db.js                # Database initialization
│
├── docs/                         # Scientific Documentation
│   ├── pest-activity-calculation.md      # Pest science
│   ├── river-crossing-calculation.md     # River science
│   ├── snow-accumulation-calculation.md  # Snow science
│   └── precipitation-rating.md           # Precipitation science
│   # Note: Rock drying docs at backend/internal/weather/rock_drying/README.md
│
├── README.md                     # This file
├── ADDING_LOCATIONS.md           # Add a location guide
└── SUMMARY.md                    # Project summary
```

---

## Architecture

woulder uses a **backend-centric architecture** that separates domain calculations from presentation:

### Backend (Go) - Domain Logic & Calculations
- **Location**: `backend/internal/`
- **Purpose**: All weather intelligence, pest analysis, and condition calculations
- **Modules**:
  - `weather/rock_drying/` - Multi-factor rock drying calculations with snow/ice melt estimation
  - `weather/calculator/` - Snow accumulation physics (SWE-based model)
  - `weather/conditions.go` - Climbing condition analysis (temperature, wind, precipitation)
  - `pests/analyzer.go` - Pest activity forecasting (mosquitoes, outdoor pests)
  - `rivers/usgs_client.go` - River crossing safety assessment
- **Benefits**: Consistent calculations, easier testing, single source of truth
- **Testing**: Comprehensive Go unit tests for all domain logic

### Frontend (React/TypeScript) - Presentation & Display
- **Location**: `frontend/src/`
- **Purpose**: UI presentation, formatting, and user interaction
- **Layers**:
  - **Display Helpers** (`utils/weather/analyzers/`) - Minimal UI logic (condition colors, labels)
    - `ConditionCalculator.ts` - Condition level to color/label mapping
    - `TemperatureAnalyzer.ts` - Temperature display formatting
    - `WindAnalyzer.ts` - Wind display formatting
  - **Formatters** (`utils/weather/formatters.ts`) - Display string formatting (dry time, snow depth)
  - **Components** (`components/`) - Visual components and user interaction
- **Data Flow**: Backend API → React Query cache → Component display
- **Testing**: Vitest unit tests for formatters and display logic

---

## Weather Conditions

woulder analyzes multiple factors to determine climbing suitability:

### Condition Levels

| Level | Color | Criteria |
|-------|-------|----------|
| **Good** | Green | Ideal climbing conditions across all factors |
| **Marginal** | Yellow | One or more factors are suboptimal but manageable |
| **Bad** | Red | One or more factors make climbing unsafe or unpleasant |
| **Do Not Climb** | Dark Red | Critical safety concerns (wet-sensitive rock when wet) |

### Factors Analyzed

1. **Precipitation** - Current rain, recent rain, drying conditions
2. **Temperature** - Ideal (41-65°F), Cold (30-40°F), Warm (66-79°F), Extreme (<30°F, >79°F)
3. **Wind** - Calm (<12 mph), Moderate (12-20 mph), High (20-30 mph), Dangerous (>30 mph)
4. **Humidity** - Normal (<85%), High (85-95%), Very High (>95%)
5. **Rock Status** - Wet-sensitive rocks (sandstone, arkose, graywacke) override overall condition when wet

### Rock Drying Intelligence

woulder provides detailed rock drying estimates with:
- **Smart Snow/Ice Handling** - Season-aware estimates (no more "unknown")
  - Summer: 2-3 days for snow to melt
  - Spring/Fall: 4-7 days
  - Winter: 1-2 weeks
- **Warming Trend Detection** - Analyzes last 12 hours of temperature data
- **Dry Time Display** - Shows hours (<72h) or days (≥72h) for clarity
- **Critical Safety Override** - Wet sandstone automatically sets condition to "DO NOT CLIMB"

### Snow Probability

Based on temperature, precipitation, and elevation:
- **High** - Multiple freezing periods with precipitation
- **Moderate** - Some freezing with precipitation
- **Low** - Freezing without precipitation
- **None** - Warm weather

---

## Locations

### Pacific Northwest (Area 1)

| Location | Coordinates | Elevation | Rivers |
|----------|-------------|-----------|--------|
| Skykomish - Money Creek | 47.70, -121.48 | 1,000 ft | Money Creek |
| Skykomish - Paradise | 47.64, -121.38 | 1,500 ft | West Fork Miller, East Fork Miller |
| Index | 47.82, -121.56 | 500 ft | North Fork Skykomish |
| Gold Bar | 47.85, -121.70 | 200 ft | Skykomish River |
| Bellingham | 48.75, -122.48 | 100 ft | - |
| Icicle Creek (Leavenworth) | 47.60, -120.66 | 1,200 ft | - |
| Squamish | 49.70, -123.16 | 200 ft | - |
| Treasury | 47.76, -121.13 | 3,650 ft | - |
| Calendar Butte | 48.36, -122.08 | 1,600 ft | - |

### Southern California (Area 2)

| Location | Coordinates | Elevation |
|----------|-------------|-----------|
| Joshua Tree | 34.02, -116.16 | 2,700 ft |
| Black Mountain | 33.83, -116.76 | 7,500 ft |
| Buttermilks (Bishop) | 37.33, -118.58 | 6,400 ft |
| Happy / Sad Boulders | 37.42, -118.44 | 4,400 ft |

---

## Dark Mode

woulder includes a full dark mode implementation:

- **Toggle**: Click the Settings icon → Toggle dark mode
- **Persistence**: Preference saved to localStorage
- **Instant Apply**: No page reload required
- **Comprehensive**: All components styled for both themes
- **System Preference**: Respects OS-level dark mode setting on first visit

---

## Deployment

### Production Build

**Backend:**
```bash
cd backend
GOOS=linux GOARCH=amd64 go build -o woulder-api cmd/server/main.go
# Binary: woulder-api
```

**Frontend:**
```bash
cd frontend
npm run build
# Output: dist/
# Serve with nginx, Apache, or any static file server
```

### Environment Variables

**Backend (.env):**
```env
DB_TYPE=postgres              # postgres, sqlite, or mysql
DB_HOST=localhost
DB_PORT=5432
DB_USER=woulder
DB_PASSWORD=your_password
DB_NAME=woulder_db
PORT=8080
```

**Frontend (.env):**
```env
VITE_API_URL=http://localhost:8080
```

---

## Roadmap

### Phase 1: Core Platform ✅
- [x] Backend API with Go + Gin
- [x] Frontend with React + TypeScript
- [x] PostgreSQL database with migrations
- [x] Open-Meteo API integration
- [x] 13 default locations across 2 areas
- [x] Area-based filtering
- [x] Color-coded conditions
- [x] Auto-refresh with React Query

### Phase 2: Intelligence Features ✅
- [x] 6-day hourly forecast view
- [x] River crossing safety (USGS data)
- [x] Pest activity forecasts (mosquitoes, outdoor pests)
- [x] Snow accumulation tracking (SWE model)
- [x] 16-day historical weather data
- [x] Dark mode with persistence
- [x] Comprehensive test suite (40+ frontend, comprehensive backend)
- [x] Scientific documentation (58+ pages)
- [x] Rock drying intelligence with snow/ice melt estimation
- [x] Modular backend architecture (rock_drying module)
- [x] Critical safety overrides for wet-sensitive rocks

### Phase 3: Boulder Intelligence ✅
- [x] Individual boulder drying status API
- [x] Aspect-based drying calculations (N/S/E/W facing)
- [x] Sun exposure tracking (6-day forecast)
- [x] Tree coverage impact on drying time
- [x] 6-day boulder drying forecast timeline
- [x] Recent activity tracking for routes
- [x] Batch API endpoints for efficient boulder fetching
- [x] Consolidated forecast periods (prevents UI clutter)

### Phase 4: Enhanced Experience 🚧
- [x] Geographic area filtering
- [x] Comprehensive conditions modal (Today, Rock, Rivers, Pests)
- [x] Mobile-optimized UI (no horizontal scrolling)
- [ ] Service workers for offline support
- [ ] PWA with install prompt
- [ ] Push notifications for alerts

### Phase 5: Advanced Features 🔮
- [ ] Temperature/speed unit preferences (F/C, mph/km/h)
- [ ] User-created custom areas
- [ ] Location search with autocomplete
- [ ] Add/remove custom locations
- [ ] Weather alerts and notifications
- [ ] Trip planning mode (multi-day forecasts)
- [ ] Historical condition trends
- [ ] Avalanche danger integration
- [ ] Approach trail conditions
- [ ] User-submitted condition reports

---

## Contributing

Contributions are welcome! Here's how to contribute:

1. **Fork the repository**
2. **Create a feature branch** (`git checkout -b feature/amazing-feature`)
3. **Make your changes**
4. **Run tests** (`npm test` in frontend/, `go test ./...` in backend/)
5. **Commit your changes** (`git commit -m 'Add amazing feature'`)
6. **Push to branch** (`git push origin feature/amazing-feature`)
7. **Open a Pull Request**

### Development Guidelines

- Follow existing code style and architecture patterns
- Add tests for new features
- Update documentation for user-facing changes
- Keep commits focused and descriptive
- Reference issues in commit messages

---

## License

GNU General Public License v3.0 - see [LICENSE](LICENSE) for details

---

## Credits

- **Inspiration**: [toorainy.com](https://toorainy.com) by Miles Crawford
- **Weather Data**: [Open-Meteo](https://open-meteo.com/) - Free weather API
- **River Data**: [USGS Water Services](https://waterservices.usgs.gov/) - Real-time stream gauges
- **Icons**: [Lucide](https://lucide.dev/) - Beautiful icon library
- **Framework**: [React](https://react.dev/) + [Tailwind CSS](https://tailwindcss.com/)

---

## Contact

**Alex Scott** - [alexscott.io](https://alexscott.io)

**Project Link**: [github.com/alexscott64/woulder](https://github.com/alexscott64/woulder)

---

**Built for the v0 crushers** 🧗‍♀️

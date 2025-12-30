# woulder

> A comprehensive weather intelligence platform for climbers in the Pacific Northwest

[![Live Demo](https://img.shields.io/badge/demo-live-brightgreen)](https://woulder.com)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![React](https://img.shields.io/badge/React-18-61DAFB?logo=react)](https://reactjs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5-3178C6?logo=typescript)](https://www.typescriptlang.org/)

Track comprehensive climbing conditions including weather, river crossings, pest activity, and snow accumulation for locations across the Pacific Northwest.

[woulder.com](https://woulder.com)

---

## Features

### Core Weather Intelligence
- **13 Climbing Locations** across Washington and British Columbia
- **Geographic Areas** - Pacific Northwest (9 locations) and Southern California (4 locations)
- **Real-time Weather** - Temperature, precipitation, wind, humidity, cloud cover
- **Intelligent Condition Analysis** - Multi-factor climbing suitability assessment
- **16-Day Forecast** - Extended hourly forecasts with daily summaries
- **Sunrise/Sunset Times** - Daily solar data for each location
- **Historical Weather** - Past 14 days for trend analysis

### Advanced Condition Monitoring

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

See [QUICKSTART.md](QUICKSTART.md) for detailed instructions.

---

## Documentation

### User Guides
- **[QUICKSTART.md](QUICKSTART.md)** - Installation and setup guide

### Scientific Documentation
- **[docs/pest-activity-calculation.md](docs/pest-activity-calculation.md)** - How pest forecasts are calculated
- **[docs/river-crossing-calculation.md](docs/river-crossing-calculation.md)** - River safety assessment methodology
- **[docs/snow-accumulation-calculation.md](docs/snow-accumulation-calculation.md)** - Snow physics and modeling
- **[docs/precipitation-rating.md](docs/precipitation-rating.md)** - Precipitation condition assessment

### Architecture
- **[notes/architecture-weather.md](notes/architecture-weather.md)** - Three-layer architecture pattern

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
- 189 tests across 7 test suites
- Weather analyzers (temperature, wind, precipitation, conditions)
- Pest activity calculations
- UI display components
- Regression tests for critical bugs

**Test Files:**
- `src/utils/weather/analyzers/__tests__/` - Weather calculation tests
- `src/utils/pests/analyzers/__tests__/` - Pest calculation tests
- `src/components/weather/__tests__/` - Weather UI tests
- `src/components/pests/__tests__/` - Pest UI tests

### Backend Tests

```bash
cd backend
go test ./...           # Run all tests
go test -v ./...        # Verbose output
go test -cover ./...    # With coverage
```

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
│   │   │   └── handlers.go       # Request handlers
│   │   ├── database/             # Database layer
│   │   │   ├── db.go             # Query methods
│   │   │   └── migrations/       # SQL migrations
│   │   ├── models/               # Data structures
│   │   │   ├── location.go       # Location & River models
│   │   │   └── area.go           # Area model
│   │   ├── weather/              # Weather service
│   │   │   ├── service.go        # Weather service orchestration
│   │   │   └── openmeteo_client.go  # Open-Meteo API client
│   │   └── rivers/               # River data service
│   │       └── usgs_client.go    # USGS API client
│   ├── .env                      # Configuration (not in git)
│   └── go.mod                    # Go dependencies
│
├── frontend/                     # React web application
│   ├── src/
│   │   ├── components/           # React components
│   │   │   ├── WeatherCard.tsx   # Main weather display
│   │   │   ├── ForecastView.tsx  # Detailed forecast modal
│   │   │   ├── AreaSidebar.tsx   # Area filtering sidebar
│   │   │   ├── SettingsModal.tsx # User settings
│   │   │   ├── RiverInfoModal.tsx    # River crossing details
│   │   │   └── PestInfoModal.tsx     # Pest activity details
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
│   │   │   ├── weather.ts        # Weather types
│   │   │   └── area.ts           # Area types
│   │   ├── utils/                # Utility functions
│   │   │   ├── weather/          # Weather utilities
│   │   │   │   ├── calculations/ # Pure math (Layer 1)
│   │   │   │   └── analyzers/    # Business logic (Layer 2)
│   │   │   └── pests/            # Pest utilities
│   │   │       ├── calculations/ # Pure math (Layer 1)
│   │   │       └── analyzers/    # Business logic (Layer 2)
│   │   └── App.tsx               # Root component
│   ├── .env                      # Frontend configuration
│   ├── package.json              # npm dependencies
│   └── vitest.config.ts          # Test configuration
│
├── scripts/                      # Utility scripts
│   └── init-db.js                # Database initialization
│
├── docs/                         # Documentation
│   ├── pest-activity-calculation.md      # Pest science
│   ├── river-crossing-calculation.md     # River science
│   ├── snow-accumulation-calculation.md  # Snow science
│   └── precipitation-rating.md           # Precipitation science
│
├── notes/                        # Development notes
│   └── architecture-weather.md   # Architecture documentation
│
├── README.md                     # This file
├── QUICKSTART.md                 # Quick start guide
└── SUMMARY.md                    # Project summary
```

---

## Architecture

woulder uses a **three-layer architecture** to separate concerns:

### Layer 1: Calculations (Pure Math)
- **Location**: `frontend/src/utils/*/calculations/`
- **Purpose**: Pure domain calculations with no business logic
- **Examples**: Temperature conversions, snow physics, pest scoring formulas
- **Testing**: Unit tests for mathematical correctness

### Layer 2: Analyzers (Business Logic)
- **Location**: `frontend/src/utils/*/analyzers/`
- **Purpose**: Climbing-specific condition assessments
- **Examples**: "Is it too cold to climb?", "Are surfaces dry?", "Is river safe?"
- **Testing**: Unit tests for decision logic

### Layer 3: Components (UI Presentation)
- **Location**: `frontend/src/components/*/`
- **Purpose**: Visual presentation and user interaction
- **Examples**: Color coding, labels, icons, charts
- **Testing**: Component tests for rendering

See [notes/architecture-weather.md](notes/architecture-weather.md) for detailed architecture documentation.

---

## Weather Conditions

woulder analyzes multiple factors to determine climbing suitability:

### Condition Levels

| Level | Color | Criteria |
|-------|-------|----------|
| **Good** | Green | Ideal climbing conditions across all factors |
| **Marginal** | Yellow | One or more factors are suboptimal but manageable |
| **Bad** | Red | One or more factors make climbing unsafe or unpleasant |

### Factors Analyzed

1. **Precipitation** - Current rain, recent rain, drying conditions
2. **Temperature** - Ideal (41-65°F), Cold (30-40°F), Warm (66-79°F), Extreme (<30°F, >79°F)
3. **Wind** - Calm (<12 mph), Moderate (12-20 mph), High (20-30 mph), Dangerous (>30 mph)
4. **Humidity** - Normal (<85%), High (85-95%), Very High (>95%)

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
- [x] 16-day hourly forecast view
- [x] River crossing safety (USGS data)
- [x] Pest activity forecasts (mosquitoes, outdoor pests)
- [x] Snow accumulation tracking (SWE model)
- [x] Historical weather data (14 days)
- [x] Dark mode with persistence
- [x] Comprehensive test suite (189 tests)
- [x] Scientific documentation (58+ pages)

### Phase 3: Enhanced Experience 🚧
- [x] Geographic area filtering
- [x] Collapsible sidebar
- [ ] Service workers for offline support
- [ ] PWA with install prompt
- [ ] Push notifications for alerts

### Phase 4: Advanced Features 🔮
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

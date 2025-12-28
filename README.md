# woulder

> A modern weather dashboard for climbers, inspired by toorainy.com with improved UI and offline support.

[![Live Demo](https://img.shields.io/badge/demo-live-brightgreen)](https://woulder.com)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![React](https://img.shields.io/badge/React-18-61DAFB?logo=react)](https://reactjs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5-3178C6?logo=typescript)](https://www.typescriptlang.org/)

Track rain, wind, temperature, humidity, and cloud cover for climbing locations in the Pacific Northwest - online or offline.

![woulder](https://woulder.com)

---

## Features

### Weather & Conditions
- **7 Climbing Locations** - Skykomish (Money Creek & Paradise), Index, Gold Bar, Bellingham, Icicle Creek, Squamish
- **Real-time Weather** - Temperature, precipitation, wind, humidity, cloud cover
- **Condition Indicators** - Color-coded badges (Good/Marginal/Bad) for climbing suitability
- **6-Day Forecast** - Expandable hourly forecast with daily summaries
- **Sunrise/Sunset Times** - Daily sun times for each location
- **Snow Tracking** - Snow probability and accumulation estimates
- **48-Hour Precipitation** - Past and forecasted rain/snow totals

### River & Pest Information
- **River Crossing Data** - Real-time river flow estimates with safety indicators
- **Pest Activity** - Mosquito and outdoor pest forecasts based on weather conditions

### User Experience
- **Dark Mode** - Toggle between light and dark themes (persisted in localStorage)
- **Settings Panel** - Centralized settings management
- **Responsive Design** - Optimized for mobile, tablet, and desktop
- **Auto-refresh** - Updates every 10 minutes
- **Offline Detection** - Shows online/offline status
- **Smart Caching** - React Query for instant data display
- **Modern UI** - Clean design with Tailwind CSS

---

## Quick Start

### Prerequisites

- [Go 1.21+](https://go.dev/dl/)
- [Node.js 18+](https://nodejs.org/)
- SQLite (included) or MySQL 8.0+
- [Open-Meteo API](https://open-meteo.com/) (free, no key required)

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

3. **Run backend**
   ```bash
   cd backend
   cp .env.example .env
   # Edit .env with your database credentials
   go mod download
   go run cmd/server/main.go
   ```

4. **Run frontend** (in a new terminal)
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

5. **Open browser**
   - Frontend: http://localhost:5173
   - Backend API: http://localhost:8080/api/health

See [QUICKSTART.md](QUICKSTART.md) for detailed instructions.

---

## Documentation

- **[QUICKSTART.md](QUICKSTART.md)** - Quick start guide
- **[SUMMARY.md](SUMMARY.md)** - Project summary and overview

---

## Tech Stack

### Backend
- **[Go](https://go.dev/)** - Fast, compiled language
- **[Gin](https://gin-gonic.com/)** - Lightweight HTTP framework
- **[SQLite](https://www.sqlite.org/)** / **[MySQL](https://www.mysql.com/)** - Database options
- **[Open-Meteo API](https://open-meteo.com/)** - Weather data source (free)

### Frontend
- **[React 18](https://react.dev/)** - UI library
- **[TypeScript](https://www.typescriptlang.org/)** - Type safety
- **[Vite](https://vitejs.dev/)** - Fast build tool
- **[Tailwind CSS](https://tailwindcss.com/)** - Utility-first CSS with dark mode support
- **[React Query](https://tanstack.com/query)** - Data fetching and caching
- **[Axios](https://axios-http.com/)** - HTTP client
- **[Lucide React](https://lucide.dev/)** - Icons
- **[date-fns](https://date-fns.org/)** - Date formatting

---

## API Endpoints

| Method | Endpoint                      | Description                     |
|--------|-------------------------------|---------------------------------|
| GET    | `/api/health`                 | Health check                    |
| GET    | `/api/locations`              | Get all locations               |
| GET    | `/api/weather/all`            | Weather for all locations       |
| GET    | `/api/weather/:id`            | Weather for specific location   |
| GET    | `/api/weather/coordinates?lat=X&lon=Y` | Weather by coordinates |
| GET    | `/api/rivers/location/:id`    | River data for a location       |

---

## Project Structure

```
woulder/
├── backend/                    # Go API server
│   ├── cmd/
│   │   └── server/
│   │       └── main.go         # Entry point
│   ├── internal/
│   │   ├── api/                # HTTP handlers
│   │   ├── database/           # Database layer
│   │   ├── models/             # Data models
│   │   ├── weather/            # Open-Meteo client
│   │   └── rivers/             # River data client
│   ├── .env                    # Configuration (not in git)
│   └── go.mod                  # Dependencies
│
├── frontend/                   # React web app
│   ├── src/
│   │   ├── components/         # React components
│   │   │   ├── WeatherCard.tsx
│   │   │   ├── ForecastView.tsx
│   │   │   ├── SettingsModal.tsx
│   │   │   ├── RiverInfoModal.tsx
│   │   │   └── PestInfoModal.tsx
│   │   ├── contexts/           # React contexts
│   │   │   └── SettingsContext.tsx
│   │   ├── services/           # API client
│   │   ├── types/              # TypeScript types
│   │   ├── utils/              # Helper functions
│   │   │   ├── weatherConditions.ts
│   │   │   └── pestConditions.ts
│   │   └── App.tsx             # Main component
│   ├── .env                    # Frontend config
│   └── package.json            # Dependencies
│
├── scripts/                    # Utility scripts
│   └── init-db.js              # Database initialization
│
├── README.md                   # This file
├── QUICKSTART.md               # Quick start guide
└── SUMMARY.md                  # Project summary
```

---

## Weather Conditions

Weather cards display a colored badge indicating climbing conditions:

| Color | Condition | Criteria |
|-------|-----------|----------|
| Green | **Good** | Dry, low winds (<12 mph), comfortable temps (35-90°F) |
| Yellow | **Marginal** | Light rain (0.05-0.1"), moderate winds (12-20 mph), extreme temps, high humidity (>85%) |
| Red | **Bad** | Heavy rain (>0.1"), high winds (>20 mph) |

---

## Locations

| Location | Coordinates | Region |
|----------|-------------|--------|
| Skykomish - Money Creek | 47.70, -121.48 | Washington |
| Skykomish - Paradise | 47.64, -121.38 | Washington |
| Index | 47.82, -121.56 | Washington |
| Gold Bar | 47.85, -121.70 | Washington |
| Bellingham | 48.75, -122.48 | Washington |
| Icicle Creek (Leavenworth) | 47.60, -120.66 | Washington |
| Squamish | 49.70, -123.16 | British Columbia |

---

## Dark Mode

woulder supports dark mode with automatic persistence:

- Toggle via the Settings icon in the header
- Preference saved to localStorage
- Applies immediately without page reload
- Respects system preference on first visit (coming soon)

---

## Deployment

### Production Build

**Backend:**
```bash
cd backend
GOOS=linux GOARCH=amd64 go build -o woulder-api cmd/server/main.go
```

**Frontend:**
```bash
cd frontend
npm run build
# Output: dist/
```

---

## Testing

### Backend
```bash
cd backend
go run cmd/server/main.go

# Test endpoints
curl http://localhost:8080/api/health
curl http://localhost:8080/api/locations
curl http://localhost:8080/api/weather/all
curl http://localhost:8080/api/rivers/location/1
```

### Frontend
```bash
cd frontend
npm run dev
# Open http://localhost:5173
```

---

## Roadmap

### Phase 1: MVP
- [x] Backend API with Go + Gin
- [x] Frontend dashboard with React + TypeScript
- [x] Database integration (SQLite/MySQL)
- [x] Open-Meteo API integration
- [x] 7 default locations
- [x] Color-coded conditions
- [x] Online/offline detection
- [x] Auto-refresh

### Phase 2: Enhanced Features
- [x] 6-day hourly forecast view
- [x] Dark mode toggle
- [x] Settings panel
- [x] Sunrise/sunset times
- [x] River crossing information
- [x] Pest activity forecasts
- [x] Historical weather data (48h)
- [ ] Service workers for offline support
- [ ] PWA with install prompt

### Phase 3: Advanced Features
- [ ] Temperature/speed unit preferences (F/C, mph/kmh)
- [ ] Additional regions
- [ ] Location search/autocomplete
- [ ] Add/remove custom locations
- [ ] Weather alerts
- [ ] Trip planning mode

---

## License

GNU License - see [LICENSE](LICENSE) for details

---

## Credits

- **Inspiration:** [toorainy.com](https://toorainy.com) by Miles Crawford
- **Weather Data:** [Open-Meteo](https://open-meteo.com/)
- **River Data:** [USGS Water Services](https://waterservices.usgs.gov/)
- **Icons:** [Lucide](https://lucide.dev/)

---

## Contributing

Contributions welcome! Please open an issue or submit a pull request.

---

## Contact

Alex Scott - [alexscott.io](https://alexscott.io)

---

**Built for the v0 crushers**

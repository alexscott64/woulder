# Woulder - Project Summary

**Built:** December 2025
**Status:** Phase 2 complete - Enhanced features implemented
**Live Demo:** https://woulder.com

---

## What We Built

A modern weather dashboard for climbers inspired by toorainy.com, with improved UI, dark mode support, and comprehensive weather data.

### Key Features

#### Weather & Forecasting
- Real-time weather data for 7 climbing locations
- Temperature, precipitation, wind, humidity, cloud cover
- 6-day hourly forecast with expandable views
- 48-hour historical and forecasted precipitation
- Snow probability and accumulation estimates
- Sunrise/sunset times for each location

#### Condition Assessment
- Color-coded climbing conditions (Good/Marginal/Bad)
- Detailed condition reasoning
- River crossing safety indicators
- Pest activity forecasts (mosquitoes, outdoor pests)

#### User Experience
- Dark mode toggle with localStorage persistence
- Settings panel for user preferences
- Responsive design (mobile/tablet/desktop)
- Auto-refresh every 10 minutes
- Online/offline detection
- Smart caching with React Query

### Locations Included
1. Skykomish - Money Creek, WA
2. Skykomish - Paradise, WA
3. Index, WA
4. Gold Bar, WA
5. Bellingham, WA
6. Icicle Creek (Leavenworth), WA
7. Squamish, BC

---

## Tech Stack

### Backend
- **Go** with Gin framework
- **SQLite** / **MySQL** database options
- **Open-Meteo API** (free, no API key required)
- **USGS Water Services** for river data
- REST API with CORS support

### Frontend
- **React 18** + **TypeScript**
- **Vite** (build tool)
- **Tailwind CSS** (styling with dark mode)
- **React Query** (data fetching/caching)
- **Axios** (HTTP client)
- **Lucide React** (icons)
- **date-fns** (date formatting)

---

## Project Structure

```
woulder/
├── backend/              # Go API server
│   ├── cmd/server/       # Entry point
│   ├── internal/         # Core logic
│   │   ├── api/          # HTTP handlers
│   │   ├── database/     # Database layer
│   │   ├── models/       # Data models
│   │   ├── weather/      # Open-Meteo client
│   │   └── rivers/       # USGS river client
│   ├── .env              # Configuration
│   └── go.mod            # Dependencies
│
├── frontend/             # React web app
│   ├── src/
│   │   ├── components/   # React components
│   │   │   ├── WeatherCard.tsx
│   │   │   ├── ForecastView.tsx
│   │   │   ├── SettingsModal.tsx
│   │   │   ├── RiverInfoModal.tsx
│   │   │   └── PestInfoModal.tsx
│   │   ├── contexts/     # React contexts
│   │   │   └── SettingsContext.tsx
│   │   ├── services/     # API client
│   │   ├── types/        # TypeScript types
│   │   ├── utils/        # Helper functions
│   │   │   ├── weatherConditions.ts
│   │   │   └── pestConditions.ts
│   │   └── App.tsx       # Main component
│   ├── .env              # Frontend config
│   └── package.json      # Dependencies
│
├── scripts/              # Utility scripts
│   └── init-db.js        # Database initialization
│
├── notes/                # Documentation
│
├── README.md             # Project overview
├── QUICKSTART.md         # Quick start guide
└── SUMMARY.md            # This file
```

---

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/health` | Health check |
| GET | `/api/locations` | All locations |
| GET | `/api/weather/all` | Weather for all locations |
| GET | `/api/weather/:id` | Weather for specific location |
| GET | `/api/weather/coordinates?lat=X&lon=Y` | Weather by coordinates |
| GET | `/api/rivers/location/:id` | River data for location |

---

## How to Run

### Prerequisites
- Go 1.21+
- Node.js 18+
- SQLite (included) or MySQL

### Quick Start

**Terminal 1 - Backend:**
```bash
cd backend
go mod download  # First time only
go run cmd/server/main.go
```

**Terminal 2 - Frontend:**
```bash
cd frontend
npm install      # First time only
npm run dev
```

**Open:** http://localhost:5173

See [QUICKSTART.md](QUICKSTART.md) for detailed instructions.

---

## Database

### Schema
- `locations` - Climbing locations (7 pre-populated)
- `weather_data` - Weather observations (current + historical)
- `rivers` - River crossing information

### Already Initialized
- Schema created
- 7 locations inserted
- Indexes configured

---

## Configuration

### Backend (.env)
```env
PORT=8080
DB_TYPE=sqlite
DB_PATH=./woulder.db
# Or for MySQL:
# DB_HOST=localhost
# DB_PORT=3306
# DB_USER=woulder
# DB_PASSWORD=yourpassword
# DB_NAME=woulder
```

### Frontend (.env)
```env
VITE_API_URL=http://localhost:8080/api
```

---

## Weather Condition Logic

The app color-codes conditions for climbing:

### Green (Good)
- Dry (no rain)
- Low winds (<12 mph)
- Comfortable temperature (35-90°F)
- Moderate humidity (<85%)

### Yellow (Marginal)
- Light rain (0.05-0.1")
- Moderate winds (12-20 mph)
- Cold (<35°F) or hot (>90°F)
- High humidity (>85%)

### Red (Bad)
- Heavy rain (>0.1")
- High winds (>20 mph)
- Multiple adverse conditions

---

## What's Working

### Phase 1 (MVP)
- Backend API with all endpoints
- Frontend dashboard with weather cards
- Real-time data from Open-Meteo
- Database storage and retrieval
- Online/offline detection
- Auto-refresh every 10 minutes
- Responsive design
- Color-coded conditions
- Error handling
- Loading states

### Phase 2 (Enhanced Features)
- 6-day hourly forecast view
- Dark mode toggle with persistence
- Settings panel
- Sunrise/sunset times
- River crossing information
- Pest activity forecasts
- 48-hour precipitation tracking
- Snow probability estimates

---

## What's Next (Phase 3)

### High Priority
- [ ] Service workers for true offline support
- [ ] PWA manifest for "install to homescreen"
- [ ] Temperature unit toggle (F/C)
- [ ] Speed unit toggle (mph/kmh)

### Medium Priority
- [ ] Location search/autocomplete
- [ ] Add/remove custom locations
- [ ] Weather alerts/notifications
- [ ] Additional regions

### Nice to Have
- [ ] Radar map overlay
- [ ] Trip planning mode (multi-day)
- [ ] Share dashboard links
- [ ] Export data (CSV)

---

## Testing Checklist

### Backend
- [x] Server starts without errors
- [x] `/api/health` returns OK
- [x] `/api/locations` returns 7 locations
- [x] `/api/weather/all` returns weather data
- [x] `/api/rivers/location/:id` returns river data
- [x] Database connection works

### Frontend
- [x] App loads at http://localhost:5173
- [x] All 7 location cards display
- [x] Weather data populates
- [x] Online indicator shows green WiFi
- [x] Auto-refresh works
- [x] Responsive on mobile
- [x] Dark mode toggle works
- [x] Settings persist after refresh
- [x] 6-day forecast expands/collapses
- [x] River info modal opens
- [x] Pest info modal opens

### Integration
- [x] Frontend calls backend successfully
- [x] CORS doesn't block requests
- [x] Data updates reflect in UI
- [x] Timestamps show correctly

---

## Deployment

See [notes/deployment-guide.md](notes/deployment-guide.md) for full instructions.

### Quick Deployment Steps

**Backend:**
1. Build: `GOOS=linux GOARCH=amd64 go build -o woulder-api cmd/server/main.go`
2. Upload binary + `.env` to server
3. Run as systemd service or background process
4. Configure Nginx reverse proxy

**Frontend:**
1. Update `.env` with production API URL
2. Build: `npm run build`
3. Upload `dist/` contents to web server
4. Configure routing for SPA

**Live URL:** https://woulder.com

---

## Documentation

All documentation is in the [notes/](notes/) folder:

1. **[project-plan.md](notes/project-plan.md)** - Full architecture, design decisions, and roadmap
2. **[setup-instructions.md](notes/setup-instructions.md)** - Installation and setup
3. **[technical-implementation.md](notes/technical-implementation.md)** - Implementation details
4. **[deployment-guide.md](notes/deployment-guide.md)** - Deployment instructions

Also see:
- **[README.md](README.md)** - Project overview
- **[QUICKSTART.md](QUICKSTART.md)** - Quick start guide
- **This file (SUMMARY.md)** - Project summary

---

## Performance

### Current Metrics
- **Backend:** ~5-10ms response time (localhost)
- **Frontend:** Initial load <2s (dev mode)
- **API Calls:** Single batch request for all weather data
- **Database Queries:** <5ms average

### Optimization Strategy
- Database indexes on location_id and timestamp
- Connection pooling
- React Query caching (5 min stale, 10 min gc)
- Auto-refresh limited to 10 minutes
- Dark mode CSS uses Tailwind's class strategy

---

## Cost Breakdown

### Current: $0/month
- Open-Meteo: Free (unlimited calls)
- USGS Water Services: Free
- Database: SQLite (local) or existing MySQL
- Hosting: Existing infrastructure
- Domain: Already owned

---

## Credits

- **Inspiration:** toorainy.com by Miles Crawford
- **Weather Data:** Open-Meteo API
- **River Data:** USGS Water Services
- **Location Data:** User-provided coordinates
- **Built with:** Go, React, TypeScript, Tailwind CSS

---

## Recent Updates

### December 2025
- Added dark mode with settings panel
- Implemented 6-day hourly forecast view
- Added river crossing safety indicators
- Added pest activity forecasts
- Added sunrise/sunset times
- Updated to 7 locations (split Skykomish into Money Creek and Paradise)
- Improved condition badge design
- Added 48-hour precipitation tracking

---

**Project Status:** Phase 2 Complete
**Live Demo:** https://woulder.com

Happy climbing!

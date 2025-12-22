# Woulder - Project Plan & Architecture

**Date:** December 13, 2025
**Goal:** Build a modern weather dashboard for climbers, similar to toorainy.com but with better UI and offline support

---

## Requirements

### Core Features
1. Display weather data for specific GPS coordinates
2. Metrics: Rain, Wind, Humidity, Temperature, Cloud Cover
3. Multiple location dashboards (add/remove)
4. Historical view (past few days)
5. Offline-first functionality with smart caching
6. Modern, functional UI optimized for outdoor use

### Target Locations
- Skykomish, WA (47.70000522442753, -121.46672102024145)
- Index, WA
- Gold Bar, WA
- Bellingham, WA
- Icicle Creek (Leavenworth), WA
- Squamish, BC

### Deployment
- Backend: Namecheap hosting at alexscott.io/woulder
- Frontend: GitHub Pages or Namecheap
- Database: AWS RDS MySQL (woulder schema)

---

## Technology Stack

### Backend (Go)
- **Framework:** Gin (lightweight, fast HTTP router)
- **Database:** MySQL 8.0 on AWS RDS
- **Weather API:** OpenWeatherMap (5-day/3-hour forecast API)
- **Caching:** In-database + client-side
- **Key Libraries:**
  - `gin-gonic/gin` - HTTP framework
  - `go-sql-driver/mysql` - MySQL driver
  - `godotenv` - Environment configuration

### Frontend (React + TypeScript)
- **Build Tool:** Vite (fast, modern)
- **UI Framework:** Shadcn/ui (clean, customizable components)
- **Styling:** Tailwind CSS (utility-first)
- **State:** React Query (data fetching/caching)
- **Charts:** Recharts (weather visualizations)
- **PWA:** Vite PWA plugin + Service Workers
- **Storage:** IndexedDB (offline data caching)

---

## Architecture Overview

### Backend Architecture
```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ HTTP/JSON
       ▼
┌─────────────────────────────────┐
│     Gin HTTP Server             │
│  (CORS enabled for frontend)    │
└────────┬───────────────┬────────┘
         │               │
         ▼               ▼
┌──────────────┐  ┌────────────────┐
│  Database    │  │ OpenWeatherMap │
│  (MySQL)     │  │   API Client   │
└──────────────┘  └────────────────┘
```

### Frontend Architecture (PWA)
```
┌──────────────────────────────────┐
│         React App                │
│  ┌────────────────────────────┐  │
│  │  Location Dashboard        │  │
│  │  - Weather Cards           │  │
│  │  - Forecast Charts         │  │
│  │  - Historical View         │  │
│  └────────────────────────────┘  │
└─────────────┬────────────────────┘
              │
              ▼
┌──────────────────────────────────┐
│      Service Worker              │
│  (Intercepts Network Requests)   │
│                                  │
│  Cache Strategy:                 │
│  1. Try Network                  │
│  2. Fall back to Cache           │
│  3. Update cache on success      │
└─────────┬────────────┬───────────┘
          │            │
          ▼            ▼
    ┌─────────┐  ┌──────────┐
    │ Network │  │ IndexedDB│
    │(API)    │  │ (Cache)  │
    └─────────┘  └──────────┘
```

---

## API Design

### Endpoints

#### Health Check
```
GET /api/health
Response: { status: "ok", message: "...", time: "..." }
```

#### Get All Locations
```
GET /api/locations
Response: { locations: [...] }
```

#### Get Weather for Location
```
GET /api/weather/:locationId
Response: {
  location_id: 1,
  location: { id, name, lat, lon, ... },
  current: { temp, precipitation, wind, ... },
  hourly: [...],  // Next 5 days, 3-hour intervals
  historical: [...] // Past 7 days
}
```

#### Get Weather by Coordinates
```
GET /api/weather/coordinates?lat=47.7&lon=-121.46
Response: {
  current: {...},
  forecast: [...]
}
```

#### Get All Weather (Dashboard Endpoint)
```
GET /api/weather/all
Response: {
  weather: [
    { location_id, location, current, hourly, historical },
    ...
  ],
  updated_at: "..."
}
```

---

## Database Schema

### Tables

#### locations
- `id` - Primary key
- `name` - Location name (e.g., "Skykomish")
- `latitude` - Decimal(10,8)
- `longitude` - Decimal(11,8)
- `created_at` - Timestamp
- `updated_at` - Timestamp

#### weather_data
- `id` - Primary key
- `location_id` - Foreign key to locations
- `timestamp` - Weather observation time
- `temperature` - Fahrenheit
- `feels_like` - Feels like temp
- `precipitation` - Inches (3-hour accumulation)
- `humidity` - Percentage (0-100)
- `wind_speed` - MPH
- `wind_direction` - Degrees (0-360)
- `cloud_cover` - Percentage (0-100)
- `pressure` - hPa
- `description` - Text description
- `icon` - OpenWeatherMap icon code
- `created_at` - Record creation time

**Indexes:**
- `idx_location_timestamp` on (location_id, timestamp)
- `idx_timestamp` on (timestamp)
- `unique_location_time` unique constraint

---

## Offline-First Strategy

### Goal
App should work perfectly at 1 bar or offline, displaying cached data

### Implementation

1. **Service Worker Registration**
   - Cache all static assets (HTML, CSS, JS, icons)
   - Implement network-first strategy with cache fallback

2. **IndexedDB Storage**
   - Store last 7 days of weather data per location
   - Store timestamps for cache invalidation
   - Structured by location ID

3. **Network Detection**
   - Use `navigator.onLine` API
   - Implement connection quality checks (ping API)
   - Show "Offline Mode" badge when disconnected

4. **Data Sync**
   - On app load: show cached data immediately
   - In background: fetch fresh data if online
   - On reconnect: auto-refresh all locations
   - Manual refresh button always available

5. **User Experience**
   - Timestamp on each card: "Updated 2 hours ago"
   - Spinner during background refresh
   - Success/error notifications
   - Graceful degradation: show what we have

---

## Weather Data Flow

### Normal Flow (Online)
1. User opens app
2. Display cached data immediately (fast UI)
3. Background: fetch fresh data from API
4. Update UI when new data arrives
5. Save to IndexedDB
6. Save to MySQL database

### Offline Flow
1. User opens app
2. Service Worker intercepts requests
3. Return cached data from IndexedDB
4. Show "Last updated: X minutes ago"
5. Queue refresh attempts
6. When online: sync data

### Low Signal Flow
1. Detect slow connection (timeout threshold)
2. Show cached data immediately
3. Display "Refreshing..." indicator
4. Continue fetching in background
5. Update when complete or fail gracefully

---

## Color Coding (Similar to toorainy.com)

### Conditions
- **Red (Bad):** Heavy rain, high winds, poor climbing conditions
- **Yellow (Marginal):** Light rain, moderate wind, acceptable
- **Green (Good):** Dry, calm, ideal conditions

### Thresholds (for climbing)
```javascript
function getConditionColor(weather) {
  // Rain check (inches in 3h)
  if (weather.precipitation > 0.1) return 'red';

  // Wind check (mph)
  if (weather.wind_speed > 20) return 'red';
  if (weather.wind_speed > 12) return 'yellow';

  // Temperature check
  if (weather.temperature < 40) return 'yellow';
  if (weather.temperature > 85) return 'yellow';

  // Humidity check
  if (weather.humidity > 80) return 'yellow';

  return 'green';
}
```

---

## UI Components (Frontend)

### Pages
1. **Dashboard** (/) - All locations overview
2. **Location Detail** (/location/:id) - Single location deep dive
3. **Settings** (/settings) - Add/remove locations, preferences

### Key Components
- `<LocationCard>` - Weather summary card
- `<WeatherChart>` - Hourly forecast visualization
- `<HistoricalView>` - Past week weather
- `<MetricBadge>` - Individual metric display
- `<OfflineIndicator>` - Network status
- `<RefreshButton>` - Manual refresh
- `<LastUpdated>` - Timestamp display

### Layout
- Responsive grid: 1 column mobile, 2-3 columns desktop
- Dark mode support (optional)
- Touch-friendly for mobile use in field

---

## Caching Strategy

### Backend Caching
- Store all weather data in MySQL
- Keep 7 days of historical data
- Clean up old data daily (cron job or on-demand)

### Frontend Caching
- IndexedDB: 7 days of weather data
- Service Worker: Static assets
- React Query: In-memory API response cache (5 min)
- Cache invalidation on manual refresh

### Cache Timing
- **Fresh:** < 10 minutes
- **Stale but OK:** 10-60 minutes
- **Very Stale:** > 60 minutes (show warning)

---

## Development Timeline

### Phase 1: Backend (Days 1-2) ✓
- [x] Project structure
- [x] Database schema
- [x] MySQL connection
- [x] OpenWeatherMap client
- [x] API endpoints
- [ ] Test with Postman

### Phase 2: Frontend Setup (Days 2-3)
- [ ] Vite + React + TypeScript
- [ ] Tailwind + Shadcn/ui setup
- [ ] Basic layout and routing
- [ ] API service layer

### Phase 3: Core Features (Days 3-4)
- [ ] Location dashboard
- [ ] Weather card components
- [ ] Fetch and display current weather
- [ ] Forecast charts
- [ ] Historical view

### Phase 4: Offline Support (Day 5)
- [ ] Service Worker setup
- [ ] IndexedDB integration
- [ ] Network detection
- [ ] Background sync
- [ ] Offline indicators

### Phase 5: Polish (Day 6)
- [ ] Responsive design
- [ ] Loading states
- [ ] Error handling
- [ ] Color coding
- [ ] Animations/transitions

### Phase 6: Deployment (Day 7)
- [ ] Build Go binary
- [ ] Deploy to Namecheap
- [ ] Configure Nginx/Apache
- [ ] Deploy frontend
- [ ] Test production

---

## OpenWeatherMap API Notes

### API Key
Stored in `.env` file (not committed to git)

### Rate Limits
- Free tier: 1,000 calls/day
- With 6 locations + caching strategy: ~288 calls/day (well under limit)
- Refresh every 10 minutes = 6 locations * 6 calls/hour * 24 hours = 864 calls/day

### Endpoints Used
- Current Weather: `/data/2.5/weather`
- 5-day Forecast: `/data/2.5/forecast` (3-hour intervals, 40 data points)

### Data Format
```json
{
  "dt": 1702479600,
  "main": {
    "temp": 45.2,
    "feels_like": 42.1,
    "humidity": 78,
    "pressure": 1013
  },
  "weather": [
    { "description": "light rain", "icon": "10d" }
  ],
  "wind": {
    "speed": 8.5,
    "deg": 210
  },
  "clouds": { "all": 85 },
  "rain": { "3h": 0.15 }
}
```

---

## Deployment Notes

### Backend Deployment (Namecheap)
- Build Go binary for Linux: `GOOS=linux GOARCH=amd64 go build -o woulder-api cmd/server/main.go`
- Upload via SSH/SFTP
- Configure reverse proxy (Nginx/Apache) for `/woulder` path
- Set up systemd service for auto-restart
- Environment variables via `.env` file

### Frontend Deployment
- Build: `npm run build` in frontend/
- Upload `dist/` folder contents
- Configure base path for `/woulder` route

### Database Setup
- Schema already created
- Default locations pre-populated
- Connection configured in `.env`

---

## Future Enhancements

### Phase 2 Features
- [ ] Location search/autocomplete
- [ ] Custom location add/remove
- [ ] Push notifications for bad weather
- [ ] Radar map overlay
- [ ] Trip planning mode (multi-day view)
- [ ] Share dashboard links
- [ ] Export weather data (CSV)

### Performance
- [ ] Redis caching layer
- [ ] CDN for static assets
- [ ] GraphQL API (optional)

---

## Notes

- Design prioritizes **functionality over aesthetics** (but still looks good!)
- **Offline-first** is critical - climbers often have spotty signal
- **Mobile-first** design - most use will be on phones
- Keep it **simple** - don't over-engineer for v1
- **Learn as you build** - Go + React + PWA concepts

---

## Resources

- OpenWeatherMap Docs: https://openweathermap.org/api
- Vite PWA Plugin: https://vite-pwa-org.netlify.app/
- Shadcn/ui: https://ui.shadcn.com/
- Gin Framework: https://gin-gonic.com/
- React Query: https://tanstack.com/query

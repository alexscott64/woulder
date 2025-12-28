# Woulder - Quick Start Guide

## Prerequisites Completed

- [x] Go installed
- [x] Node.js installed
- [x] Database initialized with schema
- [x] All 9 default locations added to database

## Running the Application

### Option 1: Run Both (Recommended)

Open **two terminals**:

**Terminal 1 - Backend:**
```bash
cd backend
go mod download   # First time only
go run cmd/server/main.go
```

You should see:
```
Database connection established
Starting Woulder API server on port 8080
```

**Terminal 2 - Frontend:**
```bash
cd frontend
npm run dev
```

You should see:
```
  VITE v7.x.x  ready in xxx ms

  ➜  Local:   http://localhost:5173/
  ➜  Network: use --host to expose
```

### Option 2: Test Backend First

```bash
cd backend
go run cmd/server/main.go
```

Test with browser or curl:
```bash
# Health check
curl http://localhost:8080/api/health

# Get all locations
curl http://localhost:8080/api/locations

# Get all weather
curl http://localhost:8080/api/weather/all

# Get river data for a location
curl http://localhost:8080/api/rivers/location/1
```

### Option 3: Frontend Only (Mock Data)

If backend isn't running, the frontend will show error messages but the UI structure will be visible.

```bash
cd frontend
npm run dev
```

## What Should You See?

### Backend Running:
- Server starts on port 8080
- Database connection confirmed
- API endpoints available at `http://localhost:8080/api/*`

### Frontend Running:
- App opens at `http://localhost:5173`
- Header shows "woulder" title with logo
- Online/Offline indicator (green WiFi icon when online)
- Settings button (gear icon)
- Grid of weather cards for all 9 locations:
  - Skykomish - Money Creek
  - Skykomish - Paradise
  - Index
  - Gold Bar
  - Bellingham
  - Icicle Creek (Leavenworth)
  - Squamish
  - Treasury
  - Calendar Butte

### Each Weather Card Shows:
- Location name
- Current timestamp
- Condition badge (Good/Marginal/Bad with color)
- Weather icon
- Temperature in °F
- Weather description
- Sunrise/Sunset times
- Pest activity indicator (bug icon)
- River crossing indicator (wave icon, if applicable)
- Last 48h and Next 48h precipitation
- Snow probability
- Wind speed & direction (mph)
- Humidity (%)
- Cloud cover (%)
- Condition reasons
- Expandable 6-day forecast button

### Settings Panel:
- Click the gear icon in the header
- Toggle dark mode on/off
- Settings are saved to localStorage

### 6-Day Forecast (Expanded):
- Click "Show 6-Day Forecast" on any card
- View hourly data for each day
- See daily summaries with highs/lows
- Close with "Hide Forecast" button

## Troubleshooting

### Backend Issues

**"go: command not found"**
- Restart your terminal/IDE
- Go should be in PATH after installation

**"Failed to connect to database"**
- Check database file exists (SQLite) or credentials (MySQL)
- Verify credentials in `backend/.env`
- Try running the init script again: `node scripts/init-db.js`

**"Failed to fetch weather"**
- Open-Meteo API is free and doesn't require a key
- Check internet connection
- API may be temporarily unavailable

### Frontend Issues

**"Cannot find module"**
- Run `npm install` in frontend directory

**"Failed to load weather data"**
- Ensure backend is running on port 8080
- Check browser console for CORS errors
- Verify VITE_API_URL in `frontend/.env`

**Port 5173 already in use**
```bash
# Kill existing process
npx kill-port 5173
# Or change port in vite.config.ts
```

**Dark mode not working**
- Clear Vite cache: `rm -rf node_modules/.vite`
- Restart dev server
- Check localStorage for `woulder-settings` key

### Network/CORS Issues

If you see CORS errors in browser console:
- Backend CORS is configured to allow all origins (`AllowOrigins: []string{"*"}`)
- This is fine for development
- For production, update `backend/cmd/server/main.go` to specific domains

## Testing Features

### Dark Mode
1. Open the app in browser
2. Click the gear icon (Settings) in the header
3. Toggle "Dark Mode" switch
4. UI should switch between light and dark themes
5. Refresh the page - setting should persist

### 6-Day Forecast
1. Click "Show 6-Day Forecast" on any weather card
2. View the expanded hourly forecast
3. Switch between days using the day headers
4. Note the condition color bar matches the card

### River Crossing Info
1. Look for the wave icon on weather cards (not all locations have rivers)
2. Click the wave icon to open the River Info modal
3. View flow rates and safety indicators

### Pest Activity
1. Look for the bug icon on weather cards
2. Click to view pest activity details
3. See mosquito and outdoor pest forecasts

### Offline Mode
1. Open the app in browser (`http://localhost:5173`)
2. Load the weather data
3. Open browser DevTools (F12)
4. Go to Network tab
5. Select "Offline" from throttling dropdown
6. Observe:
   - WiFi icon turns red
   - Status shows "Offline"
   - Cached data still displays
7. Re-enable network to see auto-refresh

## Development Workflow

### Making Backend Changes

1. Edit Go files in `backend/`
2. Save the file
3. Stop the server (Ctrl+C)
4. Restart: `go run cmd/server/main.go`

### Making Frontend Changes

1. Edit React files in `frontend/src/`
2. Save the file
3. Vite automatically hot-reloads (no restart needed)
4. Check browser for updates
5. If styles aren't updating, restart dev server

### Adding New Locations

#### Method 1: Direct Database Insert
```sql
INSERT INTO locations (name, latitude, longitude, elevation_ft) VALUES
('Your Location', 47.1234, -122.5678, 1000);
```

#### Method 2: Via init script
Edit `scripts/init-db.js` and re-run.

## API Endpoints

### GET /api/health
Health check endpoint
```json
{
  "status": "ok",
  "message": "Woulder API is running",
  "time": "2025-12-27T..."
}
```

### GET /api/locations
Get all saved locations
```json
{
  "locations": [
    {
      "id": 1,
      "name": "Skykomish - Money Creek",
      "latitude": 47.70000522,
      "longitude": -121.46672102,
      "elevation_ft": 1000,
      "created_at": "...",
      "updated_at": "..."
    }
  ]
}
```

### GET /api/weather/all
Get weather for all locations (use this for the dashboard)
```json
{
  "weather": [
    {
      "location_id": 1,
      "location": { ... },
      "current": { ... },
      "hourly": [ ... ],
      "historical": [ ... ],
      "sunrise": "2025-12-27T07:54",
      "sunset": "2025-12-27T16:22",
      "daily_sun_times": [ ... ]
    }
  ],
  "updated_at": "2025-12-27T..."
}
```

### GET /api/weather/:id
Get weather for specific location

### GET /api/weather/coordinates?lat=47.7&lon=-121.46
Get weather for custom coordinates

### GET /api/rivers/location/:id
Get river data for a location
```json
{
  "rivers": [
    {
      "name": "North Fork Skykomish River",
      "current_flow_cfs": 2235,
      "safe_threshold_cfs": 3000,
      "is_safe": true,
      "status": "safe"
    }
  ]
}
```

## Color Coding System

Weather cards show a colored badge indicating climbing conditions:

- **Green (Good):** Dry, calm, ideal conditions
- **Yellow (Marginal):** Light rain OR moderate wind OR extreme temps OR high humidity
- **Red (Bad):** Heavy rain (>0.1") OR high winds (>20mph)

## Settings

Settings are stored in localStorage under the key `woulder-settings`:

```json
{
  "darkMode": false,
  "temperatureUnit": "fahrenheit",
  "speedUnit": "mph"
}
```

Note: Temperature and speed unit preferences are coming in a future update.

## Next Steps

After confirming everything works:

1. Test all locations load properly
2. Verify weather data updates every 10 minutes
3. Test dark mode toggle and persistence
4. Expand forecast views on different cards
5. Check river and pest info modals
6. Test offline mode functionality

## Tips

- **Refresh Interval:** Weather data auto-refreshes every 10 minutes
- **Cache Duration:** React Query caches for 5 minutes (stale) / 10 minutes (garbage collection)
- **Dark Mode:** Toggle persists across browser sessions
- **Forecast View:** Different cards can be expanded independently on desktop
- **Mobile:** Only one forecast can be expanded at a time on mobile

## Need Help?

Check the notes folder for detailed documentation:
- [notes/project-plan.md](notes/project-plan.md) - Full architecture and plan
- [notes/setup-instructions.md](notes/setup-instructions.md) - Detailed setup steps

---

**Ready to climb!**

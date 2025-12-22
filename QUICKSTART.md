# Woulder - Quick Start Guide

## ✅ Prerequisites Completed

- [x] Go installed
- [x] Node.js installed
- [x] MySQL database initialized with schema
- [x] OpenWeatherMap API key configured
- [x] All 6 default locations added to database

## 🚀 Running the Application

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
```

### Option 3: Frontend Only (Mock Data)

If backend isn't running, the frontend will show error messages but the UI structure will be visible.

```bash
cd frontend
npm run dev
```

## 🎯 What Should You See?

### Backend Running:
- Server starts on port 8080
- Database connection confirmed
- API endpoints available at `http://localhost:8080/api/*`

### Frontend Running:
- App opens at `http://localhost:5173`
- Header shows "Woulder" title
- Online/Offline indicator (green WiFi icon when online)
- Refresh button
- Grid of weather cards for all 6 locations:
  - Skykomish
  - Index
  - Gold Bar
  - Bellingham
  - Icicle Creek (Leavenworth)
  - Squamish

### Each Weather Card Shows:
- Location name
- Current timestamp
- Condition indicator (red/yellow/green dot)
- Weather icon
- Temperature in °F
- Weather description
- Precipitation (inches)
- Wind speed & direction (mph)
- Humidity (%)
- Cloud cover (%)
- Condition summary

## 🐛 Troubleshooting

### Backend Issues

**"go: command not found"**
- Restart your terminal/IDE
- Go should be in PATH after installation

**"Failed to connect to database"**
- Check internet connection (AWS RDS is remote)
- Verify credentials in `backend/.env`
- Try running the init script again: `node scripts/init-db.js`

**"Failed to fetch weather"**
- Check OpenWeatherMap API key in `backend/.env`
- Free tier has 1,000 calls/day limit
- Wait a minute if rate-limited

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

### Network/CORS Issues

If you see CORS errors in browser console:
- Backend CORS is configured to allow all origins (`AllowOrigins: []string{"*"}`)
- This is fine for development
- For production, update `backend/cmd/server/main.go` to specific domains

## 📱 Testing Offline Mode

1. Open the app in browser (`http://localhost:5173`)
2. Load the weather data (click Refresh)
3. Open browser DevTools (F12)
4. Go to Network tab
5. Select "Offline" from throttling dropdown
6. Observe:
   - WiFi icon turns red
   - Status shows "Offline"
   - Refresh button is disabled
   - Cached data still displays
7. Re-enable network to see auto-refresh

## 🔄 Development Workflow

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

### Adding New Locations

#### Method 1: Direct Database Insert
```sql
INSERT INTO locations (name, latitude, longitude) VALUES
('Your Location', 47.1234, -122.5678);
```

#### Method 2: Programmatically (future feature)
We'll add a UI for this in the next phase.

## 📊 API Endpoints

### GET /api/health
Health check endpoint
```json
{
  "status": "ok",
  "message": "Woulder API is running",
  "time": "2025-12-13T..."
}
```

### GET /api/locations
Get all saved locations
```json
{
  "locations": [
    {
      "id": 1,
      "name": "Skykomish",
      "latitude": 47.70000522,
      "longitude": -121.46672102,
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
      "historical": [ ... ]
    }
  ],
  "updated_at": "2025-12-13T..."
}
```

### GET /api/weather/:id
Get weather for specific location
```json
{
  "location_id": 1,
  "location": { ... },
  "current": { ... },
  "hourly": [ ... ],
  "historical": [ ... ]
}
```

### GET /api/weather/coordinates?lat=47.7&lon=-121.46
Get weather for custom coordinates
```json
{
  "current": { ... },
  "forecast": [ ... ]
}
```

## 🎨 Color Coding System

Weather cards show a colored dot indicating climbing conditions:

- **🟢 Green (Good):** Dry, calm, ideal conditions
- **🟡 Yellow (Marginal):** Light rain OR moderate wind OR extreme temps OR high humidity
- **🔴 Red (Bad):** Heavy rain (>0.1") OR high winds (>20mph)

## 📝 Next Steps

After confirming everything works:

1. ✅ Test all locations load properly
2. ✅ Verify weather data updates every 10 minutes
3. ✅ Test offline mode functionality
4. ⏳ Add historical weather view (past 7 days)
5. ⏳ Implement PWA with service workers
6. ⏳ Add weather charts/graphs
7. ⏳ Create deployment scripts

## 💡 Tips

- **Refresh Interval:** Weather data auto-refreshes every 10 minutes
- **Cache Duration:** React Query caches for 5 minutes (stale) / 10 minutes (garbage collection)
- **API Rate Limit:** Free tier = 1,000 calls/day. With 6 locations refreshing every 10 minutes = ~864 calls/day (safe)
- **Database Cleanup:** Old weather data (>7 days) should be cleaned periodically

## 📞 Need Help?

Check the notes folder for detailed documentation:
- [notes/project-plan.md](notes/project-plan.md) - Full architecture and plan
- [notes/setup-instructions.md](notes/setup-instructions.md) - Detailed setup steps

---

**Ready to climb!** 🧗‍♂️⛰️

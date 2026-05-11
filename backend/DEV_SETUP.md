# Backend Development Setup

## Hot-Reloading with Air

Air automatically rebuilds and restarts your Go server when you save changes.

### Installation

```bash
# Install air globally
go install github.com/air-verse/air@latest
```

### Usage

```bash
# Navigate to backend directory
cd backend

# Start with hot-reloading (watches for file changes)
air

# That's it! Now any changes to .go files will automatically:
# 1. Rebuild the binary
# 2. Restart the server
# 3. Show you any build errors
```

### What It Does

- **Watches**: All `.go` files in `cmd/`, `internal/`, etc.
- **Ignores**: `*_test.go`, `tmp/`, `vendor/`
- **Rebuilds**: On save (1 second delay)
- **Restarts**: Automatically kills old process and starts new one
- **Shows**: Build errors in `build-errors.log`

### Manual Build (if needed)

```bash
# One-time build without hot-reload
go build -o woulder.exe ./cmd/server

# Run manually
./woulder.exe
```

## Development Workflow

### With Hot-Reloading (Recommended)

1. Open terminal in `backend/` directory
2. Run `air`
3. Edit any Go file
4. Save - server rebuilds and restarts automatically!

### Without Hot-Reloading (Old Way)

1. Edit Go files
2. Stop server (Ctrl+C or kill process)
3. Run `go build -o woulder.exe ./cmd/server`
4. Run `./woulder.exe`
5. Repeat for every change 😫

## Tips

- **Air runs in foreground** - keep the terminal open
- **See logs immediately** - stdout/stderr shows in terminal
- **Build errors** - shown immediately in terminal
- **Port conflicts** - Air will error if port 8080 is already in use

## Frontend Hot-Reloading

Frontend already has hot-reloading via Vite:

```bash
cd frontend
npm run dev
```

Changes to React/TypeScript files update instantly in browser!

## Full Stack Development

**Terminal 1** (Backend):
```bash
cd backend
air
```

**Terminal 2** (Frontend):
```bash
cd frontend
npm run dev
```

Now both frontend and backend hot-reload automatically! 🚀

## Offline weather mode for development

The API server fetches weather from Open-Meteo on cache miss / stale data. In
active development this can quickly trip Open-Meteo's free-tier rate limit,
because every page refresh exercises the weather endpoints. To avoid this, the
server supports an **offline weather mode** that serves all weather data from
the local DB cache and skips Open-Meteo entirely on the per-request hot path.

### Enable it

Add to your `backend/.env`:

```bash
WEATHER_OFFLINE_MODE=true
```

Restart the server. You'll see this on startup:

```
WeatherService: offline mode ENABLED — Open-Meteo API calls disabled, serving from DB only
```

While enabled:

- `GET /api/weather/...` reads cached rows from `woulder.weather_data` only
- The hourly background refresh job is a no-op
- Stale cached data is served as-is (the `< 1h` freshness check is bypassed)
- If a location has no cached data, you get an empty/synthetic response
  (downstream calculators handle this gracefully)

### Refresh the DB on demand

Use the standalone `sync_weather` command to repopulate the cache whenever
you want fresh data:

```bash
# Refresh every location
cd backend && go run ./cmd/sync_weather --all

# Refresh just one location
cd backend && go run ./cmd/sync_weather --location-id 12

# Dry run to see what would be fetched
cd backend && go run ./cmd/sync_weather --all --dry-run
```

See [`cmd/sync_weather/README.md`](cmd/sync_weather/README.md:1) for the full
flag reference and behavior details.

### Production

Leave `WEATHER_OFFLINE_MODE` unset (or `false`) in production — the server's
built-in hourly background refresh will keep the cache warm. The flag only
exists to make local UI iteration painless.

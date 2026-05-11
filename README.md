# woulder

> Weather intelligence for bouldering — predicts what the **rock** is doing, not just the air.

[![Live Demo](https://img.shields.io/badge/demo-live-brightgreen)](https://woulder.com)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![React](https://img.shields.io/badge/React-18-61DAFB?logo=react)](https://reactjs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5-3178C6?logo=typescript)](https://www.typescriptlang.org/)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)

[woulder.com](https://woulder.com) · Built for climbers

---

## What it does

woulder pulls weather, river, and climbing-activity data for boulder fields across the western US and BC, then runs domain-specific physics on top to answer climber-shaped questions:

- *Is the rock dry?*
- *Will it have friction at 2pm, or do I need to wait until shade hits?*
- *Can I cross the creek?*
- *Are the mosquitoes going to ruin my session?*
- *Where is everyone actually climbing right now?*

---

## Key features

| Feature | What it gives you |
|---|---|
| 🌡️ **Rock Temperature & Friction** | Predicted rock surface temperature, friction quality, and "send window" detection per boulder face. See [Rock Temperature](#rock-temperature) below. |
| 💧 **Rock Drying** | Multi-factor drying time with snow/ice melt estimation, wet-sensitive rock overrides ([`backend/internal/weather/rock_drying/README.md`](backend/internal/weather/rock_drying/README.md:1)). |
| 🪨 **Boulder Drying** | Per-boulder estimates using GPS, aspect, dip, sun exposure, and tree cover. 6-day drying timeline. |
| ❄️ **Snow Accumulation** | SWE-based physics model with degree-day melt and elevation lapse-rate corrections. |
| 🌊 **River Crossings** | Live USGS gauge readings with safe/caution/unsafe thresholds and drainage-area estimation. |
| 🦟 **Pest Activity** | Mosquito and outdoor-pest forecasts with breeding-cycle and seasonal adjustments. |
| 🗺️ **Activity Heat Map** | DeckGL-powered map of recent climbing activity, sourced from Mountain Project ticks and Kaya sends. |
| ☀️ **Sun-Exposure Profiles** | Per-location daylight/shade profiles tunable via [`ADDING_LOCATIONS.md`](ADDING_LOCATIONS.md:1). |
| 💬 **Plain-English Reasoning** | Confidence scores and human-readable reasons for every condition call. |

---

## Rock Temperature

woulder predicts the **surface temperature of the rock** itself — not the air — using a heat-balance model that combines solar irradiance, sky radiation, wind convection, and thermal lag for the rock's mass and color. It then layers in:

- **Friction quality** scoring (cold-and-dry vs. greasy-warm) with rock-type-aware thresholds
- **Send-window detection** — the hours when temperature, friction, and condensation align
- **Daily rollups** for at-a-glance forecast cards
- A custom [`RockTempIcon`](frontend/src/components/icons/RockTempIcon.tsx:1) used throughout the UI

Inputs are tuned per location via the `location_sun_exposure` profile (canopy, horizon, aspect bias) — see [`ADDING_LOCATIONS.md`](ADDING_LOCATIONS.md:1) for how to set one up.

📖 **For the full physics, formulas, constants, and worked examples, read [`docs/rock-temperature-calculations.md`](docs/rock-temperature-calculations.md:1).**

Implementation lives in [`backend/internal/weather/rock_temp/`](backend/internal/weather/rock_temp/calculator.go:1) (`calculator`, `surface_temp`, `irradiance`, `thermal_lag`, `friction`, `send_window`, `condensation`, `confidence`, `daily`).

---

## Tech stack

**Backend:** Go 1.21+ · Gin · PostgreSQL 18 (SQLite/MySQL also supported) · [Open-Meteo](https://open-meteo.com/) (weather) · [USGS Water Services](https://waterservices.usgs.gov/) (rivers) · Mountain Project + Kaya scrapers for climbing activity.

**Frontend:** React 18 · TypeScript 5 · Vite · Tailwind CSS 3 · TanStack Query · DeckGL (heat map) · Lucide icons · Vitest.

**Tooling:** [Air](https://github.com/air-verse/air) for Go hot-reload · [`mprocs`](https://github.com/pvolok/mprocs) to run both processes side-by-side · [`golang-migrate`](https://github.com/golang-migrate/migrate) for SQL migrations.

---

## Quick start

### Prerequisites

- Go 1.21+, Node.js 18+, PostgreSQL 18 (recommended)
- [`air`](https://github.com/air-verse/air): `go install github.com/air-verse/air@latest`
- [`mprocs`](https://github.com/pvolok/mprocs): `cargo install mprocs` (or `npm i -g mprocs`)

### Set up

```bash
git clone https://github.com/alexscott64/woulder.git
cd woulder

# Backend config
cp backend/.env.example backend/.env
# edit backend/.env with your DB credentials

# Install deps
(cd backend && go mod download)
(cd frontend && npm install)

# Run database migrations
(cd backend && go run cmd/migrate/main.go up)
```

### Run

From the repo root:

```bash
mprocs
```

This launches the backend (Air hot-reload, port 8080) and frontend (Vite, port 5173) together — see [`mprocs.yaml`](mprocs.yaml:1). Open http://localhost:5173.

To run them separately, use `cd backend && air` and `cd frontend && npm run dev`. Full backend dev notes are in [`backend/DEV_SETUP.md`](backend/DEV_SETUP.md:1).

---

## Project layout

```
woulder/
├── backend/                          Go API server
│   ├── cmd/                          Entry points (server, migrate, sync_*, job_monitor, …)
│   ├── internal/
│   │   ├── api/                      HTTP handlers & middleware
│   │   ├── database/                 Repositories (per domain) + SQL migrations
│   │   ├── service/                  Business logic orchestration
│   │   ├── weather/
│   │   │   ├── client/               Open-Meteo & OpenWeatherMap clients
│   │   │   ├── calculator/           Snow accumulation & precipitation
│   │   │   ├── rock_drying/          Drying-time / snow-melt / ice-melt
│   │   │   ├── rock_temp/            Rock temperature, friction, send window
│   │   │   ├── boulder_drying/       Per-boulder drying with GPS/aspect/trees
│   │   │   └── sun/                  Sun-position math
│   │   ├── pests/                    Mosquito & outdoor-pest analyzer
│   │   ├── rivers/                   USGS client
│   │   ├── kaya/ · mountainproject/  Climbing-activity scrapers
│   │   └── monitoring/               Background-job monitoring
│   ├── deployment/                   systemd units for Kaya sync
│   └── DEV_SETUP.md
├── frontend/                         React + Vite UI
│   └── src/
│       ├── components/               UI (WeatherCard, ConditionsModal, map/, icons/, …)
│       ├── contexts/ · hooks/        Settings, climb-activity hooks
│       ├── services/api.ts           HTTP client
│       ├── types/ · utils/           Type defs and display helpers
├── docs/                             Scientific & methodology docs
├── tools/                            One-off scripts (Kaya capture, etc.)
├── mprocs.yaml                       Dev-process orchestration
├── ADDING_LOCATIONS.md
├── CLAUDE.md
└── LICENSE
```

---

## Documentation

### Setup & operations
- [`backend/DEV_SETUP.md`](backend/DEV_SETUP.md:1) — Air hot-reload and backend workflow
- [`backend/.env.example`](backend/.env.example:1) — environment variables
- [`backend/deployment/SYSTEMD_SETUP.md`](backend/deployment/SYSTEMD_SETUP.md:1) — systemd units for the Kaya sync job
- [`ADDING_LOCATIONS.md`](ADDING_LOCATIONS.md:1) — onboard a new climbing location (incl. sun-exposure profile)

### Science & methodology
- [`docs/rock-temperature-calculations.md`](docs/rock-temperature-calculations.md:1) — **rock temperature, friction, send window, condensation**
- [`docs/rock-drying-algorithm.md`](docs/rock-drying-algorithm.md:1) — rock drying model
- [`backend/internal/weather/rock_drying/README.md`](backend/internal/weather/rock_drying/README.md:1) — rock-drying module reference
- [`docs/snow-accumulation-calculation.md`](docs/snow-accumulation-calculation.md:1) — SWE-based snow model
- [`docs/precipitation-rating.md`](docs/precipitation-rating.md:1) — precipitation scoring
- [`docs/river-crossing-calculation.md`](docs/river-crossing-calculation.md:1) — river safety thresholds
- [`docs/pest-activity-calculation.md`](docs/pest-activity-calculation.md:1) — pest activity model
- [`docs/rock-types.md`](docs/rock-types.md:1) — rock-type properties used in calculations

### Internal subsystems
- [`backend/internal/database/PRIORITY_SYSTEM.md`](backend/internal/database/PRIORITY_SYSTEM.md:1) — sync-job priority logic
- [`backend/internal/weather/boulder_drying/PERFORMANCE.md`](backend/internal/weather/boulder_drying/PERFORMANCE.md:1) — boulder-drying perf notes
- [`backend/cmd/sync_kaya/README.md`](backend/cmd/sync_kaya/README.md:1) · [`backend/cmd/sync_weather/README.md`](backend/cmd/sync_weather/README.md:1) · [`backend/cmd/job_monitor/README.md`](backend/cmd/job_monitor/README.md:1) — sync & monitoring CLIs
- [`CLAUDE.md`](CLAUDE.md:1) — high-level project context for AI coding sessions

---

## Testing

```bash
# Backend
cd backend && go test ./...

# Frontend
cd frontend && npm test
```

Backend tests cover the weather calculators (rock temp, rock drying, boulder drying, snow), pests, services, and repositories. Frontend tests use Vitest for display formatters, hooks, and components.

---

## Contributing

PRs welcome. Before opening one:

1. `go test ./...` and `npm test` both pass
2. New domain logic lives on the **backend** (the frontend is for presentation only — see [`CLAUDE.md`](CLAUDE.md:1))
3. User-facing changes update the relevant doc in `docs/`

---

## License

GNU General Public License v3.0 — see [`LICENSE`](LICENSE:1).

---

**Credits:** Inspired by [toorainy.com](https://toorainy.com). Weather by [Open-Meteo](https://open-meteo.com/), river data by [USGS](https://waterservices.usgs.gov/), icons by [Lucide](https://lucide.dev/).

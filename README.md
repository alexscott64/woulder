# Woulder 🌧️

> A modern weather dashboard for climbers, inspired by toorainy.com with improved UI and offline support.

[![Live Demo](https://img.shields.io/badge/demo-live-brightgreen)](https://alexscott.io/woulder)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![React](https://img.shields.io/badge/React-18-61DAFB?logo=react)](https://reactjs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5-3178C6?logo=typescript)](https://www.typescriptlang.org/)

Track rain, wind, temperature, humidity, and cloud cover for climbing locations in the Pacific Northwest - online or offline.

![Woulder Dashboard](https://via.placeholder.com/800x400?text=Woulder+Dashboard+Screenshot)

---

## ✨ Features

- 🗺️ **6 Climbing Locations** - Skykomish, Index, Gold Bar, Bellingham, Icicle Creek, Squamish
- ☁️ **Real-time Weather** - Temperature, precipitation, wind, humidity, cloud cover
- 🟢🟡🔴 **Condition Indicators** - Color-coded for climbing suitability
- 📱 **Responsive Design** - Optimized for mobile, tablet, and desktop
- 🔄 **Auto-refresh** - Updates every 10 minutes
- 🌐 **Offline Detection** - Shows online/offline status
- ⚡ **Smart Caching** - React Query for instant data display
- 🎨 **Modern UI** - Clean design with Tailwind CSS

---

## 🚀 Quick Start

### Prerequisites

- [Go 1.21+](https://go.dev/dl/)
- [Node.js 18+](https://nodejs.org/)
- MySQL 8.0+ (or use existing database)
- [OpenWeatherMap API key](https://openweathermap.org/api) (free tier)

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
   # Edit .env with your API key and database credentials
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

## 📚 Documentation

- **[QUICKSTART.md](QUICKSTART.md)** - Quick start guide
- **[SUMMARY.md](SUMMARY.md)** - Project summary and overview
- **[notes/project-plan.md](notes/project-plan.md)** - Full architecture and plan
- **[notes/setup-instructions.md](notes/setup-instructions.md)** - Detailed setup
- **[notes/technical-implementation.md](notes/technical-implementation.md)** - Implementation details
- **[notes/deployment-guide.md](notes/deployment-guide.md)** - Deployment to Namecheap

---

## 🛠️ Tech Stack

### Backend
- **[Go](https://go.dev/)** - Fast, compiled language
- **[Gin](https://gin-gonic.com/)** - Lightweight HTTP framework
- **[MySQL](https://www.mysql.com/)** - Relational database
- **[OpenWeatherMap API](https://openweathermap.org/api)** - Weather data source

### Frontend
- **[React 18](https://react.dev/)** - UI library
- **[TypeScript](https://www.typescriptlang.org/)** - Type safety
- **[Vite](https://vitejs.dev/)** - Fast build tool
- **[Tailwind CSS](https://tailwindcss.com/)** - Utility-first CSS
- **[React Query](https://tanstack.com/query)** - Data fetching and caching
- **[Axios](https://axios-http.com/)** - HTTP client
- **[Lucide React](https://lucide.dev/)** - Icons
- **[date-fns](https://date-fns.org/)** - Date formatting

---

## 📊 API Endpoints

| Method | Endpoint                      | Description                     |
|--------|-------------------------------|---------------------------------|
| GET    | `/api/health`                 | Health check                    |
| GET    | `/api/locations`              | Get all locations               |
| GET    | `/api/weather/all`            | Weather for all locations       |
| GET    | `/api/weather/:id`            | Weather for specific location   |
| GET    | `/api/weather/coordinates?lat=X&lon=Y` | Weather by coordinates |

---

## 🗂️ Project Structure

```
woulder/
├── backend/                    # Go API server
│   ├── cmd/
│   │   └── server/
│   │       └── main.go         # Entry point
│   ├── internal/
│   │   ├── api/                # HTTP handlers
│   │   ├── database/           # MySQL layer
│   │   ├── models/             # Data models
│   │   └── weather/            # OpenWeatherMap client
│   ├── .env                    # Configuration (not in git)
│   └── go.mod                  # Dependencies
│
├── frontend/                   # React web app
│   ├── src/
│   │   ├── components/         # React components
│   │   ├── services/           # API client
│   │   ├── types/              # TypeScript types
│   │   ├── utils/              # Helper functions
│   │   └── App.tsx             # Main component
│   ├── .env                    # Frontend config
│   └── package.json            # Dependencies
│
├── scripts/                    # Utility scripts
│   └── init-db.js              # Database initialization
│
├── notes/                      # Documentation
│   ├── project-plan.md
│   ├── setup-instructions.md
│   ├── technical-implementation.md
│   └── deployment-guide.md
│
├── README.md                   # This file
├── QUICKSTART.md               # Quick start guide
└── SUMMARY.md                  # Project summary
```

---

## 🎨 Weather Conditions

Weather cards display a colored indicator for climbing conditions:

| Color | Condition | Criteria |
|-------|-----------|----------|
| 🟢 Green | **Good** | Dry, low winds (<12 mph), comfortable temps (35-90°F) |
| 🟡 Yellow | **Marginal** | Light rain (0.05-0.1"), moderate winds (12-20 mph), extreme temps, high humidity (>85%) |
| 🔴 Red | **Bad** | Heavy rain (>0.1"), high winds (>20 mph) |

---

## 📍 Locations

| Location | Coordinates | Region |
|----------|-------------|--------|
| Skykomish | 47.70, -121.47 | Washington |
| Index | 47.82, -121.56 | Washington |
| Gold Bar | 47.86, -121.70 | Washington |
| Bellingham | 48.76, -122.49 | Washington |
| Icicle Creek (Leavenworth) | 47.60, -120.78 | Washington |
| Squamish | 49.70, -123.16 | British Columbia |

---

## 🚢 Deployment

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

See [notes/deployment-guide.md](notes/deployment-guide.md) for full deployment instructions.

---

## 🧪 Testing

### Backend
```bash
cd backend
go run cmd/server/main.go

# Test endpoints
curl http://localhost:8080/api/health
curl http://localhost:8080/api/locations
curl http://localhost:8080/api/weather/all
```

### Frontend
```bash
cd frontend
npm run dev
# Open http://localhost:5173
```

---

## 🛣️ Roadmap

### Phase 1: MVP ✅
- [x] Backend API with Go + Gin
- [x] Frontend dashboard with React + TypeScript
- [x] MySQL database integration
- [x] OpenWeatherMap API integration
- [x] 6 default locations
- [x] Color-coded conditions
- [x] Online/offline detection
- [x] Auto-refresh

### Phase 2: Enhanced Features
- [ ] Service workers for offline support
- [ ] IndexedDB for persistent caching
- [ ] Hourly forecast view
- [ ] 7-day historical chart
- [ ] PWA with install prompt

### Phase 3: Advanced Features
- [ ] Location search/autocomplete
- [ ] Add/remove custom locations
- [ ] Weather alerts
- [ ] Share dashboard links
- [ ] Trip planning mode

---

## 📝 License

MIT License - see [LICENSE](LICENSE) for details

---

## 🙏 Credits

- **Inspiration:** [toorainy.com](https://toorainy.com) by Miles Crawford
- **Weather Data:** [OpenWeatherMap](https://openweathermap.org/)
- **Icons:** [Lucide](https://lucide.dev/)

---

## 🤝 Contributing

Contributions welcome! Please open an issue or submit a pull request.

---

## 📧 Contact

Alex Scott - [alexscott.io](https://alexscott.io)

---

**Built with ❤️ for climbers**

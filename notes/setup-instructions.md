# Woulder Setup Instructions

## Prerequisites Installation

### 1. Install Go (Backend)

**Windows:**
1. Download Go from: https://go.dev/dl/
2. Download the Windows installer (.msi file) for Go 1.21 or later
3. Run the installer and follow the prompts
4. Verify installation:
   ```bash
   go version
   ```

### 2. Install Node.js (Frontend)

**Windows:**
1. Download Node.js from: https://nodejs.org/
2. Download the LTS version (18.x or later)
3. Run the installer
4. Verify installation:
   ```bash
   node --version
   npm --version
   ```

### 3. MySQL Client (Optional)

For direct database access:
```bash
# Install MySQL client
npm install -g mysql

# Or use MySQL Workbench GUI
# Download from: https://dev.mysql.com/downloads/workbench/
```

---

## Backend Setup

### Step 1: Install Go Dependencies

```bash
cd backend
go mod download
go mod tidy
```

This will download:
- gin-gonic/gin (HTTP framework)
- gin-contrib/cors (CORS middleware)
- go-sql-driver/mysql (MySQL driver)
- joho/godotenv (Environment variables)

### Step 2: Configure Environment

The `.env` file is already configured with your database credentials.

**Verify `.env` contains:**
```env
PORT=8080
GIN_MODE=release
OPENWEATHERMAP_API_KEY=4df3c0436f6dd4f0b6af69e97cb4f2bb
DB_HOST=leasecalcs-development.c0xbv45gqu40.us-west-2.rds.amazonaws.com
DB_PORT=3306
DB_USER=woulder
DB_PASSWORD=j32JgmxzycbaoLet9F#9C%wFfN*RF98O
DB_NAME=woulder
CACHE_DURATION=10
```

### Step 3: Initialize Database Schema

**Option A: Using MySQL CLI (if installed)**
```bash
mysql -h leasecalcs-development.c0xbv45gqu40.us-west-2.rds.amazonaws.com -P 3306 -u woulder -p woulder < internal/database/schema.sql
# Password: j32JgmxzycbaoLet9F#9C%wFfN*RF98O
```

**Option B: Using MySQL Workbench GUI**
1. Open MySQL Workbench
2. Create new connection:
   - Hostname: leasecalcs-development.c0xbv45gqu40.us-west-2.rds.amazonaws.com
   - Port: 3306
   - Username: woulder
   - Password: j32JgmxzycbaoLet9F#9C%wFfN*RF98O
   - Default Schema: woulder
3. Open `internal/database/schema.sql`
4. Execute the SQL script

**Option C: Using Node.js Script (easiest)**
See `scripts/init-db.js` below

### Step 4: Run the Backend

```bash
cd backend
go run cmd/server/main.go
```

You should see:
```
Starting Woulder API server on port 8080
Database connection established
```

### Step 5: Test the API

Open a browser or use curl:
```bash
# Health check
curl http://localhost:8080/api/health

# Get all locations
curl http://localhost:8080/api/locations

# Get weather for all locations
curl http://localhost:8080/api/weather/all
```

---

## Frontend Setup

### Step 1: Create React App

```bash
cd frontend
npm install
```

### Step 2: Run Development Server

```bash
npm run dev
```

The app should open at http://localhost:5173

---

## Database Initialization Script

Create `scripts/init-db.js`:

```javascript
const mysql = require('mysql2/promise');
const fs = require('fs');
const path = require('path');

async function initDatabase() {
  const connection = await mysql.createConnection({
    host: 'leasecalcs-development.c0xbv45gqu40.us-west-2.rds.amazonaws.com',
    port: 3306,
    user: 'woulder',
    password: 'j32JgmxzycbaoLet9F#9C%wFfN*RF98O',
    database: 'woulder',
    multipleStatements: true
  });

  const schema = fs.readFileSync(
    path.join(__dirname, '../backend/internal/database/schema.sql'),
    'utf8'
  );

  await connection.query(schema);
  console.log('Database schema initialized successfully!');

  await connection.end();
}

initDatabase().catch(console.error);
```

Run with:
```bash
npm install mysql2
node scripts/init-db.js
```

---

## Troubleshooting

### Backend Issues

**"go: command not found"**
- Go is not installed or not in PATH
- Restart terminal after installing Go

**"Failed to connect to database"**
- Check AWS RDS security group allows your IP
- Verify credentials in `.env` file
- Test connection with MySQL Workbench

**"Failed to fetch weather"**
- Verify OpenWeatherMap API key is valid
- Check network connection
- API might be rate-limited (wait a minute)

### Frontend Issues

**Port 5173 already in use**
```bash
# Kill process on port
npx kill-port 5173
# Or change port in vite.config.ts
```

**CORS errors**
- Backend must be running on port 8080
- Check CORS configuration in backend/cmd/server/main.go

---

## Development Workflow

### 1. Start Backend
```bash
cd backend
go run cmd/server/main.go
```

### 2. Start Frontend (in another terminal)
```bash
cd frontend
npm run dev
```

### 3. Make Changes
- Backend: Save Go files, server auto-restarts
- Frontend: Save React files, Vite hot-reloads

---

## Building for Production

### Backend
```bash
cd backend

# Build for Linux (Namecheap server)
GOOS=linux GOARCH=amd64 go build -o woulder-api cmd/server/main.go

# Build for Windows (local testing)
go build -o woulder-api.exe cmd/server/main.go
```

### Frontend
```bash
cd frontend
npm run build

# Output will be in frontend/dist/
```

---

## Next Steps

1. ✅ Install prerequisites (Go, Node.js)
2. ✅ Initialize database schema
3. ✅ Test backend API
4. ⏳ Build frontend React app
5. ⏳ Implement PWA features
6. ⏳ Deploy to Namecheap

See [deployment-guide.md](deployment-guide.md) for deployment instructions.

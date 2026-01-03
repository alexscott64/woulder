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

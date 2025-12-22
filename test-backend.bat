@echo off
echo Testing Woulder Backend...
echo.
echo Step 1: Downloading Go dependencies...
cd backend
go mod download
echo.
echo Step 2: Starting backend server...
echo Backend will run on http://localhost:8080
echo Press Ctrl+C to stop
echo.
go run cmd/server/main.go

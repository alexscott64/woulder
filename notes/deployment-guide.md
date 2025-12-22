# Woulder - Deployment Guide for Namecheap

**Target URL:** https://alexscott.io/woulder

---

## Prerequisites

- Namecheap cPanel access
- SSH access (if available)
- FTP/SFTP credentials
- Go binary compiled for Linux
- Frontend built

---

## Part 1: Backend Deployment

### Step 1: Build Go Binary for Linux

On your Windows machine:

```bash
cd backend

# Build for Linux (your Namecheap server)
set GOOS=linux
set GOARCH=amd64
go build -o woulder-api cmd/server/main.go
```

This creates `woulder-api` binary (Linux executable).

### Step 2: Upload Backend Files

**Upload via FTP/SFTP:**
- `woulder-api` (binary)
- `.env` file (with correct paths)

**Recommended location:**
```
/home/username/apps/woulder/
├── woulder-api
└── .env
```

### Step 3: Make Binary Executable

```bash
chmod +x woulder-api
```

### Step 4: Run Backend

**Option A: Direct Run (Testing)**
```bash
./woulder-api
```

**Option B: Background Process**
```bash
nohup ./woulder-api > woulder.log 2>&1 &
```

**Option C: Systemd Service (Best - if you have root access)**

Create `/etc/systemd/system/woulder.service`:
```ini
[Unit]
Description=Woulder Weather API
After=network.target

[Service]
Type=simple
User=your-username
WorkingDirectory=/home/username/apps/woulder
ExecStart=/home/username/apps/woulder/woulder-api
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Then:
```bash
sudo systemctl daemon-reload
sudo systemctl enable woulder
sudo systemctl start woulder
sudo systemctl status woulder
```

### Step 5: Configure Reverse Proxy

Your backend runs on port 8080, but you want it accessible at `alexscott.io/woulder`.

**Nginx Configuration:**

Edit `/etc/nginx/sites-available/alexscott.io` (or your config file):

```nginx
server {
    listen 80;
    server_name alexscott.io www.alexscott.io;

    # Existing location blocks...

    # Woulder API
    location /woulder/api/ {
        proxy_pass http://localhost:8080/api/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }

    # Woulder Frontend (static files)
    location /woulder/ {
        alias /home/username/public_html/woulder/;
        try_files $uri $uri/ /woulder/index.html;
    }
}
```

Test and reload Nginx:
```bash
sudo nginx -t
sudo systemctl reload nginx
```

**Apache Configuration (.htaccess):**

If using Apache, create `/public_html/woulder/.htaccess`:

```apache
# Enable rewrite engine
RewriteEngine On

# Proxy API requests to backend
RewriteCond %{REQUEST_URI} ^/woulder/api/
RewriteRule ^api/(.*)$ http://localhost:8080/api/$1 [P,L]

# SPA routing - serve index.html for non-file requests
RewriteCond %{REQUEST_FILENAME} !-f
RewriteCond %{REQUEST_FILENAME} !-d
RewriteRule ^.*$ /woulder/index.html [L]
```

---

## Part 2: Frontend Deployment

### Step 1: Update Frontend API URL

Edit `frontend/.env`:
```env
VITE_API_URL=https://alexscott.io/woulder/api
```

### Step 2: Build Frontend

```bash
cd frontend
npm run build
```

This creates `dist/` folder with optimized production files.

### Step 3: Update Base Path (for /woulder route)

Edit `frontend/vite.config.ts`:

```typescript
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  base: '/woulder/', // Add this line
})
```

Rebuild:
```bash
npm run build
```

### Step 4: Upload Frontend Files

Upload everything from `frontend/dist/` to:
```
/home/username/public_html/woulder/
```

Your folder should look like:
```
/public_html/woulder/
├── index.html
├── assets/
│   ├── index-*.js
│   ├── index-*.css
│   └── ...
└── vite.svg
```

---

## Part 3: Verify Deployment

### Test Backend
```bash
curl https://alexscott.io/woulder/api/health
```

Expected response:
```json
{
  "status": "ok",
  "message": "Woulder API is running",
  "time": "2025-12-13T..."
}
```

### Test Frontend
Open browser:
```
https://alexscott.io/woulder
```

You should see the Woulder dashboard with weather data.

### Check Browser Console
Press F12, look for:
- No CORS errors
- Successful API calls to `/woulder/api/weather/all`
- No 404 errors

---

## Part 4: SSL/HTTPS Setup

### Option A: Let's Encrypt (Free)

If you have SSH access:

```bash
sudo certbot --nginx -d alexscott.io -d www.alexscott.io
```

Certbot automatically:
- Obtains SSL certificate
- Updates Nginx configuration
- Sets up auto-renewal

### Option B: Namecheap SSL

1. Purchase SSL from Namecheap
2. Follow Namecheap's guide to install
3. Update Nginx to use SSL certificate

### Update Frontend API URL

After enabling HTTPS, update `frontend/.env`:
```env
VITE_API_URL=https://alexscott.io/woulder/api
```

Rebuild and re-upload.

---

## Part 5: Monitoring & Maintenance

### Check Backend Logs

```bash
# If using nohup
tail -f woulder.log

# If using systemd
sudo journalctl -u woulder -f
```

### Restart Backend

```bash
# Direct process
pkill woulder-api
./woulder-api &

# Systemd
sudo systemctl restart woulder
```

### Update Application

**Backend Update:**
1. Build new binary locally
2. Upload to server
3. Restart service

**Frontend Update:**
1. Make changes
2. Run `npm run build`
3. Upload new `dist/` contents
4. No restart needed (static files)

### Database Maintenance

**Clean old weather data (run weekly):**

```sql
DELETE FROM weather_data WHERE timestamp < DATE_SUB(NOW(), INTERVAL 30 DAY);
```

Or create a cron job:
```bash
# Add to crontab
0 0 * * 0 mysql -h your-db-host -u woulder -p'password' woulder -e "DELETE FROM weather_data WHERE timestamp < DATE_SUB(NOW(), INTERVAL 30 DAY);"
```

---

## Troubleshooting

### Backend Won't Start

**Check logs:**
```bash
tail -n 50 woulder.log
```

**Common issues:**
- Port 8080 already in use: `sudo lsof -i :8080`
- Database connection failed: Check `.env` credentials
- Permission denied: `chmod +x woulder-api`
- Missing .env file: Upload from local

### Frontend Shows Blank Page

**Check:**
- `base: '/woulder/'` in `vite.config.ts`
- Files uploaded to correct directory
- No console errors in browser (F12)

### API Calls Fail (CORS Error)

**Update backend CORS:**

Edit `backend/cmd/server/main.go`:
```go
router.Use(cors.New(cors.Config{
    AllowOrigins: []string{"https://alexscott.io"},
    AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders: []string{"Origin", "Content-Type", "Accept"},
    MaxAge:       12 * time.Hour,
}))
```

Rebuild and redeploy backend.

### 502 Bad Gateway

**Backend not running:**
```bash
sudo systemctl status woulder
sudo systemctl start woulder
```

**Wrong port in proxy:**
- Verify backend runs on 8080
- Check Nginx/Apache proxy config

---

## Performance Optimization

### Enable Gzip Compression

**Nginx:**
```nginx
gzip on;
gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript;
```

**Apache (.htaccess):**
```apache
<IfModule mod_deflate.c>
    AddOutputFilterByType DEFLATE text/html text/plain text/xml text/css application/json application/javascript
</IfModule>
```

### Cache Static Assets

**Nginx:**
```nginx
location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
    expires 1y;
    add_header Cache-Control "public, immutable";
}
```

### CDN (Optional)

Use Cloudflare free tier:
1. Add site to Cloudflare
2. Update nameservers
3. Enable caching and optimization
4. SSL auto-configured

---

## Backup Strategy

### Database Backups

**Daily backup script:**
```bash
#!/bin/bash
mysqldump -h your-db-host -u woulder -p'password' woulder > "woulder-backup-$(date +%Y%m%d).sql"

# Keep only last 7 days
find . -name "woulder-backup-*.sql" -mtime +7 -delete
```

Add to cron:
```bash
0 2 * * * /home/username/scripts/backup-woulder.sh
```

### Code Backups

Your code is already in git:
```bash
git add .
git commit -m "Update"
git push origin main
```

---

## Security Checklist

- [ ] HTTPS enabled (SSL certificate)
- [ ] CORS limited to your domain
- [ ] Database credentials in `.env`, not hardcoded
- [ ] `.env` file not in git
- [ ] Firewall configured (only 80, 443, SSH open)
- [ ] SSH key authentication (disable password)
- [ ] Regular security updates
- [ ] Rate limiting on API (future)
- [ ] Input validation on user input (future)

---

## Cost Summary

### Monthly Costs
- **Hosting:** $0 (using existing Namecheap)
- **Database:** $0 (using existing AWS RDS)
- **OpenWeatherMap:** $0 (free tier, 1,000 calls/day)
- **SSL:** $0 (Let's Encrypt)
- **Domain:** $0 (existing)

**Total: $0/month** 🎉

---

## Alternative Deployment: Docker (Advanced)

If your host supports Docker:

### Dockerfile (Backend)
```dockerfile
FROM golang:1.21-alpine
WORKDIR /app
COPY . .
RUN go build -o woulder-api cmd/server/main.go
CMD ["./woulder-api"]
```

### docker-compose.yml
```yaml
version: '3.8'
services:
  backend:
    build: ./backend
    ports:
      - "8080:8080"
    env_file:
      - ./backend/.env
    restart: always
```

Deploy:
```bash
docker-compose up -d
```

---

## Support & Updates

### Check for Updates
- OpenWeatherMap API changes
- Go security updates
- Node.js/React updates
- Database driver updates

### Update Commands
```bash
# Backend
cd backend
go get -u
go mod tidy

# Frontend
cd frontend
npm update
```

---

**Deployment Complete!** 🚀

Your app should now be live at:
- **Frontend:** https://alexscott.io/woulder
- **API:** https://alexscott.io/woulder/api/health

---

## Quick Reference Commands

```bash
# Check backend status
sudo systemctl status woulder

# View logs
sudo journalctl -u woulder -f

# Restart backend
sudo systemctl restart woulder

# Test API
curl https://alexscott.io/woulder/api/health

# Rebuild frontend
cd frontend && npm run build

# Nginx reload
sudo systemctl reload nginx
```

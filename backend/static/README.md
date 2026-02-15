# Job Monitoring Dashboard

A beautiful web-based dashboard for monitoring Woulder's background sync jobs in real-time.

## 🌐 Access the Dashboard

### Local Development
```
http://localhost:8080/jtrack
```

### Production
```
https://woulder.com/jtrack
```

## ✨ Features

- **Real-time Updates** - Automatically refreshes every 5 seconds
- **Active Jobs View** - See running jobs with live progress bars
- **Job History** - Browse past executions with filtering
- **Summary View** - Overview of all job types and their schedules
- **Auto-detection** - Automatically connects to the correct API
- **Beautiful UI** - Modern gradient design with smooth animations
- **Responsive** - Works on desktop, tablet, and mobile

## 📊 Dashboard Views

### 1. Active Jobs
Shows all currently running jobs with:
- Real-time progress bars (e.g., `[████████████░░░░░░░░] 53.5%`)
- Items processed out of total (`1523/2847`)
- Elapsed time and estimated remaining time
- Processing rate (items/second)
- Success/failure counts

### 2. History
Browse past job executions with:
- Job status (completed, failed, running)
- Start time and duration
- Total items processed
- Error messages for failed jobs

### 3. Summary
Overview of all job types showing:
- Last run time
- Current status
- Duration
- Next scheduled run
- Progress for running jobs

## 🚀 Setup & Deployment

### Prerequisites

The dashboard is already integrated into the Woulder backend server. No separate server needed!

### Local Setup

1. **Start the Woulder Backend**:
   ```bash
   cd backend
   go run cmd/server/main.go
   ```

2. **Open the Dashboard**:
   ```
   http://localhost:8080/jtrack
   ```

The dashboard will automatically connect to `http://localhost:8080`

### Production Deployment

#### Automatic Deployment via GitHub Actions

The dashboard auto-deploys when you push changes:

1. **Set up GitHub Secrets** (in your repo settings):
   - `DEPLOY_HOST` - Your server hostname or IP
   - `DEPLOY_USER` - SSH username
   - `DEPLOY_SSH_KEY` - Private SSH key for authentication
   - `DEPLOY_PATH` - Path to Woulder directory (default: `/opt/woulder`)

2. **Push changes**:
   ```bash
   git add backend/static/jtrack.html
   git commit -m "Update job monitor dashboard"
   git push origin main
   ```

3. **GitHub Actions will**:
   - Pull latest code on your server
   - Restart the Woulder service
   - Dashboard is live at `https://yourdomain.com/jtrack`

#### Manual Deployment

1. **Copy dashboard to server**:
   ```bash
   scp backend/static/jtrack.html user@server:/opt/woulder/backend/static/
   ```

2. **Restart Woulder service**:
   ```bash
   ssh user@server
   sudo systemctl restart woulder
   ```

3. **Access dashboard**:
   ```
   https://yourdomain.com/jtrack
   ```

## 🔧 Configuration

### Custom API URL

If your API is on a different host, you can override the auto-detected URL:

1. Open the dashboard
2. Change the API URL in the input field
3. Click "Connect"
4. The setting is saved in localStorage for future visits

### Refresh Interval

The dashboard refreshes every 5 seconds by default. To change:

Edit `backend/static/jtrack.html` line ~417:
```javascript
refreshInterval = setInterval(refreshData, 5000); // Change 5000 to your preferred ms
```

## 📱 Mobile Access

The dashboard is fully responsive. Access it on your phone:

1. Set up port forwarding or public access
2. Navigate to `https://yourdomain.com/jtrack`
3. Monitor jobs on the go!

## 🎨 Customization

### Change Colors

Edit the CSS in `backend/static/jtrack.html`:

```css
body {
    /* Change gradient colors */
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.progress-fill {
    /* Change progress bar color */
    background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
}
```

### Add Custom Metrics

To display additional job metadata:

1. Add fields to `JobExecution` in  `monitoring/job_monitor.go`
2. Update API response in `handlers_monitoring.go`
3. Display in dashboard by modifying `renderJobCard()` function

## 🔒 Security

### Authentication (Optional)

The dashboard currently has no authentication. To add:

**Backend** - Modify `cmd/server/main.go`:
```go
// Add authentication middleware
router.Use(authMiddleware())
router.StaticFile("/jtrack", "./backend/static/jtrack.html")
```

**Dashboard** - Modify `jtrack.html`:
```javascript
async function loadActiveJobs() {
    const response = await fetch(`${apiUrl}/api/monitoring/jobs/active`, {
        headers: {
            'Authorization': `Bearer ${localStorage.getItem('auth_token')}`
        }
    });
    // ...
}
```

### HTTPS

Always use HTTPS in production:

1. Set up SSL certificate (Let's Encrypt recommended)
2. Configure reverse proxy (Nginx/Caddy)
3. Dashboard will use HTTPS automatically

## 📊 Monitoring Tips

### Watch Long-Running Jobs

For monthly low-priority syncs (6-8 hours):

1. Open dashboard on second monitor or tablet
2. Keep "Active Jobs" tab selected
3. Auto-refresh shows live progress

### Check After Deployments

After deploying code changes:

1. Visit `/jtrack`
2. Check "History" tab
3. Verify recent jobs completed successfully

### Debug Failures

When a job fails:

1. Go to "History" tab
2. Find the failed job
3. Click to see error message
4. Check server logs for details

## 🐛 Troubleshooting

### Dashboard shows "Disconnected"

**Check:**
1. Backend server is running
2. API URL is correct
3. CORS is enabled for your domain
4. Firewall allows connections

**Fix:**
```bash
# Check server status
systemctl status woulder

# Check logs
tail -f /var/log/woulder/server.log

# Test API directly
curl http://localhost:8080/api/monitoring/jobs/active
```

### "No Active Jobs" but jobs are running

**Check:**
1. Jobs are using JobMonitor
2. Database migration 000022 ran
3. Sync methods call `StartJob()`

### Progress not updating

**Check:**
1. Browser console for errors (F12)
2. Network tab shows successful API calls
3. Auto-refresh is enabled

## 📝 Development

### Local Development

1. **Edit dashboard**:
   ```bash
   vim backend/static/jtrack.html
   ```

2. **Test changes**:
   - Save file
   - Refresh browser (Ctrl+R)
   - Changes appear immediately (no restart needed)

3. **Test different scenarios**:
   - Start a sync job
   - Watch it appear in dashboard
   - Verify progress updates
   - Check completion status

### Adding New Features

See `renderJobCard()` function to customize job display:

```javascript
function renderJobCard(job) {
    // Add custom fields here
    const customMetric = job.metadata?.custom_field;
    
    return `
        <div class="job-card ${job.status}">
            <!-- Your custom HTML -->
        </div>
    `;
}
```

## 📚 Related Documentation

- **Architecture**: [`plans/job-monitoring-system.md`](../plans/job-monitoring-system.md)
- **Usage Guide**: [`docs/job-monitoring-usage.md`](../docs/job-monitoring-usage.md)
- **API Docs**: [`backend/internal/api/handlers_monitoring.go`](../backend/internal/api/handlers_monitoring.go)
- **CLI Tool**: [`backend/cmd/job_monitor/README.md`](../backend/cmd/job_monitor/README.md)

## 🎯 Example Screenshots

### Active Jobs View
```
╔════════════════════════════════════════════╗
║ Job: high_priority_tick_sync               ║
║ Type: tick_sync                            ║
║ Progress: 1523/2847 (53.5%)                ║
║ [████████████████░░░░░░░░░░░░] 53.5%      ║
║ Success: 1520 | Failed: 3                  ║
║ Elapsed: 3m 35s | Remaining: ~3m 6s       ║
║ Rate: 7.08 items/sec                       ║
╚════════════════════════════════════════════╝
```

## 🤝 Contributing

To improve the dashboard:

1. Edit `backend/static/jtrack.html`
2. Test locally
3. Commit and push
4. GitHub Actions auto-deploys

## ⚡ Performance

- **Load time**: < 1 second
- **Refresh overhead**: Minimal (just JSON API calls)
- **Memory usage**: < 10MB
- **Works offline**: No (requires API connection)

## 🎉 Success!

Your dashboard is now live! Monitor your jobs at:
- **Local**: http://localhost:8080/jtrack
- **Production**: https://woulder.com/jtrack

Enjoy real-time monitoring of all your background sync jobs! 🚀

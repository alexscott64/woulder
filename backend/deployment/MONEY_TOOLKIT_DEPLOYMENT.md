# Money Creek Toolkit Deployment

## Required database migration

The Money Creek toolkit uses migration `000038_add_money_toolkit` to create:

- `woulder.users`
- `woulder.auth_refresh_tokens`
- `woulder.money_projects`
- `woulder.money_features`
- `woulder.money_notes`
- `woulder.money_uploads`

The repo-managed GitHub deployment builds `woulder-migrate`, uploads `backend/internal/database/migrations` to `/opt/woulder/migrations`, and runs:

```bash
MIGRATIONS_PATH=/opt/woulder/migrations /opt/woulder/woulder-migrate up
```

For local development from `backend/`, run:

```bash
go run cmd/migrate/main.go up
```

## Required production environment

Set these in GitHub Actions secrets or the deployed `/opt/woulder/.env`:

```bash
APP_JWT_SECRET=replace-with-a-long-random-secret
APP_ACCESS_TOKEN_MINUTES=15
APP_REFRESH_TOKEN_DAYS=30
APP_ADMIN_EMAIL=admin@example.com
APP_ADMIN_PASSWORD=replace-with-strong-password
APP_ADMIN_DISPLAY_NAME=Money Creek Admin
UPLOAD_STORAGE_DRIVER=local
UPLOAD_DIR=/var/lib/woulder/uploads
UPLOAD_MAX_BYTES=10485760
```

`MONEY_USERNAME` and `MONEY_PASSWORD` are still accepted as backwards-compatible fallbacks when `APP_ADMIN_EMAIL` and `APP_ADMIN_PASSWORD` are not set.

On startup, the server creates or reconciles the configured Money Creek admin account. Restart the service after changing admin credentials.

## Upload directory persistence and permissions

Use persistent storage outside `/opt/woulder` so deploys do not remove uploads:

```bash
sudo mkdir -p /var/lib/woulder/uploads
sudo chown -R woulder:woulder /var/lib/woulder
sudo chmod 750 /var/lib/woulder/uploads
```

If the service still runs as `root`, ownership is less strict, but moving to a dedicated `woulder` service user is recommended. The deployment workflow creates `/var/lib/woulder/uploads` and attempts to set ownership for a `woulder` user when present.

## SPA rewrite for `/money`

Direct navigation to `/money` must serve `frontend/dist/index.html`. The repo setup scripts include this Nginx fallback:

```nginx
location / {
    try_files $uri $uri/ /index.html;
}
```

Keep `/api/` proxied to the Go backend before the SPA fallback.

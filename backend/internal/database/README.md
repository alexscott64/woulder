# Woulder Database Setup

This directory contains the PostgreSQL database setup for Woulder.

## Files

- **`setup_postgres.sql`** - Complete database setup file (schema + seed data)
- **`db.go`** - Database connection and query functions

## Automatic Setup

The backend automatically initializes the database on first run:

1. Set environment variables in `.env`:
   ```bash
   DB_HOST=your-database-host
   DB_PORT=5432
   DB_USER=your-username
   DB_PASSWORD=your-password
   DB_NAME=your-database-name
   DB_SSLMODE=require
   ```

2. Start the backend:
   ```bash
   cd backend
   go run ./cmd/server
   ```

3. The backend will:
   - Check if the `woulder` schema exists
   - If not, run `setup_postgres.sql` automatically
   - Create all tables, indexes, triggers
   - Seed initial location and river data

## Manual Setup

If you want to set up the database manually (optional):

```bash
psql -h your-host -U your-user -d your-database -f setup_postgres.sql
```

## Database Schema

### Tables

- **`woulder.locations`** - Climbing locations (9 locations)
- **`woulder.weather_data`** - Historical and forecast weather
- **`woulder.rivers`** - River crossing information with USGS gauge data

### Features

- ✓ Timezone-aware timestamps (`TIMESTAMPTZ`)
- ✓ Precise decimal values for coordinates and weather
- ✓ Data validation with CHECK constraints
- ✓ Automatic `updated_at` triggers
- ✓ Performance indexes on frequently queried columns
- ✓ Foreign key CASCADE rules for data integrity

## Adding New Locations

See [ADDING_LOCATIONS.md](../../../ADDING_LOCATIONS.md) in the project root.

## Troubleshooting

**Schema already exists?**
The setup is idempotent - it uses `CREATE TABLE IF NOT EXISTS` and `ON CONFLICT DO NOTHING`, so it's safe to run multiple times.

**Connection issues?**
- Verify environment variables are set correctly
- Check that your database allows SSL connections (`DB_SSLMODE=require`)
- Ensure your IP is whitelisted in database firewall rules
- Test connection: `psql -h host -U user -d dbname`

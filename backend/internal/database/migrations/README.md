# Woulder Database Migrations

This directory contains versioned database migrations for Woulder using a custom Go migration tool.

## Migration Files

Migrations are stored as numbered SQL files with `.up.sql` and `.down.sql` extensions:

- `000001_initial_schema.up.sql` - Creates all tables, indexes, triggers
- `000001_initial_schema.down.sql` - Drops all tables and schema
- `000002_seed_data.up.sql` - Inserts initial areas, locations, and rivers
- `000002_seed_data.down.sql` - Removes seeded data

## Usage

### Check Current Version

```bash
cd backend/cmd/migrate
go run main.go version
```

### Apply All Pending Migrations

```bash
cd backend/cmd/migrate
go run main.go up
```

### Rollback All Migrations

```bash
cd backend/cmd/migrate
go run main.go down
```

### Step Migrations

Apply next migration:
```bash
go run main.go step 1
```

Rollback last migration:
```bash
go run main.go step -1
```

### Force Version (Use with Caution)

Force database to specific version without running migrations:
```bash
go run main.go force 2
```

**Warning**: This only updates the version tracking, it doesn't actually run migrations. Use only when manually fixing migration state.

## Creating New Migrations

1. Create new migration files with the next version number:
   ```bash
   000003_your_migration_name.up.sql
   000003_your_migration_name.down.sql
   ```

2. Write the forward migration in the `.up.sql` file
3. Write the rollback migration in the `.down.sql` file
4. Test locally:
   ```bash
   go run cmd/migrate/main.go up
   go run cmd/migrate/main.go down
   go run cmd/migrate/main.go up
   ```

## Migration Best Practices

- **Always write down migrations** - Every up migration must have a corresponding down migration
- **Test rollbacks** - Always test that migrations can be rolled back successfully
- **Use transactions** - The migration tool wraps each migration in a transaction
- **Idempotent operations** - Use `IF EXISTS` / `IF NOT EXISTS` where possible
- **No data loss** - Down migrations should preserve data when possible
- **Small migrations** - Keep migrations focused on single changes

## Fresh Install vs Migrations

### Fresh Install (Recommended for New Users)

For new databases, use the complete setup script:

```bash
# PostgreSQL
psql -f backend/internal/database/setup_postgres.sql

# SQLite
sqlite3 woulder.db < backend/internal/database/schema.sql
sqlite3 woulder.db < backend/internal/database/seed.sql
```

### Existing Database Upgrade

For existing Woulder installations, use migrations:

```bash
cd backend/cmd/migrate
go run main.go up
```

## Migration State Tracking

The migration tool creates a `schema_migrations` table to track applied migrations:

```sql
CREATE TABLE schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
```

## Example Migration

**000003_add_user_favorites.up.sql**:
```sql
CREATE TABLE IF NOT EXISTS woulder.user_favorites (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    location_id INTEGER NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (location_id) REFERENCES woulder.locations(id) ON DELETE CASCADE,
    UNIQUE(user_id, location_id)
);

CREATE INDEX IF NOT EXISTS idx_user_favorites_user_id
    ON woulder.user_favorites(user_id);
```

**000003_add_user_favorites.down.sql**:
```sql
DROP TABLE IF EXISTS woulder.user_favorites CASCADE;
```

## Troubleshooting

### "Migration failed" Error

If a migration fails midway:
1. Check the database state
2. Fix the issue manually if needed
3. Use `force` to set the correct version
4. Re-run migrations

### "Missing up/down migration" Error

Ensure both `.up.sql` and `.down.sql` files exist for each version.

### "Already up to date"

This means all available migrations have been applied. This is normal.

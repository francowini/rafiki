# External Database Setup

Quick guide for using external PostgreSQL (PlanetScale, Neon, AWS RDS, etc.) instead of local Docker PostgreSQL.

## Configuration

Add to `/opt/rafiki/.env` on production server:

```bash
POSTGRES_HOST=your-db-host.example.com  # Triggers external DB mode
POSTGRES_PORT=5432
POSTGRES_USER=your_username
POSTGRES_PASSWORD=your_password
POSTGRES_DB=your_database_name
POSTGRES_SSLMODE=require                # Options: disable, require, verify-ca, verify-full
POSTGRES_DISABLETLS=false
```

## How It Works

`devops/deploy.sh` detects `POSTGRES_HOST` and:
- Uses `docker-compose.external-db.yml` overlay
- Disables local PostgreSQL container (**saves ~256MB RAM**)
- Connects app to external database with SSL

## Database Permissions

Your database user needs DDL + DML permissions (migrations run on app startup):

```sql
GRANT CREATE ON SCHEMA public TO your_username;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO your_username;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO your_username;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO your_username;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO your_username;
```

## Deployment

```bash
# Local machine
make deploy

# On server
cd /opt/rafiki && git pull && sudo ./devops/deploy.sh
```

Deploy script auto-detects external DB and applies correct configuration.

## Rollback

To revert to local Docker PostgreSQL:

```bash
# Comment out POSTGRES_HOST in .env
# POSTGRES_HOST=your-db-host.example.com
make deploy  # Automatically uses local PostgreSQL
```

## Code Changes

- **[api/services/partners/main.go:86-96](../api/services/partners/main.go#L86-L96)** - Added Port + SSLMode config fields
- **[business/sdk/sqldb/sqldb.go:37-102](../business/sdk/sqldb/sqldb.go#L37-L102)** - Connection logic with SSL validation
- **[docker-compose.external-db.yml](../docker-compose.external-db.yml)** - Overlay that disables local postgres
- **[devops/deploy.sh:108-114](../devops/deploy.sh#L108-L114)** - External DB detection logic

## Verified Providers

- ✅ **PlanetScale PostgreSQL** (2025 offering - NOT MySQL/Vitess)
- ✅ Neon, Supabase, AWS RDS, DigitalOcean, Aiven

## Notes

- Local development **unchanged** (uses Docker PostgreSQL)
- Migrations run **automatically** on app startup
- SSL **required** for external databases
- Expected latency: 5-20ms (vs <1ms local)
- Backward compatible with existing configs

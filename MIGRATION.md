# Database Migration Guide

## From In-Memory to PostgreSQL

The application has been migrated from in-memory storage to PostgreSQL. This guide explains the changes and how to migrate existing data.

## What Changed

1. **Storage Backend**: Changed from in-memory map to PostgreSQL database
2. **ORM**: Added GORM for database operations
3. **Schema**: Secrets are now stored in a `secrets` table with JSONB for key-value data

## Database Schema

```sql
CREATE TABLE secrets (
    id UUID PRIMARY KEY,
    name VARCHAR NOT NULL,
    description TEXT,
    data JSONB,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

## Migration Steps

### 1. Install PostgreSQL

See the main README.md for PostgreSQL installation instructions.

### 2. Create Database

```bash
createdb vault
# Or using psql:
psql -U postgres
CREATE DATABASE vault;
CREATE USER vault WITH PASSWORD 'vault';
GRANT ALL PRIVILEGES ON DATABASE vault TO vault;
```

### 3. Configure Environment Variables

Set the database connection in your environment or `.env` file:

```bash
export DB_HOST=localhost
export DB_USER=vault
export DB_PASSWORD=vault
export DB_NAME=vault
export DB_PORT=5432
export DB_SSLMODE=disable
```

### 4. Run the Application

The schema will be automatically created on first run:

```bash
cd api
go run main.go
```

You should see: `Database connected and migrated successfully`

## Data Migration (If Needed)

If you had existing in-memory data, you'll need to manually recreate it through the UI or API, as in-memory data is not persisted.

## Rollback

If you need to rollback to in-memory storage, you would need to:
1. Revert to the previous version of `main.go`
2. Remove GORM dependencies from `go.mod`

However, this is not recommended as PostgreSQL provides persistent storage.

## Benefits of PostgreSQL

- **Persistence**: Data survives application restarts
- **ACID Compliance**: Transactional guarantees
- **Scalability**: Can handle large numbers of secrets
- **Backup**: Easy to backup and restore
- **JSONB Support**: Efficient storage and querying of key-value pairs
- **Production Ready**: Industry-standard database


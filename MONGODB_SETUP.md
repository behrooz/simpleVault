# MongoDB Migration Complete

The application has been successfully migrated from PostgreSQL to MongoDB.

## What Changed

1. **Database**: Changed from PostgreSQL to MongoDB
2. **Driver**: Replaced GORM with official MongoDB Go Driver
3. **Storage**: Secrets stored in MongoDB `secrets` collection
4. **Schema**: No manual schema creation needed - MongoDB creates collections automatically

## Installation

If you encounter network issues downloading dependencies, run:

```bash
cd api
go mod download
go mod tidy
```

If you still have issues, try:
```bash
export GOPROXY=direct
go mod download
```

## Configuration

### Environment Variables

**Option 1: Connection String (Recommended)**
```bash
export MONGODB_URI="mongodb://vault:vault@localhost:27017/vault?authSource=admin"
```

**Option 2: Individual Variables**
```bash
export DB_HOST=localhost
export DB_USER=vault
export DB_PASSWORD=vault
export DB_NAME=vault
export DB_PORT=27017
```

## Quick Start

1. **Start MongoDB** (if not using Docker Compose):
```bash
# Using Docker
docker run -d -p 27017:27017 --name mongodb \
  -e MONGO_INITDB_ROOT_USERNAME=vault \
  -e MONGO_INITDB_ROOT_PASSWORD=vault \
  mongo:7

# Or using local MongoDB
mongod
```

2. **Run the API**:
```bash
cd api
go run main.go
```

## Docker Compose

The `docker-compose.yml` has been updated to use MongoDB:
```bash
docker-compose up -d
```

This will start:
- MongoDB on port 27017
- API on port 8080
- UI on port 3000

## Benefits of MongoDB

- **No Schema**: Collections created automatically
- **Flexible**: Easy to store nested objects (key-value pairs)
- **JSON Native**: Perfect for storing secret data as objects
- **Scalable**: Horizontal scaling support
- **Simple**: No migrations needed

## Data Structure

Secrets are stored as documents in the `secrets` collection:

```json
{
  "_id": "uuid-string",
  "name": "my-secret",
  "description": "Description",
  "data": {
    "key1": "value1",
    "key2": "value2"
  },
  "createdAt": ISODate("2024-01-01T00:00:00Z"),
  "updatedAt": ISODate("2024-01-01T00:00:00Z")
}
```

## Troubleshooting

### Connection Issues

1. Check MongoDB is running:
```bash
mongosh --eval "db.adminCommand('ping')"
```

2. Verify connection string format:
```
mongodb://[username:password@]host[:port][/database][?options]
```

3. Check authentication:
```bash
mongosh -u vault -p vault --authenticationDatabase admin
```

### Build Issues

If you see dependency errors:
```bash
cd api
rm go.sum
go mod tidy
go build
```


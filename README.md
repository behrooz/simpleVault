# Simple Vault

A simple secret management system with a React UI and Go API backend for securely storing and managing secrets.

## Features

- 🔐 Create, read, update, and delete secrets
- 🎨 Modern React UI with a beautiful interface
- 🚀 Fast Go API backend with Gin framework
- 💾 MongoDB for persistent storage

## Architecture

- **Frontend**: React with Vite
- **Backend**: Go with Gin framework
- **Database**: MongoDB (persistent storage)
- **Driver**: Official MongoDB Go Driver

## Prerequisites

- Go 1.21 or later
- Node.js 18+ and npm
- MongoDB 4.4+ (or use Docker Compose)
- Docker (optional, for containerized deployment)

## Getting Started

### Database Setup

#### Option 1: Using Docker Compose (Recommended)

The easiest way is to use Docker Compose which includes MongoDB:

```bash
docker-compose up -d
```

This will start MongoDB, API, and UI services.

#### Option 2: Local MongoDB

1. Install and start MongoDB:
```bash
# On Ubuntu/Debian
sudo apt-get install mongodb
sudo systemctl start mongodb

# On macOS (using Homebrew)
brew install mongodb-community
brew services start mongodb-community
```

2. Create database and user (optional, MongoDB creates databases automatically):
```bash
mongosh
```

Then in the MongoDB shell:
```javascript
use vault
db.createUser({
  user: "vault",
  pwd: "vault",
  roles: [{ role: "readWrite", db: "vault" }]
})
```

3. Set environment variables:
```bash
export DB_HOST=localhost
export DB_USER=vault
export DB_PASSWORD=vault
export DB_NAME=vault
export DB_PORT=27017
```

Or use a connection string:
```bash
export MONGODB_URI="mongodb://vault:vault@localhost:27017/vault?authSource=admin"
```

### Backend Setup

1. Navigate to the API directory:
```bash
cd api
```

2. Install dependencies:
```bash
go mod download
```

3. Run the API server:
```bash
go run main.go
```

The API will start on `http://localhost:8080` by default. You can change the port by setting the `PORT` environment variable.

**Note**: MongoDB will automatically create the database and collections on first use. No manual schema creation is needed.

### Frontend Setup

1. Navigate to the UI directory:
```bash
cd ui
```

2. Install dependencies:
```bash
npm install
```

3. Start the development server:
```bash
npm run dev
```

The UI will start on `http://localhost:5173` by default.

4. Configure the API URL (optional):
Create a `.env` file in the `ui` directory:
```
VITE_API_URL=http://localhost:8080/api/v1
```

## API Endpoints

- `GET /api/v1/secrets` - Get all secrets
- `GET /api/v1/secrets/:id` - Get a specific secret
- `POST /api/v1/secrets` - Create a new secret
- `PUT /api/v1/secrets/:id` - Update a secret
- `DELETE /api/v1/secrets/:id` - Delete a secret

## Docker Deployment

### Build Images

```bash
# Build API image
cd api
docker build -t simple-vault-api:latest .

# Build UI image
cd ../ui
docker build -t simple-vault-ui:latest .
```

### Run with Docker Compose

```bash
docker-compose up -d
```

This will start both the API and UI services.

## Usage

1. **Create a Secret**:
   - Click "Create Secret" button
   - Enter a name and optional description
   - Add key-value pairs for your secrets
   - Click "Create"

2. **Edit a Secret**:
   - Click the edit icon (✏️) on any secret card
   - Modify the secret data
   - Click "Update"

3. **Delete a Secret**:
   - Click the delete icon (🗑️) on any secret card
   - Confirm the deletion

## Database Configuration

The application uses MongoDB for persistent storage. Secrets are stored in a `secrets` collection with the following structure:

- `_id` (string, primary key - UUID)
- `name` (string, required)
- `description` (string, optional)
- `data` (object/map, stores key-value pairs)
- `createdAt` (timestamp)
- `updatedAt` (timestamp)

### Environment Variables

The API supports the following database configuration options:

**Option 1: Individual variables**
- `DB_HOST` - Database host (default: localhost)
- `DB_USER` - Database user (optional, for authentication)
- `DB_PASSWORD` - Database password (optional, for authentication)
- `DB_NAME` - Database name (default: vault)
- `DB_PORT` - Database port (default: 27017)

**Option 2: Connection string (recommended)**
- `MONGODB_URI` - Full MongoDB connection string
  - Example: `mongodb://user:password@host:port/dbname?authSource=admin`
  - Example (no auth): `mongodb://localhost:27017/vault`

### Database Collections

MongoDB automatically creates collections when first used. The `secrets` collection will be created automatically when you create your first secret. An index is created on the `name` field for faster lookups.

## Security Considerations

⚠️ **Important**: This is a simple vault for development/testing purposes. For production use, consider:

- Adding authentication and authorization
- Encrypting secrets at rest (MongoDB supports encryption at rest, or use application-level encryption)
- Using MongoDB's built-in encryption features
- Implementing audit logging
- Adding rate limiting
- Using connection pooling (already implemented in MongoDB driver)
- Regular database backups
- Using external secret management systems (e.g., HashiCorp Vault) for production workloads

## Development

### Running Tests

```bash
# Backend tests
cd api
go test ./...

# Frontend tests (if configured)
cd ui
npm test
```

### Project Structure

```
simple-vault/
├── api/
│   ├── main.go          # Go API server
│   └── go.mod           # Go dependencies
├── ui/
│   ├── src/
│   │   ├── App.jsx      # Main React component
│   │   ├── App.css      # Styles
│   │   └── main.jsx     # React entry point
│   ├── package.json     # Node dependencies
│   └── vite.config.js   # Vite configuration
├── docker-compose.yml   # Docker Compose configuration
└── README.md            # This file
```

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.


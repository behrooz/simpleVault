#!/bin/bash

# Simple Vault - Quick Start Script

echo "🔐 Starting Simple Vault..."
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or later."
    exit 1
fi

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo "❌ Node.js is not installed. Please install Node.js 18 or later."
    exit 1
fi

# Check if PostgreSQL is running (optional check)
if ! command -v psql &> /dev/null; then
    echo "⚠️  PostgreSQL client not found. Make sure PostgreSQL is running and accessible."
    echo "   You can use Docker Compose instead: docker-compose up"
else
    # Try to connect to PostgreSQL
    if [ -z "$DATABASE_URL" ]; then
        export DB_HOST=${DB_HOST:-localhost}
        export DB_USER=${DB_USER:-vault}
        export DB_PASSWORD=${DB_PASSWORD:-vault}
        export DB_NAME=${DB_NAME:-vault}
        export DB_PORT=${DB_PORT:-5432}
        export DB_SSLMODE=${DB_SSLMODE:-disable}
    fi
fi

# Start API in background
echo "🚀 Starting API server..."
cd api
go run main.go &
API_PID=$!
cd ..

# Wait for API to start
sleep 3

# Start UI
echo "🎨 Starting UI..."
cd ui
npm run dev &
UI_PID=$!
cd ..

echo ""
echo "✅ Simple Vault is running!"
echo "   API: http://localhost:8080"
echo "   UI:  http://localhost:5173"
echo ""
echo "📝 Make sure PostgreSQL is running and accessible"
echo "   Or use: docker-compose up"
echo ""
echo "Press Ctrl+C to stop all services"

# Wait for user interrupt
trap "kill $API_PID $UI_PID 2>/dev/null; exit" INT TERM
wait


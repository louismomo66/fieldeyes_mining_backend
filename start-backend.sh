#!/bin/bash

echo "🚀 Starting Mining Finance Backend..."

# Check if PostgreSQL container is running
if ! docker ps | grep -q "mining_postgres"; then
    echo "📦 Starting PostgreSQL container..."
    docker-compose up -d postgres redis
    sleep 5
fi

# Set environment variables for Docker PostgreSQL
export DB_HOST=localhost
export DB_PORT=5433
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=mining_data
export JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
export PORT=9006

echo "🔗 Connecting to PostgreSQL on port 5433..."
echo "🌐 Starting server on port 9006..."

# Start the Go application
go run cmd/api/main.go

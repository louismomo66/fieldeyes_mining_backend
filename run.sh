#!/bin/bash

# Mining Finance System Backend Startup Script

echo "Starting Mining Finance System Backend..."

# Check if .env file exists
if [ ! -f .env ]; then
    echo "Creating .env file from template..."
    cp env.example .env
    echo "Please edit .env file with your configuration before running again."
    exit 1
fi

# Load environment variables
export $(cat .env | grep -v '^#' | xargs)

# Check if PostgreSQL is running
echo "Checking PostgreSQL connection..."
if ! pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER > /dev/null 2>&1; then
    echo "PostgreSQL is not running or not accessible."
    echo "Please start PostgreSQL and ensure it's accessible at $DB_HOST:$DB_PORT"
    exit 1
fi

# Install dependencies
echo "Installing dependencies..."
go mod tidy

# Run the application
echo "Starting server on port $PORT..."
go run cmd/api/main.go




#!/bin/bash
# Startup script for Mining Finance Backend
# This script is called by systemd service

set -e  # Exit on error

# Add Go to PATH (Go is installed at /usr/lib/go-1.22/bin/go)
export PATH="$PATH:/usr/lib/go-1.22/bin"

# Source profile to get any other environment variables
source /home/administrator/.bashrc 2>/dev/null || source /home/administrator/.profile 2>/dev/null || true

# Change to backend directory
cd /var/www/field-eyes/mining/fieldeyes_mining_backend

# Source environment if .env exists
if [ -f .env ]; then
    set -a
    source .env
    set +a
fi

# Ensure Docker is running
if command -v docker &> /dev/null; then
    # Start postgres container if docker-compose is available
    if [ -f docker-compose.yml ]; then
        docker-compose up -d postgres || true
        # Wait for postgres
        echo "Waiting for Postgres..."
        for i in {1..30}; do
            if nc -z ${DB_HOST:-localhost} ${DB_PORT:-5433} 2>/dev/null; then
                echo "Postgres is ready"
                break
            fi
            if [ $i -eq 30 ]; then
                echo "ERROR: Postgres not reachable on ${DB_HOST:-localhost}:${DB_PORT:-5433}"
                exit 1
            fi
            sleep 1
        done
    fi
fi

# Run the backend using make start
exec make start


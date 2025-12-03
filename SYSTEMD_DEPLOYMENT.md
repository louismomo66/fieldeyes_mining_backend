# Systemd Service Deployment Guide for Mining Finance Backend

## Service File Location

The systemd service file is located at:
- **Service file**: `mining-finance-backend.service`
- **Installation path**: `/etc/systemd/system/mining-finance-backend.service`
- **Backend directory**: `/var/www/field-eyes/mining/fieldeyes_mining_backend`

## Installation Steps

### 1. Copy the Service File

```bash
# Copy the service file to systemd directory
sudo cp mining-finance-backend.service /etc/systemd/system/

# Set proper permissions
sudo chmod 644 /etc/systemd/system/mining-finance-backend.service
```

### 2. Reload Systemd

```bash
# Reload systemd to recognize the new service
sudo systemctl daemon-reload
```

### 3. Enable the Service (Optional - for auto-start on boot)

```bash
# Enable service to start on boot
sudo systemctl enable mining-finance-backend.service
```

### 4. Start the Service

```bash
# Start the backend service
sudo systemctl start mining-finance-backend.service

# Check status
sudo systemctl status mining-finance-backend.service
```

## Service Management Commands

### Start/Stop/Restart

```bash
# Start the service
sudo systemctl start mining-finance-backend

# Stop the service
sudo systemctl stop mining-finance-backend

# Restart the service
sudo systemctl restart mining-finance-backend

# Reload configuration (if service file changed)
sudo systemctl daemon-reload
sudo systemctl restart mining-finance-backend
```

### Check Status

```bash
# Check service status
sudo systemctl status mining-finance-backend

# Check if service is running
sudo systemctl is-active mining-finance-backend

# Check if service is enabled
sudo systemctl is-enabled mining-finance-backend
```

### View Logs

```bash
# View service logs
sudo journalctl -u mining-finance-backend -f

# View last 100 lines
sudo journalctl -u mining-finance-backend -n 100

# View logs from today
sudo journalctl -u mining-finance-backend --since today

# View application logs (from StandardOutput/StandardError)
sudo tail -f /var/log/mining-finance-backend.log
```

## Configuration

### Environment Variables

The service file includes default environment variables. You can:

1. **Use .env file** (Recommended): Create a `.env` file in `/var/www/field-eyes/mining/fieldeyes_mining_backend/`
   ```bash
   cd /var/www/field-eyes/mining/fieldeyes_mining_backend
   cp env.example .env
   # Edit .env with your configuration
   nano .env
   ```

2. **Override in service file**: Edit `/etc/systemd/system/mining-finance-backend.service` and add/modify Environment variables:
   ```ini
   Environment="DB_HOST=your-db-host"
   Environment="DB_PORT=5433"
   # etc.
   ```

3. **Use systemd override** (Best practice):
   ```bash
   # Create override directory
   sudo systemctl edit mining-finance-backend
   
   # Add your overrides:
   [Service]
   Environment="DB_HOST=custom-host"
   Environment="DB_PORT=5432"
   ```

### Working Directory

The service runs from `/var/www/field-eyes/mining/fieldeyes_mining_backend`. Ensure:
- The directory exists
- The `administrator` user has read/write access
- The Makefile is present
- Go is installed and in PATH

## Troubleshooting

### Service Won't Start

1. **Check service status**:
   ```bash
   sudo systemctl status mining-finance-backend
   ```

2. **Check logs**:
   ```bash
   sudo journalctl -u mining-finance-backend -n 50
   sudo tail -f /var/log/mining-finance-backend.log
   ```

3. **Verify paths**:
   ```bash
   # Check if directory exists
   ls -la /var/www/field-eyes/mining/fieldeyes_mining_backend
   
   # Check if Makefile exists
   ls -la /var/www/field-eyes/mining/fieldeyes_mining_backend/Makefile
   
   # Check if make is available
   which make
   ```

4. **Test make command manually**:
   ```bash
   cd /var/www/field-eyes/mining/fieldeyes_mining_backend
   make start
   ```

### Database Connection Issues

1. **Check if PostgreSQL is running**:
   ```bash
   # Check docker containers
   docker ps
   
   # Or check if postgres is running on port 5433
   nc -z localhost 5433
   ```

2. **Start database**:
   ```bash
   cd /var/www/field-eyes/mining/fieldeyes_mining_backend
   make docker-up
   ```

### Permission Issues

1. **Check ownership**:
   ```bash
   ls -la /var/www/field-eyes/mining/fieldeyes_mining_backend
   ```

2. **Fix ownership if needed**:
   ```bash
   sudo chown -R administrator:administrator /var/www/field-eyes/mining/fieldeyes_mining_backend
   ```

### Port Already in Use

If port 9006 is already in use:

1. **Check what's using the port**:
   ```bash
   sudo lsof -i :9006
   ```

2. **Kill the process** (if it's an old instance):
   ```bash
   sudo kill -9 <PID>
   ```

3. **Or change the port** in the service file or .env:
   ```bash
   # In .env file
   PORT=9007
   ```

## Service File Details

### Key Settings

- **Type=simple**: The service runs a single process (go run)
- **User=administrator**: Runs as administrator user
- **WorkingDirectory**: Set to backend directory
- **ExecStart**: Runs `make start` which handles docker and starts the Go server
- **ExecStop**: Runs `make stop` to gracefully stop the server
- **Restart=on-failure**: Automatically restarts on failure
- **RestartSec=10**: Waits 10 seconds before restarting

### Dependencies

- **After=network.target docker.service**: Ensures network and docker are available
- **Requires=network.target**: Requires network to be up

## Verification

After installation, verify everything works:

```bash
# 1. Check service is running
sudo systemctl status mining-finance-backend

# 2. Check backend is responding
curl http://localhost:9006/api/v1/health

# 3. Check logs for errors
sudo tail -f /var/log/mining-finance-backend.log

# 4. Test API endpoint
curl http://localhost:9006/api/v1/health
```

## Notes

- The service uses `make start` which automatically:
  - Starts docker-compose postgres if available
  - Waits for postgres to be ready
  - Starts the Go server with proper environment variables

- Logs are written to:
  - Systemd journal: `sudo journalctl -u mining-finance-backend`
  - Application log: `/var/log/mining-finance-backend.log`

- The service will automatically restart on failure (after 10 seconds)

- To disable auto-start on boot:
  ```bash
  sudo systemctl disable mining-finance-backend
  ```




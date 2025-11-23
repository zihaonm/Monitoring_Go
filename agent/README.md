# Monitoring Agent

A lightweight system monitoring agent that exposes system metrics via HTTP API.

## Features

- ✅ **CPU Metrics** - Usage percentage, core count, load average
- ✅ **Memory Metrics** - Total, used, available, swap
- ✅ **Disk Metrics** - Usage per mount point
- ✅ **Network Metrics** - Traffic statistics per interface
- ✅ **Service Monitoring** - Check if common services are running
- ✅ **Port Monitoring** - Check which ports are listening
- ✅ **Lightweight** - Single binary, ~5-8MB, minimal resource usage
- ✅ **Authentication** - Optional token-based authentication
- ✅ **JSON API** - RESTful JSON endpoints

## Quick Start

### Build

```bash
# Build for current platform
go build -o monitoring-agent

# Build for Linux (from Mac/Windows)
GOOS=linux GOARCH=amd64 go build -o monitoring-agent-linux-amd64

# Build for ARM (Raspberry Pi, etc.)
GOOS=linux GOARCH=arm64 go build -o monitoring-agent-linux-arm64
```

### Run

```bash
# Simple run (port 9100, no auth)
./monitoring-agent

# Custom port
./monitoring-agent -port 8080

# With authentication
./monitoring-agent -token YOUR_SECRET_TOKEN

# Custom metrics interval (default: 10s)
./monitoring-agent -interval 5
```

### Install as Service (Linux)

```bash
# Make install script executable
chmod +x install.sh

# Run installation (requires root)
sudo ./install.sh

# Or with custom settings
sudo AGENT_PORT=9100 AUTH_TOKEN=mysecret ./install.sh
```

### Manual Installation

```bash
# Copy binary
sudo cp monitoring-agent /usr/local/bin/

# Create systemd service
sudo nano /etc/systemd/system/monitoring-agent.service
# (paste service file content - see install.sh)

# Enable and start
sudo systemctl daemon-reload
sudo systemctl enable monitoring-agent
sudo systemctl start monitoring-agent
```

## API Endpoints

### GET /metrics

Returns all system metrics in JSON format.

**Without authentication:**
```bash
curl http://localhost:9100/metrics
```

**With authentication (header):**
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" http://localhost:9100/metrics
```

**With authentication (query param):**
```bash
curl http://localhost:9100/metrics?token=YOUR_TOKEN
```

**Example Response:**
```json
{
  "hostname": "web-server-01",
  "timestamp": "2025-10-30T10:30:00Z",
  "status": "healthy",
  "cpu": {
    "usage_percent": 45.2,
    "cores": 4,
    "load_average": "1.23 1.45 1.67"
  },
  "memory": {
    "total_mb": 8192,
    "used_mb": 4521,
    "available_mb": 3671,
    "usage_percent": 55.2,
    "swap_total_mb": 2048,
    "swap_used_mb": 128,
    "swap_percent": 6.25
  },
  "disk": [
    {
      "mount": "/",
      "total_gb": 100.0,
      "used_gb": 65.5,
      "available_gb": 34.5,
      "usage_percent": 65.5,
      "filesystem": "/dev/sda1"
    }
  ],
  "network": {
    "rx_bytes": 1234567890,
    "tx_bytes": 9876543210,
    "rx_mb": 1177.38,
    "tx_mb": 9418.04,
    "interfaces": [
      {
        "name": "eth0",
        "rx_bytes": 1234567890,
        "tx_bytes": 9876543210,
        "rx_mb": 1177.38,
        "tx_mb": 9418.04
      }
    ]
  },
  "services": [
    {
      "name": "nginx",
      "status": "running",
      "pid": 1234
    },
    {
      "name": "mysql",
      "status": "running",
      "pid": 5678
    }
  ],
  "ports": [
    {
      "port": 80,
      "status": "listening"
    },
    {
      "port": 443,
      "status": "listening"
    },
    {
      "port": 3306,
      "status": "listening"
    }
  ]
}
```

### GET /health

Simple health check endpoint.

```bash
curl http://localhost:9100/health
```

**Response:**
```json
{
  "status": "ok",
  "last_update": "2025-10-30T10:30:00Z"
}
```

### GET /

Web interface showing agent information.

```bash
curl http://localhost:9100/
# Or open in browser: http://localhost:9100/
```

## Configuration

### Command Line Flags

- `-port` - Port to listen on (default: 9100)
- `-token` - Authentication token (optional)
- `-interval` - Metrics collection interval in seconds (default: 10)

### Environment Variables

```bash
# Set via environment
export AGENT_PORT=9100
export AUTH_TOKEN=mysecret
./monitoring-agent
```

## Integration with Monitoring Dashboard

Add this agent as an HTTP service in your monitoring dashboard:

1. **Service Type:** HTTP/HTTPS
2. **URL:** `http://server-ip:9100/metrics`
3. **Headers:** `Authorization: Bearer YOUR_TOKEN` (if using auth)
4. **Check Interval:** 60 seconds (or as needed)

The dashboard can parse the JSON response and alert on thresholds:
- CPU usage > 80%
- Memory usage > 90%
- Disk usage > 85%
- Service not running
- Port not listening

## Monitoring Multiple Servers

Deploy the agent on each server and add them to your dashboard:

```
Server 1: http://192.168.1.10:9100/metrics
Server 2: http://192.168.1.11:9100/metrics
Server 3: http://192.168.1.12:9100/metrics
```

## Troubleshooting

### Check if agent is running
```bash
systemctl status monitoring-agent
```

### View logs
```bash
journalctl -u monitoring-agent -f
```

### Test locally
```bash
curl http://localhost:9100/metrics
```

### Check port
```bash
netstat -tuln | grep 9100
# or
ss -tuln | grep 9100
```

### Firewall
```bash
# Allow port through firewall (if needed)
sudo ufw allow 9100/tcp
# or
sudo firewall-cmd --add-port=9100/tcp --permanent
sudo firewall-cmd --reload
```

## Uninstall

```bash
# Stop and disable service
sudo systemctl stop monitoring-agent
sudo systemctl disable monitoring-agent

# Remove files
sudo rm /usr/local/bin/monitoring-agent
sudo rm /etc/systemd/system/monitoring-agent.service

# Reload systemd
sudo systemctl daemon-reload
```

## Security Considerations

1. **Use Authentication** - Always use `-token` flag in production
2. **Firewall** - Only expose port to your monitoring server
3. **HTTPS** - Consider putting behind reverse proxy with SSL
4. **Minimal Permissions** - Agent runs as root but only reads system files

## Building for Multiple Platforms

```bash
# Linux AMD64 (most common)
GOOS=linux GOARCH=amd64 go build -o monitoring-agent-linux-amd64

# Linux ARM64 (ARM servers, Raspberry Pi 4)
GOOS=linux GOARCH=arm64 go build -o monitoring-agent-linux-arm64

# Linux ARM (older Raspberry Pi)
GOOS=linux GOARCH=arm go build -o monitoring-agent-linux-arm

# macOS (for testing)
GOOS=darwin GOARCH=amd64 go build -o monitoring-agent-darwin-amd64

# Windows (for testing)
GOOS=windows GOARCH=amd64 go build -o monitoring-agent-windows-amd64.exe
```

## Performance

- **CPU Usage:** ~0.1% idle, ~1-2% during collection
- **Memory Usage:** ~10-20 MB
- **Network:** Minimal (only when metrics requested)
- **Disk I/O:** Minimal (reading /proc files)

## License

MIT License - Feel free to use and modify

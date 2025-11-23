# Service Monitoring Tool

A lightweight monitoring tool built with Go and Gin framework to watch your services and websites.

## Features

- Real-time service/website health monitoring
- **System resource monitoring (CPU, Memory, Disk, Uptime)**
- Web-based dashboard with auto-refresh
- RESTful API for service management
- Configurable check intervals and timeouts
- Response time tracking
- Automatic periodic health checks (every 30 seconds)
- **Telegram notifications for service down/up alerts**
- **Persistent storage - all configurations and services saved to JSON file**
- Auto-save on every change
- Color-coded resource usage indicators (green/yellow/red)

## Project Structure

```
.
├── main.go                 # Application entry point
├── models/
│   ├── service.go         # Data models for monitored services
│   └── store.go           # In-memory service storage
├── services/
│   ├── monitor.go         # Health check logic
│   └── scheduler.go       # Periodic check scheduler
├── handlers/
│   └── service_handler.go # HTTP request handlers
└── templates/
    └── index.html         # Web dashboard UI
```

## Getting Started

### Prerequisites

- Go 1.21 or higher

### Installation

1. Install dependencies:
```bash
go mod tidy
```

2. Run the application:
```bash
go run main.go
```

3. Open your browser and navigate to:
```
http://localhost:8080
```

## Usage

### Web Dashboard

1. Open http://localhost:8080 in your browser

2. **Configure Telegram Notifications (Optional)**:
   - Get your Telegram Bot Token from [@BotFather](https://t.me/botfather)
   - Get your Chat ID (you can use [@userinfobot](https://t.me/userinfobot) or create a group and add your bot)
   - Enter Bot Token and Chat ID in the dashboard
   - Check "Enable" and click "Save Config"
   - Click "Test" to verify the connection

3. Add a new service by filling in:
   - Service Name (e.g., "My API")
   - URL (e.g., "https://api.example.com")
   - Check Interval (seconds, default: 60)
   - Timeout (seconds, default: 10)
4. Click "Add Service"
5. View real-time status updates on the dashboard
6. Receive Telegram alerts when services go down or recover

### API Endpoints

#### Get all services
```bash
GET /api/services
```

#### Get a specific service
```bash
GET /api/services/:id
```

#### Add a new service
```bash
POST /api/services
Content-Type: application/json

{
  "name": "My Service",
  "url": "https://example.com",
  "check_interval": 60,
  "timeout": 10
}
```

#### Update a service
```bash
PUT /api/services/:id
Content-Type: application/json

{
  "name": "Updated Service",
  "url": "https://example.com",
  "check_interval": 120,
  "timeout": 15
}
```

#### Delete a service
```bash
DELETE /api/services/:id
```

#### Check a service immediately
```bash
POST /api/services/:id/check
```

#### Get Telegram configuration
```bash
GET /api/telegram/config
```

#### Update Telegram configuration
```bash
PUT /api/telegram/config
Content-Type: application/json

{
  "bot_token": "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
  "chat_id": "-1001234567890",
  "enabled": true
}
```

#### Test Telegram notification
```bash
POST /api/telegram/test
```

#### Get system information
```bash
GET /api/system/info
```

Returns:
```json
{
  "hostname": "server.local",
  "platform": "darwin",
  "os": "darwin",
  "uptime": 1392533,
  "cpu": {
    "cores": 8,
    "usage_percent": 15.45
  },
  "memory": {
    "total": 8589934592,
    "used": 7296122880,
    "available": 1293811712,
    "used_percent": 84.93
  },
  "disks": [
    {
      "device": "/dev/disk1",
      "mountpoint": "/",
      "total": 245107195904,
      "used": 128847544320,
      "free": 116259651584,
      "used_percent": 52.56
    }
  ]
}
```

## Configuration

### Port

Set the `PORT` environment variable to change the server port (default: 8080):
```bash
PORT=3000 go run main.go
```

### Check Interval

The default background check interval is 30 seconds. To modify this, edit `services/scheduler.go`:
```go
// Change "*/30 * * * * *" to your desired interval
// Format: second minute hour day month weekday
s.cron.AddFunc("*/30 * * * * *", func() {
    // ...
})
```

### Data Persistence

All data is automatically saved to `monitoring_data.json` in the application directory:
- Services and their configurations
- Telegram bot settings
- Service status (preserved across restarts)

The file is created automatically and saved whenever:
- A service is added, updated, or deleted
- Telegram configuration is changed
- Service status is updated (during health checks)

**Backup your data**: Simply copy `monitoring_data.json` to a safe location.

**Restore data**: Replace `monitoring_data.json` with your backup and restart the application.

**Reset everything**: Delete `monitoring_data.json` and restart the application.

## Service Status

- **UP**: Service is responding with HTTP status 200-399
- **DOWN**: Service is not responding or returning HTTP status 400+
- **UNKNOWN**: Service has not been checked yet

## Telegram Notifications

When enabled, you'll receive:
- 🔴 **Down Alert**: Sent when a service goes from UP/UNKNOWN to DOWN
- 🟢 **Recovery Alert**: Sent when a service goes from DOWN to UP

Notifications include:
- Service name and URL
- Current status
- Error message (for down alerts)
- Response time (for up alerts)
- Timestamp

### How to Set Up Telegram Bot

1. Open Telegram and search for [@BotFather](https://t.me/botfather)
2. Send `/newbot` and follow the instructions
3. Copy the bot token (format: `123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11`)
4. Create a new group or use an existing one
5. Add your bot to the group
6. Get the chat ID:
   - Use [@userinfobot](https://t.me/userinfobot) for personal chats
   - For groups, send a message in the group and visit: `https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getUpdates`
   - Look for the `"chat":{"id":` field (usually negative for groups like `-1001234567890`)
7. Enter both values in the dashboard and click "Save Config"
8. Click "Test" to verify the connection

## Features to Add

Consider enhancing the tool with:
- [x] Telegram notifications (✅ Implemented!)
- [x] Persistent storage (✅ Implemented - JSON file)
- [x] System resource monitoring (✅ Implemented - CPU, RAM, Disk, Uptime)
- [ ] Database storage (SQLite, PostgreSQL) for large scale
- [ ] Email/Slack notifications
- [ ] Historical uptime statistics and charts
- [ ] Multiple check types (TCP, ICMP ping, custom scripts)
- [ ] User authentication
- [ ] Alert thresholds and rules for system resources
- [ ] Export monitoring data (CSV)

## License

MIT
